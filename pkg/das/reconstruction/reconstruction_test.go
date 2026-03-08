package reconstruction

import (
	"testing"

	"github.com/eth2030/eth2030/das/dastypes"
	"github.com/eth2030/eth2030/das/field"
)

// --- validateCellInputs tests ---

func TestValidateCellInputsMismatchedLengths(t *testing.T) {
	cells := make([]dastypes.Cell, 5)
	indices := make([]uint64, 3)
	if err := validateCellInputs(cells, indices); err == nil {
		t.Error("expected error for mismatched lengths")
	}
}

func TestValidateCellInputsTooFew(t *testing.T) {
	cells := make([]dastypes.Cell, 10)
	indices := make([]uint64, 10)
	for i := range indices {
		indices[i] = uint64(i)
	}
	if err := validateCellInputs(cells, indices); err == nil {
		t.Error("expected error for insufficient cells")
	}
}

func TestValidateCellInputsIndexOutOfRange(t *testing.T) {
	cells := make([]dastypes.Cell, dastypes.ReconstructionThreshold)
	indices := make([]uint64, dastypes.ReconstructionThreshold)
	for i := range indices {
		indices[i] = uint64(i)
	}
	indices[0] = dastypes.CellsPerExtBlob // out of range
	if err := validateCellInputs(cells, indices); err == nil {
		t.Error("expected error for index out of range")
	}
}

func TestValidateCellInputsDuplicateIndex(t *testing.T) {
	cells := make([]dastypes.Cell, dastypes.ReconstructionThreshold)
	indices := make([]uint64, dastypes.ReconstructionThreshold)
	for i := range indices {
		indices[i] = uint64(i)
	}
	indices[1] = indices[0] // duplicate
	if err := validateCellInputs(cells, indices); err == nil {
		t.Error("expected error for duplicate index")
	}
}

func TestValidateCellInputsValid(t *testing.T) {
	cells := make([]dastypes.Cell, dastypes.ReconstructionThreshold)
	indices := make([]uint64, dastypes.ReconstructionThreshold)
	for i := range indices {
		indices[i] = uint64(i)
	}
	if err := validateCellInputs(cells, indices); err != nil {
		t.Fatalf("valid inputs: %v", err)
	}
}

// --- ReconstructPolynomial tests ---

func TestReconstructPolynomialMismatch(t *testing.T) {
	xs := []field.FieldElement{field.NewFieldElementFromUint64(1), field.NewFieldElementFromUint64(2)}
	ys := []field.FieldElement{field.NewFieldElementFromUint64(3)}
	_, err := ReconstructPolynomial(xs, ys, 2)
	if err == nil {
		t.Error("expected error for mismatched xs/ys")
	}
}

func TestReconstructPolynomialTooFewPoints(t *testing.T) {
	xs := []field.FieldElement{field.NewFieldElementFromUint64(1)}
	ys := []field.FieldElement{field.NewFieldElementFromUint64(2)}
	_, err := ReconstructPolynomial(xs, ys, 3)
	if err == nil {
		t.Error("expected error for too few points")
	}
}

func TestReconstructPolynomialConstant(t *testing.T) {
	// Polynomial f(x) = 5 (constant). Any 1 evaluation is enough.
	five := field.NewFieldElementFromUint64(5)
	xs := []field.FieldElement{field.NewFieldElementFromUint64(0)}
	ys := []field.FieldElement{five}
	coeffs, err := ReconstructPolynomial(xs, ys, 1)
	if err != nil {
		t.Fatalf("ReconstructPolynomial: %v", err)
	}
	if len(coeffs) != 1 {
		t.Fatalf("expected 1 coefficient, got %d", len(coeffs))
	}
	if !coeffs[0].Equal(five) {
		t.Errorf("coefficient = %v, want 5", coeffs[0].BigInt())
	}
}

