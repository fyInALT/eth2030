# Story 12.4 — Balance recording edge cases: net-zero delta and gas refunds

> **Sprint context:** Sprint 12 — Recording Semantics Edge Cases
> **Sprint Goal:** SSTORE no-op writes, gas refunds, exceptional halts, SELFDESTRUCT, SENDALL, and unaltered balances all produce correct BAL entries per the normative edge cases.

**Spec reference:** Lines 224-226, 256. Net-zero balance delta MUST NOT appear in `balance_changes`; sender balance MUST be recorded after the gas refund is applied.

**Files:**
- Modify: `pkg/bal/builder.go`
- Modify: `pkg/core/processor.go`
- Test: `pkg/bal/builder_balance_test.go`
- Test: `pkg/core/gas_refund_bal_test.go`

**Acceptance Criteria:**
1. An account whose balance changes mid-transaction but ends equal to its pre-tx balance is in the BAL with an empty `balance_changes` list.
2. When a transaction triggers a gas refund, the sender's `balance_changes` entry reflects the post-refund balance.

#### Task 12.4.1 — Write failing tests

```go
func TestBAL_UnalteredBalance_OmittedFromBalanceChanges(t *testing.T) {
    // Account receives 1 ETH then sends 1 ETH back in same transaction
    // Post-tx balance == pre-tx balance
    // Assert: address IS in BAL (was accessed)
    // Assert: balance_changes is empty
}

func TestBAL_GasRefund_SenderFinalBalance(t *testing.T) {
    // Transaction clears a storage slot (earns gas refund)
    // Assert: sender's balance_changes[0].PostBalance == actual post-refund balance
}

func TestBAL_GasRefund_RecordedAfterRefundApplied(t *testing.T) {
    // Two transactions from same sender, second clears storage → refund
    // Assert balance_changes for second tx reflects the refunded amount
}
```

#### Task 12.4.2 — Omit balance_changes when net delta is zero

In `BuildFromEvents` / `buildAccountEntry`:

```go
// Only emit balance_change if post-tx balance != pre-tx balance
if preTxBalance[addr] != postTxBalance {
    entry.BalanceChanges = append(...)
}
// Address still appears in BAL (empty lists) if it was accessed
```

#### Task 12.4.3 — Record sender balance after gas refund

In `pkg/core/processor.go`, ensure `RecordBalanceChange` is called **after** `refundGas`:

```go
result, err := applyTransaction(tx, stateDB, evm, gp)
refundGas(tx, result, stateDB, header)

// EIP-7928: record sender balance AFTER refund — this is the final balance
postBalance := stateDB.GetBalance(sender)
tracker.RecordBalanceChange(sender.Bytes20(), uint256ToBytes32(postBalance), txIndex)
```

**Step: Run tests**

```
cd /projects/eth2030/pkg && go test ./bal/... ./core/... -run "TestBAL_UnalteredBalance|TestBAL_GasRefund" -v
```

Expected: PASS.

**Step: Commit**

```bash
cd /projects/eth2030/pkg && go fmt ./bal/... ./core/...
git add pkg/bal/builder.go pkg/bal/builder_balance_test.go \
        pkg/core/processor.go pkg/core/gas_refund_bal_test.go
git commit -m "feat(bal): net-zero balance omission + post-refund sender balance"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
Lines 224-226:
For unaltered account balances:

If an account's balance changes during a transaction, but its post-transaction balance is equal to its pre-transaction balance, then the change **MUST NOT** be recorded in `balance_changes`. The sender and recipient address **MUST** be included in `AccountChanges`.

Line 256:
- **Gas refunds:** Record the **final** balance of the sender after each transaction.
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/processor.go` (lines 207-222) | `populateTracker` — compares `preBal` to `postBal` and only calls `RecordBalanceChange` if they differ; the net-zero guard is already present here |
| `pkg/core/processor.go` (lines 916-966) | `applyMessage` — gas refund logic: `refund` is computed and subtracted from `gasUsed`; remaining gas is then refunded to the sender via `statedb.AddBalance` at line 955 |
| `pkg/core/processor.go` (lines 113-161) | `ProcessWithBAL` transaction loop — `capturePreState` runs before `applyTransaction`, and `populateTracker` runs after; the pre-snapshot happens before the refund is applied |
| `pkg/bal/tracker.go` (lines 57-63) | `RecordBalanceChange` — accepts old and new `*big.Int` values; the caller in `populateTracker` supplies the post-tx state values |

