package das

// sampling_compat.go re-exports types, functions, and variables from
// das/sampling for backward compatibility with existing callers.

import "github.com/eth2030/eth2030/das/sampling"

// Error variables re-exported from das/sampling.
var (
	ErrInvalidCustodyCount = sampling.ErrInvalidCustodyCount
	ErrInvalidColumnIndex  = sampling.ErrInvalidColumnIndex
	ErrInvalidSidecar      = sampling.ErrInvalidSidecar
	ErrMismatchedLengths   = sampling.ErrMismatchedLengths

	ErrSchedClosed         = sampling.ErrSchedClosed
	ErrSchedSlotZero       = sampling.ErrSchedSlotZero
	ErrSchedRoundComplete  = sampling.ErrSchedRoundComplete
	ErrSchedColumnOOB      = sampling.ErrSchedColumnOOB
	ErrSchedQuotaExhausted = sampling.ErrSchedQuotaExhausted
	ErrSchedNoActiveRound  = sampling.ErrSchedNoActiveRound
	ErrSchedInvalidMode    = sampling.ErrSchedInvalidMode

	ErrColSamplingSlotZero       = sampling.ErrColSamplingSlotZero
	ErrColSamplingColumnOOB      = sampling.ErrColSamplingColumnOOB
	ErrColSamplingProofMismatch  = sampling.ErrColSamplingProofMismatch
	ErrColSamplingNotAssigned    = sampling.ErrColSamplingNotAssigned
	ErrColSamplingAlreadyTracked = sampling.ErrColSamplingAlreadyTracked

	ErrPeerSchedClosed      = sampling.ErrPeerSchedClosed
	ErrPeerSchedNoColumns   = sampling.ErrPeerSchedNoColumns
	ErrPeerSchedNoPeers     = sampling.ErrPeerSchedNoPeers
	ErrPeerSchedSlotUnknown = sampling.ErrPeerSchedSlotUnknown
)

// SamplingMode constants re-exported from das/sampling.
const (
	RegularSampling  = sampling.RegularSampling
	ExtendedSampling = sampling.ExtendedSampling
)

// DAVerdict constants re-exported from das/sampling.
const (
	VerdictPending     = sampling.VerdictPending
	VerdictAvailable   = sampling.VerdictAvailable
	VerdictUnavailable = sampling.VerdictUnavailable
)

// Type aliases re-exported from das/sampling.
type (
	SamplingMode          = sampling.SamplingMode
	SchedulerConfig       = sampling.SchedulerConfig
	SamplingRound         = sampling.SamplingRound
	SamplingStats         = sampling.SamplingStats
	SamplingScheduler     = sampling.SamplingScheduler
	ColumnSample          = sampling.ColumnSample
	ColumnAvailability    = sampling.ColumnAvailability
	ColumnSamplerConfig   = sampling.ColumnSamplerConfig
	ColumnSampler         = sampling.ColumnSampler
	DAVerdict             = sampling.DAVerdict
	PeerSamplingConfig    = sampling.PeerSamplingConfig
	SamplingPeerInfo      = sampling.SamplingPeerInfo
	PeerColumnAssignment  = sampling.PeerColumnAssignment
	PeerSamplingPlan      = sampling.PeerSamplingPlan
	SlotSamplingStatus    = sampling.SlotSamplingStatus
	PeerSamplingScheduler = sampling.PeerSamplingScheduler
)

// Function aliases re-exported from das/sampling.
var (
	GetCustodyGroups              = sampling.GetCustodyGroups
	ComputeColumnsForCustodyGroup = sampling.ComputeColumnsForCustodyGroup
	GetCustodyColumns             = sampling.GetCustodyColumns
	ShouldCustodyColumn           = sampling.ShouldCustodyColumn
	VerifyDataColumnSidecar       = sampling.VerifyDataColumnSidecar
	ColumnSubnet                  = sampling.ColumnSubnet
	DefaultSchedulerConfig        = sampling.DefaultSchedulerConfig
	NewSamplingScheduler          = sampling.NewSamplingScheduler
	DefaultColumnSamplerConfig    = sampling.DefaultColumnSamplerConfig
	NewColumnSampler              = sampling.NewColumnSampler
	DefaultPeerSamplingConfig     = sampling.DefaultPeerSamplingConfig
	NewPeerSamplingScheduler      = sampling.NewPeerSamplingScheduler
)
