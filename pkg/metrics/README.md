# Package metrics

Lightweight, zero-dependency metrics primitives with Prometheus export for the ETH2030 Ethereum execution client.

## Overview

The `metrics` package provides the core instrumentation layer for ETH2030. It defines three primitive metric types — `Counter`, `Gauge`, and `Histogram` — all backed by atomic operations or a single mutex for lock-free concurrent access. A `Meter` type adds EWMA-based rate tracking (1-, 5-, and 15-minute moving averages) modelled on Unix load averages.

All metrics are held in a `Registry` with get-or-create semantics, so callers never check for nil. A `DefaultRegistry` is pre-populated with standard metrics covering chain, transaction pool, P2P, RPC, EVM, and Engine API subsystems. The `PrometheusExporter` serves these metrics in Prometheus text exposition format at `/metrics` over HTTP, augmented by optional Go runtime stats (goroutines, memory, GC) and pluggable `CustomCollector` implementations.

Subpackages provide supporting utilities: EWMA computation, CPU usage tracking, and a general-purpose metric collector interface.

## Table of Contents

- [Functionality](#functionality)
- [Subpackages](#subpackages)
- [Standard Metrics](#standard-metrics)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### Counter

Monotonically incrementing counter backed by `atomic.Int64`:

```go
NewCounter(name string) *Counter
```
- `Inc()` — increment by 1
- `Add(n int64)` — increment by n (negative values are ignored)
- `Value() int64` — current count

### Gauge

Bidirectional value backed by `atomic.Int64`:

```go
NewGauge(name string) *Gauge
```
- `Set(v int64)` / `Inc()` / `Dec()`
- `Value() int64`

### Histogram

Distribution tracker (count, sum, min, max, mean) protected by a mutex:

```go
NewHistogram(name string) *Histogram
```
- `Observe(v float64)` — record a value
- `Count() / Sum() / Min() / Max() / Mean()`

### Timer

Convenience wrapper that records elapsed milliseconds into a `Histogram`:

```go
t := metrics.NewTimer(metrics.BlockProcessTime)
defer t.Stop()
```

### Meter

Event-rate tracker with 1-, 5-, and 15-minute EWMA rates (Unix load-average model):

```go
NewMeter() *Meter
```
- `Mark(n int64)` — record n events
- `Rate1() / Rate5() / Rate15()` — EWMA rate per second
- `RateMean()` — mean rate since creation
- `Count()` — total events

### Registry

Get-or-create container for all metric types:

```go
NewRegistry() *Registry
DefaultRegistry // process-wide global
```
- `Counter(name) / Gauge(name) / Histogram(name)` — return existing or create new
- `Snapshot() map[string]interface{}` — point-in-time copy of all values

### Prometheus Exporter

HTTP handler serving Prometheus text exposition format:

```go
NewPrometheusExporter(registry *Registry, config PrometheusConfig) *PrometheusExporter
```
- `Handler() http.Handler` — mounts at `config.Path` (default `/metrics`)
- `RegisterCollector(name, c)` / `UnregisterCollector(name)` — plug in custom `CustomCollector` implementations
- `DefaultPrometheusConfig()` — `Namespace="ETH2030"`, `EnableRuntime=true`, `Path="/metrics"`

When `EnableRuntime` is set, the exporter emits: goroutine count, OS threads, heap/sys/stack memory, GC cycle count, GC pause total, last GC time, and process start time.

### System Metrics

`SystemMetrics` aggregates runtime stats on demand via callback functions:

```go
type PeerCountFunc func() int
type BlockHeightFunc func() uint64
type SyncProgressFunc func() float64
```

## Subpackages

| Subpackage | Description |
|------------|-------------|
| [`collector/`](./collector/) | Generic `Collector` interface and aggregation helpers for custom metric sources |
| [`cpu/`](./cpu/) | CPU usage tracker using OS-level process stats |
| [`ewmautil/`](./ewmautil/) | EWMA (exponentially weighted moving average) computation primitives |

## Standard Metrics

Pre-defined metrics in `DefaultRegistry` (see `standard.go`):

| Name | Type | Description |
|------|------|-------------|
| `chain.height` | Gauge | Latest block number |
| `chain.block_process_ms` | Histogram | Block processing duration (ms) |
| `chain.blocks_inserted` | Counter | Blocks appended to chain |
| `chain.reorgs` | Counter | Chain reorganisation events |
| `txpool.pending` | Gauge | Pending transaction count |
| `txpool.queued` | Gauge | Queued transaction count |
| `txpool.added` | Counter | Transactions added |
| `txpool.dropped` | Counter | Transactions dropped |
| `p2p.peers` | Gauge | Connected peer count |
| `p2p.messages_received` | Counter | devp2p messages received |
| `p2p.messages_sent` | Counter | devp2p messages sent |
| `rpc.requests` | Counter | JSON-RPC requests |
| `rpc.errors` | Counter | JSON-RPC error responses |
| `rpc.latency_ms` | Histogram | JSON-RPC request latency (ms) |
| `evm.executions` | Counter | EVM call/create invocations |
| `evm.gas_used` | Counter | Total gas consumed |
| `engine.new_payload` | Counter | `engine_newPayload` calls |
| `engine.forkchoice_updated` | Counter | `engine_forkchoiceUpdated` calls |

## Usage

```go
// Use a pre-defined standard metric
metrics.ChainHeight.Set(int64(block.Number().Uint64()))
metrics.BlocksInserted.Inc()

// Time a block processing operation
t := metrics.NewTimer(metrics.BlockProcessTime)
processBlock(block)
t.Stop()

// Create a custom metric in the default registry
myCounter := metrics.DefaultRegistry.Counter("mysubsystem.events")
myCounter.Add(5)

// Start a Prometheus HTTP server
exporter := metrics.NewPrometheusExporter(metrics.DefaultRegistry, metrics.DefaultPrometheusConfig())
http.ListenAndServe(":9001", exporter.Handler())

// Register a custom collector
exporter.RegisterCollector("mempool", &myMempoolCollector{})
```

## Documentation References

- [Prometheus exposition format](https://prometheus.io/docs/instrumenting/exposition_formats/)
- ETH2030 node metrics server: `pkg/node/node.go`
