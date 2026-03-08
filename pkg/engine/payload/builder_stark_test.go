package payload

import (
	"sync"
	"testing"

	"github.com/eth2030/eth2030/proofs"
)

// TestPayloadBuilder_ProverNilByDefault verifies that a new PayloadBuilder has no prover set.
func TestPayloadBuilder_ProverNilByDefault(t *testing.T) {
	pb := NewPayloadBuilder(nil, nil, nil)
	if pb.prover != nil {
		t.Fatal("prover should be nil by default")
	}
}

// TestPayloadBuilder_SetValidationFrameProver verifies that SetValidationFrameProver stores the prover.
func TestPayloadBuilder_SetValidationFrameProver(t *testing.T) {
	pb := NewPayloadBuilder(nil, nil, nil)
	stub := &proofs.StubValidationFrameProver{}
	pb.SetValidationFrameProver(stub)

	pb.mu.RLock()
	defer pb.mu.RUnlock()
	if pb.prover == nil {
		t.Fatal("prover should not be nil after SetValidationFrameProver")
	}
}

// TestPayloadBuilder_SetValidationFrameProver_Nil verifies that setting the prover to nil clears it.
func TestPayloadBuilder_SetValidationFrameProver_Nil(t *testing.T) {
	pb := NewPayloadBuilder(nil, nil, nil)
	stub := &proofs.StubValidationFrameProver{}
	pb.SetValidationFrameProver(stub)
	pb.SetValidationFrameProver(nil)

	pb.mu.RLock()
	defer pb.mu.RUnlock()
	if pb.prover != nil {
		t.Fatal("prover should be nil after setting to nil")
	}
}

// TestPayloadBuilder_SetValidationFrameProver_Concurrent verifies no data race with concurrent sets.
func TestPayloadBuilder_SetValidationFrameProver_Concurrent(t *testing.T) {
	pb := NewPayloadBuilder(nil, nil, nil)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stub := &proofs.StubValidationFrameProver{}
			pb.SetValidationFrameProver(stub)
		}()
	}
	wg.Wait()

	pb.mu.RLock()
	defer pb.mu.RUnlock()
	if pb.prover == nil {
		t.Fatal("prover should be set after concurrent calls")
	}
}

// TestPayloadBuilder_SetValidationFrameProver_Replace verifies that replacing a prover works.
func TestPayloadBuilder_SetValidationFrameProver_Replace(t *testing.T) {
	pb := NewPayloadBuilder(nil, nil, nil)
	stub1 := &proofs.StubValidationFrameProver{}
	real := proofs.NewSTARKValidationFrameProver()

	// Set stub first, then replace with STARK prover.
	pb.SetValidationFrameProver(stub1)
	pb.mu.RLock()
	first := pb.prover
	pb.mu.RUnlock()

	if first == nil {
		t.Fatal("prover should be set after first call")
	}

	pb.SetValidationFrameProver(real)
	pb.mu.RLock()
	second := pb.prover
	pb.mu.RUnlock()

	if second == nil {
		t.Fatal("prover should be set after second call")
	}

	// Verify the final prover is the STARK prover, not the stub.
	if second == proofs.ValidationFrameProver(stub1) {
		t.Fatal("prover should have been replaced, still points to stub1")
	}
	if second != proofs.ValidationFrameProver(real) {
		t.Fatal("final prover should be the STARK prover")
	}
}
