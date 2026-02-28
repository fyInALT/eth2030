# Story 14.1 — Engine API: `getPayloadBodiesByHashV2` and `getPayloadBodiesByRangeV2`

> **Sprint context:** Sprint 14 — Engine API Retrieval Methods & BAL Retention
> **Sprint Goal:** `engine_getPayloadBodiesByHashV2` and `engine_getPayloadBodiesByRangeV2` return `blockAccessList`; a BAL store retains data for 3533 epochs (WSP).

**Spec reference:** Lines 300-304. Both methods return `ExecutionPayloadBodyV2` with `blockAccessList` field; null for pre-Amsterdam or pruned.

**Files:**
- Modify: `pkg/engine/engine_glamsterdam.go`
- Modify: `pkg/engine/handler.go`
- Modify: `pkg/engine/types.go`
- Test: `pkg/engine/payload_bodies_test.go`

**Acceptance Criteria:** By-hash returns the stored BAL or `null` for unknown/pre-Amsterdam/pruned blocks; by-range returns BAL for each block in [start, start+count).

#### Task 14.1.1 — Write failing tests

```go
func TestGetPayloadBodiesByHashV2_IncludesBAL(t *testing.T) {
    // Store a block with a known BAL
    // Call engine_getPayloadBodiesByHashV2 with that block's hash
    // Assert response[0].blockAccessList is non-null
    // Assert keccak256(blockAccessList) == header.blockAccessListHash
}

func TestGetPayloadBodiesByHashV2_PrunedData_ReturnsNull(t *testing.T) {
    // Store a block, advance the BAL store past WSP
    // Assert response[0].blockAccessList is null
}

func TestGetPayloadBodiesByRangeV2_IncludesBAL(t *testing.T) {
    // Store 3 blocks, call by-range for all 3
    // Assert each response has non-null blockAccessList
}
```

#### Task 14.1.2 — Add `ExecutionPayloadBodyV2` type

In `pkg/engine/types.go`:

```go
// ExecutionPayloadBodyV2 extends V1 with blockAccessList for EIP-7928.
type ExecutionPayloadBodyV2 struct {
    Transactions    []hexutil.Bytes  `json:"transactions"`
    Withdrawals     []*Withdrawal    `json:"withdrawals"`
    BlockAccessList json.RawMessage  `json:"blockAccessList"` // null for pre-Amsterdam
}
```

#### Task 14.1.3 — Implement `GetPayloadBodiesByHashV2`

In `pkg/engine/engine_glamsterdam.go`:

```go
func (b *glamsterdamBackend) GetPayloadBodiesByHashV2(ctx context.Context, hashes []common.Hash) ([]*ExecutionPayloadBodyV2, error) {
    results := make([]*ExecutionPayloadBodyV2, len(hashes))
    for i, hash := range hashes {
        body, bal := b.store.GetBodyAndBAL(hash)
        if body == nil {
            results[i] = nil
            continue
        }
        balBytes, _ := rlp.EncodeToBytes(bal)
        results[i] = &ExecutionPayloadBodyV2{
            Transactions:    body.Transactions,
            Withdrawals:     body.Withdrawals,
            BlockAccessList: balBytes,
        }
    }
    return results, nil
}
```

#### Task 14.1.4 — Implement `GetPayloadBodiesByRangeV2`

```go
func (b *glamsterdamBackend) GetPayloadBodiesByRangeV2(ctx context.Context, start, count uint64) ([]*ExecutionPayloadBodyV2, error) {
    results := make([]*ExecutionPayloadBodyV2, count)
    for i := uint64(0); i < count; i++ {
        body, bal := b.store.GetBodyAndBALByNumber(start + i)
        if body == nil {
            results[i] = nil
            continue
        }
        balBytes, _ := rlp.EncodeToBytes(bal)
        results[i] = &ExecutionPayloadBodyV2{
            Transactions:    body.Transactions,
            Withdrawals:     body.Withdrawals,
            BlockAccessList: balBytes,
        }
    }
    return results, nil
}
```

**Step: Run tests**

```
cd /projects/eth2030/pkg && go test ./engine/... -run "TestGetPayloadBodiesByHash|TestGetPayloadBodiesByRange" -v
```

Expected: PASS.

**Step: Commit**

