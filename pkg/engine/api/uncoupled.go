// uncoupled.go implements EIP-7898 uncoupled execution payload logic.
package api

import (
	"errors"
	"fmt"
	"sync"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
	"github.com/eth2030/eth2030/engine/apierrors"
	"github.com/eth2030/eth2030/engine/payload"
)

// InclusionProofDepth is the depth of the Merkle inclusion proof tree.
const InclusionProofDepth = 4

// UncoupledPayloadStatus values.
const (
	UncoupledStatusPending  = "PENDING"
	UncoupledStatusVerified = "VERIFIED"
	UncoupledStatusInvalid  = "INVALID"
)

// InclusionProof is a Merkle proof demonstrating that the execution payload
// was committed to in the beacon block via the ExecutionPayloadHeader.
type InclusionProof struct {
	// Leaf is the hash of the full execution payload that was committed.
	Leaf types.Hash `json:"leaf"`
	// Branch contains the sibling hashes along the path from the leaf
	// to the beacon block body root.
	Branch []types.Hash `json:"branch"`
	// Index is the generalized index of the execution payload leaf.
	Index uint64 `json:"index"`
}

// UncoupledPayloadEnvelope wraps a full execution payload with the inclusion
// proof linking it to a beacon block. This is gossiped independently of the
// beacon block per EIP-7898.
type UncoupledPayloadEnvelope struct {
	BeaconBlockRoot types.Hash                  `json:"beaconBlockRoot"`
	Slot            uint64                      `json:"slot"`
	Payload         *payload.ExecutionPayloadV5 `json:"executionPayload"`
	Proof           *InclusionProof             `json:"inclusionProof"`
}

// Errors for uncoupled payload handling.
var (
	ErrMissingInclusionProof  = errors.New("engine: missing inclusion proof")
	ErrInvalidInclusionProof  = errors.New("engine: invalid inclusion proof")
	ErrInclusionProofMismatch = errors.New("engine: inclusion proof root mismatch")
	// ErrMissingBeaconRoot is aliased to apierrors to match the engine package sentinel.
	ErrMissingBeaconRoot = apierrors.ErrMissingBeaconRoot
)

// Validate performs basic structural validation of the envelope.
func (e *UncoupledPayloadEnvelope) Validate() error {
	if e.BeaconBlockRoot == (types.Hash{}) {
		return ErrMissingBeaconRoot
	}
	if e.Slot == 0 {
		return apierrors.ErrInvalidParams
	}
	if e.Payload == nil {
		return apierrors.ErrInvalidPayloadAttributes
	}
	if e.Proof == nil {
		return ErrMissingInclusionProof
	}
	return nil
}

// PayloadHash computes the hash of the execution payload for proof verification.
func (e *UncoupledPayloadEnvelope) PayloadHash() types.Hash {
	return e.Payload.BlockHash
}

// ValidateInclusionProof verifies that the inclusion proof correctly links
// the payload to the beacon block root.
func ValidateInclusionProof(proof *InclusionProof, beaconBlockRoot types.Hash) error {
	if proof == nil {
		return ErrMissingInclusionProof
	}
	if proof.Leaf == (types.Hash{}) {
		return ErrInvalidInclusionProof
	}
	if len(proof.Branch) == 0 {
		return ErrInvalidInclusionProof
	}

	computed := computeMerkleRoot(proof.Leaf, proof.Branch, proof.Index)
	if computed != beaconBlockRoot {
		return fmt.Errorf("%w: computed root %s != beacon block root %s",
			ErrInclusionProofMismatch, computed.Hex(), beaconBlockRoot.Hex())
	}
	return nil
}

// computeMerkleRoot walks the proof branch from leaf to root.
func computeMerkleRoot(leaf types.Hash, branch []types.Hash, index uint64) types.Hash {
	current := leaf
	idx := index
	for _, sibling := range branch {
		if idx%2 == 0 {
			current = HashPair(current, sibling)
		} else {
			current = HashPair(sibling, current)
		}
		idx /= 2
	}
	return current
}

// HashPair hashes two 32-byte values together using Keccak-256.
func HashPair(left, right types.Hash) types.Hash {
	var data [64]byte
	copy(data[:32], left[:])
	copy(data[32:], right[:])
	return crypto.Keccak256Hash(data[:])
}

// BuildInclusionProof creates an inclusion proof for an execution payload
// given the beacon block body fields.
func BuildInclusionProof(payloadHash types.Hash, bodyFieldHashes []types.Hash, payloadIndex int) (*InclusionProof, error) {
	if len(bodyFieldHashes) == 0 {
		return nil, errors.New("engine: empty body field hashes")
	}
	if payloadIndex < 0 || payloadIndex >= len(bodyFieldHashes) {
		return nil, fmt.Errorf("engine: payload index %d out of range [0, %d)", payloadIndex, len(bodyFieldHashes))
	}

	leaves := PadToPowerOfTwo(bodyFieldHashes)
	branch := collectProofBranch(leaves, payloadIndex)

	return &InclusionProof{
		Leaf:   payloadHash,
		Branch: branch,
		Index:  uint64(payloadIndex),
	}, nil
}

