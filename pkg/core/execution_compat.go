package core

// execution_compat.go re-exports types and functions from core/execution for
// backward compatibility with callers inside the core/ package root.

import (
	"math/big"

	"github.com/eth2030/eth2030/bal"
	"github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/execution"
	"github.com/eth2030/eth2030/core/gaspool"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/core/vm"
)

// --- Type aliases ---

// StateProcessor is an alias for execution.StateProcessor.
type StateProcessor = execution.StateProcessor

// ProcessResult is an alias for execution.ProcessResult.
type ProcessResult = execution.ProcessResult

// ExecutionResult is an alias for execution.ExecutionResult.
type ExecutionResult = execution.ExecutionResult

// ExtendedExecutionResult is an alias for execution.ExtendedExecutionResult.
type ExtendedExecutionResult = execution.ExtendedExecutionResult

// GasBreakdown is an alias for execution.GasBreakdown.
type GasBreakdown = execution.GasBreakdown

// TraceOutput is an alias for execution.TraceOutput.
type TraceOutput = execution.TraceOutput

// AccessListEntry is an alias for execution.AccessListEntry.
type AccessListEntry = execution.AccessListEntry

// ReceiptGenerator is an alias for execution.ReceiptGenerator.
type ReceiptGenerator = execution.ReceiptGenerator

// ReceiptGeneratorConfig is an alias for execution.ReceiptGeneratorConfig.
type ReceiptGeneratorConfig = execution.ReceiptGeneratorConfig

// TxExecutionOutcome is an alias for execution.TxExecutionOutcome.
type TxExecutionOutcome = execution.TxExecutionOutcome

// ReceiptProcessor is an alias for execution.ReceiptProcessor.
type ReceiptProcessor = execution.ReceiptProcessor

// ReceiptProcessorConfig is an alias for execution.ReceiptProcessorConfig.
type ReceiptProcessorConfig = execution.ReceiptProcessorConfig

// DependencyGraph is an alias for execution.DependencyGraph.
type DependencyGraph = execution.DependencyGraph

// TxGroup is an alias for execution.TxGroup.
type TxGroup = execution.TxGroup

// ParallelProcessor is an alias for execution.ParallelProcessor.
type ParallelProcessor = execution.ParallelProcessor

// RichDataStore is an alias for execution.RichDataStore.
type RichDataStore = execution.RichDataStore

// SchemaField is an alias for execution.SchemaField.
type SchemaField = execution.SchemaField

// DataType is an alias for execution.DataType.
type DataType = execution.DataType

// TxExecutor is an alias for execution.TxExecutor.
type TxExecutor = execution.TxExecutor

// --- Constant re-exports ---

const (
	TxGas               = execution.TxGas
	TxDataZeroGas       = execution.TxDataZeroGas
	TxDataNonZeroGas    = execution.TxDataNonZeroGas
	TxCreateGas         = execution.TxCreateGas
	PerAuthBaseCost     = execution.PerAuthBaseCost
	PerEmptyAccountCost = execution.PerEmptyAccountCost
	BlobGasPerBlob      = execution.BlobGasPerBlob

	TotalCostFloorPerToken       = execution.TotalCostFloorPerToken
	StandardTokenCost            = execution.StandardTokenCost
	FloorTokenCost               = execution.FloorTokenCost
	TotalCostFloorPerTokenGlamst = execution.TotalCostFloorPerTokenGlamst

	TypeUint256 = execution.TypeUint256
	TypeAddress = execution.TypeAddress
	TypeBytes32 = execution.TypeBytes32
	TypeString  = execution.TypeString
	TypeBool    = execution.TypeBool
	TypeArray   = execution.TypeArray
)

// --- Variable re-exports ---

var (
	ErrNonceTooLow         = execution.ErrNonceTooLow
	ErrNonceTooHigh        = execution.ErrNonceTooHigh
	ErrInsufficientBalance = execution.ErrInsufficientBalance
	ErrGasLimitExceeded    = execution.ErrGasLimitExceeded
	ErrIntrinsicGasTooLow  = execution.ErrIntrinsicGasTooLow
	ErrContractCreation    = execution.ErrContractCreation
	ErrContractCall        = execution.ErrContractCall

	ErrBALFeasibilityViolated = execution.ErrBALFeasibilityViolated
	ErrBALHashMismatch        = execution.ErrBALHashMismatch

	ErrExecutionReverted = execution.ErrExecutionReverted

	ErrNilReceipt          = execution.ErrNilReceipt
	ErrMaxReceiptsExceeded = execution.ErrMaxReceiptsExceeded

	ErrSchemaExists       = execution.ErrSchemaExists
	ErrSchemaNotFound     = execution.ErrSchemaNotFound
	ErrDataNotFound       = execution.ErrDataNotFound
	ErrFieldNotInSchema   = execution.ErrFieldNotInSchema
	ErrMissingRequired    = execution.ErrMissingRequired
	ErrFieldTooLarge      = execution.ErrFieldTooLarge
	ErrEmptySchema        = execution.ErrEmptySchema
	ErrDuplicateFieldName = execution.ErrDuplicateFieldName
	ErrDataExists         = execution.ErrDataExists
)

