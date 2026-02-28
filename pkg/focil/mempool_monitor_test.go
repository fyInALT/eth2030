package focil

import (
	"sync"
	"testing"
)

func TestMempoolMonitor_NewMonitor(t *testing.T) {
	mm := NewMempoolMonitor()
	if mm == nil {
		t.Fatal("expected non-nil monitor")
	}
	if mm.PendingCount() != 0 {
		t.Fatalf("expected 0 pending, got %d", mm.PendingCount())
	}
}

func TestMempoolMonitor_AddPendingTx(t *testing.T) {
	mm := NewMempoolMonitor()
	hash := [32]byte{1, 2, 3}
	sender := [20]byte{0xaa}

	mm.AddPendingTx(hash, sender, 1000, 100)
	if mm.PendingCount() != 1 {
		t.Fatalf("expected 1 pending, got %d", mm.PendingCount())
	}

	// Duplicate is a no-op.
	mm.AddPendingTx(hash, sender, 2000, 200)
	if mm.PendingCount() != 1 {
		t.Fatalf("expected 1 pending after dup, got %d", mm.PendingCount())
	}
}

func TestMempoolMonitor_RecordInclusion(t *testing.T) {
	mm := NewMempoolMonitor()
	hash := [32]byte{1}
	sender := [20]byte{0xaa}

	mm.AddPendingTx(hash, sender, 1000, 100)
	mm.RecordInclusion(hash, 10)

	metrics := mm.Metrics()
	if metrics.IncludedCount != 1 {
		t.Fatalf("expected 1 included, got %d", metrics.IncludedCount)
	}
}

func TestMempoolMonitor_RecordExclusion(t *testing.T) {
	mm := NewMempoolMonitor()
	hash := [32]byte{2}
	sender := [20]byte{0xbb}

	mm.AddPendingTx(hash, sender, 500, 50)
	mm.RecordExclusion(hash, 20)

	metrics := mm.Metrics()
	if metrics.ExcludedCount != 1 {
		t.Fatalf("expected 1 excluded, got %d", metrics.ExcludedCount)
	}
}

func TestMempoolMonitor_RecordInclusionUnknownTx(t *testing.T) {
	mm := NewMempoolMonitor()
	hash := [32]byte{99}
	mm.RecordInclusion(hash, 5)

	// Should still be tracked.
	if mm.PendingCount() != 1 {
		t.Fatalf("expected 1 pending, got %d", mm.PendingCount())
	}
	metrics := mm.Metrics()
	if metrics.IncludedCount != 1 {
		t.Fatalf("expected 1 included, got %d", metrics.IncludedCount)
	}
}

func TestMempoolMonitor_ComplianceRate(t *testing.T) {
	mm := NewMempoolMonitor()

	// No data -> 1.0.
	rate := mm.ComplianceRate()
	if rate != 1.0 {
		t.Fatalf("expected 1.0 for empty, got %f", rate)
	}

	// 3 included, 1 excluded -> 0.75.
	for i := 0; i < 3; i++ {
		h := [32]byte{byte(i)}
		mm.AddPendingTx(h, [20]byte{}, 100, 1)
		mm.RecordInclusion(h, 1)
	}
	h := [32]byte{99}
	mm.AddPendingTx(h, [20]byte{}, 100, 1)
	mm.RecordExclusion(h, 1)

	rate = mm.ComplianceRate()
	if rate < 0.74 || rate > 0.76 {
		t.Fatalf("expected ~0.75 compliance rate, got %f", rate)
	}
}

