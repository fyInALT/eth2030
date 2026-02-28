// gigagas_scheduler.go implements a gigagas execution scheduler with
// work-stealing, conflict resolution, and parallelism control for the
// L1 strawmap gigagas (1 Ggas/sec) target. It builds on the existing
// GigagasExecutor and WorkStealingPool by adding fine-grained read/write
// set conflict detection, dependency graph construction, and greedy batch
// scheduling for maximal transaction parallelism.
package core

import (
	"sort"
	"sync"
)

// GigagasSchedulerConfig configures the GigagasScheduler. It controls how many
// parallel execution lanes are available, how large transaction batches
// can be, and how many times a conflicting transaction may be retried.
type GigagasSchedulerConfig struct {
	// MaxLanes is the maximum number of parallel execution lanes.
	MaxLanes int
	// BatchSize is the maximum number of transactions per batch.
	BatchSize int
	// ConflictRetryLimit is how many times a conflicting tx can be retried.
	ConflictRetryLimit int
}

// DefaultGigagasSchedulerConfig returns production defaults for the scheduler.
func DefaultGigagasSchedulerConfig() GigagasSchedulerConfig {
	return GigagasSchedulerConfig{
		MaxLanes:           16,
		BatchSize:          256,
		ConflictRetryLimit: 3,
	}
}

// WorkUnit represents a unit of parallel work: a single transaction with its
// predicted read and write sets, gas estimate, and retry metadata.
type WorkUnit struct {
	// Index is the original transaction index in the block.
	Index int
	// ReadSet is the set of state keys this transaction reads.
	ReadSet map[string]bool
	// WriteSet is the set of state keys this transaction writes.
	WriteSet map[string]bool
	// GasEstimate is the predicted gas consumption.
	GasEstimate uint64
	// RetryCount tracks how many times this unit has been retried.
	RetryCount int
}

// NewWorkUnit creates a WorkUnit with the given index and gas estimate.
func NewWorkUnit(index int, gasEstimate uint64) *WorkUnit {
	return &WorkUnit{
		Index:       index,
		ReadSet:     make(map[string]bool),
		WriteSet:    make(map[string]bool),
		GasEstimate: gasEstimate,
	}
}

// AddRead marks a key as read by this work unit.
func (wu *WorkUnit) AddRead(key string) {
	wu.ReadSet[key] = true
}

// AddWrite marks a key as written by this work unit.
func (wu *WorkUnit) AddWrite(key string) {
	wu.WriteSet[key] = true
}

// ConflictResolver detects and resolves read-write conflicts between work
// units. It maintains a registry of units and can compute conflict pairs,
// dependency graphs, and topological orderings.
type ConflictResolver struct {
	mu    sync.RWMutex
	units []*WorkUnit
}

// NewConflictResolver creates an empty conflict resolver.
func NewConflictResolver() *ConflictResolver {
	return &ConflictResolver{}
}

// AddUnit registers a work unit for conflict analysis.
func (cr *ConflictResolver) AddUnit(unit *WorkUnit) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.units = append(cr.units, unit)
}

// UnitCount returns the number of registered units.
func (cr *ConflictResolver) UnitCount() int {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return len(cr.units)
}

// HasConflict checks whether two work units have a read-write or
// write-write conflict. A conflict exists when one unit writes to a key
// that the other reads or writes.
func (cr *ConflictResolver) HasConflict(a, b *WorkUnit) bool {
	// Check a's writes vs b's reads and writes.
	for key := range a.WriteSet {
		if b.ReadSet[key] || b.WriteSet[key] {
			return true
		}
	}
	// Check b's writes vs a's reads.
	for key := range b.WriteSet {
		if a.ReadSet[key] {
			return true
		}
	}
	return false
}

// FindConflictingPairs returns all pairs of unit indices that conflict.
// Each pair [i, j] satisfies i < j.
func (cr *ConflictResolver) FindConflictingPairs() [][2]int {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	var pairs [][2]int
	for i := 0; i < len(cr.units); i++ {
		for j := i + 1; j < len(cr.units); j++ {
			if cr.HasConflict(cr.units[i], cr.units[j]) {
				pairs = append(pairs, [2]int{i, j})
			}
		}
	}
	return pairs
}

// BuildDependencyGraph returns an adjacency list where graph[i] contains
// the indices of units that unit i depends on (must execute after). A unit
// i depends on unit j (j < i) if they conflict and j appears earlier.
func (cr *ConflictResolver) BuildDependencyGraph() map[int][]int {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	graph := make(map[int][]int)
	for i := 0; i < len(cr.units); i++ {
		for j := 0; j < i; j++ {
			if cr.HasConflict(cr.units[i], cr.units[j]) {
				graph[i] = append(graph[i], j)
			}
		}
	}
	return graph
}

// ConflictCount returns the total number of conflicting pairs.
func (cr *ConflictResolver) ConflictCount() int {
	return len(cr.FindConflictingPairs())
}

// ExecutionLane represents an independent execution lane with its own
// state view and assigned work units. Lanes run in parallel; units
// within a lane run sequentially.
type ExecutionLane struct {
	// ID is the lane identifier.
	ID int
	// AssignedUnits holds the work unit indices assigned to this lane.
	AssignedUnits []int
	// TotalGas is the total estimated gas across all assigned units.
	TotalGas uint64
}

