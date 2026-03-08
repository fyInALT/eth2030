# state/snapshot — Layered state snapshot tree

[← state](../README.md)

## Overview

Package `snapshot` implements a journalled, dynamic state dump as a layered structure. It consists of one persistent base layer (`diskLayer`) backed by a key-value store, on top of which arbitrarily many in-memory diff layers (`diffLayer`) are stacked. The goal is fast account and storage access without expensive multi-level trie traversals, and sorted iteration for snap-sync.

Layers form a linear chain. When a diff layer accumulates enough history, the `LayerMerger` collapses adjacent diffs. A `CompactionScheduler` triggers periodic merges. Iterators (`AccountIterator`, `StorageIterator`) walk across layers in sorted key order.

## Functionality

### Core interface

```go
type Snapshot interface {
    Root() types.Hash
    Account(hash types.Hash) (*types.Account, error)
    Storage(accountHash, storageHash types.Hash) ([]byte, error)
}
```

### Tree

```go
func NewTree(db snapshotDB, diskRoot types.Hash) *Tree

func (t *Tree) Snapshot(blockRoot types.Hash) Snapshot
func (t *Tree) Update(blockRoot, parentRoot types.Hash,
    accounts map[types.Hash][]byte,
    storage  map[types.Hash]map[types.Hash][]byte) error
func (t *Tree) Cap(root types.Hash, layers int) error
func (t *Tree) AccountIterator(root, seek types.Hash) (AccountIterator, error)
func (t *Tree) StorageIterator(root, account, seek types.Hash) (StorageIterator, error)
```

### Layers

| Type | File | Description |
|------|------|-------------|
| `diskLayer` | `disklayer.go` | Persistent base layer |
| `diffLayer` | `difflayer.go` | In-memory incremental diff |
| `LayerMerger` | `layer_merger.go` | Merges adjacent diffs |
| `CompactionScheduler` | `compaction_scheduler.go` | Schedules periodic compaction |
| `AccountCache` | `account_cache.go` | Per-layer LRU account cache |

### Errors

`ErrSnapshotStale`, `ErrNotFound`

## Usage

```go
tree := snapshot.NewTree(db, genesisRoot)

// Add a new block's state changes on top.
tree.Update(block.Root(), block.ParentHash(), accounts, storage)

// Read an account from the tip snapshot.
snap := tree.Snapshot(block.Root())
acct, err := snap.Account(addrHash)
```
