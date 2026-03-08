# sync/inserter — Block insertion pipeline

Sequentially executes and validates blocks, verifying state roots, receipt roots, logs blooms, and gas used before committing each block to the chain.

[← sync](../README.md)

## Overview

`ChainInserter` bridges the sync layer and the chain backend. For each block it calls a `BlockExecutor` to obtain the computed state root, receipts, and gas used, then cross-checks these against the header's declared values. Validation checks are individually configurable (`VerifyStateRoot`, `VerifyReceipts`, `VerifyBloom`, `VerifyGasUsed`). After validation passes, the block is inserted via `BlockInserter.InsertChain` and optionally committed through a `BlockCommitter`.

`InsertBatch` groups blocks into batches for more efficient chain insertion, validating parent-hash linkage within the batch before forwarding to the underlying inserter.

Prometheus-compatible metrics (`CIMetrics`) track blocks inserted, failures, transactions processed, gas processed, and per-operation latency histograms.

## Functionality

### Types

| Type | Purpose |
|------|---------|
| `ChainInserter` | Validates and inserts blocks with pluggable executor/committer |
| `BlockExecutor` | Interface: `ExecuteBlock(block)` → stateRoot, receipts, error |
| `BlockCommitter` | Interface: `CommitBlock(block)` |
| `BlockInserter` | Interface: `InsertChain(blocks)` + `CurrentBlock()` |
| `CIMetrics` | Counters and histograms for insertion performance |
| `CIProgress` | Blocks/txs/gas inserted, throughput (blocks/s, tx/s) |

### Key Functions

- `NewChainInserter(config, inserter)` / `SetExecutor(exec)` / `SetCommitter(c)`
- `InsertBlocks(blocks)` — single-block loop with full validation
- `InsertBatch(blocks)` — batched insertion optimised for sync throughput
- `Progress()` / `Metrics()` / `Close()` / `Reset()`

## Usage

```go
ci := inserter.NewChainInserter(inserter.DefaultChainInserterConfig(), chainBackend)
ci.SetExecutor(myExecutor)
n, err := ci.InsertBatch(downloadedBlocks)
fmt.Printf("inserted %d blocks at %.1f blocks/s\n", n, ci.Progress().BlocksPerSecond())
```
