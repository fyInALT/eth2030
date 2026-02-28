# Story 4.1 — Engine API: BAL validation (`newPayloadV5`) and retrieval (`getPayloadV6`)

> **Sprint context:** Sprint 4 — Engine API: newPayloadV5 & getPayloadV6
> **Sprint Goal:** The Engine API handlers fully validate the BAL in `engine_newPayloadV5` and return the built BAL in `engine_getPayloadV6`, matching the Amsterdam spec.

**Files:**
- Modify: `pkg/engine/engine_glamsterdam.go`
- Modify: `pkg/engine/handler.go`
- Test: `pkg/engine/newpayloadv5_test.go`
- Test: `pkg/engine/getpayloadv6_test.go`

**Acceptance Criteria:** `newPayloadV5` returns `INVALID` when the BAL hash mismatches; `getPayloadV6` returns a non-null `blockAccessList` whose hash matches `block_access_list_hash` in the header.

#### Task 4.1.1 — Write failing tests

File: `pkg/engine/newpayloadv5_test.go`

```go
package engine_test

import "testing"

func TestNewPayloadV5_RejectsMismatchedBAL(t *testing.T) {
	// Build a valid ExecutionPayloadV5
	// Corrupt the blockAccessList field
	// Call handler.handleNewPayloadV5()
	// Assert response.Status == "INVALID"
	// Assert validationError contains "BAL"
}

func TestNewPayloadV5_AcceptsCorrectBAL(t *testing.T) {
	// Build a valid ExecutionPayloadV5 with correct BAL
	// Assert response.Status == "VALID"
}
```

File: `pkg/engine/getpayloadv6_test.go`

```go
func TestGetPayloadV6_IncludesBAL(t *testing.T) {
	// Build a payload
	// Call GetPayloadV6()
	// Assert response.ExecutionPayload.BlockAccessList != nil
	// Assert keccak256(blockAccessList) == header.BlockAccessListHash
}
```

#### Task 4.1.2 — Implement BAL validation in `NewPayloadV5`

In `pkg/engine/engine_glamsterdam.go`:

```go
func (b *glamsterdamBackend) NewPayloadV5(ctx context.Context, params *ExecutionPayloadV5, ...) (*PayloadStatusV1, error) {
	var receivedBAL bal.BlockAccessList
	if err := rlp.DecodeBytes(params.BlockAccessList, &receivedBAL); err != nil {
		return &PayloadStatusV1{Status: "INVALID",
			ValidationError: strPtr("BAL decode error: " + err.Error())}, nil
	}

	result, err := b.processor.ProcessWithBAL(ctx, params)
	if err != nil {
		return &PayloadStatusV1{Status: "INVALID", ValidationError: strPtr(err.Error())}, nil
	}

	receivedHash := receivedBAL.Hash()
	computedHash := result.BlockAccessList.Hash()
	if receivedHash != computedHash {
		return &PayloadStatusV1{
			Status:          "INVALID",
			ValidationError: strPtr(fmt.Sprintf("BAL mismatch: got %x want %x", receivedHash, computedHash)),
		}, nil
	}

	return &PayloadStatusV1{Status: "VALID", LatestValidHash: &result.BlockHash}, nil
}
```

#### Task 4.1.3 — Implement BAL retrieval in `GetPayloadV6`

```go
func (b *glamsterdamBackend) GetPayloadV6(ctx context.Context, payloadID PayloadID) (*GetPayloadV6Response, error) {
	payload, err := b.blockBuilder.GetPayload(ctx, payloadID)
	if err != nil {
		return nil, err
	}
	balBytes, err := rlp.EncodeToBytes(payload.BlockAccessList)
	if err != nil {
		return nil, fmt.Errorf("encoding BAL: %w", err)
	}
	return &GetPayloadV6Response{
		ExecutionPayload: &ExecutionPayloadV5{
			// ... existing fields ...
			BlockAccessList: balBytes,
		},
		BlockValue: payload.BlockValue,
	}, nil
}
```

**Step: Run tests**

```
cd /projects/eth2030/pkg && go test ./engine/... -run "TestNewPayloadV5|TestGetPayloadV6" -v
```

Expected: PASS.

**Step: Format & commit**

