# sync/beacon — Beacon chain and blob sidecar sync

Syncs beacon chain blocks and EIP-4844 blob sidecars, implementing the Deneb `BlobSidecarsByRange` and `BlobSidecarsByRoot` protocols with peer scoring, rate limiting, and partial-availability blob recovery.

[← sync](../README.md)

## Overview

The package provides two cooperating subsystems. `BeaconSyncer` downloads beacon blocks and blob sidecars for a range of slots concurrently, with configurable parallelism and retry logic. `BlobRecovery` reconstructs missing blobs from available ones using a 50%-threshold erasure scheme when some sidecars are absent.

`BlobSyncProtocol` manages the lower-level peer interaction: it validates individual `BlobSidecarV2` objects against Deneb spec rules (index bounds, non-zero KZG commitment and proof, slot matching), tracks per-peer quality scores (rewarding good responses, penalising bad or empty ones), enforces per-peer rate limits within a sliding time window, and caches validated sidecars indexed by slot.

`BlobSyncManager` is a higher-level download manager that tracks which blob indices have been requested per slot, accepts responses from identified peers, deduplicates blobs, verifies consistency across a slot's blobs via Keccak256 content hashing, and marks slots complete.

## Functionality

### Types

| Type | Purpose |
|------|---------|
| `BeaconSyncer` | Concurrent slot-range downloader for blocks and blob sidecars |
| `BlobRecovery` | Erasure-code-style reconstruction of missing sidecars |
| `BlobSyncProtocol` | Peer-scored, rate-limited sidecar request/response handler |
| `BlobSyncManager` | Per-slot blob download tracker with consistency verification |
| `BeaconBlock` | Beacon block with slot, proposer index, parent/state roots |
| `BlobSidecar` / `BlobSidecarV2` | 128 KiB blob with KZG commitment, proof, and inclusion proof |

### Key Functions

- `NewBeaconSyncer(config)` / `SyncSlotRange(from, to)` / `Cancel()`
- `NewBlobRecovery(custody)` / `AttemptRecovery(slot, available)`
- `NewBlobSyncProtocol(config)` / `ValidateSidecar(sc)` / `ProcessSidecarResponse(resp)` / `RequestBlobRange(start, end)`
- `NewBlobSyncManager(config)` / `RequestBlobs(slot, indices)` / `ProcessBlobResponseFromPeer(slot, index, blob, peer)` / `VerifyBlobConsistency(slot)`

## Usage

```go
syncer := beacon.NewBeaconSyncer(beacon.DefaultBeaconSyncConfig())
syncer.SetFetcher(myFetcher)
if err := syncer.SyncSlotRange(1000, 2000); err != nil {
    log.Fatal(err)
}
status := syncer.GetSyncStatus()
```
