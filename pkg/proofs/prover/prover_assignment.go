// prover_assignment.go implements distributed prover assignment for the
// mandatory 3-of-5 proof system. It manages a pool of provers with
// reputation scoring, capacity tracking, and geographic diversity for
// balanced and reliable proof generation across the network.
package prover

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
)

// Prover assignment errors.
var (
	ErrPoolEmpty           = errors.New("prover_pool: no registered provers")
	ErrNotEnoughProvers    = errors.New("prover_pool: not enough available provers")
	ErrProverAlreadyExists = errors.New("prover_pool: prover already registered")
	ErrProverNotRegistered = errors.New("prover_pool: prover not registered")
	ErrZeroCapacity        = errors.New("prover_pool: capacity must be positive")
	ErrInvalidCount        = errors.New("prover_pool: requested count must be positive")
	ErrInvalidMinProvers   = errors.New("prover_pool: minProvers must be positive")
	ErrProverAtCapacity    = errors.New("prover_pool: prover is at capacity")
	ErrInvalidDecayFactor  = errors.New("reputation: decay factor must be in (0, 1]")
)

// Default reputation parameters.
const (
	initialReputation = 0.5
	maxReputation     = 1.0
	minReputation     = 0.01
	successIncrement  = 0.1
	failureDecrement  = 0.2
)

// ProverCandidate represents a prover with its current reputation and capacity.
type ProverCandidate struct {
	ID         string
	Region     string
	Capacity   int
	InFlight   int
	Reputation float64
}

// Available returns the remaining capacity.
func (pc *ProverCandidate) Available() int {
	return pc.Capacity - pc.InFlight
}

// AssignmentResult holds the result of assigning provers to a block.
type AssignmentResult struct {
	BlockNum  uint64
	ProverIDs []string
	Scores    []float64
	Regions   []string
}

// ReputationScorer tracks prover reliability as the ratio of successful
// proofs to total assignments, with time-decay to deprioritize stale records.
type ReputationScorer struct {
	mu       sync.RWMutex
	scores   map[string]float64
	attempts map[string]int
}

// NewReputationScorer creates a new reputation scorer.
func NewReputationScorer() *ReputationScorer {
	return &ReputationScorer{
		scores:   make(map[string]float64),
		attempts: make(map[string]int),
	}
}

// Init sets the initial reputation for a prover.
func (rs *ReputationScorer) Init(proverID string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if _, exists := rs.scores[proverID]; !exists {
		rs.scores[proverID] = initialReputation
		rs.attempts[proverID] = 0
	}
}

// Score returns the current reputation score for a prover.
// Returns 0 if the prover is unknown.
func (rs *ReputationScorer) Score(proverID string) float64 {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.scores[proverID]
}

// RecordSuccess increases a prover's reputation after a successful proof.
func (rs *ReputationScorer) RecordSuccess(proverID string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	s, ok := rs.scores[proverID]
	if !ok {
		s = initialReputation
	}
	rs.attempts[proverID]++
	s += successIncrement
	if s > maxReputation {
		s = maxReputation
	}
	rs.scores[proverID] = s
}

// RecordFailure decreases a prover's reputation after a failed assignment.
func (rs *ReputationScorer) RecordFailure(proverID string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	s, ok := rs.scores[proverID]
	if !ok {
		s = initialReputation
	}
	rs.attempts[proverID]++
	s -= failureDecrement
	if s < minReputation {
		s = minReputation
	}
	rs.scores[proverID] = s
}

// Decay applies a multiplicative time-decay to all reputation scores.
// factor must be in (0, 1]. A factor of 0.9 reduces all scores by 10%.
func (rs *ReputationScorer) Decay(factor float64) error {
	if factor <= 0 || factor > 1 {
		return ErrInvalidDecayFactor
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	for id, s := range rs.scores {
		s *= factor
		if s < minReputation {
			s = minReputation
		}
		rs.scores[id] = s
	}
	return nil
}

// Attempts returns the total number of assignments for a prover.
func (rs *ReputationScorer) Attempts(proverID string) int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.attempts[proverID]
}

// ProverPool manages a pool of available provers with reputation-based
// assignment. Thread-safe for concurrent use.
type ProverPool struct {
	mu         sync.RWMutex
	provers    map[string]*ProverCandidate
	scorer     *ReputationScorer
	minProvers int
}

// NewProverPool creates a new prover pool requiring at least minProvers
// registered provers before assignments can proceed.
func NewProverPool(minProvers int) *ProverPool {
	if minProvers < 1 {
		minProvers = 1
	}
	return &ProverPool{
		provers:    make(map[string]*ProverCandidate),
		scorer:     NewReputationScorer(),
		minProvers: minProvers,
	}
}

// RegisterProver adds a prover to the pool with the given capacity.
// Region is optional and used for geographic diversity.
func (pp *ProverPool) RegisterProver(id string, capacity int) error {
	return pp.RegisterProverWithRegion(id, capacity, "")
}

// RegisterProverWithRegion adds a prover with a geographic region tag.
func (pp *ProverPool) RegisterProverWithRegion(id string, capacity int, region string) error {
	if capacity <= 0 {
		return ErrZeroCapacity
	}
	pp.mu.Lock()
	defer pp.mu.Unlock()

	if _, exists := pp.provers[id]; exists {
		return ErrProverAlreadyExists
	}
	pp.provers[id] = &ProverCandidate{
		ID:         id,
		Region:     region,
		Capacity:   capacity,
		InFlight:   0,
		Reputation: initialReputation,
	}
	pp.scorer.Init(id)
	return nil
}

