// checkpoint_validator.go implements checkpoint trust validation with weak
// subjectivity proofs and historical consistency checking for light client
// security. This is part of the CL roadmap: fast confirmation -> 1-epoch
// finality -> light client checkpoint bootstrapping.
package light

import (
	"errors"
	"math"
	"sync"
)

// Checkpoint validator errors.
var (
	ErrCheckpointRootZero      = errors.New("light: checkpoint root is zero")
	ErrCheckpointStateRootZero = errors.New("light: checkpoint state root is zero")
	ErrCheckpointEpochZero     = errors.New("light: checkpoint epoch is zero")
	ErrCheckpointTooOld        = errors.New("light: checkpoint exceeds maximum age")
	ErrCheckpointTrustLow      = errors.New("light: checkpoint trust score below minimum")
	ErrWSProofInvalid          = errors.New("light: weak subjectivity proof is invalid")
	ErrWSProofExpired          = errors.New("light: weak subjectivity proof has expired")
	ErrWSPeriodExceeded        = errors.New("light: checkpoint is outside weak subjectivity period")
	ErrValidatorCountZero      = errors.New("light: validator count must be non-zero")
	ErrCurrentEpochBehind      = errors.New("light: current epoch is behind checkpoint epoch")
)

// Default constants for weak subjectivity and trust scoring.
const (
	// BaseWSPeriod is the base weak subjectivity period in epochs when the
	// validator set is large (>=262144 validators, ~8.4M ETH staked).
	BaseWSPeriod uint64 = 256

	// MinWSPeriod is the minimum weak subjectivity period in epochs,
	// applicable when validator count is very low.
	MinWSPeriod uint64 = 128

	// MaxWSPeriod is the maximum weak subjectivity period in epochs.
	MaxWSPeriod uint64 = 1024

	// WSValidatorThreshold is the validator count threshold above which
	// the weak subjectivity period scales logarithmically.
	WSValidatorThreshold uint64 = 262144

	// DefaultMaxCheckpointAge is the default maximum checkpoint age in epochs.
	DefaultMaxCheckpointAge uint64 = 512

	// DefaultMinTrustScore is the default minimum trust score for acceptance.
	DefaultMinTrustScore float64 = 0.5

	// DefaultFinalityDepth is the default number of finalized epochs
	// required for maximum trust contribution.
	DefaultFinalityDepth uint64 = 64

	// DefaultPeerThreshold is the default minimum number of peer attestations
	// for full peer trust contribution.
	DefaultPeerThreshold int = 10
)

// CheckpointValidatorConfig configures the checkpoint validator.
type CheckpointValidatorConfig struct {
	// MaxAge is the maximum acceptable age of a checkpoint in epochs.
	MaxAge uint64

	// MinTrustScore is the minimum trust score required for acceptance.
	MinTrustScore float64

	// FinalityDepth is the number of finalized epochs for maximum finality trust.
	FinalityDepth uint64

	// PeerThreshold is the number of peer attestations for full peer trust.
	PeerThreshold int
}

// DefaultCheckpointValidatorConfig returns sensible defaults.
func DefaultCheckpointValidatorConfig() CheckpointValidatorConfig {
	return CheckpointValidatorConfig{
		MaxAge:        DefaultMaxCheckpointAge,
		MinTrustScore: DefaultMinTrustScore,
		FinalityDepth: DefaultFinalityDepth,
		PeerThreshold: DefaultPeerThreshold,
	}
}

// TrustScore encapsulates a numerical trust assessment for a checkpoint.
// Values range from 0.0 (no trust) to 1.0 (fully trusted).
type TrustScore struct {
	// Overall is the combined trust score in [0.0, 1.0].
	Overall float64

	// AgeScore is the trust component from checkpoint freshness.
	AgeScore float64

	// FinalityScore is the trust component from finality depth.
	FinalityScore float64

	// PeerScore is the trust component from peer attestations.
	PeerScore float64

	// Epoch is the checkpoint epoch that was scored.
	Epoch uint64

	// Accepted indicates whether the score meets the minimum threshold.
	Accepted bool
}

// WeakSubjectivityProof proves that a checkpoint is within the weak
// subjectivity period, making it safe to trust for chain bootstrapping.
type WeakSubjectivityProof struct {
	// CheckpointEpoch is the epoch of the checkpoint being validated.
	CheckpointEpoch uint64

	// CurrentEpoch is the current epoch at proof generation time.
	CurrentEpoch uint64

	// ValidatorCount is the active validator count used for WS calculation.
	ValidatorCount uint64

	// WSPeriod is the computed weak subjectivity period in epochs.
	WSPeriod uint64

	// EpochsRemaining is how many epochs remain in the WS window.
	EpochsRemaining uint64

	// Root is the checkpoint block root.
	Root [32]byte

	// StateRoot is the checkpoint state root.
	StateRoot [32]byte

	// Valid indicates whether the proof passes all checks.
	Valid bool
}