// PadToPowerOfTwo pads a slice of hashes to the next power of two with zero hashes.
func PadToPowerOfTwo(hashes []types.Hash) []types.Hash {
	n := len(hashes)
	size := 1
	for size < n {
		size *= 2
	}
	padded := make([]types.Hash, size)
	copy(padded, hashes)
	return padded
}

// collectProofBranch builds a Merkle tree from leaves and collects the
// sibling hashes along the path from the target leaf to the root.
func collectProofBranch(leaves []types.Hash, targetIndex int) []types.Hash {
	var branch []types.Hash
	layer := make([]types.Hash, len(leaves))
	copy(layer, leaves)
	idx := targetIndex

	for len(layer) > 1 {
		if idx%2 == 0 {
			if idx+1 < len(layer) {
				branch = append(branch, layer[idx+1])
			} else {
				branch = append(branch, types.Hash{})
			}
		} else {
			branch = append(branch, layer[idx-1])
		}

		nextLayer := make([]types.Hash, len(layer)/2)
		for i := 0; i < len(layer); i += 2 {
			nextLayer[i/2] = HashPair(layer[i], layer[i+1])
		}
		layer = nextLayer
		idx /= 2
	}
	return branch
}

// pendingUncoupled tracks an uncoupled payload waiting for or having received
// its execution payload.
type pendingUncoupled struct {
	envelope *UncoupledPayloadEnvelope
	status   string
}

// UncoupledBackend is the minimal backend interface for UncoupledPayloadHandler.
// It is intentionally empty — the handler does not need to call the backend
// for validation, only for forwarding validated payloads (future use).
type UncoupledBackend interface{}

// UncoupledPayloadHandler manages the receipt and validation of uncoupled
// execution payloads per EIP-7898.
type UncoupledPayloadHandler struct {
	mu      sync.RWMutex
	pending map[types.Hash]*pendingUncoupled
	backend UncoupledBackend
}

// NewUncoupledPayloadHandler creates a new handler for uncoupled payloads.
func NewUncoupledPayloadHandler(backend UncoupledBackend) *UncoupledPayloadHandler {
	return &UncoupledPayloadHandler{
		pending: make(map[types.Hash]*pendingUncoupled),
		backend: backend,
	}
}

// SubmitUncoupledPayload receives an uncoupled execution payload with its
// inclusion proof.
func (h *UncoupledPayloadHandler) SubmitUncoupledPayload(envelope *UncoupledPayloadEnvelope) (string, error) {
	if err := envelope.Validate(); err != nil {
		return UncoupledStatusInvalid, err
	}

	if err := ValidateInclusionProof(envelope.Proof, envelope.BeaconBlockRoot); err != nil {
		return UncoupledStatusInvalid, err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if existing, ok := h.pending[envelope.BeaconBlockRoot]; ok {
		if existing.status == UncoupledStatusVerified {
			return UncoupledStatusVerified, nil
		}
	}

	h.pending[envelope.BeaconBlockRoot] = &pendingUncoupled{
		envelope: envelope,
		status:   UncoupledStatusVerified,
	}

	return UncoupledStatusVerified, nil
}

// VerifyInclusion checks whether a payload's inclusion proof is valid
// against the specified beacon block root without storing the payload.
func (h *UncoupledPayloadHandler) VerifyInclusion(envelope *UncoupledPayloadEnvelope) error {
	if err := envelope.Validate(); err != nil {
		return err
	}
	return ValidateInclusionProof(envelope.Proof, envelope.BeaconBlockRoot)
}

// GetPendingPayload retrieves a previously submitted uncoupled payload by
// its beacon block root.
func (h *UncoupledPayloadHandler) GetPendingPayload(beaconBlockRoot types.Hash) *UncoupledPayloadEnvelope {
	h.mu.RLock()
	defer h.mu.RUnlock()

	p, ok := h.pending[beaconBlockRoot]
	if !ok {
		return nil
	}
	return p.envelope
}

// GetPayloadStatus returns the status of an uncoupled payload.
func (h *UncoupledPayloadHandler) GetPayloadStatus(beaconBlockRoot types.Hash) string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	p, ok := h.pending[beaconBlockRoot]
	if !ok {
		return UncoupledStatusPending
	}
	return p.status
}

// RemovePending removes a pending uncoupled payload entry.
func (h *UncoupledPayloadHandler) RemovePending(beaconBlockRoot types.Hash) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.pending, beaconBlockRoot)
}

// PendingCount returns the number of pending uncoupled payloads.
func (h *UncoupledPayloadHandler) PendingCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.pending)
}
