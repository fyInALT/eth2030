package light

// cache_compat.go re-exports types from light/cache for backward compatibility.

import (
	"time"

	lightcache "github.com/eth2030/eth2030/light/cache"
)

// Cache type aliases.
type (
	ProofType   = lightcache.ProofType
	CacheKey    = lightcache.CacheKey
	CachedProof = lightcache.CachedProof
	CacheStats  = lightcache.CacheStats
	ProofCache  = lightcache.ProofCache
)

// Cache constants.
const (
	ProofTypeHeader  = lightcache.ProofTypeHeader
	ProofTypeAccount = lightcache.ProofTypeAccount
	ProofTypeStorage = lightcache.ProofTypeStorage
)

// Cache function wrappers.
func NewProofCache(maxSize int, ttl time.Duration) *ProofCache {
	return lightcache.NewProofCache(maxSize, ttl)
}
