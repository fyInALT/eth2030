# Story 1.2 — Implement `BALAccessTracker` (real tracker)

> **Sprint context:** Sprint 1 — EVM Opcode Access Tracking
> **Sprint Goal:** Every EVM opcode that touches state emits a structured access event into a per-transaction tracker, so the BAL can be built from real execution data.

**Files:**
- Create: `pkg/core/vm/bal_access_tracker.go`
- Test: `pkg/core/vm/bal_access_tracker_test.go`

**Acceptance Criteria:** The real tracker buffers all events in memory, indexed by `txIndex`, and exposes a method to drain them into the `pkg/bal` types.

---

#### Task 1.2.1 — Write failing tests

File: `pkg/core/vm/bal_access_tracker_test.go`

```go
package vm_test

import (
	"testing"
	"github.com/eth2030/eth2030/core/vm"
)

func TestBALAccessTracker_RecordsStorageWrite(t *testing.T) {
	tr := vm.NewBALAccessTracker()
	addr := [20]byte{0xaa}
	slot := [32]byte{0x01}
	val  := [32]byte{0xff}
	tr.RecordStorageWrite(addr, slot, val, 1)

	events := tr.Drain()
	if len(events) == 0 {
		t.Fatal("expected events")
	}
	got := events[addr]
	if len(got.StorageWrites) == 0 {
		t.Fatal("expected storage write recorded")
	}
	if got.StorageWrites[slot][0].TxIndex != 1 {
		t.Fatalf("expected txIndex=1 got %d", got.StorageWrites[slot][0].TxIndex)
	}
}

func TestBALAccessTracker_DistinctTxIndices(t *testing.T) {
	tr := vm.NewBALAccessTracker()
	addr := [20]byte{0xbb}
	slot := [32]byte{0x02}
	tr.RecordStorageWrite(addr, slot, [32]byte{0x01}, 1)
	tr.RecordStorageWrite(addr, slot, [32]byte{0x02}, 2)

	events := tr.Drain()
	writes := events[addr].StorageWrites[slot]
	if len(writes) != 2 {
		t.Fatalf("expected 2 writes got %d", len(writes))
	}
}
```

Run: `cd /projects/eth2030/pkg && go test ./core/vm/... -run TestBALAccessTracker -v`
Expected: FAIL.

---

#### Task 1.2.2 — Implement `BALAccessTracker`

File: `pkg/core/vm/bal_access_tracker.go`

