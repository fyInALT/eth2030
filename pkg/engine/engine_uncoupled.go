// engine_uncoupled.go re-exports EIP-7898 uncoupled payload symbols from
// engine/api and provides backward-compatible unexported wrappers for tests.
package engine

import (
	"github.com/eth2030/eth2030/core/types"
	engapi "github.com/eth2030/eth2030/engine/api"
)

// UncoupledPayloadStatus values (re-exported for backward compatibility).
const (
	UncoupledStatusPending  = engapi.UncoupledStatusPending
	UncoupledStatusVerified = engapi.UncoupledStatusVerified
	UncoupledStatusInvalid  = engapi.UncoupledStatusInvalid
	InclusionProofDepth     = engapi.InclusionProofDepth
)

// Error re-exports for backward compatibility.
var (
	ErrMissingInclusionProof  = engapi.ErrMissingInclusionProof
	ErrInvalidInclusionProof  = engapi.ErrInvalidInclusionProof
	ErrInclusionProofMismatch = engapi.ErrInclusionProofMismatch
)

// NewUncoupledPayloadHandler creates a new handler for uncoupled payloads.
func NewUncoupledPayloadHandler(backend interface{}) *UncoupledPayloadHandler {
	return engapi.NewUncoupledPayloadHandler(backend)
}

// ValidateInclusionProof re-exports engapi.ValidateInclusionProof for backward compatibility.
func ValidateInclusionProof(proof *InclusionProof, beaconBlockRoot types.Hash) error {
	return engapi.ValidateInclusionProof(proof, beaconBlockRoot)
}

// BuildInclusionProof re-exports engapi.BuildInclusionProof for backward compatibility.
func BuildInclusionProof(payloadHash types.Hash, bodyFieldHashes []types.Hash, payloadIndex int) (*InclusionProof, error) {
	return engapi.BuildInclusionProof(payloadHash, bodyFieldHashes, payloadIndex)
}

// hashPair wraps the exported HashPair for package-internal test use.
func hashPair(left, right types.Hash) types.Hash {
	return engapi.HashPair(left, right)
}

// padToPowerOfTwo wraps the exported PadToPowerOfTwo for package-internal test use.
func padToPowerOfTwo(hashes []types.Hash) []types.Hash {
	return engapi.PadToPowerOfTwo(hashes)
}

// computeMerkleRoot is an unexported wrapper for package-internal test use.
func computeMerkleRoot(leaf types.Hash, branch []types.Hash, index uint64) types.Hash {
	// Replicate the logic from engapi.computeMerkleRoot (unexported there).
	current := leaf
	idx := index
	for _, sibling := range branch {
		if idx%2 == 0 {
			current = hashPair(current, sibling)
		} else {
			current = hashPair(sibling, current)
		}
		idx /= 2
	}
	return current
}
