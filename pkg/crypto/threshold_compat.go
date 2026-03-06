package crypto

// threshold_compat.go provides backward-compatible re-exports from crypto/threshold.
// Existing code that imports "github.com/eth2030/eth2030/crypto" and uses
// threshold types continues to work unchanged. New code should import
// "github.com/eth2030/eth2030/crypto/threshold" directly.

import (
	"math/big"

	"github.com/eth2030/eth2030/crypto/threshold"
)

// Threshold type aliases — keep crypto.XXX working for all existing importers.
type (
	ThresholdScheme  = threshold.ThresholdScheme
	Share            = threshold.Share
	VerifiableShare  = threshold.VerifiableShare
	KeyGenResult     = threshold.KeyGenResult
	EncryptedMessage = threshold.EncryptedMessage
	DecryptionShare  = threshold.DecryptionShare
)

// Threshold error aliases.
var (
	ErrInvalidThreshold   = threshold.ErrInvalidThreshold
	ErrInsufficientShares = threshold.ErrInsufficientShares
	ErrDuplicateShareIndex = threshold.ErrDuplicateShareIndex
	ErrInvalidShare       = threshold.ErrInvalidShare
	ErrDecryptionFailed   = threshold.ErrDecryptionFailed
	ErrInvalidCiphertext  = threshold.ErrInvalidCiphertext
)

// Threshold function wrappers.

// NewThresholdScheme creates a new threshold scheme with the given parameters.
func NewThresholdScheme(t, n int) (*ThresholdScheme, error) {
	return threshold.NewThresholdScheme(t, n)
}

// VerifyShare checks that a share is consistent with the Feldman VSS commitments.
func VerifyShare(share Share, commitments []*big.Int) bool {
	return threshold.VerifyShare(share, commitments)
}

// MakeVerifiableShare bundles a share with its VSS commitments.
func MakeVerifiableShare(share Share, commitments []*big.Int) VerifiableShare {
	return threshold.MakeVerifiableShare(share, commitments)
}

// ShareEncrypt encrypts a message under the given public key.
func ShareEncrypt(publicKey *big.Int, message []byte) (*EncryptedMessage, error) {
	return threshold.ShareEncrypt(publicKey, message)
}

// ShareDecrypt produces a decryption share for an encrypted message.
func ShareDecrypt(share Share, ephemeral *big.Int) DecryptionShare {
	return threshold.ShareDecrypt(share, ephemeral)
}

// CombineShares reconstructs the plaintext from threshold-many decryption shares.
func CombineShares(shares []DecryptionShare, encrypted *EncryptedMessage) ([]byte, error) {
	return threshold.CombineShares(shares, encrypted)
}

// LagrangeInterpolate reconstructs the secret from t shares via Lagrange interpolation.
func LagrangeInterpolate(shares []Share) (*big.Int, error) {
	return threshold.LagrangeInterpolate(shares)
}
