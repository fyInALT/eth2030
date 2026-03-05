package pqc

import "testing"

var benchData = []byte("benchmark input for hash function performance comparison in PQC signing")

// BenchmarkSHA256Backend benchmarks SHA-256 hashing.
func BenchmarkSHA256Backend(b *testing.B) {
	backend := &SHA256Backend{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backend.Hash(benchData)
	}
}

// BenchmarkBlake3Backend benchmarks real BLAKE3-256 hashing.
func BenchmarkBlake3Backend(b *testing.B) {
	backend := &Blake3Backend{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backend.Hash(benchData)
	}
}

// BenchmarkKeccak256Backend benchmarks Keccak-256 hashing.
func BenchmarkKeccak256Backend(b *testing.B) {
	backend := &Keccak256Backend{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backend.Hash(benchData)
	}
}
