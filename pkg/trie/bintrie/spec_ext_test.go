package bintrie

import (
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// --- Extended GetTreeKey tests ---

func TestGetTreeKey_DifferentAddresses(t *testing.T) {
	addr1 := types.Address{0x01}
	addr2 := types.Address{0x02}
	k1 := GetTreeKeyForBasicData(addr1)
	k2 := GetTreeKeyForBasicData(addr2)
	if k1 == k2 {
		t.Error("different addresses should produce different tree keys")
	}
}

func TestGetTreeKey_Deterministic(t *testing.T) {
	addr := types.Address{0xAB, 0xCD}
	k1 := GetTreeKeyForBasicData(addr)
	k2 := GetTreeKeyForBasicData(addr)
	if k1 != k2 {
		t.Error("GetTreeKeyForBasicData is not deterministic")
	}
}

func TestGetTreeKey_CodeHashVsBasicData(t *testing.T) {
	addr := types.Address{0x42}
	basic := GetTreeKeyForBasicData(addr)
	codeHash := GetTreeKeyForCodeHash(addr)
	// Both use tree_index=0 but different sub_index → same stem, different last byte.
	if basic == codeHash {
		t.Error("BasicData and CodeHash keys should differ")
	}
	// First 31 bytes (stem) should be the same.
	for i := 0; i < 31; i++ {
		if basic[i] != codeHash[i] {
			t.Errorf("stem differs at byte %d: %02x vs %02x", i, basic[i], codeHash[i])
		}
	}
	// Last byte (sub_index) must differ.
	if basic[31] == codeHash[31] {
		t.Error("sub_index byte should differ between BasicData and CodeHash")
	}
}

func TestGetTreeKeyForCodeChunk_Sequential(t *testing.T) {
	addr := types.Address{0x01}
	k0 := GetTreeKeyForCodeChunk(addr, 0)
	k1 := GetTreeKeyForCodeChunk(addr, 1)
	// Adjacent chunks: may share stem or differ, but must be distinct.
	if k0 == k1 {
		t.Error("adjacent code chunks should have different keys")
	}
}

func TestGetTreeKeyForCodeChunk_CrossStemBoundary(t *testing.T) {
	addr := types.Address{0x01}
	// Chunk at StemNodeWidth boundary forces a new stem.
	k127 := GetTreeKeyForCodeChunk(addr, StemNodeWidth-128-1) // chunk 127 in group 0
	k128 := GetTreeKeyForCodeChunk(addr, StemNodeWidth-128)   // chunk 128 in group 1
	// These should have different stems.
	for i := 0; i < 31; i++ {
		if k127[i] != k128[i] {
			return // stems differ → pass
		}
	}
	t.Error("chunks across stem boundary should have different stems")
}

// --- Extended PackBasicDataLeaf tests ---

func TestPackBasicDataLeaf_VersionField(t *testing.T) {
	leaf := PackBasicDataLeaf(0xFF, 0, 0, nil)
	if leaf[0] != 0xFF {
		t.Errorf("version byte: got 0x%02x, want 0xFF", leaf[0])
	}
}

func TestPackBasicDataLeaf_ReservedBytesZero(t *testing.T) {
	leaf := PackBasicDataLeaf(1, 0xFFFFFF, 0xFF, nil)
	// Bytes 1-4 must be zero (reserved).
	for i := 1; i <= 4; i++ {
		if leaf[i] != 0 {
			t.Errorf("reserved byte %d = 0x%02x, want 0", i, leaf[i])
		}
	}
}

func TestPackBasicDataLeaf_MaxCodeSize(t *testing.T) {
	// codeSize fits in 3 bytes → max 0xFFFFFF.
	leaf := PackBasicDataLeaf(0, 0xFFFFFF, 0, nil)
	codeSize := uint32(leaf[5])<<16 | uint32(leaf[6])<<8 | uint32(leaf[7])
	if codeSize != 0xFFFFFF {
		t.Errorf("codeSize: got 0x%06x, want 0xFFFFFF", codeSize)
	}
}

func TestPackBasicDataLeaf_RoundTripZeroBalance(t *testing.T) {
	leaf := PackBasicDataLeaf(1, 100, 5, nil)
	_, _, nonce, balance := UnpackBasicDataLeaf(leaf)
	if nonce != 5 {
		t.Errorf("nonce: got %d, want 5", nonce)
	}
	if balance.Sign() != 0 {
		t.Errorf("nil balance should unpack as 0, got %s", balance)
	}
}

// --- Extended ChunkifyCode tests ---

func TestChunkifyCode_LargePUSH32Sequence(t *testing.T) {
	// Multiple consecutive PUSH32 instructions.
	code := make([]byte, 0)
	for i := 0; i < 3; i++ {
		code = append(code, 0x7f) // PUSH32
		pushdata := make([]byte, 32)
		for j := range pushdata {
			pushdata[j] = byte(i*32 + j)
		}
		code = append(code, pushdata...)
	}
	// Total = 3 * 33 = 99 bytes → ceil(99/31) = 4 chunks
	chunks := ChunkifyCode(code)
	if len(chunks) == 0 {
		t.Fatal("expected chunks for PUSH32 sequence")
	}
	for i, c := range chunks {
		if len(c) != 32 {
			t.Errorf("chunk %d len = %d, want 32", i, len(c))
		}
	}
}

func TestChunkifyCode_AllSTOP(t *testing.T) {
	// All STOP opcodes → all leading bytes should be 0.
	code := make([]byte, 62) // exactly 2 chunks of 31 bytes
	chunks := ChunkifyCode(code)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	for i, c := range chunks {
		if c[0] != 0 {
			t.Errorf("chunk %d leading byte = %d, want 0", i, c[0])
		}
	}
}
