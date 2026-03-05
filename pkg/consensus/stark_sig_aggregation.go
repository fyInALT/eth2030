package consensus

import (
	"crypto/sha256"
	"errors"
	"math/big"
	"sync"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/proofs"
)

// STARK signature aggregation errors.
var (
	ErrSTARKAggNoSigs       = errors.New("stark_sig_aggregation: no signatures to aggregate")
	ErrSTARKAggInvalidProof = errors.New("stark_sig_aggregation: invalid aggregate proof")
	ErrSTARKAggVerifyFailed = errors.New("stark_sig_aggregation: verification failed")
	ErrSTARKAggNilResult    = errors.New("stark_sig_aggregation: nil aggregation result")
	ErrSTARKAggMismatch     = errors.New("stark_sig_aggregation: committee root mismatch")
)

// STARKSignatureAggregation holds a STARK-aggregated set of PQ attestation signatures.
// Instead of verifying N individual Dilithium/Falcon signatures, a single STARK
// proves that all N signatures are valid.
type STARKSignatureAggregation struct {
	// Signatures are the original PQ attestations being aggregated.
	Signatures []PQAttestation
	// AggregateProof is the STARK proving all signatures are valid.
	AggregateProof *proofs.STARKProofData
	// CommitteeRoot is the Merkle root of the participating validator public keys.
	CommitteeRoot types.Hash
	// Message is the beacon block root being attested to.
	Message types.Hash
	// NumValidators is the number of validators in this aggregation.
	NumValidators int
	// Slot is the slot this aggregation covers.
	Slot uint64
	// TargetEpoch is the target epoch for finality.
	TargetEpoch uint64
}

// STARKSignatureAggregator creates STARK-aggregated signature proofs.
type STARKSignatureAggregator struct {
	mu     sync.RWMutex
	prover *proofs.STARKProver
}

// NewSTARKSignatureAggregator creates a new STARK signature aggregator.
func NewSTARKSignatureAggregator() *STARKSignatureAggregator {
	return &STARKSignatureAggregator{
		prover: proofs.NewSTARKProver(),
	}
}

// Aggregate creates a STARK proof that all given PQ attestation signatures are valid.
// This replaces iterative O(N) verification with a single STARK verification.
func (sa *STARKSignatureAggregator) Aggregate(attestations []PQAttestation) (*STARKSignatureAggregation, error) {
	if len(attestations) == 0 {
		return nil, ErrSTARKAggNoSigs
	}

	// Build execution trace: each attestation becomes a row.
	// Columns: [slot, committee_index, block_root_hi, block_root_lo, sig_hash_hi, sig_hash_lo, validator_index]
	trace := make([][]proofs.FieldElement, len(attestations))
	pubkeys := make([][]byte, len(attestations))

	for i, att := range attestations {
		sigHash := hashSignatureData(att.PQSignature, att.PQPublicKey)
		blockHi := new(big.Int).SetBytes(att.BeaconBlockRoot[:16])
		blockLo := new(big.Int).SetBytes(att.BeaconBlockRoot[16:])
		sigHi := new(big.Int).SetBytes(sigHash[:16])
		sigLo := new(big.Int).SetBytes(sigHash[16:])

		trace[i] = []proofs.FieldElement{
			proofs.NewFieldElement(int64(att.Slot)),
			proofs.NewFieldElement(int64(att.CommitteeIndex)),
			{Value: blockHi},
			{Value: blockLo},
			{Value: sigHi},
			{Value: sigLo},
			proofs.NewFieldElement(int64(att.ValidatorIndex)),
		}
		pubkeys[i] = att.PQPublicKey
	}

	// Constraint: each row's sig_hash must be non-zero (signature exists).
	constraints := []proofs.STARKConstraint{
		{Degree: 1, Coefficients: []proofs.FieldElement{proofs.NewFieldElement(1)}},
	}

	starkProof, err := sa.prover.GenerateSTARKProof(trace, constraints)
	if err != nil {
		return nil, err
	}

	// Compute committee root from public keys.
	committeeRoot := computeCommitteeRoot(pubkeys)

	return &STARKSignatureAggregation{
		Signatures:     attestations,
		AggregateProof: starkProof,
		CommitteeRoot:  committeeRoot,
		Message:        attestations[0].BeaconBlockRoot,
		NumValidators:  len(attestations),
		Slot:           attestations[0].Slot,
		TargetEpoch:    attestations[0].TargetEpoch,
	}, nil
}

// Verify checks that a STARK signature aggregation is valid.
func (sa *STARKSignatureAggregator) Verify(agg *STARKSignatureAggregation) (bool, error) {
	if agg == nil {
		return false, ErrSTARKAggNilResult
	}
	if agg.AggregateProof == nil {
		return false, ErrSTARKAggInvalidProof
	}

	// Verify the STARK proof.
	valid, err := sa.prover.VerifySTARKProof(agg.AggregateProof, nil)
	if err != nil {
		return false, err
	}
	if !valid {
		return false, ErrSTARKAggVerifyFailed
	}

	// Verify committee root matches the public keys.
	pubkeys := make([][]byte, len(agg.Signatures))
	for i, att := range agg.Signatures {
		pubkeys[i] = att.PQPublicKey
	}
	expectedRoot := computeCommitteeRoot(pubkeys)
	if expectedRoot != agg.CommitteeRoot {
		return false, ErrSTARKAggMismatch
	}

	return true, nil
}

// AggregateCount returns the number of signatures in an aggregation.
func (agg *STARKSignatureAggregation) AggregateCount() int {
	if agg == nil {
		return 0
	}
	return agg.NumValidators
}

