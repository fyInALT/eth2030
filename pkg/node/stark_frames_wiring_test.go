package node

import (
	"testing"
)

// TestStarkFrames_ConfigDefaults verifies STARK frames and lean chain config defaults.
func TestStarkFrames_ConfigDefaults(t *testing.T) {
	cfg := makeTestConfig(t)

	if cfg.LeanAvailableChainValidators != 512 {
		t.Errorf("LeanAvailableChainValidators = %d, want 512", cfg.LeanAvailableChainValidators)
	}
	if cfg.LeanAvailableChainMode {
		t.Error("LeanAvailableChainMode should default to false")
	}
	if cfg.StarkValidationFrames {
		t.Error("StarkValidationFrames should default to false")
	}
}

// TestStarkFrames_TopicManagerWired verifies that topicMgr is initialized after New().
func TestStarkFrames_TopicManagerWired(t *testing.T) {
	cfg := makeTestConfig(t)
	n := newTestNode(t, &cfg)

	if n.topicMgr == nil {
		t.Fatal("topicMgr should be non-nil after New()")
	}
}

// TestStarkFrames_STARKAggWired verifies that starkAgg is initialized after New().
func TestStarkFrames_STARKAggWired(t *testing.T) {
	cfg := makeTestConfig(t)
	n := newTestNode(t, &cfg)

	if n.starkAgg == nil {
		t.Fatal("starkAgg should be non-nil after New()")
	}
}

// TestStarkFrames_ProverNilByDefault verifies prover is nil when StarkValidationFrames=false.
func TestStarkFrames_ProverNilByDefault(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.StarkValidationFrames = false
	n := newTestNode(t, &cfg)

	if n.starkFrameProver != nil {
		t.Fatal("starkFrameProver should be nil when StarkValidationFrames=false")
	}
}

// TestStarkFrames_ProverCreatedWhenEnabled verifies prover is created when StarkValidationFrames=true.
func TestStarkFrames_ProverCreatedWhenEnabled(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.StarkValidationFrames = true
	n := newTestNode(t, &cfg)

	if n.starkFrameProver == nil {
		t.Fatal("starkFrameProver should be non-nil when StarkValidationFrames=true")
	}
}

// TestStarkFrames_AggregatorStartStop verifies aggregator lifecycle through node Start/Stop.
func TestStarkFrames_AggregatorStartStop(t *testing.T) {
	cfg := makeTestConfig(t)
	n, err := New(&cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	if n.starkAgg.IsRunning() {
		t.Fatal("aggregator should not be running before Start()")
	}

	if err := n.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	if !n.starkAgg.IsRunning() {
		t.Fatal("aggregator should be running after Start()")
	}

	if err := n.Stop(); err != nil {
		t.Fatalf("Stop() error: %v", err)
	}

	if n.starkAgg.IsRunning() {
		t.Fatal("aggregator should not be running after Stop()")
	}
}

// TestStarkFrames_NodeLifecycle is an e2e test verifying the full lifecycle with StarkValidationFrames=true.
func TestStarkFrames_NodeLifecycle(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.StarkValidationFrames = true

	n, err := New(&cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	if err := n.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	if !n.starkAgg.IsRunning() {
		t.Fatal("aggregator should be running after Start()")
	}
	if n.starkFrameProver == nil {
		t.Fatal("starkFrameProver should be non-nil with StarkValidationFrames=true")
	}

	if err := n.Stop(); err != nil {
		t.Fatalf("Stop() error: %v", err)
	}

	if n.starkAgg.IsRunning() {
		t.Fatal("aggregator should not be running after Stop()")
	}
}

// TestLeanChain_ValidatorRange_Valid verifies a valid lean validator count passes config validation.
func TestLeanChain_ValidatorRange_Valid(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.LeanAvailableChainMode = true
	cfg.LeanAvailableChainValidators = 512

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() should succeed for valid lean validators: %v", err)
	}
}

// TestLeanChain_ValidatorRange_TooSmall verifies lean validator count below 256 fails validation.
func TestLeanChain_ValidatorRange_TooSmall(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.LeanAvailableChainMode = true
	cfg.LeanAvailableChainValidators = 100

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() should fail for LeanAvailableChainValidators=100 (below 256)")
	}
}

// TestLeanChain_ValidatorRange_TooLarge verifies lean validator count above 1024 fails validation.
func TestLeanChain_ValidatorRange_TooLarge(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.LeanAvailableChainMode = true
	cfg.LeanAvailableChainValidators = 2000

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() should fail for LeanAvailableChainValidators=2000 (above 1024)")
	}
}
