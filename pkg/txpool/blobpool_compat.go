package txpool

// blobpool_compat.go re-exports types from txpool/blobpool for backward compatibility.

import (
	"math/big"

	"github.com/eth2030/eth2030/txpool/blobpool"
)

// Blob pool type aliases.
type (
	BlobTxPoolConfig      = blobpool.BlobTxPoolConfig
	BlobTxPool            = blobpool.BlobTxPool
	BlobMetadata          = blobpool.BlobMetadata
	BlobSidecar           = blobpool.BlobSidecar
	CustodyConfig         = blobpool.CustodyConfig
	BlobPoolConfig        = blobpool.BlobPoolConfig
	BlobPool              = blobpool.BlobPool
	SparseBlobEntry       = blobpool.SparseBlobEntry
	SparseBlobPoolConfig  = blobpool.SparseBlobPoolConfig
	SparseBlobPool        = blobpool.SparseBlobPool
)

// Blob pool constants.
const (
	MaxBlobsPerBlock         = blobpool.MaxBlobsPerBlock
	TargetBlobsPerBlock      = blobpool.TargetBlobsPerBlock
	BlobTxPoolCapacity       = blobpool.BlobTxPoolCapacity
	BlobTxPoolPerAccountMax  = blobpool.BlobTxPoolPerAccountMax
	BlobGasPerBlobUnit       = blobpool.BlobGasPerBlobUnit
	MaxBlobGasPerBlock       = blobpool.MaxBlobGasPerBlock
	TargetBlobGasPerBlock    = blobpool.TargetBlobGasPerBlock
	MinBlobBaseFee           = blobpool.MinBlobBaseFee
	BlobBaseFeeUpdateFraction = blobpool.BlobBaseFeeUpdateFraction
	DefaultMaxBlobs          = blobpool.DefaultMaxBlobs
	DefaultMaxBlobsPerAccount = blobpool.DefaultMaxBlobsPerAccount
	DefaultMaxBlobSize       = blobpool.DefaultMaxBlobSize
	BlobGasPerBlob           = blobpool.BlobGasPerBlob
	DefaultDatacap           = blobpool.DefaultDatacap
	DefaultBlobPriceBump     = blobpool.DefaultBlobPriceBump
	DefaultEvictionTipThreshold = blobpool.DefaultEvictionTipThreshold
	CellsPerBlob             = blobpool.CellsPerBlob
	DefaultCustodyColumns    = blobpool.DefaultCustodyColumns
	DefaultCellSampleCount   = blobpool.DefaultCellSampleCount
	DefaultSparseMaxBlobs    = blobpool.DefaultSparseMaxBlobs
	DefaultSparseMaxPerAccount = blobpool.DefaultSparseMaxPerAccount
	DefaultSparseExpirySlots = blobpool.DefaultSparseExpirySlots
	VersionedHashPrefix      = blobpool.VersionedHashPrefix
)

// Blob pool error variables.
var (
	ErrBlobTxPoolFull      = blobpool.ErrBlobTxPoolFull
	ErrBlobTxNotType3      = blobpool.ErrBlobTxNotType3
	ErrBlobTxDuplicate     = blobpool.ErrBlobTxDuplicate
	ErrBlobTxNonceLow      = blobpool.ErrBlobTxNonceLow
	ErrBlobTxNoHashes      = blobpool.ErrBlobTxNoHashes
	ErrBlobTxFeeTooLow     = blobpool.ErrBlobTxFeeTooLow
	ErrBlobTxGasExceeded   = blobpool.ErrBlobTxGasExceeded
	ErrBlobTxAccountMax    = blobpool.ErrBlobTxAccountMax
	ErrBlobTxReplaceTooLow = blobpool.ErrBlobTxReplaceTooLow
	ErrBlobPoolFull        = blobpool.ErrBlobPoolFull
	ErrNotBlobTx           = blobpool.ErrNotBlobTx
	ErrBlobAccountLimit    = blobpool.ErrBlobAccountLimit
	ErrBlobAlreadyKnown    = blobpool.ErrBlobAlreadyKnown
	ErrBlobNonceTooLow     = blobpool.ErrBlobNonceTooLow
	ErrBlobMissingHashes   = blobpool.ErrBlobMissingHashes
	ErrBlobFeeCapTooLow    = blobpool.ErrBlobFeeCapTooLow
	ErrBlobReplaceTooLow   = blobpool.ErrBlobReplaceTooLow
	ErrBlobNotCustodied    = blobpool.ErrBlobNotCustodied
	ErrSparseBlobPoolFull  = blobpool.ErrSparseBlobPoolFull
	ErrSparseNotBlobTx     = blobpool.ErrSparseNotBlobTx
	ErrSparseBlobDuplicate = blobpool.ErrSparseBlobDuplicate
	ErrSparseBlobMissing   = blobpool.ErrSparseBlobMissing
	ErrSparseBlobExpired   = blobpool.ErrSparseBlobExpired
	ErrSparseAccountLimit  = blobpool.ErrSparseAccountLimit
	ErrSparseBlobNotFound  = blobpool.ErrSparseBlobNotFound
	ErrSparseInvalidVersion = blobpool.ErrSparseInvalidVersion
)

// Blob pool function wrappers.
func DefaultBlobTxPoolConfig() BlobTxPoolConfig { return blobpool.DefaultBlobTxPoolConfig() }
func NewBlobTxPool(config BlobTxPoolConfig, state StateReader) *BlobTxPool {
	return blobpool.NewBlobTxPool(config, state)
}
func DefaultCustodyConfig() CustodyConfig   { return blobpool.DefaultCustodyConfig() }
func DefaultBlobPoolConfig() BlobPoolConfig { return blobpool.DefaultBlobPoolConfig() }
func NewBlobPool(config BlobPoolConfig, state StateReader) *BlobPool {
	return blobpool.NewBlobPool(config, state)
}
func CalcBlobBaseFee(excessBlobGas uint64) *big.Int {
	return blobpool.CalcBlobBaseFee(excessBlobGas)
}
func CalcExcessBlobGas(parentExcess, parentBlobGasUsed uint64) uint64 {
	return blobpool.CalcExcessBlobGas(parentExcess, parentBlobGasUsed)
}
func DefaultSparseBlobPoolConfig() SparseBlobPoolConfig {
	return blobpool.DefaultSparseBlobPoolConfig()
}
func NewSparseBlobPool(config SparseBlobPoolConfig) *SparseBlobPool {
	return blobpool.NewSparseBlobPool(config)
}
