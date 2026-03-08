# zkvm/poseidon — ZK-friendly Poseidon and Poseidon2 hash functions

## Overview

`zkvm/poseidon` implements the Poseidon (v1) and Poseidon2 hash functions over the BN254 scalar field. Both are ZK-friendly: they are designed to minimise R1CS/PLONK constraint counts compared to SHA-256 or Keccak-256, making them suitable for use inside zero-knowledge circuits. The BN254 scalar field is used throughout for compatibility with Ethereum's `ecAdd`/`ecMul`/`ecPairing` precompiles.

Poseidon1 uses a full Cauchy MDS matrix with full and partial S-box rounds. Poseidon2 improves circuit efficiency by replacing the MDS layer with a diagonal linear layer, reducing constraints per round. Both implementations derive round constants from a Grain LFSR seeded per the Poseidon paper (eprint.iacr.org/2019/458).

## Functionality

**Poseidon1**

- `PoseidonParams` — parameters: state width `T`, round counts, MDS matrix, field modulus.
- `DefaultPoseidonParams() *PoseidonParams` — BN254, T=3, full rounds=8, partial rounds=57.
- `PoseidonHash(params, inputs ...*big.Int) *big.Int` — sponge hash of arbitrary field elements.
- `PoseidonSponge` — incremental absorb/squeeze API via `NewPoseidonSponge`, `Absorb`, `Squeeze`.
- `SBox(x, field *big.Int) *big.Int` — x^5 mod field (the S-box primitive).
- `MDSMul(state, mds, field)` — MDS matrix multiplication.
- `Bn254ScalarField() *big.Int` — copy of the BN254 scalar field modulus.

**Poseidon2**

- `Poseidon2Params` — parameters: `T`, external/internal round counts, diagonal MDS, field.
- `DefaultPoseidon2Params() *Poseidon2Params` — BN254, T=3, external rounds=8, internal rounds=56.
- `Poseidon2Hash(params, inputs ...*big.Int) *big.Int` — sponge hash.
- `Poseidon2HashBytes(data []byte) [32]byte` — convenience wrapper: converts bytes (8 bytes → one field element) and returns a 32-byte digest.
- `Poseidon2Sponge` — incremental API via `NewPoseidon2Sponge`, `Absorb`, `Squeeze`.

## Usage

```go
// One-shot hash with default BN254 parameters.
params := poseidon.DefaultPoseidonParams()
a := new(big.Int).SetUint64(42)
b := new(big.Int).SetUint64(99)
digest := poseidon.PoseidonHash(params, a, b) // *big.Int

// Poseidon2 bytes helper (for arbitrary byte slices).
hash := poseidon.Poseidon2HashBytes([]byte("hello world")) // [32]byte

// Incremental sponge.
sponge := poseidon.NewPoseidon2Sponge(nil) // nil = default params
sponge.Absorb(a, b)
out := sponge.Squeeze(1) // []*big.Int
```

---

Parent package: [`zkvm`](../)
