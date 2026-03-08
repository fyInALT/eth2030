# Plan: Refactor `pkg/core` — Break Monolith into Functional Packages

**Branch**: `feat/split-pkgs-v2`
**Date**: 2026-03-07
**Scope**: `pkg/core/` root (~134 files, ~29,500 LOC)

---

## 1. Problem Statement

`pkg/core/` is a classic Go monolith: 134 source files in a single package with overlapping concerns, tightly coupled test suites, and no clear sub-package boundaries for the gas system, EIP-specific logic, or execution pipeline. This makes it hard to:

- Understand which files belong to which feature
- Test individual subsystems in isolation
- Onboard contributors to a specific area (gas, receipts, AA, etc.)
- Avoid circular imports when extracting functionality

---

## 2. Goals

1. Extract cohesive subsystems from `core/` root into dedicated sub-packages
2. Replace any remaining `*EngineBackend` concrete type assertions with interface types
3. Maintain 100% test coverage — every test must still pass after each move
4. Each commit must compile and pass `go vet ./...`
5. No circular imports between new packages

---

## 3. Current Structure Analysis

### 3.1 File Groups Identified

| Group | Files (approx) | LOC | Current location |
|-------|---------------|-----|-----------------|
| Gas system | 14 | ~3,700 | `core/` root |
| EIP implementations | 12 | ~3,200 | `core/` root |
| Genesis & chain config | 7 | ~1,845 | `core/` root |
| Blockchain state machine | 6 | ~2,643 | `core/` root |
| Execution pipeline | 9 | ~2,196 | `core/` root |
| Block building & validation | 3 | ~1,681 | `core/` root |
| Receipt processing | 3 | ~597 | `core/` root |
| MEV / ordering | 1 | ~379 | `core/` root |
| Already extracted | gaspool/, ratemeter/ | ~900 | sub-packages |

### 3.2 Already-Extracted Sub-packages (keep as-is)

- `core/gaspool/` — GasPool
- `core/ratemeter/` — RateMeter
- `core/gigagas/` — gigagas scheduler
- `core/vops/` — VOPS validator
- `core/eftest/` — EF state tests
- `core/state/` — StateDB + impls
- `core/rawdb/` — FileDB + WAL
- `core/vm/` — EVM interpreter
- `core/types/` — canonical types

---

## 4. Target Package Layout

```
pkg/core/
├── gas/              NEW — multidimensional gas system
│   ├── pool.go           (from gaspool_compat.go → gaspool/)
│   ├── multidim.go       (multidim.go, multidim_gas.go, multidim_market.go)
│   ├── blob.go           (blob_gas.go, blob_schedule.go, blob_validation.go)
│   ├── calldata.go       (calldata_gas.go, eip7623_floor.go)
│   ├── cap.go            (gas_cap.go, gas_cap_extended.go)
│   ├── limit.go          (gas_limit.go)
│   ├── repricing.go      (glamsterdam_repricing.go, hogota_repricing.go)
│   ├── market.go         (gas_market.go, multidim_market.go)
│   ├── estimator.go      (gas_estimator.go)
│   ├── futures.go        (gas_futures.go, gas_settlement.go)
│   └── fee.go            (fee.go)
│
├── execution/        NEW — transaction execution pipeline
│   ├── processor.go      (processor.go — StateProcessor)
│   ├── result.go         (execution_result.go, execution_result_extended.go)
│   ├── parallel.go       (parallel_processor.go, dependency_graph.go)
│   ├── receipt.go        (receipt_generation.go, receipt_processor.go)
│   └── rich.go           (rich_data.go)
│
├── eips/             NEW — per-EIP implementations
│   ├── eip2935.go        (eip2935.go — block hashes)
│   ├── eip6110.go        (eip6110.go — deposits)
│   ├── eip7002.go        (eip7002.go — exits)
│   ├── eip7702.go        (eip7702.go — SetCode auth)
│   ├── eip7997.go        (eip7997.go — ParentBeaconRoot)
│   ├── beacon_root.go    (beacon_root.go — EIP-4788)
│   ├── block_in_blobs.go (block_in_blobs.go — EIP-7898)
│   ├── payload_chunking.go (payload_chunking.go, payload_shrink.go)
│   ├── frame.go          (frame_execution.go)
│   ├── aa.go             (aa_entrypoint.go, paymaster_registry.go)
│   ├── focil.go          (inclusion_list.go)
│   ├── bal.go            (elsa.go — BAL memory accounting)
│   ├── access.go         (access_gas.go — EIP-4762)
│   └── tx_assert.go      (tx_assertions.go)
│
├── chain/            NEW — blockchain state machine
│   ├── blockchain.go     (blockchain.go)
│   ├── forkchoice.go     (forkchoice.go)
│   ├── reorg.go          (chain_reorg.go)
│   ├── reader.go         (chain_reader.go, chain_reader_ext.go)
│   └── header.go         (headerchain.go, header_verification.go)
│
├── block/            NEW — block building & validation
│   ├── builder.go        (block_builder.go)
│   ├── executor.go       (block_executor.go)
│   └── validator.go      (block_validator.go)
│
├── config/           NEW — genesis & chain config
│   ├── chain_config.go   (chain_config.go, chain_config_forks.go, chain_config_ext.go)
│   ├── genesis.go        (genesis.go, genesis_init.go, genesis_alloc.go, genesis_utils.go)
│   └── message.go        (message.go)
│
├── mev/              NEW — MEV/ordering protection
│   └── mev.go            (mev.go — commit-reveal ordering)
│
├── teragas/          NEW — teragas bandwidth scheduler
│   └── scheduler.go      (teragas_scheduler.go)
│
│   ── KEEP IN core/ ROOT ──
├── state_transition.go   (orchestration glue — thin wrapper)
├── gaspool_compat.go     → remove (inline or delete)
├── rate_meter_compat.go  → remove (inline or delete)
└── bloombits.go          (tiny util — stays)
```

