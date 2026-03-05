package rawdb

import (
	"errors"
	"testing"
)

func makeTestHash(n byte) [32]byte {
	var h [32]byte
	h[0] = n
	return h
}

// writeCanonicalBlock writes canonical hash + header-number mapping for a block.
// WriteHeader stores both the header data and the hash→number reverse index.
func writeCanonicalBlock(db Database, num uint64, hash [32]byte) {
	WriteCanonicalHash(db, num, hash)
	WriteHeader(db, num, hash, []byte("hdr"))
}

// --- ChainIterator ---

func TestForwardIterator_Basic(t *testing.T) {
	db := NewMemoryDB()
	for i := uint64(0); i < 5; i++ {
		writeCanonicalBlock(db, i, makeTestHash(byte(i+1)))
	}

	it, err := NewForwardIterator(db, 0, 4)
	if err != nil {
		t.Fatal(err)
	}
	defer it.Close()

	var nums []uint64
	for it.Next() {
		nums = append(nums, it.Number())
	}
	if len(nums) != 5 {
		t.Fatalf("expected 5 blocks, got %d", len(nums))
	}
	for i, n := range nums {
		if n != uint64(i) {
			t.Fatalf("nums[%d] = %d, want %d", i, n, i)
		}
	}
}

func TestForwardIterator_InvalidRange(t *testing.T) {
	db := NewMemoryDB()
	_, err := NewForwardIterator(db, 5, 3)
	if !errors.Is(err, ErrRangeInvalid) {
		t.Fatalf("expected ErrRangeInvalid, got %v", err)
	}
}

func TestForwardIterator_ClosePreventsNext(t *testing.T) {
	db := NewMemoryDB()
	writeCanonicalBlock(db, 0, makeTestHash(1))
	it, _ := NewForwardIterator(db, 0, 0)
	it.Close()
	if it.Next() {
		t.Fatal("Next should return false after Close")
	}
}

func TestForwardIterator_Hash(t *testing.T) {
	db := NewMemoryDB()
	h := makeTestHash(0xAB)
	writeCanonicalBlock(db, 0, h)

	it, _ := NewForwardIterator(db, 0, 0)
	it.Next()
	if it.Hash() != h {
		t.Fatalf("Hash mismatch: got %x, want %x", it.Hash(), h)
	}
}

func TestForwardIterator_MissingBlockSkipped(t *testing.T) {
	db := NewMemoryDB()
	// Only write blocks 0 and 2 — skip 1.
	writeCanonicalBlock(db, 0, makeTestHash(1))
	writeCanonicalBlock(db, 2, makeTestHash(3))

	it, _ := NewForwardIterator(db, 0, 2)
	defer it.Close()

	var nums []uint64
	for it.Next() {
		nums = append(nums, it.Number())
	}
	// Block 1 missing: iterator stops at 0, then tries 1 (missing → false).
	if len(nums) == 0 {
		t.Fatal("expected at least block 0")
	}
	if nums[0] != 0 {
		t.Fatalf("expected first block 0, got %d", nums[0])
	}
}

