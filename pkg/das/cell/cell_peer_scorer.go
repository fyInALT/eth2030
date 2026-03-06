// cell_peer_scorer.go adds per-peer cell delivery scoring for the PeerDAS
// gossip protocol. It tracks delivery reliability, penalizes redundant and
// invalid cell deliveries, factors in latency, and supports time-based score
// decay and peer banning. The scorer produces weighted sampling selections
// to bias cell requests toward reliable peers.
//
// This complements cell_gossip_scorer.go (which provides subnet-level gossip
// scoring) by adding cell-level delivery tracking per peer identity string.
package cell

import (
	"errors"
	"math"
	"sort"
	"sync"
)

// Cell peer scorer errors.
var (
	ErrCellScorerNilConfig   = errors.New("das/cell_scorer: nil config")
	ErrCellScorerEmptyPeerID = errors.New("das/cell_scorer: empty peer ID")
	ErrCellScorerZeroDecay   = errors.New("das/cell_scorer: decay rate must be > 0")
)

// CellReputationConfig configures the cell peer scoring system.
type CellReputationConfig struct {
	// BaselineScore is the initial score for a new peer.
	BaselineScore float64

	// DeliveryReward is the score increase for a valid cell delivery.
	DeliveryReward float64

	// InvalidPenalty is the score decrease for delivering an invalid cell.
	InvalidPenalty float64

	// DuplicatePenalty is the score decrease for delivering a duplicate cell.
	DuplicatePenalty float64

	// LatencyThresholdMs is the latency (in ms) above which deliveries
	// incur a penalty proportional to the excess latency.
	LatencyThresholdMs int64

	// LatencyPenaltyFactor is multiplied by (latency - threshold) / threshold
	// to compute the latency penalty. E.g., 0.5 means 50% of base reward
	// is deducted per threshold multiple of excess latency.
	LatencyPenaltyFactor float64

	// DecayRate is the multiplicative factor applied during decay (0..1).
	// Scores decay toward BaselineScore by this factor each cycle.
	DecayRate float64

	// BanThreshold is the score below which a peer is considered banned.
	BanThreshold float64
}

// DefaultCellReputationConfig returns production defaults.
func DefaultCellReputationConfig() CellReputationConfig {
	return CellReputationConfig{
		BaselineScore:        50.0,
		DeliveryReward:       2.0,
		InvalidPenalty:       -20.0,
		DuplicatePenalty:     -1.0,
		LatencyThresholdMs:   200,
		LatencyPenaltyFactor: 0.5,
		DecayRate:            0.95,
		BanThreshold:         -50.0,
	}
}

// ValidateCellReputationConfig checks the config for correctness.
func ValidateCellReputationConfig(cfg CellReputationConfig) error {
	if cfg.DecayRate <= 0 || cfg.DecayRate > 1.0 {
		return ErrCellScorerZeroDecay
	}
	return nil
}

// CellPeerStats tracks per-peer cell delivery statistics.
type CellPeerStats struct {
	// CellsDelivered is the count of valid cells delivered.
	CellsDelivered uint64

	// CellsInvalid is the count of invalid cells delivered.
	CellsInvalid uint64

	// CellsDuplicate is the count of duplicate cells delivered.
	CellsDuplicate uint64

	// TotalLatencyMs is the cumulative latency of all deliveries.
	TotalLatencyMs int64

	// Score is the current reputation score.
	Score float64
}

// AverageLatencyMs returns the average delivery latency in milliseconds.
// Returns 0 if no cells have been delivered.
func (s CellPeerStats) AverageLatencyMs() int64 {
	total := s.CellsDelivered + s.CellsInvalid + s.CellsDuplicate
	if total == 0 {
		return 0
	}
	return s.TotalLatencyMs / int64(total)
}

// SamplingWeight represents a peer's weight for sampling selection,
// derived from its reputation score.
type SamplingWeight struct {
	// PeerID is the peer's identifier.
	PeerID string

	// Weight is the normalized sampling weight (>= 0).
	Weight float64

	// Score is the raw reputation score.
	Score float64
}

// CellPeerScorer tracks per-peer cell delivery metrics and produces
// weighted sampling selections. Thread-safe.
type CellPeerScorer struct {
	mu     sync.RWMutex
	config CellReputationConfig
	peers  map[string]*cellPeerEntry
}

// cellPeerEntry is the internal per-peer tracking state.
type cellPeerEntry struct {
	stats CellPeerStats
}

// NewCellPeerScorer creates a new cell peer scorer with the given config.
func NewCellPeerScorer(cfg CellReputationConfig) (*CellPeerScorer, error) {
	if err := ValidateCellReputationConfig(cfg); err != nil {
		return nil, err
	}
	return &CellPeerScorer{
		config: cfg,
		peers:  make(map[string]*cellPeerEntry),
	}, nil
}

