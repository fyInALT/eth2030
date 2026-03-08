package das

// gossip_compat.go re-exports types, functions, and variables from
// das/gossip for backward compatibility with existing callers.

import "github.com/eth2030/eth2030/das/gossip"

// Error variables re-exported from das/gossip.
var (
	ErrGossipHandlerClosed    = gossip.ErrGossipHandlerClosed
	ErrGossipCellDuplicate    = gossip.ErrGossipCellDuplicate
	ErrGossipCellValidation   = gossip.ErrGossipCellValidation
	ErrGossipCellNilMessage   = gossip.ErrGossipCellNilMessage
	ErrGossipBlobNotTracked   = gossip.ErrGossipBlobNotTracked
	ErrGossipBroadcastNoPeers = gossip.ErrGossipBroadcastNoPeers

	ErrGossipScoreNilPeer      = gossip.ErrGossipScoreNilPeer
	ErrGossipScorePeerNotFound = gossip.ErrGossipScorePeerNotFound
	ErrSidecarBuildNoCells     = gossip.ErrSidecarBuildNoCells
	ErrSidecarBuildMismatch    = gossip.ErrSidecarBuildMismatch
	ErrReconstructNotNeeded    = gossip.ErrReconstructNotNeeded
	ErrColumnAlreadyReceived   = gossip.ErrColumnAlreadyReceived

	ErrBuilderNilBlobs       = gossip.ErrBuilderNilBlobs
	ErrBuilderEmptyBlobs     = gossip.ErrBuilderEmptyBlobs
	ErrBuilderBlobTooLarge   = gossip.ErrBuilderBlobTooLarge
	ErrBuilderColumnOOB      = gossip.ErrBuilderColumnOOB
	ErrBuilderDuplicateCol   = gossip.ErrBuilderDuplicateCol
	ErrBuilderInvalidCell    = gossip.ErrBuilderInvalidCell
	ErrBuilderMsgNil         = gossip.ErrBuilderMsgNil
	ErrBuilderMsgColOOB      = gossip.ErrBuilderMsgColOOB
	ErrBuilderMsgBlobOOB     = gossip.ErrBuilderMsgBlobOOB
	ErrBuilderMsgDataInvalid = gossip.ErrBuilderMsgDataInvalid
	ErrBuilderAlreadySeen    = gossip.ErrBuilderAlreadySeen
)

// Type aliases re-exported from das/gossip.
type (
	CellGossipMessage       = gossip.CellGossipMessage
	CellValidator           = gossip.CellValidator
	SimpleCellValidator     = gossip.SimpleCellValidator
	CellGossipCallback      = gossip.CellGossipCallback
	CellGossipHandler       = gossip.CellGossipHandler
	GossipHandlerStats      = gossip.GossipHandlerStats
	CellGossipHandlerConfig = gossip.CellGossipHandlerConfig

	GossipScoreConfig     = gossip.GossipScoreConfig
	GossipScorer          = gossip.GossipScorer
	ReconstructionTrigger = gossip.ReconstructionTrigger

	ColumnBuilderConfig     = gossip.ColumnBuilderConfig
	BuiltColumn             = gossip.BuiltColumn
	ColumnGossipMessage     = gossip.ColumnGossipMessage
	CustodyAssignmentResult = gossip.CustodyAssignmentResult
	ColumnBuilder           = gossip.ColumnBuilder
)

// Function aliases re-exported from das/gossip.
var (
	NewSimpleCellValidator     = gossip.NewSimpleCellValidator
	NewCellGossipHandler       = gossip.NewCellGossipHandler
	ComputeCellHash            = gossip.ComputeCellHash
	DefaultGossipScoreConfig   = gossip.DefaultGossipScoreConfig
	NewGossipScorer            = gossip.NewGossipScorer
	BuildDataColumnSidecar     = gossip.BuildDataColumnSidecar
	ComputeSidecarHash         = gossip.ComputeSidecarHash
	NewReconstructionTrigger   = gossip.NewReconstructionTrigger
	VerifyGossipColumn         = gossip.VerifyGossipColumn
	DefaultColumnBuilderConfig = gossip.DefaultColumnBuilderConfig
	NewColumnBuilder           = gossip.NewColumnBuilder
)
