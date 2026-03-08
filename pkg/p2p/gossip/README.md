# Package gossip

Beacon chain pub/sub gossip: topic management, message validation, deduplication, mesh scoring, and per-topic scoring for the consensus-layer P2P network.

## Overview

The `gossip` package implements the Ethereum consensus-layer gossip protocol. The central type is `TopicManager`, which tracks subscribed topics, dispatches incoming messages to registered `TopicHandler` callbacks, deduplicates messages via a SHA-256-based 20-byte `MessageID` cache with TTL expiry, and tracks per-topic scoring metrics. The lower-level `GossipManager` provides a topic-agnostic publish/subscribe bus with peer banning, score-based gating, and fanout.

Topic parameters follow the consensus spec: mesh target D=8, Dlo=6, Dhi=12, heartbeat 700 ms, seen-TTL 384 s.

## Functionality

### GossipTopic enumeration

`BeaconBlock`, `BeaconAggregateAndProof`, `VoluntaryExit`, `ProposerSlashing`, `AttesterSlashing`, `BlobSidecar`, `SyncCommitteeContribution`, `STARKMempoolTick`, `PQAggRequest`, `PQAggResult`, `ProposerPreferences`.

### TopicManager

- `NewTopicManager(params TopicParams) *TopicManager`
- `Subscribe(topic GossipTopic, handler TopicHandler) error`
- `Unsubscribe(topic GossipTopic) error`
- `Publish(topic GossipTopic, data []byte) error` — local publish with deduplication
- `Deliver(topic GossipTopic, data []byte, isValid bool) error` — inbound from peer
- `TopicScore(topic GossipTopic) (TopicScoreSnapshot, bool)`
- `UpdatePeerTopicScore(topic GossipTopic, peerID string, delta float64)`
- `PruneSeenMessages() int` — periodic TTL-based cleanup
- `DefaultTopicParams() TopicParams`

### GossipManager

- `NewGossipManager(config GossipConfig) *GossipManager`
- `PublishMessage(topic string, data []byte) error`
- `Subscribe(topic string) *GossipSubscription`
- `Unsubscribe(sub *GossipSubscription) error`
- `ValidateMessage(msg *GossipMessage) error`
- `BanPeer(peerID types.Hash, reason string, duration uint64)`
- `UpdatePeerScore(peerID types.Hash, delta float64)`

### Message ID computation

- `ComputeMessageID(decompressedData []byte) MessageID` — SHA-256(domain + data)[:20]
- `ComputeInvalidMessageID(rawData []byte) MessageID`
- `ParseGossipTopic(name string) (GossipTopic, error)`
- `(GossipTopic).TopicString(forkDigest string) string` — `/eth2/<digest>/<name>/ssz_snappy`

## Usage

```go
tm := gossip.NewTopicManager(gossip.DefaultTopicParams())
tm.Subscribe(gossip.BeaconBlock, func(topic gossip.GossipTopic, msgID gossip.MessageID, data []byte) {
    // deserialise and process the signed beacon block
})
tm.Subscribe(gossip.STARKMempoolTick, func(_ gossip.GossipTopic, _ gossip.MessageID, data []byte) {
    // validate STARK proof and merge tick
})
// publish a beacon block we have produced
tm.Publish(gossip.BeaconBlock, sszBlock)
```

[← p2p](../README.md)
