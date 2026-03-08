package builder

import (
	"sync"
	"testing"
)

func TestBuilderReputationTracker_NewDefault(t *testing.T) {
	cfg := DefaultReputationConfig()
	rt := NewBuilderReputationTracker(cfg)
	if rt == nil {
		t.Fatal("expected non-nil tracker")
	}
	if rt.BuilderCount() != 0 {
		t.Fatalf("expected 0 builders, got %d", rt.BuilderCount())
	}
}

func TestBuilderReputationTracker_RecordDelivery(t *testing.T) {
	rt := NewBuilderReputationTracker(DefaultReputationConfig())
	rt.RecordDelivery("builder1", 100, 250)
	rt.RecordDelivery("builder1", 101, 300)

	rec, ok := rt.GetRecord("builder1")
	if !ok {
		t.Fatal("expected record for builder1")
	}
	if rec.Deliveries != 2 {
		t.Fatalf("expected 2 deliveries, got %d", rec.Deliveries)
	}
	if rec.TotalLatencyMs != 550 {
		t.Fatalf("expected 550ms total latency, got %d", rec.TotalLatencyMs)
	}
	if rec.AverageLatencyMs() != 275 {
		t.Fatalf("expected 275ms avg latency, got %d", rec.AverageLatencyMs())
	}
}

func TestBuilderReputationTracker_RecordFailure(t *testing.T) {
	rt := NewBuilderReputationTracker(DefaultReputationConfig())
	rt.RecordFailure("builder1", 100, "timeout")
	rt.RecordFailure("builder1", 101, "invalid")

	rec, ok := rt.GetRecord("builder1")
	if !ok {
		t.Fatal("expected record for builder1")
	}
	if rec.Failures != 2 {
		t.Fatalf("expected 2 failures, got %d", rec.Failures)
	}
	if rec.Deliveries != 0 {
		t.Fatalf("expected 0 deliveries, got %d", rec.Deliveries)
	}
}

func TestBuilderReputationTracker_ScoreWithSuccesses(t *testing.T) {
	cfg := DefaultReputationConfig()
	cfg.MinDeliveries = 2
	rt := NewBuilderReputationTracker(cfg)

	// All successes.
	for i := 0; i < 10; i++ {
		rt.RecordDelivery("builder1", uint64(i), 100)
	}

	score := rt.Score("builder1")
	if score < 0.9 {
		t.Fatalf("expected high score, got %f", score)
	}
}

func TestBuilderReputationTracker_ScoreWithFailures(t *testing.T) {
	cfg := DefaultReputationConfig()
	cfg.MinDeliveries = 2
	rt := NewBuilderReputationTracker(cfg)

	// All failures.
	for i := 0; i < 10; i++ {
		rt.RecordFailure("builder1", uint64(i), "failed")
	}

	score := rt.Score("builder1")
	if score > 0.1 {
		t.Fatalf("expected low score for all failures, got %f", score)
	}
}

func TestBuilderReputationTracker_ScoreMixed(t *testing.T) {
	cfg := DefaultReputationConfig()
	cfg.MinDeliveries = 2
	rt := NewBuilderReputationTracker(cfg)

	// 7 successes, 3 failures -> 70% in window.
	for i := 0; i < 7; i++ {
		rt.RecordDelivery("builder1", uint64(i), 100)
	}
	for i := 7; i < 10; i++ {
		rt.RecordFailure("builder1", uint64(i), "timeout")
	}

	score := rt.Score("builder1")
	// Expected around 0.7.
	if score < 0.6 || score > 0.8 {
		t.Fatalf("expected score around 0.7, got %f", score)
	}
}

func TestBuilderReputationTracker_ScoreUnknownBuilder(t *testing.T) {
	rt := NewBuilderReputationTracker(DefaultReputationConfig())
	score := rt.Score("unknown")
	if score != 0.5 {
		t.Fatalf("expected default score 0.5 for unknown builder, got %f", score)
	}
}

func TestBuilderReputationTracker_GetRecordNotFound(t *testing.T) {
	rt := NewBuilderReputationTracker(DefaultReputationConfig())
	rec, ok := rt.GetRecord("nonexistent")
	if ok || rec != nil {
		t.Fatal("expected nil record for unknown builder")
	}
}

