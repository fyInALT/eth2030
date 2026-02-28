# Story 2.1 — Build `BlockAccessList` from `BALAccessTracker.Drain()`

> **Sprint context:** Sprint 2 — BAL Assembly & Header Integration
> **Sprint Goal:** After block execution, the processor assembles a valid, sorted `BlockAccessList` from tracker events, computes its Keccak256 hash, and sets `block_access_list_hash` in the block header.

**Files:**
- Create: `pkg/bal/builder.go`
- Test: `pkg/bal/builder_test.go`

**Acceptance Criteria:** `BuildFromEvents(events map[[20]byte]*vm.AccountEvents) *BlockAccessList` produces a properly sorted BAL from raw tracker events; round-trip RLP encode → decode → re-encode produces identical bytes.

#### Task 2.1.1 — Write failing tests

File: `pkg/bal/builder_test.go`

```go
package bal_test

import (
	"testing"
	"github.com/eth2030/eth2030/bal"
	"github.com/eth2030/eth2030/core/vm"
)

func TestBuildFromEvents_SingleAddress(t *testing.T) {
	events := map[[20]byte]*vm.AccountEvents{
		{0xaa}: {
			StorageWrites: map[[32]byte][]vm.StorageEvent{
				{0x01}: {{TxIndex: 1, Value: [32]byte{0xff}}},
			},
			BalanceChange: map[uint16][32]byte{1: uint256Bytes(100)},
			NonceChange:   map[uint16]uint64{1: 5},
		},
	}

	bl := bal.BuildFromEvents(events)
	if len(bl.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(bl.Entries))
	}
}

func TestBuildFromEvents_SortedAddresses(t *testing.T) {
	events := map[[20]byte]*vm.AccountEvents{
		{0xbb}: {},
		{0xaa}: {},
	}
	bl := bal.BuildFromEvents(events)
	if bl.Entries[0].Address != ([20]byte{0xaa}) {
		t.Fatal("addresses must be sorted lexicographically")
	}
}
```

Run: `cd /projects/eth2030/pkg && go test ./bal/... -run TestBuildFromEvents -v`
Expected: FAIL.

#### Task 2.1.2 — Implement `BuildFromEvents`

File: `pkg/bal/builder.go`

