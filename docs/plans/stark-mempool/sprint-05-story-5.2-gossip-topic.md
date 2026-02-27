# Sprint 5, Story 5.2 — P2P Gossip Topic for STARK Mempool Ticks

**Sprint goal:** Add network-level propagation for STARK mempool ticks.
**Files modified:** `pkg/p2p/gossip_topics.go`

## Overview

STARK mempool aggregation only works locally without a gossip topic. Ticks must propagate between peers for recursive accumulation.

## Gap (GAP-STARK3 + AUDIT-6)

**Severity:** CRITICAL
**File:** `pkg/p2p/gossip_topics.go`

**Round 1:** No gossip topic existed for STARK mempool ticks.
**Round 2:** Topic was added but handler registration pattern was undocumented.

## Implement

### Step 1: Add topic constant

```go
// pkg/p2p/gossip_topics.go
const (
    // ... existing topics ...
    SyncCommitteeContribution
    // STARKMempoolTick propagates recursive STARK mempool aggregation ticks.
    STARKMempoolTick
)
```

### Step 2: Register in topic names map

```go
var gossipTopicNames = map[GossipTopic]string{
    // ... existing entries ...
    STARKMempoolTick: "stark_mempool_tick",
}
```

### Step 3: Document handler registration pattern

```go
// Handler registration: at application startup, the node should call
// TopicManager.Subscribe(STARKMempoolTick, handler) where the handler
// deserialises the tick via MempoolAggregationTick.UnmarshalBinary,
// validates the STARK proof, and calls STARKAggregator.MergeTick.
```

### Expected handler pattern (future implementation)

```go
func handleSTARKMempoolTick(msg *GossipMessage) error {
    var tick txpool.MempoolAggregationTick
    if err := tick.UnmarshalBinary(msg.Data); err != nil {
        return err
    }
    // Verify serialized size <= MaxTickSize (128KB).
    if len(msg.Data) > txpool.MaxTickSize {
        return errors.New("tick exceeds bandwidth limit")
    }
    return aggregator.MergeTick(&tick)
}
```

## ethresear.ch Spec Reference

> Nodes gossip aggregation ticks every 500ms via a dedicated gossip topic. Upon receiving a tick, nodes verify the STARK proof and merge the validated transactions into their local set.

## Codebase Locations

| File | Line | Purpose |
|------|------|---------|
| `pkg/p2p/gossip_topics.go` | 32 | STARKMempoolTick constant |
| `pkg/p2p/gossip_topics.go` | 50 | Topic name in gossipTopicNames map |