// --- Function re-exports (public API) ---

// NewStateProcessor creates a new state processor.
func NewStateProcessor(cfg *config.ChainConfig) *StateProcessor {
	return execution.NewStateProcessor(cfg)
}

// NewParallelProcessor creates a new parallel processor.
func NewParallelProcessor(cfg *config.ChainConfig) *ParallelProcessor {
	return execution.NewParallelProcessor(cfg)
}

// NewReceiptGenerator creates a new receipt generator.
func NewReceiptGenerator(cfg ReceiptGeneratorConfig) *ReceiptGenerator {
	return execution.NewReceiptGenerator(cfg)
}

// DefaultReceiptGeneratorConfig returns default config for receipt generation.
func DefaultReceiptGeneratorConfig() ReceiptGeneratorConfig {
	return execution.DefaultReceiptGeneratorConfig()
}

// NewReceiptProcessor creates a new receipt processor.
func NewReceiptProcessor(cfg ReceiptProcessorConfig) *ReceiptProcessor {
	return execution.NewReceiptProcessor(cfg)
}

// DefaultReceiptProcessorConfig returns default config for receipt processing.
func DefaultReceiptProcessorConfig() ReceiptProcessorConfig {
	return execution.DefaultReceiptProcessorConfig()
}

// NewRichDataStore creates an empty RichDataStore.
func NewRichDataStore() *RichDataStore {
	return execution.NewRichDataStore()
}

// NewDependencyGraph builds a dependency graph from transactions and their BAL.
func NewDependencyGraph(txs []*types.Transaction, accessList *bal.BlockAccessList) *DependencyGraph {
	return execution.NewDependencyGraph(txs, accessList)
}

// ApplyTransaction applies a single transaction to the state.
func ApplyTransaction(cfg *config.ChainConfig, statedb state.StateDB, header *types.Header, tx *types.Transaction, gp *gaspool.GasPool) (*types.Receipt, uint64, error) {
	return execution.ApplyTransaction(cfg, statedb, header, tx, gp)
}

// ApplyTransactionWithBAL applies a transaction with EIP-7928 BAL tracking.
func ApplyTransactionWithBAL(cfg *config.ChainConfig, statedb state.StateDB, header *types.Header, tx *types.Transaction, gp *gaspool.GasPool, tracker vm.BALTracker) (*types.Receipt, uint64, error) {
	return execution.ApplyTransactionWithBAL(cfg, statedb, header, tx, gp, tracker)
}

// ProcessRequests collects EIP-7685 execution layer requests.
func ProcessRequests(cfg *config.ChainConfig, statedb state.StateDB, header *types.Header) (types.Requests, error) {
	return execution.ProcessRequests(cfg, statedb, header)
}

// ProcessWithdrawals applies EIP-4895 beacon chain withdrawals to the state.
func ProcessWithdrawals(statedb state.StateDB, withdrawals []*types.Withdrawal) {
	execution.ProcessWithdrawals(statedb, withdrawals)
}

// CalcWithdrawalsHash computes the withdrawals root hash.
func CalcWithdrawalsHash(withdrawals []*types.Withdrawal) types.Hash {
	return execution.CalcWithdrawalsHash(withdrawals)
}

// ReceiptTrieRoot computes the receipt trie root hash.
func ReceiptTrieRoot(receipts []*types.Receipt) types.Hash {
	return execution.ReceiptTrieRoot(receipts)
}

// ComputeBlockBloomFromReceipts computes the aggregate bloom filter.
func ComputeBlockBloomFromReceipts(receipts []*types.Receipt) types.Bloom {
	return execution.ComputeBlockBloomFromReceipts(receipts)
}

// DeriveReceiptStatus returns the receipt status code.
func DeriveReceiptStatus(failed bool) uint64 {
	return execution.DeriveReceiptStatus(failed)
}

// CalcBlobGasUsed returns blob gas consumed by a transaction.
func CalcBlobGasUsed(numBlobs int) uint64 {
	return execution.CalcBlobGasUsed(numBlobs)
}

// CalcEffectiveGasPrice computes the effective gas price.
func CalcEffectiveGasPrice(baseFee, gasFeeCap, gasTipCap *big.Int) *big.Int {
	return execution.CalcEffectiveGasPrice(baseFee, gasFeeCap, gasTipCap)
}

// DecodeRevertReason decodes revert reason from raw return data.
func DecodeRevertReason(data []byte) string {
	return execution.DecodeRevertReason(data)
}

// IsRevert returns true if the error represents an EVM revert.
func IsRevert(err error) bool {
	return execution.IsRevert(err)
}

// ParseRevertData parses revert data.
func ParseRevertData(data []byte) (reason string, ok bool) {
	return execution.ParseRevertData(data)
}

// NewGasBreakdown creates a GasBreakdown.
func NewGasBreakdown(executionGas, executionRefund, calldataGas, blobGas, intrinsicGas uint64, baseFee, calldataBaseFee, blobBaseFee *big.Int) *GasBreakdown {
	return execution.NewGasBreakdown(executionGas, executionRefund, calldataGas, blobGas, intrinsicGas, baseFee, calldataBaseFee, blobBaseFee)
}

