package sync

// statesync_compat.go re-exports types from sync/statesync for backward compatibility.

import "github.com/eth2030/eth2030/sync/statesync"

// StateSyncScheduler type aliases.
type (
	StateSyncProgress  = statesync.StateSyncProgress
	ProgressCallback   = statesync.ProgressCallback
	StateSyncScheduler = statesync.StateSyncScheduler
	StateSyncConfig    = statesync.StateSyncConfig
	StateAccount       = statesync.StateAccount
	StateRangeResponse = statesync.StateRangeResponse
	SSMProgress        = statesync.SSMProgress
	StateSyncManager   = statesync.StateSyncManager
	StateSynCheckpoint = statesync.StateSynCheckpoint
	StateSynProgress   = statesync.StateSynProgress
	StateSyn           = statesync.StateSyn
)

// StateSyncScheduler error variables.
var (
	ErrStateSyncRunning     = statesync.ErrStateSyncRunning
	ErrStateSyncStopped     = statesync.ErrStateSyncStopped
	ErrPivotTooOld          = statesync.ErrPivotTooOld
	ErrTaskFailed           = statesync.ErrTaskFailed
	ErrAllTasksExhausted    = statesync.ErrAllTasksExhausted
	ErrSSMAlreadySyncing    = statesync.ErrSSMAlreadySyncing
	ErrSSMNotSyncing        = statesync.ErrSSMNotSyncing
	ErrSSMPaused            = statesync.ErrSSMPaused
	ErrSSMInvalidProof      = statesync.ErrSSMInvalidProof
	ErrSSMEmptyRange        = statesync.ErrSSMEmptyRange
	ErrSSMRetryExhausted    = statesync.ErrSSMRetryExhausted
	ErrStateSynRunning      = statesync.ErrStateSynRunning
	ErrStateSynCancelled    = statesync.ErrStateSynCancelled
	ErrStateSynVerifyFailed = statesync.ErrStateSynVerifyFailed
	ErrStateSynRetryLimit   = statesync.ErrStateSynRetryLimit
	ErrStateSynNoPeer       = statesync.ErrStateSynNoPeer
)

// StateSyncScheduler function wrappers.
func SyncPhaseName(phase uint32) string { return statesync.SyncPhaseName(phase) }
func NewStateSyncScheduler(writer statesync.StateWriter, cb ProgressCallback) *StateSyncScheduler {
	return statesync.NewStateSyncScheduler(writer, cb)
}
func DefaultStateSyncConfig() *StateSyncConfig { return statesync.DefaultStateSyncConfig() }
func NewStateSyncManager(config *StateSyncConfig) *StateSyncManager {
	return statesync.NewStateSyncManager(config)
}
func DecodeStateSynCheckpoint(data []byte) (*StateSynCheckpoint, error) {
	return statesync.DecodeStateSynCheckpoint(data)
}
func StateSynPhaseName(phase uint32) string { return statesync.StateSynPhaseName(phase) }
func NewStateSyn(writer statesync.StateWriter) *StateSyn {
	return statesync.NewStateSyn(writer)
}
