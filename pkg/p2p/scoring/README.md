# Package scoring

Per-peer reputation scoring with event-driven score adjustments, disconnect thresholds, and a concurrent `ScoreMap` for the full peer set.

## Overview

The `scoring` package tracks how well each connected peer is behaving. A `PeerScore` accumulates rewards and penalties based on observed events and clamps the value to `[-100, +100]`. When a peer's score drops at or below `ScoreDisconnect = -50.0`, the server should disconnect it. The `ScoreMap` type stores one `PeerScore` per peer ID and creates entries on first access. Additional types (`PeerReputation`, `PeerScorer`, `ReputationSystem`) in sibling files provide extended reputation tracking.

## Functionality

### PeerScore

- `NewPeerScore() *PeerScore`
- `Value() float64` / `ShouldDisconnect() bool`
- `GoodResponse()` (+1.0) — valid, timely response
- `BadResponse()` (−5.0) — invalid or useless response
- `Timeout()` (−10.0) — request timeout
- `UsefulBlock()` (+2.0) — announced block was needed
- `UselessBlock()` (−0.5) — announced block already known
- `HandshakeOK()` (+5.0) / `HandshakeFail()` (−20.0)
- `Stats() ScoreStats`

### ScoreMap

- `NewScoreMap() *ScoreMap`
- `Get(id string) *PeerScore` — creates entry if absent
- `Remove(id string)`
- `Len() int`
- `All() map[string]float64`

### Constants

`MaxScore = 100.0`, `MinScore = -100.0`, `DefaultScore = 0.0`, `ScoreDisconnect = -50.0`

Exported aliases for test packages: `ScoreHandshakeOK = +5.0`, `ScoreHandshakeFail = -20.0`.

## Usage

```go
scores := scoring.NewScoreMap()

// on successful handshake
scores.Get(peerID).HandshakeOK()

// on each good protocol response
scores.Get(peerID).GoodResponse()

// decide whether to disconnect
if scores.Get(peerID).ShouldDisconnect() {
    peer.Disconnect()
    scores.Remove(peerID)
}
```

[← p2p](../README.md)
