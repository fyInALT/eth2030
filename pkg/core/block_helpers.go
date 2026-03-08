package core

// block_helpers.go provides unexported wrappers around core/block functions
// so that tests within the core/ package can call them by the same names they
// used before the functions were moved to core/block/.

import (
	"math/big"

	"github.com/eth2030/eth2030/core/block"
	"github.com/eth2030/eth2030/core/types"
)

// effectiveGasPrice returns the effective gas price for a transaction.
func effectiveGasPrice(tx *types.Transaction, baseFee *big.Int) *big.Int {
	return block.EffectiveGasPrice(tx, baseFee)
}

// deriveReceiptsRoot computes the receipts root using a Merkle Patricia Trie.
func deriveReceiptsRoot(receipts []*types.Receipt) types.Hash {
	return block.DeriveReceiptsRoot(receipts)
}

// deriveTxsRoot computes the transactions root using a Merkle Patricia Trie.
func deriveTxsRoot(txs []*types.Transaction) types.Hash {
	return block.DeriveTxsRoot(txs)
}

// deriveWithdrawalsRoot computes the withdrawals root using a Merkle Patricia Trie.
func deriveWithdrawalsRoot(ws []*types.Withdrawal) types.Hash {
	return block.DeriveWithdrawalsRoot(ws)
}

// validateBlobHashes checks versioned hash version bytes.
func validateBlobHashes(hashes []types.Hash) error {
	return block.ValidateBlobHashes(hashes)
}

// calldataFloorDelta computes additional gas under EIP-7623.
func calldataFloorDelta(tx *types.Transaction, standardGasUsed uint64) uint64 {
	return block.CalldataFloorDelta(tx, standardGasUsed)
}

// sortedTxLists separates and sorts transactions by gas price.
func sortedTxLists(pending []*types.Transaction, baseFee *big.Int) (regular, blobs []*types.Transaction) {
	return block.SortedTxLists(pending, baseFee)
}

// calcGasLimit calculates the gas limit for the next block.
func calcGasLimit(parentGasLimit, parentGasUsed uint64) uint64 {
	return block.CalcGasLimit(parentGasLimit, parentGasUsed)
}

// verifyGasLimit checks that the gas limit change is within bounds.
func verifyGasLimit(parentGasLimit, headerGasLimit uint64) error {
	return block.VerifyGasLimit(parentGasLimit, headerGasLimit)
}
