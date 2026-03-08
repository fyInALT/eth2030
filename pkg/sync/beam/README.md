# sync/beam — On-demand stateless block execution (beam sync)

Implements beam sync: fetching Ethereum state on-demand from peers instead of downloading the full state trie before executing blocks. Enables block execution to start immediately after header sync.

[← sync](../README.md)

## Overview

`BeamSync` wraps a `BeamStateFetcher` with an in-process cache. When the EVM needs an account or storage slot that is not cached locally, `BeamSync` fetches it from a peer in real time and stores the result. A `BeamPrefetcher` fires background goroutines to warm the cache from transaction access lists before execution begins.

`BeamStateSync` is the higher-level component used for full stateless execution. It accepts execution witnesses (`ExecutionWitness`) from a `WitnessFetcher`, uses them to warm an LRU account/storage cache, and exposes `ExecuteBlock` for witness-backed state root computation. It also monitors cache hit rates and consecutive miss counts, triggering an automatic fallback to full sync when the hit rate drops below a configured threshold.

`OnDemandDB` bridges `BeamSync` into the state-read interface expected by the EVM (`GetBalance`, `GetNonce`, `GetCode`, `GetStorage`).

## Functionality

### Types

| Type | Purpose |
|------|---------|
| `BeamSync` | Per-address/slot on-demand fetch with local cache |
| `BeamPrefetcher` | Background prefetch of accounts and storage keys |
| `OnDemandDB` | EVM-facing state reader backed by `BeamSync` |
| `BeamStateSync` | Witness-based execution with LRU cache and fallback detection |
| `StatePrefill` | Queues access-list prefetches before block execution |
| `ExecutionWitness` | Block-level state snapshot (accounts, storage, bytecodes) |

### Key Functions

- `NewBeamSync(fetcher)` / `FetchAccount(addr)` / `FetchStorage(addr, key)` / `CacheHitRate()`
- `NewBeamStateSync(fetcher, cacheConfig, fallbackConfig)` / `ExecuteBlock(blockRoot)` / `ShouldFallback()` / `ResetFallback()`
- `NewOnDemandDB(beam)` — returns a `GetBalance`/`GetNonce`/`GetCode`/`GetStorage` provider
- `DefaultBeamCacheConfig()` / `DefaultFallbackConfig()`

## Usage

```go
bs := beam.NewBeamSync(myNetworkFetcher)
bs.Prefetcher().PrefetchAccounts([]types.Address{addr1, addr2})
bs.Prefetcher().Wait()

bal, err := beam.NewOnDemandDB(bs).GetBalance(addr1)
```
