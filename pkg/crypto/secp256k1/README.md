# crypto/secp256k1 — secp256k1 ECDSA, P-256, and ECIES

[← crypto](../README.md)

## Overview

This package provides secp256k1 ECDSA operations used throughout Ethereum for transaction signing and address derivation, along with P-256 (NIST P-256) support and ECIES (Elliptic Curve Integrated Encryption Scheme) for encrypted communication. The secp256k1 implementation exposes key generation, signing with low-S normalization (EIP-2), public key recovery, address derivation, and compressed/uncompressed key encoding.

The P-256 sub-files (`p256.go`, `p256_extended.go`) add an additional elliptic curve implementation for use cases requiring NIST compliance. The `ecies.go` file provides hybrid public-key encryption using secp256k1 (or P-256) for key agreement and AES-GCM for payload encryption.

## Functionality

**Key operations**
- `GenerateKey() (*ecdsa.PrivateKey, error)`
- `BLSPubkeyFromSecret` / `PubkeyToAddress(p ecdsa.PublicKey) types.Address`
- `CompressPubkey(pubkey *ecdsa.PublicKey) []byte` — 33-byte compressed form
- `DecompressPubkey(pubkey []byte) (*ecdsa.PublicKey, error)`
- `FromECDSAPub(pub *ecdsa.PublicKey) []byte` — 65-byte uncompressed `[0x04 || X || Y]`

**Signing and verification**
- `Sign(hash []byte, prv *ecdsa.PrivateKey) ([]byte, error)` — returns 65-byte `[R || S || V]` with low-S normalization
- `Ecrecover(hash, sig []byte) ([]byte, error)` — recovers uncompressed public key
- `SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error)`
- `ValidateSignature(pubkey, hash, sig []byte) bool` — verifies 64-byte sig against 65-byte pubkey
- `ValidateSignatureValues(v byte, r, s *big.Int, homestead bool) bool`

**Curve accessors**
- `S256() elliptic.Curve` — returns the secp256k1 curve
- `Secp256k1N() *big.Int` / `Secp256k1HalfN() *big.Int`

## Usage

```go
prv, _ := secp256k1.GenerateKey()
sig, _ := secp256k1.Sign(hash32, prv)        // [R || S || V], low-S
pub, _ := secp256k1.Ecrecover(hash32, sig)   // 65-byte uncompressed pubkey
addr := secp256k1.PubkeyToAddress(prv.PublicKey)
```
