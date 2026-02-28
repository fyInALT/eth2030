# Story 12.2 — Exceptional Halts: Reverted Calls Still Include Accessed Addresses

> **Sprint context:** Sprint 12 — Recording Semantics Edge Cases
> **Sprint Goal:** SSTORE no-op writes, gas refunds, exceptional halts, SELFDESTRUCT, SENDALL, and unaltered balances all produce correct BAL entries per the normative edge cases.

**Spec reference:** Lines 258. "Exceptional halts: Record the final nonce and balance of the sender, and the final balance of the fee recipient after each transaction. State changes from the reverted call are discarded, but all accessed addresses MUST be included. If storage was read, the keys MUST appear in storage_reads."

**Files:**
- Modify: `pkg/core/processor.go`
- Test: `pkg/core/exceptional_halt_test.go`

**Acceptance Criteria:** When a transaction reverts (OOG, invalid opcode, assertion failure), the BAL still contains all addresses that were accessed before the revert, with storage reads preserved; balance/nonce of sender are recorded as final post-revert values.

#### Task 12.2.1 — Write failing tests

```go
func TestBAL_RevertedTx_AccessedAddressesPreserved(t *testing.T) {
    // Transaction that calls EXTCODEHASH on 0xdeadbeef... then OOGs
    // Assert 0xdeadbeef... IS in BAL (was accessed before the OOG)
    // Assert sender IS in BAL with final nonce + balance (gas deducted)
}

func TestBAL_RevertedTx_StorageReadsPreserved(t *testing.T) {
    // Transaction that SLOADs slot 0x01 then reverts
    // Assert slot 0x01 appears in storage_reads (not discarded with revert)
}
```

#### Task 12.2.2 — Separate revert-state rollback from BAL rollback

The key implementation insight: when a transaction reverts, the state is rolled back but the tracker events are NOT rolled back. After revert, still emit the final sender nonce/balance:

```go
snap := stateDB.Snapshot()
err := executeTx(tx, stateDB, evm, tracker)
if err != nil {
    stateDB.RevertToSnapshot(snap)
    // BAL events from before the revert are PRESERVED (tracker not rolled back)
    // But still record final sender balance/nonce after revert
}
// Always record sender final state regardless of revert
tracker.RecordNonceChange(sender.Bytes20(), stateDB.GetNonce(sender), txIndex)
tracker.RecordBalanceChange(sender.Bytes20(), uint256ToBytes32(stateDB.GetBalance(sender)), txIndex)
tracker.RecordBalanceChange(coinbase.Bytes20(), uint256ToBytes32(stateDB.GetBalance(coinbase)), txIndex)
```

**Step: Run tests**

```
cd /projects/eth2030/pkg && go test ./core/... -run TestBAL_Reverted -v
```

Expected: PASS.

**Step: Commit**

```bash
git add pkg/core/processor.go pkg/core/exceptional_halt_test.go
git commit -m "feat(bal): reverted txs preserve BAL events before revert"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
Line 258:
Exceptional halts: Record the final nonce and balance of the sender, and
the final balance of the fee recipient after each transaction. State changes
from the reverted call are discarded, but all accessed addresses MUST be
included. If no changes remain, addresses are included with empty lists;
if storage was read, the corresponding keys MUST appear in storage_reads.
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/processor.go` | `applyTransaction` (line 438): takes a snapshot before calling `applyMessage` (line 441) and calls `statedb.RevertToSnapshot(snapshot)` on error (line 445); this rolls back all state changes but the BAL tracker is not consulted |
| `pkg/core/processor.go` | `ProcessWithBAL` (lines 152-160): creates a fresh `bal.NewTracker()` and calls `populateTracker` after a successful `applyTransaction`; if `applyTransaction` returns an error the block processing aborts entirely (line 127-128), so no BAL entry is produced for failed transactions |
| `pkg/core/processor.go` | `populateTracker` (line 207): only records balance and nonce deltas; no storage reads or address-touch events are captured here |
| `pkg/bal/tracker.go` | `AccessTracker`: no snapshot/revert mechanism — once an event is recorded via `RecordStorageRead` etc., it cannot be selectively reverted; this is the property the story exploits (tracker is never rolled back on revert) |
| `pkg/core/vm/access_list_tracker.go` | `AccessListTracker.RevertToSnapshot` (line 171): rolls back EIP-2929 warm-set entries added after a snapshot; this is separate from the BAL tracker and controls gas pricing, not BAL inclusion |

---

## Implementation Assessment

### Current Status

Not implemented. The current `ProcessWithBAL` in `processor.go` short-circuits on any `applyTransaction` error (line 127: `return nil, fmt.Errorf(...)`), so the BAL receives no entry at all for a failing transaction. Even for successful transactions, only balance and nonce deltas are recorded — addresses accessed without state change (e.g. via `EXTCODEHASH` before an OOG) are never emitted. The key design requirement — that BAL address and storage-read events survive a state revert — cannot be implemented with the current architecture because no opcode-level BAL events exist to preserve.

### Architecture Notes

The story's plan draws a clean line between state rollback (via `statedb.RevertToSnapshot`) and BAL preservation (tracker not rolled back). The `bal.AccessTracker` in `pkg/bal/tracker.go` already satisfies the "no revert" property by design — it has no snapshot mechanism. However, this only matters once opcode-level BAL events are actually emitted during execution. Today, BAL events are only emitted post-execution via `populateTracker`, by which point the state has already been reverted for failed transactions, making the pre-revert accesses unobservable.

A second issue is that `applyTransaction` in `processor.go` currently treats any execution error as a block-level error that aborts processing (line 126-128 in `ProcessWithBAL`). In a correct EVM implementation, transaction-level reverts (OOG, REVERT opcode, invalid opcode) should produce a failed receipt but allow block processing to continue. This is a separate correctness issue that must be resolved before the exceptional-halt BAL story can be tested end-to-end.

### Gaps and Proposed Solutions

1. **No opcode-level BAL event emission**: All accessed addresses and storage reads must be recorded during EVM execution (not post-hoc) so they are available before any state revert. This is a prerequisite shared with Stories 10.3, 11.1, and 12.1. The BAL tracker reference must be threaded into the EVM execution context.

2. **Transaction errors abort block processing**: The `ProcessWithBAL` loop calls `return nil, fmt.Errorf(...)` on the first failed transaction. This must be changed to distinguish between block-invalid errors (nonce mismatch, gas limit exceeded) and transaction-level execution failures (OOG, revert). For the latter, state is rolled back but processing continues and a failed receipt is emitted.

3. **Post-revert sender/coinbase recording**: After a revert, the spec requires the sender's final nonce and balance and the coinbase's final balance to still be recorded. The `populateTracker` approach (post-tx diff) would capture these correctly if called after the state revert, because the reverted state is the "final" state for a failed transaction. This part of the design is sound — it just needs to be wired to also execute on the failure path, not only on the success path.
