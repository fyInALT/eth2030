# engine/chunking — Execution payload chunking for network propagation

Splits large execution payloads into ordered, integrity-checked chunks for
efficient P2P propagation. Part of the J+ era "payload chunking" roadmap item.

## Overview

`PayloadChunker` divides a payload byte slice into chunks of up to
`DefaultChunkSize` (128 KiB). Each `PayloadChunk` carries its index, total
count, raw data, a Keccak-256 hash of that data, and the hash of the full
payload (`ParentHash`). Receivers verify `VerifyChunkIntegrity` on arrival and
reassemble with `ReassemblePayload`, which checks the final hash.

`ChunkSet` provides a thread-safe accumulator for out-of-order chunk reception,
tracking which chunks have arrived and signalling completion.

## Functionality

**Types**
- `PayloadChunk` — `Index`, `Total`, `Data`, `Hash`, `ParentHash`
- `PayloadChunker` — stateless splitter/reassembler
- `ChunkSet` — concurrent chunk collector

**Functions**
- `NewPayloadChunker(chunkSize int) *PayloadChunker`
- `(*PayloadChunker).ChunkPayload(payload []byte) ([]*PayloadChunk, error)`
- `(*PayloadChunker).ReassemblePayload(chunks []*PayloadChunk) ([]byte, error)`
- `VerifyChunkIntegrity(chunk *PayloadChunk) bool`
- `NewChunkSet(parentHash [32]byte, total uint32) *ChunkSet`
- `(*ChunkSet).AddChunk(chunk) (complete bool, err error)`
- `(*ChunkSet).IsComplete() bool`
- `(*ChunkSet).Received() int`
- `(*ChunkSet).Missing() []uint32`
- `(*ChunkSet).Chunks() []*PayloadChunk`
- `(*ChunkSet).Progress() float64`

## Usage

```go
chunker := chunking.NewPayloadChunker(0) // uses DefaultChunkSize
chunks, _ := chunker.ChunkPayload(payloadBytes)

// On the receiver side:
set := chunking.NewChunkSet(chunks[0].ParentHash, chunks[0].Total)
for _, c := range receivedChunks {
    complete, _ := set.AddChunk(c)
    if complete {
        data, _ := chunker.ReassemblePayload(set.Chunks())
    }
}
```

[← engine](../README.md)