func TestReconstructPolynomialLinear(t *testing.T) {
	// Polynomial f(x) = 2x + 3. Need 2 points.
	// f(0) = 3, f(1) = 5
	three := field.NewFieldElementFromUint64(3)
	five := field.NewFieldElementFromUint64(5)
	xs := []field.FieldElement{field.NewFieldElementFromUint64(0), field.NewFieldElementFromUint64(1)}
	ys := []field.FieldElement{three, five}
	coeffs, err := ReconstructPolynomial(xs, ys, 2)
	if err != nil {
		t.Fatalf("ReconstructPolynomial: %v", err)
	}
	if len(coeffs) != 2 {
		t.Fatalf("expected 2 coefficients, got %d", len(coeffs))
	}
	// coeffs[0] = 3 (constant), coeffs[1] = 2 (linear)
	if !coeffs[0].Equal(three) {
		t.Errorf("c0 = %v, want 3", coeffs[0].BigInt())
	}
	two := field.NewFieldElementFromUint64(2)
	if !coeffs[1].Equal(two) {
		t.Errorf("c1 = %v, want 2", coeffs[1].BigInt())
	}
}

func TestReconstructPolynomialQuadratic(t *testing.T) {
	// Polynomial f(x) = x^2 + 1. Need 3 points.
	// f(0) = 1, f(1) = 2, f(2) = 5
	xs := []field.FieldElement{
		field.NewFieldElementFromUint64(0),
		field.NewFieldElementFromUint64(1),
		field.NewFieldElementFromUint64(2),
	}
	ys := []field.FieldElement{
		field.NewFieldElementFromUint64(1),
		field.NewFieldElementFromUint64(2),
		field.NewFieldElementFromUint64(5),
	}
	coeffs, err := ReconstructPolynomial(xs, ys, 3)
	if err != nil {
		t.Fatalf("ReconstructPolynomial: %v", err)
	}
	if len(coeffs) != 3 {
		t.Fatalf("expected 3 coefficients, got %d", len(coeffs))
	}
	// coeffs = [1, 0, 1] for 1 + 0*x + 1*x^2
	if !coeffs[0].Equal(field.NewFieldElementFromUint64(1)) {
		t.Errorf("c0 = %v, want 1", coeffs[0].BigInt())
	}
	if !coeffs[1].Equal(field.FieldZero()) {
		t.Errorf("c1 = %v, want 0", coeffs[1].BigInt())
	}
	if !coeffs[2].Equal(field.NewFieldElementFromUint64(1)) {
		t.Errorf("c2 = %v, want 1", coeffs[2].BigInt())
	}
}

// --- evaluatePolynomial tests ---

func TestEvaluatePolynomialEmpty(t *testing.T) {
	result := evaluatePolynomial(nil, field.NewFieldElementFromUint64(5))
	if !result.IsZero() {
		t.Error("evaluating empty polynomial should return zero")
	}
}

func TestEvaluatePolynomialConstant(t *testing.T) {
	coeffs := []field.FieldElement{field.NewFieldElementFromUint64(7)}
	result := evaluatePolynomial(coeffs, field.NewFieldElementFromUint64(100))
	if !result.Equal(field.NewFieldElementFromUint64(7)) {
		t.Errorf("constant polynomial: got %v, want 7", result.BigInt())
	}
}

func TestEvaluatePolynomialLinear(t *testing.T) {
	// f(x) = 3 + 2x. f(5) = 3 + 10 = 13.
	coeffs := []field.FieldElement{
		field.NewFieldElementFromUint64(3),
		field.NewFieldElementFromUint64(2),
	}
	result := evaluatePolynomial(coeffs, field.NewFieldElementFromUint64(5))
	if !result.Equal(field.NewFieldElementFromUint64(13)) {
		t.Errorf("linear polynomial at x=5: got %v, want 13", result.BigInt())
	}
}

// --- ReconstructBlob tests ---

func TestReconstructBlobZeroCellsRoundtrip(t *testing.T) {
	// Zero-filled cells should produce a zero blob.
	cells := make([]dastypes.Cell, dastypes.ReconstructionThreshold)
	indices := make([]uint64, dastypes.ReconstructionThreshold)
	for i := range indices {
		indices[i] = uint64(i)
	}

	result, err := ReconstructBlob(cells, indices)
	if err != nil {
		t.Fatalf("ReconstructBlob: %v", err)
	}

	expectedSize := dastypes.FieldElementsPerBlob * dastypes.BytesPerFieldElement
	if len(result) != expectedSize {
		t.Fatalf("result size = %d, want %d", len(result), expectedSize)
	}

	for i, b := range result {
		if b != 0 {
			t.Fatalf("result[%d] = %d, want 0", i, b)
		}
	}
}

// --- RecoverCellsAndProofs tests ---

