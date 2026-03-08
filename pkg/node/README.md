# Package node

Top-level ETH2030 full node lifecycle, wiring together blockchain, Engine API, JSON-RPC, P2P, and transaction pool.

## Overview

The `node` package is the integration point for all ETH2030 subsystems. `Node` owns the persistent database, blockchain, transaction pool, P2P server, JSON-RPC server (HTTP + WebSocket), Engine API server, metrics server, and anonymous transaction transport (mixnet). It initializes each subsystem in the correct order, resolves genesis state from a file or built-in network config, manages graceful startup and shutdown, and propagates configuration to every subsystem.

The `Config` struct mirrors the CLI flags exposed by `cmd/eth2030`: HTTP-RPC, Engine API (authenticated), WebSocket, metrics, P2P, fork overrides, finality mode, BLS backend selection, slot duration, attester sampling, and mixnet mode. A `ServiceRegistry` provides dependency-aware lifecycle management for subsystems registered as independent services.

The `eventbus` subpackage provides a typed publish/subscribe bus for internal subsystem communication (new blocks, new transactions, peer events, sync state changes). The `healthcheck` and `lifecycle` subpackages offer reusable health check and service lifecycle abstractions.

## Table of Contents

- [Functionality](#functionality)
- [Subpackages](#subpackages)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### Node

```go
New(config *Config) (*Node, error)
```

`New` performs all initialization:
1. Validates `Config`; generates a JWT secret if none is configured
2. Opens the persistent `rawdb.FileDB` at `<datadir>/chaindata`
3. Loads genesis from `--override.genesis` JSON file or built-in network config (mainnet / sepolia / holesky); applies fork timestamp overrides
4. Initializes `chain.Blockchain`, `txpool.TxPool`
5. Sets up STARK mempool gossip (`TopicManager`, `STARKAggregator`)
6. Compiles AA proof circuit in a background goroutine
7. Initializes `p2p.Server` with bootnodes, discovery port, and NAT
8. Auto-probes anonymous transport (Tor → Nym → simulated mixnet)
9. Initializes `rpc.ExtServer` (HTTP, auth, rate limit, CORS, body limits) and `engine.EngineAPI`

Key methods on `Node`:
- `Start() error` — starts all subsystems (P2P, RPC, Engine API, WebSocket, metrics)
- `Stop() error` — graceful shutdown in reverse order
- `Wait()` — blocks until `Stop` is called
- `Blockchain() *chain.Blockchain`
- `TxPool() *txpool.TxPool`
- `Config() *Config`
- `Running() bool`

### Config

`Config` holds all node parameters with `DefaultConfig()` providing sensible defaults:

| Field | Default | Description |
|-------|---------|-------------|
| `DataDir` | `~/.ETH2030` | Root data directory |
| `Network` | `mainnet` | Network name (`mainnet`, `sepolia`, `holesky`) |
| `P2PPort` | 30303 | TCP devp2p port |
| `RPCPort` | 8545 | HTTP JSON-RPC port |
| `HTTPAddr` | `0.0.0.0` | HTTP listen address |
| `EnginePort` | 8551 | Engine API port |
| `AuthAddr` | `0.0.0.0` | Engine API listen address |
| `WSPort` | 8546 | WebSocket RPC port |
| `MetricsPort` | 9001 | Prometheus metrics port |
| `SyncMode` | `snap` | Sync strategy (`full` or `snap`) |
| `GCMode` | `full` | State pruning (`full` or `archive`) |
| `MaxPeers` | 50 | Maximum P2P peer count |
| `FinalityMode` | `ssf` | Finality engine (`ssf` or `minimmit`) |
| `BLSBackend` | `blst` | BLS backend (`blst` or `pure-go`) |
| `SlotDuration` | `6s` | Slot timing (`4s` or `6s`) |
| `MixnetMode` | `simulated` | Anonymous transport (`simulated`, `tor`, `nym`) |

`Config.Validate()` checks all port ranges, module names, and enum values. Address helpers: `P2PAddr()`, `RPCAddr()`, `AuthListenAddr()`, `WSListenAddr()`, `MetricsListenAddr()`, `JWTSecretPath()`.

### Genesis Loader

`genesis_loader.go` parses Kurtosis-compatible `genesis.json` files (geth genesis format). Supported fields include chain config with timestamp forks (`shanghaiTime`, `cancunTime`, `pragueTime`, `amsterdamTime`, `hogotaTime`, `iPlusTime`), initial account alloc, gas limit, base fee, and difficulty.

### Service Registry

`ServiceRegistry` provides dependency-aware lifecycle management:

```go
NewServiceRegistry(maxSize int) *ServiceRegistry
```

- `Register(desc *ServiceDescriptor)` — registers a service with optional dependencies, priority, and health function
- `Start() []error` — topological sort (Kahn's algorithm) with priority-based tiebreaking; skips services whose dependencies failed
- `Stop() []error` — stops in reverse start order
- `HealthCheck() map[string]bool` — returns running status per service
- `GetState(name)` / `RunningCount()` / `Names()`

`ServiceState` values: `StateCreated`, `StateStarting`, `StateRunning`, `StateStopping`, `StateStopped`, `StateFailed`.

### RPC Handler

`rpc_handler.go` wires the node backend into the RPC dispatch layer, exposing `eth_`, `net_`, `web3_`, `admin_`, `debug_`, `txpool_` namespaces. The WebSocket handler in `node.go` upgrades HTTP connections and dispatches JSON-RPC over persistent WebSocket connections using an in-memory round-trip.

## Subpackages

| Subpackage | Description |
|------------|-------------|
| [`eventbus/`](./eventbus/) | Typed publish/subscribe event bus for inter-subsystem communication (chain.newBlock, tx.new, p2p.newPeer, sync.started, etc.) |
| [`healthcheck/`](./healthcheck/) | Reusable health checker abstraction for subsystem readiness probes |
| [`lifecycle/`](./lifecycle/) | Service lifecycle primitives (start/stop/wait patterns) |

## Usage

```go
cfg := node.DefaultConfig()
cfg.DataDir  = "/data/eth2030"
cfg.Network  = "mainnet"
cfg.JWTSecret = "/etc/eth2030/jwtsecret"
cfg.Bootnodes = "enode://..."

n, err := node.New(&cfg)
if err != nil {
    log.Fatal(err)
}
if err := n.Start(); err != nil {
    log.Fatal(err)
}
n.Wait() // blocks until Stop()
```

To load a custom devnet genesis:

```go
cfg.GenesisPath = "/path/to/genesis.json"
cfg.GlamsterdamOverride = &timestamp  // optional fork timestamp override
```

## Documentation References

- [ETH2030 CLI: cmd/eth2030/](../cmd/eth2030/)
- [ETH2030 geth binary: cmd/eth2030-geth/](../cmd/eth2030-geth/)
- ETH2030 Engine API: `pkg/engine/`
- ETH2030 RPC server: `pkg/rpc/`
- ETH2030 P2P: `pkg/p2p/`
- [Kurtosis devnet configs](../devnet/kurtosis/)
