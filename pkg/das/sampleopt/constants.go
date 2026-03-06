package sampleopt

// Constants copied from das/types.go to avoid circular import.
const (
	NumberOfColumns    = 128  // from das/types.go
	SamplesPerSlot     = 8    // from das/types.go
	BytesPerFieldElement = 32 // from das/types.go
	FieldElementsPerCell = 64 // from das/types.go
	BytesPerCell         = FieldElementsPerCell * BytesPerFieldElement // 2048
)
