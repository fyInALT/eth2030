# Story 15.1 — Spurious Entry Validation

> **Sprint context:** Sprint 15 — Spurious Entry Validation & Spec Test Vectors
> **Sprint Goal:** The validator catches spurious BAL entries (index > n+1); a golden test reproduces the concrete example from the EIP spec exactly.

**Spec reference:** Line 400. "Spurious entries MAY be detected by validating BAL indices, which MUST never be higher than `len(transactions) + 1`."

**Files:**
- Modify: `pkg/core/block_validator.go`
- Test: `pkg/core/spurious_entry_test.go`

**Acceptance Criteria:** A BAL containing an entry with `block_access_index > len(transactions)+1` causes `ValidateBlock` to return `ErrSpuriousBALEntry`.

#### Task 15.1.1 — Write failing test

```go
func TestBALValidator_SpuriousIndex_Rejected(t *testing.T) {
    // Block with 2 transactions → valid indices are 0, 1, 2, 3
    // Inject a BAL entry with block_access_index = 4
    // Assert ValidateBlock returns ErrSpuriousBALEntry
}

func TestBALValidator_ValidMaxIndex_Accepted(t *testing.T) {
    // Block with 2 transactions, BAL entry with index = 3 (n+1 = 2+1 = 3)
    // Assert ValidateBlock succeeds
}
```

#### Task 15.1.2 — Implement validation

In `pkg/core/block_validator.go`:

```go
var ErrSpuriousBALEntry = errors.New("BAL contains spurious entry with invalid block_access_index")

// validateBALIndices checks that no BAL entry has a block_access_index > len(txs)+1.
func validateBALIndices(bl *bal.BlockAccessList, txCount int) error {
    maxIndex := uint16(txCount + 1)
    for _, entry := range bl.Entries {
        for _, sc := range entry.StorageChanges {
            for _, c := range sc.Changes {
                if c.BlockAccessIndex > maxIndex {
                    return fmt.Errorf("%w: index %d > max %d", ErrSpuriousBALEntry, c.BlockAccessIndex, maxIndex)
                }
            }
        }
        for _, bc := range entry.BalanceChanges {
            if bc.BlockAccessIndex > maxIndex {
                return fmt.Errorf("%w: balance index %d > max %d", ErrSpuriousBALEntry, bc.BlockAccessIndex, maxIndex)
            }
        }
        for _, nc := range entry.NonceChanges {
            if nc.BlockAccessIndex > maxIndex {
                return fmt.Errorf("%w: nonce index %d > max %d", ErrSpuriousBALEntry, nc.BlockAccessIndex, maxIndex)
            }
        }
        for _, cc := range entry.CodeChanges {
            if cc.BlockAccessIndex > maxIndex {
                return fmt.Errorf("%w: code index %d > max %d", ErrSpuriousBALEntry, cc.BlockAccessIndex, maxIndex)
            }
        }
    }
    return nil
}
```

**Step: Run tests**

```
cd /projects/eth2030/pkg && go test ./core/... -run TestBALValidator_Spurious -v
```

Expected: PASS.

**Step: Commit**

```bash
go fmt ./core/...
git add pkg/core/block_validator.go pkg/core/spurious_entry_test.go
git commit -m "feat(bal): validate BAL indices <= len(txs)+1"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
Lines 400-402:
The BAL MUST be complete and accurate. Missing or spurious entries invalidate the block. Spurious entries MAY be detected by validating BAL indices, which MUST never be higher than `len(transactions) + 1`.

Clients MAY invalidate immediately if any transaction exceeds declared state.
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/block_validator.go` | `ValidateBlockAccessList` checks BAL hash presence and hash equality; no index-range check; `ErrSpuriousBALEntry` not defined |
| `pkg/core/block_validator.go` | Defines `ErrInvalidBlockAccessList` and `ErrMissingBlockAccessList`; `ErrSpuriousBALEntry` is absent from the error var block |
| `pkg/bal/types.go` | `AccessEntry.AccessIndex uint64` is a single flat index per entry; the spec's per-change `BlockAccessIndex uint16` model is not yet represented |
| `pkg/bal/tracker.go` | `AccessTracker.Build(txIndex uint64)` stamps all changes with a single `txIndex`; no multi-index change lists to validate against |

---

## Implementation Assessment

### Current Status

Not implemented.

### Architecture Notes

The story places `validateBALIndices` and `ErrSpuriousBALEntry` in `pkg/core/block_validator.go`. The existing `ValidateBlockAccessList` method in that file checks only hash presence and hash equality — it knows nothing about per-change `BlockAccessIndex` values. The plan's `validateBALIndices` function expects `entry.StorageChanges[i].Changes[j].BlockAccessIndex` (a slice of per-change structs with a `uint16` index), but the current `bal.AccessEntry` type in `pkg/bal/types.go` uses a single flat `AccessIndex uint64` on the entry, not a slice of `StorageChange` structs that each carry their own `BlockAccessIndex`. This means both the `bal` type model and the `block_validator.go` function need changes before the validation can be implemented.

### Gaps and Proposed Solutions

1. **`ErrSpuriousBALEntry` is not defined.** Add it to the `var (...)` error block in `pkg/core/block_validator.go` alongside the existing `ErrInvalidBlockAccessList` and `ErrMissingBlockAccessList`.

2. **`bal.AccessEntry` does not carry per-change `BlockAccessIndex` fields.** The current model stores one `AccessIndex` per entry. The spec model stores `[block_access_index, value]` pairs inside `balance_changes`, `nonce_changes`, `code_changes`, and `StorageChange.Changes`. The `bal` types must be extended (or a parallel `RLPAccountChanges` type added) before `validateBALIndices` can iterate `entry.BalanceChanges[i].BlockAccessIndex`.

3. **`validateBALIndices` function is absent.** Once the type gap above is resolved, add the function to `block_validator.go` and call it from `ValidateBlockAccessList` (or from a new `ValidateBALIndices` method on `BlockValidator`) after the hash check.

4. **Test file `pkg/core/spurious_entry_test.go` does not exist.** Create it with `TestBALValidator_SpuriousIndex_Rejected` and `TestBALValidator_ValidMaxIndex_Accepted` as specified in Task 15.1.1.
