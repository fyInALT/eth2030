package blobs

// constants.go defines local copies of das package constants needed by this sub-package.
// These must stay in sync with das/variable_blobs.go.

const (
	// DefaultBlobSize is the standard blob size in bytes (128 KiB).
	// FieldElementsPerBlob(4096) * BytesPerFieldElement(32) = 131072
	DefaultBlobSize = 131072 // from das/variable_blobs.go
)
