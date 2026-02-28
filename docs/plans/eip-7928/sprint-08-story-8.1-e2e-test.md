# Story 8.1 — Comprehensive E2E test

> **Sprint context:** Sprint 8 — End-to-End & Regression
> **Sprint Goal:** Full end-to-end test suite covering the complete BAL pipeline: build block → compute BAL → hash → validate via Engine API → parallel execution matches sequential → state reconstruction from BAL.

**Files:**
- Create: `pkg/core/e2e_bal_test.go`

**Acceptance Criteria:** Single test function `TestBAL_FullPipeline` covers:
1. Build 20-tx block (mixed: ETH transfers, contract deploys, storage ops)
2. Execute via `ProcessWithBAL` → get BAL
3. BAL contains all expected addresses (sender, recipient, coinbase, contract)
4. BAL hash matches header field
5. Engine API round-trip: `getPayloadV6` → `newPayloadV5` returns VALID
6. `ProcessParallel` produces same state root as `Process`
7. `ApplyBAL` on fresh state produces same state root

#### Task 8.1.1 — Write the full pipeline test

```go
package core_test

import (
	"testing"
	"github.com/eth2030/eth2030/bal"
	"github.com/eth2030/eth2030/core"
)

func TestBAL_FullPipeline(t *testing.T) {
	chain := setupTestChain(t) // test helper

	// 1. Build a block with 20 transactions
	block := buildTestBlock(t, chain, 20)

	// 2. Process with BAL
	result, err := chain.Processor.ProcessWithBAL(block)
	if err != nil {
		t.Fatalf("ProcessWithBAL: %v", err)
	}
	if result.BlockAccessList == nil {
		t.Fatal("expected non-nil BlockAccessList")
	}

	// 3. Check address count
	if len(result.BlockAccessList.Entries) < 3 {
		t.Fatalf("expected at least 3 addresses in BAL, got %d",
			len(result.BlockAccessList.Entries))
	}

	// 4. BAL hash in header
	expectedHash := result.BlockAccessList.Hash()
	if block.Header().BlockAccessListHash == nil {
		t.Fatal("header.BlockAccessListHash not set")
	}
	if *block.Header().BlockAccessListHash != expectedHash {
		t.Fatalf("BAL hash mismatch: header=%x computed=%x",
			*block.Header().BlockAccessListHash, expectedHash)
	}

	// 5. Parallel execution matches sequential
	seqRoot := result.StateRoot
	parRoot := chain.ParallelProcessor.ProcessParallel(block, result.BlockAccessList)
	if seqRoot != parRoot {
		t.Fatalf("parallel state root mismatch: seq=%x par=%x", seqRoot, parRoot)
	}

	// 6. State reconstruction from BAL
	freshState := chain.NewEmptyState()
	chain.Processor.ProcessPreState(block, freshState) // apply genesis-level pre-state
	bal.ApplyBAL(freshState, result.BlockAccessList)
	reconRoot := freshState.IntermediateRoot(false)
	if seqRoot != reconRoot {
		t.Fatalf("state reconstruction mismatch: exec=%x recon=%x", seqRoot, reconRoot)
	}
}
```

Run:
```
cd /projects/eth2030/pkg && go test ./core/... -run TestBAL_FullPipeline -v -timeout 120s
```

Expected: PASS.

**Step: Run full test suite and race detector**

```
cd /projects/eth2030/pkg && go test ./... -count=1 -timeout 300s 2>&1 | tail -30
cd /projects/eth2030/pkg && go test -race ./bal/... ./core/... ./engine/... -timeout 120s
```

Expected: All 18,000+ tests pass; no data races reported.

**Step: Format & commit**

