package bintrie

import (
	"encoding/binary"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

func TestBinaryTrieEpochUpdater_ImplementsInterface(t *testing.T) {
	trie := New()
	var _ TrieEpochUpdater = &BinaryTrieEpochUpdater{Trie: trie}
}

func TestBinaryTrieEpochUpdater_UpdateAccountEpoch(t *testing.T) {
	trie := New()
	updater := &BinaryTrieEpochUpdater{Trie: trie}

	addr := types.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	epoch := uint64(12345)

	updater.UpdateAccountEpoch(addr, epoch)

	// Verify that the trie now contains a value for the metadata key.
	stem := GetBinaryTreeKeyBasicData(addr)
	key := make([]byte, HashSize)
	copy(key[:StemSize], stem[:StemSize])
	key[StemSize] = byte(2) // subindex 2

	val, err := trie.Get(key)
	if err != nil {
		t.Fatalf("Get epoch key error: %v", err)
	}
	if val == nil {
		t.Fatal("epoch value should be stored")
	}
	// Epoch is stored as big-endian uint64 in the last 8 bytes.
	got := binary.BigEndian.Uint64(val[HashSize-8:])
	if got != epoch {
		t.Fatalf("epoch mismatch: want %d, got %d", epoch, got)
	}
}

func TestBinaryTrieEpochUpdater_UpdateStorageEpoch(t *testing.T) {
	trie := New()
	updater := &BinaryTrieEpochUpdater{Trie: trie}

	addr := types.HexToAddress("0xabcdef0123456789abcdef0123456789abcdef01")
	storageKey := types.HexToHash("0000000000000000000000000000000000000000000000000000000000000001")
	epoch := uint64(99999)

	updater.UpdateStorageEpoch(addr, storageKey, epoch)

	// Verify the epoch was stored.
	stem := GetBinaryTreeKeyStorageSlot(addr, storageKey[:])
	key := make([]byte, HashSize)
	copy(key[:StemSize], stem[:StemSize])
	key[StemSize] = byte(2)

	val, err := trie.Get(key)
	if err != nil {
		t.Fatalf("Get storage epoch key error: %v", err)
	}
	if val == nil {
		t.Fatal("storage epoch value should be stored")
	}
	got := binary.BigEndian.Uint64(val[HashSize-8:])
	if got != epoch {
		t.Fatalf("storage epoch mismatch: want %d, got %d", epoch, got)
	}
}

func TestBinaryTrieEpochUpdater_UpdateEpochOverwrite(t *testing.T) {
	trie := New()
	updater := &BinaryTrieEpochUpdater{Trie: trie}
	addr := types.HexToAddress("0x1111111111111111111111111111111111111111")

	updater.UpdateAccountEpoch(addr, 100)
	updater.UpdateAccountEpoch(addr, 200)

	stem := GetBinaryTreeKeyBasicData(addr)
	key := make([]byte, HashSize)
	copy(key[:StemSize], stem[:StemSize])
	key[StemSize] = 2

	val, err := trie.Get(key)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	got := binary.BigEndian.Uint64(val[HashSize-8:])
	if got != 200 {
		t.Fatalf("expected overwritten epoch 200, got %d", got)
	}
}

func TestBinaryTrieEpochUpdater_ZeroEpoch(t *testing.T) {
	trie := New()
	updater := &BinaryTrieEpochUpdater{Trie: trie}
	addr := types.HexToAddress("0x2222222222222222222222222222222222222222")

	// Should not panic.
	updater.UpdateAccountEpoch(addr, 0)
}

func TestBinaryTrieEpochUpdater_MultipleAddresses(t *testing.T) {
	trie := New()
	updater := &BinaryTrieEpochUpdater{Trie: trie}

	addrs := []types.Address{
		types.HexToAddress("0xAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
		types.HexToAddress("0xBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB"),
		types.HexToAddress("0xCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC"),
	}
	epochs := []uint64{1, 2, 3}

	for i, addr := range addrs {
		updater.UpdateAccountEpoch(addr, epochs[i])
	}

	for i, addr := range addrs {
		stem := GetBinaryTreeKeyBasicData(addr)
		key := make([]byte, HashSize)
		copy(key[:StemSize], stem[:StemSize])
		key[StemSize] = 2

		val, err := trie.Get(key)
		if err != nil {
			t.Fatalf("Get addr[%d] error: %v", i, err)
		}
		got := binary.BigEndian.Uint64(val[HashSize-8:])
		if got != epochs[i] {
			t.Fatalf("addr[%d]: want epoch %d, got %d", i, epochs[i], got)
		}
	}
}
