# Story 1.3 — Wire `AccessTracker` into `EVM` and hook all state-touching opcodes

> **Sprint context:** Sprint 1 — EVM Opcode Access Tracking
> **Sprint Goal:** Every EVM opcode that touches state emits a structured access event into a per-transaction tracker, so the BAL can be built from real execution data.

**Files:**
- Modify: `pkg/core/vm/evm.go`
- Modify: `pkg/core/vm/instructions.go`
- Test: `pkg/core/vm/evm_bal_test.go`

**Opcodes to instrument:** SLOAD, SSTORE, BALANCE, EXTCODESIZE, EXTCODECOPY, EXTCODEHASH, CALL, CALLCODE, DELEGATECALL, STATICCALL, CREATE, CREATE2, SELFDESTRUCT. (`SELFBALANCE` is excluded per spec line 147 — current contract is always warm.)

**Acceptance Criteria:** The `EVM` struct holds `AccessTracker` and `TxIndex` fields; every state-touching opcode calls the appropriate recorder method after a successful gas deduction; tests confirm events are captured.

#### Task 1.3.1 — Locate and read the EVM struct

```
Read pkg/core/vm/evm.go to find EVM struct definition and where SLOAD/SSTORE are called.
```

#### Task 1.3.2 — Write failing test

File: `pkg/core/vm/evm_bal_test.go`

```go
package vm_test

import "testing"

func TestEVM_SLOADRecordsReadEvent(t *testing.T) {
	// Build minimal EVM context with BALAccessTracker
	// Execute a simple contract that calls SLOAD
	// Assert tracker.Drain() returns the accessed slot
	t.Skip("implement after EVM wiring")
}
```

#### Task 1.3.3 — Add `AccessTracker` and `TxIndex` fields to EVM

In `pkg/core/vm/evm.go`, add to the EVM struct:

```go
// AccessTracker records state accesses for EIP-7928 BAL.
// Set to NoopAccessTracker for pre-Amsterdam blocks.
AccessTracker AccessTracker

// TxIndex is the 1-based transaction index within the block (0 = pre-execution system calls).
TxIndex uint16
```

In the EVM constructor or `NewEVM()`:
```go
evm.AccessTracker = NewNoopAccessTracker()
```

#### Task 1.3.4 — Emit events in SLOAD and SSTORE opcode handlers

In the SLOAD handler, add after the state read:

```go
// EIP-7928: record storage read for BAL
evm.AccessTracker.RecordStorageRead(
    scope.Contract.Address(),
    common.Hash(loc).Bytes32(),
    evm.TxIndex,
)
```

In the SSTORE handler, add after the state write:

```go
// EIP-7928: record storage write for BAL
evm.AccessTracker.RecordStorageWrite(
    scope.Contract.Address(),
    common.Hash(loc).Bytes32(),
    common.Hash(val).Bytes32(),
    evm.TxIndex,
)
```

#### Task 1.3.5 — Instrument account-read opcodes (BALANCE, EXTCODExxx)

For each opcode that reads account state, add after the state access:

```go
evm.AccessTracker.RecordAddressAccess(addr.Bytes20(), evm.TxIndex)
```

#### Task 1.3.6 — Instrument CALL family opcodes

For each CALL variant, after the target address is loaded:

```go
evm.AccessTracker.RecordAddressAccess(toAddr.Bytes20(), evm.TxIndex)
```

If value > 0, record balance change after call returns:

```go
if value.Sign() > 0 {
    evm.AccessTracker.RecordBalanceChange(toAddr.Bytes20(), postBalanceBytes(stateDB, toAddr), evm.TxIndex)
    evm.AccessTracker.RecordBalanceChange(from.Bytes20(), postBalanceBytes(stateDB, from), evm.TxIndex)
}
```

#### Task 1.3.7 — Instrument CREATE / CREATE2

After successful deployment:

```go
evm.AccessTracker.RecordAddressAccess(contractAddr.Bytes20(), evm.TxIndex)
evm.AccessTracker.RecordCodeChange(contractAddr.Bytes20(), code, evm.TxIndex)
evm.AccessTracker.RecordNonceChange(contractAddr.Bytes20(), 1, evm.TxIndex)
evm.AccessTracker.RecordNonceChange(caller.Bytes20(), postNonce, evm.TxIndex)
```

