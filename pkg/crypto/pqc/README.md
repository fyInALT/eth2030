# crypto/pqc — Post-quantum cryptography (Dilithium3, Falcon512, SPHINCS+, ML-DSA-65)

[← crypto](../README.md)

## Overview

This package implements post-quantum cryptographic primitives for the Post-Quantum L1 North Star. It covers four signature schemes: Dilithium3 (real lattice operations, CRYSTALS-Dilithium NIST level 3), Falcon512 (NTRU-lattice, NIST level 1), SPHINCS+-SHA256 (stateless hash-based, NIST level 1), and ML-DSA-65 (FIPS 204 compliant, backed by `github.com/cloudflare/circl`). A unified `PQSigner` interface allows algorithm-agnostic code, while a hybrid signer combines classical ECDSA with a PQ algorithm for migration.

Additional sub-systems include: a PQ algorithm registry for on-chain algorithm selection, a PQ transaction signer for signing Ethereum transactions, lattice-based blob commitments (MLWE), hash-based multi-signatures (XMSS/WOTS+ unified hash signer), a threshold PQ scheme, and a custody replacer for upgrading existing custody proofs to PQ-safe equivalents.

## Functionality

**Core types**
- `PQAlgorithm` — enum: `DILITHIUM3`, `FALCON512`, `SPHINCSSHA256`
- `PQKeyPair` — `Algorithm`, `PublicKey []byte`, `SecretKey []byte`
- `PQSignature` — `Algorithm`, `PublicKey []byte`, `Signature []byte`
- `PQSigner` interface — `GenerateKey() (*PQKeyPair, error)`, `Sign(sk, msg []byte) ([]byte, error)`, `Verify(pk, msg, sig []byte) bool`

**Signers**
- `GetSigner(alg PQAlgorithm) PQSigner` — returns `DilithiumSigner` or `FalconSigner`
- `MLDSASigner` — FIPS 204 ML-DSA-65 via cloudflare/circl
- `SPHINCSSigner` — stateless hash-based signatures
- `HybridSigner` — ECDSA + PQ combined signature
- `UnifiedHashSigner` — XMSS/WOTS+ hash-based scheme

**Supporting infrastructure**
- `PQAlgorithmRegistry` — on-chain algorithm capability registry
- `PQTxSigner` — signs Ethereum transactions with PQ keys
- `LatticeCommitment` / `BlobCommitment` — MLWE lattice-based commitments for blobs
- `HybridThresholdScheme` — combines threshold crypto with PQ
- `CustodyReplacer` / `CustodyReplacerV2` — upgrades custody proofs to PQ
- `BatchBlobVerifier` — verifies multiple blob commitments in batch
- `SignaturePipeline` — end-to-end PQ signing pipeline
- `PubKeyRegistry` — maps lean-sig public keys to PQ counterparts

**Size constants**
- Dilithium3: pubkey 1952 B, seckey 4000 B, sig 3293 B
- Falcon512: pubkey 897 B, seckey 1281 B, sig 690 B
- SPHINCS+-SHA256: pubkey 32 B, seckey 64 B, sig 49216 B

## Usage

```go
signer := pqc.GetSigner(pqc.DILITHIUM3)
kp, _ := signer.GenerateKey()
sig, _ := signer.Sign(kp.SecretKey, msg)
ok := signer.Verify(kp.PublicKey, msg, sig)
```
