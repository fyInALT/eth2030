package groth16

import (
	"testing"
)

func TestGroth16_CompileAACircuit(t *testing.T) {
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
}

func TestGroth16_SetupKeys(t *testing.T) {
	circuit, err := CompileAACircuitGnark()
	if err != nil {
		t.Fatalf("CompileAACircuitGnark: %v", err)
	}
	keys, err := SetupGnarkAAKeys(circuit)
	if err != nil {
		t.Fatalf("SetupGnarkAAKeys: %v", err)
	}
	if keys == nil || keys.ProvingKey == nil || keys.VerifyingKey == nil {
		t.Fatal("expected non-nil keys")
	}
}

func TestGroth16_ProveAndVerify(t *testing.T) {
	circuit, err := CompileAACircuitGnark()
	if err != nil {
		t.Fatalf("CompileAACircuitGnark: %v", err)
	}
	keys, err := SetupGnarkAAKeys(circuit)
	if err != nil {
		t.Fatalf("SetupGnarkAAKeys: %v", err)
	}

	proof, err := ProveGnarkAA(circuit, keys, 1, 21000, 0)
	if err != nil {
		t.Fatalf("ProveGnarkAA: %v", err)
	}

	ok, err := VerifyGnarkAAProof(keys, proof, 1, 21000)
	if err != nil {
		t.Fatalf("VerifyGnarkAAProof: %v", err)
	}
	if !ok {
		t.Fatal("proof verification failed")
	}
}

func TestGroth16_NilHandling(t *testing.T) {
	if _, err := SetupGnarkAAKeys(nil); err != ErrGnarkNilCS {
		t.Fatalf("expected ErrGnarkNilCS, got %v", err)
	}

	circuit, _ := CompileAACircuitGnark()
	keys, _ := SetupGnarkAAKeys(circuit)
	if _, err := ProveGnarkAA(circuit, keys, 0, 21000, 0); err != ErrGnarkBadArgs {
		t.Fatalf("expected ErrGnarkBadArgs for zero nonce, got %v", err)
	}
}

func TestGroth16_IntegrationStatus(t *testing.T) {
	if GnarkIntegrationStatus() != "gnark-groth16-bn254-real" {
		t.Errorf("unexpected status %q", GnarkIntegrationStatus())
	}
}
