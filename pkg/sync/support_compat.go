package sync

// support_compat.go re-exports types from sync/support for backward compatibility.

import "github.com/eth2030/eth2030/sync/support"

// Support type aliases.
type (
	StageStatus         = support.StageStatus
	PipelineConfig      = support.PipelineConfig
	PipelineStage       = support.PipelineStage
	SyncPipeline        = support.SyncPipeline
	ProgressStage       = support.ProgressStage
	ProgressInfo        = support.ProgressInfo
	ProgressTracker     = support.ProgressTracker
	SchedulerPriority   = support.SchedulerPriority
	SchedulerHealTask   = support.SchedulerHealTask
	HealSchedulerConfig = support.HealSchedulerConfig
	ResourceBudget      = support.ResourceBudget
	ConcurrentHealer    = support.ConcurrentHealer
)

// Support constants.
const (
	StageStatusPending    = support.StageStatusPending
	StageStatusRunning    = support.StageStatusRunning
	StageStatusCompleted  = support.StageStatusCompleted
	StageStatusFailed     = support.StageStatusFailed
	PriorityCritical      = support.PriorityCritical
	PriorityUrgent        = support.PriorityUrgent
	PriorityNormal        = support.PriorityNormal
	PriorityBackground    = support.PriorityBackground
	StageProgressIdle     = support.StageProgressIdle
	StageProgressHeaders  = support.StageProgressHeaders
	StageProgressBodies   = support.StageProgressBodies
	StageProgressReceipts = support.StageProgressReceipts
	StageProgressState    = support.StageProgressState
	StageProgressBeacon   = support.StageProgressBeacon
	StageProgressSnap     = support.StageProgressSnap
	StageProgressComplete = support.StageProgressComplete
)

// Support error variables.
var (
	ErrPipelineStageNotFound  = support.ErrPipelineStageNotFound
	ErrPipelineStageDeps      = support.ErrPipelineStageDeps
	ErrPipelineStageActive    = support.ErrPipelineStageActive
	ErrPipelineRetryExhausted = support.ErrPipelineRetryExhausted
	ErrSchedulerFull          = support.ErrSchedulerFull
	ErrSchedulerClosed        = support.ErrSchedulerClosed
	ErrBudgetExceeded         = support.ErrBudgetExceeded
	ErrHealTaskNotFound       = support.ErrHealTaskNotFound
	ErrInvalidBudget          = support.ErrInvalidBudget
	ErrDuplicateTask          = support.ErrDuplicateTask
	ErrBudgetInsufficient     = support.ErrBudgetInsufficient
)

// Support function wrappers.
func DefaultPipelineConfig() PipelineConfig     { return support.DefaultPipelineConfig() }
func NewSyncPipeline(config PipelineConfig) *SyncPipeline { return support.NewSyncPipeline(config) }
func NewProgressTracker() *ProgressTracker      { return support.NewProgressTracker() }
func DefaultHealSchedulerConfig() HealSchedulerConfig { return support.DefaultHealSchedulerConfig() }
func NewResourceBudget(memLimit, bwLimit, maxPending int) *ResourceBudget {
	return support.NewResourceBudget(memLimit, bwLimit, maxPending)
}
func NewConcurrentHealer(cfg HealSchedulerConfig) *ConcurrentHealer {
	return support.NewConcurrentHealer(cfg)
}
