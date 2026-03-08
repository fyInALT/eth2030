package announce

import (
	"errors"

	"github.com/eth2030/eth2030/core/types"
)

// ErrNotFound is returned when a key is not found in the trie.
var ErrNotFound = errors.New("trie: key not found")

// getBit returns bit pos (0-indexed, MSB-first) of hash h.
func getBit(h types.Hash, pos int) byte {
	byteIdx := pos / 8
	bitIdx := uint(7 - pos%8)
	return (h[byteIdx] >> bitIdx) & 1
}

// copyBytes returns a copy of the byte slice.
func copyBytes(b []byte) []byte {
	if len(b) == 0 {
		return nil
	}
	cp := make([]byte, len(b))
	copy(cp, b)
	return cp
}