---

## 5. Dependency Graph (new packages)

```
types/ ←──────────────────────────────┐
rawdb/ ←──────────────────────────────┤
state/ ←──────────────────────────────┤
vm/   ←──────────────────────────────┤
                                      │
config/  ←── no internal deps        │
gas/     ←── types/, config/         │
eips/    ←── types/, state/, gas/    │
receipt/ ←── types/, gas/            │
execution/ ←── types/, state/, gas/, eips/
block/  ←── types/, state/, execution/, eips/
chain/  ←── types/, state/, block/, execution/
mev/    ←── types/
teragas/ ←── types/, gas/
```

No cycles. All arrows point downward.

---

## 6. High-Use Types & Interface Extraction

### 6.1 Usage Frequency (ranked)

The following concrete types appear most often across `core/` root source files as function parameters, return types, or struct fields:

| Rank | Type | Usages | Defined In | Cross-file scope |
|------|------|--------|-----------|-----------------|
| 1 | `*types.Transaction` | 237 | types/transaction.go | 17+ files |
| 2 | `*Blockchain` | 94 | blockchain.go | 12+ files |
| 3 | `*FeeMarket` | 90 | multidim_market.go | 3+ files |
| 4 | `*types.Header` | 110 | types/header.go | 31+ files |
| 5 | `*types.Block` | 83 | types/block.go | 12+ files |
| 6 | `*types.Receipt` | 64 | types/receipt.go | 8+ files |
| 7 | `*BlockValidator` | 42 | block_validator.go | 8+ files |
| 8 | `*StateProcessor` | 39 | processor.go | 13+ files |
| 9 | `*BlockBuilder` | 37 | block_builder.go | 10+ files |
| 10 | `*ExecutionResult` | 34 | execution_result.go | 8+ files |
| 11 | `*BlockExecutor` | 32 | block_executor.go | 2+ files |
| 12 | `*HeaderChain` | 29 | headerchain.go | 3+ files |
| 13 | `*DependencyGraph` | 23 | dependency_graph.go | 2+ files |
| 14 | `*ReceiptGenerator` | 20 | receipt_generation.go | 3+ files |
| 15 | `*StateTransition` | 11 | state_transition.go | 5+ files |

**Value types that stay as structs** (well-defined, no behaviour swap needed):
`*types.Transaction`, `*types.Header`, `*types.Block`, `*types.Receipt` — these are canonical wire/storage types, not execution behaviour.

---

### 6.2 Existing Interfaces (keep as-is)

These are already well-designed — do **not** break them:

| Interface | File | Methods | Status |
|-----------|------|---------|--------|
| `state.StateDB` | core/state/statedb.go | 23 | ✓ Excellent — used everywhere |
| `ChainReader` | chain_reader.go | 10 | ✓ Good read-only chain access |
| `FullChainReader` | chain_reader_ext.go | extends ChainReader | ✓ Keep |
| `TxPoolReader` | block_builder.go | 1 (Pending) | ✓ Minimal, ideal |
| `PaymasterSlasher` | message.go | 1 (SlashOnBadSettlement) | ✓ Keep |
| `vm.BALTracker` | vm/bal_tracker.go | EIP-7928 hooks | ✓ Keep |

---

### 6.3 Type Assertions Requiring Remediation

These concrete type assertions create hard coupling and must be replaced:

