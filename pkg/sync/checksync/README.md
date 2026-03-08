# sync/checksync — Checkpoint sync state machine

Lightweight checkpoint-based sync that starts the chain from a trusted block instead of genesis, implementing the Glamsterdam fast-confirmation roadmap item.

[← sync](../README.md)

## Overview

`CheckpointSyncer` manages a single trusted `Checkpoint` (epoch, block number, block hash, state root) and drives a three-state machine: idle → syncing → complete. It validates checkpoint fields for internal consistency (non-zero hash and state root, block number > 0, block number ≥ epoch × 32) and caches verified checkpoints by their deterministic Keccak256 hash.

Progress tracking records the current and target block numbers and recomputes a percentage as `UpdateProgress` is called. The syncer does not perform any network I/O itself; it is designed to be called by the main sync coordinator which drives header downloads.

This package is distinct from `checkpoint/` in that it provides a simpler, self-contained sync state machine rather than a full multi-checkpoint store with range request management.

## Functionality

### Types

| Type | Purpose |
|------|---------|
| `Checkpoint` | Trusted sync point: epoch, block number, block hash, state root |
| `CheckpointSyncer` | Idle/syncing/complete state machine with progress tracking |
| `CheckpointProgress` | Current block, target block, start time, percentage |

### Key Functions

- `NewCheckpointSyncer(config)` / `SetCheckpoint(cp)` / `VerifyCheckpoint(cp)`
- `SetTarget(blockNumber)` / `SyncFromCheckpoint()` / `UpdateProgress(currentBlock)`
- `IsComplete()` / `Progress()` / `Reset()` / `IsVerified(cp)`

## Usage

```go
cs := checksync.NewCheckpointSyncer(checksync.DefaultCheckpointConfig())
cs.SetCheckpoint(checksync.Checkpoint{
    Epoch: 500, BlockNumber: 16000, BlockHash: bh, StateRoot: sr,
})
cs.SetTarget(20000)
cs.SyncFromCheckpoint()
// As blocks are downloaded:
cs.UpdateProgress(currentBlock)
```
