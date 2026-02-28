# Story 10.1 — COINBASE Tracking with Zero-Reward Rule

> **Sprint context:** Sprint 10 — Special Address Tracking
> **Sprint Goal:** COINBASE, precompiles, EIP-2930 exclusion, and the SYSTEM_ADDRESS rule are all handled correctly by the tracker.

**Spec reference:** Lines 104, 233, 250. "If the COINBASE reward is zero, the COINBASE address MUST be included as a *read*. Zero-value block reward recipients MUST NOT trigger a balance change. MUST NOT be included for blocks with no transactions provided there are no other state changes."

**Files:**
- Modify: `pkg/core/processor.go`
- Test: `pkg/core/coinbase_bal_test.go`

**Acceptance Criteria:**
1. Block with non-zero fees → COINBASE in BAL with balance_changes after each tx
2. Block with zero reward, zero fees → COINBASE NOT in BAL
3. Block with zero block reward but non-zero tx fees → COINBASE in BAL as read-only (no balance_changes)

#### Task 10.1.1 — Write failing tests

File: `pkg/core/coinbase_bal_test.go`

```go
func TestBAL_COINBASE_NonZeroFees_HasBalanceChange(t *testing.T) {
    // Block with 1 tx paying 21000 * basefee in fees
    // Execute ProcessWithBAL
    // Assert COINBASE appears in BAL with balance_changes
}

func TestBAL_COINBASE_NoTxsAndNoWithdrawals_Absent(t *testing.T) {
    // Empty block (no transactions, no withdrawals)
    // Assert COINBASE absent from BAL
}

func TestBAL_COINBASE_ZeroRewardBlock_ReadOnly(t *testing.T) {
    // Block where COINBASE reward is exactly 0 (e.g. all fees burned, no block reward)
    // Assert COINBASE present as address-read entry with empty balance_changes
}
```

#### Task 10.1.2 — Implement COINBASE tracking in processor

After each transaction's fee application, in `pkg/core/processor.go`:

```go
// COINBASE: record balance change after every transaction
preBalance := stateDB.GetBalance(header.Coinbase)
applyFees(tx, stateDB, header)
postBalance := stateDB.GetBalance(header.Coinbase)

if postBalance.Cmp(preBalance) != 0 {
    tracker.RecordBalanceChange(header.Coinbase.Bytes20(), uint256ToBytes32(postBalance), txIndex)
} else {
    // Balance unchanged → still record as address read (coinbase was touched)
    tracker.RecordAddressAccess(header.Coinbase.Bytes20(), txIndex)
}
```

After all processing, apply the zero-reward rule:

```go
// EIP-7928: zero block reward means coinbase is read-only in BAL
if blockReward.Sign() == 0 && len(block.Transactions()) == 0 {
    // Remove coinbase from BAL entirely
    delete(events, coinbaseAddr)
}
```

**Step: Run tests**

```
cd /projects/eth2030/pkg && go test ./core/... -run TestBAL_COINBASE -v
```

Expected: PASS.

**Step: Commit**

```bash
git add pkg/core/processor.go pkg/core/coinbase_bal_test.go
git commit -m "feat(bal): COINBASE zero-reward and empty-block rules"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
  - COINBASE address if the block contains transactions or withdrawals to the COINBASE address
```

From the Balance section:

```
- **COINBASE** (rewards + fees).
```

From the Edge Cases section:

```
- **COINBASE / Fee Recipient:** The COINBASE address MUST be included if it experiences any state change. It MUST NOT be included for blocks with no transactions, provided there are no other state changes (e.g., from EIP-4895 withdrawals). If the COINBASE reward is zero, the COINBASE address MUST be included as a *read*.

Zero-value block reward recipients MUST NOT trigger a balance change in the block access list and MUST NOT cause the recipient address to be included as a read (e.g. without changes). Zero-value block reward recipients MUST only be included with a balance change in blocks where the reward is greater than zero.
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/processor.go` | `ProcessWithBAL`: main block execution loop, fee payment to `header.Coinbase` via `statedb.AddBalance`; `populateTracker` called after each tx but does not include coinbase; `applyTransaction` adds coinbase to access list and pays tip/full fees |
| `pkg/core/block_builder.go` | `BuildBlock`/`BuildBlockLegacy`: also calls `ApplyTransaction` and uses `bal.NewTracker` + `populateTracker`; same gap — no coinbase BAL recording |
| `pkg/bal/tracker.go` | `AccessTracker`: `RecordBalanceChange`, `RecordAddressAccess` (missing), `touchedAddrs`; BAL population methods |

---

## Implementation Assessment

### Current Status

Partially implemented — fee payment to coinbase is performed, but no BAL events are emitted for the coinbase address.

### Architecture Notes

In `pkg/core/processor.go`, the function `applyTransaction` (line 968–984) pays the tip to `header.Coinbase` via `statedb.AddBalance(header.Coinbase, tipPayment)`. This state change occurs after the EVM execution and gas refunds. However, the coinbase address is never passed to `capturePreState`, so `populateTracker` has no pre-tx coinbase balance snapshot and cannot detect the coinbase balance change.

The `capturePreState` function (line 188) captures only the tx sender (from `tx.Sender()`) and the tx recipient (from `tx.To()`). The coinbase address is not included.

The block_builder's `populateTracker` call has exactly the same gap.

There is also no logic anywhere that implements the three-case coinbase rule:
- Non-zero fee block: coinbase appears with balance_changes.
- Zero-reward, zero-fee block with no transactions: coinbase must NOT appear.
- Zero block reward with non-zero tx fees: coinbase appears as address-read only (empty change list).

The `bal.AccessTracker` does not have a `RecordAddressAccess` method for the read-only case, which would be needed for the "zero reward, non-zero fees" scenario (or if the coinbase balance happened to be unchanged after fees and any other adjustments).

### Gaps and Proposed Solutions

1. **Coinbase not in `capturePreState`**: The pre-tx balance snapshot for the coinbase address is never taken. Solution: extend `capturePreState` to also capture `header.Coinbase` balance, passing the header into the function. Then `populateTracker` will detect the post-tx balance delta and call `RecordBalanceChange` for the coinbase.

2. **Zero-reward / empty-block rules not enforced**: After all transactions are processed, the current code makes no decision about whether to include or exclude the coinbase from the BAL. Solution: after the transaction loop in `ProcessWithBAL`, inspect the coinbase entry in the tracker and apply the spec rules:
   - If the block has no transactions and no withdrawals touched the coinbase, delete any coinbase entry.
   - If the coinbase balance is unchanged (tip = 0, e.g. all fees burned), emit a read-only access (add `RecordAddressAccess` to `AccessTracker`) rather than a balance change.

3. **`bal.AccessTracker` missing `RecordAddressAccess`**: The read-only coinbase case (zero reward) requires emitting the address with an empty change list. This method must be added to `AccessTracker`, adding the address to `touchedAddrs` without any change records.

4. **`processor.go` vs `block_builder.go` duplication**: Both files duplicate the `populateTracker` + BAL-building pattern. The coinbase fix must be applied in both, or the logic centralized into a shared helper.
