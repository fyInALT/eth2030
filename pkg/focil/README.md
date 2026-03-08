# Package focil

Fork-Choice Enforced Inclusion Lists (EIP-7805) for censorship resistance.

## Overview

The `focil` package implements Fork-Choice Enforced Inclusion Lists as specified in EIP-7805. FOCIL is a censorship resistance mechanism that requires block builders to include transactions nominated by a randomly sampled committee of validators. If a block does not include the required transactions (without valid exemptions), attesters will refuse to vote for it, making the block non-canonical.

Each slot, an IL committee of `IL_COMMITTEE_SIZE` (16) validators is selected. Each committee member independently scans the mempool, selects high-priority pending transactions, and broadcasts a signed `InclusionList` to the network. Block builders must satisfy all received ILs — including each valid transaction or demonstrating a legitimate gas or state exemption — before their block will pass fork-choice validation.

The package covers the full FOCIL lifecycle: IL building from the mempool, structural validation, per-validator equivocation detection via `ILStore`, satisfaction checking per the EIP-7805 algorithm, compliance enforcement at the block level, and monitoring for violations.

## Table of Contents

- [Functionality](#functionality)
- [Subpackages](#subpackages)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### Core Types

| Type | Description |
|------|-------------|
| `InclusionListEntry` | A single RLP-encoded transaction with its index in the IL |
| `InclusionList` | Slot, proposer index, committee root, and a list of entries |
| `SignedInclusionList` | `InclusionList` with a 96-byte BLS signature |

### Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `MAX_TRANSACTIONS_PER_INCLUSION_LIST` | 16 (2^4) | Maximum entries per IL |
| `MAX_GAS_PER_INCLUSION_LIST` | 2,097,152 (2^21) | Maximum total gas across all IL entries |
| `MAX_BYTES_PER_INCLUSION_LIST` | 8,192 (8 KiB) | Maximum total byte size of IL transactions |
| `IL_COMMITTEE_SIZE` | 16 | Number of validators in the IL committee |
| `InclusionListUnsatisfied` | `"INCLUSION_LIST_UNSATISFIED"` | Engine API status when a valid IL tx is absent |

### IL Building

`BuildInclusionList(pending []*types.Transaction, slot uint64) *InclusionList` — constructs an IL from a slice of pending transactions:

- Sorts by gas price descending (priority fee ordering)
- Respects all three limits: `MAX_TRANSACTIONS_PER_INCLUSION_LIST`, `MAX_GAS_PER_INCLUSION_LIST`, `MAX_BYTES_PER_INCLUSION_LIST`
- Skips transactions that would violate a limit rather than truncating the list

`BuildInclusionListFromRaw(rawTxs [][]byte, slot uint64) *InclusionList` — builds an IL from already-encoded transaction bytes received from the CL layer.

Helper methods on `*InclusionList`:
- `TotalGas() uint64` — sum of gas limits across all entries
- `TotalBytes() int` — sum of encoded byte lengths
- `TransactionHashes() []types.Hash` — decoded transaction hashes

### Validation

`ValidateInclusionList(il *InclusionList) error` — structural validation per EIP-7805:

- Slot must be > 0
- At least one entry must be present
- Transaction count, total gas, and total bytes must not exceed their respective maxima
- Each entry must decode as a valid RLP transaction

Note: nonce and balance validity are intentionally deferred to attestation time (EL-side state checks happen in `CheckILSatisfaction`, not here).

### Equivocation Detection (ILStore)

`ILStore` tracks inclusion lists per validator per slot and detects equivocations (a validator submitting two different ILs for the same slot):

```go
store := focil.NewILStore()

// Returns false if the validator already equivocated for this slot.
accepted := store.AddIL(validatorIdx, slot, il)

// Check if a validator equivocated.
if store.IsEquivocator(validatorIdx, slot) { ... }

// Count equivocators for a slot.
count := store.EquivocatorCount(slot)
```

`AddIL` behavior:
- First IL from a validator for a slot is always accepted and stored
- If a second, *different* IL arrives for the same slot, the validator is marked as an equivocator
- Subsequent ILs from a known equivocator for the same slot are silently dropped

Comparison uses byte-level equality on each transaction entry, not tx hash.

### IL Satisfaction Check

`CheckILSatisfaction(block, ils, postState, gasRemaining) ILSatisfactionResult` implements the EIP-7805 satisfaction algorithm:

For each transaction T in each IL:
1. **Block inclusion**: if T is in the block — satisfied, skip
2. **Gas exemption**: if `gasRemaining < T.gasLimit` — exempt, skip
3. **State validity**: check T's nonce against `postState.GetNonce(sender)` and estimated cost against `postState.GetBalance(sender)` — if invalid, the tx is exempt
4. **Unsatisfied**: if a valid tx is absent with sufficient gas remaining — return `ILUnsatisfied`

Returns `ILSatisfied` or `ILUnsatisfied`. The `InclusionListUnsatisfied` string constant matches the Engine API status code expected by the CL.

### Compliance Checking

`CheckInclusionCompliance(block, ils) (bool, []int)` — a simplified presence-only compliance check (no gas or state exemptions):

- Returns `(true, nil)` if all ILs are satisfied
- Returns `(false, []int{...})` with the indices of any unsatisfied ILs

This is used at block gossip time before full EL execution. The full `CheckILSatisfaction` with post-state is used after block execution.

### Violation Detection

`violation_detector.go` tracks builders who repeatedly produce non-compliant blocks and accumulates evidence for potential slashing or reputation penalties.

### Enhanced IL Operations

`enhanced.go` provides extended IL operations used by the validator:

- Multi-IL merging and deduplication
- IL set coverage analysis (which pending txs are covered by at least one IL)
- Filtering ILs from equivocators before compliance evaluation

### State Checker

`state_checker.go` implements the `PostStateReader` interface backed by the local state DB for use in `CheckILSatisfaction` after block execution.

### List Validator

`list_validator.go` validates a set of ILs received over the network before admitting them to the local store, checking signatures and structural constraints.

## Subpackages

| Subpackage | Description |
|------------|-------------|
| [`committee/`](./committee/) | IL committee selection and validator index resolution |
| [`compliance/`](./compliance/) | Block compliance engine with full gas and state exemption checks |
| [`monitor/`](./monitor/) | FOCIL compliance monitoring and violation tracking |

## Usage

```go
import "github.com/eth2030/eth2030/focil"

// --- IL Committee Member: build and broadcast ---

// Build an IL from the mempool.
il := focil.BuildInclusionList(txPool.Pending(), currentSlot)

// Sign and broadcast via the CL.
signed := focil.SignedInclusionList{
    Message:   *il,
    Signature: blsSign(il),
}

// --- Node: track and validate received ILs ---

store := focil.NewILStore()

// Receive an IL from a validator.
if err := focil.ValidateInclusionList(&signed.Message); err != nil {
    return fmt.Errorf("invalid IL: %w", err)
}
if !store.AddIL(signed.Message.ProposerIndex, signed.Message.Slot, &signed.Message) {
    log.Println("dropped: validator equivocated")
}

// --- After block execution: check satisfaction ---

ils := getStoredILsForSlot(slot)
result := focil.CheckILSatisfaction(block, ils, postStateReader, gasRemaining)
if result == focil.ILUnsatisfied {
    // Return INCLUSION_LIST_UNSATISFIED to the CL via Engine API.
    return focil.InclusionListUnsatisfied
}

// --- Block gossip: quick presence check ---

allOK, unsatisfied := focil.CheckInclusionCompliance(block, ils)
if !allOK {
    log.Printf("block missing IL txs from committees: %v", unsatisfied)
}
```

## Documentation References

- [Design Doc](../../docs/DESIGN.md)
- [EIP Spec Implementation](../../docs/EIP_SPEC_IMPL.md)
- [GAP Analysis](../../docs/GAP_ANALYSIS.md)
- [EIP-7805: Fork-Choice Enforced Inclusion Lists](https://eips.ethereum.org/EIPS/eip-7805)