```go
package vm

import "sync"

// StorageEvent records a single storage write at a given transaction index.
type StorageEvent struct {
	TxIndex uint16
	Value   [32]byte
}

// AccountEvents accumulates all events for one address across the whole block.
type AccountEvents struct {
	AddressReads  map[uint16]struct{} // txIndex -> accessed
	StorageReads  map[[32]byte]map[uint16]struct{}
	StorageWrites map[[32]byte][]StorageEvent
	BalanceChange map[uint16][32]byte // txIndex -> postBalance
	NonceChange   map[uint16]uint64
	CodeChange    map[uint16][]byte
}

func newAccountEvents() *AccountEvents {
	return &AccountEvents{
		AddressReads:  make(map[uint16]struct{}),
		StorageReads:  make(map[[32]byte]map[uint16]struct{}),
		StorageWrites: make(map[[32]byte][]StorageEvent),
		BalanceChange: make(map[uint16][32]byte),
		NonceChange:   make(map[uint16]uint64),
		CodeChange:    make(map[uint16][]byte),
	}
}

// BALAccessTracker is a thread-safe AccessTracker that collects all EVM state accesses.
type BALAccessTracker struct {
	mu     sync.Mutex
	events map[[20]byte]*AccountEvents
}

// NewBALAccessTracker returns a live access tracker for EIP-7928 BAL building.
func NewBALAccessTracker() *BALAccessTracker {
	return &BALAccessTracker{events: make(map[[20]byte]*AccountEvents)}
}

func (t *BALAccessTracker) account(addr [20]byte) *AccountEvents {
	ev, ok := t.events[addr]
	if !ok {
		ev = newAccountEvents()
		t.events[addr] = ev
	}
	return ev
}

func (t *BALAccessTracker) RecordAddressAccess(addr [20]byte, txIndex uint16) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.account(addr).AddressReads[txIndex] = struct{}{}
}

func (t *BALAccessTracker) RecordStorageRead(addr [20]byte, slot [32]byte, txIndex uint16) {
	t.mu.Lock()
	defer t.mu.Unlock()
	ev := t.account(addr)
	if ev.StorageReads[slot] == nil {
		ev.StorageReads[slot] = make(map[uint16]struct{})
	}
	ev.StorageReads[slot][txIndex] = struct{}{}
}

func (t *BALAccessTracker) RecordStorageWrite(addr [20]byte, slot [32]byte, newVal [32]byte, txIndex uint16) {
	t.mu.Lock()
	defer t.mu.Unlock()
	ev := t.account(addr)
	ev.StorageWrites[slot] = append(ev.StorageWrites[slot], StorageEvent{TxIndex: txIndex, Value: newVal})
}

func (t *BALAccessTracker) RecordBalanceChange(addr [20]byte, postBalance [32]byte, txIndex uint16) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.account(addr).BalanceChange[txIndex] = postBalance
}

func (t *BALAccessTracker) RecordNonceChange(addr [20]byte, newNonce uint64, txIndex uint16) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.account(addr).NonceChange[txIndex] = newNonce
}

func (t *BALAccessTracker) RecordCodeChange(addr [20]byte, code []byte, txIndex uint16) {
	t.mu.Lock()
	defer t.mu.Unlock()
	cp := make([]byte, len(code))
	copy(cp, code)
	t.account(addr).CodeChange[txIndex] = cp
}

// Drain returns the collected events and resets the tracker.
func (t *BALAccessTracker) Drain() map[[20]byte]*AccountEvents {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := t.events
	t.events = make(map[[20]byte]*AccountEvents)
	return out
}
```

Run: `cd /projects/eth2030/pkg && go test ./core/vm/... -run TestBALAccessTracker -v`
Expected: PASS.

**Step 5: Format & commit**

```bash
cd /projects/eth2030/pkg && go fmt ./core/vm/...
git add pkg/core/vm/bal_access_tracker.go pkg/core/vm/bal_access_tracker_test.go
git commit -m "feat(bal): implement BALAccessTracker"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
### BlockAccessIndex Assignment

`BlockAccessIndex` values **MUST** be assigned as follows:

- `0` for **pre‑execution** system contract calls.
- `1 … n` for transactions (in block order).
- `n + 1` for **post‑execution** system contract calls.

### Ordering and Determinism

The following ordering rules **MUST** apply:

- **Accounts**: Lexicographic by address
- **storage_changes**: Slots lexicographic by storage key; within each slot, changes by block access index (ascending)
- **storage_reads**: Lexicographic by storage key
- **balance_changes, nonce_changes, code_changes**: By block access index (ascending)

def track_state_changes(tx, accesses, block_access_index):
    """Track all state changes from a transaction"""
    for addr in get_touched_addresses(tx):
        if addr not in accesses:
            accesses[addr] = {
                'storage_writes': {},  # slot -> [(index, value)]
                'storage_reads': set(),
                'balance_changes': [],
                'nonce_changes': [],
                'code_changes': []
            }

        # Track storage changes
        for slot, value in get_storage_writes(addr).items():
            if slot not in accesses[addr]['storage_writes']:
                accesses[addr]['storage_writes'][slot] = []
            accesses[addr]['storage_writes'][slot].append((block_access_index, value))

        # Track reads (slots accessed but not written)
        for slot in get_storage_reads(addr):
            if slot not in accesses[addr]['storage_writes']:
                accesses[addr]['storage_reads'].add(slot)

        # Track balance, nonce, code changes
        if balance_changed(addr):
            accesses[addr]['balance_changes'].append((block_access_index, get_balance(addr)))
        if nonce_changed(addr):
            accesses[addr]['nonce_changes'].append((block_access_index, get_nonce(addr)))
        if code_changed(addr):
            accesses[addr]['code_changes'].append((block_access_index, get_code(addr)))

### Block Structure Modification

The `BlockAccessList` is not included in the block body. The EL stores BALs separately
and transmits them as a field in the `ExecutionPayload` via the engine API.
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/bal/tracker.go` | Existing per-tx `AccessTracker` struct — tracks reads/changes/balance/nonce/code for one tx at a time; `Build(txIndex)` produces a single `BlockAccessList`; no multi-tx aggregation |
| `pkg/bal/types.go` | Current `AccessEntry` has `AccessIndex uint64` (single per-entry); plan wants all entries for an address across all txs under one entry with per-change `BlockAccessIndex` |
| `pkg/core/vm/access_list_tracker.go` | EIP-2929 warm/cold tracker only; not related to BAL accumulation |
| `pkg/core/vm/parallel_executor.go` | Contains a `TxIndex int` field in `TxAccessProfile` — shows pattern for per-tx indexing within the VM layer |
| `pkg/core/vm/parallel_executor_deep.go` | Also has `TxIndex int` in `TxAccessProfile` and slot access tracking per tx; architecturally similar to what `BALAccessTracker` needs but purpose-built for parallelism scheduling |
| `pkg/core/vm/interpreter.go` | EVM struct with `balTracker BALTracker` and `txIndex uint64` fields; `SetBALTracker()` / `GetBALTracker()` methods |

