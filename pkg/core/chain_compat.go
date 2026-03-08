package core

// chain_compat.go re-exports types and functions from core/chain for
// backward compatibility with callers inside the core/ package root.

import "github.com/eth2030/eth2030/core/chain"

// --- Blockchain types ---

// Blockchain is an alias for chain.Blockchain.
type Blockchain = chain.Blockchain

// TxLookupEntry is an alias for chain.TxLookupEntry.
type TxLookupEntry = chain.TxLookupEntry

// NewBlockchain creates a new blockchain.
var NewBlockchain = chain.NewBlockchain

// Blockchain error vars.
var (
	ErrNoGenesis     = chain.ErrNoGenesis
	ErrGenesisExists = chain.ErrGenesisExists
	ErrBlockNotFound = chain.ErrBlockNotFound
	ErrInvalidChain  = chain.ErrInvalidChain
	ErrFutureBlock2  = chain.ErrFutureBlock2
	ErrStateNotFound = chain.ErrStateNotFound
)

// --- ForkChoice types ---

// ForkChoice is an alias for chain.ForkChoice.
type ForkChoice = chain.ForkChoice

// NewForkChoice creates a new fork choice tracker.
var NewForkChoice = chain.NewForkChoice

// FindCommonAncestor finds the common ancestor of two chain heads.
var FindCommonAncestor = chain.FindCommonAncestor

// ForkChoice error vars.
var (
	ErrFinalizedBlockUnknown  = chain.ErrFinalizedBlockUnknown
	ErrSafeBlockUnknown       = chain.ErrSafeBlockUnknown
	ErrHeadBlockUnknown       = chain.ErrHeadBlockUnknown
	ErrReorgPastFinalized     = chain.ErrReorgPastFinalized
	ErrCommonAncestorNotFound = chain.ErrCommonAncestorNotFound
	ErrInvalidFinalizedChain  = chain.ErrInvalidFinalizedChain
	ErrInvalidSafeChain       = chain.ErrInvalidSafeChain
	ErrSafeNotFinalized       = chain.ErrSafeNotFinalized
)

// --- ChainReorgHandler types ---

// ChainReorgHandler is an alias for chain.ChainReorgHandler.
type ChainReorgHandler = chain.ChainReorgHandler

// ReorgConfig is an alias for chain.ReorgConfig.
type ReorgConfig = chain.ReorgConfig

// ReorgEvent is an alias for chain.ReorgEvent.
type ReorgEvent = chain.ReorgEvent

// NewChainReorgHandler creates a new chain reorg handler.
var NewChainReorgHandler = chain.NewChainReorgHandler

// DefaultReorgConfig returns sensible defaults for reorg handling.
var DefaultReorgConfig = chain.DefaultReorgConfig

// ChainReorg error vars.
var (
	ErrReorgTooDeep      = chain.ErrReorgTooDeep
	ErrReorgZeroHash     = chain.ErrReorgZeroHash
	ErrReorgUnknownBlock = chain.ErrReorgUnknownBlock
)

// --- ChainReader types ---

// ChainReader is an alias for chain.ChainReader.
type ChainReader = chain.ChainReader

// MemoryChain is an alias for chain.MemoryChain.
type MemoryChain = chain.MemoryChain

// ChainIterator is an alias for chain.ChainIterator.
type ChainIterator = chain.ChainIterator

// NewMemoryChain creates a new in-memory chain.
var NewMemoryChain = chain.NewMemoryChain

// NewChainIterator creates a new chain iterator.
var NewChainIterator = chain.NewChainIterator

// GetAncestor walks back from a given block to find an ancestor.
var GetAncestor = chain.GetAncestor

// GetTD returns the total difficulty for the given block.
var GetTD = chain.GetTD

// --- FullChainReader types ---

// FullChainReader is an alias for chain.FullChainReader.
type FullChainReader = chain.FullChainReader

// MemoryFullChain is an alias for chain.MemoryFullChain.
type MemoryFullChain = chain.MemoryFullChain

// NewMemoryFullChain creates a new in-memory full chain reader.
var NewMemoryFullChain = chain.NewMemoryFullChain

// --- HeaderChain types ---

// HeaderChain is an alias for chain.HeaderChain.
type HeaderChain = chain.HeaderChain

// NewHeaderChain creates a new header chain.
var NewHeaderChain = chain.NewHeaderChain

// HeaderChain error vars.
var (
	ErrKnownBlock    = chain.ErrKnownBlock
	ErrInsertStopped = chain.ErrInsertStopped
)

// --- HeaderVerifier types ---

// HeaderVerifier is an alias for chain.HeaderVerifier.
type HeaderVerifier = chain.HeaderVerifier

// NewHeaderVerifier creates a new header chain verifier.
var NewHeaderVerifier = chain.NewHeaderVerifier

// VerifyTimestampWindow checks a header's timestamp against a wall clock time.
var VerifyTimestampWindow = chain.VerifyTimestampWindow

// CalcGasLimitRange returns the min/max gas limit allowed for the next block.
var CalcGasLimitRange = chain.CalcGasLimitRange

// VerifyBaseFeeFromScratch computes and compares the expected base fee.
var VerifyBaseFeeFromScratch = chain.VerifyBaseFeeFromScratch

// HeaderVerifier error vars.
var (
	ErrTimestampNonMonotonic = chain.ErrTimestampNonMonotonic
	ErrHeaderChainBroken     = chain.ErrHeaderChainBroken
	ErrGasLimitJump          = chain.ErrGasLimitJump
	ErrBaseFeeComputation    = chain.ErrBaseFeeComputation
	ErrBlobGasComputation    = chain.ErrBlobGasComputation
	ErrDifficultyPostMerge   = chain.ErrDifficultyPostMerge
	ErrNoncePostMerge        = chain.ErrNoncePostMerge
	ErrUnclesPostMerge       = chain.ErrUnclesPostMerge
	ErrExtraDataOverflow     = chain.ErrExtraDataOverflow
	ErrBlockNumberGap        = chain.ErrBlockNumberGap
	ErrGasUsedExceedsLimit   = chain.ErrGasUsedExceedsLimit
	ErrBaseFeeNil            = chain.ErrBaseFeeNil
	ErrBlobFieldsMissing     = chain.ErrBlobFieldsMissing
	ErrCalldataFieldsMissing = chain.ErrCalldataFieldsMissing
)
