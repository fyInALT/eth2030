# cache

LRU hash cache mapping block numbers to block hashes (and reverse).

[← rawdb](../README.md)

## Overview

Package `cache` provides `HashCache`, a thread-safe LRU cache that maps block
numbers to block hashes and supports reverse lookups from hash to number. It is
used by the chain reader and the `BLOCKHASH` opcode implementation to avoid
repeated database queries for recently seen blocks.

## Functionality

### HashCacheConfig

```go
type HashCacheConfig struct {
    MaxEntries    int  // default: 1024
    EnableMetrics bool // whether to track hit/miss/eviction stats
}
```

`DefaultHashCacheConfig()` — returns the above defaults.

### HashCacheEntry

```go
type HashCacheEntry struct {
    Number    uint64
    Hash      types.Hash
    Timestamp int64 // unix timestamp when entry was added
}
```

### HashCacheStats

```go
type HashCacheStats struct {
    Hits      uint64
    Misses    uint64
    Evictions uint64
    Size      int
}
```

### HashCache

- `NewHashCache(config HashCacheConfig) *HashCache`
- `Put(number uint64, hash types.Hash)` — stores a mapping; evicts the LRU
  entry if at capacity.
- `Get(number uint64) (types.Hash, bool)` — forward lookup; moves entry to
  front (MRU).
- `GetByHash(hash types.Hash) (uint64, bool)` — reverse lookup.
- `Contains(number uint64) bool` — membership test without moving to MRU.
- `Remove(number uint64)` — evicts a specific entry.
- `Len() int` — current number of entries.
- `Purge()` — removes all entries.
- `Entries() []HashCacheEntry` — returns all entries in MRU order.
- `Stats() HashCacheStats` — returns hit/miss/eviction counters (atomic reads).

The cache uses a doubly-linked list for LRU ordering and two maps
(`byNumber`, `byHash`) for O(1) lookups in both directions.

## Usage

```go
cache := cache.NewHashCache(cache.DefaultHashCacheConfig())

// Populate from a new block.
cache.Put(block.NumberU64(), block.Hash())

// BLOCKHASH opcode lookup.
hash, ok := cache.Get(blockNumber)
if !ok {
    // fall back to database
}

// Reverse lookup.
num, ok := cache.GetByHash(knownHash)

stats := cache.Stats()
log.Printf("hit rate: %d/%d", stats.Hits, stats.Hits+stats.Misses)
```
