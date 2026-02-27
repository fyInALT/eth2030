package proofs

import (
	"math/big"
	"testing"
)

func TestNewSTARKProver(t *testing.T) {
	p := NewSTARKProver()
	if p.blowupFactor != DefaultBlowupFactor {
		t.Errorf("expected blowup %d, got %d", DefaultBlowupFactor, p.blowupFactor)
	}
	if p.numQueries != DefaultNumQueries {
		t.Errorf("expected queries %d, got %d", DefaultNumQueries, p.numQueries)
	}
	if p.fieldModulus.Cmp(GoldilocksModulus) != 0 {
		t.Error("expected Goldilocks modulus")
	}
}

func TestNewSTARKProverWithParams(t *testing.T) {
	// Valid params.
	p, err := NewSTARKProverWithParams(2, 30, big.NewInt(97))
	if err != nil {
		t.Fatal(err)
	}
	if p.blowupFactor != 2 {
		t.Errorf("expected blowup 2, got %d", p.blowupFactor)
	}

	// Invalid blowup.
	_, err = NewSTARKProverWithParams(3, 30, big.NewInt(97))
	if err != ErrSTARKInvalidBlowup {
		t.Errorf("expected ErrSTARKInvalidBlowup, got %v", err)
	}

	// Nil modulus.
	_, err = NewSTARKProverWithParams(4, 30, nil)
	if err != ErrSTARKInvalidField {
		t.Errorf("expected ErrSTARKInvalidField, got %v", err)
	}
}

func TestSTARKGenerateAndVerify(t *testing.T) {
	p := NewSTARKProver()

	// Simple execution trace: 4 rows, 2 columns.
	trace := [][]FieldElement{
		{NewFieldElement(1), NewFieldElement(2)},
		{NewFieldElement(3), NewFieldElement(4)},
		{NewFieldElement(5), NewFieldElement(6)},
		{NewFieldElement(7), NewFieldElement(8)},
	}
	constraints := []STARKConstraint{
		{Degree: 1, Coefficients: []FieldElement{NewFieldElement(1)}},
	}

	proof, err := p.GenerateSTARKProof(trace, constraints)
	if err != nil {
		t.Fatal(err)
	}

	if proof.TraceLength != 4 {
		t.Errorf("expected trace length 4, got %d", proof.TraceLength)
	}
	if proof.BlowupFactor != DefaultBlowupFactor {
		t.Errorf("expected blowup %d, got %d", DefaultBlowupFactor, proof.BlowupFactor)
	}
	if proof.ConstraintCount != 1 {
		t.Errorf("expected 1 constraint, got %d", proof.ConstraintCount)
	}
	if len(proof.QueryResponses) != int(DefaultNumQueries) {
		t.Errorf("expected %d queries, got %d", DefaultNumQueries, len(proof.QueryResponses))
	}

	// Verify.
	valid, err := p.VerifySTARKProof(proof, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("proof should be valid")
	}
}

func TestSTARKEmptyTrace(t *testing.T) {
	p := NewSTARKProver()
	_, err := p.GenerateSTARKProof(nil, []STARKConstraint{{Degree: 1}})
	if err != ErrSTARKEmptyTrace {
		t.Errorf("expected ErrSTARKEmptyTrace, got %v", err)
	}
}

func TestSTARKNoConstraints(t *testing.T) {
	p := NewSTARKProver()
	trace := [][]FieldElement{{NewFieldElement(1)}}
	_, err := p.GenerateSTARKProof(trace, nil)
	if err != ErrSTARKNoConstraints {
		t.Errorf("expected ErrSTARKNoConstraints, got %v", err)
	}
}

func TestSTARKVerifyNil(t *testing.T) {
	p := NewSTARKProver()
	_, err := p.VerifySTARKProof(nil, nil)
	if err != ErrSTARKInvalidProof {
		t.Errorf("expected ErrSTARKInvalidProof, got %v", err)
	}
}

func TestSTARKProofSize(t *testing.T) {
	p := NewSTARKProver()
	trace := [][]FieldElement{
		{NewFieldElement(1)},
		{NewFieldElement(2)},
	}
	constraints := []STARKConstraint{{Degree: 1, Coefficients: []FieldElement{NewFieldElement(1)}}}

	proof, err := p.GenerateSTARKProof(trace, constraints)
	if err != nil {
		t.Fatal(err)
	}
	size := proof.ProofSize()
	if size <= 0 {
		t.Error("proof size should be positive")
	}
}

func TestGoldilocksModulus(t *testing.T) {
	// Goldilocks: p = 2^64 - 2^32 + 1
	expected := new(big.Int).Lsh(big.NewInt(1), 64)
	expected.Sub(expected, new(big.Int).Lsh(big.NewInt(1), 32))
	expected.Add(expected, big.NewInt(1))

	if GoldilocksModulus.Cmp(expected) != 0 {
		t.Errorf("Goldilocks modulus incorrect: got %s, expected %s", GoldilocksModulus.String(), expected.String())
	}
}

func TestFRILayerCount(t *testing.T) {
	tests := []struct {
		size     uint64
		expected int
	}{
		{1, 0},
		{2, 1},
		{4, 2},
		{8, 3},
		{16, 4},
		{1024, 10},
	}
	for _, tt := range tests {
		got := friLayerCount(tt.size)
		if got != tt.expected {
			t.Errorf("friLayerCount(%d) = %d, want %d", tt.size, got, tt.expected)
		}
	}
}

func TestSTARKLargeTrace(t *testing.T) {
	p := NewSTARKProver()
	// Reasonable size trace.
	trace := make([][]FieldElement, 256)
	for i := range trace {
		trace[i] = []FieldElement{NewFieldElement(int64(i))}
	}
	constraints := []STARKConstraint{{Degree: 2, Coefficients: []FieldElement{NewFieldElement(1), NewFieldElement(2)}}}

	proof, err := p.GenerateSTARKProof(trace, constraints)
	if err != nil {
		t.Fatal(err)
	}

	valid, err := p.VerifySTARKProof(proof, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("large trace proof should be valid")
	}
}
