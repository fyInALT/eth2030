package sync

// inserter_compat.go re-exports types from sync/inserter for backward compatibility.

import "github.com/eth2030/eth2030/sync/inserter"

// Inserter type aliases.
type (
	ChainInserterConfig   = inserter.ChainInserterConfig
	CIMetrics             = inserter.CIMetrics
	CIProgress            = inserter.CIProgress
	BlockExecutor         = inserter.BlockExecutor
	BlockCommitter        = inserter.BlockCommitter
	ChainInserter         = inserter.ChainInserter
	BlockProcessorConfig  = inserter.BlockProcessorConfig
	BlockProcessorMetrics = inserter.BlockProcessorMetrics
	ReceiptHasher         = inserter.ReceiptHasher
	StateExecutor         = inserter.StateExecutor
	AncestorLookup        = inserter.AncestorLookup
	BlockProcessor        = inserter.BlockProcessor
)

// Inserter error variables.
var (
	ErrCIClosedState     = inserter.ErrCIClosedState
	ErrCIRunning         = inserter.ErrCIRunning
	ErrCIStateRoot       = inserter.ErrCIStateRoot
	ErrCIReceiptRoot     = inserter.ErrCIReceiptRoot
	ErrCILogsBloom       = inserter.ErrCILogsBloom
	ErrCIGasUsed         = inserter.ErrCIGasUsed
	ErrCIParentMismatch  = inserter.ErrCIParentMismatch
	ErrCIEmptyBatch      = inserter.ErrCIEmptyBatch
	ErrCIExecutionFailed = inserter.ErrCIExecutionFailed
	ErrCIInsertFailed    = inserter.ErrCIInsertFailed
)

// Inserter function wrappers.
func DefaultChainInserterConfig() ChainInserterConfig { return inserter.DefaultChainInserterConfig() }
func NewCIMetrics() *CIMetrics                        { return inserter.NewCIMetrics() }
func NewChainInserter(config ChainInserterConfig, ins inserter.BlockInserter) *ChainInserter {
	return inserter.NewChainInserter(config, ins)
}
func DefaultBlockProcessorConfig() BlockProcessorConfig {
	return inserter.DefaultBlockProcessorConfig()
}
func NewBlockProcessorMetrics() *BlockProcessorMetrics { return inserter.NewBlockProcessorMetrics() }
func NewBlockProcessor(config BlockProcessorConfig, ins inserter.BlockInserter) *BlockProcessor {
	return inserter.NewBlockProcessor(config, ins)
}
