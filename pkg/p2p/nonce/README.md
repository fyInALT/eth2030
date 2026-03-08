# Package nonce

EIP-8077 announce-nonce support — per-peer LRU caches of nonce records with TTL expiry for lightweight block validation before full download.

## Overview

The `nonce` package implements the nonce announcement protocol from the J+ era roadmap. Peers broadcast a `(blockHash, nonce)` pair before sending a full block, allowing receivers to perform a lightweight proof-of-work check without downloading the entire block body. `NonceAnnouncer` maintains a per-peer `NonceCache` (LRU, capacity 1024, TTL 5 min). Cache entries are indexed by block hash; `PruneStale` reclaims expired entries across all peers.

## Functionality

- `NonceAnnouncer` — global store of per-peer nonce caches
  - `NewNonceAnnouncer() *NonceAnnouncer`
  - `NewNonceAnnouncerWithConfig(cacheSize int, ttl time.Duration, maxPeers int) *NonceAnnouncer`
  - `AnnounceNonce(peerID string, blockHash types.Hash, nonce uint64) error`
  - `ValidateNonce(peerID string, blockHash types.Hash, nonce uint64) bool`
  - `GetPeerNonces(peerID string) []NonceRecord`
  - `HasNonce(peerID string, blockHash types.Hash) bool`
  - `PruneStale(maxAge time.Duration) int`
  - `RemovePeer(peerID string)`
  - `PeerCount() int` / `RecordCount() int`

- `NonceCache` — per-peer LRU cache
  - `NewNonceCache(maxSize int, ttl time.Duration) *NonceCache`

- `NonceRecord` — `{PeerID, BlockHash, Nonce uint64, Timestamp}`

- Default constants: `DefaultNonceCacheSize=1024`, `DefaultNonceTTL=5min`, `DefaultMaxPeers=256`

## Usage

```go
announcer := nonce.NewNonceAnnouncer()

// peer announces a nonce for an upcoming block
announcer.AnnounceNonce(peerID, blockHash, nonce)

// validate before downloading the full block
if !announcer.ValidateNonce(peerID, blockHash, receivedNonce) {
    // mismatch — penalise peer
}

// periodic cleanup
announcer.PruneStale(5 * time.Minute)
```

[← p2p](../README.md)
