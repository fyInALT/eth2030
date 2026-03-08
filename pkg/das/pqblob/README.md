# das/pqblob — Post-quantum blob commitments and signatures

Provides lattice-based blob commitments, chunk proofs, and PQ signature
integration for PeerDAS data availability. Part of the L+ era roadmap for
quantum-resistant blob security (Gap #31).

## Overview

`pqblob` implements a two-layer security model. The first layer (`pq_blobs.go`)
computes a lattice Merkle commitment over 32-byte blob chunks using modular
arithmetic modulo the NTRU prime `LatticeModulus` (12289). The second layer
(`pq_blob_integrity.go`, `pq_blob_signer.go`) authenticates those commitments
with one of three PQ signature schemes: ML-DSA-65 (FIPS 204), Falcon-512
(NTRU-lattice NTT), or SPHINCS+ (hash-based Merkle OTS).

The `PQBlobValidator` type wires both layers together into the DAS sampling
pipeline, providing a single validation entry point for DAS samplers.

## Functionality

**Types**
- `PQBlobCommitment` — 64-byte lattice Merkle digest with chunk/data metadata
- `PQBlobProof` — per-chunk lattice witness (96 bytes) linking chunk to commitment
- `PQBlobProofV2` — extended proof carrying a PQ signature over the Merkle root
- `PQBlobSignature` — serializable PQ signature over a commitment digest
- `PQBlobSigner` — wraps a `pqc.PQSigner` with key management
- `MLDSABlobIntegritySigner`, `FalconBlobIntegritySigner`, `SPHINCSBlobIntegritySigner` — algorithm-specific signers implementing `PQBlobIntegritySigner`
- `BatchBlobIntegrityVerifier` — parallel batch verifier using worker goroutines
- `PQBlobIntegrityReport` — thread-safe counters for sign/verify statistics
- `PQBlobValidator` — unified DAS validation entry point

**Functions**
- `CommitBlob(data []byte) (*PQBlobCommitment, error)` — lattice Merkle commitment
- `VerifyBlobCommitment(commitment, data)` — recompute and compare commitment
- `GenerateBlobProof(data, index)` — per-chunk lattice witness
- `VerifyBlobProof(proof, commitment)` — verify lattice witness
- `BatchVerifyProofs(proofs, commitments)` — parallel batch verification
- `ValidatePQBlob`, `ValidatePQBlobProof` — structural validation
- `SignBlobCommitment(commitment, signer)` — produce `PQBlobSignature`
- `VerifyBlobSignature(commitment, sig)` — verify `PQBlobSignature`
- `BatchVerifyBlobSignatures(commitments, sigs)` — parallel batch verify
- `CommitAndSignBlob(data, signer)` — combined commit+sign in one call
- `PQBlobValidator.ValidateBlobCommitment`, `.ValidateBlobProof`, `.GenerateCommitmentProof`, `.BatchValidateCommitments`
- `EstimateValidationGas(algorithm, blobSize)` — gas cost estimate

## Usage

```go
// Commit a blob and sign with ML-DSA-65.
signer, _ := pqblob.NewMLDSABlobIntegritySigner()
commitment, sig, err := pqblob.CommitAndSignBlob(blobData, signer)

// Verify the signature.
ok := signer.VerifyIntegrity(sig, commitment)

// Validate via DAS pipeline entry point.
validator := pqblob.NewPQBlobValidator(pqblob.PQAlgDilithium)
err = validator.ValidateBlobCommitment(blobData, commitment.Digest[:])
```

[← das](../README.md)
