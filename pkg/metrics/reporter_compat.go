package metrics

// reporter_compat.go re-exports types from metrics/reporter for backward compatibility.

import (
	"time"

	"github.com/eth2030/eth2030/metrics/reporter"
)

// Type aliases.
type (
	ReportBackend   = reporter.ReportBackend
	MetricsReporter = reporter.MetricsReporter
)

// NewMetricsReporter creates a new MetricsReporter with the given interval.
func NewMetricsReporter(interval time.Duration) *MetricsReporter {
	return reporter.NewMetricsReporter(interval)
}
