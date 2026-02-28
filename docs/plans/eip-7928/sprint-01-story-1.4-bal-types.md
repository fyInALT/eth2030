# Story 1.4 — Define `pkg/bal` core types

> **Sprint context:** Sprint 1 — EVM Opcode Access Tracking
> **Sprint Goal:** Every EVM opcode that touches state emits a structured access event into a per-transaction tracker, so the BAL can be built from real execution data.

**Files:**
- Create: `pkg/bal/types.go`
- Test: `pkg/bal/types_test.go`

**Acceptance Criteria:** All RLP-encodable BAL types are defined in one file; a compile-only test confirms every type referenced by the builder and apply packages is present.

#### Task 1.4.1 — Write failing test

File: `pkg/bal/types_test.go`

```go
package bal_test

import (
	"testing"
	"github.com/eth2030/eth2030/bal"
)

func TestTypes_Compile(t *testing.T) {
	_ = bal.BlockAccessList{}
	_ = bal.AccessEntry{}
	_ = bal.StorageChange{}
	_ = bal.StorageAccess{}
	_ = bal.BalanceChange{}
	_ = bal.NonceChange{}
	_ = bal.CodeChange{}
}
```

Run: `cd /projects/eth2030/pkg && go test ./bal/... -run TestTypes_Compile -v`
Expected: FAIL — types undefined.

#### Task 1.4.2 — Implement `pkg/bal/types.go`

File: `pkg/bal/types.go`

```go
package bal

// BlockAccessList is the top-level EIP-7928 structure: a sorted list of per-address entries.
type BlockAccessList struct {
	Entries []AccessEntry
}

// AccessEntry holds all state access records for one address across the whole block.
type AccessEntry struct {
	Address        [20]byte
	StorageChanges []StorageChange
	StorageReads   [][32]byte
	BalanceChanges []BalanceChange
	NonceChanges   []NonceChange
	CodeChanges    []CodeChange
}

// StorageChange groups all writes to a single storage slot.
type StorageChange struct {
	Slot    [32]byte
	Changes []StorageAccess
}

// StorageAccess is one (block_access_index, new_value) pair for a storage write.
type StorageAccess struct {
	BlockAccessIndex uint16
	Value            [32]byte
}

// BalanceChange records the post-transaction balance at a given block access index.
type BalanceChange struct {
	BlockAccessIndex uint16
	PostBalance      [32]byte
}

// NonceChange records the post-transaction nonce at a given block access index.
type NonceChange struct {
	BlockAccessIndex uint16
	NewNonce         uint64
}

// CodeChange records the post-transaction bytecode at a given block access index.
type CodeChange struct {
	BlockAccessIndex uint16
	Code             []byte
}
```

Run: `cd /projects/eth2030/pkg && go test ./bal/... -run TestTypes_Compile -v`
Expected: PASS.

**Step: Format & commit**

```bash
cd /projects/eth2030/pkg && go fmt ./bal/...
git add pkg/bal/types.go pkg/bal/types_test.go
git commit -m "feat(bal): define core BAL types"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
### RLP Data Structures

BALs use RLP encoding following the pattern: `address -> field -> block_access_index -> change`.

# Type aliases for RLP encoding
Address = bytes20  # 20-byte Ethereum address
StorageKey = uint256  # Storage slot key
StorageValue = uint256  # Storage value
Bytecode = bytes  # Variable-length contract bytecode
BlockAccessIndex = uint16  # Block access index (0 for pre-execution, 1..n for transactions, n+1 for post-execution)
Balance = uint256  # Post-transaction balance in wei
Nonce = uint64  # Account nonce

# Core change structures (RLP encoded as lists)
# StorageChange: [block_access_index, new_value]
StorageChange = [BlockAccessIndex, StorageValue]

# BalanceChange: [block_access_index, post_balance]
BalanceChange = [BlockAccessIndex, Balance]

# NonceChange: [block_access_index, new_nonce]
NonceChange = [BlockAccessIndex, Nonce]

# CodeChange: [block_access_index, new_code]
CodeChange = [BlockAccessIndex, Bytecode]

# SlotChanges: [slot, [changes]]
# All changes to a single storage slot
SlotChanges = [StorageKey, List[StorageChange]]

# AccountChanges: [address, storage_changes, storage_reads, balance_changes, nonce_changes, code_changes]
# All changes for a single account, grouped by field type
AccountChanges = [
    Address,                    # address
    List[SlotChanges],          # storage_changes (slot -> [block_access_index -> new_value])
    List[StorageKey],           # storage_reads (read-only storage keys)
    List[BalanceChange],        # balance_changes ([block_access_index -> post_balance])
    List[NonceChange],          # nonce_changes ([block_access_index -> new_nonce])
    List[CodeChange]            # code_changes ([block_access_index -> new_code])
]

# BlockAccessList: List of AccountChanges
BlockAccessList = List[AccountChanges]
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/bal/types.go` | Existing BAL types — **already defines** `BlockAccessList`, `AccessEntry`, `StorageAccess`, `StorageChange`, `BalanceChange`, `NonceChange`, `CodeChange`; but with a different schema from the spec and the plan |
| `pkg/bal/types_extended.go` | Adds `DetailedAccess`, `DetailedAccessEntry`, `AccessMode` for parallelism conflict analysis; uses `AccessEntry` from `types.go` |
| `pkg/bal/tracker.go` | Uses all types from `types.go` to build a per-tx `BlockAccessList` |
| `pkg/bal/hash.go` | `EncodeRLP()` and `Hash()` on `BlockAccessList`; will need to remain consistent with any type redesign |
| `pkg/bal/scheduler.go` | Depends on `BlockAccessList` and `AccessEntry` for scheduling decisions |
| `pkg/bal/conflict_detector.go` | Reads `AccessEntry` fields (`StorageReads`, `StorageChanges`) for conflict detection |
| `pkg/bal/types_test.go` | Existing compile test that references all 7 type names — confirms types currently exist |
| `pkg/core/bal_integration_test.go` | Integration test using `BlockAccessList.Len()` and `ProcessWithBAL`; depends on type stability |

