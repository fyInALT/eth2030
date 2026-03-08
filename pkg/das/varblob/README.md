# das/varblob — Variable-size blob support (J+ upgrade)

Provides variable-length blob payloads that are split into power-of-two-sized
chunks for data availability sampling. Implements the "variable-size blobs" item
from the J+ (2027-2028) upgrade tier.

## Overview

`VarBlob` wraps a data payload aligned to a configurable `ChunkSize` (must be a
power of two in `[128, 4096]` bytes). Data is zero-padded to fill the last
chunk. The blob hash is computed over the original (unpadded) bytes so that
verifiers can detect non-zero padding via `ValidatePaddingProof`.

`VarBlobTx` attaches a `VarBlob` to a transaction-level envelope carrying a
destination address and value.

Gas costs scale linearly with chunk count: `VarBlobBaseGas + VarBlobPerChunkGas * numChunks`.

## Functionality

**Types**
- `VarBlobConfig` — `MinChunkSize`, `MaxChunkSize`, `MaxBlobSize`
- `VarBlob` — `Data`, `ChunkSize`, `NumChunks`, `BlobHash`
- `VarBlobTx` — `Blob`, `To`, `Value`, `Data`

**Functions**
- `DefaultVarBlobConfig() VarBlobConfig`
- `NewVarBlob(data []byte, chunkSize int) (*VarBlob, error)` — create with zero-pad
- `(*VarBlob).Chunks() [][]byte` — split into fixed-size chunks
- `(*VarBlob).Encode() []byte` — serialize: `chunkSize[4] || numChunks[4] || data`
- `DecodeVarBlob(data []byte) (*VarBlob, error)`
- `(*VarBlob).Verify(expectedHash types.Hash) bool`
- `ValidateVarBlob(vb *VarBlob) error` — structural consistency check
- `ValidatePaddingProof(vb *VarBlob, dataLen int) error` — zero-padding check
- `EstimateVarBlobGas(blobSize, chunkSize int) uint64`

## Usage

```go
blob, err := varblob.NewVarBlob(payload, 512) // 512-byte chunks
chunks := blob.Chunks()                        // len == blob.NumChunks

encoded := blob.Encode()
decoded, _ := varblob.DecodeVarBlob(encoded)

gas := varblob.EstimateVarBlobGas(len(payload), 512)
```

[← das](../README.md)
