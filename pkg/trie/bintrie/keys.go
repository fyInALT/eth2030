// keys.go implements EIP-7864 binary trie key generation functions.
// All state (account data, code, storage) is embedded in a single key/value
// space using 32-byte keys produced by these functions.
package bintrie

import (
	"math/big"

	"github.com/eth2030/eth2030/core/types"
	"lukechampine.com/blake3"
)

// EIP-7864 tree embedding constants.
// BasicDataLeafKey (0) and CodeHashLeafKey (1) are defined in bintrie.go.
// StemNodeWidth (256) and StemSize (31) are defined in node.go.
const (
	// HeaderStorageOffset is the starting sub_index for inline storage slots 0–63.
	HeaderStorageOffset = 64
	// CodeOffset is the starting sub_index for code chunks in the account stem.
	CodeOffset = 128
)

// treeMainStorageOffset is MAIN_STORAGE_OFFSET = 256^31 per EIP-7864.
var treeMainStorageOffset *big.Int

func init() {
	treeMainStorageOffset = new(big.Int).Exp(big.NewInt(256), big.NewInt(31), nil)
}

// GetTreeKey returns the 32-byte tree key for (address32, tree_index, sub_index).
// key = blake3(address32 || tree_index.to_bytes(32, "little"))[:31] || sub_index
// Per EIP-7864 §key-generation.
func GetTreeKey(address32 [32]byte, treeIndex *big.Int, subIndex uint8) [32]byte {
	// Encode tree_index as 32-byte little-endian.
	var tiBytes [32]byte
	if treeIndex != nil && treeIndex.Sign() != 0 {
		raw := treeIndex.Bytes() // big-endian
		// Reverse to little-endian.
		for i := 0; i < len(raw) && i < 32; i++ {
			tiBytes[i] = raw[len(raw)-1-i]
		}
	}

	h := blake3.New(32, nil)
	h.Write(address32[:])
	h.Write(tiBytes[:])

	var out [32]byte
	digest := h.Sum(nil)
	copy(out[:31], digest[:31])
	out[31] = subIndex
	return out
}

// GetTreeKeyForBasicData returns the tree key for an account's basic data leaf
// (version, nonce, balance, code_size packed at offsets per spec).
func GetTreeKeyForBasicData(addr types.Address) [32]byte {
	return GetTreeKey(addrTo32(addr), big.NewInt(0), BasicDataLeafKey)
}

// GetTreeKeyForCodeHash returns the tree key for an account's code hash leaf.
func GetTreeKeyForCodeHash(addr types.Address) [32]byte {
	return GetTreeKey(addrTo32(addr), big.NewInt(0), CodeHashLeafKey)
}

// GetTreeKeyForCodeChunk returns the tree key for code chunk chunkID of an account.
// tree_index = (CODE_OFFSET + chunkID) / StemNodeWidth
// sub_index  = (CODE_OFFSET + chunkID) % StemNodeWidth
func GetTreeKeyForCodeChunk(addr types.Address, chunkID uint64) [32]byte {
	pos := uint64(CodeOffset) + chunkID
	treeIndex := new(big.Int).SetUint64(pos / StemNodeWidth)
	subIndex := uint8(pos % StemNodeWidth)
	return GetTreeKey(addrTo32(addr), treeIndex, subIndex)
}

// GetTreeKeyForStorageSlot returns the tree key for a storage slot.
// Slots 0..63 are inline (co-located with account data at sub_index 64..127).
// Slots 64+ are stored at MAIN_STORAGE_OFFSET + slot.
// Per EIP-7864 §storage.
func GetTreeKeyForStorageSlot(addr types.Address, storageKey *big.Int) [32]byte {
	// Inline threshold: CODE_OFFSET - HEADER_STORAGE_OFFSET = 64.
	const inlineThreshold int64 = CodeOffset - HeaderStorageOffset

	var pos *big.Int
	if storageKey.IsInt64() && storageKey.Int64() >= 0 && storageKey.Int64() < inlineThreshold {
		pos = new(big.Int).SetUint64(HeaderStorageOffset + uint64(storageKey.Int64()))
	} else {
		pos = new(big.Int).Add(treeMainStorageOffset, storageKey)
	}

	stemWidth := new(big.Int).SetUint64(StemNodeWidth)
	treeIndex := new(big.Int).Div(pos, stemWidth)
	subMod := new(big.Int).Mod(pos, stemWidth)
	return GetTreeKey(addrTo32(addr), treeIndex, uint8(subMod.Uint64()))
}

// addrTo32 pads a 20-byte Ethereum address to Address32 by prepending 12 zero bytes.
func addrTo32(addr types.Address) [32]byte {
	var addr32 [32]byte
	copy(addr32[12:], addr[:])
	return addr32
}
