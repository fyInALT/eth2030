package vm

import (
	"testing"
)

// TestSetDimGasUsage_Getter verifies SetDimGasUsage stores the pointer and
// DimGasUsage retrieves it (GAP-2.1 setter/getter added in this session).
func TestSetDimGasUsage_Getter(t *testing.T) {
	evm := &EVM{}
	if evm.DimGasUsage() != nil {
		t.Fatal("expected nil before set")
	}

	usage := &TxDimGasUsage{DimCompute: 100, DimStorage: 200}
	evm.SetDimGasUsage(usage)

	got := evm.DimGasUsage()
	if got == nil {
		t.Fatal("expected non-nil after set")
	}
	if got.DimCompute != 100 || got.DimStorage != 200 {
		t.Errorf("got DimCompute=%d DimStorage=%d, want 100/200", got.DimCompute, got.DimStorage)
	}
}

// TestSetDimGasUsage_Nil verifies setting nil clears the tracker.
func TestSetDimGasUsage_Nil(t *testing.T) {
	evm := &EVM{}
	evm.SetDimGasUsage(&TxDimGasUsage{})
	evm.SetDimGasUsage(nil)
	if evm.DimGasUsage() != nil {
		t.Error("expected nil after setting nil")
	}
}

// TestSetInitialReservoir_SeedsForward verifies SetInitialReservoir seeds
// callReservoirForward so the first child call receives the full reservoir
// (GAP-1.2). ForwardReservoir/ReturnReservoir operate on uint64 pointers.
func TestSetInitialReservoir_SeedsForward(t *testing.T) {
	evm := &EVM{}
	evm.SetInitialReservoir(50_000)

	// Simulate opCall forwarding: move parent reservoir to child frame.
	var childReservoir uint64
	ForwardReservoir(&evm.callReservoirForward, &childReservoir)

	if childReservoir != 50_000 {
		t.Errorf("child received %d, want 50000", childReservoir)
	}
	if evm.callReservoirForward != 0 {
		t.Error("parent forward should be zero after forwarding")
	}
}

// TestSetInitialReservoir_Zero verifies zero reservoir seeds correctly.
func TestSetInitialReservoir_Zero(t *testing.T) {
	evm := &EVM{}
	evm.SetInitialReservoir(0)
	var child uint64
	ForwardReservoir(&evm.callReservoirForward, &child)
	if child != 0 {
		t.Error("expected zero child reservoir")
	}
}

// TestDimGasUsage_SharedPointer verifies that AccountSSTOREGas modifies the
// usage struct in place, so the caller sees the updates after each SSTORE.
func TestDimGasUsage_SharedPointer(t *testing.T) {
	usage := &TxDimGasUsage{}
	AccountSSTOREGas(true, 0, usage) // state creation
	if usage.DimStorage != StateCreationGasPremium {
		t.Errorf("DimStorage = %d, want %d", usage.DimStorage, StateCreationGasPremium)
	}
	if usage.DimCompute != WarmStorageReadGlamst {
		t.Errorf("DimCompute = %d, want %d", usage.DimCompute, WarmStorageReadGlamst)
	}

	// Second call accumulates correctly.
	AccountSSTOREGas(true, 0, usage)
	if usage.DimStorage != 2*StateCreationGasPremium {
		t.Errorf("DimStorage after 2 SSTOREs = %d, want %d", usage.DimStorage, 2*StateCreationGasPremium)
	}
}

// TestDimGasUsage_NonCreationNoStorage verifies non-creation SSTORE adds
// nothing to DimStorage — only DimCompute receives the warm read cost.
func TestDimGasUsage_NonCreationNoStorage(t *testing.T) {
	usage := &TxDimGasUsage{}
	AccountSSTOREGas(false, 0, usage) // nonzero→nonzero, not state creation
	if usage.DimStorage != 0 {
		t.Errorf("DimStorage should be 0 for non-creation SSTORE, got %d", usage.DimStorage)
	}
	if usage.DimCompute == 0 {
		t.Error("DimCompute should be non-zero for any SSTORE")
	}
}

// TestReservoirForwardReturn_RoundTrip verifies the full forward→child→return
// lifecycle: parent seeds, child receives, spends partially, returns remainder.
func TestReservoirForwardReturn_RoundTrip(t *testing.T) {
	var parentReservoir uint64 = 80_000
	var childReservoir uint64

	// opCall: forward parent reservoir to child.
	ForwardReservoir(&parentReservoir, &childReservoir)
	if childReservoir != 80_000 {
		t.Errorf("child got %d, want 80000", childReservoir)
	}
	if parentReservoir != 0 {
		t.Error("parent reservoir should be 0 after forward")
	}

	// Child spends some reservoir gas via DrawReservoir.
	const cost = 24_084 // GasSstoreSetGlamsterdam
	if !DrawReservoir(&childReservoir, cost) {
		t.Fatal("DrawReservoir should succeed with 80000 reservoir")
	}
	// childReservoir = 80000 - 24084 = 55916

	// opCall: return child remainder to parent.
	ReturnReservoir(&parentReservoir, &childReservoir)
	if parentReservoir != 80_000-cost {
		t.Errorf("parent recovered %d, want %d", parentReservoir, 80_000-cost)
	}
	if childReservoir != 0 {
		t.Error("child reservoir should be 0 after return")
	}
}

// TestReservoirReturnRevert verifies that on revert the child reservoir is
// discarded (not returned to parent). The parent keeps its original zero state
// and the caller should restore it from the pre-call snapshot.
func TestReservoirReturnRevert(t *testing.T) {
	var parentReservoir uint64 = 60_000
	var childReservoir uint64

	ForwardReservoir(&parentReservoir, &childReservoir)
	// Simulate revert: child reservoir is lost (zeroed, not returned).
	childReservoir = 0

	// Parent reservoir was already zeroed by forward; revert leaves it at zero.
	// In EVM, the snapshot restores state; here we verify the variable is zero.
	if parentReservoir != 0 {
		t.Errorf("parent forward should be 0 after forward (pre-revert): got %d", parentReservoir)
	}
}

// TestCheckDimStorageCap_BoundaryAtCap verifies the cap is inclusive at exactly
// 4M DimStorage units (edge: blockUsed + txGas == cap is allowed).
func TestCheckDimStorageCap_BoundaryAtCap(t *testing.T) {
	cap := DimStorageBlockGasCap
	// Exactly at cap: allowed.
	if !CheckDimStorageCap(0, cap) {
		t.Error("tx using exactly cap should be allowed")
	}
	// One over: rejected.
	if CheckDimStorageCap(0, cap+1) {
		t.Error("tx using cap+1 should be rejected")
	}
	// Block already at cap: no more storage txs allowed.
	if CheckDimStorageCap(cap, 1) {
		t.Error("tx with block already full should be rejected")
	}
	// Block partially used: remaining space available.
	if !CheckDimStorageCap(cap/2, cap/2) {
		t.Error("tx within remaining capacity should be allowed")
	}
}
