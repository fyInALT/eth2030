# light/cache - LRU proof cache for light clients

## Overview

Package `cache` provides a thread-safe LRU cache for Merkle proofs used by the
ETH2030 light client. It stores header, account, and storage proofs keyed by
block number, address, and storage slot, avoiding redundant network round-trips
when the same proof is needed across multiple validation steps.

Each entry carries a configurable TTL so proofs are automatically evicted once
the relevant state root is stale. The cache supports background prefetching to
warm entries for anticipated upcoming blocks before they are needed.

## Functionality

**Types**

- `ProofType` (`ProofTypeHeader`, `ProofTypeAccount`, `ProofTypeStorage`) -
  distinguishes proof categories.
- `CacheKey` - composite key: `BlockNumber uint64`, `Address types.Address`,
  `StorageKey types.Hash`, `Type ProofType`.
- `CachedProof` - stored entry: raw `Proof []byte`, `Value []byte`,
  `ExpiresAt time.Time`.
- `CacheStats` - counters: `Hits`, `Misses`, `Evictions`, `Inserts`,
  `MemoryUsed`; `HitRate() float64` helper.
- `ProofCache` - main cache struct (mutex-protected LRU + map).

**Constructor**

- `NewProofCache(maxSize int, ttl time.Duration) *ProofCache`

**Operations**

- `Get(key CacheKey) (*CachedProof, bool)` - returns a proof on hit; evicts and
  returns false if the entry is expired.
- `Put(key CacheKey, proof, value []byte)` - inserts or updates an entry; evicts
  the LRU entry if at capacity.
- `Evict(key CacheKey) bool` - explicit removal by key.
- `Prefetch(baseBlock uint64, addr, storageKey, proofType, count int, fetch func) ` -
  asynchronously fills the cache for a range of upcoming blocks using a caller-
  supplied fetch function.

**Accessors**

- `Stats() CacheStats`, `Len() int`, `MaxSize() int`

Parent package: [`light`](../)
