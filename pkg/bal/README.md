# bal

Block Access Lists (EIP-7928) for parallel transaction execution.

## Overview

Package `bal` implements Block Access Lists as specified by EIP-7928. A Block Access
List (BAL) is a per-block data structure that records every state access (storage
reads, storage writes, balance/nonce/code changes) that occurred during block
execution, tagged with the access index (0 = pre-execution system calls, 1..n =
transaction indices, n+1 = post-execution system calls).

The primary purpose of the BAL is to enable parallel transaction scheduling. By
declaring what state each transaction touches before or during execution, the scheduler
can identify independent transactions and group them into execution waves that run
concurrently. Two transactions conflict when they access the same (address, slot) pair
and at least one of them writes to it. The conflict detector classifies conflicts as
read-write, write-read, write-write, or account-level.

The package also implements EIP-7928's ordering invariant: entries within a BAL must be
sorted in strict ascending lexicographic order by (Address, AccessIndex). The validator
enforces this at block validation time, and the `BALItemCost` constant (2000 gas units
per item) bounds the total number of BAL entries to `gasLimit / 2000`.

## Table of Contents

- [Functionality](#functionality)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### Core Types

`BlockAccessList` is the top-level container holding a slice of `AccessEntry` values.
Each `AccessEntry` records the address, access index, storage reads
(`[]StorageAccess`), storage changes (`[]StorageChange`), and optional balance, nonce,
and code changes. `BlockAccessList.Sort()` sorts entries into the EIP-7928-required
lexicographic order.

### Access Tracking

`AccessTracker` collects raw state accesses during execution of a single transaction.
It exposes `RecordStorageRead`, `RecordStorageChange`, `RecordBalanceChange`,
`RecordNonceChange`, `RecordCodeChange`, and `RecordAddressTouch`. After execution,
`Build(txIndex)` produces a `BlockAccessList` with all accesses tagged to the given
transaction index.

`BlockBALTracker` tracks accesses across an entire block, managing the
pre/tx/post-execution phases via `BeginPreExecution()`, `BeginTx(txIndex)`, and
`BeginPostExecution(txCount)`. It enforces the `gasLimit / BALItemCost` size cap and
returns `ErrBALSizeExceeded` if the limit is exceeded.

### Conflict Detection

`BALConflictDetector` analyzes a `BlockAccessList` and identifies all conflicting
transaction pairs. It exposes:

- `DetectConflicts(bal)` — returns a sorted slice of `Conflict` structs, each
  identifying the pair (TxA, TxB), the `ConflictType`, and the conflicting address/slot.
- `IsParallelFeasible(bal)` — returns true if at least one conflict-free pair exists.
- `BuildDependencyGraph(bal)` — produces a DAG mapping each transaction index to its
  predecessor dependencies.
- `ResolveConflicts(conflicts)` — applies the configured `ResolutionStrategy`
  (Serialize, Abort, or Retry) and returns per-transaction action strings.

`ConflictMetrics` tracks running totals for total pairs analyzed, conflicts found by
type, parallel-feasible runs, and serial-required runs via `atomic.Uint64` counters.

An advanced conflict detector (`conflict_detector_advanced.go`) and a graph-based
detector (`conflict_detector_graph.go`) provide additional analysis strategies.

### Parallel Scheduling

`ComputeParallelSets(bal)` performs graph coloring on the conflict graph to find the
maximum set of non-conflicting transactions that can run simultaneously, returning
`[]ExecutionGroup`.

`BALScheduler` uses a dependency graph from `BALConflictDetector` to schedule
transactions into sequential waves of independent tasks:

- `Schedule(bal)` — runs topological sort (Kahn's algorithm) and partitions the result
  into `[]Wave` via level assignment.
- `AssignWorkers(wave)` — distributes wave tasks across workers using round-robin.
- `ExecuteSpeculative(wave, conflictSet)` — runs tasks speculatively in parallel
  goroutines, marking conflicting tasks as rolled back.
- `ReExecute(results)` — serially re-runs rolled-back transactions for correctness.
- `SchedulerMetrics` — tracks waves formed, transactions scheduled, rollbacks, and
  re-executions via atomic counters.

A pipeline-mode scheduler (`scheduler_pipeline.go`) provides a streaming variant
suitable for pipelined block processing.

### Validation

`ValidateBALOrdering(bal)` enforces the EIP-7928 lexicographic ordering requirement.
It returns a descriptive error identifying the first violation. `BALItemCost = 2000`
gas units per item.

### Extended Types and Hashing

`types_extended.go` provides additional BAL-level types and helpers used by the
extended conflict detector and pipeline scheduler. `hash.go` / `hash_extended.go`
provide deterministic content hashing for BAL structures.

## Usage

```go
import "github.com/eth2030/eth2030/bal"

// Track accesses for a single transaction.
tracker := bal.NewTracker()
tracker.RecordStorageRead(addr, slot, value)
tracker.RecordStorageChange(addr, slot, oldVal, newVal)
txBAL := tracker.Build(1) // tx index 1

// Block-level tracking with EIP-7928 size enforcement.
blockTracker := bal.NewBlockBALTracker(gasLimit)
blockTracker.BeginPreExecution()
// ... system calls ...
blockTracker.BeginTx(1)
if err := blockTracker.RecordAccess(addr); err != nil {
    // ErrBALSizeExceeded
}
entries := blockTracker.Build()

// Validate ordering before including in a block.
if err := bal.ValidateBALOrdering(blockBAL); err != nil {
    // ordering violation
}

// Detect conflicts and schedule parallel execution.
detector := bal.NewBALConflictDetector(bal.StrategySerialize)
conflicts := detector.DetectConflicts(blockBAL)
scheduler, _ := bal.NewBALScheduler(8, detector)
waves, _ := scheduler.Schedule(blockBAL)
for _, wave := range waves {
    assignments := scheduler.AssignWorkers(wave)
    results := scheduler.ExecuteSpeculative(wave, conflictSet)
    results = scheduler.ReExecute(results)
}
```

## Documentation References

- [EIP-7928: Block Access Lists](https://eips.ethereum.org/EIPS/eip-7928)
- [Design Doc](../../docs/DESIGN.md)
- [EIP Spec Implementation](../../docs/EIP_SPEC_IMPL.md)
- [Roadmap](../../docs/ROADMAP.md)
