package txpool

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// --- DefaultBumperConfig ---

func TestDefaultBumperConfig(t *testing.T) {
	cfg := DefaultBumperConfig()
	if cfg.HistoryDepth != DefaultFeeHistoryDepth {
		t.Errorf("HistoryDepth = %d, want %d", cfg.HistoryDepth, DefaultFeeHistoryDepth)
	}
	if cfg.MinSuggestedTip == nil || cfg.MinSuggestedTip.Int64() != BumperMinSuggestedTip {
		t.Errorf("MinSuggestedTip = %v, want %d", cfg.MinSuggestedTip, BumperMinSuggestedTip)
	}
	if cfg.BaseFeeMultiplier != DefaultBaseFeeMultiplier {
		t.Errorf("BaseFeeMultiplier = %d, want %d", cfg.BaseFeeMultiplier, DefaultBaseFeeMultiplier)
	}
}

// --- NewPriceBumper ---

func TestNewPriceBumper_DefaultsApplied(t *testing.T) {
	pb := NewPriceBumper(BumperConfig{}) // all zeroes
	if pb.config.HistoryDepth != DefaultFeeHistoryDepth {
		t.Errorf("HistoryDepth = %d, want default", pb.config.HistoryDepth)
	}
	if pb.config.MinSuggestedTip == nil {
		t.Error("MinSuggestedTip should not be nil")
	}
	if pb.config.BaseFeeMultiplier != DefaultBaseFeeMultiplier {
		t.Errorf("BaseFeeMultiplier = %d, want default", pb.config.BaseFeeMultiplier)
	}
}

func TestNewPriceBumper_CustomConfig(t *testing.T) {
	cfg := BumperConfig{HistoryDepth: 5, MinSuggestedTip: big.NewInt(500), BaseFeeMultiplier: 3}
	pb := NewPriceBumper(cfg)
	if pb.config.HistoryDepth != 5 {
		t.Errorf("HistoryDepth = %d, want 5", pb.config.HistoryDepth)
	}
}

// --- RecordBlock ---

func TestPriceBumper_RecordBlock_Basic(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	pb.RecordBlock(BumperBlockFeeData{
		BaseFee:     big.NewInt(1e9),
		GasUsedRatio: 0.5,
		BlockNumber:  1,
	})
	if pb.HistoryLen() != 1 {
		t.Errorf("HistoryLen = %d, want 1", pb.HistoryLen())
	}
	if got := pb.LatestBaseFee(); got == nil || got.Cmp(big.NewInt(1e9)) != 0 {
		t.Errorf("LatestBaseFee = %v, want 1e9", got)
	}
}

func TestPriceBumper_RecordBlock_BlobBaseFee(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	pb.RecordBlock(BumperBlockFeeData{
		BaseFee:     big.NewInt(1e9),
		BlobBaseFee: big.NewInt(2e9),
		BlockNumber:  1,
	})
	if got := pb.LatestBlobBaseFee(); got == nil || got.Cmp(big.NewInt(2e9)) != 0 {
		t.Errorf("LatestBlobBaseFee = %v, want 2e9", got)
	}
}

func TestPriceBumper_RecordBlock_CircularOverwrite(t *testing.T) {
	cfg := BumperConfig{HistoryDepth: 3, MinSuggestedTip: big.NewInt(1)}
	pb := NewPriceBumper(cfg)
	for i := range 5 {
		pb.RecordBlock(BumperBlockFeeData{BaseFee: big.NewInt(int64(i + 1)), BlockNumber: uint64(i)})
	}
	if pb.HistoryLen() != 3 {
		t.Errorf("HistoryLen = %d, want 3 (capacity)", pb.HistoryLen())
	}
	// Latest base fee should be from the last RecordBlock call.
	if got := pb.LatestBaseFee(); got == nil || got.Int64() != 5 {
		t.Errorf("LatestBaseFee = %v, want 5", got)
	}
}

func TestPriceBumper_RecordBlock_NilBaseFee(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	pb.RecordBlock(BumperBlockFeeData{BaseFee: nil, BlockNumber: 1})
	if pb.LatestBaseFee() != nil {
		t.Error("LatestBaseFee should remain nil when nil base fee recorded first")
	}
}

// --- RecordBlockFromHeader ---

