# engine/blobval — EIP-4844 blob sidecar validation

Validates blob sidecars for structural correctness, counts blob gas, and
computes blob base fee using the EIP-4844 fake exponential formula.

## Overview

`BlobValidator` performs structural (non-cryptographic) validation: sequential
index ordering, no duplicates, non-zero KZG commitments, and count consistency
with block transactions. Full KZG proof verification requires the trusted setup
from `crypto/kzg` and is delegated to callers.

`FixedBlobSidecar` uses fixed-size arrays (`[131072]byte` blob, `[48]byte`
commitment/proof) matching the consensus-layer representation, complementing the
dynamic-slice `BlobSidecar` in `engine/blobsbundle`.

Gas helpers compute total blob gas, excess blob gas, and blob base fee using the
same formulas as `core/types`.

## Functionality

**Types**
- `FixedBlobSidecar` — `BlobIndex`, `Blob [131072]byte`, `KZGCommitment [48]byte`, `KZGProof [48]byte`
- `BlobValidator` — configurable max/target blobs and gas per blob

**Functions**
- `NewBlobValidator() *BlobValidator` — Cancun defaults
- `NewBlobValidatorWithConfig(maxBlobs, targetBlobs int, gasPerBlob uint64) *BlobValidator`
- `(*BlobValidator).ValidateBlobSidecars(sidecars, blockHash) error`
- `(*BlobValidator).ValidateKZGCommitments(commitments, blobs) error`
- `(*BlobValidator).VerifySidecarCount(sidecars, txs) error`
- `ComputeBlobGas(numBlobs int) uint64`
- `ComputeExcessBlobGas(parentExcess, parentUsed uint64) uint64`
- `ComputeBlobBaseFee(excessBlobGas uint64) *big.Int`
- `ValidateBlobTransactionSidecar(tx, sidecar) error`
- `CountBlobsInTransactions(txs []*types.Transaction) int`

## Usage

```go
v := blobval.NewBlobValidator()
err := v.ValidateBlobSidecars(sidecars, blockHash)
err = v.VerifySidecarCount(sidecars, blockTxs)

baseFee := blobval.ComputeBlobBaseFee(header.ExcessBlobGas)
```

[← engine](../README.md)
