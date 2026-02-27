package pqc

import (
	"crypto/sha256"

	"github.com/eth2030/eth2030/crypto"
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

// Blake3Backend implements a minimal BLAKE3-like hash using iterative SHA-256 mixing.
// NOTE: This is a structural placeholder; production use should integrate
// lukechampine.com/blake3 or zeebo/blake3 via go.mod.
type Blake3Backend struct{}

func (b *Blake3Backend) Hash(data []byte) [32]byte {
	// Iterative mixing: hash with domain separation to approximate BLAKE3 structure.
	// Round 1: SHA256("blake3-r1" || data)
	h1 := sha256.New()
	h1.Write([]byte("blake3-r1"))
	h1.Write(data)
	var r1 [32]byte
	copy(r1[:], h1.Sum(nil))
	// Round 2: SHA256("blake3-r2" || r1)
	h2 := sha256.New()
	h2.Write([]byte("blake3-r2"))
	h2.Write(r1[:])
	var result [32]byte
	copy(result[:], h2.Sum(nil))
	return result
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
