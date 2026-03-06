// iblt.go implements an Invertible Bloom Lookup Table (IBLT) for
// rateless set reconciliation of 64-bit items. IBLTs allow two parties
// to compute the symmetric difference of their sets without full
// enumeration, using only a compact sketch.
//
// Reference: Goodrich & Mitzenmacher, "Invertible Bloom Lookup Tables", 2011.
package netutil

// IBLTCell represents one cell in the IBLT.
type IBLTCell struct {
	Count   int32  // number of insertions minus deletions
	IDSum   uint64 // XOR of all item IDs in this cell
	HashSum uint64 // XOR of ibltItemHash(item) for all items in this cell
}

// IBLT is an Invertible Bloom Lookup Table for 64-bit items.
type IBLT struct {
	cells []IBLTCell
	k     int // number of hash functions
}

// NewIBLT creates an IBLT with the given number of cells and hash functions.
// Recommended parameters: cells = 1.5 * expected_diff_size, k = 3.
func NewIBLT(cells, k int) *IBLT {
	return &IBLT{
		cells: make([]IBLTCell, cells),
		k:     k,
	}
}

// Insert adds an item to the IBLT.
func (ib *IBLT) Insert(item uint64) {
	h := ibltItemHash(item)
	for i := 0; i < ib.k; i++ {
		idx := ibltCellIndex(item, i, len(ib.cells))
		ib.cells[idx].Count++
		ib.cells[idx].IDSum ^= item
		ib.cells[idx].HashSum ^= h
	}
}

// Delete removes an item from the IBLT (inverse of Insert).
func (ib *IBLT) Delete(item uint64) {
	h := ibltItemHash(item)
	for i := 0; i < ib.k; i++ {
		idx := ibltCellIndex(item, i, len(ib.cells))
		ib.cells[idx].Count--
		ib.cells[idx].IDSum ^= item
		ib.cells[idx].HashSum ^= h
	}
}

// Subtract returns a new IBLT representing the difference: ib - other.
// Neither ib nor other is modified.
// Used for set reconciliation: A.Subtract(B) encodes symmetric difference.
func (ib *IBLT) Subtract(other *IBLT) *IBLT {
	n := len(ib.cells)
	result := &IBLT{
		cells: make([]IBLTCell, n),
		k:     ib.k,
	}
	for i := 0; i < n; i++ {
		result.cells[i] = IBLTCell{
			Count:   ib.cells[i].Count - other.cells[i].Count,
			IDSum:   ib.cells[i].IDSum ^ other.cells[i].IDSum,
			HashSum: ib.cells[i].HashSum ^ other.cells[i].HashSum,
		}
	}
	return result
}

// Decode peels the IBLT to recover inserted and deleted items.
// Returns ok=false if decoding fails (too many differences for cell count).
//
// Items appearing in inserted are those with net positive count (in ib but not
// in the subtracted table); items in deleted have net negative count.
func (ib *IBLT) Decode() (inserted, deleted []uint64, ok bool) {
	// Work on a copy so the original IBLT is not mutated.
	work := make([]IBLTCell, len(ib.cells))
	copy(work, ib.cells)

	// Iteratively peel pure cells (|count| == 1).
	for {
		peeled := false
		for i := range work {
			c := &work[i]
			if c.Count == 0 {
				continue
			}
			// A pure cell: exactly one item (or its inverse).
			if c.Count == 1 || c.Count == -1 {
				item := c.IDSum
				// Verify the cell is consistent: only peel if the fingerprint
				// matches. A mismatch means the cell contains multiple items
				// whose counts happened to sum to ±1 — skip it and try others.
				if ibltItemHash(item) != c.HashSum {
					continue
				}
				// Capture count before mutation (pointer c may alias work[idx]).
				origCount := c.Count
				if origCount == 1 {
					inserted = append(inserted, item)
				} else {
					deleted = append(deleted, item)
				}
				// Remove the item from all k cells.
				h := ibltItemHash(item)
				for ki := 0; ki < ib.k; ki++ {
					idx := ibltCellIndex(item, ki, len(work))
					work[idx].Count -= origCount
					work[idx].IDSum ^= item
					work[idx].HashSum ^= h
				}
				peeled = true
				break // restart scan after each peel
			}
		}
		if !peeled {
			break
		}
	}

	// Check whether all cells are empty.
	for i := range work {
		if work[i].Count != 0 {
			return nil, nil, false
		}
	}
	return inserted, deleted, true
}

// ibltKPrimes provides per-k additive constants for independent cell hashing.
// Using additive perturbation (rather than XOR) avoids correlated outputs.
var ibltKPrimes = [8]uint64{
	0x517cc1b727220a95,
	0x6c62272e07bb0142,
	0x4be98134a5976fd3,
	0xea1b52c80b2b4b2f,
	0x3bc0993a5ad19f85,
	0xc3a5c85c97cb3127,
	0x96de1b173f119089,
	0xb492b66fbe98f273,
}

// ibltCellIndex returns the cell index for item under hash function k.
// Each k uses a distinct additive constant before the finaliser to produce
// k statistically independent hash functions.
func ibltCellIndex(item uint64, k, numCells int) int {
	x := item + ibltKPrimes[k%len(ibltKPrimes)]
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	x ^= x >> 31
	return int(uint32(x)) % numCells
}

// ibltItemHash returns a 64-bit fingerprint for item used to detect purity in
// cells. A 64-bit checksum reduces false-positive peel rate to ~2^-64.
func ibltItemHash(item uint64) uint64 {
	// Murmur3-inspired 64-bit finalizer.
	x := item ^ (item >> 33)
	x *= 0xff51afd7ed558ccd
	x ^= x >> 33
	x *= 0xc4ceb9fe1a85ec53
	x ^= x >> 33
	return x
}
