package rawdb

// hashcache_compat.go re-exports types from core/rawdb/cache for backward compatibility.

import "github.com/eth2030/eth2030/core/rawdb/cache"

// HashCache type aliases.
type (
	HashCacheConfig = cache.HashCacheConfig
	HashCacheEntry  = cache.HashCacheEntry
	HashCacheStats  = cache.HashCacheStats
	HashCache       = cache.HashCache
)

// HashCache function wrappers.
func DefaultHashCacheConfig() HashCacheConfig        { return cache.DefaultHashCacheConfig() }
func NewHashCache(config HashCacheConfig) *HashCache { return cache.NewHashCache(config) }