// NewExecutionLane creates a new execution lane with the given ID.
func NewExecutionLane(id int) *ExecutionLane {
	return &ExecutionLane{
		ID: id,
	}
}

// Assign adds a work unit index to this lane and accumulates its gas.
func (el *ExecutionLane) Assign(unitIndex int, gasEstimate uint64) {
	el.AssignedUnits = append(el.AssignedUnits, unitIndex)
	el.TotalGas += gasEstimate
}

// Load returns the total estimated gas load on this lane.
func (el *ExecutionLane) Load() uint64 {
	return el.TotalGas
}

// UnitCount returns the number of units assigned to this lane.
func (el *ExecutionLane) UnitCount() int {
	return len(el.AssignedUnits)
}

// GigagasScheduler manages parallel transaction execution with conflict
// detection, dependency-aware batch building, and lane assignment.
type GigagasScheduler struct {
	config GigagasSchedulerConfig
}

// NewGigagasScheduler creates a scheduler with the given configuration.
func NewGigagasScheduler(cfg GigagasSchedulerConfig) *GigagasScheduler {
	if cfg.MaxLanes <= 0 {
		cfg.MaxLanes = 16
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 256
	}
	if cfg.ConflictRetryLimit <= 0 {
		cfg.ConflictRetryLimit = 3
	}
	return &GigagasScheduler{config: cfg}
}

// DetectConflicts checks whether two work units have a read-write or
// write-write conflict.
func (gs *GigagasScheduler) DetectConflicts(a, b *WorkUnit) bool {
	for key := range a.WriteSet {
		if b.ReadSet[key] || b.WriteSet[key] {
			return true
		}
	}
	for key := range b.WriteSet {
		if a.ReadSet[key] {
			return true
		}
	}
	return false
}

// ScheduleTransactions takes transaction count and their read/write sets,
// builds work units, and returns batches of non-conflicting transaction
// indices. Each batch can be executed in parallel.
func (gs *GigagasScheduler) ScheduleTransactions(
	txCount int,
	readSets, writeSets []map[string]bool,
) [][]int {
	if txCount == 0 {
		return nil
	}

	units := make([]*WorkUnit, txCount)
	for i := 0; i < txCount; i++ {
		units[i] = NewWorkUnit(i, 21000)
		if i < len(readSets) && readSets[i] != nil {
			units[i].ReadSet = readSets[i]
		}
		if i < len(writeSets) && writeSets[i] != nil {
			units[i].WriteSet = writeSets[i]
		}
	}
	return gs.BuildExecutionPlan(units)
}

// BuildExecutionPlan takes a slice of work units and partitions them into
// batches where no two units within the same batch conflict. Uses a greedy
// coloring algorithm: for each unit, assign it to the first batch where it
// has no conflicts; if none exists, create a new batch.
func (gs *GigagasScheduler) BuildExecutionPlan(units []*WorkUnit) [][]int {
	if len(units) == 0 {
		return nil
	}

	// Sort units by gas estimate descending so that expensive txs are placed
	// first, reducing the chance of later txs creating new batches.
	sorted := make([]*WorkUnit, len(units))
	copy(sorted, units)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GasEstimate > sorted[j].GasEstimate
	})

	type batch struct {
		indices []int
		units   []*WorkUnit
	}

	var batches []batch

	for _, unit := range sorted {
		placed := false
		for bi := range batches {
			conflict := false
			for _, existing := range batches[bi].units {
				if gs.DetectConflicts(unit, existing) {
					conflict = true
					break
				}
			}
			if !conflict {
				batches[bi].indices = append(batches[bi].indices, unit.Index)
				batches[bi].units = append(batches[bi].units, unit)
				placed = true
				break
			}
		}
		if !placed {
			batches = append(batches, batch{
				indices: []int{unit.Index},
				units:   []*WorkUnit{unit},
			})
		}
	}

	result := make([][]int, len(batches))
	for i, b := range batches {
		// Sort indices within each batch for determinism.
		sort.Ints(b.indices)
		result[i] = b.indices
	}
	return result
}

// AssignToLanes distributes work units across execution lanes using a
// least-loaded-first strategy. Returns the lane assignments.
func (gs *GigagasScheduler) AssignToLanes(units []*WorkUnit) []*ExecutionLane {
	numLanes := gs.config.MaxLanes
	if len(units) < numLanes {
		numLanes = len(units)
	}
	if numLanes == 0 {
		return nil
	}

	lanes := make([]*ExecutionLane, numLanes)
	for i := 0; i < numLanes; i++ {
		lanes[i] = NewExecutionLane(i)
	}

	// Assign each unit to the lane with the smallest current gas load.
	for _, unit := range units {
		minIdx := 0
		for j := 1; j < len(lanes); j++ {
			if lanes[j].Load() < lanes[minIdx].Load() {
				minIdx = j
			}
		}
		lanes[minIdx].Assign(unit.Index, unit.GasEstimate)
	}

	return lanes
}

// MaxLanes returns the configured maximum lane count.
func (gs *GigagasScheduler) MaxLanes() int {
	return gs.config.MaxLanes
}

// BatchSize returns the configured batch size.
func (gs *GigagasScheduler) BatchSize() int {
	return gs.config.BatchSize
}

// ConflictRetryLimit returns the configured conflict retry limit.
func (gs *GigagasScheduler) ConflictRetryLimit() int {
	return gs.config.ConflictRetryLimit
}
