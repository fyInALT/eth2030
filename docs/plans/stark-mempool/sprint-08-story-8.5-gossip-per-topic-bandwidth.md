# Sprint 8, Story 8.5 — Per-Topic Gossip Bandwidth Enforcement

**Sprint goal:** Enforce the 128KB per-topic message size limit at the gossip layer.
**Files modified:** `pkg/p2p/gossip_topics.go`
**Files tested:** `pkg/p2p/gossip_topics_test.go`

## Overview

The ethresear.ch proposal specifies a 128KB bandwidth budget per STARK mempool tick. Story 5.3 enforced this at the `MergeTick()` application layer, but the gossip layer itself had no per-topic size enforcement. An oversized message could still be published and delivered on the `STARKMempoolTick` topic, consuming network bandwidth before the application-layer check rejected it.

## Gap (GAP-STARK5)

**Severity:** LOW
**File:** `pkg/p2p/gossip_topics.go` — `Publish()` at line 278 and `Deliver()` at line 317

**Evidence:** Both `Publish()` and `Deliver()` only checked the global `MaxPayloadSize` (10 MiB). The `STARKMempoolTick` topic had no specific limit. A 5 MB STARK tick would pass the gossip layer and only be rejected later at `MergeTick()`.

**Impact:** Network-level amplification attack — a peer could flood the STARK mempool gossip topic with messages up to 10 MB each. Defense-in-depth requires the gossip layer to reject oversized messages before handler dispatch.

## ethresear.ch Spec Reference

> The bandwidth budget for mempool aggregation is 128KB × peers per tick interval (500ms).

The 128KB limit should be enforced as close to the wire as possible — at the gossip layer, not just the application handler.

## Implement

### Step 1: Add TopicMessageSizeLimit map

```go
// TopicMessageSizeLimit defines per-topic maximum message sizes.
// Topics not in this map use the global MaxPayloadSize limit.
var TopicMessageSizeLimit = map[GossipTopic]int{
    STARKMempoolTick: 128 * 1024, // 128KB per ethresear.ch
}

var ErrTopicMsgTooLarge = errors.New("gossip_topics: message exceeds per-topic size limit")
```

### Step 2: Enforce in Publish()

After the global `MaxPayloadSize` check:

```go
if limit, ok := TopicMessageSizeLimit[topic]; ok && len(data) > limit {
    return ErrTopicMsgTooLarge
}
```

### Step 3: Enforce in Deliver()

Same check, same position:

```go
if limit, ok := TopicMessageSizeLimit[topic]; ok && len(data) > limit {
    return ErrTopicMsgTooLarge
}
```

### Step 4: Update MergeTick to use actual serialized size

In `pkg/txpool/stark_aggregation.go`, replace the approximate formula `len(hashes)*32 + 1024` with `remote.MarshalBinary()` and check `len(serialized) > MaxTickSize`. This ensures the application-layer check matches the gossip-layer check.

## Tests

- `TestTopicMessageSizeLimit` — oversized STARKMempoolTick data rejected in Publish and Deliver
- `TestTopicMessageSizeLimit_NonLimitedTopic` — BeaconBlock allows 200KB (under 10MB global limit)
- `TestMergeTick_ActualSerializedSize` — serialized tick for 10 txs is under 128KB
- `TestMergeTick_BandwidthLimit` — 4100-tx tick exceeds 128KB via actual serialization

## Codebase Locations

| File | Line | Purpose |
|------|------|---------|
| `pkg/p2p/gossip_topics.go` | 180 | `TopicMessageSizeLimit` map |
| `pkg/p2p/gossip_topics.go` | 186 | `ErrTopicMsgTooLarge` error |
| `pkg/p2p/gossip_topics.go` | 294 | `Publish()` — per-topic check |
| `pkg/p2p/gossip_topics.go` | 340 | `Deliver()` — per-topic check |
| `pkg/txpool/stark_aggregation.go` | 388 | `MergeTick()` — actual serialized size |
