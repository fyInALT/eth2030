package bintrie

import (
	"testing"
)

func TestChunkifyCode_EmptyCode(t *testing.T) {
	chunks := ChunkifyCode(nil)
	if len(chunks) != 0 {
		t.Errorf("empty code should produce 0 chunks, got %d", len(chunks))
	}
}

func TestChunkifyCode_SingleChunk(t *testing.T) {
	// Simple 3-byte code: PUSH1 0x60 ADD — fits in one 31-byte chunk.
	code := []byte{0x60, 0x60, 0x01}
	chunks := ChunkifyCode(code)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	// Each chunk is [32]byte: [leading_pushdata | 31-byte slice].
	if len(chunks[0]) != 32 {
		t.Errorf("chunk element should be [32]byte, len=%d", len(chunks[0]))
	}
	// PUSH1 starts as opcode, so leading pushdata bytes = 0.
	if chunks[0][0] != 0 {
		t.Errorf("leading pushdata byte should be 0, got %d", chunks[0][0])
	}
	// Bytes 1-3 should match the code.
	if chunks[0][1] != 0x60 || chunks[0][2] != 0x60 || chunks[0][3] != 0x01 {
		t.Errorf("chunk bytes 1-3 wrong: %v", chunks[0][1:4])
	}
}

func TestChunkifyCode_PUSH32_SpansBoundary(t *testing.T) {
	// PUSH32 followed by 32 bytes of data, total 33 bytes.
	// First chunk (31 bytes): PUSH32 opcode + 30 pushdata bytes.
	// Second chunk starts at byte 31 (1 remaining pushdata byte of PUSH32) → leading byte = 1.
	code := make([]byte, 33)
	code[0] = 0x7f // PUSH32
	for i := 1; i <= 32; i++ {
		code[i] = byte(i)
	}
	chunks := ChunkifyCode(code)
	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}
	// Chunk 0 contains PUSH32 opcode (pos 0) + bytes 1..30 (30 pushdata bytes).
	// Chunk 1 starts at pos 31: positions 31 and 32 are both pushdata, so execData[31]=2.
	if chunks[1][0] != 2 {
		t.Errorf("second chunk leading byte should be 2 (2 pushdata bytes remaining), got %d", chunks[1][0])
	}
}

func TestChunkifyCode_Deterministic(t *testing.T) {
	code := make([]byte, 100)
	for i := range code {
		code[i] = byte(i % 256)
	}
	c1 := ChunkifyCode(code)
	c2 := ChunkifyCode(code)
	if len(c1) != len(c2) {
		t.Fatal("non-deterministic chunk count")
	}
	for i := range c1 {
		if c1[i] != c2[i] {
			t.Fatalf("chunk %d differs between runs", i)
		}
	}
}

func TestChunkifyCode_ExactlyOneFull31ByteChunk(t *testing.T) {
	// 31 bytes of STOP opcodes — no PUSH data.
	code := make([]byte, 31)
	chunks := ChunkifyCode(code)
	if len(chunks) != 1 {
		t.Errorf("31 bytes should produce 1 chunk, got %d", len(chunks))
	}
	if chunks[0][0] != 0 {
		t.Errorf("STOP opcodes: leading byte should be 0, got %d", chunks[0][0])
	}
}

func TestChunkifyCode_32Bytes(t *testing.T) {
	// 32 bytes of code → pads to 62 (next multiple of 31) → 2 chunks.
	code := make([]byte, 32)
	chunks := ChunkifyCode(code)
	if len(chunks) != 2 {
		t.Errorf("32 bytes should produce 2 chunks, got %d", len(chunks))
	}
}

func TestChunkifyCode_ChunkSize(t *testing.T) {
	// All chunks must be exactly [32]byte.
	code := make([]byte, 200)
	chunks := ChunkifyCode(code)
	for i, c := range chunks {
		if len(c) != 32 {
			t.Errorf("chunk %d: len=%d, want 32", i, len(c))
		}
	}
}
