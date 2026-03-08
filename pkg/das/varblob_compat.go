package das

// varblob_compat.go re-exports types from das/varblob for backward compatibility.

import "github.com/eth2030/eth2030/das/varblob"

// VarBlob type aliases.
type (
	VarBlobConfig     = varblob.VarBlobConfig
	VarBlob           = varblob.VarBlob
	VarBlobTx         = varblob.VarBlobTx
	BlobConfig        = varblob.BlobConfig
	BlobScheduleEntry = varblob.BlobScheduleEntry
	BlobSchedule      = varblob.BlobSchedule
)

// VarBlob constants.
const (
	DefaultBlobSize  = varblob.DefaultBlobSize
	MinBlobSizeBytes = varblob.MinBlobSizeBytes
	MaxBlobSizeBytes = varblob.MaxBlobSizeBytes
)

// VarBlob error variables.
var (
	ErrVarBlobTooLarge     = varblob.ErrVarBlobTooLarge
	ErrVarBlobInvalidChunk = varblob.ErrVarBlobInvalidChunk
	ErrVarBlobDecodeShort  = varblob.ErrVarBlobDecodeShort
	ErrVarBlobDecodeLen    = varblob.ErrVarBlobDecodeLen
	ErrVarBlobEmptyData    = varblob.ErrVarBlobEmptyData
	ErrInvalidBlobConfig   = varblob.ErrInvalidBlobConfig
	ErrBlobCountOutOfRange = varblob.ErrBlobCountOutOfRange
	ErrBlobSizeOutOfRange  = varblob.ErrBlobSizeOutOfRange
	ErrNoScheduleEntries   = varblob.ErrNoScheduleEntries
	ErrScheduleNotSorted   = varblob.ErrScheduleNotSorted
)

// Variable blob schedule.
var DefaultBlobSchedule = varblob.DefaultBlobSchedule

// VarBlob function wrappers.
func DefaultVarBlobConfig() VarBlobConfig { return varblob.DefaultVarBlobConfig() }
func NewVarBlob(data []byte, chunkSize int) (*VarBlob, error) {
	return varblob.NewVarBlob(data, chunkSize)
}
func DecodeVarBlob(data []byte) (*VarBlob, error) { return varblob.DecodeVarBlob(data) }
func ValidateVarBlob(vb *VarBlob) error           { return varblob.ValidateVarBlob(vb) }
func ValidatePaddingProof(vb *VarBlob, dataLen int) error {
	return varblob.ValidatePaddingProof(vb, dataLen)
}
func EstimateVarBlobGas(blobSize, chunkSize int) uint64 {
	return varblob.EstimateVarBlobGas(blobSize, chunkSize)
}
func GetBlobConfigAtTime(schedule BlobSchedule, timestamp uint64) BlobConfig {
	return varblob.GetBlobConfigAtTime(schedule, timestamp)
}
func ValidateBlobCount(schedule BlobSchedule, timestamp uint64, blobCount uint64) error {
	return varblob.ValidateBlobCount(schedule, timestamp, blobCount)
}
func ValidateVariableBlobSize(schedule BlobSchedule, timestamp uint64, blobDataSize uint64) error {
	return varblob.ValidateVariableBlobSize(schedule, timestamp, blobDataSize)
}
