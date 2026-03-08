# gas

Complete gas pricing stack: EIP-1559 base fee, blob gas, multidimensional pricing, repricing bundles, gas cap, gas limit schedule, estimator, and gas futures market.

[← core](../README.md)

## Overview

Package `gas` implements all gas-related computation for the ETH2030 execution
layer. It covers the historical EIP-1559 base fee, the EIP-4844 blob gas
mechanism with BPO blob schedules, the EIP-7706 multidimensional (5D) gas
engine, Glamsterdam and Hogotá repricing bundles, the EIP-7623/7976 calldata
floor, the EIP-7825 gas cap, a binary-search gas estimator, and the long-dated
gas futures market (M+ roadmap).

## Functionality

### EIP-1559 Base Fee (`fee.go`)

- `CalcBaseFee(parent *types.Header) *big.Int` — standard EIP-1559 base fee
  computation; returns at least `MinBaseFee` (7 wei).
- Constants: `InitialBaseFee` (1 Gwei), `MinBaseFee` (7), `ElasticityMultiplier` (2),
  `BaseFeeChangeDenominator` (8).

### Blob Gas (`blob_gas.go`, `blob_schedule.go`, `blob_validation.go`)

- `BlobSchedule` — `{Target, Max, UpdateFraction}` per fork.
- `CalcExcessBlobGas(parentExcess, parentUsed)` — EIP-4844 excess blob gas.
- `CalcBlobBaseFee(excessBlobGas)` — exponential blob base fee.
- `ValidateBlockBlobGas(header, parent)` — validates blob gas fields.
- `CountBlobGas(tx)` — gas consumed by a single blob transaction.
- BPO1/BPO2 blob schedule constants and `GetBlobSchedule(config, time)`.
- Constants: `MaxBlobGasPerBlock` (786432), `TargetBlobGasPerBlock` (393216),
  `GasPerBlob` (131072), `MaxBlobsPerBlock` (6), `BlobTxHashVersion` (0x01).
- Fusaka (EIP-7691): `FusakaMaxBlobsPerBlock` (9), `FusakaTargetBlobsPerBlock` (6),
  `MinBaseFeePerBlobGas`, `BlobBaseCost` (EIP-7918 reserve price).

### Multidimensional Gas (`multidim_gas.go`, `multidim.go`, `multidim_market.go`)

Five independent gas dimensions, each with its own EIP-1559-style base fee:

| Dimension | Constant | Tracks |
|---|---|---|
| `DimCompute` | 0 | Traditional EVM gas |
| `DimStorage` | 1 | SLOAD/SSTORE state access |
| `DimBandwidth` | 2 | Calldata (data availability) |
| `DimBlob` | 3 | EIP-4844 blobs |
| `DimWitness` | 4 | Stateless execution witness |

- `MultidimGasEngine` — thread-safe 5D pricing engine; `NewMultidimGasEngine()`,
  `CalcNextBaseFees(block)`, `GetBaseFee(dim)`.
- `NumGasDimensions` = 5.
- `AllGasDimensions()` — returns all five `GasDimension` values.

### Calldata Gas (`calldata_gas.go`, `eip7623_floor.go`)

- `CalldataGas(data, isCreate)` — standard calldata gas cost.
- `GetCalldataGas(data)` — EIP-7706 alias.
- `CalldataFloorGas(data, isCreate)` — EIP-7623/7976 calldata floor cost.
- `CalcCalldataExcessGas(parentExcess, parentUsed, gasLimit)` — EIP-7706 excess.
- `CalcCalldataGasLimit(blockGasLimit)` — derives calldata gas limit.

### Gas Cap (`gas_cap.go`, `gas_cap_extended.go`)

- `ValidateTxGasCap(tx, blockGasLimit)` — EIP-7825 per-transaction cap.
- `ValidateBlockGasCap(block)` — per-block gas cap enforcement.

### Gas Limit Schedule (`gas_limit.go`)

- `CalcGasLimitSchedule(config, parent)` — scheduled gas limit increases per
  fork (Hogotá 3x/year schedule).

### Repricing (`glamsterdam_repricing.go`, `hogota_repricing.go`, `conversion_repricing.go`)

- `ApplyGlamsterdamRepricing(opcode)` — applies the Glamsterdam repricing bundle
  (18 EIPs) to EVM opcode gas costs.
- `ApplyHogotaRepricing(opcode)` — applies the Hogotá repricing adjustments.

### Gas Estimator (`gas_estimator.go`)

- `EstimateGas(tx, statedb, header, config)` — binary-search gas estimation.

### Gas Futures Market (`gas_futures.go`, `gas_market.go`, `gas_settlement.go`)

- `GasFuturesContract` — on-chain long-dated gas futures (M+ roadmap).
- `GasMarket` — order book and matching engine.
- `GasSettlement` — delivery and settlement logic.

### Gas Pool Extension (`gas_pool_extended.go`)

- `MultidimGasPool` — extends `gaspool.GasPool` with per-dimension tracking.

### Constants (`constants.go`)

`TxGas` (21000), `TxDataZeroGas` (4), `TxDataNonZeroGas` (16),
`TxCreateGas` (32000), `TotalCostFloorPerToken` (10),
`GasLimitBoundDivisor` (1024), `MinGasLimit` (5000),
`ErrGasLimitExceeded`, `ErrIntrinsicGasTooLow`.

## Usage

```go
// EIP-1559 base fee for the next block.
nextBaseFee := gas.CalcBaseFee(parentHeader)

// EIP-4844 blob gas.
excessBlobGas := gas.CalcExcessBlobGas(parentExcess, parentUsed)
blobBaseFee := gas.CalcBlobBaseFee(excessBlobGas)

// EIP-7623 calldata floor.
floor := gas.CalldataFloorGas(tx.Data(), tx.To() == nil)
```
