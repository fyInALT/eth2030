# netapi — net namespace JSON-RPC methods

[← rpc](../README.md)

## Overview

Package `netapi` implements the `net_` namespace JSON-RPC methods for network
status information. It exposes the network ID, listening state, connected peer
count, and maximum peer limit through a thin `Backend` interface, keeping the
API layer independent of the P2P implementation.

## Functionality

**`Backend` interface**

- `NetworkID() uint64` — network identifier (1 = mainnet)
- `IsListening() bool` — whether the node accepts inbound connections
- `PeerCount() int` — number of connected peers
- `MaxPeers() int` — configured peer limit

**`API`** — constructed with `NewAPI(backend Backend)`; `GetBackend()` for testing

**Methods and dispatch (`HandleNetRequest`)**

| JSON-RPC method | Go method | Returns |
|---|---|---|
| `net_version` | `API.Version()` | Network ID as decimal string |
| `net_listening` | `API.Listening()` | `bool` |
| `net_peerCount` | `API.PeerCount()` | Hex-encoded peer count |
| `net_maxPeers` | `API.MaxPeers()` | Hex-encoded max peers |

All methods return `ErrBackendNil` when called with a nil backend.
