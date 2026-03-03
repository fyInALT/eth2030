package core

import (
	"sort"
	"sync"

	"github.com/eth2030/eth2030/bal"
	"github.com/eth2030/eth2030/core/types"
)

// TxGroup is a set of non-conflicting transactions that can be
// assigned to a single builder for parallel execution.
type TxGroup struct {
	Transactions []*types.Transaction
	GroupID      int
	AccessSet    map[types.Address]map[types.Hash]bool
}

// DependencyGraph builds a DAG of transactions based on state access conflicts.
// Transactions with conflicting state accesses are connected by edges.
type DependencyGraph struct {
	mu        sync.RWMutex
	txs       []*types.Transaction
	conflicts map[int][]int // tx index -> conflicting tx indices
	readSets  map[int]map[types.Address]map[types.Hash]bool
	writeSets map[int]map[types.Address]map[types.Hash]bool
}

// NewDependencyGraph builds a dependency graph from transactions and their BAL.
func NewDependencyGraph(txs []*types.Transaction, accessList *bal.BlockAccessList) *DependencyGraph {
	dg := &DependencyGraph{
		txs:       txs,
		conflicts: make(map[int][]int),
		readSets:  make(map[int]map[types.Address]map[types.Hash]bool),
		writeSets: make(map[int]map[types.Address]map[types.Hash]bool),
	}

	if accessList == nil {
		return dg
	}

	// Build read/write sets per transaction from BAL entries.
	for _, entry := range accessList.Entries {
		// AccessIndex 0 = pre-execution, 1..n = tx index (1-based).
		if entry.AccessIndex == 0 {
			continue
		}
		txIdx := int(entry.AccessIndex) - 1
		if txIdx < 0 || txIdx >= len(txs) {
			continue
		}

		// Initialize maps.
		if dg.readSets[txIdx] == nil {
			dg.readSets[txIdx] = make(map[types.Address]map[types.Hash]bool)
		}
		if dg.writeSets[txIdx] == nil {
			dg.writeSets[txIdx] = make(map[types.Address]map[types.Hash]bool)
		}

		addr := entry.Address

		// Storage reads.
		if dg.readSets[txIdx][addr] == nil {
			dg.readSets[txIdx][addr] = make(map[types.Hash]bool)
		}
		for _, sr := range entry.StorageReads {
			dg.readSets[txIdx][addr][sr.Slot] = true
		}

		// Storage writes.
		if dg.writeSets[txIdx][addr] == nil {
			dg.writeSets[txIdx][addr] = make(map[types.Hash]bool)
		}
		for _, sc := range entry.StorageChanges {
			dg.writeSets[txIdx][addr][sc.Slot] = true
		}

		// Balance/nonce/code changes count as writes to a sentinel slot.
		if entry.BalanceChange != nil || entry.NonceChange != nil || entry.CodeChange != nil {
			sentinel := types.Hash{0xFF} // sentinel for account-level changes
			dg.writeSets[txIdx][addr][sentinel] = true
		}
	}

	// Detect conflicts: read-write, write-read, write-write.
	for i := 0; i < len(txs); i++ {
		for j := i + 1; j < len(txs); j++ {
			if dg.hasConflict(i, j) {
				dg.conflicts[i] = append(dg.conflicts[i], j)
				dg.conflicts[j] = append(dg.conflicts[j], i)
			}
		}
	}

	return dg
}

// hasConflict returns true if transactions i and j have conflicting state accesses.
func (dg *DependencyGraph) hasConflict(i, j int) bool {
	// Check write-write conflicts.
	for addr, slotsI := range dg.writeSets[i] {
		if slotsJ, ok := dg.writeSets[j][addr]; ok {
			for slot := range slotsI {
				if slotsJ[slot] {
					return true
				}
			}
		}
	}

	// Check read-write conflicts (i reads, j writes).
	for addr, slotsI := range dg.readSets[i] {
		if slotsJ, ok := dg.writeSets[j][addr]; ok {
			for slot := range slotsI {
				if slotsJ[slot] {
					return true
				}
			}
		}
	}

	// Check write-read conflicts (i writes, j reads).
	for addr, slotsI := range dg.writeSets[i] {
		if slotsJ, ok := dg.readSets[j][addr]; ok {
			for slot := range slotsI {
				if slotsJ[slot] {
					return true
				}
			}
		}
	}

	return false
}

// Partition splits transactions into non-conflicting groups using graph coloring.
// maxGroups limits the number of groups; 0 means unlimited.
func (dg *DependencyGraph) Partition(maxGroups int) []TxGroup {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	if len(dg.txs) == 0 {
		return nil
	}

	// Greedy graph coloring.
	colors := make([]int, len(dg.txs))
	for i := range colors {
		colors[i] = -1 // uncolored
	}

	maxColor := 0
	for i := 0; i < len(dg.txs); i++ {
		// Find colors used by neighbors.
		used := make(map[int]bool)
		for _, j := range dg.conflicts[i] {
			if colors[j] >= 0 {
				used[colors[j]] = true
			}
		}

		// Assign smallest available color.
		color := 0
		for used[color] {
			color++
		}

		// Enforce max groups limit.
		if maxGroups > 0 && color >= maxGroups {
			color = 0 // fall back to first group
		}

		colors[i] = color
		if color > maxColor {
			maxColor = color
		}
	}

	// Build groups from colors.
	groupMap := make(map[int]*TxGroup)
	for i, c := range colors {
		if _, ok := groupMap[c]; !ok {
			groupMap[c] = &TxGroup{
				GroupID:   c,
				AccessSet: make(map[types.Address]map[types.Hash]bool),
			}
		}
		groupMap[c].Transactions = append(groupMap[c].Transactions, dg.txs[i])

		// Merge access sets.
		for addr, slots := range dg.writeSets[i] {
			if groupMap[c].AccessSet[addr] == nil {
				groupMap[c].AccessSet[addr] = make(map[types.Hash]bool)
			}
			for slot := range slots {
				groupMap[c].AccessSet[addr][slot] = true
			}
		}
	}

	// Convert to sorted slice.
	groups := make([]TxGroup, 0, len(groupMap))
	for _, g := range groupMap {
		groups = append(groups, *g)
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].GroupID < groups[j].GroupID
	})

	return groups
}

// ConflictCount returns the total number of conflict edges in the graph.
func (dg *DependencyGraph) ConflictCount() int {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	total := 0
	for _, c := range dg.conflicts {
		total += len(c)
	}
	return total / 2 // each edge counted twice
}

// IsLocal returns true if the transaction is a LocalTx type.
func IsLocal(tx *types.Transaction) bool {
	return types.IsLocalTx(tx)
}

// ClassifyTransactions splits transactions into local (scope-hinted) and global groups.
func ClassifyTransactions(txs []*types.Transaction) (local, global []*types.Transaction) {
	for _, tx := range txs {
		if types.IsLocalTx(tx) {
			local = append(local, tx)
		} else {
			global = append(global, tx)
		}
	}
	return
}
