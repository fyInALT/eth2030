package cell

// constants.go defines local copies of das package constants needed by this sub-package.
// These must stay in sync with das/types.go.

const (
	// NumberOfColumns is the number of columns in the extended data matrix.
	NumberOfColumns = 128 // from das/types.go

	// DataColumnSidecarSubnetCount is the number of subnets for cell gossip.
	DataColumnSidecarSubnetCount = 64 // from das/types.go

	// CustodyRequirement is the minimum number of custody groups.
	CustodyRequirement = 4 // from das/types.go

	// BytesPerCell is the byte size of a single cell.
	BytesPerCell = 64 * 32 // FieldElementsPerCell * BytesPerFieldElement, from das/types.go

	// MaxBlobCommitmentsPerBlock is the maximum blob commitments per block.
	MaxBlobCommitmentsPerBlock = 9 // from das/types.go

	// CellsPerExtBlob is the number of cells in an extended blob.
	CellsPerExtBlob = (2 * 4096) / 64 // FieldElementsPerExtBlob / FieldElementsPerCell = 128
)