func TestBackwardIterator_Basic(t *testing.T) {
	db := NewMemoryDB()
	for i := uint64(0); i < 5; i++ {
		writeCanonicalBlock(db, i, makeTestHash(byte(i+1)))
	}

	it, err := NewBackwardIterator(db, 4, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer it.Close()

	var nums []uint64
	for it.Next() {
		nums = append(nums, it.Number())
	}
	if len(nums) == 0 {
		t.Fatal("expected blocks")
	}
	if nums[0] != 4 {
		t.Fatalf("expected first block 4, got %d", nums[0])
	}
}

func TestBackwardIterator_InvalidRange(t *testing.T) {
	db := NewMemoryDB()
	_, err := NewBackwardIterator(db, 2, 5)
	if !errors.Is(err, ErrRangeInvalid) {
		t.Fatalf("expected ErrRangeInvalid, got %v", err)
	}
}

// --- ReadCanonicalRange ---

func TestReadCanonicalRange_Basic(t *testing.T) {
	db := NewMemoryDB()
	for i := uint64(0); i < 4; i++ {
		writeCanonicalBlock(db, i, makeTestHash(byte(i+1)))
	}

	blocks, err := ReadCanonicalRange(db, 0, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 4 {
		t.Fatalf("expected 4 blocks, got %d", len(blocks))
	}
	for i, b := range blocks {
		if b.Number != uint64(i) {
			t.Fatalf("blocks[%d].Number = %d, want %d", i, b.Number, i)
		}
	}
}

func TestReadCanonicalRange_InvalidRange(t *testing.T) {
	db := NewMemoryDB()
	_, err := ReadCanonicalRange(db, 5, 3)
	if !errors.Is(err, ErrRangeInvalid) {
		t.Fatalf("expected ErrRangeInvalid, got %v", err)
	}
}

func TestReadCanonicalRange_SkipsMissing(t *testing.T) {
	db := NewMemoryDB()
	writeCanonicalBlock(db, 0, makeTestHash(1))
	writeCanonicalBlock(db, 2, makeTestHash(3))

	blocks, err := ReadCanonicalRange(db, 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks (skip missing 1), got %d", len(blocks))
	}
}

// --- ReadBlockNumberByHash ---

func TestReadBlockNumberByHash(t *testing.T) {
	db := NewMemoryDB()
	h := makeTestHash(0x7)
	WriteHeader(db, 42, h, []byte("hdr"))

	num, err := ReadBlockNumberByHash(db, h)
	if err != nil {
		t.Fatal(err)
	}
	if num != 42 {
		t.Fatalf("expected 42, got %d", num)
	}
}

func TestReadBlockNumberByHash_NotFound(t *testing.T) {
	db := NewMemoryDB()
	_, err := ReadBlockNumberByHash(db, makeTestHash(0xFF))
	if err == nil {
		t.Fatal("expected error for missing hash")
	}
}

// --- IsCanonicalHash ---

func TestIsCanonicalHash_True(t *testing.T) {
	db := NewMemoryDB()
	h := makeTestHash(0x1)
	writeCanonicalBlock(db, 10, h)

	ok, err := IsCanonicalHash(db, h)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected canonical")
	}
}

func TestIsCanonicalHash_False(t *testing.T) {
	db := NewMemoryDB()
	canonical := makeTestHash(0xAA)
	other := makeTestHash(0xBB)
	writeCanonicalBlock(db, 10, canonical)
	// other hash points to block 10 but is not canonical.
	WriteHeader(db, 10, other, []byte("hdr-other"))

	ok, err := IsCanonicalHash(db, other)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected non-canonical")
	}
}

func TestIsCanonicalHash_NotFound(t *testing.T) {
	db := NewMemoryDB()
	_, err := IsCanonicalHash(db, makeTestHash(0xFF))
	if err == nil {
		t.Fatal("expected error for unknown hash")
	}
}

// --- FindCommonAncestor ---

func TestFindCommonAncestor_FoundAtFirst(t *testing.T) {
	db := NewMemoryDB()
	h := makeTestHash(0x05)
	writeCanonicalBlock(db, 5, h)

	num, got, err := FindCommonAncestor(db, 5, func(n uint64) ([32]byte, error) {
		if n == 5 {
			return h, nil
		}
		return [32]byte{}, errors.New("not found")
	})
	if err != nil {
		t.Fatal(err)
	}
	if num != 5 || got != h {
		t.Fatalf("expected num=5 h=%x, got num=%d h=%x", h, num, got)
	}
}

func TestFindCommonAncestor_NotFound(t *testing.T) {
	db := NewMemoryDB()
	writeCanonicalBlock(db, 0, makeTestHash(0x01))

	_, _, err := FindCommonAncestor(db, 0, func(n uint64) ([32]byte, error) {
		return makeTestHash(0xFF), nil // always mismatch
	})
	if err == nil {
		t.Fatal("expected error when no common ancestor")
	}
}

// --- BuildNumberIndex ---

