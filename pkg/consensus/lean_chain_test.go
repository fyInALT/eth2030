package consensus

import (
	"testing"

	"github.com/eth2030/eth2030/consensus/pq"
	"github.com/eth2030/eth2030/core/types"
)

// toLeanCfg converts a root ConsensusConfig to the minimal pq.LeanConfig.
func toLeanCfg(cfg *ConsensusConfig) *pq.LeanConfig {
	if cfg == nil {
		return nil
	}
	return &pq.LeanConfig{
		LeanAvailableChainMode:       cfg.LeanAvailableChainMode,
		LeanAvailableChainValidators: cfg.LeanAvailableChainValidators,
	}
}

func TestLeanAvailableChainModeConfig(t *testing.T) {
	// Default config should have lean mode disabled and validators = 512.
	cfg := DefaultConfig()
	if cfg.LeanAvailableChainMode {
		t.Fatal("default config should have LeanAvailableChainMode = false")
	}
	if cfg.LeanAvailableChainValidators != 512 {
		t.Fatalf("default LeanAvailableChainValidators = %d, want 512", cfg.LeanAvailableChainValidators)
	}

	// Validation: lean mode off, any value is fine.
	cfg.LeanAvailableChainMode = false
	cfg.LeanAvailableChainValidators = 0
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error with lean mode off, got: %v", err)
	}

	// Validation: lean mode on, validators in range [256, 1024].
	cfg.LeanAvailableChainMode = true
	cfg.LeanAvailableChainValidators = 512
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error with validators=512, got: %v", err)
	}

	cfg.LeanAvailableChainValidators = 256
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error with validators=256, got: %v", err)
	}

	cfg.LeanAvailableChainValidators = 1024
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error with validators=1024, got: %v", err)
	}

	// Out of range: too low.
	cfg.LeanAvailableChainValidators = 100
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error with validators=100")
	}

	// Out of range: too high.
	cfg.LeanAvailableChainValidators = 2000
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error with validators=2000")
	}

	// Lean mode on, validators = 0 should pass (means use default).
	cfg.LeanAvailableChainValidators = 0
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error with validators=0 (use default), got: %v", err)
	}
}

func TestSelectLeanPQAttestors(t *testing.T) {
	validators := make([]uint64, 1000)
	for i := range validators {
		validators[i] = uint64(i)
	}

	var epochSeed types.Hash
	epochSeed[0] = 0xAB
	epochSeed[1] = 0xCD

	result := SelectLeanPQAttestors(validators, 512, 42, epochSeed)
	if len(result) != 512 {
		t.Fatalf("got %d attestors, want 512", len(result))
	}

	// All returned indices must be unique.
	seen := make(map[uint64]bool)
	for _, v := range result {
		if seen[v] {
			t.Fatalf("duplicate validator index %d", v)
		}
		seen[v] = true
	}

	// All returned indices must be valid.
	for _, v := range result {
		if v >= 1000 {
			t.Fatalf("invalid validator index %d", v)
		}
	}

	// Deterministic: same args -> same result.
	result2 := SelectLeanPQAttestors(validators, 512, 42, epochSeed)
	for i := range result {
		if result[i] != result2[i] {
			t.Fatalf("non-deterministic at index %d: %d vs %d", i, result[i], result2[i])
		}
	}

	// Count >= len -> return all.
	resultAll := SelectLeanPQAttestors(validators, 2000, 42, epochSeed)
	if len(resultAll) != 1000 {
		t.Fatalf("got %d attestors, want 1000 (all)", len(resultAll))
	}
}