func TestBuilderReputationTracker_TopBuilders(t *testing.T) {
	cfg := DefaultReputationConfig()
	cfg.MinDeliveries = 1
	rt := NewBuilderReputationTracker(cfg)

	// builder_a: all successes.
	for i := 0; i < 10; i++ {
		rt.RecordDelivery("builder_a", uint64(i), 100)
	}
	// builder_b: half and half.
	for i := 0; i < 5; i++ {
		rt.RecordDelivery("builder_b", uint64(i+10), 100)
	}
	for i := 0; i < 5; i++ {
		rt.RecordFailure("builder_b", uint64(i+15), "fail")
	}
	// builder_c: all failures.
	for i := 0; i < 10; i++ {
		rt.RecordFailure("builder_c", uint64(i+20), "fail")
	}

	top := rt.TopBuilders(3)
	if len(top) != 3 {
		t.Fatalf("expected 3 top builders, got %d", len(top))
	}
	if top[0] != "builder_a" {
		t.Fatalf("expected builder_a at top, got %s", top[0])
	}
	if top[2] != "builder_c" {
		t.Fatalf("expected builder_c at bottom, got %s", top[2])
	}
}

func TestBuilderReputationTracker_TopBuildersEmpty(t *testing.T) {
	rt := NewBuilderReputationTracker(DefaultReputationConfig())
	top := rt.TopBuilders(5)
	if len(top) != 0 {
		t.Fatalf("expected empty top builders, got %d", len(top))
	}
}

func TestBuilderReputationTracker_TopBuildersExceedsCount(t *testing.T) {
	cfg := DefaultReputationConfig()
	cfg.MinDeliveries = 1
	rt := NewBuilderReputationTracker(cfg)
	rt.RecordDelivery("only_one", 1, 100)

	top := rt.TopBuilders(10)
	if len(top) != 1 {
		t.Fatalf("expected 1 builder, got %d", len(top))
	}
}

func TestBuilderReputationTracker_SlidingWindow(t *testing.T) {
	cfg := DefaultReputationConfig()
	cfg.WindowSize = 5
	cfg.MinDeliveries = 1
	rt := NewBuilderReputationTracker(cfg)

	// Record 5 failures then 5 successes.
	for i := 0; i < 5; i++ {
		rt.RecordFailure("builder1", uint64(i), "fail")
	}
	for i := 5; i < 10; i++ {
		rt.RecordDelivery("builder1", uint64(i), 100)
	}

	// Window should only see the 5 successes.
	score := rt.Score("builder1")
	if score < 0.9 {
		t.Fatalf("expected high score after window slides, got %f", score)
	}
}

func TestBuilderReputationTracker_DecayScores(t *testing.T) {
	cfg := DefaultReputationConfig()
	cfg.MinDeliveries = 1
	cfg.DecayFactor = 0.5
	rt := NewBuilderReputationTracker(cfg)

	// All successes -> score near 1.0.
	for i := 0; i < 10; i++ {
		rt.RecordDelivery("builder1", uint64(i), 100)
	}
	scoreBefore := rt.Score("builder1")

	rt.DecayScores()
	scoreAfter := rt.Score("builder1")

	// After decay, score should move toward default (0.5).
	if scoreAfter >= scoreBefore {
		t.Fatalf("expected score to decay: before=%f, after=%f", scoreBefore, scoreAfter)
	}
}

func TestSlashingDetector_CheckLateDelivery(t *testing.T) {
	sd := NewSlashingDetector()

	// On time: no event.
	ev := sd.CheckLateDelivery("builder1", 100, 2000, 1500)
	if ev != nil {
		t.Fatal("expected nil for on-time delivery")
	}

	// Late: should produce event.
	ev = sd.CheckLateDelivery("builder1", 100, 2000, 3000)
	if ev == nil {
		t.Fatal("expected slashing event for late delivery")
	}
	if ev.Type != SlashEventLateDelivery {
		t.Fatalf("expected late delivery type, got %v", ev.Type)
	}
	if ev.BuilderID != "builder1" {
		t.Fatalf("expected builder1, got %s", ev.BuilderID)
	}
}

