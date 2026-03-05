package bintrie

import (
	"crypto/sha256"
	"encoding/binary"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// ---- HashedNode tests (hasher.go) ----

func TestHashedNode_Hash(t *testing.T) {
	h := types.HexToHash("deadbeef00000000000000000000000000000000000000000000000000000000")
	hn := HashedNode(h)
	if hn.Hash() != h {
		t.Fatalf("HashedNode.Hash: want %x, got %x", h, hn.Hash())
	}
}

func TestHashedNode_Copy(t *testing.T) {
	h := types.HexToHash("deadbeef00000000000000000000000000000000000000000000000000000000")
	hn := HashedNode(h)
	cp := hn.Copy()
	cpHn, ok := cp.(HashedNode)
	if !ok {
		t.Fatal("Copy must return HashedNode")
	}
	if types.Hash(cpHn) != h {
		t.Fatalf("Copy value mismatch: want %x, got %x", h, types.Hash(cpHn))
	}
}

func TestHashedNode_Get_Error(t *testing.T) {
	hn := HashedNode(types.Hash{})
	_, err := hn.Get(nil, nil)
	if err == nil {
		t.Fatal("HashedNode.Get must return error")
	}
}

func TestHashedNode_Insert_Error(t *testing.T) {
	hn := HashedNode(types.Hash{})
	_, err := hn.Insert(nil, nil, nil, 0)
	if err == nil {
		t.Fatal("HashedNode.Insert must return error")
	}
}

func TestHashedNode_GetValuesAtStem_Error(t *testing.T) {
	hn := HashedNode(types.Hash{})
	_, err := hn.GetValuesAtStem(nil, nil)
	if err == nil {
		t.Fatal("HashedNode.GetValuesAtStem must return error")
	}
}

func TestHashedNode_InsertValuesAtStem_NilResolver(t *testing.T) {
	hn := HashedNode(types.Hash{})
	stem := make([]byte, StemSize)
	_, err := hn.InsertValuesAtStem(stem, nil, nil, 0)
	if err == nil {
		t.Fatal("InsertValuesAtStem with nil resolver must return error")
	}
}

func TestHashedNode_InsertValuesAtStem_WithResolver(t *testing.T) {
	// Build a real stem node, serialize it, then wrap it as a HashedNode
	// and resolve via a fake resolver.
	sn := &StemNode{
		Stem:   make([]byte, StemSize),
		Values: make([][]byte, StemNodeWidth),
		depth:  0,
	}
	data := SerializeNode(sn)

	hn := HashedNode(types.Hash{1})
	resolver := func(_ []byte, _ types.Hash) ([]byte, error) {
		return data, nil
	}
	stem := make([]byte, StemSize)
	vals := make([][]byte, StemNodeWidth)
	vals[0] = make([]byte, HashSize)
	vals[0][0] = 0xAB

	result, err := hn.InsertValuesAtStem(stem, vals, resolver, 0)
	if err != nil {
		t.Fatalf("InsertValuesAtStem with resolver error: %v", err)
	}
	if result == nil {
		t.Fatal("result must not be nil")
	}
}

func TestHashedNode_CollectNodes(t *testing.T) {
	hn := HashedNode(types.Hash{})
	err := hn.CollectNodes(nil, nil)
	if err != nil {
		t.Fatalf("HashedNode.CollectNodes must return nil, got: %v", err)
	}
}

func TestHashedNode_GetHeight_Panics(t *testing.T) {
	hn := HashedNode(types.Hash{})
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("HashedNode.GetHeight must panic")
		}
	}()
	_ = hn.GetHeight()
}

// ---- BinaryHasher tests (hasher_extended.go) ----

func TestNewBinaryHasher(t *testing.T) {
	bh := NewBinaryHasher(5)
	if bh == nil {
		t.Fatal("NewBinaryHasher returned nil")
	}
}

func TestDefaultBinaryHasher(t *testing.T) {
	bh := DefaultBinaryHasher()
	if bh == nil {
		t.Fatal("DefaultBinaryHasher returned nil")
	}
}

func TestBinaryHasher_NilNode(t *testing.T) {
	bh := NewBinaryHasher(0)
	h := bh.Hash(nil)
	if h != (types.Hash{}) {
		t.Fatalf("Hash(nil): want zero, got %x", h)
	}
}

func TestBinaryHasher_EmptyNode(t *testing.T) {
	bh := NewBinaryHasher(0)
	h := bh.Hash(Empty{})
	if h != (types.Hash{}) {
		t.Fatalf("Hash(Empty{}): want zero, got %x", h)
	}
}

