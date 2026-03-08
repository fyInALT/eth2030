# crypto/bn254 — BN254 (alt_bn128) curve and private L1 shielded transfers

[← crypto](../README.md)

## Overview

This package implements BN254 (also known as alt_bn128) elliptic curve operations for the EVM precompiles defined in EIP-196 and EIP-197, as well as the private L1 shielded transfer infrastructure targeting the Private L1 North Star. It provides point addition, scalar multiplication, and pairing checks (precompiles 0x06–0x08), along with BN254 Pedersen commitments, a nullifier-based double-spend prevention set, and a commitment Merkle tree for the `ShieldedPool`.

The shielded transfer subsystem uses Pedersen commitments over BN254 to hide transfer amounts and participants. Notes are spent by revealing nullifiers; the `ShieldedPool` tracks both unspent commitments and revealed nullifiers in thread-safe maps, and computes a Merkle root over the nullifier set via `SparseMerkleTree`.

## Functionality

**EIP-196/197 precompile entry points**
- `BN254Add(input []byte) ([]byte, error)` — G1 point addition (precompile 0x06)
- `BN254ScalarMul(input []byte) ([]byte, error)` — G1 scalar multiplication (0x07)
- `BN254PairingCheck(input []byte) ([]byte, error)` — multi-pairing equality check (0x08)

**Shielded transfer pool**
- `ShieldedTx` — private transfer with `NullifierHash`, `Commitment`, `EncryptedData`, `Proof`
- `ShieldedPool` — thread-safe pool tracking commitments and nullifiers
- `NewShieldedPool() *ShieldedPool`
- `CreateShieldedTx(sender, recipient types.Address, amount uint64, blinding [32]byte) *ShieldedTx`
- `VerifyShieldedTx(tx *ShieldedTx) bool`
- `(sp *ShieldedPool) AddCommitment / HasCommitment / CheckNullifier / RevealNullifier`
- `(sp *ShieldedPool) NullifierRoot() types.Hash` — sparse Merkle root of spent nullifiers

**Zero-knowledge transfer**
- `ZKTransfer` — full zero-knowledge shielded transfer with circuit proof
- `ShieldedCircuit` — gnark Groth16 circuit for range proof structure
- `CommitmentTree` — Merkle tree over commitments (inclusion proofs)
- `SparseMerkleTree` — sparse Merkle tree for the nullifier set

## Usage

```go
pool := bn254.NewShieldedPool()

tx := bn254.CreateShieldedTx(sender, recipient, 1e18, blinding)
if bn254.VerifyShieldedTx(tx) {
    pool.AddCommitment(tx.Commitment)
    pool.RevealNullifier(tx.NullifierHash)
}
root := pool.NullifierRoot()
```