func TestRecoverCellsAndProofsValidation(t *testing.T) {
	// Too few cells.
	cells := make([]dastypes.Cell, 10)
	indices := make([]uint64, 10)
	for i := range indices {
		indices[i] = uint64(i)
	}
	_, _, err := RecoverCellsAndProofs(cells, indices)
	if err == nil {
		t.Error("expected error for too few cells")
	}
}

func TestRecoverCellsAndProofsZeroCells(t *testing.T) {
	// Use zero cells -- all zeros should reconstruct to all zeros.
	cells := make([]dastypes.Cell, dastypes.ReconstructionThreshold)
	indices := make([]uint64, dastypes.ReconstructionThreshold)
	for i := range indices {
		indices[i] = uint64(i)
	}

	allCells, allProofs, err := RecoverCellsAndProofs(cells, indices)
	if err != nil {
		t.Fatalf("RecoverCellsAndProofs: %v", err)
	}
	if len(allCells) != int(dastypes.CellsPerExtBlob) {
		t.Fatalf("expected %d cells, got %d", dastypes.CellsPerExtBlob, len(allCells))
	}
	if len(allProofs) != int(dastypes.CellsPerExtBlob) {
		t.Fatalf("expected %d proofs, got %d", dastypes.CellsPerExtBlob, len(allProofs))
	}
}

// --- RecoverMatrix tests ---

func TestRecoverMatrixInvalidBlobCount(t *testing.T) {
	_, err := RecoverMatrix(nil, -1)
	if err == nil {
		t.Error("expected error for negative blob count")
	}
}

func TestRecoverMatrixMissingRow(t *testing.T) {
	// 2-blob matrix but only provide entries for row 0.
	entries := make([]dastypes.MatrixEntry, dastypes.ReconstructionThreshold)
	for i := range entries {
		entries[i] = dastypes.MatrixEntry{RowIndex: 0, ColumnIndex: dastypes.ColumnIndex(i)}
	}
	_, err := RecoverMatrix(entries, 2)
	if err == nil {
		t.Error("expected error when row 1 has no entries")
	}
}

func TestRecoverMatrixSufficientEntries(t *testing.T) {
	// Single blob with exactly threshold entries.
	entries := make([]dastypes.MatrixEntry, dastypes.ReconstructionThreshold)
	for i := range entries {
		entries[i] = dastypes.MatrixEntry{RowIndex: 0, ColumnIndex: dastypes.ColumnIndex(i)}
	}
	result, err := RecoverMatrix(entries, 1)
	if err != nil {
		t.Fatalf("RecoverMatrix: %v", err)
	}
	if len(result) != int(dastypes.CellsPerExtBlob) {
		t.Fatalf("expected %d entries, got %d", dastypes.CellsPerExtBlob, len(result))
	}
}

// --- cellToFieldElements / fieldElementsToBytes roundtrip ---

func TestCellFieldElementsRoundtrip(t *testing.T) {
	var cell dastypes.Cell
	// Set some bytes in the cell.
	cell[31] = 42 // last byte of first field element
	cell[63] = 99 // last byte of second field element

	elems := cellToFieldElements(&cell)
	if len(elems) != dastypes.FieldElementsPerCell {
		t.Fatalf("expected %d elements, got %d", dastypes.FieldElementsPerCell, len(elems))
	}

	// Verify the first field element encodes 42.
	if !elems[0].Equal(field.NewFieldElementFromUint64(42)) {
		t.Errorf("elem[0] = %v, want 42", elems[0].BigInt())
	}
	// Second element encodes 99.
	if !elems[1].Equal(field.NewFieldElementFromUint64(99)) {
		t.Errorf("elem[1] = %v, want 99", elems[1].BigInt())
	}
}

func TestFieldElementsToBytesSize(t *testing.T) {
	elems := []field.FieldElement{
		field.NewFieldElementFromUint64(1),
		field.NewFieldElementFromUint64(2),
	}
	result := fieldElementsToBytes(elems, 64)
	if len(result) != 64 {
		t.Fatalf("expected 64 bytes, got %d", len(result))
	}
	// Verify first element is at bytes 0..31 and second at 32..63.
	if result[31] != 1 {
		t.Errorf("result[31] = %d, want 1 (first elem)", result[31])
	}
	if result[63] != 2 {
		t.Errorf("result[63] = %d, want 2 (second elem)", result[63])
	}
}
