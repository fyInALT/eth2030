package sync

// checkpoint_compat.go re-exports types from sync/checkpoint for backward compatibility.

import "github.com/eth2030/eth2030/sync/checkpoint"

// Checkpoint type aliases.
type (
	SyncState              = checkpoint.SyncState
	TrustedCheckpoint      = checkpoint.TrustedCheckpoint
	CheckpointSyncProgress = checkpoint.CheckpointSyncProgress
	HeaderRangeRequest     = checkpoint.HeaderRangeRequest
	CheckpointStoreConfig  = checkpoint.CheckpointStoreConfig
	CheckpointStore        = checkpoint.CheckpointStore
)

// Checkpoint constants.
const (
	StateCheckpointIdle                = checkpoint.StateCheckpointIdle
	StateCheckpointDownloadingHeaders  = checkpoint.StateCheckpointDownloadingHeaders
	StateCheckpointDownloadingBodies   = checkpoint.StateCheckpointDownloadingBodies
	StateCheckpointDownloadingReceipts = checkpoint.StateCheckpointDownloadingReceipts
	StateCheckpointProcessing          = checkpoint.StateCheckpointProcessing
	StateCheckpointComplete            = checkpoint.StateCheckpointComplete
)

// Checkpoint error variables.
var (
	ErrStoreCheckpointExists  = checkpoint.ErrStoreCheckpointExists
	ErrStoreCheckpointUnknown = checkpoint.ErrStoreCheckpointUnknown
	ErrStoreEmpty             = checkpoint.ErrStoreEmpty
	ErrStoreSyncActive        = checkpoint.ErrStoreSyncActive
	ErrStoreSyncInactive      = checkpoint.ErrStoreSyncInactive
	ErrStoreInvalidRange      = checkpoint.ErrStoreInvalidRange
	ErrStoreRangeOverlap      = checkpoint.ErrStoreRangeOverlap
	ErrStoreTooManyPending    = checkpoint.ErrStoreTooManyPending
)

// Checkpoint function wrappers.
func DefaultCheckpointStoreConfig() CheckpointStoreConfig {
	return checkpoint.DefaultCheckpointStoreConfig()
}
func NewCheckpointStore(config CheckpointStoreConfig) *CheckpointStore {
	return checkpoint.NewCheckpointStore(config)
}
