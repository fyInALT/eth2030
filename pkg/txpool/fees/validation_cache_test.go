package fees

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"testing"
	"time"
)

// helper to create a tx hash from an integer.
func testHash(i int) [32]byte {
	return sha256.Sum256([]byte(fmt.Sprintf("tx-%d", i)))
}

// helper to create a valid cache entry.
func validEntry() *CacheEntry {
	return &CacheEntry{
		Valid:     true,
		GasUsed:   21000,
		Timestamp: time.Now(),
	}
}

func TestValidationCacheGetMiss(t *testing.T) {
	vc := NewValidationCache(DefaultValidationCacheConfig())
	_, found := vc.Get(testHash(0))
	if found {
		t.Fatal("expected cache miss for unknown key")
	}
	if vc.Misses() != 1 {
		t.Fatalf("expected 1 miss, got %d", vc.Misses())
	}
}

func TestValidationCachePutAndGet(t *testing.T) {
	vc := NewValidationCache(DefaultValidationCacheConfig())
	h := testHash(1)
	entry := validEntry()
	vc.Put(h, entry)

	got, found := vc.Get(h)
	if !found {
		t.Fatal("expected cache hit after put")
	}
	if !got.Valid {
		t.Fatal("expected valid entry")
	}
	if got.GasUsed != 21000 {
		t.Fatalf("expected gas 21000, got %d", got.GasUsed)
	}
}

func TestValidationCacheSize(t *testing.T) {
	vc := NewValidationCache(ValidationCacheConfig{MaxEntries: 100, TTL: time.Minute})
	for i := 0; i < 50; i++ {
		vc.Put(testHash(i), validEntry())
	}
	if vc.Size() != 50 {
		t.Fatalf("expected size 50, got %d", vc.Size())
	}
}

func TestValidationCacheLRUEviction(t *testing.T) {
	cfg := ValidationCacheConfig{MaxEntries: 3, TTL: time.Minute}
	vc := NewValidationCache(cfg)

	// Fill cache.
	vc.Put(testHash(0), validEntry())
	vc.Put(testHash(1), validEntry())
	vc.Put(testHash(2), validEntry())

	// Insert a 4th entry -- should evict the LRU (testHash(0)).
	vc.Put(testHash(3), validEntry())

	if vc.Size() != 3 {
		t.Fatalf("expected size 3 after eviction, got %d", vc.Size())
	}
	_, found := vc.Get(testHash(0))
	if found {
		t.Fatal("expected testHash(0) to be evicted")
	}
	_, found = vc.Get(testHash(3))
	if !found {
		t.Fatal("expected testHash(3) to be present")
	}
}

func TestValidationCacheLRUAccessOrder(t *testing.T) {
	cfg := ValidationCacheConfig{MaxEntries: 3, TTL: time.Minute}
	vc := NewValidationCache(cfg)

	vc.Put(testHash(0), validEntry())
	vc.Put(testHash(1), validEntry())
	vc.Put(testHash(2), validEntry())

	// Access testHash(0) to make it most recently used.
	vc.Get(testHash(0))

	// Insert testHash(3) -- should evict testHash(1) (the LRU).
	vc.Put(testHash(3), validEntry())

	_, found := vc.Get(testHash(1))
	if found {
		t.Fatal("expected testHash(1) to be evicted (LRU)")
	}
	_, found = vc.Get(testHash(0))
	if !found {
		t.Fatal("expected testHash(0) to survive (recently accessed)")
	}
}

func TestValidationCacheTTLExpiry(t *testing.T) {
	cfg := ValidationCacheConfig{MaxEntries: 100, TTL: 50 * time.Millisecond}
	vc := NewValidationCache(cfg)

	h := testHash(0)
	vc.Put(h, validEntry())

	// Entry should be available immediately.
	_, found := vc.Get(h)
	if !found {
		t.Fatal("expected cache hit before TTL expiry")
	}

	// Wait for TTL to expire.
	time.Sleep(80 * time.Millisecond)

	_, found = vc.Get(h)
	if found {
		t.Fatal("expected cache miss after TTL expiry")
	}
}

