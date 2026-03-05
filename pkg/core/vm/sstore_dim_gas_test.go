package vm

import "testing"

// TestSSTOREDimensionRouting verifies that SSTORE zero→nonzero increments
// DimStorage and charges only the base cost to DimCompute (GAP-2.1).
func TestSSTOREDimensionRouting(t *testing.T) {
	var usage TxDimGasUsage

	// State creation (zero→nonzero).
	AccountSSTOREGas(true, 0, &usage)

	if usage.DimCompute != WarmStorageReadGlamst {
		t.Errorf("DimCompute = %d, want %d (WarmStorageReadGlamst)", usage.DimCompute, WarmStorageReadGlamst)
	}
	if usage.DimStorage != StateCreationGasPremium {
		t.Errorf("DimStorage = %d, want %d (StateCreationGasPremium)", usage.DimStorage, StateCreationGasPremium)
	}
}

// TestSSTOREDimensionNonCreation verifies that non-creation SSTORE ops only
// charge DimCompute; DimStorage is not incremented (GAP-2.1).
func TestSSTOREDimensionNonCreation(t *testing.T) {
	var usage TxDimGasUsage

	// Non-creation update (nonzero→nonzero).
	AccountSSTOREGas(false, 0, &usage)

	if usage.DimStorage != 0 {
		t.Errorf("DimStorage = %d, want 0 for non-creation SSTORE", usage.DimStorage)
	}
	if usage.DimCompute != WarmStorageReadGlamst {
		t.Errorf("DimCompute = %d, want %d", usage.DimCompute, WarmStorageReadGlamst)
	}
}

// TestSSTOREDimensionColdPenalty verifies that cold access penalty is routed
// to DimCompute (not DimStorage) for both creation and non-creation (GAP-2.1).
func TestSSTOREDimensionColdPenalty(t *testing.T) {
	var usageCreation TxDimGasUsage
	AccountSSTOREGas(true, ColdSloadGlamst, &usageCreation)

	expectedCompute := WarmStorageReadGlamst + ColdSloadGlamst
	if usageCreation.DimCompute != expectedCompute {
		t.Errorf("DimCompute (creation+cold) = %d, want %d", usageCreation.DimCompute, expectedCompute)
	}
	if usageCreation.DimStorage != StateCreationGasPremium {
		t.Errorf("DimStorage (creation+cold) = %d, want %d", usageCreation.DimStorage, StateCreationGasPremium)
	}

	var usageUpdate TxDimGasUsage
	AccountSSTOREGas(false, ColdSloadGlamst, &usageUpdate)

	if usageUpdate.DimStorage != 0 {
		t.Errorf("DimStorage (update+cold) = %d, want 0", usageUpdate.DimStorage)
	}
}

// TestSSTOREDimensionNilUsage verifies no panic when usage is nil (GAP-2.1).
func TestSSTOREDimensionNilUsage(t *testing.T) {
	// Must not panic.
	AccountSSTOREGas(true, 0, nil)
	AccountSSTOREGas(false, 0, nil)
}

// TestSSTOREDimensionMultiple verifies that multiple SSTORE creation ops
// accumulate DimStorage correctly (GAP-2.1).
func TestSSTOREDimensionMultiple(t *testing.T) {
	var usage TxDimGasUsage

	AccountSSTOREGas(true, 0, &usage)  // +StateCreationGasPremium
	AccountSSTOREGas(true, 0, &usage)  // +StateCreationGasPremium
	AccountSSTOREGas(false, 0, &usage) // no DimStorage

	if usage.DimStorage != 2*StateCreationGasPremium {
		t.Errorf("DimStorage = %d, want %d (2 creations)", usage.DimStorage, 2*StateCreationGasPremium)
	}
	if usage.DimCompute != 3*WarmStorageReadGlamst {
		t.Errorf("DimCompute = %d, want %d (3 warm reads)", usage.DimCompute, 3*WarmStorageReadGlamst)
	}
}

// TestDimStorageCap verifies that the 4M DimStorage block cap is enforced.
// A transaction that would push usage over the cap should be rejected (GAP-2.2).
func TestDimStorageCap(t *testing.T) {
	// Just below cap: should be allowed.
	if !CheckDimStorageCap(3_999_000, 900) {
		t.Error("expected cap check to pass when usage is below cap")
	}

	// Exactly at cap: allowed.
	if !CheckDimStorageCap(2_000_000, 2_000_000) {
		t.Error("expected cap check to pass when usage equals cap exactly")
	}

	// Over cap: should be rejected.
	if CheckDimStorageCap(3_999_000, 2_000) {
		t.Error("expected cap check to fail when usage exceeds cap")
	}

	// Full cap used: any additional storage rejected.
	if CheckDimStorageCap(DimStorageBlockGasCap, 1) {
		t.Error("expected cap check to fail when block is full")
	}

	// Empty block + full tx: allowed as long as tx does not exceed cap.
	if !CheckDimStorageCap(0, DimStorageBlockGasCap) {
		t.Error("expected cap check to pass for first tx filling the cap")
	}
}

// TestDimStorageCapConstant verifies the cap is 4M as specified (GAP-2.2).
func TestDimStorageCapConstant(t *testing.T) {
	if DimStorageBlockGasCap != 4_000_000 {
		t.Errorf("DimStorageBlockGasCap = %d, want 4000000", DimStorageBlockGasCap)
	}
}
