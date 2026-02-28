package light

import (
	"testing"
)

func TestCheckpointValidatorNew(t *testing.T) {
	cv := NewCheckpointValidator(DefaultCheckpointValidatorConfig())
	if cv == nil {
		t.Fatal("expected non-nil checkpoint validator")
	}
	if cv.config.MaxAge != DefaultMaxCheckpointAge {
		t.Fatalf("expected max age %d, got %d", DefaultMaxCheckpointAge, cv.config.MaxAge)
	}
	if cv.config.PeerThreshold != DefaultPeerThreshold {
		t.Fatalf("expected peer threshold %d, got %d", DefaultPeerThreshold, cv.config.PeerThreshold)
	}
}

func TestCheckpointValidatorDefaults(t *testing.T) {
	// Zero values should be replaced with defaults.
	cv := NewCheckpointValidator(CheckpointValidatorConfig{})
	if cv.config.MaxAge != DefaultMaxCheckpointAge {
		t.Fatalf("expected default max age, got %d", cv.config.MaxAge)
	}
	if cv.config.MinTrustScore != DefaultMinTrustScore {
		t.Fatalf("expected default min trust, got %f", cv.config.MinTrustScore)
	}
	if cv.config.FinalityDepth != DefaultFinalityDepth {
		t.Fatalf("expected default finality depth, got %d", cv.config.FinalityDepth)
	}
}

func TestCheckpointValidateSuccess(t *testing.T) {
	cv := NewCheckpointValidator(DefaultCheckpointValidatorConfig())
	root := [32]byte{0x01}
	stateRoot := [32]byte{0x02}

	// Recent checkpoint with good finality and peers.
	ts, err := cv.ValidateCheckpoint(100, root, stateRoot, 110, 64, 15)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if !ts.Accepted {
		t.Fatal("expected checkpoint to be accepted")
	}
	if ts.Overall <= 0 || ts.Overall > 1.0 {
		t.Fatalf("trust score out of range: %f", ts.Overall)
	}
	if ts.Epoch != 100 {
		t.Fatalf("expected epoch 100, got %d", ts.Epoch)
	}
}

func TestCheckpointValidateEpochZero(t *testing.T) {
	cv := NewCheckpointValidator(DefaultCheckpointValidatorConfig())
	_, err := cv.ValidateCheckpoint(0, [32]byte{1}, [32]byte{2}, 100, 10, 5)
	if err != ErrCheckpointEpochZero {
		t.Fatalf("expected ErrCheckpointEpochZero, got %v", err)
	}
}

func TestCheckpointValidateRootZero(t *testing.T) {
	cv := NewCheckpointValidator(DefaultCheckpointValidatorConfig())
	_, err := cv.ValidateCheckpoint(10, [32]byte{}, [32]byte{2}, 100, 10, 5)
	if err != ErrCheckpointRootZero {
		t.Fatalf("expected ErrCheckpointRootZero, got %v", err)
	}
}

func TestCheckpointValidateStateRootZero(t *testing.T) {
	cv := NewCheckpointValidator(DefaultCheckpointValidatorConfig())
	_, err := cv.ValidateCheckpoint(10, [32]byte{1}, [32]byte{}, 100, 10, 5)
	if err != ErrCheckpointStateRootZero {
		t.Fatalf("expected ErrCheckpointStateRootZero, got %v", err)
	}
}

func TestCheckpointValidateCurrentBehind(t *testing.T) {
	cv := NewCheckpointValidator(DefaultCheckpointValidatorConfig())
	_, err := cv.ValidateCheckpoint(200, [32]byte{1}, [32]byte{2}, 100, 10, 5)
	if err != ErrCurrentEpochBehind {
		t.Fatalf("expected ErrCurrentEpochBehind, got %v", err)
	}
}

func TestCheckpointValidateTooOld(t *testing.T) {
	cv := NewCheckpointValidator(CheckpointValidatorConfig{
		MaxAge:        100,
		MinTrustScore: 0.5,
		FinalityDepth: 64,
		PeerThreshold: 10,
	})
	_, err := cv.ValidateCheckpoint(10, [32]byte{1}, [32]byte{2}, 200, 10, 5)
	if err != ErrCheckpointTooOld {
		t.Fatalf("expected ErrCheckpointTooOld, got %v", err)
	}
}

