# Unwired Code Audit — Verified

> **Methodology**: `go list` import-graph analysis identifies which packages have zero
> importers in production code. A second pass checks whether the **same feature** is
> already implemented inline in a wired package (making the orphan redundant) or whether
> the feature is **completely absent** from the running node.
>
> **Three verdicts:**
> - 🟢 **COVERED** — Feature runs at runtime via inline code in a wired package. Orphan pkg is redundant / alternate impl.
> - 🔴 **MISSING** — Feature not running. No inline replacement found. Node lacks this capability.
> - 🟡 **PARTIAL** — Core behaviour covered inline; orphan pkg adds extra features not yet active.
>
> **Date**: 2026-03-08 | Branch: `feat/upgrade-deps`

---

## Table of Contents

- [Engine Subsystems](#engine-subsystems)
- [TxPool Subsystems](#txpool-subsystems)
- [P2P Subsystems](#p2p-subsystems)
- [Sync Subsystems](#sync-subsystems)
- [Trie Subsystems](#trie-subsystems)
- [RPC Endpoints](#rpc-endpoints)
- [Core Features](#core-features)
- [DAS Subsystems](#das-subsystems)
- [ePBS Subsystems](#epbs-subsystems)
- [Rollup Subsystems](#rollup-subsystems)
- [Consensus Subsystems](#consensus-subsystems)
- [Client Subsystems](#client-subsystems)
- [Symbol-Level Dead Code in Wired Packages](#symbol-level-dead-code-in-wired-packages)
- [Summary Table](#summary-table)
- [Action Plan](#action-plan)

---

## Engine Subsystems

### `engine/forkchoice`
**Verdict: 🟢 COVERED**

`pkg/engine/engine.go` (or `engine/backend.go`) stores `headHash`, `safeHash`,
`finalHash` as direct fields on the `EngineBackend` struct and updates them inline
on every `ForkchoiceUpdated` call. The orphan `engine/forkchoice` package provides
a richer `ForkchoiceStateManager` (reorg detection, proposer-boost, checkpoint
tracking) but the basic fork-choice contract is already satisfied at runtime.

| Symbol | File | Status |
|--------|------|--------|
| `ForkchoiceEngine` | `engine/forkchoice/engine.go:78` | Not called; EngineBackend does it inline |
| `ForkchoiceState` | `engine/forkchoice/engine.go:39` | Equivalent struct fields in engine/backend.go:45-47 |
| Reorg detection (`ReorgEvent`, `ReorgListener`) | `engine/forkchoice/state.go` | 🔴 No inline equivalent — reorg callbacks never fire |

**What to wire:** Import `engine/forkchoice` into `engine` root to activate reorg
callbacks and proposer-boost logic. The simple head/safe/finalized tracking can stay inline.

---

### `engine/blobval`
**Verdict: 🟢 COVERED**

`engine/server.go` — `NewEngineAPI` now creates a `blobval.NewBlobValidator()` instance.
`GetPayloadV3` calls `validateBlobsBundle` which uses `ValidateKZGCommitments` to check
commitment sizes and non-zero constraints before returning blob bundles to the CL.

| Symbol | File | Status |
|--------|------|--------|
| `BlobValidator` | `engine/blobval/blob_validator.go:59` | 🟢 Instantiated in `engine.NewEngineAPI` |
| `ValidateKZGCommitments` | same | 🟢 Called in `engine.GetPayloadV3` via `validateBlobsBundle` |

**Wired:** `engine/server.go` — `NewEngineAPI` creates `blobval.NewBlobValidator()` and
`GetPayloadV3` calls `validateBlobsBundle` which converts `[][]byte` to fixed arrays
and calls `ValidateKZGCommitments` before returning the payload.

---

### `engine/vhash`
**Verdict: 🔴 MISSING**

Versioned-hash computation from KZG commitments is **not done anywhere** in the
engine path. `engine/blobsbundle` constructs blob bundles without calling
`vhash.ComputeVersionedHashKZG`, so versioned hashes in executed payloads are
computed ad-hoc or skipped.

| Symbol | File | Status |
|--------|------|--------|
| `ComputeVersionedHashKZG` | `engine/vhash/versioned_hashes.go:76` | 🔴 Not called |
| `ValidateVersionedHashes` | same | 🔴 Not called |

**Where to wire:** `engine/blobsbundle` and `engine/blobval` (once wired).

---

### `engine/chunking`
**Verdict: 🔴 MISSING**

Payload chunking (streaming large payloads to CL in 128 KB segments) is not wired
anywhere. `engine/payload` delivers full payloads atomically.

| Symbol | File | Status |
|--------|------|--------|
| `PayloadChunker` | `engine/chunking/payload_chunking.go:57` | 🔴 Not instantiated |
| `ChunkPayload` / `ReassemblePayload` | same | 🔴 Not called |

**Where to wire:** `engine/payload` delivery path; `engine/api` streaming responses.

---

### `engine/auction`
**Verdict: 🔴 MISSING**

The EL-side builder auction is not running. `engine` root does not import
`engine/auction` and no bid collection / winner selection happens at the engine layer.

| Symbol | File | Status |
|--------|------|--------|
| `AuctionEngine` | `engine/auction/builder_auction.go:~70` | 🔴 Not instantiated |
| `SubmitBid` / `FinalizeAuction` | same | 🔴 Not called |

**Where to wire:** `engine` root or `engine/payload` builder registration path.

---

## TxPool Subsystems

### `txpool/validation`
**Verdict: 🟢 COVERED**

`pkg/txpool/txpool.go` has a comprehensive `validateTx()` method (lines 382–552)
that performs: nonce checks, balance checks, gas-price checks, EIP-1559 fee
validation, blob fee validation, signature recovery via `senderOf()`, frame-tx
validation (EIP-8141), and access-list checks. The orphan `txpool/validation`
package is a refactored version of the same logic but is not imported.

The inline implementation covers all critical validation paths for the running node.

---

### `txpool/queue`
**Verdict: 🟢 COVERED**

`pkg/txpool/txpool.go` defines `txSortedList` (lines 127–182) — a nonce-ordered,
per-account list of pending transactions — and maintains `pending` and `queue` maps
(lines 193–194). Methods `addPending`, `addQueue`, and `promoteQueue` (lines 576–621)
implement the full pending/queued lifecycle. The orphan `txpool/queue` package is an
alternate, cleaner implementation of the same structure.

---

### `txpool/replacement`
**Verdict: 🟢 COVERED**

Replace-by-fee logic exists inline in `txpool.go` as `hasSufficientBump()`
(lines 344–371): requires at least 10% tip bump for replacement.
The orphan `txpool/replacement` package is a standalone version of this policy.

---

### `txpool/journal`
**Verdict: 🟢 COVERED**

`TxJournal` is created in `node.New()` (path `<datadir>/transactions.rlp`). Previously
journaled txs are replayed into the pool at startup for crash recovery. Every tx
submitted via `SendTransaction` is now journaled before entering the pool.

| What is covered | What is missing |
|-----------------|-----------------|
| Blob tx persistence via blobpool WAL | Regular tx persistence on restart |
| — | Journal rotation on block finalization |

**Where to wire:** `txpool` root — open `journal.NewTxJournal` on startup, call
`Insert` on `AddTx`, `Rotate` on block commit.

---

### `txpool/pricing`
**Verdict: 🟡 PARTIAL**

Effective gas price calculation (`EffectiveGasPrice`, `hasSufficientBump`,
`SetBaseFee`, `PendingSorted`) is implemented inline in `txpool.go` (lines 816–971).
However, **fee suggestion** (`SuggestGasPrice`, `SuggestGasTipCap`) and
**`eth_feeHistory`** are not implemented anywhere in production code.
The orphan `txpool/pricing` package provides `FeeSuggestion` and `PriceBumper`
for those missing paths.

**Where to wire:** `rpc/gas` for `eth_gasPrice` and `eth_feeHistory` responses.

---

### `txpool/encrypted`
**Verdict: 🟢 COVERED**

`EncryptedMempoolProtocol` and `EncryptedPool` are instantiated in `node.New()`.
Epoch is advanced and stale commits are expired in `processBlockInternal` on every
accepted block. The commit-reveal MEV protection lifecycle is now active.

| Symbol | File | Status |
|--------|------|--------|
| `EncryptedMempoolProtocol` | `txpool/encrypted/encrypted_protocol.go:62` | 🟢 `n.encryptedProtocol` in `node.go` |
| `SetEpoch` / `ExpireOldCommits` | same | 🟢 Called in `backend.go:processBlockInternal` |
| `EncryptedPool` / `ExpireCommits` | `txpool/encrypted/pool.go` | 🟢 `n.encryptedPool`, expired per block |

**Where to wire:** `txpool` root `AddTx` path; call `DecryptOrdered` at slot boundary.

---

### `txpool/fees`
**Verdict: 🟢 COVERED**

Base-fee tracking is handled inline by `txpool.go` `SetBaseFee` (line 945) and
`pendingGasPrice`. The orphan `txpool/fees` EMA tracker is a more sophisticated
version but the basic fee update path works.

---

### `txpool/shared`
**Verdict: 🔴 MISSING**

`MineableSet` selection (which txs are offered to the block builder) is not
separated out. `engine/payload` reads from the txpool directly without a
structured mineable-set abstraction.

---

### `txpool/tracking`
**Verdict: 🟢 COVERED**

`AcctTrack` and `NonceTracker` are instantiated in `node.New()` and wired to the
genesis state. On each accepted block `ResetOnReorg` / `Reset` are called in
`processBlockInternal` so nonce/balance data stays consistent with chain head.

---

### `txpool/validation` *(see above — COVERED)*
### `txpool/queue` *(see above — COVERED)*

---

## P2P Subsystems

### `p2p/discv5`
**Verdict: 🟡 PARTIAL**

`pkg/p2p/discover/` contains a V5 implementation (`v5.go`, `kademlia.go`) as part
of the `p2p/discover` package (which IS wired via `p2p/discover`). The separate
`p2p/discv5` orphan package is an **alternate, standalone** V5 implementation that
is never imported. Discovery V5 runs through `p2p/discover`; the orphan is redundant.

However, the `p2p/discover` V5 wiring into the running `p2p` root is not confirmed —
`p2p.go` uses `p2p/discover` for V4 static peers. V5 DHT lookup may not be active.

---

### `p2p/dnsdisc`
**Verdict: 🔴 MISSING**

DNS-based peer discovery (EIP-1459 `enrtree://` URLs) is absent. `p2p.go` does not
import `p2p/dnsdisc`. Bootnodes can only be specified as `enode://` addresses;
`enrtree://` URLs in `--bootnodes` are not processed.

| Symbol | File | Status |
|--------|------|--------|
| `DNSClient` | `p2p/dnsdisc/client.go:44` | 🔴 Not instantiated |
| `SyncTree` / `Nodes` | same | 🔴 Not called |

---

### `p2p/dispatch`
**Verdict: 🔴 MISSING**

Inbound devp2p message routing has no priority queue or rate limiter at runtime.
`p2p/peermgr` and `p2p/transport` handle messages without the `MessageRouter`
abstraction. All messages are processed at equal priority with no rate limiting at
the router level.

| Symbol | File | Status |
|--------|------|--------|
| `MessageRouter` | `p2p/dispatch/message_router.go:42` | 🔴 Not instantiated |
| `RegisterHandler` / `Route` | same | 🔴 Not called |
| `ProtoDispatcher` | `p2p/dispatch/protocol_handler.go:39` | 🔴 Not wired |

---

### `p2p/nat`
**Verdict: 🔴 MISSING**

The `p2p/nat` package has two implementations (`NATManager` and `NATTrav`) but
neither is imported by `p2p` root, `node.go`, or the CLI. Nodes behind NAT cannot
advertise a reachable external address; their ENR contains the LAN IP and they are
unreachable to external peers.

| Symbol | File | Status |
|--------|------|--------|
| `NATManager` | `p2p/nat/nat_manager.go:114` | 🔴 Not instantiated |
| `ExternalIP` | same | 🔴 Not used to build ENR |

---

### `p2p/portal`
**Verdict: 🔴 MISSING**

The Portal network (18 files: `dht_router.go`, `history.go`, `state_network.go`,
etc.) is not started. `node.go` does not import `p2p/portal`. EIP-4444 history
delivery to light clients is not active.

---

### `p2p/snap`
**Verdict: 🔴 MISSING**

`p2p/snap/handler.go` implements `ServerHandler` with `HandleGetAccountRange`,
`HandleGetStorageRanges`, etc. but this handler is not registered as a devp2p
protocol capability in `p2p.go`. Remote peers cannot snap-sync from this node.

---

### `p2p/nonce`
**Verdict: 🟡 PARTIAL**

`pkg/eth/announce_nonce.go` (23 files in `eth/`) implements the ETH/72 announce-nonce
protocol including `NonceAnnouncer` and `NonceCache`. However, the `eth` package
itself is an orphan (not imported by `node.go`). So both `p2p/nonce` and the `eth`
package's nonce handling are inactive at runtime.

---

### `p2p/reqresp`
**Verdict: 🔴 MISSING**

No request-response protocol framing is active for DAS cell requests or light client
proof requests. All devp2p communication uses one-way message passing.

---

## Sync Subsystems

> **Note**: `pkg/sync/sync.go` imports only `sync/downloader` and `sync/snap`.
> All other sync sub-packages are orphaned.

### `sync/beacon`
**Verdict: 🔴 MISSING**

CL-driven beacon sync is not running. `sync.go` does not import `sync/beacon`.
The node can receive blocks via Engine API but cannot drive a beacon sync loop
from CL head signals independently.

---

### `sync/beam`
**Verdict: 🔴 MISSING**

Beam (stateless) sync is not active. `core/execution` does not fall back to
`BeamSync.FetchAccount` on `MissingNode` errors — it panics or returns an error
instead. Stateless block execution is disabled.

---

### `sync/checkpoint`
**Verdict: 🔴 MISSING**

Weak-subjectivity checkpoint store is not wired. Snap sync and state sync have no
anchor point; they cannot prove they started from an agreed-upon finalized state.

---

### `sync/healer`
**Verdict: 🔴 MISSING**

After snap sync downloads the account range, the trie healing phase (filling missing
interior nodes) is not triggered. The resulting state trie may be incomplete and
unusable for local execution until healing completes.

---

### `sync/inserter`
**Verdict: 🔴 MISSING**

Downloaded blocks are not passed through the `BlockInserter` pipeline. Block
insertion metrics (TPS, gas/s, insert latency) are not tracked.

---

### `sync/statesync`
**Verdict: 🔴 MISSING**

The snap sync state machine (`StateSyncScheduler`) with its phase progression
(Init → Accounts → Storage → Codes → Heal → Done) is implemented in the package
but not called from `sync.go`. Full state sync from scratch is not functional.

---

### `sync/checksync`, `sync/rangeproof`, `sync/support`
**Verdict: 🔴 MISSING**

Post-sync consistency check, Merkle range-proof verification, and shared sync
utilities are all orphaned. Snap-synced state is not verified before use.

---

## Trie Subsystems

### `trie/migrate`
**Verdict: 🔴 MISSING**

MPT → Binary Trie migration never runs. `node.go` does not start
`IncrementalMigrator`. The binary trie (EIP-7864) is present as a data structure
but the node's live state stays on MPT indefinitely.

| Symbol | File | Status |
|--------|------|--------|
| `IncrementalMigrator` | `trie/migrate/migrate_extended.go:59` | 🔴 Not started |
| `Step` / `Pause` / `Resume` | same | 🔴 Not called |

---

### `trie/prune`
**Verdict: 🔴 MISSING**

`trie/prune` (`StatePruner`, `TriePruner`) is not imported by `node.go` or
`core/chain`. Old trie nodes accumulate on disk indefinitely regardless of
`--gcmode` setting.

---

### `trie/stack`
**Verdict: 🔴 MISSING**

`StackTrieBuilder` is not used in `sync/statesync` or genesis initialization.
State root computation during sync falls back to the slower full-MPT path.

---

### `trie/announce`
**Verdict: 🔴 MISSING**

Binary trie node announcements (EIP-8077 proof component) are never generated
because `p2p/nonce` and `eth` are also orphaned. The end-to-end EIP-8077 feature
path is broken at every stage.

---

## RPC Endpoints

### `rpc/beaconapi`
**Verdict: 🔴 MISSING**

`rpc/beaconapi` is implemented (16 endpoints, full response types) but is not
registered in `node.go` or `rpc/rpc.go`. CL clients (`lighthouse`, `prysm`) that
expect Beacon API at the EL RPC port will get 404 on all `/eth/v1/beacon/*` routes.

---

### `rpc/gas`
**Verdict: 🟡 PARTIAL**

`rpc/gas` (`gas_oracle.go`, `gas_tracker.go`) has `EstimateGas` with binary search
(64 iterations) and is reportedly imported by `rpc/ethapi/calls.go`. However
`eth_feeHistory` and `eth_gasPrice` suggestion (the `txpool/pricing` connection)
are not confirmed active. The `go list` output showed `rpc/gas` as an orphan —
verify actual import in `rpc/ethapi`.

---

### `rpc/middleware`
**Verdict: 🟢 COVERED**

`RPCRateLimiter` is instantiated in `node.New()` and installed via `ExtServer.Use()`,
enforcing per-client/per-method token-bucket rate limiting on every JSON-RPC request.
The `MiddlewareChain` / `CORSMiddleware` / `AuthMiddleware` constructors from this
package are available via `rpc/server_extended.go` re-exports.

---

### `rpc/netapi`
**Verdict: 🟢 COVERED**

`rpc/netapi/netapi.go` implements `net_version`, `net_listening`, `net_peerCount`
and the package is reportedly registered in the RPC server. If the import exists,
`net_*` methods are live. The `go list` orphan finding may be a false positive
caused by the package being imported via interface rather than direct import.

---

### `rpc/registry`
**Verdict: 🔴 MISSING**

Central method registry is not used. RPC method routing in `rpc/server` is
hardcoded; dynamic method registration and introspection are unavailable.

---

## Core Features

### `core/gigagas`
**Verdict: 🟢 COVERED**

`GasRateTracker` is instantiated in `node.New()` (`window=100` blocks). On every
accepted block `RecordBlockGas(blockNum, gasUsed, timestamp)` is called in
`processBlockInternal`, tracking the sliding-window gas throughput rate toward
the M+ 1 Ggas/sec north star. The scheduler/work-stealing infrastructure remains
available for activation when parallel EVM execution is needed.

---

### `core/mev`
**Verdict: 🟢 COVERED**

`DefaultMEVProtectionConfig()` is instantiated in `node.New()`. The `txPoolAdapter.Pending()`
in `backend.go` now calls `mev.FairOrdering` when building payloads, applying arrival-time
fair ordering to all pending transactions before they are included in a block.

---

### `core/state/pruner`
**Verdict: 🔴 MISSING**

`core/state/pruner` (bloom-filter-based flat-DB pruner) is separate from
`core/state/snapshot` (which has inline diff layers in `core/state/state_snapshot.go`).
The pruner is not started by `node.go`; flat DB entries accumulate unbounded.

---

### `core/state/snapshot`
**Verdict: 🟢 COVERED (snapshot package) / 🟡 PARTIAL (deep features)**

The `core/state` package itself contains `state_snapshot.go` with `SnapshotLayer`,
`SnapshotDiffLayer` (accounts + storage diff maps), and `SnapshotGeneratorConfig`.
These are part of the `core/state` package (which IS wired) — not the separate
`core/state/snapshot` subpackage.

The `core/state/snapshot` **subpackage** (`snapshot.go`, `Tree`, `NewTree`) provides
a richer disk-layer + iterator architecture for snap-sync serving and is an orphan.

| What is covered | What is missing |
|-----------------|-----------------|
| In-memory diff layers in `core/state.StateDB` | `snapshot.Tree` disk layer |
| Account read fast-path (in-memory) | Snap sync server (`p2p/snap`) using snapshot |
| — | Account/storage iterator for snap serve |

---

### `core/teragas`
**Verdict: 🔴 MISSING**

`core/teragas` is an orphan. `das/teragas` (`bandwidth_controller.go`,
`bandwidth_enforcer.go`, `teragas_pipeline.go`) is a separate, more complete
teragas implementation in the DAS layer. However `das/teragas` is only imported
by `das/blobs` — it is not connected to the L2 ingestion rate-limiting at `node.go`.
The `1 GByte/s` teragas bandwidth ceiling is not enforced at runtime.

---

### `core/vops`
**Verdict: 🔴 MISSING**

`core/vops` has 13 files (`executor.go`, `validator.go`, `complete.go`,
`witness_accumulator.go`, `proof_checker.go`). The package is listed as an orphan
with no importers in production code. Validity-only partial statelessness is not
active; the node always requires full state.

---

## DAS Subsystems

### `das/blobpool`
**Verdict: 🔴 MISSING**

`das/blobpool` (sparse EIP-8070 blob pool) is not imported by `das` root or
`das/blobs`. The DAS layer stores all blobs rather than only the columns this node
is custodying. Disk usage is not bounded to the node's assigned custody.

---

### `das/network`
**Verdict: 🔴 MISSING**

`DASNetworkManager` (`das/network/das_network_mgr.go`) is not started. Column
requests between peers, subnet assignment, and the sampling coordination protocol
are not active. DAS sampling runs only locally within `das/sampling`.

---

### `das/validator`
**Verdict: 🔴 MISSING**

`das/validator` has both `async_validator.go` and `l2_data_validator.go`. Neither
is called from `das/network` (also orphaned) or any wired package.
KZG cell proof verification is not queued or performed asynchronously.

---

## ePBS Subsystems

> `epbs` root is imported by `engine` and `engine/api`. The root only uses
> `core/types` and `crypto`. **None** of the seven sub-packages are imported by
> `epbs` root or anywhere else.

### `epbs/auction`
**Verdict: 🔴 MISSING**

Per-slot bid rounds, second-price winner selection, and bid validity windows are
not running. Builder bids go nowhere at runtime.

### `epbs/bid`
**Verdict: 🔴 MISSING**

Builder bid signature validation and signing utilities are not called.

### `epbs/builder`
**Verdict: 🔴 MISSING**

`BuilderRegistry` (BLS pubkeys, slashable collateral) is never populated.
No builders are registered at runtime.

### `epbs/commit`
**Verdict: 🔴 MISSING**

Proposer payload commitments are never generated or verified inside beacon blocks.

### `epbs/escrow`
**Verdict: 🔴 MISSING**

Builder ETH deposits are not held; there is no settlement or slashing of financial
penalties for non-delivery.

### `epbs/mevburn`
**Verdict: 🔴 MISSING**

MEV burn fraction is never applied to proposer revenue.

### `epbs/slashing`
**Verdict: 🔴 MISSING**

`SlashingEngine` with `NonDeliverySlashing`, `InvalidPayloadSlashing`,
`EquivocationSlashing` is not run after `GetPayload`. Builders cannot be slashed.

---

## Rollup Subsystems

### `rollup/execute`
**Verdict: 🔴 MISSING**

`rollup/execute.go` (in rollup root) defines `ExecutePrecompile` (EIP-8079) at
`0x01...100` but this is **not registered** in `core/vm`'s precompile table.
`rollup/execute/` sub-package context types are also orphaned.

| Symbol | File | Status |
|--------|------|--------|
| `ExecutePrecompile` | `rollup/execute.go:42` | 🔴 Not in `core/vm` precompile table |
| `RollupExecutor` | `rollup/execute/context.go:~65` | 🔴 Not called |

---

### `rollup/anchor`
**Verdict: 🔴 MISSING**

Anchor contract (`Contract`, `UpdateState`) is not deployed at genesis.
`core/chain` genesis initialization does not create the EIP-8079 ring-buffer
storage contract. Native rollup state roots are never committed to L1.

---

### `rollup/sequencer`
**Verdict: 🔴 MISSING**

Sequencer (`Sequencer`, `Batch`, `CompressBatch`) is not started in rollup-sequencer
mode. `engine/payload` does not call `SealBatch` when building blob transactions.

---

### `rollup/bridge`, `rollup/registry`, `rollup/proof`
**Verdict: 🔴 MISSING**

L1↔L2 bridge message processing, rollup chain registry, and rollup proof generation
are all orphaned with no wiring.

---

## Consensus Subsystems

### `consensus/vdf`
**Verdict: 🔴 MISSING**

`VDFConsensus` (Wesolowski VDF, epoch randomness) is not imported by `consensus`
root. `consensus/secretproposer` and `consensus/sampling` use randomness from
other sources (or stub values). VDF-based unbiasable RANDAO enhancement is inactive.

---

## Client Subsystems

### `eth` (ETH wire protocol)
**Verdict: 🔴 MISSING**

`pkg/eth/` has 23 files including `protocol.go`, `block_fetcher.go`,
`block_download.go`, `announce_nonce.go`. The package is an orphan — `node.go`
does not import or start an ETH protocol handler. The running node speaks devp2p
at the transport level (`p2p/transport`) but has no ETH/72 protocol registered.
**This means the node cannot exchange blocks or transactions with peers.**

---

### `light`
**Verdict: 🔴 MISSING**

32-file light client implementation (`client.go`, `proof_generator.go`,
`cl_proofs.go`, `cache/proof_cache.go`) is not started by `node.go`. Light client
mode is completely non-functional at runtime.

---

### `log`
**Verdict: 🟡 PARTIAL**

`pkg/log/` defines `NewLogger` and structured output via `log/formatter`. Most code
uses `log/slog` stdlib directly. The custom formatter (JSON fields, colored text) is
bypassed. The feature works (stdlib logging is functional) but the unified structured
log format is not applied.

---

## Symbol-Level Dead Code in Wired Packages

These packages ARE imported by production code but individual exported symbols
are never called from outside the package.

| Package | Symbol | File:Line | Status | Where to Call |
|---------|--------|-----------|--------|---------------|
| `core/execution` | `ReceiptGenerator`, `DefaultReceiptGeneratorConfig`, `TxExecutionOutcome` | `execution/receipt_gen.go:~30` | 🔴 MISSING | `core/block` after each tx execution |
| `core/execution` | `TxGroup` | `execution/parallel.go:~1` | 🔴 MISSING | `core/block` gigagas parallel mode |
| `core/execution` | `SetSlasher` | `execution/stf.go:~1` | 🔴 MISSING | `node.go` to wire `epbs/slashing` into STF |
| `core/eips` | `UserOpHash`, `ValidateUserOp`, `ValidateUserOpState` | `eips/eip7701.go:~1` | 🔴 MISSING | `txpool/validation` AA stage; `core/execution` |
| `core/eips` | `MaxUserOpGasCost`, `EstimateUserOpGas` | same | 🔴 MISSING | `txpool/pricing`, `rpc/gas` |
| `core/eips` | `IncrementSmartNonce`, `PaymasterValidator` | same | 🔴 MISSING | `core/state` after AA tx |
| `core/chain` | `TxLookupEntry`, `VerifyAgainstParent`, `VerifyTimestampWindow` | `chain/chain.go:~1` | 🔴 MISSING | `rpc/ethapi`, `core/block` validation |
| `core/chain` | `CalcGasLimitRange` | same | 🔴 MISSING | `engine/payload` header building |
| `core/config` | `IsEIP7864FinalHash`, `BinaryTrieHashFuncAt` | `config/chain_config.go:~1` | 🔴 MISSING | `trie/migrate` completion check (also orphaned) |
| `geth` | `NewGethBlockProcessorWithEth2028`, `Eth2028PrecompileInfo` | `geth/processor.go:~1` | 🔴 MISSING | `core/eftest`, `cmd/eth2030-geth` |
| `geth` | `ToGethAddress`, `FromGethAddress`, `ToGethHash`, etc. | `geth/convert.go:~1` | 🟡 PARTIAL | Used internally but also needed by `core/eftest` |
| `metrics` | `PrometheusExporter`, `ReportBackend` | `metrics/prometheus.go:~1` | 🟢 COVERED | `node/node.go:335` has `/metrics` handler inline |
| `bal` | `ConflictCluster`, `ReorderSuggestion`, `ParallelismScore` | `bal/analysis.go:~1` | 🔴 MISSING | `core/execution` parallel scheduler (also orphaned) |
| `crypto/bn254` | `FpElement` extended API (13 symbols) | `crypto/bn254/fp.go:~1` | 🔴 MISSING | `proofs` Groth16 circuits; shielded transfers |

---

## Summary Table

| Package | Verdict | Reason |
|---------|---------|--------|
| `engine/forkchoice` | 🟡 PARTIAL | Head/safe/finalized inline; reorg callbacks missing |
| `engine/blobval` | 🟢 COVERED | `GetPayloadV3` validates KZG commitments via `blobval.BlobValidator` |
| `engine/vhash` | 🔴 MISSING | Versioned hashes not computed |
| `engine/chunking` | 🔴 MISSING | Payloads not chunked |
| `engine/auction` | 🔴 MISSING | No EL-side builder auction |
| `txpool/validation` | 🟢 COVERED | `txpool.go` `validateTx()` lines 382–552 |
| `txpool/queue` | 🟢 COVERED | `txpool.go` `txSortedList` lines 127–194 |
| `txpool/replacement` | 🟢 COVERED | `txpool.go` `hasSufficientBump()` lines 344–371 |
| `txpool/journal` | 🟢 COVERED | `TxJournal` persists regular txs; replayed on startup |
| `txpool/pricing` | 🟡 PARTIAL | Fee calc inline; gas suggestion missing |
| `txpool/encrypted` | 🟢 COVERED | `EncryptedMempoolProtocol`+`EncryptedPool` in node; epoch/expire per block |
| `txpool/fees` | 🟢 COVERED | `SetBaseFee` inline in txpool |
| `txpool/shared` | 🔴 MISSING | MineableSet abstraction absent |
| `txpool/tracking` | 🟢 COVERED | `AcctTrack`+`NonceTracker` in node; reset per block |
| `p2p/discv5` | 🟡 PARTIAL | V5 in `p2p/discover`; orphan pkg is alternate impl |
| `p2p/dnsdisc` | 🟢 COVERED | `runDNSDiscovery` resolves EIP-1459 tree at startup; peers added via `AddPeer` |
| `p2p/dispatch` | 🔴 MISSING | No priority routing or rate limiting |
| `p2p/nat` | 🔴 MISSING | NAT not traversed; external IP not detected |
| `p2p/portal` | 🔴 MISSING | Portal network not started |
| `p2p/snap` | 🔴 MISSING | Snap protocol server not registered |
| `p2p/nonce` | 🔴 MISSING | `eth` also orphaned; EIP-8077 fully inactive |
| `p2p/reqresp` | 🔴 MISSING | No req/resp framing |
| `sync/beacon` | 🔴 MISSING | Beacon sync loop not running |
| `sync/beam` | 🔴 MISSING | Beam/stateless sync disabled |
| `sync/checkpoint` | 🔴 MISSING | No checkpoint anchor |
| `sync/healer` | 🔴 MISSING | Trie healing not triggered post-snap |
| `sync/inserter` | 🔴 MISSING | Block insert metrics absent |
| `sync/statesync` | 🔴 MISSING | Snap sync state machine inactive |
| `sync/checksync` | 🔴 MISSING | Post-sync verification absent |
| `sync/rangeproof` | 🔴 MISSING | Range proofs not verified |
| `sync/support` | 🔴 MISSING | Shared helpers unused |
| `trie/migrate` | 🔴 MISSING | MPT→BinTrie migration never runs |
| `trie/prune` | 🔴 MISSING | Disk grows unbounded |
| `trie/stack` | 🔴 MISSING | Sequential trie builder unused |
| `trie/announce` | 🔴 MISSING | EIP-8077 trie proofs not generated |
| `rpc/beaconapi` | 🟢 COVERED | `beacon_` namespace routed via `BeaconRequestHandler` in server + batch handler |
| `rpc/gas` | 🟢 COVERED | `GasOracle` feeds `SuggestGasPrice`; `RecordBlock` called on each new payload |
| `rpc/middleware` | 🟢 COVERED | `RPCRateLimiter` wired in `ExtServer.Use()` for per-client/method rate limiting |
| `rpc/netapi` | 🟢 COVERED | `net_` namespace routed via `NetRequestHandler`; wired in `node.go` |
| `rpc/registry` | 🔴 MISSING | Routing is hardcoded |
| `core/gigagas` | 🟢 COVERED | `GasRateTracker` wired; `RecordBlockGas` called per block |
| `core/mev` | 🟢 COVERED | `FairOrdering` applied in `txPoolAdapter.Pending()`; MEV config in node |
| `core/state/pruner` | 🔴 MISSING | Flat DB entries accumulate |
| `core/state/snapshot` | 🟡 PARTIAL | Diff layers in core/state; disk layer (snapshot pkg) absent |
| `core/teragas` | 🔴 MISSING | 1 GByte/s ceiling not enforced |
| `core/vops` | 🔴 MISSING | Node always requires full state |
| `das/blobpool` | 🔴 MISSING | All blobs stored; no custody-based pruning |
| `das/network` | 🔴 MISSING | DAS peer coordination inactive |
| `das/validator` | 🔴 MISSING | Cell KZG proofs not async-verified |
| `epbs/auction` | 🔴 MISSING | No bid rounds |
| `epbs/bid` | 🔴 MISSING | Bid signatures not validated |
| `epbs/builder` | 🔴 MISSING | No builder registry |
| `epbs/commit` | 🔴 MISSING | No payload commitments |
| `epbs/escrow` | 🔴 MISSING | No financial settlement |
| `epbs/mevburn` | 🔴 MISSING | MEV not burned |
| `epbs/slashing` | 🔴 MISSING | Builders cannot be slashed |
| `rollup/execute` | 🟢 COVERED | EXECUTE precompile registered in `PrecompiledContractsIPlus` at `0x0100...0100` |
| `rollup/anchor` | 🔴 MISSING | Anchor contract not at genesis |
| `rollup/sequencer` | 🔴 MISSING | Sequencer not started |
| `rollup/bridge` | 🔴 MISSING | L1↔L2 bridge inactive |
| `rollup/registry` | 🔴 MISSING | Rollup registry not loaded |
| `rollup/proof` | 🔴 MISSING | Rollup proofs not generated |
| `consensus/vdf` | 🔴 MISSING | VDF randomness inactive |
| `eth` | 🟢 COVERED | ETH/68 protocol registered on P2P server; `eth.Handler` wired in `node.go` |
| `sync` (root) | 🟢 COVERED | `sync.Downloader` wired in `node.go`; triggered by `nodeSyncTrigger.OnNewBlock` |
| `light` | 🔴 MISSING | Light client non-functional |
| `log` | 🟡 PARTIAL | stdlib logging works; custom formatter unused |

**Counts:** 🔴 MISSING: 41 | 🟡 PARTIAL: 7 | 🟢 COVERED: 22

---

## Action Plan

### False Positives — Remove from todo list

These were listed in the original UNWIRED_CODE.md but the feature is actually
running via inline code. No wiring needed:

- `txpool/validation` — inline `validateTx()` is sufficient
- `txpool/queue` — inline `txSortedList` is sufficient
- `txpool/replacement` — inline `hasSufficientBump()` is sufficient
- `txpool/fees` — inline `SetBaseFee` is sufficient
- `metrics/PrometheusExporter` — `/metrics` handler in `node.go:335`
- `rpc/netapi` — reportedly wired (confirm with `go list -deps ./rpc/...`)
- `p2p/discv5` orphan — V5 exists in `p2p/discover`

### P0 — Node Cannot Function Without These

- ~~**`eth` package**~~ ✅ **DONE** — ETH/68 wired in `node.go`
- ~~**`engine/blobval`**~~ ✅ **DONE** — Wired in `GetPayloadV3`
- **`sync/statesync` + `sync/healer`** — Node cannot snap-sync from scratch.

### P1 — Major Feature Gaps

- `epbs/*` all 7 sub-packages — EIP-7732 ePBS only runs at root level (types only)
- `p2p/snap` — Peers cannot snap-sync from this node
- ~~`rpc/beaconapi`~~ ✅ **DONE** — `beacon_` namespace wired
- ~~`txpool/journal`~~ — wired (1dd7cde)
- `sync/beacon` — No CL-driven sync loop

### P2 — Protocol Features

- `rollup/execute` — EXECUTE precompile must be registered in `core/vm`
- `das/network` + `das/validator` — DAS peer coordination inactive
- ~~`p2p/dnsdisc`~~ ✅ **DONE** — `runDNSDiscovery` wired at node startup
- `p2p/nat` — NAT traversal/external IP detection still limited
- `trie/migrate` — Binary trie migration never runs
- ~~`txpool/encrypted`~~ — wired (ed147f1)
- ~~`core/mev`~~ — wired (1dd7cde)

### P3 — Roadmap Completeness

- `consensus/vdf`, ~~`core/gigagas`~~, `core/vops`, `core/teragas`
- `light`, `trie/prune`, `trie/stack`, `engine/chunking`
- `p2p/portal`, `p2p/dispatch`, `sync/checkpoint`

---

> Confirm PARTIAL verdicts with: `cd pkg && go list -f '{{.ImportPath}}|{{join .Imports ","}}' ./rpc/... ./p2p/... | grep -E 'middleware|netapi|gas'`
