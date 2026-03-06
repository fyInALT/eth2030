// iblt_test.go tests the Invertible Bloom Lookup Table implementation.
package netutil

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

// TestIBLTBasicInsertDecode verifies a simple insert and decode cycle.
func TestIBLTBasicInsertDecode(t *testing.T) {
	ib := NewIBLT(32, 3)
	ib.Insert(42)
	ib.Insert(99)

	ins, del, ok := ib.Decode()
	if !ok {
		t.Fatal("Decode failed")
	}
	if len(del) != 0 {
		t.Errorf("expected no deleted items, got %v", del)
	}
	sort.Slice(ins, func(i, j int) bool { return ins[i] < ins[j] })
	if len(ins) != 2 || ins[0] != 42 || ins[1] != 99 {
		t.Errorf("inserted = %v, want [42 99]", ins)
	}
}

// TestIBLT verifies the core set reconciliation property:
// setA = {1,2,3,4,5}, setB = {3,4,5,6,7}
// A\B = {1,2}, B\A = {6,7}
func TestIBLT(t *testing.T) {
	cells := 32
	k := 3
	sketchA := NewIBLT(cells, k)
	sketchB := NewIBLT(cells, k)

	setA := []uint64{1, 2, 3, 4, 5}
	setB := []uint64{3, 4, 5, 6, 7}

	for _, item := range setA {
		sketchA.Insert(item)
	}
	for _, item := range setB {
		sketchB.Insert(item)
	}

	diff := sketchA.Subtract(sketchB)
	inserted, deleted, ok := diff.Decode()
	if !ok {
		t.Fatal("Decode of difference IBLT failed")
	}

	sort.Slice(inserted, func(i, j int) bool { return inserted[i] < inserted[j] })
	sort.Slice(deleted, func(i, j int) bool { return deleted[i] < deleted[j] })

	wantInserted := []uint64{1, 2} // in A but not B
	wantDeleted := []uint64{6, 7}  // in B but not A

	if !uint64SliceEqual(inserted, wantInserted) {
		t.Errorf("inserted = %v, want %v", inserted, wantInserted)
	}
	if !uint64SliceEqual(deleted, wantDeleted) {
		t.Errorf("deleted = %v, want %v", deleted, wantDeleted)
	}
}

// TestIBLTSubtract verifies that Subtract does not mutate either IBLT.
func TestIBLTSubtract(t *testing.T) {
	a := NewIBLT(32, 3)
	b := NewIBLT(32, 3)
	a.Insert(10)
	a.Insert(20)
	b.Insert(20)
	b.Insert(30)

	diff := a.Subtract(b)
	// a and b should be unchanged.
	ins, _, ok := a.Decode()
	if !ok {
		t.Fatal("a.Decode failed after Subtract")
	}
	sort.Slice(ins, func(i, j int) bool { return ins[i] < ins[j] })
	if !uint64SliceEqual(ins, []uint64{10, 20}) {
		t.Errorf("a was mutated by Subtract: got %v", ins)
	}

	// diff should have inserted={10}, deleted={30}
	dIns, dDel, ok := diff.Decode()
	if !ok {
		t.Fatal("diff.Decode failed")
	}
	sort.Slice(dIns, func(i, j int) bool { return dIns[i] < dIns[j] })
	sort.Slice(dDel, func(i, j int) bool { return dDel[i] < dDel[j] })
	if !uint64SliceEqual(dIns, []uint64{10}) {
		t.Errorf("diff inserted = %v, want [10]", dIns)
	}
	if !uint64SliceEqual(dDel, []uint64{30}) {
		t.Errorf("diff deleted = %v, want [30]", dDel)
	}
}

// TestIBLTDecodeFailure verifies that decoding fails gracefully when the
// IBLT is too small for the number of differences.
func TestIBLTDecodeFailure(t *testing.T) {
	// Use a tiny IBLT that cannot hold many differences.
	tiny := NewIBLT(3, 3)
	for i := uint64(0); i < 100; i++ {
		tiny.Insert(i)
	}
	// We don't subtract anything, so this is 100 insertions in 3 cells.
	// Decode should fail (cannot peel 100 items from 3 cells).
	_, _, ok := tiny.Decode()
	if ok {
		t.Error("expected Decode to fail for over-saturated IBLT, but it succeeded")
	}
}

