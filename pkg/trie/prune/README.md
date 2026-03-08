# trie/prune — State and trie node pruning

Coordinates garbage collection of old state trie snapshots and unreachable trie nodes using a sliding window of recent state roots and a permanently-alive set.

[← trie](../README.md)

## Overview

**`StatePruner`** (`state_pruner.go`) manages which complete state snapshots (identified by block number + state root) should be retained. It keeps a sliding window of the most recent `maxRecent` state roots and a separate set of roots explicitly marked alive (`MarkAlive`) — used for checkpoint roots, finalized roots, or snap sync pivots. When a new root is added (`AddRoot`) and the window overflows, the oldest non-alive roots are evicted. `Prune(keepRecent)` can be called explicitly to evict roots outside the desired retention window while always preserving alive-marked roots. `RetainedRoots()` returns the full set for driving the underlying node database's garbage collector.

**`TriePruner`** (`trie_pruner.go`) operates at the node level, using a Bloom filter to track reachable node hashes and purging anything not referenced from a retained root.

## Functionality

### Types

| Type | Purpose |
|------|---------|
| `StatePruner` | Sliding-window state-root retention manager |
| `StateRoot` | Block number + root hash pair |
| `StatePrunerStats` | Window size, alive count, pruned total, head block |

### Key Functions (`StatePruner`)

- `NewStatePruner(maxRecent)` / `AddRoot(block, root)` / `MarkAlive(root)` / `UnmarkAlive(root)`
- `IsAlive(root)` / `Prune(keepRecent)` / `RetainedRoots()` / `RecentRoots()` / `AliveRoots()`
- `HeadRoot()` / `WindowSize()` / `PrunedTotal()` / `Stats()` / `Stop()` / `Reset()`

## Usage

```go
pruner := prune.NewStatePruner(128)
pruner.MarkAlive(checkpointRoot)

evicted := pruner.AddRoot(blockNum, stateRoot)
// evicted contains roots that can now be purged from the node database

toKeep := pruner.RetainedRoots()
```