func TestSlashingDetector_CheckEquivocation(t *testing.T) {
	sd := NewSlashingDetector()

	h1 := [32]byte{1, 2, 3}
	h2 := [32]byte{4, 5, 6}

	// Same hash: no event.
	ev := sd.CheckEquivocation("builder1", 100, h1, h1)
	if ev != nil {
		t.Fatal("expected nil for same hashes")
	}

	// Different hashes: equivocation.
	ev = sd.CheckEquivocation("builder1", 100, h1, h2)
	if ev == nil {
		t.Fatal("expected slashing event for equivocation")
	}
	if ev.Type != SlashEventEquivocation {
		t.Fatalf("expected equivocation type, got %v", ev.Type)
	}
}

func TestSlashingDetector_CheckInvalidPayload(t *testing.T) {
	sd := NewSlashingDetector()
	evidence := [32]byte{0xab, 0xcd}
	ev := sd.CheckInvalidPayload("builder1", 200, evidence)
	if ev == nil {
		t.Fatal("expected event for invalid payload")
	}
	if ev.Type != SlashEventInvalidPayload {
		t.Fatalf("expected invalid payload type, got %v", ev.Type)
	}
	if ev.EvidenceHash != evidence {
		t.Fatal("evidence hash mismatch")
	}
}

func TestSlashingDetector_Events(t *testing.T) {
	sd := NewSlashingDetector()
	sd.CheckLateDelivery("b1", 1, 100, 200)
	sd.CheckEquivocation("b2", 2, [32]byte{1}, [32]byte{2})
	sd.CheckInvalidPayload("b3", 3, [32]byte{3})

	events := sd.Events()
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
}

func TestSlashingDetector_EventsForBuilder(t *testing.T) {
	sd := NewSlashingDetector()
	sd.CheckLateDelivery("b1", 1, 100, 200)
	sd.CheckLateDelivery("b1", 2, 100, 200)
	sd.CheckLateDelivery("b2", 3, 100, 200)

	b1Events := sd.EventsForBuilder("b1")
	if len(b1Events) != 2 {
		t.Fatalf("expected 2 events for b1, got %d", len(b1Events))
	}
	b2Events := sd.EventsForBuilder("b2")
	if len(b2Events) != 1 {
		t.Fatalf("expected 1 event for b2, got %d", len(b2Events))
	}
}

func TestBuilderReputationTracker_ApplySlashing(t *testing.T) {
	cfg := DefaultReputationConfig()
	cfg.MinDeliveries = 1
	rt := NewBuilderReputationTracker(cfg)

	// Build up a good score.
	for i := 0; i < 10; i++ {
		rt.RecordDelivery("builder1", uint64(i), 100)
	}
	scoreBefore := rt.Score("builder1")

	// Apply slashing.
	ev := &SlashingEvent{
		BuilderID: "builder1",
		Slot:      100,
		Type:      SlashEventLateDelivery,
	}
	rt.ApplySlashing(ev)

	scoreAfter := rt.Score("builder1")
	if scoreAfter >= scoreBefore {
		t.Fatalf("expected score to decrease after slashing: before=%f, after=%f",
			scoreBefore, scoreAfter)
	}

	rec, _ := rt.GetRecord("builder1")
	if rec.SlashingCount != 1 {
		t.Fatalf("expected 1 slashing, got %d", rec.SlashingCount)
	}
}

func TestBuilderReputationTracker_ApplySlashingNil(t *testing.T) {
	rt := NewBuilderReputationTracker(DefaultReputationConfig())
	// Should not panic.
	rt.ApplySlashing(nil)
	if rt.BuilderCount() != 0 {
		t.Fatal("expected no builders after nil slashing")
	}
}

func TestBuilderReputationTracker_ConcurrentAccess(t *testing.T) {
	rt := NewBuilderReputationTracker(DefaultReputationConfig())
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(3)
		id := "builder"
		slot := uint64(i)
		go func() {
			defer wg.Done()
			rt.RecordDelivery(id, slot, 100)
		}()
		go func() {
			defer wg.Done()
			rt.RecordFailure(id, slot+1000, "fail")
		}()
		go func() {
			defer wg.Done()
			_ = rt.Score(id)
		}()
	}
	wg.Wait()

	// Should not have panicked.
	if rt.BuilderCount() == 0 {
		t.Fatal("expected at least one builder after concurrent access")
	}
}

