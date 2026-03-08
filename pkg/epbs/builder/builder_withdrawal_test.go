package builder

import (
	"testing"
)

func TestBuilderWithdrawalDelay(t *testing.T) {
	// MIN_BUILDER_WITHDRAWABILITY_DELAY = 64 epochs.
	if MinBuilderWithdrawabilityDelay != 64 {
		t.Errorf("MinBuilderWithdrawabilityDelay = %d, want 64", MinBuilderWithdrawabilityDelay)
	}

	reg := NewBuilderWithdrawalRegistry()
	const currentEpoch uint64 = 10

	// Register and request withdrawal.
	addr := [20]byte{0x01}
	reg.AddBuilder(addr, 1000)
	err := reg.RequestWithdrawal(addr, 500, currentEpoch)
	if err != nil {
		t.Fatalf("RequestWithdrawal: %v", err)
	}

	// Should not be withdrawable before 64 epochs.
	ready := reg.WithdrawableBuilders(currentEpoch + 63)
	for _, b := range ready {
		if b.Address == addr {
			t.Error("builder withdrawable before delay elapsed")
		}
	}

	// Should be withdrawable at currentEpoch + 64.
	ready = reg.WithdrawableBuilders(currentEpoch + 64)
	found := false
	for _, b := range ready {
		if b.Address == addr {
			found = true
		}
	}
	if !found {
		t.Error("builder not withdrawable after 64-epoch delay")
	}
}

func TestBuilderWithdrawalSweepLimit(t *testing.T) {
	// MAX_BUILDERS_PER_WITHDRAWALS_SWEEP = 16384.
	if MaxBuildersPerWithdrawalsSweep != 16384 {
		t.Errorf("MaxBuildersPerWithdrawalsSweep = %d, want 16384", MaxBuildersPerWithdrawalsSweep)
	}

	reg := NewBuilderWithdrawalRegistry()
	// Register 20,000 builders all requesting withdrawal at epoch 0.
	for i := 0; i < 20000; i++ {
		addr := [20]byte{byte(i >> 8), byte(i)}
		reg.AddBuilder(addr, 1000)
		_ = reg.RequestWithdrawal(addr, 1000, 0)
	}

	// At epoch MinBuilderWithdrawabilityDelay, sweep returns at most 16384.
	ready := reg.WithdrawableBuilders(MinBuilderWithdrawabilityDelay)
	if len(ready) > MaxBuildersPerWithdrawalsSweep {
		t.Errorf("sweep returned %d builders, want at most %d", len(ready), MaxBuildersPerWithdrawalsSweep)
	}
}

func TestBuilderWithdrawalPrefix(t *testing.T) {
	if BuilderWithdrawalPrefix != 0x03 {
		t.Errorf("BuilderWithdrawalPrefix = 0x%02x, want 0x03", BuilderWithdrawalPrefix)
	}
}

func TestBuilderWithdrawal_ExactBoundary(t *testing.T) {
	reg := NewBuilderWithdrawalRegistry()
	addr := [20]byte{0x01}
	reg.AddBuilder(addr, 1000)
	_ = reg.RequestWithdrawal(addr, 500, 0)

	if ready := reg.WithdrawableBuilders(63); len(ready) != 0 {
		t.Errorf("epoch 63: expected 0 withdrawable, got %d", len(ready))
	}
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
	reg.AddBuilder(addrs[0], 1000)
	reg.AddBuilder(addrs[1], 1000)
	reg.AddBuilder(addrs[2], 1000)
	_ = reg.RequestWithdrawal(addrs[0], 500, 0)  // withdrawable at 64
	_ = reg.RequestWithdrawal(addrs[1], 500, 10) // withdrawable at 74
	_ = reg.RequestWithdrawal(addrs[2], 500, 20) // withdrawable at 84

	ready := reg.WithdrawableBuilders(65)
	if len(ready) != 1 || ready[0].Address != addrs[0] {
		t.Errorf("epoch 65: expected only addr[0], got %d entries", len(ready))
	}
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
