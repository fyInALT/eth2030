# sync/statesync — State trie download manager

Coordinates batched state trie downloads, Merkle proof validation, and pause/resume lifecycle for state synchronization.

[← sync](../README.md)

## Overview

`StateSyncManager` manages the download of the full Ethereum state trie from a target state root. It divides the 256-bit account key space into ranges and issues `RequestStateRange` calls, each returning a `StateRangeResponse` containing up to `BatchSize` accounts with boundary proofs.

`ValidateStateRange` verifies that proof nodes are non-trivially hashed, that accounts are in ascending key order, and that the response is internally consistent. The manager supports pause/resume (`PauseSync`/`ResumeSync`) and tracks progress atomically: accounts synced, bytes downloaded, pending request count.

The `snap_interfaces.go` file defines shared snap-sync interfaces used by both this package and the `healer` package.

## Functionality

### Types

| Type | Purpose |
|------|---------|
| `StateSyncManager` | Multi-range state download coordinator |
| `StateSyncConfig` | `MaxConcurrent`, `BatchSize`, `RetryAttempts`, `TargetRoot` |
| `StateAccount` | Downloaded account: hash, nonce, balance, storage/code roots |
| `StateRangeResponse` | Slice of `StateAccount` with proofs and continuation key |
| `SSMProgress` | Accounts synced, bytes downloaded, current phase |

### Key Functions

- `NewStateSyncManager(config)` / `StartSync(targetRoot)` / `StopSync()`
- `PauseSync()` / `ResumeSync()` / `IsSyncing()` / `IsPaused()`
- `RequestStateRange(startKey, endKey)` / `ValidateStateRange(resp)`
- `Progress()` / `PendingRequests()` / `TotalRequests()`

## Usage

```go
mgr := statesync.NewStateSyncManager(statesync.DefaultStateSyncConfig())
mgr.StartSync(targetStateRoot)

resp, err := mgr.RequestStateRange(startKey, endKey)
if err := mgr.ValidateStateRange(resp); err != nil {
    log.Fatal(err)
}
fmt.Println("accounts synced:", mgr.Progress().AccountsSynced)
```
