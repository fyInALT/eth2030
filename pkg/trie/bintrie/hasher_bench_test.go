package bintrie

import (
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// BenchmarkHashPairSHA256 benchmarks SHA-256 binary trie internal-node hashing.
func BenchmarkHashPairSHA256(b *testing.B) {
	left := types.HexToHash("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
	right := types.HexToHash("2122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f40")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HashPair(left, right)
	}
}

// BenchmarkHashPairBlake3 benchmarks BLAKE3-256 binary trie internal-node hashing.
func BenchmarkHashPairBlake3(b *testing.B) {
	left := types.HexToHash("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
	right := types.HexToHash("2122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f40")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HashPairBlake3(left, right)
	}
}

// BenchmarkBuildMerkleRootSHA256 benchmarks building a 1024-leaf SHA-256 trie root.
func BenchmarkBuildMerkleRootSHA256(b *testing.B) {
	leaves := make([]types.Hash, 1024)
	for i := range leaves {
		leaves[i][0] = byte(i)
		leaves[i][1] = byte(i >> 8)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildMerkleRoot(leaves)
	}
}

// BenchmarkBuildMerkleRootBlake3 benchmarks building a 1024-leaf BLAKE3 trie root.
func BenchmarkBuildMerkleRootBlake3(b *testing.B) {
	leaves := make([]types.Hash, 1024)
	for i := range leaves {
		leaves[i][0] = byte(i)
		leaves[i][1] = byte(i >> 8)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildMerkleRootBlake3(leaves)
	}
}
