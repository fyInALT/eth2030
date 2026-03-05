// account.go implements EIP-7864 BASIC_DATA_LEAF_KEY packing and unpacking.
// Basic account data (version, code_size, nonce, balance) is packed into a
// single 32-byte value at the leaf slot identified by BASIC_DATA_LEAF_KEY.
//
// Layout per EIP-7864 §header-values:
//
//	byte offset 0:   version   (1 byte)
//	byte offsets 1-4: reserved  (4 bytes, must be zero)
//	byte offsets 5-7: code_size (3 bytes, big-endian)
//	byte offsets 8-15: nonce    (8 bytes, big-endian)
//	byte offsets 16-31: balance (16 bytes, big-endian)
package bintrie

import (
	"math/big"
)

// Basic data layout offsets.
const (
	basicDataVersionOffset = 0
	basicDataBalanceOffset = 16
)

// PackBasicDataLeaf packs version, codeSize, nonce, and balance into a 32-byte
// BASIC_DATA leaf value per EIP-7864 §header-values.
// Reserved bytes 1-4 are set to zero.
func PackBasicDataLeaf(version uint8, codeSize uint32, nonce uint64, balance *big.Int) [32]byte {
	var leaf [32]byte

	// Byte 0: version.
	leaf[basicDataVersionOffset] = version

	// Bytes 1-4: reserved (already zero from [32]byte zero value).

	// Bytes 5-7: code_size as 3-byte big-endian.
	leaf[5] = byte(codeSize >> 16)
	leaf[6] = byte(codeSize >> 8)
	leaf[7] = byte(codeSize)

	// Bytes 8-15: nonce as 8-byte big-endian.
	leaf[8] = byte(nonce >> 56)
	leaf[9] = byte(nonce >> 48)
	leaf[10] = byte(nonce >> 40)
	leaf[11] = byte(nonce >> 32)
	leaf[12] = byte(nonce >> 24)
	leaf[13] = byte(nonce >> 16)
	leaf[14] = byte(nonce >> 8)
	leaf[15] = byte(nonce)

	// Bytes 16-31: balance as 16-byte big-endian.
	if balance != nil && balance.Sign() > 0 {
		b := balance.Bytes() // big-endian, minimal length
		// Copy right-aligned into bytes 16-31.
		if len(b) > 16 {
			b = b[len(b)-16:] // truncate to 16 bytes (overflow silently)
		}
		copy(leaf[basicDataBalanceOffset+16-len(b):], b)
	}

	return leaf
}

// UnpackBasicDataLeaf unpacks a BASIC_DATA leaf value into its constituent fields.
func UnpackBasicDataLeaf(leaf [32]byte) (version uint8, codeSize uint32, nonce uint64, balance *big.Int) {
	version = leaf[basicDataVersionOffset]

	// Bytes 5-7: code_size (3-byte big-endian).
	codeSize = uint32(leaf[5])<<16 | uint32(leaf[6])<<8 | uint32(leaf[7])

	// Bytes 8-15: nonce (8-byte big-endian).
	nonce = uint64(leaf[8])<<56 | uint64(leaf[9])<<48 |
		uint64(leaf[10])<<40 | uint64(leaf[11])<<32 |
		uint64(leaf[12])<<24 | uint64(leaf[13])<<16 |
		uint64(leaf[14])<<8 | uint64(leaf[15])

	// Bytes 16-31: balance (16-byte big-endian).
	balance = new(big.Int).SetBytes(leaf[basicDataBalanceOffset : basicDataBalanceOffset+16])

	return
}