---

## Implementation Assessment

### Current Status

Complete. All BAL types exist in `pkg/bal/types.go` with a working schema used throughout the codebase. The schema differs from the plan's proposed design but is functionally correct and used by `tracker.go`, `hash.go`, `scheduler.go`, `conflict_detector.go`, and integration tests.

### Architecture Notes

The existing `pkg/bal/types.go` diverges from the plan in three critical ways:

**1. Per-transaction `AccessEntry` vs. per-address-per-block `AccessEntry`**

The spec and plan group **all transactions** for a single address under one `AccountChanges` entry, where each individual change (balance, nonce, storage write) carries its own `BlockAccessIndex`. The existing code uses a flat `AccessEntry` with a single `AccessIndex uint64` field — one entry per address per transaction. A block with 50 txs touching the same address would produce 50 separate `AccessEntry` records rather than one record with 50 indexed changes.

**2. Change type fields differ from spec**

| Field | Spec / Plan | Existing `types.go` |
|-------|-------------|---------------------|
| `StorageChange` | `{BlockAccessIndex uint16, Value [32]byte}` | `{Slot Hash, OldValue Hash, NewValue Hash}` — no index, includes old value |
| `StorageAccess` (read) | `[][32]byte` (list of slot keys only) | `{Slot Hash, Value Hash}` — includes the read value |
| `BalanceChange` | `{BlockAccessIndex uint16, PostBalance [32]byte}` | `{OldValue *big.Int, NewValue *big.Int}` — no index, uses `*big.Int`, includes old value |
| `NonceChange` | `{BlockAccessIndex uint16, NewNonce uint64}` | `{OldValue uint64, NewValue uint64}` — no index, includes old value |
| `CodeChange` | `{BlockAccessIndex uint16, Code []byte}` | `{OldCode []byte, NewCode []byte}` — no index, includes old code |

**3. `StorageChange` struct wraps slot+changes in `AccessEntry` differently**

The plan's `AccessEntry` has a `StorageChanges []StorageChange` where each `StorageChange` is `{Slot [32]byte, Changes []StorageAccess}` — i.e., slot-first with a list of indexed writes. The existing `StorageChange` is flat: `{Slot Hash, OldValue Hash, NewValue Hash}`, one record per write with no multi-write grouping per slot.

### Gaps and Proposed Solutions

| Gap | Proposed Solution |
|-----|-------------------|
| `AccessEntry` uses a single `AccessIndex` instead of per-change indices | Redesign `AccessEntry` to match the plan: remove `AccessIndex`; each change type carries its own `BlockAccessIndex uint16` |
| `BalanceChange` uses `*big.Int` with old+new values | Replace with `{BlockAccessIndex uint16, PostBalance [32]byte}` matching spec and plan; old value is not recorded in the BAL |
| `NonceChange` records old+new nonce | Replace with `{BlockAccessIndex uint16, NewNonce uint64}` |
| `CodeChange` records old+new code | Replace with `{BlockAccessIndex uint16, Code []byte}` |
| `StorageAccess` (read) records the value read | Replace with just the slot key `[32]byte`; spec `storage_reads` is a list of slot keys only |
| `StorageChange` is flat, no per-slot write grouping | Introduce inner grouping: `StorageChange{Slot [32]byte, Changes []StorageAccess}` where `StorageAccess{BlockAccessIndex uint16, Value [32]byte}` |
| Existing types use `types.Address` / `types.Hash` | The plan uses raw `[20]byte` and `[32]byte`; either is functionally equivalent but should be kept consistent across the BAL package |
| Downstream code (`tracker.go`, `hash.go`, `scheduler.go`, `conflict_detector.go`, `types_extended.go`) depends on current schema | Changing types will break all dependents — requires coordinated refactor; plan for it as part of this story with a migration pass over dependent files |
