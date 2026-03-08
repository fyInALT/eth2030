# das/sampleopt — Adaptive DAS sample size optimizer

Dynamically optimises the number of DAS samples per blob based on security
requirements and observed network conditions. Implements the "decreased sample
size" optimization from the Data Layer roadmap.

## Overview

`SampleOptimizer` computes the minimum sample count needed to reach a target
security level (in bits) using the information-theoretic model from the PeerDAS
spec: with `k` samples drawn from `N = 128` columns, the probability that at
least one withheld cell is detected is `1 − ((N−1)/N)^k`. The optimizer clamps
results to `[MinSamples, MaxSamples]` and exposes an adaptive variant that
scales the sample count up when network health degrades.

A `SamplingPlan` describes a concrete per-slot strategy (samples per blob, total
samples, security level in bits, confidence). A `SamplingVerdict` reports
whether a completed sampling round met the plan's threshold.

## Functionality

**Types**
- `SampleOptimizerConfig` — `MinSamples`, `MaxSamples`, `TargetConfidence`, `SecurityMargin`
- `SampleOptimizer` — thread-safe optimizer
- `SamplingPlan` — `SamplesPerBlob`, `TotalSamples`, `SecurityLevel`, `ConfidenceLevel`
- `SamplingVerdict` — `Sufficient`, `Confidence`, `MissingSamples`

**Functions**
- `DefaultSampleOptimizerConfig() SampleOptimizerConfig`
- `NewSampleOptimizer(config) *SampleOptimizer`
- `(*SampleOptimizer).CalculateOptimalSamples(blobCount, securityParam int) int`
- `(*SampleOptimizer).AdaptiveSampling(blobCount int, networkHealth float64) *SamplingPlan`
- `(*SampleOptimizer).ValidateSamplingResult(plan, receivedSamples) *SamplingVerdict`
- `(*SampleOptimizer).AdjustSampleSize(currentSize int, failureRate float64) int`
- `(*SampleOptimizer).EstimateNetworkLoad(blobCount, sampleSize int) uint64`
- `(*SampleOptimizer).ValidateSamplingPlan(plan) error`

## Usage

```go
opt := sampleopt.NewSampleOptimizer(sampleopt.DefaultSampleOptimizerConfig())

plan := opt.AdaptiveSampling(blobCount, networkHealth) // networkHealth in [0,1]
// ... collect samples ...
verdict := opt.ValidateSamplingResult(plan, receivedCount)
if !verdict.Sufficient {
    // request verdict.MissingSamples more samples
}
```

[← das](../README.md)
