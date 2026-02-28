# Story 11.1 — System contract tracking: pre-execution (index 0) and post-execution (index n+1)

> **Sprint context:** Sprint 11 — System Contracts & Withdrawal Tracking
> **Sprint Goal:** Pre-execution system contracts use `block_access_index = 0`; post-execution (withdrawals, EIP-7002, EIP-7251) use `block_access_index = n+1`; EIP-4895 withdrawal recipients are tracked.

**Spec reference:** Lines 106, 193, 259-266. EIP-2935/4788 at index 0; EIP-4895 withdrawals, EIP-7002, EIP-7251 at index `n+1`.

**Files:**
- Modify: `pkg/core/processor.go`
- Test: `pkg/core/syscall_bal_test.go`
- Test: `pkg/core/postcall_bal_test.go`

**Acceptance Criteria:**
1. EIP-2935 block hash contract: 1 storage write at index 0; EIP-4788: 2 storage writes at index 0.
2. EIP-4895 withdrawal recipients appear at index `n+1` (even zero-amount withdrawals); EIP-7002/7251 system contracts write slots 0-3 at index `n+1`.

#### Task 11.1.1 — Write failing tests

File: `pkg/core/syscall_bal_test.go`

```go
func TestBAL_EIP2935_PreExecution_Index0(t *testing.T) {
    // Execute a block with EIP-2935 active
    // Assert BlockHashContract (0x0000F908...) in BAL
    // Assert storage_change for the ring buffer slot has block_access_index = 0
}

func TestBAL_EIP4788_PreExecution_TwoSlots_Index0(t *testing.T) {
    // Execute a block with EIP-4788 active
    // Assert BeaconRootsContract in BAL
    // Assert exactly 2 storage_changes, both with block_access_index = 0
}
```

File: `pkg/core/postcall_bal_test.go`

```go
func TestBAL_EIP4895_WithdrawalRecipients_PostIndex(t *testing.T) {
    // Block with 1 transaction and 1 withdrawal to address 0xwith...
    // Assert 0xwith... in BAL with balance_changes at index 2 (= 1 tx + 1)
}

func TestBAL_EIP4895_ZeroAmountWithdrawal_RecipientIncluded(t *testing.T) {
    // Block with 1 withdrawal of amount 0 to address 0xwith...
    // Assert 0xwith... IS in BAL at index n+1
    // Assert balance_changes is EMPTY (no value transferred)
}

func TestBAL_EIP7002_PostExecution_Slots0to3(t *testing.T) {
    // Execute block with EIP-7002 dequeuing active
    // Assert EIP-7002 system contract has storage_changes for slots 0..3
    // Assert all storage_changes have block_access_index = len(txs) + 1
}
```

#### Task 11.1.2 — Set TxIndex=0 before pre-execution system calls

In `pkg/core/processor.go`:

```go
// Pre-execution system calls use block_access_index = 0
evm.TxIndex = 0
processEIP2935SystemCall(evm, stateDB, header)
processEIP4788SystemCall(evm, stateDB, header)
// Restore TxIndex for user transactions (1..n)
```

#### Task 11.1.3 — Set TxIndex=n+1 before post-execution phase

```go
// Post-execution index: len(transactions) + 1
postIndex := uint16(len(block.Transactions()) + 1)
evm.TxIndex = postIndex

// Process EIP-4895 withdrawals
for _, w := range block.Withdrawals() {
    applyWithdrawal(stateDB, w)
    postBalance := stateDB.GetBalance(w.Address)
    tracker.RecordBalanceChange(w.Address.Bytes20(), uint256ToBytes32(postBalance), postIndex)
}

// Process EIP-7002 / EIP-7251 system contracts
processEIP7002SystemCall(evm, stateDB, header)
processEIP7251SystemCall(evm, stateDB, header)
```

**Step: Run tests**

```
cd /projects/eth2030/pkg && go test ./core/... -run "TestBAL_EIP2935|TestBAL_EIP4788|TestBAL_EIP4895|TestBAL_EIP7002" -v
```

Expected: PASS.

**Step: Commit**

