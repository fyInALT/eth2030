package sync

// rangeproof_compat.go re-exports types from sync/rangeproof for backward compatibility.

import (
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/sync/rangeproof"
)

// RangeProof type aliases.
type (
	RangeProof   = rangeproof.RangeProof
	RangeRequest = rangeproof.RangeRequest
	AccountRange = rangeproof.AccountRange
	RangeProver  = rangeproof.RangeProver
)

// RangeProof error variables.
var (
	ErrEmptyRangeProof   = rangeproof.ErrEmptyRangeProof
	ErrUnsortedKeys      = rangeproof.ErrUnsortedKeys
	ErrKeyValueMismatch  = rangeproof.ErrKeyValueMismatch
	ErrInvalidProofRoot  = rangeproof.ErrInvalidProofRoot
	ErrInvalidSplitCount = rangeproof.ErrInvalidSplitCount
	ErrEmptyMerge        = rangeproof.ErrEmptyMerge
)

// RangeProof function wrappers.
func NewRangeProver() *RangeProver { return rangeproof.NewRangeProver() }
func ComputeRangeHash(keys, values [][]byte) types.Hash {
	return rangeproof.ComputeRangeHash(keys, values)
}
func PadTo32(b []byte) []byte { return rangeproof.PadTo32(b) }
