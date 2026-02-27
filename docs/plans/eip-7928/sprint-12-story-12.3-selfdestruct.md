# Story 12.3 — SELFDESTRUCT In-Transaction Semantics

> **Sprint context:** Sprint 12 — Recording Semantics Edge Cases
> **Sprint Goal:** SSTORE no-op writes, gas refunds, exceptional halts, SELFDESTRUCT, SENDALL, and unaltered balances all produce correct BAL entries per the normative edge cases.

**Spec reference:** Lines 252-253. "SELFDESTRUCT: included without nonce or code changes. If positive balance, balance change to zero MUST be recorded. Storage reads MUST be included as storage_reads."

**Files:**
- Modify: `pkg/core/vm/instructions.go` (SELFDESTRUCT handler)
- Test: `pkg/core/vm/selfdestruct_bal_test.go`

**Acceptance Criteria:**
1. SELFDESTRUCT with positive pre-balance → address in BAL with balance_change to zero; no nonce_change, no code_change
2. SELFDESTRUCT with zero pre-balance → address in BAL with empty lists
3. Beneficiary receives balance → balance_change recorded for beneficiary
4. Any storage slots accessed in the selfdestructed contract → appear in storage_reads

#### Task 12.3.1 — Write failing tests

```go
func TestBAL_SELFDESTRUCT_PositiveBalance_BalanceZero(t *testing.T) {
    // Contract at 0xdead... has 5 ETH, selfdestructs to beneficiary 0xbene...
    // Assert 0xdead... in BAL with balance_changes = [[txIdx, 0]], no nonce/code
    // Assert 0xbene... in BAL with balance_changes = [[txIdx, 5 ETH]]
}

func TestBAL_SELFDESTRUCT_ZeroBalance_EmptyLists(t *testing.T) {
    // Contract at 0xdead... has 0 ETH, selfdestructs
    // Assert 0xdead... in BAL with all empty lists
}

func TestBAL_SELFDESTRUCT_StorageReadsPreserved(t *testing.T) {
    // Contract SLOADs slot 0x05 then selfdestructs
    // Assert slot 0x05 in storage_reads for that address
}
```

#### Task 12.3.2 — Implement in SELFDESTRUCT handler

```go
// SELFDESTRUCT: record beneficiary and sender
evm.AccessTracker.RecordAddressAccess(beneficiary.Bytes20(), evm.TxIndex)

preBalance := stateDB.GetBalance(contract.Address())
if preBalance.Sign() > 0 {
    // Record zero balance for selfdestructed account
    evm.AccessTracker.RecordBalanceChange(contract.Address().Bytes20(), [32]byte{}, evm.TxIndex)
    // Record beneficiary receiving the balance
    postBeneficiaryBalance := new(big.Int).Add(stateDB.GetBalance(beneficiary), preBalance)
    evm.AccessTracker.RecordBalanceChange(beneficiary.Bytes20(), uint256ToBytes32(postBeneficiaryBalance), evm.TxIndex)
}
// Note: NO nonce_change, NO code_change for selfdestructed account (per spec)
```

**Step: Run tests**

```
cd /projects/eth2030/pkg && go test ./core/vm/... -run TestBAL_SELFDESTRUCT -v
```

Expected: PASS.

**Step: Commit**

```bash
git add pkg/core/vm/instructions.go pkg/core/vm/selfdestruct_bal_test.go
git commit -m "feat(bal): SELFDESTRUCT correct BAL semantics"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
Line 252: - **SENDALL:** For positive-value selfdestructs, the sender and beneficiary are recorded with a balance change.
Line 253: - **SELFDESTRUCT (in-transaction):** Accounts destroyed within a transaction **MUST** be included in `AccountChanges` without nonce or code changes. However, if the account had a positive balance pre-transaction, the balance change to zero **MUST** be recorded. Storage keys within the self-destructed contracts that were modified or read **MUST** be included as a `storage_reads` entry.

Supporting context (line 105):
  - Beneficiary addresses for `SELFDESTRUCT`

Supporting context (lines 220-221):
  - **SELFDESTRUCT/SENDALL** beneficiaries. [recorded with a balance change]

Pre-state validation (line 159):
| `SELFDESTRUCT` | `GAS_SELF_DESTRUCT` + `access_cost` |
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/vm/instructions.go` (lines 1124-1156) | `opSelfdestruct` — the SELFDESTRUCT opcode handler; calls `AddBalance`/`SubBalance` but contains no BAL recording calls |
| `pkg/core/processor.go` (lines 152-160) | Post-tx BAL population via `populateTracker`; only captures sender/recipient balance and nonce deltas, not SELFDESTRUCT-specific semantics |
| `pkg/core/processor.go` (lines 207-222) | `populateTracker` — compares pre/post balances for sender and recipient only; does not cover beneficiary or the selfdestructed address |
| `pkg/bal/tracker.go` | `AccessTracker` with `RecordBalanceChange`, `RecordNonceChange`, `RecordCodeChange`, `RecordStorageRead` — the recording API that needs to be called from the SELFDESTRUCT handler |

---

## Implementation Assessment

### Current Status

Not implemented.

### Architecture Notes

The `opSelfdestruct` handler in `pkg/core/vm/instructions.go` (lines 1129-1156) correctly transfers the balance via `AddBalance`/`SubBalance` but has no BAL tracking at all. The post-transaction `populateTracker` in `pkg/core/processor.go` only snapshots the sender and `tx.To()` addresses before/after each transaction; it never sees the selfdestructed contract address or the beneficiary unless they coincide with the transaction sender/recipient.

The EVM struct itself (`pkg/core/vm/`) does not hold a reference to the `AccessTracker` — the tracker lives in the processor layer and is populated after the EVM returns, not during execution. This means opcode-level hooks (like inserting recording calls inside `opSelfdestruct`) would require either: (a) threading the tracker through the EVM context, or (b) extending the post-execution snapshot mechanism in `capturePreState`/`populateTracker` to cover the selfdestructed address and its beneficiary.

The post-EIP-6780 semantics add a wrinkle: the code notes that account destruction only occurs when the contract was created in the same transaction, tracked externally. The BAL spec (line 253) requires that the selfdestructed account be included without nonce or code changes regardless of whether the account is actually deleted.

### Gaps and Proposed Solutions

1. **No BAL recording in `opSelfdestruct`**: The handler does not call any tracker method. Solution: extend `capturePreState` to also snapshot the SELFDESTRUCT beneficiary and the contract address (accessible via EVM context/logs), then in `populateTracker` emit the zero-balance change for the selfdestructed account (if pre-balance was positive) and the balance increase for the beneficiary — while explicitly suppressing nonce and code changes for the selfdestructed address.

2. **No mechanism to identify SELFDESTRUCT at the processor level**: The processor has no way to know which addresses were selfdestructed. Solution: add a field to `ExecutionResult` (or use `statedb.HasSelfDestructed(addr)`) to expose selfdestructed addresses after execution, so `populateTracker` can apply the correct BAL semantics.

3. **Storage reads in selfdestructed contracts not tracked**: `capturePreState` does not snapshot storage. Solution: storage reads inside a selfdestructed contract are already tracked via `RecordStorageRead` if the EVM calls the tracker at SLOAD/SSTORE time; this requires the tracker to be wired into the EVM's storage opcode handlers, which is not yet done.

4. **Beneficiary address not included in `capturePreState`**: Since `capturePreState` only snapshots `tx.Sender()` and `tx.To()`, beneficiary addresses that differ from these will not appear in the BAL. Solution: expand `capturePreState` to accept a broader set of addresses, discoverable post-execution via the state journal or a new EVM hook.