```bash
cd /projects/eth2030/pkg && go fmt ./...
git add pkg/core/e2e_bal_test.go
git commit -m "test(bal): full pipeline E2E test + regression check"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
### State Transition Function

The state transition function must validate that the provided BAL matches
the actual state accesses:

def validate_block(execution_payload, block_header):
    # 1. Compute hash from received BAL and set in header
    block_header.block_access_list_hash = keccak(execution_payload.blockAccessList)

    # 2. Execute block and collect actual accesses
    actual_bal = execute_and_collect_accesses(execution_payload)

    # 3. Verify actual execution matches provided BAL
    assert rlp.encode(actual_bal) == execution_payload.blockAccessList

### Block Structure Modification

The BlockAccessList is not included in the block body. The EL stores BALs
separately and transmits them as a field in the ExecutionPayload via the
engine API. The BAL is RLP-encoded as a list of AccountChanges.

### Recording Semantics

BALs include:
- Transaction senders and recipients (even for zero-value transfers)
- COINBASE address for each transaction
- All accessed storage slots (reads and writes)
- Post-transaction balances, nonces, and code for changed accounts
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/e2e_block_test.go` | Existing E2E tests (`TestE2E_BlockCreation`, `TestE2E_TransactionProcessing`, `TestE2E_StateTransition`) — does not test BAL pipeline; provides helpers (`e2eChain`, `buildAndInsert`, `signLegacyTx`) reusable in the new test |
| `pkg/core/bal_integration_test.go` | Existing BAL-specific integration tests: `TestProcessWithBAL_WithTransactions`, `TestBALHash_Computed`, `TestBlockchain_BALValidation_EndToEnd`, `TestBlockchain_RejectsWrongBALHash` — partial pipeline coverage; no parallel execution or `ApplyBAL` check |
| `pkg/core/e2e_bal_test.go` | Does not exist — the file to be created by this story |
| `pkg/bal/apply.go` | Does not exist (Sprint 6 dependency) — `ApplyBAL` is unavailable until Story 6.1 is implemented |
| `pkg/engine/engine_glamsterdam.go` | `HandleNewPayloadV5` — checks `payload.BlockAccessList != nil`; full Engine API round-trip validation is not yet wired into the test framework |

---

## Implementation Assessment

### Current Status

Partially implemented. Significant portions of the pipeline exist and are tested, but the full `TestBAL_FullPipeline` as described in the plan does not yet exist, and several of its steps depend on unimplemented components.

### Architecture Notes

The plan's `TestBAL_FullPipeline` test references several APIs and types that either do not exist yet or differ from the current codebase:

1. **`chain.Processor.ProcessWithBAL(block)`** — `ProcessWithBAL` exists in `pkg/core/block_executor.go` and is tested in `bal_integration_test.go`, but its signature is `ProcessWithBAL(block, statedb)` (takes a `statedb` parameter), not `chain.Processor.ProcessWithBAL(block)`.

2. **`result.BlockAccessList.Entries`** — `BlockAccessList.Entries` exists as `[]AccessEntry` in `pkg/bal/types.go`, matching the plan.

3. **`block.Header().BlockAccessListHash`** — This field exists in the `types.Header` struct and is set by the block builder, confirmed by `TestBlockBuilder_SetsBALHash`.

4. **`result.BlockAccessList.Hash()`** — `Hash()` exists in `pkg/bal/hash.go` and works correctly (confirmed by `TestBALHash_Computed`).

5. **`chain.ParallelProcessor.ProcessParallel(block, bal)`** — No `ParallelProcessor` field or `ProcessParallel` method exists in the codebase. Parallel execution scheduling infrastructure is in `pkg/bal/scheduler.go` and `pkg/bal/parallel.go`, but it is not wired into the core block processor.

6. **Engine API round-trip** (`getPayloadV6` → `newPayloadV5`) — The plan references `engine_getPayloadV6` which is not implemented; the current Glamsterdam engine exposes `engine_getPayloadV5`. The round-trip validation via Engine API is not covered by any existing test.

7. **`bal.ApplyBAL`** — Does not exist (Story 6.1 dependency). Step 7 of the test cannot be written until `pkg/bal/apply.go` is created.

### Gaps and Proposed Solutions

1. **Write `TestBAL_FullPipeline` in phases**: Start with steps 1–4 (block building, BAL presence, address count, hash match) which are fully supported by the existing codebase. Gate steps 5–7 on implementation completion of their respective dependencies (parallel execution wiring, ApplyBAL).

2. **Fix `ProcessWithBAL` call signature**: The test helper `setupTestChain` must expose a `statedb` alongside the chain, or `ProcessWithBAL` must be called as `proc.ProcessWithBAL(block, statedb)` — not `chain.Processor.ProcessWithBAL(block)` as the plan shows.

3. **Parallel execution assertion (step 6)**: Add a `ProcessParallel` method to the block executor that uses the BAL scheduler from `pkg/bal/scheduler.go` to compute an execution wave order, then executes in parallel and asserts the state root matches the sequential result. This is non-trivial and should be tracked as a separate prerequisite sub-story.

4. **`ApplyBAL` assertion (step 7)**: Blocked on Story 6.1. Once `pkg/bal/apply.go` exists, add a sub-test that applies the BAL to a fresh state and compares state roots.

5. **Engine API round-trip (step 5)**: The test can validate the BAL hash embedded in the header and the BAL returned by `ProcessWithBAL` without a full Engine API call. A full round-trip test (`getPayloadV6` → `newPayloadV5`) requires implementing the missing endpoint and should be a separate story (or deferred to Sprint 14).
