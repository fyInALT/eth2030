// builder_coordinator.go implements distributed builder coordination for ePBS
// with Vickrey auction settlement, bid validation, builder reputation tracking,
// and MEV burn calculation.
package util

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Builder coordinator errors.
var (
	ErrCoordNoBids          = errors.New("coordinator: no bids for slot")
	ErrCoordBuilderNotFound = errors.New("coordinator: builder not found")
	ErrCoordBuilderExists   = errors.New("coordinator: builder already registered")
	ErrCoordMaxBuilders     = errors.New("coordinator: max builders reached")
	ErrCoordBidDeadline     = errors.New("coordinator: bid submitted after deadline")
	ErrCoordBidIncrement    = errors.New("coordinator: bid does not meet min increment")
	ErrCoordZeroStake       = errors.New("coordinator: builder stake must be > 0")
	ErrCoordDuplicateBid    = errors.New("coordinator: duplicate bid for slot")
	ErrCoordSlotZero        = errors.New("coordinator: slot must be > 0")
)

// CoordinatorConfig configures the BuilderCoordinator.
type CoordinatorConfig struct {
	// MaxBuilders is the maximum number of concurrently registered builders.
	MaxBuilders int

	// BidTimeout is the duration after which bids for a slot are no longer accepted.
	BidTimeout time.Duration

	// MinBidIncrement is the minimum amount a new bid must exceed the current
	// best bid for the same slot from the same builder.
	MinBidIncrement uint64

	// ReputationDecay controls how much a failed delivery reduces reputation.
	// Value between 0.0 and 1.0; higher means more penalty per failure.
	ReputationDecay float64
}

// DefaultCoordinatorConfig returns sensible defaults.
func DefaultCoordinatorConfig() CoordinatorConfig {
	return CoordinatorConfig{
		MaxBuilders:     64,
		BidTimeout:      4 * time.Second,
		MinBidIncrement: 1,
		ReputationDecay: 0.1,
	}
}

// BuilderReputation tracks the reliability score for a single builder.
type BuilderReputation struct {
	BuilderID          string
	TotalWins          uint64
	SuccessfulDelivery uint64
	FailedDelivery     uint64
	LastActive         time.Time
}

// Score returns the reputation score as successfulDeliveries / totalWins.
// A builder with no wins has a default score of 0.5 (neutral).
func (br *BuilderReputation) Score() float64 {
	if br.TotalWins == 0 {
		return 0.5
	}
	return float64(br.SuccessfulDelivery) / float64(br.TotalWins)
}

// coordinatorBuilder represents a registered builder within the coordinator.
type coordinatorBuilder struct {
	ID     string
	Stake  uint64
	Active bool
	Rep    *BuilderReputation
}

// coordinatorBid represents a bid submitted to the coordinator.
type coordinatorBid struct {
	BuilderID   string
	Slot        uint64
	Amount      uint64
	PayloadHash [32]byte
	SubmitTime  time.Time
}

// AuctionSettlement is the result of settling a Vickrey auction for a slot.
type AuctionSettlement struct {
	Slot        uint64
	WinnerID    string
	WinnerBid   uint64
	SettlePrice uint64 // Vickrey: second-highest bid (or winner's own if sole bidder)
	RunnerUpID  string // empty if only one bidder
	TotalBids   int
}

// BuilderCoordinator manages the full lifecycle of builder bids for a slot:
// registration, bidding, Vickrey settlement, delivery recording, and reputation.
type BuilderCoordinator struct {
	mu        sync.RWMutex
	config    CoordinatorConfig
	builders  map[string]*coordinatorBuilder
	bids      map[uint64][]*coordinatorBid // slot -> bids
	deadlines map[uint64]time.Time         // slot -> first-bid timestamp (deadline anchor)
}

// NewBuilderCoordinator creates a new coordinator with the given config.
func NewBuilderCoordinator(cfg CoordinatorConfig) *BuilderCoordinator {
	return &BuilderCoordinator{
		config:    cfg,
		builders:  make(map[string]*coordinatorBuilder),
		bids:      make(map[uint64][]*coordinatorBid),
		deadlines: make(map[uint64]time.Time),
	}
}