func TestBuildNumberIndex_Basic(t *testing.T) {
	db := NewMemoryDB()
	hashes := map[uint64][32]byte{
		0: makeTestHash(1),
		1: makeTestHash(2),
		2: makeTestHash(3),
	}
	count, err := BuildNumberIndex(db, 0, 2, func(n uint64) ([32]byte, bool) {
		h, ok := hashes[n]
		return h, ok
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("expected 3, got %d", count)
	}
	// Verify we can read back.
	for num, h := range hashes {
		got, err := ReadCanonicalHash(db, num)
		if err != nil || got != h {
			t.Fatalf("block %d: got %x want %x (err=%v)", num, got, h, err)
		}
	}
}

func TestBuildNumberIndex_InvalidRange(t *testing.T) {
	db := NewMemoryDB()
	_, err := BuildNumberIndex(db, 5, 3, func(n uint64) ([32]byte, bool) { return [32]byte{}, false })
	if !errors.Is(err, ErrRangeInvalid) {
		t.Fatalf("expected ErrRangeInvalid, got %v", err)
	}
}

func TestBuildNumberIndex_SkipsMissing(t *testing.T) {
	db := NewMemoryDB()
	count, err := BuildNumberIndex(db, 0, 2, func(n uint64) ([32]byte, bool) {
		if n == 1 {
			return [32]byte{}, false // skip block 1
		}
		return makeTestHash(byte(n + 1)), true
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("expected 2 (skip 1), got %d", count)
	}
}

// --- DeleteCanonicalRange ---

func TestDeleteCanonicalRange_Basic(t *testing.T) {
	db := NewMemoryDB()
	for i := uint64(0); i < 5; i++ {
		writeCanonicalBlock(db, i, makeTestHash(byte(i+1)))
	}

	if err := DeleteCanonicalRange(db, 2, 4); err != nil {
		t.Fatal(err)
	}
	// Blocks 0-1 should still exist.
	if _, err := ReadCanonicalHash(db, 0); err != nil {
		t.Fatal("block 0 should still exist")
	}
	// Block 2 should be gone.
	if _, err := ReadCanonicalHash(db, 2); err == nil {
		t.Fatal("block 2 should be deleted")
	}
}

func TestDeleteCanonicalRange_InvalidRange(t *testing.T) {
	db := NewMemoryDB()
	err := DeleteCanonicalRange(db, 5, 3)
	if !errors.Is(err, ErrRangeInvalid) {
		t.Fatalf("expected ErrRangeInvalid, got %v", err)
	}
}

// --- CountCanonicalBlocks ---

func TestCountCanonicalBlocks_Basic(t *testing.T) {
	db := NewMemoryDB()
	writeCanonicalBlock(db, 0, makeTestHash(1))
	writeCanonicalBlock(db, 1, makeTestHash(2))
	// skip 2
	writeCanonicalBlock(db, 3, makeTestHash(4))

	count, err := CountCanonicalBlocks(db, 0, 3)
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("expected 3, got %d", count)
	}
}

func TestCountCanonicalBlocks_InvalidRange(t *testing.T) {
	db := NewMemoryDB()
	_, err := CountCanonicalBlocks(db, 5, 3)
	if !errors.Is(err, ErrRangeInvalid) {
		t.Fatalf("expected ErrRangeInvalid, got %v", err)
	}
}

func TestCountCanonicalBlocks_Empty(t *testing.T) {
	db := NewMemoryDB()
	count, err := CountCanonicalBlocks(db, 0, 9)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected 0, got %d", count)
	}
}

// --- WalkCanonicalChain ---

func TestWalkCanonicalChain_Basic(t *testing.T) {
	db := NewMemoryDB()
	for i := uint64(0); i < 4; i++ {
		writeCanonicalBlock(db, i, makeTestHash(byte(i+1)))
	}

	var visited []uint64
	err := WalkCanonicalChain(db, 0, 3, func(num uint64, _ [32]byte) error {
		visited = append(visited, num)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(visited) != 4 {
		t.Fatalf("expected 4 blocks visited, got %d", len(visited))
	}
}

func TestWalkCanonicalChain_StopsOnError(t *testing.T) {
	db := NewMemoryDB()
	writeCanonicalBlock(db, 0, makeTestHash(1))
	writeCanonicalBlock(db, 1, makeTestHash(2))

	count := 0
	err := WalkCanonicalChain(db, 0, 1, func(num uint64, _ [32]byte) error {
		count++
		return errors.New("stop")
	})
	if err == nil {
		t.Fatal("expected error from fn")
	}
	if count != 1 {
		t.Fatalf("expected fn called once, got %d", count)
	}
}

func TestWalkCanonicalChain_InvalidRange(t *testing.T) {
	db := NewMemoryDB()
	err := WalkCanonicalChain(db, 5, 3, nil)
	if !errors.Is(err, ErrRangeInvalid) {
		t.Fatalf("expected ErrRangeInvalid, got %v", err)
	}
}
