# das/proof — Block-in-blob Merkle proof system

Validates block data encoded within blob commitments using Keccak-256 Merkle
proofs. Part of the K+ roadmap item "block in blobs → mandatory proofs →
canonical guest".

## Overview

`proof` splits execution block data into 31-byte field-element chunks (matching
EIP-4844's usable bytes per field element), hashes each chunk as a Merkle leaf,
and constructs a binary Merkle tree over those leaves. The resulting
`BlockBlobProof` carries the block hash, blob indices, Merkle root, proof path,
and encoding metadata.

Verification re-encodes the block data with the same parameters and checks that
the recomputed Merkle root matches the proof, binding the proof to the original
block data without re-downloading blobs.

## Functionality

**Types**
- `BlockBlobProverConfig` — `MaxBlockSize`, `BlobFieldElementSize`, `MaxBlobsPerBlock`
- `BlockBlobEncoding` — split chunks, padding info, chunk count
- `BlockBlobProof` — `BlockHash`, `BlobIndices`, `MerkleRoot`, `ProofPath`, `EncodingMetadata`
- `BlockBlobEncodingMeta` — encoding parameters for proof verification
- `BlockBlobProver` — thread-safe prover with an LRU proof cache

**Functions**
- `DefaultBlockBlobProverConfig() BlockBlobProverConfig`
- `NewBlockBlobProver(config) *BlockBlobProver`
- `(*BlockBlobProver).EncodeBlock(blockData []byte) (*BlockBlobEncoding, error)`
- `(*BlockBlobProver).CreateProof(encoding) (*BlockBlobProof, error)`
- `(*BlockBlobProver).VerifyProof(proof, blockData) (bool, error)`
- `(*BlockBlobProver).EstimateBlobCount(blockSize int) int`

## Usage

```go
prover := proof.NewBlockBlobProver(proof.DefaultBlockBlobProverConfig())

encoding, _ := prover.EncodeBlock(blockData)
prf, _ := prover.CreateProof(encoding)

ok, err := prover.VerifyProof(prf, blockData)
```

[← das](../README.md)
