// leansig_pubkey.go implements leanSig 50-byte wire format serialization for
// XMSS public keys, matching the leanSig protocol's compact on-wire encoding.
package pqc

import (
	"crypto/sha256"
	"errors"
)

// leanSig wire format constants.
const (
	leanSigPubKeySize   = 50 // total wire bytes
	leanSigRootElements = 8  // 8 chunks of 4 bytes from the 32-byte root
	leanSigRandElements = 5  // 5 randomiser elements of 2 bytes each
	leanSigElemBytes    = 5  // bytes per root element (1 zero prefix + 4 data)
	leanSigRandBytes    = 2  // bytes per randomiser element
)

// leanSigRandLabel is the domain label used to derive randomiser bytes.
var leanSigRandLabel = []byte("leansig-rand")

// ErrLeanSigInvalidSize is returned when the input buffer has wrong length.
var ErrLeanSigInvalidSize = errors.New("leansig: pubkey must be exactly 50 bytes")

// ErrLeanSigNilPubKey is returned when a nil XMSSPublicKey is passed.
var ErrLeanSigNilPubKey = errors.New("leansig: nil pubkey")

// SerializeLeanSigPubKey serializes an XMSSPublicKey to the leanSig 50-byte wire format.
//
// Format:
//   - 40 bytes: 8 root elements, each 5 bytes [0x00, b0, b1, b2, b3]
//     where [b0,b1,b2,b3] is one 4-byte chunk of the 32-byte Root.
//   - 10 bytes: 5 randomiser elements of 2 bytes each, derived from
//     SHA256(root || "leansig-rand")[0:10].
func SerializeLeanSigPubKey(pk *XMSSPublicKey) ([]byte, error) {
	if pk == nil {
		return nil, ErrLeanSigNilPubKey
	}

	out := make([]byte, leanSigPubKeySize)

	// Encode 8 root elements (5 bytes each = 40 bytes total).
	for i := 0; i < leanSigRootElements; i++ {
		srcOffset := i * 4
		dstOffset := i * leanSigElemBytes
		out[dstOffset] = 0x00 // leading zero
		out[dstOffset+1] = pk.Root[srcOffset]
		out[dstOffset+2] = pk.Root[srcOffset+1]
		out[dstOffset+3] = pk.Root[srcOffset+2]
		out[dstOffset+4] = pk.Root[srcOffset+3]
	}

	// Derive 10 randomiser bytes from SHA256(root || "leansig-rand").
	h := sha256.New()
	h.Write(pk.Root[:])
	h.Write(leanSigRandLabel)
	randHash := h.Sum(nil)
	copy(out[40:50], randHash[:10])

	return out, nil
}

// DeserializeLeanSigPubKey deserializes a 50-byte leanSig pubkey back to an XMSSPublicKey.
//
// It validates the exact 50-byte length and reconstructs the 32-byte root by
// extracting bytes [1:5] from each 5-byte root element. The 10 randomiser bytes
// are validated for presence but not stored in XMSSPublicKey. The returned key
// uses XMSSHeight10 as the default height.
func DeserializeLeanSigPubKey(data []byte) (*XMSSPublicKey, error) {
	if len(data) != leanSigPubKeySize {
		return nil, ErrLeanSigInvalidSize
	}

	var root [32]byte
	for i := 0; i < leanSigRootElements; i++ {
		srcOffset := i * leanSigElemBytes
		dstOffset := i * 4
		// bytes [1:5] of each element are the 4 source bytes.
		root[dstOffset] = data[srcOffset+1]
		root[dstOffset+1] = data[srcOffset+2]
		root[dstOffset+2] = data[srcOffset+3]
		root[dstOffset+3] = data[srcOffset+4]
	}

	return &XMSSPublicKey{
		Root:   root,
		Height: XMSSHeight10,
	}, nil
}
