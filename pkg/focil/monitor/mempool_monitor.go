// mempool_monitor.go implements real-time inclusion list compliance monitoring
// against the live mempool with fairness metrics and censorship detection.
//
// This complements the existing ComplianceEngine (compliance_engine.go) and
// InclusionMonitor (inclusion_monitor.go) by providing mempool-level
// transaction tracking, per-sender fairness analysis, and censorship
// indicators based on repeated exclusion patterns.
package monitor

import (
	"math"
	"sort"
	"sync"
)

// pendingTx tracks a single mempool transaction.
type pendingTx struct {
	txHash    [32]byte
	sender    [20]byte
	gasPrice  uint64
	timestamp int64
	included  bool
	excluded  bool
	slot      uint64 // slot of inclusion/exclusion
}

// ComplianceMetrics summarizes mempool compliance statistics.
type ComplianceMetrics struct {
	// TotalPending is the number of pending transactions tracked.
	TotalPending int

	// IncludedCount is how many pending txs were included in blocks.
	IncludedCount int

	// ExcludedCount is how many pending txs were explicitly excluded.
	ExcludedCount int

	// UnresolvedCount is how many txs have no inclusion/exclusion record.
	UnresolvedCount int

	// ComplianceRate is IncludedCount / (IncludedCount + ExcludedCount).
	// Returns 1.0 if no inclusions or exclusions are recorded.
	ComplianceRate float64
}

// MempoolMonitor monitors mempool transactions against inclusion lists for
// compliance. It tracks pending transactions and records their
// inclusion/exclusion outcomes. Thread-safe.
type MempoolMonitor struct {
	mu  sync.RWMutex
	txs map[[32]byte]*pendingTx
}

// NewMempoolMonitor creates a new mempool monitor.
func NewMempoolMonitor() *MempoolMonitor {
	return &MempoolMonitor{
		txs: make(map[[32]byte]*pendingTx),
	}
}

// AddPendingTx registers a new pending transaction in the mempool.
// If the tx is already tracked, it is a no-op.
func (mm *MempoolMonitor) AddPendingTx(txHash [32]byte, sender [20]byte, gasPrice uint64, timestamp int64) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if _, exists := mm.txs[txHash]; exists {
		return
	}
	mm.txs[txHash] = &pendingTx{
		txHash:    txHash,
		sender:    sender,
		gasPrice:  gasPrice,
		timestamp: timestamp,
	}
}

// RecordInclusion marks a transaction as included in a block at the given slot.
func (mm *MempoolMonitor) RecordInclusion(txHash [32]byte, slot uint64) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	tx, ok := mm.txs[txHash]
	if !ok {
		// Track it even if not previously added.
		mm.txs[txHash] = &pendingTx{
			txHash:   txHash,
			included: true,
			slot:     slot,
		}
		return
	}
	tx.included = true
	tx.slot = slot
}

// RecordExclusion marks a transaction as excluded from a block at the given slot.
func (mm *MempoolMonitor) RecordExclusion(txHash [32]byte, slot uint64) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	tx, ok := mm.txs[txHash]
	if !ok {
		mm.txs[txHash] = &pendingTx{
			txHash:   txHash,
			excluded: true,
			slot:     slot,
		}
		return
	}
	tx.excluded = true
	tx.slot = slot
}

// ComplianceRate returns the ratio of included transactions to total resolved
// (included + excluded). Returns 1.0 if no transactions have been resolved.
func (mm *MempoolMonitor) ComplianceRate() float64 {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	included, excluded := 0, 0
	for _, tx := range mm.txs {
		if tx.included {
			included++
		}
		if tx.excluded {
			excluded++
		}
	}
	total := included + excluded
	if total == 0 {
		return 1.0
	}
	return float64(included) / float64(total)
}

