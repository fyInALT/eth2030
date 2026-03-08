# metrics/ewmautil - Exponentially weighted moving average (EWMA)

## Overview

Package `ewmautil` implements a concurrency-safe exponentially weighted moving
average for computing per-second event rates over sliding time windows. It follows
the same decay model used by Unix load averages and the go-metrics library: every
5-second tick the uncounted samples are incorporated into the running rate using
`rate += alpha * (instantRate - rate)`.

Three standard windows are provided (1-minute, 5-minute, 15-minute) alongside a
general `StandardEWMA` constructor for custom alpha values.

## Functionality

**Type**

- `EWMA` - holds the alpha decay factor, an `atomic.Int64` for lock-free
  accumulation between ticks, and a mutex-protected rate float.

**Constructors**

- `StandardEWMA(alpha float64) *EWMA` - base constructor; tick interval defaults
  to 5 seconds.
- `NewEWMA1() *EWMA` - 1-minute window, `alpha = 1 - exp(-5/60)`.
- `NewEWMA5() *EWMA` - 5-minute window, `alpha = 1 - exp(-5/300)`.
- `NewEWMA15() *EWMA` - 15-minute window, `alpha = 1 - exp(-5/900)`.

**Methods**

- `Update(n int64)` - adds `n` to the uncounted total (atomic, no lock).
- `Tick()` - drains the uncounted total, computes the instant rate
  (`count / interval`), and decays the running rate. Should be called every 5
  seconds by a background ticker.
- `Rate() float64` - returns the current smoothed rate in events per second.

Parent package: [`metrics`](../)
