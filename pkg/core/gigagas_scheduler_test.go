package core

import (
	"testing"
)

// --- GigagasScheduler tests ---

func TestGigagasSchedulerDefaults(t *testing.T) {
	gs := NewGigagasScheduler(GigagasSchedulerConfig{})
	if gs.MaxLanes() != 16 {
		t.Fatalf("expected default MaxLanes=16, got %d", gs.MaxLanes())
	}
	if gs.BatchSize() != 256 {
		t.Fatalf("expected default BatchSize=256, got %d", gs.BatchSize())
	}
	if gs.ConflictRetryLimit() != 3 {
		t.Fatalf("expected default ConflictRetryLimit=3, got %d", gs.ConflictRetryLimit())
	}
}

func TestGigagasSchedulerCustomConfig(t *testing.T) {
	cfg := GigagasSchedulerConfig{MaxLanes: 8, BatchSize: 64, ConflictRetryLimit: 5}
	gs := NewGigagasScheduler(cfg)
	if gs.MaxLanes() != 8 {
		t.Fatalf("expected MaxLanes=8, got %d", gs.MaxLanes())
	}
	if gs.BatchSize() != 64 {
		t.Fatalf("expected BatchSize=64, got %d", gs.BatchSize())
	}
	if gs.ConflictRetryLimit() != 5 {
		t.Fatalf("expected ConflictRetryLimit=5, got %d", gs.ConflictRetryLimit())
	}
}

func TestGigagasSchedulerScheduleEmpty(t *testing.T) {
	gs := NewGigagasScheduler(DefaultGigagasSchedulerConfig())
	batches := gs.ScheduleTransactions(0, nil, nil)
	if batches != nil {
		t.Fatalf("expected nil for empty input, got %v", batches)
	}
}

func TestGigagasSchedulerNoConflicts(t *testing.T) {
	gs := NewGigagasScheduler(DefaultGigagasSchedulerConfig())
	// 3 txs with disjoint write sets: all go into one batch.
	reads := []map[string]bool{
		{"a": true},
		{"b": true},
		{"c": true},
	}
	writes := []map[string]bool{
		{"x": true},
		{"y": true},
		{"z": true},
	}
	batches := gs.ScheduleTransactions(3, reads, writes)
	if len(batches) != 1 {
		t.Fatalf("expected 1 batch for non-conflicting txs, got %d", len(batches))
	}
	if len(batches[0]) != 3 {
		t.Fatalf("expected 3 txs in batch, got %d", len(batches[0]))
	}
}

func TestGigagasSchedulerAllConflicting(t *testing.T) {
	gs := NewGigagasScheduler(DefaultGigagasSchedulerConfig())
	// 3 txs all writing to the same key: each goes in its own batch.
	writes := []map[string]bool{
		{"shared": true},
		{"shared": true},
		{"shared": true},
	}
	batches := gs.ScheduleTransactions(3, nil, writes)
	if len(batches) != 3 {
		t.Fatalf("expected 3 batches for all-conflicting txs, got %d", len(batches))
	}
}

func TestGigagasSchedulerReadWriteConflict(t *testing.T) {
	gs := NewGigagasScheduler(DefaultGigagasSchedulerConfig())
	// Tx 0 writes "a", tx 1 reads "a" => conflict.
	reads := []map[string]bool{
		nil,
		{"a": true},
	}
	writes := []map[string]bool{
		{"a": true},
		nil,
	}
	batches := gs.ScheduleTransactions(2, reads, writes)
	if len(batches) != 2 {
		t.Fatalf("expected 2 batches for read-write conflict, got %d", len(batches))
	}
}

func TestGigagasSchedulerPartialConflicts(t *testing.T) {
	gs := NewGigagasScheduler(DefaultGigagasSchedulerConfig())
	// Tx 0 and tx 1 conflict (both write "x"), tx 2 is independent.
	writes := []map[string]bool{
		{"x": true},
		{"x": true},
		{"y": true},
	}
	batches := gs.ScheduleTransactions(3, nil, writes)
	// Tx 0 and tx 2 can be in one batch, tx 1 in another.
	if len(batches) != 2 {
		t.Fatalf("expected 2 batches, got %d", len(batches))
	}
	totalTxs := 0
	for _, b := range batches {
		totalTxs += len(b)
	}
	if totalTxs != 3 {
		t.Fatalf("expected 3 total txs across batches, got %d", totalTxs)
	}
}