// RegisterBuilder registers a new builder with the given ID and stake.
func (bc *BuilderCoordinator) RegisterBuilder(builderID string, stake uint64) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if _, exists := bc.builders[builderID]; exists {
		return ErrCoordBuilderExists
	}
	if len(bc.builders) >= bc.config.MaxBuilders {
		return ErrCoordMaxBuilders
	}
	if stake == 0 {
		return ErrCoordZeroStake
	}

	bc.builders[builderID] = &coordinatorBuilder{
		ID:     builderID,
		Stake:  stake,
		Active: true,
		Rep: &BuilderReputation{
			BuilderID: builderID,
		},
	}
	return nil
}

// UnregisterBuilder marks a builder as inactive.
func (bc *BuilderCoordinator) UnregisterBuilder(builderID string) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	b, exists := bc.builders[builderID]
	if !exists {
		return ErrCoordBuilderNotFound
	}
	b.Active = false
	return nil
}

// SubmitBid submits a bid from a registered builder for a given slot.
// Validates builder existence, deadline, duplicate bids, and min increment.
func (bc *BuilderCoordinator) SubmitBid(builderID string, slot uint64, amount uint64, payloadHash [32]byte) error {
	if slot == 0 {
		return ErrCoordSlotZero
	}

	bc.mu.Lock()
	defer bc.mu.Unlock()

	b, exists := bc.builders[builderID]
	if !exists {
		return ErrCoordBuilderNotFound
	}
	if !b.Active {
		return ErrCoordBuilderNotFound
	}

	now := time.Now()

	// Check deadline: if this slot already has bids, enforce timeout from
	// the first bid's timestamp.
	if deadline, ok := bc.deadlines[slot]; ok {
		if now.After(deadline.Add(bc.config.BidTimeout)) {
			return ErrCoordBidDeadline
		}
	}

	// Check for duplicate bid from same builder for same slot.
	for _, existing := range bc.bids[slot] {
		if existing.BuilderID == builderID {
			// Allow replacement only if increment is sufficient.
			if amount < existing.Amount+bc.config.MinBidIncrement {
				return ErrCoordDuplicateBid
			}
			// Replace the existing bid.
			existing.Amount = amount
			existing.PayloadHash = payloadHash
			existing.SubmitTime = now
			return nil
		}
	}

	// Check min increment against current best bid for new bidders.
	if len(bc.bids[slot]) > 0 {
		best := bc.bestBidForSlot(slot)
		if best != nil && amount > 0 && amount < best.Amount && best.Amount-amount > best.Amount {
			// New bid doesn't need to beat current best; Vickrey auction accepts all.
		}
	}

	bid := &coordinatorBid{
		BuilderID:   builderID,
		Slot:        slot,
		Amount:      amount,
		PayloadHash: payloadHash,
		SubmitTime:  now,
	}
	bc.bids[slot] = append(bc.bids[slot], bid)

	// Record the deadline anchor on first bid.
	if _, ok := bc.deadlines[slot]; !ok {
		bc.deadlines[slot] = now
	}

	b.Rep.LastActive = now
	return nil
}

// SettleAuction executes the Vickrey (second-price sealed-bid) auction for a slot.
// The winner is the highest bidder, but they pay the second-highest price.
func (bc *BuilderCoordinator) SettleAuction(slot uint64) (*AuctionSettlement, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	bids := bc.bids[slot]
	if len(bids) == 0 {
		return nil, ErrCoordNoBids
	}

	// Sort bids descending by amount, then by submit time (earlier wins ties).
	sorted := make([]*coordinatorBid, len(bids))
	copy(sorted, bids)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Amount != sorted[j].Amount {
			return sorted[i].Amount > sorted[j].Amount
		}
		return sorted[i].SubmitTime.Before(sorted[j].SubmitTime)
	})

	winner := sorted[0]
	settlement := &AuctionSettlement{
		Slot:        slot,
		WinnerID:    winner.BuilderID,
		WinnerBid:   winner.Amount,
		SettlePrice: winner.Amount, // default: pay own bid
		TotalBids:   len(bids),
	}

	// Vickrey: if there is a second bidder, winner pays second-highest price.
	if len(sorted) > 1 {
		settlement.SettlePrice = sorted[1].Amount
		settlement.RunnerUpID = sorted[1].BuilderID
	}

	return settlement, nil
}

