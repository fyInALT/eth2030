# Story 13.1 — Implement `ValidateBALSizeConstraint`

> **Sprint context:** Sprint 13 — BAL Size Constraint Validation
> **Sprint Goal:** The block validator enforces the spec's gas-based size constraint on the BAL.

**Spec reference:** Lines 114-129.
```
bal_items * ITEM_COST <= available_gas + system_allowance
```
Where:
- `bal_items = storage_reads + addresses`
- `ITEM_COST = GAS_WARM_ACCESS + TX_ACCESS_LIST_STORAGE_KEY_COST` = 100 + 1900 = 2000
- `available_gas = block_gas_limit - tx_count * TX_BASE_COST`
- `system_allowance = (15 + 3 * (MAX_WITHDRAWAL_REQUESTS_PER_BLOCK + MAX_CONSOLIDATION_REQUESTS_PER_BLOCK)) * ITEM_COST`

**Files:**
- Modify: `pkg/core/block_validator.go`
- Test: `pkg/core/bal_size_constraint_test.go`

**Acceptance Criteria:** A BAL whose `bal_items` exceeds the gas-based limit causes `ValidateBlock` to return `ErrBALSizeExceeded`; one within limits passes.

#### Task 13.1.1 — Write failing tests

File: `pkg/core/bal_size_constraint_test.go`

```go
const (
    ItemCost       = 100 + 1900 // GAS_WARM_ACCESS + TX_ACCESS_LIST_STORAGE_KEY_COST
    TxBaseCost     = 21000
    MaxWithdrawReq = 16  // EIP-7002
    MaxConsolidReq = 2   // EIP-7251
)

func TestBALSizeConstraint_WithinLimit_Passes(t *testing.T) {
    blockGasLimit := uint64(30_000_000)
    txCount := 1
    // Compute max allowed items
    availableGas := blockGasLimit - uint64(txCount)*TxBaseCost
    sysAllowance := uint64(15+3*(MaxWithdrawReq+MaxConsolidReq)) * ItemCost
    maxItems := (availableGas + sysAllowance) / ItemCost

    bal := makeBALWithItems(int(maxItems) - 1) // one under limit
    err := ValidateBALSizeConstraint(bal, blockGasLimit, txCount)
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
}

func TestBALSizeConstraint_ExceedsLimit_Fails(t *testing.T) {
    blockGasLimit := uint64(30_000_000)
    txCount := 0
    bal := makeBALWithItems(20_000) // definitely over limit for empty block
    err := ValidateBALSizeConstraint(bal, blockGasLimit, txCount)
    if err == nil {
        t.Fatal("expected ErrBALSizeExceeded")
    }
}
```

#### Task 13.1.2 — Implement `ValidateBALSizeConstraint`

In `pkg/core/block_validator.go`:

```go
var ErrBALSizeExceeded = errors.New("block access list exceeds gas-based size limit")

const (
    balItemCost            = 100 + 1900 // GAS_WARM_ACCESS + TX_ACCESS_LIST_STORAGE_KEY_COST
    txBaseCost             = 21000
    maxWithdrawalRequests  = 16  // EIP-7002
    maxConsolidationReqs   = 2   // EIP-7251
    systemAccessAllowance  = 15  // system contract slots outside user txs
)

// ValidateBALSizeConstraint checks that the BAL does not exceed the gas-proportional size limit.
func ValidateBALSizeConstraint(bl *bal.BlockAccessList, blockGasLimit uint64, txCount int) error {
    // Count bal_items = storage_reads + addresses
    var storageReads, addresses uint64
    for _, entry := range bl.Entries {
        addresses++
        storageReads += uint64(len(entry.StorageReads))
    }
    balItems := storageReads + addresses

    availableGas := blockGasLimit - uint64(txCount)*txBaseCost
    sysAllowance := uint64(systemAccessAllowance+3*(maxWithdrawalRequests+maxConsolidationReqs)) * balItemCost
    maxItems := (availableGas + sysAllowance) / balItemCost

    if balItems > maxItems {
        return fmt.Errorf("%w: %d items > %d limit (gas=%d)",
            ErrBALSizeExceeded, balItems, maxItems, blockGasLimit)
    }
    return nil
}
```

#### Task 13.1.3 — Wire into `ValidateBlock`

In `ValidateBlock()`, after BAL hash validation:

```go
if err := ValidateBALSizeConstraint(computedBAL, header.GasLimit, len(block.Transactions())); err != nil {
    return err
}
```

**Step: Run tests**

```
cd /projects/eth2030/pkg && go test ./core/... -run TestBALSizeConstraint -v
```

Expected: PASS.

**Step: Commit**

