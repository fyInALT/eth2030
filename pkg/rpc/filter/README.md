# filter — log and block filter system

[← rpc](../README.md)

## Overview

Package `rpcfilter` implements the server-side filter system that backs
`eth_newFilter`, `eth_newBlockFilter`, `eth_getFilterChanges`, and
`eth_getLogs`. It maintains a registry of active log and block filters,
evaluates incoming log events against address/topic criteria, and expires idle
filters after a configurable timeout.

Additional files add an event hub (`filter_event_hub.go`) for broadcasting new
blocks and logs to registered filters, and an extended system
(`filter_extended.go`) with iterator-style access.

## Functionality

**Configuration**

- `FilterConfig` — `MaxFilters int`, `FilterTimeout time.Duration`, `MaxLogs int`
- `DefaultFilterConfig()` — returns `MaxFilters=100`, `FilterTimeout=5m`, `MaxLogs=10000`

**Core types**

- `FilterSystem` — constructed with `NewFilterSystem(config FilterConfig)`; manages the filter registry
- `FSLogFilter` — log filter entry: `ID Hash`, `FromBlock/ToBlock uint64`, `Addresses []Address`, `Topics [][]Hash`, `Logs []*types.Log`
- `FSBlockFilter` — block hash filter entry: `ID Hash`, `BlockHashes []Hash`
- `FilterQuery` — query parameters: `Addresses`, `Topics`, `FromBlock`, `ToBlock` (shared with `subscription` and `ethapi`)

**`FilterSystem` methods**

| Method | Description |
|---|---|
| `NewLogFilter(from, to, addrs, topics)` | Install a log filter; returns filter ID |
| `NewBlockFilter()` | Install a block hash filter; returns filter ID |
| `GetFilterChanges(id Hash)` | Drain and return accumulated events since last poll |
| `UninstallFilter(id Hash) bool` | Remove a filter |
| `ExpireFilters(now time.Time) int` | Remove filters idle longer than `FilterTimeout` |

**Log matching**

- `MatchLog(log *types.Log, addrs []Address, topics [][]Hash) bool` — standalone predicate used by both polling and streaming paths
