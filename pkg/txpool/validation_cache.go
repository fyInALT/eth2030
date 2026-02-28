// validation_cache.go implements smart validation caching for the transaction
// pool. It provides an LRU cache for transaction validation results keyed by
// tx hash, a batch signature verifier that processes multiple signatures in
// parallel, and hit rate tracking for monitoring cache effectiveness.
package txpool

import (
	"crypto/sha256"
	"sync"
	"sync/atomic"
	"time"
)

// ValidationCacheConfig configures the validation cache.
type ValidationCacheConfig struct {
	// MaxEntries is the maximum number of cached entries. Defaults to 8192.
	MaxEntries int
	// TTL is how long a cache entry remains valid. Defaults to 5 minutes.
	TTL time.Duration
	// BatchSize is the default batch size for the signature verifier.
	BatchSize int
}

// DefaultValidationCacheConfig returns production defaults.
func DefaultValidationCacheConfig() ValidationCacheConfig {
	return ValidationCacheConfig{
		MaxEntries: 8192,
		TTL:        5 * time.Minute,
		BatchSize:  64,
	}
}

// CacheEntry holds a cached transaction validation result.
type CacheEntry struct {
	// Valid indicates whether the transaction passed validation.
	Valid bool
	// GasUsed is the estimated gas for this transaction.
	GasUsed uint64
	// Timestamp is when this entry was created.
	Timestamp time.Time
	// Error is the validation error, if any.
	Error error
}

// IsExpired returns true if the entry has exceeded the given TTL.
func (ce *CacheEntry) IsExpired(ttl time.Duration) bool {
	return time.Since(ce.Timestamp) > ttl
}

// cacheNode is a doubly-linked list node for LRU tracking.
type cacheNode struct {
	key   [32]byte
	entry *CacheEntry
	prev  *cacheNode
	next  *cacheNode
}

// ValidationCache is an LRU cache for transaction validation results.
// It is safe for concurrent use.
type ValidationCache struct {
	mu        sync.RWMutex
	config    ValidationCacheConfig
	entries   map[[32]byte]*cacheNode
	head      *cacheNode // most recently used
	tail      *cacheNode // least recently used
	hits      atomic.Uint64
	misses    atomic.Uint64
	evictions atomic.Uint64
}

// NewValidationCache creates a new validation cache with the given config.
func NewValidationCache(cfg ValidationCacheConfig) *ValidationCache {
	if cfg.MaxEntries <= 0 {
		cfg.MaxEntries = 8192
	}
	if cfg.TTL <= 0 {
		cfg.TTL = 5 * time.Minute
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 64
	}
	return &ValidationCache{
		config:  cfg,
		entries: make(map[[32]byte]*cacheNode, cfg.MaxEntries),
	}
}

// Get retrieves a cached validation result for the given tx hash.
// Returns the entry and true if found and not expired, nil and false otherwise.
func (vc *ValidationCache) Get(txHash [32]byte) (*CacheEntry, bool) {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	node, ok := vc.entries[txHash]
	if !ok {
		vc.misses.Add(1)
		return nil, false
	}
	// Check TTL expiry.
	if node.entry.IsExpired(vc.config.TTL) {
		vc.removeNode(node)
		delete(vc.entries, txHash)
		vc.misses.Add(1)
		return nil, false
	}
	// Move to front (most recently used).
	vc.moveToFront(node)
	vc.hits.Add(1)
	return node.entry, true
}

// Put stores a validation result in the cache. If the cache is full,
// the least recently used entry is evicted.
func (vc *ValidationCache) Put(txHash [32]byte, entry *CacheEntry) {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	// If key already exists, update and move to front.
	if node, ok := vc.entries[txHash]; ok {
		node.entry = entry
		vc.moveToFront(node)
		return
	}

	// Evict LRU if at capacity.
	for len(vc.entries) >= vc.config.MaxEntries {
		vc.evictLRU()
	}

	// Insert new node at front.
	node := &cacheNode{
		key:   txHash,
		entry: entry,
	}
	vc.entries[txHash] = node
	vc.pushFront(node)
}

// Invalidate removes a specific entry from the cache.
func (vc *ValidationCache) Invalidate(txHash [32]byte) {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	node, ok := vc.entries[txHash]
	if !ok {
		return
	}
	vc.removeNode(node)
	delete(vc.entries, txHash)
}

// Prune removes all expired entries from the cache. Returns the number
// of entries removed.
func (vc *ValidationCache) Prune() int {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	pruned := 0
	// Walk from tail (oldest) to head.
	node := vc.tail
	for node != nil {
		prev := node.prev
		if node.entry.IsExpired(vc.config.TTL) {
			vc.removeNode(node)
			delete(vc.entries, node.key)
			pruned++
		}
		node = prev
	}
	return pruned
}

// Size returns the number of entries currently in the cache.
func (vc *ValidationCache) Size() int {
	vc.mu.RLock()
	defer vc.mu.RUnlock()
	return len(vc.entries)
}

// HitRate returns the cache hit rate as a value between 0.0 and 1.0.
// Returns 0.0 if no lookups have been performed.
func (vc *ValidationCache) HitRate() float64 {
	hits := vc.hits.Load()
	misses := vc.misses.Load()
	total := hits + misses
	if total == 0 {
		return 0.0
	}
	return float64(hits) / float64(total)
}

