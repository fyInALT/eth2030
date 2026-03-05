// groth16_gnark_test.go tests the real gnark Groth16 integration (PQ-6.1/6.2).
// These tests use actual gnark proving/verification with BN254 pairings.
package proofs

import (
	"testing"
)

// TestAACircuitCompileGnark verifies that the AA validation circuit compiles
// to an R1CS with the expected constraint structure (PQ-6.2).
func TestAACircuitCompileGnark(t *testing.T) {
	circuit, err := CompileAACircuitGnark()
	if err != nil {
		t.Fatalf("CompileAACircuitGnark: %v", err)
	}
	if circuit == nil {
		t.Fatal("expected non-nil circuit")
	}
	if circuit.ConstraintCount() <= 0 {
		t.Errorf("expected constraint count > 0, got %d", circuit.ConstraintCount())
	}
	t.Logf("AAValidationGnarkCircuit: %d R1CS constraints", circuit.ConstraintCount())
}

// TestSetupGnarkAAKeys verifies that proving and verifying keys are generated
// from the compiled AA circuit (PQ-6.2).
func TestSetupGnarkAAKeys(t *testing.T) {
	circuit, err := CompileAACircuitGnark()
	if err != nil {
		t.Fatalf("CompileAACircuitGnark: %v", err)
	}

	keys, err := SetupGnarkAAKeys(circuit)
	if err != nil {
		t.Fatalf("SetupGnarkAAKeys: %v", err)
	}
	if keys == nil {
		t.Fatal("expected non-nil keys")
	}
	if keys.ProvingKey == nil {
		t.Fatal("expected non-nil proving key")
	}
	if keys.VerifyingKey == nil {
		t.Fatal("expected non-nil verifying key")
	}
}

// TestGroth16RealVerify verifies the complete prove-then-verify round-trip
// using real gnark BN254 Groth16 pairings (PQ-6.1).
func TestGroth16RealVerify(t *testing.T) {
	// Step 1: Compile circuit.
	circuit, err := CompileAACircuitGnark()
	if err != nil {
		t.Fatalf("CompileAACircuitGnark: %v", err)
	}

	// Step 2: Generate proving and verifying keys.
	keys, err := SetupGnarkAAKeys(circuit)
	if err != nil {
		t.Fatalf("SetupGnarkAAKeys: %v", err)
	}

	// Step 3: Prove AA validation: nonce=1, gasLimit=21000, prevNonce=0.
	proof, err := ProveGnarkAA(circuit, keys, 1, 21000, 0)
	if err != nil {
		t.Fatalf("ProveGnarkAA: %v", err)
	}
	if proof == nil {
		t.Fatal("expected non-nil proof")
	}

	// Step 4: Verify proof against public inputs.
	ok, err := VerifyGnarkAAProof(keys, proof, 1, 21000)
	if err != nil {
		t.Fatalf("VerifyGnarkAAProof: %v", err)
	}
	if !ok {
		t.Fatal("proof verification failed")
	}
}

// TestGroth16RealVerify_TamperedPublicInput verifies that wrong public inputs
// cause verification to fail (PQ-6.1 tamper test).
func TestGroth16RealVerify_TamperedPublicInput(t *testing.T) {
	circuit, err := CompileAACircuitGnark()
	if err != nil {
		t.Fatalf("CompileAACircuitGnark: %v", err)
	}
	keys, err := SetupGnarkAAKeys(circuit)
	if err != nil {
		t.Fatalf("SetupGnarkAAKeys: %v", err)
	}

	// Prove for nonce=2, prevNonce=1.
	proof, err := ProveGnarkAA(circuit, keys, 2, 50000, 1)
	if err != nil {
		t.Fatalf("ProveGnarkAA: %v", err)
	}

	// Tampered: verify with wrong nonce.
	_, err = VerifyGnarkAAProof(keys, proof, 99, 50000)
	if err == nil {
		t.Fatal("expected verification error with wrong nonce, got nil")
	}
	t.Logf("tampered nonce correctly rejected: %v", err)
}

// TestGroth16RealVerify_MultipleNonces verifies proofs for different nonces.
func TestGroth16RealVerify_MultipleNonces(t *testing.T) {
	circuit, err := CompileAACircuitGnark()
	if err != nil {
		t.Fatalf("CompileAACircuitGnark: %v", err)
	}
	keys, err := SetupGnarkAAKeys(circuit)
	if err != nil {
		t.Fatalf("SetupGnarkAAKeys: %v", err)
	}

	cases := []struct {
		nonce uint64
		gas   uint64
		prevN uint64
	}{
		{1, 21000, 0},
		{2, 50000, 1},
		{10, 100000, 9},
	}

	for _, tc := range cases {
		proof, err := ProveGnarkAA(circuit, keys, tc.nonce, tc.gas, tc.prevN)
		if err != nil {
			t.Errorf("ProveGnarkAA(nonce=%d): %v", tc.nonce, err)
			continue
		}
		ok, err := VerifyGnarkAAProof(keys, proof, tc.nonce, tc.gas)
		if err != nil {
			t.Errorf("VerifyGnarkAAProof(nonce=%d): %v", tc.nonce, err)
			continue
		}
		if !ok {
			t.Errorf("verification failed for nonce=%d", tc.nonce)
		}
	}
}

// TestProveGnarkAA_ZeroNonce verifies that zero nonce is rejected (invalid AA op).
func TestProveGnarkAA_ZeroNonce(t *testing.T) {
	circuit, err := CompileAACircuitGnark()
	if err != nil {
		t.Fatalf("CompileAACircuitGnark: %v", err)
	}
	keys, err := SetupGnarkAAKeys(circuit)
	if err != nil {
		t.Fatalf("SetupGnarkAAKeys: %v", err)
	}

	_, err = ProveGnarkAA(circuit, keys, 0, 21000, 0)
	if err != ErrGnarkBadArgs {
		t.Fatalf("expected ErrGnarkBadArgs for zero nonce, got %v", err)
	}
}

// TestSetupGnarkAAKeys_NilCircuit verifies nil handling.
func TestSetupGnarkAAKeys_NilCircuit(t *testing.T) {
	_, err := SetupGnarkAAKeys(nil)
	if err != ErrGnarkNilCS {
		t.Fatalf("expected ErrGnarkNilCS, got %v", err)
	}
}

// TestGnarkIntegrationStatus verifies the status string indicates real gnark.
func TestGnarkIntegrationStatus(t *testing.T) {
	status := GnarkIntegrationStatus()
	if status == "" {
		t.Fatal("expected non-empty gnark status")
	}
	if status != "gnark-groth16-bn254-real" {
		t.Errorf("unexpected status %q", status)
	}
}