// Metrics returns the current compliance metrics snapshot.
func (mm *MempoolMonitor) Metrics() ComplianceMetrics {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	var m ComplianceMetrics
	m.TotalPending = len(mm.txs)

	for _, tx := range mm.txs {
		if tx.included {
			m.IncludedCount++
		}
		if tx.excluded {
			m.ExcludedCount++
		}
		if !tx.included && !tx.excluded {
			m.UnresolvedCount++
		}
	}

	resolved := m.IncludedCount + m.ExcludedCount
	if resolved == 0 {
		m.ComplianceRate = 1.0
	} else {
		m.ComplianceRate = float64(m.IncludedCount) / float64(resolved)
	}
	return m
}

// PruneOlderThan removes all pending transactions with a timestamp strictly
// less than the given threshold. Returns the number of transactions pruned.
func (mm *MempoolMonitor) PruneOlderThan(timestamp int64) int {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	pruned := 0
	for hash, tx := range mm.txs {
		if tx.timestamp < timestamp {
			delete(mm.txs, hash)
			pruned++
		}
	}
	return pruned
}

// PendingCount returns the number of tracked transactions.
func (mm *MempoolMonitor) PendingCount() int {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return len(mm.txs)
}

// --- FairnessAnalyzer ---

// senderStats tracks inclusion/exclusion for a single sender.
type senderStats struct {
	inclusions int
	exclusions int
}

// FairnessAnalyzer analyzes fairness of inclusion across transaction senders.
// It computes a Gini-like fairness score and identifies potentially censored
// senders. Thread-safe.
type FairnessAnalyzer struct {
	mu      sync.RWMutex
	senders map[[20]byte]*senderStats
}

// NewFairnessAnalyzer creates a new fairness analyzer.
func NewFairnessAnalyzer() *FairnessAnalyzer {
	return &FairnessAnalyzer{
		senders: make(map[[20]byte]*senderStats),
	}
}

// RecordSenderInclusion records that a sender's transaction was included.
func (fa *FairnessAnalyzer) RecordSenderInclusion(sender [20]byte) {
	fa.mu.Lock()
	defer fa.mu.Unlock()

	s := fa.getOrCreate(sender)
	s.inclusions++
}

// RecordSenderExclusion records that a sender's transaction was excluded.
func (fa *FairnessAnalyzer) RecordSenderExclusion(sender [20]byte) {
	fa.mu.Lock()
	defer fa.mu.Unlock()

	s := fa.getOrCreate(sender)
	s.exclusions++
}

// FairnessScore computes a fairness measure across all senders, where 1.0
// means perfectly fair (all senders have identical inclusion rates) and 0.0
// means maximally unfair. Based on 1 - Gini coefficient of inclusion rates.
func (fa *FairnessAnalyzer) FairnessScore() float64 {
	fa.mu.RLock()
	defer fa.mu.RUnlock()

	if len(fa.senders) == 0 {
		return 1.0
	}

	// Collect inclusion rates per sender.
	rates := make([]float64, 0, len(fa.senders))
	for _, s := range fa.senders {
		total := s.inclusions + s.exclusions
		if total == 0 {
			rates = append(rates, 1.0) // no data = assume fair
			continue
		}
		rates = append(rates, float64(s.inclusions)/float64(total))
	}

	if len(rates) <= 1 {
		return 1.0
	}

	// Compute Gini coefficient.
	gini := computeGini(rates)

	// Return 1 - Gini so 1.0 = perfectly fair.
	score := 1.0 - gini
	if score < 0 {
		return 0
	}
	return math.Round(score*1e12) / 1e12
}

// SuspectedCensored returns senders whose inclusion rate is below 0.25
// (i.e., more than 75% exclusion rate) and who have at least 3 total events.
func (fa *FairnessAnalyzer) SuspectedCensored() [][20]byte {
	fa.mu.RLock()
	defer fa.mu.RUnlock()

	var result [][20]byte
	for addr, s := range fa.senders {
		total := s.inclusions + s.exclusions
		if total < 3 {
			continue
		}
		rate := float64(s.inclusions) / float64(total)
		if rate < 0.25 {
			result = append(result, addr)
		}
	}

	// Sort for deterministic output.
	sort.Slice(result, func(i, j int) bool {
		for k := 0; k < 20; k++ {
			if result[i][k] != result[j][k] {
				return result[i][k] < result[j][k]
			}
		}
		return false
	})
	return result
}

