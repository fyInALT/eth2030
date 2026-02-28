package das

import (
	"math"
	"sort"
	"sync"
	"testing"
)

func defaultTestCellRepConfig() CellReputationConfig {
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

func TestCellPeerScorerRecordDeliveryIncreasesScore(t *testing.T) {
	cfg := defaultTestCellRepConfig()
	scorer, err := NewCellPeerScorer(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scorer.RecordDelivery("peer-a", 0, 100)
	score := scorer.PeerScore("peer-a")
	expected := cfg.BaselineScore + cfg.DeliveryReward
	if math.Abs(score-expected) > 0.001 {
		t.Errorf("score = %f, want %f", score, expected)
	}

	// Second delivery should increase further.
	scorer.RecordDelivery("peer-a", 1, 50)
	score2 := scorer.PeerScore("peer-a")
	if score2 <= score {
		t.Errorf("score should increase after second delivery: %f <= %f", score2, score)
	}
}

func TestCellPeerScorerRecordInvalidDecreasesScore(t *testing.T) {
	cfg := defaultTestCellRepConfig()
	scorer, err := NewCellPeerScorer(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scorer.RecordInvalid("peer-b", 5)
	score := scorer.PeerScore("peer-b")
	expected := cfg.BaselineScore + cfg.InvalidPenalty
	if math.Abs(score-expected) > 0.001 {
		t.Errorf("score = %f, want %f", score, expected)
	}
}

func TestCellPeerScorerRecordDuplicateMinorPenalty(t *testing.T) {
	cfg := defaultTestCellRepConfig()
	scorer, err := NewCellPeerScorer(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scorer.RecordDuplicate("peer-c", 3)
	score := scorer.PeerScore("peer-c")
	expected := cfg.BaselineScore + cfg.DuplicatePenalty
	if math.Abs(score-expected) > 0.001 {
		t.Errorf("score = %f, want %f", score, expected)
	}

	// Duplicate penalty should be smaller than invalid penalty.
	scorer.RecordInvalid("peer-d", 3)
	scoreInvalid := scorer.PeerScore("peer-d")
	if score <= scoreInvalid {
		t.Error("duplicate penalty should be less severe than invalid")
	}
}

func TestCellPeerScorerBaselineForUnknownPeer(t *testing.T) {
	cfg := defaultTestCellRepConfig()
	scorer, err := NewCellPeerScorer(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	score := scorer.PeerScore("unknown-peer")
	if math.Abs(score-cfg.BaselineScore) > 0.001 {
		t.Errorf("unknown peer score = %f, want baseline %f", score, cfg.BaselineScore)
	}
}

func TestCellPeerScorerBanThreshold(t *testing.T) {
	cfg := defaultTestCellRepConfig()
	cfg.BanThreshold = 0.0
	cfg.BaselineScore = 10.0
	cfg.InvalidPenalty = -15.0
	scorer, err := NewCellPeerScorer(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After 1 invalid: 10 - 15 = -5, below ban threshold of 0.
	scorer.RecordInvalid("bad-peer", 0)

	banned := scorer.BannedPeers()
	if len(banned) != 1 || banned[0] != "bad-peer" {
		t.Errorf("banned = %v, want [bad-peer]", banned)
	}

	if !scorer.IsBanned("bad-peer") {
		t.Error("expected bad-peer to be banned")
	}
}

func TestCellPeerScorerDecayReducesScores(t *testing.T) {
	cfg := defaultTestCellRepConfig()
	cfg.DecayRate = 0.5
	scorer, err := NewCellPeerScorer(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Give peer a high score.
	for i := 0; i < 10; i++ {
		scorer.RecordDelivery("peer-e", uint64(i), 50)
	}
	scoreBefore := scorer.PeerScore("peer-e")
	if scoreBefore <= cfg.BaselineScore {
		t.Fatalf("expected score above baseline, got %f", scoreBefore)
	}

	// Apply decay -- score should move toward baseline.
	scorer.DecayScores()
	scoreAfter := scorer.PeerScore("peer-e")

	// After decay with rate 0.5: new = baseline + (old - baseline) * 0.5
	expectedAfter := cfg.BaselineScore + (scoreBefore-cfg.BaselineScore)*0.5
	if math.Abs(scoreAfter-expectedAfter) > 0.01 {
		t.Errorf("score after decay = %f, want ~%f", scoreAfter, expectedAfter)
	}
}

func TestCellPeerScorerSamplingWeightsSorted(t *testing.T) {
	cfg := defaultTestCellRepConfig()
	scorer, err := NewCellPeerScorer(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Create peers with different scores.
	scorer.RecordDelivery("peer-high", 0, 50)
	scorer.RecordDelivery("peer-high", 1, 50)
	scorer.RecordDelivery("peer-high", 2, 50)

	scorer.RecordDelivery("peer-mid", 0, 50)

	scorer.RecordInvalid("peer-low", 0)

	weights := scorer.SamplingWeights()
	if len(weights) == 0 {
		t.Fatal("expected non-empty weights")
	}

	// Verify sorted by score descending.
	for i := 1; i < len(weights); i++ {
		if weights[i].Score > weights[i-1].Score {
			t.Errorf("weights not sorted: index %d score %f > index %d score %f",
				i, weights[i].Score, i-1, weights[i-1].Score)
		}
	}

	// All weights should be non-negative.
	for _, w := range weights {
		if w.Weight < 0 {
			t.Errorf("negative weight for %s: %f", w.PeerID, w.Weight)
		}
	}
}

func TestCellPeerScorerUnknownPeerDefaultStats(t *testing.T) {
	cfg := defaultTestCellRepConfig()
	scorer, err := NewCellPeerScorer(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stats, found := scorer.PeerStats("nonexistent")
	if found {
		t.Error("expected not found for unknown peer")
	}
	if stats.CellsDelivered != 0 || stats.CellsInvalid != 0 {
		t.Error("expected zero stats for unknown peer")
	}
}

func TestCellPeerScorerHighLatencyPenalized(t *testing.T) {
	cfg := defaultTestCellRepConfig()
	scorer, err := NewCellPeerScorer(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Low latency peer.
	scorer.RecordDelivery("fast-peer", 0, 50)
	fastScore := scorer.PeerScore("fast-peer")

	// High latency peer (well above threshold of 200ms).
	scorer.RecordDelivery("slow-peer", 0, 1000)
	slowScore := scorer.PeerScore("slow-peer")

	if slowScore >= fastScore {
		t.Errorf("slow peer score (%f) should be less than fast peer (%f)", slowScore, fastScore)
	}
}

func TestCellPeerScorerConcurrentRecording(t *testing.T) {
	cfg := defaultTestCellRepConfig()
	scorer, err := NewCellPeerScorer(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(3)
		go func(idx int) {
			defer wg.Done()
			scorer.RecordDelivery("concurrent-peer", uint64(idx), 100)
		}(i)
		go func(idx int) {
			defer wg.Done()
			scorer.RecordInvalid("concurrent-peer", uint64(idx))
		}(i)
		go func(idx int) {
			defer wg.Done()
			scorer.RecordDuplicate("concurrent-peer", uint64(idx))
		}(i)
	}
	wg.Wait()

	stats, found := scorer.PeerStats("concurrent-peer")
	if !found {
		t.Fatal("expected peer to be tracked")
	}
	if stats.CellsDelivered != 50 {
		t.Errorf("delivered = %d, want 50", stats.CellsDelivered)
	}
	if stats.CellsInvalid != 50 {
		t.Errorf("invalid = %d, want 50", stats.CellsInvalid)
	}
	if stats.CellsDuplicate != 50 {
		t.Errorf("duplicate = %d, want 50", stats.CellsDuplicate)
	}
}

func TestCellPeerScorerManyPeers(t *testing.T) {
	cfg := defaultTestCellRepConfig()
	scorer, err := NewCellPeerScorer(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < 100; i++ {
		peerID := string(rune('A'+i/26)) + string(rune('a'+i%26))
		for j := 0; j < (i + 1); j++ {
			scorer.RecordDelivery(peerID, uint64(j), int64(50+i))
		}
	}

	if scorer.PeerCount() != 100 {
		t.Errorf("peer count = %d, want 100", scorer.PeerCount())
	}

	weights := scorer.SamplingWeights()
	if len(weights) != 100 {
		t.Errorf("weights count = %d, want 100", len(weights))
	}
}

func TestCellPeerScorerRecoveryAfterDecay(t *testing.T) {
	cfg := defaultTestCellRepConfig()
	cfg.DecayRate = 0.5
	cfg.BanThreshold = 0.0
	cfg.BaselineScore = 50.0
	cfg.InvalidPenalty = -40.0
	scorer, err := NewCellPeerScorer(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Drive score below ban threshold.
	scorer.RecordInvalid("recoverable", 0)
	scorer.RecordInvalid("recoverable", 1)
	if !scorer.IsBanned("recoverable") {
		t.Fatal("expected peer to be banned after 2 invalids")
	}

	// Apply decay multiple times -- score should move toward baseline.
	for i := 0; i < 20; i++ {
		scorer.DecayScores()
	}

	// After enough decay, score should be close to baseline and no longer banned.
	score := scorer.PeerScore("recoverable")
	if math.Abs(score-cfg.BaselineScore) > 1.0 {
		t.Errorf("after decay, score = %f, want near baseline %f", score, cfg.BaselineScore)
	}
	if scorer.IsBanned("recoverable") {
		t.Error("expected peer to be unbanned after sufficient decay")
	}
}

func TestCellPeerScorerConfigValidation(t *testing.T) {
	// Zero decay rate.
	cfg := defaultTestCellRepConfig()
	cfg.DecayRate = 0
	_, err := NewCellPeerScorer(cfg)
	if err != ErrCellScorerZeroDecay {
		t.Errorf("expected ErrCellScorerZeroDecay, got: %v", err)
	}

	// Negative decay rate.
	cfg.DecayRate = -0.5
	_, err = NewCellPeerScorer(cfg)
	if err != ErrCellScorerZeroDecay {
		t.Errorf("expected ErrCellScorerZeroDecay, got: %v", err)
	}

	// Decay rate > 1.
	cfg.DecayRate = 1.5
	_, err = NewCellPeerScorer(cfg)
	if err != ErrCellScorerZeroDecay {
		t.Errorf("expected ErrCellScorerZeroDecay, got: %v", err)
	}

	// Valid config.
	cfg.DecayRate = 0.95
	_, err = NewCellPeerScorer(cfg)
	if err != nil {
		t.Errorf("expected nil error for valid config, got: %v", err)
	}
}

func TestCellPeerScorerEmptyState(t *testing.T) {
	cfg := defaultTestCellRepConfig()
	scorer, err := NewCellPeerScorer(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if scorer.PeerCount() != 0 {
		t.Errorf("empty scorer peer count = %d, want 0", scorer.PeerCount())
	}

	weights := scorer.SamplingWeights()
	if len(weights) != 0 {
		t.Errorf("empty scorer weights = %d, want 0", len(weights))
	}

	banned := scorer.BannedPeers()
	if len(banned) != 0 {
		t.Errorf("empty scorer banned = %d, want 0", len(banned))
	}

	// Decay on empty scorer should not panic.
	scorer.DecayScores()
}

func TestCellPeerScorerStatsAccuracy(t *testing.T) {
	cfg := defaultTestCellRepConfig()
	scorer, err := NewCellPeerScorer(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scorer.RecordDelivery("stats-peer", 0, 100)
	scorer.RecordDelivery("stats-peer", 1, 200)
	scorer.RecordInvalid("stats-peer", 2)
	scorer.RecordDuplicate("stats-peer", 3)

	stats, found := scorer.PeerStats("stats-peer")
	if !found {
		t.Fatal("expected peer to be found")
	}
	if stats.CellsDelivered != 2 {
		t.Errorf("delivered = %d, want 2", stats.CellsDelivered)
	}
	if stats.CellsInvalid != 1 {
		t.Errorf("invalid = %d, want 1", stats.CellsInvalid)
	}
	if stats.CellsDuplicate != 1 {
		t.Errorf("duplicate = %d, want 1", stats.CellsDuplicate)
	}
	if stats.TotalLatencyMs != 300 {
		t.Errorf("total latency = %d, want 300", stats.TotalLatencyMs)
	}

	avgLatency := stats.AverageLatencyMs()
	// 300ms / 4 events = 75ms
	if avgLatency != 75 {
		t.Errorf("avg latency = %d, want 75", avgLatency)
	}

	expectedScore := cfg.BaselineScore + 2*cfg.DeliveryReward + cfg.InvalidPenalty + cfg.DuplicatePenalty
	if math.Abs(stats.Score-expectedScore) > 0.01 {
		t.Errorf("score = %f, want %f", stats.Score, expectedScore)
	}
}

func TestCellPeerScorerBannedPeerExcludedFromWeights(t *testing.T) {
	cfg := defaultTestCellRepConfig()
	cfg.BanThreshold = 0.0
	cfg.BaselineScore = 10.0
	cfg.InvalidPenalty = -15.0
	scorer, err := NewCellPeerScorer(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scorer.RecordDelivery("good-peer", 0, 50)
	scorer.RecordInvalid("bad-peer", 0) // score = 10 - 15 = -5, banned

	weights := scorer.SamplingWeights()
	for _, w := range weights {
		if w.PeerID == "bad-peer" {
			t.Error("banned peer should not appear in sampling weights")
		}
	}
	found := false
	for _, w := range weights {
		if w.PeerID == "good-peer" {
			found = true
		}
	}
	if !found {
		t.Error("good peer should appear in sampling weights")
	}
}

func TestCellPeerScorerBannedPeersSorted(t *testing.T) {
	cfg := defaultTestCellRepConfig()
	cfg.BanThreshold = 0.0
	cfg.BaselineScore = 10.0
	cfg.InvalidPenalty = -15.0
	scorer, err := NewCellPeerScorer(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scorer.RecordInvalid("zebra-peer", 0)
	scorer.RecordInvalid("alpha-peer", 0)
	scorer.RecordInvalid("mid-peer", 0)

	banned := scorer.BannedPeers()
	if len(banned) != 3 {
		t.Fatalf("banned count = %d, want 3", len(banned))
	}
	if !sort.StringsAreSorted(banned) {
		t.Errorf("banned peers not sorted: %v", banned)
	}
}
