# crypto/vdf — Verifiable Delay Function (Wesolowski scheme)

[← crypto](../README.md)

## Overview

This package implements a Verifiable Delay Function (VDF) using Wesolowski's repeated-squaring protocol modulo an RSA modulus. VDFs produce outputs that require a prescribed number of sequential squarings to compute but can be verified in logarithmic time. They are targeted at the L+ / M+ roadmap items for unbiasable beacon chain randomness (replacing or supplementing RANDAO).

The RSA modulus is generated as `N = p * q` where `p` and `q` are random primes. For production use, `N` should be generated via an MPC ceremony so that nobody knows its factorization. The Fiat-Shamir challenge prime `l` is derived deterministically from `H(x || y)` using Keccak-256 followed by a primality search. A `beacon.go` file wraps the VDF into slot-indexed beacon randomness derivation. An `enhanced.go` file adds batch evaluation and caching.

## Functionality

**Types**
- `VDFParams` — `T uint64` (squarings), `Lambda uint64` (security bits)
- `VDFProof` — `Input []byte`, `Output []byte`, `Proof []byte`, `Iterations uint64`
- `VDFEvaluator` interface — `Evaluate(input []byte, iterations uint64) (*VDFProof, error)`, `Verify(proof *VDFProof) bool`
- `WesolowskiVDF` — concrete implementation

**Construction**
- `DefaultVDFParams() *VDFParams` — T=2^20, Lambda=128
- `NewWesolowskiVDF(params *VDFParams) *WesolowskiVDF` — generates random RSA modulus
- `NewWesolowskiVDFWithModulus(params *VDFParams, n *big.Int) *WesolowskiVDF` — explicit modulus (testing)

**Evaluation and verification**
- `(v *WesolowskiVDF) Evaluate(input []byte, iterations uint64) (*VDFProof, error)`
- `(v *WesolowskiVDF) Verify(proof *VDFProof) bool` — checks `pi^l * x^r == y mod N`
- `(v *WesolowskiVDF) Modulus() *big.Int`

**Validation helpers**
- `ValidateVDFParams(params *VDFParams) error`
- `ValidateVDFProof(proof *VDFProof) error`

## Usage

```go
vdf := vdf.NewWesolowskiVDF(vdf.DefaultVDFParams())
proof, err := vdf.Evaluate(slotSeed, 1<<20)
ok := vdf.Verify(proof) // fast: O(log T) modular exponentiations
```
