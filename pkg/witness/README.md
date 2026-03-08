# witness

Execution witness collection, stateless re-execution, and proof generation for EIP-8025 stateless block validation.

## Overview

The `witness` package implements the execution witness framework that enables stateless block validation in ETH2030. Rather than requiring a full copy of the world state, a witness bundles exactly the pre-state data accessed during block execution—account balances, nonces, code, storage slots, and ancestor headers—so that any node can re-execute and verify a block without the full state trie.

The package is organized around three complementary roles. The `WitnessCollector` wraps a live `vm.StateDB` during block execution, recording every state read into a `BlockWitness`. The `WitnessStateDB` consumes that witness to enable full EVM re-execution without any trie access, acting as a self-contained in-memory state database backed entirely by witness data. The `WitnessAggregator` merges witnesses from multiple blocks into range witnesses for efficient light client verification and multi-block proofs.

Proof types follow the EIP-8025 specification: an `ExecutionProof` wraps serialized proof bytes tagged with a `ProofType` identifier (SP1 = 1, ZisK = 2, RISC Zero = 3) and enforces a 300 KiB maximum proof size. The `ProofGenerator` creates proofs from block witnesses, and the `BlockWitnessProducer` coordinates witness collection and proof production across a block's full transaction set.

## Table of Contents

- [Functionality](#functionality)
- [Subpackages](#subpackages)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### Core Witness Types

**`BlockWitness`** is the primary witness container:
- `State map[types.Address]*AccountWitness` — pre-state for all accessed accounts
- `Codes map[types.Hash][]byte` — contract bytecode keyed by code hash
- `Headers map[uint64]*types.Header` — ancestor headers for `BLOCKHASH` opcode

**`AccountWitness`** records an account's pre-state:
- `Balance`, `Nonce`, `CodeHash`, `Exists` — top-level account fields
- `Storage map[types.Hash]types.Hash` — accessed storage slot pre-values

**`ExecutionWitness`** is the Verkle-tree-oriented witness structure:
- `State []StemStateDiff` — leaf diffs organized by 31-byte Verkle stem
- Each `SuffixStateDiff` records `CurrentValue` (pre-state) and `NewValue` (post-state) for a single leaf

### Witness Collection

`WitnessCollector` implements `vm.StateDB` as a transparent proxy. It wraps any `StateDB` implementation and intercepts state reads to populate a `BlockWitness`:

- On first access to an address, `recordAccount` snapshots balance, nonce, code hash, and existence into `AccountWitness`
- `GetState` / `GetCommittedState` record storage slot pre-values on first slot access
- `GetCode` also stores bytecode in `witness.Codes` keyed by code hash
- Write operations (`AddBalance`, `SetState`, `SetCode`, etc.) delegate to the inner `StateDB` and snapshot pre-state before the write
- Snapshots via `Snapshot()` / `RevertToSnapshot()` do not revert witness data — pre-state values are preserved regardless of transaction reversion

```go
var _ vm.StateDB = (*WitnessCollector)(nil) // compile-time interface check
```

### Stateless Re-execution

`WitnessStateDB` implements `vm.StateDB` backed entirely by a `BlockWitness`. It initializes an in-memory overlay from the witness pre-state and applies writes to that overlay. Account reads that are not present in the witness return errors (`ErrAccountNotInWitness`, `ErrSlotNotInWitness`, `ErrCodeNotInWitness`). Snapshot and revert semantics are fully implemented via `witnessSnapshot` copies.

`NewWitnessStateDB(w *BlockWitness) *WitnessStateDB` creates a state DB ready for stateless block re-execution.

### Execution Proofs (EIP-8025)

**`ExecutionProof`** wraps a ZK proof of correct block execution:
- `ProofType uint8` — prover system identifier (SP1 / ZisK / RISC Zero)
- `ProofBytes []byte` — serialized proof data
- `MaxProofSize = 300 * 1024` bytes (per EIP-8025)
- `Validate()` checks non-empty proof, size bound, and known proof type
- `ProofTypeName(pt)` returns a human-readable prover name

### Proof Generation

`ProofGenerator` derives an `ExecutionProof` from a `BlockWitness`. The proof encodes the witness state diffs and execution parameters into a format suitable for ZK verification.

`BlockWitnessProducer` wraps block execution to produce both the witness and the proof in a single pass.

### State Witnesses

`StateWitness` (in `state_witness.go`) captures the minimal state needed to verify a single account or storage value: a Merkle proof path from the state root to the leaf.

### Witness Aggregation

`WitnessAggregator` merges multiple single-block `BlockExecutionWitness` values into a range witness covering a contiguous block range. It deduplicates account and storage entries that appear in multiple blocks, reducing total witness size for light clients. Thread-safe for concurrent block additions.

Key methods:
- `AddBlockWitness(blockNum, witness)` — register a block's witness (any order)
- `Aggregate(from, to)` — produce a merged `RangeWitness` for blocks `[from, to]`
- `GetBlockWitness(blockNum)` — retrieve a single block's witness
- `BlockRange()` — returns `(minBlock, maxBlock)` currently tracked

### Witness Encoding and Caching

`encoding.go` implements binary serialization for `BlockWitness` and `ExecutionWitness`, suitable for network transmission. The `cache/` subpackage provides an LRU cache for recently produced witnesses to avoid redundant recomputation during sync.

## Subpackages

| Subpackage | Description |
|------------|-------------|
| [`cache/`](./cache/) | LRU witness cache for recently produced block witnesses |

## Usage

```go
import "github.com/eth2030/eth2030/witness"

// Wrap a live StateDB during block execution to collect a witness.
bw := witness.NewBlockWitness()
collector := witness.NewWitnessCollector(liveStateDB, bw)

// Execute the block using the collector as the StateDB.
// All state reads are automatically recorded into bw.
processor.Process(block, collector)

// bw now contains all pre-state data accessed during execution.
```

```go
// Re-execute a block statelessly using only the witness.
stateDB := witness.NewWitnessStateDB(bw)
processor.Process(block, stateDB)
// No trie access required — all state comes from the witness.
```

```go
// Validate an execution proof.
proof := &witness.ExecutionProof{
    ProofType:  witness.ProofTypeSP1,
    ProofBytes: proofData,
}
if err := proof.Validate(); err != nil {
    return fmt.Errorf("invalid proof: %w", err)
}
```

```go
// Aggregate witnesses from a range of blocks for light client delivery.
agg := witness.NewWitnessAggregator()
for num, w := range blockWitnesses {
    agg.AddBlockWitness(num, w)
}
rangeWitness, err := agg.Aggregate(fromBlock, toBlock)
```

## Documentation References

- [EIP-8025: Execution Witness](https://eips.ethereum.org/EIPS/eip-8025)
- [EIP-6800: Verkle State Trie](https://eips.ethereum.org/EIPS/eip-6800)
- [L1 Strawmap Roadmap — Beam Sync (I+)](../../docs/ROADMAP.md)
