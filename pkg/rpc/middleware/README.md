# middleware — RPC rate limiting and HTTP middleware

[← rpc](../README.md)

## Overview

Package `middleware` provides HTTP-level and JSON-RPC-level middleware for the
RPC server. It includes a token-bucket rate limiter with per-client, per-method,
and global limits, as well as a lifecycle manager for admin endpoints. A
separate `http.go` provides generic HTTP middleware helpers (CORS, logging,
request-size limiting).

## Functionality

**Rate limiting (`rate_limiter.go`)**

- `RPCRateLimitConfig` — `GlobalRPS`, `PerClientRPS`, `PerMethodRPS`, `BurstMultiplier`, `BanDurationSecs`; `DefaultRPCRateLimitConfig()` sets `GlobalRPS=1000`, `PerClientRPS=100`, `PerMethodRPS=50`
- `RPCRateLimiter` — constructed with `NewRPCRateLimiter(config *RPCRateLimitConfig)`

| Method | Description |
|---|---|
| `Allow(clientIP, method string) bool` | Token-bucket check; updates counters |
| `Ban(clientIP string, durationSecs int64)` | Manually ban a client |
| `Unban(clientIP string)` | Remove a ban |
| `IsBanned(clientIP string) bool` | Query ban status |
| `ClientStats(clientIP string) *ClientRateStats` | Per-client totals and ban expiry |
| `MethodStats(method string) *MethodRateStats` | Per-method totals and avg latency |
| `RecordLatency(method string, latencyNs int64)` | Feed latency sample for `AvgLatencyMs` |
| `PruneInactive(beforeTimestamp int64) int` | Evict idle client entries |
| `GlobalStats() *GlobalRateStats` | Aggregate totals and active/banned counts |

- Stat types: `ClientRateStats`, `MethodRateStats`, `GlobalRateStats`

**Admin lifecycle (`admin_lifecycle.go`)**

- `AdminLifecycleManager` — tracks server start/stop lifecycle and exposes state for `admin_startRPC` / `admin_stopRPC`
