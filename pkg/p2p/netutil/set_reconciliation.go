// set_reconciliation.go implements a rateless set reconciliation protocol
// using IBLT sketches. Two peers can each build a local sketch, exchange it,
// and compute the symmetric difference without sending their full sets.
package netutil

// SetReconciliationProtocol reconciles two item sets using IBLT sketches.
// The cell count and k parameters determine the capacity (maximum symmetric
// difference size) that can be decoded reliably.
type SetReconciliationProtocol struct {
	ibltCells int
	ibltK     int
}

// NewSetReconciliationProtocol creates a new reconciliation protocol.
// cells should be >= 1.5 * expected_max_diff; k=3 is a good default.
func NewSetReconciliationProtocol(cells, k int) *SetReconciliationProtocol {
	return &SetReconciliationProtocol{
		ibltCells: cells,
		ibltK:     k,
	}
}

// BuildSketch creates an IBLT sketch of the given item set.
func (p *SetReconciliationProtocol) BuildSketch(items []uint64) *IBLT {
	ib := NewIBLT(p.ibltCells, p.ibltK)
	for _, item := range items {
		ib.Insert(item)
	}
	return ib
}

// Reconcile takes two sketches and returns the symmetric difference.
//
//   - missing: items present in sketchA but absent in sketchB (in A \ B).
//   - extra:   items present in sketchB but absent in sketchA (in B \ A).
//
// Returns ok=false when the difference exceeds the IBLT capacity.
func (p *SetReconciliationProtocol) Reconcile(sketchA, sketchB *IBLT) (missing, extra []uint64, ok bool) {
	diff := sketchA.Subtract(sketchB)
	// After A.Subtract(B):
	//   inserted (count=+1) = items in A but not B  => "missing" from B's perspective
	//   deleted  (count=-1) = items in B but not A  => "extra" in B
	inserted, deleted, ok := diff.Decode()
	if !ok {
		return nil, nil, false
	}
	return inserted, deleted, true
}