```bash
git add pkg/core/processor.go pkg/core/syscall_bal_test.go pkg/core/postcall_bal_test.go
git commit -m "feat(bal): system contract tracking at index 0 and n+1"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
Line 106:
System contract addresses accessed during pre/post-execution; the system
caller address, SYSTEM_ADDRESS (0xfffffffffffffffffffffffffffffffffffffffe),
MUST NOT be included unless it experiences state access itself.

Lines 191-193:
BlockAccessIndex values MUST be assigned as follows:
- 0 for pre-execution system contract calls.
- 1...n for transactions (in block order).
- n+1 for post-execution system contract calls.

Lines 259-266 (Edge Cases):
- Pre-execution system contract calls: All state changes MUST use
  block_access_index = 0.
- Post-execution system contract calls: All state changes MUST use
  block_access_index = len(transactions) + 1.
- EIP-2935 (block hash): Record system contract storage diffs of the
  single updated storage slot in the ring buffer.
- EIP-4788 (beacon root): Record system contract storage diffs of the
  two updated storage slots in the ring buffer.
- EIP-7002 (withdrawals): Record system contract storage diffs of storage
  slots 0-3 (4 slots) after the dequeuing call.
- EIP-7251 (consolidations): Record system contract storage diffs of
  storage slots 0-3 (4 slots) after the dequeuing call.
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/processor.go` | `ProcessWithBAL` (lines 72-183): orchestrates pre-execution system calls (EIP-4788 at line 82, EIP-2935 at line 95), transaction loop (lines 113-161), and post-execution withdrawals (lines 174-177); BAL tracker is created per-transaction (line 154) |
| `pkg/core/beacon_root.go` | `ProcessBeaconBlockRoot`: writes two storage slots (`timestampSlot`, `rootSlot`) to `BeaconRootAddress` (0x000F3df6...) via `statedb.SetState`; no BAL index is set or passed |
| `pkg/core/eip2935.go` | `ProcessParentBlockHash`: writes one storage slot (parent hash at `slot = parentNumber % 8192`) to `HistoryStorageAddress` (0x0F792be4...); no BAL index is set or passed |
| `pkg/core/eip7002.go` | `ProcessWithdrawalRequests`: reads and writes queue slots 0-3 and excess/count slots; no BAL index is set or passed; called indirectly via `ProcessRequests` (not from `ProcessWithBAL`) |
| `pkg/bal/tracker.go` | `AccessTracker.Build(txIndex uint64)`: accepts a single `txIndex` applied to all entries in the tracker; single-index granularity is the current design |

---

## Implementation Assessment

### Current Status

Not implemented. The current `ProcessWithBAL` in `processor.go` creates a fresh `bal.NewTracker()` per transaction (line 154) and calls `tracker.Build(uint64(i + 1))` to tag entries with indices `1..n`. No BAL tracking is performed for pre-execution system contracts (EIP-2935, EIP-4788) or post-execution system contracts (EIP-7002, EIP-7251). The withdrawal application (`ProcessWithdrawals`, lines 174-177) modifies state but is never tracked in the BAL at all. No `block_access_index = 0` path exists.

### Architecture Notes

The story's design calls for `evm.TxIndex` to be set before each phase, implying the EVM struct carries the current BAL index. In the actual codebase no `TxIndex` field exists on the EVM struct (`pkg/core/vm/`). Additionally, `ProcessBeaconBlockRoot` and `ProcessParentBlockHash` are standalone functions that receive `statedb` and `header` — they have no awareness of the BAL tracker and no index parameter. The BAL tracker in `pkg/bal/tracker.go` has no method to record pre-existing state changes with a per-call index override; `Build(txIndex)` stamps a single index on all entries recorded since the last reset.

### Gaps and Proposed Solutions

1. **No pre-execution tracking at index 0**: `ProcessBeaconBlockRoot` (2 slot writes) and `ProcessParentBlockHash` (1 slot write) are called in `ProcessWithBAL` before the transaction loop but produce no BAL entries. Solution: capture pre/post storage state around each system call and emit entries tagged with `block_access_index = 0` into the block-level BAL before the transaction loop begins.

2. **No post-execution tracking at index n+1**: `ProcessWithdrawals` modifies balances but the processor never records those changes in the BAL. `ProcessWithdrawalRequests` (EIP-7002) and its EIP-7251 counterpart are called from `ProcessWithRequests`, a separate code path that never touches the BAL. Solution: after the transaction loop, set `postIndex = len(txs) + 1`, capture pre/post state for all withdrawal recipients and system contract slot writes, and emit BAL entries tagged with `postIndex`.

3. **EIP-7251 file absent**: No `pkg/core/eip7251.go` exists yet. It must be created alongside a consolidation system contract address before post-execution tracking can be completed.

4. **Tracker design limitation**: `AccessTracker.Build(txIndex)` applies one index to all recorded entries. For a block containing multiple phases (pre, n transactions, post), the tracker must be reset and rebuilt separately for each phase, or the tracker API must be extended to support per-entry index assignment.
