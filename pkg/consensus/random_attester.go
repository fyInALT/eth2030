// random_attester.go implements random attester sampling for the ETH2030
// consensus layer. It selects a deterministic subset of active validators
// per slot (256-1024) using SHA-256-based hashing with optional balance
// weighting, supporting the fast consensus roadmap (EP-2).
package consensus

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math"

	"github.com/eth2030/eth2030/crypto"
)

// Random attester errors.
var (
	ErrRANoValidators     = errors.New("random_attester: no validators")
	ErrRASampleTooLarge   = errors.New("random_attester: sample size exceeds validator count")
	ErrRAZeroSampleSize   = errors.New("random_attester: zero sample size")
	ErrRAInvalidValidator = errors.New("random_attester: invalid validator")
)

// RandomAttesterConfig configures random attester sampling.
type RandomAttesterConfig struct {
	// MinSampleSize is the minimum number of attesters per slot.
	MinSampleSize uint64 // Default: 256
	// MaxSampleSize is the maximum number of attesters per slot.
	MaxSampleSize uint64 // Default: 1024
	// BalanceWeighted enables balance-weighted sampling.
	BalanceWeighted bool // Default: true
}

// ValidatorInfo holds minimal validator data for sampling.
type ValidatorInfo struct {
	Index            ValidatorIndex
	EffectiveBalance uint64 // in Gwei
	Active           bool
}

// RandomAttesterSelector performs random attester sampling.
type RandomAttesterSelector struct {
	config RandomAttesterConfig
}

// DefaultRandomAttesterConfig returns defaults: 256 min, 1024 max, balance-weighted.
func DefaultRandomAttesterConfig() *RandomAttesterConfig {
	return &RandomAttesterConfig{
		MinSampleSize:   256,
		MaxSampleSize:   1024,
		BalanceWeighted: true,
	}
}

// NewRandomAttesterSelector creates a new selector with the given config.
// If config is nil, defaults are used.
func NewRandomAttesterSelector(config *RandomAttesterConfig) *RandomAttesterSelector {
	if config == nil {
		config = DefaultRandomAttesterConfig()
	}
	return &RandomAttesterSelector{config: *config}
}

// SampleValidators selects a deterministic subset of active validators for
// the given slot and seed. The selection is unique (no duplicates) and
// optionally balance-weighted. Returns the selected validator indices.
func (s *RandomAttesterSelector) SampleValidators(
	validators []ValidatorInfo,
	slot uint64,
	seed [32]byte,
) ([]ValidatorIndex, error) {
	if len(validators) == 0 {
		return nil, ErrRANoValidators
	}

	// Filter active validators.
	active := make([]ValidatorInfo, 0, len(validators))
	for i := range validators {
		if validators[i].Active {
			active = append(active, validators[i])
		}
	}
	if len(active) == 0 {
		return nil, ErrRANoValidators
	}

	activeCount := uint64(len(active))
	sampleSize := s.ComputeSampleSize(activeCount)
	if sampleSize == 0 {
		return nil, ErrRAZeroSampleSize
	}

	// If sample size >= active count, return all active validators.
	if sampleSize >= activeCount {
		result := make([]ValidatorIndex, activeCount)
		for i, v := range active {
			result[i] = v.Index
		}
		return result, nil
	}

	// Compute total effective balance for weighted sampling.
	var totalBalance uint64
	if s.config.BalanceWeighted {
		for _, v := range active {
			totalBalance += v.EffectiveBalance
		}
		// Fallback to uniform if total balance is zero.
		if totalBalance == 0 {
			totalBalance = 0 // will trigger uniform path below
		}
	}

	// Select validators deterministically.
	selected := make(map[ValidatorIndex]struct{})
	result := make([]ValidatorIndex, 0, sampleSize)

	// We iterate through candidates using increasing index values.
	// For each position, we hash (seed || slot || attempt) to get randomness.
	for attempt := uint64(0); len(result) < int(sampleSize); attempt++ {
		// Safety bound: prevent infinite loop for edge cases.
		if attempt > activeCount*100 {
			break
		}

		h := computeSamplingHash(seed, slot, attempt)
		candidateIdx := binary.LittleEndian.Uint64(h[:8]) % activeCount
		candidate := active[candidateIdx]

		// Skip if already selected.
		if _, exists := selected[candidate.Index]; exists {
			continue
		}

		// Balance-weighted acceptance check (similar to ComputeProposerIndex).
		if s.config.BalanceWeighted && totalBalance > 0 {
			randByte := h[8]
			// Max effective balance for normalization (32 ETH = 32e9 Gwei).
			maxEffBal := uint64(32_000_000_000)
			if candidate.EffectiveBalance*255 < maxEffBal*uint64(randByte) {
				continue
			}
		}

		selected[candidate.Index] = struct{}{}
		result = append(result, candidate.Index)
	}

	if len(result) == 0 {
		return nil, ErrRANoValidators
	}

	return result, nil
}

