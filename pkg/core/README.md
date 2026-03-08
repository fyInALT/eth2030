# core

Ethereum Execution Layer (EL) state machine: state transition, gas pricing, and block processing.

## Overview

Package `core` is the top-level entry point for Ethereum's execution layer state
machine. It exports `StateTransition`, which orchestrates block-level execution:
validating transactions, applying them against the world state, computing EIP-1559
base fee burns, EIP-4844 blob gas accounting, withdrawal processing, and
post-block validation. The package brings together the functionality of all
subpackages described below.

The `core` package targets the full L1 Strawmap roadmap from Glamsterdam through
M+. It implements 18+ gas repricing EIPs, 5-dimensional gas pricing, BPO blob
schedules, EIP-7928 BAL integration, frame transaction execution (EIP-8141),
native account abstraction (EIP-7701/7702), gigagas infrastructure (1 Ggas/sec),
MEV protection, validity-only partial statelessness (VOPS), and the gas futures
market.

## Table of Contents

- [Functionality](#functionality)
- [Subpackages](#subpackages)
- [Documentation References](#documentation-references)

## Functionality

### State Transition

`StateTransition.ApplyBlock(block, statedb)` is the main entry point. It iterates
over all transactions in the block, validates each one, applies it to the state via
the EVM executor, burns EIP-1559 base fees, tracks blob gas consumed (EIP-4844), and
processes EIP-4895 withdrawals. It returns a `TransitionResult` containing receipts,
gas used, blob gas used, logs bloom, and the post-state root.

### Configuration and Genesis

The `config/` subpackage defines `ChainConfig` with all fork activation timestamps
and block numbers, including Glamsterdam, Hegotá, I+, J+, K+, L+, and M+ fork
levels. It includes genesis allocation helpers (`GenesisAlloc`), chain config
extensions for custom fork ordering and validation, and `GenesisInit` for
programmatic genesis creation.

### Gas Pricing (18+ EIPs)

The `gas/` subpackage implements the complete gas pricing stack:

- **EIP-1559 base fee**: `fee.go` computes next-block base fees.
- **EIP-4844 blob gas**: `blob_gas.go` / `blob_schedule.go` implement blob gas
  accounting, BPO1/BPO2 blob schedules, and EIP-7918 blob base fee floor.
- **EIP-7706 multidimensional gas**: `multidim_gas.go` implements a 5-dimensional
  pricing engine (compute, storage, bandwidth, blob, witness) with independent
  EIP-1559-style base fee adjustment per dimension.
- **Calldata gas (EIP-7623)**: `calldata_gas.go` and `eip7623_floor.go` implement
  the floor cost and EIP-7706 calldata gas alias `GetCalldataGas()`.
- **Gas cap (EIP-7825)**: `gas_cap.go` enforces per-transaction and per-block gas
  caps.
- **Gas limit schedule**: `gas_limit.go` implements scheduled gas limit increases.
- **Glamsterdam and Hegotá repricing**: `glamsterdam_repricing.go` and
  `hogota_repricing.go` implement the 18-EIP repricing bundles for those forks.
- **Gas futures market**: `gas_futures.go`, `gas_market.go`, and `gas_settlement.go`
  implement long-dated gas futures contracts and settlement (M+ roadmap).
- **Gas pool**: `gas_pool_extended.go` provides pool extensions for multidimensional
  gas.
- **Gas estimator**: `gas_estimator.go` implements binary-search gas estimation.

### EIPs (Execution Layer)

The `eips/` subpackage implements each EIP as a self-contained module:

- `eip6110.go` — in-protocol deposit receipts
- `eip7002.go` — EL-triggered withdrawals
- `eip7702.go` — SetCode (native EOA delegation)
- `beacon_root.go` — EIP-4788 beacon root storage
- `eip2935.go` — historical block hash access
- `frame_execution.go` — EIP-8141 frame transaction execution
- `inclusion_list.go` — FOCIL inclusion list enforcement
- `payload_chunking.go` — Hegotá payload chunking
- `block_in_blobs.go` — Hegotá block-in-blobs encoding
- `tx_assertions.go` — EVM transaction assertions
- `aa_entrypoint.go` — EIP-7701 AA entry point
- `paymaster_registry.go` — EIP-7701 paymaster registry
- `access_gas.go` — EIP-2929/2930 access list gas

### Execution

The `execution/` subpackage provides the transaction execution pipeline:

- `Processor` — applies transactions against the EVM, generates receipts, and
  hooks into the BAL tracker.
- `ReceiptProcessor` and `ReceiptGeneration` — receipt RLP encoding and root
  computation.
- `ParallelExecutor` — BAL-guided parallel execution using dependency graphs.
- `RichData` — block-level metadata collection (inclusion delays, gas breakdown).
- `DependencyGraph` — constructs execution dependency graphs from BAL entries.

### State

The `state/` subpackage implements the `StateDB` interface:

- In-memory state (`MemStateDB`) and trie-backed state.
- Access events (EIP-4762 statelessness gas, `access_events.go`).
- `StatelessStateDB` — witness-backed state for stateless execution.
- `StatePrefetcher` — pre-fetches state for upcoming blocks.
- `AccountTrie` — direct account-level trie access.
- `BalsEngine` — integrates with the BAL package for opcode-level tracking.
- `Snapshots` (via `state/snapshot/`) — layered diff/disk snapshot architecture
  with account/storage iterators and pruner.

### Chain

The `chain/` subpackage manages the canonical chain:

- `Blockchain` — full chain manager with fork choice integration, reorg handling,
  block import, and state cache.
- `Forkchoice` — EL fork choice adapter (receives head/safe/finalized from CL via
  Engine API).
- `HeaderChain` — header-only chain management.
- `HeaderVerification` — block header validation.
- `StateCache` — LRU state cache for recent blocks.

### Block

The `block/` subpackage provides block construction and validation:

- `Builder` — constructs blocks from a transaction pool, applying BAL tracking,
  blob scheduling, and EIP-8141 frame execution.
- `Executor` — executes a pre-built block against a state.
- `Validator` — post-execution block validation (state root, receipt root, bloom).
- `iface.go` — `BlockBuilder`, `BlockExecutor`, `BlockValidator` interfaces.

### GigaGas (1 Ggas/sec, M+)

The `gigagas/` subpackage implements the gigagas infrastructure for the M+ North
Star goal of 1 billion gas per second:

- `GasRateTracker` — sliding-window gas throughput measurement.
- `GigagasScheduler` — work-stealing parallel scheduler for gigagas-scale blocks
  (target: `DefaultGigagasConfig.TargetGasPerSecond = 1_000_000_000`).
- `GigagasExecutor` — gigagas-capable transaction executor.
- `WorkStealing` — work-stealing goroutine pool for load balancing.

### VOPS (Validity-Only Partial Statelessness)

The `vops/` subpackage implements VOPS for stateless block validation (I+ roadmap):

- `PartialState` — partial world state subset (accounts, storage, code).
- `Executor` — executes transactions against a partial state.
- `Validator` — validates state transitions using partial state and access proofs.
- `ProofChecker` — verifies storage proofs against the state root.
- `WitnessAccumulator` — accumulates the execution witness during validation.
- `CompleteVOPS` — full VOPS validator with access list and storage proof checks.

### MEV Protection

The `mev/` subpackage provides MEV (Maximal Extractable Value) protection:

- `FlashbotsBundle` — represents a bundle of transactions for atomic inclusion with
  target block, timestamp constraints, and revert protection.
- `SandwichDetector` — detects sandwich attack patterns.
- `FrontrunDetector` — detects frontrunning patterns.
- `FairOrderEnforcer` — enforces fair ordering rules (MEV commit-reveal ordering).

### Teragas (L2 Throughput)

The `teragas/` subpackage implements teragas-scale scheduling infrastructure for
the L+ era (1 Gbyte/sec L2):

- `Scheduler` — bandwidth-enforced teragas scheduler with streaming pipeline.

### Raw Database

The `rawdb/` subpackage provides the persistent storage layer:

- `FileDB` — append-only file database with WAL for crash safety.
- `ChainDB` — block, receipt, and transaction storage with EIP-4444 history expiry.
- `AncientStore` — ancient (frozen) chain segment storage.
- `Batch` — atomic write batches.

### Rate Metering

The `ratemeter/` subpackage provides `RateMeter` — an EWMA-based rate meter for
tracking gas throughput, blob bandwidth, and sync rates.

## Subpackages

| Subpackage | Description |
|------------|-------------|
| [`block/`](./block/) | Block construction, execution, and validation interfaces |
| [`chain/`](./chain/) | Canonical chain manager, fork choice, and reorg handling |
| [`config/`](./config/) | Chain configuration, fork activation, and genesis |
| [`eftest/`](./eftest/) | EF state test runner (36,126/36,126 passing via go-ethereum backend) |
| [`eips/`](./eips/) | Individual EIP implementations (EIP-6110, 7002, 7702, 8141, etc.) |
| [`execution/`](./execution/) | Transaction execution pipeline with receipt generation |
| [`gas/`](./gas/) | Gas pricing: EIP-1559, multidim (EIP-7706), blob (EIP-4844/7918), repricing, futures |
| [`gaspool/`](./gaspool/) | Gas pool tracking for block gas limit enforcement |
| [`gigagas/`](./gigagas/) | Gigagas (1 Ggas/sec) infrastructure with work-stealing scheduler (M+) |
| [`mev/`](./mev/) | MEV protection: bundle validation, sandwich/frontrun detection, fair ordering |
| [`ratemeter/`](./ratemeter/) | EWMA rate meter for gas and bandwidth throughput |
| [`rawdb/`](./rawdb/) | Persistent storage: FileDB with WAL, block/receipt/tx storage, EIP-4444 expiry |
| [`state/`](./state/) | StateDB: in-memory, trie-backed, stateless (witness-backed), access events |
| [`teragas/`](./teragas/) | Teragas (1 Gbyte/sec L2) scheduling infrastructure (L+) |
| [`types/`](./types/) | Core EL types: Header, Transaction (7 types), Receipt, Block, Account |
| [`vm/`](./vm/) | EVM interpreter: 164+ opcodes, 24 precompiles, EOF (EIP-3540) |
| [`vops/`](./vops/) | Validity-Only Partial Statelessness: partial executor, proof checker, witness |

## Documentation References

- [Roadmap](../../docs/ROADMAP.md)
- [Design Doc](../../docs/DESIGN.md)
- [Roadmap Deep-Dive](../../docs/ROADMAP-DEEP-DIVE.md)
- [EIP Spec Implementation](../../docs/EIP_SPEC_IMPL.md)
- [Gap Analysis](../../docs/GAP_ANALYSIS.md)
- [EIP-1559](https://eips.ethereum.org/EIPS/eip-1559)
- [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844)
- [EIP-7706: Separate Gas for Calldata](https://eips.ethereum.org/EIPS/eip-7706)
- [EIP-7928: Block Access Lists](https://eips.ethereum.org/EIPS/eip-7928)
- [EIP-8141: Frame Transactions](https://eips.ethereum.org/EIPS/eip-8141)
