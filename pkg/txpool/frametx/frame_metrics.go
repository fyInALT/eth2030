package frametx

import "github.com/eth2030/eth2030/metrics"

// FrameTxMetrics tracks acceptance/rejection counters for frame transactions.
// These counters are exported via /metrics (Prometheus text format).
type FrameTxMetrics struct {
	accepted             *metrics.Counter
	rejectedConservative *metrics.Counter
	rejectedAggressive   *metrics.Counter
}

// NewFrameTxMetrics creates and registers frame tx counters.
func NewFrameTxMetrics() *FrameTxMetrics {
	return &FrameTxMetrics{
		accepted:             metrics.NewCounter("frame_tx_accepted_total"),
		rejectedConservative: metrics.NewCounter("frame_tx_rejected_conservative_total"),
		rejectedAggressive:   metrics.NewCounter("frame_tx_rejected_aggressive_total"),
	}
}

// IncAccepted increments the accepted-frame-tx counter.
func (m *FrameTxMetrics) IncAccepted() { m.accepted.Inc() }

// IncRejectedConservative increments the conservative-reject counter.
func (m *FrameTxMetrics) IncRejectedConservative() { m.rejectedConservative.Inc() }

// IncRejectedAggressive increments the aggressive-reject counter.
func (m *FrameTxMetrics) IncRejectedAggressive() { m.rejectedAggressive.Inc() }

// Accepted returns the current accepted count (for testing).
func (m *FrameTxMetrics) Accepted() int64 { return m.accepted.Value() }

// RejectedConservative returns the current conservative-reject count (for testing).
func (m *FrameTxMetrics) RejectedConservative() int64 { return m.rejectedConservative.Value() }

// RejectedAggressive returns the current aggressive-reject count (for testing).
func (m *FrameTxMetrics) RejectedAggressive() int64 { return m.rejectedAggressive.Value() }
