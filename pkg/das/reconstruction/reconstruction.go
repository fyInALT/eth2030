package reconstruction

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/eth2030/eth2030/das/dastypes"
	"github.com/eth2030/eth2030/das/field"
)

var (
	ErrInsufficientCells  = errors.New("das: insufficient cells for reconstruction")
	ErrInvalidCellIndex   = errors.New("das: cell index out of range")
	ErrDuplicateCellIndex = errors.New("das: duplicate cell index")
)

// CanReconstruct returns true if the number of received cells/columns is
// sufficient to reconstruct the full extended blob via Reed-Solomon decoding.
// At least 50% of columns (NUMBER_OF_COLUMNS / 2 = 64) are required.
func CanReconstruct(receivedCount int) bool {
	return receivedCount >= dastypes.ReconstructionThreshold
}

// validateCellInputs checks common preconditions for reconstruction.
func validateCellInputs(cells []dastypes.Cell, cellIndices []uint64) error {
	if len(cells) != len(cellIndices) {
		return fmt.Errorf("%w: %d cells but %d indices",
			ErrInsufficientCells, len(cells), len(cellIndices))
	}
	if !CanReconstruct(len(cells)) {
		return fmt.Errorf("%w: have %d, need %d",
			ErrInsufficientCells, len(cells), dastypes.ReconstructionThreshold)
	}

	seen := make(map[uint64]bool, len(cellIndices))
	for _, idx := range cellIndices {
		if idx >= dastypes.CellsPerExtBlob {
			return fmt.Errorf("%w: index %d >= %d",
				ErrInvalidCellIndex, idx, dastypes.CellsPerExtBlob)
		}
		if seen[idx] {
			return fmt.Errorf("%w: index %d", ErrDuplicateCellIndex, idx)
		}
		seen[idx] = true
	}
	return nil
}

// ReconstructPolynomial recovers a polynomial of degree < k from k evaluation
// points using Lagrange interpolation over the BLS12-381 scalar field.
//
// xs and ys must have the same length >= k. Only the first k points are used.
// The returned slice has length k, holding the polynomial coefficients
// [c_0, c_1, ..., c_{k-1}].
func ReconstructPolynomial(xs, ys []field.FieldElement, k int) ([]field.FieldElement, error) {
	if len(xs) != len(ys) {
		return nil, fmt.Errorf("das: xs and ys length mismatch: %d vs %d", len(xs), len(ys))
	}
	if len(xs) < k {
		return nil, fmt.Errorf("das: need at least %d points, have %d", k, len(xs))
	}

	// Use exactly k points.
	xs = xs[:k]
	ys = ys[:k]

	// Lagrange interpolation to recover coefficients.
	// First compute the Lagrange basis polynomials and accumulate.
	coeffs := make([]field.FieldElement, k)
	for i := range coeffs {
		coeffs[i] = field.FieldZero()
	}

	for i := 0; i < k; i++ {
		// Compute the i-th Lagrange basis polynomial's coefficient contribution.
		// L_i(x) = y_i * prod_{j!=i} (x - x_j) / (x_i - x_j)

		// First compute the denominator: prod_{j!=i} (x_i - x_j).
		denom := field.FieldOne()
		for j := 0; j < k; j++ {
			if j == i {
				continue
			}
			denom = denom.Mul(xs[i].Sub(xs[j]))
		}
		factor := ys[i].Div(denom)

		// Build the numerator polynomial prod_{j!=i} (x - x_j).
		// Start with [1] and multiply by (x - x_j) for each j != i.
		poly := make([]field.FieldElement, 1, k)
		poly[0] = field.FieldOne()
		for j := 0; j < k; j++ {
			if j == i {
				continue
			}
			// Multiply poly by (x - x_j).
			newPoly := make([]field.FieldElement, len(poly)+1)
			for m := range newPoly {
				newPoly[m] = field.FieldZero()
			}
			for m := range poly {
				// poly[m] * x -> newPoly[m+1]
				newPoly[m+1] = newPoly[m+1].Add(poly[m])
				// poly[m] * (-x_j) -> newPoly[m]
				newPoly[m] = newPoly[m].Sub(poly[m].Mul(xs[j]))
			}
			poly = newPoly
		}

		// Add factor * poly to coeffs.
		for m := range poly {
			if m < k {
				coeffs[m] = coeffs[m].Add(factor.Mul(poly[m]))
			}
		}
	}

	return coeffs, nil
}

