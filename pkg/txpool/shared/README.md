# txpool/shared — Cross-node and cross-validator transaction propagation

## Overview

Package `shared` provides two complementary pools for mempool propagation beyond a single node. `SharedMempool` handles cross-node gossip: it maintains a peer-aware transaction cache with a built-in bloom filter for fast deduplication, tracks which peers a transaction has been relayed to, and sorts pending transactions by gas price for priority relay. `SharedPool` handles cross-validator shard coordination: transactions carry shard origin metadata and relay hop counts, and the pool enforces a relay limit to prevent infinite propagation; `CrossShardSync` gathers transactions from peer shards not yet present locally.

## Functionality

**Types**
- `SharedMempool` — cross-node cache; `AddTransaction`, `GetPendingTxs`, `MarkRelayed`, `IsRelayedTo`, `IsKnown`, `AddPeer`, `RemovePeer`, `EvictStale`, `TxCount`
- `SharedMempoolTx{Hash, Sender, GasPrice, ReceivedFrom, ReceivedAt}` — gossip record
- `SharedPool` — cross-validator pool; `AddSharedTx`, `GetPendingForShard`, `RelayTx`, `PruneShard`, `CrossShardSync`, `TxCount`, `ShardCount`, `Close`
- `SharedTx{Tx, Origin ShardID, RelayCount, Priority, AddedAt}` — shard-aware tx wrapper
- `SharedMempoolConfig`, `SharedPoolConfig` — configuration

**Functions**
- `NewSharedMempool(config)`, `NewSharedPool(config)` — constructors
- `DefaultSharedMempoolConfig` — 50 peers, 4096 cache, 65536-bit bloom, 500ms relay interval
- `DefaultSharedPoolConfig` — 1024 tx/shard, 64 shards, relay limit 3

## Usage

```go
mem := shared.NewSharedMempool(shared.DefaultSharedMempoolConfig())
mem.AddPeer("peer-1")
mem.AddTransaction(shared.SharedMempoolTx{Hash: h, GasPrice: 1e9})
if !mem.IsRelayedTo(h, "peer-2") {
    mem.MarkRelayed(h, "peer-2")
    // forward to peer-2
}
```

[← txpool](../README.md)
