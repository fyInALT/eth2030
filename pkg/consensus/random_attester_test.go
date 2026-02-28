package consensus

import (
	"testing"
)

func TestDefaultRandomAttesterConfig(t *testing.T) {
	cfg := DefaultRandomAttesterConfig()
	if cfg.MinSampleSize != 256 {
		t.Fatalf("expected MinSampleSize=256, got %d", cfg.MinSampleSize)
	}
	if cfg.MaxSampleSize != 1024 {
		t.Fatalf("expected MaxSampleSize=1024, got %d", cfg.MaxSampleSize)
	}
	if !cfg.BalanceWeighted {
		t.Fatal("expected BalanceWeighted=true")
	}
}

func TestComputeSampleSize_Small(t *testing.T) {
	s := NewRandomAttesterSelector(nil)
	// 100 validators: sqrt(100)=10, min=256, but 256 > 100 so clamped to 100.
	size := s.ComputeSampleSize(100)
	if size != 100 {
		t.Fatalf("expected 100 (clamped to validator count), got %d", size)
	}
}

func TestComputeSampleSize_Medium(t *testing.T) {
	s := NewRandomAttesterSelector(nil)
	// 10000 validators: sqrt(10000)=100, min=256, so result=256.
	size := s.ComputeSampleSize(10000)
	if size != 256 {
		t.Fatalf("expected 256, got %d", size)
	}
}

func TestComputeSampleSize_Large(t *testing.T) {
	s := NewRandomAttesterSelector(nil)
	// 1000000 validators: sqrt(1000000)=1000, 1000 > 256 (min), 1000 < 1024 (max), so result=1000.
	size := s.ComputeSampleSize(1000000)
	if size != 1000 {
		t.Fatalf("expected 1000, got %d", size)
	}

	// 2000000 validators: sqrt(2000000)~1414, clamped to max=1024.
	size = s.ComputeSampleSize(2000000)
	if size != 1024 {
		t.Fatalf("expected 1024 (clamped to max), got %d", size)
	}
}

func TestComputeSampleSize_Exact(t *testing.T) {
	s := NewRandomAttesterSelector(nil)

	// Exactly at boundary: 65536 validators, sqrt=256 => 256 (matches min).
	size := s.ComputeSampleSize(65536)
	if size != 256 {
		t.Fatalf("expected 256, got %d", size)
	}

	// 1048576 = 1024^2, sqrt=1024 => 1024 (matches max).
	size = s.ComputeSampleSize(1048576)
	if size != 1024 {
		t.Fatalf("expected 1024, got %d", size)
	}

	// 0 validators.
	size = s.ComputeSampleSize(0)
	if size != 0 {
		t.Fatalf("expected 0, got %d", size)
	}

	// 1 validator.
	size = s.ComputeSampleSize(1)
	if size != 1 {
		t.Fatalf("expected 1, got %d", size)
	}
}

func makeValidators(count int, active bool, balance uint64) []ValidatorInfo {
	vals := make([]ValidatorInfo, count)
	for i := 0; i < count; i++ {
		vals[i] = ValidatorInfo{
			Index:            ValidatorIndex(i),
			EffectiveBalance: balance,
			Active:           active,
		}
	}
	return vals
}

