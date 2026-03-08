# crypto/merkle — Binary Merkle multi-proof generation and verification

[← crypto](../README.md)

## Overview

This package provides binary Merkle multi-proof generation and verification for use with SSZ beacon chain proofs (EIP-4881) and execution witness inclusion proofs. A multi-proof proves multiple leaf values simultaneously against a single root using a minimal set of sibling hashes, reducing proof size compared to independent single-leaf proofs when leaves share internal ancestors.

The tree uses generalized indices: the root is at index 1; node `i` has children at `2i` (left) and `2i+1` (right). Leaves of a depth-`d` tree are at indices `[2^d, 2^(d+1)-1]`. Internal hashing uses Keccak-256 of the concatenated left and right children.

## Functionality

**Types**
- `MerkleMultiProof` — contains `Leaves []MerkleLeaf`, `Proof []MerkleNode`, and `Depth uint`
- `MerkleLeaf` — `GeneralizedIndex uint64` + `Hash [32]byte`
- `MerkleNode` — `GeneralizedIndex uint64` + `Hash [32]byte`

**Tree construction**
- `BuildMerkleTree(leaves [][32]byte) ([][32]byte, uint)` — builds flat tree array indexed by generalized index
- `MerkleRoot(leaves [][32]byte) [32]byte` — convenience wrapper returning just the root

**Proof operations**
- `GenerateMultiProof(tree [][32]byte, depth uint, leafIndices []uint64) (*MerkleMultiProof, error)`
- `VerifyMultiProof(root [32]byte, proof *MerkleMultiProof) bool`
- `CompactMultiProof(proof *MerkleMultiProof) *MerkleMultiProof` — removes redundant proof nodes

**Generalized index helpers**
- `GeneralizedIndex(depth uint, leafPos uint64) uint64`
- `Parent(gi uint64) uint64` / `Sibling(gi uint64) uint64`
- `IsLeft(gi uint64) bool` / `DepthOfGI(gi uint64) uint`
- `PathToRoot(gi uint64) []uint64`
- `ProofSize(depth uint, k int) int` — upper-bound estimate

## Usage

```go
tree, depth := merkle.BuildMerkleTree(leaves)
proof, err := merkle.GenerateMultiProof(tree, depth, []uint64{2, 5})
proof = merkle.CompactMultiProof(proof)
ok := merkle.VerifyMultiProof(tree[1], proof)
```
