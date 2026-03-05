package consensus

import (
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// makeTestGossipAttestation creates a PQAttestation suitable for gossip encoding tests.
func makeTestGossipAttestation(slot, valIdx uint64) *PQAttestation {
	var root types.Hash
	root[0] = byte(slot)
	root[1] = byte(valIdx)
	// Use a 32-byte PQPublicKey so the root extraction works.
	pubkey := make([]byte, 32)
	for i := range pubkey {
		pubkey[i] = byte(valIdx + uint64(i))
	}
	sig := make([]byte, 64)
	for i := range sig {
		sig[i] = byte(slot + uint64(i))
	}
	return &PQAttestation{
		Slot:            slot,
		CommitteeIndex:  1,
		BeaconBlockRoot: root,
		SourceEpoch:     slot / 32,
		TargetEpoch:     slot/32 + 1,
		ValidatorIndex:  valIdx,
		PQPublicKey:     pubkey,
		PQSignature:     sig,
	}
}

// TestPQAttestationLeanSigFormat verifies the gossip encoding has the correct byte layout.
func TestPQAttestationLeanSigFormat(t *testing.T) {
	att := makeTestGossipAttestation(100, 7)

	encoded, err := EncodePQAttestationForGossip(att)
	if err != nil {
		t.Fatalf("EncodePQAttestationForGossip: %v", err)
	}

	// Minimum expected size:
	// 8 (slot) + 8 (committeeIndex) + 32 (beaconBlockRoot) +
	// 8 (sourceEpoch) + 8 (targetEpoch) + 8 (validatorIndex) +
	// 50 (leanSig pubkey) + 2 (pubkeyFormat) + 4 (sigLen) + len(sig).
	minSize := 8 + 8 + 32 + 8 + 8 + 8 + 50 + 2 + 4
	if len(encoded) < minSize {
		t.Fatalf("encoded size too small: %d < %d", len(encoded), minSize)
	}
}

// TestPQAttestationGossipRoundTrip verifies that encode→decode gives back the original attestation.
func TestPQAttestationGossipRoundTrip(t *testing.T) {
	original := makeTestGossipAttestation(200, 42)

	encoded, err := EncodePQAttestationForGossip(original)
	if err != nil {
		t.Fatalf("EncodePQAttestationForGossip: %v", err)
	}

	decoded, err := DecodePQAttestationFromGossip(encoded)
	if err != nil {
		t.Fatalf("DecodePQAttestationFromGossip: %v", err)
	}

	if decoded.Slot != original.Slot {
		t.Errorf("Slot: got %d, want %d", decoded.Slot, original.Slot)
	}
	if decoded.CommitteeIndex != original.CommitteeIndex {
		t.Errorf("CommitteeIndex: got %d, want %d", decoded.CommitteeIndex, original.CommitteeIndex)
	}
	if decoded.BeaconBlockRoot != original.BeaconBlockRoot {
		t.Errorf("BeaconBlockRoot: got %x, want %x", decoded.BeaconBlockRoot, original.BeaconBlockRoot)
	}
	if decoded.SourceEpoch != original.SourceEpoch {
		t.Errorf("SourceEpoch: got %d, want %d", decoded.SourceEpoch, original.SourceEpoch)
	}
	if decoded.TargetEpoch != original.TargetEpoch {
		t.Errorf("TargetEpoch: got %d, want %d", decoded.TargetEpoch, original.TargetEpoch)
	}
	if decoded.ValidatorIndex != original.ValidatorIndex {
		t.Errorf("ValidatorIndex: got %d, want %d", decoded.ValidatorIndex, original.ValidatorIndex)
	}
	if len(decoded.PQSignature) != len(original.PQSignature) {
		t.Errorf("PQSignature len: got %d, want %d", len(decoded.PQSignature), len(original.PQSignature))
	}
}

// TestPQAttestationGossipShortPubkey verifies encoding still works with a short PQPublicKey.
func TestPQAttestationGossipShortPubkey(t *testing.T) {
	att := makeTestGossipAttestation(50, 3)
	att.PQPublicKey = []byte{0xAB, 0xCD} // shorter than 32 bytes

	encoded, err := EncodePQAttestationForGossip(att)
	if err != nil {
		t.Fatalf("EncodePQAttestationForGossip (short pubkey): %v", err)
	}

	decoded, err := DecodePQAttestationFromGossip(encoded)
	if err != nil {
		t.Fatalf("DecodePQAttestationFromGossip (short pubkey): %v", err)
	}
	if decoded.Slot != att.Slot {
		t.Errorf("Slot: got %d, want %d", decoded.Slot, att.Slot)
	}
}

// TestPQAttestationGossipDecodeInvalid verifies that truncated data returns an error.
func TestPQAttestationGossipDecodeInvalid(t *testing.T) {
	_, err := DecodePQAttestationFromGossip([]byte{0x01, 0x02})
	if err == nil {
		t.Error("expected error for too-short gossip data, got nil")
	}
}
