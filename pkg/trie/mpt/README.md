# trie/mpt — Merkle Patricia Trie

A full Merkle Patricia Trie implementation with iterator, proof generation and verification, sync scheduler, reference-counted node database, and an in-memory binary trie companion.

[← trie](../README.md)

## Overview

`Trie` is the in-memory MPT. It supports `Get`, `Put`, `Delete`, and `Hash` (Keccak256 root). Node encoding follows Ethereum's compact hex-prefix encoding for short nodes and RLP for full nodes. A `TrieCommitter` flushes dirty nodes to a `Database` (or `RefcountDB`) that reference-counts nodes so they can be safely shared across multiple trie versions.

**Proofs** — `proof.go` exports `Prove(key)` which walks the trie and collects the Merkle proof nodes. `ProofVerifier` and `ProofVerifierDeep` verify inclusion and non-inclusion. `AccountProof` provides an Ethereum-style JSON account proof.

**Iterator** — `Iterator` traverses key-value pairs in order. `NodeIterator` visits trie nodes for state export or diff computation. `BinaryIterator` and its extended variant provide bit-based traversal for binary trie contexts.

**Sync** — `SyncScheduler` coordinates trie healing by maintaining a set of missing node hashes and dispatching requests.

**DiffTracker** tracks modified nodes between two trie versions. `KVHasher` computes a hash over a sorted key-value set without constructing the trie.

**Binary companion** — `BinaryTrie` (in `binary.go`) is a simple binary Merkle trie backed by a sorted key-value list. `NewBinaryTrie()` / `PutHashed(hash, value)` / `Hash()` are used by the migration package.

## Functionality

### Key Types

| Type | Purpose |
|------|---------|
| `Trie` | Merkle Patricia Trie: Get/Put/Delete/Hash/Prove |
| `Database` | Node database with dirty/clean caches |
| `RefcountDB` | Reference-counted node store |
| `TrieCommitter` | Flushes dirty trie nodes to storage |
| `Iterator` / `NodeIterator` | Sequential and node-level traversal |
| `SyncScheduler` | Missing-node tracking for trie healing |
| `ProofVerifier` / `ProofVerifierDeep` | Merkle proof verification |
| `AccountProof` | Ethereum JSON-RPC style account + storage proof |
| `BinaryTrie` | Simple binary Merkle trie (used by migration) |
| `DiffTracker` | Tracks node changes between trie versions |

### Key Functions

- `New()` / `Get(key)` / `Put(key, value)` / `Delete(key)` / `Hash()` / `Prove(key)`
- `NewIterator(trie)` / `NewNodeIterator(trie)`
- `NewSyncScheduler(root)` / `Missing()` / `Process(hash, data)`
- `NewBinaryTrie()` / `PutHashed(hash, value)` / `Hash()`
- `NewDatabase()` / `NewRefcountDB()` / `NewTrieCommitter(db)`

## Usage

```go
t := mpt.New()
t.Put([]byte("key"), []byte("value"))
root := t.Hash()

proof, _ := t.Prove([]byte("key"))
v := mpt.NewProofVerifier()
ok, _ := v.VerifyProof(root, []byte("key"), proof)
```
