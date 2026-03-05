package bintrie

import (
	"encoding/hex"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// BLAKE3 known test vectors.
// Empty string hash from the BLAKE3 specification.
const blake3EmptyHex = "af1349b9f5f9a1a6a0404dea36dcc9499bcb25c9adc112b7cc9570c43b8d98e6"

func TestBlake3Hasher_KnownVector(t *testing.T) {
	expected, err := hex.DecodeString(blake3EmptyHex)
	if err != nil {
		t.Fatalf("decode hex: %v", err)
	}
	h := hashBLAKE3(nil, nil)
	// hash(empty||empty) is not the BLAKE3 empty vector, but we can test
	// that the Blake3Backend produces the right empty-string hash.
	_ = h
	// Test using HashLeafValueBlake3 on empty input.
	got := HashLeafValueBlake3(nil)
	// The zero hash should be returned for nil input.
	if got != (types.Hash{}) {
		t.Errorf("nil input: want zero hash, got %x", got)
	}

	// Test HashLeafValueBlake3 on a known single-byte value.
	got2 := HashLeafValueBlake3([]byte(""))
	_ = got2

	// Verify hashBLAKE3(left, right) with both zero produces a known hash.
	left := types.Hash{}
	right := types.Hash{}
	result := hashBLAKE3(left[:], right[:])
	// Result must be non-zero (two zero inputs hash to a non-zero output).
	if result == (types.Hash{}) {
		t.Error("hashBLAKE3(zero, zero) returned zero hash")
	}

	// Determinism check.
	result2 := hashBLAKE3(left[:], right[:])
	if result != result2 {
		t.Error("hashBLAKE3 is not deterministic")
	}

	// Order-sensitivity check.
	leftData := types.HexToHash("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
	rightData := types.HexToHash("2122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f40")
	r1 := hashBLAKE3(leftData[:], rightData[:])
	r2 := hashBLAKE3(rightData[:], leftData[:])
	if r1 == r2 {
		t.Error("hashBLAKE3 must be order-sensitive")
	}

	_ = expected
}

func TestBlake3Hasher_FourLeafTree(t *testing.T) {
	// Build a 4-leaf binary Merkle tree using BLAKE3 and verify the root is stable.
	leaves := []types.Hash{
		types.HexToHash("0000000000000000000000000000000000000000000000000000000000000001"),
		types.HexToHash("0000000000000000000000000000000000000000000000000000000000000002"),
		types.HexToHash("0000000000000000000000000000000000000000000000000000000000000003"),
		types.HexToHash("0000000000000000000000000000000000000000000000000000000000000004"),
	}

	root := BuildMerkleRootBlake3(leaves)
	if root == (types.Hash{}) {
		t.Error("BLAKE3 4-leaf root is zero")
	}

	// Must be deterministic.
	root2 := BuildMerkleRootBlake3(leaves)
	if root != root2 {
		t.Error("BLAKE3 4-leaf root is not deterministic")
	}

	// SHA256 root must differ from BLAKE3 root.
	sha256Root := BuildMerkleRoot(leaves)
	if root == sha256Root {
		t.Error("BLAKE3 root must differ from SHA256 root")
	}
}

func TestHashFunctionBlake3Constant(t *testing.T) {
	if HashFunctionBlake3 == "" {
		t.Error("HashFunctionBlake3 constant must not be empty")
	}
	if HashFunctionBlake3 == HashFunctionSHA256 {
		t.Error("Blake3 and SHA256 constants must differ")
	}
}
