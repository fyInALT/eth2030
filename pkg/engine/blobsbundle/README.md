# engine/blobsbundle — EIP-4844 blob bundle construction and sidecar preparation

Builds, validates, and decomposes EIP-4844 blobs bundles for the Engine API
`engine_getPayload` response. Derives versioned hashes and prepares blob
sidecars for gossip propagation.

## Overview

`BlobsBundleBuilder` incrementally accumulates (blob, KZG commitment, KZG proof)
triples, optionally verifying each via a `KZGVerifier` interface before storing.
`Build` produces a `payload.BlobsBundleV1` with parallel `Commitments`, `Proofs`,
and `Blobs` arrays.

`VersionedHash` implements the EIP-4844 versioned hash derivation:
`SHA-256(commitment)` with byte 0 replaced by `0x01`. `DeriveVersionedHashes`
applies this to every commitment in a bundle, and `ValidateVersionedHashes`
compares the derived hashes against expected values from the CL.

`PrepareSidecars` splits a bundle into individual `BlobSidecar` structs for
gossip, each carrying a binary Merkle inclusion proof built by
`api.BuildInclusionProof`.

## Functionality

**Types**
- `BlobsBundleBuilder` — incremental thread-safe builder
- `BlobSidecar` — `Index`, `Blob`, `KZGCommitment`, `KZGProof`, `SignedBlockHeader`, `CommitmentInclusionProof`
- `KZGVerifier` — interface for commitment/proof verification

**Functions**
- `NewBlobsBundleBuilder(verifier KZGVerifier) *BlobsBundleBuilder`
- `(*BlobsBundleBuilder).AddBlob(blob, commitment, proof) error`
- `(*BlobsBundleBuilder).Build() (*payload.BlobsBundleV1, error)`
- `(*BlobsBundleBuilder).Count() int`
- `(*BlobsBundleBuilder).Reset()`
- `ValidateBundle(bundle *payload.BlobsBundleV1) error`
- `VersionedHash(commitment []byte) types.Hash`
- `DeriveVersionedHashes(bundle) []types.Hash`
- `ValidateVersionedHashes(bundle, expected) error`
- `PrepareSidecars(bundle, blockHash) ([]*BlobSidecar, error)`
- `GetSidecar(bundle, index, blockHash) (*BlobSidecar, error)`

## Usage

```go
builder := blobsbundle.NewBlobsBundleBuilder(nil)
builder.AddBlob(blobData, commitment, proof)
bundle, _ := builder.Build()

hashes := blobsbundle.DeriveVersionedHashes(bundle)
sidecars, _ := blobsbundle.PrepareSidecars(bundle, blockHash)
```

[← engine](../README.md)