```bash
go fmt ./engine/...
git add pkg/engine/engine_glamsterdam.go pkg/engine/handler.go \
        pkg/engine/types.go pkg/engine/payload_bodies_test.go
git commit -m "feat(engine): getPayloadBodiesV2 by hash and range with BAL"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
Lines 299-304:
**Retrieval methods** for historical BALs:

- `engine_getPayloadBodiesByHashV2`: Returns `ExecutionPayloadBodyV2` objects containing transactions, withdrawals, and `blockAccessList`
- `engine_getPayloadBodiesByRangeV2`: Returns `ExecutionPayloadBodyV2` objects containing transactions, withdrawals, and `blockAccessList`

The `blockAccessList` field contains the RLP-encoded BAL or `null` for pre-Amsterdam blocks or when data has been pruned.

The EL MUST retain BALs for at least the duration of the weak subjectivity period (`=3533 epochs`) to support synchronization with re-execution after being offline for less than the WSP.

Supporting context (lines 272-297):
**ExecutionPayloadV4** extends ExecutionPayloadV3 with:
- `blockAccessList`: RLP-encoded block access list

**engine_newPayloadV5** validates execution payloads:
- Accepts `ExecutionPayloadV4` structure
- Validates that computed access list matches provided `blockAccessList`

**engine_getPayloadV6** builds execution payloads:
- Returns `ExecutionPayloadV4` structure
- Populates `blockAccessList` field with RLP-encoded access list
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/engine/types.go` (lines 65-69) | `ExecutionPayloadV5` — already extends V4 with `BlockAccessList json.RawMessage`; this is the Amsterdam payload type; no `ExecutionPayloadBodyV2` type exists |
| `pkg/engine/types.go` (lines 120-128) | `GetPayloadV6Response` — returns `*ExecutionPayloadV5`; demonstrates the V6 pattern for payload retrieval |
| `pkg/engine/engine_glamsterdam.go` | `EngineGlamsterdam` struct with `HandleNewPayloadV5`, `HandleGetPayloadV5`, `HandleGetBlobsV2` — no `GetPayloadBodiesByHashV2` or `GetPayloadBodiesByRangeV2` methods |
| `pkg/engine/engine_glamsterdam.go` (line 150-153) | Validates that `payload.BlockAccessList != nil` for Amsterdam payloads — BAL presence is checked but no retrieval methods exist |
| `pkg/engine/handler.go` | Engine handler for JSON-RPC dispatch; would need new case entries for `engine_getPayloadBodiesByHashV2` and `engine_getPayloadBodiesByRangeV2` |
| `pkg/engine/backend.go` (lines 391-410) | BAL validation logic: decodes `payload.BlockAccessList`, computes expected BAL, compares — confirms BAL is handled during `newPayload` but no body retrieval path exists |

---

## Implementation Assessment

### Current Status

Not implemented.

### Architecture Notes

The codebase does not contain `GetPayloadBodiesByHashV2`, `GetPayloadBodiesByRangeV2`, or `ExecutionPayloadBodyV2` anywhere in `pkg/engine/`. There is no V1 equivalent either — neither `getPayloadBodiesByHashV1` nor `getPayloadBodiesByRangeV1` exists in the current engine handler, meaning the BAL-extended V2 methods would be the first implementation of the bodies-retrieval API family.

The `ExecutionPayloadV5` type in `pkg/engine/types.go` already carries `BlockAccessList json.RawMessage` (line 68), which is the Amsterdam-era payload type. The `GetPayloadV6Response` (lines 121-128) returns this type for `engine_getPayloadV6`. What is missing is:

1. A `ExecutionPayloadBodyV2` type (distinct from the full payload) containing only `transactions`, `withdrawals`, and `blockAccessList`.
2. A BAL store interface (`store.GetBodyAndBAL(hash)`, `store.GetBodyAndBALByNumber(number)`) backed by a persistent store that retains BALs for the WSP duration.
3. The two handler methods on `EngineGlamsterdam` (or a new type) and their JSON-RPC dispatch entries in `HandleJSONRPC`.

The story's proposed `glamsterdamBackend` interface and `GetBodyAndBAL` store method do not exist in `pkg/engine/backend.go`. The actual backend (`EngineBackend` in `backend.go`) processes payloads and validates BALs during `newPayload`, but does not expose a body-retrieval API.

The `engine_glamsterdam.go` file currently handles five methods (`newPayloadV5`, `forkchoiceUpdatedV4`, `getPayloadV5`, `getBlobsV2`, `getClientVersionV2`). Adding the two bodies-retrieval methods fits naturally into this file's pattern.

### Gaps and Proposed Solutions

1. **`ExecutionPayloadBodyV2` type is absent**: Add to `pkg/engine/types.go`. The story's definition is appropriate; `BlockAccessList json.RawMessage` should be `null`-able for pre-Amsterdam/pruned blocks.

2. **No BAL persistence or retrieval store**: The most significant gap. The engine backend validates the BAL during `newPayload` but does not store it separately for later retrieval. Need to design a `BALStore` interface with `Store(blockHash, bal []byte)` and `Get(blockHash) ([]byte, bool)` methods, backed by a key-value store with TTL or epoch-based pruning at the WSP boundary (3533 epochs ≈ 3533 * 32 * 12 seconds ≈ 45.5 days).

3. **`GetPayloadBodiesByHashV2` and `GetPayloadBodiesByRangeV2` not in `GlamsterdamBackend` interface**: The `GlamsterdamBackend` interface in `engine_glamsterdam.go` (lines 25-44) defines only the five existing methods. Both new methods need to be added to the interface and implemented by the concrete backend.

4. **JSON-RPC dispatch missing**: `HandleJSONRPC` in `engine_glamsterdam.go` (lines 247-265) dispatches on method name; cases for `engine_getPayloadBodiesByHashV2` and `engine_getPayloadBodiesByRangeV2` must be added with their RPC parsing helpers.

5. **Block body retrieval path**: To return `transactions` and `withdrawals` alongside the BAL, the implementation needs access to the block body (transactions and withdrawals by hash or number). This may require a `ChainReader` or `BlockReader` dependency in the engine backend that is not currently wired in for this purpose.
