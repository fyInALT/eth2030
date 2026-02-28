package rollup

import (
	"testing"

	"github.com/eth2030/eth2030/crypto"
)

func TestExecutionStepHash(t *testing.T) {
	step := BuildValidStep(0, 0x60, 21000, [32]byte{0x01})
	h := step.Hash()
	if h == ([32]byte{}) {
		t.Fatal("step hash should not be zero")
	}
	// Deterministic.
	h2 := step.Hash()
	if h != h2 {
		t.Fatal("step hash should be deterministic")
	}
}

func TestExecutionStepHashNil(t *testing.T) {
	var step *ExecutionStep
	h := step.Hash()
	if h != ([32]byte{}) {
		t.Fatal("nil step hash should be zero")
	}
}

func TestExecutionStepIsValid(t *testing.T) {
	step := BuildValidStep(0, 0x60, 21000, [32]byte{0x01})
	if !step.IsValid() {
		t.Fatal("valid step should report IsValid=true")
	}

	// Step with zero input state.
	bad := ExecutionStep{InputStateHash: [32]byte{}, OutputStateHash: [32]byte{0x01}}
	if bad.IsValid() {
		t.Fatal("step with zero input should report IsValid=false")
	}

	// Nil step.
	var nilStep *ExecutionStep
	if nilStep.IsValid() {
		t.Fatal("nil step should report IsValid=false")
	}
}

func TestBuildValidStep(t *testing.T) {
	input := [32]byte{0xaa}
	step := BuildValidStep(5, 0x01, 3000, input)
	if step.Index != 5 {
		t.Fatalf("expected index 5, got %d", step.Index)
	}
	if step.Opcode != 0x01 {
		t.Fatalf("expected opcode 0x01, got %d", step.Opcode)
	}
	if step.InputStateHash != input {
		t.Fatal("input state hash mismatch")
	}
	// Output should match the deterministic computation.
	expected := computeExpectedOutput(&step)
	if step.OutputStateHash != expected {
		t.Fatal("output state hash should match expected computation")
	}
}

func TestNewExecutionTrace(t *testing.T) {
	trace := BuildValidTrace(100, 5, [32]byte{0x01})
	if trace == nil {
		t.Fatal("expected non-nil trace")
	}
	if trace.Len() != 5 {
		t.Fatalf("expected 5 steps, got %d", trace.Len())
	}
	if trace.BlockNumber != 100 {
		t.Fatalf("expected block 100, got %d", trace.BlockNumber)
	}
}

func TestExecutionTraceLen(t *testing.T) {
	var nilTrace *ExecutionTrace
	if nilTrace.Len() != 0 {
		t.Fatal("nil trace should have length 0")
	}

	empty := &ExecutionTrace{}
	if empty.Len() != 0 {
		t.Fatal("empty trace should have length 0")
	}
}

func TestExecutionTraceHash(t *testing.T) {
	trace := BuildValidTrace(1, 4, [32]byte{0x01})
	h := trace.Hash()
	if h == ([32]byte{}) {
		t.Fatal("trace hash should not be zero")
	}
	// Deterministic.
	h2 := trace.Hash()
	if h != h2 {
		t.Fatal("trace hash should be deterministic")
	}

	// Different trace should produce different hash.
	trace2 := BuildValidTrace(1, 4, [32]byte{0x02})
	h3 := trace2.Hash()
	if h == h3 {
		t.Fatal("different traces should have different hashes")
	}
}

func TestExecutionTraceHashNil(t *testing.T) {
	var nilTrace *ExecutionTrace
	h := nilTrace.Hash()
	if h != ([32]byte{}) {
		t.Fatal("nil trace hash should be zero")
	}
}

func TestExecutionTraceVerifyChain(t *testing.T) {
	trace := BuildValidTrace(1, 5, [32]byte{0x01})
	if err := trace.VerifyChain(); err != nil {
		t.Fatalf("valid trace should verify: %v", err)
	}
}

func TestExecutionTraceVerifyChainBroken(t *testing.T) {
	trace := BuildValidTrace(1, 5, [32]byte{0x01})
	// Break the chain by modifying a step's output.
	trace.Steps[2].OutputStateHash = [32]byte{0xff}
	err := trace.VerifyChain()
	if err != ErrTraceHashMismatch {
		t.Fatalf("expected ErrTraceHashMismatch, got %v", err)
	}
}

func TestExecutionTraceVerifyChainPreRootMismatch(t *testing.T) {
	trace := BuildValidTrace(1, 3, [32]byte{0x01})
	trace.PreStateRoot = [32]byte{0xff} // Mismatch.
	err := trace.VerifyChain()
	if err != ErrTraceRootMismatch {
		t.Fatalf("expected ErrTraceRootMismatch, got %v", err)
	}
}

