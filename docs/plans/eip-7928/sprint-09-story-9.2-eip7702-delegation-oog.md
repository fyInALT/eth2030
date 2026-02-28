# Story 9.2 — EIP-7702 Delegation Access-Cost Failure

> **Sprint context:** Sprint 9 — Two-Phase Gas Validation (BAL Inclusion Gate)
> **Sprint Goal:** Opcodes that fail pre-state gas validation MUST NOT add the target address/slot to the BAL. This is a spec requirement (lines 131–176): pre-state validation must pass before any state access occurs.

**Spec reference:** Lines 168–172. "If this check fails, the delegated address MUST NOT appear in the BAL, though the original call target is included."

**Files:**
- Modify: `pkg/core/eip7702.go`
- Test: `pkg/core/eip7702_bal_test.go`

**Acceptance Criteria:** When a call resolves a delegation but then fails the `access_cost` check for the delegated address, the delegated address is absent from the BAL; the call target (the authority) is present.

#### Task 9.2.1 — Write failing test

```go
func TestEIP7702_DelegatedAddressOOG_ExcludedFromBAL(t *testing.T) {
    // Authority has delegation to contractAddr
    // Call with exactly COLD_ACCOUNT_ACCESS_COST - 1 gas remaining after resolving authority
    // Assert: authority address IS in BAL (was accessed for delegation resolution)
    // Assert: contractAddr is NOT in BAL (access_cost check failed)
}
```

#### Task 9.2.2 — Implement

In the EVM's call handler, when resolving a 7702 delegation:

```go
// The call target (authority) was accessed to resolve delegation → include
evm.AccessTracker.RecordAddressAccess(callTarget.Bytes20(), evm.TxIndex)

// Now check access_cost for the delegated address
if !scope.Contract.UseGas(delegatedAccessCost) {
    // delegated address MUST NOT be included
    return nil, ErrOutOfGas
}
// Access_cost paid → include delegated address
evm.AccessTracker.RecordAddressAccess(delegatedAddr.Bytes20(), evm.TxIndex)
```

**Step: Run tests**

```
cd /projects/eth2030/pkg && go test ./core/... -run TestEIP7702_Delegated -v
```

Expected: PASS.

**Step: Commit**

```bash
git add pkg/core/eip7702.go pkg/core/eip7702_bal_test.go
git commit -m "feat(bal): EIP-7702 delegated addr excluded if access_cost OOG"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
#### EIP-7702 Delegation

When a call target has an EIP-7702 delegation, the target is accessed to resolve the delegation. If a delegation exists, the delegated address requires its own `access_cost` check before being accessed. If this check fails, the delegated address MUST NOT appear in the BAL, though the original call target is included (having been accessed to resolve the delegation).

Note: Delegated accounts cannot be empty, so `GAS_NEW_ACCOUNT` never applies when resolving delegations.
```

And from the Edge Cases section:

```
- **EIP-7702 Delegations:** The authority address MUST be included with nonce and code changes after any successful delegation set, update, or clear. If authorization fails after the authority address has been loaded and added to `accessed_addresses` (per EIP-2929), it MUST still be included with an empty change set; if authorization fails before the authority is loaded, it MUST NOT be included. The delegation target MUST NOT be included during delegation creation or modification and MUST only be included once it is actually loaded as an execution target (e.g., via `CALL`/`CALLCODE`/`DELEGATECALL`/`STATICCALL` under authority execution).
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/eip7702.go` | `ProcessAuthorizations`, `processOneAuthorization`, `RecoverAuthority`, `ResolveDelegation`, `IsDelegated` — full EIP-7702 authorization processing at tx application time |
| `pkg/core/vm/evm_call_handlers.go` | `CallHandler.HandleCall`: precompile routing and code execution; `WarmTarget` uses `StateDB.AddAddressToAccessList`; does not resolve EIP-7702 delegations |
| `pkg/core/vm/instructions.go` | `opCall`, `opCallCode`, `opDelegateCall`, `opStaticCall`: dispatch to `evm.Call` etc.; no delegation resolution here |
| `pkg/core/vm/interpreter.go` | `gasEIP2929AccountCheck`: warms address before `UseGas`; called from `dynamicGas` functions for CALL-family opcodes |
| `pkg/core/vm/gas_table.go` | `gasCallEIP2929`, `gasCallCodeEIP2929`, `gasDelegateCallEIP2929`, `gasStaticCallEIP2929`: charge `access_cost` for the call target |
| `pkg/core/processor.go` | `applyTransaction`: calls `ProcessAuthorizations` for SetCode txs and then executes the EVM; `capturePreState` tracks sender/recipient only |
| `pkg/bal/tracker.go` | `AccessTracker`: no delegation-specific recording hooks exist |

---

## Implementation Assessment

### Current Status

Not implemented.

### Architecture Notes

EIP-7702 delegation processing in this codebase occurs in two separate places with different responsibilities:

1. **Authorization processing** (`pkg/core/eip7702.go`): `ProcessAuthorizations` runs during transaction application (before EVM execution). It verifies chain ID, nonce, and signature, then calls `statedb.SetCode(signerAddr, delegationCode)` and increments the signer nonce. This handles delegation *creation*. No BAL events are emitted here.

2. **Call-time delegation resolution**: When the EVM executes a CALL/CALLCODE/DELEGATECALL/STATICCALL targeting an account that has a delegation designator as its code, the delegated address should be resolved and accessed. However, `evm_call_handlers.go`'s `CallHandler.HandleCall` calls `ch.evm.StateDB.GetCode(codeAddr)` to retrieve code, but does not check for the `0xef0100` delegation prefix or perform a separate `access_cost` check for the delegated address. The `IsDelegated` / `ResolveDelegation` helpers in `eip7702.go` exist but are not wired into the EVM call path.

3. **`gasEIP2929AccountCheck` warms the call target inside `dynamicGas`** (before `UseGas` is called), so the call target is always warmed regardless of whether gas suffices. No separate `access_cost` check for the delegated address is performed at all.

4. The BAL tracker (`pkg/bal/tracker.go`) has no `RecordAddressAccess` method. The BAL is populated by `populateTracker` using a pre/post snapshot comparison restricted to the tx sender and recipient; it cannot observe delegation resolution access patterns.

### Gaps and Proposed Solutions

1. **No delegation resolution in the EVM call path**: The EVM call handler does not inspect code for the `0xef0100` prefix or call `ResolveDelegation`. Solution: in `HandleCall` (or the `Call`/`CallCode` methods on the EVM), after retrieving the code for `codeAddr`, check `IsDelegated(code)`. If true, record the call target (authority) in the BAL (it was accessed), then perform a separate `access_cost` gas check for the delegated address before including it in the BAL and loading its code.

2. **No `access_cost` gate for the delegated address**: The spec requires a fresh `access_cost` check for the delegated address (warm=100, cold=2600) after the call target has been accessed. If this check fails OOG, the delegated address must not appear in the BAL. Currently no such check exists anywhere in the call path.

3. **`bal.AccessTracker` lacks `RecordAddressAccess`**: The tracker has `RecordBalanceChange`, `RecordStorageRead`, etc., but no method for a pure address touch (no state change). A new `RecordAddressAccess(addr)` method must be added to `AccessTracker` that adds the address to `touchedAddrs` without recording any change, so that empty-change-list entries appear in the BAL for delegation targets.

4. **Authorization-time BAL recording missing**: When `ProcessAuthorizations` successfully sets delegation code and increments nonce, the authority address must appear in the BAL with nonce and code changes. Currently `populateTracker` is called with only sender/recipient pre-state; it misses the authority addresses in the authorization list.