// IsLocal returns true if the transaction is a LocalTx type.
func IsLocal(tx *types.Transaction) bool {
	return execution.IsLocal(tx)
}

// ClassifyTransactions splits transactions into local and global groups.
func ClassifyTransactions(txs []*types.Transaction) (local, global []*types.Transaction) {
	return execution.ClassifyTransactions(txs)
}

// --- Internal helpers exposed for use within core/ sub-packages ---
// These wrappers allow core/block_builder.go and similar files to access
// internal helpers that now reside in core/execution/.

// calldataFloorGas delegates to execution.CalldataFloorGas.
func calldataFloorGas(data []byte, isCreate bool) uint64 {
	return execution.CalldataFloorGas(data, isCreate)
}

// calldataFloorGasGlamst delegates to execution.CalldataFloorGasGlamst.
func calldataFloorGasGlamst(data []byte, accessList types.AccessList, isCreate bool) uint64 {
	return execution.CalldataFloorGasGlamst(data, accessList, isCreate)
}

// capturePreState delegates to execution.CapturePreState.
func capturePreState(statedb state.StateDB, tx *types.Transaction) (map[types.Address]*big.Int, map[types.Address]uint64) {
	return execution.CapturePreState(statedb, tx)
}

// balTrackerOrNil delegates to execution.BalTrackerOrNil.
func balTrackerOrNil(t *bal.AccessTracker) vm.BALTracker {
	return execution.BalTrackerOrNil(t)
}

// populateTracker delegates to execution.PopulateTracker.
func populateTracker(tracker *bal.AccessTracker, statedb state.StateDB, preBalances map[types.Address]*big.Int, preNonces map[types.Address]uint64) {
	execution.PopulateTracker(tracker, statedb, preBalances, preNonces)
}

// calcBlobBaseFee delegates to execution.CalcBlobBaseFee.
func calcBlobBaseFee(excessBlobGas uint64) *big.Int {
	return execution.CalcBlobBaseFee(excessBlobGas)
}

// applyTransaction delegates to the internal execution path (no GetHash).
func applyTransaction(cfg *config.ChainConfig, getHash vm.GetHashFunc, statedb state.StateDB, header *types.Header, tx *types.Transaction, gp *gaspool.GasPool) (*types.Receipt, uint64, error) {
	return execution.ApplyTransactionInternal(cfg, getHash, statedb, header, tx, gp)
}

// intrinsicGasGlamst delegates to execution.IntrinsicGasGlamst.
func intrinsicGasGlamst(data []byte, isCreate bool, hasValue bool, toExists bool, authCount, emptyAuthCount uint64) uint64 {
	return execution.IntrinsicGasGlamst(data, isCreate, hasValue, toExists, authCount, emptyAuthCount)
}

// intrinsicGas delegates to execution.IntrinsicGas.
func intrinsicGas(data []byte, isCreate, isShanghai bool, authCount, emptyAuthCount uint64) uint64 {
	return execution.IntrinsicGas(data, isCreate, isShanghai, authCount, emptyAuthCount)
}

// calldataTokens delegates to execution.CalldataTokens.
func calldataTokens(data []byte) uint64 {
	return execution.CalldataTokens(data)
}

// accessListDataTokens delegates to execution.AccessListDataTokens.
func accessListDataTokens(accessList types.AccessList) uint64 {
	return execution.AccessListDataTokens(accessList)
}

// accessListGasGlamst delegates to execution.AccessListGasGlamst.
func accessListGasGlamst(accessList types.AccessList) uint64 {
	return execution.AccessListGasGlamst(accessList)
}

// applyMessage delegates to execution.ApplyMessage.
func applyMessage(cfg *config.ChainConfig, getHash vm.GetHashFunc, statedb state.StateDB, header *types.Header, msg *config.Message, gp *gaspool.GasPool, balTrackers ...vm.BALTracker) (*ExecutionResult, error) {
	return execution.ApplyMessage(cfg, getHash, statedb, header, msg, gp, balTrackers...)
}

// validateBAL delegates to execution.ValidateBAL.
func validateBAL(header *types.Header, accessList *bal.BlockAccessList) error {
	return execution.ValidateBAL(header, accessList)
}

// requestCountSlot mirrors execution.RequestCountSlot.
var requestCountSlot = execution.RequestCountSlot

// requestDataSlotBase mirrors execution.RequestDataSlotBase.
var requestDataSlotBase = execution.RequestDataSlotBase

// incrementSlot delegates to execution.IncrementSlot.
func incrementSlot(base types.Hash, offset uint64) types.Hash {
	return execution.IncrementSlot(base, offset)
}

// countToUint64 delegates to execution.CountToUint64.
func countToUint64(val types.Hash) uint64 {
	return execution.CountToUint64(val)
}

// trimTrailingZeros delegates to execution.TrimTrailingZeros.
func trimTrailingZeros(b []byte) []byte {
	return execution.TrimTrailingZeros(b)
}
