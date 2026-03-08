# das/rpo — Rows-per-operation (RPO) management for blob throughput

Manages the RPO parameter that governs DAS blob throughput scaling across the
BPO upgrade schedule (J+ BPO3 and L+ BPO4 eras).

## Overview

RPO (rows-per-operation) controls how many blob rows a node processes per
sampling operation. `RPOManager` enforces monotonically increasing RPO
transitions bounded by `[MinRPO, MaxRPO]` with a configurable step-size
constraint per upgrade. It maintains a planned `RPOSchedule` (keyed by epoch)
and a history of past transitions.

`CalculateThroughput` estimates blobs-per-slot, data rate, samples needed, and
validation time at a given RPO. Predefined schedule constructors `BPO3Schedule`
and `BPO4Schedule` return the J+ and L+ upgrade schedules respectively, which
can be merged with `MergeBPOSchedules`.

## Functionality

**Types**
- `RPOConfig` — `InitialRPO`, `MaxRPO`, `MinRPO`, `RPOStepSize`
- `RPOManager` — thread-safe manager with current value, schedule, and history
- `RPOSchedule` — `{Epoch, TargetRPO, Description}`
- `RPOHistoryEntry` — `{Epoch, OldRPO, NewRPO}`
- `ThroughputEstimate` — `BlobsPerSlot`, `DataRateKBps`, `SamplesNeeded`, `ValidationTimeMs`

**Functions**
- `DefaultRPOConfig() RPOConfig`
- `NewRPOManager(config) *RPOManager`
- `(*RPOManager).CurrentRPO() uint64`
- `(*RPOManager).IncreaseRPO(newRPO) error`
- `(*RPOManager).ValidateRPOTransition(current, target) error`
- `(*RPOManager).CalculateThroughput(rpo) *ThroughputEstimate`
- `(*RPOManager).SetSchedule(schedule) error`
- `(*RPOManager).GetScheduledRPO(epoch) uint64`
- `(*RPOManager).GetHistory() []*RPOHistoryEntry`
- `ValidateBlobSchedule(schedule, config) error`
- `BPO3Schedule() []*RPOSchedule`
- `BPO4Schedule() []*RPOSchedule`
- `MergeBPOSchedules(phases...) ([]*RPOSchedule, error)`

## Usage

```go
mgr := rpo.NewRPOManager(rpo.DefaultRPOConfig())

schedule, _ := rpo.MergeBPOSchedules(rpo.BPO3Schedule(), rpo.BPO4Schedule())
mgr.SetSchedule(schedule)

effective := mgr.GetScheduledRPO(350000) // epoch 350000
est := mgr.CalculateThroughput(effective)
```

[← das](../README.md)
