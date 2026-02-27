# Story 15.2 — Spec vector event fixture (`buildSpecVectorEvents` helper)

> **Sprint context:** Sprint 15 — Spurious Entry Validation & Spec Test Vectors
> **Sprint Goal:** The validator catches spurious BAL entries (index > n+1); a golden test reproduces the concrete example from the EIP spec exactly.

**Spec reference:** Lines 406-511. The EIP provides an exact BAL structure for a 2-transaction block.

**Files:**
- Create: `pkg/bal/spec_vector_test.go` (helpers only; test in 15.3)

**Acceptance Criteria:** `buildSpecVectorEvents` correctly simulates all tracker events for the spec's concrete block example (pre-execution index 0, tx indices 1-2, post-execution index 3); `mustAddr` and `findEntry` helpers compile and are callable from tests.

#### Task 15.2.1 — Write the fixture helpers

File: `pkg/bal/spec_vector_test.go`

```go
package bal_test

import (
    "encoding/hex"
    "testing"
)

// mustAddr parses a hex address string; panics if invalid.
func mustAddr(s string) [20]byte {
    b, err := hex.DecodeString(s[2:]) // strip "0x"
    if err != nil || len(b) != 20 {
        panic("invalid addr: " + s)
    }
    var a [20]byte
    copy(a[:], b)
    return a
}

// findEntry returns the BAL entry for addr, or nil if not found.
func findEntry(bl *BlockAccessList, addr [20]byte) *AccessEntry {
    for i := range bl.Entries {
        if bl.Entries[i].Address == addr {
            return &bl.Entries[i]
        }
    }
    return nil
}

// buildSpecVectorEvents simulates the tracker events for EIP-7928 spec lines 406-511.
//
// Block structure:
//   Pre-execution (index=0): EIP-2935 stores parent hash at block hash contract
//   Tx 1 (index=1): Alice sends 1 ETH to Bob, checks 0x2222...
//   Tx 2 (index=2): Charlie calls factory, deploying new contract
//   Post-execution (index=3): Withdrawal of 100 ETH to Eve
func buildSpecVectorEvents(
    blockHashContract, alice, bob, checked,
    charlie, factory, deployed, coinbase, eve [20]byte,
) map[[20]byte]*AccountEvents {
    tr := NewBALAccessTracker()

    // Pre-execution index=0: EIP-2935 write to block hash contract
    tr.RecordStorageWrite(blockHashContract, [32]byte{0x01}, [32]byte{0xAB}, 0)

    // Tx 1 (index=1): Alice → Bob (ETH transfer) + address check
    tr.RecordAddressAccess(alice, 1)
    tr.RecordNonceChange(alice, 1, 1)
    tr.RecordBalanceChange(alice, [32]byte{0x10}, 1)
    tr.RecordAddressAccess(bob, 1)
    tr.RecordBalanceChange(bob, [32]byte{0x11}, 1)
    tr.RecordAddressAccess(checked, 1) // read-only EXTCODEHASH check
    tr.RecordBalanceChange(coinbase, [32]byte{0xCC}, 1)

    // Tx 2 (index=2): Charlie calls factory, deploys new contract
    tr.RecordAddressAccess(charlie, 2)
    tr.RecordNonceChange(charlie, 1, 2)
    tr.RecordBalanceChange(charlie, [32]byte{0x20}, 2)
    tr.RecordAddressAccess(factory, 2)
    tr.RecordAddressAccess(deployed, 2)
    tr.RecordNonceChange(deployed, 1, 2)
    tr.RecordCodeChange(deployed, []byte{0x60, 0x00}, 2)
    tr.RecordBalanceChange(coinbase, [32]byte{0xCD}, 2)

    // Post-execution index=3: EIP-4895 withdrawal to Eve
    tr.RecordAddressAccess(eve, 3)
    tr.RecordBalanceChange(eve, [32]byte{0x64}, 3)

    return tr.Drain()
}
```

**Step: Verify helpers compile**

```
cd /projects/eth2030/pkg && go build ./bal/...
```

Expected: No errors.

**Step: Commit**