```go
package bal

import (
	"sort"
	"github.com/eth2030/eth2030/core/vm"
)

// BuildFromEvents converts raw tracker events into a sorted BlockAccessList.
func BuildFromEvents(events map[[20]byte]*vm.AccountEvents) *BlockAccessList {
	bal := &BlockAccessList{}

	// Collect and sort addresses lexicographically (EIP-7928 ordering rule)
	addrs := make([][20]byte, 0, len(events))
	for addr := range events {
		addrs = append(addrs, addr)
	}
	sort.Slice(addrs, func(i, j int) bool {
		return string(addrs[i][:]) < string(addrs[j][:])
	})

	for _, addr := range addrs {
		ev := events[addr]
		entry := buildAccountEntry(addr, ev)
		bal.Entries = append(bal.Entries, entry)
	}
	return bal
}

func buildAccountEntry(addr [20]byte, ev *vm.AccountEvents) AccessEntry {
	entry := AccessEntry{Address: addr}

	// Storage writes: sorted by slot, then by txIndex ascending
	slots := make([][32]byte, 0, len(ev.StorageWrites))
	for slot := range ev.StorageWrites {
		slots = append(slots, slot)
	}
	sort.Slice(slots, func(i, j int) bool {
		return string(slots[i][:]) < string(slots[j][:])
	})
	for _, slot := range slots {
		writes := ev.StorageWrites[slot]
		sort.Slice(writes, func(i, j int) bool {
			return writes[i].TxIndex < writes[j].TxIndex
		})
		sc := StorageChange{Slot: slot}
		for _, w := range writes {
			sc.Changes = append(sc.Changes, StorageAccess{
				BlockAccessIndex: w.TxIndex,
				Value:            w.Value,
			})
		}
		entry.StorageChanges = append(entry.StorageChanges, sc)
	}

	// Storage reads: slots written are NOT also listed as reads
	writtenSlots := make(map[[32]byte]struct{})
	for slot := range ev.StorageWrites {
		writtenSlots[slot] = struct{}{}
	}
	readSlots := make([][32]byte, 0)
	for slot := range ev.StorageReads {
		if _, written := writtenSlots[slot]; !written {
			readSlots = append(readSlots, slot)
		}
	}
	sort.Slice(readSlots, func(i, j int) bool {
		return string(readSlots[i][:]) < string(readSlots[j][:])
	})
	entry.StorageReads = readSlots

	// Balance changes: sorted by txIndex
	balIdxs := sortedUint16Keys(ev.BalanceChange)
	for _, idx := range balIdxs {
		entry.BalanceChanges = append(entry.BalanceChanges, BalanceChange{
			BlockAccessIndex: idx,
			PostBalance:      ev.BalanceChange[idx],
		})
	}

	// Nonce changes
	nonceIdxs := sortedUint16Keys64(ev.NonceChange)
	for _, idx := range nonceIdxs {
		entry.NonceChanges = append(entry.NonceChanges, NonceChange{
			BlockAccessIndex: idx,
			NewNonce:         ev.NonceChange[idx],
		})
	}

	// Code changes
	codeIdxs := sortedUint16KeysBytes(ev.CodeChange)
	for _, idx := range codeIdxs {
		entry.CodeChanges = append(entry.CodeChanges, CodeChange{
			BlockAccessIndex: idx,
			Code:             ev.CodeChange[idx],
		})
	}

	return entry
}

func sortedUint16Keys(m map[uint16][32]byte) []uint16 {
	keys := make([]uint16, 0, len(m))
	for k := range m { keys = append(keys, k) }
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func sortedUint16Keys64(m map[uint16]uint64) []uint16 {
	keys := make([]uint16, 0, len(m))
	for k := range m { keys = append(keys, k) }
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func sortedUint16KeysBytes(m map[uint16][]byte) []uint16 {
	keys := make([]uint16, 0, len(m))
	for k := range m { keys = append(keys, k) }
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}
```

Run: `cd /projects/eth2030/pkg && go test ./bal/... -run TestBuildFromEvents -v`
Expected: PASS.

**Step: Format & commit**

```bash
cd /projects/eth2030/pkg && go fmt ./bal/...
git add pkg/bal/builder.go pkg/bal/builder_test.go
git commit -m "feat(bal): BuildFromEvents assembles sorted BlockAccessList"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
### Ordering and Determinism

The following ordering rules **MUST** apply:

- **Accounts**: Lexicographic by address
- **storage_changes**: Slots lexicographic by storage key; within each slot, changes by block access index (ascending)
- **storage_reads**: Lexicographic by storage key
- **balance_changes, nonce_changes, code_changes**: By block access index (ascending)

def build_bal(accesses):
    """Convert collected accesses to BAL format"""
    bal = []
    for addr in sorted(accesses.keys()):  # Sort addresses lexicographically
        data = accesses[addr]

        # Format storage changes: [slot, [[index, value], ...]]
        storage_changes = [[slot, sorted(changes)]
                          for slot, changes in sorted(data['storage_writes'].items())]

        # Account entry: [address, storage_changes, reads, balance_changes, nonce_changes, code_changes]
        bal.append([
            addr,
            storage_changes,
            sorted(list(data['storage_reads'])),
            sorted(data['balance_changes']),
            sorted(data['nonce_changes']),
            sorted(data['code_changes'])
        ])

    return bal

# AccountChanges: [address, storage_changes, storage_reads, balance_changes, nonce_changes, code_changes]
AccountChanges = [
    Address,                    # address
    List[SlotChanges],          # storage_changes (slot -> [block_access_index -> new_value])
    List[StorageKey],           # storage_reads (read-only storage keys)
    List[BalanceChange],        # balance_changes ([block_access_index -> post_balance])
    List[NonceChange],          # nonce_changes ([block_access_index -> new_nonce])
    List[CodeChange]            # code_changes ([block_access_index -> new_code])
]
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/bal/types.go` | Defines `BlockAccessList`, `AccessEntry` (with `AccessIndex uint64`), `StorageAccess`, `StorageChange`, `BalanceChange`, `NonceChange`, `CodeChange` — all single-change structs, not per-tx collections |
| `pkg/bal/tracker.go` | `AccessTracker` with `Build(txIndex uint64) *BlockAccessList`; records accesses for a single transaction and tags all entries with one `AccessIndex`; has `Reset()` for reuse |
| `pkg/bal/hash.go` | `BlockAccessList.EncodeRLP()` and `BlockAccessList.Hash()` — Keccak256 of RLP encoding |
| `pkg/core/processor.go` | `ProcessWithBAL()` calls `bal.NewTracker()` then `tracker.Build(uint64(i+1))` per transaction and merges entries into a block-level `BlockAccessList` |

