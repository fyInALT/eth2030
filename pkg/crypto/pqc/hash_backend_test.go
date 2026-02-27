package pqc

import (
	"bytes"
	"testing"
)

func TestHashBackendInterface(t *testing.T) {
	backends := []HashBackend{
		&Keccak256Backend{},
		&SHA256Backend{},
		&Blake3Backend{},
	}

	data := []byte("hello post-quantum world")

	for _, b := range backends {
		t.Run(b.Name(), func(t *testing.T) {
			// Basic hash
			h1 := b.Hash(data)
			h2 := b.Hash(data)
			if h1 != h2 {
				t.Error("hash is not deterministic")
			}

			// Different data produces different hash
			h3 := b.Hash([]byte("different data"))
			if h1 == h3 {
				t.Error("different inputs produced same hash")
			}

			// Empty input works
			h4 := b.Hash(nil)
			h5 := b.Hash([]byte{})
			if h4 != h5 {
				t.Error("nil and empty should produce same hash")
			}

			// BlockSize is positive
			if b.BlockSize() <= 0 {
				t.Error("block size must be positive")
			}

			// Name is non-empty
			if b.Name() == "" {
				t.Error("name must be non-empty")
			}
		})
	}
}

func TestDefaultHashBackend(t *testing.T) {
	b := DefaultHashBackend()
	if b.Name() != "keccak256" {
		t.Errorf("default backend should be keccak256, got %s", b.Name())
	}
}

func TestHashBackendByName(t *testing.T) {
	tests := []struct {
		name     string
		wantName string
		wantNil  bool
	}{
		{"keccak256", "keccak256", false},
		{"sha256", "sha256", false},
		{"blake3", "blake3", false},
		{"unknown", "", true},
	}
	for _, tt := range tests {
		b := HashBackendByName(tt.name)
		if tt.wantNil {
			if b != nil {
				t.Errorf("expected nil for name %q", tt.name)
			}
		} else {
			if b == nil {
				t.Fatalf("expected non-nil for name %q", tt.name)
			}
			if b.Name() != tt.wantName {
				t.Errorf("got name %q, want %q", b.Name(), tt.wantName)
			}
		}
	}
}

func TestHashBackendDifferentOutputs(t *testing.T) {
	// Each backend should produce different outputs for the same input
	data := []byte("test input for cross-backend comparison")
	k := (&Keccak256Backend{}).Hash(data)
	s := (&SHA256Backend{}).Hash(data)
	b := (&Blake3Backend{}).Hash(data)

	if bytes.Equal(k[:], s[:]) {
		t.Error("keccak256 and sha256 produced same output")
	}
	if bytes.Equal(k[:], b[:]) {
		t.Error("keccak256 and blake3 produced same output")
	}
	if bytes.Equal(s[:], b[:]) {
		t.Error("sha256 and blake3 produced same output")
	}
}
