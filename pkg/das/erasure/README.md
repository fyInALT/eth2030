# das/erasure — Reed-Solomon erasure coding for blob reconstruction

[← das](../README.md)

## Overview

This package provides Reed-Solomon erasure coding used to encode and reconstruct blob data in PeerDAS. The top-level `Encode`/`Decode` functions use an XOR-based parity scheme over uniform shards. For full Galois-field correctness, `galois_field.go` and `gf_field.go` implement GF(2^8) arithmetic, and `polynomial_ops.go` provides polynomial evaluation and interpolation. `reed_solomon_encoder.go` wraps these into a proper RS codec using GF(2^8) Vandermonde matrices.

The package allows any `dataShards` non-nil shards (out of `dataShards + parityShards` total) to reconstruct the original data, matching the PeerDAS requirement that the extended blob be recoverable from any 50% of columns.

## Functionality

**Simple XOR RS interface**
- `Encode(data []byte, dataShards, parityShards int) ([][]byte, error)` — splits into `k+m` uniform shards
- `Decode(shards [][]byte, dataShards, parityShards int) ([]byte, error)` — reconstructs from any `k` non-nil shards

**GF(2^8) arithmetic**
- `GF256` type with `Add`, `Mul`, `Div`, `Pow`, `Inv` methods
- `NewGF256(v byte) GF256`

**Polynomial operations**
- `PolyEval(poly []GF256, x GF256) GF256`
- `PolyInterpolate(xs, ys []GF256) []GF256` — Lagrange interpolation

**Full RS encoder**
- `NewReedSolomonEncoder(dataShards, parityShards int) (*ReedSolomonEncoder, error)`
- `(e *ReedSolomonEncoder) Encode(data []byte) ([][]byte, error)`
- `(e *ReedSolomonEncoder) Decode(shards [][]byte) ([]byte, error)`

**Errors**
- `ErrInvalidShardConfig`, `ErrTooFewShards`, `ErrShardSizeMismatch`, `ErrInvalidShardCount`

## Usage

```go
// Simple interface
shards, _ := erasure.Encode(blobData, 64, 64) // extend to 128 shards
// shards[32] = nil // simulate missing shard
original, _ := erasure.Decode(shards, 64, 64)
```