// CheckpointValidator validates checkpoint trust using finality proofs,
// weak subjectivity bounds, and peer attestation counts.
// All methods are safe for concurrent use.
type CheckpointValidator struct {
	mu     sync.RWMutex
	config CheckpointValidatorConfig

	// validated tracks previously validated checkpoint epochs and their scores.
	validated map[uint64]*TrustScore
}

// NewCheckpointValidator creates a new checkpoint validator with the given config.
func NewCheckpointValidator(cfg CheckpointValidatorConfig) *CheckpointValidator {
	if cfg.MaxAge == 0 {
		cfg.MaxAge = DefaultMaxCheckpointAge
	}
	if cfg.MinTrustScore <= 0 || cfg.MinTrustScore > 1.0 {
		cfg.MinTrustScore = DefaultMinTrustScore
	}
	if cfg.FinalityDepth == 0 {
		cfg.FinalityDepth = DefaultFinalityDepth
	}
	if cfg.PeerThreshold <= 0 {
		cfg.PeerThreshold = DefaultPeerThreshold
	}
	return &CheckpointValidator{
		config:    cfg,
		validated: make(map[uint64]*TrustScore),
	}
}

// ValidateCheckpoint validates a checkpoint at the given epoch by computing
// its trust score from age, finality depth, and peer attestation count.
// The currentEpoch, finalityDepth, and peerAttestations are provided by the
// caller from their local chain view.
func (cv *CheckpointValidator) ValidateCheckpoint(
	epoch uint64, root [32]byte, stateRoot [32]byte,
	currentEpoch uint64, finalityDepth uint64, peerAttestations int,
) (*TrustScore, error) {
	if epoch == 0 {
		return nil, ErrCheckpointEpochZero
	}
	if root == ([32]byte{}) {
		return nil, ErrCheckpointRootZero
	}
	if stateRoot == ([32]byte{}) {
		return nil, ErrCheckpointStateRootZero
	}
	if currentEpoch < epoch {
		return nil, ErrCurrentEpochBehind
	}

	age := currentEpoch - epoch
	if age > cv.config.MaxAge {
		return nil, ErrCheckpointTooOld
	}

	score := cv.ComputeTrustScore(age, finalityDepth, peerAttestations)

	ts := &TrustScore{
		Overall:       score,
		AgeScore:      cv.computeAgeScore(age),
		FinalityScore: cv.computeFinalityScore(finalityDepth),
		PeerScore:     cv.computePeerScore(peerAttestations),
		Epoch:         epoch,
		Accepted:      score >= cv.config.MinTrustScore,
	}

	cv.mu.Lock()
	cv.validated[epoch] = ts
	cv.mu.Unlock()

	if !ts.Accepted {
		return ts, ErrCheckpointTrustLow
	}
	return ts, nil
}

// BuildWeakSubjectivityProof builds a proof that the given checkpoint epoch
// is within the weak subjectivity period for the current epoch and validator set.
func (cv *CheckpointValidator) BuildWeakSubjectivityProof(
	epoch uint64, currentEpoch uint64, validatorCount uint64,
	root [32]byte, stateRoot [32]byte,
) (*WeakSubjectivityProof, error) {
	if epoch == 0 {
		return nil, ErrCheckpointEpochZero
	}
	if validatorCount == 0 {
		return nil, ErrValidatorCountZero
	}
	if currentEpoch < epoch {
		return nil, ErrCurrentEpochBehind
	}

	wsPeriod := WeakSubjectivityPeriod(validatorCount)
	age := currentEpoch - epoch
	valid := age <= wsPeriod

	var remaining uint64
	if valid {
		remaining = wsPeriod - age
	}

	return &WeakSubjectivityProof{
		CheckpointEpoch: epoch,
		CurrentEpoch:    currentEpoch,
		ValidatorCount:  validatorCount,
		WSPeriod:        wsPeriod,
		EpochsRemaining: remaining,
		Root:            root,
		StateRoot:       stateRoot,
		Valid:           valid,
	}, nil
}

