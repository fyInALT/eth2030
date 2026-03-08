# das/custody — PeerDAS column custody management (EIP-7594)

[← das](../README.md)

## Overview

This package implements data custody for PeerDAS as specified in the Fulu consensus spec (`consensus-specs/specs/fulu/das-core.md`). A node custodies a set of column groups determined by its node ID and rotates custody assignments on epoch boundaries. The package coordinates three sub-managers: `CustodyManager` (top-level), `ColumnCustodyManager` (per-column data storage and retrieval), and `CustodySubnetManager` (subnet subscription assignments).

In addition, `custody_proof.go` and `proof_custody.go` provide cryptographic custody proofs that allow a node to demonstrate to peers which columns it is custodying, using KZG or hash-based proofs. `custody_verify.go` verifies incoming custody claims.

## Functionality

**Types**
- `CustodyManagerConfig` — `CustodyRequirement`, `NumberOfColumns`, `NumberOfCustodyGroups`, `SlotsPerEpoch`, `RetentionEpochs`, `MaxTrackedSlots`
- `CustodyEpochState` — columns custodied, completeness tracking for one epoch
- `CustodyManager` — top-level manager; thread-safe
- `ColumnCustodyManager` — per-column data store with WAL-backed retention
- `CustodySubnetManager` — manages gossip subnet subscriptions for custodied columns
- `CustodyProof` / `ProofCustody` — cryptographic proof of custody

**Construction**
- `DefaultCustodyManagerConfig() CustodyManagerConfig`
- `NewCustodyManager(cfg CustodyManagerConfig, nodeID [32]byte) *CustodyManager`

**Key operations**
- `(m *CustodyManager) Initialize(epoch uint64) error`
- `(m *CustodyManager) IsCustodied(col dastypes.ColumnIndex) bool`
- `(m *CustodyManager) StoreColumn(slot uint64, col dastypes.ColumnIndex, data []byte) error`
- `(m *CustodyManager) GetColumn(slot uint64, col dastypes.ColumnIndex) ([]byte, error)`
- `(m *CustodyManager) IsSlotComplete(slot uint64) bool`
- `(m *CustodyManager) RotateEpoch(newEpoch uint64) error`
- `(m *CustodyManager) GenerateCustodyProof(epoch uint64) (*CustodyProof, error)`
- `VerifyCustodyProof(proof *CustodyProof, nodeID [32]byte) bool`

## Usage

```go
cfg := custody.DefaultCustodyManagerConfig()
mgr := custody.NewCustodyManager(cfg, localNodeID)
mgr.Initialize(currentEpoch)

if mgr.IsCustodied(colIdx) {
    mgr.StoreColumn(slot, colIdx, cellData)
}
if mgr.IsSlotComplete(slot) {
    // all custodied columns for this slot are available
}
```
