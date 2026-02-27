# Sprint 5, Story 5.1 — Tick Serialization (MarshalBinary/UnmarshalBinary)

**Sprint goal:** Enable P2P transmission of STARK mempool ticks.
**Files modified:** `pkg/txpool/stark_aggregation.go`
**Files tested:** `pkg/txpool/stark_recursion_test.go`

## Overview

The MempoolAggregationTick must be serializable for gossip transmission. Without `MarshalBinary`/`UnmarshalBinary`, the gossip handler has no way to encode/decode ticks.

## Gap (AUDIT-4)

**Severity:** CRITICAL
**File:** `pkg/txpool/stark_aggregation.go:46`
**Evidence:** No serialization methods existed on `MempoolAggregationTick`. The struct was only usable in-process.

## Implement

Binary encoding format (144 lines added):

```
MarshalBinary wire format:
┌─────────────────┬──────────────────┬────────────────────────┐
│ tick_number (8B) │ timestamp (8B)   │ peer_id_len (2B) + ID  │
├─────────────────┴──────────────────┴────────────────────────┤
│ valid_tx_count (4B) + [32B hash × N]                        │
├─────────────────────────────────────────────────────────────┤
│ discard_count (4B) + [32B hash × M]                         │
├──────────────────────────────────────────���──────────────────┤
│ bitfield_len (4B) + bitfield bytes                          │
├─────────────────────────────────────────────────────────────┤
│ merkle_root (32B)                                           │
├─────────────────────────────────────────────────────────────┤
│ has_proof (1B) + [trace_commitment (32B) if present]        │
└─────────────────────────────────────────────────────────────┘
```

**Design decisions:**
- Big-endian byte order (matches Ethereum convention)
- Tick number as uint64 (sufficient for ~584 billion years at 500ms ticks)
- Timestamp as UnixNano for sub-second precision
- Proof serialization is partial (trace commitment only) — full FRI layer serialization is future work

## Tests

```go
func TestTickMarshalUnmarshalRoundtrip(t *testing.T) {
    tick := &MempoolAggregationTick{
        TickNumber:    42,
        Timestamp:     time.Now(),
        PeerID:        "peer-1",
        ValidTxHashes: []types.Hash{{0x01}, {0x02}},
        ValidBitfield: []byte{0x03},
        TxMerkleRoot:  types.Hash{0xAA},
    }
    data, err := tick.MarshalBinary()
    require.NoError(t, err)
    require.True(t, len(data) <= MaxTickSize)

    var decoded MempoolAggregationTick
    err = decoded.UnmarshalBinary(data)
    require.NoError(t, err)
    require.Equal(t, tick.TickNumber, decoded.TickNumber)
    require.Equal(t, tick.PeerID, decoded.PeerID)
    require.Equal(t, len(tick.ValidTxHashes), len(decoded.ValidTxHashes))
}
```

## Codebase Locations

| File | Line | Purpose |
|------|------|---------|
| `pkg/txpool/stark_aggregation.go` | 67 | MarshalBinary |
| `pkg/txpool/stark_aggregation.go` | 133 | UnmarshalBinary |
| `pkg/txpool/stark_recursion_test.go` | — | Round-trip tests |
