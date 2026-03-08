# das/blobs — Block-in-blobs encoding and variable blob support

[← das](../README.md)

## Overview

This package implements block-in-blobs encoding for the EL throughput roadmap track (Hegotá / I+). It encodes full execution payloads into blob space, enabling L1 blocks to be reconstructed entirely from blob data. Each blob carries a 53-byte header (`originalLen[4] + index[8] + totalChunks[8] + isLast[1] + blockHash[32]`) followed by payload bytes. A block is split across as many blobs as needed up to a configurable maximum.

The package also provides `forward_cast.go` for casting between blob representations, and `teradata.go` for teragas-scale data handling.

## Functionality

**Types and config**
- `BlobBlockConfig` — `MaxBlobSize uint64`, `MaxBlobsPerBlock uint64`, `CompressionEnabled bool`
- `DefaultBlobBlockConfig() BlobBlockConfig` — 128 KiB blobs, configurable max blob count
- `DefaultBlobSize = 131072` (128 KiB)

**Encoding / decoding**
- `EncodeBlockToBlobs(block *types.Block, cfg BlobBlockConfig) ([][]byte, error)` — splits block RLP into sequenced blobs
- `DecodeBlockFromBlobs(blobs [][]byte) (*types.Block, error)` — reconstructs block from ordered blobs; validates header consistency and hash integrity

**Errors**
- `ErrBlockDataEmpty`, `ErrBlockDataTooLarge`, `ErrNoBlobsProvided`, `ErrBlobOrderMismatch`, `ErrBlobHashMismatch`, `ErrBlobCountMismatch`, `ErrMissingLastBlob`, `ErrBlobDataCorrupt`, `ErrCommitmentMismatch`, `ErrMaxBlobsExceeded`

## Usage

```go
cfg := blobs.DefaultBlobBlockConfig()
encoded, err := blobs.EncodeBlockToBlobs(block, cfg)
// transmit encoded blobs via blob transaction
recovered, err := blobs.DecodeBlockFromBlobs(received)
```
