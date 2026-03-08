# das/blockerasure — Block-level Reed-Solomon erasure coding

[← das](../README.md)

## Overview

This package provides k-of-n Reed-Solomon erasure coding of execution blocks over GF(2^8). It splits a block into `dataShards` data pieces and `parityShards` parity pieces, allowing the original block to be reconstructed from any `k` of the `n = k + m` pieces. The default configuration is k=4, m=4 (n=8 total). A standard configuration of k=8, m=8 (n=16) is also provided via `StandardDataShards` and `StandardParityShards` constants.

Each piece includes a Keccak-256 hash for integrity verification and a unique `PieceIndex`. The `BlockErasureEncoder` wraps the lower-level `erasure.Encode`/`Decode` functions and adds hashing, validation, and a piece accumulator (`BlockReconstructionAccumulator`) that collects incoming pieces and signals when reconstruction is possible.

## Functionality

**Types**
- `BlockErasureConfig` — `DataShards int`, `ParityShards int`, `MaxBlockSize uint64`
- `DefaultBlockErasureConfig() BlockErasureConfig` — k=4, m=4, 10 MB max
- `BlockPiece` — `Index int`, `Data []byte`, `Hash [32]byte`, `IsData bool`
- `BlockErasureEncoder` — encodes/decodes blocks with integrity hashing
- `BlockReconstructionAccumulator` — accumulates pieces and reconstructs when threshold reached

**Operations**
- `NewBlockErasureEncoder(cfg BlockErasureConfig) (*BlockErasureEncoder, error)`
- `(e *BlockErasureEncoder) Encode(block *types.Block) ([]*BlockPiece, error)`
- `(e *BlockErasureEncoder) Decode(pieces []*BlockPiece) (*types.Block, error)`
- `NewBlockReconstructionAccumulator(dataShards, parityShards int) *BlockReconstructionAccumulator`
- `(a *BlockReconstructionAccumulator) AddPiece(piece *BlockPiece) (bool, error)` — returns true when reconstruction is ready
- `(a *BlockReconstructionAccumulator) Reconstruct(enc *BlockErasureEncoder) (*types.Block, error)`

## Usage

```go
enc, _ := blockerasure.NewBlockErasureEncoder(blockerasure.DefaultBlockErasureConfig())
pieces, _ := enc.Encode(block)  // distribute pieces to peers

acc := blockerasure.NewBlockReconstructionAccumulator(4, 4)
for _, piece := range receivedPieces {
    ready, _ := acc.AddPiece(piece)
    if ready {
        block, _ := acc.Reconstruct(enc)
    }
}
```
