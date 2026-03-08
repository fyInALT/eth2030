# trie

Merkle Patricia Trie (MPT) and binary Merkle trie implementations for Ethereum state storage, including EIP-7864 binary trie support and MPT-to-binary migration.

## Overview

The `trie` package provides the core cryptographic data structures used to represent and authenticate Ethereum world state. It exposes the Merkle Patricia Trie (MPT) as the current production trie and the binary Merkle trie (EIP-7864) as the next-generation ZK-proof-friendly alternative.

The top-level package re-exports the MPT public API from `trie/mpt` for backward compatibility, exposing `Trie`, `AccountProof`, `StorageProof`, `New`, `VerifyProof`, and `ProveAccountWithStorage`. All callers using the canonical import path continue to work unchanged as subpackages evolve.

The binary trie subpackage (`bintrie`) implements the EIP-7864 specification: a SHA-256-based binary tree where each 32-byte key is split into a 31-byte stem (navigating internal nodes) and a 1-byte suffix (selecting a leaf among 256 within a StemNode). This structure is designed for efficient inclusion proof generation in stateless clients and ZK circuits.

## Table of Contents

- [Functionality](#functionality)
- [Subpackages](#subpackages)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### Merkle Patricia Trie (MPT)

The `mpt` subpackage implements a full in-memory MPT with four node types: `shortNode` (extension/leaf), `fullNode` (branch with 17 children), `valueNode` (leaf data), and `hashNode` (unresolved reference). Key operations:

- `New() *Trie` — creates an empty trie
- `Get(key []byte) ([]byte, error)` — retrieves a value; returns `ErrNotFound` if absent
- `Put(key, value []byte) error` — inserts or updates; deletes if value is empty
- `Delete(key []byte) error` — removes a key (no-op if absent)
- `Hash() types.Hash` — computes the Keccak-256 root hash with caching of dirty nodes
- `Len() int` / `Empty() bool` — cardinality helpers

Node splitting and branch collapsing follow the Ethereum Yellow Paper MPT spec. The `emptyRoot` constant is `Keccak256(RLP(""))`.

### MPT Proof Generation

`ProveAccountWithStorage(stateTrie, addr, storageTrie, storageKeys)` generates an `AccountProof` containing Merkle proof paths for an account and any requested storage slots. `VerifyProof(rootHash, key, proof)` verifies a proof path against a known root.

### Binary Merkle Trie (EIP-7864)

The `bintrie` subpackage implements the EIP-7864 binary trie with SHA-256 hashing. Key functions:

- `GetBinaryTreeKey(addr, key)` — derives the 32-byte tree key: `SHA256(zeroHash[:12] || addr || key[:31] || 0x00)` with `key[31]` as the suffix
- `GetBinaryTreeKeyBasicData(addr)` — key for account basic data leaf (suffix `0x00`)
- `GetBinaryTreeKeyCodeHash(addr)` — key for code hash leaf (suffix `0x01`)
- `GetBinaryTreeKeyStorageSlot(addr, key)` — key for a storage slot, with header storage and main storage offset encoding

The `mpt` package also contains a `BinaryTrie` implementation keyed by keccak256-hashed 32-byte keys. Tree path traversal walks bits MSB-first (bit 0 = left, bit 1 = right).

### MPT-to-Binary Migration

`migrate.MigrateFromMPT(source *mpt.Trie) *mpt.BinaryTrie` iterates all key-value pairs in an MPT and re-inserts them into a binary trie with keys re-derived via keccak256, implementing the on-chain state migration path described in the EIP-7864 specification.

### Supporting Infrastructure

- `announce/` — trie announcement protocol for peer state synchronization
- `nodecache/` — LRU cache for resolved trie nodes, reducing redundant hashing
- `prune/` — bloom-filter-based reachability pruner to remove unreferenced nodes
- `stack/` — stack-based trie iterator used by migration and sync

## Subpackages

| Subpackage | Description |
|------------|-------------|
| [`mpt/`](./mpt/) | Merkle Patricia Trie: Get/Put/Delete/Hash, binary trie, iterator, account proofs |
| [`bintrie/`](./bintrie/) | EIP-7864 binary Merkle trie with SHA-256 hashing and StemNode structure |
| [`migrate/`](./migrate/) | MPT-to-binary trie migration: `MigrateFromMPT` |
| [`announce/`](./announce/) | Trie announcement protocol for peer state sync |
| [`nodecache/`](./nodecache/) | LRU node cache for resolved trie nodes |
| [`prune/`](./prune/) | Bloom-filter reachability pruner for unreferenced nodes |
| [`stack/`](./stack/) | Stack-based trie iterator for traversal and migration |

## Usage

```go
import "github.com/eth2030/eth2030/trie"

// Create and use a Merkle Patricia Trie.
t := trie.New()
t.Put([]byte("account:0xdead"), encodedAccount)
val, err := t.Get([]byte("account:0xdead"))
root := t.Hash()

// Verify a Merkle proof.
value, err := trie.VerifyProof(rootHash, key, proofPath)

// Generate account + storage proofs.
proof, err := trie.ProveAccountWithStorage(stateTrie, addr, storageTrie, storageKeys)
```

```go
import "github.com/eth2030/eth2030/trie/bintrie"

// Derive EIP-7864 binary trie keys.
basicKey := bintrie.GetBinaryTreeKeyBasicData(addr)
codeKey  := bintrie.GetBinaryTreeKeyCodeHash(addr)
slotKey  := bintrie.GetBinaryTreeKeyStorageSlot(addr, storageKey)
```

```go
import (
    "github.com/eth2030/eth2030/trie/mpt"
    "github.com/eth2030/eth2030/trie/migrate"
)

// Migrate existing MPT state to binary trie.
mptTrie := mpt.New()
// ... populate mptTrie ...
binaryTrie := migrate.MigrateFromMPT(mptTrie)
```

## Documentation References

- [EIP-7864: Binary Merkle Trie](https://eips.ethereum.org/EIPS/eip-7864)
- [Ethereum Yellow Paper: MPT Specification](https://ethereum.github.io/yellowpaper/paper.pdf)
- [L1 Strawmap Roadmap](../../docs/ROADMAP.md)