func TestExecutionTraceVerifyChainPostRootMismatch(t *testing.T) {
	trace := BuildValidTrace(1, 3, [32]byte{0x01})
	trace.PostStateRoot = [32]byte{0xff} // Mismatch.
	err := trace.VerifyChain()
	if err != ErrTraceRootMismatch {
		t.Fatalf("expected ErrTraceRootMismatch, got %v", err)
	}
}

func TestExecutionTraceVerifyChainEmpty(t *testing.T) {
	trace := &ExecutionTrace{}
	err := trace.VerifyChain()
	if err != ErrTraceEmpty {
		t.Fatalf("expected ErrTraceEmpty, got %v", err)
	}
}

func TestExecutionTraceSubTrace(t *testing.T) {
	trace := BuildValidTrace(1, 10, [32]byte{0x01})
	sub, err := trace.SubTrace(2, 5)
	if err != nil {
		t.Fatalf("expected success: %v", err)
	}
	if sub.Len() != 3 {
		t.Fatalf("expected 3 steps, got %d", sub.Len())
	}
	if sub.PreStateRoot != trace.Steps[2].InputStateHash {
		t.Fatal("sub-trace pre-state should match step 2 input")
	}
	if sub.PostStateRoot != trace.Steps[4].OutputStateHash {
		t.Fatal("sub-trace post-state should match step 4 output")
	}
}

func TestExecutionTraceSubTraceErrors(t *testing.T) {
	trace := BuildValidTrace(1, 5, [32]byte{0x01})
	_, err := trace.SubTrace(-1, 3)
	if err != ErrDisputeIdxRange {
		t.Fatalf("expected ErrDisputeIdxRange, got %v", err)
	}
	_, err = trace.SubTrace(3, 2)
	if err != ErrDisputeIdxRange {
		t.Fatalf("expected ErrDisputeIdxRange for start >= end, got %v", err)
	}
	_, err = trace.SubTrace(0, 100)
	if err != ErrDisputeIdxRange {
		t.Fatalf("expected ErrDisputeIdxRange for end out of range, got %v", err)
	}
}

func TestTraceProofBuilderGenerateProof(t *testing.T) {
	trace := BuildValidTrace(100, 10, [32]byte{0x01})
	// Corrupt one step to create fraud.
	trace.Steps[5].OutputStateHash = [32]byte{0xff}

	builder := NewTraceProofBuilder(nil) // Use default verifier.
	proof, err := builder.GenerateProof(
		100, trace.PreStateRoot, [32]byte{0xaa}, [32]byte{0xbb}, trace,
	)
	if err != nil {
		t.Fatalf("expected proof generation to succeed: %v", err)
	}
	if proof.InvalidStepIndex != 5 {
		t.Fatalf("expected invalid step index 5, got %d", proof.InvalidStepIndex)
	}
	if proof.BlockNumber != 100 {
		t.Fatalf("expected block 100, got %d", proof.BlockNumber)
	}
}

func TestTraceProofBuilderGenerateProofErrors(t *testing.T) {
	builder := NewTraceProofBuilder(nil)

	// Block zero.
	_, err := builder.GenerateProof(0, [32]byte{1}, [32]byte{2}, [32]byte{3}, BuildValidTrace(1, 1, [32]byte{1}))
	if err != ErrTraceBlockZero {
		t.Fatalf("expected ErrTraceBlockZero, got %v", err)
	}

	// Nil trace.
	_, err = builder.GenerateProof(1, [32]byte{1}, [32]byte{2}, [32]byte{3}, nil)
	if err != ErrTraceEmpty {
		t.Fatalf("expected ErrTraceEmpty, got %v", err)
	}

	// Matching roots.
	_, err = builder.GenerateProof(1, [32]byte{1}, [32]byte{2}, [32]byte{2}, BuildValidTrace(1, 1, [32]byte{1}))
	if err != ErrFraudProofRootsMatch {
		t.Fatalf("expected ErrFraudProofRootsMatch, got %v", err)
	}

	// All valid steps (no fraud).
	validTrace := BuildValidTrace(1, 5, [32]byte{0x01})
	_, err = builder.GenerateProof(1, [32]byte{1}, [32]byte{2}, [32]byte{3}, validTrace)
	if err != ErrDisputeNoInvalid {
		t.Fatalf("expected ErrDisputeNoInvalid, got %v", err)
	}
}

