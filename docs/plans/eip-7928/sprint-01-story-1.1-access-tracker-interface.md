# Story 1.1 — Define `AccessTracker` Interface in `pkg/core/vm/`

> **Sprint context:** Sprint 1 — EVM Opcode Access Tracking
> **Sprint Goal:** Every EVM opcode that touches state emits a structured access event into a per-transaction tracker, so the BAL can be built from real execution data.

**Files:**
- Create: `pkg/core/vm/access_tracker.go`
- Test: `pkg/core/vm/access_tracker_test.go`

**Acceptance Criteria:** The `AccessTracker` interface is defined; a `NoopAccessTracker` ships for pre-Amsterdam blocks; tests confirm the interface compiles and noop methods are callable.

---

#### Task 1.1.1 — Write failing test for interface & noop

**Step 1: Write the failing test**

File: `pkg/core/vm/access_tracker_test.go`

```go
package vm_test

import (
	"testing"
	"github.com/your-org/eth2030/pkg/core/vm"
)

func TestNoopAccessTracker_ImplementsInterface(t *testing.T) {
	var tracker vm.AccessTracker = vm.NewNoopAccessTracker()
	if tracker == nil {
		t.Fatal("expected non-nil tracker")
	}
}

func TestNoopAccessTracker_RecordAddress(t *testing.T) {
	tracker := vm.NewNoopAccessTracker()
	// Should not panic
	tracker.RecordAddressAccess([20]byte{0x01}, 1)
	tracker.RecordStorageRead([20]byte{0x01}, [32]byte{0x02}, 1)
	tracker.RecordStorageWrite([20]byte{0x01}, [32]byte{0x02}, [32]byte{0xaa}, 1)
	tracker.RecordBalanceChange([20]byte{0x01}, uint256Bytes(100), 1)
	tracker.RecordNonceChange([20]byte{0x01}, 5, 1)
	tracker.RecordCodeChange([20]byte{0x01}, []byte{0x60, 0x00}, 1)
}
```

**Step 2: Run to verify it fails**

```
cd /projects/eth2030/pkg && go test ./core/vm/... -run TestNoopAccessTracker -v
```

Expected: `FAIL` — `vm.AccessTracker` undefined.

**Step 3: Implement**

File: `pkg/core/vm/access_tracker.go`

```go
package vm

// AccessTracker records state accesses during EVM execution for EIP-7928 BAL building.
// One tracker instance is created per block; the TxIndex must be set before each transaction.
type AccessTracker interface {
	// RecordAddressAccess records that an address was touched (read or written).
	RecordAddressAccess(addr [20]byte, txIndex uint16)
	// RecordStorageRead records a storage slot that was read but not written.
	RecordStorageRead(addr [20]byte, slot [32]byte, txIndex uint16)
	// RecordStorageWrite records a storage write with the post-transaction value.
	RecordStorageWrite(addr [20]byte, slot [32]byte, newVal [32]byte, txIndex uint16)
	// RecordBalanceChange records the post-transaction balance for an address.
	RecordBalanceChange(addr [20]byte, postBalance [32]byte, txIndex uint16)
	// RecordNonceChange records the post-transaction nonce for an address.
	RecordNonceChange(addr [20]byte, newNonce uint64, txIndex uint16)
	// RecordCodeChange records the post-transaction bytecode for an address.
	RecordCodeChange(addr [20]byte, code []byte, txIndex uint16)
}

// NoopAccessTracker satisfies AccessTracker with empty implementations.
// Used for pre-Amsterdam blocks where BAL is not required.
type NoopAccessTracker struct{}

// NewNoopAccessTracker returns a tracker that discards all events.
func NewNoopAccessTracker() AccessTracker { return &NoopAccessTracker{} }

func (*NoopAccessTracker) RecordAddressAccess([20]byte, uint16)          {}
func (*NoopAccessTracker) RecordStorageRead([20]byte, [32]byte, uint16)  {}
func (*NoopAccessTracker) RecordStorageWrite([20]byte, [32]byte, [32]byte, uint16) {}
func (*NoopAccessTracker) RecordBalanceChange([20]byte, [32]byte, uint16) {}
func (*NoopAccessTracker) RecordNonceChange([20]byte, uint64, uint16)     {}
func (*NoopAccessTracker) RecordCodeChange([20]byte, []byte, uint16)      {}
```

**Step 4: Run tests to verify they pass**

```
cd /projects/eth2030/pkg && go test ./core/vm/... -run TestNoopAccessTracker -v
```

Expected: PASS.

**Step 5: Format & commit**

