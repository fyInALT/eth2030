package das

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/eth2030/eth2030/crypto"
)

func TestDefaultBlockErasureConfig(t *testing.T) {
	cfg := DefaultBlockErasureConfig()
	if cfg.DataShards != 4 {
		t.Errorf("DataShards = %d, want 4", cfg.DataShards)
	}
	if cfg.ParityShards != 4 {
		t.Errorf("ParityShards = %d, want 4", cfg.ParityShards)
	}
	if cfg.MaxBlockSize != 10*1024*1024 {
		t.Errorf("MaxBlockSize = %d, want %d", cfg.MaxBlockSize, 10*1024*1024)
	}
}

func TestBlockErasureEncode_Basic(t *testing.T) {
	cfg := DefaultBlockErasureConfig()
	enc, err := NewBlockErasureEncoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureEncoder: %v", err)
	}

	data := []byte("hello erasure coded block")
	pieces, err := enc.Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	expected := cfg.DataShards + cfg.ParityShards
	if len(pieces) != expected {
		t.Fatalf("got %d pieces, want %d", len(pieces), expected)
	}

	for i, p := range pieces {
		if p.Index != i {
			t.Errorf("piece %d: Index = %d", i, p.Index)
		}
		if p.TotalPieces != expected {
			t.Errorf("piece %d: TotalPieces = %d, want %d", i, p.TotalPieces, expected)
		}
		if p.BlockSize != uint64(len(data)) {
			t.Errorf("piece %d: BlockSize = %d, want %d", i, p.BlockSize, len(data))
		}
		if len(p.Data) == 0 {
			t.Errorf("piece %d: empty data", i)
		}
	}
}

func TestBlockErasureEncode_EmptyBlock(t *testing.T) {
	enc, err := NewBlockErasureEncoder(DefaultBlockErasureConfig())
	if err != nil {
		t.Fatalf("NewBlockErasureEncoder: %v", err)
	}
	_, err = enc.Encode(nil)
	if err != ErrBlockErasureEmpty {
		t.Errorf("got %v, want ErrBlockErasureEmpty", err)
	}
	_, err = enc.Encode([]byte{})
	if err != ErrBlockErasureEmpty {
		t.Errorf("got %v, want ErrBlockErasureEmpty", err)
	}
}

func TestBlockErasureEncode_TooLarge(t *testing.T) {
	cfg := BlockErasureConfig{
		DataShards:   4,
		ParityShards: 4,
		MaxBlockSize: 100,
	}
	enc, err := NewBlockErasureEncoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureEncoder: %v", err)
	}
	data := make([]byte, 101)
	_, err = enc.Encode(data)
	if err == nil {
		t.Fatal("expected error for oversized block")
	}
}

func TestBlockErasureEncode_PieceHashes(t *testing.T) {
	enc, err := NewBlockErasureEncoder(DefaultBlockErasureConfig())
	if err != nil {
		t.Fatalf("NewBlockErasureEncoder: %v", err)
	}

	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
	}

	pieces, err := enc.Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	for i, p := range pieces {
		computed := crypto.Keccak256Hash(p.Data)
		if computed != p.PieceHash {
			t.Errorf("piece %d: hash mismatch: got %x, want %x", i, computed, p.PieceHash)
		}
	}
}

func TestBlockErasureEncode_BlockHash(t *testing.T) {
	enc, err := NewBlockErasureEncoder(DefaultBlockErasureConfig())
	if err != nil {
		t.Fatalf("NewBlockErasureEncoder: %v", err)
	}

	data := []byte("block hash consistency test data")
	blockHash := crypto.Keccak256Hash(data)

	pieces, err := enc.Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	for i, p := range pieces {
		if p.BlockHash != blockHash {
			t.Errorf("piece %d: BlockHash mismatch", i)
		}
	}
}

func TestBlockErasureDecode_AllPieces(t *testing.T) {
	cfg := DefaultBlockErasureConfig()
	enc, err := NewBlockErasureEncoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureEncoder: %v", err)
	}
	dec, err := NewBlockErasureDecoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureDecoder: %v", err)
	}

	data := []byte("decode with all 8 pieces test data block")
	pieces, err := enc.Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	decoded, err := dec.Decode(pieces)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if !bytes.Equal(decoded, data) {
		t.Errorf("decoded mismatch:\n got: %q\nwant: %q", decoded, data)
	}
}

