# das/blobpool — Sparse blob pool (EIP-8070 / Glamsterdam)

[← das](../README.md)

## Overview

This package implements the sparse blob pool for PeerDAS as specified by EIP-8070 (Glamsterdam). Rather than storing all blobs locally, the pool retains only a configurable fraction based on a sparsity filter: a blob is stored only if the first 8 bytes of its hash satisfy `hash_prefix mod sparsity == 0`. This reduces disk and memory requirements for nodes participating in the PeerDAS network that do not need to serve the full blob dataset.

The pool also supports slot-based pruning to evict blobs older than a configurable retention window, a write-ahead log (WAL) path for durability, and per-pool statistics tracking accepted, rejected, and pruned blob counts.

## Functionality

**Types**
- `SparseBlobPool` — thread-safe pool with sparsity filtering
- `PoolStats` — `Stored`, `TotalAdded`, `Pruned`, `Rejected uint64`

**Construction**
- `NewSparseBlobPool(sparsity uint64) *SparseBlobPool` — panics if sparsity is 0; sparsity=1 stores all blobs

**Operations**
- `(p *SparseBlobPool) AddBlob(blobHash [32]byte, data []byte, slot uint64) bool` — returns true if stored
- `(p *SparseBlobPool) GetBlob(blobHash [32]byte) ([]byte, bool)`
- `(p *SparseBlobPool) HasBlob(blobHash [32]byte) bool`
- `(p *SparseBlobPool) PruneOlderThan(slot uint64) int` — removes blobs from slots older than given slot
- `(p *SparseBlobPool) Stats() PoolStats`
- `(p *SparseBlobPool) Sparsity() uint64`
- `(p *SparseBlobPool) Size() int` — current number of stored blobs

## Usage

```go
// Store roughly 25% of blobs (sparsity=4)
pool := blobpool.NewSparseBlobPool(4)

stored := pool.AddBlob(blobHash, blobData, slotNumber)
data, ok := pool.GetBlob(blobHash)

// Prune blobs older than the current slot minus retention
pool.PruneOlderThan(currentSlot - retentionSlots)
```