// evaluatePolynomial evaluates a polynomial at point x.
// coeffs[i] is the coefficient of x^i.
func evaluatePolynomial(coeffs []field.FieldElement, x field.FieldElement) field.FieldElement {
	if len(coeffs) == 0 {
		return field.FieldZero()
	}
	// Horner's method.
	result := coeffs[len(coeffs)-1]
	for i := len(coeffs) - 2; i >= 0; i-- {
		result = result.Mul(x).Add(coeffs[i])
	}
	return result
}

// cellToFieldElements converts a cell's raw bytes to field elements.
// Each field element is 32 bytes (big-endian within BLS12-381 scalar field).
func cellToFieldElements(cell *dastypes.Cell) []field.FieldElement {
	elems := make([]field.FieldElement, dastypes.FieldElementsPerCell)
	for i := 0; i < dastypes.FieldElementsPerCell; i++ {
		b := new(big.Int).SetBytes(cell[i*dastypes.BytesPerFieldElement : (i+1)*dastypes.BytesPerFieldElement])
		elems[i] = field.NewFieldElement(b)
	}
	return elems
}

// fieldElementsToBytes converts field elements back to raw bytes.
func fieldElementsToBytes(elems []field.FieldElement, size int) []byte {
	result := make([]byte, size)
	for i, elem := range elems {
		offset := i * dastypes.BytesPerFieldElement
		if offset+dastypes.BytesPerFieldElement > size {
			break
		}
		b := elem.BigInt().Bytes()
		// Right-align in 32-byte slot.
		start := offset + dastypes.BytesPerFieldElement - len(b)
		copy(result[start:offset+dastypes.BytesPerFieldElement], b)
	}
	return result
}

// ReconstructBlob reconstructs a full blob from a partial set of cells using
// Reed-Solomon erasure coding recovery via Lagrange interpolation over the
// BLS12-381 scalar field.
//
// The blob is encoded as a polynomial of degree < FieldElementsPerBlob (4096),
// evaluated at CellsPerExtBlob (128) positions. Each cell contains
// FieldElementsPerCell (64) consecutive field elements from the evaluation
// domain. Given >= 50% of cells, we recover each field element column
// independently using Lagrange interpolation.
//
// Parameters:
//   - cells: the received cells (at least ReconstructionThreshold required)
//   - cellIndices: the column indices corresponding to each cell
//
// Returns the reconstructed blob data or an error.
func ReconstructBlob(cells []dastypes.Cell, cellIndices []uint64) ([]byte, error) {
	if err := validateCellInputs(cells, cellIndices); err != nil {
		return nil, err
	}

	numCells := len(cells)
	blobSize := dastypes.FieldElementsPerBlob * dastypes.BytesPerFieldElement

	// Compute the evaluation domain roots of unity.
	// The extended blob uses CellsPerExtBlob evaluation positions.
	// Each cell i holds field elements evaluated at positions related to cell index i.
	// We use the cell indices as the x-coordinates for interpolation.

	// For RS reconstruction, we treat each "field element position" within
	// a cell independently. For field element position j within cells,
	// we have evaluations at x = cellIndex for each received cell.
	// We need to interpolate and recover the original data blob (first half).

	// Precompute x-coordinates: use cell indices as evaluation points.
	xs := make([]field.FieldElement, numCells)
	for i, idx := range cellIndices {
		xs[i] = field.NewFieldElementFromUint64(idx)
	}

	// Parse all cells into field elements.
	cellElems := make([][]field.FieldElement, numCells)
	for i := range cells {
		cellElems[i] = cellToFieldElements(&cells[i])
	}

	// For each field element position within a cell, reconstruct the
	// polynomial and evaluate at the original positions (0..63).
	originalCells := dastypes.CellsPerExtBlob / 2
	resultElems := make([]field.FieldElement, dastypes.FieldElementsPerBlob)

	for fePos := 0; fePos < dastypes.FieldElementsPerCell; fePos++ {
		// Gather y-values for this field element position.
		ys := make([]field.FieldElement, numCells)
		for i := range cells {
			ys[i] = cellElems[i][fePos]
		}

		// Recover the polynomial (degree < CellsPerExtBlob).
		// We only need evaluations at the first half of indices (0..63).
		coeffs, err := ReconstructPolynomial(xs, ys, numCells)
		if err != nil {
			return nil, fmt.Errorf("das: polynomial reconstruction failed at position %d: %w", fePos, err)
		}

		// Evaluate at original cell indices 0..63 to get blob field elements.
		for cellIdx := 0; cellIdx < originalCells; cellIdx++ {
			x := field.NewFieldElementFromUint64(uint64(cellIdx))
			val := evaluatePolynomial(coeffs, x)
			resultElems[cellIdx*dastypes.FieldElementsPerCell+fePos] = val
		}
	}

	return fieldElementsToBytes(resultElems, blobSize), nil
}

