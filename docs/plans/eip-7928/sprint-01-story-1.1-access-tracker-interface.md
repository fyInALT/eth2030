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
	"github.com/eth2030/eth2030/core/vm"
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
| `pkg/core/vm/bal_tracker.go` | `BALTracker` interface definition: `RecordStorageRead`, `RecordStorageChange`, `RecordBalanceChange`, `RecordAddressTouch` |
| `pkg/core/vm/access_list_tracker.go` | EIP-2929 warm/cold `AccessListTracker` — unrelated to EIP-7928 |
| `pkg/core/vm/interpreter.go` | EVM struct with `balTracker BALTracker` and `txIndex uint64` fields; `SetBALTracker()` / `GetBALTracker()` methods |
| `pkg/core/vm/instructions.go` | Raw opcode implementations (`opSload`, `opSstore`, `opBalance`, `opCall`, `opCreate`, etc.) |

---

## Implementation Assessment

### Current Status

Complete. The BAL tracker interface exists in `pkg/core/vm/bal_tracker.go` as `BALTracker`, and the EVM struct in `pkg/core/vm/interpreter.go` already has `balTracker BALTracker` and `txIndex uint64` fields with `SetBALTracker(t BALTracker, txIdx uint64)` / `GetBALTracker()` methods.

### Architecture Notes

The actual implementation differs from the plan in naming and signatures but achieves the same goal:

- `pkg/core/vm/bal_tracker.go`: defines the `BALTracker` interface (not `AccessTracker`) with methods: `RecordStorageRead`, `RecordStorageChange`, `RecordBalanceChange`, `RecordAddressTouch`. Uses `types.Address`, `types.Hash`, and `*big.Int` rather than raw `[20]byte`/`[32]byte` arrays.
- `pkg/bal/tracker.go`: the concrete `AccessTracker` struct satisfies `BALTracker` via Go structural typing. It has `RecordStorageRead`, `RecordStorageChange`, `RecordBalanceChange`, `RecordNonceChange`, `RecordCodeChange`, `RecordAddressTouch`, `Build(txIndex)`, and `Reset()`.
- The EVM struct in `pkg/core/vm/interpreter.go` carries `balTracker BALTracker` and `txIndex uint64` fields. A nil `balTracker` functions as the noop case (no pre-Amsterdam `NoopAccessTracker` is needed; nil checks guard all call sites).
- `pkg/core/vm/access_list_tracker.go`: the `AccessListTracker` handles EIP-2929 warm/cold accounting only and is entirely separate from EIP-7928 BAL tracking.

### Gaps and Proposed Solutions

All gaps from the original plan are resolved:

| Original Gap | Resolution |
|-----|-------------------|
| No `AccessTracker` interface in `pkg/core/vm/` | `BALTracker` interface exists in `pkg/core/vm/bal_tracker.go` |
| No `NoopAccessTracker` | Not needed; nil `balTracker` field serves the same purpose (all call sites check `evm.balTracker != nil`) |
| Naming collision with `pkg/bal.AccessTracker` | No collision: the interface is named `BALTracker` in `vm` package; the concrete struct is named `AccessTracker` in `bal` package |
| Method signatures differ from plan | The actual interface uses `types.Address` / `types.Hash` / `*big.Int` rather than raw arrays; this is consistent with the rest of the codebase |
| `EVM` struct lacks tracker fields | `balTracker BALTracker` and `txIndex uint64` exist on the `EVM` struct (see Story 1.3) |
