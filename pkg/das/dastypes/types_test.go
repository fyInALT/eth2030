package dastypes

import "testing"

func TestConstants(t *testing.T) {
	if NumberOfColumns != 128 {
		t.Errorf("NumberOfColumns = %d, want 128", NumberOfColumns)
	}
	if BytesPerCell != FieldElementsPerCell*BytesPerFieldElement {
		t.Errorf("BytesPerCell = %d, want %d", BytesPerCell, FieldElementsPerCell*BytesPerFieldElement)
	}
	if CellsPerExtBlob != FieldElementsPerExtBlob/FieldElementsPerCell {
		t.Errorf("CellsPerExtBlob = %d, want %d", CellsPerExtBlob, FieldElementsPerExtBlob/FieldElementsPerCell)
	}
	if ReconstructionThreshold != NumberOfColumns/2 {
		t.Errorf("ReconstructionThreshold = %d, want %d", ReconstructionThreshold, NumberOfColumns/2)
	}
}

func TestCellSize(t *testing.T) {
	var c Cell
	if len(c) != BytesPerCell {
		t.Errorf("Cell size = %d, want %d", len(c), BytesPerCell)
	}
}

func TestKZGTypes(t *testing.T) {
	var commitment KZGCommitment
	var proof KZGProof
	if len(commitment) != 48 {
		t.Errorf("KZGCommitment size = %d, want 48", len(commitment))
	}
	if len(proof) != 48 {
		t.Errorf("KZGProof size = %d, want 48", len(proof))
	}
}