// --- DetectConflicts tests ---

func TestDetectConflictsNoOverlap(t *testing.T) {
	gs := NewGigagasScheduler(DefaultGigagasSchedulerConfig())
	a := NewWorkUnit(0, 21000)
	a.AddWrite("x")
	b := NewWorkUnit(1, 21000)
	b.AddWrite("y")
	if gs.DetectConflicts(a, b) {
		t.Fatal("expected no conflict for disjoint write sets")
	}
}

func TestDetectConflictsWriteWrite(t *testing.T) {
	gs := NewGigagasScheduler(DefaultGigagasSchedulerConfig())
	a := NewWorkUnit(0, 21000)
	a.AddWrite("x")
	b := NewWorkUnit(1, 21000)
	b.AddWrite("x")
	if !gs.DetectConflicts(a, b) {
		t.Fatal("expected conflict for overlapping write sets")
	}
}

func TestDetectConflictsReadWrite(t *testing.T) {
	gs := NewGigagasScheduler(DefaultGigagasSchedulerConfig())
	a := NewWorkUnit(0, 21000)
	a.AddRead("x")
	b := NewWorkUnit(1, 21000)
	b.AddWrite("x")
	if !gs.DetectConflicts(a, b) {
		t.Fatal("expected conflict: a reads what b writes")
	}
}

func TestDetectConflictsReadRead(t *testing.T) {
	gs := NewGigagasScheduler(DefaultGigagasSchedulerConfig())
	a := NewWorkUnit(0, 21000)
	a.AddRead("x")
	b := NewWorkUnit(1, 21000)
	b.AddRead("x")
	if gs.DetectConflicts(a, b) {
		t.Fatal("expected no conflict for read-read")
	}
}

// --- WorkUnit tests ---

func TestWorkUnitCreation(t *testing.T) {
	wu := NewWorkUnit(5, 50000)
	if wu.Index != 5 {
		t.Fatalf("expected index 5, got %d", wu.Index)
	}
	if wu.GasEstimate != 50000 {
		t.Fatalf("expected gas 50000, got %d", wu.GasEstimate)
	}
	if len(wu.ReadSet) != 0 || len(wu.WriteSet) != 0 {
		t.Fatal("expected empty read/write sets")
	}
}

func TestWorkUnitAddReadWrite(t *testing.T) {
	wu := NewWorkUnit(0, 21000)
	wu.AddRead("balance_0xABC")
	wu.AddWrite("nonce_0xABC")
	if !wu.ReadSet["balance_0xABC"] {
		t.Fatal("expected read set to contain balance_0xABC")
	}
	if !wu.WriteSet["nonce_0xABC"] {
		t.Fatal("expected write set to contain nonce_0xABC")
	}
}

func TestWorkUnitRetryCount(t *testing.T) {
	wu := NewWorkUnit(0, 21000)
	if wu.RetryCount != 0 {
		t.Fatalf("expected initial retry count 0, got %d", wu.RetryCount)
	}
	wu.RetryCount++
	if wu.RetryCount != 1 {
		t.Fatalf("expected retry count 1, got %d", wu.RetryCount)
	}
}

// --- ConflictResolver tests ---

func TestConflictResolverEmpty(t *testing.T) {
	cr := NewConflictResolver()
	pairs := cr.FindConflictingPairs()
	if len(pairs) != 0 {
		t.Fatalf("expected no conflicts for empty resolver, got %d", len(pairs))
	}
	graph := cr.BuildDependencyGraph()
	if len(graph) != 0 {
		t.Fatalf("expected empty dependency graph, got %d entries", len(graph))
	}
}

func TestConflictResolverAddUnit(t *testing.T) {
	cr := NewConflictResolver()
	cr.AddUnit(NewWorkUnit(0, 21000))
	cr.AddUnit(NewWorkUnit(1, 21000))
	if cr.UnitCount() != 2 {
		t.Fatalf("expected 2 units, got %d", cr.UnitCount())
	}
}

