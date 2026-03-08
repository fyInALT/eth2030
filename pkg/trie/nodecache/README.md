# trie/nodecache — LRU trie node cache

Thread-safe LRU cache for RLP-encoded trie nodes, keyed by Keccak-256 hash, with byte-level memory accounting and hit/miss/eviction statistics.

[← trie](../README.md)

## Overview

`TrieCache` stores raw trie node bytes up to a configured `maxSize` (in bytes). It maintains a doubly-linked list for O(1) LRU ordering: `Get` moves the accessed entry to the front; `Put` inserts at the front, evicting the tail if the byte budget would be exceeded. All reads return copies to prevent external mutation.

`Prune(targetSize)` actively evicts least-recently-used entries until total byte usage is at or below `targetSize`, returning the count of evicted entries. `Stats()` returns a `CacheStats` snapshot with cumulative hit/miss/eviction counts, current byte size, and entry count.

## Functionality

### Types

| Type | Purpose |
|------|---------|
| `TrieCache` | LRU cache: `Get`/`Put`/`Delete`/`Prune` with byte-level tracking |
| `CacheStats` | Hits, misses, evictions, current size (bytes), entry count |

### Key Functions

- `NewTrieCache(maxSize int)` — maxSize in bytes; 0 means unlimited
- `Get(hash [32]byte)` / `Put(hash [32]byte, data []byte)` / `Delete(hash [32]byte)`
- `Len()` / `Size()` / `Prune(targetSize)` / `Stats()` / `HitRate()` / `Reset()`

## Usage

```go
cache := nodecache.NewTrieCache(64 * 1024 * 1024) // 64 MiB
cache.Put(nodeHash, rlpData)

data, ok := cache.Get(nodeHash)
stats := cache.Stats()
fmt.Printf("hit rate: %.1f%%\n", cache.HitRate()*100)
```