// VerifyWeakSubjectivityProof verifies a weak subjectivity proof by
// recomputing the WS period and checking the age constraint.
func (cv *CheckpointValidator) VerifyWeakSubjectivityProof(proof *WeakSubjectivityProof) bool {
	if proof == nil {
		return false
	}
	if proof.ValidatorCount == 0 {
		return false
	}
	if proof.CurrentEpoch < proof.CheckpointEpoch {
		return false
	}
	if proof.Root == ([32]byte{}) || proof.StateRoot == ([32]byte{}) {
		return false
	}

	// Recompute and verify the WS period.
	expectedPeriod := WeakSubjectivityPeriod(proof.ValidatorCount)
	if proof.WSPeriod != expectedPeriod {
		return false
	}

	age := proof.CurrentEpoch - proof.CheckpointEpoch
	withinPeriod := age <= expectedPeriod

	// Verify the remaining epochs calculation.
	if withinPeriod {
		if proof.EpochsRemaining != expectedPeriod-age {
			return false
		}
	}

	return proof.Valid == withinPeriod
}

// IsWithinWeakSubjectivityPeriod checks whether a checkpoint at
// checkpointEpoch is within the weak subjectivity period relative
// to currentEpoch, given the active validator count.
func (cv *CheckpointValidator) IsWithinWeakSubjectivityPeriod(
	checkpointEpoch, currentEpoch, validatorCount uint64,
) bool {
	if validatorCount == 0 || currentEpoch < checkpointEpoch {
		return false
	}
	age := currentEpoch - checkpointEpoch
	return age <= WeakSubjectivityPeriod(validatorCount)
}

// ComputeTrustScore computes a combined trust score from age, finality
// depth, and peer attestation count. Each component contributes equally
// (1/3 weight) to the final score in [0.0, 1.0].
func (cv *CheckpointValidator) ComputeTrustScore(
	age uint64, finalityDepth uint64, peerAttestations int,
) float64 {
	ageScore := cv.computeAgeScore(age)
	finalityScore := cv.computeFinalityScore(finalityDepth)
	peerScore := cv.computePeerScore(peerAttestations)

	// Equal weighting: 1/3 each.
	return (ageScore + finalityScore + peerScore) / 3.0
}

// GetValidatedScore returns a previously computed trust score for an epoch.
func (cv *CheckpointValidator) GetValidatedScore(epoch uint64) *TrustScore {
	cv.mu.RLock()
	defer cv.mu.RUnlock()
	return cv.validated[epoch]
}

// ValidatedCount returns the number of checkpoints that have been validated.
func (cv *CheckpointValidator) ValidatedCount() int {
	cv.mu.RLock()
	defer cv.mu.RUnlock()
	return len(cv.validated)
}

// WeakSubjectivityPeriod computes the weak subjectivity period in epochs
// based on the active validator count. The formula follows the Ethereum
// consensus spec: for large validator sets, the period scales logarithmically.
func WeakSubjectivityPeriod(validatorCount uint64) uint64 {
	if validatorCount == 0 {
		return MinWSPeriod
	}

	if validatorCount < WSValidatorThreshold {
		// For small validator sets, scale linearly between min and base.
		ratio := float64(validatorCount) / float64(WSValidatorThreshold)
		period := float64(MinWSPeriod) + ratio*float64(BaseWSPeriod-MinWSPeriod)
		return uint64(period)
	}

	// For large validator sets, scale logarithmically above the base.
	logScale := math.Log2(float64(validatorCount) / float64(WSValidatorThreshold))
	period := float64(BaseWSPeriod) + logScale*float64(BaseWSPeriod)
	if period > float64(MaxWSPeriod) {
		return MaxWSPeriod
	}
	return uint64(period)
}

// --- Internal scoring helpers ---

// computeAgeScore returns a score in [0.0, 1.0] based on checkpoint age.
// Newer checkpoints score higher. Score decays linearly with age.
func (cv *CheckpointValidator) computeAgeScore(age uint64) float64 {
	if age == 0 {
		return 1.0
	}
	if age >= cv.config.MaxAge {
		return 0.0
	}
	return 1.0 - float64(age)/float64(cv.config.MaxAge)
}

// computeFinalityScore returns a score in [0.0, 1.0] based on finality depth.
// Deeper finality means more confidence.
func (cv *CheckpointValidator) computeFinalityScore(depth uint64) float64 {
	if depth == 0 {
		return 0.0
	}
	if depth >= cv.config.FinalityDepth {
		return 1.0
	}
	return float64(depth) / float64(cv.config.FinalityDepth)
}

// computePeerScore returns a score in [0.0, 1.0] based on peer attestation count.
func (cv *CheckpointValidator) computePeerScore(attestations int) float64 {
	if attestations <= 0 {
		return 0.0
	}
	if attestations >= cv.config.PeerThreshold {
		return 1.0
	}
	return float64(attestations) / float64(cv.config.PeerThreshold)
}
