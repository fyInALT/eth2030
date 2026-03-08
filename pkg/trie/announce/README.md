# trie/announce — Announcement binary trie for state change broadcasts

Provides a compact, proof-friendly binary Merkle trie for broadcasting state diffs, implementing the EL sustainability "announce binary tree" track.

[← trie](../README.md)

## Overview

`AnnounceBinaryTrie` is a thread-safe binary Merkle trie keyed by `keccak256(rawKey)`. Bits are traversed MSB-first, placing each key at a unique leaf. The `Root()` method computes the Keccak256 Merkle root with lazy caching: dirty nodes are rehashed on demand. `Prove(key)` generates a `BinaryProofAnnounce` (sibling hashes + direction bits from leaf to root), and `VerifyAnnounceProof` reconstructs the path to verify inclusion.

`AnnouncementSet` accumulates `StateChange` records (address, storage slot, old/new values) and materialises them into an `AnnounceBinaryTrie` via `BuildAnnouncementTree`. Each change is keyed by `keccak256(addr || slot)` with a 64-byte value encoding the old and new values side-by-side.

## Functionality

### Types

| Type | Purpose |
|------|---------|
| `AnnounceBinaryTrie` | Thread-safe binary Merkle trie for key-value state announcements |
| `BinaryProofAnnounce` | Inclusion proof: key, value, sibling hashes, direction bits |
| `AnnouncementSet` | Collects `StateChange` records and builds an announcement trie |
| `StateChange` | Address, storage slot, old value, new value |
| `BinaryNode` | Exported tree node for external inspection |

### Key Functions

- `NewAnnounceBinaryTrie()` / `Insert(key, value)` / `Get(key)` / `Delete(key)`
- `Root()` / `Prove(key)` / `VerifyAnnounceProof(root, key, proof)`
- `NewAnnouncementSet()` / `AddStateChange(addr, slot, old, new)` / `BuildAnnouncementTree()`
- `ExportBinaryNode()` / `Len()`

## Usage

```go
as := announce.NewAnnouncementSet()
as.AddStateChange(addr, slot, oldVal, newVal)
trie := as.BuildAnnouncementTree()
root := trie.Root()

proof, _ := trie.Prove(key)
valid := announce.VerifyAnnounceProof(root, key, proof)
```
