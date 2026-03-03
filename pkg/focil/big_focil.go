package focil

import (
	"sort"
	"sync"

	"github.com/eth2030/eth2030/core/types"
)

// Big FOCIL constants.
const (
	// NumPartitions is the number of sender-hex partitions (16 = one per hex char).
	NumPartitions = 16

	// DefaultCarryoverSlots is the max slots a tx can carry over before expiry.
	DefaultCarryoverSlots = 4
)

// BigFOCILConfig configures Big FOCIL with sender-hex partitioning.
type BigFOCILConfig struct {
	EnablePartitioning bool
	CarryoverSlots     uint64
}

// DefaultBigFOCILConfig returns a config with partitioning enabled and default carryover.
func DefaultBigFOCILConfig() *BigFOCILConfig {
	return &BigFOCILConfig{
		EnablePartitioning: true,
		CarryoverSlots:     DefaultCarryoverSlots,
	}
}

// SenderHexPartition returns the FOCIL partition index (0-15)
// based on the high nibble of the first byte of the sender address.
func SenderHexPartition(sender types.Address) uint8 {
	return sender[0] >> 4
}

// PartitionedList is a FOCIL inclusion list scoped to a single sender-hex partition.
type PartitionedList struct {
	Partition    uint8
	MemberIndex  int
	Transactions []types.Hash
	CarryoverTxs []types.Hash
	SlotNumber   uint64
}

// AssignPartitions maps each of 16 hex partitions to a committee member index.
// With fewer than 16 members, partitions wrap around.
func AssignPartitions(committeeSize int) map[uint8]int {
	assignments := make(map[uint8]int, NumPartitions)
	for i := uint8(0); i < NumPartitions; i++ {
		assignments[i] = int(i) % committeeSize
	}
	return assignments
}

// FilterByPartition returns only transactions whose sender matches the given partition.
func FilterByPartition(txs []*types.Transaction, partition uint8) []*types.Transaction {
	var result []*types.Transaction
	for _, tx := range txs {
		sender := tx.Sender()
		if sender == nil {
			continue
		}
		if SenderHexPartition(*sender) == partition {
			result = append(result, tx)
		}
	}
	return result
}

// BuildPartitionedList constructs a Big FOCIL inclusion list for a specific partition.
// It filters pending transactions by partition, prepends carryover txs, and
// builds the IL up to per-partition limits.
func BuildPartitionedList(pending []*types.Transaction, partition uint8, memberIndex int, slot uint64, carryover []CarryoverEntry) *PartitionedList {
	partitioned := FilterByPartition(pending, partition)

	// Build carryover hash list and prepend carried-over txs.
	var carryoverHashes []types.Hash
	var carryoverTxMap = make(map[types.Hash]bool)
	for _, c := range carryover {
		carryoverHashes = append(carryoverHashes, c.TxHash)
		carryoverTxMap[c.TxHash] = true
	}

	// Sort partitioned txs by gas price descending.
	sort.Slice(partitioned, func(i, j int) bool {
		pi := partitioned[i].GasPrice()
		pj := partitioned[j].GasPrice()
		if pi == nil || pj == nil {
			return pi != nil
		}
		return pi.Cmp(pj) > 0
	})

	var txHashes []types.Hash
	for _, tx := range partitioned {
		h := tx.Hash()
		if !carryoverTxMap[h] {
			txHashes = append(txHashes, h)
		}
	}

	return &PartitionedList{
		Partition:    partition,
		MemberIndex:  memberIndex,
		Transactions: txHashes,
		CarryoverTxs: carryoverHashes,
		SlotNumber:   slot,
	}
}

// TotalTransactions returns the total number of transactions (carryover + fresh).
func (pl *PartitionedList) TotalTransactions() int {
	return len(pl.CarryoverTxs) + len(pl.Transactions)
}

// --- Carryover Tracker ---

// CarryoverEntry tracks an un-included tx across slots.
type CarryoverEntry struct {
	TxHash    types.Hash
	Partition uint8
	FirstSlot uint64
	Priority  uint8 // increases each slot not included
}

// CarryoverTracker maintains un-included txs across slots with priority escalation.
type CarryoverTracker struct {
	mu       sync.Mutex
	pending  map[types.Hash]*CarryoverEntry
	maxSlots uint64
}

// NewCarryoverTracker creates a tracker with the given max carryover slots.
func NewCarryoverTracker(maxSlots uint64) *CarryoverTracker {
	if maxSlots == 0 {
		maxSlots = DefaultCarryoverSlots
	}
	return &CarryoverTracker{
		pending:  make(map[types.Hash]*CarryoverEntry),
		maxSlots: maxSlots,
	}
}

// AddUnincluded registers a tx that was not included in its slot.
// If already tracked, its priority is incremented.
func (ct *CarryoverTracker) AddUnincluded(txHash types.Hash, partition uint8, slot uint64) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if entry, exists := ct.pending[txHash]; exists {
		if entry.Priority < 255 {
			entry.Priority++
		}
		return
	}
	ct.pending[txHash] = &CarryoverEntry{
		TxHash:    txHash,
		Partition: partition,
		FirstSlot: slot,
		Priority:  1,
	}
}

// MarkIncluded removes a tx from the carryover tracker.
func (ct *CarryoverTracker) MarkIncluded(txHash types.Hash) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	delete(ct.pending, txHash)
}

// GetCarryover returns carryover entries for a partition, sorted by priority descending.
func (ct *CarryoverTracker) GetCarryover(partition uint8, currentSlot uint64) []CarryoverEntry {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	var result []CarryoverEntry
	for _, entry := range ct.pending {
		if entry.Partition != partition {
			continue
		}
		if currentSlot-entry.FirstSlot > ct.maxSlots {
			continue // expired
		}
		result = append(result, *entry)
	}

	// Sort by priority descending.
	sort.Slice(result, func(i, j int) bool {
		return result[i].Priority > result[j].Priority
	})
	return result
}

// Prune removes entries that have exceeded the maximum carryover slots.
func (ct *CarryoverTracker) Prune(currentSlot uint64) int {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	pruned := 0
	for hash, entry := range ct.pending {
		if currentSlot > entry.FirstSlot && currentSlot-entry.FirstSlot > ct.maxSlots {
			delete(ct.pending, hash)
			pruned++
		}
	}
	return pruned
}

// Len returns the number of pending carryover entries.
func (ct *CarryoverTracker) Len() int {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	return len(ct.pending)
}
