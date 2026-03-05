// code_chunker.go implements EIP-7864 code chunkification.
// Contract bytecode is split into 31-byte chunks. Each chunk is stored as
// a 32-byte value: byte 0 is the count of leading PUSHDATA bytes in the chunk,
// bytes 1..31 are the raw code bytes.
package bintrie

// PUSH opcode range constants for code chunking.
const (
	pushOffset = 0x5f // PUSH_OFFSET: PUSH1 = pushOffset+1, PUSH32 = pushOffset+32
	push1Byte  = byte(0x60)
	push32Byte = byte(0x7f)
)

// ChunkifyCode splits EVM bytecode into 32-byte chunks per EIP-7864 §code-chunking.
// Each chunk: byte 0 = count of leading PUSHDATA bytes; bytes 1..31 = code slice.
// Returns a slice of [32]byte (each entry is one chunk).
func ChunkifyCode(code []byte) [][32]byte {
	if len(code) == 0 {
		return nil
	}

	// Pad code to a multiple of StemSize (31) bytes.
	padded := code
	if rem := len(code) % StemSize; rem != 0 {
		padded = make([]byte, len(code)+(StemSize-rem))
		copy(padded, code)
	}

	// bytes_to_exec_data[i] = number of remaining pushdata bytes at position i.
	// Sized as len(padded)+32 to guard against PUSH32 at the very end.
	execData := make([]int, len(padded)+32)
	pos := 0
	for pos < len(padded) {
		b := padded[pos]
		var pushdataBytes int
		if b >= push1Byte && b <= push32Byte {
			pushdataBytes = int(b - pushOffset)
		}
		pos++
		for x := 0; x < pushdataBytes; x++ {
			execData[pos+x] = pushdataBytes - x
		}
		pos += pushdataBytes
	}

	// Build chunks.
	numChunks := len(padded) / StemSize
	chunks := make([][32]byte, numChunks)
	for i := 0; i < numChunks; i++ {
		offset := i * StemSize
		var chunk [32]byte
		// Leading byte: min(execData[offset], 31) — capped at StemSize.
		leading := execData[offset]
		if leading > StemSize {
			leading = StemSize
		}
		chunk[0] = byte(leading)
		copy(chunk[1:], padded[offset:offset+StemSize])
		chunks[i] = chunk
	}
	return chunks
}
