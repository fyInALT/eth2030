package stack

import (
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
	"github.com/eth2030/eth2030/rlp"
)

// emptyRoot is the root hash of an empty trie: Keccak256(RLP("")).
var emptyRoot = crypto.Keccak256Hash(func() []byte {
	b, _ := rlp.EncodeToBytes([]byte{})
	return b
}())

// NodeWriter stores trie nodes by hash.
type NodeWriter interface {
	Put(hash types.Hash, data []byte) error
}

const terminatorByte = 16

func keybytesToHex(str []byte) []byte {
	l := len(str)*2 + 1
	nibbles := make([]byte, l)
	for i, b := range str {
		nibbles[i*2] = b / 16
		nibbles[i*2+1] = b % 16
	}
	nibbles[l-1] = terminatorByte
	return nibbles
}

func hexToCompact(hex []byte) []byte {
	terminator := byte(0)
	if hasTerm(hex) {
		terminator = 1
		hex = hex[:len(hex)-1]
	}
	buf := make([]byte, len(hex)/2+1)
	buf[0] = terminator << 5
	if len(hex)&1 == 1 {
		buf[0] |= 1 << 4
		buf[0] |= hex[0]
		hex = hex[1:]
	}
	decodeNibbles(hex, buf[1:])
	return buf
}

func hasTerm(s []byte) bool {
	return len(s) > 0 && s[len(s)-1] == terminatorByte
}

func prefixLen(a, b []byte) int {
	var i, length int
	if len(a) < len(b) {
		length = len(a)
	} else {
		length = len(b)
	}
	for ; i < length; i++ {
		if a[i] != b[i] {
			break
		}
	}
	return i
}

func keysEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func decodeNibbles(nibbles []byte, bytes []byte) {
	for bi, ni := 0, 0; ni < len(nibbles); bi, ni = bi+1, ni+2 {
		bytes[bi] = nibbles[ni]<<4 | nibbles[ni+1]
	}
}

func compareBytesLess(a, b []byte) bool {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return true
		}
		if a[i] > b[i] {
			return false
		}
	}
	return len(a) < len(b)
}

// wrapListPayload wraps the given payload bytes in an RLP list header.
func wrapListPayload(payload []byte) []byte {
	n := len(payload)
	if n <= 55 {
		buf := make([]byte, 1+n)
		buf[0] = 0xc0 + byte(n)
		copy(buf[1:], payload)
		return buf
	}
	lenBytes := putUintBigEndian(uint64(n))
	buf := make([]byte, 1+len(lenBytes)+n)
	buf[0] = 0xf7 + byte(len(lenBytes))
	copy(buf[1:], lenBytes)
	copy(buf[1+len(lenBytes):], payload)
	return buf
}

// putUintBigEndian encodes u as big-endian with no leading zeros.
func putUintBigEndian(u uint64) []byte {
	switch {
	case u < (1 << 8):
		return []byte{byte(u)}
	case u < (1 << 16):
		return []byte{byte(u >> 8), byte(u)}
	case u < (1 << 24):
		return []byte{byte(u >> 16), byte(u >> 8), byte(u)}
	case u < (1 << 32):
		return []byte{byte(u >> 24), byte(u >> 16), byte(u >> 8), byte(u)}
	default:
		return []byte{byte(u >> 56), byte(u >> 48), byte(u >> 40), byte(u >> 32),
			byte(u >> 24), byte(u >> 16), byte(u >> 8), byte(u)}
	}
}