// TestIBLTLargeSet verifies that the IBLT can handle moderate difference sets
// with high reliability (failure rate < 1%).
func TestIBLTLargeSet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large-set IBLT test in short mode")
	}

	const iterations = 200
	const diffSize = 5 // 5 per side = 10 total diff items
	const cells = 200  // 20x total diff for high reliability

	failures := 0
	rng := rand.New(rand.NewSource(42))

	for iter := 0; iter < iterations; iter++ {
		// Build two sets with a known symmetric difference.
		shared := make([]uint64, 500)
		for i := range shared {
			shared[i] = rng.Uint64()
		}

		onlyA := make([]uint64, diffSize)
		onlyB := make([]uint64, diffSize)
		for i := range onlyA {
			onlyA[i] = rng.Uint64()
			onlyB[i] = rng.Uint64()
		}

		skA := NewIBLT(cells, 3)
		skB := NewIBLT(cells, 3)
		for _, x := range shared {
			skA.Insert(x)
			skB.Insert(x)
		}
		for _, x := range onlyA {
			skA.Insert(x)
		}
		for _, x := range onlyB {
			skB.Insert(x)
		}

		diff := skA.Subtract(skB)
		ins, del, ok := diff.Decode()
		if !ok {
			failures++
			continue
		}
		if len(ins) != diffSize || len(del) != diffSize {
			failures++
		}
		_ = ins
		_ = del
	}

	failRate := float64(failures) / float64(iterations)
	if failRate > 0.01 {
		t.Errorf("failure rate %.2f%% exceeds 1%% threshold (%d/%d)", failRate*100, failures, iterations)
	}
}

// TestIBLTDeleteThenDecode verifies the Delete path.
func TestIBLTDeleteThenDecode(t *testing.T) {
	ib := NewIBLT(32, 3)
	ib.Insert(100)
	ib.Insert(200)
	ib.Insert(300)
	ib.Delete(200)

	ins, del, ok := ib.Decode()
	if !ok {
		t.Fatal("Decode failed")
	}
	sort.Slice(ins, func(i, j int) bool { return ins[i] < ins[j] })
	if !uint64SliceEqual(ins, []uint64{100, 300}) {
		t.Errorf("inserted = %v, want [100 300]", ins)
	}
	if len(del) != 0 {
		t.Errorf("deleted = %v, want []", del)
	}
}

// TestIBLTEmptyDecode verifies that an empty IBLT decodes successfully.
func TestIBLTEmptyDecode(t *testing.T) {
	ib := NewIBLT(16, 3)
	ins, del, ok := ib.Decode()
	if !ok {
		t.Fatal("empty IBLT Decode should succeed")
	}
	if len(ins) != 0 || len(del) != 0 {
		t.Errorf("empty IBLT Decode returned non-empty: ins=%v del=%v", ins, del)
	}
}

// uint64SliceEqual returns true if a and b have the same elements in order.
func uint64SliceEqual(a, b []uint64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestIBLTHashDistribution verifies that items spread across cells reasonably.
func TestIBLTHashDistribution(t *testing.T) {
	cells := 64
	k := 3
	ib := NewIBLT(cells, k)
	for i := uint64(0); i < 100; i++ {
		ib.Insert(i)
	}
	// Count non-empty cells.
	nonEmpty := 0
	for _, cell := range ib.cells {
		if cell.Count != 0 {
			nonEmpty++
		}
	}
	// With 100 items and k=3, we expect the vast majority of cells to be non-empty.
	if nonEmpty < cells/2 {
		t.Errorf("only %d/%d cells non-empty; distribution seems too sparse", nonEmpty, cells)
	}
	_ = fmt.Sprintf("non-empty cells: %d/%d", nonEmpty, cells)
}
