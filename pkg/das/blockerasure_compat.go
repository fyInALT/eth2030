package das

// blockerasure_compat.go re-exports types, functions, and variables from
// das/blockerasure for backward compatibility with existing callers.

import "github.com/eth2030/eth2030/das/blockerasure"

// Constants re-exported from das/blockerasure.
const (
	StandardDataShards   = blockerasure.StandardDataShards
	StandardParityShards = blockerasure.StandardParityShards
	DefaultMaxBlockSize  = blockerasure.DefaultMaxBlockSize
)

// Error variables re-exported from das/blockerasure.
var (
	ErrBlockErasureEmpty      = blockerasure.ErrBlockErasureEmpty
	ErrBlockErasureTooLarge   = blockerasure.ErrBlockErasureTooLarge
	ErrBlockErasureNilEncoder = blockerasure.ErrBlockErasureNilEncoder
	ErrBlockPieceInvalid      = blockerasure.ErrBlockPieceInvalid
	ErrBlockPieceHashMismatch = blockerasure.ErrBlockPieceHashMismatch
	ErrBlockPieceDuplicate    = blockerasure.ErrBlockPieceDuplicate
	ErrInsufficientPieces     = blockerasure.ErrInsufficientPieces
	ErrBlockReconstructFailed = blockerasure.ErrBlockReconstructFailed
)

// Type aliases re-exported from das/blockerasure.
type (
	BlockErasureConfig         = blockerasure.BlockErasureConfig
	BlockPiece                 = blockerasure.BlockPiece
	BlockErasureEncoder        = blockerasure.BlockErasureEncoder
	BlockErasureDecoder        = blockerasure.BlockErasureDecoder
	BlockAssemblyManager       = blockerasure.BlockAssemblyManager
	BlockAssemblyManagerConfig = blockerasure.BlockAssemblyManagerConfig
)

// Function aliases re-exported from das/blockerasure.
var (
	DefaultBlockErasureConfig         = blockerasure.DefaultBlockErasureConfig
	StandardBlockErasureConfig        = blockerasure.StandardBlockErasureConfig
	NewBlockErasureEncoder            = blockerasure.NewBlockErasureEncoder
	NewBlockErasureDecoder            = blockerasure.NewBlockErasureDecoder
	DefaultBlockAssemblyManagerConfig = blockerasure.DefaultBlockAssemblyManagerConfig
	NewBlockAssemblyManager           = blockerasure.NewBlockAssemblyManager
)
