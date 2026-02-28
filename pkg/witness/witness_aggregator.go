// witness_aggregator.go implements witness aggregation: merging multiple
// single-block witnesses into range witnesses for light clients and
// stateless verification. This enables efficient multi-block proof
// verification by deduplicating shared accounts and storage across blocks.
package witness

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/eth2030/eth2030/core/types"
)

// Aggregator errors.
var (
	ErrAggregatorEmpty    = errors.New("aggregator: no block witnesses added")
	ErrAggregatorGap      = errors.New("aggregator: gap in block range")
	ErrAggregatorNilBlock = errors.New("aggregator: nil witness")
	ErrAggregatorRange    = errors.New("aggregator: invalid range (from > to)")
	ErrAggregatorMissing  = errors.New("aggregator: block not found")
)

// WitnessAggregator merges multiple single-block witnesses into a range
// witness. Thread-safe for concurrent block addition.
type WitnessAggregator struct {
	mu       sync.RWMutex
	blocks   map[uint64]*BlockExecutionWitness
	minBlock uint64
	maxBlock uint64
	count    int
}

// NewWitnessAggregator creates an empty witness aggregator.
func NewWitnessAggregator() *WitnessAggregator {
	return &WitnessAggregator{
		blocks: make(map[uint64]*BlockExecutionWitness),
	}
}

// AddBlockWitness adds a single block's execution witness to the aggregator.
// Blocks may be added in any order.
func (wa *WitnessAggregator) AddBlockWitness(blockNum uint64, witness *BlockExecutionWitness) error {
	if witness == nil {
		return ErrAggregatorNilBlock
	}

	wa.mu.Lock()
	defer wa.mu.Unlock()

	wa.blocks[blockNum] = witness

	if wa.count == 0 {
		wa.minBlock = blockNum
		wa.maxBlock = blockNum
	} else {
		if blockNum < wa.minBlock {
			wa.minBlock = blockNum
		}
		if blockNum > wa.maxBlock {
			wa.maxBlock = blockNum
		}
	}
	wa.count++
	return nil
}

// WitnessRange represents the aggregated witness data for a contiguous
// range of blocks [From, To].
type WitnessRange struct {
	From      uint64
	To        uint64
	Witnesses []*BlockExecutionWitness
	// MergedPreState holds the union of all pre-state accounts across blocks.
	MergedPreState map[types.Address]*PreStateAccount
	// MergedCodes holds the union of all code entries across blocks.
	MergedCodes map[types.Hash][]byte
	// AllDiffs holds all state diffs from all blocks, in block order.
	AllDiffs []StateDiff
}

// BuildRangeWitness constructs a range witness for the given block range.
// Returns an error if the range is invalid or any block in the range is missing.
func (wa *WitnessAggregator) BuildRangeWitness(from, to uint64) (*WitnessRange, error) {
	if from > to {
		return nil, ErrAggregatorRange
	}

	wa.mu.RLock()
	defer wa.mu.RUnlock()

	if wa.count == 0 {
		return nil, ErrAggregatorEmpty
	}

	wr := &WitnessRange{
		From:           from,
		To:             to,
		MergedPreState: make(map[types.Address]*PreStateAccount),
		MergedCodes:    make(map[types.Hash][]byte),
	}

	for num := from; num <= to; num++ {
		bw, ok := wa.blocks[num]
		if !ok {
			return nil, fmt.Errorf("%w: block %d", ErrAggregatorGap, num)
		}
		wr.Witnesses = append(wr.Witnesses, bw)

		// Merge pre-state: keep the earliest pre-state for each account.
		for addr, psa := range bw.PreState {
			if _, exists := wr.MergedPreState[addr]; !exists {
				wr.MergedPreState[addr] = copyPreStateAccount(psa)
			} else {
				// Merge storage keys we haven't seen yet.
				existing := wr.MergedPreState[addr]
				for k, v := range psa.Storage {
					if _, seen := existing.Storage[k]; !seen {
						existing.Storage[k] = v
					}
				}
			}
		}

		// Merge codes.
		for hash, code := range bw.Codes {
			if _, exists := wr.MergedCodes[hash]; !exists {
				cp := make([]byte, len(code))
				copy(cp, code)
				wr.MergedCodes[hash] = cp
			}
		}

		// Append all state diffs.
		wr.AllDiffs = append(wr.AllDiffs, bw.StateDiffs...)
	}

	return wr, nil
}

// MinimalWitness is an optimized witness with deduplicated accounts and
// storage across the full block range. It represents the minimal set of
// pre-state data needed to verify the range.
type MinimalWitness struct {
	From     uint64
	To       uint64
	Accounts map[types.Address]*PreStateAccount
	Codes    map[types.Hash][]byte
	// FinalStorage holds the latest storage values for each address/key
	// after applying all diffs in the range.
	FinalStorage map[types.Address]map[types.Hash]types.Hash
}

