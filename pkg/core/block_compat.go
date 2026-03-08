package core

// block_compat.go re-exports types and functions from core/block for
// backward compatibility with callers inside the core/ package root.

import "github.com/eth2030/eth2030/core/block"

// --- Type aliases ---

// BlockValidator is an alias for block.BlockValidator.
type BlockValidator = block.BlockValidator

// BlockExecutor is an alias for block.BlockExecutor.
type BlockExecutor = block.BlockExecutor

// BlockBuilder is an alias for block.BlockBuilder.
type BlockBuilder = block.BlockBuilder

// TxPoolReader is an alias for block.TxPoolReader.
type TxPoolReader = block.TxPoolReader

// BuildBlockAttributes is an alias for block.BuildBlockAttributes.
type BuildBlockAttributes = block.BuildBlockAttributes

// BlockExecutionResult is an alias for block.BlockExecutionResult.
type BlockExecutionResult = block.BlockExecutionResult

// ExecutorConfig is an alias for block.ExecutorConfig.
type ExecutorConfig = block.ExecutorConfig

// ExecutorStats is an alias for block.ExecutorStats.
type ExecutorStats = block.ExecutorStats

// --- Constructors ---

// NewBlockValidator creates a new block validator.
var NewBlockValidator = block.NewBlockValidator

// NewBlockExecutor creates a new block executor.
var NewBlockExecutor = block.NewBlockExecutor

// NewBlockBuilder creates a new block builder.
var NewBlockBuilder = block.NewBlockBuilder

// DefaultExecutorConfig returns sensible defaults for the executor.
var DefaultExecutorConfig = block.DefaultExecutorConfig

// --- Constants (must be re-declared, not aliased) ---

const (
	// MaxExtraDataSize is the maximum allowed extra data in a block header.
	MaxExtraDataSize = block.MaxExtraDataSize

	// GasLimitBoundDivisor is the divisor for max gas limit change per block.
	GasLimitBoundDivisor = block.GasLimitBoundDivisor

	// MinGasLimit is the minimum gas limit.
	MinGasLimit = block.MinGasLimit

	// MaxGasLimit is the maximum gas limit.
	MaxGasLimit = block.MaxGasLimit
)

// --- Variables ---

// EmptyUncleHash is the hash of an empty uncle list.
var EmptyUncleHash = block.EmptyUncleHash

// --- Errors ---

var (
	ErrUnknownParent          = block.ErrUnknownParent
	ErrFutureBlock            = block.ErrFutureBlock
	ErrInvalidNumber          = block.ErrInvalidNumber
	ErrInvalidGasLimit        = block.ErrInvalidGasLimit
	ErrInvalidGasUsed         = block.ErrInvalidGasUsed
	ErrInvalidTimestamp       = block.ErrInvalidTimestamp
	ErrExtraDataTooLong       = block.ErrExtraDataTooLong
	ErrInvalidBaseFee         = block.ErrInvalidBaseFee
	ErrInvalidDifficulty      = block.ErrInvalidDifficulty
	ErrInvalidUncleHash       = block.ErrInvalidUncleHash
	ErrInvalidNonce           = block.ErrInvalidNonce
	ErrInvalidRequestHash     = block.ErrInvalidRequestHash
	ErrInvalidBlockAccessList = block.ErrInvalidBlockAccessList
	ErrMissingBlockAccessList = block.ErrMissingBlockAccessList
	ErrInvalidStateRoot       = block.ErrInvalidStateRoot
	ErrInvalidReceiptRoot     = block.ErrInvalidReceiptRoot
	ErrInvalidTxRoot          = block.ErrInvalidTxRoot
	ErrInvalidGasUsedTotal    = block.ErrInvalidGasUsedTotal
	ErrInvalidCalldataGas     = block.ErrInvalidCalldataGas
	ErrBlobGasLimitExceeded   = block.ErrBlobGasLimitExceeded
	ErrInvalidBlobHash        = block.ErrInvalidBlobHash

	// Executor errors.
	ErrNilHeader       = block.ErrNilHeader
	ErrNilTransaction  = block.ErrNilTransaction
	ErrGasExceeded     = block.ErrGasExceeded
	ErrNoTransactions  = block.ErrNoTransactions
	ErrExecutionFailed = block.ErrExecutionFailed
	ErrGasMismatch     = block.ErrGasMismatch
	ErrTxCountMismatch = block.ErrTxCountMismatch
	ErrRootMismatch    = block.ErrRootMismatch
	ErrReceiptMismatch = block.ErrReceiptMismatch
)

// --- Exported functions ---

// ValidateCalldataGas validates the EIP-7706 calldata gas header fields.
var ValidateCalldataGas = block.ValidateCalldataGas

// ComputeReceiptsRoot computes the receipt trie root from a list of receipts.
var ComputeReceiptsRoot = block.ComputeReceiptsRoot

// DeriveTxsRoot is the exported version of deriveTxsRoot.
var DeriveTxsRoot = block.DeriveTxsRoot

// DeriveReceiptsRoot is the exported version of deriveReceiptsRoot.
var DeriveReceiptsRoot = block.DeriveReceiptsRoot

// DeriveWithdrawalsRoot is the exported version of deriveWithdrawalsRoot.
var DeriveWithdrawalsRoot = block.DeriveWithdrawalsRoot