func TestCheckpointValidateLowTrust(t *testing.T) {
	cv := NewCheckpointValidator(CheckpointValidatorConfig{
		MaxAge:        512,
		MinTrustScore: 0.9, // Very high threshold.
		FinalityDepth: 64,
		PeerThreshold: 10,
	})
	// Old checkpoint with low finality and few peers.
	ts, err := cv.ValidateCheckpoint(10, [32]byte{1}, [32]byte{2}, 500, 1, 1)
	if err != ErrCheckpointTrustLow {
		t.Fatalf("expected ErrCheckpointTrustLow, got %v", err)
	}
	if ts.Accepted {
		t.Fatal("expected checkpoint to be rejected")
	}
}

func TestWeakSubjectivityPeriodSmallValidators(t *testing.T) {
	// With 0 validators, should return MinWSPeriod.
	period := WeakSubjectivityPeriod(0)
	if period != MinWSPeriod {
		t.Fatalf("expected MinWSPeriod (%d), got %d", MinWSPeriod, period)
	}

	// With half the threshold validators.
	halfThreshold := WSValidatorThreshold / 2
	period = WeakSubjectivityPeriod(halfThreshold)
	if period <= MinWSPeriod || period >= BaseWSPeriod {
		t.Fatalf("expected period between %d and %d, got %d", MinWSPeriod, BaseWSPeriod, period)
	}
}

func TestWeakSubjectivityPeriodLargeValidators(t *testing.T) {
	// At threshold, should be approximately BaseWSPeriod.
	period := WeakSubjectivityPeriod(WSValidatorThreshold)
	if period != BaseWSPeriod {
		t.Fatalf("expected BaseWSPeriod (%d), got %d", BaseWSPeriod, period)
	}

	// At 2x threshold, should be above BaseWSPeriod.
	period = WeakSubjectivityPeriod(WSValidatorThreshold * 2)
	if period <= BaseWSPeriod {
		t.Fatalf("expected period > %d for double threshold, got %d", BaseWSPeriod, period)
	}

	// Very large count should not exceed MaxWSPeriod.
	period = WeakSubjectivityPeriod(WSValidatorThreshold * 1024)
	if period > MaxWSPeriod {
		t.Fatalf("expected period <= %d, got %d", MaxWSPeriod, period)
	}
}

func TestBuildWeakSubjectivityProofSuccess(t *testing.T) {
	cv := NewCheckpointValidator(DefaultCheckpointValidatorConfig())
	root := [32]byte{0xaa}
	stateRoot := [32]byte{0xbb}

	proof, err := cv.BuildWeakSubjectivityProof(100, 200, 300000, root, stateRoot)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if proof == nil {
		t.Fatal("expected non-nil proof")
	}
	if !proof.Valid {
		t.Fatal("expected proof to be valid")
	}
	if proof.CheckpointEpoch != 100 {
		t.Fatalf("expected checkpoint epoch 100, got %d", proof.CheckpointEpoch)
	}
	if proof.EpochsRemaining == 0 {
		t.Fatal("expected non-zero epochs remaining")
	}
}

func TestBuildWeakSubjectivityProofExpired(t *testing.T) {
	cv := NewCheckpointValidator(DefaultCheckpointValidatorConfig())
	root := [32]byte{0xaa}
	stateRoot := [32]byte{0xbb}

	// Build a proof where the checkpoint is very old compared to current epoch.
	proof, err := cv.BuildWeakSubjectivityProof(100, 100000, 300000, root, stateRoot)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if proof.Valid {
		t.Fatal("expected proof to be invalid (expired)")
	}
	if proof.EpochsRemaining != 0 {
		t.Fatalf("expected 0 epochs remaining, got %d", proof.EpochsRemaining)
	}
}

