package vm

import "testing"

func TestDefaultReservoirConfig(t *testing.T) {
	cfg := DefaultReservoirConfig()
	if !cfg.Enabled {
		t.Error("expected Enabled to be true")
	}
	if cfg.ReservoirFraction != 0.25 {
		t.Errorf("ReservoirFraction = %f, want 0.25", cfg.ReservoirFraction)
	}
	if cfg.MinReservoir != 5000 {
		t.Errorf("MinReservoir = %d, want 5000", cfg.MinReservoir)
	}
	if cfg.MaxReservoir != 500_000 {
		t.Errorf("MaxReservoir = %d, want 500000", cfg.MaxReservoir)
	}
}

func TestInitReservoir_Enabled(t *testing.T) {
	cfg := DefaultReservoirConfig()
	// 100,000 intrinsic gas -> 25% = 25,000 reservoir, 75,000 exec.
	exec, res := InitReservoir(100_000, cfg)
	if res != 25_000 {
		t.Errorf("reservoir = %d, want 25000", res)
	}
	if exec != 75_000 {
		t.Errorf("execGas = %d, want 75000", exec)
	}
	if exec+res != 100_000 {
		t.Error("exec + reservoir should equal intrinsic gas")
	}
}

func TestInitReservoir_Disabled(t *testing.T) {
	cfg := DefaultReservoirConfig()
	cfg.Enabled = false
	exec, res := InitReservoir(100_000, cfg)
	if res != 0 {
		t.Errorf("reservoir = %d, want 0 when disabled", res)
	}
	if exec != 100_000 {
		t.Errorf("execGas = %d, want 100000 when disabled", exec)
	}
}

func TestInitReservoir_NilConfig(t *testing.T) {
	exec, res := InitReservoir(100_000, nil)
	if res != 0 {
		t.Errorf("reservoir = %d, want 0 with nil config", res)
	}
	if exec != 100_000 {
		t.Errorf("execGas = %d, want 100000 with nil config", exec)
	}
}

func TestInitReservoir_ClampMin(t *testing.T) {
	cfg := DefaultReservoirConfig()
	// 10,000 * 0.25 = 2,500, but min is 5,000.
	exec, res := InitReservoir(10_000, cfg)
	if res != 5_000 {
		t.Errorf("reservoir = %d, want 5000 (clamped to min)", res)
	}
	if exec != 5_000 {
		t.Errorf("execGas = %d, want 5000", exec)
	}
}

func TestInitReservoir_ClampMax(t *testing.T) {
	cfg := DefaultReservoirConfig()
	// 10,000,000 * 0.25 = 2,500,000, but max is 500,000.
	exec, res := InitReservoir(10_000_000, cfg)
	if res != 500_000 {
		t.Errorf("reservoir = %d, want 500000 (clamped to max)", res)
	}
	if exec != 9_500_000 {
		t.Errorf("execGas = %d, want 9500000", exec)
	}
}

func TestInitReservoir_ZeroGas(t *testing.T) {
	cfg := DefaultReservoirConfig()
	// 0 * 0.25 = 0, but min is 5000, which exceeds intrinsicGas (0).
	// Reservoir clamped to intrinsicGas.
	exec, res := InitReservoir(0, cfg)
	if res != 0 {
		t.Errorf("reservoir = %d, want 0 (clamped to intrinsicGas)", res)
	}
	if exec != 0 {
		t.Errorf("execGas = %d, want 0", exec)
	}
}

func TestDrawReservoir_Success(t *testing.T) {
	var res uint64 = 30_000
	ok := DrawReservoir(&res, 10_000)
	if !ok {
		t.Error("expected draw to succeed")
	}
	if res != 20_000 {
		t.Errorf("reservoir = %d, want 20000 after draw", res)
	}
}

func TestDrawReservoir_Insufficient(t *testing.T) {
	var res uint64 = 5_000
	ok := DrawReservoir(&res, 10_000)
	if ok {
		t.Error("expected draw to fail with insufficient reservoir")
	}
	if res != 5_000 {
		t.Errorf("reservoir = %d, want 5000 (unchanged)", res)
	}
}

func TestDrawReservoir_Exact(t *testing.T) {
	var res uint64 = 10_000
	ok := DrawReservoir(&res, 10_000)
	if !ok {
		t.Error("expected draw to succeed for exact amount")
	}
	if res != 0 {
		t.Errorf("reservoir = %d, want 0 after exact draw", res)
	}
}

