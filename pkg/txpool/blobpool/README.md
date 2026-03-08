# txpool/blobpool — EIP-4844 and EIP-8070 blob transaction pool

## Overview

Package `blobpool` provides dedicated transaction pools for EIP-4844 type-3 blob transactions. It contains three pool implementations that share the same acceptance rules but differ in storage strategy: `BlobTxPool` (simple in-memory), `BlobPool` (EIP-8070 sparse memory with WAL journal and PeerDAS custody filtering), and `SparseBlobPool` (metadata-only tracking with slot-based expiry).

All pools enforce blob gas limits, per-account caps, blob fee cap thresholds, and eviction by lowest effective tip. `BlobPool` additionally persists sidecar data (KZG commitments, proofs, cell indices) to disk with JSONL write-ahead logging and applies custody column filtering for PeerDAS (EIP-7594).

## Functionality

**Types**
- `BlobTxPool` — simple concurrent pool; `Add`, `Get`, `Remove`, `Pending`, `SetExcessBlobGas`, `BlobBaseFee`
- `BlobPool` — EIP-8070 pool; `Add`, `AddBlobTx`, `RemoveBlobTx`, `GetBlobSidecar`, `PruneSidecars`, `SetBlobBaseFee`, `Close`
- `SparseBlobPool` — metadata-only pool; `AddBlobTx`, `GetBlobTx`, `RemoveBlobTx`, `PendingBlobTxs`, `PruneExpired`
- `BlobMetadata`, `BlobSidecar`, `SparseBlobEntry` — data structs for sidecar and metadata
- `CustodyConfig` — configures PeerDAS custody columns; `IsCustodyColumn`, `CustodyFilter`
- `BlobPoolConfig`, `BlobTxPoolConfig`, `SparseBlobPoolConfig` — per-pool configuration

**Functions**
- `CalcBlobBaseFee(excessBlobGas)` — EIP-4844 exponential base fee formula
- `CalcExcessBlobGas(parentExcess, parentBlobGasUsed)` — next-block excess blob gas
- `DefaultBlobPoolConfig`, `DefaultBlobTxPoolConfig`, `DefaultSparseBlobPoolConfig`

## Usage

```go
pool := blobpool.NewBlobPool(blobpool.DefaultBlobPoolConfig(), stateReader)
if err := pool.Add(tx); err != nil { ... }
pool.SetBlobBaseFee(newBaseFee)
pending := pool.PendingSorted()
```

[← txpool](../README.md)
