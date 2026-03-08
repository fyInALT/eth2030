# Package peermgr

Low-level peer set management: `Peer` struct, `PeerSet`, `ManagedPeerSet`, and an advanced peer manager for concurrent peer tracking and request correlation.

## Overview

The `peermgr` package defines the core peer data model used throughout the ETH2030 P2P stack. A `Peer` holds its identity, negotiated capabilities, best-known block head and total difficulty, and maps for last-response and delivered-response correlation (used by the request-response layer). `PeerSet` is an unbounded thread-safe map of peers; `ManagedPeerSet` adds a configurable capacity cap and close semantics that clear all entries. `AdvancedPeerManager` (in `adv_peer_manager.go`) provides higher-level lifecycle hooks.

## Functionality

### Peer

- `NewPeer(id, remoteAddr string, caps []wire.Cap) *Peer`
- `ID() string` / `RemoteAddr() string` / `Caps() []wire.Cap`
- `Head() types.Hash` / `TD() *big.Int` / `Version() uint32` / `HeadNumber() uint64`
- `SetHead(hash types.Hash, td *big.Int)` / `SetVersion(v uint32)` / `SetHeadNumber(num uint64)`
- `SetLastResponse(code uint64, value any)` / `LastResponse(code uint64) any`
- `DeliverResponse(requestID uint64, value any)` / `GetDeliveredResponse(requestID uint64) (any, bool)`

### PeerSet

- `NewPeerSet() *PeerSet`
- `Register(p *Peer) error` / `Unregister(id string) error`
- `Peer(id string) *Peer` / `Len() int`
- `BestPeer() *Peer` — highest total difficulty
- `Peers() []*Peer`

### ManagedPeerSet

- `NewManagedPeerSet(maxPeers int) *ManagedPeerSet`
- `Add(p *Peer) error` — returns `ErrMaxPeers` when full
- `Remove(id string) error` / `Get(id string) *Peer`
- `Len() int` / `Peers() []*Peer`
- `Close()` — marks closed and clears all entries

## Usage

```go
ps := peermgr.NewManagedPeerSet(50)
p := peermgr.NewPeer(enodeID, "192.168.1.2:30303", caps)
if err := ps.Add(p); err == peermgr.ErrMaxPeers {
    // evict someone first
}
best := ps.Get(enodeID)
best.SetHead(headHash, headTD)
```

[← p2p](../README.md)