---

## Implementation Assessment

### Current Status

Complete. The concrete BAL tracker exists at `pkg/bal/tracker.go` as `AccessTracker`, and it satisfies the `BALTracker` interface defined in `pkg/core/vm/bal_tracker.go` via Go structural typing.

### Architecture Notes

The actual implementation uses a different architecture than the plan proposed, but achieves the same goal:

- `pkg/bal/tracker.go` (`AccessTracker`): a per-transaction tracker that records storage reads, storage changes, balance changes, nonce changes, code changes, and address touches. After each transaction, `Build(txIndex)` produces `AccessEntry` records tagged with that transaction's index, and `Reset()` clears state for the next transaction. The block processor aggregates entries across all transactions by calling `Build()` + `Reset()` in a loop.
- `pkg/core/vm/bal_tracker.go` (`BALTracker` interface): defines `RecordStorageRead`, `RecordStorageChange`, `RecordBalanceChange`, `RecordAddressTouch`. The EVM holds a `balTracker BALTracker` field and calls these methods from opcode handlers (15+ call sites in `instructions.go` and `interpreter.go`).
- The import cycle concern is avoided because the `vm` package defines the `BALTracker` interface, and the `bal` package provides the concrete implementation. The EVM never imports `bal` directly; the processor wires them together.
- Method signatures use `types.Address`, `types.Hash`, and `*big.Int` (not raw `[20]byte`/`[32]byte` arrays), consistent with the rest of the codebase.

### Gaps and Proposed Solutions

All gaps from the original plan are resolved:

| Original Gap | Resolution |
|-----|-------------------|
| No `BALAccessTracker` in `pkg/core/vm/` | `BALTracker` interface in `pkg/core/vm/bal_tracker.go`; concrete `AccessTracker` in `pkg/bal/tracker.go` |
| Existing `pkg/bal.AccessTracker` is per-tx, not multi-tx | Per-tx design is used with `Build(txIndex)` + `Reset()` loop; the processor aggregates entries per block |
| No `AccessTracker` interface yet | `BALTracker` interface exists in `pkg/core/vm/bal_tracker.go` |
| Thread-safety | Not needed in current design; per-tx tracker is single-threaded within the EVM execution loop |
| Ordering not applied by tracker | `Build()` in `pkg/bal/tracker.go` sorts addresses lexicographically and slot keys within each address |
