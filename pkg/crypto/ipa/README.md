# crypto/ipa — Inner Product Argument proof system over Banderwagon

[← crypto](../README.md)

## Overview

This package implements an Inner Product Argument (IPA) proof system over the Banderwagon group, intended for Verkle tree proof generation and verification (EIP-6800). The IPA scheme proves that a committed vector `a` satisfies `<a, b> = v` for a public vector `b`, using a Pedersen commitment `C = <a, G>` over the Banderwagon curve. The proof size is O(log n) curve points, achieved through a recursive vector-halving protocol similar to Bulletproofs.

The `banderwagon.go` file provides the underlying Banderwagon group: a prime-order subgroup of the Twisted Edwards curve BLS12-381 Jubjub variant, supporting point serialization, multi-scalar multiplication (`BanderMSM`), and scalar field arithmetic. Fiat-Shamir challenges are derived from a SHA-256 transcript to make the protocol non-interactive.

## Functionality

**Core IPA types**
- `IPAProofData` — proof consisting of `L`, `R` commitment pairs per round and final scalar `A`
- `BanderPoint` — Banderwagon group element in extended twisted Edwards coordinates

**Proof generation and verification**
- `IPAProve(generators, a, b []*big.Int, commitment *BanderPoint) (*IPAProofData, *big.Int, error)`
- `IPAVerify(generators []*BanderPoint, commitment *BanderPoint, b []*big.Int, v *big.Int, proof *IPAProofData) (bool, error)`
- `IPAProofSize(vectorLen int) int` — returns expected number of rounds for vector length

**Serialization**
- `IPASerialize(proof *IPAProofData) []byte` — compact byte encoding
- `IPADeserialize(data []byte) (*IPAProofData, error)`

**Banderwagon primitives**
- `BanderAdd`, `BanderScalarMul`, `BanderMSM`
- `BanderSerialize(p *BanderPoint) [32]byte` / `BanderDeserialize`
- `BanderEqual(a, b *BanderPoint) bool`

## Usage

```go
n := 8
generators := make([]*ipa.BanderPoint, n) // load Pedersen generators
commitment := ipa.BanderMSM(generators, a)

proof, v, err := ipa.IPAProve(generators, a, b, commitment)
ok, err := ipa.IPAVerify(generators, commitment, b, v, proof)
```
