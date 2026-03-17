package engine

import (
	"testing"

	coreconfig "github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
)

// --- SPEC-4.3: IL satisfaction check in engine_newPayload ---

// TestEngineBackend_ProcessInclusionList verifies that ILs can be stored.
func TestEngineBackend_ProcessInclusionList(t *testing.T) {
	statedb := state.NewMemoryStateDB()
	genesis := makeGenesis()
	backend := NewEngineBackend(coreconfig.TestConfig, statedb, genesis)

	il := &types.InclusionList{
		Slot:           1,
		ValidatorIndex: 42,
		Transactions:   [][]byte{{0xaa, 0xbb}},
	}
	if err := backend.ProcessInclusionList(il); err != nil {
		t.Fatalf("ProcessInclusionList: %v", err)
	}
	backend.ilMu.RLock()
	count := len(backend.ils)
	backend.ilMu.RUnlock()
	if count != 1 {
		t.Errorf("expected 1 stored IL, got %d", count)
	}
}

// TestEngineBackend_ProcessInclusionList_Multiple verifies multiple ILs accumulate.
func TestEngineBackend_ProcessInclusionList_Multiple(t *testing.T) {
	statedb := state.NewMemoryStateDB()
	genesis := makeGenesis()
	backend := NewEngineBackend(coreconfig.TestConfig, statedb, genesis)

	for i := 0; i < 3; i++ {
		il := &types.InclusionList{
			Slot:           uint64(i + 1),
			ValidatorIndex: uint64(i),
		}
		if err := backend.ProcessInclusionList(il); err != nil {
			t.Fatalf("ProcessInclusionList[%d]: %v", i, err)
		}
	}
	backend.ilMu.RLock()
	count := len(backend.ils)
	backend.ilMu.RUnlock()
	if count != 3 {
		t.Errorf("expected 3 stored ILs, got %d", count)
	}
}

// TestEngineBackend_GetInclusionList_ReturnsEmpty verifies stub returns empty IL.
func TestEngineBackend_GetInclusionList_ReturnsEmpty(t *testing.T) {
	statedb := state.NewMemoryStateDB()
	genesis := makeGenesis()
	backend := NewEngineBackend(coreconfig.TestConfig, statedb, genesis)

	il := backend.GetInclusionList()
	if il == nil {
		t.Fatal("GetInclusionList returned nil")
	}
	if il.Transactions == nil {
		t.Error("Transactions should be non-nil empty slice")
	}
	if len(il.Transactions) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(il.Transactions))
	}
}

// TestEngineBackend_IlsAsFocil verifies conversion of types.InclusionList to focil format.
func TestEngineBackend_IlsAsFocil(t *testing.T) {
	statedb := state.NewMemoryStateDB()
	genesis := makeGenesis()
	backend := NewEngineBackend(coreconfig.TestConfig, statedb, genesis)

	tx1 := []byte{0x01, 0x02}
	tx2 := []byte{0x03, 0x04}
	il := &types.InclusionList{
		Slot:           5,
		ValidatorIndex: 10,
		Transactions:   [][]byte{tx1, tx2},
	}
	backend.ils = append(backend.ils, il)

	focilILs := backend.ilsAsFocil()
	if len(focilILs) != 1 {
		t.Fatalf("expected 1 focil IL, got %d", len(focilILs))
	}
	f := focilILs[0]
	if f.Slot != 5 {
		t.Errorf("Slot = %d, want 5", f.Slot)
	}
	if f.ProposerIndex != 10 {
		t.Errorf("ProposerIndex = %d, want 10", f.ProposerIndex)
	}
	if len(f.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(f.Entries))
	}
	if f.Entries[0].Index != 0 {
		t.Errorf("Entries[0].Index = %d, want 0", f.Entries[0].Index)
	}
	if f.Entries[1].Index != 1 {
		t.Errorf("Entries[1].Index = %d, want 1", f.Entries[1].Index)
	}
}

// TestStatusInclusionListUnsatisfied_Value verifies the constant value matches the spec.
func TestStatusInclusionListUnsatisfied_Value(t *testing.T) {
	const want = "INCLUSION_LIST_UNSATISFIED"
	if StatusInclusionListUnsatisfied != want {
		t.Errorf("StatusInclusionListUnsatisfied = %q, want %q", StatusInclusionListUnsatisfied, want)
	}
}

// TestEngineBackend_ImplementsInclusionListBackend verifies interface compliance.
func TestEngineBackend_ImplementsInclusionListBackend(t *testing.T) {
	statedb := state.NewMemoryStateDB()
	genesis := makeGenesis()
	backend := NewEngineBackend(coreconfig.TestConfig, statedb, genesis)

	// This is a compile-time check via interface assertion.
	var _ InclusionListBackend = backend
}
