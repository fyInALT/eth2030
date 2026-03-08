# ratemeter

EWMA rate meter for gas throughput, bandwidth, and sync rate tracking.

[← core](../README.md)

## Overview

Package `ratemeter` provides `RateMeter`, a sliding-window gas rate meter with
exponential moving average (EMA) smoothing and adaptive parallelism
recommendations. It is used by the gigagas infrastructure to track whether the
node is approaching the 1 Ggas/sec target and to dynamically scale the number
of parallel execution workers.

## Functionality

### RateMeterConfig

```go
type RateMeterConfig struct {
    WindowSize      int     // blocks in sliding window (default 64)
    TargetGasPerSec float64 // target rate in gas/sec (default 1_000_000_000)
    EMAAlpha        float64 // EMA smoothing factor 0 < α ≤ 1 (default 0.1)
    MinWorkers      int     // minimum parallel workers (default 2)
    MaxWorkers      int     // maximum parallel workers (default 128)
}
```

`DefaultRateMeterConfig()` — returns the above defaults.

### RateMeter

- `NewRateMeter(config RateMeterConfig) *RateMeter`
- `RecordBlock(blockNumber, gasUsed, timestamp uint64)` — records a block's
  gas usage (timestamp in seconds) and updates the EMA.
- `CurrentRate() float64` — EMA-smoothed gas rate in gas/sec.
- `RollingAverageRate() float64` — simple rolling average over the full window.
- `RecommendedWorkers() int` — adaptive worker count based on rate vs target.
- `UtilizationRatio() float64` — `currentRate / targetRate`; 1.0 = at target.
- `IsAtTarget() bool` — true if within ±20% of the target rate.
- `Reset()` — clears all records and resets EMA and worker count.
- `WindowSize() int`, `RecordCount() int`, `TargetGasPerSec() float64` —
  configuration and state accessors.

### Adaptive Worker Scaling

After each `RecordBlock` call, `adaptWorkers` adjusts the recommended worker
count based on the ratio of current EMA rate to target:

| Ratio | Action |
|---|---|
| < 0.5 | Double workers |
| 0.5 – 0.8 | Increase by 25% |
| 0.8 – 1.2 | No change |
| 1.2 – 1.5 | Decrease by 25% |
| > 1.5 | Halve workers |

Workers are always clamped to `[MinWorkers, MaxWorkers]`.

## Usage

```go
rm := ratemeter.NewRateMeter(ratemeter.DefaultRateMeterConfig())

// After each block is processed:
rm.RecordBlock(block.NumberU64(), block.GasUsed(), block.Time())

rate := rm.CurrentRate()          // EMA gas/sec
workers := rm.RecommendedWorkers() // adaptive parallelism hint
ratio := rm.UtilizationRatio()    // 0.0 = idle, 1.0 = at target

if rm.IsAtTarget() {
    log.Println("gigagas target reached")
}
```
