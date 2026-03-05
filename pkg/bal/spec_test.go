package bal

import (
	"bytes"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// --- SPEC-5.1: BAL account ordering ---

func TestBALOrderingValid(t *testing.T) {
	bal := &BlockAccessList{
		Entries: []AccessEntry{
			{Address: types.Address{0x01}, AccessIndex: 1},
			{Address: types.Address{0x02}, AccessIndex: 1},
			{Address: types.Address{0x03}, AccessIndex: 1},
		},
	}
	if err := ValidateBALOrdering(bal); err != nil {
		t.Errorf("valid ordering rejected: %v", err)
	}
}

func TestBALOrderingOutOfOrder(t *testing.T) {
	bal := &BlockAccessList{
		Entries: []AccessEntry{
			{Address: types.Address{0x02}, AccessIndex: 1},
			{Address: types.Address{0x01}, AccessIndex: 1}, // out of order
		},
	}
	if err := ValidateBALOrdering(bal); err == nil {
		t.Error("expected error for out-of-order addresses")
	}
}

func TestBALOrderingDuplicate(t *testing.T) {
	// Duplicate address with same AccessIndex → invalid (strict ordering).
	bal := &BlockAccessList{
		Entries: []AccessEntry{
			{Address: types.Address{0x01}, AccessIndex: 1},
			{Address: types.Address{0x01}, AccessIndex: 1},
		},
	}
	if err := ValidateBALOrdering(bal); err == nil {
		t.Error("expected error for duplicate (addr, accessIndex)")
	}
}

// --- SPEC-5.2: ITEM_COST=2000 ---

func TestBALItemCost(t *testing.T) {
	if BALItemCost != 2000 {
		t.Errorf("BALItemCost = %d, want 2000", BALItemCost)
	}
}

func TestBALItemCostLimitExceeded(t *testing.T) {
	// Block gas limit 30M → max 15000 items.
	const gasLimit uint64 = 30_000_000
	tracker := NewBlockBALTracker(gasLimit)

	// Add items up to the limit.
	for i := 0; i < 15000; i++ {
		addr := types.Address{byte(i >> 8), byte(i)}
		tracker.RecordAccess(addr)
	}

	// Adding one more should return ErrBALSizeExceeded.
	err := tracker.RecordAccess(types.Address{0xFF, 0xFF, 0x01})
	if err != ErrBALSizeExceeded {
		t.Errorf("expected ErrBALSizeExceeded, got %v", err)
	}
}

// --- SPEC-5.3: BlockAccessIndex assignment ---

func TestBlockAccessIndexPreExecution(t *testing.T) {
	tracker := NewBlockBALTracker(30_000_000)
	tracker.BeginPreExecution()
	tracker.RecordAccess(types.Address{0x01})
	entries := tracker.Build()
	for _, e := range entries {
		if bytes.Equal(e.Address[:], []byte{0x01}) {
			if e.AccessIndex != 0 {
				t.Errorf("pre-execution entry AccessIndex = %d, want 0", e.AccessIndex)
			}
		}
	}
}

func TestBlockAccessIndexTxExecution(t *testing.T) {
	tracker := NewBlockBALTracker(30_000_000)
	tracker.BeginTx(1) // tx at index 1
	tracker.RecordAccess(types.Address{0x01})
	entries := tracker.Build()
	for _, e := range entries {
		if bytes.Equal(e.Address[:], []byte{0x01}) {
			if e.AccessIndex != 1 {
				t.Errorf("tx entry AccessIndex = %d, want 1", e.AccessIndex)
			}
		}
	}
}

func TestBlockAccessIndexPostExecution(t *testing.T) {
	tracker := NewBlockBALTracker(30_000_000)
	tracker.BeginPostExecution(3) // n+1 = 4 for 3 txs
	tracker.RecordAccess(types.Address{0x01})
	entries := tracker.Build()
	for _, e := range entries {
		if bytes.Equal(e.Address[:], []byte{0x01}) {
			if e.AccessIndex != 4 {
				t.Errorf("post-execution entry AccessIndex = %d, want 4", e.AccessIndex)
			}
		}
	}
}
