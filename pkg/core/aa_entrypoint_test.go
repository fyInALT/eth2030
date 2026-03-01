package core

import (
	"math/big"
	"testing"
)

func TestEncodeDecodeNonce2D(t *testing.T) {
	// Test zero key.
	n := EncodeNonce2D(big.NewInt(0), 42)
	key, seq := DecodeNonce2D(n)
	if key.Sign() != 0 {
		t.Errorf("zero key: got key %s, want 0", key)
	}
	if seq != 42 {
		t.Errorf("zero key: got seq %d, want 42", seq)
	}

	// Test non-zero key (privacy pool).
	privKey := new(big.Int).SetBytes([]byte{0x12, 0x34, 0x56, 0x78})
	n2 := EncodeNonce2D(privKey, 100)
	key2, seq2 := DecodeNonce2D(n2)
	if key2.Cmp(privKey) != 0 {
		t.Errorf("non-zero key: got key %s, want %s", key2, privKey)
	}
	if seq2 != 100 {
		t.Errorf("non-zero key: got seq %d, want 100", seq2)
	}

	// Test nil nonce.
	key3, seq3 := DecodeNonce2D(nil)
	if key3.Sign() != 0 {
		t.Errorf("nil nonce: got key %s, want 0", key3)
	}
	if seq3 != 0 {
		t.Errorf("nil nonce: got seq %d, want 0", seq3)
	}

	// Test max sequence.
	n4 := EncodeNonce2D(big.NewInt(1), ^uint64(0))
	key4, seq4 := DecodeNonce2D(n4)
	if key4.Cmp(big.NewInt(1)) != 0 {
		t.Errorf("max seq: got key %s, want 1", key4)
	}
	if seq4 != ^uint64(0) {
		t.Errorf("max seq: got seq %d, want %d", seq4, ^uint64(0))
	}
}

func TestEncodeNonce2D_NilKey(t *testing.T) {
	// Nil key should behave like zero key.
	n := EncodeNonce2D(nil, 77)
	key, seq := DecodeNonce2D(n)
	if key.Sign() != 0 {
		t.Errorf("nil key: got key %s, want 0", key)
	}
	if seq != 77 {
		t.Errorf("nil key: got seq %d, want 77", seq)
	}
}

func TestEncodeDecodeNonce2D_LargeKey(t *testing.T) {
	// 192-bit key (max for 2D nonce model).
	largeKey := new(big.Int).Lsh(big.NewInt(1), 191) // 2^191
	n := EncodeNonce2D(largeKey, 999)
	key, seq := DecodeNonce2D(n)
	if key.Cmp(largeKey) != 0 {
		t.Errorf("large key: got %s, want %s", key, largeKey)
	}
	if seq != 999 {
		t.Errorf("large key: got seq %d, want 999", seq)
	}
}

func TestDecodeNonce2D_ZeroValue(t *testing.T) {
	key, seq := DecodeNonce2D(big.NewInt(0))
	if key.Sign() != 0 {
		t.Errorf("zero value: got key %s, want 0", key)
	}
	if seq != 0 {
		t.Errorf("zero value: got seq %d, want 0", seq)
	}
}

func TestEncodeDecodeNonce2D_SequenceOnly(t *testing.T) {
	// When key is 0, the encoded nonce should equal the sequence.
	for _, seq := range []uint64{0, 1, 42, 1000000, ^uint64(0)} {
		n := EncodeNonce2D(big.NewInt(0), seq)
		if n.Uint64() != seq {
			t.Errorf("seq-only: EncodeNonce2D(0, %d) = %s, want %d", seq, n, seq)
		}
		_, gotSeq := DecodeNonce2D(n)
		if gotSeq != seq {
			t.Errorf("seq-only: DecodeNonce2D round-trip got seq %d, want %d", gotSeq, seq)
		}
	}
}
