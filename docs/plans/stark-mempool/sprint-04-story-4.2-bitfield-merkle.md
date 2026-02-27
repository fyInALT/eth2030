# Sprint 4, Story 4.2 — Bitfield + Merkle Root Public Inputs

**Sprint goal:** Add compact public inputs for efficient tick verification.
**Files modified:** `pkg/txpool/stark_aggregation.go`
**Files tested:** `pkg/txpool/stark_recursion_test.go`

## Overview

Per the ethresear.ch proposal, the STARK proof should include compact public inputs — a bitfield of valid transaction indices and a Merkle root of valid transaction hashes — enabling O(1) membership checks by receivers instead of iterating all transactions.

## Gap (GAP-STARK2)

**Severity:** IMPORTANT
**File:** `pkg/txpool/stark_aggregation.go` — `GenerateTick()` at line 315
**Evidence:** The STARK proof proved "these txs are valid" but didn't include a compact bitfield or hash list as public input.

## Implement

### Step 1: Add fields to MempoolAggregationTick

```go
type MempoolAggregationTick struct {
    // ... existing fields ...
    ValidBitfield []byte     // bit i set if tx i is valid
    TxMerkleRoot  types.Hash // Merkle root of valid tx hashes
}
```

### Step 2: Populate in GenerateTick

```go
func (sa *STARKAggregator) GenerateTick() (*MempoolAggregationTick, error) {
    // ... generate STARK proof ...

    // Build bitfield: one bit per transaction in the pool.
    bitfieldSize := (len(sa.validTxs) + 7) / 8
    bitfield := make([]byte, bitfieldSize)
    for i := range sa.validTxs {
        bitfield[i/8] |= 1 << (uint(i) % 8)
    }

    // Compute Merkle root of valid tx hashes.
    merkleRoot := computeTxMerkleRoot(validHashes)

    tick.ValidBitfield = bitfield
    tick.TxMerkleRoot = merkleRoot
    return tick, nil
}
```

### Step 3: computeTxMerkleRoot helper

```go
// computeTxMerkleRoot builds a binary Merkle tree over transaction hashes
// using SHA-256 and returns the root.
func computeTxMerkleRoot(hashes []types.Hash) types.Hash {
    if len(hashes) == 0 {
        return types.Hash{}
    }
    layer := make([]types.Hash, len(hashes))
    copy(layer, hashes)
    // Pad to power of 2.
    for len(layer)&(len(layer)-1) != 0 {
        layer = append(layer, types.Hash{})
    }
    for len(layer) > 1 {
        var next []types.Hash
        for i := 0; i < len(layer); i += 2 {
            h := sha256.New()
            h.Write(layer[i][:])
            h.Write(layer[i+1][:])
            var parent types.Hash
            copy(parent[:], h.Sum(nil))
            next = append(next, parent)
        }
        layer = next
    }
    return layer[0]
}
```

## ethresear.ch Spec Reference

> The public inputs to the STARK proof should include a bitfield of valid transaction indices and a Merkle root of the valid transaction hashes. This allows receivers to check membership in O(1) time using Merkle proofs, rather than iterating over the full transaction list.

## Codebase Locations

| File | Line | Purpose |
|------|------|---------|
| `pkg/txpool/stark_aggregation.go` | 46 | ValidBitfield and TxMerkleRoot fields |
| `pkg/txpool/stark_aggregation.go` | 350 | Bitfield computation in GenerateTick |
| `pkg/txpool/stark_aggregation.go` | 360 | computeTxMerkleRoot helper |
