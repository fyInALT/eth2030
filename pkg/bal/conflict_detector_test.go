package bal

import (
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// --- ConflictType.String ---

func TestConflictTypeString(t *testing.T) {
	cases := []struct {
		ct   ConflictType
		want string
	}{
		{ConflictReadWrite, "read-write"},
		{ConflictWriteRead, "write-read"},
		{ConflictWriteWrite, "write-write"},
		{ConflictAccountLevel, "account-level"},
		{ConflictType(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.ct.String(); got != tc.want {
			t.Errorf("ConflictType(%d).String() = %q, want %q", tc.ct, got, tc.want)
		}
	}
}

// --- ResolutionStrategy.String ---

func TestResolutionStrategyString(t *testing.T) {
	cases := []struct {
		rs   ResolutionStrategy
		want string
	}{
		{StrategySerialize, "serialize"},
		{StrategyAbort, "abort"},
		{StrategyRetry, "retry"},
		{ResolutionStrategy(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.rs.String(); got != tc.want {
			t.Errorf("ResolutionStrategy(%d).String() = %q, want %q", tc.rs, got, tc.want)
		}
	}
}

// --- ConflictMetrics.Snapshot ---

func TestConflictMetricsSnapshot(t *testing.T) {
	var m ConflictMetrics
	m.TotalPairs.Store(10)
	m.ConflictsFound.Store(4)
	m.ReadWriteCount.Store(1)
	m.WriteReadCount.Store(2)
	m.WriteWriteCount.Store(3)
	m.AccountConflicts.Store(5)
	m.ParallelFeasible.Store(6)
	m.SerialRequired.Store(7)

	s := m.Snapshot()
	if s.TotalPairs != 10 {
		t.Errorf("TotalPairs = %d, want 10", s.TotalPairs)
	}
	if s.ConflictsFound != 4 {
		t.Errorf("ConflictsFound = %d, want 4", s.ConflictsFound)
	}
	if s.ReadWriteCount != 1 {
		t.Errorf("ReadWriteCount = %d, want 1", s.ReadWriteCount)
	}
	if s.WriteReadCount != 2 {
		t.Errorf("WriteReadCount = %d, want 2", s.WriteReadCount)
	}
	if s.WriteWriteCount != 3 {
		t.Errorf("WriteWriteCount = %d, want 3", s.WriteWriteCount)
	}
	if s.AccountConflicts != 5 {
		t.Errorf("AccountConflicts = %d, want 5", s.AccountConflicts)
	}
	if s.ParallelFeasible != 6 {
		t.Errorf("ParallelFeasible = %d, want 6", s.ParallelFeasible)
	}
	if s.SerialRequired != 7 {
		t.Errorf("SerialRequired = %d, want 7", s.SerialRequired)
	}
}

// --- NewBALConflictDetector / Strategy / SetStrategy ---

func TestNewBALConflictDetectorStrategy(t *testing.T) {
	d := NewBALConflictDetector(StrategyAbort)
	if d == nil {
		t.Fatal("NewBALConflictDetector returned nil")
	}
	if d.Strategy() != StrategyAbort {
		t.Errorf("Strategy() = %v, want StrategyAbort", d.Strategy())
	}
}

func TestSetStrategy(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	d.SetStrategy(StrategyRetry)
	if d.Strategy() != StrategyRetry {
		t.Errorf("after SetStrategy: got %v, want StrategyRetry", d.Strategy())
	}
}

// --- Metrics ---

func TestMetricsReference(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	m := d.Metrics()
	if m == nil {
		t.Fatal("Metrics() returned nil")
	}
	// Mutate via reference and verify it reflects in the detector.
	m.TotalPairs.Store(42)
	if d.Metrics().TotalPairs.Load() != 42 {
		t.Error("Metrics() not returning reference to same struct")
	}
}

// helpers for building BAL entries for tests.

func makeBALWithStorageReads(txIdx uint64, addr types.Address, slots ...types.Hash) *BlockAccessList {
	b := NewBlockAccessList()
	entry := AccessEntry{
		Address:     addr,
		AccessIndex: txIdx,
	}
	for _, s := range slots {
		entry.StorageReads = append(entry.StorageReads, StorageAccess{Slot: s})
	}
	b.AddEntry(entry)
	return b
}

func makeBALWithStorageChanges(txIdx uint64, addr types.Address, slots ...types.Hash) *BlockAccessList {
	b := NewBlockAccessList()
	entry := AccessEntry{
		Address:     addr,
		AccessIndex: txIdx,
	}
	for _, s := range slots {
		entry.StorageChanges = append(entry.StorageChanges, StorageChange{Slot: s})
	}
	b.AddEntry(entry)
	return b
}

// mergeBAL creates a new BAL containing all entries from multiple BALs.
func mergeBAL(bals ...*BlockAccessList) *BlockAccessList {
	result := NewBlockAccessList()
	for _, b := range bals {
		for _, e := range b.Entries {
			result.AddEntry(e)
		}
	}
	return result
}

// --- IsParallelFeasible ---

func TestIsParallelFeasibleNil(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	if d.IsParallelFeasible(nil) {
		t.Error("expected false for nil BAL")
	}
}

func TestIsParallelFeasibleSingleTx(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	addr := types.HexToAddress("0x01")
	slot := types.HexToHash("0x01")
	b := makeBALWithStorageReads(1, addr, slot)
	if d.IsParallelFeasible(b) {
		t.Error("expected false for single tx BAL")
	}
}

func TestIsParallelFeasibleNoConflicts(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	addr1 := types.HexToAddress("0xA1")
	addr2 := types.HexToAddress("0xA2")
	slot1 := types.HexToHash("0x01")
	slot2 := types.HexToHash("0x02")

	// tx1 reads slot1@addr1, tx2 reads slot2@addr2 — no conflict.
	b := mergeBAL(
		makeBALWithStorageReads(1, addr1, slot1),
		makeBALWithStorageReads(2, addr2, slot2),
	)
	if !d.IsParallelFeasible(b) {
		t.Error("expected true: two txs with no shared slots")
	}
}

func TestIsParallelFeasibleAllConflict(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	addr := types.HexToAddress("0xAA")
	slot := types.HexToHash("0x01")

	// tx1 writes slot, tx2 writes slot → write-write conflict on only pair.
	b := mergeBAL(
		makeBALWithStorageChanges(1, addr, slot),
		makeBALWithStorageChanges(2, addr, slot),
	)
	if d.IsParallelFeasible(b) {
		t.Error("expected false: every pair conflicts")
	}
}

func TestIsParallelFeasibleThreeTxsMixed(t *testing.T) {
	// tx0 and tx1 share a slot (conflict), but tx0 and tx2 do not.
	d := NewBALConflictDetector(StrategySerialize)
	addr := types.HexToAddress("0xBB")
	sharedSlot := types.HexToHash("0x01")
	otherSlot := types.HexToHash("0x02")

	b := mergeBAL(
		makeBALWithStorageChanges(1, addr, sharedSlot),
		makeBALWithStorageChanges(2, addr, sharedSlot),
		makeBALWithStorageReads(3, addr, otherSlot),
	)
	// tx0(idx=0) and tx2(idx=2) should be conflict-free → feasible.
	if !d.IsParallelFeasible(b) {
		t.Error("expected true: at least one pair has no conflict")
	}
}

// --- DetectConflicts (via BuildDependencyGraph indirectly, and directly) ---

func TestDetectConflictsNil(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	if got := d.DetectConflicts(nil); got != nil {
		t.Errorf("expected nil conflicts for nil BAL, got %v", got)
	}
}

func TestDetectConflictsEmpty(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	b := NewBlockAccessList()
	if got := d.DetectConflicts(b); got != nil {
		t.Errorf("expected nil conflicts for empty BAL, got %v", got)
	}
}

func TestDetectConflictsPreExecutionSkipped(t *testing.T) {
	// AccessIndex 0 = pre-execution, should be skipped.
	d := NewBALConflictDetector(StrategySerialize)
	addr := types.HexToAddress("0x01")
	slot := types.HexToHash("0x01")
	b := NewBlockAccessList()
	b.AddEntry(AccessEntry{
		Address:        addr,
		AccessIndex:    0,
		StorageChanges: []StorageChange{{Slot: slot}},
	})
	b.AddEntry(AccessEntry{
		Address:        addr,
		AccessIndex:    0,
		StorageChanges: []StorageChange{{Slot: slot}},
	})
	if got := d.DetectConflicts(b); len(got) != 0 {
		t.Errorf("expected no conflicts from pre-execution entries, got %d", len(got))
	}
}

func TestDetectConflictsWriteWrite(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	addr := types.HexToAddress("0xCC")
	slot := types.HexToHash("0x10")

	b := mergeBAL(
		makeBALWithStorageChanges(1, addr, slot),
		makeBALWithStorageChanges(2, addr, slot),
	)
	conflicts := d.DetectConflicts(b)
	if len(conflicts) == 0 {
		t.Fatal("expected write-write conflict")
	}
	found := false
	for _, c := range conflicts {
		if c.Type == ConflictWriteWrite && c.Address == addr && c.Slot == slot {
			found = true
		}
	}
	if !found {
		t.Error("write-write conflict not found in results")
	}
	if d.Metrics().WriteWriteCount.Load() == 0 {
		t.Error("WriteWriteCount metric not incremented")
	}
}

func TestDetectConflictsReadWrite(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	addr := types.HexToAddress("0xDD")
	slot := types.HexToHash("0x11")

	// tx1 reads slot, tx2 writes slot → read-write conflict.
	b := mergeBAL(
		makeBALWithStorageReads(1, addr, slot),
		makeBALWithStorageChanges(2, addr, slot),
	)
	conflicts := d.DetectConflicts(b)
	found := false
	for _, c := range conflicts {
		if c.Type == ConflictReadWrite && c.Address == addr && c.Slot == slot {
			found = true
		}
	}
	if !found {
		t.Error("read-write conflict not found")
	}
	if d.Metrics().ReadWriteCount.Load() == 0 {
		t.Error("ReadWriteCount metric not incremented")
	}
}

func TestDetectConflictsWriteRead(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	addr := types.HexToAddress("0xEE")
	slot := types.HexToHash("0x12")

	// tx1 writes slot, tx2 reads slot → write-read conflict.
	b := mergeBAL(
		makeBALWithStorageChanges(1, addr, slot),
		makeBALWithStorageReads(2, addr, slot),
	)
	conflicts := d.DetectConflicts(b)
	found := false
	for _, c := range conflicts {
		if c.Type == ConflictWriteRead && c.Address == addr && c.Slot == slot {
			found = true
		}
	}
	if !found {
		t.Error("write-read conflict not found")
	}
	if d.Metrics().WriteReadCount.Load() == 0 {
		t.Error("WriteReadCount metric not incremented")
	}
}

func TestDetectConflictsAccountLevel(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	addr := types.HexToAddress("0xFF")

	// Both txs modify the same account's balance → account-level conflict.
	b := NewBlockAccessList()
	b.AddEntry(AccessEntry{
		Address:       addr,
		AccessIndex:   1,
		BalanceChange: &BalanceChange{},
	})
	b.AddEntry(AccessEntry{
		Address:       addr,
		AccessIndex:   2,
		BalanceChange: &BalanceChange{},
	})

	conflicts := d.DetectConflicts(b)
	found := false
	for _, c := range conflicts {
		if c.Type == ConflictAccountLevel && c.Address == addr {
			found = true
		}
	}
	if !found {
		t.Error("account-level conflict not found")
	}
	if d.Metrics().AccountConflicts.Load() == 0 {
		t.Error("AccountConflicts metric not incremented")
	}
}

func TestDetectConflictsNoConflict(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	addr1 := types.HexToAddress("0x01")
	addr2 := types.HexToAddress("0x02")

	b := mergeBAL(
		makeBALWithStorageReads(1, addr1, types.HexToHash("0x01")),
		makeBALWithStorageReads(2, addr2, types.HexToHash("0x02")),
	)
	conflicts := d.DetectConflicts(b)
	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts, got %d", len(conflicts))
	}
	if d.Metrics().TotalPairs.Load() == 0 {
		t.Error("TotalPairs not incremented")
	}
}

// --- BuildDependencyGraph ---

func TestBuildDependencyGraphEmpty(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	graph := d.BuildDependencyGraph(NewBlockAccessList())
	if len(graph) != 0 {
		t.Errorf("expected empty graph, got %v", graph)
	}
}

func TestBuildDependencyGraphNoConflicts(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	addr1 := types.HexToAddress("0x11")
	addr2 := types.HexToAddress("0x22")

	b := mergeBAL(
		makeBALWithStorageReads(1, addr1, types.HexToHash("0x01")),
		makeBALWithStorageReads(2, addr2, types.HexToHash("0x02")),
	)
	graph := d.BuildDependencyGraph(b)
	// Both txs should have no dependencies.
	for _, deps := range graph {
		if len(deps) != 0 {
			t.Errorf("expected no deps, got %v", deps)
		}
	}
}

func TestBuildDependencyGraphWithConflict(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	addr := types.HexToAddress("0xAB")
	slot := types.HexToHash("0x01")

	// tx1 writes, tx2 writes → tx2 depends on tx1.
	b := mergeBAL(
		makeBALWithStorageChanges(1, addr, slot),
		makeBALWithStorageChanges(2, addr, slot),
	)
	graph := d.BuildDependencyGraph(b)

	// tx index in graph is 0-based (AccessIndex - 1).
	tx1Idx := 0
	tx2Idx := 1
	deps, ok := graph[tx2Idx]
	if !ok {
		t.Fatalf("tx2 not in graph")
	}
	found := false
	for _, dep := range deps {
		if dep == tx1Idx {
			found = true
		}
	}
	if !found {
		t.Errorf("tx2 should depend on tx1, deps = %v", deps)
	}
}

func TestBuildDependencyGraphDeduplicatesEdges(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	addr := types.HexToAddress("0xCC")
	slot1 := types.HexToHash("0x01")
	slot2 := types.HexToHash("0x02")

	// tx1 writes two slots, tx2 writes both same slots → two conflicts but one edge.
	b := NewBlockAccessList()
	b.AddEntry(AccessEntry{
		Address:     addr,
		AccessIndex: 1,
		StorageChanges: []StorageChange{
			{Slot: slot1},
			{Slot: slot2},
		},
	})
	b.AddEntry(AccessEntry{
		Address:     addr,
		AccessIndex: 2,
		StorageChanges: []StorageChange{
			{Slot: slot1},
			{Slot: slot2},
		},
	})

	graph := d.BuildDependencyGraph(b)
	tx2Idx := 1
	deps := graph[tx2Idx]
	// Should be deduplicated: only one dep on tx1.
	count := 0
	for _, dep := range deps {
		if dep == 0 {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 edge from tx2 to tx1, got %d", count)
	}
}

// --- ResolveConflicts ---

func TestResolveConflictsEmpty(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	actions := d.ResolveConflicts(nil)
	if len(actions) != 0 {
		t.Errorf("expected empty actions for no conflicts, got %v", actions)
	}
}

func TestResolveConflictsSerialize(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	conflicts := []Conflict{
		{TxA: 0, TxB: 1, Type: ConflictWriteWrite},
	}
	actions := d.ResolveConflicts(conflicts)
	if actions[0] != "execute" {
		t.Errorf("tx0 action = %q, want execute", actions[0])
	}
	if actions[1] != "serialize" {
		t.Errorf("tx1 action = %q, want serialize", actions[1])
	}
}

func TestResolveConflictsAbort(t *testing.T) {
	d := NewBALConflictDetector(StrategyAbort)
	conflicts := []Conflict{
		{TxA: 0, TxB: 2, Type: ConflictReadWrite},
	}
	actions := d.ResolveConflicts(conflicts)
	if actions[0] != "execute" {
		t.Errorf("tx0 action = %q, want execute", actions[0])
	}
	if actions[2] != "abort" {
		t.Errorf("tx2 action = %q, want abort", actions[2])
	}
}

func TestResolveConflictsRetry(t *testing.T) {
	d := NewBALConflictDetector(StrategyRetry)
	conflicts := []Conflict{
		{TxA: 1, TxB: 3, Type: ConflictWriteRead},
	}
	actions := d.ResolveConflicts(conflicts)
	if actions[1] != "execute" {
		t.Errorf("tx1 action = %q, want execute", actions[1])
	}
	if actions[3] != "retry" {
		t.Errorf("tx3 action = %q, want retry", actions[3])
	}
}

func TestResolveConflictsMultiple(t *testing.T) {
	d := NewBALConflictDetector(StrategyAbort)
	conflicts := []Conflict{
		{TxA: 0, TxB: 1, Type: ConflictWriteWrite},
		{TxA: 0, TxB: 2, Type: ConflictReadWrite},
	}
	actions := d.ResolveConflicts(conflicts)
	if actions[0] != "execute" {
		t.Errorf("tx0 should execute, got %q", actions[0])
	}
	if actions[1] != "abort" {
		t.Errorf("tx1 should abort, got %q", actions[1])
	}
	if actions[2] != "abort" {
		t.Errorf("tx2 should abort, got %q", actions[2])
	}
}

// --- ConflictRate ---

func TestConflictRateZero(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	if r := d.ConflictRate(); r != 0.0 {
		t.Errorf("ConflictRate() = %f, want 0.0 when no pairs analyzed", r)
	}
}

func TestConflictRateAfterDetection(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	addr := types.HexToAddress("0x01")
	slot := types.HexToHash("0x01")

	// 2 txs, 1 pair → conflict rate should be > 0 when they conflict.
	b := mergeBAL(
		makeBALWithStorageChanges(1, addr, slot),
		makeBALWithStorageChanges(2, addr, slot),
	)
	d.DetectConflicts(b)

	rate := d.ConflictRate()
	if rate <= 0 {
		t.Errorf("ConflictRate() = %f, expected > 0 after a conflict", rate)
	}
}

func TestConflictRateNoConflicts(t *testing.T) {
	d := NewBALConflictDetector(StrategySerialize)
	addr1 := types.HexToAddress("0x01")
	addr2 := types.HexToAddress("0x02")

	b := mergeBAL(
		makeBALWithStorageReads(1, addr1, types.HexToHash("0x01")),
		makeBALWithStorageReads(2, addr2, types.HexToHash("0x02")),
	)
	d.DetectConflicts(b)

	if r := d.ConflictRate(); r != 0.0 {
		t.Errorf("ConflictRate() = %f, want 0.0 for no-conflict case", r)
	}
}
