package bintrie

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

func TestComputeEmptyHash(t *testing.T) {
	h := sha256.New()
	h.Write(zero[:])
	h.Write(zero[:])
	expected := types.BytesToHash(h.Sum(nil))

	got := computeEmptyHash()
	if got != expected {
		t.Fatalf("computeEmptyHash: want %x, got %x", expected, got)
	}
}

func TestEmptyHash(t *testing.T) {
	h := EmptyHash()
	if h == (types.Hash{}) {
		t.Fatal("EmptyHash should not be zero (it is SHA256(zero||zero))")
	}
	// Calling twice must return same value.
	if EmptyHash() != h {
		t.Fatal("EmptyHash not stable")
	}
}

func TestEmptyMerkleRoot(t *testing.T) {
	if EmptyMerkleRoot() != EmptyHash() {
		t.Fatal("EmptyMerkleRoot must equal EmptyHash")
	}
}

func TestEmptyProof_ShortKey(t *testing.T) {
	key := []byte{1, 2, 3}
	p := EmptyProof(key)

	if p == nil {
		t.Fatal("EmptyProof returned nil")
	}
	if len(p.Key) != HashSize {
		t.Fatalf("Key length: want %d, got %d", HashSize, len(p.Key))
	}
	// First bytes match the provided key.
	if !bytes.Equal(p.Key[:len(key)], key) {
		t.Fatalf("Key prefix mismatch")
	}
	if p.Value != nil {
		t.Fatal("Value must be nil for empty proof")
	}
	if p.Siblings != nil {
		t.Fatal("Siblings must be nil for empty proof")
	}
	if len(p.Stem) != StemSize {
		t.Fatalf("Stem length: want %d, got %d", StemSize, len(p.Stem))
	}
	if p.LeafIndex != 0 {
		t.Fatalf("LeafIndex: want 0, got %d", p.LeafIndex)
	}
}

func TestEmptyProof_LongKey(t *testing.T) {
	key := make([]byte, HashSize+5)
	for i := range key {
		key[i] = byte(i)
	}
	p := EmptyProof(key)
	if len(p.Key) != HashSize {
		t.Fatalf("Key length: want %d, got %d", HashSize, len(p.Key))
	}
	// Key should be the first HashSize bytes of the input.
	if !bytes.Equal(p.Key, key[:HashSize]) {
		t.Fatalf("Key content mismatch for long input")
	}
}

func TestEmptyProof_ExactKey(t *testing.T) {
	key := make([]byte, HashSize)
	for i := range key {
		key[i] = byte(i + 1)
	}
	p := EmptyProof(key)
	if !bytes.Equal(p.Key, key) {
		t.Fatal("Key must match for exact-size input")
	}
}

func TestIsEmptyNode(t *testing.T) {
	if !IsEmptyNode(Empty{}) {
		t.Fatal("Empty{} must be detected as empty node")
	}

	// Non-empty nodes must return false.
	stem := &StemNode{Stem: make([]byte, StemSize), Values: make([][]byte, StemNodeWidth)}
	if IsEmptyNode(stem) {
		t.Fatal("StemNode must not be detected as empty node")
	}
}

func TestEmptyNodeHash(t *testing.T) {
	h := EmptyNodeHash()
	if h != (types.Hash{}) {
		t.Fatalf("EmptyNodeHash: want zero hash, got %x", h)
	}
}

func TestSerializeEmpty(t *testing.T) {
	b := SerializeEmpty()
	if len(b) != 1 || b[0] != 0 {
		t.Fatalf("SerializeEmpty: want [0x00], got %v", b)
	}
}

func TestDeserializeEmpty_ZeroByte(t *testing.T) {
	e, ok := DeserializeEmpty([]byte{0})
	if !ok {
		t.Fatal("expected ok=true for [0x00]")
	}
	_ = e
}

func TestDeserializeEmpty_EmptySlice(t *testing.T) {
	e, ok := DeserializeEmpty([]byte{})
	if !ok {
		t.Fatal("expected ok=true for empty slice")
	}
	_ = e
}

