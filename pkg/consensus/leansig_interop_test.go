package consensus

import (
	"bytes"
	"testing"

	"github.com/eth2030/eth2030/crypto/pqc"
)

// TestLeanSigInterop exercises the full serialize/deserialize round-trip
// using the XMSS leanSig 50-byte format.
func TestLeanSigInterop(t *testing.T) {
	// Generate an XMSS key pair.
	pk, _, err := pqc.GenerateXMSSKeyPair(pqc.XMSSHeight10)
	if err != nil {
		t.Fatalf("GenerateXMSSKeyPair: %v", err)
	}

	// Serialize to leanSig 50-byte format.
	data, err := pqc.SerializeLeanSigPubKey(pk)
	if err != nil {
		t.Fatalf("SerializeLeanSigPubKey: %v", err)
	}
	if len(data) != 50 {
		t.Fatalf("expected 50 bytes, got %d", len(data))
	}

	// Deserialize back.
	pk2, err := pqc.DeserializeLeanSigPubKey(data)
	if err != nil {
		t.Fatalf("DeserializeLeanSigPubKey: %v", err)
	}
	if pk2.Root != pk.Root {
		t.Errorf("root mismatch after round-trip:\n  original:     %x\n  deserialized: %x", pk.Root, pk2.Root)
	}

	// Re-serialize and verify idempotence.
	data2, err := pqc.SerializeLeanSigPubKey(pk2)
	if err != nil {
		t.Fatalf("second SerializeLeanSigPubKey: %v", err)
	}
	if !bytes.Equal(data, data2) {
		t.Error("re-serialized data differs from original serialization")
	}

	// Verify known test vector: Root=[0x01..0x20], serialized[0:40].
	var knownRoot [32]byte
	for i := 0; i < 32; i++ {
		knownRoot[i] = byte(i + 1)
	}
	knownPK := &pqc.XMSSPublicKey{Root: knownRoot, Height: pqc.XMSSHeight10}
	knownData, err := pqc.SerializeLeanSigPubKey(knownPK)
	if err != nil {
		t.Fatalf("SerializeLeanSigPubKey (known vector): %v", err)
	}

	// Each 5-byte root element: [0x00, b0, b1, b2, b3].
	wantRootElements := [][]byte{
		{0x00, 0x01, 0x02, 0x03, 0x04},
		{0x00, 0x05, 0x06, 0x07, 0x08},
		{0x00, 0x09, 0x0a, 0x0b, 0x0c},
		{0x00, 0x0d, 0x0e, 0x0f, 0x10},
		{0x00, 0x11, 0x12, 0x13, 0x14},
		{0x00, 0x15, 0x16, 0x17, 0x18},
		{0x00, 0x19, 0x1a, 0x1b, 0x1c},
		{0x00, 0x1d, 0x1e, 0x1f, 0x20},
	}
	for i, want := range wantRootElements {
		offset := i * 5
		got := knownData[offset : offset+5]
		if !bytes.Equal(got, want) {
			t.Errorf("known vector root element %d: got %x, want %x", i, got, want)
		}
	}
}
