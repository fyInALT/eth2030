package txpool

import (
	"math/big"
	"strings"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// makeFrameTx builds a FrameTx with the given frames for testing.
func makeFrameTx(frames []types.Frame) *types.FrameTx {
	return &types.FrameTx{
		ChainID:      big.NewInt(1),
		Nonce:        big.NewInt(0),
		Sender:       types.BytesToAddress([]byte{0x01}),
		Frames:       frames,
		MaxFeePerGas: big.NewInt(1e9),
	}
}

// makeAddr returns a deterministic address from a single byte.
func makeAddr(b byte) types.Address {
	return types.BytesToAddress([]byte{b})
}

// --- Conservative frame rules tests ---

func TestConservativeFrameRules_Valid(t *testing.T) {
	tx := makeFrameTx([]types.Frame{
		{Mode: types.ModeVerify, GasLimit: 50000},
	})
	if err := ValidateFrameTxConservative(tx); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestConservativeFrameRules_FirstFrameNotVerify(t *testing.T) {
	tx := makeFrameTx([]types.Frame{
		{Mode: types.ModeDefault, GasLimit: 21000},
	})
	err := ValidateFrameTxConservative(tx)
	if err == nil {
		t.Fatal("expected error for non-VERIFY first frame, got nil")
	}
	if err != ErrNoVerifyFirst {
		t.Fatalf("expected ErrNoVerifyFirst, got: %v", err)
	}
}

func TestConservativeFrameRules_VerifyGasExceedsLimit(t *testing.T) {
	tx := makeFrameTx([]types.Frame{
		{Mode: types.ModeVerify, GasLimit: 50001},
	})
	err := ValidateFrameTxConservative(tx)
	if err == nil {
		t.Fatal("expected error for VERIFY gas > 50000, got nil")
	}
	if !strings.Contains(err.Error(), "gas limit") {
		t.Fatalf("expected error message to contain 'gas limit', got: %v", err)
	}
}

func TestConservativeFrameRules_VerifyGasAtLimit(t *testing.T) {
	tx := makeFrameTx([]types.Frame{
		{Mode: types.ModeVerify, GasLimit: 50000},
	})
	if err := ValidateFrameTxConservative(tx); err != nil {
		t.Fatalf("expected no error at exactly the limit, got: %v", err)
	}
}

func TestConservativeFrameRules_ZeroGasFrame(t *testing.T) {
	tx := makeFrameTx([]types.Frame{
		{Mode: types.ModeVerify, GasLimit: 0},
	})
	err := ValidateFrameTxConservative(tx)
	if err == nil {
		t.Fatal("expected error for zero gas frame, got nil")
	}
}

func TestConservativeFrameRules_NilTx(t *testing.T) {
	err := ValidateFrameTxConservative(nil)
	if err == nil {
		t.Fatal("expected error for nil tx, got nil")
	}
}

func TestConservativeFrameRules_MultipleVerifyFrames(t *testing.T) {
	tx := makeFrameTx([]types.Frame{
		{Mode: types.ModeVerify, GasLimit: 50000},
		{Mode: types.ModeVerify, GasLimit: 49000},
	})
	if err := ValidateFrameTxConservative(tx); err != nil {
		t.Fatalf("expected no error for multiple valid VERIFY frames, got: %v", err)
	}

	// Second VERIFY frame also validated — exceeding limit on second frame.
	tx2 := makeFrameTx([]types.Frame{
		{Mode: types.ModeVerify, GasLimit: 50000},
		{Mode: types.ModeVerify, GasLimit: 50001},
	})
	err := ValidateFrameTxConservative(tx2)
	if err == nil {
		t.Fatal("expected error when second VERIFY frame exceeds limit, got nil")
	}
	if !strings.Contains(err.Error(), "gas limit") {
		t.Fatalf("expected 'gas limit' in error, got: %v", err)
	}
}

func TestConservativeFrameRules_MixedFrames(t *testing.T) {
	// VERIFY(50000) + SENDER(1M) + DEFAULT(500K) — only VERIFY is capped.
	tx := makeFrameTx([]types.Frame{
		{Mode: types.ModeVerify, GasLimit: 50000},
		{Mode: types.ModeSender, GasLimit: 1_000_000},
		{Mode: types.ModeDefault, GasLimit: 500_000},
	})
	if err := ValidateFrameTxConservative(tx); err != nil {
		t.Fatalf("expected no error for mixed frames (only VERIFY capped), got: %v", err)
	}
}

// --- Aggressive frame rules tests ---

// stakedPaymasters is a simple mock PaymasterApprover for tests.
type stakedPaymasters map[types.Address]bool

func (s stakedPaymasters) IsApprovedPaymaster(addr types.Address) bool {
	return s[addr]
}

func TestAggressiveFrameRules_StakedPaymaster_Valid(t *testing.T) {
	paymasterAddr := makeAddr(0xca)
	registry := stakedPaymasters{paymasterAddr: true}

	target := paymasterAddr
	tx := makeFrameTx([]types.Frame{
		{Mode: types.ModeVerify, GasLimit: 50000, Target: &target},
	})
	if err := ValidateFrameTxAggressive(tx, registry); err != nil {
		t.Fatalf("expected no error with staked paymaster, got: %v", err)
	}
}

func TestAggressiveFrameRules_UnstakedPaymaster_Rejected(t *testing.T) {
	paymasterAddr := makeAddr(0xca)
	// Registry has a different address staked — paymasterAddr is not staked.
	otherAddr := makeAddr(0xbb)
	registry := stakedPaymasters{otherAddr: true}

	target := paymasterAddr
	tx := makeFrameTx([]types.Frame{
		{Mode: types.ModeVerify, GasLimit: 50001, Target: &target},
	})
	err := ValidateFrameTxAggressive(tx, registry)
	if err == nil {
		t.Fatal("expected error for unstaked paymaster exceeding conservative cap, got nil")
	}
}

func TestAggressiveFrameRules_NoPaymasterFrame_UsesConservative(t *testing.T) {
	// No SENDER frame, no staked paymaster → falls back to conservative check.
	registry := stakedPaymasters{}
	tx := makeFrameTx([]types.Frame{
		{Mode: types.ModeVerify, GasLimit: 50000},
	})
	if err := ValidateFrameTxAggressive(tx, registry); err != nil {
		t.Fatalf("expected no error (within conservative limit), got: %v", err)
	}

	tx2 := makeFrameTx([]types.Frame{
		{Mode: types.ModeVerify, GasLimit: 50001},
	})
	err := ValidateFrameTxAggressive(tx2, registry)
	if err == nil {
		t.Fatal("expected error: no staked paymaster, falls back to conservative cap 50000, got nil")
	}
}

func TestAggressiveFrameRules_VerifyGasExceedsConservative_WithStakedPaymaster(t *testing.T) {
	// VERIFY gas=100000 with staked paymaster → no error (aggressive allows up to 200K).
	paymasterAddr := makeAddr(0xca)
	registry := stakedPaymasters{paymasterAddr: true}

	target := paymasterAddr
	tx := makeFrameTx([]types.Frame{
		{Mode: types.ModeVerify, GasLimit: 100_000, Target: &target},
	})
	if err := ValidateFrameTxAggressive(tx, registry); err != nil {
		t.Fatalf("expected no error (aggressive cap 200K, gas=100K, staked paymaster), got: %v", err)
	}
}

func TestAggressiveFrameRules_VerifyGasExceedsConservative_NoPaymaster(t *testing.T) {
	// VERIFY gas=100000 without staked paymaster → error (conservative cap 50K applies).
	registry := stakedPaymasters{}
	tx := makeFrameTx([]types.Frame{
		{Mode: types.ModeVerify, GasLimit: 100_000},
	})
	err := ValidateFrameTxAggressive(tx, registry)
	if err == nil {
		t.Fatal("expected error: gas=100K exceeds conservative cap 50K without staked paymaster, got nil")
	}
}

// --- Metrics tests ---

func TestFrameTxMetrics_Accept(t *testing.T) {
	m := NewFrameTxMetrics()
	m.IncAccepted()
	if got := m.Accepted(); got != 1 {
		t.Fatalf("expected Accepted()=1, got %d", got)
	}
}

func TestFrameTxMetrics_RejectConservative(t *testing.T) {
	m := NewFrameTxMetrics()
	m.IncRejectedConservative()
	m.IncRejectedConservative()
	if got := m.RejectedConservative(); got != 2 {
		t.Fatalf("expected RejectedConservative()=2, got %d", got)
	}
}

func TestFrameTxMetrics_RejectAggressive(t *testing.T) {
	m := NewFrameTxMetrics()
	m.IncRejectedAggressive()
	if got := m.RejectedAggressive(); got != 1 {
		t.Fatalf("expected RejectedAggressive()=1, got %d", got)
	}
}
