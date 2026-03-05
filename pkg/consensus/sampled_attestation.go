// sampled_attestation.go defines the SampledAttestation and SampledAggregate
// types for the committee-less attestation format (GAP-3.2). Unlike the
// post-Electra Attestation (EIP-7549) which includes CommitteeBits, sampled
// attestations list their validators explicitly, eliminating the CommitteeBits
// field entirely.
//
// SampledAttestation is produced by RandomAttesterSelector.SelectAttesters
// and aggregated into SampledAggregate for fork-choice weight computation.
package consensus

import (
	"errors"

	"github.com/eth2030/eth2030/core/types"
)

// Sampled attestation errors.
var (
	ErrSampledAttNilData      = errors.New("sampled_att: nil attestation data")
	ErrSampledAttEmptySig     = errors.New("sampled_att: empty signature")
	ErrSampledAttNoValidators = errors.New("sampled_att: no validator indices")
	ErrSampledAttDataMismatch = errors.New("sampled_att: data mismatch in aggregate")
	ErrSampledAttDuplicate    = errors.New("sampled_att: duplicate validator in aggregate")
)

// SampledAttestation is a single attestation from a randomly selected
// validator subset. It omits CommitteeBits because validators are identified
// by explicit index, not committee position (GAP-3.2).
type SampledAttestation struct {
	// Data is the signed attestation data (slot, block root, source/target).
	Data AttestationData
	// ValidatorIndices lists the validators included in this attestation.
	// For a single-validator attestation this has exactly one element.
	ValidatorIndices []ValidatorIndex
	// Signature is the BLS signature over Data.
	Signature [96]byte
}

// SampledAggregate is an aggregate of multiple SampledAttestations that
// all share the same AttestationData. It accumulates the combined stake
// weight for fork-choice computation (GAP-3.2).
type SampledAggregate struct {
	// Data is the common attestation data shared by all aggregated attestations.
	Data AttestationData
	// ValidatorIndices is the union of all validator indices in the aggregate.
	ValidatorIndices []ValidatorIndex
	// AggregatedSignature is the BLS aggregate signature.
	AggregatedSignature [96]byte
	// TotalStake is the sum of effective balances for all included validators.
	TotalStake uint64
}

// NewSampledAttestation creates a SampledAttestation for a single validator.
func NewSampledAttestation(
	data AttestationData,
	validatorIdx ValidatorIndex,
	sig [96]byte,
) *SampledAttestation {
	return &SampledAttestation{
		Data:             data,
		ValidatorIndices: []ValidatorIndex{validatorIdx},
		Signature:        sig,
	}
}

// Validate checks that the attestation has non-nil data and a non-empty signature.
func (a *SampledAttestation) Validate() error {
	if len(a.ValidatorIndices) == 0 {
		return ErrSampledAttNoValidators
	}
	var emptySig [96]byte
	if a.Signature == emptySig {
		return ErrSampledAttEmptySig
	}
	return nil
}

// AggregateSampledAttestations aggregates N SampledAttestations with identical
// AttestationData into a single SampledAggregate. Returns an error if data
// mismatches or duplicate validators are detected (GAP-3.2).
//
// NOTE: This implementation XOR-accumulates signatures as a placeholder for
// the real BLS aggregate. In production, use BLS FastAggregateVerify.
func AggregateSampledAttestations(atts []*SampledAttestation) (*SampledAggregate, error) {
	if len(atts) == 0 {
		return nil, ErrSampledAttNoValidators
	}

	ref := atts[0].Data
	seen := make(map[ValidatorIndex]bool)
	allIndices := make([]ValidatorIndex, 0)
	var aggSig [96]byte

	for _, att := range atts {
		// All attestations must share the same data.
		if !IsEqualAttestationData(&att.Data, &ref) {
			return nil, ErrSampledAttDataMismatch
		}
		for _, idx := range att.ValidatorIndices {
			if seen[idx] {
				return nil, ErrSampledAttDuplicate
			}
			seen[idx] = true
			allIndices = append(allIndices, idx)
		}
		// XOR aggregate signatures (placeholder; replace with BLS aggregate).
		for i := range aggSig {
			aggSig[i] ^= att.Signature[i]
		}
	}

	return &SampledAggregate{
		Data:                ref,
		ValidatorIndices:    allIndices,
		AggregatedSignature: aggSig,
	}, nil
}

// ScaleSampledWeight computes the fork-choice vote weight for a SampledAggregate.
// Because only a sample of the full committee voted, the weight is scaled by
// full_committee_size / sample_size to reflect the expected full-committee weight
// (GAP-3.3).
//
//	scaledWeight = baseWeight * fullCommitteeSize / sampleSize
func ScaleSampledWeight(baseWeight, fullCommitteeSize, sampleSize uint64) uint64 {
	if sampleSize == 0 {
		return 0
	}
	return baseWeight * fullCommitteeSize / sampleSize
}

// SampledForkChoiceWeight returns the effective fork-choice weight for a
// SampledAggregate by scaling each validator's stake by the committee ratio
// and summing the result (GAP-3.3).
//
// Parameters:
//   - agg:               the sampled aggregate to weight
//   - validatorStakes:   map from ValidatorIndex to effective balance in Gwei
//   - fullCommitteeSize: the total number of validators in the full committee
func SampledForkChoiceWeight(
	agg *SampledAggregate,
	validatorStakes map[ValidatorIndex]uint64,
	fullCommitteeSize uint64,
) uint64 {
	if agg == nil || fullCommitteeSize == 0 {
		return 0
	}

	sampleSize := uint64(len(agg.ValidatorIndices))
	if sampleSize == 0 {
		return 0
	}

	var totalBase uint64
	for _, idx := range agg.ValidatorIndices {
		totalBase += validatorStakes[idx]
	}

	return ScaleSampledWeight(totalBase, fullCommitteeSize, sampleSize)
}

// AttesterSampleSize is the number of attesters selected per slot when sampled
// mode is active. Valid values are 0 (full committee), 256, 512, or 1024.
type AttesterSampleSize uint64

const (
	// AttesterSampleFull uses the full committee (sampled mode off).
	AttesterSampleFull AttesterSampleSize = 0
	// AttesterSample256 samples 256 attesters per slot.
	AttesterSample256 AttesterSampleSize = 256
	// AttesterSample512 samples 512 attesters per slot (default for lean mode).
	AttesterSample512 AttesterSampleSize = 512
	// AttesterSample1024 samples 1024 attesters per slot.
	AttesterSample1024 AttesterSampleSize = 1024
)

// AttestationMode determines which attestation format is used.
type AttestationMode uint8

const (
	// AttestationModeFull uses the standard Attestation type with CommitteeBits.
	AttestationModeFull AttestationMode = iota
	// AttestationModeSampled uses SampledAttestation without CommitteeBits.
	AttestationModeSampled
)

// ResolveAttestationMode returns the AttestationMode for the given sample size.
// A sample size of 0 means full-committee mode; any non-zero size means sampled.
func ResolveAttestationMode(sampleSize AttesterSampleSize) AttestationMode {
	if sampleSize == AttesterSampleFull {
		return AttestationModeFull
	}
	return AttestationModeSampled
}

// SampledBlockRoot is a block root and its associated full-weight from sampled
// attestations. Used by the fork-choice rule to select the canonical head.
type SampledBlockRoot struct {
	Root   types.Hash
	Weight uint64
}
