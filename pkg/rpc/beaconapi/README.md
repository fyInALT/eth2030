# beaconapi — Beacon API JSON-RPC methods

[← rpc](../README.md)

## Overview

Package `beaconapi` implements a subset of the Ethereum Beacon API as
JSON-RPC methods, allowing consensus-layer clients to query consensus state
through the standard RPC server without a separate HTTP server. It exposes 10
`beacon_*` methods covering genesis info, blocks, state roots, finality
checkpoints, validators, node health, and sync status.

Consensus state is held in a mutex-protected `ConsensusState` struct that
external components update as the chain progresses.

## Functionality

**Types**

- `BeaconAPI` — constructed with `NewBeaconAPI(state *ConsensusState, backend rpcbackend.Backend)`
- `ConsensusState` — holds `GenesisTime`, `GenesisValRoot`, `HeadSlot`, `FinalizedEpoch/Root`, `JustifiedEpoch/Root`, `Validators`, `IsSyncing`, `SyncDistance`, `Peers`; created with `NewConsensusState()`
- Response types: `GenesisResponse`, `BlockResponse`, `HeaderResponse`, `SignedHeaderData`, `BeaconHeaderMessage`, `StateRootResponse`, `FinalityCheckpointsResponse`, `Checkpoint`, `ValidatorListResponse`, `ValidatorEntry`, `ValidatorData`, `VersionResponse`, `SyncingResponse`, `PeerListResponse`, `BeaconPeer`
- `BeaconError` — error with numeric code (`BeaconErrNotFound=404`, `BeaconErrBadRequest=400`, `BeaconErrInternal=500`, `BeaconErrNotImplemented=501`)

**Registered methods** (via `RegisterBeaconRoutes(api *BeaconAPI)`)

| Method | Description |
|---|---|
| `beacon_getGenesis` | Genesis time, validators root, fork version |
| `beacon_getBlock` | Block fields at a given slot |
| `beacon_getBlockHeader` | Signed header at a given slot |
| `beacon_getStateRoot` | State root for `head`, `finalized`, `justified`, or slot |
| `beacon_getStateFinalityCheckpoints` | Justified and finalized checkpoints |
| `beacon_getStateValidators` | Full validator list |
| `beacon_getNodeVersion` | Client version string |
| `beacon_getNodeSyncing` | Sync distance and optimistic status |
| `beacon_getNodePeers` | Connected beacon peers |
| `beacon_getNodeHealth` | `healthy` or `syncing` status |
