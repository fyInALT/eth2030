package core

import (
	"errors"
	"math/big"
	"sync"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// TestPaymasterRegistry_EndToEnd runs a full deposit-slash-ban lifecycle.
func TestPaymasterRegistry_EndToEnd(t *testing.T) {
	r := NewPaymasterRegistry(DefaultRegistryConfig())
	testPaymaster := types.BytesToAddress([]byte{0xAA, 0xBB})

	// Deposit 1 ETH at block 1.
	if err := r.Deposit(testPaymaster, new(big.Int).Set(oneETH), 1); err != nil {
		t.Fatalf("Deposit: %v", err)
	}

	// Should be approved.
	if !r.IsApprovedPaymaster(testPaymaster) {
		t.Fatal("expected approved after 1 ETH deposit")
	}

	// Slash 3 times → banned.
	r.Slash(testPaymaster)
	r.Slash(testPaymaster)
	r.Slash(testPaymaster)

	// No longer approved.
	if r.IsApprovedPaymaster(testPaymaster) {
		t.Fatal("expected NOT approved after 3 slashes")
	}

	// Additional slashes must not panic (idempotent after ban).
	r.Slash(testPaymaster)
	r.Slash(testPaymaster)

	e := r.Entry(testPaymaster)
	if e == nil {
		t.Fatal("expected entry, got nil")
	}
	if !e.Banned {
		t.Fatal("expected Banned=true")
	}
	if e.SlashCount < 3 {
		t.Fatalf("expected SlashCount >= 3, got %d", e.SlashCount)
	}
}

// TestPaymasterRegistry_SlashOnBadSettlement verifies the error-returning variant.
func TestPaymasterRegistry_SlashOnBadSettlement(t *testing.T) {
	r := NewPaymasterRegistry(DefaultRegistryConfig())
	addr := types.BytesToAddress([]byte{0xCC, 0xDD})

	if err := r.Deposit(addr, new(big.Int).Set(oneETH), 1); err != nil {
		t.Fatalf("Deposit: %v", err)
	}

	// First slash: not yet at limit (MaxSlashes=3), no error.
	if err := r.SlashOnBadSettlement(addr); err != nil {
		t.Fatalf("SlashOnBadSettlement #1: unexpected error: %v", err)
	}

	// Second slash: still below limit, no error.
	if err := r.SlashOnBadSettlement(addr); err != nil {
		t.Fatalf("SlashOnBadSettlement #2: unexpected error: %v", err)
	}

	// Third slash: reaches MaxSlashes → ErrBannedPaymaster wrapped.
	err := r.SlashOnBadSettlement(addr)
	if err == nil {
		t.Fatal("expected ErrBannedPaymaster on 3rd SlashOnBadSettlement, got nil")
	}
	if !errors.Is(err, ErrBannedPaymaster) {
		t.Fatalf("expected errors.Is(err, ErrBannedPaymaster), got: %v", err)
	}
}

// TestPaymasterRegistry_ConcurrentAccess verifies goroutine safety under the race detector.
func TestPaymasterRegistry_ConcurrentAccess(t *testing.T) {
	r := NewPaymasterRegistry(DefaultRegistryConfig())
	addr := types.BytesToAddress([]byte{0xEE})

	// Seed initial stake so Slash has an entry to work with.
	if err := r.Deposit(addr, new(big.Int).Set(oneETH), 0); err != nil {
		t.Fatalf("initial Deposit: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			switch i % 3 {
			case 0:
				// Deposit a small amount.
				amt := new(big.Int).SetInt64(int64(i + 1))
				_ = r.Deposit(addr, amt, uint64(i))
			case 1:
				r.Slash(addr)
			case 2:
				_ = r.IsApprovedPaymaster(addr)
			}
		}(i)
	}
	wg.Wait()
	// No data races (verified by -race detector) and no panic.
}