// Minimize takes a WitnessRange and produces a MinimalWitness by deduplicating
// accounts and merging storage diffs so that only the final state of each
// slot is retained.
func (wa *WitnessAggregator) Minimize(wr *WitnessRange) *MinimalWitness {
	mw := &MinimalWitness{
		From:         wr.From,
		To:           wr.To,
		Accounts:     make(map[types.Address]*PreStateAccount),
		Codes:        make(map[types.Hash][]byte),
		FinalStorage: make(map[types.Address]map[types.Hash]types.Hash),
	}

	// Copy merged pre-state accounts (already deduplicated).
	for addr, psa := range wr.MergedPreState {
		mw.Accounts[addr] = copyPreStateAccount(psa)
	}

	// Copy codes (already deduplicated).
	for hash, code := range wr.MergedCodes {
		cp := make([]byte, len(code))
		copy(cp, code)
		mw.Codes[hash] = cp
	}

	// Walk all diffs and keep only the latest value for each storage slot.
	for _, diff := range wr.AllDiffs {
		addr := types.Address(diff.Address)
		if _, ok := mw.FinalStorage[addr]; !ok {
			mw.FinalStorage[addr] = make(map[types.Hash]types.Hash)
		}
		for _, sc := range diff.StorageChanges {
			key := types.Hash(sc.Key)
			mw.FinalStorage[addr][key] = types.Hash(sc.NewValue)
		}

		// Update balance/nonce to latest if changed.
		if diff.BalanceDiff.Changed {
			if psa, ok := mw.Accounts[addr]; ok {
				psa.Balance = make([]byte, len(diff.BalanceDiff.NewBalance))
				copy(psa.Balance, diff.BalanceDiff.NewBalance)
			}
		}
		if diff.NonceDiff.Changed {
			if psa, ok := mw.Accounts[addr]; ok {
				psa.Nonce = diff.NonceDiff.NewNonce
			}
		}
	}

	return mw
}

// Size returns the estimated byte size of a WitnessRange.
func (wa *WitnessAggregator) Size(wr *WitnessRange) int {
	size := 0

	// Pre-state accounts.
	for _, psa := range wr.MergedPreState {
		size += 20 // address
		size += len(psa.Balance)
		size += 8  // nonce
		size += 32 // code hash
		size += 1  // exists flag
		size += len(psa.Storage) * 64 // key + value
	}

	// Codes.
	for _, code := range wr.MergedCodes {
		size += 32 // hash key
		size += len(code)
	}

	// State diffs.
	for _, diff := range wr.AllDiffs {
		size += 20 // address
		if diff.BalanceDiff.Changed {
			size += len(diff.BalanceDiff.OldBalance) + len(diff.BalanceDiff.NewBalance)
		}
		if diff.NonceDiff.Changed {
			size += 16 // old + new nonce
		}
		size += len(diff.StorageChanges) * 96 // key + old + new
	}

	return size
}

// BlockCount returns the number of block witnesses currently held.
func (wa *WitnessAggregator) BlockCount() int {
	wa.mu.RLock()
	defer wa.mu.RUnlock()
	return wa.count
}

// PruneOlderThan removes all block witnesses with block number < blockNum.
// Returns the number of blocks pruned.
func (wa *WitnessAggregator) PruneOlderThan(blockNum uint64) int {
	wa.mu.Lock()
	defer wa.mu.Unlock()

	pruned := 0
	for num := range wa.blocks {
		if num < blockNum {
			delete(wa.blocks, num)
			pruned++
			wa.count--
		}
	}

	// Recalculate min/max.
	if wa.count == 0 {
		wa.minBlock = 0
		wa.maxBlock = 0
	} else {
		wa.minBlock = ^uint64(0)
		wa.maxBlock = 0
		for num := range wa.blocks {
			if num < wa.minBlock {
				wa.minBlock = num
			}
			if num > wa.maxBlock {
				wa.maxBlock = num
			}
		}
	}

	return pruned
}

// MergeAccessSets merges multiple boolean access sets into one.
// The result is the union of all input sets.
func MergeAccessSets(sets ...map[string]bool) map[string]bool {
	result := make(map[string]bool)
	for _, set := range sets {
		for k, v := range set {
			if v {
				result[k] = true
			}
		}
	}
	return result
}

// SortedAddresses returns the addresses from a WitnessRange's merged
// pre-state, sorted lexicographically.
func (wr *WitnessRange) SortedAddresses() []types.Address {
	addrs := make([]types.Address, 0, len(wr.MergedPreState))
	for addr := range wr.MergedPreState {
		addrs = append(addrs, addr)
	}
	sort.Slice(addrs, func(i, j int) bool {
		for b := 0; b < types.AddressLength; b++ {
			if addrs[i][b] != addrs[j][b] {
				return addrs[i][b] < addrs[j][b]
			}
		}
		return false
	})
	return addrs
}

// copyPreStateAccount creates a deep copy of a PreStateAccount.
func copyPreStateAccount(src *PreStateAccount) *PreStateAccount {
	dst := &PreStateAccount{
		Nonce:    src.Nonce,
		CodeHash: src.CodeHash,
		Exists:   src.Exists,
		Storage:  make(map[types.Hash]types.Hash, len(src.Storage)),
	}
	dst.Balance = make([]byte, len(src.Balance))
	copy(dst.Balance, src.Balance)
	for k, v := range src.Storage {
		dst.Storage[k] = v
	}
	return dst
}