---

## Implementation Assessment

### Current Status

Partially implemented.

### Architecture Notes

**Net-zero balance omission** is already correctly implemented. In `populateTracker` (lines 210-214 of `processor.go`), a balance change is only recorded when `preBal.Cmp(postBal) != 0`. If a transaction changes the sender balance mid-execution but restores it to the pre-transaction value, the post-execution read will equal the pre-snapshot, and no entry is recorded. The address is still present in the BAL because `capturePreState` adds it to `preBalances`, and all keys of `preBalances` are iterated in `populateTracker` — but since no `RecordBalanceChange` is called for the unchanged case, `touchedAddrs` is not marked from the balance path. This means the address will only appear in the BAL if it was touched via another path (nonce change, etc.). This is a potential correctness gap: the spec requires the address to be present in `AccountChanges` even with empty change lists.

**Gas refund timing** is the primary gap. `capturePreState` snapshots the sender balance at line 123, before `applyTransaction` is called. Inside `applyTransaction` → `applyMessage`, the gas refund is applied at line 929 (`gasUsed -= refund`) and then the remaining gas (including refunded gas) is credited back to the sender at line 955 (`statedb.AddBalance(msg.From, refundAmount)`). The `populateTracker` call at line 155 then reads `statedb.GetBalance(addr)` which will already include the refund — so the recorded post-balance is the correct post-refund value. However, the pre-snapshot at line 123 was taken before gas deduction (which happens inside `applyMessage` at line 763). The net effect is that the balance delta in the BAL does reflect the full gas deduction minus the refund, which is the correct final balance per spec line 256.

The correctness concern is more subtle: the pre-snapshot at `capturePreState` does not deduct the gas upfront, so if execution reverts and no refund is issued, the delta is computed from the state as left by `applyMessage` — which handles the full deduct-then-partially-refund cycle internally. This works correctly only because `populateTracker` reads the post-execution state, which includes all intermediate modifications.

### Gaps and Proposed Solutions

1. **Accessed-but-unchanged address not guaranteed in BAL**: If an address appears in `preBalances` (because it is the sender or recipient) but has no net balance change and no nonce change, it will not be added to `touchedAddrs` in the tracker and thus will not appear in the BAL at all. The spec (lines 224-226) requires the address to be present in `AccountChanges`. Solution: in `populateTracker`, always call `tracker.RecordTouchedAddress(addr)` (a new method to add) for every address in `preBalances`, even when no change is detected.

2. **No `pkg/bal/builder.go` exists**: The story references `pkg/bal/builder.go` but the actual BAL population logic resides in `pkg/bal/tracker.go` and the `populateTracker` helper in `pkg/core/processor.go`. The `pkg/bal/` package has no `builder.go` file. The story's architecture diagram does not match the actual implementation structure.

3. **Gas refund correctness verified, but coinbase tip timing is off**: The coinbase tip is paid inside `applyMessage` (line 972-984) after the refund, but the coinbase address is not in `capturePreState`. Coinbase balance changes are therefore not tracked in the BAL for transaction fees. This is a separate gap from Story 12.4 but intersects with it.

4. **`pkg/core/processor.go` is the correct file to modify** (not `pkg/core/processor.go` as named in the story — but note the story references the file correctly). No change needed here; the gas refund path produces the correct final balance for the sender automatically.
