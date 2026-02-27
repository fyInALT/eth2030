# Sprint 5, Story 5.3 — Bandwidth Limit Enforcement

**Sprint goal:** Enforce the 128KB tick size limit from ethresear.ch.
**Files modified:** `pkg/txpool/stark_aggregation.go`

## Overview

The ethresear.ch proposal specifies a bandwidth model of 128KB per tick. Oversized ticks could consume excessive bandwidth, defeating the proposal's efficiency goals.

## Gap (GAP-STARK5)

**Severity:** LOW (structural, not security-critical)
**File:** `pkg/txpool/stark_aggregation.go`
**Evidence:** No size limit enforcement existed on `MempoolAggregationTick` serialization.

## Implement

```go
// pkg/txpool/stark_aggregation.go
const MaxTickSize = 128 * 1024 // 128KB per ethresear.ch bandwidth model

var ErrAggTickTooLarge = errors.New("stark_aggregation: tick exceeds 128KB bandwidth limit")
```

The bandwidth limit is enforced at two points:

1. **GenerateTick**: After serialization, check `len(data) <= MaxTickSize`
2. **Gossip handler**: Before deserialization, check `len(msg.Data) <= MaxTickSize`

## ethresear.ch Spec Reference

> The bandwidth budget for mempool aggregation is 128KB × peers per tick interval (500ms). Each individual tick message must not exceed 128KB.

## Codebase Locations

| File | Line | Purpose |
|------|------|---------|
| `pkg/txpool/stark_aggregation.go` | 40 | MaxTickSize constant |
| `pkg/txpool/stark_aggregation.go` | 42 | ErrAggTickTooLarge error |