func TestPriceBumper_RecordBlockFromHeader(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	h := &types.Header{
		Number:   big.NewInt(1),
		BaseFee:  big.NewInt(1e9),
		GasLimit: 30_000_000,
		GasUsed:  15_000_000,
	}
	pb.RecordBlockFromHeader(h, nil)
	if pb.HistoryLen() != 1 {
		t.Errorf("HistoryLen = %d, want 1", pb.HistoryLen())
	}
}

func TestPriceBumper_RecordBlockFromHeader_WithExcessBlobGas(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	excess := uint64(0)
	h := &types.Header{
		Number:        big.NewInt(1),
		BaseFee:       big.NewInt(1e9),
		ExcessBlobGas: &excess,
	}
	pb.RecordBlockFromHeader(h, nil)
	if pb.LatestBlobBaseFee() == nil {
		t.Error("expected non-nil blob base fee when ExcessBlobGas is set")
	}
}

// --- SuggestFee ---

func TestPriceBumper_SuggestFee_NoHistory(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	s := pb.SuggestFee(TierStandard)
	// With no history, tip should be min tip.
	if s.MaxPriorityFeePerGas == nil {
		t.Error("MaxPriorityFeePerGas should not be nil")
	}
	if s.MaxPriorityFeePerGas.Cmp(big.NewInt(BumperMinSuggestedTip)) < 0 {
		t.Error("tip below minimum")
	}
}

func TestPriceBumper_SuggestFee_AllTiers(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	pb.RecordBlock(BumperBlockFeeData{
		BaseFee: big.NewInt(10e9),
		Tips:    []*big.Int{big.NewInt(1e9), big.NewInt(2e9), big.NewInt(5e9)},
	})

	urgent := pb.SuggestFee(TierUrgent)
	fast := pb.SuggestFee(TierFast)
	standard := pb.SuggestFee(TierStandard)
	slow := pb.SuggestFee(TierSlow)

	// Urgent tip >= slow tip.
	if urgent.MaxPriorityFeePerGas.Cmp(slow.MaxPriorityFeePerGas) < 0 {
		t.Errorf("urgent tip %v should be >= slow tip %v", urgent.MaxPriorityFeePerGas, slow.MaxPriorityFeePerGas)
	}
	_ = fast
	_ = standard
}

func TestPriceBumper_SuggestFee_UnknownTier(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	// Should default to standard.
	s := pb.SuggestFee("unknown")
	if s.MaxPriorityFeePerGas == nil {
		t.Error("expected non-nil suggestion")
	}
}

// --- SuggestAllTiers ---

func TestPriceBumper_SuggestAllTiers(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	pb.RecordBlock(BumperBlockFeeData{
		BaseFee: big.NewInt(5e9),
		Tips:    []*big.Int{big.NewInt(2e9)},
	})
	ts := pb.SuggestAllTiers()
	if ts.BaseFee == nil || ts.BaseFee.Cmp(big.NewInt(5e9)) != 0 {
		t.Errorf("BaseFee = %v, want 5e9", ts.BaseFee)
	}
	if ts.Urgent.MaxFeePerGas == nil || ts.Fast.MaxFeePerGas == nil ||
		ts.Standard.MaxFeePerGas == nil || ts.Slow.MaxFeePerGas == nil {
		t.Error("all tier fee caps should be non-nil")
	}
}

// --- SuggestReplacementFee ---

func TestPriceBumper_SuggestReplacementFee_DefaultBump(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	tx := makeTx(0, 1e9, 21000)
	s := pb.SuggestReplacementFee(tx, 0) // 0 → uses PriceBump (10%)
	if s.MaxFeePerGas == nil {
		t.Fatal("MaxFeePerGas nil")
	}
	// Should be at least 10% higher than original gas price.
	orig := big.NewInt(1e9)
	minExpected := new(big.Int).Mul(orig, big.NewInt(110))
	minExpected.Div(minExpected, big.NewInt(100))
	if s.MaxFeePerGas.Cmp(minExpected) < 0 {
		t.Errorf("replacement fee %v < expected minimum %v", s.MaxFeePerGas, minExpected)
	}
}

func TestPriceBumper_SuggestReplacementFee_CustomBump(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	tx := makeTx(0, 1000, 21000)
	s := pb.SuggestReplacementFee(tx, 20) // 20% bump
	expected := big.NewInt(1200) // 1000 * 120/100
	if s.MaxFeePerGas.Cmp(expected) != 0 {
		t.Errorf("replacement fee = %v, want %v", s.MaxFeePerGas, expected)
	}
}

