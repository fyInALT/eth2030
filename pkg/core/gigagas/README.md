# gigagas

Gigagas (1 Ggas/sec) infrastructure: gas rate tracking, work-stealing scheduler, and parallel executor (M+ roadmap).

[← core](../README.md)

## Overview

Package `gigagas` implements the execution infrastructure needed to reach the
M+ North Star goal of 1 billion gas per second (gigagas). It combines a
sliding-window `GasRateTracker` for real-time throughput measurement, a
`GigagasScheduler` with conflict-aware batching and work-stealing parallelism,
and a `GigagasExecutor` that drives the full pipeline.

## Functionality

### Configuration (`gigagas.go`)

```go
type GigagasConfig struct {
    TargetGasPerSecond     uint64 // target: 1_000_000_000
    MaxBlockGas            uint64 // 500_000_000
    ParallelExecutionSlots uint32 // 16
}

var DefaultGigagasConfig = GigagasConfig{...}
```

### GasRateTracker (`gigagas.go`)

Sliding-window gas throughput measurement.

- `NewGasRateTracker(windowSize int) *GasRateTracker`
- `RecordBlockGas(blockNum, gasUsed, timestamp uint64)`
- `CurrentGasRate() float64` — gas/sec over the window; returns 0 if fewer
  than 2 records.

### GigagasScheduler (`gigagas_scheduler.go`)

Conflict-aware batch scheduler that groups independent transactions into
parallel execution lanes.

- `GigagasSchedulerConfig` — `MaxLanes` (16), `BatchSize` (256),
  `ConflictRetryLimit` (3).
- `DefaultGigagasSchedulerConfig() GigagasSchedulerConfig`
- `WorkUnit` — `{Index, ReadSet, WriteSet, GasEstimate, RetryCount}`.
- `NewWorkUnit(index int, gasEstimate uint64) *WorkUnit`
- `AddRead(key)`, `AddWrite(key)` — annotate the read/write set.
- `GigagasScheduler.Schedule(units []*WorkUnit)` — returns batches of
  non-conflicting `WorkUnit` slices for parallel execution.
- `GigagasScheduler.Conflicting(a, b *WorkUnit) bool` — returns true if two
  units have overlapping write sets or read-write conflicts.

### Work-Stealing Pool (`work_stealing.go`)

Lock-free goroutine pool where idle workers steal tasks from busy workers'
local deques.

- `WorkStealingTask` — `{ID int, GasCost uint64, Execute func() uint64}`.
- `WorkStealingPool` — pool of worker goroutines with per-worker deques.
- `NewWorkStealingPool(numWorkers int) *WorkStealingPool`
- `Submit(task *WorkStealingTask)` — adds a task to the pool.
- `Wait()` — blocks until all submitted tasks complete.
- `TotalGasExecuted() uint64` — sum of gas returned by all `Execute` calls.

### GigagasExecutor (`gigagas_executor.go`)

Wraps the scheduler and pool into a single execution entry point.

- `NewGigagasExecutor(config GigagasConfig) *GigagasExecutor`
- `Execute(units []*WorkUnit) (totalGas uint64, err error)` — schedules and
  executes all units through the work-stealing pool.

### Integration (`gigagas_integration.go`)

Wires the gigagas infrastructure into the broader block execution pipeline,
connecting `GasRateTracker`, `GigagasScheduler`, and `GigagasExecutor`.

## Usage

```go
tracker := gigagas.NewGasRateTracker(100)
tracker.RecordBlockGas(blockNum, block.GasUsed(), block.Time())
rate := tracker.CurrentGasRate() // gas/sec

executor := gigagas.NewGigagasExecutor(gigagas.DefaultGigagasConfig)
units := buildWorkUnits(block.Transactions())
totalGas, err := executor.Execute(units)
```
