package proofs

import (
	"crypto/sha256"
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

func TestSTARKConstraintEvaluation(t *testing.T) {
	p := NewSTARKProver()

	trace1 := [][]FieldElement{
		{NewFieldElement(1), NewFieldElement(2)},
		{NewFieldElement(3), NewFieldElement(4)},
	}
	trace2 := [][]FieldElement{
		{NewFieldElement(10), NewFieldElement(20)},
		{NewFieldElement(30), NewFieldElement(40)},
	}
	constraints := []STARKConstraint{
		{Degree: 1, Coefficients: []FieldElement{NewFieldElement(1), NewFieldElement(1)}},
	}

	proof1, err := p.GenerateSTARKProof(trace1, constraints)
	if err != nil {
		t.Fatal(err)
	}
	proof2, err := p.GenerateSTARKProof(trace2, constraints)
	if err != nil {
		t.Fatal(err)
	}

	// Different traces should produce different constraint eval commitments.
	if proof1.ConstraintEvalCommitment == proof2.ConstraintEvalCommitment {
		t.Error("different traces should produce different constraint eval commitments")
	}

	// Neither should be zero.
	var zero [32]byte
	if proof1.ConstraintEvalCommitment == zero {
		t.Error("constraint eval commitment should not be zero")
	}
	if proof2.ConstraintEvalCommitment == zero {
		t.Error("constraint eval commitment should not be zero")
	}
}

func TestSTARKMerkleAuthPath(t *testing.T) {
	// Create some leaves.
	leaves := make([][32]byte, 4)
	for i := range leaves {
		h := sha256.New()
		h.Write([]byte{byte(i)})
		copy(leaves[i][:], h.Sum(nil))
	}

	root := merkleRoot(leaves)

	// Test auth path for each leaf.
	for i := uint64(0); i < 4; i++ {
		path := merkleAuthPath(leaves, i)
		if !verifyMerkleAuthPath(leaves[i], i, path, root) {
			t.Errorf("auth path verification failed for leaf %d", i)
		}
	}

	// Verify wrong leaf fails.
	wrongLeaf := [32]byte{0xFF}
	path := merkleAuthPath(leaves, 0)
	if verifyMerkleAuthPath(wrongLeaf, 0, path, root) {
		t.Error("auth path should fail for wrong leaf")
	}
}

func TestSTARKFRIFolding(t *testing.T) {
	p := NewSTARKProver()

	trace1 := [][]FieldElement{
		{NewFieldElement(1), NewFieldElement(2)},
		{NewFieldElement(3), NewFieldElement(4)},
	}
	trace2 := [][]FieldElement{
		{NewFieldElement(100), NewFieldElement(200)},
		{NewFieldElement(300), NewFieldElement(400)},
	}
	constraints := []STARKConstraint{
		{Degree: 1, Coefficients: []FieldElement{NewFieldElement(1)}},
	}

	proof1, err := p.GenerateSTARKProof(trace1, constraints)
	if err != nil {
		t.Fatal(err)
	}
	proof2, err := p.GenerateSTARKProof(trace2, constraints)
	if err != nil {
		t.Fatal(err)
	}

	// FRI commitments should differ for different traces.
	if len(proof1.FRICommitments) != len(proof2.FRICommitments) {
		t.Fatal("FRI commitment counts should match for same-size traces")
	}

	allSame := true
	for i := range proof1.FRICommitments {
		if proof1.FRICommitments[i] != proof2.FRICommitments[i] {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("FRI commitments should differ for different trace data")
	}
}

func TestSTARKAggregator_EndToEnd_WithConstraints(t *testing.T) {
	// Create a prover and generate a proof with multiple constraints.
	p := NewSTARKProver()

	trace := [][]FieldElement{
		{NewFieldElement(100), NewFieldElement(200)},
		{NewFieldElement(300), NewFieldElement(400)},
		{NewFieldElement(500), NewFieldElement(600)},
		{NewFieldElement(700), NewFieldElement(800)},
	}

	// Two meaningful constraints like the aggregator uses.
	constraints := []STARKConstraint{
		{Degree: 1, Coefficients: []FieldElement{NewFieldElement(1), NewFieldElement(1)}},
		{Degree: 1, Coefficients: []FieldElement{NewFieldElement(0), NewFieldElement(0), NewFieldElement(1)}},
	}

	proof, err := p.GenerateSTARKProof(trace, constraints)
	if err != nil {
		t.Fatal(err)
	}

	// Verify constraint eval commitment is non-zero.
	var zero [32]byte
	if proof.ConstraintEvalCommitment == zero {
		t.Error("constraint eval commitment should be non-zero")
	}

	// Verify constraint count matches.
	if proof.ConstraintCount != 2 {
		t.Errorf("expected 2 constraints, got %d", proof.ConstraintCount)
	}

	// Verify the proof is valid.
	valid, err := p.VerifySTARKProof(proof, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("proof should be valid")
	}

	// Verify that tampering with constraint eval commitment causes rejection.
	tampered := *proof
	tampered.ConstraintEvalCommitment = [32]byte{}
	valid, err = p.VerifySTARKProof(&tampered, nil)
	if valid || err == nil {
		t.Error("tampered proof with zero constraint eval commitment should fail")
	}
}