// ComputeSampleSize returns the sample size for a given active validator count.
// It computes sqrt(activeValidatorCount) and clamps to [MinSampleSize, MaxSampleSize].
// If activeValidatorCount is less than MinSampleSize, returns activeValidatorCount.
func (s *RandomAttesterSelector) ComputeSampleSize(activeValidatorCount uint64) uint64 {
	if activeValidatorCount == 0 {
		return 0
	}

	sqrtCount := uint64(math.Sqrt(float64(activeValidatorCount)))
	if sqrtCount < 1 {
		sqrtCount = 1
	}

	// Clamp to [MinSampleSize, MaxSampleSize].
	size := sqrtCount
	if size < s.config.MinSampleSize {
		size = s.config.MinSampleSize
	}
	if size > s.config.MaxSampleSize {
		size = s.config.MaxSampleSize
	}

	// Cannot exceed active validator count.
	if size > activeValidatorCount {
		size = activeValidatorCount
	}

	return size
}

// computeSamplingHash computes a deterministic hash for sampling given
// a seed, slot number, and index. Uses SHA-256 over a 48-byte buffer.
func computeSamplingHash(seed [32]byte, slot, index uint64) [32]byte {
	var buf [48]byte
	copy(buf[:32], seed[:])
	binary.LittleEndian.PutUint64(buf[32:40], slot)
	binary.LittleEndian.PutUint64(buf[40:48], index)
	return sha256.Sum256(buf[:])
}

// SelectAttesters selects a deterministic subset of validators for the given
// slot and RANDAO reveal (GAP-3.1). The seed is keccak256(slot_bytes || randao)
// and the selection uses a Fisher-Yates shuffle seeded from the keccak hash,
// returning the first SampleSize elements.
//
// This is the committee-less API: validators are identified by ValidatorIndex
// directly (no CommitteeBits). The sample size is clamped to [MinSampleSize,
// MaxSampleSize] and cannot exceed len(validators).
func (s *RandomAttesterSelector) SelectAttesters(
	slot uint64,
	randao []byte,
	validators []ValidatorIndex,
) []ValidatorIndex {
	if len(validators) == 0 {
		return nil
	}

	// Derive deterministic seed: keccak256(slot_bytes || randao).
	var slotBytes [8]byte
	binary.LittleEndian.PutUint64(slotBytes[:], slot)
	seed := crypto.Keccak256(append(slotBytes[:], randao...))
	if len(seed) < 32 {
		return nil
	}

	n := uint64(len(validators))
	sampleSize := s.ComputeSampleSize(n)
	if sampleSize == 0 {
		return nil
	}

	// Fisher-Yates partial shuffle: swap first sampleSize elements.
	// Work on a copy to avoid mutating the caller's slice.
	shuffled := make([]ValidatorIndex, n)
	copy(shuffled, validators)

	for i := uint64(0); i < sampleSize; i++ {
		// Hash position i with the base seed to get a random swap index.
		var posBuf [40]byte
		copy(posBuf[:32], seed[:32])
		binary.LittleEndian.PutUint64(posBuf[32:40], i)
		h := sha256.Sum256(posBuf[:])
		j := binary.LittleEndian.Uint64(h[:8])%(n-i) + i
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled[:sampleSize]
}
