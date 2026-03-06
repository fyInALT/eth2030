package util

import (
	"sync"
	"testing"
	"time"
)

func TestBuilderCoordinatorRegisterAndBid(t *testing.T) {
	bc := NewBuilderCoordinator(DefaultCoordinatorConfig())

	if err := bc.RegisterBuilder("builder-1", 1000); err != nil {
		t.Fatalf("register: %v", err)
	}

	var ph [32]byte
	ph[0] = 0xAB
	if err := bc.SubmitBid("builder-1", 1, 500, ph); err != nil {
		t.Fatalf("submit bid: %v", err)
	}

	result, err := bc.SettleAuction(1)
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
	if result.WinnerID != "builder-1" {
		t.Errorf("expected winner builder-1, got %s", result.WinnerID)
	}
	if result.SettlePrice != 500 {
		t.Errorf("single bidder should pay own bid: got %d", result.SettlePrice)
	}
}

func TestBuilderCoordinatorVickreyPricing(t *testing.T) {
	bc := NewBuilderCoordinator(DefaultCoordinatorConfig())

	bc.RegisterBuilder("alice", 1000)
	bc.RegisterBuilder("bob", 1000)
	bc.RegisterBuilder("carol", 1000)

	var ph [32]byte
	bc.SubmitBid("alice", 10, 300, ph)
	bc.SubmitBid("bob", 10, 500, ph)
	bc.SubmitBid("carol", 10, 400, ph)

	result, err := bc.SettleAuction(10)
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
	if result.WinnerID != "bob" {
		t.Errorf("expected winner bob, got %s", result.WinnerID)
	}
	// Vickrey: winner pays second-highest price.
	if result.SettlePrice != 400 {
		t.Errorf("expected settle price 400 (second-highest), got %d", result.SettlePrice)
	}
	if result.RunnerUpID != "carol" {
		t.Errorf("expected runner-up carol, got %s", result.RunnerUpID)
	}
	if result.TotalBids != 3 {
		t.Errorf("expected 3 total bids, got %d", result.TotalBids)
	}
}

func TestBuilderCoordinatorSingleBidderPaysOwnBid(t *testing.T) {
	bc := NewBuilderCoordinator(DefaultCoordinatorConfig())
	bc.RegisterBuilder("solo", 500)

	var ph [32]byte
	bc.SubmitBid("solo", 5, 1000, ph)

	result, err := bc.SettleAuction(5)
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
	if result.SettlePrice != 1000 {
		t.Errorf("single bidder should pay own bid: expected 1000, got %d", result.SettlePrice)
	}
	if result.RunnerUpID != "" {
		t.Errorf("expected empty runner-up, got %s", result.RunnerUpID)
	}
}

func TestBuilderCoordinatorNoBidsError(t *testing.T) {
	bc := NewBuilderCoordinator(DefaultCoordinatorConfig())

	_, err := bc.SettleAuction(99)
	if err != ErrCoordNoBids {
		t.Errorf("expected ErrCoordNoBids, got %v", err)
	}
}

func TestBuilderCoordinatorBidAfterDeadline(t *testing.T) {
	cfg := DefaultCoordinatorConfig()
	cfg.BidTimeout = 50 * time.Millisecond
	bc := NewBuilderCoordinator(cfg)

	bc.RegisterBuilder("fast", 1000)
	bc.RegisterBuilder("slow", 1000)

	var ph [32]byte
	bc.SubmitBid("fast", 1, 100, ph)

	// Wait for deadline to expire.
	time.Sleep(100 * time.Millisecond)

	err := bc.SubmitBid("slow", 1, 200, ph)
	if err != ErrCoordBidDeadline {
		t.Errorf("expected ErrCoordBidDeadline, got %v", err)
	}
}

func TestBuilderCoordinatorDuplicateBidSameSlot(t *testing.T) {
	bc := NewBuilderCoordinator(DefaultCoordinatorConfig())
	bc.RegisterBuilder("dup", 1000)

	var ph [32]byte
	bc.SubmitBid("dup", 1, 100, ph)

	// Same builder, same slot, insufficient increment -> duplicate error.
	err := bc.SubmitBid("dup", 1, 100, ph)
	if err != ErrCoordDuplicateBid {
		t.Errorf("expected ErrCoordDuplicateBid, got %v", err)
	}

	// With sufficient increment, replacement should succeed.
	err = bc.SubmitBid("dup", 1, 200, ph)
	if err != nil {
		t.Errorf("expected replacement bid to succeed, got %v", err)
	}
}

func TestBuilderCoordinatorReputationIncreases(t *testing.T) {
	bc := NewBuilderCoordinator(DefaultCoordinatorConfig())
	bc.RegisterBuilder("reliable", 1000)

	// Default score for builder with no wins.
	score := bc.BuilderScore("reliable")
	if score != 0.5 {
		t.Errorf("expected default score 0.5, got %f", score)
	}

	// Record successful deliveries.
	bc.RecordDelivery("reliable", 1, true)
	bc.RecordDelivery("reliable", 2, true)
	bc.RecordDelivery("reliable", 3, true)

	score = bc.BuilderScore("reliable")
	if score != 1.0 {
		t.Errorf("expected perfect score 1.0, got %f", score)
	}
}

