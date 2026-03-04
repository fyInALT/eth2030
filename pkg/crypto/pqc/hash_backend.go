package pqc

import (
	"crypto/sha256"

	"github.com/eth2030/eth2030/crypto"
	"lukechampine.com/blake3"
)

// HashBackend is a pluggable hash function interface for hash-based signatures.
// It allows swapping the underlying hash function (Keccak256, SHA-256, BLAKE3, Poseidon2)
// without changing signature scheme logic.
type HashBackend interface {
	// Hash computes a 32-byte digest of the input.
	Hash(data []byte) [32]byte
	// Name returns the hash function name (e.g., "keccak256", "sha256", "blake3").
	Name() string
	// BlockSize returns the hash function block size in bytes.
	BlockSize() int
}

// Keccak256Backend wraps the existing Keccak256 implementation.
type Keccak256Backend struct{}

func (k *Keccak256Backend) Hash(data []byte) [32]byte {
	h := crypto.Keccak256(data)
	var result [32]byte
	copy(result[:], h)
	return result
}
func (k *Keccak256Backend) Name() string   { return "keccak256" }
func (k *Keccak256Backend) BlockSize() int { return 136 }

// SHA256Backend wraps crypto/sha256.
type SHA256Backend struct{}

func (s *SHA256Backend) Hash(data []byte) [32]byte {
	return sha256.Sum256(data)
}
func (s *SHA256Backend) Name() string   { return "sha256" }
func (s *SHA256Backend) BlockSize() int { return 64 }

// Blake3Backend uses the real BLAKE3-256 hash (lukechampine.com/blake3).
// BLAKE3 is a fast, parallel cryptographic hash function that is
// 2–4× faster than SHA-256 on modern CPUs.
type Blake3Backend struct{}

func (b *Blake3Backend) Hash(data []byte) [32]byte {
	return blake3.Sum256(data)
}
func (b *Blake3Backend) Name() string   { return "blake3" }
func (b *Blake3Backend) BlockSize() int { return 64 }

// DefaultHashBackend returns the Keccak256 backend (Ethereum's current default).
func DefaultHashBackend() HashBackend {
	return &Keccak256Backend{}
}

// HashBackendByName returns a HashBackend by name, or nil if unknown.
func HashBackendByName(name string) HashBackend {
	switch name {
	case "keccak256":
		return &Keccak256Backend{}
	case "sha256":
		return &SHA256Backend{}
	case "blake3":
		return &Blake3Backend{}
	default:
		return nil
	}
}