```bash
git add pkg/bal/spec_vector_test.go
git commit -m "test(bal): spec vector fixture helpers for EIP-7928 golden test"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
Lines 406-511 (Concrete Example):

Example block:

Pre-execution (index=0): EIP-2935 stores parent hash at block hash contract (0x0000F90827F1C53a10cb7A02335B175320002935)
Tx 1 (index=1): Alice (0xaaaa...) sends 1 ETH to Bob (0xbbbb...), checks balance of 0x2222...
Tx 2 (index=2): Charlie (0xcccc...) calls factory (0xffff...) deploying contract at 0xdddd...
Post-execution (index=3): Withdrawal of 100 ETH to Eve (0xabcd...)

Note: Pre-execution system contract uses block_access_index = 0.
      Post-execution withdrawal uses block_access_index = 3 (len(transactions) + 1)

Addresses sorted lexicographically in the resulting BAL:
  0x0000F908... (block hash contract) — storage_change at index 0
  0x2222...    (checked address)       — all empty lists
  0xaaaa...    (Alice)                 — balance_change + nonce_change at index 1
  0xabcd...    (Eve)                   — balance_change at index 3
  0xbbbb...    (Bob)                   — balance_change at index 1
  0xcccc...    (Charlie)               — balance_change + nonce_change at index 2
  0xdddd...    (deployed contract)     — nonce_change + code_change at index 2
  0xeeee...    (COINBASE)              — balance_changes at indices 1 and 2
  0xffff...    (factory)               — storage_change + nonce_change at index 2
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/bal/tracker.go` | `AccessTracker` records storage reads/changes and balance/nonce/code changes; `Build(txIndex)` stamps all with one index — cannot record multi-index events |
| `pkg/bal/types.go` | `BlockAccessList`, `AccessEntry`, `StorageAccess`, `StorageChange`, `BalanceChange`, `NonceChange`, `CodeChange`; no `BlockAccessIndex` field on change structs |
| `pkg/bal/types_extended.go` | `DetailedAccessEntry`, `ConflictMatrix`, `DependencyGraph`; no `AccountEvents` or `BALAccessTracker` type |
| `pkg/bal/hash.go` | `EncodeRLP()` and `Hash()` on `BlockAccessList`; no `BuildFromEvents` function |

---

## Implementation Assessment

### Current Status

Not implemented.

### Architecture Notes

The fixture helper file `pkg/bal/spec_vector_test.go` does not exist. The plan references three functions that are absent from the current `pkg/bal` package: `NewBALAccessTracker()`, `BuildFromEvents()`, and the `AccountEvents` type. The existing `AccessTracker` in `tracker.go` is a single-phase tracker — it accepts one `txIndex` at `Build` time and stamps every recorded event with that same index. It cannot represent the multi-index event model required by the spec (pre-execution at index 0, per-transaction indices 1..n, post-execution at index n+1). `mustAddr` and `findEntry` are purely test-side helpers and can be written in the `_test.go` file without modifying production code, but the three production-side symbols must be created before the test file can compile.

### Gaps and Proposed Solutions

1. **`NewBALAccessTracker` does not exist.** A new tracker type (distinct from `AccessTracker`) must be added to `pkg/bal/` that can accept a `blockAccessIndex uint16` alongside each recorded event, accumulating a per-address map of indexed changes matching the spec's `AccountChanges` structure.

2. **`AccountEvents` type is absent.** Define a struct (or use the existing `AccessEntry` extended with per-change index slices) to hold all events recorded for one address across all phases. The `Drain()` method on the new tracker must return `map[[20]byte]*AccountEvents`.

3. **`BuildFromEvents` does not exist.** This function converts the `map[[20]byte]*AccountEvents` returned by `Drain()` into a sorted `*BlockAccessList`. It must sort addresses lexicographically and populate the multi-index change slices on each `AccessEntry`.

4. **`spec_vector_test.go` file is absent.** Once the three production symbols above exist, create the file in `package bal_test` with `mustAddr`, `findEntry`, and `buildSpecVectorEvents` as shown in the plan.
