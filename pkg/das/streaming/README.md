# das/streaming — Progressive blob streaming protocol

Implements the blob streaming protocol for PeerDAS, transmitting blobs as
ordered chunks with per-chunk KZG sub-proof verification. Corresponds to the
Data Layer "blob streaming" item in the Hegotá/L+ roadmap.

## Overview

`BlobStream` tracks the progressive reception of a single blob broken into
fixed-size chunks (default 2048 bytes, matching one DAS cell). Each chunk
carries an optional proof that is verified on receipt via
`hash(commitment || index || data)`. When all chunks arrive `IsComplete`
returns `true` and `Assemble` concatenates them into the original blob data.

`BlobStreamer` manages up to `MaxConcurrentStreams` active streams, keyed by
blob commitment hash. `stream_pipeline.go` provides a higher-level pipeline
that processes multiple blobs through the streaming layer. `streaming_proto.go`
defines wire-format encoding for stream messages.

## Functionality

**Types**
- `StreamConfig` — `ChunkSize`, `MaxConcurrentStreams`, `Timeout`
- `BlobChunk` — `Index`, `Data`, `Proof`
- `BlobStream` — per-blob chunk tracker with progressive verification
- `ChunkCallback` — invoked on each verified chunk arrival
- `BlobStreamer` — multi-stream manager

**Functions**
- `DefaultStreamConfig() StreamConfig`
- `NewBlobStreamer(config) *BlobStreamer`
- `(*BlobStreamer).StartStream(blobHash, totalSize, cb) (*BlobStream, error)`
- `(*BlobStreamer).GetStream(blobHash) (*BlobStream, error)`
- `(*BlobStreamer).CloseStream(blobHash)`
- `(*BlobStreamer).ActiveStreams() int`
- `(*BlobStream).AddChunk(chunk) error`
- `(*BlobStream).IsComplete() bool`
- `(*BlobStream).Progress() float64`
- `(*BlobStream).Assemble() ([]byte, error)`
- `ValidateStreamConfig(cfg) error`
- `ValidateBlobChunk(chunk, chunkSize) error`

## Usage

```go
streamer := streaming.NewBlobStreamer(streaming.DefaultStreamConfig())
stream, _ := streamer.StartStream(blobHash, uint32(blobSize), nil)

for _, chunk := range receivedChunks {
    stream.AddChunk(chunk)
}
if stream.IsComplete() {
    data, _ := stream.Assemble()
}
```

[← das](../README.md)
