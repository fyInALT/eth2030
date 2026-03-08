# das/network — DAS network layer and sampling manager

[← das](../README.md)

## Overview

This package implements the network layer for PeerDAS data availability sampling. It manages sample request/response cycles, tracks which fragments have been received, and coordinates blob reconstruction when enough fragments arrive. The `DASNetworkManager` (in `das_network_mgr.go`) provides a higher-level interface that integrates subnet-based request routing, peer selection, and sampling retry logic, wrapping the lower-level `DASNetwork` sampling engine.

Default parameters follow the Fulu PeerDAS spec: 64 gossip subnets, 8 samples per slot, minimum 4 custody subnets, 2048 bytes per column cell.

## Functionality

**Types**
- `DASNetworkConfig` — `NumSubnets`, `SamplesPerSlot`, `MinCustodySubnets`, `ColumnSize uint64`
- `SampleResponse` — `BlobIndex uint64`, `CellIndex uint64`, `Data []byte`, `Proof []byte`
- `DASNetwork` — core sampling engine with per-blob fragment tracking
- `DASNetworkManager` — full DAS manager with peer routing and retry

**Construction**
- `DefaultDASNetworkConfig() DASNetworkConfig`
- `NewDASNetwork(cfg DASNetworkConfig) *DASNetwork`
- `NewDASNetworkManager(cfg DASNetworkConfig) *DASNetworkManager`

**Sampling operations**
- `(n *DASNetwork) Start() error` / `Stop()`
- `(n *DASNetwork) RequestSample(blobIndex, cellIndex uint64) (*SampleResponse, error)`
- `(n *DASNetwork) StoreSample(resp *SampleResponse) error`
- `(n *DASNetwork) IsAvailable(blobIndex, cellIndex uint64) bool`
- `(n *DASNetwork) AddFragment(blobIndex, cellIndex uint64, data []byte) (bool, error)` — returns true when reconstruction threshold reached
- `(n *DASNetwork) ReconstructBlob(blobIndex uint64) ([]byte, error)`

**Manager operations**
- `(m *DASNetworkManager) SampleSlot(slot uint64, blobCount int) error`
- `(m *DASNetworkManager) GetSamplingStatus(slot uint64) map[uint64]bool`

**Errors**
- `ErrDASNotStarted`, `ErrInvalidBlobIdx`, `ErrInvalidCellIdx`, `ErrSampleNotAvailable`, `ErrVerificationFailed`, `ErrReconstructNotReady`, `ErrReconstructDone`, `ErrDuplicateFragment`, `ErrFragmentOutOfRange`

## Usage

```go
net := network.NewDASNetwork(network.DefaultDASNetworkConfig())
net.Start()

resp, err := net.RequestSample(blobIdx, cellIdx)
ready, _ := net.AddFragment(blobIdx, cellIdx, resp.Data)
if ready {
    blobData, _ := net.ReconstructBlob(blobIdx)
}
```
