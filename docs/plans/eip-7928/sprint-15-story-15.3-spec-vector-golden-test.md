# Story 15.3 — Spec vector golden assertions (`TestSpecVector_ConcreteExample`)

> **Sprint context:** Sprint 15 — Spurious Entry Validation & Spec Test Vectors
> **Sprint Goal:** The validator catches spurious BAL entries (index > n+1); a golden test reproduces the concrete example from the EIP spec exactly.

**Spec reference:** Lines 406-511. Verifies the exact BAL structure from the EIP's concrete example.

**Files:**
- Modify: `pkg/bal/spec_vector_test.go`

**Acceptance Criteria:** `TestSpecVector_ConcreteExample` passes with all assertions: correct address order, correct `block_access_index` per entry, correct change types (storage_changes vs storage_reads), and correct empty lists for read-only addresses.

#### Task 15.3.1 — Write the golden test

Add to `pkg/bal/spec_vector_test.go`:

```go
// TestSpecVector_ConcreteExample reproduces the example from EIP-7928 spec lines 406-511.
// Expected address order (lexicographic):
//   0x0000F908..., 0x2222..., 0xaaaa..., 0xabcd..., 0xbbbb...,
//   0xcccc..., 0xdddd..., 0xeeee... (COINBASE), 0xffff...
func TestSpecVector_ConcreteExample(t *testing.T) {
    blockHashContract := mustAddr("0x0000F90827F1C53a10cb7A02335B175320002935")
    alice             := mustAddr("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
    bob               := mustAddr("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
    checked           := mustAddr("0x2222222222222222222222222222222222222222")
    charlie           := mustAddr("0xCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC")
    factory           := mustAddr("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
    deployed          := mustAddr("0xDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD")
    coinbase          := mustAddr("0xEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE")
    eve               := mustAddr("0xABCDABCDABCDABCDABCDABCDABCDABCDABCDABCD")

    events := buildSpecVectorEvents(
        blockHashContract, alice, bob, checked,
        charlie, factory, deployed, coinbase, eve,
    )
    bl := BuildFromEvents(events)

    // Verify address order (lexicographic)
    expectedOrder := [][20]byte{
        blockHashContract, checked, alice, eve, bob,
        charlie, deployed, coinbase, factory,
    }
    if len(bl.Entries) != len(expectedOrder) {
        t.Fatalf("expected %d entries, got %d", len(expectedOrder), len(bl.Entries))
    }
    for i, expected := range expectedOrder {
        if bl.Entries[i].Address != expected {
            t.Errorf("entry[%d]: expected %x, got %x", i, expected, bl.Entries[i].Address)
        }
    }

    // Block hash contract: 1 storage_change at index 0
    bhc := bl.Entries[0]
    if len(bhc.StorageChanges) != 1 || bhc.StorageChanges[0].Changes[0].BlockAccessIndex != 0 {
        t.Error("block hash contract: expected 1 storage write at index 0")
    }

    // Alice: balance_change and nonce_change at index 1
    aliceEntry := findEntry(bl, alice)
    if len(aliceEntry.BalanceChanges) != 1 || aliceEntry.BalanceChanges[0].BlockAccessIndex != 1 {
        t.Error("Alice: expected balance_change at index 1")
    }
    if len(aliceEntry.NonceChanges) != 1 || aliceEntry.NonceChanges[0].BlockAccessIndex != 1 {
        t.Error("Alice: expected nonce_change at index 1")
    }

    // Eve: balance_change at index 3 (post-execution)
    eveEntry := findEntry(bl, eve)
    if len(eveEntry.BalanceChanges) != 1 || eveEntry.BalanceChanges[0].BlockAccessIndex != 3 {
        t.Error("Eve: expected balance_change at index 3 (post-execution)")
    }

    // COINBASE: two balance_changes at indices 1 and 2
    cbEntry := findEntry(bl, coinbase)
    if len(cbEntry.BalanceChanges) != 2 {
        t.Errorf("COINBASE: expected 2 balance_changes, got %d", len(cbEntry.BalanceChanges))
    }
    if cbEntry.BalanceChanges[0].BlockAccessIndex != 1 || cbEntry.BalanceChanges[1].BlockAccessIndex != 2 {
        t.Error("COINBASE: expected balance_changes at indices 1 and 2")
    }

    // Checked address (0x2222): all empty lists (read-only)
    checkedEntry := findEntry(bl, checked)
    if len(checkedEntry.StorageChanges) != 0 || len(checkedEntry.BalanceChanges) != 0 {
        t.Error("checked address: expected all empty lists")
    }

    // Deployed contract: nonce_change and code_change at index 2
    deployedEntry := findEntry(bl, deployed)
    if len(deployedEntry.NonceChanges) != 1 || deployedEntry.NonceChanges[0].BlockAccessIndex != 2 {
        t.Error("deployed contract: expected nonce_change at index 2")
    }
    if len(deployedEntry.CodeChanges) != 1 || deployedEntry.CodeChanges[0].BlockAccessIndex != 2 {
        t.Error("deployed contract: expected code_change at index 2")
    }
}
```