func TestConflictResolverFindPairs(t *testing.T) {
	cr := NewConflictResolver()
	u0 := NewWorkUnit(0, 21000)
	u0.AddWrite("x")
	u1 := NewWorkUnit(1, 21000)
	u1.AddWrite("x")
	u2 := NewWorkUnit(2, 21000)
	u2.AddWrite("y")

	cr.AddUnit(u0)
	cr.AddUnit(u1)
	cr.AddUnit(u2)

	pairs := cr.FindConflictingPairs()
	if len(pairs) != 1 {
		t.Fatalf("expected 1 conflicting pair, got %d", len(pairs))
	}
	if pairs[0] != [2]int{0, 1} {
		t.Fatalf("expected pair [0,1], got %v", pairs[0])
	}
}

func TestConflictResolverDependencyGraph(t *testing.T) {
	cr := NewConflictResolver()
	u0 := NewWorkUnit(0, 21000)
	u0.AddWrite("x")
	u1 := NewWorkUnit(1, 21000)
	u1.AddRead("x")
	u2 := NewWorkUnit(2, 21000)
	u2.AddRead("y") // no conflict

	cr.AddUnit(u0)
	cr.AddUnit(u1)
	cr.AddUnit(u2)

	graph := cr.BuildDependencyGraph()
	// u1 depends on u0 (u0 writes x, u1 reads x).
	deps, ok := graph[1]
	if !ok || len(deps) != 1 || deps[0] != 0 {
		t.Fatalf("expected u1 -> [u0], got %v", graph)
	}
	// u2 has no deps.
	if _, ok := graph[2]; ok {
		t.Fatal("expected u2 to have no dependencies")
	}
}

func TestConflictResolverConflictCount(t *testing.T) {
	cr := NewConflictResolver()
	u0 := NewWorkUnit(0, 21000)
	u0.AddWrite("a")
	u0.AddWrite("b")
	u1 := NewWorkUnit(1, 21000)
	u1.AddWrite("a")
	u2 := NewWorkUnit(2, 21000)
	u2.AddWrite("b")

	cr.AddUnit(u0)
	cr.AddUnit(u1)
	cr.AddUnit(u2)

	// u0 conflicts with u1 (a), u0 conflicts with u2 (b).
	if cr.ConflictCount() != 2 {
		t.Fatalf("expected 2 conflicts, got %d", cr.ConflictCount())
	}
}

// --- ExecutionLane tests ---

func TestExecutionLaneCreation(t *testing.T) {
	lane := NewExecutionLane(3)
	if lane.ID != 3 {
		t.Fatalf("expected lane ID 3, got %d", lane.ID)
	}
	if lane.UnitCount() != 0 {
		t.Fatalf("expected 0 units, got %d", lane.UnitCount())
	}
	if lane.Load() != 0 {
		t.Fatalf("expected 0 load, got %d", lane.Load())
	}
}

func TestExecutionLaneAssign(t *testing.T) {
	lane := NewExecutionLane(0)
	lane.Assign(0, 50000)
	lane.Assign(1, 30000)
	if lane.UnitCount() != 2 {
		t.Fatalf("expected 2 units, got %d", lane.UnitCount())
	}
	if lane.Load() != 80000 {
		t.Fatalf("expected load 80000, got %d", lane.Load())
	}
}

func TestGigagasSchedulerAssignToLanes(t *testing.T) {
	gs := NewGigagasScheduler(GigagasSchedulerConfig{MaxLanes: 3})
	units := []*WorkUnit{
		NewWorkUnit(0, 100000),
		NewWorkUnit(1, 50000),
		NewWorkUnit(2, 80000),
		NewWorkUnit(3, 30000),
		NewWorkUnit(4, 60000),
	}
	lanes := gs.AssignToLanes(units)
	if len(lanes) != 3 {
		t.Fatalf("expected 3 lanes, got %d", len(lanes))
	}
	totalUnits := 0
	for _, lane := range lanes {
		totalUnits += lane.UnitCount()
	}
	if totalUnits != 5 {
		t.Fatalf("expected 5 total units across lanes, got %d", totalUnits)
	}
}

