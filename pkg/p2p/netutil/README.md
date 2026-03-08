# Package netutil

Network utility primitives: connection limiting with per-subnet caps, sliding-window bandwidth tracking with priority-class allocation, IBLT-based set reconciliation, and connection deduplication.

## Overview

The `netutil` package provides three independent building blocks for connection management and network accounting. `ConnLim` enforces per-subnet (/16 and /24), per-direction (inbound/outbound), and global connection limits, with reserved slots for static and trusted peers, inbound rate limiting, and connection deduplication. `BandwidthTracker` measures per-peer and global upload/download rates using a sliding-window bucketed counter and enforces priority-class bandwidth shares (consensus 40%, blocks 30%, transactions 20%, blobs 10%). `IBLT` and `SetReconciliation` implement Invertible Bloom Lookup Tables for efficient set difference discovery between peers.

## Functionality

### ConnLim

- `NewConnLim(cfg ConnLimConfig) *ConnLim`
- `CanConnect(peerID string, remoteIP net.IP, dir ConnDirection, isStatic, isTrusted bool) error`
- `AddConn(...)  error` / `RemoveConn(peerID string)`
- `RecordAttempt(peerID string)` — starts 5 s dedup window
- `EvictLowestPriority() string` — returns peerID of best eviction candidate
- `ConnCount()` / `InboundConnCount()` / `OutboundConnCount()`
- `Subnet16Count(ip net.IP) int` / `Subnet24Count(ip net.IP) int`
- `AvailableSlots() int`
- `DefaultConnLimConfig() ConnLimConfig`

### BandwidthTracker

- `NewBandwidthTracker(config BandwidthTrackerConfig) *BandwidthTracker`
- `RegisterPeer(peerID string)` / `RemovePeer(peerID string)`
- `RecordUpload(peerID string, bytes int64, priority int) error`
- `RecordDownload(peerID string, bytes int64, priority int) error`
- `PeerStats(peerID string) (BandwidthStats, error)`
- `GlobalStats() GlobalBandwidthStats`
- `UploadRate() float64` / `DownloadRate() float64`
- Priority constants: `BandwidthPriorityConsensus`, `BandwidthPriorityBlocks`, `BandwidthPriorityTxs`, `BandwidthPriorityBlobs`

## Usage

```go
cl := netutil.NewConnLim(netutil.DefaultConnLimConfig())
if err := cl.AddConn(peerID, remoteIP, netutil.ConnInbound, false, false); err != nil {
    // reject: limit reached
}

bt := netutil.NewBandwidthTracker(netutil.DefaultBandwidthTrackerConfig())
bt.RegisterPeer(peerID)
bt.RecordUpload(peerID, int64(len(data)), netutil.BandwidthPriorityBlocks)
```

[← p2p](../README.md)