func TestBinaryHasher_StemNode(t *testing.T) {
	sn := &StemNode{
		Stem:   make([]byte, StemSize),
		Values: make([][]byte, StemNodeWidth),
		depth:  0,
	}
	sn.Values[0] = make([]byte, HashSize)
	sn.Values[0][0] = 0xAB

	bh := NewBinaryHasher(0)
	h := bh.Hash(sn)
	if h == (types.Hash{}) {
		t.Fatal("Hash(StemNode): should not be zero for non-empty stem")
	}
	// Deterministic.
	h2 := bh.Hash(sn)
	if h != h2 {
		t.Fatal("Hash(StemNode) not deterministic")
	}
}

func TestBinaryHasher_HashedNode(t *testing.T) {
	expected := types.HexToHash("cafebabe00000000000000000000000000000000000000000000000000000000")
	hn := HashedNode(expected)
	bh := NewBinaryHasher(0)
	h := bh.Hash(hn)
	if h != expected {
		t.Fatalf("Hash(HashedNode): want %x, got %x", expected, h)
	}
}

func TestBinaryHasher_InternalNode_Sequential(t *testing.T) {
	tree := New()
	key1 := types.HexToHash("0000000000000000000000000000000000000000000000000000000000000001")
	key2 := types.HexToHash("8000000000000000000000000000000000000000000000000000000000000001")
	val := types.HexToHash("deadbeef00000000000000000000000000000000000000000000000000000000")
	_ = tree.Put(key1[:], val[:])
	_ = tree.Put(key2[:], val[:])

	bh := NewBinaryHasher(0) // no parallel
	h := bh.Hash(tree.Root())
	if h == (types.Hash{}) {
		t.Fatal("Hash(InternalNode): should not be zero")
	}
}

func TestBinaryHasher_InternalNode_Parallel(t *testing.T) {
	tree := New()
	// Add many keys to ensure a tall tree.
	for i := 0; i < 16; i++ {
		var key [32]byte
		binary.BigEndian.PutUint64(key[:], uint64(i)*0x1000000000000000)
		_ = tree.Put(key[:], key[:])
	}

	bhSeq := NewBinaryHasher(0)
	bhPar := NewBinaryHasher(1) // low threshold to trigger parallel
	hSeq := bhSeq.Hash(tree.Root())
	hPar := bhPar.Hash(tree.Root())
	if hSeq != hPar {
		t.Fatalf("sequential and parallel hashes differ: %x vs %x", hSeq, hPar)
	}
}

func TestBinaryHasher_CacheHit(t *testing.T) {
	tree := New()
	key := types.HexToHash("0000000000000000000000000000000000000000000000000000000000000001")
	val := types.HexToHash("deadbeef00000000000000000000000000000000000000000000000000000000")
	_ = tree.Put(key[:], val[:])

	bh := NewBinaryHasher(0)
	h1 := bh.Hash(tree.Root())
	if bh.CacheSize() == 0 {
		t.Fatal("cache should be non-empty after first hash")
	}
	h2 := bh.Hash(tree.Root())
	if h1 != h2 {
		t.Fatal("second Hash must return same result from cache")
	}
}

func TestBinaryHasher_InvalidateCache(t *testing.T) {
	tree := New()
	key := types.HexToHash("0000000000000000000000000000000000000000000000000000000000000001")
	val := types.HexToHash("deadbeef00000000000000000000000000000000000000000000000000000000")
	_ = tree.Put(key[:], val[:])

	bh := NewBinaryHasher(0)
	_ = bh.Hash(tree.Root())
	if bh.CacheSize() == 0 {
		t.Fatal("cache should not be empty")
	}
	bh.InvalidateCache()
	if bh.CacheSize() != 0 {
		t.Fatalf("cache should be empty after InvalidateCache, got %d", bh.CacheSize())
	}
}

func TestBinaryHasher_HashWithStats(t *testing.T) {
	tree := New()
	key1 := types.HexToHash("0000000000000000000000000000000000000000000000000000000000000001")
	key2 := types.HexToHash("8000000000000000000000000000000000000000000000000000000000000001")
	val := types.HexToHash("deadbeef00000000000000000000000000000000000000000000000000000000")
	_ = tree.Put(key1[:], val[:])
	_ = tree.Put(key2[:], val[:])

	bh := NewBinaryHasher(0)
	h, stats := bh.HashWithStats(tree.Root())
	if h == (types.Hash{}) {
		t.Fatal("HashWithStats: expected non-zero hash")
	}
	if stats.NodesHashed == 0 {
		t.Fatal("HashWithStats: NodesHashed should be > 0")
	}
}

