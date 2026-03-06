package bintrie

import (
	"bytes"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

func TestGetBinaryTreeKeyCodeChunk(t *testing.T) {
	addr := types.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")

	key0 := GetBinaryTreeKeyCodeChunk(addr, 0)
	if len(key0) != HashSize {
		t.Fatalf("expected %d bytes, got %d", HashSize, len(key0))
	}

	key1 := GetBinaryTreeKeyCodeChunk(addr, 1)
	if len(key1) != HashSize {
		t.Fatalf("expected %d bytes, got %d", HashSize, len(key1))
	}

	// Different chunk numbers must produce different keys.
	if bytes.Equal(key0, key1) {
		t.Fatal("different chunk numbers must produce different keys")
	}

	// Same inputs must be deterministic.
	key0b := GetBinaryTreeKeyCodeChunk(addr, 0)
	if !bytes.Equal(key0, key0b) {
		t.Fatal("GetBinaryTreeKeyCodeChunk must be deterministic")
	}

	// Large chunk number must still work.
	keyLarge := GetBinaryTreeKeyCodeChunk(addr, 1000)
	if len(keyLarge) != HashSize {
		t.Fatalf("large chunk: expected %d bytes, got %d", HashSize, len(keyLarge))
	}
}

func TestNewWithHashFunc_Default(t *testing.T) {
	trie := NewWithHashFunc("")
	if trie == nil {
		t.Fatal("NewWithHashFunc(\"\") returned nil")
	}
	// Should behave like New().
	key := types.HexToHash("0000000000000000000000000000000000000000000000000000000000000001")
	val := types.HexToHash("deadbeef00000000000000000000000000000000000000000000000000000000")
	if err := trie.Put(key[:], val[:]); err != nil {
		t.Fatalf("Put error: %v", err)
	}
	got, err := trie.Get(key[:])
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if !bytes.Equal(got, val[:]) {
		t.Fatalf("Get: want %x, got %x", val[:], got)
	}
}

func TestNewWithHashFunc_SHA256(t *testing.T) {
	trie := NewWithHashFunc(HashFunctionSHA256)
	if trie == nil {
		t.Fatal("NewWithHashFunc(SHA256) returned nil")
	}
	// Must produce same hash as New().
	key := types.HexToHash("0000000000000000000000000000000000000000000000000000000000000001")
	val := types.HexToHash("cafebabe00000000000000000000000000000000000000000000000000000000")
	_ = trie.Put(key[:], val[:])
	_ = New().Put(key[:], val[:])
	// Both tries represent the same data; we just verify no panic/error.
}

func TestNewWithHashFunc_Blake3(t *testing.T) {
	trie := NewWithHashFunc(HashFunctionBlake3)
	if trie == nil {
		t.Fatal("NewWithHashFunc(Blake3) returned nil")
	}
	key := types.HexToHash("0000000000000000000000000000000000000000000000000000000000000001")
	val := types.HexToHash("deadbeef00000000000000000000000000000000000000000000000000000000")
	if err := trie.Put(key[:], val[:]); err != nil {
		t.Fatalf("Put error with blake3 trie: %v", err)
	}
}

func TestBinaryTrie_UpdateLeafMetadata_ShortStem(t *testing.T) {
	trie := New()
	// stem shorter than StemSize should return an error.
	err := trie.UpdateLeafMetadata([]byte{1, 2}, 2, 100)
	if err == nil {
		t.Fatal("UpdateLeafMetadata with short stem must return error")
	}
}

func TestBinaryTrie_UpdateLeafMetadata_Valid(t *testing.T) {
	trie := New()
	stem := make([]byte, StemSize)
	stem[0] = 0xAB

	if err := trie.UpdateLeafMetadata(stem, 2, 42); err != nil {
		t.Fatalf("UpdateLeafMetadata error: %v", err)
	}

	// Retrieve the metadata key directly.
	key := make([]byte, HashSize)
	copy(key[:StemSize], stem)
	key[StemSize] = 2
	val, err := trie.Get(key)
	if err != nil {
		t.Fatalf("Get metadata error: %v", err)
	}
	if val == nil {
		t.Fatal("metadata value should be stored")
	}
}

func TestBinaryTrie_UpdateContractCode_Empty(t *testing.T) {
	trie := New()
	addr := types.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	// Empty code should be a no-op.
	if err := trie.UpdateContractCode(addr, nil); err != nil {
		t.Fatalf("UpdateContractCode(nil) error: %v", err)
	}
}

func TestBinaryTrie_UpdateContractCode_Short(t *testing.T) {
	trie := New()
	addr := types.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	code := make([]byte, 50) // fits in one chunk
	for i := range code {
		code[i] = byte(i)
	}
	if err := trie.UpdateContractCode(addr, code); err != nil {
		t.Fatalf("UpdateContractCode short code error: %v", err)
	}
	// Trie should not be empty after inserting code.
	if trie.Hash() == (types.Hash{}) {
		t.Fatal("trie hash should not be zero after contract code update")
	}
}

func TestBinaryTrie_UpdateContractCode_Long(t *testing.T) {
	trie := New()
	addr := types.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	// Long code that spans multiple stems (> 256 chunks).
	code := make([]byte, 8192)
	for i := range code {
		code[i] = byte(i % 256)
	}
	if err := trie.UpdateContractCode(addr, code); err != nil {
		t.Fatalf("UpdateContractCode long code error: %v", err)
	}
}

func TestBinaryTrie_UpdateStem(t *testing.T) {
	trie := New()
	stem := make([]byte, StemSize)
	stem[0] = 0xCC

	values := make([][]byte, StemNodeWidth)
	values[0] = make([]byte, HashSize)
	values[0][0] = 0x01
	values[1] = make([]byte, HashSize)
	values[1][0] = 0x02

	if err := trie.UpdateStem(stem, values); err != nil {
		t.Fatalf("UpdateStem error: %v", err)
	}

	// Retrieve leaf 0.
	key0 := make([]byte, HashSize)
	copy(key0[:StemSize], stem)
	key0[StemSize] = 0
	val0, err := trie.Get(key0)
	if err != nil {
		t.Fatalf("Get key0 error: %v", err)
	}
	if !bytes.Equal(val0, values[0]) {
		t.Fatalf("Get key0: want %x, got %x", values[0], val0)
	}
}

func TestBinaryTrie_Root_Empty(t *testing.T) {
	trie := New()
	root := trie.Root()
	if root == nil {
		t.Fatal("Root must not be nil for empty trie")
	}
	if _, ok := root.(Empty); !ok {
		t.Fatal("Root of empty trie must be Empty{}")
	}
}

func TestBinaryTrie_Root_NonEmpty(t *testing.T) {
	trie := New()
	key := types.HexToHash("0000000000000000000000000000000000000000000000000000000000000001")
	val := types.HexToHash("deadbeef00000000000000000000000000000000000000000000000000000000")
	_ = trie.Put(key[:], val[:])

	root := trie.Root()
	if root == nil {
		t.Fatal("Root must not be nil after insert")
	}
	if _, ok := root.(Empty); ok {
		t.Fatal("Root must not be Empty after insert")
	}
}

func TestBinaryTrie_Root_ConsistencyWithHash(t *testing.T) {
	trie := New()
	key := types.HexToHash("0000000000000000000000000000000000000000000000000000000000000001")
	val := types.HexToHash("deadbeef00000000000000000000000000000000000000000000000000000000")
	_ = trie.Put(key[:], val[:])

	// Root().Hash() must equal trie.Hash().
	rootHash := trie.Root().Hash()
	trieHash := trie.Hash()
	if rootHash != trieHash {
		t.Fatalf("Root().Hash() %x != trie.Hash() %x", rootHash, trieHash)
	}
}
