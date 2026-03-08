# txpool/sharding — Consistent-hash sharded transaction pool

## Overview

Package `sharding` implements a sharded mempool that distributes transactions across multiple independent shards using consistent hashing on the transaction hash (first 4 bytes mod numShards). Each shard has its own mutex, enabling parallel access to disjoint transaction subsets and reducing lock contention at high throughput. The pool supports dynamic rebalancing of hotspot shards, online resizing of the shard count with full transaction migration, and per-shard capacity limits.

## Functionality

**Types**
- `ShardedPool` — `AddTx`, `RemoveTx`, `GetTx`, `PendingByAddress`, `Count`, `GetShardStats`, `RebalanceShards`, `ResizeShards`
- `ShardConfig{NumShards, ShardCapacity, ReplicationFactor}` — must have NumShards as power of two
- `ShardStats{ID, TxCount, Utilization}` — per-shard metrics

**Functions**
- `NewShardedPool(config)` — create pool with given number of shards
- `DefaultShardConfig()` — 16 shards, 1024 capacity each, replication 1
- `ValidateShardAssignment(config)` — validates NumShards is power-of-two, capacity > 0, replication in bounds

**Key behaviours**
- `ShardForTx(hash)` / `ShardForAddress(addr)` — deterministic shard routing
- `RebalanceShards()` — moves overflow from shards >150% average to underloaded ones
- `ResizeShards(n)` — drains all shards, creates new set, re-inserts via new hash ring

## Usage

```go
pool := sharding.NewShardedPool(sharding.DefaultShardConfig())
pool.AddTx(tx)
tx := pool.GetTx(hash)
stats := pool.GetShardStats() // per-shard utilization
pool.ResizeShards(32)         // grow at epoch boundary
```

[← txpool](../README.md)
