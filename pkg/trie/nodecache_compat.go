package trie

// nodecache_compat.go re-exports types from trie/nodecache for backward compatibility.

import "github.com/eth2030/eth2030/trie/nodecache"

// Cache type aliases.
type (
	CacheStats = nodecache.CacheStats
	TrieCache  = nodecache.TrieCache
)

// Cache function wrappers.
func NewTrieCache(maxSize int) *TrieCache { return nodecache.NewTrieCache(maxSize) }
