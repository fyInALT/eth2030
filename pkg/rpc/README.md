# rpc

JSON-RPC 2.0 server with 50+ Ethereum execution-layer methods, WebSocket subscriptions, Beacon API, and batch request handling.

## Overview

The `rpc` package is the top-level entry point for ETH2030's JSON-RPC layer. It re-exports types and constructors from a set of focused subpackages and wires them together into a unified server. The package covers the standard `eth_`, `debug_`, `net_`, and `admin_` namespaces as well as a Beacon API implementation for consensus-layer clients.

The server supports HTTP, WebSocket, and authenticated Engine API transports. Batch request handling follows the JSON-RPC 2.0 specification. An event filter and subscription system provides `eth_newFilter`, `eth_getLogs`, and `eth_subscribe` (newHeads, logs, newPendingTransactions) over WebSocket. The `admin_` namespace exposes node identity, peer management, and dynamic RPC configuration.

The `Backend` interface is the sole dependency of the RPC layer on the rest of the node; it provides access to blocks, state, transactions, the transaction pool, and chain configuration without importing concrete implementations.

## Table of Contents

- [Functionality](#functionality)
- [Subpackages](#subpackages)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### Top-Level Re-Exports

The `rpc` package re-exports the following types as aliases for backward compatibility:

```go
type Backend          = rpcbackend.Backend
type EthAPI           = ethapi.EthAPI
type BatchHandler     = rpcbatch.BatchHandler
type BatchResponse    = rpcbatch.BatchResponse
type AdminBackend     = adminapi.Backend
type AdminDispatchAPI = adminapi.DispatchAPI
type ServerConfig     = rpcserver.ServerConfig
```

Block number sentinels: `LatestBlockNumber`, `PendingBlockNumber`, `EarliestBlockNumber`, `SafeBlockNumber`, `FinalizedBlockNumber`.

Error codes: `ErrCodeParse`, `ErrCodeInvalidRequest`, `ErrCodeMethodNotFound`, `ErrCodeInvalidParams`, `ErrCodeInternal`, `ErrCodeHistoryPruned`.

### EthAPI (`eth_` namespace)

`NewEthAPI(backend Backend) *EthAPI` creates the primary Ethereum RPC handler. Methods include:

- **Block queries**: `eth_blockNumber`, `eth_getBlockByHash`, `eth_getBlockByNumber`, `eth_getBlockTransactionCountByHash/Number`, `eth_getUncleCountByBlockHash/Number`
- **Transaction methods**: `eth_getTransactionByHash`, `eth_getTransactionByBlockHashAndIndex`, `eth_getTransactionReceipt`, `eth_sendRawTransaction`, `eth_call`, `eth_estimateGas`
- **State queries**: `eth_getBalance`, `eth_getCode`, `eth_getStorageAt`, `eth_getTransactionCount`, `eth_getProof`
- **Fee methods**: `eth_gasPrice`, `eth_maxPriorityFeePerGas`, `eth_feeHistory`, `eth_blobBaseFee`
- **Filter and subscription**: `eth_newFilter`, `eth_newBlockFilter`, `eth_newPendingTransactionFilter`, `eth_getFilterChanges`, `eth_getFilterLogs`, `eth_getLogs`, `eth_uninstallFilter`
- **AA methods**: `eth_sendUserOperation`, `eth_getUserOperationByHash`
- **Access list**: `eth_createAccessList`
- **Chain info**: `eth_chainId`, `eth_syncing`, `eth_coinbase`, `eth_accounts`

### Beacon API (`beacon_` namespace)

The `beaconapi` subpackage implements 16 Beacon API endpoints used by consensus-layer clients:

- `beacon_getGenesis` — genesis time, validators root, fork version
- `beacon_getBlockHeader`, `beacon_getBlock` — beacon block headers and bodies
- `beacon_getState`, `beacon_getStateValidators`, `beacon_getStateValidator` — beacon state and validator queries
- `beacon_getStateFinalityCheckpoints` — justified/finalized checkpoints
- `beacon_getHeadSlot`, `beacon_getForkSchedule` — chain head and fork information
- `beacon_submitBlock` — block submission from validators
- `beacon_getAttestations`, `beacon_submitAttestation` — attestation propagation
- `beacon_getBlobSidecars` — EIP-4844 blob sidecar retrieval

### Debug API (`debug_` namespace)

The `debugapi` subpackage provides:

- `debug_traceTransaction` — EVM execution trace returning opcode-level `TraceResult`
- `debug_traceBlock`, `debug_traceBlockByNumber` — block-level tracing
- `debug_getBadBlocks` — retrieve rejected blocks

### Admin API (`admin_` namespace)

The `adminapi` subpackage exposes:

- `admin_nodeInfo` — node identity (`NodeInfoData` with enode URL, peer count, ports)
- `admin_peers` — connected peer list (`PeerInfoData`)
- `admin_addPeer` — dynamic peer addition
- `admin_removePeer` — peer removal
- `admin_startHTTP`, `admin_stopHTTP` — runtime HTTP control

### Batch Requests

`BatchHandler` processes JSON-RPC 2.0 batch requests (JSON arrays). `IsBatchRequest(body []byte) bool` detects batch format. `MarshalBatchResponse(responses []BatchResponse) ([]byte, error)` encodes the response array.

### WebSocket Subscriptions

`websocket_handler.go` implements the WebSocket upgrade and subscription dispatch. Real-time event channels: `newHeads` (new block headers), `logs` (filtered log events), `newPendingTransactions` (txpool insertions), `syncing` (sync state changes).

### Response Formatting

Exported formatting functions convert internal types to JSON-serializable RPC representations:

```go
var FormatBlock       = rpctypes.FormatBlock
var FormatHeader      = rpctypes.FormatHeader
var FormatTransaction = rpctypes.FormatTransaction
var FormatReceipt     = rpctypes.FormatReceipt
var FormatLog         = rpctypes.FormatLog
```

### Core RPC Types

```go
type RPCBlock         // block with transaction hashes
type RPCBlockWithTxs  // block with full transaction objects
type RPCTransaction   // transaction fields (hash, from, to, value, data, etc.)
type RPCReceipt       // transaction receipt (status, gas used, logs, bloom)
type RPCLog           // event log (address, topics, data, block/tx context)
type CallArgs         // eth_call / eth_estimateGas parameters
type FilterCriteria   // eth_getLogs filter (from/to block, addresses, topics)
type FeeHistoryResult // eth_feeHistory response
type AccessListResult // eth_createAccessList response
type AccountProof     // eth_getProof Merkle proof response
```

## Subpackages

| Subpackage | Description |
|------------|-------------|
| [`adminapi/`](./adminapi/) | `admin_` namespace: node info, peer management |
| [`backend/`](./backend/) | `Backend` interface used by all RPC handlers |
| [`batch/`](./batch/) | JSON-RPC 2.0 batch request parsing and dispatch |
| [`beaconapi/`](./beaconapi/) | 16 Beacon API endpoints for CL client integration |
| [`debugapi/`](./debugapi/) | `debug_` namespace: EVM tracing, bad blocks |
| [`ethapi/`](./ethapi/) | `eth_` namespace: blocks, txs, state, filters, AA |
| [`filter/`](./filter/) | Event filter and log indexing |
| [`gas/`](./gas/) | Gas price oracle and fee history |
| [`middleware/`](./middleware/) | Request/response middleware (logging, rate limiting) |
| [`netapi/`](./netapi/) | `net_` namespace: peer count, network ID, listening status |
| [`registry/`](./registry/) | Method registry for dynamic handler lookup |
| [`server/`](./server/) | HTTP/WebSocket server with `ServerConfig` |
| [`subscription/`](./subscription/) | Real-time subscription manager (newHeads, logs, etc.) |
| [`types/`](./types/) | Shared JSON-serializable types and formatting functions |

## Usage

```go
// Create the EthAPI with a backend.
api := rpc.NewEthAPI(backend)

// Create a batch handler for HTTP dispatch.
batchHandler := rpc.NewBatchHandler(api)

// Detect and dispatch batch requests.
if rpc.IsBatchRequest(body) {
    responses := batchHandler.HandleBatch(ctx, body)
    out, _ := rpc.MarshalBatchResponse(responses)
    w.Write(out)
}

// Admin API setup.
adminAPI := rpc.NewAdminDispatchAPI(adminBackend)
```

## Documentation References

- [Roadmap](../../docs/ROADMAP.md)
- [Ethereum JSON-RPC API](https://ethereum.github.io/execution-apis/api-documentation/)
- [Beacon API](https://ethereum.github.io/beacon-APIs/)
- EIP-4844: `eth_blobBaseFee`, blob sidecar retrieval
- EIP-7701: `eth_sendUserOperation` (native AA)
