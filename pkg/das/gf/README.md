# das/gf — GF(2^16) Galois field arithmetic and Reed-Solomon encoder

[← das](../README.md)

## Overview

This package implements GF(2^16) (Galois field with 2^16 = 65536 elements) for Reed-Solomon erasure coding. The field uses the irreducible polynomial `x^16 + x^12 + x^3 + x + 1` (0x1100B) as the reduction polynomial with primitive element `g = 2`. Multiplication and division are O(1) via precomputed log/antilog lookup tables (`gfLogTable` and `gfExpTable`) initialized once via `sync.Once`.

`reed_solomon_encode.go` builds a Reed-Solomon encoder on top of the GF(2^16) arithmetic, targeting the extended-blob encoding for PeerDAS where a 4096-element blob is extended to 8192 elements using polynomial evaluation at roots of unity in GF(2^16).

## Functionality

**GF(2^16) type and operations**
- `GF2_16 uint16` — field element
- `GFAdd(a, b GF2_16) GF2_16` — XOR addition
- `GFMul(a, b GF2_16) GF2_16` — multiplication via log tables (O(1))
- `GFDiv(a, b GF2_16) GF2_16` — division via log tables
- `GFPow(a GF2_16, n int) GF2_16` — exponentiation
- `GFInv(a GF2_16) GF2_16` — multiplicative inverse
- `initGFTables()` — called once to populate log/exp tables

**Reed-Solomon encoder**
- `RSEncoder` — polynomial-based GF(2^16) RS codec
- `NewRSEncoder(dataSymbols, totalSymbols int) (*RSEncoder, error)`
- `(e *RSEncoder) Encode(data []GF2_16) ([]GF2_16, error)` — returns `totalSymbols` codeword
- `(e *RSEncoder) Decode(received []GF2_16, erasures []int) ([]GF2_16, error)` — Berlekamp-Welch style reconstruction

## Usage

```go
// GF(2^16) arithmetic
a, b := gf.GF2_16(12345), gf.GF2_16(6789)
product := gf.GFMul(a, b)

// Reed-Solomon: extend 4096 data symbols to 8192
enc, _ := gf.NewRSEncoder(4096, 8192)
codeword, _ := enc.Encode(dataSymbols)
```
