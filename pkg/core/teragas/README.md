# core/teragas — Teragas L2 blob scheduling (1 GiB/sec target)

[← core](../README.md)

## Overview

Package `teragas` implements the blob scheduling infrastructure for the Teragas L2 North Star goal: 1 GiB/sec (1 Gbyte/sec) of L2 data throughput. It manages a priority-ordered queue of blob requests, enforces per-slot bandwidth budgets, and produces scheduling decisions with estimated delivery times and allocated bandwidth.

Blobs are ordered by priority (descending), then by earliest deadline, then by arrival sequence. The scheduler runs in a slot-based model: `ProcessQueue` drains blobs up to the slot's bandwidth budget and advances the slot counter.

## Functionality

```go
const TeragasTarget int64 = 1 << 30  // 1 GiB/sec

type SchedulerConfig struct {
    MaxQueueSize    int           // default 4096
    TargetBps       int64         // default TeragasTarget
    DefaultPriority int
    SlotDuration    time.Duration // default 12s
    MaxBlobSize     int64         // default 4 MiB
}

func DefaultSchedulerConfig() SchedulerConfig
func NewTeragasScheduler(config SchedulerConfig) *TeragasScheduler

func (ts *TeragasScheduler) ScheduleBlob(req BlobRequest) (ScheduleResult, error)
func (ts *TeragasScheduler) ProcessQueue() (processedCount int, processedBytes int64)
func (ts *TeragasScheduler) QueueLength() int
func (ts *TeragasScheduler) Metrics() (total, processed, dropped int64, avgLatency time.Duration)
func (ts *TeragasScheduler) Stop()
func (ts *TeragasScheduler) IsStopped() bool

type BlobRequest struct {
    Data         []byte
    Priority     int
    Deadline     time.Time
    MaxBandwidth int64
    ID           string
    SubmitTime   time.Time
}

type ScheduleResult struct {
    Slot              uint64
    EstimatedDelivery time.Time
    AllocatedBps      int64
    QueuePosition     int
    WaitEstimate      time.Duration
    RequestID         string
}
```

### Errors

`ErrTeragasQueueFull`, `ErrTeragasDeadlineExpired`, `ErrTeragasInvalidPriority`, `ErrTeragasEmptyData`, `ErrTeragasBandwidthZero`, `ErrTeragasSchedulerStopped`

## Usage

```go
sched := teragas.NewTeragasScheduler(teragas.DefaultSchedulerConfig())

result, err := sched.ScheduleBlob(teragas.BlobRequest{
    Data:     blobData,
    Priority: 10,
    ID:       "blob-001",
})
fmt.Printf("slot=%d delivery=%v bps=%d\n",
    result.Slot, result.EstimatedDelivery, result.AllocatedBps)

// Per-slot processing loop:
count, bytes := sched.ProcessQueue()
```
