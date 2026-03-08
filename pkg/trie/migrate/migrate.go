package migrate

import (
	"github.com/eth2030/eth2030/crypto"
	"github.com/eth2030/eth2030/trie/mpt"
)

// MigrateFromMPT converts an MPT trie to a binary Merkle trie. Each key-value
// pair from the MPT is re-inserted into the binary trie with the key hashed
// via keccak256 (matching the binary trie's key derivation).
func MigrateFromMPT(source *mpt.Trie) *mpt.BinaryTrie {
	bt := mpt.NewBinaryTrie()
	it := mpt.NewIterator(source)
	for it.Next() {
		hk := crypto.Keccak256Hash(it.Key)
		bt.PutHashed(hk, it.Value)
	}
	return bt
}
