package rpo

// constants.go defines local copies of das package constants.
// Must stay in sync with das/types.go.

const (
	// NumberOfColumns is the number of columns in the extended data matrix.
	NumberOfColumns = 128

	// SamplesPerSlot is the minimum number of samples for an honest node.
	SamplesPerSlot = 8

	// FieldElementsPerBlob is the number of field elements in a blob.
	FieldElementsPerBlob = 4096

	// BytesPerFieldElement is the byte size of a BLS scalar field element.
	BytesPerFieldElement = 32

	// MaxBlobCommitmentsPerBlock is the maximum blob commitments per block.
	MaxBlobCommitmentsPerBlock = 9
)
