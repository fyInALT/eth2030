// Package trie re-exports the public API from trie/mpt so that callers
// using "github.com/eth2030/eth2030/trie" continue to work unchanged.
package trie

import (
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/trie/mpt"
)

// Trie is the Merkle Patricia Trie.
type Trie = mpt.Trie

// AccountProof holds Merkle proofs for an account and its storage slots.
type AccountProof = mpt.AccountProof

// StorageProof holds a Merkle proof for a single storage slot.
type StorageProof = mpt.StorageProof

// New creates a new, empty Merkle Patricia Trie.
var New = mpt.New

// ErrNotFound is returned when a key is not present in the trie.
var ErrNotFound = mpt.ErrNotFound

// VerifyProof verifies a Merkle proof for the given key against rootHash.
func VerifyProof(rootHash types.Hash, key []byte, proof [][]byte) ([]byte, error) {
	return mpt.VerifyProof(rootHash, key, proof)
}

// ProveAccountWithStorage generates an account proof and proofs for the given
// storage keys against the provided state and storage tries.
func ProveAccountWithStorage(stateTrie *mpt.Trie, addr types.Address, storageTrie *mpt.Trie, storageKeys []types.Hash) (*mpt.AccountProof, error) {
	return mpt.ProveAccountWithStorage(stateTrie, addr, storageTrie, storageKeys)
}
