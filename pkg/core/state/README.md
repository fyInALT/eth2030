# core/state — Ethereum world-state management

[← core](../README.md)

## Overview

Package `state` provides the `StateDB` interface and all concrete implementations used during block execution. It abstracts account balances, nonces, contract code, storage slots, transient storage (EIP-1153), access lists (EIP-2929), and log emission behind a single interface so the EVM, state transition, and VOPS subsystems can operate against interchangeable backends.

Two production backends are provided: `MemoryStateDB` (in-memory, journalled, used for testing and single-block execution) and `TrieBackedStateDB` (MPT-rooted, used for full node operation). Additional files implement stateless state (`stateless.go`), validity-only partial state (`validity_only.go`), sharded state for gigagas parallel execution (`sharded_state.go`), state expiry and migration (EIP-4444-adjacent), Block Access List tracking (`bals_engine.go`), endgame state, and misc EVM purges.

## Functionality

### Core interface

```go
type StateDB interface {
    CreateAccount(addr types.Address)
    AddBalance / SubBalance / GetBalance
    GetNonce / SetNonce
    GetCode / SetCode / GetCodeHash / GetCodeSize
    SelfDestruct / HasSelfDestructed
    GetState / SetState / GetCommittedState      // storage
    Exist / Empty
    Snapshot() int / RevertToSnapshot(id int)    // tx-level atomicity
    AddLog / GetLogs / SetTxContext
    AddRefund / SubRefund / GetRefund
    AddAddressToAccessList / AddSlotToAccessList  // EIP-2929
    GetTransientState / SetTransientState / ClearTransientStorage  // EIP-1153
    GetRoot / StorageRoot / Commit() (types.Hash, error)
}
```

### Concrete implementations

| Type | File | Purpose |
|------|------|---------|
| `MemoryStateDB` | `memory_statedb.go` | Pure in-memory, journalled |
| `TrieBackedStateDB` | `trie_backed.go` | MPT-rooted persistent state |
| `StatelessStateDB` | `stateless.go` | Witness-backed stateless execution |
| `ValidityOnlyState` | `validity_only.go` | VOPS partial-state execution |
| `ShardedState` | `sharded_state.go` | Parallel shard-partitioned state |

### Supporting types

- `journal` / `JournalManager` — change-log for snapshot/revert
- `AccessTracker` / `AccessEventTracker` — EIP-4762 access events
- `BALSEngine` — Block Access List state engine (EIP-7928)
- `ExpiryEngine` — state expiry scheduler
- `MigrationScheduler` / `TechDebtMigration` — MPT → binary trie migration
- `ConflictDetector` — parallel execution conflict detection
- `StatePrefetcher` / `TxPrefetcher` — read-ahead for state access
- `SnapshotGen` — snapshot generation from state root

## Usage

```go
db := state.NewMemoryStateDB()
db.CreateAccount(addr)
db.AddBalance(addr, big.NewInt(1e18))
snap := db.Snapshot()
db.SetState(addr, key, value)
db.RevertToSnapshot(snap)   // undo SetState
root, err := db.Commit()
```