// RecoverCellsAndProofs recovers all missing cells and their indices from
// a partial set of cells. Returns the full set of CellsPerExtBlob cells.
//
// Note: KZG proof recovery requires the commitment and is not implemented
// here. The returned proofs slice will contain zero proofs for recovered cells.
func RecoverCellsAndProofs(cells []dastypes.Cell, cellIndices []uint64) ([]dastypes.Cell, []dastypes.KZGProof, error) {
	if err := validateCellInputs(cells, cellIndices); err != nil {
		return nil, nil, err
	}

	numCells := len(cells)

	// Build index set for quick lookup.
	indexSet := make(map[uint64]int, numCells)
	for i, idx := range cellIndices {
		indexSet[idx] = i
	}

	// Precompute evaluation points.
	xs := make([]field.FieldElement, numCells)
	for i, idx := range cellIndices {
		xs[i] = field.NewFieldElementFromUint64(idx)
	}

	// Parse all cells into field elements.
	cellElems := make([][]field.FieldElement, numCells)
	for i := range cells {
		cellElems[i] = cellToFieldElements(&cells[i])
	}

	// Output: full set of cells and proofs.
	allCells := make([]dastypes.Cell, dastypes.CellsPerExtBlob)
	allProofs := make([]dastypes.KZGProof, dastypes.CellsPerExtBlob)

	// Copy existing cells.
	for i, idx := range cellIndices {
		allCells[idx] = cells[i]
	}

	// For each field element position, reconstruct and evaluate missing cells.
	for fePos := 0; fePos < dastypes.FieldElementsPerCell; fePos++ {
		ys := make([]field.FieldElement, numCells)
		for i := range cells {
			ys[i] = cellElems[i][fePos]
		}

		coeffs, err := ReconstructPolynomial(xs, ys, numCells)
		if err != nil {
			return nil, nil, fmt.Errorf("das: recovery failed at position %d: %w", fePos, err)
		}

		// Fill in missing cells.
		for cellIdx := uint64(0); cellIdx < dastypes.CellsPerExtBlob; cellIdx++ {
			if _, exists := indexSet[cellIdx]; exists {
				continue // Already have this cell.
			}
			x := field.NewFieldElementFromUint64(cellIdx)
			val := evaluatePolynomial(coeffs, x)
			b := val.BigInt().Bytes()
			start := fePos*dastypes.BytesPerFieldElement + dastypes.BytesPerFieldElement - len(b)
			copy(allCells[cellIdx][start:fePos*dastypes.BytesPerFieldElement+dastypes.BytesPerFieldElement], b)
		}
	}

	return allCells, allProofs, nil
}

// RecoverMatrix recovers the full matrix of cells from a partial set of
// matrix entries, following the consensus spec's recover_matrix helper.
// Each row (blob) is reconstructed independently.
func RecoverMatrix(entries []dastypes.MatrixEntry, blobCount int) ([]dastypes.MatrixEntry, error) {
	if blobCount <= 0 {
		return nil, fmt.Errorf("das: invalid blob count %d", blobCount)
	}

	// Group entries by row (blob).
	byRow := make(map[dastypes.RowIndex][]dastypes.MatrixEntry)
	for _, e := range entries {
		byRow[e.RowIndex] = append(byRow[e.RowIndex], e)
	}

	// Check each row has enough entries for reconstruction.
	for row := 0; row < blobCount; row++ {
		rowEntries := byRow[dastypes.RowIndex(row)]
		if !CanReconstruct(len(rowEntries)) {
			return nil, fmt.Errorf("%w: row %d has %d cells, need %d",
				ErrInsufficientCells, row, len(rowEntries), dastypes.ReconstructionThreshold)
		}
	}

	// Recover each row independently.
	var result []dastypes.MatrixEntry
	for row := 0; row < blobCount; row++ {
		rowEntries := byRow[dastypes.RowIndex(row)]

		cells := make([]dastypes.Cell, len(rowEntries))
		indices := make([]uint64, len(rowEntries))
		for i, e := range rowEntries {
			cells[i] = e.Cell
			indices[i] = uint64(e.ColumnIndex)
		}

		allCells, allProofs, err := RecoverCellsAndProofs(cells, indices)
		if err != nil {
			return nil, fmt.Errorf("das: row %d recovery failed: %w", row, err)
		}

		for colIdx := uint64(0); colIdx < dastypes.CellsPerExtBlob; colIdx++ {
			result = append(result, dastypes.MatrixEntry{
				Cell:        allCells[colIdx],
				KZGProof:    allProofs[colIdx],
				ColumnIndex: dastypes.ColumnIndex(colIdx),
				RowIndex:    dastypes.RowIndex(row),
			})
		}
	}

	return result, nil
}
