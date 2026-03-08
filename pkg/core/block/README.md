# block

Block construction, execution, and validation for the ETH2030 execution layer.

[← core](../README.md)

## Overview

Package `block` provides the three building blocks of the EL block lifecycle: a
`BlockBuilder` that assembles new blocks from a transaction pool, a
`BlockExecutor` that processes a pre-built block to produce execution results,
and a `BlockValidator` that enforces consensus rules on headers, bodies, and
post-execution state commitments.

The builder integrates tightly with the Glamsterdam/Amsterdam fork stack: it
applies FOCIL inclusion-list transactions first (EIP-7547/7805), tracks
per-transaction state changes in a `BlockAccessList` (EIP-7928), enforces the
per-block `DimStorage` gas cap (EIP-7706), processes EIP-4895 withdrawals,
accumulates EIP-7685 EL requests, and computes the BAL hash for the header.

## Functionality

### Interfaces

- `Validator` — `ValidateHeader`, `ValidateBody`, `ValidateRequests`,
  `ValidateBlockAccessList`.
- `BlockchainReader` — read-only chain access needed by the builder
  (`Config`, `CurrentBlock`, `Genesis`, `GetBlock`, `StateAtBlock`).
- `TxPoolReader` — `Pending() []*types.Transaction`.

### Types

- `BlockBuilder` — builds blocks via `BuildBlock(parent, attrs)` or
  `BuildBlockLegacy(parent, txs, ...)` for backward compatibility.
  Constructed with `NewBlockBuilder(config, chain, pool)`.
- `BuildBlockAttributes` — payload attributes: `Timestamp`, `FeeRecipient`,
  `Random`, `Withdrawals`, `BeaconRoot`, `GasLimit`, `InclusionListTxs`.
- `BlockExecutor` — executes blocks via `Execute(header, txs)` and validates
  results via `ValidateExecution(result, header)`. Constructed with
  `NewBlockExecutor(config ExecutorConfig)`.
- `ExecutorConfig` — `ParallelTxs`, `MaxGasPerBlock`, `TraceExecution`.
- `BlockExecutionResult` — `StateRoot`, `ReceiptsRoot`, `LogsBloom`,
  `GasUsed`, `TxCount`, `Success`.
- `BlockValidator` — validates headers, bodies, requests, and BAL hashes.
  Constructed with `NewBlockValidator(config)`.

### Key Constants

| Constant | Value | Source |
|---|---|---|
| `MinGasLimit` | 5000 | EIP-1559 |
| `MaxGasLimit` | 2^63 - 1 | |
| `GasLimitBoundDivisor` | 1024 | EIP-1559 |
| `MaxExtraDataSize` | 32 | |
| `ElasticityMultiplier` | 2 | EIP-1559 |
| `BaseFeeChangeDenominator` | 8 | EIP-1559 |

### Key Functions

- `CalcGasLimit(parentGasLimit, parentGasUsed)` — EIP-1559 gas limit adjustment.
- `EffectiveGasPrice(tx, baseFee)` — `min(gasFeeCap, baseFee + gasTipCap)`.
- `ValidateBlobHashes(hashes)` — ensures every hash starts with `0x01`.
- `SortedTxLists(pending, baseFee)` — separates and sorts regular/blob txs.
- `ValidateCalldataGas(header, parent)` — EIP-7706 calldata gas header check.
- `ComputeReceiptsRoot(receipts)` — MPT receipt root.
- `DeriveTxsRoot`, `DeriveReceiptsRoot`, `DeriveWithdrawalsRoot` — Merkle
  Patricia Trie root derivations.

### Errors

`ErrUnknownParent`, `ErrFutureBlock`, `ErrInvalidNumber`,
`ErrInvalidGasLimit`, `ErrInvalidGasUsed`, `ErrInvalidTimestamp`,
`ErrInvalidBaseFee`, `ErrInvalidRequestHash`, `ErrInvalidBlockAccessList`,
`ErrMissingBlockAccessList`, `ErrBlobGasLimitExceeded`, `ErrInvalidBlobHash`.

## Usage

```go
builder := block.NewBlockBuilder(chainConfig, chainReader, txPool)
blk, receipts, err := builder.BuildBlock(parentHeader, &block.BuildBlockAttributes{
    Timestamp:    uint64(time.Now().Unix()),
    FeeRecipient: coinbase,
    GasLimit:     30_000_000,
    Withdrawals:  withdrawals,
    BeaconRoot:   &beaconRoot,
})

validator := block.NewBlockValidator(chainConfig)
if err := validator.ValidateHeader(blk.Header(), parentHeader); err != nil {
    // handle validation failure
}
```
