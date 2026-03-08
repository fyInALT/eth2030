# Package broadcast

Block and transaction broadcast, EIP-7702 SetCode authorization dissemination, and STARK mempool tick propagation over the P2P gossip network.

## Overview

The `broadcast` package implements message dissemination strategies for ETH2030's peer-to-peer network. Block announcements are propagated using a sqrt(n) fanout to limit bandwidth while ensuring wide coverage. EIP-7702 SetCode authorization tuples are disseminated via topic-based gossip with per-authority rate limiting and bloom-filter deduplication to prevent spam. Encrypted mempool tick messages are forwarded over the `STARKMempoolTick` gossip topic.

Block gossip tracks a seen-block cache so duplicate announcements are filtered. The SetCode broadcaster validates secp256k1 signatures before queuing and enforces a per-authority epoch rate cap (default 16 per epoch).

## Functionality

### Types

- `BlockGossipHandler` — manages block announcement receipt and propagation
  - `NewBlockGossipHandler(config BlockGossipConfig) *BlockGossipHandler`
  - `HandleAnnouncement(ann BlockAnnouncement) error`
  - `PropagateBlock(hash types.Hash, number uint64) []string` — sqrt(n) fanout
  - `AddPeer(peerID string)` / `RemovePeer(peerID string)`
  - `SeenBlock(hash types.Hash) bool`
  - `Stats() GossipStats`
  - `RecentAnnouncements(limit int) []BlockAnnouncement`

- `SetCodeBroadcaster` — EIP-7702 authorization gossip
  - `NewSetCodeBroadcaster(chainID *big.Int) *SetCodeBroadcaster`
  - `Submit(msg *SetCodeMessage) error`
  - `DrainPending() []*SetCodeMessage`
  - `AddHandler(handler SetCodeGossipHandler)`
  - `ResetEpoch()` — clears bloom filter and rate counters

- `MempoolBroadcaster` — STARK mempool tick gossip
  - `NewMempoolBroadcaster(tm *gossip.TopicManager) *MempoolBroadcaster`
  - `GossipMempoolStarkTick(data []byte) error`

- `ValidateSetCodeAuth(msg *SetCodeMessage) bool` — secp256k1 signature recovery
- `ValidateSetCodeAuthWithChainID(msg *SetCodeMessage, localChainID *big.Int) bool`

### Supporting types

`BlockAnnouncement`, `GossipStats`, `BlockGossipConfig`, `SetCodeMessage`, `SetCodeGossipHandler` (interface), `SetCodeGossipHandlerFunc` (adapter).

## Usage

```go
// Block gossip with sqrt(n) fanout
h := broadcast.NewBlockGossipHandler(broadcast.DefaultBlockGossipConfig())
h.AddPeer("peer1")
h.AddPeer("peer2")
err := h.HandleAnnouncement(broadcast.BlockAnnouncement{
    Hash: blockHash, Number: 100, PeerID: "peer1",
})
selected := h.PropagateBlock(blockHash, 100)

// EIP-7702 SetCode broadcast
b := broadcast.NewSetCodeBroadcaster(chainID)
b.AddHandler(broadcast.SetCodeGossipHandlerFunc(func(msg *broadcast.SetCodeMessage) error {
    // forward to peers
    return nil
}))
b.Submit(msg)
```

[← p2p](../README.md)