// SenderCount returns the number of tracked senders.
func (fa *FairnessAnalyzer) SenderCount() int {
	fa.mu.RLock()
	defer fa.mu.RUnlock()
	return len(fa.senders)
}

// getOrCreate returns or creates stats for a sender.
// Must be called with the write lock held.
func (fa *FairnessAnalyzer) getOrCreate(sender [20]byte) *senderStats {
	s, ok := fa.senders[sender]
	if !ok {
		s = &senderStats{}
		fa.senders[sender] = s
	}
	return s
}

// --- CensorshipIndicator ---

// CensorshipIndicator tracks per-sender inclusion outcomes and flags senders
// whose exclusion rate exceeds a configurable threshold. Thread-safe.
type CensorshipIndicator struct {
	mu        sync.RWMutex
	threshold float64 // exclusion rate above which a sender is considered censored
	senders   map[[20]byte]*senderStats
}

// NewCensorshipIndicator creates a new indicator with the given exclusion
// rate threshold (0.0 to 1.0). A threshold of 0.7 means a sender is
// flagged if 70% or more of their outcomes are exclusions.
func NewCensorshipIndicator(threshold float64) *CensorshipIndicator {
	if threshold <= 0 || threshold > 1.0 {
		threshold = 0.7
	}
	return &CensorshipIndicator{
		threshold: threshold,
		senders:   make(map[[20]byte]*senderStats),
	}
}

// RecordOutcome records an inclusion (true) or exclusion (false) for a sender.
func (ci *CensorshipIndicator) RecordOutcome(sender [20]byte, included bool) {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	s, ok := ci.senders[sender]
	if !ok {
		s = &senderStats{}
		ci.senders[sender] = s
	}
	if included {
		s.inclusions++
	} else {
		s.exclusions++
	}
}

// IsCensored returns true if the sender's exclusion rate exceeds the threshold
// and the sender has at least 2 total outcomes.
func (ci *CensorshipIndicator) IsCensored(sender [20]byte) bool {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	s, ok := ci.senders[sender]
	if !ok {
		return false
	}
	total := s.inclusions + s.exclusions
	if total < 2 {
		return false
	}
	exclusionRate := float64(s.exclusions) / float64(total)
	return exclusionRate >= ci.threshold
}

// CensoredSenders returns all senders currently flagged as censored.
func (ci *CensorshipIndicator) CensoredSenders() [][20]byte {
	ci.mu.RLock()
	defer ci.mu.RUnlock()

	var result [][20]byte
	for addr, s := range ci.senders {
		total := s.inclusions + s.exclusions
		if total < 2 {
			continue
		}
		exclusionRate := float64(s.exclusions) / float64(total)
		if exclusionRate >= ci.threshold {
			result = append(result, addr)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		for k := 0; k < 20; k++ {
			if result[i][k] != result[j][k] {
				return result[i][k] < result[j][k]
			}
		}
		return false
	})
	return result
}

// SenderCount returns the number of tracked senders.
func (ci *CensorshipIndicator) SenderCount() int {
	ci.mu.RLock()
	defer ci.mu.RUnlock()
	return len(ci.senders)
}

// --- Internal helpers ---

// computeGini calculates the Gini coefficient for a slice of values.
// Values should be non-negative. Returns 0.0 for empty or single-element input.
func computeGini(values []float64) float64 {
	n := len(values)
	if n <= 1 {
		return 0.0
	}

	// Sort values ascending.
	sorted := make([]float64, n)
	copy(sorted, values)
	sort.Float64s(sorted)

	// Gini = (2 * sum(i * x_i) / (n * sum(x_i))) - (n+1)/n
	var sumXi, sumIXi float64
	for i, x := range sorted {
		sumXi += x
		sumIXi += float64(i+1) * x
	}

	if sumXi == 0 {
		return 0.0
	}

	gini := (2.0*sumIXi)/(float64(n)*sumXi) - (float64(n)+1.0)/float64(n)
	if gini < 0 {
		return 0.0
	}
	if gini > 1.0 {
		return 1.0
	}
	return gini
}
