package metrics

// collector_compat.go re-exports types from metrics/collector for backward compatibility.

import "github.com/eth2030/eth2030/metrics/collector"

// Collector type aliases.
type (
	CollectorConfig  = collector.CollectorConfig
	MetricEntry      = collector.MetricEntry
	MetricsCollector = collector.MetricsCollector
)

// Collector function wrappers.
func NewMetricsCollector(config CollectorConfig) *MetricsCollector {
	return collector.NewMetricsCollector(config)
}
