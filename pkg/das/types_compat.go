package das

// types_compat.go re-exports types from das/dastypes for backward compatibility.

import "github.com/eth2030/eth2030/das/dastypes"

// Constants.
const (
	NumberOfColumns              = dastypes.NumberOfColumns
	NumberOfCustodyGroups        = dastypes.NumberOfCustodyGroups
	CustodyRequirement           = dastypes.CustodyRequirement
	SamplesPerSlot               = dastypes.SamplesPerSlot
	DataColumnSidecarSubnetCount = dastypes.DataColumnSidecarSubnetCount
	FieldElementsPerBlob         = dastypes.FieldElementsPerBlob
	FieldElementsPerExtBlob      = dastypes.FieldElementsPerExtBlob
	FieldElementsPerCell         = dastypes.FieldElementsPerCell
	BytesPerFieldElement         = dastypes.BytesPerFieldElement
	BytesPerCell                 = dastypes.BytesPerCell
	CellsPerExtBlob              = dastypes.CellsPerExtBlob
	MaxBlobCommitmentsPerBlock   = dastypes.MaxBlobCommitmentsPerBlock
	ReconstructionThreshold      = dastypes.ReconstructionThreshold
)

// Type aliases.
type (
	SubnetID          = dastypes.SubnetID
	CustodyGroup      = dastypes.CustodyGroup
	ColumnIndex       = dastypes.ColumnIndex
	RowIndex          = dastypes.RowIndex
	Cell              = dastypes.Cell
	KZGCommitment     = dastypes.KZGCommitment
	KZGProof          = dastypes.KZGProof
	DataColumn        = dastypes.DataColumn
	DataColumnSidecar = dastypes.DataColumnSidecar
	MatrixEntry       = dastypes.MatrixEntry
	STARKCommitment   = dastypes.STARKCommitment
)
