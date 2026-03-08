package sampling

import (
	"testing"

	"github.com/eth2030/eth2030/das/dastypes"
)

// --- GetCustodyGroups tests ---

func TestGetCustodyGroupsZeroCount(t *testing.T) {
	nodeID := [32]byte{0x01}
	groups, err := GetCustodyGroups(nodeID, 0)
	if err != nil {
		t.Fatalf("GetCustodyGroups(0): %v", err)
	}
	if len(groups) != 0 {
		t.Fatalf("expected 0 groups, got %d", len(groups))
	}
}

func TestGetCustodyGroupsOverflowWrap(t *testing.T) {
	// Use a node ID close to 2^256 - 1 to test overflow wrapping.
	var nodeID [32]byte
	for i := range nodeID {
		nodeID[i] = 0xff
	}
	groups, err := GetCustodyGroups(nodeID, dastypes.CustodyRequirement)
	if err != nil {
		t.Fatalf("GetCustodyGroups near max: %v", err)
	}
	if len(groups) != int(dastypes.CustodyRequirement) {
		t.Fatalf("expected %d groups, got %d", dastypes.CustodyRequirement, len(groups))
	}
	// Verify sorted and unique.
	seen := make(map[dastypes.CustodyGroup]bool)
	for i, g := range groups {
		if uint64(g) >= dastypes.NumberOfCustodyGroups {
			t.Errorf("group %d out of range", g)
		}
		if seen[g] {
			t.Errorf("duplicate group %d", g)
		}
		seen[g] = true
		if i > 0 && groups[i] <= groups[i-1] {
			t.Errorf("groups not sorted at index %d", i)
		}
	}
}

// --- ComputeColumnsForCustodyGroup tests ---

func TestComputeColumnsForCustodyGroupAllGroups(t *testing.T) {
	// Verify that each group maps to the correct column.
	// Since NumberOfColumns == NumberOfCustodyGroups, each group gets 1 column.
	for g := uint64(0); g < dastypes.NumberOfCustodyGroups; g++ {
		cols, err := ComputeColumnsForCustodyGroup(dastypes.CustodyGroup(g))
		if err != nil {
			t.Fatalf("group %d: %v", g, err)
		}
		if len(cols) != 1 {
			t.Fatalf("group %d: expected 1 column, got %d", g, len(cols))
		}
		if uint64(cols[0]) != g {
			t.Errorf("group %d: expected column %d, got %d", g, g, cols[0])
		}
	}
}

// --- GetCustodyColumns tests ---

func TestGetCustodyColumnsInvalidCount(t *testing.T) {
	nodeID := [32]byte{}
	_, err := GetCustodyColumns(nodeID, dastypes.NumberOfCustodyGroups+1)
	if err != ErrInvalidCustodyCount {
		t.Fatalf("expected ErrInvalidCustodyCount, got %v", err)
	}
}

func TestGetCustodyColumnsAll(t *testing.T) {
	nodeID := [32]byte{0x42}
	columns, err := GetCustodyColumns(nodeID, dastypes.NumberOfCustodyGroups)
	if err != nil {
		t.Fatalf("GetCustodyColumns(all): %v", err)
	}
	if len(columns) != int(dastypes.NumberOfColumns) {
		t.Fatalf("expected %d columns, got %d", dastypes.NumberOfColumns, len(columns))
	}
	// All columns should be present and sorted.
	for i := 0; i < int(dastypes.NumberOfColumns); i++ {
		if columns[i] != dastypes.ColumnIndex(i) {
			t.Errorf("columns[%d] = %d, want %d", i, columns[i], i)
		}
	}
}

// --- ShouldCustodyColumn tests ---

func TestShouldCustodyColumnEmpty(t *testing.T) {
	if ShouldCustodyColumn(0, nil) {
		t.Error("should not custody any column with nil set")
	}
	if ShouldCustodyColumn(0, []dastypes.ColumnIndex{}) {
		t.Error("should not custody any column with empty set")
	}
}

func TestShouldCustodyColumnAllColumns(t *testing.T) {
	allCols := make([]dastypes.ColumnIndex, dastypes.NumberOfColumns)
	for i := range allCols {
		allCols[i] = dastypes.ColumnIndex(i)
	}
	for i := uint64(0); i < dastypes.NumberOfColumns; i++ {
		if !ShouldCustodyColumn(dastypes.ColumnIndex(i), allCols) {
			t.Errorf("should custody column %d with all columns", i)
		}
	}
}

// --- VerifyDataColumnSidecar tests ---

func TestVerifyDataColumnSidecarMaxBlobs(t *testing.T) {
	// Valid sidecar with maximum blobs per block.
	cells := make([]dastypes.Cell, dastypes.MaxBlobCommitmentsPerBlock)
	commits := make([]dastypes.KZGCommitment, dastypes.MaxBlobCommitmentsPerBlock)
	proofs := make([]dastypes.KZGProof, dastypes.MaxBlobCommitmentsPerBlock)
	sidecar := &dastypes.DataColumnSidecar{
		Index:          0,
		Column:         cells,
		KZGCommitments: commits,
		KZGProofs:      proofs,
	}
	if err := VerifyDataColumnSidecar(sidecar); err != nil {
		t.Fatalf("valid max-blob sidecar failed: %v", err)
	}
}

func TestVerifyDataColumnSidecarMismatchedProofs(t *testing.T) {
	sidecar := &dastypes.DataColumnSidecar{
		Index:          0,
		Column:         []dastypes.Cell{{}, {}},
		KZGCommitments: []dastypes.KZGCommitment{{}, {}},
		KZGProofs:      []dastypes.KZGProof{{}}, // only 1 proof for 2 cells
	}
	if err := VerifyDataColumnSidecar(sidecar); err == nil {
		t.Error("expected error for mismatched proof count")
	}
}

// --- ColumnSubnet tests ---

func TestColumnSubnetAllColumns(t *testing.T) {
	// Verify subnet assignment for all columns.
	for i := uint64(0); i < dastypes.NumberOfColumns; i++ {
		subnet := ColumnSubnet(dastypes.ColumnIndex(i))
		expected := dastypes.SubnetID(i % dastypes.DataColumnSidecarSubnetCount)
		if subnet != expected {
			t.Errorf("ColumnSubnet(%d) = %d, want %d", i, subnet, expected)
		}
	}
}
