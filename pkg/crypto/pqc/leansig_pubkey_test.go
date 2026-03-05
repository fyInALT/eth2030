package pqc

import (
	"bytes"
	"testing"
)

// TestLeanSigPubKeyFormat verifies the 50-byte wire format structure.
func TestLeanSigPubKeyFormat(t *testing.T) {
	// Use a known root: bytes 0x01..0x20 (1..32).
	var root [32]byte
	for i := 0; i < 32; i++ {
		root[i] = byte(i + 1)
	}
	pk := &XMSSPublicKey{Root: root, Height: XMSSHeight10}

	data, err := SerializeLeanSigPubKey(pk)
	if err != nil {
		t.Fatalf("SerializeLeanSigPubKey: %v", err)
	}
	if len(data) != 50 {
		t.Fatalf("expected 50 bytes, got %d", len(data))
	}

	// Check that the root elements (first 40 bytes) follow the format.
	// Root is split into 8 chunks of 4 bytes each.
	// Each chunk is encoded as [0x00, b0, b1, b2, b3].
	for i := 0; i < 8; i++ {
		offset := i * 5
		if data[offset] != 0x00 {
			t.Errorf("root element %d: expected leading 0x00, got 0x%02x", i, data[offset])
		}
		// The 4 source bytes from the root.
		srcOffset := i * 4
		for j := 0; j < 4; j++ {
			if data[offset+1+j] != root[srcOffset+j] {
				t.Errorf("root element %d byte %d: expected 0x%02x, got 0x%02x",
					i, j, root[srcOffset+j], data[offset+1+j])
			}
		}
	}

	// Randomiser elements occupy bytes [40:50] – 5 elements of 2 bytes each.
	// Just verify they are not all zero (derived from SHA256).
	allZero := true
	for _, b := range data[40:50] {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("randomiser bytes should not all be zero")
	}
}

// TestLeanSigPubKeyRoundTrip verifies that serialize→deserialize gives back an equivalent key.
func TestLeanSigPubKeyRoundTrip(t *testing.T) {
	pk, _, err := GenerateXMSSKeyPair(XMSSHeight10)
	if err != nil {
		t.Fatalf("GenerateXMSSKeyPair: %v", err)
	}

	data, err := SerializeLeanSigPubKey(pk)
	if err != nil {
		t.Fatalf("SerializeLeanSigPubKey: %v", err)
	}
	if len(data) != 50 {
		t.Fatalf("expected 50 bytes, got %d", len(data))
	}

	pk2, err := DeserializeLeanSigPubKey(data)
	if err != nil {
		t.Fatalf("DeserializeLeanSigPubKey: %v", err)
	}
	if pk2 == nil {
		t.Fatal("expected non-nil deserialized key")
	}

	// Root must round-trip exactly.
	if pk.Root != pk2.Root {
		t.Errorf("root mismatch:\n  original:     %x\n  deserialized: %x", pk.Root, pk2.Root)
	}

	// Height defaults to XMSSHeight10.
	if pk2.Height != XMSSHeight10 {
		t.Errorf("height: got %d, want %d", pk2.Height, XMSSHeight10)
	}
}

// TestLeanSigPubKeyKnownVector verifies the exact encoding for a known input.
func TestLeanSigPubKeyKnownVector(t *testing.T) {
	// Root = [0x01, 0x02, ..., 0x20].
	var root [32]byte
	for i := 0; i < 32; i++ {
		root[i] = byte(i + 1)
	}
	pk := &XMSSPublicKey{Root: root, Height: XMSSHeight10}

	data, err := SerializeLeanSigPubKey(pk)
	if err != nil {
		t.Fatalf("SerializeLeanSigPubKey: %v", err)
	}

	// Verify each 5-byte root element explicitly.
	want := [][]byte{
		{0x00, 0x01, 0x02, 0x03, 0x04},
		{0x00, 0x05, 0x06, 0x07, 0x08},
		{0x00, 0x09, 0x0a, 0x0b, 0x0c},
		{0x00, 0x0d, 0x0e, 0x0f, 0x10},
		{0x00, 0x11, 0x12, 0x13, 0x14},
		{0x00, 0x15, 0x16, 0x17, 0x18},
		{0x00, 0x19, 0x1a, 0x1b, 0x1c},
		{0x00, 0x1d, 0x1e, 0x1f, 0x20},
	}
	for i, wantElem := range want {
		offset := i * 5
		got := data[offset : offset+5]
		if !bytes.Equal(got, wantElem) {
			t.Errorf("root element %d: got %x, want %x", i, got, wantElem)
		}
	}
}

// TestLeanSigPubKeyInvalidInput verifies that invalid inputs are rejected.
func TestLeanSigPubKeyInvalidInput(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{"nil", nil},
		{"too short", make([]byte, 49)},
		{"too long", make([]byte, 51)},
		{"empty", []byte{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := DeserializeLeanSigPubKey(tc.data)
			if err == nil {
				t.Error("expected error for invalid input, got nil")
			}
		})
	}
}

// TestSerializeLeanSigPubKeyNilInput verifies that nil pk returns an error.
func TestSerializeLeanSigPubKeyNilInput(t *testing.T) {
	_, err := SerializeLeanSigPubKey(nil)
	if err == nil {
		t.Error("expected error for nil pubkey, got nil")
	}
}
