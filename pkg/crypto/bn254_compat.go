package crypto

// bn254_compat.go provides backward-compatible re-exports from crypto/bn254.
// Existing code that imports "github.com/eth2030/eth2030/crypto" and uses
// BN254/shielded types continues to work unchanged. New code should import
// "github.com/eth2030/eth2030/crypto/bn254" directly.

import (
	"github.com/eth2030/eth2030/crypto/bn254"
)

// BN254 precompile-compatible function wrappers.

func BN254Add(input []byte) ([]byte, error)          { return bn254.BN254Add(input) }
func BN254ScalarMul(input []byte) ([]byte, error)    { return bn254.BN254ScalarMul(input) }
func BN254PairingCheck(input []byte) ([]byte, error) { return bn254.BN254PairingCheck(input) }
