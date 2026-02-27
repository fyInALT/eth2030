# Story 12.1 — SSTORE No-Op Writes → `storage_reads`

> **Sprint context:** Sprint 12 — Recording Semantics Edge Cases
> **Sprint Goal:** SSTORE no-op writes, gas refunds, exceptional halts, SELFDESTRUCT, SENDALL, and unaltered balances all produce correct BAL entries per the normative edge cases.

**Spec reference:** Lines 207-209. "Slots written with unchanged values (SSTORE where post-value equals pre-value, also known as 'no-op writes') → storage_reads. Implementations MUST check the pre-transaction value."

**Files:**
- Modify: `pkg/core/vm/instructions.go` (SSTORE handler)
- Modify: `pkg/bal/builder.go`
- Test: `pkg/core/vm/sstore_noop_test.go`

**Acceptance Criteria:** SSTORE that writes the same value already in storage produces a `storage_reads` entry, NOT a `storage_changes` entry.

#### Task 12.1.1 — Write failing test

```go
func TestBAL_SSTORE_NoOp_GoesToStorageReads(t *testing.T) {
    // Pre-set slot 0x01 to value 0xff
    // Execute SSTORE(0x01, 0xff) — same value, no-op write
    // Assert: slot 0x01 is in storage_reads (NOT storage_changes)
}

func TestBAL_SSTORE_RealWrite_GoesToStorageChanges(t *testing.T) {
    // Pre-set slot 0x01 to value 0xff
    // Execute SSTORE(0x01, 0xab) — different value
    // Assert: slot 0x01 is in storage_changes
}
```

#### Task 12.1.2 — Implement pre-tx value check in SSTORE

In the SSTORE handler, compare new value against the pre-TRANSACTION (not pre-opcode) value:

```go
preTxValue := stateDB.GetPreTransactionStorageValue(addr, slot) // snapshot taken at tx start
newValue := scope.Stack.peek()

if newValue.Eq(preTxValue) {
    // No-op write: emit as storage READ
    evm.AccessTracker.RecordStorageRead(addr.Bytes20(), slot.Bytes32(), evm.TxIndex)
} else {
    // Real write
    evm.AccessTracker.RecordStorageWrite(addr.Bytes20(), slot.Bytes32(), newValue.Bytes32(), evm.TxIndex)
}
```

**Step: Run tests**

```
cd /projects/eth2030/pkg && go test ./core/vm/... -run TestBAL_SSTORE_NoOp -v
```

Expected: PASS.

**Step: Commit**

```bash
git add pkg/core/vm/instructions.go pkg/core/vm/sstore_noop_test.go
git commit -m "feat(bal): SSTORE no-op writes recorded as storage_reads"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
Lines 199-209:
Storage

- Writes include:
  - Any value change (post-value != pre-value).
  - Zeroing a slot (pre-value exists, post-value is zero).

- Reads include:
  - Slots accessed via SLOAD that are not written.
  - Slots written with unchanged values (i.e., SSTORE where post-value equals
    pre-value, also known as "no-op writes").

Note: Implementations MUST check the pre-transaction value to correctly
distinguish between actual writes and no-op writes.

Line 176 (SSTORE pre-state validation):
SSTORE performs an implicit read of the current storage value for gas
calculation. The GAS_CALL_STIPEND check prevents this state access when
operating within the call stipend. If SSTORE fails this check, the storage
slot MUST NOT appear in storage_reads or storage_changes.
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/vm/instructions.go` | `opSstore` (line 594): reads `current` via `GetState` and `original` via `GetCommittedState`, computes refund, then calls `SetState`; no BAL event emitted |
| `pkg/core/vm/instructions.go` | `opSstoreGlamst` (line 629): Glamsterdam variant — only writes state, no refund tracking, no BAL event emitted |
| `pkg/core/vm/evm_storage_ops.go` | `StorageOpHandler.SstoreGasCost` (line 97): the no-op case (`current == newVal`) is already detected at line 110 with comment "No-op case: writing the same value that already exists" — returns `WarmStorageReadGas` gas; this is the gas metering equivalent of what the BAL needs for storage classification |
| `pkg/bal/tracker.go` | `RecordStorageRead` (line 43) and `RecordStorageChange` (line 50): the two destination methods for the no-op vs real-write distinction |
| `pkg/core/processor.go` | `populateTracker` (line 207): current BAL population only records balance and nonce deltas; SSTORE events are not propagated to the BAL tracker at all |

---

## Implementation Assessment

### Current Status

Not implemented. Neither `opSstore` nor `opSstoreGlamst` in `pkg/core/vm/instructions.go` emit any BAL event. The `populateTracker` function in `processor.go` does not inspect storage state at all — it only compares balances and nonces. As a result, no SSTORE operation, no-op or otherwise, produces any entry in the BAL today.

### Architecture Notes

The spec requires comparing the **pre-transaction** value (`GetCommittedState`, i.e. the value at the start of the transaction before any writes in the current call stack) against the new value being written. This is distinct from the **pre-opcode** current value (`GetState`), which may already reflect earlier writes within the same transaction. The existing `opSstore` implementation already fetches both `current` (`GetState`) and `original` (`GetCommittedState`) for EIP-2200 gas calculation purposes (lines 604-605) — the same `original` value is the correct pre-transaction baseline for BAL classification.

The story proposes a `GetPreTransactionStorageValue(addr, slot)` API on the state DB. In the actual codebase `statedb.GetCommittedState(addr, slot)` (already used in `opSstore`) provides exactly this semantic: the value committed before the current transaction began.

### Gaps and Proposed Solutions

1. **No BAL storage recording in opSstore / opSstoreGlamst**: Both handlers must be extended to call the BAL tracker after determining write vs no-op. Concretely, after the existing `original`/`current` comparison, add:
   - If `original == newVal` (no-op write): call `tracker.RecordStorageRead(addr, slot, original)`.
   - If `original != newVal` (real write): call `tracker.RecordStorageChange(addr, slot, original, newVal)`.
   The `original` value from `GetCommittedState` is the correct pre-transaction baseline per the spec note.

2. **EVM struct has no reference to the BAL tracker**: The `evm *EVM` passed to opcode handlers does not currently carry a `bal.AccessTracker` reference. The tracker must either be added as a field on `EVM`, passed via the `StateDB` interface, or stored on the `BlockContext` so opcode handlers can reach it.

3. **Glamsterdam variant**: `opSstoreGlamst` does not read `original` at all today (EIP-7778 eliminates refunds). For BAL purposes the `GetCommittedState` read must be added back — but only for BAL classification, not for refund computation.
