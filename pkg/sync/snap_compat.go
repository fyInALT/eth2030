package sync

// snap_compat.go re-exports types from sync/snap for backward compatibility.

import (
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/sync/snap"
)

// Snap type aliases.
type (
	AccountData          = snap.AccountData
	StorageData          = snap.StorageData
	BytecodeData         = snap.BytecodeData
	AccountRangeRequest  = snap.AccountRangeRequest
	AccountRangeResponse = snap.AccountRangeResponse
	StorageRangeRequest  = snap.StorageRangeRequest
	StorageRangeResponse = snap.StorageRangeResponse
	BytecodeRequest      = snap.BytecodeRequest
	BytecodeResponse     = snap.BytecodeResponse
	SnapPeer             = snap.SnapPeer
	StateWriter          = snap.StateWriter
	SnapProgress         = snap.SnapProgress
	SnapSyncer           = snap.SnapSyncer
	SnapSync             = snap.SnapSync
	SnapSyncPeer         = snap.SnapSyncPeer
	SnapSyncProgress     = snap.SnapSyncProgress
)

// Snap constants.
const (
	MaxAccountRange  = snap.MaxAccountRange
	MaxStorageRange  = snap.MaxStorageRange
	MaxBytecodeItems = snap.MaxBytecodeItems
	PivotOffset      = snap.PivotOffset
	MinPivotBlock    = snap.MinPivotBlock
	MaxHealNodes     = snap.MaxHealNodes

	PhaseIdle     = snap.PhaseIdle
	PhaseAccounts = snap.PhaseAccounts
	PhaseStorage  = snap.PhaseStorage
	PhaseBytecode = snap.PhaseBytecode
	PhaseHealing  = snap.PhaseHealing
	PhaseComplete = snap.PhaseComplete

	SnapSyncSoftByteLimit = snap.SnapSyncSoftByteLimit
)

// Snap error variables.
var (
	ErrSnapAlreadyRunning = snap.ErrSnapAlreadyRunning
	ErrSnapCancelled      = snap.ErrSnapCancelled
	ErrNoPivotBlock       = snap.ErrNoPivotBlock
	ErrBadStateRoot       = snap.ErrBadStateRoot
	ErrBadAccountProof    = snap.ErrBadAccountProof
	ErrBadStorageProof    = snap.ErrBadStorageProof
	ErrBadBytecode        = snap.ErrBadBytecode
	ErrNoSnapPeer         = snap.ErrNoSnapPeer
	ErrRangeExhausted     = snap.ErrRangeExhausted
)

// Snap function wrappers.
func PhaseName(phase uint32) string                 { return snap.PhaseName(phase) }
func NewSnapSyncer(writer StateWriter) *SnapSyncer  { return snap.NewSnapSyncer(writer) }
func SelectPivot(headNumber uint64) (uint64, error) { return snap.SelectPivot(headNumber) }
func NewSnapSync(writer StateWriter) *SnapSync      { return snap.NewSnapSync(writer) }
func SelectSnapPivot(headBlock uint64) (uint64, error) {
	return snap.SelectSnapPivot(headBlock)
}
func VerifyAccountRange(root types.Hash, accounts []AccountData, proof [][]byte) error {
	return snap.VerifyAccountRange(root, accounts, proof)
}
func SplitAccountRange(origin, limit types.Hash, n int) []AccountRangeRequest {
	return snap.SplitAccountRange(origin, limit, n)
}
func MergeAccountRanges(a, b []AccountData) []AccountData {
	return snap.MergeAccountRanges(a, b)
}
func DetectHealingNeeded(writer StateWriter, root types.Hash) bool {
	return snap.DetectHealingNeeded(writer, root)
}
func SnapSyncPhaseName(phase uint32) string { return snap.SnapSyncPhaseName(phase) }
func VerifyAccountRangeProof(root types.Hash, accounts []AccountData, proof [][]byte) error {
	return snap.VerifyAccountRangeProof(root, accounts, proof)
}
func VerifyStorageRangeProof(root types.Hash, slots []StorageData, proof [][]byte) error {
	return snap.VerifyStorageRangeProof(root, slots, proof)
}
func SplitSnapSyncRange(origin, limit types.Hash, n int) []AccountRangeRequest {
	return snap.SplitSnapSyncRange(origin, limit, n)
}
