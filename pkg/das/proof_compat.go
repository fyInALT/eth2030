package das

// proof_compat.go re-exports types, functions, and variables from
// das/proof for backward compatibility with existing callers.

import "github.com/eth2030/eth2030/das/proof"

// Error variables re-exported from das/proof.
var (
	ErrBlockBlobTooLarge       = proof.ErrBlockBlobTooLarge
	ErrBlockBlobInvalidProof   = proof.ErrBlockBlobInvalidProof
	ErrBlockBlobEncodingFailed = proof.ErrBlockBlobEncodingFailed
)

// Type aliases re-exported from das/proof.
type (
	BlockBlobProverConfig = proof.BlockBlobProverConfig
	BlockBlobEncoding     = proof.BlockBlobEncoding
	BlockBlobProof        = proof.BlockBlobProof
	BlockBlobEncodingMeta = proof.BlockBlobEncodingMeta
	BlockBlobProver       = proof.BlockBlobProver
)

// Function aliases re-exported from das/proof.
var (
	DefaultBlockBlobProverConfig = proof.DefaultBlockBlobProverConfig
	NewBlockBlobProver           = proof.NewBlockBlobProver
)
