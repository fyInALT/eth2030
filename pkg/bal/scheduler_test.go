package bal

import (
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// helpers

func makeSchedulerBAL(txCount int) *BlockAccessList {
	b := NewBlockAccessList()
	addr := types.HexToAddress("0x01")
	for i := 1; i <= txCount; i++ {
		b.AddEntry(AccessEntry{
			Address:     addr,
			AccessIndex: uint64(i),
			StorageReads: []StorageAccess{
				{Slot: types.Hash{byte(i)}},
			},
		})
	}
	return b
}

// makeConflictingBAL creates a BAL where every tx writes the same slot
// (all pairs conflict, forcing sequential waves).
func makeConflictingBAL(txCount int) *BlockAccessList {
	b := NewBlockAccessList()
	addr := types.HexToAddress("0x01")
	slot := types.HexToHash("0x01")
	for i := 1; i <= txCount; i++ {
		b.AddEntry(AccessEntry{
			Address:     addr,
			AccessIndex: uint64(i),
			StorageChanges: []StorageChange{
				{Slot: slot},
			},
		})
	}
	return b
}

// makeIndependentBAL creates a BAL where every tx accesses unique slots
// (no conflicts, all txs in one wave).
func makeIndependentBAL(txCount int) *BlockAccessList {
	b := NewBlockAccessList()
	for i := 1; i <= txCount; i++ {
		addr := types.Address{byte(i)}
		b.AddEntry(AccessEntry{
			Address:     addr,
			AccessIndex: uint64(i),
			StorageReads: []StorageAccess{
				{Slot: types.Hash{byte(i)}},
			},
		})
	}
	return b
}

// --- NewBALScheduler ---

func TestNewBALSchedulerValid(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, err := NewBALScheduler(4, d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
}

func TestNewBALSchedulerZeroWorkers(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	_, err := NewBALScheduler(0, d)
	if err != ErrWorkerCountInvalid {
		t.Errorf("expected ErrWorkerCountInvalid, got %v", err)
	}
}

func TestNewBALSchedulerNegativeWorkers(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	_, err := NewBALScheduler(-1, d)
	if err != ErrWorkerCountInvalid {
		t.Errorf("expected ErrWorkerCountInvalid, got %v", err)
	}
}

// --- Workers ---

func TestWorkers(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(8, d)
	if s.Workers() != 8 {
		t.Errorf("Workers() = %d, want 8", s.Workers())
	}
}

// --- SchedulerMetricsSnapshot ---

func TestSchedulerMetricsSnapshotInitial(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(2, d)
	m := s.SchedulerMetricsSnapshot()
	if m == nil {
		t.Fatal("SchedulerMetricsSnapshot returned nil")
	}
	if m.WavesFormed.Load() != 0 {
		t.Errorf("WavesFormed = %d, want 0", m.WavesFormed.Load())
	}
	if m.TxsScheduled.Load() != 0 {
		t.Errorf("TxsScheduled = %d, want 0", m.TxsScheduled.Load())
	}
}

func TestSchedulerMetricsSnapshotAfterSchedule(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(2, d)

	b := makeIndependentBAL(3)
	_, err := s.Schedule(b)
	if err != nil {
		t.Fatalf("Schedule failed: %v", err)
	}

	m := s.SchedulerMetricsSnapshot()
	if m.WavesFormed.Load() == 0 {
		t.Error("WavesFormed should be > 0 after Schedule")
	}
	if m.TxsScheduled.Load() == 0 {
		t.Error("TxsScheduled should be > 0 after Schedule")
	}
}

// --- Schedule ---

func TestScheduleNilBAL(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(2, d)
	_, err := s.Schedule(nil)
	if err != ErrNoTransactions {
		t.Errorf("expected ErrNoTransactions, got %v", err)
	}
}

func TestScheduleEmptyBAL(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(2, d)
	_, err := s.Schedule(NewBlockAccessList())
	if err != ErrNoTransactions {
		t.Errorf("expected ErrNoTransactions, got %v", err)
	}
}

func TestScheduleOnlyPreExecution(t *testing.T) {
	// AccessIndex=0 entries are skipped by the conflict detector.
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(2, d)
	b := NewBlockAccessList()
	b.AddEntry(AccessEntry{
		Address:     types.HexToAddress("0x01"),
		AccessIndex: 0,
	})
	_, err := s.Schedule(b)
	if err != ErrNoTransactions {
		t.Errorf("expected ErrNoTransactions for pre-execution-only BAL, got %v", err)
	}
}

func TestScheduleSingleTx(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(2, d)
	b := makeIndependentBAL(1)
	waves, err := s.Schedule(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(waves) != 1 {
		t.Errorf("expected 1 wave, got %d", len(waves))
	}
	if len(waves[0].Tasks) != 1 {
		t.Errorf("expected 1 task in wave, got %d", len(waves[0].Tasks))
	}
}

func TestScheduleIndependentTxsOneWave(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(4, d)
	// 4 txs, each accessing unique slots → no conflicts → single wave.
	b := makeIndependentBAL(4)
	waves, err := s.Schedule(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(waves) != 1 {
		t.Errorf("expected 1 wave for independent txs, got %d", len(waves))
	}
	if len(waves[0].Tasks) != 4 {
		t.Errorf("expected 4 tasks in wave, got %d", len(waves[0].Tasks))
	}
}

func TestScheduleChainedConflicts(t *testing.T) {
	// tx1 writes slot, tx2 writes slot, tx3 writes slot → chain of deps → 3 waves.
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(2, d)
	b := makeConflictingBAL(3)
	waves, err := s.Schedule(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(waves) < 2 {
		t.Errorf("expected at least 2 waves for chained conflicts, got %d", len(waves))
	}
}

func TestScheduleMetricsUpdated(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(2, d)
	b := makeIndependentBAL(3)
	_, err := s.Schedule(b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.metrics.WavesFormed.Load() == 0 {
		t.Error("WavesFormed not updated")
	}
	if s.metrics.TxsScheduled.Load() != 3 {
		t.Errorf("TxsScheduled = %d, want 3", s.metrics.TxsScheduled.Load())
	}
	if s.metrics.MaxWaveSize.Load() == 0 {
		t.Error("MaxWaveSize not updated")
	}
}

// --- AssignWorkers ---

func TestAssignWorkersRoundRobin(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(3, d)

	wave := Wave{
		Tasks: []TxTask{
			{Index: 0, GasLimit: 21000},
			{Index: 1, GasLimit: 21000},
			{Index: 2, GasLimit: 21000},
			{Index: 3, GasLimit: 21000},
		},
	}
	assignments := s.AssignWorkers(wave)
	if len(assignments) != 4 {
		t.Fatalf("expected 4 assignments, got %d", len(assignments))
	}
	// Round-robin: task 0→worker 0, 1→1, 2→2, 3→0.
	expected := []int{0, 1, 2, 0}
	for i, a := range assignments {
		if a.WorkerID != expected[i] {
			t.Errorf("assignment[%d].WorkerID = %d, want %d", i, a.WorkerID, expected[i])
		}
	}
}

func TestAssignWorkersEmptyWave(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(2, d)
	assignments := s.AssignWorkers(Wave{})
	if len(assignments) != 0 {
		t.Errorf("expected 0 assignments for empty wave, got %d", len(assignments))
	}
}

func TestAssignWorkersTxIndexPreserved(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(2, d)
	wave := Wave{
		Tasks: []TxTask{
			{Index: 10, GasLimit: 21000},
			{Index: 20, GasLimit: 21000},
		},
	}
	assignments := s.AssignWorkers(wave)
	if assignments[0].TxIndex != 10 {
		t.Errorf("TxIndex = %d, want 10", assignments[0].TxIndex)
	}
	if assignments[1].TxIndex != 20 {
		t.Errorf("TxIndex = %d, want 20", assignments[1].TxIndex)
	}
}

// --- ExecuteSpeculative ---

func TestExecuteSpeculativeNoConflicts(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(2, d)

	wave := Wave{
		Tasks: []TxTask{
			{Index: 0, GasLimit: 100_000},
			{Index: 1, GasLimit: 100_000},
		},
	}
	results := s.ExecuteSpeculative(wave, map[int]struct{}{})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Rolled {
			t.Errorf("tx %d rolled back unexpectedly", r.TxIndex)
		}
		if !r.Success {
			t.Errorf("tx %d should succeed without conflict", r.TxIndex)
		}
	}
}

func TestExecuteSpeculativeWithConflict(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(2, d)

	wave := Wave{
		Tasks: []TxTask{
			{Index: 0, GasLimit: 100_000},
			{Index: 1, GasLimit: 100_000},
		},
	}
	conflictSet := map[int]struct{}{1: {}}
	results := s.ExecuteSpeculative(wave, conflictSet)

	var tx0, tx1 *SpeculativeResult
	for i := range results {
		if results[i].TxIndex == 0 {
			tx0 = &results[i]
		}
		if results[i].TxIndex == 1 {
			tx1 = &results[i]
		}
	}
	if tx0 == nil || tx1 == nil {
		t.Fatal("missing results for tx0 or tx1")
	}
	if tx0.Rolled {
		t.Error("tx0 should not be rolled back")
	}
	if !tx1.Rolled {
		t.Error("tx1 should be rolled back")
	}
	if tx1.Success {
		t.Error("tx1 success should be false after rollback")
	}
	if s.metrics.Rollbacks.Load() != 1 {
		t.Errorf("Rollbacks = %d, want 1", s.metrics.Rollbacks.Load())
	}
}

func TestExecuteSpeculativeLowGasLimit(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(1, d)

	// GasLimit lower than 21000 → gasUsed is capped at GasLimit.
	// Because gasUsed == GasLimit, success = (gasUsed <= GasLimit) = true.
	wave := Wave{
		Tasks: []TxTask{
			{Index: 0, GasLimit: 1000},
		},
	}
	results := s.ExecuteSpeculative(wave, map[int]struct{}{})
	if results[0].GasUsed != 1000 {
		t.Errorf("GasUsed = %d, want 1000 (capped at GasLimit)", results[0].GasUsed)
	}
	// success = gasUsed <= GasLimit → 1000 <= 1000 → true.
	if !results[0].Success {
		t.Error("expected success=true when gasUsed == gasLimit")
	}
}

// --- ReExecute ---

func TestReExecuteNoRollbacks(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(2, d)

	results := []SpeculativeResult{
		{TxIndex: 0, GasUsed: 21000, Success: true, Rolled: false},
		{TxIndex: 1, GasUsed: 21000, Success: true, Rolled: false},
	}
	updated := s.ReExecute(results)
	for i, r := range updated {
		if r.ReExecuted {
			t.Errorf("result[%d] should not be re-executed", i)
		}
	}
	if s.metrics.ReExecutions.Load() != 0 {
		t.Errorf("ReExecutions = %d, want 0", s.metrics.ReExecutions.Load())
	}
}

func TestReExecuteRolledBack(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(2, d)

	results := []SpeculativeResult{
		{TxIndex: 0, GasUsed: 21000, Success: true, Rolled: false},
		{TxIndex: 1, GasUsed: 0, Success: false, Rolled: true},
	}
	updated := s.ReExecute(results)
	if updated[1].ReExecuted != true {
		t.Error("expected tx1 to be marked ReExecuted")
	}
	if updated[1].Success != true {
		t.Error("expected tx1 success=true after re-execution")
	}
	if updated[1].Rolled != false {
		t.Error("expected tx1 Rolled=false after re-execution")
	}
	if s.metrics.ReExecutions.Load() != 1 {
		t.Errorf("ReExecutions = %d, want 1", s.metrics.ReExecutions.Load())
	}
}

func TestReExecutePreservesNonRolled(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(2, d)

	results := []SpeculativeResult{
		{TxIndex: 0, GasUsed: 50000, Success: true, Rolled: false},
		{TxIndex: 1, GasUsed: 0, Success: false, Rolled: true},
	}
	updated := s.ReExecute(results)
	if updated[0].GasUsed != 50000 {
		t.Errorf("non-rolled tx GasUsed changed: got %d, want 50000", updated[0].GasUsed)
	}
}

// --- topoSort ---

func TestTopoSortEmpty(t *testing.T) {
	order, err := topoSort(map[int][]int{})
	if err != nil {
		t.Fatalf("topoSort empty: unexpected error: %v", err)
	}
	if len(order) != 0 {
		t.Errorf("expected empty order, got %v", order)
	}
}

func TestTopoSortNoDeps(t *testing.T) {
	graph := map[int][]int{
		0: nil,
		1: nil,
		2: nil,
	}
	order, err := topoSort(graph)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(order))
	}
}

func TestTopoSortLinearChain(t *testing.T) {
	// 0 → 1 → 2 (0 is predecessor of 1, etc.)
	graph := map[int][]int{
		0: nil,
		1: {0},
		2: {1},
	}
	order, err := topoSort(graph)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(order))
	}
	// node 0 must come before 1, 1 before 2.
	pos := make(map[int]int)
	for i, n := range order {
		pos[n] = i
	}
	if pos[0] >= pos[1] {
		t.Error("0 must come before 1")
	}
	if pos[1] >= pos[2] {
		t.Error("1 must come before 2")
	}
}

func TestTopoSortCycleDetected(t *testing.T) {
	// Cycle: 0 → 1 → 0.
	graph := map[int][]int{
		0: {1},
		1: {0},
	}
	_, err := topoSort(graph)
	if err != ErrCyclicDependency {
		t.Errorf("expected ErrCyclicDependency, got %v", err)
	}
}

// --- buildWaves ---

func TestBuildWavesEmpty(t *testing.T) {
	waves := buildWaves(nil, map[int][]int{})
	if waves != nil {
		t.Errorf("expected nil waves for empty order, got %v", waves)
	}
}

func TestBuildWavesNoDeps(t *testing.T) {
	// Three independent nodes → all in wave 0.
	graph := map[int][]int{0: nil, 1: nil, 2: nil}
	order := []int{0, 1, 2}
	waves := buildWaves(order, graph)
	if len(waves) != 1 {
		t.Fatalf("expected 1 wave for independent txs, got %d", len(waves))
	}
	if len(waves[0].Tasks) != 3 {
		t.Errorf("expected 3 tasks in wave 0, got %d", len(waves[0].Tasks))
	}
}

func TestBuildWavesLinearChain(t *testing.T) {
	graph := map[int][]int{
		0: nil,
		1: {0},
		2: {1},
	}
	order := []int{0, 1, 2}
	waves := buildWaves(order, graph)
	if len(waves) != 3 {
		t.Fatalf("expected 3 waves for chain, got %d", len(waves))
	}
}

func TestBuildWavesDiamond(t *testing.T) {
	// Diamond: 0 is root; 1 and 2 depend on 0; 3 depends on 1 and 2.
	graph := map[int][]int{
		0: nil,
		1: {0},
		2: {0},
		3: {1, 2},
	}
	order, err := topoSort(graph)
	if err != nil {
		t.Fatalf("topoSort error: %v", err)
	}
	waves := buildWaves(order, graph)
	// Wave 0: node 0; Wave 1: nodes 1,2; Wave 2: node 3.
	if len(waves) != 3 {
		t.Fatalf("expected 3 waves for diamond, got %d", len(waves))
	}
	if len(waves[1].Tasks) != 2 {
		t.Errorf("expected 2 tasks in wave 1, got %d", len(waves[1].Tasks))
	}
}

// --- ParallelismRatio ---

func TestParallelismRatioInitial(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(2, d)
	if r := s.ParallelismRatio(); r != 1.0 {
		t.Errorf("ParallelismRatio() = %f, want 1.0 when no scheduling done", r)
	}
}

func TestParallelismRatioAfterIndependent(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(4, d)
	// 4 independent txs → 1 wave → ratio = 4/1 = 4.
	b := makeIndependentBAL(4)
	_, err := s.Schedule(b)
	if err != nil {
		t.Fatalf("Schedule failed: %v", err)
	}
	ratio := s.ParallelismRatio()
	if ratio != 4.0 {
		t.Errorf("ParallelismRatio() = %f, want 4.0", ratio)
	}
}

func TestParallelismRatioChained(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(2, d)
	// 3 chained conflicting txs → 3 waves → ratio = 3/3 = 1.
	b := makeConflictingBAL(3)
	_, err := s.Schedule(b)
	if err != nil {
		t.Fatalf("Schedule failed: %v", err)
	}
	ratio := s.ParallelismRatio()
	if ratio != 1.0 {
		t.Errorf("ParallelismRatio() = %f, want 1.0 for fully chained txs", ratio)
	}
}

func TestMaxWaveSizeUpdated(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	s, _ := NewBALScheduler(4, d)
	// 4 independent txs → 1 wave of size 4.
	b := makeIndependentBAL(4)
	_, err := s.Schedule(b)
	if err != nil {
		t.Fatalf("Schedule failed: %v", err)
	}
	if s.metrics.MaxWaveSize.Load() != 4 {
		t.Errorf("MaxWaveSize = %d, want 4", s.metrics.MaxWaveSize.Load())
	}
}