---

## Implementation Assessment

### Current Status

Partially implemented.

### Architecture Notes

The plan's story calls for a standalone `BuildFromEvents(events map[[20]byte]*vm.AccountEvents) *BlockAccessList` function in a new `pkg/bal/builder.go` file. This function is intended to accept a map keyed by address whose values hold per-address, multi-transaction event collections (storage writes as `map[slot][]StorageEvent`, balance changes as `map[uint16][32]byte`, etc.).

The actual codebase has taken a different approach: `pkg/bal/tracker.go` contains `AccessTracker` with a `Build(txIndex uint64)` method that emits a `BlockAccessList` containing entries for a single transaction at a time. In `pkg/core/processor.go`, `ProcessWithBAL()` creates a fresh `AccessTracker` per transaction, calls `Build(i+1)`, and appends the resulting single-tx entries directly into the block-level `BlockAccessList`. The `BlockAccessList.Entries` field holds one `AccessEntry` per (address, txIndex) pair, not one entry per address across all transactions.

The plan's `AccessEntry` type shows slices for `BalanceChanges`, `NonceChanges`, `CodeChanges`, and `StorageChanges` that can hold multiple per-tx records for one address. The actual `AccessEntry` in `pkg/bal/types.go` uses pointer fields (`BalanceChange *BalanceChange`, `NonceChange *NonceChange`, `CodeChange *CodeChange`) — one change per entry — which means the multi-tx aggregation model assumed by the plan does not exist yet.

The `BuildFromEvents` function itself, the `pkg/bal/builder.go` file, and the `vm.AccountEvents` type referenced in the plan's test code do not exist in the codebase.

### Gaps and Proposed Solutions

1. **`vm.AccountEvents` type is missing.** The plan expects a `vm.AccountEvents` struct with fields like `StorageWrites map[[32]byte][]StorageEvent`, `BalanceChange map[uint16][32]byte`, etc. These do not exist in `pkg/core/vm`. Solution: define `AccountEvents` in `pkg/core/vm` or in `pkg/bal` as a multi-tx accumulator struct.

2. **`BuildFromEvents` function does not exist.** No `pkg/bal/builder.go` file is present. Solution: create it, implementing address-lexicographic sorting and the intra-address ordering rules (slot-lex for storage, ascending block access index for balance/nonce/code changes).

3. **`AccessEntry` does not support multi-tx aggregation.** The current `AccessEntry` holds a single `AccessIndex` and single-value pointers for each change type, which prevents representing multiple balance or nonce changes for the same address across different transactions. Solution: update `AccessEntry` to use slice fields (`[]BalanceChange`, `[]NonceChange`, `[]CodeChange`) with `BlockAccessIndex uint16` on each change struct, matching the spec's `BalanceChange = [BlockAccessIndex, Balance]` structure. This is a breaking change to the existing type.

4. **`pkg/core/processor.go` uses a per-tx tracker and flat merge rather than a block-level accumulator.** The `populateTracker`/`Build(i+1)` loop produces multiple entries for the same address if it appears in multiple transactions; the plan expects one entry per address across the entire block. Solution: after implementing `BuildFromEvents`, replace the per-tx tracker loop in `ProcessWithBAL` with a block-scoped accumulator that collects `AccountEvents` across all transactions and calls `BuildFromEvents` once after the loop.
