package epbs

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