func TestGigagasSchedulerAssignToLanesFewerUnits(t *testing.T) {
	gs := NewGigagasScheduler(GigagasSchedulerConfig{MaxLanes: 8})
	units := []*WorkUnit{
		NewWorkUnit(0, 100000),
		NewWorkUnit(1, 50000),
	}
	lanes := gs.AssignToLanes(units)
	// Should only create 2 lanes, not 8.
	if len(lanes) != 2 {
		t.Fatalf("expected 2 lanes (fewer units than max), got %d", len(lanes))
	}
}

func TestGigagasSchedulerAssignToLanesEmpty(t *testing.T) {
	gs := NewGigagasScheduler(DefaultGigagasSchedulerConfig())
	lanes := gs.AssignToLanes(nil)
	if lanes != nil {
		t.Fatalf("expected nil for empty units, got %v", lanes)
	}
}

func TestBuildExecutionPlanDeterminism(t *testing.T) {
	gs := NewGigagasScheduler(DefaultGigagasSchedulerConfig())
	units := make([]*WorkUnit, 5)
	for i := range units {
		units[i] = NewWorkUnit(i, 21000)
		units[i].AddWrite("unique_" + string(rune('a'+i)))
	}
	plan1 := gs.BuildExecutionPlan(units)
	plan2 := gs.BuildExecutionPlan(units)

	if len(plan1) != len(plan2) {
		t.Fatalf("plans differ in batch count: %d vs %d", len(plan1), len(plan2))
	}
	for i := range plan1 {
		if len(plan1[i]) != len(plan2[i]) {
			t.Fatalf("batch %d differs in size: %d vs %d", i, len(plan1[i]), len(plan2[i]))
		}
	}
}

func TestSchedulerConfigDefault(t *testing.T) {
	cfg := DefaultGigagasSchedulerConfig()
	if cfg.MaxLanes != 16 {
		t.Fatalf("expected MaxLanes=16, got %d", cfg.MaxLanes)
	}
	if cfg.BatchSize != 256 {
		t.Fatalf("expected BatchSize=256, got %d", cfg.BatchSize)
	}
	if cfg.ConflictRetryLimit != 3 {
		t.Fatalf("expected ConflictRetryLimit=3, got %d", cfg.ConflictRetryLimit)
	}
}

func TestGigagasSchedulerLargeNonConflicting(t *testing.T) {
	gs := NewGigagasScheduler(DefaultGigagasSchedulerConfig())
	n := 100
	units := make([]*WorkUnit, n)
	for i := 0; i < n; i++ {
		units[i] = NewWorkUnit(i, 21000)
		// Each unit writes to a unique key.
		units[i].AddWrite("key_" + string(rune(i)))
	}
	plan := gs.BuildExecutionPlan(units)
	// All should fit in a single batch since no conflicts.
	if len(plan) != 1 {
		t.Fatalf("expected 1 batch for %d non-conflicting units, got %d", n, len(plan))
	}
	if len(plan[0]) != n {
		t.Fatalf("expected %d units in single batch, got %d", n, len(plan[0]))
	}
}

func TestConflictResolverChainDependencies(t *testing.T) {
	cr := NewConflictResolver()
	// Chain: u0 writes x, u1 reads x and writes y, u2 reads y.
	u0 := NewWorkUnit(0, 21000)
	u0.AddWrite("x")
	u1 := NewWorkUnit(1, 21000)
	u1.AddRead("x")
	u1.AddWrite("y")
	u2 := NewWorkUnit(2, 21000)
	u2.AddRead("y")

	cr.AddUnit(u0)
	cr.AddUnit(u1)
	cr.AddUnit(u2)

	graph := cr.BuildDependencyGraph()
	// u1 depends on u0, u2 depends on u1.
	if len(graph[1]) != 1 || graph[1][0] != 0 {
		t.Fatalf("expected u1 depends on u0, got %v", graph[1])
	}
	if len(graph[2]) != 1 || graph[2][0] != 1 {
		t.Fatalf("expected u2 depends on u1, got %v", graph[2])
	}
}