func TestBuildWSProofEpochZero(t *testing.T) {
	cv := NewCheckpointValidator(DefaultCheckpointValidatorConfig())
	_, err := cv.BuildWeakSubjectivityProof(0, 100, 1000, [32]byte{1}, [32]byte{2})
	if err != ErrCheckpointEpochZero {
		t.Fatalf("expected ErrCheckpointEpochZero, got %v", err)
	}
}

func TestBuildWSProofValidatorCountZero(t *testing.T) {
	cv := NewCheckpointValidator(DefaultCheckpointValidatorConfig())
	_, err := cv.BuildWeakSubjectivityProof(100, 200, 0, [32]byte{1}, [32]byte{2})
	if err != ErrValidatorCountZero {
		t.Fatalf("expected ErrValidatorCountZero, got %v", err)
	}
}

func TestBuildWSProofCurrentBehind(t *testing.T) {
	cv := NewCheckpointValidator(DefaultCheckpointValidatorConfig())
	_, err := cv.BuildWeakSubjectivityProof(200, 100, 1000, [32]byte{1}, [32]byte{2})
	if err != ErrCurrentEpochBehind {
		t.Fatalf("expected ErrCurrentEpochBehind, got %v", err)
	}
}

func TestVerifyWeakSubjectivityProofValid(t *testing.T) {
	cv := NewCheckpointValidator(DefaultCheckpointValidatorConfig())
	root := [32]byte{0xcc}
	stateRoot := [32]byte{0xdd}

	proof, err := cv.BuildWeakSubjectivityProof(100, 200, 300000, root, stateRoot)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if !cv.VerifyWeakSubjectivityProof(proof) {
		t.Fatal("expected proof to verify successfully")
	}
}

func TestVerifyWeakSubjectivityProofNil(t *testing.T) {
	cv := NewCheckpointValidator(DefaultCheckpointValidatorConfig())
	if cv.VerifyWeakSubjectivityProof(nil) {
		t.Fatal("expected nil proof to fail verification")
	}
}

func TestVerifyWeakSubjectivityProofTampered(t *testing.T) {
	cv := NewCheckpointValidator(DefaultCheckpointValidatorConfig())
	root := [32]byte{0xcc}
	stateRoot := [32]byte{0xdd}

	proof, _ := cv.BuildWeakSubjectivityProof(100, 200, 300000, root, stateRoot)
	// Tamper with the WS period.
	proof.WSPeriod = 9999
	if cv.VerifyWeakSubjectivityProof(proof) {
		t.Fatal("expected tampered proof to fail verification")
	}
}

func TestVerifyWSProofZeroRoot(t *testing.T) {
	cv := NewCheckpointValidator(DefaultCheckpointValidatorConfig())
	proof := &WeakSubjectivityProof{
		CheckpointEpoch: 100,
		CurrentEpoch:    200,
		ValidatorCount:  300000,
		WSPeriod:        WeakSubjectivityPeriod(300000),
		Root:            [32]byte{}, // Zero root.
		StateRoot:       [32]byte{0xdd},
		Valid:           true,
	}
	if cv.VerifyWeakSubjectivityProof(proof) {
		t.Fatal("expected zero root proof to fail")
	}
}

func TestIsWithinWeakSubjectivityPeriod(t *testing.T) {
	cv := NewCheckpointValidator(DefaultCheckpointValidatorConfig())

	// Recent checkpoint should be within WS period.
	if !cv.IsWithinWeakSubjectivityPeriod(100, 200, 300000) {
		t.Fatal("expected recent checkpoint to be within WS period")
	}

	// Very old checkpoint should be outside WS period.
	if cv.IsWithinWeakSubjectivityPeriod(100, 100000, 300000) {
		t.Fatal("expected old checkpoint to be outside WS period")
	}

	// Zero validator count should return false.
	if cv.IsWithinWeakSubjectivityPeriod(100, 200, 0) {
		t.Fatal("expected false for zero validator count")
	}

	// Current behind checkpoint should return false.
	if cv.IsWithinWeakSubjectivityPeriod(200, 100, 300000) {
		t.Fatal("expected false when current epoch is behind")
	}
}

