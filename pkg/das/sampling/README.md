# das/sampling — PeerDAS custody and column sampling

Implements the consensus-spec custody group and column sampling algorithms for
PeerDAS (EIP-7594, Fulu). Provides deterministic column selection, per-validator
download/verification tracking, and structural sidecar validation.

## Overview

`sampling.go` implements the two core spec algorithms: `GetCustodyGroups` (maps
a node ID to a set of custody groups via a keccak hash chain) and
`ComputeColumnsForCustodyGroup` (maps each group to its column indices). Together
they determine which columns each node must custody.

`column_sampling.go` builds the per-validator `ColumnSampler` on top of those
primitives. For each slot it deterministically selects `SamplesPerSlot` (8)
columns to download, tracks download and verification state, computes an
availability score, and supports pruning of old slots.

`sampling_scheduler.go` and `peer_sampling_scheduler.go` implement higher-level
scheduling of sampling tasks across peers.

## Functionality

**Types**
- `ColumnSampler` — per-validator download/verification state tracker
- `ColumnSamplerConfig` — `SamplesPerSlot`, `NumberOfColumns`, `CustodyGroupCount`, `TrackSlots`
- `ColumnSample` — single downloaded sample record
- `ColumnAvailability` — per-slot required/downloaded/verified sets and score

**Functions**
- `GetCustodyGroups(nodeID [32]byte, count uint64) ([]CustodyGroup, error)`
- `ComputeColumnsForCustodyGroup(group CustodyGroup) ([]ColumnIndex, error)`
- `GetCustodyColumns(nodeID, count) ([]ColumnIndex, error)`
- `ShouldCustodyColumn(columnIndex, custodyColumns) bool`
- `VerifyDataColumnSidecar(sidecar *dastypes.DataColumnSidecar) error`
- `ColumnSubnet(columnIndex) SubnetID`
- `NewColumnSampler(config, nodeID) *ColumnSampler`
- `(*ColumnSampler).SelectColumns(slot) ([]ColumnIndex, error)`
- `(*ColumnSampler).InitSlot(slot) error`
- `(*ColumnSampler).RecordDownload(slot, col, dataSize) error`
- `(*ColumnSampler).VerifySample(slot, col, data, expectedRoot) error`
- `(*ColumnSampler).GetAvailability(slot) (*ColumnAvailability, error)`
- `(*ColumnSampler).IsAvailable(slot) bool`
- `(*ColumnSampler).PruneBefore(slot) int`

## Usage

```go
cs := sampling.NewColumnSampler(sampling.DefaultColumnSamplerConfig(), nodeID)
cs.InitSlot(slot)

required, _ := cs.SelectColumns(slot)
// download required columns ...
cs.RecordDownload(slot, col, len(data))
cs.VerifySample(slot, col, data, expectedRoot)

avail, _ := cs.GetAvailability(slot)
fmt.Println("available:", avail.Available, "score:", avail.Score)
```

[← das](../README.md)
