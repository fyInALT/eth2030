# node/lifecycle - Node service startup and shutdown manager

## Overview

Package `lifecycle` manages the ordered startup and graceful shutdown of all
services that make up an ETH2030 node. Services implement a minimal `Service`
interface (`Start`, `Stop`, `Name`) and are registered with an integer priority.
`StartAll` launches them in ascending priority order; `StopAll` tears them down in
reverse, mirroring the dependency order.

Each service entry transitions through a well-defined state machine
(`created → starting → running → stopping → stopped`, with `failed` as an error
sink). The manager tracks per-service state, start time, and the first error
returned by `Start` or `Stop`.

## Functionality

**Interface**

- `Service` - `Start() error`, `Stop() error`, `Name() string`.

**Types**

- `ServiceState` (`StateCreated`, `StateStarting`, `StateRunning`, `StateStopping`,
  `StateStopped`, `StateFailed`) with `String() string`.
- `ServiceEntry` - `Svc Service`, `State ServiceState`, `StartedAt time.Time`,
  `Error error`, `Priority int`.
- `LifecycleConfig` - `ShutdownTimeout time.Duration`, `GracePeriod time.Duration`,
  `MaxServices int`.
- `LifecycleManager` - main manager (mutex-protected service list and name index).

**Constructors**

- `DefaultLifecycleConfig() LifecycleConfig` - 30 s shutdown timeout, 5 s grace
  period, max 32 services.
- `NewLifecycleManager(config LifecycleConfig) *LifecycleManager`

**Methods**

- `Register(svc Service, priority int) error` - adds a service; returns an error
  if the name is already registered or the cap is reached.
- `StartAll() []error` - starts all services in priority order; collects errors
  without aborting remaining services.
- `StopAll() []error` - stops all running services in reverse priority order.
- `GetState(name string) ServiceState`
- `ServiceCount() int`, `RunningCount() int`
- `HealthCheck() map[string]bool` - returns a name-to-running map for all services.

Parent package: [`node`](../)