func TestTrustScoreComponents(t *testing.T) {
	cv := NewCheckpointValidator(CheckpointValidatorConfig{
		MaxAge:        100,
		MinTrustScore: 0.3,
		FinalityDepth: 50,
		PeerThreshold: 20,
	})

	// Perfect scores across all dimensions.
	score := cv.ComputeTrustScore(0, 50, 20)
	if score < 0.99 {
		t.Fatalf("expected score ~1.0 for perfect inputs, got %f", score)
	}

	// Zero scores across all dimensions.
	score = cv.ComputeTrustScore(100, 0, 0)
	if score > 0.01 {
		t.Fatalf("expected score ~0.0 for worst inputs, got %f", score)
	}

	// Mixed: half age, half finality, half peers.
	score = cv.ComputeTrustScore(50, 25, 10)
	if score < 0.45 || score > 0.55 {
		t.Fatalf("expected score ~0.5 for mid inputs, got %f", score)
	}
}

func TestTrustScoreAgeDecay(t *testing.T) {
	cv := NewCheckpointValidator(CheckpointValidatorConfig{
		MaxAge:        100,
		MinTrustScore: 0.1,
		FinalityDepth: 64,
		PeerThreshold: 10,
	})

	// Younger should score higher than older.
	score1 := cv.ComputeTrustScore(10, 32, 5)
	score2 := cv.ComputeTrustScore(80, 32, 5)
	if score1 <= score2 {
		t.Fatalf("expected younger checkpoint to score higher: %f <= %f", score1, score2)
	}
}

func TestGetValidatedScore(t *testing.T) {
	cv := NewCheckpointValidator(DefaultCheckpointValidatorConfig())

	// No validated checkpoints yet.
	if cv.GetValidatedScore(100) != nil {
		t.Fatal("expected nil for unvalidated epoch")
	}

	// Validate a checkpoint.
	cv.ValidateCheckpoint(100, [32]byte{1}, [32]byte{2}, 110, 64, 15)

	ts := cv.GetValidatedScore(100)
	if ts == nil {
		t.Fatal("expected non-nil for validated epoch")
	}
	if ts.Epoch != 100 {
		t.Fatalf("expected epoch 100, got %d", ts.Epoch)
	}
}

func TestValidatedCount(t *testing.T) {
	cv := NewCheckpointValidator(DefaultCheckpointValidatorConfig())
	if cv.ValidatedCount() != 0 {
		t.Fatalf("expected 0, got %d", cv.ValidatedCount())
	}

	cv.ValidateCheckpoint(100, [32]byte{1}, [32]byte{2}, 110, 64, 15)
	cv.ValidateCheckpoint(200, [32]byte{3}, [32]byte{4}, 210, 64, 15)

	if cv.ValidatedCount() != 2 {
		t.Fatalf("expected 2, got %d", cv.ValidatedCount())
	}
}

func TestWeakSubjectivityPeriodMonotonic(t *testing.T) {
	// WS period should not decrease as validator count increases.
	prev := WeakSubjectivityPeriod(1)
	for _, count := range []uint64{100, 1000, 10000, 100000, 262144, 500000, 1000000} {
		period := WeakSubjectivityPeriod(count)
		if period < prev {
			t.Fatalf("WS period decreased from %d to %d at count %d", prev, period, count)
		}
		prev = period
	}
}

func TestCheckpointValidateSameEpochAsCurrent(t *testing.T) {
	cv := NewCheckpointValidator(DefaultCheckpointValidatorConfig())
	root := [32]byte{0x01}
	stateRoot := [32]byte{0x02}

	// Age = 0 should be a perfect age score.
	ts, err := cv.ValidateCheckpoint(100, root, stateRoot, 100, 64, 15)
	if err != nil {
		t.Fatalf("expected success for zero-age checkpoint, got %v", err)
	}
	if ts.AgeScore != 1.0 {
		t.Fatalf("expected age score 1.0 for zero age, got %f", ts.AgeScore)
	}
}