func TestMempoolMonitor_Metrics(t *testing.T) {
	mm := NewMempoolMonitor()

	mm.AddPendingTx([32]byte{1}, [20]byte{}, 100, 10)
	mm.AddPendingTx([32]byte{2}, [20]byte{}, 200, 20)
	mm.AddPendingTx([32]byte{3}, [20]byte{}, 300, 30)

	mm.RecordInclusion([32]byte{1}, 1)
	mm.RecordExclusion([32]byte{2}, 2)

	m := mm.Metrics()
	if m.TotalPending != 3 {
		t.Fatalf("expected 3 total, got %d", m.TotalPending)
	}
	if m.IncludedCount != 1 {
		t.Fatalf("expected 1 included, got %d", m.IncludedCount)
	}
	if m.ExcludedCount != 1 {
		t.Fatalf("expected 1 excluded, got %d", m.ExcludedCount)
	}
	if m.UnresolvedCount != 1 {
		t.Fatalf("expected 1 unresolved, got %d", m.UnresolvedCount)
	}
}

func TestMempoolMonitor_PruneOlderThan(t *testing.T) {
	mm := NewMempoolMonitor()

	mm.AddPendingTx([32]byte{1}, [20]byte{}, 100, 10)
	mm.AddPendingTx([32]byte{2}, [20]byte{}, 200, 20)
	mm.AddPendingTx([32]byte{3}, [20]byte{}, 300, 30)

	pruned := mm.PruneOlderThan(25)
	if pruned != 2 {
		t.Fatalf("expected 2 pruned, got %d", pruned)
	}
	if mm.PendingCount() != 1 {
		t.Fatalf("expected 1 remaining, got %d", mm.PendingCount())
	}
}

func TestMempoolMonitor_PruneNone(t *testing.T) {
	mm := NewMempoolMonitor()
	mm.AddPendingTx([32]byte{1}, [20]byte{}, 100, 50)

	pruned := mm.PruneOlderThan(10)
	if pruned != 0 {
		t.Fatalf("expected 0 pruned, got %d", pruned)
	}
}

func TestFairnessAnalyzer_NewAnalyzer(t *testing.T) {
	fa := NewFairnessAnalyzer()
	if fa == nil {
		t.Fatal("expected non-nil analyzer")
	}
	if fa.SenderCount() != 0 {
		t.Fatalf("expected 0 senders, got %d", fa.SenderCount())
	}
}

func TestFairnessAnalyzer_PerfectlyFair(t *testing.T) {
	fa := NewFairnessAnalyzer()

	// All senders have the same inclusion rate.
	for i := 0; i < 5; i++ {
		sender := [20]byte{byte(i)}
		for j := 0; j < 10; j++ {
			fa.RecordSenderInclusion(sender)
		}
	}

	score := fa.FairnessScore()
	if score < 0.99 {
		t.Fatalf("expected ~1.0 for perfectly fair, got %f", score)
	}
}

func TestFairnessAnalyzer_Unfair(t *testing.T) {
	fa := NewFairnessAnalyzer()

	// Sender A: always included.
	senderA := [20]byte{0xaa}
	for i := 0; i < 10; i++ {
		fa.RecordSenderInclusion(senderA)
	}

	// Sender B: never included.
	senderB := [20]byte{0xbb}
	for i := 0; i < 10; i++ {
		fa.RecordSenderExclusion(senderB)
	}

	score := fa.FairnessScore()
	if score > 0.6 {
		t.Fatalf("expected low fairness score for unfair distribution, got %f", score)
	}
}

func TestFairnessAnalyzer_EmptyScore(t *testing.T) {
	fa := NewFairnessAnalyzer()
	if fa.FairnessScore() != 1.0 {
		t.Fatalf("expected 1.0 for empty, got %f", fa.FairnessScore())
	}
}

func TestFairnessAnalyzer_SuspectedCensored(t *testing.T) {
	fa := NewFairnessAnalyzer()

	// Sender A: 1 inclusion, 9 exclusions (10% rate < 25%).
	senderA := [20]byte{0xaa}
	fa.RecordSenderInclusion(senderA)
	for i := 0; i < 9; i++ {
		fa.RecordSenderExclusion(senderA)
	}

	// Sender B: 8 inclusions, 2 exclusions (80% rate).
	senderB := [20]byte{0xbb}
	for i := 0; i < 8; i++ {
		fa.RecordSenderInclusion(senderB)
	}
	for i := 0; i < 2; i++ {
		fa.RecordSenderExclusion(senderB)
	}

	suspected := fa.SuspectedCensored()
	if len(suspected) != 1 {
		t.Fatalf("expected 1 suspected censored, got %d", len(suspected))
	}
	if suspected[0] != senderA {
		t.Fatalf("expected sender A as suspected, got %v", suspected[0])
	}
}