func TestBlockErasureDecode_MinPieces(t *testing.T) {
	cfg := DefaultBlockErasureConfig()
	enc, err := NewBlockErasureEncoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureEncoder: %v", err)
	}
	dec, err := NewBlockErasureDecoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureDecoder: %v", err)
	}

	data := []byte("decode with exactly k=4 pieces")
	pieces, err := enc.Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	// Use only the first k=4 pieces.
	subset := pieces[:cfg.DataShards]
	decoded, err := dec.Decode(subset)
	if err != nil {
		t.Fatalf("Decode with k pieces: %v", err)
	}
	if !bytes.Equal(decoded, data) {
		t.Errorf("decoded mismatch:\n got: %q\nwant: %q", decoded, data)
	}
}

func TestBlockErasureDecode_KOfN_AllCombinations(t *testing.T) {
	cfg := DefaultBlockErasureConfig() // k=4, m=4, total=8
	enc, err := NewBlockErasureEncoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureEncoder: %v", err)
	}
	dec, err := NewBlockErasureDecoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureDecoder: %v", err)
	}

	data := []byte("test all C(8,4)=70 combinations for k-of-n recovery")
	pieces, err := enc.Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	n := len(pieces)    // 8
	k := cfg.DataShards // 4
	combos := 0

	// Enumerate all C(8,4) = 70 combinations using a bitmask.
	for mask := 0; mask < (1 << n); mask++ {
		bits := 0
		for b := 0; b < n; b++ {
			if mask&(1<<b) != 0 {
				bits++
			}
		}
		if bits != k {
			continue
		}
		combos++

		subset := make([]*BlockPiece, 0, k)
		for b := 0; b < n; b++ {
			if mask&(1<<b) != 0 {
				subset = append(subset, pieces[b])
			}
		}

		decoded, err := dec.Decode(subset)
		if err != nil {
			t.Fatalf("combo mask=%08b: Decode: %v", mask, err)
		}
		if !bytes.Equal(decoded, data) {
			t.Fatalf("combo mask=%08b: decoded mismatch", mask)
		}
	}

	if combos != 70 {
		t.Errorf("tested %d combos, expected 70", combos)
	}
}

func TestBlockErasureDecode_InsufficientPieces(t *testing.T) {
	cfg := DefaultBlockErasureConfig()
	enc, err := NewBlockErasureEncoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureEncoder: %v", err)
	}
	dec, err := NewBlockErasureDecoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureDecoder: %v", err)
	}

	data := []byte("insufficient pieces test")
	pieces, err := enc.Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	// k-1 = 3 pieces should fail.
	subset := pieces[:cfg.DataShards-1]
	_, err = dec.Decode(subset)
	if err == nil {
		t.Fatal("expected error with k-1 pieces")
	}
}

func TestBlockErasureDecode_DuplicatePieces(t *testing.T) {
	cfg := DefaultBlockErasureConfig()
	enc, err := NewBlockErasureEncoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureEncoder: %v", err)
	}
	dec, err := NewBlockErasureDecoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureDecoder: %v", err)
	}

	data := []byte("duplicate piece detection test")
	pieces, err := enc.Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	// Supply the same piece twice.
	duped := []*BlockPiece{pieces[0], pieces[1], pieces[2], pieces[0]}
	_, err = dec.Decode(duped)
	if err == nil {
		t.Fatal("expected error with duplicate pieces")
	}
}

func TestBlockErasureDecode_HashMismatch(t *testing.T) {
	cfg := DefaultBlockErasureConfig()
	enc, err := NewBlockErasureEncoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureEncoder: %v", err)
	}
	dec, err := NewBlockErasureDecoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureDecoder: %v", err)
	}

	data := []byte("tamper detection test data")
	pieces, err := enc.Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	// Tamper with piece data (the hash will no longer match).
	tampered := make([]*BlockPiece, len(pieces))
	for i, p := range pieces {
		cp := *p
		cp.Data = make([]byte, len(p.Data))
		copy(cp.Data, p.Data)
		tampered[i] = &cp
	}
	tampered[0].Data[0] ^= 0xFF

	_, err = dec.Decode(tampered)
	if err == nil {
		t.Fatal("expected error with tampered piece")
	}
}

func TestBlockErasureDecode_LargeBlock(t *testing.T) {
	cfg := DefaultBlockErasureConfig()
	enc, err := NewBlockErasureEncoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureEncoder: %v", err)
	}
	dec, err := NewBlockErasureDecoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureDecoder: %v", err)
	}

	// 1 MB block.
	data := make([]byte, 1024*1024)
	if _, err := rand.Read(data); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}

	pieces, err := enc.Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	decoded, err := dec.Decode(pieces)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if !bytes.Equal(decoded, data) {
		t.Error("1 MB block round-trip mismatch")
	}
}