func TestTraceProofBuilderValidateProof(t *testing.T) {
	trace := BuildValidTrace(100, 10, [32]byte{0x01})
	trace.Steps[3].OutputStateHash = [32]byte{0xff} // Inject fault.

	builder := NewTraceProofBuilder(nil)
	proof, _ := builder.GenerateProof(100, trace.PreStateRoot, [32]byte{0xaa}, [32]byte{0xbb}, trace)

	if !builder.ValidateProof(proof) {
		t.Fatal("expected proof to validate")
	}

	// Nil proof.
	if builder.ValidateProof(nil) {
		t.Fatal("nil proof should not validate")
	}
}

func TestDisputeResolverBisectTrace(t *testing.T) {
	trace := BuildValidTrace(1, 10, [32]byte{0x01})
	dr := NewDisputeResolver(64, nil)

	left, right, mid, err := dr.BisectTrace(trace, 5)
	if err != nil {
		t.Fatalf("expected bisect to succeed: %v", err)
	}
	if mid != 5 {
		t.Fatalf("expected mid=5, got %d", mid)
	}
	if left.Len() != 5 {
		t.Fatalf("expected left length 5, got %d", left.Len())
	}
	if right.Len() != 5 {
		t.Fatalf("expected right length 5, got %d", right.Len())
	}
}

func TestDisputeResolverBisectSingleStep(t *testing.T) {
	trace := BuildValidTrace(1, 1, [32]byte{0x01})
	dr := NewDisputeResolver(64, nil)
	_, _, _, err := dr.BisectTrace(trace, 0)
	if err != ErrBisectSingleStep {
		t.Fatalf("expected ErrBisectSingleStep, got %v", err)
	}
}

func TestDisputeResolverVerifyStep(t *testing.T) {
	dr := NewDisputeResolver(64, nil)

	validStep := BuildValidStep(0, 0x60, 21000, [32]byte{0x01})
	if !dr.VerifyStep(&validStep) {
		t.Fatal("expected valid step to verify")
	}

	invalidStep := validStep
	invalidStep.OutputStateHash = [32]byte{0xff}
	if dr.VerifyStep(&invalidStep) {
		t.Fatal("expected invalid step to fail verification")
	}

	if dr.VerifyStep(nil) {
		t.Fatal("nil step should fail verification")
	}
}

func TestDisputeResolverResolve(t *testing.T) {
	trace := BuildValidTrace(1, 16, [32]byte{0x01})
	// Inject fault at step 7.
	trace.Steps[7].OutputStateHash = [32]byte{0xff}

	dr := NewDisputeResolver(64, nil)
	idx, err := dr.Resolve(trace)
	if err != nil {
		t.Fatalf("expected resolve to succeed: %v", err)
	}
	if idx != 7 {
		t.Fatalf("expected invalid step at index 7, got %d", idx)
	}
	if !dr.IsResolved() {
		t.Fatal("expected dispute to be resolved")
	}
	if dr.ResolvedIndex() != 7 {
		t.Fatalf("expected resolved index 7, got %d", dr.ResolvedIndex())
	}
}

func TestDisputeResolverResolveFirstStep(t *testing.T) {
	trace := BuildValidTrace(1, 8, [32]byte{0x01})
	// Inject fault at the very first step.
	trace.Steps[0].OutputStateHash = [32]byte{0xff}

	dr := NewDisputeResolver(64, nil)
	idx, err := dr.Resolve(trace)
	if err != nil {
		t.Fatalf("expected resolve to succeed: %v", err)
	}
	if idx != 0 {
		t.Fatalf("expected invalid step at index 0, got %d", idx)
	}
}

func TestDisputeResolverResolveLastStep(t *testing.T) {
	trace := BuildValidTrace(1, 8, [32]byte{0x01})
	// Inject fault at the last step.
	trace.Steps[7].OutputStateHash = [32]byte{0xff}

	dr := NewDisputeResolver(64, nil)
	idx, err := dr.Resolve(trace)
	if err != nil {
		t.Fatalf("expected resolve to succeed: %v", err)
	}
	if idx != 7 {
		t.Fatalf("expected invalid step at index 7, got %d", idx)
	}
}

func TestDisputeResolverNoInvalid(t *testing.T) {
	trace := BuildValidTrace(1, 8, [32]byte{0x01})
	dr := NewDisputeResolver(64, nil)
	_, err := dr.Resolve(trace)
	if err != ErrDisputeNoInvalid {
		t.Fatalf("expected ErrDisputeNoInvalid, got %v", err)
	}
}