func TestDrawReservoir_Zero(t *testing.T) {
	var res uint64 = 5_000
	ok := DrawReservoir(&res, 0)
	if !ok {
		t.Error("expected draw of zero to succeed")
	}
	if res != 5_000 {
		t.Errorf("reservoir = %d, want 5000 (unchanged after zero draw)", res)
	}
}

func TestForwardReservoir(t *testing.T) {
	var parent uint64 = 25_000
	var child uint64 = 0
	ForwardReservoir(&parent, &child)
	if parent != 0 {
		t.Errorf("parent = %d, want 0 after forward", parent)
	}
	if child != 25_000 {
		t.Errorf("child = %d, want 25000 after forward", child)
	}
}

func TestReturnReservoir(t *testing.T) {
	var parent uint64 = 0
	var child uint64 = 15_000
	ReturnReservoir(&parent, &child)
	if parent != 15_000 {
		t.Errorf("parent = %d, want 15000 after return", parent)
	}
	if child != 0 {
		t.Errorf("child = %d, want 0 after return", child)
	}
}

func TestForwardReturnRoundTrip(t *testing.T) {
	var parent uint64 = 50_000
	var child uint64 = 0
	original := parent

	ForwardReservoir(&parent, &child)
	// Child uses some gas.
	DrawReservoir(&child, 10_000)
	ReturnReservoir(&parent, &child)

	if parent != original-10_000 {
		t.Errorf("parent = %d, want %d after round-trip with 10000 spent", parent, original-10_000)
	}
	if child != 0 {
		t.Errorf("child = %d, want 0 after return", child)
	}
}

func TestReservoirGasCost_ZeroToNonzero(t *testing.T) {
	var original, current [32]byte // both zero
	var newVal [32]byte
	newVal[31] = 1 // non-zero

	var res uint64 = 50_000
	gas, drew := ReservoirGasCost(original, current, newVal, false, &res)

	if !drew {
		t.Error("expected drewFromReservoir to be true for zero->nonzero")
	}
	// When reservoir covers state creation, caller pays only warm read.
	if gas != WarmStorageReadGlamst {
		t.Errorf("gas = %d, want %d (WarmStorageReadGlamst)", gas, WarmStorageReadGlamst)
	}
	// Reservoir should have been reduced by GasSstoreSetGlamsterdam.
	expectedRes := uint64(50_000) - GasSstoreSetGlamsterdam
	if res != expectedRes {
		t.Errorf("reservoir = %d, want %d", res, expectedRes)
	}
}

func TestReservoirGasCost_ZeroToNonzeroCold(t *testing.T) {
	var original, current [32]byte
	var newVal [32]byte
	newVal[31] = 1

	var res uint64 = 50_000
	gas, drew := ReservoirGasCost(original, current, newVal, true, &res)

	if !drew {
		t.Error("expected drewFromReservoir for cold zero->nonzero")
	}
	// Warm read + cold penalty.
	expected := WarmStorageReadGlamst + ColdSloadGlamst
	if gas != expected {
		t.Errorf("gas = %d, want %d", gas, expected)
	}
}

func TestReservoirGasCost_NonzeroToNonzero(t *testing.T) {
	var original, current [32]byte
	original[31] = 1
	current[31] = 1 // original == current
	var newVal [32]byte
	newVal[31] = 2

	var res uint64 = 50_000
	gas, drew := ReservoirGasCost(original, current, newVal, false, &res)

	if drew {
		t.Error("expected no reservoir draw for nonzero->nonzero update")
	}
	if gas != GasSstoreReset {
		t.Errorf("gas = %d, want %d (GasSstoreReset)", gas, GasSstoreReset)
	}
	if res != 50_000 {
		t.Errorf("reservoir = %d, want 50000 (unchanged)", res)
	}
}

func TestReservoirGasCost_ClearSlot(t *testing.T) {
	var original, current [32]byte
	original[31] = 5
	current[31] = 5     // original == current
	var newVal [32]byte // zero (clearing)

	var res uint64 = 50_000
	gas, drew := ReservoirGasCost(original, current, newVal, false, &res)

	if drew {
		t.Error("expected no reservoir draw for clearing a slot")
	}
	// Non-zero to zero with original == current: GasSstoreReset.
	if gas != GasSstoreReset {
		t.Errorf("gas = %d, want %d (GasSstoreReset for clear)", gas, GasSstoreReset)
	}
	if res != 50_000 {
		t.Errorf("reservoir = %d, want 50000 (unchanged)", res)
	}
}