```bash
cd /projects/eth2030/pkg && go fmt ./core/...
git add pkg/core/block_validator.go pkg/core/bal_size_constraint_test.go
git commit -m "feat(bal): enforce gas-based BAL size constraint"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
Lines 114-129:
### Block Access List Size Constraint

The block access list is constrained by available gas rather than a fixed maximum number of items. The constraint is defined as:

```
bal_items * ITEM_COST <= available_gas + system_allowance
```

Where:

- `bal_items = storage_reads + addresses`
- `ITEM_COST = GAS_WARM_ACCESS + TX_ACCESS_LIST_STORAGE_KEY_COST`
- `available_gas = block_gas_limit - tx_count * TX_BASE_COST`
- `system_allowance = (15 + 3 * (MAX_WITHDRAWAL_REQUESTS_PER_BLOCK + MAX_CONSOLIDATION_REQUESTS_PER_BLOCK)) * ITEM_COST`

The `storage_reads` is the total number of storage accesses across all accounts, and `addresses` is the total number of unique addresses accessed in the block. The `system_allowance` term accounts for system contract accesses that occur outside of user transactions. `MAX_WITHDRAWAL_REQUESTS_PER_BLOCK` is defined in [EIP-7002](./eip-7002.md) and `MAX_CONSOLIDATION_REQUESTS_PER_BLOCK` is defined in [EIP-7251](./eip-7251.md).
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/block_validator.go` (lines 14-35) | Defines all block validation errors; `ErrBALSizeExceeded` is absent — only `ErrInvalidBlockAccessList` and `ErrMissingBlockAccessList` exist |
| `pkg/core/block_validator.go` (lines 254-287) | `ValidateBlockAccessList` — validates the BAL hash against the header field; no size constraint check is present |
| `pkg/core/block_validator.go` (lines 77-139) | `ValidateHeader` — the main validation chain; calls `ValidateBlockBlobGas` and `ValidateCalldataGas` but has no call to any BAL size constraint function |
| `pkg/core/block_validator.go` (lines 141-189) | `ValidateBody` — validates transactions, blob gas, calldata gas, and withdrawals; no BAL size constraint check |
| `pkg/bal/types.go` | `BlockAccessList` and `AccessEntry` types; `AccessEntry` has `StorageReads []StorageAccess` and `Address` fields needed for counting `bal_items` |

---

## Implementation Assessment

### Current Status

Not implemented.

### Architecture Notes

`pkg/core/block_validator.go` contains `ValidateBlockAccessList` (lines 254-287) which only checks that the `BlockAccessListHash` in the header is present and matches the computed hash. There is no `ValidateBALSizeConstraint` function anywhere in the codebase, and `ErrBALSizeExceeded` does not exist.

The story's proposed implementation is structurally sound and aligns with the existing validator pattern. The `BlockAccessList` type in `pkg/bal/types.go` exposes `Entries []AccessEntry`, and each `AccessEntry` has `StorageReads []StorageAccess` and an `Address` field, which maps directly to the spec's `storage_reads + addresses` formula.

One naming discrepancy: the spec says `ITEM_COST = GAS_WARM_ACCESS + TX_ACCESS_LIST_STORAGE_KEY_COST = 100 + 1900 = 2000`. The story's implementation uses `balItemCost = 100 + 1900`, which is correct. However, it must also verify that `storage_reads` in the spec counts total storage accesses across all accounts (both `StorageReads` and `StorageChanges` entries), not just the read-only slots. Looking at the spec text (line 124): `bal_items = storage_reads + addresses` where "storage_reads is the total number of storage accesses across all accounts" — this appears to mean all storage accesses (reads + writes), not just `storage_reads` entries in the BAL structure.

The `TX_BASE_COST` in the story is set to 21000. Under Glamsterdam (EIP-2780), the base tx cost is 4500. This constant needs to be fork-aware.

The constraint must be wired into `ValidateBody` or a new `ValidateBAL` method called from the block processing flow. Currently `ValidateHeader` does not receive the BAL object, only the header and parent.

### Gaps and Proposed Solutions

1. **`ValidateBALSizeConstraint` and `ErrBALSizeExceeded` do not exist**: Need to add both to `pkg/core/block_validator.go` as described in the story. The implementation is straightforward given the existing `BlockAccessList` type.

2. **`storage_reads` definition ambiguity**: The spec's `storage_reads` in the size formula likely means total storage accesses (both `StorageReads` and `StorageChanges` entries), not just the read-only slots. The story's implementation only counts `len(entry.StorageReads)`. Need to verify and likely add `len(entry.StorageChanges)` to the count.

3. **`TX_BASE_COST` must be fork-aware**: Under Glamsterdam (EIP-2780, `vm.TxBaseGlamsterdam = 4500`), the base transaction cost differs from the pre-Glamsterdam 21000. The `ValidateBALSizeConstraint` function needs to accept or derive the fork-appropriate `TX_BASE_COST`.

4. **Wiring into the validation flow**: `ValidateBlockAccessList` already receives the computed BAL hash, but not the BAL object itself. The size constraint check requires the full BAL. The call site in the block processing pipeline needs to pass the `*bal.BlockAccessList` to a combined validation step alongside the hash check.

5. **`system_allowance` constants**: `MAX_WITHDRAWAL_REQUESTS_PER_BLOCK = 16` (EIP-7002) and `MAX_CONSOLIDATION_REQUESTS_PER_BLOCK = 2` (EIP-7251) should be referenced from their canonical definitions rather than hardcoded, if those constants exist in the codebase.
