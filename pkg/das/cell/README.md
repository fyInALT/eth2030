# das/cell — Cell-level gossip routing and subnet management

[← das](../README.md)

## Overview

This package implements cell-level gossip message routing and peer scoring for PeerDAS (EIP-7594). It defines the `CellMessage` type carrying individual 2048-byte cells with KZG proofs, a `GossipRouter` that maps cell indices to gossip subnets, and a `CellPeerScorer` that rewards and penalizes peers based on their cell-serving behavior.

The subnet topology follows the PeerDAS spec: 64 gossip subnets (`DataColumnSidecarSubnetCount`), each node subscribing to at least 4 (`CustodyRequirement`). Subnet assignment is computed deterministically from node ID using a hash-based permutation so that assignments are stable within an epoch.

## Functionality

**Types**
- `CellMessage` — `BlobIndex uint64`, `CellIndex uint64`, `Data []byte`, `Proof []byte`
- `SubnetConfig` — `NumSubnets uint64`, `SubnetsPerNode uint64`
- `GossipRouter` — routes cells to subnets; tracks per-node subnet assignments
- `CellPeerScorer` — tracks per-peer cell availability scores

**Constants** (from `constants.go`)
- `DataColumnSidecarSubnetCount = 64`
- `CustodyRequirement = 4`

**Router operations**
- `DefaultSubnetConfig() SubnetConfig`
- `NewGossipRouter(cfg SubnetConfig) *GossipRouter`
- `(r *GossipRouter) SubnetsForNode(nodeID [32]byte) []uint64`
- `(r *GossipRouter) SubnetForCell(cellIndex uint64) uint64`
- `(r *GossipRouter) RouteCell(msg *CellMessage, nodeID [32]byte) ([]uint64, error)`

**Peer scorer**
- `NewCellPeerScorer() *CellPeerScorer`
- `(s *CellPeerScorer) RecordSuccess(peer [32]byte, cellIndex uint64)`
- `(s *CellPeerScorer) RecordFailure(peer [32]byte, cellIndex uint64)`
- `(s *CellPeerScorer) Score(peer [32]byte) float64`

## Usage

```go
router := cell.NewGossipRouter(cell.DefaultSubnetConfig())
subnets, err := router.RouteCell(msg, localNodeID)
// publish msg on each of the returned subnets
```
