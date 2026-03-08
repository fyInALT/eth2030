package sync

import (
	"github.com/eth2030/eth2030/sync/downloader"
	"github.com/eth2030/eth2030/sync/snap"
)

// Re-exported types from snap sub-package used by sync.go.
type (
	SnapProgress = snap.SnapProgress
	SnapSyncer   = snap.SnapSyncer
	SnapPeer     = snap.SnapPeer
	StateWriter  = snap.StateWriter
)

// Re-exported phase constants from snap sub-package.
const (
	PhaseIdle     = snap.PhaseIdle
	PhaseAccounts = snap.PhaseAccounts
	PhaseStorage  = snap.PhaseStorage
	PhaseBytecode = snap.PhaseBytecode
	PhaseHealing  = snap.PhaseHealing
	PhaseComplete = snap.PhaseComplete
)

// Re-exported constructors from snap sub-package.
var (
	NewSnapSyncer = snap.NewSnapSyncer
	SelectPivot   = snap.SelectPivot
)

// Re-exported types from downloader sub-package used by sync.go.
type (
	HeaderSource = downloader.HeaderSource
	BodySource   = downloader.BodySource
	HeaderData   = downloader.HeaderData
	BlockData    = downloader.BlockData
)
