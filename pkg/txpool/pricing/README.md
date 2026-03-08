# txpool/pricing — Gas price heaps and priority ordering

## Overview

Package `pricing` provides the data structures and algorithms for ordering transactions by effective gas price. `PriceHeap` maintains parallel min-heap (for eviction) and max-heap (for block building) over pending transactions, plus a min-heap for queued (future) transactions. It supports lazy deletion, per-sender nonce-gap detection, and full re-heaping when the base fee changes. Helper functions compute EIP-1559-aware effective prices and tips for use throughout the txpool.

## Functionality

**Types**
- `PriceHeap` — concurrent dual-heap; `AddPending`, `AddQueued`, `Remove`, `PopCheapest`, `PeekHighestTip`, `PopCheapestQueued`, `DetectNonceGaps`, `SetBaseFee`, `Cleanup`
- `PriorityQueue` — generic max-heap for block building by tip

**Functions**
- `EffectiveGasPrice(tx, baseFee)` — `min(feeCap, baseFee+tipCap)` for EIP-1559; `gasPrice` for legacy
- `NewPriceHeap(baseFee)` — create heap with initial base fee

**Key behaviours**
- Lazy deletion: `Remove` marks entries deleted; `PopCheapest` skips them; `Cleanup` rebuilds
- `SetBaseFee` triggers full re-heap recomputing all effective prices
- `DetectNonceGaps(sender, baseNonce)` returns missing nonces for a sender

## Usage

```go
ph := pricing.NewPriceHeap(currentBaseFee)
ph.AddPending(tx, senderAddr)

for victim := ph.PopCheapest(); victim != nil; victim = ph.PopCheapest() {
    // evict lowest-priced tx to make room
}
ph.SetBaseFee(newBaseFee) // re-orders entire heap
```

[← txpool](../README.md)
