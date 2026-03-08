# subscription — WebSocket subscription manager

[← rpc](../README.md)

## Overview

Package `rpcsub` provides the subscription management layer for
`eth_subscribe` and `eth_unsubscribe` over WebSocket connections. It supports
four subscription types (`newHeads`, `logs`, `newPendingTransactions`,
`syncing`), enforces per-connection limits and event-rate limits, and delivers
notifications through buffered channels. A `Manager` tracks all active
subscriptions; a `Dispatcher` pushes events to matching subscribers.

## Functionality

**Subscription kinds**

- `SubKindNewHeads`, `SubKindLogs`, `SubKindPendingTx`, `SubKindSyncStatus`
- `ParseSubKind(name string) (SubKind, error)` — converts subscription type name

**Core types**

- `SubEntry` — a single subscription: `ID string`, `Kind SubKind`, `ConnID string`, `Query rpcfilter.FilterQuery`, `CreatedAt time.Time`; `Channel() <-chan interface{}` for event delivery
- `SubRateLimitConfig` — `MaxSubsPerConn`, `WindowDuration`, `MaxEventsPerSec`; `DefaultSubRateLimitConfig()`
- `Manager` — constructed with `NewManager(config SubRateLimitConfig)`

**`Manager` methods**

| Method | Description |
|---|---|
| `Subscribe(connID, typeName string, query FilterQuery) (*SubEntry, error)` | Create subscription; enforces per-conn limit |
| `Unsubscribe(id string) bool` | Remove a subscription |
| `UnsubscribeConn(connID string) int` | Remove all subscriptions for a connection |
| `Count() int` | Total active subscriptions |
| `ConnCount(connID string) int` | Subscriptions for one connection |

**`Dispatcher`** — pushes events to matching `SubEntry` channels; `DispatchNewHead`, `DispatchLog`, `DispatchPendingTx`, `DispatchSyncStatus`

**Errors**: `ErrSubManagerClosed`, `ErrSubManagerCapacity`, `ErrSubManagerNotFound`, `ErrSubManagerRateLimit`, `ErrSubManagerInvalidTyp`
