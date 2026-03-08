# txpool/stark — STARK-proof mempool aggregation and P2P tick broadcasting

## Overview

Package `stark` implements Vitalik's recursive STARK mempool proposal. Every 500ms (configurable), `STARKAggregator` collects all locally validated transactions, builds an execution trace of their hashes and gas usage, generates a `STARKProof` over the trace via the proof aggregation framework, and emits a `MempoolAggregationTick`. Ticks are broadcast to peers via a `P2PBroadcaster` interface and merged from remote peers after STARK proof verification. The `peertick.PeerTickCache` prevents re-validating transactions already proven by peers.

## Functionality

**Types**
- `STARKAggregator` — `Start`, `Stop`, `AddValidatedTx`, `RemoveTx`, `GenerateTick`, `MergeTick`, `MergeTickAtSlot`, `BroadcastTick`, `SetBroadcaster`, `PendingCount`, `TickChannel`
- `MempoolAggregationTick{TickNumber, Timestamp, ValidTxHashes, AggregateProof, DiscardList, ValidBitfield, TxMerkleRoot, PeerID}` — wire format with `MarshalBinary` / `UnmarshalBinary`
- `ValidatedTx{TxHash, ValidationProof, GasUsed, RemoteProven}` — per-tx state in aggregator
- `P2PBroadcaster` — interface for `GossipMempoolStarkTick(data []byte) error`

**Functions**
- `NewSTARKAggregator(peerID)` — default 500ms tick interval
- `NewSTARKAggregatorWithInterval(peerID, interval)`
- `TickHash(tick)` — deterministic SHA-256 hash of a tick for comparison

**Constants**
- `DefaultTickInterval = 500ms`, `MaxTickTransactions = 10000`, `MaxTickSize = 128KB`

## Usage

```go
agg := stark.NewSTARKAggregator("node-1")
agg.SetBroadcaster(p2pLayer)
agg.Start()
agg.AddValidatedTx(txHash, proof, gasUsed)

// Receive remote tick:
var tick stark.MempoolAggregationTick
tick.UnmarshalBinary(data)
agg.MergeTickAtSlot(&tick, currentSlot)
```

[← txpool](../README.md)