func TestDisputeResolverAlreadyResolved(t *testing.T) {
	trace := BuildValidTrace(1, 4, [32]byte{0x01})
	trace.Steps[2].OutputStateHash = [32]byte{0xff}

	dr := NewDisputeResolver(64, nil)
	dr.Resolve(trace)

	_, err := dr.Resolve(trace)
	if err != ErrDisputeConverged {
		t.Fatalf("expected ErrDisputeConverged, got %v", err)
	}
}

func TestDisputeResolverResolveEmpty(t *testing.T) {
	dr := NewDisputeResolver(64, nil)
	_, err := dr.Resolve(nil)
	if err != ErrTraceEmpty {
		t.Fatalf("expected ErrTraceEmpty, got %v", err)
	}
}

func TestTotalGasUsed(t *testing.T) {
	trace := BuildValidTrace(1, 3, [32]byte{0x01})
	total := TotalGasUsed(trace)
	// Gas: 21000+0*100, 21000+1*100, 21000+2*100 = 21000+21100+21200 = 63300
	expected := uint64(21000 + 21100 + 21200)
	if total != expected {
		t.Fatalf("expected total gas %d, got %d", expected, total)
	}

	// Nil trace.
	if TotalGasUsed(nil) != 0 {
		t.Fatal("nil trace should have 0 gas")
	}
}

func TestSortStepsByGas(t *testing.T) {
	steps := []ExecutionStep{
		{GasUsed: 300},
		{GasUsed: 100},
		{GasUsed: 200},
	}
	sorted := SortStepsByGas(steps)
	if sorted[0].GasUsed != 100 || sorted[1].GasUsed != 200 || sorted[2].GasUsed != 300 {
		t.Fatalf("expected ascending gas order, got %d %d %d",
			sorted[0].GasUsed, sorted[1].GasUsed, sorted[2].GasUsed)
	}
	// Original should be unmodified.
	if steps[0].GasUsed != 300 {
		t.Fatal("sort should not modify original slice")
	}
}

func TestBuildValidTraceChaining(t *testing.T) {
	trace := BuildValidTrace(1, 5, [32]byte{0x01})
	// Verify all steps chain correctly.
	for i := 0; i < len(trace.Steps)-1; i++ {
		if trace.Steps[i].OutputStateHash != trace.Steps[i+1].InputStateHash {
			t.Fatalf("step %d output does not chain to step %d input", i, i+1)
		}
	}
	// First input matches pre-state.
	if trace.Steps[0].InputStateHash != trace.PreStateRoot {
		t.Fatal("first step input should match pre-state root")
	}
	// Last output matches post-state.
	if trace.Steps[4].OutputStateHash != trace.PostStateRoot {
		t.Fatal("last step output should match post-state root")
	}
}

func TestCustomStepVerifier(t *testing.T) {
	// Create a verifier that always considers step index 3 invalid.
	custom := func(step *ExecutionStep) bool {
		return step.Index != 3
	}

	trace := BuildValidTrace(1, 5, [32]byte{0x01})
	builder := NewTraceProofBuilder(custom)
	idx, err := builder.FindInvalidStep(trace)
	if err != nil {
		t.Fatalf("expected to find invalid step: %v", err)
	}
	if idx != 3 {
		t.Fatalf("expected invalid step at index 3, got %d", idx)
	}
}

// computeExpectedOutputTest is a helper for the test to verify independently.
func computeExpectedOutputTest(input [32]byte, opcode byte, gas uint64) [32]byte {
	var data []byte
	data = append(data, input[:]...)
	data = append(data, opcode)
	var buf [8]byte
	_ = buf
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0)
	copy(data[len(data)-8:], make([]byte, 8))
	// Recompute properly.
	data = data[:0]
	data = append(data, input[:]...)
	data = append(data, opcode)
	b := make([]byte, 8)
	b[0] = byte(gas >> 56)
	b[1] = byte(gas >> 48)
	b[2] = byte(gas >> 40)
	b[3] = byte(gas >> 32)
	b[4] = byte(gas >> 24)
	b[5] = byte(gas >> 16)
	b[6] = byte(gas >> 8)
	b[7] = byte(gas)
	data = append(data, b...)
	h := crypto.Keccak256(data)
	var result [32]byte
	copy(result[:], h)
	return result
}

func TestComputeExpectedOutputConsistency(t *testing.T) {
	input := [32]byte{0x01}
	step := BuildValidStep(0, 0x60, 21000, input)
	expected := computeExpectedOutputTest(input, 0x60, 21000)
	if step.OutputStateHash != expected {
		t.Fatal("output state hash does not match independent computation")
	}
}