func TestBinaryHasher_HashWithStats_CacheHit(t *testing.T) {
	tree := New()
	key := types.HexToHash("0000000000000000000000000000000000000000000000000000000000000001")
	val := types.HexToHash("deadbeef00000000000000000000000000000000000000000000000000000000")
	_ = tree.Put(key[:], val[:])

	bh := NewBinaryHasher(0)
	_, _ = bh.HashWithStats(tree.Root())
	_, stats2 := bh.HashWithStats(tree.Root())
	if stats2.CacheHits == 0 {
		t.Fatal("second HashWithStats should report cache hits")
	}
}

func TestBinaryHasher_HashWithStats_NilNode(t *testing.T) {
	bh := NewBinaryHasher(0)
	h, stats := bh.HashWithStats(nil)
	if h != (types.Hash{}) {
		t.Fatal("HashWithStats(nil): expected zero hash")
	}
	if stats.NodesHashed != 0 || stats.CacheHits != 0 {
		t.Fatal("HashWithStats(nil): expected zero stats")
	}
}

func TestHashLeafValue_Nil(t *testing.T) {
	h := HashLeafValue(nil)
	if h != (types.Hash{}) {
		t.Fatalf("HashLeafValue(nil): want zero, got %x", h)
	}
}

func TestHashLeafValue_NonNil(t *testing.T) {
	data := []byte{1, 2, 3}
	h := HashLeafValue(data)
	expected := sha256.Sum256(data)
	if h != types.BytesToHash(expected[:]) {
		t.Fatalf("HashLeafValue: want %x, got %x", expected, h)
	}
}

func TestHashPair_Deterministic(t *testing.T) {
	left := types.HexToHash("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
	right := types.HexToHash("2122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f40")
	h1 := HashPair(left, right)
	h2 := HashPair(left, right)
	if h1 != h2 {
		t.Fatal("HashPair not deterministic")
	}
	// Order matters.
	h3 := HashPair(right, left)
	if h1 == h3 {
		t.Fatal("HashPair must be order-sensitive")
	}
}

func TestHashPair_KnownValue(t *testing.T) {
	left := types.Hash{}
	right := types.Hash{}
	got := HashPair(left, right)
	h := sha256.New()
	h.Write(left[:])
	h.Write(right[:])
	expected := types.BytesToHash(h.Sum(nil))
	if got != expected {
		t.Fatalf("HashPair(zero, zero): want %x, got %x", expected, got)
	}
}

func TestBuildMerkleRoot_Empty(t *testing.T) {
	h := BuildMerkleRoot(nil)
	if h != (types.Hash{}) {
		t.Fatalf("BuildMerkleRoot(nil): want zero, got %x", h)
	}
}

func TestBuildMerkleRoot_SingleLeaf(t *testing.T) {
	leaf := types.HexToHash("deadbeef00000000000000000000000000000000000000000000000000000000")
	h := BuildMerkleRoot([]types.Hash{leaf})
	// n starts at 1; 1 < 1 is false so no padding occurs.
	// padded has length 1 and the reduction loop never runs.
	// Result is the leaf itself.
	if h != leaf {
		t.Fatalf("BuildMerkleRoot([leaf]): want leaf %x, got %x", leaf, h)
	}
}

func TestBuildMerkleRoot_TwoLeaves(t *testing.T) {
	l1 := types.HexToHash("0000000000000000000000000000000000000000000000000000000000000001")
	l2 := types.HexToHash("0000000000000000000000000000000000000000000000000000000000000002")
	h := BuildMerkleRoot([]types.Hash{l1, l2})
	expected := HashPair(l1, l2)
	if h != expected {
		t.Fatalf("BuildMerkleRoot([l1,l2]): want %x, got %x", expected, h)
	}
}

func TestBuildMerkleRoot_FourLeaves(t *testing.T) {
	leaves := []types.Hash{
		types.HexToHash("0000000000000000000000000000000000000000000000000000000000000001"),
		types.HexToHash("0000000000000000000000000000000000000000000000000000000000000002"),
		types.HexToHash("0000000000000000000000000000000000000000000000000000000000000003"),
		types.HexToHash("0000000000000000000000000000000000000000000000000000000000000004"),
	}
	h := BuildMerkleRoot(leaves)
	if h == (types.Hash{}) {
		t.Fatal("BuildMerkleRoot(4 leaves): must not be zero")
	}
	// Verify determinism.
	if BuildMerkleRoot(leaves) != h {
		t.Fatal("BuildMerkleRoot not deterministic")
	}
}
