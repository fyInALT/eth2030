# Package engine

Engine API server (V3-V7) for CL-EL communication in ETH2030.

## Overview

The `engine` package implements the Ethereum Engine API JSON-RPC interface that the Consensus Layer (CL) uses to drive the Execution Layer (EL). It covers all Engine API versions from V3 (Cancun/Deneb) through V7 (Amsterdam and beyond), including the ePBS builder extensions and FOCIL inclusion list methods.

The package exposes an `EngineAPI` struct that binds a `Backend` (the EL block processor) to an HTTP server with bearer-token JWT authentication. Incoming JSON-RPC 2.0 requests are dispatched to typed handlers for each method. All Engine API type definitions, conversion utilities, and sub-concerns live in dedicated sub-packages, while the top-level package re-exports them for backward compatibility.

The implementation supports the full ETH2030 fork progression — Cancun, Prague/Electra, Glamsterdam, and Amsterdam — selecting the correct payload version, fork validation, and blob bundle handling automatically based on the timestamp in the payload or forkchoice attributes.

## Table of Contents

- [Functionality](#functionality)
- [Subpackages](#subpackages)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### Core Engine API Server

`EngineAPI` is the central type. It wraps a `Backend` and serves requests over HTTP:

```
NewEngineAPI(backend Backend) *EngineAPI
```

Key methods on `EngineAPI`:

| Method | Description |
|--------|-------------|
| `Start(addr string) error` | Starts the HTTP server (blocking) |
| `Stop() error` | Gracefully shuts down the server |
| `SetAuthSecret(secret string)` | Configures bearer-token JWT authentication |
| `SetMaxRequestSize(bytes int64)` | Sets the 5 MiB default request size limit |
| `SetEthHandler(h http.Handler)` | Registers a handler for non-engine_ methods (eth_, net_, admin_) |
| `HandleRequest(data []byte) []byte` | Processes a raw JSON-RPC 2.0 request |
| `ExchangeCapabilities([]string) []string` | Returns the list of supported Engine API methods |

### Payload Processing (newPayload)

Three `newPayload` versions are dispatched based on the fork present in the timestamp:

- `NewPayloadV3` — Cancun/Deneb: validates EIP-4844 blob versioned hashes and EIP-4788 `parentBeaconBlockRoot`
- `NewPayloadV4` — Prague/Electra: adds EIP-7685 `executionRequests` (deposits, withdrawals, consolidations)
- `NewPayloadV5` — Amsterdam (Glamsterdam): adds `blockAccessList` (EIP-7928 BAL field)

All versions verify blob versioned hash ordering and count against the transactions in the payload before forwarding to the backend.

### Forkchoice Updates

- `ForkchoiceUpdatedV3` — validates `parentBeaconBlockRoot` in payload attributes, enforces timestamp progression
- `ForkchoiceUpdatedV4` — Amsterdam fork check on payload attributes

### Payload Building (getPayload)

- `GetPayloadV3` — returns `ExecutionPayloadV3` + `blockValue` + `BlobsBundleV1`
- `GetPayloadV4` — adds `executionRequests` to the response
- `GetPayloadV6` — returns `ExecutionPayloadV4` (with BAL) + execution requests for Amsterdam

### ePBS Builder API

The Engine API is extended with ePBS (EIP-7732) methods dispatched through the same JSON-RPC endpoint:

- `engine_submitBuilderBidV1` — builder submits a signed `BuilderBid`; stored in `BuilderRegistry`
- `engine_getBuilderBidsV1` — CL retrieves pending bids for a slot
- `engine_getPayloadHeaderV1` — returns a blinded payload header (commitment without body)
- `engine_submitBlindedBlockV1` — builder reveals the full payload matching a previous bid

### FOCIL Inclusion List API

- `engine_newInclusionListV1` — CL submits an inclusion list for EL-side storage
- `engine_getInclusionListV1` — EL returns stored ILs for a given slot

### Payload Bodies

- `engine_getPayloadBodiesByHashV2` — batch retrieval of payload bodies by block hash
- `engine_getPayloadBodiesByRangeV2` — batch retrieval of payload bodies by block number range

### Payload Types (V1-V7)

A full hierarchy of payload types tracks the fork progression:

| Type | Fork | New Fields |
|------|------|------------|
| `ExecutionPayloadV1` | Paris | Base fields |
| `ExecutionPayloadV2` | Shanghai | `Withdrawals` |
| `ExecutionPayloadV3` | Cancun | `BlobGasUsed`, `ExcessBlobGas` |
| `ExecutionPayloadV4` | Amsterdam | `BlockAccessList` (EIP-7928) |
| `ExecutionPayloadV5` | Amsterdam | Combined V3 + BAL |
| `ExecutionPayloadV7` | M+ | DALayer config, proof requirements |

### Blob Bundle Validation (blobsbundle)

The `blobsbundle` sub-package validates that blob data, KZG commitments, and KZG proofs in a `BlobsBundleV1` are internally consistent:

- `ValidateBundle(bundle, txs)` — checks commitment count, sizes, and versioned hash derivation
- `DeriveVersionedHashes(commitments)` — computes `kzg_to_versioned_hash` for each commitment
- `ValidateVersionedHashes(bundle, expected)` — cross-checks CL-provided hashes
- `PrepareSidecars(bundle, txs)` — assembles `BlobSidecar` objects for network propagation

### Distributed Block Builder (distbuilder)

The `distbuilder` sub-package coordinates a distributed builder network where multiple builders compete to construct execution payloads:

- `DistributedBuilder` — maintains a registry of up to `MaxBuilders` builders
- Builder registration, bid submission, and auction-based winner selection
- `BuilderConfig` — configures timeout, minimum bid, and auction duration

### Vickrey Auction (auction)

The `auction` sub-package implements a second-price sealed-bid (Vickrey) auction for builder selection:

- Winning builder pays the second-highest bid price
- Slashing integration for builders who fail to reveal the committed payload
- `BuilderAuction` tracks per-slot bids and determines winners

### Payload Conversion (convert)

Utilities for converting between payload versions and extracting data for block assembly:

- `PayloadToHeaderV1/V2/V3/V5` — extract a block header from a payload
- `HeaderToPayloadV2/V3` — construct a payload from a block header
- `DeterminePayloadVersion(timestamp, forkTimestamps)` — selects the correct version constant
- `ConvertV1ToV2`, `ConvertV2ToV3`, etc. — upgrade payload structs across fork boundaries
- `ValidatePayloadConsistency(payload)` — cross-field consistency checks

### Error Codes

Standard Engine API JSON-RPC error codes from `errors.go`:

| Constant | Code | Meaning |
|----------|------|---------|
| `UnknownPayloadCode` | -38001 | Payload ID not found |
| `InvalidForkchoiceStateCode` | -38002 | Invalid forkchoice state |
| `InvalidPayloadAttributeCode` | -38003 | Invalid payload attributes |
| `TooLargeRequestCode` | -38004 | Request body too large |
| `UnsupportedForkCode` | -38005 | Fork not supported by this method version |

## Subpackages

| Subpackage | Description |
|------------|-------------|
| [`api/`](./api/) | Extended Engine API types: V4 execution requests, uncoupled payload (EIP-7898), ePBS, Glamsterdam, V7 |
| [`auction/`](./auction/) | Vickrey second-price sealed-bid builder auction |
| [`backendapi/`](./backendapi/) | `Backend` interface definition consumed by `EngineAPI` |
| [`blobsbundle/`](./blobsbundle/) | BlobsBundleV1 validation, versioned hash derivation, sidecar preparation |
| [`blobval/`](./blobval/) | Blob validation utilities |
| [`blocks/`](./blocks/) | Block assembly helpers |
| [`builder/`](./builder/) | Builder registry, bid and envelope types, BLS key management |
| [`chunking/`](./chunking/) | Payload chunking for block-in-blobs (Hegotá) |
| [`convert/`](./convert/) | Payload version conversion and header extraction |
| [`distbuilder/`](./distbuilder/) | Distributed block builder network coordination |
| [`errors/`](./errors/) | Engine API error code definitions |
| [`forkchoice/`](./forkchoice/) | Forkchoice state management |
| [`payload/`](./payload/) | Canonical payload type definitions (V1-V7) and status constants |
| [`requests/`](./requests/) | EIP-7685 execution request parsing |
| [`util/`](./util/) | Shared utilities (hex encoding, hash helpers) |
| [`vhash/`](./vhash/) | Versioned hash computation (EIP-4844) |

## Usage

```go
import "github.com/eth2030/eth2030/engine"

// Create and start the Engine API server.
api := engine.NewEngineAPI(myBackend)
api.SetAuthSecret(jwtSecret)

// Register a handler for eth_* methods on the same port.
api.SetEthHandler(ethRPCHandler)

// Start serving (blocks until Stop is called).
go func() {
    if err := api.Start("0.0.0.0:8551"); err != nil {
        log.Fatal(err)
    }
}()

// The CL connects and calls engine_exchangeCapabilities first.
// Supported methods include all versions V3-V7 plus ePBS and FOCIL extensions.

// Validate a blobs bundle before returning it to the CL.
if err := engine.ValidateBundle(bundle, txs); err != nil {
    return fmt.Errorf("invalid blobs bundle: %w", err)
}

// Convert a payload to a block header for local chain insertion.
header, err := engine.PayloadToHeaderV3(&payload)
if err != nil {
    return err
}
```

## Documentation References

- [Design Doc](../../docs/DESIGN.md)
- [Roadmap](../../docs/ROADMAP.md)
- [EIP Spec Implementation](../../docs/EIP_SPEC_IMPL.md)
- [Engine API Specification](https://github.com/ethereum/execution-apis/blob/main/src/engine/README.md)
- [EIP-7898: Uncoupled Execution Payload](https://eips.ethereum.org/EIPS/eip-7898)
- [EIP-7685: General Purpose EL Requests](https://eips.ethereum.org/EIPS/eip-7685)
- [EIP-4844: Blob Transactions](https://eips.ethereum.org/EIPS/eip-4844)
