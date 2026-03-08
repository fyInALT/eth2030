# trie/bintrie — EIP-7864 binary Merkle trie

Implements the EIP-7864 binary state trie that replaces the Merkle Patricia Trie with a SHA-256 (or BLAKE3) binary tree supporting efficient stateless proofs.

[← trie](../README.md)

## Overview

The binary trie organises state under 32-byte keys. The first 31 bytes form a **stem** that routes through `InternalNode` branch nodes; the final byte selects one of 256 leaves in a `StemNode`. Hashing uses SHA-256 by default and optionally BLAKE3 (`NewWithHashFunc(HashFunctionBlake3)`).

Key derivation (`GetBinaryTreeKey`) follows the EIP-7864 spec: `SHA256(zeroHash[:12] || addr || key[:31] || 0x00)` with `key[31]` overwriting the last byte, making each account's storage disjoint. Dedicated helpers exist for the account basic-data leaf (`GetBinaryTreeKeyBasicData`), code hash leaf (`GetBinaryTreeKeyCodeHash`), storage slots (`GetBinaryTreeKeyStorageSlot`), and code chunks (`GetBinaryTreeKeyCodeChunk`).

`BinaryTrie` exposes high-level account and storage APIs (`GetAccount`, `UpdateAccount`, `GetStorage`, `UpdateStorage`, `UpdateContractCode`) that encode/decode the `BasicDataLeaf` format (version, code size, nonce, balance) and manage code chunking via `ChunkifyCode` (which returns `[][32]byte` per spec).

Proof generation and verification are in `proof.go` and `proof_verifier.go`. Epoch metadata for state expiry is written via `UpdateLeafMetadata`. An `EpochUpdater` in `epoch_updater.go` performs batch expiry updates.

## Functionality

### Types

| Type | Purpose |
|------|---------|
| `BinaryTrie` | Top-level trie: Get/Put/Delete, account and storage helpers, Hash/Copy |
| `InternalNode` | Branch node routing on key bits |
| `StemNode` | Leaf array node holding 256 values under a 31-byte stem |
| `Empty` | Sentinel for an empty subtree |

### Key Functions

- `New()` / `NewWithHashFunc(hashFunc)` / `Get(key)` / `Put(key, value)` / `Delete(key)` / `Hash()`
- `GetAccount(addr)` / `UpdateAccount(addr, acc, codeLen)` / `GetStorage(addr, key)` / `UpdateStorage(addr, key, value)` / `DeleteStorage(addr, key)`
- `UpdateContractCode(addr, code)` / `UpdateStem(key, values)` / `UpdateLeafMetadata(stem, subindex, epoch)`
- `GetBinaryTreeKey(addr, key)` / `GetBinaryTreeKeyBasicData(addr)` / `GetBinaryTreeKeyCodeHash(addr)` / `GetBinaryTreeKeyStorageSlot(addr, key)` / `StorageIndex(storageKey)`
- `ChunkifyCode(code)` (returns `[][32]byte`) / `PackBasicDataLeaf` / `UnpackBasicDataLeaf`

## Usage

```go
t := bintrie.New()
t.UpdateAccount(addr, &types.Account{Nonce: 1, Balance: big.NewInt(1e18)}, 0)
t.UpdateStorage(addr, storageKey, value)
root := t.Hash()

acc, _ := t.GetAccount(addr)
```
