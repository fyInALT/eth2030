package witness

// cache_compat.go re-exports types from witness/cache for backward compatibility.

import "github.com/eth2030/eth2030/witness/cache"

// Cache type aliases.
type (
	CachedWitness    = cache.CachedWitness
	WitnessCacheStats = cache.WitnessCacheStats
	WitnessCache     = cache.WitnessCache
)

// Cache function wrappers.
func NewWitnessCache(maxBlocks int) *WitnessCache { return cache.NewWitnessCache(maxBlocks) }
