package zkvm

import (
	"math/big"
	"testing"
)

func TestDefaultPoseidon2Params(t *testing.T) {
	params := DefaultPoseidon2Params()
	if params.T != 3 {
		t.Fatalf("expected T=3, got %d", params.T)
	}
	if params.ExternalRounds != 8 {
		t.Fatalf("expected 8 external rounds, got %d", params.ExternalRounds)
	}
	if params.InternalRounds != 56 {
		t.Fatalf("expected 56 internal rounds, got %d", params.InternalRounds)
	}

	// Verify round constant count: T * externalRounds + internalRounds.
	expectedRC := params.T*params.ExternalRounds + params.InternalRounds
	if len(params.RoundConstants) != expectedRC {
		t.Fatalf("expected %d round constants, got %d", expectedRC, len(params.RoundConstants))
	}

	if len(params.DiagMDS) != params.T {
		t.Fatalf("expected DiagMDS of length %d, got %d", params.T, len(params.DiagMDS))
	}

	if params.Field.Cmp(bn254ScalarField) != 0 {
		t.Fatal("field should be BN254 scalar field")
	}
}

func TestPoseidon2Hash_Deterministic(t *testing.T) {
	params := DefaultPoseidon2Params()
	a := big.NewInt(42)
	b := big.NewInt(99)

	h1 := Poseidon2Hash(params, a, b)
	h2 := Poseidon2Hash(params, a, b)

	if h1.Cmp(h2) != 0 {
		t.Fatal("Poseidon2 hash should be deterministic")
	}
}

func TestPoseidon2Hash_SingleElement(t *testing.T) {
	params := DefaultPoseidon2Params()
	h := Poseidon2Hash(params, big.NewInt(42))

	if h == nil || h.Sign() < 0 {
		t.Fatal("hash should be non-nil and non-negative")
	}
	if h.Cmp(params.Field) >= 0 {
		t.Fatal("hash should be less than field modulus")
	}
}

func TestPoseidon2Hash_MultipleElements(t *testing.T) {
	params := DefaultPoseidon2Params()

	// Test 2 elements.
	h2 := Poseidon2Hash(params, big.NewInt(1), big.NewInt(2))
	if h2 == nil || h2.Cmp(params.Field) >= 0 {
		t.Fatal("hash of 2 elements should be valid field element")
	}

	// Test 3 elements (more than rate=2 for t=3).
	h3 := Poseidon2Hash(params, big.NewInt(1), big.NewInt(2), big.NewInt(3))
	if h3 == nil || h3.Cmp(params.Field) >= 0 {
		t.Fatal("hash of 3 elements should be valid field element")
	}

	// Test 5 elements (multi-block absorption).
	h5 := Poseidon2Hash(params,
		big.NewInt(1), big.NewInt(2), big.NewInt(3),
		big.NewInt(4), big.NewInt(5))
	if h5 == nil || h5.Cmp(params.Field) >= 0 {
		t.Fatal("hash of 5 elements should be valid field element")
	}

	// All should be different.
	if h2.Cmp(h3) == 0 {
		t.Fatal("h2 and h3 should differ")
	}
	if h3.Cmp(h5) == 0 {
		t.Fatal("h3 and h5 should differ")
	}
}

func TestPoseidon2Hash_EmptyInput(t *testing.T) {
	params := DefaultPoseidon2Params()
	h := Poseidon2Hash(params)

	if h == nil {
		t.Fatal("hash of empty input should be non-nil")
	}
	if h.Cmp(params.Field) >= 0 {
		t.Fatal("hash should be less than field modulus")
	}
}

func TestPoseidon2Hash_DifferentInputs(t *testing.T) {
	params := DefaultPoseidon2Params()

	h1 := Poseidon2Hash(params, big.NewInt(1), big.NewInt(2))
	h2 := Poseidon2Hash(params, big.NewInt(3), big.NewInt(4))

	if h1.Cmp(h2) == 0 {
		t.Fatal("different inputs should produce different hashes")
	}
}

func TestPoseidon2Hash_InField(t *testing.T) {
	params := DefaultPoseidon2Params()

	// Test with various inputs.
	inputs := []*big.Int{
		big.NewInt(0),
		big.NewInt(1),
		big.NewInt(1000000),
		new(big.Int).Sub(params.Field, big.NewInt(1)),
	}
	for _, inp := range inputs {
		h := Poseidon2Hash(params, inp)
		if h.Sign() < 0 || h.Cmp(params.Field) >= 0 {
			t.Fatalf("hash of %s should be in field [0, p), got %s", inp.String(), h.String())
		}
	}
}

func TestPoseidon2Permutation_Deterministic(t *testing.T) {
	params := DefaultPoseidon2Params()

	// Create two identical states.
	s1 := []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)}
	s2 := []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)}

	r1 := poseidon2Permutation(s1, params)
	r2 := poseidon2Permutation(s2, params)

	for i := 0; i < params.T; i++ {
		if r1[i].Cmp(r2[i]) != 0 {
			t.Fatalf("permutation element %d differs: %s vs %s",
				i, r1[i].String(), r2[i].String())
		}
	}
}

