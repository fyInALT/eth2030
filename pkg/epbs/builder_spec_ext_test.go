package epbs

import (
	"testing"
)

// --- Extended builder withdrawal tests ---

func TestBuilderWithdrawal_ExactBoundary(t *testing.T) {
	reg := NewBuilderWithdrawalRegistry()
	addr := [20]byte{0x01}
	reg.AddBuilder(addr, 1000)
	_ = reg.RequestWithdrawal(addr, 500, 0)

	// epoch 63 → not yet withdrawable
	if ready := reg.WithdrawableBuilders(63); len(ready) != 0 {
		t.Errorf("epoch 63: expected 0 withdrawable, got %d", len(ready))
	}
	// epoch 64 → exactly withdrawable
	ready := reg.WithdrawableBuilders(64)
	if len(ready) != 1 {
		t.Errorf("epoch 64: expected 1 withdrawable, got %d", len(ready))
	}
}

func TestBuilderWithdrawal_BuilderNotFound(t *testing.T) {
	reg := NewBuilderWithdrawalRegistry()
	addr := [20]byte{0xFF}
	err := reg.RequestWithdrawal(addr, 100, 0)
	if err == nil {
		t.Error("expected error for unregistered builder, got nil")
	}
}

func TestBuilderWithdrawal_MultipleBuilders(t *testing.T) {
	reg := NewBuilderWithdrawalRegistry()
	addrs := [][20]byte{{0x01}, {0x02}, {0x03}}
	// Register at different request epochs.
	reg.AddBuilder(addrs[0], 1000)
	reg.AddBuilder(addrs[1], 1000)
	reg.AddBuilder(addrs[2], 1000)
	_ = reg.RequestWithdrawal(addrs[0], 500, 0) // withdrawable at 64
	_ = reg.RequestWithdrawal(addrs[1], 500, 10) // withdrawable at 74
	_ = reg.RequestWithdrawal(addrs[2], 500, 20) // withdrawable at 84

	// At epoch 65: only addr[0] is ready.
	ready := reg.WithdrawableBuilders(65)
	if len(ready) != 1 || ready[0].Address != addrs[0] {
		t.Errorf("epoch 65: expected only addr[0], got %d entries", len(ready))
	}
	// At epoch 80: addr[0] and addr[1] are ready.
	ready = reg.WithdrawableBuilders(80)
	if len(ready) != 2 {
		t.Errorf("epoch 80: expected 2 withdrawable, got %d", len(ready))
	}
}

func TestBuilderWithdrawal_EmptyRegistry(t *testing.T) {
	reg := NewBuilderWithdrawalRegistry()
	ready := reg.WithdrawableBuilders(1000)
	if len(ready) != 0 {
		t.Errorf("empty registry: expected 0 withdrawable, got %d", len(ready))
	}
}

// --- Extended builder pending payments tests ---

func TestBuilderPendingPayments_SameBuilderTwice(t *testing.T) {
	state := NewBuilderEpochState()
	state.SetBuilderBalance(BuilderIndex(1), 5000)
	state.AddPendingPayment(BuilderIndex(1), 1000)
	state.AddPendingPayment(BuilderIndex(1), 2000) // two payments, total 3000

	ProcessBuilderPendingPayments(state)
	// 5000 - 1000 - 2000 = 2000
	if bal := state.GetBuilderBalance(BuilderIndex(1)); bal != 2000 {
		t.Errorf("balance after two payments: got %d, want 2000", bal)
	}
}

func TestBuilderPendingPayments_ZeroPayment(t *testing.T) {
	state := NewBuilderEpochState()
	state.SetBuilderBalance(BuilderIndex(1), 5000)
	state.AddPendingPayment(BuilderIndex(1), 0)
	ProcessBuilderPendingPayments(state)
	if bal := state.GetBuilderBalance(BuilderIndex(1)); bal != 5000 {
		t.Errorf("zero payment changed balance: got %d", bal)
	}
}

func TestBuilderPendingPayments_ExactBalance(t *testing.T) {
	state := NewBuilderEpochState()
	state.SetBuilderBalance(BuilderIndex(1), 1000)
	state.AddPendingPayment(BuilderIndex(1), 1000)
	ProcessBuilderPendingPayments(state)
	if bal := state.GetBuilderBalance(BuilderIndex(1)); bal != 0 {
		t.Errorf("exact payment: expected 0, got %d", bal)
	}
}

// --- Extended self-build tests ---

func TestSelfBuildIsUINT64Max(t *testing.T) {
	if uint64(BuilderIndexSelfBuild) != ^uint64(0) {
		t.Errorf("BuilderIndexSelfBuild is not UINT64_MAX: %d", uint64(BuilderIndexSelfBuild))
	}
}

func TestNormalBidNotSkipsAuction(t *testing.T) {
	engine := NewAuctionEngine(DefaultAuctionEngineConfig())
	bid := &BuilderBid{
		BuilderIndex: BuilderIndex(42),
		Slot:         1,
		Value:        1000,
	}
	_, skipped := engine.ProcessBidWithSelfBuild(bid)
	if skipped {
		t.Error("normal bid (non-self-build) should NOT skip auction")
	}
}
