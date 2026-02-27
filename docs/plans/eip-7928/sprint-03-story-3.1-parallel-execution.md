# Story 3.1 — Implement speculative execution and wire parallel BAL scheduler

> **Sprint context:** Sprint 3 — Parallel Execution with Real Rollback
> **Sprint Goal:** Replace the skeleton `ExecuteSpeculative()` with real transaction re-execution that rolls back on conflict and retries, while the pipeline tests prove parallel results match sequential.

**Files:**
- Modify: `pkg/bal/scheduler.go`
- Modify: `pkg/core/parallel_processor.go`
- Test: `pkg/bal/scheduler_rollback_test.go`
- Test: `pkg/core/parallel_vs_sequential_test.go`

**Acceptance Criteria:** `BALScheduler.ExecuteWave()` runs non-conflicting txs concurrently, retries conflicts sequentially; `ProcessParallel()` produces an identical state root to `Process()` for the same block.

#### Task 3.1.1 — Write failing tests

File: `pkg/bal/scheduler_rollback_test.go`

```go
package bal_test

import "testing"

func TestScheduler_ConflictingTxsRetried(t *testing.T) {
	// Create two txs that write to same storage slot
	// Schedule them
	// Assert only one succeeds in first wave; the other retries
	t.Skip("implement after speculative execution")
}
```

File: `pkg/core/parallel_vs_sequential_test.go`

```go
package core_test

import "testing"

func TestParallelProcessor_MatchesSequential(t *testing.T) {
	// Build 10-tx block: 5 independent + 5 conflicting
	// Run Process() -> get stateRoot1
	// Run ProcessParallel() with BAL -> get stateRoot2
	// Assert stateRoot1 == stateRoot2
}
```

#### Task 3.1.2 — Define `StateSnapshot` interface and `ExecutorFunc`

In `pkg/bal/scheduler.go`:

```go
// ExecutorFunc executes a single transaction on a snapshotted state.
// Returns gas used and error. On conflict, returns ErrConflict.
type ExecutorFunc func(txIndex int, snap StateSnapshot) (gasUsed uint64, err error)

// StateSnapshot supports copy-on-write for speculative execution.
type StateSnapshot interface {
	Snapshot() int
	RevertToSnapshot(int)
	Commit() error
}
```

#### Task 3.1.3 — Implement `ExecuteWave` with goroutines

```go
// ExecuteWave executes all transactions in a wave in parallel.
// Conflicts are retried sequentially after the first pass.
func (s *BALScheduler) ExecuteWave(wave Wave, exec ExecutorFunc, state StateSnapshot) error {
	results := make(chan struct{ idx int; err error }, len(wave.TxIndices))
	for _, txIdx := range wave.TxIndices {
		go func(idx int) {
			snap := state.Snapshot()
			_, err := exec(idx, state)
			if err != nil {
				state.RevertToSnapshot(snap)
			}
			results <- struct{ idx int; err error }{idx, err}
		}(txIdx)
	}

	var conflicts []int
	for range wave.TxIndices {
		r := <-results
		if r.err != nil {
			conflicts = append(conflicts, r.idx)
		}
	}
	for _, idx := range conflicts {
		if _, err := exec(idx, state); err != nil {
			return fmt.Errorf("tx %d failed after retry: %w", idx, err)
		}
	}
	return nil
}
```

#### Task 3.1.4 — Wire scheduler into `ProcessParallel`

In `pkg/core/parallel_processor.go`:

```go
for _, wave := range waves {
	if len(wave.TxIndices) == 1 {
		executeTx(wave.TxIndices[0], stateDB)
		continue
	}
	scheduler.ExecuteWave(wave, func(idx int, snap bal.StateSnapshot) (uint64, error) {
		return executeTxOnState(idx, snap)
	}, stateDB)
}
```

**Step: Run tests**

```
cd /projects/eth2030/pkg && go test ./bal/... ./core/... -run "TestScheduler|TestParallelProcessor" -v -timeout 60s
```

Expected: PASS with matching state roots.

**Step: Format & commit**

```bash
cd /projects/eth2030/pkg && go fmt ./bal/... ./core/...
git add pkg/bal/scheduler.go pkg/bal/scheduler_rollback_test.go \
        pkg/core/parallel_processor.go pkg/core/parallel_vs_sequential_test.go
git commit -m "feat(bal): speculative execution with conflict retry + parallel processor"
```

---

## EIP-7928 Spec Reference

> *Relevant excerpt from `refs/EIPs/EIPS/eip-7928.md`:*

