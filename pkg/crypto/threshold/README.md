# crypto/threshold — t-of-n threshold encryption for the encrypted mempool

[← crypto](../README.md)

## Overview

This package implements threshold cryptography for the encrypted mempool (Hegotá roadmap item). It uses Shamir's Secret Sharing over a safe-prime group with Feldman Verifiable Secret Sharing (VSS) commitments, combined with ElGamal-style key encapsulation and AES-GCM symmetric encryption. The scheme requires at least `t` of `n` parties to cooperate for decryption, preventing any single node from observing transaction contents before ordering is finalized.

The group parameters use a safe prime `p = 2q+1` (where both `p` and `q` are prime), generator `g = 4` of order `q` in `Z_p*`. Secret sharing polynomials and Lagrange interpolation operate mod `q`; group element operations operate mod `p`. The shared symmetric key is derived from the reconstructed group secret via Keccak-256.

## Functionality

**Types**
- `ThresholdScheme` — holds `T` (threshold) and `N` (total parties)
- `Share` — `Index int`, `Value *big.Int` (one party's Shamir share)
- `VerifiableShare` — `Share` plus Feldman VSS `Commitments []*big.Int`
- `KeyGenResult` — `Shares []Share`, `PublicKey *big.Int`, `Commitments []*big.Int`
- `EncryptedMessage` — `Ephemeral *big.Int`, `Ciphertext []byte`, `Nonce []byte`
- `DecryptionShare` — `Index int`, `Value *big.Int` (per-party decryption contribution)

**Operations**
- `NewThresholdScheme(t, n int) (*ThresholdScheme, error)`
- `(ts *ThresholdScheme) KeyGeneration() (*KeyGenResult, error)` — generates secret, VSS shares, and commitments
- `VerifyShare(share Share, commitments []*big.Int) bool` — Feldman VSS verification
- `MakeVerifiableShare(share Share, commitments []*big.Int) VerifiableShare`
- `ShareEncrypt(publicKey *big.Int, message []byte) (*EncryptedMessage, error)` — ElGamal + AES-GCM
- `ShareDecrypt(share Share, ephemeral *big.Int) DecryptionShare`
- `CombineShares(shares []DecryptionShare, encrypted *EncryptedMessage) ([]byte, error)` — Lagrange interpolation in exponent
- `LagrangeInterpolate(shares []Share) (*big.Int, error)` — plain secret reconstruction

## Usage

```go
ts, _ := threshold.NewThresholdScheme(3, 5)
result, _ := ts.KeyGeneration()

enc, _ := threshold.ShareEncrypt(result.PublicKey, txBytes)

// Each party computes their decryption share:
ds := threshold.ShareDecrypt(result.Shares[i], enc.Ephemeral)

// Combine t=3 shares to decrypt:
plain, _ := threshold.CombineShares(decryptionShares[:3], enc)
```