func TestBuilderCoordinatorReputationDecreases(t *testing.T) {
	bc := NewBuilderCoordinator(DefaultCoordinatorConfig())
	bc.RegisterBuilder("flaky", 1000)

	bc.RecordDelivery("flaky", 1, true)
	bc.RecordDelivery("flaky", 2, false)
	bc.RecordDelivery("flaky", 3, true)
	bc.RecordDelivery("flaky", 4, false)

	score := bc.BuilderScore("flaky")
	// 2 successes / 4 wins = 0.5
	if score != 0.5 {
		t.Errorf("expected score 0.5, got %f", score)
	}
}

func TestBuilderCoordinatorLowReputationDeprioritized(t *testing.T) {
	bc := NewBuilderCoordinator(DefaultCoordinatorConfig())
	bc.RegisterBuilder("good", 1000)
	bc.RegisterBuilder("bad", 1000)

	// Good builder: all successes.
	bc.RecordDelivery("good", 1, true)
	bc.RecordDelivery("good", 2, true)

	// Bad builder: all failures.
	bc.RecordDelivery("bad", 1, false)
	bc.RecordDelivery("bad", 2, false)

	goodScore := bc.BuilderScore("good")
	badScore := bc.BuilderScore("bad")

	if badScore >= goodScore {
		t.Errorf("bad builder score (%f) should be less than good builder score (%f)",
			badScore, goodScore)
	}
}

func TestMEVBurnCalculator50Percent(t *testing.T) {
	calc := NewMEVBurnCalculator(0.5)
	burned, pay := calc.Calculate(1000)
	if burned != 500 {
		t.Errorf("expected 500 burned, got %d", burned)
	}
	if pay != 500 {
		t.Errorf("expected 500 proposer pay, got %d", pay)
	}
}

func TestMEVBurnCalculatorZeroRate(t *testing.T) {
	calc := NewMEVBurnCalculator(0.0)
	burned, pay := calc.Calculate(1000)
	if burned != 0 {
		t.Errorf("expected 0 burned, got %d", burned)
	}
	if pay != 1000 {
		t.Errorf("expected 1000 proposer pay, got %d", pay)
	}
}

func TestMEVBurnCalculator100PercentRate(t *testing.T) {
	calc := NewMEVBurnCalculator(1.0)
	burned, pay := calc.Calculate(1000)
	if burned != 1000 {
		t.Errorf("expected 1000 burned, got %d", burned)
	}
	if pay != 0 {
		t.Errorf("expected 0 proposer pay, got %d", pay)
	}
}

func TestBuilderCoordinatorConcurrentBids(t *testing.T) {
	bc := NewBuilderCoordinator(DefaultCoordinatorConfig())
	for i := 0; i < 20; i++ {
		bc.RegisterBuilder(builderName(i), 1000)
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			var ph [32]byte
			ph[0] = byte(idx)
			bc.SubmitBid(builderName(idx), 42, uint64(100+idx), ph)
		}(i)
	}
	wg.Wait()

	result, err := bc.SettleAuction(42)
	if err != nil {
		t.Fatalf("settle: %v", err)
	}
	if result.TotalBids != 20 {
		t.Errorf("expected 20 bids, got %d", result.TotalBids)
	}
}

func TestBuilderCoordinatorMultipleSlots(t *testing.T) {
	bc := NewBuilderCoordinator(DefaultCoordinatorConfig())
	bc.RegisterBuilder("a", 1000)
	bc.RegisterBuilder("b", 1000)

	var ph [32]byte
	bc.SubmitBid("a", 1, 100, ph)
	bc.SubmitBid("b", 1, 200, ph)
	bc.SubmitBid("a", 2, 300, ph)
	bc.SubmitBid("b", 2, 150, ph)

	r1, _ := bc.SettleAuction(1)
	r2, _ := bc.SettleAuction(2)

	if r1.WinnerID != "b" {
		t.Errorf("slot 1 winner: expected b, got %s", r1.WinnerID)
	}
	if r2.WinnerID != "a" {
		t.Errorf("slot 2 winner: expected a, got %s", r2.WinnerID)
	}
}

func TestBuilderCoordinatorDeregistrationCleanup(t *testing.T) {
	bc := NewBuilderCoordinator(DefaultCoordinatorConfig())
	bc.RegisterBuilder("keep", 1000)
	bc.RegisterBuilder("remove", 1000)

	bc.UnregisterBuilder("remove")
	active := bc.ActiveBuilders()
	if len(active) != 1 || active[0] != "keep" {
		t.Errorf("expected only keep active, got %v", active)
	}

	removed := bc.CleanupInactive()
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}
	if bc.BuilderCount() != 1 {
		t.Errorf("expected 1 builder remaining, got %d", bc.BuilderCount())
	}
}

