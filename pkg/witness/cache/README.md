# witness/cache — Block-scoped execution witness cache

## Overview

`witness/cache` provides a thread-safe, block-keyed LRU-style cache for execution witnesses used in stateless validation. Witnesses are stored by block hash and evicted in insertion order when the cache reaches its configured capacity, bounding memory use without external coordination.

The cache is intended for stateless nodes and validators that need fast lookups of previously received witnesses without re-downloading or recomputing them. All public methods are safe for concurrent use via `sync.RWMutex`.

## Functionality

**Types**

- `CachedWitness` — holds a witness for one block: `BlockHash`, `BlockNumber`, `StateRoot`, `AccountProofs` (`map[[32]byte][]byte`), `StorageProofs` (`map[[32]byte]map[[32]byte][]byte`), `CodeChunks`, and `Size` (estimated bytes).
- `WitnessCache` — the cache itself; created with `NewWitnessCache(maxBlocks int)` (default 128 if `maxBlocks <= 0`).
- `WitnessCacheStats` — snapshot of `Entries`, `TotalSize`, `Hits`, `Misses`.

**Methods on `WitnessCache`**

| Method | Description |
|---|---|
| `StoreWitness(blockHash [32]byte, w *CachedWitness)` | Insert or update; evicts oldest entry when at capacity. |
| `GetWitness(blockHash [32]byte) (*CachedWitness, bool)` | Lookup; increments hit/miss counters. |
| `HasWitness(blockHash [32]byte) bool` | Existence check without affecting stats. |
| `RemoveWitness(blockHash [32]byte)` | Explicit removal. |
| `PruneBeforeBlock(blockNumber uint64) int` | Remove all witnesses with `BlockNumber < blockNumber`; returns count removed. |
| `TotalSize() uint64` | Sum of `CachedWitness.Size` across all entries. |
| `Stats() *WitnessCacheStats` | Atomic snapshot of entries, size, hits, misses. |
| `MaxBlocks() int` | Configured capacity. |

## Usage

```go
c := cache.NewWitnessCache(256)

c.StoreWitness(blockHash, &cache.CachedWitness{
    BlockHash:   blockHash,
    BlockNumber: 1000,
    StateRoot:   stateRoot,
    Size:        4096,
})

if w, ok := c.GetWitness(blockHash); ok {
    // use w.AccountProofs, w.StorageProofs, etc.
}

// Free witnesses older than a checkpoint.
pruned := c.PruneBeforeBlock(900)

stats := c.Stats()
fmt.Printf("entries=%d hits=%d misses=%d\n", stats.Entries, stats.Hits, stats.Misses)
```

---

Parent package: [`witness`](../)