// Size returns the number of registered provers.
func (pp *ProverPool) Size() int {
	pp.mu.RLock()
	defer pp.mu.RUnlock()
	return len(pp.provers)
}

// AssignProvers selects the best `count` provers for the given block,
// sorted by reputation. Provers at capacity are skipped. When possible,
// geographic diversity is preferred (no two provers from the same region).
func (pp *ProverPool) AssignProvers(blockNum uint64, count int) (*AssignmentResult, error) {
	if count <= 0 {
		return nil, ErrInvalidCount
	}

	pp.mu.Lock()
	defer pp.mu.Unlock()

	if len(pp.provers) == 0 {
		return nil, ErrPoolEmpty
	}

	// Gather available candidates.
	candidates := make([]*ProverCandidate, 0, len(pp.provers))
	for _, p := range pp.provers {
		if p.Available() > 0 {
			// Sync reputation from scorer.
			p.Reputation = pp.scorer.Score(p.ID)
			if p.Reputation == 0 {
				p.Reputation = initialReputation
			}
			candidates = append(candidates, p)
		}
	}

	if len(candidates) < count {
		// Also check if we have the minimum total provers.
		if len(pp.provers) < pp.minProvers {
			return nil, fmt.Errorf("%w: have %d, need %d",
				ErrNotEnoughProvers, len(pp.provers), pp.minProvers)
		}
		return nil, fmt.Errorf("%w: %d available, %d requested",
			ErrNotEnoughProvers, len(candidates), count)
	}

	// Sort by reputation (highest first), tie-break by ID for determinism.
	sort.Slice(candidates, func(i, j int) bool {
		if math.Abs(candidates[i].Reputation-candidates[j].Reputation) < 1e-9 {
			return candidates[i].ID < candidates[j].ID
		}
		return candidates[i].Reputation > candidates[j].Reputation
	})

	// Select with geographic diversity: prefer provers from different regions.
	selected := diverseSelect(candidates, count)

	// Mark in-flight.
	result := &AssignmentResult{
		BlockNum:  blockNum,
		ProverIDs: make([]string, len(selected)),
		Scores:    make([]float64, len(selected)),
		Regions:   make([]string, len(selected)),
	}
	for i, c := range selected {
		c.InFlight++
		result.ProverIDs[i] = c.ID
		result.Scores[i] = c.Reputation
		result.Regions[i] = c.Region
	}

	return result, nil
}

// diverseSelect picks `count` provers favoring geographic diversity.
// It first picks one prover per unique region from the top candidates,
// then fills remaining slots from the remaining highest-reputation provers.
func diverseSelect(candidates []*ProverCandidate, count int) []*ProverCandidate {
	if len(candidates) <= count {
		return candidates
	}

	selected := make([]*ProverCandidate, 0, count)
	usedRegions := make(map[string]bool)
	used := make(map[string]bool)

	// First pass: one per region (skip empty regions).
	for _, c := range candidates {
		if len(selected) >= count {
			break
		}
		if c.Region != "" && !usedRegions[c.Region] {
			selected = append(selected, c)
			usedRegions[c.Region] = true
			used[c.ID] = true
		}
	}

	// Second pass: fill remaining slots by reputation order.
	for _, c := range candidates {
		if len(selected) >= count {
			break
		}
		if !used[c.ID] {
			selected = append(selected, c)
			used[c.ID] = true
		}
	}

	return selected
}

// RecordSuccess records a successful proof completion for a prover.
func (pp *ProverPool) RecordSuccess(proverID string, blockNum uint64) error {
	_ = blockNum
	pp.mu.Lock()
	defer pp.mu.Unlock()

	p, ok := pp.provers[proverID]
	if !ok {
		return ErrProverNotRegistered
	}
	if p.InFlight > 0 {
		p.InFlight--
	}
	pp.scorer.RecordSuccess(proverID)
	p.Reputation = pp.scorer.Score(proverID)
	return nil
}

// RecordFailure records a failed proof attempt for a prover.
func (pp *ProverPool) RecordFailure(proverID string, blockNum uint64) error {
	_ = blockNum
	pp.mu.Lock()
	defer pp.mu.Unlock()

	p, ok := pp.provers[proverID]
	if !ok {
		return ErrProverNotRegistered
	}
	if p.InFlight > 0 {
		p.InFlight--
	}
	pp.scorer.RecordFailure(proverID)
	p.Reputation = pp.scorer.Score(proverID)
	return nil
}

// GetReputation returns the current reputation score for a prover.
func (pp *ProverPool) GetReputation(proverID string) (float64, error) {
	pp.mu.RLock()
	defer pp.mu.RUnlock()
	if _, ok := pp.provers[proverID]; !ok {
		return 0, ErrProverNotRegistered
	}
	return pp.scorer.Score(proverID), nil
}

// GetProver returns the prover candidate info, or error if not found.
func (pp *ProverPool) GetProver(proverID string) (*ProverCandidate, error) {
	pp.mu.RLock()
	defer pp.mu.RUnlock()
	p, ok := pp.provers[proverID]
	if !ok {
		return nil, ErrProverNotRegistered
	}
	// Return a copy.
	cp := *p
	return &cp, nil
}

// Scorer returns the underlying reputation scorer for direct access.
func (pp *ProverPool) Scorer() *ReputationScorer {
	return pp.scorer
}

// DecayAllReputations applies a time-decay factor to all prover reputations.
func (pp *ProverPool) DecayAllReputations(factor float64) error {
	return pp.scorer.Decay(factor)
}