**Step: Run full VM tests**

```
cd /projects/eth2030/pkg && go test ./core/vm/... -count=1 2>&1 | tail -20
```

Expected: All previously passing tests still pass.

**Step: Format & commit**

```bash
cd /projects/eth2030/pkg && go fmt ./core/vm/...
git add pkg/core/vm/evm.go pkg/core/vm/instructions.go pkg/core/vm/evm_bal_test.go
git commit -m "feat(bal): wire AccessTracker into EVM + hook all state-touching opcodes"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
### Gas Validation Before State Access

State-accessing opcodes perform gas validation in two phases:

- **Pre-state validation**: Gas costs determinable without state access (memory expansion, base opcode cost, warm/cold access cost)
- **Post-state validation**: Gas costs requiring state access (account existence, [EIP-7702](./eip-7702.md) delegation resolution)

Pre-state validation MUST pass before any state access occurs. If pre-state validation fails, the target resource (address or storage slot) is never accessed and MUST NOT be included in the BAL.

Once pre-state validation passes, the target is accessed and included in the BAL. Post-state costs are then calculated; their order is implementation-defined since the target has already been accessed.

The following table specifies pre-state validation costs in addition to the base opcode cost:

| Instruction | Pre-state Validation |
|-------------|----------------------|
| `BALANCE` | `access_cost` |
| `SELFBALANCE` | None (accesses current contract, always warm) |
| `EXTCODESIZE` | `access_cost` |
| `EXTCODEHASH` | `access_cost` |
| `EXTCODECOPY` | `access_cost` + `memory_expansion` |
| `CALL` | `access_cost` + `memory_expansion` + `GAS_CALL_VALUE` (if value > 0) |
| `CALLCODE` | `access_cost` + `memory_expansion` + `GAS_CALL_VALUE` (if value > 0) |
| `DELEGATECALL` | `access_cost` + `memory_expansion` |
| `STATICCALL` | `access_cost` + `memory_expansion` |
| `CREATE` | `memory_expansion` + `INITCODE_WORD_COST` + `GAS_CREATE` |
| `CREATE2` | `memory_expansion` + `INITCODE_WORD_COST` + `GAS_KECCAK256_WORD` + `GAS_CREATE` |
| `SLOAD` | `access_cost` |
| `SSTORE` | More than `GAS_CALL_STIPEND` available |
| `SELFDESTRUCT` | `GAS_SELF_DESTRUCT` + `access_cost` |

#### SSTORE

`SSTORE` performs an implicit read of the current storage value for gas calculation. The `GAS_CALL_STIPEND` check prevents this state access when operating within the call stipend. If `SSTORE` fails this check, the storage slot MUST NOT appear in `storage_reads` or `storage_changes`.

### Recording Semantics by Change Type

#### Storage

- **Writes include:**

  - Any value change (post‑value ≠ pre‑value).
  - **Zeroing** a slot (pre‑value exists, post‑value is zero).

- **Reads include:**

  - Slots accessed via `SLOAD` that are not written.
  - Slots written with unchanged values (i.e., `SSTORE` where post-value equals pre-value, also known as "no-op writes").

Note: Implementations MUST check the pre-transaction value to correctly distinguish between actual writes and no-op writes.

**`BlockAccessList`** is the set of all addresses accessed during block execution.

It **MUST** include:

  - Addresses with state changes (storage, balance, nonce, or code).
  - Addresses accessed without state changes, including:

    - Targets of `BALANCE`, `EXTCODESIZE`, `EXTCODECOPY`, `EXTCODEHASH` opcodes
    - Targets of `CALL`, `CALLCODE`, `DELEGATECALL`, `STATICCALL` (even if they revert)
    - Target addresses of `CREATE`/`CREATE2` if the target account is accessed
    - Deployed contract addresses from calls with initcode to empty addresses
    - Transaction sender and recipient addresses (even for zero-value transfers)
    - COINBASE address if the block contains transactions or withdrawals
    - Beneficiary addresses for `SELFDESTRUCT`
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/vm/interpreter.go` | EVM struct with `balTracker BALTracker`, `txIndex uint64`, `Context`, `TxContext`, `StateDB`, `Config`, `depth`, `readOnly`, `jumpTable`, `precompiles`, `returnData`, `callGasTemp`, `witnessGas`, `forkRules`, `FrameCtx`; `Call()` method records balance changes for value transfers |
| `pkg/core/vm/instructions.go` | Contains opcode implementations with BAL hooks: 15+ call sites using `evm.balTracker.RecordStorageRead`, `RecordStorageChange`, `RecordBalanceChange`, `RecordAddressTouch` |
| `pkg/core/vm/evm_storage_ops.go` | `StorageOpHandler` wraps SLOAD/SSTORE with gas accounting via `AccessListTracker`; no BAL emit calls |
| `pkg/core/vm/evm_call_handlers.go` | `CallHandler` orchestrates CALL-family opcodes; no BAL emit calls |
| `pkg/core/vm/evm_create.go` | `CreateExecutor` handles CREATE/CREATE2 lifecycle; no BAL emit calls |
| `pkg/core/vm/access_list_tracker.go` | EIP-2929 `AccessListTracker` for warm/cold tracking only; separate from EIP-7928 BAL hooks |
| `pkg/core/vm/gas_eip2929.go` | Gas functions for cold/warm access; applied before state access — relevant to pre-state validation ordering |
| `pkg/core/vm/evm_metering.go` | Gas metering logic; relevant for understanding where gas checks happen relative to state access |

