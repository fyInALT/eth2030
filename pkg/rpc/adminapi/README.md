# adminapi — admin namespace JSON-RPC backend

[← rpc](../README.md)

## Overview

Package `adminapi` implements the `admin_` namespace JSON-RPC methods for node
administration. It exposes peer management, node identity, and data-directory
queries that Kurtosis and other devnet tooling rely on (particularly
`admin_nodeInfo` for enode discovery).

The package is split into two layers: `API` holds the business logic against the
`Backend` interface, and `DispatchAPI` translates raw `rpctypes.Request` /
`rpctypes.Response` envelopes to and from `API` calls.

## Functionality

**Types**

- `Backend` — interface: `NodeInfo`, `Peers`, `AddPeer`, `RemovePeer`, `ChainID`, `DataDir`
- `NodeInfoData` — geth-compatible node info struct (`name`, `id`, `enr`, `enode`, `listenAddr`, `ports`, `protocols`)
- `NodePorts` — discovery and listener ports
- `PeerInfoData` — connected peer descriptor (`id`, `name`, `remoteAddr`, `caps`, `static`, `trusted`)
- `API` — business-logic handler; constructed with `NewAPI(backend Backend)`
- `DispatchAPI` — JSON-RPC dispatcher; constructed with `NewDispatchAPI(backend Backend)`

**API methods (via `DispatchAPI.HandleAdminRequest`)**

| JSON-RPC method | Go method |
|---|---|
| `admin_nodeInfo` | `API.NodeInfo` |
| `admin_peers` | `API.Peers` |
| `admin_addPeer` | `API.AddPeer(url string) (bool, error)` |
| `admin_removePeer` | `API.RemovePeer(url string) (bool, error)` |
| `admin_datadir` | `API.DataDir() (string, error)` |
| `admin_startRPC` | `API.StartRPC(host string, port int) (bool, error)` |
| `admin_stopRPC` | `API.StopRPC() (bool, error)` |
| `admin_chainId` | `API.ChainID() (string, error)` |

Each method has a corresponding `Admin*` alias for backward compatibility. `API.GetBackend()` exposes the underlying `Backend` for testing.
