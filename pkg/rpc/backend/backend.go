// Package rpcbackend defines the Backend interface and related service
// accessors used by the Ethereum JSON-RPC layer.
package rpcbackend

import (
	"math/big"

	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/core/vm"
	rpctypes "github.com/eth2030/eth2030/rpc/types"
	"github.com/eth2030/eth2030/trie"
)

// Backend provides access to chain data for the JSON-RPC API.
// This interface decouples the RPC layer from the chain implementation,
// following go-ethereum's ethapi.Backend pattern.
type Backend interface {
	// Chain data
	HeaderByNumber(number rpctypes.BlockNumber) *types.Header
	HeaderByHash(hash types.Hash) *types.Header
	BlockByNumber(number rpctypes.BlockNumber) *types.Block
	BlockByHash(hash types.Hash) *types.Block
	CurrentHeader() *types.Header
	ChainID() *big.Int

	// State access
	StateAt(root types.Hash) (state.StateDB, error)

	// Transaction pool
	SendTransaction(tx *types.Transaction) error
	GetTransaction(hash types.Hash) (*types.Transaction, uint64, uint64) // tx, blockNum, index

	// Gas estimation
	SuggestGasPrice() *big.Int

	// Receipts and logs
	GetReceipts(blockHash types.Hash) []*types.Receipt
	GetLogs(blockHash types.Hash) []*types.Log
	GetBlockReceipts(number uint64) []*types.Receipt

	// Proofs
	GetProof(addr types.Address, storageKeys []types.Hash, blockNumber rpctypes.BlockNumber) (*trie.AccountProof, error)

	// EVM execution
	EVMCall(from types.Address, to *types.Address, data []byte, gas uint64, value *big.Int, blockNumber rpctypes.BlockNumber) ([]byte, uint64, error)

	// Tracing
	TraceTransaction(txHash types.Hash) (*vm.StructLogTracer, error)

	// History availability (EIP-4444)
	// HistoryOldestBlock returns the oldest block number for which bodies
	// and receipts are still available. Returns 0 if no pruning occurred.
	HistoryOldestBlock() uint64

	// Blob schedule
	// BlobSchedule returns the target, max blobs per block and base fee update
	// fraction for the given block timestamp. Used by eth_blobBaseFee and
	// eth_feeHistory to compute correct per-fork blob fees and ratios.
	BlobSchedule(blockTime uint64) (target, max uint64, updateFraction uint64)
}
