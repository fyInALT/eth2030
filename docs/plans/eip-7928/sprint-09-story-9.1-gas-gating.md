# Story 9.1 — Gas-gate BAL emission: account access OOG and SSTORE stipend

> **Sprint context:** Sprint 9 — Two-Phase Gas Validation (BAL Inclusion Gate)
> **Sprint Goal:** Opcodes that fail pre-state gas validation MUST NOT add the target address/slot to the BAL. This is a spec requirement (lines 131–176): pre-state validation must pass before any state access occurs.

**Spec reference:** Lines 131–176. Pre-state gas validation table and SSTORE GAS_CALL_STIPEND rule.

**Files:**
- Modify: `pkg/core/vm/instructions.go`
- Test: `pkg/core/vm/gas_gating_test.go`

**Acceptance Criteria:** (1) If CALL/BALANCE/EXTCODESIZE/SLOAD exhausts gas before the pre-state cost is paid, no BAL event is emitted for that address/slot. (2) SSTORE inside a call with ≤ 2300 gas stipend does NOT emit any tracker event.

#### Task 9.1.1 — Write failing tests

File: `pkg/core/vm/gas_gating_test.go`

```go
func TestBAL_OOGBeforeAccess_ExcludesAddress(t *testing.T) {
    // Build EVM with exactly (COLD_ACCOUNT_ACCESS_COST - 1) gas remaining
    // Execute BALANCE on a cold address
    // Assert tracker contains NO entry for that address
}

func TestBAL_SufficientGas_IncludesAddress(t *testing.T) {
    // Same setup but with enough gas
    // Execute BALANCE
    // Assert tracker DOES contain an entry for that address
}

func TestBAL_SSTOREWithinStipend_ExcludesSlot(t *testing.T) {
    // Set up contract that tries SSTORE with exactly GAS_CALL_STIPEND (2300) gas
    // Execute
    // Assert tracker has NO storage event for that slot
}
```

#### Task 9.1.2 — Guard `RecordAddressAccess` calls with gas check

In each opcode handler that calls `tracker.RecordAddressAccess`, the call must only happen **after** the pre-state gas deduction succeeds. Pattern:

```go
// Pre-state: deduct access_cost first
if !scope.Contract.UseGas(accessCost) {
    return nil, ErrOutOfGas
}
// Only here is address actually accessed — emit BAL event
evm.AccessTracker.RecordAddressAccess(target.Bytes20(), evm.TxIndex)
```

Verify ordering in: BALANCE, EXTCODESIZE, EXTCODECOPY, EXTCODEHASH, CALL, CALLCODE, DELEGATECALL, STATICCALL, SELFDESTRUCT.

#### Task 9.1.3 — Guard SSTORE with GAS_CALL_STIPEND check

```go
// SSTORE: check GAS_CALL_STIPEND before accessing storage
if scope.Contract.Gas <= params.CallStipend {
    // MUST NOT appear in storage_reads or storage_changes
    return nil, ErrWriteProtection
}
// Only emit BAL event after passing the stipend check
evm.AccessTracker.RecordStorageRead(addr, slot, evm.TxIndex)
```

**Step: Run tests**

```
cd /projects/eth2030/pkg && go test ./core/vm/... -run "TestBAL_OOG|TestBAL_SSTORE" -v
```

Expected: PASS.

**Step: Commit**

```bash
git add pkg/core/vm/instructions.go pkg/core/vm/gas_gating_test.go
git commit -m "feat(bal): gas-gate BAL emission for OOG and SSTORE stipend"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
### Gas Validation Before State Access

State-accessing opcodes perform gas validation in two phases:

- **Pre-state validation**: Gas costs determinable without state access (memory expansion, base opcode cost, warm/cold access cost)
- **Post-state validation**: Gas costs requiring state access (account existence, EIP-7702 delegation resolution)

Pre-state validation MUST pass before any state access occurs. If pre-state validation fails, the target resource (address or storage slot) is never accessed and MUST NOT be included in the BAL.

Once pre-state validation passes, the target is accessed and included in the BAL. Post-state costs are then calculated; their order is implementation-defined since the target has already been accessed.

| Instruction | Pre-state Validation |
|-------------|----------------------|
| `BALANCE`   | `access_cost`        |
| `EXTCODESIZE` | `access_cost`      |
| `EXTCODEHASH` | `access_cost`      |
| `EXTCODECOPY` | `access_cost` + `memory_expansion` |
| `CALL`      | `access_cost` + `memory_expansion` + `GAS_CALL_VALUE` (if value > 0) |
| `CALLCODE`  | `access_cost` + `memory_expansion` + `GAS_CALL_VALUE` (if value > 0) |
| `DELEGATECALL` | `access_cost` + `memory_expansion` |
| `STATICCALL` | `access_cost` + `memory_expansion` |
| `SLOAD`     | `access_cost`        |
| `SSTORE`    | More than `GAS_CALL_STIPEND` available |
| `SELFDESTRUCT` | `GAS_SELF_DESTRUCT` + `access_cost` |

#### SSTORE

`SSTORE` performs an implicit read of the current storage value for gas calculation. The `GAS_CALL_STIPEND` check prevents this state access when operating within the call stipend. If `SSTORE` fails this check, the storage slot MUST NOT appear in `storage_reads` or `storage_changes`.
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/vm/interpreter.go` | EVM `Run` loop: calls `dynamicGas` then `UseGas`; defines `gasEIP2929AccountCheck` and `gasEIP2929SlotCheck` |
| `pkg/core/vm/gas_table.go` | Per-opcode dynamic gas functions: `gasBalanceEIP2929`, `gasSloadEIP2929`, `gasCallEIP2929`, `gasCallCodeEIP2929`, `gasExtCodeHashEIP2929`, `gasExtCodeSizeEIP2929`, `gasExtCodeCopyEIP2929`, `gasSelfdestructEIP2929` |
| `pkg/core/vm/instructions.go` | Opcode execution functions: `opBalance`, `opSload`, `opSstore`, `opCall`, `opCallCode`, `opDelegateCall`, `opStaticCall`, `opExtcodesize`, `opExtcodehash`, `opExtcodecopy`, `opSelfdestruct` |
| `pkg/core/vm/evm_storage_ops.go` | `StorageOpHandler`: higher-level SLOAD/SSTORE with access list integration |
| `pkg/core/vm/access_list_tracker.go` | `AccessListTracker`: warm/cold tracking with journaling; `TouchAddress`/`TouchSlot` |
| `pkg/bal/tracker.go` | `AccessTracker`: BAL emission methods `RecordStorageRead`, `RecordStorageChange`, `RecordBalanceChange` |

