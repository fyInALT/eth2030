# das/gossip — Cell-level gossip handler and column builder for PeerDAS

[← das](../README.md)

## Overview

This package implements the cell-level gossip protocol handler for PeerDAS data availability sampling. It sits between the network gossip layer and the reconstruction engine: `CellGossipHandler` receives, validates, deduplicates, and stores individual cells from gossip, then signals when enough cells are available for Reed-Solomon reconstruction. `CellGossipScorer` scores peers based on their cell-serving reliability. `ColumnBuilder` assembles complete `DataColumnSidecar` objects from accumulated gossip cells.

The handler supports a pluggable `CellValidator` interface so callers can inject KZG proof verification, size checks, or other integrity logic without coupling to a specific proof backend.

## Functionality

**Types**
- `CellGossipMessage` — `BlobIndex int`, `CellIndex int`, `Data []byte`, `KZGProof [48]byte`, `Slot uint64`
- `CellValidator` interface — `ValidateCell(msg CellGossipMessage) bool`
- `CellGossipHandler` — receives and tracks cells per blob; signals reconstruction readiness
- `CellGossipScorer` — per-peer scoring based on cell availability
- `ColumnBuilder` — assembles complete `DataColumnSidecar` from gossip cells

**Handler operations**
- `NewCellGossipHandler(numBlobs, cellsPerBlob int, validator CellValidator) *CellGossipHandler`
- `(h *CellGossipHandler) HandleCell(msg CellGossipMessage) error`
- `(h *CellGossipHandler) IsReadyForReconstruction(blobIndex int) bool`
- `(h *CellGossipHandler) GetCells(blobIndex int) []CellGossipMessage`
- `(h *CellGossipHandler) MissingCells(blobIndex int) []int`
- `(h *CellGossipHandler) Close()`

**Column builder**
- `NewColumnBuilder(numBlobs int) *ColumnBuilder`
- `(b *ColumnBuilder) AddCell(msg CellGossipMessage) error`
- `(b *ColumnBuilder) BuildSidecar(colIndex int) (*dastypes.DataColumnSidecar, error)`

**Errors**
- `ErrGossipHandlerClosed`, `ErrGossipCellDuplicate`, `ErrGossipCellValidation`, `ErrGossipCellNilMessage`, `ErrGossipBlobNotTracked`, `ErrGossipBroadcastNoPeers`

## Usage

```go
handler := gossip.NewCellGossipHandler(numBlobs, 128, myKZGValidator)
err := handler.HandleCell(cellMsg)
if handler.IsReadyForReconstruction(blobIdx) {
    cells := handler.GetCells(blobIdx)
    // pass to RS decoder
}
```
