# state/acctcache — Thread-safe LRU account cache

[← state](../README.md)

## Overview

Package `acctcache` provides a bounded, thread-safe LRU cache for `types.Account` objects. It sits between the state layer and the underlying trie or database to reduce redundant account lookups during block execution.

All values are deep-copied on both read and write to prevent callers from mutating cached state. Hit/miss counters are tracked atomically and accessible via `Stats()`.

## Functionality

```go
type AccountCache struct { ... }

func NewAccountCache(maxSize int) *AccountCache
func (c *AccountCache) Get(addr types.Address) (*types.Account, bool)
func (c *AccountCache) Put(addr types.Address, acct *types.Account)
func (c *AccountCache) Delete(addr types.Address)
func (c *AccountCache) Len() int
func (c *AccountCache) Clear()
func (c *AccountCache) Stats() CacheStats

type CacheStats struct {
    Hits   uint64
    Misses uint64
}
```

Eviction policy: least-recently-used. When `Len() == maxSize` and a new entry is inserted, the tail (LRU) entry is removed.

## Usage

```go
cache := acctcache.NewAccountCache(1024)
cache.Put(addr, &types.Account{Nonce: 1, Balance: big.NewInt(100)})

if acct, ok := cache.Get(addr); ok {
    fmt.Println(acct.Nonce) // 1
}

stats := cache.Stats()
fmt.Printf("hits=%d misses=%d\n", stats.Hits, stats.Misses)
```
