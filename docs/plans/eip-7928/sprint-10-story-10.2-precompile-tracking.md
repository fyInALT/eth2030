# Story 10.2 — Precompiled Contract Tracking

> **Sprint context:** Sprint 10 — Special Address Tracking
> **Sprint Goal:** COINBASE, precompiles, EIP-2930 exclusion, and the SYSTEM_ADDRESS rule are all handled correctly by the tracker.

**Spec reference:** Line 108, 251. "Precompiles MUST be included when accessed. If a precompile receives value, it is recorded with a balance change. Otherwise, it is included with empty change lists."

**Files:**
- Modify: `pkg/core/vm/evm.go` (precompile call path)
- Test: `pkg/core/vm/precompile_bal_test.go`

**Acceptance Criteria:** Calling a precompile (e.g., ecrecover at 0x01) produces an address entry in BAL; if value is sent, a balance_change entry is also produced.

#### Task 10.2.1 — Write failing tests

```go
func TestBAL_Precompile_NoValue_EmptyEntry(t *testing.T) {
    // Call ecrecover (0x01) with no value
    // Assert 0x01 appears in BAL with all empty lists
}

func TestBAL_Precompile_WithValue_BalanceChange(t *testing.T) {
    // Call sha256 (0x02) with 1 wei value
    // Assert 0x02 appears in BAL with balance_changes entry
}
```

#### Task 10.2.2 — Emit events in the precompile call path

In `pkg/core/vm/evm.go`, in the precompile execution branch:

```go
// EIP-7928: precompiles are always included in BAL when accessed
evm.AccessTracker.RecordAddressAccess(addr.Bytes20(), evm.TxIndex)

if value != nil && value.Sign() > 0 {
    postBalance := stateDB.GetBalance(addr)
    evm.AccessTracker.RecordBalanceChange(addr.Bytes20(), uint256ToBytes32(postBalance), evm.TxIndex)
}
```

**Step: Run tests**

```
cd /projects/eth2030/pkg && go test ./core/vm/... -run TestBAL_Precompile -v
```

Expected: PASS.

**Step: Commit**

```bash
git add pkg/core/vm/evm.go pkg/core/vm/precompile_bal_test.go
git commit -m "feat(bal): track precompile calls in BAL"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
    - Precompiled contracts when called or accessed
```

From the Edge Cases section:

```
- **Precompiled contracts:** Precompiles MUST be included when accessed. If a precompile receives value, it is recorded with a balance change. Otherwise, it is included with empty change lists.
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/vm/evm_call_handlers.go` | `CallHandler.HandleCall`: detects precompiles via `ch.evm.precompile(params.Target)` and dispatches to `ch.runPrecompile`; no BAL emission here |
| `pkg/core/vm/evm_call_handlers.go` | `runPrecompile`: executes the precompile, charges gas, returns result; no `RecordAddressAccess` call |
| `pkg/core/vm/interpreter.go` | `warmPrecompiles`: adds all 19 precompile addresses (0x01–0x13) to the state access list at transaction start via `evm.StateDB.AddAddressToAccessList`; does not emit BAL events |
| `pkg/core/vm/access_list_tracker.go` | `PrePopulate`: also pre-warms precompile addresses (0x01–0x13) without journal; this is EIP-2929 warm tracking only, not BAL recording |
| `pkg/bal/tracker.go` | `AccessTracker`: `RecordBalanceChange`, `RecordAddressTouch` (line 74), `touchedAddrs`; `RecordAddressTouch` exists for pure address-touch |

---

## Implementation Assessment

### Current Status

Not implemented.

### Architecture Notes

Precompile addresses (0x01–0x13) are pre-warmed in the EIP-2929 access list at the start of every transaction by `warmPrecompiles` in `interpreter.go` (line 757) and by `AccessListTracker.PrePopulate` in `access_list_tracker.go` (line 63). This warming ensures cold gas is not charged for precompile calls, but it is purely an EIP-2929 concern and has no connection to the BAL.

When a precompile is actually called, `CallHandler.HandleCall` in `evm_call_handlers.go` detects it at line 74 (`ch.evm.precompile(params.Target)`) and calls `runPrecompile`. The `runPrecompile` function charges gas and executes the precompile, but emits no BAL events of any kind — no `RecordAddressAccess` and no `RecordBalanceChange`.

The BAL population path (`populateTracker` in `processor.go`) uses pre/post balance comparisons restricted to the tx sender and recipient. A precompile receiving value (e.g., `sha256` at `0x02` called with 1 wei) would have its balance changed via `statedb.AddBalance(params.Target, params.Value)` in `HandleCall` (line 98), but since the precompile address was not captured in `capturePreState`, the balance change is invisible to `populateTracker`.

There is also a discrepancy in target files: the story plan refers to `pkg/core/vm/evm.go`, but the actual precompile call path lives in `pkg/core/vm/evm_call_handlers.go`. There is no file named `evm.go` in `pkg/core/vm/`.

### Gaps and Proposed Solutions

1. **No BAL emission in `runPrecompile`**: When a precompile is called, `RecordAddressAccess` must be called for the precompile's address. Solution: add a BAL event recorder to the EVM (or pass it into `HandleCall`), and emit `RecordAddressAccess(addr)` inside `runPrecompile` after the gas check passes but before (or after) execution. If `params.Value.Sign() > 0`, also emit `RecordBalanceChange` for the precompile address using the post-execution balance.

2. **`bal.AccessTracker` has `RecordAddressTouch`**: The method `RecordAddressTouch(addr types.Address)` exists at `pkg/bal/tracker.go` line 74, adding the address to `touchedAddrs` without recording any change. It is also part of the `BALTracker` interface in `pkg/core/vm/bal_tracker.go`. The EVM's `Call()` path already calls `evm.balTracker.RecordAddressTouch` for precompile targets indirectly via the value-transfer path, but a direct call may be needed for zero-value precompile calls.

3. **`capturePreState` does not include precompile addresses**: Even if `populateTracker` were extended to detect value-bearing precompile calls, the pre-balance of the precompile is not captured. Precompile tracking requires either in-EVM instrumentation (in `runPrecompile`) or extending `capturePreState` to iterate over all call targets parsed from EVM bytecode — the former is far simpler.

4. **Wrong target file in the story plan**: The plan refers to `pkg/core/vm/evm.go` but the precompile dispatch lives in `pkg/core/vm/interpreter.go` (the `Call()` method handles precompile detection at line 409 via `evm.precompile(addr)` and runs them via `runPrecompile`). Note: there is no file named `evm.go` in `pkg/core/vm/`. The correct file to modify for precompile BAL tracking is `pkg/core/vm/interpreter.go`.
