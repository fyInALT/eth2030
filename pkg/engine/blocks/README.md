# engine/blocks — Block assembly and pipeline

Assembles execution blocks from pending transactions and provides a multi-stage
block-building pipeline for the Engine API payload construction workflow.

## Overview

`TxBlockBuilder` (`block_builder.go`) maintains a list of pending transactions,
enforces a gas limit, and produces a `types.Block` with transactions sorted by
effective gas price (EIP-1559 `GasFeeCap` descending). It respects the gas
limit by skipping transactions that would overflow.

`block_assembler.go` provides a higher-level `BlockAssembler` that coordinates
header construction with state-root computation, receipts, and withdrawals.

`block_pipeline.go` implements a `BlockPipeline` that runs multiple `BlockStage`
processors in sequence, enabling modular payload enrichment (e.g. BAL tracking,
FOCIL inclusion checks).

## Functionality

**Types**
- `TxBlockBuilder` — gas-aware transaction accumulator
- `BlockAssembler` — header + body assembler with state-root hooks
- `BlockPipeline` — composable multi-stage block builder

**Functions**
- `NewTxBlockBuilder() *TxBlockBuilder`
- `(*TxBlockBuilder).AddTransaction(tx) error`
- `(*TxBlockBuilder).SetGasLimit(limit uint64)`
- `(*TxBlockBuilder).BuildBlock(parentHash, timestamp, coinbase, gasLimit) (*types.Block, error)`
- `(*TxBlockBuilder).GasUsed() uint64`
- `(*TxBlockBuilder).PendingCount() int`
- `(*TxBlockBuilder).Reset()`

## Usage

```go
bb := blocks.NewTxBlockBuilder()
bb.SetGasLimit(30_000_000)
for _, tx := range pendingTxs {
    bb.AddTransaction(tx)
}
block, err := bb.BuildBlock(parentHash, timestamp, coinbase, gasLimit)
```

[← engine](../README.md)
