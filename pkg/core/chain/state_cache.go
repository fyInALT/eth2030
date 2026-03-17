package chain

import (
	"sync"

	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
)

const (
	// defaultMaxCachedStates is the default state snapshot cache capacity.
	// With TrieStateDB, each Dup() entry is cheap (~200 bytes — only the
	// struct; committedTrie/frozenAccounts/persistedStorage are shared via
	// pointer). 1024 entries cover blockscout-style backfill scenarios
	// (e.g. block 0 to current) without triggering expensive re-execution.
	defaultMaxCachedStates = 1024

	// stateSnapshotInterval determines how often we cache a state snapshot.
	// Every N blocks, a snapshot is taken to avoid re-execution from genesis.
	stateSnapshotInterval = 16
)

// stateCache caches state snapshots at regular block intervals to avoid
// expensive re-execution from genesis when building state for arbitrary blocks.
type stateCache struct {
	mu        sync.RWMutex
	maxSize   int                             // maximum number of snapshots to retain
	snapshots map[types.Hash]*stateCacheEntry // block hash → state snapshot
	order     []types.Hash                    // insertion order for eviction
	protected types.Hash                      // protected block hash (e.g., current head) - never evicted
}

type stateCacheEntry struct {
	blockNumber uint64
	stateDB     state.StateDB
}

func newStateCache(maxSize int) *stateCache {
	if maxSize <= 0 {
		maxSize = defaultMaxCachedStates
	}
	return &stateCache{
		maxSize:   maxSize,
		snapshots: make(map[types.Hash]*stateCacheEntry),
	}
}

// get returns a copy of the cached state for the given block hash.
func (sc *stateCache) get(blockHash types.Hash) (state.StateDB, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	entry, ok := sc.snapshots[blockHash]
	if !ok {
		return nil, false
	}
	return entry.stateDB.Dup(), true
}

// put stores a state snapshot for the given block.
func (sc *stateCache) put(blockHash types.Hash, blockNumber uint64, stateDB state.StateDB) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if _, ok := sc.snapshots[blockHash]; ok {
		return // already cached
	}

	// Evict oldest non-protected entries if at capacity.
	// We track how many entries we've checked to avoid infinite loops
	// if all entries are protected.
	checked := 0
	for len(sc.snapshots) >= sc.maxSize && checked < len(sc.order) {
		if len(sc.order) == 0 {
			break
		}
		oldest := sc.order[0]
		checked++
		// Skip protected entry (e.g., current head block state).
		// Only check if protected is set (non-zero hash).
		if sc.protected != (types.Hash{}) && oldest == sc.protected {
			// Move to end of order list and try next.
			sc.order = append(sc.order[1:], oldest)
			continue
		}
		sc.order = sc.order[1:]
		delete(sc.snapshots, oldest)
		// Reset checked counter since we successfully evicted an entry.
		checked = 0
	}

	sc.snapshots[blockHash] = &stateCacheEntry{
		blockNumber: blockNumber,
		stateDB:     stateDB.Dup(),
	}
	sc.order = append(sc.order, blockHash)
}

// closest finds the cached state snapshot closest to (but not after) the target
// block number. Returns the state copy, the block number it corresponds to,
// and whether a match was found.
func (sc *stateCache) closest(targetNumber uint64) (state.StateDB, uint64, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	var best *stateCacheEntry
	for _, entry := range sc.snapshots {
		if entry.blockNumber <= targetNumber {
			if best == nil || entry.blockNumber > best.blockNumber {
				best = entry
			}
		}
	}
	if best == nil {
		return nil, 0, false
	}
	return best.stateDB.Dup(), best.blockNumber, true
}

// remove deletes a cached state entry (e.g. during reorg).
func (sc *stateCache) remove(blockHash types.Hash) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	delete(sc.snapshots, blockHash)
	// Clean order list.
	for i := 0; i < len(sc.order); i++ {
		if sc.order[i] == blockHash {
			sc.order = append(sc.order[:i], sc.order[i+1:]...)
			break
		}
	}
}

// clear removes all cached states except the protected entry (e.g. after a major reorg).
func (sc *stateCache) clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	// Preserve protected entry if it exists.
	if sc.protected != (types.Hash{}) {
		if entry, ok := sc.snapshots[sc.protected]; ok {
			sc.snapshots = map[types.Hash]*stateCacheEntry{sc.protected: entry}
			sc.order = []types.Hash{sc.protected}
			return
		}
	}
	sc.snapshots = make(map[types.Hash]*stateCacheEntry)
	sc.order = nil
}

// protect marks a block hash as protected from eviction.
// This is used to ensure the current head block's state is never evicted,
// preventing expensive re-execution during payload building.
func (sc *stateCache) protect(blockHash types.Hash) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.protected = blockHash
}

// getProtected returns the currently protected block hash.
func (sc *stateCache) getProtected() types.Hash {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.protected
}
