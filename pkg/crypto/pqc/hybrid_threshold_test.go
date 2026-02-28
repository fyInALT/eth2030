package pqc

import (
	"errors"
	"sync"
	"testing"
)

func TestHybridThresholdConfigValid(t *testing.T) {
	cfg, err := NewHybridThresholdConfig(3, 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Threshold != 3 {
		t.Errorf("expected threshold=3, got %d", cfg.Threshold)
	}
	if cfg.Total != 5 {
		t.Errorf("expected total=5, got %d", cfg.Total)
	}
	if cfg.RequireBoth {
		t.Error("expected RequireBoth=false by default")
	}
}

func TestHybridThresholdConfigTGreaterThanN(t *testing.T) {
	_, err := NewHybridThresholdConfig(6, 5)
	if !errors.Is(err, ErrHTThresholdExceeds) {
		t.Fatalf("expected ErrHTThresholdExceeds, got %v", err)
	}
}

func TestHybridThresholdConfigTZero(t *testing.T) {
	_, err := NewHybridThresholdConfig(0, 5)
	if !errors.Is(err, ErrHTInvalidThreshold) {
		t.Fatalf("expected ErrHTInvalidThreshold, got %v", err)
	}
}

func TestHybridThresholdConfigNZero(t *testing.T) {
	_, err := NewHybridThresholdConfig(1, 0)
	if !errors.Is(err, ErrHTInvalidTotal) {
		t.Fatalf("expected ErrHTInvalidTotal, got %v", err)
	}
}

func TestHybridThresholdAddSharesAndMeetsThreshold(t *testing.T) {
	cfg, _ := NewHybridThresholdConfig(2, 5)
	agg := NewHybridThresholdAggregator(cfg)

	// Add first share - below threshold.
	err := agg.AddShare(&HybridThresholdShare{
		SignerIndex:  0,
		ClassicalSig: []byte("bls-sig-0"),
		PQSig:        []byte("pq-sig-0"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if agg.MeetsThreshold() {
		t.Error("should not meet threshold with 1 share (t=2)")
	}

	// Add second share - at threshold.
	err = agg.AddShare(&HybridThresholdShare{
		SignerIndex:  1,
		ClassicalSig: []byte("bls-sig-1"),
		PQSig:        []byte("pq-sig-1"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !agg.MeetsThreshold() {
		t.Error("should meet threshold with 2 shares (t=2)")
	}
}

func TestHybridThresholdBelowThreshold(t *testing.T) {
	cfg, _ := NewHybridThresholdConfig(3, 5)
	agg := NewHybridThresholdAggregator(cfg)

	for i := 0; i < 2; i++ {
		agg.AddShare(&HybridThresholdShare{
			SignerIndex:  i,
			ClassicalSig: []byte("sig"),
			PQSig:        []byte("pq"),
		})
	}
	if agg.MeetsThreshold() {
		t.Error("should not meet threshold with 2 shares (t=3)")
	}
}

func TestHybridThresholdDuplicateSigner(t *testing.T) {
	cfg, _ := NewHybridThresholdConfig(2, 5)
	agg := NewHybridThresholdAggregator(cfg)

	share := &HybridThresholdShare{
		SignerIndex:  0,
		ClassicalSig: []byte("sig"),
		PQSig:        []byte("pq"),
	}
	if err := agg.AddShare(share); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err := agg.AddShare(share)
	if !errors.Is(err, ErrHTDuplicateSigner) {
		t.Fatalf("expected ErrHTDuplicateSigner, got %v", err)
	}
}

func TestHybridThresholdShareOnlyClassical(t *testing.T) {
	cfg, _ := NewHybridThresholdConfig(1, 3)
	agg := NewHybridThresholdAggregator(cfg)

	err := agg.AddShare(&HybridThresholdShare{
		SignerIndex:  0,
		ClassicalSig: []byte("classical-only"),
	})
	if err != nil {
		t.Fatalf("expected no error for classical-only share, got %v", err)
	}
	if agg.ValidClassicalCount() != 1 {
		t.Errorf("expected classical count 1, got %d", agg.ValidClassicalCount())
	}
	if agg.ValidPQCount() != 0 {
		t.Errorf("expected PQ count 0, got %d", agg.ValidPQCount())
	}
}

func TestHybridThresholdShareOnlyPQ(t *testing.T) {
	cfg, _ := NewHybridThresholdConfig(1, 3)
	agg := NewHybridThresholdAggregator(cfg)

	err := agg.AddShare(&HybridThresholdShare{
		SignerIndex: 1,
		PQSig:       []byte("pq-only"),
	})
	if err != nil {
		t.Fatalf("expected no error for PQ-only share, got %v", err)
	}
	if agg.ValidClassicalCount() != 0 {
		t.Errorf("expected classical count 0, got %d", agg.ValidClassicalCount())
	}
	if agg.ValidPQCount() != 1 {
		t.Errorf("expected PQ count 1, got %d", agg.ValidPQCount())
	}
}

func TestHybridThresholdShareBothSigs(t *testing.T) {
	cfg, _ := NewHybridThresholdConfig(1, 3)
	agg := NewHybridThresholdAggregator(cfg)

	err := agg.AddShare(&HybridThresholdShare{
		SignerIndex:  2,
		ClassicalSig: []byte("classical"),
		PQSig:        []byte("pq"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if agg.ValidClassicalCount() != 1 || agg.ValidPQCount() != 1 {
		t.Error("expected both classical and PQ counts to be 1")
	}
}

func TestHybridThresholdShareNeitherSig(t *testing.T) {
	cfg, _ := NewHybridThresholdConfig(1, 3)
	agg := NewHybridThresholdAggregator(cfg)

	err := agg.AddShare(&HybridThresholdShare{
		SignerIndex: 0,
	})
	if !errors.Is(err, ErrHTEmptyShare) {
		t.Fatalf("expected ErrHTEmptyShare, got %v", err)
	}
}

func TestHybridThresholdRequireBothRejectsSingleSig(t *testing.T) {
	cfg, _ := NewHybridThresholdConfig(1, 3)
	cfg.RequireBoth = true
	agg := NewHybridThresholdAggregator(cfg)

	// Classical only.
	err := agg.AddShare(&HybridThresholdShare{
		SignerIndex:  0,
		ClassicalSig: []byte("classical"),
	})
	if !errors.Is(err, ErrHTRequireBoth) {
		t.Fatalf("expected ErrHTRequireBoth for classical-only, got %v", err)
	}

	// PQ only.
	err = agg.AddShare(&HybridThresholdShare{
		SignerIndex: 1,
		PQSig:       []byte("pq"),
	})
	if !errors.Is(err, ErrHTRequireBoth) {
		t.Fatalf("expected ErrHTRequireBoth for PQ-only, got %v", err)
	}

	// Both present - should succeed.
	err = agg.AddShare(&HybridThresholdShare{
		SignerIndex:  2,
		ClassicalSig: []byte("classical"),
		PQSig:        []byte("pq"),
	})
	if err != nil {
		t.Fatalf("expected no error for both sigs with RequireBoth, got %v", err)
	}
}

func TestHybridThresholdReset(t *testing.T) {
	cfg, _ := NewHybridThresholdConfig(1, 3)
	agg := NewHybridThresholdAggregator(cfg)

	agg.AddShare(&HybridThresholdShare{
		SignerIndex:  0,
		ClassicalSig: []byte("sig"),
	})
	if agg.ShareCount() != 1 {
		t.Fatalf("expected 1 share, got %d", agg.ShareCount())
	}

	agg.Reset()
	if agg.ShareCount() != 0 {
		t.Fatalf("expected 0 shares after reset, got %d", agg.ShareCount())
	}
	if agg.MeetsThreshold() {
		t.Error("should not meet threshold after reset")
	}
}

func TestHybridThresholdShareCountAccuracy(t *testing.T) {
	cfg, _ := NewHybridThresholdConfig(2, 5)
	agg := NewHybridThresholdAggregator(cfg)

	for i := 0; i < 4; i++ {
		agg.AddShare(&HybridThresholdShare{
			SignerIndex:  i,
			ClassicalSig: []byte("sig"),
			PQSig:        []byte("pq"),
		})
	}
	if agg.ShareCount() != 4 {
		t.Errorf("expected share count 4, got %d", agg.ShareCount())
	}
	if agg.ValidClassicalCount() != 4 {
		t.Errorf("expected classical count 4, got %d", agg.ValidClassicalCount())
	}
	if agg.ValidPQCount() != 4 {
		t.Errorf("expected PQ count 4, got %d", agg.ValidPQCount())
	}
}

func TestHybridThresholdResultWeights(t *testing.T) {
	cfg, _ := NewHybridThresholdConfig(2, 5)
	cfg.Weighting = SignatureWeighting{ClassicalWeight: 0.6, PQWeight: 0.4}
	agg := NewHybridThresholdAggregator(cfg)

	// Share with both sigs: weight = 0.6 + 0.4 = 1.0.
	agg.AddShare(&HybridThresholdShare{
		SignerIndex:  0,
		ClassicalSig: []byte("c"),
		PQSig:        []byte("p"),
	})
	// Share with classical only: weight = 0.6.
	agg.AddShare(&HybridThresholdShare{
		SignerIndex:  1,
		ClassicalSig: []byte("c"),
	})
	// Share with PQ only: weight = 0.4.
	agg.AddShare(&HybridThresholdShare{
		SignerIndex: 2,
		PQSig:       []byte("p"),
	})

	result := agg.Result()
	if result.ClassicalCount != 2 {
		t.Errorf("expected classical count 2, got %d", result.ClassicalCount)
	}
	if result.PQCount != 2 {
		t.Errorf("expected PQ count 2, got %d", result.PQCount)
	}
	if result.TotalShares != 3 {
		t.Errorf("expected total shares 3, got %d", result.TotalShares)
	}
	// Total weight: 1.0 + 0.6 + 0.4 = 2.0.
	expectedWeight := 2.0
	if result.TotalWeight < expectedWeight-0.001 || result.TotalWeight > expectedWeight+0.001 {
		t.Errorf("expected total weight ~%.1f, got %.4f", expectedWeight, result.TotalWeight)
	}
	if !result.MetThreshold {
		t.Error("expected threshold to be met with 3 shares (t=2)")
	}
}

func TestHybridThresholdConcurrentAddShare(t *testing.T) {
	cfg, _ := NewHybridThresholdConfig(50, 100)
	agg := NewHybridThresholdAggregator(cfg)

	var wg sync.WaitGroup
	errs := make([]error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			errs[idx] = agg.AddShare(&HybridThresholdShare{
				SignerIndex:  idx,
				ClassicalSig: []byte("sig"),
				PQSig:        []byte("pq"),
			})
		}(i)
	}

	wg.Wait()

	errCount := 0
	for _, e := range errs {
		if e != nil {
			errCount++
		}
	}
	if errCount != 0 {
		t.Errorf("expected 0 errors from concurrent adds, got %d", errCount)
	}
	if agg.ShareCount() != 100 {
		t.Errorf("expected 100 shares, got %d", agg.ShareCount())
	}
	if !agg.MeetsThreshold() {
		t.Error("should meet threshold with 100 shares (t=50)")
	}
}

func TestHybridThresholdDegenerateOneOfOne(t *testing.T) {
	cfg, _ := NewHybridThresholdConfig(1, 1)
	agg := NewHybridThresholdAggregator(cfg)

	err := agg.AddShare(&HybridThresholdShare{
		SignerIndex:  0,
		ClassicalSig: []byte("only-signer"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !agg.MeetsThreshold() {
		t.Error("1-of-1 should meet threshold immediately")
	}
}

func TestHybridThresholdLargeThreshold(t *testing.T) {
	cfg, _ := NewHybridThresholdConfig(67, 100)
	agg := NewHybridThresholdAggregator(cfg)

	for i := 0; i < 66; i++ {
		agg.AddShare(&HybridThresholdShare{
			SignerIndex: i,
			PQSig:       []byte("pq"),
		})
	}
	if agg.MeetsThreshold() {
		t.Error("should not meet threshold with 66 shares (t=67)")
	}

	agg.AddShare(&HybridThresholdShare{
		SignerIndex: 66,
		PQSig:       []byte("pq"),
	})
	if !agg.MeetsThreshold() {
		t.Error("should meet threshold with 67 shares (t=67)")
	}
}

func TestHybridThresholdSignerIndexOutOfRange(t *testing.T) {
	cfg, _ := NewHybridThresholdConfig(1, 5)
	agg := NewHybridThresholdAggregator(cfg)

	err := agg.AddShare(&HybridThresholdShare{
		SignerIndex:  5, // out of range [0, 5)
		ClassicalSig: []byte("sig"),
	})
	if !errors.Is(err, ErrHTSignerOutOfRange) {
		t.Fatalf("expected ErrHTSignerOutOfRange, got %v", err)
	}

	err = agg.AddShare(&HybridThresholdShare{
		SignerIndex:  -1,
		ClassicalSig: []byte("sig"),
	})
	if !errors.Is(err, ErrHTSignerOutOfRange) {
		t.Fatalf("expected ErrHTSignerOutOfRange for negative index, got %v", err)
	}
}

func TestHybridThresholdNilShare(t *testing.T) {
	cfg, _ := NewHybridThresholdConfig(1, 3)
	agg := NewHybridThresholdAggregator(cfg)

	err := agg.AddShare(nil)
	if !errors.Is(err, ErrHTNilShare) {
		t.Fatalf("expected ErrHTNilShare, got %v", err)
	}
}

func TestValidateShareFunction(t *testing.T) {
	msg := []byte("test message")
	both := &HybridThresholdShare{ClassicalSig: []byte("c"), PQSig: []byte("p")}
	cOnly := &HybridThresholdShare{ClassicalSig: []byte("c")}
	pOnly := &HybridThresholdShare{PQSig: []byte("p")}
	empty := &HybridThresholdShare{SignerIndex: 0}

	cases := []struct {
		name string
		s    *HybridThresholdShare
		m    []byte
		rb   bool
		want error
	}{
		{"both_ok", both, msg, false, nil},
		{"nil_share", nil, msg, false, ErrHTNilShare},
		{"empty_msg", cOnly, nil, false, ErrHTEmptyMessage},
		{"no_sigs", empty, msg, false, ErrHTEmptyShare},
		{"require_both_classical_only", cOnly, msg, true, ErrHTInvalidPQ},
		{"require_both_pq_only", pOnly, msg, true, ErrHTInvalidClassical},
	}
	for _, tc := range cases {
		err := ValidateShare(tc.s, tc.m, tc.rb)
		if tc.want == nil && err != nil {
			t.Errorf("%s: unexpected error %v", tc.name, err)
		} else if tc.want != nil && !errors.Is(err, tc.want) {
			t.Errorf("%s: expected %v, got %v", tc.name, tc.want, err)
		}
	}
}

func TestHybridThresholdConfigValidate(t *testing.T) {
	cases := []struct {
		t, n int
		err  error
	}{
		{0, 5, ErrHTInvalidThreshold},
		{3, 0, ErrHTInvalidTotal},
		{6, 5, ErrHTThresholdExceeds},
		{3, 5, nil},
	}
	for _, tc := range cases {
		cfg := &HybridThresholdConfig{Threshold: tc.t, Total: tc.n}
		err := cfg.Validate()
		if tc.err == nil && err != nil {
			t.Errorf("Validate(%d,%d): unexpected error %v", tc.t, tc.n, err)
		} else if tc.err != nil && !errors.Is(err, tc.err) {
			t.Errorf("Validate(%d,%d): expected %v, got %v", tc.t, tc.n, tc.err, err)
		}
	}
}

func TestHybridThresholdComputeThresholdWeight(t *testing.T) {
	cfg, _ := NewHybridThresholdConfig(2, 5)
	cfg.Weighting = SignatureWeighting{ClassicalWeight: 0.6, PQWeight: 0.4}
	agg := NewHybridThresholdAggregator(cfg)

	// Classical only: 0.6.
	agg.AddShare(&HybridThresholdShare{
		SignerIndex:  0,
		ClassicalSig: []byte("c"),
	})
	w := agg.ComputeThresholdWeight()
	if w < 0.59 || w > 0.61 {
		t.Errorf("expected weight ~0.6, got %f", w)
	}

	// Add PQ only: total = 0.6 + 0.4 = 1.0.
	agg.AddShare(&HybridThresholdShare{
		SignerIndex: 1,
		PQSig:       []byte("p"),
	})
	w = agg.ComputeThresholdWeight()
	if w < 0.99 || w > 1.01 {
		t.Errorf("expected weight ~1.0, got %f", w)
	}
}

func TestHybridThresholdDefaultWeighting(t *testing.T) {
	w := DefaultWeighting()
	if w.ClassicalWeight != 0.5 {
		t.Errorf("expected classical weight 0.5, got %f", w.ClassicalWeight)
	}
	if w.PQWeight != 0.5 {
		t.Errorf("expected PQ weight 0.5, got %f", w.PQWeight)
	}
}