func TestLeanAvailablePQAttestors(t *testing.T) {
	verifier := NewPQAttestationVerifier(nil)
	validators := make([]uint64, 1000)
	for i := range validators {
		validators[i] = uint64(i)
	}
	var epochSeed types.Hash
	epochSeed[0] = 0x42

	// Mode off -> all validators returned.
	cfgOff := DefaultConfig()
	cfgOff.LeanAvailableChainMode = false
	result := verifier.LeanAvailablePQAttestors(42, validators, epochSeed, toLeanCfg(cfgOff))
	if len(result) != 1000 {
		t.Fatalf("mode off: got %d, want 1000", len(result))
	}

	// Nil config -> all validators returned.
	result = verifier.LeanAvailablePQAttestors(42, validators, epochSeed, nil)
	if len(result) != 1000 {
		t.Fatalf("nil config: got %d, want 1000", len(result))
	}

	// Mode on -> 512 returned.
	cfgOn := DefaultConfig()
	cfgOn.LeanAvailableChainMode = true
	cfgOn.LeanAvailableChainValidators = 512
	result = verifier.LeanAvailablePQAttestors(42, validators, epochSeed, toLeanCfg(cfgOn))
	if len(result) != 512 {
		t.Fatalf("mode on: got %d, want 512", len(result))
	}
}

func TestLeanAvailableAggregation(t *testing.T) {
	// Create attestations with dummy data.
	attestations := make([]PQAttestation, 10)
	for i := range attestations {
		attestations[i] = PQAttestation{
			Slot:            100,
			CommitteeIndex:  0,
			BeaconBlockRoot: types.Hash{0x01, 0x02},
			SourceEpoch:     3,
			TargetEpoch:     4,
			PQSignature:     []byte{0x01, 0x02, 0x03},
			PQPublicKey:     []byte{byte(i), 0xAA, 0xBB},
			ValidatorIndex:  uint64(i),
		}
	}

	agg := NewSTARKSignatureAggregator()

	// Lean mode on: should not use STARK prover, MerkleRoot set in CommitteeRoot.
	cfg := DefaultConfig()
	cfg.LeanAvailableChainMode = true
	result, err := agg.AggregateWithConfig(attestations, toLeanCfg(cfg))
	if err != nil {
		t.Fatalf("lean aggregation failed: %v", err)
	}
	if result.AggregateProof != nil {
		t.Fatal("lean mode should not produce STARK proof")
	}
	emptyHash := types.Hash{}
	if result.CommitteeRoot == emptyHash {
		t.Fatal("lean mode should set CommitteeRoot")
	}
	if result.NumValidators != 10 {
		t.Fatalf("NumValidators = %d, want 10", result.NumValidators)
	}
	if result.Slot != 100 {
		t.Fatalf("Slot = %d, want 100", result.Slot)
	}

	// Lean mode off: should use regular STARK aggregation.
	cfgOff := DefaultConfig()
	cfgOff.LeanAvailableChainMode = false
	resultFull, err := agg.AggregateWithConfig(attestations, toLeanCfg(cfgOff))
	if err != nil {
		t.Fatalf("full aggregation failed: %v", err)
	}
	if resultFull.AggregateProof == nil {
		t.Fatal("full mode should produce STARK proof")
	}
}

func TestLeanAggregationDeterministic(t *testing.T) {
	validators := make([]uint64, 1000)
	for i := range validators {
		validators[i] = uint64(i)
	}
	var epochSeed types.Hash
	epochSeed[0] = 0xFF
	epochSeed[15] = 0x01

	r1 := SelectLeanPQAttestors(validators, 512, 99, epochSeed)
	r2 := SelectLeanPQAttestors(validators, 512, 99, epochSeed)
	if len(r1) != len(r2) {
		t.Fatalf("lengths differ: %d vs %d", len(r1), len(r2))
	}
	for i := range r1 {
		if r1[i] != r2[i] {
			t.Fatalf("mismatch at %d: %d vs %d", i, r1[i], r2[i])
		}
	}

	// Different slot -> different result.
	r3 := SelectLeanPQAttestors(validators, 512, 100, epochSeed)
	differ := false
	for i := range r1 {
		if r1[i] != r3[i] {
			differ = true
			break
		}
	}
	if !differ {
		t.Fatal("different slot should produce different selection")
	}
}
