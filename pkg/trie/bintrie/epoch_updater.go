// epoch_updater.go provides helpers for wiring binary trie epoch tracking
// into the state expiry mechanism (BL-1.4, I+ roadmap).
package bintrie

import "github.com/eth2030/eth2030/core/types"

// TrieEpochUpdater is the interface used by StateExpiryManager to record
// last-access epochs into the binary trie. Implementations call
// BinaryTrie.UpdateLeafMetadata(stem, 2, epoch).
type TrieEpochUpdater interface {
	UpdateAccountEpoch(addr types.Address, epoch uint64)
	UpdateStorageEpoch(addr types.Address, key types.Hash, epoch uint64)
}

// BinaryTrieEpochUpdater adapts a *BinaryTrie to implement TrieEpochUpdater.
// It writes epochs into metadata subindex 2 of the stem node for each address.
type BinaryTrieEpochUpdater struct {
	Trie *BinaryTrie
}

// UpdateAccountEpoch writes the epoch into subindex 2 of the stem node
// corresponding to addr's basic-data leaf.
func (u *BinaryTrieEpochUpdater) UpdateAccountEpoch(addr types.Address, epoch uint64) {
	stem := GetBinaryTreeKeyBasicData(addr)
	// stem[:StemSize] is the 31-byte stem prefix.
	_ = u.Trie.UpdateLeafMetadata(stem, 2, epoch)
}

// UpdateStorageEpoch writes the epoch into subindex 2 of the stem node
// corresponding to the given storage key.
func (u *BinaryTrieEpochUpdater) UpdateStorageEpoch(addr types.Address, key types.Hash, epoch uint64) {
	stem := GetBinaryTreeKeyStorageSlot(addr, key[:])
	_ = u.Trie.UpdateLeafMetadata(stem, 2, epoch)
}