func TestDeserializeEmpty_NonEmpty(t *testing.T) {
	_, ok := DeserializeEmpty([]byte{1})
	if ok {
		t.Fatal("expected ok=false for non-zero data")
	}

	_, ok2 := DeserializeEmpty([]byte{0, 1})
	if ok2 {
		t.Fatal("expected ok=false for multi-byte data")
	}
}

func TestNewEmptyStateProof(t *testing.T) {
	key := []byte{0xAA, 0xBB, 0xCC}
	p := NewEmptyStateProof(key)

	if !bytes.Equal(p.Key, key) {
		t.Fatal("NewEmptyStateProof: key mismatch")
	}
	if p.RootHash != EmptyMerkleRoot() {
		t.Fatal("NewEmptyStateProof: root hash must equal EmptyMerkleRoot")
	}
	// Mutation of original key must not affect proof.
	key[0] = 0xFF
	if p.Key[0] == 0xFF {
		t.Fatal("NewEmptyStateProof must copy the key")
	}
}

func TestEmptyStateProof_Verify(t *testing.T) {
	key := []byte{1, 2, 3}
	p := NewEmptyStateProof(key)

	if !p.Verify(EmptyMerkleRoot()) {
		t.Fatal("Verify must return true for empty root")
	}

	nonEmptyRoot := types.HexToHash("deadbeef00000000000000000000000000000000000000000000000000000000")
	if p.Verify(nonEmptyRoot) {
		t.Fatal("Verify must return false for non-empty root")
	}
}

func TestGetEmptyTrieStats(t *testing.T) {
	stats := GetEmptyTrieStats()

	if stats.NodeCount != 0 {
		t.Fatalf("NodeCount: want 0, got %d", stats.NodeCount)
	}
	if stats.Height != 0 {
		t.Fatalf("Height: want 0, got %d", stats.Height)
	}
	if stats.MerkleRoot != EmptyMerkleRoot() {
		t.Fatal("MerkleRoot must equal EmptyMerkleRoot")
	}
}

func TestEmptySubtreeHash_ZeroDepth(t *testing.T) {
	h := EmptySubtreeHash(0)
	if h != (types.Hash{}) {
		t.Fatalf("EmptySubtreeHash(0): want zero, got %x", h)
	}
}

func TestEmptySubtreeHash_NegativeDepth(t *testing.T) {
	h := EmptySubtreeHash(-1)
	if h != (types.Hash{}) {
		t.Fatalf("EmptySubtreeHash(-1): want zero, got %x", h)
	}
}

func TestEmptySubtreeHash_Positive(t *testing.T) {
	h1 := EmptySubtreeHash(1)
	h2 := EmptySubtreeHash(2)

	if h1 == (types.Hash{}) {
		t.Fatal("EmptySubtreeHash(1) must not be zero")
	}
	if h2 == (types.Hash{}) {
		t.Fatal("EmptySubtreeHash(2) must not be zero")
	}
	// Each level hashes the previous level with itself.
	if h1 == h2 {
		t.Fatal("different depths must yield different hashes")
	}

	// Manually verify depth=1: SHA256(zero || zero)
	hasher := sha256.New()
	hasher.Write(zero[:])
	hasher.Write(zero[:])
	expected1 := types.BytesToHash(hasher.Sum(nil))
	if h1 != expected1 {
		t.Fatalf("EmptySubtreeHash(1): want %x, got %x", expected1, h1)
	}

	// Depth=2: SHA256(h1 || h1)
	hasher.Reset()
	hasher.Write(h1[:])
	hasher.Write(h1[:])
	expected2 := types.BytesToHash(hasher.Sum(nil))
	if h2 != expected2 {
		t.Fatalf("EmptySubtreeHash(2): want %x, got %x", expected2, h2)
	}
}

func TestEmptySubtreeHash_Deterministic(t *testing.T) {
	for depth := 0; depth <= 5; depth++ {
		h1 := EmptySubtreeHash(depth)
		h2 := EmptySubtreeHash(depth)
		if h1 != h2 {
			t.Fatalf("EmptySubtreeHash(%d) not deterministic", depth)
		}
	}
}
