# Story 6.1 — Implement `ApplyBAL` state reconstruction

> **Sprint context:** Sprint 6 — State Reconstruction (Executionless Updates)
> **Sprint Goal:** Given a `BlockAccessList`, a client can apply all state changes to a stateDB without re-executing any transactions, enabling fast sync and stateless client support.

**Files:**
- Create: `pkg/bal/apply.go`
- Test: `pkg/bal/apply_test.go`

**Acceptance Criteria:** `ApplyBAL(stateDB StateWriter, bl *BlockAccessList)` correctly applies all balance, nonce, code, and storage changes from the BAL to a fresh state; resulting state root matches the root produced by full re-execution.

#### Task 6.1.1 — Write failing test

File: `pkg/bal/apply_test.go`

```go
package bal_test

import "testing"

func TestApplyBAL_MatchesExecutionStateRoot(t *testing.T) {
	// 1. Execute block normally -> stateRoot1
	// 2. Start fresh state
	// 3. ApplyBAL(freshState, bal) -> stateRoot2
	// 4. Assert stateRoot1 == stateRoot2
}
```

Run: `cd /projects/eth2030/pkg && go test ./bal/... -run TestApplyBAL -v`
Expected: FAIL.

#### Task 6.1.2 — Implement `ApplyBAL`

File: `pkg/bal/apply.go`

```go
package bal

import "math/big"

// StateWriter is the minimal interface needed to reconstruct state from a BAL.
type StateWriter interface {
	SetBalance(addr [20]byte, amount *big.Int)
	SetNonce(addr [20]byte, nonce uint64)
	SetCode(addr [20]byte, code []byte)
	SetState(addr [20]byte, key [32]byte, value [32]byte)
}

// ApplyBAL applies all state changes from a BlockAccessList to a StateWriter.
// Only post-execution values (highest block_access_index per field) are applied.
func ApplyBAL(state StateWriter, bl *BlockAccessList) {
	for _, entry := range bl.Entries {
		// Apply final balance (highest txIndex)
		if n := len(entry.BalanceChanges); n > 0 {
			last := entry.BalanceChanges[n-1]
			bal := new(big.Int).SetBytes(last.PostBalance[:])
			state.SetBalance(entry.Address, bal)
		}
		// Apply final nonce
		if n := len(entry.NonceChanges); n > 0 {
			state.SetNonce(entry.Address, entry.NonceChanges[n-1].NewNonce)
		}
		// Apply final code
		if n := len(entry.CodeChanges); n > 0 {
			state.SetCode(entry.Address, entry.CodeChanges[n-1].Code)
		}
		// Apply all storage writes (final value per slot)
		for _, sc := range entry.StorageChanges {
			if len(sc.Changes) > 0 {
				last := sc.Changes[len(sc.Changes)-1]
				state.SetState(entry.Address, sc.Slot, last.Value)
			}
		}
	}
}
```

Run: `cd /projects/eth2030/pkg && go test ./bal/... -run TestApplyBAL -v`
Expected: PASS.

**Step: Format & commit**

