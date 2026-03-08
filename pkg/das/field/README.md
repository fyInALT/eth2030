# das/field — BLS12-381 scalar field arithmetic and FFT for PeerDAS

[← das](../README.md)

## Overview

This package provides BLS12-381 scalar field (`fr`) arithmetic and FFT operations for PeerDAS polynomial computations. The `FieldElement` type wraps `gnark-crypto`'s `fr.Element` in Montgomery form (four 64-bit limbs), which is 10–50× faster than equivalent `big.Int` modular arithmetic. It exposes addition, subtraction, multiplication, inversion, exponentiation, and conversion to/from `*big.Int`.

FFT operations (`NaiveFFT` and `InverseFFT`) over the BLS scalar field support the polynomial encoding and multi-proof generation required by EIP-7594 (PeerDAS). Domain elements (roots of unity) are computed from the BLS12-381 generator using the field's 2-adic structure.

## Functionality

**Types**
- `FieldElement` — BLS12-381 scalar field element backed by `fr.Element`
- `BLSModulus *big.Int` — `fr.Modulus()` (exported for reference)

**Construction**
- `NewFieldElement(v *big.Int) FieldElement`
- `NewFieldElementFromUint64(v uint64) FieldElement`
- `FieldZero() FieldElement` / `FieldOne() FieldElement`

**Arithmetic**
- `(a FieldElement) Add(b FieldElement) FieldElement`
- `(a FieldElement) Sub(b FieldElement) FieldElement`
- `(a FieldElement) Mul(b FieldElement) FieldElement`
- `(a FieldElement) Neg() FieldElement`
- `(a FieldElement) Inv() FieldElement`
- `(a FieldElement) Exp(e *big.Int) FieldElement`
- `(a FieldElement) IsZero() bool` / `Equal(b FieldElement) bool`
- `(a FieldElement) BigInt() *big.Int` / `Bytes() [32]byte`

**FFT**
- `NaiveFFT(coeffs []FieldElement, domain []FieldElement) []FieldElement`
- `InverseFFT(evals []FieldElement, domain []FieldElement) []FieldElement`
- `ComputeDomain(size int) []FieldElement` — roots of unity for `size`

## Usage

```go
coeffs := make([]field.FieldElement, 4096)
// populate polynomial coefficients
domain := field.ComputeDomain(4096)
evals := field.NaiveFFT(coeffs, domain)
```
