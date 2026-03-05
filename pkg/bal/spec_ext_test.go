package bal

import (
	"bytes"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// --- Extended BAL ordering tests ---

func TestBALOrdering_Empty(t *testing.T) {
	bal := &BlockAccessList{}
	if err := ValidateBALOrdering(bal); err != nil {
		t.Errorf("empty BAL should be valid, got: %v", err)
	}
}

func TestBALOrdering_Single(t *testing.T) {
	bal := &BlockAccessList{
		Entries: []AccessEntry{{Address: types.Address{0x42}, AccessIndex: 1}},
	}
	if err := ValidateBALOrdering(bal); err != nil {
		t.Errorf("single entry should be valid, got: %v", err)
	}
}

func TestBALOrdering_SameAddrDifferentIndex(t *testing.T) {
	// Same address, ascending access index → valid (pre then tx1).
	bal := &BlockAccessList{
		Entries: []AccessEntry{
			{Address: types.Address{0x01}, AccessIndex: 0},
			{Address: types.Address{0x01}, AccessIndex: 1},
		},
	}
	if err := ValidateBALOrdering(bal); err != nil {
		t.Errorf("same addr ascending index should be valid, got: %v", err)
	}
}

func TestBALOrdering_SameAddrSameIndex(t *testing.T) {
	// Same addr + same index → invalid.
	bal := &BlockAccessList{
		Entries: []AccessEntry{
			{Address: types.Address{0x01}, AccessIndex: 1},
			{Address: types.Address{0x01}, AccessIndex: 1},
		},
	}
	if err := ValidateBALOrdering(bal); err == nil {
		t.Error("expected error for same (addr, index) duplicate")
	}
}

func TestBALOrdering_LexicographicByAllBytes(t *testing.T) {
	// {0x00, 0xFF} < {0x01, 0x00} lexicographically.
	bal := &BlockAccessList{
		Entries: []AccessEntry{
			{Address: types.Address{0x00, 0xFF}, AccessIndex: 1},
			{Address: types.Address{0x01, 0x00}, AccessIndex: 1},
		},
	}
	if err := ValidateBALOrdering(bal); err != nil {
		t.Errorf("lex-correct ordering should be valid, got: %v", err)
	}
}

// --- Extended BlockBALTracker tests ---

func TestBlockBALTracker_ZeroGasLimit(t *testing.T) {
	// gasLimit=0 → maxItems=0; any access should fail immediately.
	tracker := NewBlockBALTracker(0)
	err := tracker.RecordAccess(types.Address{0x01})
	if err != ErrBALSizeExceeded {
		t.Errorf("zero gas limit: expected ErrBALSizeExceeded, got %v", err)
	}
}

func TestBlockBALTracker_SameAddrMultiplePhases(t *testing.T) {
	// Same address in pre + tx + post → 3 separate entries with different AccessIndex.
	tracker := NewBlockBALTracker(30_000_000)

	tracker.BeginPreExecution()
	_ = tracker.RecordAccess(types.Address{0xAA})

	tracker.BeginTx(1)
	_ = tracker.RecordAccess(types.Address{0xAA})

	tracker.BeginPostExecution(1)
	_ = tracker.RecordAccess(types.Address{0xAA})

	entries := tracker.Build()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	indices := map[uint64]bool{}
	for _, e := range entries {
		indices[e.AccessIndex] = true
	}
	for _, expected := range []uint64{0, 1, 2} {
		if !indices[expected] {
			t.Errorf("missing AccessIndex %d", expected)
		}
	}
}

func TestBlockBALTracker_IdempotentRecordSamePhase(t *testing.T) {
	// Recording the same address twice in the same phase is a no-op (deduplicated).
	tracker := NewBlockBALTracker(30_000_000)
	tracker.BeginTx(1)
	_ = tracker.RecordAccess(types.Address{0x01})
	_ = tracker.RecordAccess(types.Address{0x01}) // duplicate

	entries := tracker.Build()
	if len(entries) != 1 {
		t.Errorf("duplicate in same phase: expected 1 entry, got %d", len(entries))
	}
	if tracker.ItemCount() != 1 {
		t.Errorf("item count: got %d, want 1", tracker.ItemCount())
	}
}

func TestBlockBALTracker_MultiTxIndexing(t *testing.T) {
	tracker := NewBlockBALTracker(30_000_000)
	addrs := []types.Address{{0x01}, {0x02}, {0x03}}
	for i, a := range addrs {
		tracker.BeginTx(uint64(i + 1))
		_ = tracker.RecordAccess(a)
	}
	entries := tracker.Build()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	for _, e := range entries {
		addr := e.Address
		found := false
		for _, a := range addrs {
			if bytes.Equal(addr[:], a[:]) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("unexpected address in entries: %x", addr)
		}
	}
}

func TestBALItemCostIsCorrect(t *testing.T) {
	// 30M gas / 2000 per item = 15000 max items.
	const gasLimit uint64 = 30_000_000
	tracker := NewBlockBALTracker(gasLimit)
	if tracker.maxItems != gasLimit/BALItemCost {
		t.Errorf("maxItems = %d, want %d", tracker.maxItems, gasLimit/BALItemCost)
	}
}
