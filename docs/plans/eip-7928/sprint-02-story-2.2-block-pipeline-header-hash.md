# Story 2.2 — Integrate BAL into block pipeline and validate header hash

> **Sprint context:** Sprint 2 — BAL Assembly & Header Integration
> **Sprint Goal:** After block execution, the processor assembles a valid, sorted `BlockAccessList` from tracker events, computes its Keccak256 hash, and sets `block_access_list_hash` in the block header.

**Files:**
- Modify: `pkg/core/processor.go`
- Modify: `pkg/core/block_validator.go`
- Test: `pkg/core/processor_bal_test.go`
- Test: `pkg/core/block_validator_test.go`

**Acceptance Criteria:** `ProcessWithBAL()` creates a `BALAccessTracker`, drains events after all transactions, and returns a fully-populated `*BlockAccessList`; `ValidateBlock()` re-computes the BAL hash and returns `ErrInvalidBlockAccessList` on mismatch.

#### Task 2.2.1 — Write failing tests

File: `pkg/core/processor_bal_test.go`

```go
package core_test

import "testing"

func TestProcessor_BALPopulated(t *testing.T) {
	// Build a block with one simple ETH transfer
	// Call ProcessWithBAL()
	// Assert returned BAL contains sender and recipient addresses
	// Assert BAL hash matches keccak256(rlp.encode(bal))
	t.Skip("wire in story 2.2")
}
```

File: `pkg/core/block_validator_test.go`

```go
func TestBlockValidator_RejectsWrongBALHash(t *testing.T) {
	// Build block with valid BAL
	// Tamper with header BAL hash
	// Assert ValidateBlock returns ErrInvalidBlockAccessList
}
```

#### Task 2.2.2 — Wire `BALAccessTracker` into `ProcessWithBAL`

In `pkg/core/processor.go`, in the `ProcessWithBAL()` function:

```go
// Create BAL tracker for Amsterdam forks
var tracker vm.AccessTracker = vm.NewNoopAccessTracker()
if p.config.IsAmsterdam(header.Time) {
	tracker = vm.NewBALAccessTracker()
}

for i, tx := range block.Transactions() {
	evm.TxIndex = uint16(i + 1) // 1-based; 0 reserved for pre-execution
	evm.AccessTracker = tracker
	// ... existing execution code ...
}

// After all transactions: drain events and build BAL
if balTracker, ok := tracker.(*vm.BALAccessTracker); ok {
	events := balTracker.Drain()
	result.BlockAccessList = bal.BuildFromEvents(events)
	result.BALHash = result.BlockAccessList.Hash()
}
```

#### Task 2.2.3 — Enforce BAL hash in `ValidateBlock`

In `pkg/core/block_validator.go`, add to `ValidateBlock()`:

```go
if p.config.IsAmsterdam(header.Time) {
	computedHash := result.BlockAccessList.Hash()
	if header.BlockAccessListHash == nil {
		return ErrMissingBlockAccessList
	}
	if *header.BlockAccessListHash != computedHash {
		return fmt.Errorf("%w: got %x want %x",
			ErrInvalidBlockAccessList, *header.BlockAccessListHash, computedHash)
	}
}
```

**Step: Run integration tests**

```
cd /projects/eth2030/pkg && go test ./core/... -run "TestBAL|TestBlockValidator" -v
```

Expected: PASS.

**Step: Format & commit**

