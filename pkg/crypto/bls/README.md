# crypto/bls — BLS12-381 signatures and KZG commitments

[← crypto](../README.md)

## Overview

This package implements BLS12-381 elliptic curve operations for the EVM precompiles defined in EIP-2537 and the KZG polynomial commitment scheme used by EIP-4844 point evaluation. It covers the full EIP-2537 precompile surface (G1/G2 add, scalar mul, MSM, pairing, hash-to-curve) together with beacon-chain BLS aggregate signature operations required by the consensus layer.

The G1 MSM path uses gnark-crypto `G1Affine.MultiExp` (Pippenger algorithm) for performance. Aggregate verification follows the Altair beacon chain spec: public keys live in G1 (48 bytes compressed), signatures in G2 (96 bytes compressed), and hash-to-curve uses the `BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_` domain-separation tag.

## Functionality

**EIP-2537 precompile entry points**
- `BLS12G1Add(input []byte) ([]byte, error)` — G1 point addition (precompile 0x0b)
- `BLS12G1Mul(input []byte) ([]byte, error)` — G1 scalar multiplication (0x0c)
- `BLS12G1MSM(input []byte) ([]byte, error)` — G1 multi-scalar multiplication via Pippenger (0x0d)
- `BLS12G2Add(input []byte) ([]byte, error)` — G2 point addition (0x0e)
- `BLS12G2Mul(input []byte) ([]byte, error)` — G2 scalar multiplication (0x0f)
- `BLS12G2MSM(input []byte) ([]byte, error)` — G2 multi-scalar multiplication (0x10)
- `BLS12Pairing(input []byte) ([]byte, error)` — multi-pairing check (0x11)
- `BLS12MapFpToG1(input []byte) ([]byte, error)` — field element → G1 (0x12)
- `BLS12MapFp2ToG2(input []byte) ([]byte, error)` — Fp2 element → G2 (0x13)

**Aggregate signatures (beacon chain)**
- `BLSSign(secret *big.Int, msg []byte) [96]byte`
- `BLSVerify(pubkey [48]byte, msg []byte, sig [96]byte) bool`
- `FastAggregateVerify(pubkeys [][48]byte, msg []byte, sig [96]byte) bool`
- `VerifyAggregate(pubkeys [][48]byte, msgs [][]byte, sig [96]byte) bool`
- `AggregatePublicKeys(pubkeys [][48]byte) [48]byte`
- `AggregateSignatures(sigs [][96]byte) [96]byte`
- `BLSPubkeyFromSecret(secret *big.Int) [48]byte`
- `SerializeG1 / DeserializeG1`, `SerializeG2 / DeserializeG2`

**KZG commitment verification (EIP-4844)**
- `KZGVerifyFromBytes(commitment, proof []byte, z, y *big.Int) error`
- `KZGVerifyProof(commitment *BlsG1Point, z, y *big.Int, proof *BlsG1Point) bool`
- `KZGDecompressG1(data []byte) (*BlsG1Point, error)` / `KZGCompressG1`
- `KZGCommit(polyAtS *big.Int) *BlsG1Point`
- `KZGComputeProof(secret, z, polyAtS, y *big.Int) *BlsG1Point`

## Usage

```go
// EIP-2537: G1 scalar multiplication
result, err := bls.BLS12G1Mul(input) // 160 bytes: G1 point (128) + scalar (32)

// BLS aggregate signature verification (Altair)
ok := bls.FastAggregateVerify(pubkeys, msg, aggSig)

// KZG point evaluation (EIP-4844 precompile)
err := bls.KZGVerifyFromBytes(commitment48, z, y, proof48)
```