// getOrCreate returns the entry for a peer, creating it if needed.
// Caller must hold cs.mu write lock.
func (cs *CellPeerScorer) getOrCreate(peerID string) *cellPeerEntry {
	entry, ok := cs.peers[peerID]
	if !ok {
		entry = &cellPeerEntry{
			stats: CellPeerStats{
				Score: cs.config.BaselineScore,
			},
		}
		cs.peers[peerID] = entry
	}
	return entry
}

// RecordDelivery records a valid cell delivery from a peer with the
// given latency in milliseconds. The cell index is tracked for metrics.
func (cs *CellPeerScorer) RecordDelivery(peerID string, cellIdx uint64, latencyMs int64) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	entry := cs.getOrCreate(peerID)
	entry.stats.CellsDelivered++
	if latencyMs > 0 {
		entry.stats.TotalLatencyMs += latencyMs
	}

	// Base reward.
	reward := cs.config.DeliveryReward

	// Apply latency penalty if above threshold.
	if cs.config.LatencyThresholdMs > 0 && latencyMs > cs.config.LatencyThresholdMs {
		excess := float64(latencyMs-cs.config.LatencyThresholdMs) / float64(cs.config.LatencyThresholdMs)
		penalty := cs.config.LatencyPenaltyFactor * excess * cs.config.DeliveryReward
		reward -= penalty
	}

	entry.stats.Score += reward
}

// RecordInvalid records an invalid cell delivery from a peer.
func (cs *CellPeerScorer) RecordInvalid(peerID string, cellIdx uint64) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	entry := cs.getOrCreate(peerID)
	entry.stats.CellsInvalid++
	entry.stats.Score += cs.config.InvalidPenalty
}

// RecordDuplicate records a duplicate cell delivery from a peer.
func (cs *CellPeerScorer) RecordDuplicate(peerID string, cellIdx uint64) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	entry := cs.getOrCreate(peerID)
	entry.stats.CellsDuplicate++
	entry.stats.Score += cs.config.DuplicatePenalty
}

// PeerScore returns the current reputation score for a peer.
// Returns the baseline score if the peer is unknown.
func (cs *CellPeerScorer) PeerScore(peerID string) float64 {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	entry, ok := cs.peers[peerID]
	if !ok {
		return cs.config.BaselineScore
	}
	return entry.stats.Score
}

// SamplingWeights returns peers sorted by score in descending order,
// with non-negative weights derived from scores. Banned peers are
// excluded from the sampling list.
func (cs *CellPeerScorer) SamplingWeights() []SamplingWeight {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	weights := make([]SamplingWeight, 0, len(cs.peers))
	for id, entry := range cs.peers {
		if entry.stats.Score <= cs.config.BanThreshold {
			continue // exclude banned peers
		}
		w := entry.stats.Score - cs.config.BanThreshold
		if w < 0 {
			w = 0
		}
		weights = append(weights, SamplingWeight{
			PeerID: id,
			Weight: w,
			Score:  entry.stats.Score,
		})
	}

	sort.Slice(weights, func(i, j int) bool {
		return weights[i].Score > weights[j].Score
	})

	return weights
}

// BannedPeers returns the list of peer IDs whose score is at or below
// the ban threshold.
func (cs *CellPeerScorer) BannedPeers() []string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	var banned []string
	for id, entry := range cs.peers {
		if entry.stats.Score <= cs.config.BanThreshold {
			banned = append(banned, id)
		}
	}
	sort.Strings(banned)
	return banned
}

// DecayScores applies time-based decay to all peer scores. Each peer's
// score moves toward the baseline score by the configured decay rate.
// Should be called periodically (e.g., once per slot).
func (cs *CellPeerScorer) DecayScores() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	baseline := cs.config.BaselineScore
	rate := cs.config.DecayRate

	for _, entry := range cs.peers {
		diff := entry.stats.Score - baseline
		entry.stats.Score = baseline + diff*rate
		// Snap to baseline if very close.
		if math.Abs(entry.stats.Score-baseline) < 0.001 {
			entry.stats.Score = baseline
		}
	}
}

// PeerStats returns the statistics for a specific peer.
// Returns the stats and true if found, or zero stats and false if unknown.
func (cs *CellPeerScorer) PeerStats(peerID string) (CellPeerStats, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	entry, ok := cs.peers[peerID]
	if !ok {
		return CellPeerStats{}, false
	}
	return entry.stats, true
}

// PeerCount returns the number of tracked peers.
func (cs *CellPeerScorer) PeerCount() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return len(cs.peers)
}

// IsBanned returns true if the peer's score is at or below the ban threshold.
func (cs *CellPeerScorer) IsBanned(peerID string) bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	entry, ok := cs.peers[peerID]
	if !ok {
		return false
	}
	return entry.stats.Score <= cs.config.BanThreshold
}
