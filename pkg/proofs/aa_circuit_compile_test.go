package proofs

import "testing"

func TestAACircuitCompile(t *testing.T) {
	circuit, err := CompileAACircuit()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if circuit == nil {
		t.Fatal("expected non-nil circuit")
	}
	if circuit.Name != "AAValidationCircuit" {
		t.Errorf("expected name AAValidationCircuit, got %s", circuit.Name)
	}
	if circuit.PublicInputCount != 5 {
		t.Errorf("expected 5 public inputs, got %d", circuit.PublicInputCount)
	}
}

func TestSetupKeys(t *testing.T) {
	circuit, err := CompileAACircuit()
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	pk, vk, err := SetupKeys(circuit)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if pk == nil {
		t.Fatal("expected non-nil proving key")
	}
	if vk == nil {
		t.Fatal("expected non-nil verifying key")
	}
}

func TestSetupKeys_NilCircuit(t *testing.T) {
	_, _, err := SetupKeys(nil)
	if err != ErrCircuitNilDef {
		t.Fatalf("expected ErrCircuitNilDef, got %v", err)
	}
}

func TestAACircuitCompileConstraintCount(t *testing.T) {
	circuit, err := CompileAACircuit()
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if circuit.ConstraintCount() <= 0 {
		t.Errorf("expected constraint count > 0, got %d", circuit.ConstraintCount())
	}
}