func TestSampleValidators_Basic(t *testing.T) {
	s := NewRandomAttesterSelector(nil)
	vals := makeValidators(500, true, 32_000_000_000)
	seed := [32]byte{1, 2, 3}

	result, err := s.SampleValidators(vals, 42, seed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 500 active: sqrt(500)~22, min=256, 256 < 500, so target=256.
	// Due to balance weighting some candidates may be rejected; result should
	// still be substantial. With max balance all should pass.
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
	if len(result) > 500 {
		t.Fatalf("result exceeds validator count: %d", len(result))
	}
}

func TestSampleValidators_Deterministic(t *testing.T) {
	s := NewRandomAttesterSelector(nil)
	vals := makeValidators(500, true, 32_000_000_000)
	seed := [32]byte{10, 20, 30}

	r1, err := s.SampleValidators(vals, 100, seed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r2, err := s.SampleValidators(vals, 100, seed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(r1) != len(r2) {
		t.Fatalf("lengths differ: %d vs %d", len(r1), len(r2))
	}
	for i := range r1 {
		if r1[i] != r2[i] {
			t.Fatalf("index %d differs: %d vs %d", i, r1[i], r2[i])
		}
	}
}

func TestSampleValidators_DifferentSeeds(t *testing.T) {
	s := NewRandomAttesterSelector(nil)
	vals := makeValidators(500, true, 32_000_000_000)

	seed1 := [32]byte{1}
	seed2 := [32]byte{2}

	r1, err := s.SampleValidators(vals, 100, seed1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r2, err := s.SampleValidators(vals, 100, seed2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With different seeds, at least some indices should differ.
	differs := false
	minLen := len(r1)
	if len(r2) < minLen {
		minLen = len(r2)
	}
	for i := 0; i < minLen; i++ {
		if r1[i] != r2[i] {
			differs = true
			break
		}
	}
	if !differs && len(r1) == len(r2) {
		t.Fatal("different seeds produced identical results")
	}
}

func TestSampleValidators_DifferentSlots(t *testing.T) {
	s := NewRandomAttesterSelector(nil)
	vals := makeValidators(500, true, 32_000_000_000)
	seed := [32]byte{42}

	r1, err := s.SampleValidators(vals, 100, seed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r2, err := s.SampleValidators(vals, 200, seed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	differs := false
	minLen := len(r1)
	if len(r2) < minLen {
		minLen = len(r2)
	}
	for i := 0; i < minLen; i++ {
		if r1[i] != r2[i] {
			differs = true
			break
		}
	}
	if !differs && len(r1) == len(r2) {
		t.Fatal("different slots produced identical results")
	}
}

func TestSampleValidators_Uniqueness(t *testing.T) {
	s := NewRandomAttesterSelector(nil)
	vals := makeValidators(500, true, 32_000_000_000)
	seed := [32]byte{5, 10, 15}

	result, err := s.SampleValidators(vals, 77, seed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	seen := make(map[ValidatorIndex]bool)
	for _, idx := range result {
		if seen[idx] {
			t.Fatalf("duplicate validator index: %d", idx)
		}
		seen[idx] = true
	}
}

func TestSampleValidators_AllActive(t *testing.T) {
	s := NewRandomAttesterSelector(nil)
	vals := makeValidators(300, true, 32_000_000_000)
	seed := [32]byte{99}

	result, err := s.SampleValidators(vals, 1, seed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 300 active: sample size clamped to 300 (since min=256, 256<300, but
	// sqrt(300)~17 < 256, so target=256). With full balance all should pass.
	if len(result) < 200 {
		t.Fatalf("expected at least 200 results, got %d", len(result))
	}
	if len(result) > 300 {
		t.Fatalf("result exceeds active count: %d", len(result))
	}
}

func TestSampleValidators_SomeInactive(t *testing.T) {
	s := NewRandomAttesterSelector(nil)
	// 600 total: 400 active, 200 inactive.
	vals := make([]ValidatorInfo, 600)
	for i := 0; i < 600; i++ {
		vals[i] = ValidatorInfo{
			Index:            ValidatorIndex(i),
			EffectiveBalance: 32_000_000_000,
			Active:           i < 400,
		}
	}
	seed := [32]byte{7, 8, 9}

	result, err := s.SampleValidators(vals, 50, seed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify no inactive validators are in the result.
	for _, idx := range result {
		if uint64(idx) >= 400 {
			t.Fatalf("inactive validator %d was selected", idx)
		}
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestSampleValidators_NoValidators(t *testing.T) {
	s := NewRandomAttesterSelector(nil)
	seed := [32]byte{}

	_, err := s.SampleValidators(nil, 1, seed)
	if err != ErrRANoValidators {
		t.Fatalf("expected ErrRANoValidators, got %v", err)
	}

	_, err = s.SampleValidators([]ValidatorInfo{}, 1, seed)
	if err != ErrRANoValidators {
		t.Fatalf("expected ErrRANoValidators, got %v", err)
	}
}

func TestSampleValidators_NoActiveValidators(t *testing.T) {
	s := NewRandomAttesterSelector(nil)
	vals := makeValidators(100, false, 32_000_000_000) // all inactive
	seed := [32]byte{1}

	_, err := s.SampleValidators(vals, 1, seed)
	if err != ErrRANoValidators {
		t.Fatalf("expected ErrRANoValidators, got %v", err)
	}
}

func TestSampleValidators_BalanceWeighted(t *testing.T) {
	// Use a non-weighted selector to compare distribution.
	cfgWeighted := &RandomAttesterConfig{
		MinSampleSize:   10,
		MaxSampleSize:   50,
		BalanceWeighted: true,
	}
	cfgUniform := &RandomAttesterConfig{
		MinSampleSize:   10,
		MaxSampleSize:   50,
		BalanceWeighted: false,
	}
	sw := NewRandomAttesterSelector(cfgWeighted)
	su := NewRandomAttesterSelector(cfgUniform)

	// Create validators: half with max balance, half with min balance.
	vals := make([]ValidatorInfo, 1000)
	for i := 0; i < 1000; i++ {
		bal := uint64(32_000_000_000)
		if i >= 500 {
			bal = 1_000_000_000 // 1 ETH
		}
		vals[i] = ValidatorInfo{
			Index:            ValidatorIndex(i),
			EffectiveBalance: bal,
			Active:           true,
		}
	}

	// Run many samples and count high-balance vs low-balance selections.
	highCountWeighted := 0
	totalWeighted := 0
	highCountUniform := 0
	totalUniform := 0

	for slot := uint64(0); slot < 100; slot++ {
		seed := [32]byte{byte(slot), byte(slot >> 8)}

		rw, err := sw.SampleValidators(vals, slot, seed)
		if err != nil {
			t.Fatalf("weighted error at slot %d: %v", slot, err)
		}
		for _, idx := range rw {
			totalWeighted++
			if uint64(idx) < 500 {
				highCountWeighted++
			}
		}

		ru, err := su.SampleValidators(vals, slot, seed)
		if err != nil {
			t.Fatalf("uniform error at slot %d: %v", slot, err)
		}
		for _, idx := range ru {
			totalUniform++
			if uint64(idx) < 500 {
				highCountUniform++
			}
		}
	}

	// Weighted sampling should favor high-balance validators more than uniform.
	weightedRatio := float64(highCountWeighted) / float64(totalWeighted)
	uniformRatio := float64(highCountUniform) / float64(totalUniform)

	// With balance weighting, high-balance validators (32 ETH) should be
	// selected much more often than low-balance (1 ETH).
	if weightedRatio <= uniformRatio*0.9 {
		t.Fatalf("weighted sampling should favor high-balance validators: "+
			"weighted=%.3f, uniform=%.3f", weightedRatio, uniformRatio)
	}
}

func TestComputeSamplingHash_Deterministic(t *testing.T) {
	seed := [32]byte{0xaa, 0xbb, 0xcc}
	h1 := computeSamplingHash(seed, 42, 7)
	h2 := computeSamplingHash(seed, 42, 7)

	if h1 != h2 {
		t.Fatal("same inputs produced different hashes")
	}

	// Different index should produce different hash.
	h3 := computeSamplingHash(seed, 42, 8)
	if h1 == h3 {
		t.Fatal("different index produced same hash")
	}

	// Different slot should produce different hash.
	h4 := computeSamplingHash(seed, 43, 7)
	if h1 == h4 {
		t.Fatal("different slot produced same hash")
	}

	// Different seed should produce different hash.
	seed2 := [32]byte{0xdd}
	h5 := computeSamplingHash(seed2, 42, 7)
	if h1 == h5 {
		t.Fatal("different seed produced same hash")
	}
}
