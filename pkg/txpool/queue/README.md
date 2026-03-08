# txpool/queue — Future transaction queue manager

## Overview

Package `queue` manages transactions whose nonces are too high to be immediately executable (they have gaps relative to the sender's current state nonce). `QueueManager` maintains per-account nonce-sorted queues with configurable per-account and global capacity limits. When capacity is exceeded it evicts the lowest-priced transaction. It supports promotion of gap-free nonce prefixes to the pending pool and automatic removal of stale nonces after chain advances.

## Functionality

**Types**
- `QueueManager` — concurrent manager; `Add`, `Remove`, `PromoteReady`, `UpdateStateNonce`, `Get`, `AccountNonces`, `EvictAll`, `SetBaseFee`, `Len`, `AccountCount`
- `QueueManagerConfig{MaxPerAccount, MaxGlobal}` — capacity configuration

**Functions**
- `NewQueueManager(config, baseFee)` — create manager with given base fee for price comparisons

**Errors**
- `ErrReplacementUnderpriced`, `ErrSenderLimitExceeded`, `ErrTxPoolFull`

**Key behaviours**
- `PromoteReady(sender, baseNonce)` — returns and removes the contiguous nonce prefix starting at `baseNonce`
- `UpdateStateNonce(sender, nonce)` — removes all queued txs with nonce < new state nonce
- Eviction selects the lowest effective gas price first (per-account), then globally

## Usage

```go
qm := queue.NewQueueManager(queue.QueueManagerConfig{MaxPerAccount: 64, MaxGlobal: 1024}, baseFee)
evicted, err := qm.Add(sender, tx)
ready := qm.PromoteReady(sender, stateNonce)
qm.UpdateStateNonce(sender, newNonce) // clean up mined txs
```

[← txpool](../README.md)
