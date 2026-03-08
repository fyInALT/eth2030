# sync/downloader — Full-sync block chain downloader

Coordinates peer selection, batched header and body downloads, chain validation, and block announcement handling for full chain synchronization.

[← sync](../README.md)

## Overview

`ChainDownloader` is the primary type. It tracks a set of connected peers (up to `MaxPeers`), selecting the best peer by total difficulty for downloads. When `Download(ctx, from, to)` is called, it fetches headers in configurable batches via a `HeaderSource`, validates sequential block numbers and parent-hash linkage with `ValidateChain`, then fetches bodies via a `BodySource`. The entire operation is context-cancellable with per-request timeouts.

Block announcements from peers are accumulated in a ring buffer and used to update each peer's known head. Peer slots are managed with LRU-by-difficulty eviction so the set always favours the heaviest peers.

Supporting types include `SkeletonChain` (sparse header skeleton for parallel download), `BlockAnnouncer` (handles new-block gossip), `BodyDownloader` and `HeaderDownloader` (standalone sub-components), and `Fetcher` (concurrent fetch scheduler).

## Functionality

### Types

| Type | Purpose |
|------|---------|
| `ChainDownloader` | High-level block range downloader |
| `PeerInfo` | Peer identity, head hash/number, and total difficulty |
| `DownloadProgress` | Start/current/highest block and connected peer count |
| `HeaderSource` | Interface: `FetchHeaders(from, count)` |
| `BodySource` | Interface: `FetchBodies(hashes)` |

### Key Functions

- `NewChainDownloader(cfg)` / `SetSources(hs, bs)` / `AddPeer(p)` / `RemovePeer(id)`
- `SelectBestPeer()` / `Download(ctx, from, to)` / `ValidateChain(headers)` / `ProcessBatch(headers, bodies)`
- `HandleAnnouncement(peerID, hash, number)` / `Progress()` / `HighestPeerBlock()`

## Usage

```go
cd := downloader.NewChainDownloader(downloader.DefaultDownloadConfig())
cd.SetSources(myHeaderSource, myBodySource)
cd.AddPeer(downloader.PeerInfo{ID: "peer1", HeadNumber: 20000, TotalDifficulty: 1e18})
if err := cd.Download(ctx, 1, 20000); err != nil {
    log.Fatal(err)
}
```