func TestBlockErasureRoundTrip(t *testing.T) {
	cfg := DefaultBlockErasureConfig()
	enc, err := NewBlockErasureEncoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureEncoder: %v", err)
	}
	dec, err := NewBlockErasureDecoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureDecoder: %v", err)
	}

	data := []byte("full round trip encode -> decode cycle verification data")
	pieces, err := enc.Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	// Use parity-only pieces (indices 4-7) to test reconstruction.
	parityOnly := pieces[cfg.DataShards:]
	decoded, err := dec.Decode(parityOnly)
	if err != nil {
		t.Fatalf("Decode from parity pieces only: %v", err)
	}
	if !bytes.Equal(decoded, data) {
		t.Errorf("round-trip mismatch:\n got: %q\nwant: %q", decoded, data)
	}
}

func TestBlockErasureRoundTrip_RandomData(t *testing.T) {
	cfg := DefaultBlockErasureConfig()
	enc, err := NewBlockErasureEncoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureEncoder: %v", err)
	}
	dec, err := NewBlockErasureDecoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureDecoder: %v", err)
	}

	// Test several random sizes.
	sizes := []int{1, 7, 16, 100, 1000, 4096, 65536}
	for _, sz := range sizes {
		data := make([]byte, sz)
		if _, err := rand.Read(data); err != nil {
			t.Fatalf("rand.Read: %v", err)
		}

		pieces, err := enc.Encode(data)
		if err != nil {
			t.Fatalf("Encode (size %d): %v", sz, err)
		}

		// Use a mix: pieces 0, 2, 5, 7 (two data + two parity).
		subset := []*BlockPiece{pieces[0], pieces[2], pieces[5], pieces[7]}
		decoded, err := dec.Decode(subset)
		if err != nil {
			t.Fatalf("Decode (size %d): %v", sz, err)
		}
		if !bytes.Equal(decoded, data) {
			t.Errorf("size %d: decoded mismatch", sz)
		}
	}
}

func TestBlockErasureCanDecode(t *testing.T) {
	cfg := DefaultBlockErasureConfig()
	enc, err := NewBlockErasureEncoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureEncoder: %v", err)
	}
	dec, err := NewBlockErasureDecoder(cfg)
	if err != nil {
		t.Fatalf("NewBlockErasureDecoder: %v", err)
	}

	data := []byte("can decode test data")
	pieces, err := enc.Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	// k-1 pieces: cannot decode.
	if dec.CanDecode(pieces[:cfg.DataShards-1]) {
		t.Error("expected CanDecode=false with k-1 pieces")
	}

	// k pieces: can decode.
	if !dec.CanDecode(pieces[:cfg.DataShards]) {
		t.Error("expected CanDecode=true with k pieces")
	}

	// All pieces: can decode.
	if !dec.CanDecode(pieces) {
		t.Error("expected CanDecode=true with all pieces")
	}

	// Empty: cannot decode.
	if dec.CanDecode(nil) {
		t.Error("expected CanDecode=false with nil")
	}
}

func TestBlockErasureEncode_Deterministic(t *testing.T) {
	enc, err := NewBlockErasureEncoder(DefaultBlockErasureConfig())
	if err != nil {
		t.Fatalf("NewBlockErasureEncoder: %v", err)
	}

	data := []byte("deterministic encoding test data")

	pieces1, err := enc.Encode(data)
	if err != nil {
		t.Fatalf("Encode (1): %v", err)
	}
	pieces2, err := enc.Encode(data)
	if err != nil {
		t.Fatalf("Encode (2): %v", err)
	}

	if len(pieces1) != len(pieces2) {
		t.Fatalf("piece count mismatch: %d vs %d", len(pieces1), len(pieces2))
	}

	for i := range pieces1 {
		if !bytes.Equal(pieces1[i].Data, pieces2[i].Data) {
			t.Errorf("piece %d: data differs between encodings", i)
		}
		if pieces1[i].PieceHash != pieces2[i].PieceHash {
			t.Errorf("piece %d: hash differs between encodings", i)
		}
		if pieces1[i].BlockHash != pieces2[i].BlockHash {
			t.Errorf("piece %d: block hash differs between encodings", i)
		}
	}
}
