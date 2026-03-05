package bal

import (
	"github.com/eth2030/eth2030/core/types"
)

// BlockBALTracker tracks BAL accesses for a single block, managing the
// BlockAccessIndex (0=pre-execution, 1..n=txs, n+1=post-execution) and
// enforcing the ITEM_COST=2000 size constraint per EIP-7928.
type BlockBALTracker struct {
	gasLimit    uint64
	maxItems    uint64
	itemCount   uint64
	currentIdx  uint64
	entries     []AccessEntry
	addressSeen map[types.Address]map[uint64]bool
}

// NewBlockBALTracker creates a tracker for a block with the given gas limit.
// Maximum BAL items = gasLimit / BALItemCost.
func NewBlockBALTracker(gasLimit uint64) *BlockBALTracker {
	return &BlockBALTracker{
		gasLimit:    gasLimit,
		maxItems:    gasLimit / BALItemCost,
		addressSeen: make(map[types.Address]map[uint64]bool),
	}
}

// BeginPreExecution sets the current AccessIndex to 0 (pre-execution system calls).
func (t *BlockBALTracker) BeginPreExecution() {
	t.currentIdx = 0
}

// BeginTx sets the current AccessIndex to txIndex (1-based, 1..n for txs).
func (t *BlockBALTracker) BeginTx(txIndex uint64) {
	t.currentIdx = txIndex
}

// BeginPostExecution sets the current AccessIndex to n+1 (post-execution system calls).
// n is the total number of user transactions.
func (t *BlockBALTracker) BeginPostExecution(txCount uint64) {
	t.currentIdx = txCount + 1
}

// RecordAccess records an address access at the current AccessIndex.
// Returns ErrBALSizeExceeded if the item count would exceed the limit.
func (t *BlockBALTracker) RecordAccess(addr types.Address) error {
	if _, ok := t.addressSeen[addr]; !ok {
		t.addressSeen[addr] = make(map[uint64]bool)
	}
	if t.addressSeen[addr][t.currentIdx] {
		return nil // already recorded
	}

	if t.itemCount >= t.maxItems {
		return ErrBALSizeExceeded
	}

	t.addressSeen[addr][t.currentIdx] = true
	t.itemCount++
	t.entries = append(t.entries, AccessEntry{
		Address:     addr,
		AccessIndex: t.currentIdx,
	})
	return nil
}

// Build returns the recorded access entries.
func (t *BlockBALTracker) Build() []AccessEntry {
	return t.entries
}

// ItemCount returns the current number of recorded items.
func (t *BlockBALTracker) ItemCount() uint64 {
	return t.itemCount
}
