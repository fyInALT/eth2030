# metrics/collector - Multi-subsystem metrics aggregator

## Overview

Package `collector` implements a concurrent metrics aggregator that subsystems use
to record gauges, counters, and histogram observations. All data points are stored
with timestamps and optional string tags, enabling label-based querying across
subsystems (e.g. filtering by `subsystem=txpool`).

The collector is intentionally append-only up to a configurable cap: once
`MaxMetrics` entries are reached, new recordings are silently dropped to protect
memory. The `Flush` method drains and resets state for periodic Prometheus-style
export cycles.

## Functionality

**Types**

- `CollectorConfig` - `FlushInterval int64` (seconds), `MaxMetrics int`,
  `EnableHistograms bool`.
- `MetricEntry` - `Name string`, `Value float64`, `Tags map[string]string`,
  `Timestamp int64`, `Type string` (`"gauge"` or `"histogram"`).
- `MetricsCollector` - main struct (RWMutex-protected append log + latest map +
  histogram buckets).

**Constructor**

- `NewMetricsCollector(config CollectorConfig) *MetricsCollector` - defaults:
  `MaxMetrics=10000`, `FlushInterval=60`.

**Recording**

- `Record(name string, value float64, tags map[string]string)` - stores a gauge
  entry and updates the latest map.
- `RecordHistogram(name string, value float64)` - stores a histogram observation;
  no-op when `EnableHistograms` is false.

**Querying**

- `Get(name string) *MetricEntry` - latest entry for a metric name.
- `GetAll() []MetricEntry` - copy of every recorded entry.
- `GetByTag(key, value string) []MetricEntry` - filter entries by tag key/value.
- `Summary() map[string]float64` - map of metric name to latest value.
- `HistogramPercentile(name string, percentile float64) float64` - linear
  interpolation percentile (0-100) over all recorded observations.

**Management**

- `Flush() []MetricEntry` - returns all entries and resets internal state.
- `MetricCount() int`, `Config() CollectorConfig`

Parent package: [`metrics`](../)