| Assertion | File | Risk | Fix |
|-----------|------|------|-----|
| `statedb.(*state.MemoryStateDB)` | blockchain.go (3×) | **HIGH** — panics on alt StateDB | Store as `state.StateDB`; remove assertion |
| `statedb.(*state.MemoryStateDB)` | genesis_utils.go | HIGH | Same fix |
| `statedb.(*state.MemoryStateDB)` | parallel_processor.go | MEDIUM (guarded) | Remove MemoryStateDB-specific fast path |
| `layer.(*diffLayer)` | state/snapshot/* (8×) | LOW — snapshot-internal | Acceptable inside snapshot package |
| `current.(*diskLayer)` | state/snapshot/* (6×) | LOW — snapshot-internal | Acceptable inside snapshot package |

---

### 6.4 Priority 1 — Extract These Interfaces (High Impact)

#### `BlockchainReader` — replaces `*Blockchain` in consumers

**Current coupling**: `BlockBuilder`, `BlockExecutor`, and `chain_reorg.go` all hold `*Blockchain` directly.
**Problem**: impossible to test builder/executor without a real Blockchain.

```go
// core/chain/reader.go  (new location after refactor)
type BlockchainReader interface {
    Config() *params.ChainConfig
    CurrentBlock() *types.Block
    CurrentHeader() *types.Header
    GetBlock(hash common.Hash, number uint64) *types.Block
    GetHeader(hash common.Hash, number uint64) *types.Header
    GetReceipts(hash common.Hash) types.Receipts
    GetTd(hash common.Hash, number uint64) *big.Int
    StateAt(root common.Hash) (state.StateDB, error)
    HasBlock(hash common.Hash, number uint64) bool
}
```

Callers that change: `BlockBuilder` constructor, `BlockExecutor`, `chain_reorg.go`.

---

#### `TxExecutor` — replaces `*StateProcessor` in Blockchain

**Current coupling**: `Blockchain` holds `processor *StateProcessor` and calls `processor.Process(...)` directly.
**Problem**: impossible to swap in parallel or zkVM-backed processor without changing Blockchain.

```go
// core/execution/processor.go
type TxExecutor interface {
    Process(block *types.Block, statedb state.StateDB, cfg vm.Config) ([]*types.Receipt, error)
    ProcessWithBAL(block *types.Block, statedb state.StateDB, cfg vm.Config) (*ProcessResult, error)
    SetGetHash(fn vm.GetHashFunc)
    SetSlasher(s PaymasterSlasher)
}
```

`StateProcessor` implements `TxExecutor`. `Blockchain` stores `processor TxExecutor`.

---

#### `Validator` — replaces `*BlockValidator` in Blockchain and HeaderChain

**Current coupling**: both `Blockchain` and `HeaderChain` hold `*BlockValidator` directly.
**Problem**: fork-specific validation rules require conditional logic scattered across one struct.

```go
// core/block/validator.go
type Validator interface {
    ValidateHeader(header *types.Header, parent *types.Header) error
    ValidateBody(block *types.Block) error
    ValidateState(block *types.Block, statedb state.StateDB, receipts types.Receipts, usedGas uint64) error
}
```

`BlockValidator` implements `Validator`. Both `Blockchain` and `HeaderChain` store `validator Validator`.

---

#### `HeaderStore` — replaces `*HeaderChain` in Blockchain

**Current coupling**: `Blockchain` holds `hc *HeaderChain` as a concrete field.
**Problem**: alternative header storage (e.g., light-client backed) requires forking Blockchain.

```go
// core/chain/header.go
type HeaderStore interface {
    GetHeader(hash common.Hash, number uint64) *types.Header
    GetHeaderByNumber(number uint64) *types.Header
    CurrentHeader() *types.Header
    InsertHeaders(headers []*types.Header) (int, error)
    SetHead(number uint64) error
}
```

`HeaderChain` implements `HeaderStore`. `Blockchain` stores `hc HeaderStore`.

---

### 6.5 Priority 2 — Consider Extracting (Medium Impact)

#### `BlockProducer` — replaces `*BlockBuilder` in engine API

Useful for ePBS separation: proposer calls `BuildBlock`, builder implements the interface.

```go
// core/block/builder.go
type BlockProducer interface {
    BuildBlock(attrs BuildBlockAttributes, statedb state.StateDB) (*types.Block, error)
    ValidateForInclusion(tx *types.Transaction) error
}
```

---

#### `PricingEngine` — replaces `*FeeMarket` in processor

`FeeMarket` has 90 usages but they all call a small set of methods. Extracting an interface makes gas market pluggable.

```go
// core/gas/market.go
type PricingEngine interface {
    BaseFee() *big.Int
    BlobBaseFee() *big.Int
    CalculateFee(tx *types.Transaction) (*DimensionalPrice, error)
    UpdateAfterBlock(gasUsed uint64, blobGasUsed uint64) error
}
```

---

#### `ExecutionOutcome` — replaces `*ExecutionResult` in receipt pipeline

Lower priority: `ExecutionResult` is a plain data struct, not a behaviour type. Extract only if receipt generation moves to its own package and needs to be decoupled from the execution package.

---

### 6.6 Do-Not-Extract (Good as Concrete Structs)

| Type | Reason |
|------|--------|
| `*types.Transaction` | Wire-format type; all consumers need full fields |
| `*types.Header` | Wire-format type; field-level access patterns throughout |
| `*types.Block` | Aggregation of Header + Transactions; no swap needed |
| `*types.Receipt` | Output record; no polymorphism required |
| `*DependencyGraph` | Internal parallel-executor detail; stays in execution/ |
| `*StateTransition` | Thin orchestration shim; will be deleted post-refactor |

---

### 6.7 Interface Introduction Order (aligned with Phase plan)

| Phase | Interface | Replaces | Introduced in file |
|-------|-----------|---------|-------------------|
| Before Phase 4 | `TxExecutor` | `*StateProcessor` | `core/execution/iface.go` |
| Before Phase 5 | `Validator` | `*BlockValidator` | `core/block/iface.go` |
| Before Phase 5 | `BlockchainReader` | `*Blockchain` | `core/chain/iface.go` |
| Before Phase 5 | `HeaderStore` | `*HeaderChain` | `core/chain/iface.go` |
| Phase 5 | `BlockProducer` | `*BlockBuilder` | `core/block/iface.go` |
| Phase 5 | `PricingEngine` | `*FeeMarket` | `core/gas/iface.go` |

Each interface file (`iface.go`) in its package declares the interface and a compile-time assertion that the concrete type satisfies it:

```go
var _ TxExecutor = (*StateProcessor)(nil)   // compile-time check
```

---

## 7. Migration Steps (Ordered)

Each step = one atomic commit that compiles + passes tests.

### Phase 1 — Config & Types (no deps on other core files)
1. `core/config/` — move `chain_config*.go`, `genesis*.go`, `message.go`
2. Update all import paths in `core/` root and `engine/`, `node/`

### Phase 2 — Gas System
3. `core/gas/` — move all 14 gas files
4. Update imports across `processor.go`, `block_builder.go`, `block_validator.go`
5. Remove `gaspool_compat.go` and `rate_meter_compat.go` (wire directly)

### Phase 3 — EIP Implementations
6. `core/eips/` — move EIP-specific files (no dependency on processor)
7. Update imports in `processor.go`, `block_builder.go`

### Phase 4 — Execution Pipeline
8. `core/execution/` — move `processor.go`, result types, parallel, receipt
9. Update imports in `block/`, `chain/`

### Phase 5 — Block & Chain
10. `core/block/` — move builder, executor, validator
11. `core/chain/` — move blockchain, forkchoice, reorg, reader, header
12. Update engine API imports

### Phase 6 — MEV & Teragas
13. `core/mev/` — move mev.go
14. `core/teragas/` — move teragas_scheduler.go

### Phase 7 — Cleanup
15. Remove `state_transition.go` if fully superseded
16. Verify `bloombits.go` still needed or move to `core/`
17. Run full test suite: `cd pkg && go test ./...`
18. Run `go vet ./...` and `go fmt ./...`

---

## 8. Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| Circular imports | Draw dep graph before moving each file; use `go build` to verify |
| Test breakage | Move test files alongside source files; run tests after every step |
| Large e2e tests | Keep `e2e_test.go` in `core/` root (integration-level) until all units stable |
| Interface mismatch | Extract interfaces first, then move impl; use compile-check pattern |
| Partial state | Never push a half-migrated commit; each commit = fully compilable |

---

## 9. Success Criteria

- [ ] `cd pkg && go build ./...` — clean
- [ ] `cd pkg && go vet ./...` — zero warnings
- [ ] `cd pkg && go test ./...` — all pass (18,400+ tests)
- [ ] No file in `core/` root that clearly belongs to a sub-package
- [ ] No `*ConcreteType` assertion in `engine/` or `node/`
- [ ] Each new sub-package has its own `_test.go` files
- [ ] `docs/GAP_ANALYSIS.md` statistics updated if package count changes

---

## 10. Out of Scope

- Moving `vm/`, `state/`, `rawdb/`, `types/` — already well-structured
- Changing any EIP logic or behavior — pure structural refactor
- Adding new features during the refactor
