# gas — EIP-1559 gas price oracle and fee tracker

[← rpc](../README.md)

## Overview

Package `gas` provides an EIP-1559-aware gas price oracle and a rolling block
fee tracker for use by `eth_gasPrice`, `eth_maxPriorityFeePerGas`, and
`eth_feeHistory`. The oracle samples priority fees from recent blocks, applies
a configurable percentile, and caps results at a maximum price to prevent
unreasonably high suggestions. A companion `GasTracker` accumulates per-block
statistics for diagnostics.

## Functionality

**Configuration**

- `GasOracleConfig` — `Blocks int`, `Percentile int`, `MaxPrice *big.Int`, `IgnorePrice *big.Int`, `MaxHeaderHistory int`
- `DefaultGasOracleConfig()` — `Blocks=20`, `Percentile=60`, `MaxPrice=500 Gwei`, `IgnorePrice=2 wei`, `MaxHeaderHistory=1024`

**`GasOracle`** — constructed with `NewGasOracle(config GasOracleConfig)`

| Method | Description |
|---|---|
| `RecordBlock(number uint64, baseFee *big.Int, tips []*big.Int)` | Feed a new block's fee data into the history |
| `BaseFee() *big.Int` | Latest known base fee |
| `SuggestGasTipCap() *big.Int` | Percentile priority fee across the last N blocks |
| `SuggestGasPrice() *big.Int` | Legacy gas price = baseFee + tip |
| `MaxPriorityFeePerGas() *big.Int` | Configured maximum priority fee |
| `FeeHistory(blockCount int) []BlockFeeData` | Per-block fee history for `eth_feeHistory` |
| `EstimateL1DataFee(dataSize uint64) *big.Int` | Rollup L1 data posting cost estimate |

**`BlockFeeData`** — `Number uint64`, `BaseFee *big.Int`, `RewardPercentile *big.Int`, `GasUsedRatio float64`

**`GasTracker`** — rolling window tracker for per-block gas statistics; updated via `RecordBlock`; query methods `AverageGasUsed()`, `AverageBaseFee()`, `BlockCount() int`