```bash
cd /projects/eth2030/pkg && go fmt ./engine/...
git add pkg/engine/engine_glamsterdam.go pkg/engine/handler.go \
        pkg/engine/newpayloadv5_test.go pkg/engine/getpayloadv6_test.go
git commit -m "feat(engine): newPayloadV5 BAL validation + getPayloadV6 retrieval"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
### Engine API

The Engine API is extended with new structures and methods to support block-level access lists:

**ExecutionPayloadV4** extends ExecutionPayloadV3 with:

- `blockAccessList`: RLP-encoded block access list

**engine_newPayloadV5** validates execution payloads:

- Accepts `ExecutionPayloadV4` structure
- Validates that computed access list matches provided `blockAccessList`
- Returns `INVALID` if access list is malformed or doesn't match

**engine_getPayloadV6** builds execution payloads:

- Returns `ExecutionPayloadV4` structure
- Collects all account accesses and state changes during transaction execution
- Populates `blockAccessList` field with RLP-encoded access list

**Block processing flow:**

When processing a block:

1. The EL receives the BAL in the ExecutionPayload
2. The EL computes `block_access_list_hash = keccak256(blockAccessList)` and includes it in the block header
3. The EL executes the block and generates the actual BAL
4. If the generated BAL doesn't match the provided BAL, the block is invalid
   (the hash in the header would be wrong)

The EL MUST retain BALs for at least the duration of the weak subjectivity period
(`=3533 epochs`) to support synchronization.
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/engine/engine_glamsterdam.go` | `GlamsterdamBackend` interface; `HandleNewPayloadV5()` checks `payload.BlockAccessList != nil` then delegates to `backend.NewPayloadV5()`; exposes `HandleGetPayloadV5()` (V5, not V6) |
| `pkg/engine/server.go` | `EngineAPI.NewPayloadV5()` validates Amsterdam fork gate, blob hashes, and non-nil BAL, then calls `backend.ProcessBlockV5()`; `EngineAPI.GetPayloadV6()` delegates to `backend.GetPayloadV6ByID()` with Amsterdam fork check |
| `pkg/engine/backend.go` | `ProcessBlockV5()` processes the block via `ProcessWithBAL()` and compares the computed BAL against the provided one; `GetPayloadV6ByID()` calls `blockToPayloadV5()` with `pending.bal`; `ForkchoiceUpdatedV4()` computes BAL during block building |
| `pkg/engine/types.go` | `ExecutionPayloadV5` (extends V4 with `BlockAccessList json.RawMessage`); `GetPayloadV6Response` (returns `*ExecutionPayloadV5`) |
| `pkg/engine/handler.go` | JSON-RPC dispatch including `engine_getPayloadV6` route |

---

## Implementation Assessment

### Current Status

Partially implemented.

### Architecture Notes

The plan expects `newPayloadV5` to:
1. RLP-decode the provided `blockAccessList`
2. Execute the block via `ProcessWithBAL()` to get the actual BAL
3. Compare `receivedBAL.Hash()` against `computedBAL.Hash()` and return `INVALID` on mismatch

The plan expects `getPayloadV6` to:
1. Retrieve the built payload
2. RLP-encode the `BlockAccessList` from the build result
3. Return it inside `ExecutionPayloadV5.BlockAccessList`

The actual `engine_newPayloadV5` path (`server.go` → `backend.ProcessBlockV5()`) is correctly wired. `ProcessBlockV5()` at `backend.go` line 383 calls `b.processor.ProcessWithBAL(block, stateCopy)` to compute the actual BAL from execution. It then compares the computed BAL encoding against the provided `blockAccessList` bytes.

The `engine_getPayloadV6` path exists at the handler and server level. `GetPayloadV6ByID()` in `backend.go` at line 553 calls `blockToPayloadV5(pending.block, ..., pending.bal)`, passing the BAL computed during `ForkchoiceUpdatedV4`. The `pendingPayload` struct has a `bal *bal.BlockAccessList` field (line 21) that is populated during block building.

The `engine_glamsterdam.go` file does not expose `getPayloadV6` — only `getPayloadV5`. The V6 route lives in `server.go`/`handler.go` as part of the main `EngineAPI`, separate from `EngineGlamsterdam`.

### Gaps and Proposed Solutions

1. **`ProcessBlockV5` correctly uses `ProcessWithBAL()`.** At `backend.go` line 383, `b.processor.ProcessWithBAL(block, stateCopy)` is called. The `result.BlockAccessList` is used for comparison against the provided BAL bytes. This gap is resolved.

2. **BAL comparison in `ProcessBlockV5` uses the actual execution result.** The comparison at lines 394-416 encodes the computed BAL via `computedBAL.EncodeRLP()` and compares against the provided bytes. A fallback to `bal.NewBlockAccessList()` handles the nil case. This gap is resolved.

3. **`GetPayloadV6ByID` returns the built BAL.** The `pendingPayload` struct has a `bal *bal.BlockAccessList` field (line 21), populated during `ForkchoiceUpdatedV4` (lines 493-500) via `b.processor.ProcessWithBAL()`. `GetPayloadV6ByID` passes `pending.bal` to `blockToPayloadV5` at line 553. This gap is resolved.

4. **`engine_newPayloadV5` BAL validation test files are absent.** The plan lists `pkg/engine/newpayloadv5_test.go` and `pkg/engine/getpayloadv6_test.go`. The existing BAL payload tests are in `pkg/engine/payload_test.go` (`TestNewPayloadV5_ValidBAL`, `TestNewPayloadV5_InvalidBAL`, `TestGetPayloadV6_Success`), but `TestNewPayloadV5_ValidBAL` passes a non-nil BAL while the backend still compares against an empty computed BAL, so the test only works for the empty-block case. Solution: once the `ProcessWithBAL` integration is fixed, update the tests to cover non-empty blocks.

5. **`GlamsterdamBackend` / `EngineGlamsterdam` does not expose `getPayloadV6`.** The Glamsterdam handler only exposes V5 get-payload. Solution: if the intent is to keep the Glamsterdam handler as the primary Amsterdam entry point, add `GetPayloadV6` to `GlamsterdamBackend` and wire it in `engine_glamsterdam.go`; otherwise, the current split where `server.go`/`handler.go` handles V6 is acceptable but should be documented.