func TestReservoirGasCost_InsufficientReservoir(t *testing.T) {
	var original, current [32]byte // both zero
	var newVal [32]byte
	newVal[31] = 1

	// Reservoir too small to cover GasSstoreSetGlamsterdam (24084).
	var res uint64 = 10_000
	gas, drew := ReservoirGasCost(original, current, newVal, false, &res)

	if drew {
		t.Error("expected drewFromReservoir to be false when reservoir insufficient")
	}
	// Falls back to normal Glamsterdam set cost.
	if gas != GasSstoreSetGlamsterdam {
		t.Errorf("gas = %d, want %d (GasSstoreSetGlamsterdam fallback)", gas, GasSstoreSetGlamsterdam)
	}
	if res != 10_000 {
		t.Errorf("reservoir = %d, want 10000 (unchanged after failed draw)", res)
	}
}

func TestReservoirGasCost_Noop(t *testing.T) {
	var original, current, newVal [32]byte
	original[31] = 3
	current[31] = 3
	newVal[31] = 3 // current == newVal, no-op

	var res uint64 = 50_000
	gas, drew := ReservoirGasCost(original, current, newVal, false, &res)

	if drew {
		t.Error("expected no reservoir draw for no-op")
	}
	if gas != WarmStorageReadGlamst {
		t.Errorf("gas = %d, want %d (WarmStorageReadGlamst for no-op)", gas, WarmStorageReadGlamst)
	}
}

func TestReservoirGasCost_DirtySlot(t *testing.T) {
	var original, current, newVal [32]byte
	original[31] = 1
	current[31] = 2 // original != current (dirty)
	newVal[31] = 3

	var res uint64 = 50_000
	gas, drew := ReservoirGasCost(original, current, newVal, false, &res)

	if drew {
		t.Error("expected no reservoir draw for dirty slot")
	}
	if gas != WarmStorageReadGlamst {
		t.Errorf("gas = %d, want %d (WarmStorageReadGlamst for dirty slot)", gas, WarmStorageReadGlamst)
	}
}

func TestReservoirGasCost_DirtyZeroToNonzero(t *testing.T) {
	// original is non-zero, current is zero (previously cleared), new is non-zero.
	// This is NOT a state creation from the original perspective.
	var original, current, newVal [32]byte
	original[31] = 5
	// current is zero (cleared in this tx)
	newVal[31] = 7

	var res uint64 = 50_000
	gas, drew := ReservoirGasCost(original, current, newVal, false, &res)

	if drew {
		t.Error("expected no reservoir draw for dirty re-creation")
	}
	// Dirty slot: original != current.
	if gas != WarmStorageReadGlamst {
		t.Errorf("gas = %d, want %d", gas, WarmStorageReadGlamst)
	}
}

func TestInitReservoir_CustomFraction(t *testing.T) {
	cfg := &ReservoirConfig{
		Enabled:           true,
		ReservoirFraction: 0.50,
		MinReservoir:      1000,
		MaxReservoir:      1_000_000,
	}
	exec, res := InitReservoir(80_000, cfg)
	if res != 40_000 {
		t.Errorf("reservoir = %d, want 40000 (50%%)", res)
	}
	if exec != 40_000 {
		t.Errorf("execGas = %d, want 40000", exec)
	}
}

func TestInitReservoir_ReservoirExceedsIntrinsic(t *testing.T) {
	// With high fraction and low intrinsic gas, min could exceed intrinsic.
	cfg := &ReservoirConfig{
		Enabled:           true,
		ReservoirFraction: 0.25,
		MinReservoir:      50_000,
		MaxReservoir:      500_000,
	}
	exec, res := InitReservoir(30_000, cfg)
	// 30000 * 0.25 = 7500, clamped to min 50000, then clamped to intrinsicGas 30000.
	if res != 30_000 {
		t.Errorf("reservoir = %d, want 30000 (clamped to intrinsicGas)", res)
	}
	if exec != 0 {
		t.Errorf("execGas = %d, want 0", exec)
	}
}