```
## Motivation

This proposal enforces access lists at the block level, enabling:

- Parallel disk reads and transaction execution
- Parallel post-state root calculation
- State reconstruction without executing transactions
- Reduced execution time to `parallel IO + parallel EVM`

## Rationale

### BAL Design Choice

4. **Transaction independence**: 60-80% of transactions access disjoint storage slots,
   enabling effective parallelization. The remaining 20-40% can be parallelized by having
   post-transaction state diffs.

### Asynchronous Validation

BAL verification occurs alongside parallel IO and EVM operations without delaying block
processing.
```

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/bal/scheduler.go` | `BALScheduler` with `NewBALScheduler(workers, detector)`, `Schedule(bal)` returning `[]Wave`, `AssignWorkers(wave)`, `ExecuteSpeculative(wave, conflictSet)`, `ReExecute(results)` |
| `pkg/bal/parallel.go` | `ComputeParallelSets(bal) []ExecutionGroup` — greedy graph-coloring to find non-conflicting groups; `MaxParallelism(bal) int` |
| `pkg/bal/scheduler_pipeline.go` | `PipelineScheduler` with `PipelinePlan`, `PipelineStage`, `PipelineBatch`, `PipelineTask` — worker-pool pipeline scheduling |
| `pkg/bal/conflict_detector.go` | `BALConflictDetector` used by `BALScheduler.Schedule()` to build the dependency graph |
| `pkg/bal/types_extended.go` | `BuildDetailedEntries`, `DependencyGraph`, `ScheduleFromGraph`, `ConflictMatrix` — graph and scheduling primitives |

---

## Implementation Assessment

### Current Status

Partially implemented.

### Architecture Notes

The plan calls for implementing `BALScheduler.ExecuteWave(wave Wave, exec ExecutorFunc, state StateSnapshot) error` as a new method that runs non-conflicting transactions concurrently with goroutines, collects conflicts, and retries them serially. It also calls for a `pkg/core/parallel_processor.go` file containing `ProcessParallel()`.

The actual `BALScheduler` in `pkg/bal/scheduler.go` implements a different set of methods. `ExecuteSpeculative(wave, conflictSet)` is present, but it only simulates execution — it increments counters and marks tasks as rolled back based on a pre-supplied conflict set rather than actually executing transactions against a state. There is no real EVM invocation, no state snapshot interface, and no `ExecutorFunc` callback type. `ReExecute(results)` similarly simulates re-execution by hard-coding `gasUsed = 21000`.

The dependency graph and wave formation logic (`topoSort`, `buildWaves`) is correctly implemented in `scheduler.go`. `ComputeParallelSets` in `parallel.go` uses greedy graph coloring to identify non-conflicting transaction groups. `PipelineScheduler` in `scheduler_pipeline.go` provides a more elaborate worker-pool pipeline. These pieces form a solid scheduling foundation, but none of them perform actual EVM execution.

The plan's `StateSnapshot` interface (`Snapshot() int`, `RevertToSnapshot(int)`, `Commit() error`) and `ExecutorFunc` type do not exist in the codebase. The file `pkg/core/parallel_processor.go` does not exist.

### Gaps and Proposed Solutions

1. **`ExecuteWave` with real goroutines and state does not exist.** The current `ExecuteSpeculative` is a simulation stub that does not accept an `ExecutorFunc` or `StateSnapshot`. Solution: add `ExecuteWave(wave Wave, exec ExecutorFunc, state StateSnapshot) error` to `BALScheduler` with the goroutine-per-task pattern described in the plan. The conflict detection must happen at runtime (after execution) rather than being pre-supplied.

2. **`StateSnapshot` interface is not defined.** No interface exists for snapshotting/reverting state during speculative execution. Solution: define `StateSnapshot` in `pkg/bal/scheduler.go` (or a new `pkg/bal/snapshot.go`) matching `state.MemoryStateDB`'s existing `Snapshot()`/`RevertToSnapshot()` API. The `state.MemoryStateDB` in `pkg/core/state` already exposes these methods, so the interface can wrap them.

3. **`pkg/core/parallel_processor.go` does not exist.** There is no `ProcessParallel()` function. Solution: create the file with a `ProcessParallel(block, statedb, bal) (*ProcessResult, error)` function that calls `BALScheduler.Schedule()` to get waves, then calls `ExecuteWave()` per wave, and verifies the resulting state root matches sequential execution.

4. **Real conflict detection during parallel execution is absent.** The plan's `ExecuteWave` detects conflicts by catching errors; however, in a speculative model, the conflict must be detected by comparing accessed state versus declared BAL. Solution: implement runtime conflict detection by comparing the actual read/write set of each goroutine against the pre-execution BAL, returning `ErrConflict` when an undeclared dependency is observed.

5. **`pkg/core/parallel_vs_sequential_test.go` does not exist.** The test that validates identical state roots is absent. Solution: create the test file as a pre-condition to merging `ExecuteWave` to prevent regressions.
