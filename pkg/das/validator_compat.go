package das

// validator_compat.go re-exports types, functions, and variables from
// das/validator for backward compatibility with existing callers.

import "github.com/eth2030/eth2030/das/validator"

// Error variables re-exported from das/validator.
var (
	ErrInvalidColumnIdx    = validator.ErrInvalidColumnIdx
	ErrInvalidCustodyGroup = validator.ErrInvalidCustodyGroup
	ErrDataUnavailable     = validator.ErrDataUnavailable
	ErrInvalidColumnProof  = validator.ErrInvalidColumnProof

	ErrBlobValidateNil        = validator.ErrBlobValidateNil
	ErrBlobValidateEmpty      = validator.ErrBlobValidateEmpty
	ErrBlobValidateSizeMax    = validator.ErrBlobValidateSizeMax
	ErrBlobValidateSizeMin    = validator.ErrBlobValidateSizeMin
	ErrBlobValidateFormat     = validator.ErrBlobValidateFormat
	ErrBlobValidateCommitment = validator.ErrBlobValidateCommitment
	ErrBlobValidateExpiry     = validator.ErrBlobValidateExpiry
	ErrBlobValidateNoRules    = validator.ErrBlobValidateNoRules

	ErrCellDataEmpty        = validator.ErrCellDataEmpty
	ErrCellDataSize         = validator.ErrCellDataSize
	ErrCellProofInvalid     = validator.ErrCellProofInvalid
	ErrCellColumnOutOfRange = validator.ErrCellColumnOutOfRange
	ErrCellRowOutOfRange    = validator.ErrCellRowOutOfRange
	ErrCellBatchEmpty       = validator.ErrCellBatchEmpty
	ErrCellReconstructFail  = validator.ErrCellReconstructFail
	ErrCellDuplicateIndex   = validator.ErrCellDuplicateIndex
	ErrCellInsufficientData = validator.ErrCellInsufficientData

	ErrL2ValidatorChainNotRegistered = validator.ErrL2ValidatorChainNotRegistered
	ErrL2ValidatorChainAlreadyExists = validator.ErrL2ValidatorChainAlreadyExists
	ErrL2ValidatorMaxChainsReached   = validator.ErrL2ValidatorMaxChainsReached
	ErrL2ValidatorInvalidCommitment  = validator.ErrL2ValidatorInvalidCommitment
	ErrL2ValidatorEmptyData          = validator.ErrL2ValidatorEmptyData
	ErrL2ValidatorDataTooLarge       = validator.ErrL2ValidatorDataTooLarge
	ErrL2ValidatorInvalidChainID     = validator.ErrL2ValidatorInvalidChainID

	ErrValidatorStopped = validator.ErrValidatorStopped
	ErrProofTimeout     = validator.ErrProofTimeout
	ErrNilProof         = validator.ErrNilProof
	ErrQueueFull        = validator.ErrQueueFull
)

// ProofPriority constants re-exported from das/validator.
const (
	PriorityCustody = validator.PriorityCustody
	PriorityRandom  = validator.PriorityRandom
)

// Type aliases re-exported from das/validator.
type (
	DAValidatorConfig     = validator.DAValidatorConfig
	DAValidator           = validator.DAValidator
	BlobMeta              = validator.BlobMeta
	BlobValidationError   = validator.BlobValidationError
	BlobValidationResult  = validator.BlobValidationResult
	BlobValidationRule    = validator.BlobValidationRule
	SizeRule              = validator.SizeRule
	FormatRule            = validator.FormatRule
	CommitmentRule        = validator.CommitmentRule
	ExpiryRule            = validator.ExpiryRule
	BlobValidationCache   = validator.BlobValidationCache
	BlobValidator         = validator.BlobValidator
	DataCell              = validator.DataCell
	BatchValidationResult = validator.BatchValidationResult
	CellValidatorConfig   = validator.CellValidatorConfig
	DataCellValidator     = validator.DataCellValidator
	CellReconstructor     = validator.CellReconstructor
	L2ChainConfig         = validator.L2ChainConfig
	L2DataReceipt         = validator.L2DataReceipt
	L2ChainMetrics        = validator.L2ChainMetrics
	L2DataValidator       = validator.L2DataValidator
	ProofPriority         = validator.ProofPriority
	DASProof              = validator.DASProof
	ValidationResult      = validator.ValidationResult
	AsyncValidatorConfig  = validator.AsyncValidatorConfig
	ValidatorMetrics      = validator.ValidatorMetrics
	AsyncValidator        = validator.AsyncValidator
)

// Function aliases re-exported from das/validator.
var (
	DefaultDAValidatorConfig    = validator.DefaultDAValidatorConfig
	NewDAValidator              = validator.NewDAValidator
	ComputeColumnProof          = validator.ComputeColumnProof
	DefaultBlobValidator        = validator.DefaultBlobValidator
	NewBlobValidator            = validator.NewBlobValidator
	NewBlobValidationCache      = validator.NewBlobValidationCache
	DefaultCellValidatorConfig  = validator.DefaultCellValidatorConfig
	NewDataCellValidator        = validator.NewDataCellValidator
	ComputeCellProof            = validator.ComputeCellProof
	NewCellReconstructor        = validator.NewCellReconstructor
	NewL2DataValidator          = validator.NewL2DataValidator
	DefaultAsyncValidatorConfig = validator.DefaultAsyncValidatorConfig
	NewAsyncValidator           = validator.NewAsyncValidator
)