func TestSlashingDetector_ConcurrentAccess(t *testing.T) {
	sd := NewSlashingDetector()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(slot uint64) {
			defer wg.Done()
			sd.CheckLateDelivery("b1", slot, 100, 200)
		}(uint64(i))
		go func() {
			defer wg.Done()
			_ = sd.Events()
		}()
	}
	wg.Wait()

	if sd.EventCount() != 50 {
		t.Fatalf("expected 50 events, got %d", sd.EventCount())
	}
}

func TestReputationRecord_SuccessRate(t *testing.T) {
	rec := &ReputationRecord{Deliveries: 7, Failures: 3}
	rate := rec.SuccessRate()
	if rate < 0.69 || rate > 0.71 {
		t.Fatalf("expected ~0.7 success rate, got %f", rate)
	}

	// Zero events.
	empty := &ReputationRecord{}
	if empty.SuccessRate() != 1.0 {
		t.Fatalf("expected 1.0 for empty record, got %f", empty.SuccessRate())
	}
}

func TestReputationRecord_AverageLatencyMs(t *testing.T) {
	rec := &ReputationRecord{Deliveries: 4, TotalLatencyMs: 1000}
	if rec.AverageLatencyMs() != 250 {
		t.Fatalf("expected 250ms avg, got %d", rec.AverageLatencyMs())
	}

	// Zero deliveries.
	empty := &ReputationRecord{}
	if empty.AverageLatencyMs() != 0 {
		t.Fatal("expected 0 for no deliveries")
	}
}

func TestBuilderReputationTracker_NegativeLatencyClamped(t *testing.T) {
	rt := NewBuilderReputationTracker(DefaultReputationConfig())
	rt.RecordDelivery("builder1", 1, -500)

	rec, ok := rt.GetRecord("builder1")
	if !ok {
		t.Fatal("expected record")
	}
	// Negative latency should be clamped to 0.
	if rec.TotalLatencyMs != 0 {
		t.Fatalf("expected 0 latency, got %d", rec.TotalLatencyMs)
	}
}

func TestSlashingEventType_String(t *testing.T) {
	tests := []struct {
		typ  SlashingEventType
		want string
	}{
		{SlashEventLateDelivery, "late_delivery"},
		{SlashEventInvalidPayload, "invalid_payload"},
		{SlashEventEquivocation, "equivocation"},
		{SlashingEventType(99), "unknown"},
	}
	for _, tc := range tests {
		if tc.typ.String() != tc.want {
			t.Fatalf("expected %q, got %q", tc.want, tc.typ.String())
		}
	}
}

func TestBuilderReputationTracker_ConfigDefaults(t *testing.T) {
	// Zero/invalid config should be corrected.
	cfg := ReputationConfig{}
	rt := NewBuilderReputationTracker(cfg)
	if rt.config.WindowSize != 100 {
		t.Fatalf("expected default window size 100, got %d", rt.config.WindowSize)
	}
	if rt.config.MinDeliveries != 5 {
		t.Fatalf("expected default min deliveries 5, got %d", rt.config.MinDeliveries)
	}
}

func TestBuilderReputationTracker_MinDeliveriesBlending(t *testing.T) {
	cfg := DefaultReputationConfig()
	cfg.MinDeliveries = 10
	rt := NewBuilderReputationTracker(cfg)

	// Only 2 deliveries (below min) -> should blend with default.
	rt.RecordDelivery("builder1", 1, 100)
	rt.RecordDelivery("builder1", 2, 100)

	score := rt.Score("builder1")
	// With 2/10 deliveries, weight=0.2, observedRate=1.0.
	// Blended = 0.2 * 1.0 + 0.8 * 0.5 = 0.6.
	if score < 0.55 || score > 0.65 {
		t.Fatalf("expected blended score around 0.6, got %f", score)
	}
}
