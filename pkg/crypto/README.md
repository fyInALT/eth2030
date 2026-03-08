# crypto

Cryptographic primitives for the ETH2030 Ethereum client.

## Overview

Package `crypto` provides all cryptographic building blocks used across the
ETH2030 codebase: classical primitives (Keccak-256, secp256k1 ECDSA, BN254),
advanced elliptic curve operations (BLS12-381 with gnark-crypto backend,
Banderwagon/IPA for Verkle proofs), post-quantum signatures (ML-DSA-65/FIPS 204,
Dilithium3, Falcon512, SPHINCS+, XMSS/WOTS+), threshold cryptography (Shamir SSS,
Feldman VSS, ElGamal encryption for the encrypted mempool), and Verifiable Delay
Functions (Wesolowski scheme for beacon chain randomness).

The package follows a split-package architecture: each major cryptographic domain
lives in a dedicated subpackage (`bls/`, `bn254/`, `ipa/`, `pqc/`, `secp256k1/`,
`threshold/`, `vdf/`, `merkle/`). The root `crypto` package re-exports all
subpackage symbols via compat files so that existing imports of
`github.com/eth2030/eth2030/crypto` continue to work unchanged. New code should
import subpackages directly.

## Table of Contents

- [Functionality](#functionality)
- [Subpackages](#subpackages)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### Keccak-256

`Keccak256(data ...[]byte) []byte` and `Keccak256Hash(data ...[]byte) types.Hash`
compute the Ethereum-standard Keccak-256 hash backed by `golang.org/x/crypto/sha3`.
Extended variants in `keccak_extended.go` support domain separation and incremental
hashing.

### ECDSA / secp256k1 (via `secp256k1/`)

The `secp256k1` subpackage provides:

- `GenerateKey()` — generate a secp256k1 private key.
- `Sign(hash, prv)` / `SigToPub(hash, sig)` / `Ecrecover(hash, sig)` — standard
  Ethereum signing and public key recovery.
- `ValidateSignatureValues` — EIP-2 (Homestead) lower-S normalization check.
- `PubkeyToAddress` — derive an Ethereum address from a public key.
- ECIES (`ECIESEncrypt` / `ECIESDecrypt`) for peer-to-peer message encryption.
- P256 (`P256GenerateKey`, `P256Sign`, `P256Verify`, etc.) for TLS/QUIC use.

`SigRecover` (root package) provides compact 65-byte signature parsing, EIP-155
replay-protected recovery, and batch verification for transaction pool use.

`SignatureCacheLRU` (`signature_cache_lru.go`) is an LRU cache for recovered sender
addresses, avoiding redundant ECDSA operations during block import.

`Keystore` (`keystore.go`) manages encrypted private keys using scrypt
(N=262144, r=8, p=1) with version-3 keystore format.

### BLS12-381 (via `bls/`)

The `bls` subpackage implements all EIP-2537 BLS12-381 precompile operations using
gnark-crypto (`github.com/consensys/gnark-crypto/ecc/bls12-381`) as the backend:

- **G1 operations**: `BLS12G1Add`, `BLS12G1Mul`, `BLS12G1MSM` (Pippenger multi-scalar
  multiplication via gnark-crypto `G1Affine.MultiExp`).
- **G2 operations**: `BLS12G2Add`, `BLS12G2Mul`, `BLS12G2MSM`.
- **Pairing**: `BLS12Pairing`.
- **Map-to-curve**: `BLS12MapFpToG1`, `BLS12MapFp2ToG2`, `HashToCurveG1`.
- **High-level**: `BLSSign`, `BLSVerify`, `AggregatePublicKeys`,
  `AggregateSignatures`, `FastAggregateVerify`, `VerifyAggregate`.
- **KZG**: `KZGCommit`, `KZGComputeProof`, `KZGVerifyFromBytes`,
  `KZGCeremonyBackend` interface and `DefaultKZGBackend`.
- **blst adapter** (`bls_blst_adapter.go`): thin adapter over the `blst` CGO library
  for production BLS performance when CGO is available.
- **Batch aggregation** (`bls_aggregate_batch.go`, `bls_aggregate_extended.go`):
  batch BLS verification for transaction pools and attestation processing.

### BN254 (via `bn254/`)

The `bn254` subpackage provides EVM precompile-compatible BN254 operations:

- `BN254Add` — point addition (EIP-196 alt_bn128_add).
- `BN254ScalarMul` — scalar multiplication (EIP-196 alt_bn128_mul).
- `BN254PairingCheck` — pairing check (EIP-197 alt_bn128_pairing).

BN254 is also used for shielded transfers: Pedersen commitments (`v*G + r*H`) and
nullifier set operations live in `shielded_circuit.go` (root package).

### IPA / Banderwagon (via `ipa/`)

The `ipa` subpackage implements Inner Product Arguments over the Banderwagon group
(used for Verkle/IPA proofs):

- `BanderPoint` — Banderwagon group element with Add, Double, Neg, ScalarMul, MSM,
  Serialize/Deserialize.
- `PedersenCommit(values)` — multi-scalar Pedersen commitment.
- `IPAProve(generators, a, b, commitment)` / `IPAVerify(...)` — IPA prover and
  verifier.
- `GoIPABackend` and `PureGoIPABackend` for pluggable backends.

### VDF / Wesolowski (via `vdf/`)

The `vdf` subpackage implements Wesolowski's Verifiable Delay Function for
unbiasable beacon chain randomness (M+ roadmap):

- `VDFEvaluator` interface: `Evaluate(input, iterations)` and `Verify(proof)`.
- `WesolowskiVDF` — concrete Wesolowski implementation. Default params: T=2^20
  squarings, Lambda=128 bits.
- `VDFv2` — enhanced VDF with configurable challenge generation.
- `VDFChain` — chains VDF evaluations for multi-step beacons.
- `VDFBeacon` — wraps `VDFChain` to produce a sequential randomness beacon.
- `BeaconOutput` — beacon output with epoch and value.

### Threshold Cryptography (via `threshold/`)

The `threshold` subpackage implements t-of-n threshold schemes for the encrypted
mempool (Hegotá threshold decryption + ordering):

- `ThresholdScheme` — t-of-n Shamir Secret Sharing with Feldman VSS verification.
- `Share` / `VerifiableShare` — share types with VSS commitments.
- `KeyGenResult` — distributed key generation output.
- `ShareEncrypt` / `ShareDecrypt` / `CombineShares` — ElGamal encryption with
  threshold decryption shares.
- `LagrangeInterpolate` — reconstruct secret from t shares.
- `VerifyShare` — Feldman VSS share verification.

### Post-Quantum Cryptography (via `pqc/`)

The `pqc` subpackage provides a complete post-quantum signature stack:

- **Dilithium3** (`dilithium.go`, `dilithium_sign.go`, `dilithium_enhanced.go`):
  real CRYSTALS-Dilithium lattice operations (1952-byte pubkeys, 3293-byte
  signatures). `DilithiumSigner` implements `PQSigner`.
- **Falcon512** (`falcon.go`, `falcon_signer.go`): FALCON-512 NTRU-lattice
  signatures (897-byte pubkeys, ~690-byte signatures). `FalconSigner` implements
  `PQSigner`.
- **SPHINCS+** (`sphincs_sign.go`, `sphincs_signer.go`): SPHINCS+-SHA256
  stateless hash-based signatures (32-byte pubkeys, 49216-byte signatures).
- **ML-DSA-65** (`mldsa_signer.go`): FIPS 204 lattice signer backed by
  `github.com/cloudflare/circl`.
- **Unified hash signer** (`unified_hash_signer.go`, `l1_hash_sig.go`): XMSS/WOTS+
  hash-based multi-signature scheme for PQ L1 hash-based security (M+ roadmap).
- **Hybrid signer** (`hybrid.go`, `hybrid_threshold.go`): ECDSA + PQ hybrid signing
  for gradual migration.
- **PQ algorithm registry** (`pq_algorithm_registry.go`): runtime-selectable
  algorithm registry mapping `PQAlgorithm` identifiers to signers.
- **PQ transaction signer** (`pq_tx_signer.go`): PQ-signed Ethereum transaction
  wrapper.
- **PQ signing pipeline** (`pq_signing_pipeline.go`): batch signing pipeline for
  high-throughput PQ attestations.
- **Lattice-based blob commitments** (`lattice_commit.go`): MLWE-based blob
  commitment scheme for PQ blob security (L+ roadmap).
- **Pubkey registry** (`pubkey_registry.go`): on-chain PQ pubkey registration
  (post-quantum pubkey registry roadmap item).
- **Signature shares** (`signature_share.go`): threshold PQ signature shares for
  distributed signing.
- **KZG/blob batch verifier** (`batch_blob_verify.go`): batched KZG proof
  verification for blob transactions.

### Merkle Trees (via `merkle/`)

The `merkle` subpackage provides SHA-256-based Merkle tree construction for beacon
chain state roots and block roots, including multi-proof generation and verification.

### Signature Cache

`SignatureCacheLRU` provides an LRU-based cache for ECDSA sender recovery, keyed on
`(txHash, signature)`. The standard `SignatureCache` is a fixed-size concurrent
cache used by the transaction pool.

## Subpackages

| Subpackage | Description |
|------------|-------------|
| [`bls/`](./bls/) | BLS12-381: G1/G2 ops, pairing, MSM, aggregation, KZG (EIP-2537, EIP-4844) |
| [`bn254/`](./bn254/) | BN254: add, scalar-mul, pairing check (EIP-196/197 precompiles) |
| [`ipa/`](./ipa/) | Banderwagon group, IPA proofs, Pedersen commitments (Verkle trees) |
| [`merkle/`](./merkle/) | SHA-256 Merkle tree, proofs, multi-proofs (beacon state roots) |
| [`pqc/`](./pqc/) | Post-quantum: ML-DSA-65, Dilithium3, Falcon512, SPHINCS+, XMSS/WOTS+, hybrid |
| [`secp256k1/`](./secp256k1/) | secp256k1 ECDSA, ECIES, P256, address derivation |
| [`threshold/`](./threshold/) | t-of-n threshold: Shamir SSS, Feldman VSS, ElGamal (encrypted mempool) |
| [`vdf/`](./vdf/) | Wesolowski VDF, VDF chain, randomness beacon (M+ roadmap) |

## Usage

```go
import "github.com/eth2030/eth2030/crypto"

// Keccak-256 hashing.
hash := crypto.Keccak256Hash(data)

// secp256k1 key generation and signing.
prv, _ := crypto.GenerateKey()
sig, _ := crypto.Sign(msgHash, prv)
pubkey, _ := crypto.SigToPub(msgHash, sig)
addr := crypto.PubkeyToAddress(*pubkey)

// EIP-2537 BLS12-381 G1 addition.
result, err := crypto.BLS12G1Add(input)

// BLS aggregate verification.
ok := crypto.FastAggregateVerify(pubkeys, msg, aggregateSig)

// BN254 pairing check (EIP-197).
result, err := crypto.BN254PairingCheck(input)

// Post-quantum signing (Dilithium3).
import "github.com/eth2030/eth2030/crypto/pqc"
signer := pqc.GetSigner(pqc.DILITHIUM3)
kp, _ := signer.GenerateKey()
sig, _ := signer.Sign(kp.SecretKey, msg)
ok := signer.Verify(kp.PublicKey, msg, sig)

// Threshold cryptography.
scheme, _ := crypto.NewThresholdScheme(3, 5) // 3-of-5
// ...distribute shares, encrypt, decrypt...

// Wesolowski VDF.
vdf := crypto.NewWesolowskiVDF(crypto.DefaultVDFParams())
proof, _ := vdf.Evaluate(seed, 1<<20)
ok = vdf.Verify(proof)
```

## Documentation References

- [Design Doc](../../docs/DESIGN.md)
- [PQ Implementation Report](../../docs/PQ_IMPLEMENTATION_REPORT.md)
- [PQ Roadmap Report](../../docs/PQ_ROADMAP_REPORT.md)
- [Roadmap](../../docs/ROADMAP.md)
- [EIP-196: BN254 add/mul](https://eips.ethereum.org/EIPS/eip-196)
- [EIP-197: BN254 pairing](https://eips.ethereum.org/EIPS/eip-197)
- [EIP-2537: BLS12-381 precompiles](https://eips.ethereum.org/EIPS/eip-2537)
- [FIPS 204: ML-DSA (Module-Lattice-Based DSA)](https://csrc.nist.gov/pubs/fips/204/final)