**Step: Run the spec vector test**

```
cd /projects/eth2030/pkg && go test ./bal/... -run TestSpecVector_ConcreteExample -v
```

Expected: PASS (every assertion in the spec example is verified).

**Step: Commit**

```bash
git add pkg/bal/spec_vector_test.go
git commit -m "test(bal): EIP-7928 spec concrete example as golden test vector"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
Lines 406-511 (Concrete Example — assertions verified by this test):

BlockAccessIndex assignment (lines 189-193):
  0   — pre-execution system contract calls
  1…n — transactions in block order
  n+1 — post-execution system contract calls

Ordering (lines 180-185):
  Accounts: lexicographic by address
  storage_changes: slots lexicographic by key; within each slot, changes by block_access_index ascending
  balance_changes, nonce_changes, code_changes: by block_access_index ascending

Expected entry order for the 2-transaction example block:
  [0] 0x0000F90827F1C53a10cb7A02335B175320002935 — 1 storage_change at index 0
  [1] 0x2222...  — all empty lists (read-only EXTCODEHASH check)
  [2] 0xaaaa...  — balance_change at index 1, nonce_change at index 1
  [3] 0xabcd...  — balance_change at index 3 (post-execution withdrawal)
  [4] 0xbbbb...  — balance_change at index 1
  [5] 0xcccc...  — balance_change at index 2, nonce_change at index 2
  [6] 0xdddd...  — nonce_change at index 2, code_change at index 2
  [7] 0xeeee...  — balance_changes at indices 1 and 2 (COINBASE)
  [8] 0xffff...  — storage_change at index 2 (slot 1 → deployed address), nonce_change at index 2
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/bal/types.go` | `AccessEntry` has single `AccessIndex uint64`; test assertions require per-change `BlockAccessIndex` on `BalanceChanges`, `NonceChanges`, `CodeChanges`, and `StorageChanges` sub-entries |
| `pkg/bal/tracker.go` | `AccessTracker.Build(txIndex)` only produces single-index entries; multi-phase simulation needed by `buildSpecVectorEvents` is not possible with current API |
| `pkg/bal/types_extended.go` | No `AccountEvents`, `NewBALAccessTracker`, or `BuildFromEvents` symbol |
| `pkg/bal/hash.go` | `BlockAccessList.Hash()` / `EncodeRLP()` present; usable once the type model supports per-change indices |

---

## Implementation Assessment

### Current Status

Not implemented.

### Architecture Notes

`TestSpecVector_ConcreteExample` does not exist. The test (in Story 15.3) directly depends on the fixture helper `buildSpecVectorEvents` from Story 15.2, which in turn depends on `NewBALAccessTracker`, `AccountEvents`, and `BuildFromEvents` — none of which exist in `pkg/bal/`. Beyond the missing symbols, the assertions themselves reveal a structural mismatch with the current type model: the test accesses `aliceEntry.BalanceChanges[0].BlockAccessIndex` and `cbEntry.BalanceChanges[1].BlockAccessIndex`, but the current `AccessEntry` holds only a pointer `BalanceChange *BalanceChange` (not a slice), and `BalanceChange` carries `OldValue`/`NewValue` with no `BlockAccessIndex` field. The same mismatch exists for `NonceChanges`, `CodeChanges`, and `StorageChanges.Changes`. The address hex values in the test use fully padded 40-character hex strings (e.g., `0xAAAA...AAAA`) and `mustAddr` strips the `0x` prefix and asserts `len(b) == 20`; this is correct and consistent with the spec's abbreviated notation.

### Gaps and Proposed Solutions

1. **`AccessEntry` must be extended** to carry slices of indexed changes rather than single-change pointers. Concretely: replace `BalanceChange *BalanceChange` with `BalanceChanges []IndexedBalanceChange`, and similarly for nonce and code changes. `IndexedBalanceChange` must carry a `BlockAccessIndex uint16` (matching the spec's `uint16` type alias) plus the post-value. `StorageChange` must become a per-slot entry whose `Changes []SlotChange` each carry `BlockAccessIndex uint16` and `NewValue`.

2. **`BuildFromEvents` must sort correctly.** The golden test verifies exact address order. `buildSpecVectorEvents` supplies addresses as fully-qualified 20-byte arrays, and `BuildFromEvents` must sort them byte-by-byte lexicographically using the same `addrLess` function already in `tracker.go`.

3. **COINBASE address has two balance changes.** The test asserts `len(cbEntry.BalanceChanges) == 2` with indices 1 and 2. The new tracker must accumulate multiple `IndexedBalanceChange` entries for the same address when called multiple times with different `blockAccessIndex` values, rather than overwriting the previous entry as the current `AccessTracker.RecordBalanceChange` does.

4. **The file `pkg/bal/spec_vector_test.go` does not exist.** It must be created (in Story 15.2) before this test can be added. Once Stories 15.2 and 15.3 are both complete, run `go test ./bal/... -run TestSpecVector_ConcreteExample -v` to confirm all nine address-order assertions and all per-entry change-type assertions pass.