// --- GasPriceAtPercentile ---

func TestPriceBumper_GasPriceAtPercentile_NoHistory(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	p := pb.GasPriceAtPercentile(50)
	if p.Sign() != 0 {
		t.Errorf("expected 0 with no history, got %s", p)
	}
}

func TestPriceBumper_GasPriceAtPercentile_WithData(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	pb.RecordBlock(BumperBlockFeeData{
		BaseFee: big.NewInt(10),
		Tips:    []*big.Int{big.NewInt(5), big.NewInt(10), big.NewInt(15)},
	})
	p50 := pb.GasPriceAtPercentile(50)
	if p50.Sign() <= 0 {
		t.Error("expected positive price at 50th percentile")
	}
}

func TestPriceBumper_GasPriceAtPercentile_ClampsBelowZero(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	pb.RecordBlock(BumperBlockFeeData{Tips: []*big.Int{big.NewInt(1)}})
	// Negative percentile should be clamped to 0.
	p := pb.GasPriceAtPercentile(-10)
	if p.Sign() < 0 {
		t.Error("expected non-negative result")
	}
}

func TestPriceBumper_GasPriceAtPercentile_ClampsAbove100(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	pb.RecordBlock(BumperBlockFeeData{Tips: []*big.Int{big.NewInt(1)}})
	// >100 should be clamped.
	p := pb.GasPriceAtPercentile(150)
	if p.Sign() < 0 {
		t.Error("expected non-negative result")
	}
}

// --- LatestBaseFee / LatestBlobBaseFee ---

func TestPriceBumper_LatestBaseFee_Initially_Nil(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	if pb.LatestBaseFee() != nil {
		t.Error("expected nil before any block recorded")
	}
}

func TestPriceBumper_LatestBlobBaseFee_Initially_Nil(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	if pb.LatestBlobBaseFee() != nil {
		t.Error("expected nil before any blob block recorded")
	}
}

// --- HistoryLen ---

func TestPriceBumper_HistoryLen_Empty(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	if pb.HistoryLen() != 0 {
		t.Errorf("HistoryLen = %d, want 0", pb.HistoryLen())
	}
}

func TestPriceBumper_HistoryLen_MaxCapacity(t *testing.T) {
	cfg := BumperConfig{HistoryDepth: 5, MinSuggestedTip: big.NewInt(1)}
	pb := NewPriceBumper(cfg)
	for i := range 10 {
		pb.RecordBlock(BumperBlockFeeData{BlockNumber: uint64(i)})
	}
	if pb.HistoryLen() != 5 {
		t.Errorf("HistoryLen = %d, want 5", pb.HistoryLen())
	}
}

// --- FeeHistory ---

func TestPriceBumper_FeeHistory_Empty(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	fees, ratios := pb.FeeHistory(5)
	if fees != nil || ratios != nil {
		t.Error("expected nil slices with no history")
	}
}

func TestPriceBumper_FeeHistory_Basic(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	for i := range 5 {
		pb.RecordBlock(BumperBlockFeeData{
			BaseFee:      big.NewInt(int64(i + 1) * 1e9),
			GasUsedRatio: float64(i) * 0.2,
			BlockNumber:  uint64(i),
		})
	}
	fees, ratios := pb.FeeHistory(3)
	if len(fees) != 3 || len(ratios) != 3 {
		t.Errorf("FeeHistory(3): len(fees)=%d len(ratios)=%d, want 3", len(fees), len(ratios))
	}
	// First entry is the most recent block.
	if fees[0] == nil || fees[0].Cmp(big.NewInt(5e9)) != 0 {
		t.Errorf("FeeHistory[0] = %v, want 5e9 (most recent)", fees[0])
	}
}

func TestPriceBumper_FeeHistory_LimitedByCount(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	pb.RecordBlock(BumperBlockFeeData{BaseFee: big.NewInt(1e9)})
	fees, _ := pb.FeeHistory(100) // ask for more than available
	if len(fees) != 1 {
		t.Errorf("FeeHistory len = %d, want 1", len(fees))
	}
}

func TestPriceBumper_FeeHistory_ZeroN(t *testing.T) {
	pb := NewPriceBumper(DefaultBumperConfig())
	pb.RecordBlock(BumperBlockFeeData{BaseFee: big.NewInt(1e9)})
	fees, ratios := pb.FeeHistory(0)
	if fees != nil || ratios != nil {
		t.Error("expected nil for n=0")
	}
}