// Hits returns the total number of cache hits.
func (vc *ValidationCache) Hits() uint64 {
	return vc.hits.Load()
}

// Misses returns the total number of cache misses.
func (vc *ValidationCache) Misses() uint64 {
	return vc.misses.Load()
}

// Clear removes all entries from the cache.
func (vc *ValidationCache) Clear() {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	vc.entries = make(map[[32]byte]*cacheNode, vc.config.MaxEntries)
	vc.head = nil
	vc.tail = nil
}

// --- LRU linked list operations ---

func (vc *ValidationCache) pushFront(node *cacheNode) {
	node.prev = nil
	node.next = vc.head
	if vc.head != nil {
		vc.head.prev = node
	}
	vc.head = node
	if vc.tail == nil {
		vc.tail = node
	}
}

func (vc *ValidationCache) removeNode(node *cacheNode) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		vc.head = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		vc.tail = node.prev
	}
	node.prev = nil
	node.next = nil
}

func (vc *ValidationCache) moveToFront(node *cacheNode) {
	if vc.head == node {
		return // already at front
	}
	vc.removeNode(node)
	vc.pushFront(node)
}

func (vc *ValidationCache) evictLRU() {
	if vc.tail == nil {
		return
	}
	victim := vc.tail
	vc.removeNode(victim)
	delete(vc.entries, victim.key)
	vc.evictions.Add(1)
}

// --- Batch Signature Verifier ---

// sigTask holds a pending signature verification request.
type sigTask struct {
	txHash   [32]byte
	sigBytes []byte
	pubkey   []byte
}

// BatchSignatureVerifier verifies multiple transaction signatures in parallel
// using a configurable worker count.
type BatchSignatureVerifier struct {
	mu      sync.Mutex
	workers int
	tasks   []sigTask
}

// NewBatchSignatureVerifier creates a new batch verifier with the given
// number of parallel workers. If workers <= 0, defaults to 4.
func NewBatchSignatureVerifier(workers int) *BatchSignatureVerifier {
	if workers <= 0 {
		workers = 4
	}
	return &BatchSignatureVerifier{
		workers: workers,
	}
}

// AddSignature enqueues a signature verification task.
func (bv *BatchSignatureVerifier) AddSignature(txHash [32]byte, sigBytes []byte, pubkey []byte) {
	bv.mu.Lock()
	defer bv.mu.Unlock()
	bv.tasks = append(bv.tasks, sigTask{
		txHash:   txHash,
		sigBytes: sigBytes,
		pubkey:   pubkey,
	})
}

// Pending returns the number of pending verification tasks.
func (bv *BatchSignatureVerifier) Pending() int {
	bv.mu.Lock()
	defer bv.mu.Unlock()
	return len(bv.tasks)
}

// VerifyAll processes all pending signature verifications in parallel and
// returns a map from tx hash to verification result. The task queue is
// cleared after verification.
//
// The verification uses a simplified check: it hashes the concatenation
// of signature bytes and public key, and considers the signature valid if
// the hash meets a basic structural check (non-zero inputs). In production,
// this would delegate to the real ECDSA/BLS/PQ verification backends.
func (bv *BatchSignatureVerifier) VerifyAll() map[[32]byte]bool {
	bv.mu.Lock()
	tasks := make([]sigTask, len(bv.tasks))
	copy(tasks, bv.tasks)
	bv.tasks = bv.tasks[:0]
	bv.mu.Unlock()

	if len(tasks) == 0 {
		return make(map[[32]byte]bool)
	}

	results := make(map[[32]byte]bool, len(tasks))
	var mu sync.Mutex

	// Split tasks across workers.
	numWorkers := bv.workers
	if len(tasks) < numWorkers {
		numWorkers = len(tasks)
	}

	chunkSize := (len(tasks) + numWorkers - 1) / numWorkers
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		start := w * chunkSize
		end := start + chunkSize
		if end > len(tasks) {
			end = len(tasks)
		}
		if start >= end {
			break
		}

		wg.Add(1)
		go func(chunk []sigTask) {
			defer wg.Done()
			for _, task := range chunk {
				valid := verifySigStructural(task.sigBytes, task.pubkey)
				mu.Lock()
				results[task.txHash] = valid
				mu.Unlock()
			}
		}(tasks[start:end])
	}

	wg.Wait()
	return results
}

// Workers returns the configured worker count.
func (bv *BatchSignatureVerifier) Workers() int {
	return bv.workers
}

// verifySigStructural performs a structural signature verification: checks
// that the signature and public key are non-empty and that their hash
// produces a valid-looking digest. In production this would call the real
// ECDSA ecrecover or BLS verify.
func verifySigStructural(sig, pubkey []byte) bool {
	if len(sig) == 0 || len(pubkey) == 0 {
		return false
	}
	// Structural check: hash sig+pubkey and verify the digest is non-zero.
	combined := make([]byte, len(sig)+len(pubkey))
	copy(combined, sig)
	copy(combined[len(sig):], pubkey)
	digest := sha256.Sum256(combined)
	// If the first byte is zero, treat as invalid (simulates ~0.4% failure rate
	// for fuzz testing; real verification would use ecrecover).
	return digest[0] != 0
}