func TestPoseidon2ExternalLinear(t *testing.T) {
	field := bn254ScalarField
	diag := generateDiagonalMDS(3, field)

	// Test with known state.
	state := []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)}
	result := poseidon2ExternalLinear(state, diag, field)

	if len(result) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(result))
	}

	// Verify: sum = 1+2+3 = 6, y[i] = state[i] + diag[i] * 6.
	sum := big.NewInt(6)
	for i := 0; i < 3; i++ {
		expected := new(big.Int).Mul(diag[i], sum)
		expected.Add(expected, state[i])
		expected.Mod(expected, field)
		if result[i].Cmp(expected) != 0 {
			t.Fatalf("element %d: got %s, want %s", i, result[i].String(), expected.String())
		}
	}
}

func TestPoseidon2InternalLinear(t *testing.T) {
	field := bn254ScalarField
	diag := generateDiagonalMDS(3, field)

	state := []*big.Int{big.NewInt(5), big.NewInt(10), big.NewInt(15)}
	result := poseidon2InternalLinear(state, diag, field)

	// Internal linear uses the same structure as external.
	expected := poseidon2ExternalLinear(
		[]*big.Int{big.NewInt(5), big.NewInt(10), big.NewInt(15)},
		diag, field,
	)
	for i := 0; i < 3; i++ {
		if result[i].Cmp(expected[i]) != 0 {
			t.Fatalf("element %d: got %s, want %s", i, result[i].String(), expected[i].String())
		}
	}
}

func TestPoseidon2Sponge_Basic(t *testing.T) {
	sponge := NewPoseidon2Sponge(nil)
	sponge.Absorb(big.NewInt(1), big.NewInt(2))
	results := sponge.Squeeze(1)
	if len(results) != 1 {
		t.Fatalf("expected 1 squeeze result, got %d", len(results))
	}
	if results[0] == nil || results[0].Sign() < 0 {
		t.Fatal("squeeze result should be non-nil and non-negative")
	}
}

func TestPoseidon2Sponge_Streaming(t *testing.T) {
	sponge := NewPoseidon2Sponge(nil)
	sponge.Absorb(big.NewInt(1))
	sponge.Absorb(big.NewInt(2))
	sponge.Absorb(big.NewInt(3))
	results := sponge.Squeeze(1)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0] == nil {
		t.Fatal("result should not be nil")
	}
}

func TestPoseidon2Sponge_MultipleSqueeze(t *testing.T) {
	sponge := NewPoseidon2Sponge(nil)
	sponge.Absorb(big.NewInt(7))
	results := sponge.Squeeze(4) // more than rate
	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}
	for i, r := range results {
		if r == nil {
			t.Fatalf("result %d is nil", i)
		}
	}
}

func TestPoseidon2HashBytes(t *testing.T) {
	data := []byte("hello world")
	h := Poseidon2HashBytes(data)

	// Result should not be all zeros (extremely unlikely for real hash).
	allZero := true
	for _, b := range h {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Fatal("hash of non-empty data should not be all zeros")
	}
}

func TestPoseidon2HashBytes_Deterministic(t *testing.T) {
	data := []byte("deterministic test input")
	h1 := Poseidon2HashBytes(data)
	h2 := Poseidon2HashBytes(data)

	if h1 != h2 {
		t.Fatal("Poseidon2HashBytes should be deterministic")
	}
}

func TestPoseidon2VsPoseidon1(t *testing.T) {
	// Poseidon2 and Poseidon1 should produce different outputs for the same input.
	a := big.NewInt(42)
	b := big.NewInt(99)

	h1 := PoseidonHash(nil, a, b)
	h2 := Poseidon2Hash(nil, a, b)

	if h1.Cmp(h2) == 0 {
		t.Fatal("Poseidon1 and Poseidon2 should produce different hashes")
	}
}

func TestPoseidon2DiagonalMDS(t *testing.T) {
	field := bn254ScalarField
	diag := generateDiagonalMDS(3, field)

	if len(diag) != 3 {
		t.Fatalf("expected 3 diagonal entries, got %d", len(diag))
	}

	// Verify values: d_i = 1 + (i+1)^2.
	for i := 0; i < 3; i++ {
		iPlusOne := int64(i + 1)
		expected := new(big.Int).SetInt64(1 + iPlusOne*iPlusOne)
		expected.Mod(expected, field)
		if diag[i].Cmp(expected) != 0 {
			t.Fatalf("diag[%d] = %s, want %s", i, diag[i].String(), expected.String())
		}
	}
}

