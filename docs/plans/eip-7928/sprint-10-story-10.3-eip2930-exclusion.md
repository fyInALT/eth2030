# Story 10.3 — EIP-2930 Access List Entries NOT Auto-Included

> **Sprint context:** Sprint 10 — Special Address Tracking
> **Sprint Goal:** COINBASE, precompiles, EIP-2930 exclusion, and the SYSTEM_ADDRESS rule are all handled correctly by the tracker.

**Spec reference:** Line 112. "Entries from an EIP-2930 access list MUST NOT be included automatically. Only addresses and storage slots that are actually touched or changed during execution are recorded."

**Files:**
- Modify: `pkg/core/processor.go` (EIP-2930 warming path)
- Test: `pkg/core/eip2930_bal_test.go`

**Acceptance Criteria:** A type-1 transaction with an address in its access list does NOT produce a BAL entry for that address unless it is actually accessed during execution.

#### Task 10.3.1 — Write failing test

```go
func TestBAL_EIP2930AccessList_NotAutoIncluded(t *testing.T) {
    // Build a type-1 tx with access list containing address 0xdeadbeef...
    // That address is never touched during execution
    // Assert 0xdeadbeef... is absent from BAL
}

func TestBAL_EIP2930AccessList_ActuallyTouched_Included(t *testing.T) {
    // Build type-1 tx with access list containing address X
    // Execution calls EXTCODEHASH on X
    // Assert X IS present in BAL (because actually accessed)
}
```

#### Task 10.3.2 — Ensure warming does not emit BAL events

In the EIP-2930 warming code (where `stateDB.AddAddressToAccessList` is called before execution), do NOT call `tracker.RecordAddressAccess`. The call to `RecordAddressAccess` happens only in the opcode handlers when the address is actually accessed during execution.

```go
// Warm up access list addresses (EIP-2929 gas optimization)
// EIP-7928: warming MUST NOT emit BAL events — only actual execution does
for _, addr := range tx.AccessList() {
    stateDB.AddAddressToAccessList(addr.Address) // gas warming only
    // DO NOT call tracker.RecordAddressAccess here
}
```

**Step: Run tests**

```
cd /projects/eth2030/pkg && go test ./core/... -run TestBAL_EIP2930 -v
```

Expected: PASS.

**Step: Commit**

```bash
git add pkg/core/processor.go pkg/core/eip2930_bal_test.go
git commit -m "feat(bal): EIP-2930 entries excluded from BAL unless actually accessed"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
Line 112:
Entries from an EIP-2930 access list MUST NOT be included automatically.
Only addresses and storage slots that are actually touched or changed during
execution are recorded.
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/vm/access_list_tracker.go` | `PrePopulate` method warms EIP-2930 access list entries for gas purposes (lines 68-73); no BAL event is emitted here — this is the correct separation point |
| `pkg/core/processor.go` | Lines 876-881: EIP-2930 warming loop calls `statedb.AddAddressToAccessList` and `statedb.AddSlotToAccessList` directly on the state DB — no BAL tracker call is present here either |
| `pkg/bal/tracker.go` | `RecordStorageRead` / `RecordStorageChange` are the BAL emission points; they are called only from `populateTracker` in `processor.go` (post-execution diff), not from the warming path |

---

## Implementation Assessment

### Current Status

Partially implemented. The EIP-2930 warming path in `processor.go` does NOT currently call any BAL tracker method, which is correct behaviour. However, the `populateTracker` function only compares pre/post balance and nonce; it does not track which addresses were actually accessed by opcode handlers during execution. As a result the BAL cannot yet distinguish "EIP-2930 warmed but never accessed" from "actually accessed" for addresses that have no balance/nonce change.

### Architecture Notes

The BAL tracker (`pkg/bal/tracker.go`) has `RecordStorageRead`, `RecordStorageChange`, `RecordBalanceChange`, `RecordNonceChange`, `RecordCodeChange`, and `RecordAddressTouch` (line 74). The `RecordAddressTouch` method adds an address to `touchedAddrs` without recording any change. The `BALTracker` interface in `pkg/core/vm/bal_tracker.go` also includes `RecordAddressTouch`. The EVM's opcode handlers in `instructions.go` call `evm.balTracker.RecordAddressTouch(addr)` for BALANCE, EXTCODESIZE, EXTCODECOPY, EXTCODEHASH, and CALL-family targets (15+ call sites). The `AccessListTracker` in `pkg/core/vm/access_list_tracker.go` handles EIP-2929 warm/cold gas tracking only; it has no connection to the BAL.

### Gaps and Proposed Solutions

1. **Opcode-level address access recording exists**: `RecordAddressTouch(addr types.Address)` exists on `bal.AccessTracker` at line 74 and is called from 15+ opcode handlers in `pkg/core/vm/instructions.go` via the `BALTracker` interface. Addresses touched by BALANCE, EXTCODESIZE, EXTCODEHASH, STATICCALL, etc. are recorded. This gap is resolved.

2. **EIP-2930 exclusion is already correctly not emitted**: The warming loop at `processor.go:876-881` does not call the BAL tracker — this part is already correct and just needs to be guarded by a comment or test to document the invariant.

3. **Missing test coverage**: `pkg/core/eip2930_bal_test.go` does not yet exist. The two tests described in the story (`TestBAL_EIP2930AccessList_NotAutoIncluded` and `TestBAL_EIP2930AccessList_ActuallyTouched_Included`) must be written once opcode-level address recording is wired up.
