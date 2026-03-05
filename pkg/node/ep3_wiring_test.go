package node

import (
	"testing"
)

// TestEP3ConfigDefaults verifies EP-3 config defaults are correct.
func TestEP3ConfigDefaults(t *testing.T) {
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

// TestEP3NodeTopicManagerWired verifies that topicMgr is initialized.
func TestEP3NodeTopicManagerWired(t *testing.T) {
	cfg := makeTestConfig(t)
	n := newTestNode(t, &cfg)

	if n.topicMgr == nil {
		t.Fatal("topicMgr should be non-nil after New()")
	}
}

// TestEP3NodeSTARKAggWired verifies that starkAgg is initialized.
func TestEP3NodeSTARKAggWired(t *testing.T) {
	cfg := makeTestConfig(t)
	n := newTestNode(t, &cfg)

	if n.starkAgg == nil {
		t.Fatal("starkAgg should be non-nil after New()")
	}
}

// TestEP3NodeProverNilByDefault verifies prover is nil when StarkValidationFrames=false.
func TestEP3NodeProverNilByDefault(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.StarkValidationFrames = false
	n := newTestNode(t, &cfg)

	if n.starkFrameProver != nil {
		t.Fatal("starkFrameProver should be nil when StarkValidationFrames=false")
	}
}

// TestEP3NodeProverCreatedWhenEnabled verifies prover is created when StarkValidationFrames=true.
func TestEP3NodeProverCreatedWhenEnabled(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.StarkValidationFrames = true
	n := newTestNode(t, &cfg)

	if n.starkFrameProver == nil {
		t.Fatal("starkFrameProver should be non-nil when StarkValidationFrames=true")
	}
}

// TestEP3NodeAggregatorStartStop verifies aggregator lifecycle through node Start/Stop.
func TestEP3NodeAggregatorStartStop(t *testing.T) {
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

// TestEP3NodeLifecycleWithStarkFrames is an e2e test with StarkValidationFrames enabled.
func TestEP3NodeLifecycleWithStarkFrames(t *testing.T) {
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

// TestEP3ConfigLeanValidators_ValidRange verifies valid lean validator count passes validation.
func TestEP3ConfigLeanValidators_ValidRange(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.LeanAvailableChainMode = true
	cfg.LeanAvailableChainValidators = 512

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() should succeed for valid lean validators: %v", err)
	}
}

// TestEP3ConfigLeanValidators_TooSmall verifies lean validator count below 256 fails validation.
func TestEP3ConfigLeanValidators_TooSmall(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.LeanAvailableChainMode = true
	cfg.LeanAvailableChainValidators = 100

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() should fail for LeanAvailableChainValidators=100 (below 256)")
	}
}

// TestEP3ConfigLeanValidators_TooBig verifies lean validator count above 1024 fails validation.
func TestEP3ConfigLeanValidators_TooBig(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.LeanAvailableChainMode = true
	cfg.LeanAvailableChainValidators = 2000

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() should fail for LeanAvailableChainValidators=2000 (above 1024)")
	}
}