func TestFairnessAnalyzer_SuspectedCensoredMinEvents(t *testing.T) {
	fa := NewFairnessAnalyzer()

	// Only 2 events: below the 3-event minimum.
	sender := [20]byte{0xcc}
	fa.RecordSenderExclusion(sender)
	fa.RecordSenderExclusion(sender)

	suspected := fa.SuspectedCensored()
	if len(suspected) != 0 {
		t.Fatalf("expected 0 suspected (below min events), got %d", len(suspected))
	}
}

func TestCensorshipIndicator_NewIndicator(t *testing.T) {
	ci := NewCensorshipIndicator(0.7)
	if ci == nil {
		t.Fatal("expected non-nil indicator")
	}
	if ci.SenderCount() != 0 {
		t.Fatalf("expected 0 senders, got %d", ci.SenderCount())
	}
}

func TestCensorshipIndicator_NotCensored(t *testing.T) {
	ci := NewCensorshipIndicator(0.7)
	sender := [20]byte{0xaa}

	// 8 inclusions, 2 exclusions (20% exclusion rate < 70%).
	for i := 0; i < 8; i++ {
		ci.RecordOutcome(sender, true)
	}
	for i := 0; i < 2; i++ {
		ci.RecordOutcome(sender, false)
	}

	if ci.IsCensored(sender) {
		t.Fatal("expected not censored")
	}
}

func TestCensorshipIndicator_IsCensored(t *testing.T) {
	ci := NewCensorshipIndicator(0.7)
	sender := [20]byte{0xbb}

	// 1 inclusion, 9 exclusions (90% exclusion rate >= 70%).
	ci.RecordOutcome(sender, true)
	for i := 0; i < 9; i++ {
		ci.RecordOutcome(sender, false)
	}

	if !ci.IsCensored(sender) {
		t.Fatal("expected censored")
	}
}

func TestCensorshipIndicator_MinimumOutcomes(t *testing.T) {
	ci := NewCensorshipIndicator(0.5)
	sender := [20]byte{0xcc}

	// Only 1 event: below minimum of 2.
	ci.RecordOutcome(sender, false)
	if ci.IsCensored(sender) {
		t.Fatal("expected not censored with only 1 event")
	}
}

func TestCensorshipIndicator_UnknownSender(t *testing.T) {
	ci := NewCensorshipIndicator(0.7)
	unknown := [20]byte{0xff}
	if ci.IsCensored(unknown) {
		t.Fatal("expected not censored for unknown sender")
	}
}

func TestCensorshipIndicator_CensoredSenders(t *testing.T) {
	ci := NewCensorshipIndicator(0.5)

	// Sender A: censored (3/4 excluded = 75% >= 50%).
	senderA := [20]byte{0x01}
	ci.RecordOutcome(senderA, true)
	ci.RecordOutcome(senderA, false)
	ci.RecordOutcome(senderA, false)
	ci.RecordOutcome(senderA, false)

	// Sender B: not censored (1/4 excluded = 25% < 50%).
	senderB := [20]byte{0x02}
	ci.RecordOutcome(senderB, true)
	ci.RecordOutcome(senderB, true)
	ci.RecordOutcome(senderB, true)
	ci.RecordOutcome(senderB, false)

	censored := ci.CensoredSenders()
	if len(censored) != 1 {
		t.Fatalf("expected 1 censored sender, got %d", len(censored))
	}
	if censored[0] != senderA {
		t.Fatalf("expected sender A, got %v", censored[0])
	}
}