---

## Implementation Assessment

### Current Status

Not implemented.

### Architecture Notes

The EVM interpreter `Run` loop in `pkg/core/vm/interpreter.go` separates gas charging into two steps: first `constantGas` is deducted, then the `dynamicGas` function is called and its result is charged via `contract.UseGas`. The `dynamicGas` functions for state-accessing opcodes (e.g. `gasBalanceEIP2929` at line 589 of `gas_table.go`) call `gasEIP2929AccountCheck`, which calls `evm.StateDB.AddAddressToAccessList(addr)` to warm the address *before* `UseGas` is called on the result. This means:

1. For account-accessing opcodes (BALANCE, EXTCODESIZE, etc.): the address is warmed inside the `dynamicGas` call. If `UseGas` subsequently fails (OOG), the opcode execution function never runs, but the address is already in the state access list. The BAL, however, is currently populated from `populateTracker` in `processor.go`, which only captures balance and nonce deltas — it does not use the EVM-level access list at all. There is therefore no mechanism that would emit a BAL address entry from the EVM level, nor any mechanism that gates such emission on `UseGas` success.

2. For SSTORE: `opSstore` in `instructions.go` (line 594) checks `evm.readOnly` but does NOT check `contract.Gas <= params.CallStipend`. The `GAS_CALL_STIPEND` guard specified in the spec (line 158 of the EIP) is entirely absent.

3. The `bal.AccessTracker` (in `pkg/bal/tracker.go`) provides `RecordStorageRead`, `RecordStorageChange`, `RecordBalanceChange`, `RecordAddressTouch`, `RecordNonceChange`, and `RecordCodeChange` methods. These ARE called from EVM opcode handlers via the `BALTracker` interface: `instructions.go` has 15+ call sites using `evm.balTracker.RecordStorageRead`, `RecordStorageChange`, `RecordBalanceChange`, and `RecordAddressTouch`. The `Call()` and `create()` methods in `interpreter.go` also record balance changes for value transfers.

### Gaps and Proposed Solutions

1. **Per-opcode BAL emission is implemented**: The EVM struct has `balTracker BALTracker` and `txIndex uint64` fields. Opcode handlers in `instructions.go` call `evm.balTracker.RecordStorageRead`, `RecordStorageChange`, `RecordBalanceChange`, and `RecordAddressTouch` at 15+ call sites, gated by `evm.balTracker != nil` checks. The `Call()` method in `interpreter.go` also records balance changes for value transfers. This gap is resolved.

2. **`gasEIP2929AccountCheck` warms the address before `UseGas` succeeds**: The current ordering means that if the cold-access surcharge causes OOG, the address has been touched in the state access list. For EIP-7928, the BAL inclusion gate must be placed after `UseGas` returns successfully. The implementation should emit the BAL event inside the opcode execution function (which only runs if `UseGas` succeeded), not inside `dynamicGas`.

3. **SSTORE stipend check absent**: `opSstore` must gate on `contract.Gas <= CallStipend` before performing the state read. The existing `opSstore` code skips this check entirely. The check must be added at the top of `opSstore` (or its Glamsterdam variant `opSstoreGlamst`), returning `ErrWriteProtection` (or a new sentinel error) without emitting any BAL event.

4. **In-EVM instrumentation is complete**: The EVM now instruments every state-touching opcode via the `BALTracker` interface. SLOAD/SSTORE, BALANCE, EXTCODESIZE, EXTCODECOPY, EXTCODEHASH, CALL/CALLCODE/DELEGATECALL/STATICCALL targets, CREATE/CREATE2, and SELFDESTRUCT all emit BAL events from the opcode handlers. The `populateTracker` path in `processor.go` supplements this with sender/recipient tracking. This gap is resolved.
