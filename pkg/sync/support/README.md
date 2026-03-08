# sync/support — Shared sync infrastructure

Provides reusable utilities for the sync subsystem: a stage-based pipeline orchestrator, a cross-stage progress tracker, and a resource-aware concurrent heal scheduler.

[← sync](../README.md)

## Overview

**`SyncPipeline`** (`pipeline.go`) models sync as a DAG of named stages (`HeaderSync`, `BodySync`, `ReceiptSync`, `StateSync`, `Verification`). Each stage transitions through `pending → running → completed/failed` with explicit dependency checking. Stages can be retried up to a configured limit. `OverallProgress()` aggregates per-stage percentages.

**`ProgressTracker`** (`progress.go`) is a thread-safe recorder for the current sync stage and block counters. It computes `PercentComplete` and `EstimatedCompletion` from blocks-per-second throughput and exposes `BlocksPerSecond()`.

**`ConcurrentHealer`** (`concurrent_heal_scheduler.go`) is a priority-queued, resource-budgeted scheduler for parallel trie healing. `SchedulerHealTask` entries carry a priority (`Critical > Urgent > Normal > Background`), an estimated memory cost, and an optional deadline. The `ResourceBudget` enforces memory and bandwidth limits. `ProcessBatch` returns the highest-priority non-expired tasks up to `MaxWorkers`. Tasks can be re-queued on failure via `FailTaskWithRetry`.

## Functionality

### Types

| Type | Purpose |
|------|---------|
| `SyncPipeline` | Dependency-aware, retriable sync stage DAG |
| `PipelineStage` | Name, status, progress %, dependencies, attempt count |
| `ProgressTracker` | Block/byte/header/body/receipt/state-node counters |
| `ConcurrentHealer` | Priority-queue heal scheduler with resource budgeting |
| `ResourceBudget` | Memory + bandwidth + pending-count limits |
| `SchedulerHealTask` | Hash, path, priority, deadline, estimated cost |

### Key Functions

- `NewSyncPipeline(config)` / `AddStage(name, deps)` / `StartStage(name)` / `CompleteStage(name)` / `FailStage(name, reason)` / `RetryStage(name)`
- `NewProgressTracker()` / `Start(highestBlock)` / `SetStage(stage)` / `UpdateBlock(n)` / `GetProgress()`
- `NewConcurrentHealer(cfg)` / `ScheduleTask(task)` / `ProcessBatch()` / `CompleteTaskWithCost(hash, cost)` / `FailTaskWithRetry(task)`

## Usage

```go
// Pipeline
p := support.NewSyncPipeline(support.DefaultPipelineConfig())
p.AddStage(support.PipelineStageHeaderSync, nil)
p.AddStage(support.PipelineStageBodySync, []string{support.PipelineStageHeaderSync})
p.StartStage(support.PipelineStageHeaderSync)
p.CompleteStage(support.PipelineStageHeaderSync)

// Progress
pt := support.NewProgressTracker()
pt.Start(20000)
pt.UpdateBlock(500)
info := pt.GetProgress() // info.PercentComplete, info.EstimatedCompletion
```
