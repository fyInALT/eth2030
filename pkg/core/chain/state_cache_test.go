package chain

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
)

func makeTestState(balance int64) *state.MemoryStateDB {
	s := state.NewMemoryStateDB()
	s.AddBalance(types.Address{1}, big.NewInt(balance))
	return s
}

func TestStateCache_PutAndGet(t *testing.T) {
	sc := newStateCache(defaultMaxCachedStates)

	hash := types.Hash{0xAA}
	sdb := makeTestState(100)
	sc.put(hash, 10, sdb)

	got, ok := sc.get(hash)
	if !ok {
		t.Fatal("expected cache hit")
	}
	bal := got.GetBalance(types.Address{1})
	if bal.Int64() != 100 {
		t.Fatalf("balance mismatch: got %d, want 100", bal.Int64())
	}
}

func TestStateCache_NotFound(t *testing.T) {
	sc := newStateCache(defaultMaxCachedStates)
	_, ok := sc.get(types.Hash{0xFF})
	if ok {
		t.Fatal("expected cache miss")
	}
}

func TestStateCache_Isolation(t *testing.T) {
	sc := newStateCache(defaultMaxCachedStates)
	hash := types.Hash{0xBB}
	sdb := makeTestState(200)
	sc.put(hash, 5, sdb)

	// Modify the returned copy.
	got, _ := sc.get(hash)
	got.AddBalance(types.Address{1}, big.NewInt(999))

	// Original should be unchanged.
	got2, _ := sc.get(hash)
	bal := got2.GetBalance(types.Address{1})
	if bal.Int64() != 200 {
		t.Fatalf("cache should be isolated: got %d, want 200", bal.Int64())
	}
}

// testHash encodes an integer into a Hash (little-endian two-byte prefix).
func testHash(i int) types.Hash {
	return types.Hash{byte(i), byte(i >> 8)}
}

func TestStateCache_Eviction(t *testing.T) {
	// Use a small fixed capacity to keep the test fast and byte-safe.
	const cap = 16
	sc := newStateCache(cap)

	// Fill beyond max.
	for i := 0; i < cap+10; i++ {
		sc.put(testHash(i), uint64(i), makeTestState(int64(i)))
	}

	// Should not exceed max.
	sc.mu.RLock()
	count := len(sc.snapshots)
	sc.mu.RUnlock()
	if count > cap {
		t.Fatalf("expected at most %d cached states, got %d", cap, count)
	}

	// Oldest entries should have been evicted.
	_, ok := sc.get(testHash(0))
	if ok {
		t.Fatal("expected oldest entry to be evicted")
	}

	// Newest entries should still be present.
	_, ok = sc.get(testHash(cap + 9))
	if !ok {
		t.Fatal("expected newest entry to be present")
	}
}

func TestStateCache_Closest(t *testing.T) {
	sc := newStateCache(defaultMaxCachedStates)

	sc.put(types.Hash{0x10}, 0, makeTestState(0))
	sc.put(types.Hash{0x20}, 16, makeTestState(16))
	sc.put(types.Hash{0x30}, 32, makeTestState(32))
	sc.put(types.Hash{0x40}, 48, makeTestState(48))

	// Find closest to block 40.
	_, num, ok := sc.closest(40)
	if !ok {
		t.Fatal("expected match")
	}
	if num != 32 {
		t.Fatalf("closest: got %d, want 32", num)
	}

	// Find closest to block 48.
	_, num, ok = sc.closest(48)
	if !ok {
		t.Fatal("expected match")
	}
	if num != 48 {
		t.Fatalf("closest: got %d, want 48", num)
	}

	// Find closest to block 5 (should get genesis at 0).
	_, num, ok = sc.closest(5)
	if !ok {
		t.Fatal("expected match")
	}
	if num != 0 {
		t.Fatalf("closest: got %d, want 0", num)
	}
}

func TestStateCache_Remove(t *testing.T) {
	sc := newStateCache(defaultMaxCachedStates)
	hash := types.Hash{0xCC}
	sc.put(hash, 10, makeTestState(100))

	sc.remove(hash)

	_, ok := sc.get(hash)
	if ok {
		t.Fatal("expected cache miss after remove")
	}
}

func TestStateCache_Clear(t *testing.T) {
	sc := newStateCache(defaultMaxCachedStates)
	for i := 0; i < 10; i++ {
		sc.put(types.Hash{byte(i)}, uint64(i), makeTestState(int64(i)))
	}
	sc.clear()

	sc.mu.RLock()
	count := len(sc.snapshots)
	sc.mu.RUnlock()
	if count != 0 {
		t.Fatalf("expected 0 after clear, got %d", count)
	}
}

func TestStateCache_Protect(t *testing.T) {
	const cap = 4
	sc := newStateCache(cap)

	// Fill cache.
	for i := 0; i < cap; i++ {
		sc.put(types.Hash{byte(i + 1)}, uint64(i), makeTestState(int64(i)))
	}

	// Protect the first entry.
	protectedHash := types.Hash{0x01}
	sc.protect(protectedHash)

	// Add more entries to trigger eviction.
	for i := cap; i < cap+5; i++ {
		sc.put(types.Hash{byte(i + 1)}, uint64(i), makeTestState(int64(i)))
	}

	// Protected entry should still be present.
	_, ok := sc.get(protectedHash)
	if !ok {
		t.Fatal("expected protected entry to be preserved")
	}

	// Cache should not exceed max size.
	sc.mu.RLock()
	count := len(sc.snapshots)
	sc.mu.RUnlock()
	if count > cap {
		t.Fatalf("expected at most %d cached states, got %d", cap, count)
	}
}

func TestStateCache_ClearPreservesProtected(t *testing.T) {
	sc := newStateCache(defaultMaxCachedStates)

	// Add entries.
	for i := 0; i < 10; i++ {
		sc.put(types.Hash{byte(i + 1)}, uint64(i), makeTestState(int64(i)))
	}

	// Protect one entry.
	protectedHash := types.Hash{0x05}
	sc.protect(protectedHash)

	// Clear cache.
	sc.clear()

	// Protected entry should still be present.
	_, ok := sc.get(protectedHash)
	if !ok {
		t.Fatal("expected protected entry to be preserved after clear")
	}

	// Other entries should be gone.
	_, ok = sc.get(types.Hash{0x01})
	if ok {
		t.Fatal("expected non-protected entry to be cleared")
	}
}
