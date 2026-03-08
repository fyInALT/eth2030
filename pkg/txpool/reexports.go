package txpool

import (
	"github.com/eth2030/eth2030/txpool/blobpool"
	"github.com/eth2030/eth2030/txpool/peertick"
	"github.com/eth2030/eth2030/txpool/sharding"
	"github.com/eth2030/eth2030/txpool/stark"
)

// Re-exported types from sub-packages.
type (
	// stark sub-package
	MempoolAggregationTick = stark.MempoolAggregationTick
	STARKAggregator        = stark.STARKAggregator

	// peertick sub-package
	PeerTickCache = peertick.PeerTickCache

	// sharding sub-package
	ShardConfig = sharding.ShardConfig
	ShardedPool = sharding.ShardedPool

	// blobpool sub-package
	BlobPoolConfig = blobpool.BlobPoolConfig
	BlobPool       = blobpool.BlobPool
)

// Re-exported errors from blobpool sub-package.
var (
	ErrBlobAccountLimit  = blobpool.ErrBlobAccountLimit
	ErrBlobPoolFull      = blobpool.ErrBlobPoolFull
	ErrBlobReplaceTooLow = blobpool.ErrBlobReplaceTooLow
)

// Re-exported constructors and functions.
var (
	NewPeerTickCache      = peertick.NewPeerTickCache
	NewSTARKAggregator    = stark.NewSTARKAggregator
	NewShardedPool        = sharding.NewShardedPool
	DefaultShardConfig    = sharding.DefaultShardConfig
	DefaultBlobPoolConfig = blobpool.DefaultBlobPoolConfig
	NewBlobPool           = blobpool.NewBlobPool
)
