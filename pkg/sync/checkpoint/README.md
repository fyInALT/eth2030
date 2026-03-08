# sync/checkpoint — Trusted checkpoint store and sync orchestrator

Stores verified chain checkpoints and coordinates checkpoint-based sync with header range request tracking and detailed phase state management.

[← sync](../README.md)

## Overview

`CheckpointStore` is the central component. It maintains a bounded set of `TrustedCheckpoint` entries (keyed by a deterministic hash over epoch, block number, block hash, and state root) and drives a sync state machine through the phases: idle → downloading-headers → downloading-bodies → downloading-receipts → processing → complete.

When `StartSync` is called, the store records the active checkpoint and target block, then exposes `UpdateProgress` to compute percentage completion and ETA as blocks are downloaded. Parallel header downloads are tracked via `HeaderRangeRequest` objects: the store enforces a maximum number of concurrent pending ranges and prevents overlapping requests.

## Functionality

### Types

| Type | Purpose |
|------|---------|
| `TrustedCheckpoint` | Verified chain point with block number, hash, state root, epoch, and source tag |
| `CheckpointStore` | Multi-checkpoint registry with sync state machine and range tracking |
| `CheckpointSyncProgress` | Snapshot of current sync state, block progress, and ETA |
| `HeaderRangeRequest` | Batch header download request with completion tracking |
| `SyncState` | Enum: `StateCheckpointIdle` through `StateCheckpointComplete` |

### Key Functions

- `NewCheckpointStore(config)` / `RegisterCheckpoint(cp)` / `GetHighestCheckpoint()`
- `StartSync(cp, targetBlock)` / `TransitionState(next)` / `UpdateProgress(current, headers, bodies, receipts)`
- `CreateRangeRequest(from, to, peerID)` / `CompleteRangeRequest(id, headers)`
- `VerifyCheckpoint(cp)` / `Progress()` / `State()` / `Reset()`

## Usage

```go
store := checkpoint.NewCheckpointStore(checkpoint.DefaultCheckpointStoreConfig())
store.RegisterCheckpoint(checkpoint.TrustedCheckpoint{
    BlockNumber: 21000000, BlockHash: hash, StateRoot: root, Source: "embedded",
})
store.StartSync(*store.GetHighestCheckpoint(), headNumber)
// As headers arrive:
store.UpdateProgress(currentBlock, headersDown, 0, 0)
```