func TestValidationCacheInvalidate(t *testing.T) {
	vc := NewValidationCache(DefaultValidationCacheConfig())
	h := testHash(0)
	vc.Put(h, validEntry())
	vc.Invalidate(h)

	if vc.Size() != 0 {
		t.Fatalf("expected size 0 after invalidation, got %d", vc.Size())
	}
	_, found := vc.Get(h)
	if found {
		t.Fatal("expected miss after invalidation")
	}
}

func TestValidationCachePrune(t *testing.T) {
	cfg := ValidationCacheConfig{MaxEntries: 100, TTL: 50 * time.Millisecond}
	vc := NewValidationCache(cfg)

	for i := 0; i < 10; i++ {
		vc.Put(testHash(i), validEntry())
	}
	time.Sleep(80 * time.Millisecond)

	pruned := vc.Prune()
	if pruned != 10 {
		t.Fatalf("expected 10 pruned, got %d", pruned)
	}
	if vc.Size() != 0 {
		t.Fatalf("expected size 0 after prune, got %d", vc.Size())
	}
}

func TestValidationCacheHitRate(t *testing.T) {
	vc := NewValidationCache(DefaultValidationCacheConfig())

	// No lookups: hit rate should be 0.
	if vc.HitRate() != 0.0 {
		t.Fatalf("expected 0.0 hit rate, got %f", vc.HitRate())
	}

	h := testHash(0)
	vc.Put(h, validEntry())

	// 1 miss, then 1 hit.
	vc.Get(testHash(99)) // miss
	vc.Get(h)            // hit

	rate := vc.HitRate()
	if rate < 0.49 || rate > 0.51 {
		t.Fatalf("expected ~0.5 hit rate, got %f", rate)
	}
}

func TestValidationCacheClear(t *testing.T) {
	vc := NewValidationCache(DefaultValidationCacheConfig())
	for i := 0; i < 20; i++ {
		vc.Put(testHash(i), validEntry())
	}
	vc.Clear()
	if vc.Size() != 0 {
		t.Fatalf("expected size 0 after clear, got %d", vc.Size())
	}
}

func TestValidationCacheUpdateExisting(t *testing.T) {
	vc := NewValidationCache(DefaultValidationCacheConfig())
	h := testHash(0)
	vc.Put(h, &CacheEntry{Valid: true, GasUsed: 21000, Timestamp: time.Now()})
	vc.Put(h, &CacheEntry{Valid: false, GasUsed: 50000, Timestamp: time.Now(), Error: fmt.Errorf("bad")})

	got, found := vc.Get(h)
	if !found {
		t.Fatal("expected hit after update")
	}
	if got.Valid {
		t.Fatal("expected updated entry to be invalid")
	}
	if got.GasUsed != 50000 {
		t.Fatalf("expected updated gas 50000, got %d", got.GasUsed)
	}
	if vc.Size() != 1 {
		t.Fatalf("expected size 1 after update, got %d", vc.Size())
	}
}

func TestValidationCacheConcurrent(t *testing.T) {
	vc := NewValidationCache(ValidationCacheConfig{MaxEntries: 1000, TTL: time.Minute})
	var wg sync.WaitGroup

	// Concurrent writes.
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			vc.Put(testHash(idx), validEntry())
		}(i)
	}
	wg.Wait()

	// Concurrent reads.
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			vc.Get(testHash(idx))
		}(i)
	}
	wg.Wait()

	if vc.Size() != 100 {
		t.Fatalf("expected 100 entries after concurrent ops, got %d", vc.Size())
	}
}

func TestCacheEntryIsExpired(t *testing.T) {
	entry := &CacheEntry{
		Valid:     true,
		Timestamp: time.Now().Add(-10 * time.Second),
	}
	if !entry.IsExpired(5 * time.Second) {
		t.Fatal("expected entry to be expired")
	}
	if entry.IsExpired(20 * time.Second) {
		t.Fatal("expected entry to not be expired")
	}
}

