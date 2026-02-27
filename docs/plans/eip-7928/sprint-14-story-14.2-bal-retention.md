# Story 14.2 — BAL Retention Policy (WSP = 3533 Epochs)

> **Sprint context:** Sprint 14 — Engine API Retrieval Methods & BAL Retention
> **Sprint Goal:** `engine_getPayloadBodiesByHashV2` and `engine_getPayloadBodiesByRangeV2` return `blockAccessList`; a BAL store retains data for 3533 epochs (WSP).

**Spec reference:** Line 306. "The EL MUST retain BALs for at least the duration of the weak subjectivity period (=3533 epochs)."

**Files:**
- Create: `pkg/engine/bal_store.go`
- Test: `pkg/engine/bal_store_test.go`

**Acceptance Criteria:** The BAL store prunes entries older than 3533 * 32 = 112,956 slots (~14 days); entries within the WSP are always retrievable.

#### Task 14.2.1 — Write failing tests

```go
const (
    SlotsPerEpoch   = 32
    WSPEpochs       = 3533
    WSPSlots        = WSPEpochs * SlotsPerEpoch // 112,956 slots
)

func TestBALStore_RetainsWithinWSP(t *testing.T) {
    store := NewBALStore()
    store.Store(blockHash, slotNumber, bal)
    // Advance time by WSP - 1 slots
    store.Prune(slotNumber + WSPSlots - 1)
    retrieved := store.Get(blockHash)
    if retrieved == nil {
        t.Fatal("BAL should still be retained within WSP")
    }
}

func TestBALStore_PrunesAfterWSP(t *testing.T) {
    store := NewBALStore()
    store.Store(blockHash, slotNumber, bal)
    // Advance past WSP
    store.Prune(slotNumber + WSPSlots + 1)
    retrieved := store.Get(blockHash)
    if retrieved != nil {
        t.Fatal("BAL should be pruned after WSP")
    }
}
```

#### Task 14.2.2 — Implement `BALStore`

File: `pkg/engine/bal_store.go`:

```go
package engine

import "sync"

const (
    wspSlots = 3533 * 32 // weak subjectivity period in slots
)

type balEntry struct {
    bal  []byte // RLP-encoded BAL
    slot uint64 // beacon slot when block was proposed
}

// BALStore persists BALs and prunes entries older than the WSP.
type BALStore struct {
    mu      sync.RWMutex
    byHash  map[common.Hash]*balEntry
    byNum   map[uint64]*balEntry
}

func NewBALStore() *BALStore {
    return &BALStore{
        byHash: make(map[common.Hash]*balEntry),
        byNum:  make(map[uint64]*balEntry),
    }
}

func (s *BALStore) Store(hash common.Hash, blockNum, slot uint64, bal []byte) {
    s.mu.Lock()
    defer s.mu.Unlock()
    e := &balEntry{bal: bal, slot: slot}
    s.byHash[hash] = e
    s.byNum[blockNum] = e
}

func (s *BALStore) Get(hash common.Hash) []byte {
    s.mu.RLock()
    defer s.mu.RUnlock()
    if e := s.byHash[hash]; e != nil {
        return e.bal
    }
    return nil
}

// Prune removes BAL entries older than the WSP relative to currentSlot.
func (s *BALStore) Prune(currentSlot uint64) {
    s.mu.Lock()
    defer s.mu.Unlock()
    cutoff := currentSlot - wspSlots
    for hash, e := range s.byHash {
        if e.slot < cutoff {
            delete(s.byHash, hash)
        }
    }
    for num, e := range s.byNum {
        if e.slot < cutoff {
            delete(s.byNum, num)
        }
    }
}
```

**Step: Run tests**

```
cd /projects/eth2030/pkg && go test ./engine/... -run TestBALStore -v
```

Expected: PASS.

**Step: Commit**

```bash
go fmt ./engine/...
git add pkg/engine/bal_store.go pkg/engine/bal_store_test.go
git commit -m "feat(engine): BAL retention store with WSP pruning (3533 epochs)"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
Line 304:
The `blockAccessList` field contains the RLP-encoded BAL or `null` for pre-Amsterdam blocks or when data has been pruned.

Line 306:
The EL MUST retain BALs for at least the duration of the weak subjectivity period (`=3533 epochs`) to support synchronization with re-execution after being offline for less than the WSP.

Lines 299-304:
**Retrieval methods** for historical BALs:

- `engine_getPayloadBodiesByHashV2`: Returns `ExecutionPayloadBodyV2` objects containing transactions, withdrawals, and `blockAccessList`
- `engine_getPayloadBodiesByRangeV2`: Returns `ExecutionPayloadBodyV2` objects containing transactions, withdrawals, and `blockAccessList`
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/engine/backend.go` | `EngineBackend` stores blocks in `b.blocks` map (keyed by hash); no slot-indexed retention or pruning logic exists |
| `pkg/engine/types.go` | Defines `ExecutionPayloadV5` with `BlockAccessList json.RawMessage`; no `ExecutionPayloadBodyV2` type present |
| `pkg/engine/engine_glamsterdam.go` | `HandleNewPayloadV5` validates `BlockAccessList != nil`; no retrieval-by-range or retention policy |
| `pkg/engine/engine_api_v4.go` | Prague V4 engine handler; no BAL-specific body retrieval methods |
| `pkg/bal/types.go` | `BlockAccessList` struct with `EncodeRLP()` via `hash.go`; no store or slot-based lifecycle |

---

## Implementation Assessment

### Current Status

Not implemented.

### Architecture Notes

The plan calls for a new `pkg/engine/bal_store.go` file with a `BALStore` type that indexes RLP-encoded BALs by both block hash and block number, keyed to a beacon slot number for WSP-based pruning. The codebase currently has no such file. Block storage in `pkg/engine/backend.go` uses a plain `map[types.Hash]*types.Block` with no slot tracking and no pruning. The engine type hierarchy (`ExecutionPayloadV5`) includes the `blockAccessList` field but there is no `ExecutionPayloadBodyV2` type, and neither `engine_getPayloadBodiesByHashV2` nor `engine_getPayloadBodiesByRangeV2` are implemented anywhere in `pkg/engine/`.

### Gaps and Proposed Solutions

1. **`BALStore` is entirely absent.** Create `pkg/engine/bal_store.go` exactly as specified in Task 14.2.2, with `byHash map[common.Hash]*balEntry` and `byNum map[uint64]*balEntry`, a `Store(hash, blockNum, slot, bal)` method, a `Get(hash)` method, and a `Prune(currentSlot)` method using the `wspSlots = 3533 * 32` constant.

2. **No `ExecutionPayloadBodyV2`.** Define the type containing `Transactions [][]byte`, `Withdrawals []*Withdrawal`, and `BlockAccessList json.RawMessage`, so that `engine_getPayloadBodiesByHashV2` and `engine_getPayloadBodiesByRangeV2` can return it.

3. **No WSP constant anywhere in `pkg/engine/`.** The `BALStore` file must define `SlotsPerEpoch = 32`, `WSPEpochs = 3533`, and `WSPSlots = WSPEpochs * SlotsPerEpoch` as exported constants to be testable.

4. **Backend does not call `Prune`.** Once `BALStore` is created and wired into `EngineBackend`, a periodic or per-slot `Prune` call must be added so the WSP guarantee is actually enforced at runtime.
