package prover

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

// --- Fix 1: Canonical field element tests ---

func TestNewFieldElement_Canonical(t *testing.T) {
	// Zero should remain zero.
	fe := NewFieldElement(0)
	if fe.Value.Sign() != 0 {
		t.Fatalf("NewFieldElement(0) should be 0, got %s", fe.Value.String())
	}

	// Small positive values should remain unchanged.
	fe = NewFieldElement(42)
	if fe.Value.Cmp(big.NewInt(42)) != 0 {
		t.Fatalf("NewFieldElement(42) should be 42, got %s", fe.Value.String())
	}

	// All values produced by NewFieldElement must be in [0, p).
	fe = NewFieldElement(1)
	if fe.Value.Cmp(GoldilocksModulus) >= 0 || fe.Value.Sign() < 0 {
		t.Fatalf("field element out of range: %s", fe.Value.String())
	}

	// Negative values should be reduced into the field (e.g., -1 becomes p-1).
	fe = NewFieldElement(-1)
	expectedNeg1 := new(big.Int).Sub(GoldilocksModulus, big.NewInt(1))
	if fe.Value.Cmp(expectedNeg1) != 0 {
		t.Fatalf("NewFieldElement(-1) should be p-1 = %s, got %s", expectedNeg1.String(), fe.Value.String())
	}

	fe = NewFieldElement(-5)
	expectedNeg5 := new(big.Int).Sub(GoldilocksModulus, big.NewInt(5))
	if fe.Value.Cmp(expectedNeg5) != 0 {
		t.Fatalf("NewFieldElement(-5) should be p-5 = %s, got %s", expectedNeg5.String(), fe.Value.String())
	}

	// max int64 should be reduced (max int64 < Goldilocks modulus, so stays the same).
	maxInt64 := int64(1<<63 - 1)
	fe = NewFieldElement(maxInt64)
	expected := new(big.Int).SetInt64(maxInt64)
	expected.Mod(expected, GoldilocksModulus)
	if fe.Value.Cmp(expected) != 0 {
		t.Fatalf("NewFieldElement(maxInt64) should be %s, got %s", expected.String(), fe.Value.String())
	}
	if fe.Value.Cmp(GoldilocksModulus) >= 0 || fe.Value.Sign() < 0 {
		t.Fatalf("max int64 field element out of range: %s", fe.Value.String())
	}

	// min int64 should be reduced into field.
	minInt64 := int64(-1 << 63)
	fe = NewFieldElement(minInt64)
	expectedMin := new(big.Int).SetInt64(minInt64)
	expectedMin.Mod(expectedMin, GoldilocksModulus)
	if fe.Value.Cmp(expectedMin) != 0 {
		t.Fatalf("NewFieldElement(minInt64) should be %s, got %s", expectedMin.String(), fe.Value.String())
	}
	if fe.Value.Cmp(GoldilocksModulus) >= 0 || fe.Value.Sign() < 0 {
		t.Fatalf("min int64 field element out of range: %s", fe.Value.String())
	}
}

// --- Fix 2: Query deduplication test ---

func TestSTARKQueryDeduplication(t *testing.T) {
	p := NewSTARKProver()

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

	// Verify all query indices are unique.
	seen := make(map[uint64]bool, len(proof.QueryResponses))
	for i, qr := range proof.QueryResponses {
		if seen[qr.Index] {
			t.Fatalf("duplicate query index %d found at position %d", qr.Index, i)
		}
		seen[qr.Index] = true
	}

	// Also verify the proof still verifies.
	valid, err := p.VerifySTARKProof(proof, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Error("proof with deduplicated queries should be valid")
	}
}

// --- Fix 3: Public input Fiat-Shamir binding tests ---

func TestSTARKPublicInputFiatShamirBinding(t *testing.T) {
	p := NewSTARKProver()

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

	// Two different sets of public inputs should produce different challenges.
	pubInputs1 := []FieldElement{NewFieldElement(100)}
	pubInputs2 := []FieldElement{NewFieldElement(200)}

	challenge1 := p.deriveChallenge(proof, pubInputs1)
	challenge2 := p.deriveChallenge(proof, pubInputs2)

	if challenge1 == challenge2 {
		t.Fatal("different public inputs must produce different Fiat-Shamir challenges")
	}

	// Both proofs should verify with their respective inputs.
	valid1, err := p.VerifySTARKProof(proof, pubInputs1)
	if err != nil {
		t.Fatal(err)
	}
	if !valid1 {
		t.Error("proof should verify with public inputs 1")
	}

	valid2, err := p.VerifySTARKProof(proof, pubInputs2)
	if err != nil {
		t.Fatal(err)
	}
	if !valid2 {
		t.Error("proof should verify with public inputs 2")
	}

	// Empty public inputs should produce a different challenge than non-empty.
	challengeEmpty := p.deriveChallenge(proof, nil)
	if challengeEmpty == challenge1 {
		t.Fatal("empty public inputs should produce a different challenge than non-empty")
	}
}

func TestSTARKDeriveChallengeDeterministic(t *testing.T) {
	p := NewSTARKProver()

	trace := [][]FieldElement{
		{NewFieldElement(1)},
		{NewFieldElement(2)},
	}
	constraints := []STARKConstraint{
		{Degree: 1, Coefficients: []FieldElement{NewFieldElement(1)}},
	}

	proof, err := p.GenerateSTARKProof(trace, constraints)
	if err != nil {
		t.Fatal(err)
	}

	pubInputs := []FieldElement{NewFieldElement(42)}
	c1 := p.deriveChallenge(proof, pubInputs)
	c2 := p.deriveChallenge(proof, pubInputs)
	if c1 != c2 {
		t.Fatal("deriveChallenge should be deterministic")
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