```bash
cd /projects/eth2030/pkg && go fmt ./core/vm/...
git add pkg/core/vm/access_tracker.go pkg/core/vm/access_tracker_test.go
git commit -m "feat(bal): define AccessTracker interface + noop"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
### RLP Data Structures

BALs use RLP encoding following the pattern: `address -> field -> block_access_index -> change`.

# BlockAccessIndex = uint16  # Block access index (0 for pre-execution, 1..n for transactions, n+1 for post-execution)

### BlockAccessIndex Assignment

`BlockAccessIndex` values **MUST** be assigned as follows:

- `0` for **pre‑execution** system contract calls.
- `1 … n` for transactions (in block order).
- `n + 1` for **post‑execution** system contract calls.

### Recording Semantics by Change Type

#### Storage

- **Writes include:**

  - Any value change (post‑value ≠ pre‑value).
  - **Zeroing** a slot (pre‑value exists, post‑value is zero).

- **Reads include:**

  - Slots accessed via `SLOAD` that are not written.
  - Slots written with unchanged values (i.e., `SSTORE` where post-value equals pre-value, also known as "no-op writes").

#### Balance (`balance_changes`)

Record **post‑transaction** balances (`uint256`) for:

- Transaction **senders** (gas + value).
- Transaction **recipients** (only if `value > 0`).
- CALL/CALLCODE **senders** (value).
- CALL/CALLCODE **recipients** (only if `value > 0`).
- CREATE/CREATE2 recipients (only if `value > 0`).
- **COINBASE** (rewards + fees).
- **SELFDESTRUCT/SENDALL** beneficiaries.
- **Withdrawal recipients** (system withdrawals, [EIP-4895](./eip-4895.md)).

#### Nonce

Record **post‑transaction nonces** for:

- EOA senders.
- Contracts that performed a successful `CREATE` or `CREATE2`.
- Deployed contracts.
- [EIP-7702](./eip-7702.md) authorities.

#### Code

Track **post‑transaction runtime bytecode** for deployed or modified contracts, and **delegation indicators** for successful delegations as defined in [EIP-7702](./eip-7702.md).
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/bal/tracker.go` | Existing `AccessTracker` struct (concrete type, not interface); implements `RecordStorageRead`, `RecordStorageChange`, `RecordBalanceChange`, `RecordNonceChange`, `RecordCodeChange`, `Build(txIndex)`, `Reset()` |
| `pkg/bal/types.go` | Defines `BlockAccessList`, `AccessEntry`, `StorageAccess`, `StorageChange`, `BalanceChange`, `NonceChange`, `CodeChange` |
| `pkg/core/vm/access_list_tracker.go` | EIP-2929 warm/cold `AccessListTracker` — unrelated to EIP-7928; no BAL interface here |
| `pkg/core/vm/interpreter.go` | EVM struct definition — no `AccessTracker` or `TxIndex` fields present |
| `pkg/core/vm/evm_storage_ops.go` | SLOAD/SSTORE handling; uses `AccessListTracker` for warmth only; no BAL emit calls |
| `pkg/core/vm/instructions.go` | Raw opcode implementations (`opSload`, `opSstore`, `opBalance`, `opCall`, `opCreate`, etc.) |

---

## Implementation Assessment

### Current Status

Not implemented.

### Architecture Notes

The plan calls for an `AccessTracker` **interface** (plus `NoopAccessTracker`) in `pkg/core/vm/access_tracker.go`. This file does not exist. The codebase has:

- `pkg/bal/tracker.go`: a **concrete struct** named `AccessTracker` that is per-transaction (single-tx focus via `Build(txIndex)`), not a Go interface, and lives in the `bal` package rather than `pkg/core/vm`. Its method signatures also differ from the plan: it takes `types.Hash` and `*big.Int` rather than raw `[32]byte` arrays.
- `pkg/core/vm/access_list_tracker.go`: the `AccessListTracker` struct handles only EIP-2929 warm/cold accounting and is entirely separate from EIP-7928 BAL tracking.

The `EVM` struct in `pkg/core/vm/interpreter.go` has no `AccessTracker` or `TxIndex` fields, so there is no hook point yet.

### Gaps and Proposed Solutions

| Gap | Proposed Solution |
|-----|-------------------|
| No `AccessTracker` interface in `pkg/core/vm/` | Create `pkg/core/vm/access_tracker.go` with the interface definition as shown in the story |
| No `NoopAccessTracker` | Ship alongside the interface in the same file; used for pre-Amsterdam blocks |
| Naming collision with `pkg/bal.AccessTracker` (struct) | The plan's interface lives in a different package (`vm`) so there is no direct naming conflict, but documentation should clarify the distinction |
| Method signatures differ from plan | The plan uses `[20]byte` and `[32]byte` raw arrays; `pkg/bal.AccessTracker` uses `types.Address` / `types.Hash` / `*big.Int`. Align the new interface to use raw arrays as the plan specifies, since it bridges the VM layer to the BAL builder |
| `EVM` struct lacks `AccessTracker` and `TxIndex` fields | Addressed in Story 1.3; Story 1.1 only defines the interface and noop |
