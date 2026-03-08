package txpool

// sharding_compat.go re-exports types from txpool/sharding for backward compatibility.

import "github.com/eth2030/eth2030/txpool/sharding"

// Sharding type aliases.
type (
	ShardConfig = sharding.ShardConfig
	ShardStats  = sharding.ShardStats
	TxShard     = sharding.TxShard
	ShardedPool = sharding.ShardedPool
)

// Sharding error variables.
var (
	ErrShardFull    = sharding.ErrShardFull
	ErrShardInvalid = sharding.ErrShardInvalid
)

// Sharding function wrappers.
func DefaultShardConfig() ShardConfig   { return sharding.DefaultShardConfig() }
func NewShardedPool(config ShardConfig) *ShardedPool {
	return sharding.NewShardedPool(config)
}
