package prover

import "testing"

func TestValidationFrameCircuit_ValidFrame(t *testing.T) {
	prover := NewSTARKValidationFrameProver()
	calldata := []byte("test calldata")
	output := []byte{1, 2, 3} // non-zero first byte

	proof, err := prover.ProveValidationFrame(calldata, output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proof == nil {
		t.Fatal("expected non-nil proof")
	}
	if proof.TraceLength != 1 {
		t.Errorf("expected trace length 1, got %d", proof.TraceLength)
	}
}

func TestValidationFrameCircuit_RevertedFrame(t *testing.T) {
	prover := NewSTARKValidationFrameProver()
	calldata := []byte("test calldata")
	output := []byte{0} // first byte is 0 -> reverted

	proof, err := prover.ProveValidationFrame(calldata, output)
	if err != ErrFrameReverted {
		t.Fatalf("expected ErrFrameReverted, got %v", err)
	}
	if proof != nil {
		t.Fatal("expected nil proof for reverted frame")
	}
}

func TestValidationFrameCircuit_NilOutput(t *testing.T) {
	prover := NewSTARKValidationFrameProver()
	calldata := []byte("test")

	proof, err := prover.ProveValidationFrame(calldata, nil)
	if err != ErrFrameNilOutput {
		t.Fatalf("expected ErrFrameNilOutput, got %v", err)
	}
	if proof != nil {
		t.Fatal("expected nil proof for nil output")
	}
}

func TestValidationFrameCircuit_Batch(t *testing.T) {
	prover := NewSTARKValidationFrameProver()

	for _, count := range []int{1, 10, 100} {
		frames := make([][]byte, count)
		for i := range frames {
			frames[i] = []byte{byte(i + 1), 0xAA, 0xBB}
		}
		proof, err := prover.ProveAllValidationFrames(frames)
		if err != nil {
			t.Fatalf("batch %d: unexpected error: %v", count, err)
		}
		if proof == nil {
			t.Fatalf("batch %d: expected non-nil proof", count)
		}
		if proof.TraceLength != uint64(count) {
			t.Errorf("batch %d: expected trace length %d, got %d", count, count, proof.TraceLength)
		}
	}
}

func TestValidationFrameProofSize100(t *testing.T) {
	prover := NewSTARKValidationFrameProver()
	frames := make([][]byte, 100)
	for i := range frames {
		frames[i] = []byte{byte(i + 1), 0xCC}
	}
	proof, err := prover.ProveAllValidationFrames(frames)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	size := ValidationFrameProofSize(proof)
	maxSize := 128 * 1024
	if size > maxSize {
		t.Errorf("proof size %d exceeds max %d", size, maxSize)
	}
}

func TestStubProver_Interface(t *testing.T) {
	var _ ValidationFrameProver = &StubValidationFrameProver{}
}

func TestStubProver_AlwaysSucceeds(t *testing.T) {
	stub := &StubValidationFrameProver{}
	proof, err := stub.ProveValidationFrame([]byte("data"), []byte{1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proof == nil {
		t.Fatal("expected non-nil proof")
	}
	if !stub.Verify(proof) {
		t.Fatal("stub verify should return true for non-nil proof")
	}
}

func TestSTARKProver_Roundtrip(t *testing.T) {
	prover := NewSTARKValidationFrameProver()
	calldata := []byte("roundtrip test")
	output := []byte{42}

	proof, err := prover.ProveValidationFrame(calldata, output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !prover.Verify(proof) {
		t.Fatal("expected proof to verify successfully")
	}
}
