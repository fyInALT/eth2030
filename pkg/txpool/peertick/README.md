# txpool/peertick — Peer-validated transaction tick cache

## Overview

Package `peertick` provides a slot-scoped cache that records which transaction hashes have been validated by remote peers via STARK mempool ticks. Entries expire automatically after a configurable number of slots, preventing stale peer validations from persisting across epoch boundaries. The cache is used by the STARK aggregator (`txpool/stark`) to skip redundant local validation for transactions already proven by peers.

## Functionality

**Types**
- `PeerTickCache` — concurrent cache; `MarkPeerValidated`, `IsPeerValidated`, `AdvanceSlot`, `Size`

**Functions**
- `NewPeerTickCache(slotTTL uint64)` — create cache with given slot time-to-live (default 2)

## Usage

```go
cache := peertick.NewPeerTickCache(2) // entries live for 2 slots
cache.MarkPeerValidated(txHash, "peer-abc", currentSlot)

if cache.IsPeerValidated(txHash) {
    // skip local validation, trust peer STARK proof
}

evicted := cache.AdvanceSlot(nextSlot) // removes entries older than slotTTL
```

[← txpool](../README.md)