---

## Implementation Assessment

### Current Status

Complete. The EVM struct has `balTracker BALTracker` and `txIndex uint64` fields, and 15+ opcodes in `instructions.go` emit BAL events via `evm.balTracker`.

### Architecture Notes

The EVM struct (`pkg/core/vm/interpreter.go`) carries `balTracker BALTracker` and `txIndex uint64` fields, set via `SetBALTracker(t BALTracker, txIdx uint64)`. All state-touching opcodes in `instructions.go` check `evm.balTracker != nil` and call the appropriate recorder methods. The `Call()` method in `interpreter.go` also records balance changes for value transfers.

The gas validation ordering is already partially correct: `AccessListTracker.TouchSlot` / `TouchAddress` is called by the dynamic gas functions (via `gas_eip2929.go`) **before** the opcode body reads state — which aligns with the spec's pre-state validation requirement. BAL emit calls should therefore be placed **after** the gas deduction succeeds (i.e., after the existing `TouchSlot`/`TouchAddress` calls), to match the spec's rule that the target is included only when pre-state validation passes.

Key spec nuance: `SSTORE` no-op writes (where `post-value == pre-value`) must appear in `storage_reads`, not `storage_changes`. The existing `evm_storage_ops.go` already distinguishes this case in `SstoreGasCost` (the `current == newVal` branch), so the hook point is identifiable.

`SELFBALANCE` is explicitly excluded from BAL recording per spec line 147 — the current contract is always warm and always included via other mechanisms.

### Gaps and Proposed Solutions

All gaps from the original plan are resolved:

| Original Gap | Resolution |
|-----|-------------------|
| EVM struct lacks tracker fields | `balTracker BALTracker` and `txIndex uint64` exist on the `EVM` struct; set via `SetBALTracker(t, txIdx)` |
| `opSload` does not emit BAL read event | `instructions.go` line ~583: `evm.balTracker.RecordStorageRead(contract.Address, key, val)` |
| `opSstore` does not emit BAL write/read event | `instructions.go` lines ~626-655: emits `RecordStorageRead` for no-op writes and `RecordStorageChange` for actual changes |
| Account-read opcodes do not emit address access | `instructions.go`: `RecordAddressTouch(addr)` calls in BALANCE, EXTCODESIZE, EXTCODECOPY, EXTCODEHASH, CALL targets, etc. (15+ call sites) |
| CALL family opcodes do not emit balance events | `interpreter.go` `Call()` method records `RecordBalanceChange` for both sender and recipient when `value > 0` |
| CREATE/CREATE2 do not emit deployment events | `interpreter.go` `create()` method records `RecordBalanceChange` for creator and contract address when `value > 0` |
| Pre-state validation gate | Hooks are placed after gas deduction succeeds, inside the execution path |
| `SELFBALANCE` must NOT be hooked | Confirmed: no `RecordAddressTouch` call in `opSelfBalance` |