func TestPoseidon2RoundConstants(t *testing.T) {
	field := bn254ScalarField
	rcs := generatePoseidon2RoundConstants(3, 80, field)

	if len(rcs) != 80 {
		t.Fatalf("expected 80 round constants, got %d", len(rcs))
	}

	// All should be within field.
	for i, rc := range rcs {
		if rc.Sign() < 0 || rc.Cmp(field) >= 0 {
			t.Fatalf("round constant %d out of field range", i)
		}
	}

	// Should be deterministic.
	rcs2 := generatePoseidon2RoundConstants(3, 80, field)
	for i := range rcs {
		if rcs[i].Cmp(rcs2[i]) != 0 {
			t.Fatalf("round constant %d differs between calls", i)
		}
	}

	// Should differ from Poseidon1 constants (different seed).
	rcs1 := generateRoundConstants(3, 80, field)
	differ := false
	for i := range rcs {
		if rcs[i].Cmp(rcs1[i]) != 0 {
			differ = true
			break
		}
	}
	if !differ {
		t.Fatal("Poseidon2 round constants should differ from Poseidon1")
	}
}

func TestPoseidon2RoundConstantsGrainLFSR(t *testing.T) {
	field := bn254ScalarField
	rcs := generatePoseidon2RoundConstants(3, 80, field)

	// All constants must be non-zero.
	for i, rc := range rcs {
		if rc.Sign() == 0 {
			t.Fatalf("Poseidon2 round constant %d is zero; Grain LFSR should not produce zero constants", i)
		}
	}

	// All constants must be strictly within the field [0, p).
	for i, rc := range rcs {
		if rc.Sign() < 0 || rc.Cmp(field) >= 0 {
			t.Fatalf("Poseidon2 round constant %d out of field range: %s", i, rc.String())
		}
	}

	// Determinism.
	rcs2 := generatePoseidon2RoundConstants(3, 80, field)
	for i := range rcs {
		if rcs[i].Cmp(rcs2[i]) != 0 {
			t.Fatalf("Poseidon2 round constant %d differs between calls (not deterministic)", i)
		}
	}

	// Must differ from Poseidon1 constants (different seed).
	rcs1 := generateRoundConstants(3, 80, field)
	differ := false
	for i := 0; i < len(rcs) && i < len(rcs1); i++ {
		if rcs[i].Cmp(rcs1[i]) != 0 {
			differ = true
			break
		}
	}
	if !differ {
		t.Fatal("Poseidon2 round constants should differ from Poseidon1 (different seed)")
	}
}

func TestPoseidon2HashBytes_Empty(t *testing.T) {
	h := Poseidon2HashBytes(nil)
	// Should produce a valid hash even for empty input.
	allZero := true
	for _, b := range h {
		if b != 0 {
			allZero = false
			break
		}
	}
	// Empty input hashes to the Poseidon2 permutation of zero state,
	// which is non-zero.
	if allZero {
		t.Fatal("hash of empty data should not be all zeros")
	}
}

func TestPoseidon2Hash_NilParams(t *testing.T) {
	h := Poseidon2Hash(nil, big.NewInt(1))
	if h == nil || h.Sign() < 0 {
		t.Fatal("nil params should use defaults")
	}
}

func TestPoseidon2Hash_InputReduction(t *testing.T) {
	params := DefaultPoseidon2Params()
	// Input larger than field should be reduced.
	large := new(big.Int).Add(params.Field, big.NewInt(5))
	h1 := Poseidon2Hash(params, large)
	h2 := Poseidon2Hash(params, big.NewInt(5))

	if h1.Cmp(h2) != 0 {
		t.Fatal("input should be reduced mod field before hashing")
	}
}

func TestPoseidon2Hash_OrderMatters(t *testing.T) {
	params := DefaultPoseidon2Params()
	h1 := Poseidon2Hash(params, big.NewInt(1), big.NewInt(2))
	h2 := Poseidon2Hash(params, big.NewInt(2), big.NewInt(1))
	if h1.Cmp(h2) == 0 {
		t.Fatal("hash should depend on input order")
	}
}

func TestPoseidon2Sponge_Deterministic(t *testing.T) {
	s1 := NewPoseidon2Sponge(nil)
	s1.Absorb(big.NewInt(42))
	r1 := s1.Squeeze(1)

	s2 := NewPoseidon2Sponge(nil)
	s2.Absorb(big.NewInt(42))
	r2 := s2.Squeeze(1)

	if r1[0].Cmp(r2[0]) != 0 {
		t.Fatal("sponge should be deterministic")
	}
}

func TestPoseidon2HashBytes_DifferentInputs(t *testing.T) {
	h1 := Poseidon2HashBytes([]byte("alice"))
	h2 := Poseidon2HashBytes([]byte("bob"))

	if h1 == h2 {
		t.Fatal("different byte inputs should produce different hashes")
	}
}

// --- Benchmarks ---

func BenchmarkPoseidon2Hash_TwoInputs(b *testing.B) {
	params := DefaultPoseidon2Params()
	a := big.NewInt(12345)
	bv := big.NewInt(67890)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Poseidon2Hash(params, a, bv)
	}
}

func BenchmarkPoseidon2HashBytes(b *testing.B) {
	data := []byte("benchmark test data for Poseidon2")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Poseidon2HashBytes(data)
	}
}
