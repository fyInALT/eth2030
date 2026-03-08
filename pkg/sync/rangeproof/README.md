# sync/rangeproof — Range proofs for snap sync state downloads

Creates, verifies, splits, and merges Merkle range proofs for snap sync, enabling cryptographic integrity checks on parallel state range downloads.

[← sync](../README.md)

## Overview

During snap sync, a server provides ranges of sorted key-value pairs from the state trie along with Merkle proofs for the range boundaries. The client uses these proofs to confirm that no keys were omitted within the delivered range.

`RangeProver` implements the core operations. `CreateRangeProof` builds a proof that anchors boundary keys to the state root. `VerifyRangeProof` checks that keys are sorted, counts match values, and the first proof node hashes back to the expected root.

`SplitRange` divides a key range into N equal sub-ranges (using `big.Int` arithmetic over the 256-bit key space) for parallel downloading from multiple peers. `MergeRangeProofs` merges sequential proofs back into a single deduplicated proof. `ComputeRangeHash` produces a Keccak256 fingerprint of a key-value set for independent integrity checks.

## Functionality

### Types

| Type | Purpose |
|------|---------|
| `RangeProver` | Creates, verifies, splits, and merges range proofs |
| `RangeProof` | Sorted keys, values, and Merkle proof nodes |
| `RangeRequest` | Root, origin, limit, byte/count limits for a download request |
| `AccountRange` | Start/end account hashes and completeness flag |

### Key Functions

- `NewRangeProver()` / `CreateRangeProof(keys, values, root)` / `VerifyRangeProof(root, proof)`
- `SplitRange(origin, limit, n)` / `MergeRangeProofs(proofs)` / `ComputeRangeHash(keys, values)`
- `PadTo32(b)` — utility for 256-bit key arithmetic

## Usage

```go
rp := rangeproof.NewRangeProver()
proof := rp.CreateRangeProof(keys, values, stateRoot)
ok, err := rp.VerifyRangeProof(stateRoot, proof)

// Split for parallel download across 4 peers:
subRanges := rp.SplitRange(origin, limit, 4)
```