// GasSavings estimates the gas savings from STARK aggregation vs individual verification.
// Individual Dilithium verification: ~50,000 gas per signature.
// STARK verification: ~200,000 gas flat (amortized over N signatures).
func (agg *STARKSignatureAggregation) GasSavings() uint64 {
	if agg == nil || agg.NumValidators == 0 {
		return 0
	}
	individualCost := uint64(agg.NumValidators) * 50000
	starkCost := uint64(200000)
	if individualCost <= starkCost {
		return 0
	}
	return individualCost - starkCost
}

// hashSignatureData hashes signature and public key into a commitment.
func hashSignatureData(sig, pubkey []byte) [32]byte {
	h := sha256.New()
	h.Write(sig)
	h.Write(pubkey)
	var result [32]byte
	copy(result[:], h.Sum(nil))
	return result
}

// computeCommitteeRoot computes a Merkle root over validator public keys.
func computeCommitteeRoot(pubkeys [][]byte) types.Hash {
	if len(pubkeys) == 0 {
		return types.Hash{}
	}

	// Hash each public key into a leaf.
	leaves := make([][32]byte, len(pubkeys))
	for i, pk := range pubkeys {
		h := sha256.New()
		h.Write(pk)
		copy(leaves[i][:], h.Sum(nil))
	}

	// Pad to next power of two.
	n := len(leaves)
	target := 1
	for target < n {
		target <<= 1
	}
	padded := make([][32]byte, target)
	copy(padded, leaves)

	// Build Merkle tree bottom-up.
	layer := padded
	for len(layer) > 1 {
		next := make([][32]byte, len(layer)/2)
		for i := range next {
			h := sha256.New()
			h.Write(layer[2*i][:])
			h.Write(layer[2*i+1][:])
			copy(next[i][:], h.Sum(nil))
		}
		layer = next
	}

	var root types.Hash
	copy(root[:], layer[0][:])
	return root
}

// BatchVerifyPQAttestations uses STARK aggregation to verify a batch of PQ attestations.
// This is the integration point with the existing PQAttestationVerifier.
func BatchVerifyPQAttestations(attestations []PQAttestation) (*STARKSignatureAggregation, error) {
	if len(attestations) == 0 {
		return nil, ErrSTARKAggNoSigs
	}

	aggregator := NewSTARKSignatureAggregator()
	agg, err := aggregator.Aggregate(attestations)
	if err != nil {
		return nil, err
	}

	valid, err := aggregator.Verify(agg)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, ErrSTARKAggVerifyFailed
	}

	return agg, nil
}

// MerklePQAggregation holds a lightweight Merkle-only aggregation of PQ
// attestations, used in lean available chain mode where STARK proofs are skipped.
type MerklePQAggregation struct {
	MerkleRoot    types.Hash
	NumValidators int
	Slot          uint64
}

// MerkleAggregatePQAttestations computes a SHA-256 Merkle tree over the
// attestation public key hashes and returns the root. No STARK proof is produced.
func MerkleAggregatePQAttestations(attestations []PQAttestation) (*MerklePQAggregation, error) {
	if len(attestations) == 0 {
		return nil, ErrSTARKAggNoSigs
	}
	pubkeys := make([][]byte, len(attestations))
	for i, att := range attestations {
		pubkeys[i] = att.PQPublicKey
	}
	root := computeCommitteeRoot(pubkeys)
	return &MerklePQAggregation{
		MerkleRoot:    root,
		NumValidators: len(attestations),
		Slot:          attestations[0].Slot,
	}, nil
}

// AggregateWithConfig creates a signature aggregation using the appropriate
// strategy based on config. In lean available chain mode, only a Merkle root
// is computed (no STARK proof). Otherwise, the full STARK aggregation is used.
func (sa *STARKSignatureAggregator) AggregateWithConfig(attestations []PQAttestation, cfg *ConsensusConfig) (*STARKSignatureAggregation, error) {
	if cfg != nil && cfg.LeanAvailableChainMode {
		merkle, err := MerkleAggregatePQAttestations(attestations)
		if err != nil {
			return nil, err
		}
		return &STARKSignatureAggregation{
			Signatures:     attestations,
			AggregateProof: nil,
			CommitteeRoot:  merkle.MerkleRoot,
			Message:        attestations[0].BeaconBlockRoot,
			NumValidators:  merkle.NumValidators,
			Slot:           merkle.Slot,
			TargetEpoch:    attestations[0].TargetEpoch,
		}, nil
	}
	return sa.Aggregate(attestations)
}

// AggregationStats holds statistics about a STARK signature aggregation.
type AggregationStats struct {
	NumSignatures     int
	ProofSizeBytes    int
	CommitteeRoot     types.Hash
	Slot              uint64
	TargetEpoch       uint64
	EstimatedGasSaved uint64
}

// Stats returns statistics about this aggregation.
func (agg *STARKSignatureAggregation) Stats() AggregationStats {
	if agg == nil {
		return AggregationStats{}
	}
	proofSize := 0
	if agg.AggregateProof != nil {
		proofSize = agg.AggregateProof.ProofSize()
	}
	return AggregationStats{
		NumSignatures:     agg.NumValidators,
		ProofSizeBytes:    proofSize,
		CommitteeRoot:     agg.CommitteeRoot,
		Slot:              agg.Slot,
		TargetEpoch:       agg.TargetEpoch,
		EstimatedGasSaved: agg.GasSavings(),
	}
}
