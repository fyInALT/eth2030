# txpool/fees — Gas price oracle and blob fee tracking

## Overview

Package `fees` provides three cooperative fee-estimation components. `FeeEstimator` maintains a sliding window of recent block gas prices and computes percentile-based suggestions (slow/medium/fast) for legacy and EIP-1559 transactions. `PriceOracle` adds per-block tip percentiles, EIP-1559 next-base-fee projection, and a structured `FeeRecommendation` with all urgency levels. `BlobFeeTracker` is a dedicated blob gas market oracle that detects fee spikes, tracks utilization ratios, and produces `BlobFeeSuggestion` with separate slow/medium/fast blob fee caps.

All three types operate on circular buffers of block records and are safe for concurrent use.

## Functionality

**Types**
- `FeeEstimator` — `AddBlock`, `SuggestGasPrice`, `SuggestGasTipCap`, `SuggestGasFeeCap`, `EstimateBlobFee`, `FeeEstByPercentile`
- `PriceOracle` — `AddBlock`, `SuggestTipCap`, `SuggestGasPrice`, `EstimateNextBaseFee`, `Recommend`, `FeeHistory`
- `BlobFeeTracker` — `AddBlock`, `SuggestBlobFee`, `Suggest`, `EstimateNextBlobFee`, `MovingAverage`, `IsCurrentSpike`, `BlobGasUtilization`
- `BlockFeeData`, `BlockFeeRecord`, `BlobFeeRecord` — per-block input records
- `FeeRecommendation`, `BlobFeeSuggestion`, `BlobFeeSpike` — output structs
- `FeeEstimatorConfig`, `PriceOracleConfig`, `BlobFeeTrackerConfig` — configuration

**Functions**
- `DefaultFeeEstimatorConfig`, `DefaultPriceOracleConfig`, `DefaultBlobFeeTrackerConfig`

## Usage

```go
oracle := fees.NewPriceOracle(fees.DefaultPriceOracleConfig())
oracle.AddBlock(fees.BlockFeeRecord{Number: 100, BaseFee: baseFee, Tips: tips})
rec := oracle.Recommend() // rec.MediumFee, rec.MediumTip, rec.NextBaseFee
```

[← txpool](../README.md)