// --- BatchSignatureVerifier tests ---

func TestBatchSignatureVerifierEmpty(t *testing.T) {
	bv := NewBatchSignatureVerifier(4)
	results := bv.VerifyAll()
	if len(results) != 0 {
		t.Fatalf("expected empty results for no tasks, got %d", len(results))
	}
}

func TestBatchSignatureVerifierPending(t *testing.T) {
	bv := NewBatchSignatureVerifier(4)
	bv.AddSignature(testHash(0), []byte("sig0"), []byte("pub0"))
	bv.AddSignature(testHash(1), []byte("sig1"), []byte("pub1"))
	if bv.Pending() != 2 {
		t.Fatalf("expected 2 pending, got %d", bv.Pending())
	}
}

func TestBatchSignatureVerifierVerifyAll(t *testing.T) {
	bv := NewBatchSignatureVerifier(2)
	for i := 0; i < 10; i++ {
		sig := []byte(fmt.Sprintf("signature-%d", i))
		pub := []byte(fmt.Sprintf("pubkey-%d", i))
		bv.AddSignature(testHash(i), sig, pub)
	}

	results := bv.VerifyAll()
	if len(results) != 10 {
		t.Fatalf("expected 10 results, got %d", len(results))
	}
	// After VerifyAll, pending should be 0.
	if bv.Pending() != 0 {
		t.Fatalf("expected 0 pending after VerifyAll, got %d", bv.Pending())
	}
}

func TestBatchSignatureVerifierInvalidSig(t *testing.T) {
	bv := NewBatchSignatureVerifier(1)
	// Empty sig should fail.
	bv.AddSignature(testHash(0), nil, []byte("pubkey"))
	results := bv.VerifyAll()
	if results[testHash(0)] {
		t.Fatal("expected invalid result for nil signature")
	}
}

func TestBatchSignatureVerifierInvalidPubkey(t *testing.T) {
	bv := NewBatchSignatureVerifier(1)
	// Empty pubkey should fail.
	bv.AddSignature(testHash(0), []byte("sig"), nil)
	results := bv.VerifyAll()
	if results[testHash(0)] {
		t.Fatal("expected invalid result for nil pubkey")
	}
}

func TestBatchSignatureVerifierWorkers(t *testing.T) {
	bv := NewBatchSignatureVerifier(8)
	if bv.Workers() != 8 {
		t.Fatalf("expected 8 workers, got %d", bv.Workers())
	}
}

func TestBatchSignatureVerifierDefaultWorkers(t *testing.T) {
	bv := NewBatchSignatureVerifier(0)
	if bv.Workers() != 4 {
		t.Fatalf("expected default 4 workers, got %d", bv.Workers())
	}
}

func TestBatchSignatureVerifierConcurrent(t *testing.T) {
	bv := NewBatchSignatureVerifier(4)
	var wg sync.WaitGroup

	// Concurrent adds.
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sig := []byte(fmt.Sprintf("sig-%d", idx))
			pub := []byte(fmt.Sprintf("pub-%d", idx))
			bv.AddSignature(testHash(idx), sig, pub)
		}(i)
	}
	wg.Wait()

	if bv.Pending() != 50 {
		t.Fatalf("expected 50 pending after concurrent adds, got %d", bv.Pending())
	}

	results := bv.VerifyAll()
	if len(results) != 50 {
		t.Fatalf("expected 50 results, got %d", len(results))
	}
}

func TestValidationCacheDefaultConfig(t *testing.T) {
	cfg := DefaultValidationCacheConfig()
	if cfg.MaxEntries != 8192 {
		t.Fatalf("expected MaxEntries=8192, got %d", cfg.MaxEntries)
	}
	if cfg.TTL != 5*time.Minute {
		t.Fatalf("expected TTL=5m, got %v", cfg.TTL)
	}
	if cfg.BatchSize != 64 {
		t.Fatalf("expected BatchSize=64, got %d", cfg.BatchSize)
	}
}
