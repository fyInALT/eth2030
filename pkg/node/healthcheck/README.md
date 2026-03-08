# node/healthcheck - Subsystem health monitoring

## Overview

Package `healthcheck` provides a pluggable health monitoring framework for the
ETH2030 node. Individual subsystems register a `SubsystemChecker` implementation;
the `HealthChecker` runs all checks on demand and aggregates the results into a
single `HealthReport` with an overall status of `"healthy"`, `"degraded"`, or
`"unhealthy"`. The overall status degrades to the worst individual status observed.

Each check result records latency (nanoseconds) and the unix timestamp of
execution, making the reports suitable for Prometheus scraping or HTTP health
endpoints. Uptime is tracked from the moment `NewHealthChecker` is called (or an
explicit `SetStartTime` call).

## Functionality

**Interface**

- `SubsystemChecker` - `Check() *SubsystemHealth`; implemented by each subsystem.

**Types**

- `SubsystemHealth` - `Name string`, `Status string`, `Message string`,
  `LastCheck int64`, `Latency int64`.
- `HealthReport` - `OverallStatus string`, `Subsystems []*SubsystemHealth`,
  `CheckedAt int64`, `NodeUptime int64`.
- `HealthChecker` - main aggregator (RWMutex-protected checker registry).

**Status constants**

`StatusHealthy = "healthy"`, `StatusDegraded = "degraded"`,
`StatusUnhealthy = "unhealthy"`

**Constructor**

- `NewHealthChecker() *HealthChecker`

**Methods**

- `RegisterSubsystem(name string, checker SubsystemChecker)` - adds or replaces a
  checker; registration order is preserved for report output.
- `CheckAll() *HealthReport` - runs every checker sequentially and returns the
  aggregate report.
- `CheckSubsystem(name string) (*SubsystemHealth, error)` - runs a single checker
  by name.
- `IsHealthy() bool` - returns true only when all subsystems report healthy.
- `RegisteredSubsystems() []string`, `SortedSubsystems() []string`
- `Uptime() int64`, `SetStartTime(t int64)`, `SubsystemCount() int`

Parent package: [`node`](../)