// RecordDelivery records whether a builder successfully delivered the payload
// after winning an auction. Updates reputation accordingly.
func (bc *BuilderCoordinator) RecordDelivery(builderID string, slot uint64, success bool) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	b, exists := bc.builders[builderID]
	if !exists {
		return
	}

	b.Rep.TotalWins++
	if success {
		b.Rep.SuccessfulDelivery++
	} else {
		b.Rep.FailedDelivery++
	}
}

// BuilderScore returns the reputation score for a builder.
// Returns 0 if the builder is not found.
func (bc *BuilderCoordinator) BuilderScore(builderID string) float64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	b, exists := bc.builders[builderID]
	if !exists {
		return 0
	}
	return b.Rep.Score()
}

// ActiveBuilders returns the IDs of all active builders, sorted alphabetically.
func (bc *BuilderCoordinator) ActiveBuilders() []string {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	var result []string
	for id, b := range bc.builders {
		if b.Active {
			result = append(result, id)
		}
	}
	sort.Strings(result)
	return result
}

// BuilderCount returns the total number of registered builders (active and inactive).
func (bc *BuilderCoordinator) BuilderCount() int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return len(bc.builders)
}

// GetReputation returns the reputation data for a builder.
func (bc *BuilderCoordinator) GetReputation(builderID string) (*BuilderReputation, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	b, exists := bc.builders[builderID]
	if !exists {
		return nil, ErrCoordBuilderNotFound
	}
	// Return a copy.
	rep := *b.Rep
	return &rep, nil
}

// PruneSlot removes all bids and deadline data for a given slot.
func (bc *BuilderCoordinator) PruneSlot(slot uint64) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	delete(bc.bids, slot)
	delete(bc.deadlines, slot)
}

// CleanupInactive removes all builders that are marked inactive.
func (bc *BuilderCoordinator) CleanupInactive() int {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	removed := 0
	for id, b := range bc.builders {
		if !b.Active {
			delete(bc.builders, id)
			removed++
		}
	}
	return removed
}

// bestBidForSlot returns the highest bid for a slot (must hold mu).
func (bc *BuilderCoordinator) bestBidForSlot(slot uint64) *coordinatorBid {
	bids := bc.bids[slot]
	if len(bids) == 0 {
		return nil
	}
	best := bids[0]
	for _, bid := range bids[1:] {
		if bid.Amount > best.Amount {
			best = bid
		}
	}
	return best
}

// MEVBurnCalculator calculates the portion of MEV to burn vs pay to the proposer.
type MEVBurnCalculator struct {
	// BurnRate is the fraction of the winning bid that is burned (0.0 to 1.0).
	BurnRate float64
}

// NewMEVBurnCalculator creates a calculator with the given burn rate.
// The rate is clamped to [0.0, 1.0].
func NewMEVBurnCalculator(burnRate float64) *MEVBurnCalculator {
	if burnRate < 0 {
		burnRate = 0
	}
	if burnRate > 1 {
		burnRate = 1
	}
	return &MEVBurnCalculator{BurnRate: burnRate}
}

// Calculate splits the winning bid into burned and proposer-paid portions.
func (m *MEVBurnCalculator) Calculate(winningBid uint64) (burned, proposerPay uint64) {
	burned = uint64(float64(winningBid) * m.BurnRate)
	proposerPay = winningBid - burned
	return
}

// CalculateWithDetails returns a formatted string describing the split.
func (m *MEVBurnCalculator) CalculateWithDetails(winningBid uint64) string {
	burned, pay := m.Calculate(winningBid)
	return fmt.Sprintf("bid=%d burned=%d proposer=%d rate=%.2f", winningBid, burned, pay, m.BurnRate)
}