func TestCensorshipIndicator_DefaultThreshold(t *testing.T) {
	// Invalid threshold should default to 0.7.
	ci := NewCensorshipIndicator(0)
	sender := [20]byte{0xaa}

	// 3 inclusions, 7 exclusions (70% exclusion rate = 0.7 threshold).
	for i := 0; i < 3; i++ {
		ci.RecordOutcome(sender, true)
	}
	for i := 0; i < 7; i++ {
		ci.RecordOutcome(sender, false)
	}

	if !ci.IsCensored(sender) {
		t.Fatal("expected censored at default 0.7 threshold")
	}
}

func TestMempoolMonitor_ConcurrentAccess(t *testing.T) {
	mm := NewMempoolMonitor()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(3)
		hash := [32]byte{byte(i)}
		go func() {
			defer wg.Done()
			mm.AddPendingTx(hash, [20]byte{}, 100, 1)
		}()
		go func() {
			defer wg.Done()
			mm.RecordInclusion(hash, 1)
		}()
		go func() {
			defer wg.Done()
			_ = mm.Metrics()
		}()
	}
	wg.Wait()

	// Should not panic.
	if mm.PendingCount() == 0 {
		t.Fatal("expected some pending txs")
	}
}

func TestFairnessAnalyzer_ConcurrentAccess(t *testing.T) {
	fa := NewFairnessAnalyzer()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(2)
		sender := [20]byte{byte(i)}
		go func() {
			defer wg.Done()
			fa.RecordSenderInclusion(sender)
		}()
		go func() {
			defer wg.Done()
			_ = fa.FairnessScore()
		}()
	}
	wg.Wait()

	if fa.SenderCount() == 0 {
		t.Fatal("expected some senders")
	}
}

func TestCensorshipIndicator_ConcurrentAccess(t *testing.T) {
	ci := NewCensorshipIndicator(0.7)
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(2)
		sender := [20]byte{byte(i)}
		go func() {
			defer wg.Done()
			ci.RecordOutcome(sender, i%2 == 0)
		}()
		go func() {
			defer wg.Done()
			_ = ci.IsCensored(sender)
		}()
	}
	wg.Wait()

	if ci.SenderCount() == 0 {
		t.Fatal("expected some senders")
	}
}

func TestMempoolMonitor_RecordExclusionUnknownTx(t *testing.T) {
	mm := NewMempoolMonitor()
	hash := [32]byte{77}
	mm.RecordExclusion(hash, 3)

	if mm.PendingCount() != 1 {
		t.Fatalf("expected 1 pending, got %d", mm.PendingCount())
	}
	metrics := mm.Metrics()
	if metrics.ExcludedCount != 1 {
		t.Fatalf("expected 1 excluded, got %d", metrics.ExcludedCount)
	}
}

func TestFairnessAnalyzer_SingleSender(t *testing.T) {
	fa := NewFairnessAnalyzer()
	sender := [20]byte{0x01}
	fa.RecordSenderInclusion(sender)
	fa.RecordSenderExclusion(sender)

	// Single sender -> perfectly fair (Gini=0 for single element).
	score := fa.FairnessScore()
	if score != 1.0 {
		t.Fatalf("expected 1.0 for single sender, got %f", score)
	}
}

func TestComplianceMetrics_AllUnresolved(t *testing.T) {
	mm := NewMempoolMonitor()
	mm.AddPendingTx([32]byte{1}, [20]byte{}, 100, 1)
	mm.AddPendingTx([32]byte{2}, [20]byte{}, 200, 2)

	m := mm.Metrics()
	if m.ComplianceRate != 1.0 {
		t.Fatalf("expected 1.0 for all unresolved, got %f", m.ComplianceRate)
	}
	if m.UnresolvedCount != 2 {
		t.Fatalf("expected 2 unresolved, got %d", m.UnresolvedCount)
	}
}