```bash
cd /projects/eth2030/pkg && go fmt ./core/...
git add pkg/core/processor.go pkg/core/processor_bal_test.go \
        pkg/core/block_validator.go pkg/core/block_validator_test.go
git commit -m "feat(bal): wire BAL into block pipeline + enforce header hash"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
### Block Structure Modification

We introduce a new field to the block header, `block_access_list_hash`, which contains the
Keccak-256 hash of the RLP-encoded block access list. When no state changes are present,
this field is the hash of an empty RLP list
`0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347`,
i.e. `keccak256(rlp.encode([]))`.

class Header:
    # Existing fields
    ...
    block_access_list_hash: Hash32 = keccak256(rlp.encode(block_access_list))

### State Transition Function

def validate_block(execution_payload, block_header):
    # 1. Compute hash from received BAL and set in header
    block_header.block_access_list_hash = keccak(execution_payload.blockAccessList)

    # 2. Execute block and collect actual accesses
    actual_bal = execute_and_collect_accesses(execution_payload)

    # 3. Verify actual execution matches provided BAL
    # If this fails, the block is invalid (the hash in the header would be wrong)
    assert rlp.encode(actual_bal) == execution_payload.blockAccessList

def execute_and_collect_accesses(block):
    accesses = {}

    # Pre-execution system contracts (block_access_index = 0)
    track_system_contracts_pre(block, accesses, block_access_index=0)

    # Execute transactions (block_access_index = 1..n)
    for i, tx in enumerate(block.transactions):
        execute_transaction(tx)
        track_state_changes(tx, accesses, block_access_index=i+1)

    # Withdrawals and post-execution (block_access_index = len(txs) + 1)
    post_index = len(block.transactions) + 1
    for withdrawal in block.withdrawals:
        apply_withdrawal(withdrawal)
        track_balance_change(withdrawal.address, accesses, post_index)
    track_system_contracts_post(block, accesses, post_index)

    return build_bal(accesses)
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/block_validator.go` | Defines `ErrInvalidBlockAccessList`, `ErrMissingBlockAccessList`; implements `ValidateBlockAccessList(header, computedBALHash)` which checks Amsterdam fork gate, nil-hash presence, and hash equality |
| `pkg/core/processor.go` | `ProcessWithBAL()` builds a `BlockAccessList` and returns it inside `ProcessResult`; uses `bal.NewTracker()` per-tx; returns `nil` BAL for pre-Amsterdam blocks |
| `pkg/core/blockchain.go` | `InsertBlock()` calls `processor.ProcessWithBAL()` then `validator.ValidateBlockAccessList()` with the computed hash |
| `pkg/core/block_builder.go` | `BuildBlockLegacy()` and `BuildBlock()` call `ProcessWithBAL()` and set `header.BlockAccessListHash` from `result.BlockAccessList.Hash()` |
| `pkg/core/bal_integration_test.go` | Integration tests: `TestProcessWithBAL_EmptyBlock`, `TestBALHash_Computed`, `TestBlockchain_RejectsWrongBALHash`, `TestValidator_RejectsBALMismatch`, etc. |

---

## Implementation Assessment

### Current Status

Implemented.

### Architecture Notes

The plan's story describes wiring `BALAccessTracker` into a `ProcessWithBAL()` function in `pkg/core/processor.go`, with `ValidateBlock()` enforcing the BAL hash. The actual implementation matches this intent at the integration level, but differs in several internal details.

The plan assumed `processor.go` would be the file, while the actual implementation lives in `pkg/core/processor.go` under `StateProcessor.ProcessWithBAL()`. The `BALAccessTracker` type from `pkg/core/vm` referenced in the plan's pseudocode (`vm.NewBALAccessTracker()`, `balTracker.Drain()`) does not exist; instead the code uses `bal.NewTracker()` (the `AccessTracker` from `pkg/bal/tracker.go`) and a `populateTracker()` helper.

The validation path (`ValidateBlockAccessList`) lives in `pkg/core/block_validator.go` as a standalone method rather than inside a `ValidateBlock()` method as the plan described. It takes a pre-computed `*types.Hash` parameter rather than re-running the BAL computation internally. The blockchain's `InsertBlock()` calls `ProcessWithBAL()` first, then passes the resulting hash to `ValidateBlockAccessList()`.

The plan mentions `pkg/core/processor.go` should be modified but the codebase has no `processor.go` that needs modification — the file already contains the complete `ProcessWithBAL()` implementation. Similarly, the tests described (`processor_bal_test.go`) are implemented in `bal_integration_test.go`.

### Gaps and Proposed Solutions

1. **Pre- and post-execution system contract tracking at `block_access_index = 0` and `n+1` is not implemented.** `ProcessWithBAL()` only tracks per-transaction state changes (indices 1..n). Pre-execution calls such as EIP-2935 parent hash storage and EIP-4788 beacon root, and post-execution withdrawal balance changes, are not recorded in the BAL. Solution: wrap `ProcessBeaconBlockRoot()`, `ProcessParentBlockHash()`, and `ProcessWithdrawals()` calls with pre/post tracker invocations using `AccessIndex = 0` and `len(txs)+1` respectively.

2. **Storage slot reads and changes are not tracked.** `populateTracker()` only compares balance and nonce before/after each transaction. Storage reads (`SLOAD`) and storage writes (`SSTORE`) that happen inside EVM execution are not captured. Solution: hook the EVM's state-access callbacks (or add a vm-level `AccessTracker` interface) so that opcode-level storage accesses are routed into the tracker.

3. **`AccessEntry` multi-tx aggregation limitation.** As noted in story 2.1, the current schema emits one `AccessEntry` per (address, txIndex) pair. If the spec requires one entry per address across the whole block with multiple `BalanceChange`/`NonceChange` records, the schema needs to change before the hash can be correct. Solution: coordinate with story 2.1 fix and update `ProcessWithBAL` accordingly.

4. **The plan's `processor_bal_test.go` file does not exist** — coverage is in `bal_integration_test.go` instead. This is a naming discrepancy only; the tests are functionally present and cover the acceptance criteria.
