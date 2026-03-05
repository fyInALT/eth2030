// pq_attestation_gossip.go implements P2P gossip encoding/decoding for
// PQAttestation using the leanSig 50-byte XMSS pubkey wire format.
package consensus

import (
	"encoding/binary"
	"errors"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto/pqc"
)

// gossip wire format constants.
const (
	gossipFixedHeader = 8 + 8 + 32 + 8 + 8 + 8 // slot+committee+root+srcEpoch+tgtEpoch+valIdx
	gossipPubKeyBytes = 50                     // leanSig pubkey size
	gossipFormatBytes = 2                      // pubkeyFormat indicator
	gossipSigLenBytes = 4                      // big-endian uint32 sig length

	gossipMinSize = gossipFixedHeader + gossipPubKeyBytes + gossipFormatBytes + gossipSigLenBytes

	// pubkeyFormat values.
	pubkeyFormatLeanSig  = uint16(0x0001)
	pubkeyFormatInternal = uint16(0x0000)
)

// gossip encoding errors.
var (
	ErrGossipTooShort  = errors.New("pq attestation gossip: data too short")
	ErrGossipTruncated = errors.New("pq attestation gossip: signature data truncated")
)

// EncodePQAttestationForGossip encodes a PQAttestation for P2P gossip using
// the leanSig 50-byte XMSS pubkey format. Falls back to a zero-padded 50-byte
// block if the pubkey cannot be represented as an XMSS key.
//
// Wire format:
//
//	[0:8]   slot           (uint64 big-endian)
//	[8:16]  committeeIndex (uint64 big-endian)
//	[16:48] beaconBlockRoot
//	[48:56] sourceEpoch    (uint64 big-endian)
//	[56:64] targetEpoch    (uint64 big-endian)
//	[64:72] validatorIndex (uint64 big-endian)
//	[72:122] pubkey        (50 bytes leanSig or zero-padded)
//	[122:124] pubkeyFormat (uint16 big-endian: 0x0001=leanSig, 0x0000=internal)
//	[124:128] sigLen       (uint32 big-endian)
//	[128:]  PQSignature
func EncodePQAttestationForGossip(att *PQAttestation) ([]byte, error) {
	// Build the XMSSPublicKey from the first 32 bytes of PQPublicKey.
	// Zero-pad if shorter than 32 bytes.
	var xmssRoot [32]byte
	copy(xmssRoot[:], att.PQPublicKey) // safe: copy only up to len(att.PQPublicKey)

	xmssPK := &pqc.XMSSPublicKey{Root: xmssRoot, Height: pqc.XMSSHeight10}
	leanPubKey, err := pqc.SerializeLeanSigPubKey(xmssPK)
	format := pubkeyFormatLeanSig
	if err != nil {
		// Fallback: 50 zero bytes.
		leanPubKey = make([]byte, gossipPubKeyBytes)
		format = pubkeyFormatInternal
	}

	sigLen := len(att.PQSignature)
	buf := make([]byte, gossipMinSize+sigLen)
	off := 0

	binary.BigEndian.PutUint64(buf[off:], att.Slot)
	off += 8
	binary.BigEndian.PutUint64(buf[off:], att.CommitteeIndex)
	off += 8
	copy(buf[off:], att.BeaconBlockRoot[:])
	off += 32
	binary.BigEndian.PutUint64(buf[off:], att.SourceEpoch)
	off += 8
	binary.BigEndian.PutUint64(buf[off:], att.TargetEpoch)
	off += 8
	binary.BigEndian.PutUint64(buf[off:], att.ValidatorIndex)
	off += 8
	copy(buf[off:], leanPubKey)
	off += gossipPubKeyBytes
	binary.BigEndian.PutUint16(buf[off:], format)
	off += 2
	binary.BigEndian.PutUint32(buf[off:], uint32(sigLen))
	off += 4
	copy(buf[off:], att.PQSignature)

	return buf, nil
}

// DecodePQAttestationFromGossip decodes a gossip-encoded PQAttestation.
// Tries leanSig format first, falls back to internal format for the pubkey.
func DecodePQAttestationFromGossip(data []byte) (*PQAttestation, error) {
	if len(data) < gossipMinSize {
		return nil, ErrGossipTooShort
	}

	off := 0
	slot := binary.BigEndian.Uint64(data[off:])
	off += 8
	committeeIndex := binary.BigEndian.Uint64(data[off:])
	off += 8

	var beaconBlockRoot types.Hash
	copy(beaconBlockRoot[:], data[off:off+32])
	off += 32

	sourceEpoch := binary.BigEndian.Uint64(data[off:])
	off += 8
	targetEpoch := binary.BigEndian.Uint64(data[off:])
	off += 8
	validatorIndex := binary.BigEndian.Uint64(data[off:])
	off += 8

	rawPubKey := data[off : off+gossipPubKeyBytes]
	off += gossipPubKeyBytes

	format := binary.BigEndian.Uint16(data[off:])
	off += 2

	sigLen := int(binary.BigEndian.Uint32(data[off:]))
	off += 4

	if off+sigLen > len(data) {
		return nil, ErrGossipTruncated
	}
	sig := make([]byte, sigLen)
	copy(sig, data[off:off+sigLen])

	// Reconstruct the pubkey bytes from the 50-byte wire block.
	var pubKeyBytes []byte
	if format == pubkeyFormatLeanSig {
		xmssPK, err := pqc.DeserializeLeanSigPubKey(rawPubKey)
		if err == nil {
			// Store the 32-byte root as pubkey bytes for compatibility with PQAttestation.
			pubKeyBytes = xmssPK.Root[:]
		}
	}
	if pubKeyBytes == nil {
		// Fallback: use raw bytes directly (first 32 bytes of the 50-byte block).
		pubKeyBytes = make([]byte, 32)
		copy(pubKeyBytes, rawPubKey)
	}

	return &PQAttestation{
		Slot:            slot,
		CommitteeIndex:  committeeIndex,
		BeaconBlockRoot: beaconBlockRoot,
		SourceEpoch:     sourceEpoch,
		TargetEpoch:     targetEpoch,
		ValidatorIndex:  validatorIndex,
		PQPublicKey:     pubKeyBytes,
		PQSignature:     sig,
	}, nil
}
