package das

// reconstruction_compat.go re-exports types and functions from
// das/reconstruction for backward compatibility.

import "github.com/eth2030/eth2030/das/reconstruction"

// Error vars re-exported from das/reconstruction.
var (
	ErrInsufficientCells  = reconstruction.ErrInsufficientCells
	ErrInvalidCellIndex   = reconstruction.ErrInvalidCellIndex
	ErrDuplicateCellIndex = reconstruction.ErrDuplicateCellIndex

	ErrNilSample            = reconstruction.ErrNilSample
	ErrInvalidSampleIndex   = reconstruction.ErrInvalidSampleIndex
	ErrSampleBlobOutOfRange = reconstruction.ErrSampleBlobOutOfRange
	ErrNoSamplesForBlob     = reconstruction.ErrNoSamplesForBlob
	ErrReconstructionFailed = reconstruction.ErrReconstructionFailed

	ErrPipelineClosed        = reconstruction.ErrPipelineClosed
	ErrPipelineInvalidBlob   = reconstruction.ErrPipelineInvalidBlob
	ErrPipelineNoCells       = reconstruction.ErrPipelineNoCells
	ErrPipelineInsufficient  = reconstruction.ErrPipelineInsufficient
	ErrPipelineValidation    = reconstruction.ErrPipelineValidation
	ErrPipelineDuplicateCell = reconstruction.ErrPipelineDuplicateCell

	ErrReconstructorClosed   = reconstruction.ErrReconstructorClosed
	ErrBlobAlreadyComplete   = reconstruction.ErrBlobAlreadyComplete
	ErrInvalidSamplePayload  = reconstruction.ErrInvalidSamplePayload
	ErrCannotReconstruct     = reconstruction.ErrCannotReconstruct
	ErrErasureRecoveryFailed = reconstruction.ErrErasureRecoveryFailed
)

// Type aliases re-exported from das/reconstruction.
type (
	Sample                    = reconstruction.Sample
	ReconstructionMetrics     = reconstruction.ReconstructionMetrics
	BlobReconstructor         = reconstruction.BlobReconstructor
	ReconstructionStatus      = reconstruction.ReconstructionStatus
	ReconstructionPriority    = reconstruction.ReconstructionPriority
	BlobReconState            = reconstruction.BlobReconState
	CellCollector             = reconstruction.CellCollector
	ReconstructionScheduler   = reconstruction.ReconstructionScheduler
	ScheduleEntry             = reconstruction.ScheduleEntry
	ReconPipelineMetrics      = reconstruction.ReconPipelineMetrics
	ValidationStep            = reconstruction.ValidationStep
	ReconstructionPipeline    = reconstruction.ReconstructionPipeline
	SampleReconstructorConfig = reconstruction.SampleReconstructorConfig
	CellSample                = reconstruction.CellSample
	BlobProgress              = reconstruction.BlobProgress
	ReconstructorMetrics      = reconstruction.ReconstructorMetrics
	SampleReconstructor       = reconstruction.SampleReconstructor
)

// ReconstructionPriority constants re-exported from das/reconstruction.
const (
	PriorityLow      = reconstruction.PriorityLow
	PriorityNormal   = reconstruction.PriorityNormal
	PriorityHigh     = reconstruction.PriorityHigh
	PriorityCritical = reconstruction.PriorityCritical
)

// Function var aliases re-exported from das/reconstruction.
var (
	CanReconstruct                   = reconstruction.CanReconstruct
	ReconstructPolynomial            = reconstruction.ReconstructPolynomial
	ReconstructBlob                  = reconstruction.ReconstructBlob
	RecoverCellsAndProofs            = reconstruction.RecoverCellsAndProofs
	RecoverMatrix                    = reconstruction.RecoverMatrix
	ValidateSample                   = reconstruction.ValidateSample
	ValidateReconstructionInput      = reconstruction.ValidateReconstructionInput
	NewBlobReconstructor             = reconstruction.NewBlobReconstructor
	ReconstructWithErasure           = reconstruction.ReconstructWithErasure
	NewCellCollector                 = reconstruction.NewCellCollector
	NewReconstructionScheduler       = reconstruction.NewReconstructionScheduler
	NewValidationStep                = reconstruction.NewValidationStep
	NewReconstructionPipeline        = reconstruction.NewReconstructionPipeline
	DefaultSampleReconstructorConfig = reconstruction.DefaultSampleReconstructorConfig
	NewSampleReconstructor           = reconstruction.NewSampleReconstructor
)