func TestBuilderCoordinatorZeroStakeBuilder(t *testing.T) {
	bc := NewBuilderCoordinator(DefaultCoordinatorConfig())
	err := bc.RegisterBuilder("zero-stake", 0)
	if err != ErrCoordZeroStake {
		t.Errorf("expected ErrCoordZeroStake, got %v", err)
	}
}

func TestBuilderCoordinatorUnknownBuilderBid(t *testing.T) {
	bc := NewBuilderCoordinator(DefaultCoordinatorConfig())
	var ph [32]byte
	err := bc.SubmitBid("unknown", 1, 100, ph)
	if err != ErrCoordBuilderNotFound {
		t.Errorf("expected ErrCoordBuilderNotFound, got %v", err)
	}
}

func TestBuilderCoordinatorSlotZero(t *testing.T) {
	bc := NewBuilderCoordinator(DefaultCoordinatorConfig())
	bc.RegisterBuilder("x", 100)
	var ph [32]byte
	err := bc.SubmitBid("x", 0, 100, ph)
	if err != ErrCoordSlotZero {
		t.Errorf("expected ErrCoordSlotZero, got %v", err)
	}
}

func TestMEVBurnCalculatorClamp(t *testing.T) {
	// Negative rate should be clamped to 0.
	calc := NewMEVBurnCalculator(-0.5)
	if calc.BurnRate != 0 {
		t.Errorf("expected clamped rate 0, got %f", calc.BurnRate)
	}

	// Rate > 1 should be clamped to 1.
	calc2 := NewMEVBurnCalculator(1.5)
	if calc2.BurnRate != 1 {
		t.Errorf("expected clamped rate 1, got %f", calc2.BurnRate)
	}
}

func TestMEVBurnCalculatorDetails(t *testing.T) {
	calc := NewMEVBurnCalculator(0.25)
	details := calc.CalculateWithDetails(1000)
	if details == "" {
		t.Error("expected non-empty details string")
	}
	burned, pay := calc.Calculate(1000)
	if burned != 250 {
		t.Errorf("expected 250 burned, got %d", burned)
	}
	if pay != 750 {
		t.Errorf("expected 750 proposer pay, got %d", pay)
	}
}

func TestBuilderCoordinatorPruneSlot(t *testing.T) {
	bc := NewBuilderCoordinator(DefaultCoordinatorConfig())
	bc.RegisterBuilder("p", 1000)
	var ph [32]byte
	bc.SubmitBid("p", 5, 100, ph)

	// Before prune, auction should settle.
	_, err := bc.SettleAuction(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bc.PruneSlot(5)

	// After prune, no bids.
	_, err = bc.SettleAuction(5)
	if err != ErrCoordNoBids {
		t.Errorf("expected ErrCoordNoBids after prune, got %v", err)
	}
}

func TestBuilderCoordinatorDuplicateRegister(t *testing.T) {
	bc := NewBuilderCoordinator(DefaultCoordinatorConfig())
	bc.RegisterBuilder("dup", 1000)
	err := bc.RegisterBuilder("dup", 2000)
	if err != ErrCoordBuilderExists {
		t.Errorf("expected ErrCoordBuilderExists, got %v", err)
	}
}

func TestBuilderCoordinatorMaxBuilders(t *testing.T) {
	cfg := DefaultCoordinatorConfig()
	cfg.MaxBuilders = 2
	bc := NewBuilderCoordinator(cfg)

	bc.RegisterBuilder("a", 100)
	bc.RegisterBuilder("b", 100)
	err := bc.RegisterBuilder("c", 100)
	if err != ErrCoordMaxBuilders {
		t.Errorf("expected ErrCoordMaxBuilders, got %v", err)
	}
}

func TestBuilderCoordinatorGetReputation(t *testing.T) {
	bc := NewBuilderCoordinator(DefaultCoordinatorConfig())
	bc.RegisterBuilder("rep-test", 1000)

	bc.RecordDelivery("rep-test", 1, true)
	bc.RecordDelivery("rep-test", 2, false)

	rep, err := bc.GetReputation("rep-test")
	if err != nil {
		t.Fatalf("get reputation: %v", err)
	}
	if rep.TotalWins != 2 {
		t.Errorf("expected 2 total wins, got %d", rep.TotalWins)
	}
	if rep.SuccessfulDelivery != 1 {
		t.Errorf("expected 1 successful, got %d", rep.SuccessfulDelivery)
	}
	if rep.FailedDelivery != 1 {
		t.Errorf("expected 1 failed, got %d", rep.FailedDelivery)
	}

	_, err = bc.GetReputation("nonexistent")
	if err != ErrCoordBuilderNotFound {
		t.Errorf("expected ErrCoordBuilderNotFound, got %v", err)
	}
}

// builderName returns a deterministic builder name for index i.
func builderName(i int) string {
	return "builder-" + string(rune('A'+i))
}
