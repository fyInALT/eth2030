# sync/healer — Snap sync state trie healing

Detects and fills missing trie nodes after snap sync's bulk state download, supporting both the main state trie and per-account storage tries.

[← sync](../README.md)

## Overview

After snap sync downloads account leaves, storage slots, and bytecodes, the local state trie may be missing interior nodes. The healer package provides two healers with increasing sophistication.

`StateHealer` is the base implementation: it calls `StateWriter.MissingTrieNodes` to discover gaps, queues them as `HealingTask` entries, batches requests to a `SnapPeer`, and writes received node data back via `StateWriter.WriteTrieNode`. Failed tasks are retried up to `MaxHealRetries` before being permanently failed.

`TrieHealer` extends this with a min-heap priority queue (shallowest nodes first), per-account storage trie healing, checkpoint/resume support via `ResumeFromCheckpoint`, and a `SetCheckpointCallback` for persisting progress. `Run(peer)` orchestrates the full healing loop for both the main state and all registered storage tries.

`snap_interfaces.go` defines the `SnapPeer` (with `RequestTrieNodes`) and `StateWriter` interfaces shared across the package.

## Functionality

### Types

| Type | Purpose |
|------|---------|
| `StateHealer` | Basic gap detection, scheduling, and healing loop |
| `TrieHealer` | Priority-queue healer with storage tries, checkpointing, and resume |
| `HealingTask` | Path + root + retry count for a missing node |
| `TrieHealNode` | Depth-aware node for priority queue scheduling |
| `TrieHealCheckpoint` | Serialisable healing state for crash recovery |
| `HealingProgress` / `TrieHealProgress` | Detected/healed/failed counts and completion flag |

### Key Functions

- `NewStateHealer(root, writer)` / `DetectGaps()` / `ScheduleHealing()` / `ProcessHealingBatch(tasks, results)` / `Run(peer)`
- `NewTrieHealer(config, root, writer)` / `AddStorageTrie(accountHash, storageRoot)` / `DetectStateGaps()` / `DetectStorageGaps()` / `ProcessResults(batch, results)` / `CheckCompletion()` / `Run(peer)`
- `ResumeFromCheckpoint(cp)` / `SetCheckpointCallback(fn)`

## Usage

```go
healer := healer.NewStateHealer(pivotRoot, myStateWriter)
if err := healer.Run(mySnapPeer); err != nil {
    log.Fatal(err)
}
fmt.Println("healed nodes:", healer.Progress().NodesHealed)
```
