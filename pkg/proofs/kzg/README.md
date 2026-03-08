# proofs/kzg — KZG commitment verifier for blob data (EIP-4844 / PeerDAS)

## Overview

This package implements KZG (Kate-Zaverucha-Goldberg) commitment verification
for EIP-4844 blob transactions and EIP-7594 PeerDAS. It provides single-proof
point evaluation, versioned-hash blob commitment checks, parallel batch
verification, and proof aggregation over multiple blobs.

The verifier is thread-safe and tracks verification statistics. The internal
verification primitive uses a SHA-256 binding commitment; a production deployment
would replace this with real BLS12-381 pairing checks via `go-eth-kzg` or
`c-kzg-4844`.

## Functionality

**Constants** — `KZGCommitmentSize=48`, `KZGProofPointSize=48`,
`BlobFieldElementCount=4096`, `BlobSize=131072`

**Types**

- `KZGCommitment` / `KZGProofPoint` — 48-byte G1 point arrays
- `PointEvaluation` — commitment, proof, evaluation point `z`, claimed value `y`
- `BlobCommitmentPair` — blob versioned hash paired with its commitment and proof
- `KZGBatchResult` — per-item results with `AllValid`, `ValidCount`, `FailedCount`
- `AggregatedKZGProof` — aggregated root and proof over multiple commitments
- `KZGVerifier` — stateful verifier; created with `NewKZGVerifier(config)`

**Methods on `KZGVerifier`**

- `VerifyPointEvaluation(eval) (bool, error)` — verifies p(z) = y for one proof
- `VerifyBlobCommitment(pair) (bool, error)` — checks EIP-4844 versioned hash
- `BatchVerify(evals) (*KZGBatchResult, error)` — parallel or sequential batch
- `AggregateProofs(pairs) (*AggregatedKZGProof, error)` — aggregate commitment root
- `VerifyAggregatedProof(agg) (bool, error)` — verify aggregated root
- `Stats() (verified, failed, batches uint64)`

**Test helpers** — `MakeTestPointEvaluation(index)`, `MakeTestBlobCommitmentPair(index)`

**Parent package:** [proofs](../)