```bash
cd /projects/eth2030/pkg && go fmt ./bal/...
git add pkg/bal/apply.go pkg/bal/apply_test.go
git commit -m "feat(bal): ApplyBAL for executionless state reconstruction"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
## Motivation

This EIP introduces Block-Level Access Lists (BALs) that record all accounts
and storage locations accessed during block execution, along with their
post-execution values. BALs enable parallel disk reads, parallel transaction
validation, parallel state root computation and executionless state updates.

This proposal enforces access lists at the block level, enabling:

- Parallel disk reads and transaction execution
- Parallel post-state root calculation
- State reconstruction without executing transactions
- Reduced execution time to `parallel IO + parallel EVM`

### BAL Design Choice

2. **Storage values for writes**: Post-execution values enable state
   reconstruction during sync without individual proofs against state root.

### State Transition Function (excerpt)

def build_bal(accesses):
    for addr in sorted(accesses.keys()):
        data = accesses[addr]
        bal.append([
            addr,
            storage_changes,          # final new_value per slot per index
            sorted(list(data['storage_reads'])),
            sorted(data['balance_changes']),   # [(index, post_balance), ...]
            sorted(data['nonce_changes']),      # [(index, new_nonce), ...]
            sorted(data['code_changes'])        # [(index, new_code), ...]
        ])
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/bal/apply.go` | Does not exist — the file to be created by this story |
| `pkg/bal/apply_test.go` | Does not exist — the test file to be created by this story |
| `pkg/bal/types.go` | Defines `BlockAccessList`, `AccessEntry`, `BalanceChange`, `NonceChange`, `CodeChange`, `StorageChange` — the input types for `ApplyBAL` |
| `pkg/bal/tracker.go` | `AccessTracker.Build(txIndex)` — produces a `BlockAccessList` from recorded accesses; the format that `ApplyBAL` will consume |
| `pkg/core/bal_integration_test.go` | Has `TestProcessWithBAL_WithTransactions` and related tests that already call `proc.ProcessWithBAL` and inspect `result.BlockAccessList`; provides a test harness that `apply_test.go` can reuse |

---

## Implementation Assessment

### Current Status

Not implemented. Neither `pkg/bal/apply.go` nor `pkg/bal/apply_test.go` exist.

### Architecture Notes

The plan's `ApplyBAL` implementation is semantically correct for the current `types.go` data model, but there are type-level mismatches that must be resolved before it can compile:

1. **`entry.BalanceChanges` (plural) does not exist**: `AccessEntry.BalanceChange` is a single `*BalanceChange` pointer, not a slice. The plan loops over `entry.BalanceChanges[n-1]` as if it were a slice; the correct access is `entry.BalanceChange` (singular, check non-nil).

2. **`entry.NonceChanges` and `entry.CodeChanges` (plural) do not exist**: Likewise, `AccessEntry.NonceChange *NonceChange` and `AccessEntry.CodeChange *CodeChange` are single pointers, not slices.

3. **`BalanceChange.PostBalance` field does not exist**: The `BalanceChange` struct has `OldValue *big.Int` and `NewValue *big.Int`. The plan references `last.PostBalance[:]`; the correct field is `entry.BalanceChange.NewValue`.

4. **`StorageChange.Changes` slice does not exist**: `StorageChange` has `Slot types.Hash`, `OldValue types.Hash`, and `NewValue types.Hash` directly, not a nested slice of changes. The plan loops `sc.Changes[len(sc.Changes)-1]`; the correct access is `sc.NewValue` directly.

5. **`StateWriter.SetBalance` signature**: The plan's `StateWriter` takes `*big.Int`, which is consistent with `BalanceChange.NewValue *big.Int`.

The current `AccessTracker.Build` collapses all changes for a given address into a single `AccessEntry` (one balance change, one nonce change, one code change). The spec requires per-`BlockAccessIndex` lists so that `ApplyBAL` applies the final value — the highest index — but with the current single-pointer model the final value is already the only stored value.

### Gaps and Proposed Solutions

1. **Create `pkg/bal/apply.go`** with a `StateWriter` interface and an `ApplyBAL` function adapted to the actual `AccessEntry` field names: read `entry.BalanceChange.NewValue`, `entry.NonceChange.NewValue`, `entry.CodeChange.NewCode`, and `sc.NewValue` for storage changes.

2. **Create `pkg/bal/apply_test.go`** that executes a block via `ProcessWithBAL`, captures the `BlockAccessList`, then calls `ApplyBAL` on a fresh state and asserts the resulting state root matches. The test can reuse the helpers already established in `pkg/core/bal_integration_test.go` (or define a self-contained test within `pkg/bal/`).

3. **Multi-index support (future)**: When the tracker is refactored to store per-`BlockAccessIndex` slices (as the spec requires), `ApplyBAL` will need to select the entry with the highest index per field. This story's implementation should be written so that such a refactor only requires changing the field access pattern, not the overall algorithm structure.
