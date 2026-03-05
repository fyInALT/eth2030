// set_reconciliation_test.go tests the IBLT-based set reconciliation protocol.
package p2p

import (
	"math/rand"
	"sort"
	"testing"
)

// TestSetReconciliation verifies basic set reconciliation where two sets
// differ by a small symmetric difference.
func TestSetReconciliation(t *testing.T) {
	proto := NewSetReconciliationProtocol(128, 3)

	setA := []uint64{1, 2, 3, 4, 5, 100, 200}
	setB := []uint64{3, 4, 5, 6, 7, 100, 200}

	sketchA := proto.BuildSketch(setA)
	sketchB := proto.BuildSketch(setB)

	missing, extra, ok := proto.Reconcile(sketchA, sketchB)
	if !ok {
		t.Fatal("Reconcile failed")
	}

	sort.Slice(missing, func(i, j int) bool { return missing[i] < missing[j] })
	sort.Slice(extra, func(i, j int) bool { return extra[i] < extra[j] })

	// missing = items in A but not in B = {1, 2}
	wantMissing := []uint64{1, 2}
	// extra = items in B but not in A = {6, 7}
	wantExtra := []uint64{6, 7}

	if !uint64SliceEqual(missing, wantMissing) {
		t.Errorf("missing = %v, want %v", missing, wantMissing)
	}
	if !uint64SliceEqual(extra, wantExtra) {
		t.Errorf("extra = %v, want %v", extra, wantExtra)
	}
}

// TestSetReconciliationHighOverlap verifies reconciliation when the two sets
// share most items but differ by a handful.
func TestSetReconciliationHighOverlap(t *testing.T) {
	const sharedCount = 1000
	const diffCount = 10

	proto := NewSetReconciliationProtocol(256, 3)
	rng := rand.New(rand.NewSource(99))

	shared := make([]uint64, sharedCount)
	for i := range shared {
		shared[i] = rng.Uint64()
	}

	onlyA := make([]uint64, diffCount)
	onlyB := make([]uint64, diffCount)
	for i := range onlyA {
		onlyA[i] = rng.Uint64()
		onlyB[i] = rng.Uint64()
	}

	setA := append(shared, onlyA...)
	setB := append(shared, onlyB...)

	sketchA := proto.BuildSketch(setA)
	sketchB := proto.BuildSketch(setB)

	missing, extra, ok := proto.Reconcile(sketchA, sketchB)
	if !ok {
		t.Fatal("Reconcile failed for high-overlap sets")
	}

	if len(missing) != diffCount {
		t.Errorf("missing count = %d, want %d", len(missing), diffCount)
	}
	if len(extra) != diffCount {
		t.Errorf("extra count = %d, want %d", len(extra), diffCount)
	}

	// Verify that missing items are exactly onlyA and extra items are onlyB.
	sort.Slice(missing, func(i, j int) bool { return missing[i] < missing[j] })
	sort.Slice(extra, func(i, j int) bool { return extra[i] < extra[j] })
	sortedOnlyA := make([]uint64, len(onlyA))
	copy(sortedOnlyA, onlyA)
	sort.Slice(sortedOnlyA, func(i, j int) bool { return sortedOnlyA[i] < sortedOnlyA[j] })
	sortedOnlyB := make([]uint64, len(onlyB))
	copy(sortedOnlyB, onlyB)
	sort.Slice(sortedOnlyB, func(i, j int) bool { return sortedOnlyB[i] < sortedOnlyB[j] })

	if !uint64SliceEqual(missing, sortedOnlyA) {
		t.Errorf("missing items incorrect: got %v, want %v", missing, sortedOnlyA)
	}
	if !uint64SliceEqual(extra, sortedOnlyB) {
		t.Errorf("extra items incorrect: got %v, want %v", extra, sortedOnlyB)
	}
}

// TestSetReconciliationIdentical verifies that reconciling identical sets
// yields empty missing and extra.
func TestSetReconciliationIdentical(t *testing.T) {
	proto := NewSetReconciliationProtocol(64, 3)

	items := []uint64{10, 20, 30, 40, 50}
	sketchA := proto.BuildSketch(items)
	sketchB := proto.BuildSketch(items)

	missing, extra, ok := proto.Reconcile(sketchA, sketchB)
	if !ok {
		t.Fatal("Reconcile failed for identical sets")
	}
	if len(missing) != 0 || len(extra) != 0 {
		t.Errorf("identical sets: missing=%v extra=%v, want both empty", missing, extra)
	}
}

// TestSetReconciliationDisjoint verifies reconciliation for completely
// disjoint sets.
func TestSetReconciliationDisjoint(t *testing.T) {
	const n = 5
	proto := NewSetReconciliationProtocol(128, 3)

	setA := []uint64{1, 2, 3, 4, 5}
	setB := []uint64{6, 7, 8, 9, 10}

	sketchA := proto.BuildSketch(setA)
	sketchB := proto.BuildSketch(setB)

	missing, extra, ok := proto.Reconcile(sketchA, sketchB)
	if !ok {
		t.Fatal("Reconcile failed for disjoint sets")
	}
	if len(missing) != n || len(extra) != n {
		t.Errorf("missing=%d extra=%d, want both %d", len(missing), len(extra), n)
	}
}

// TestSetReconciliationEmptyB verifies that reconciling A against an empty B
// yields all of A as missing items.
func TestSetReconciliationEmptyB(t *testing.T) {
	proto := NewSetReconciliationProtocol(64, 3)

	setA := []uint64{11, 22, 33}
	sketchA := proto.BuildSketch(setA)
	sketchB := proto.BuildSketch(nil)

	missing, extra, ok := proto.Reconcile(sketchA, sketchB)
	if !ok {
		t.Fatal("Reconcile failed for B=empty")
	}
	if len(extra) != 0 {
		t.Errorf("extra = %v, want empty when B is empty", extra)
	}
	sort.Slice(missing, func(i, j int) bool { return missing[i] < missing[j] })
	if !uint64SliceEqual(missing, setA) {
		t.Errorf("missing = %v, want %v", missing, setA)
	}
}
