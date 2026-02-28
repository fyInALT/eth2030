# Story 5.1 — Track EIP-7702 delegation in `AccessTracker`

> **Sprint context:** Sprint 5 — EIP-7702 Compatibility
> **Sprint Goal:** SetCode transactions (type 0x04) are correctly tracked in the BAL — both the authority address (nonce + code changes) and delegation target address (read event) appear in the right entries.

**Files:**
- Modify: `pkg/core/eip7702.go`
- Test: `pkg/core/eip7702_bal_test.go`

**Acceptance Criteria:** After processing a type-4 transaction, the BAL contains a nonce change and code change for the authority address; the delegation target appears in the BAL only when called (not on delegation setup).

#### Task 5.1.1 — Write failing tests

File: `pkg/core/eip7702_bal_test.go`

```go
package core_test

import "testing"

func TestEIP7702_AuthorityInBAL(t *testing.T) {
	// Build a SetCode transaction (type 0x04)
	// Execute via processor
	// Assert BAL has the authority address with nonce+code change
}

func TestEIP7702_DelegationTargetNotInBAL_UnlessCalledm(t *testing.T) {
	// Build SetCode tx that delegates to contract but doesn't call it
	// Execute
	// Assert delegation target address is NOT in BAL
}
```

#### Task 5.1.2 — Emit BAL events in `eip7702.go`

In `pkg/core/eip7702.go`, in the authorization processing loop, after successful delegation:

```go
// EIP-7928: authority address must appear with nonce + code change
tracker.RecordNonceChange(authority.Bytes20(), postNonce, txIndex)
tracker.RecordCodeChange(authority.Bytes20(), delegationCode, txIndex)
// authority address accessed (per EIP-2929 accessed_addresses)
tracker.RecordAddressAccess(authority.Bytes20(), txIndex)
```

If authorization fails after the authority is loaded (per EIP-7928 edge case):

```go
// still include authority with empty changes if already loaded
tracker.RecordAddressAccess(authority.Bytes20(), txIndex)
```

Run: `cd /projects/eth2030/pkg && go test ./core/... -run TestEIP7702 -v`
Expected: PASS.

**Step: Format & commit**

```bash
cd /projects/eth2030/pkg && go fmt ./core/...
git add pkg/core/eip7702.go pkg/core/eip7702_bal_test.go
git commit -m "feat(bal): EIP-7702 authority tracked in BAL"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
#### EIP-7702 Delegation

When a call target has an EIP-7702 delegation, the target is accessed to
resolve the delegation. If a delegation exists, the delegated address requires
its own `access_cost` check before being accessed. If this check fails, the
delegated address MUST NOT appear in the BAL, though the original call target
is included (having been accessed to resolve the delegation).

Note: Delegated accounts cannot be empty, so `GAS_NEW_ACCOUNT` never applies
when resolving delegations.

#### Edge Cases (Normative)

- **EIP-7702 Delegations:** The authority address MUST be included with nonce
  and code changes after any successful delegation set, update, or clear. If
  authorization fails after the authority address has been loaded and added to
  `accessed_addresses` (per EIP-2929), it MUST still be included with an empty
  change set; if authorization fails before the authority is loaded, it MUST NOT
  be included. The delegation target MUST NOT be included during delegation
  creation or modification and MUST only be included once it is actually loaded
  as an execution target (e.g., via CALL/CALLCODE/DELEGATECALL/STATICCALL under
  authority execution).
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/eip7702.go` | Contains `ProcessAuthorizations` and `processOneAuthorization` — the entry point for SetCode tx processing; no BAL emission calls present |
| `pkg/bal/tracker.go` | `AccessTracker` with `RecordNonceChange`, `RecordCodeChange`; tracker API exists but is not called from `eip7702.go` |
| `pkg/bal/types.go` | `AccessEntry` struct — holds `NonceChange *NonceChange` and `CodeChange *CodeChange` pointer fields (single change per entry, not a list) |
| `pkg/core/eip7702_test.go` | Existing tests for authorization logic; no BAL-specific assertions |

---

## Implementation Assessment

### Current Status

Not implemented.

### Architecture Notes

`pkg/core/eip7702.go` — specifically `processOneAuthorization` — correctly calls `statedb.SetCode` and `statedb.SetNonce` on the authority address after a successful delegation (lines 69–72). However, there is no `AccessTracker` parameter in the function signature and no BAL recording calls anywhere in the file. The `AccessTracker` in `pkg/bal/tracker.go` exposes `RecordNonceChange(addr, oldNonce, newNonce uint64)` and `RecordCodeChange(addr, oldCode, newCode []byte)`, but these require both old and new values — the plan's pseudocode passes only `postNonce` and `delegationCode` without old values, so the call signatures in the story must be adjusted to match `tracker.go`.

Additionally, the current tracker design stores one change per address (a `map[addrKey]*NonceChange`), overwriting previous entries. The spec requires a list of `NonceChange` entries with `BlockAccessIndex` — for EIP-7702 this means the authority nonce and code changes from the SetCode transaction must be tagged with the transaction's index `i+1`. The current `AccessEntry.NonceChange` is a single pointer, not a slice, which diverges from the spec's per-index list model.

The plan calls `tracker.RecordAddressAccess` for the failed-authorization case. The equivalent method `RecordAddressTouch(addr types.Address)` exists at `pkg/bal/tracker.go` line 74, adding the address to `touchedAddrs` without recording any change. It is also part of the `BALTracker` interface in `pkg/core/vm/bal_tracker.go`.

### Gaps and Proposed Solutions

1. **No BAL emission in `processOneAuthorization`**: Add an `*AccessTracker` parameter (or accept a callback) to `ProcessAuthorizations` / `processOneAuthorization`. After `statedb.SetCode` / `statedb.SetNonce`, call `tracker.RecordNonceChange` and `tracker.RecordCodeChange`. On failed-after-load authorization, call `tracker.RecordAddressAccess` (which needs to be added as a new method that merely adds the address to `touchedAddrs`).

2. **`RecordAddressTouch` method exists**: `func (t *AccessTracker) RecordAddressTouch(addr types.Address)` already exists at `pkg/bal/tracker.go` line 74. It inserts into `t.touchedAddrs` without recording any field change. This gap is resolved.

3. **Single-change-per-address limitation**: The `AccessEntry` structs use single-pointer fields (`NonceChange *NonceChange`) rather than slices. For EIP-7702 the authority address may have nonce and code changes from the SetCode tx and then additional changes if the delegated code executes. A deeper refactor to per-index slices is needed for full spec compliance; for this story, the minimal fix of recording the changes at the correct `txIndex` is sufficient.

4. **Tests**: `pkg/core/eip7702_bal_test.go` does not exist. It must be created with tests that wire a tracker into the processor and assert BAL contents after SetCode transaction execution.
