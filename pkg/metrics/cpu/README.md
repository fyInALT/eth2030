# metrics/cpu - Process CPU utilization tracker

## Overview

Package `cpu` provides lightweight process CPU utilization sampling for the
ETH2030 metrics subsystem. On Linux it reads `/proc/self/stat` and `/proc/stat`
to compute the fraction of total CPU cycles consumed by the current process since
the last sample. On non-Linux platforms it falls back to goroutine count as a
rough proxy.

Callers periodically invoke `RecordCPU` (e.g. once per metrics tick) and then
read the smoothed utilization via `Usage`. The tracker is safe for concurrent use.

## Functionality

**Types**

- `CPUStats` - raw sample: `GlobalTime int64` (total CPU jiffies across all
  processes), `GlobalWait int64` (I/O wait jiffies), `LocalTime int64` (jiffies
  consumed by this process).
- `CPUTracker` - stateful tracker holding the previous sample and computed
  utilization percentage.

**Functions**

- `ReadCPUStats() *CPUStats` - reads `/proc/self/stat` and `/proc/stat`; falls
  back to `runtime.NumGoroutine()` on non-Linux hosts.

**Constructor**

- `NewCPUTracker() *CPUTracker` - takes an initial sample immediately.

**Methods**

- `RecordCPU()` - takes a new sample and updates the stored utilization:
  `usage = (localDelta / globalDelta) * 100 * numCPU`.
- `Usage() float64` - returns the current utilization percentage
  (range 0 to 100 * `runtime.NumCPU()`).

Parent package: [`metrics`](../)
