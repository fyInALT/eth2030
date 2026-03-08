// subscription.go defines SubscriptionManager which tracks installed filters
// and WebSocket subscriptions. FilterType and FilterQuery are re-exported
// from rpc/filter for consistency.
package rpc

import (
	"encoding/hex"
	"sync"
	"time"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
	rpcfilter "github.com/eth2030/eth2030/rpc/filter"
)

// Re-export filter types from rpc/filter.
type (
	FilterType  = rpcfilter.FilterType
	FilterQuery = rpcfilter.FilterQuery
)

// Re-export FilterType constants.
const (
	LogFilter       = rpcfilter.LogFilter
	BlockFilter     = rpcfilter.BlockFilter
	PendingTxFilter = rpcfilter.PendingTxFilter
)

// MatchFilter re-exports rpcfilter.MatchFilter.
var MatchFilter = rpcfilter.MatchFilter

// bloomMatchesQuery is a package-level helper wrapping
// rpcfilter.BloomMatchesFilterQuery for internal use.
func bloomMatchesQuery(bloom types.Bloom, query FilterQuery) bool {
	return rpcfilter.BloomMatchesFilterQuery(bloom, query)
}

// filterTimeout is how long a filter lives without being polled.
const filterTimeout = 5 * time.Minute

// installedFilter is a stateful filter installed via eth_newFilter and friends.
type installedFilter struct {
	typ       FilterType
	query     FilterQuery
	lastPoll  time.Time
	blockLogs []*types.Log // accumulated log results (LogFilter)
	hashes    []types.Hash // accumulated block or tx hashes
	lastBlock uint64       // last scanned block (for incremental polls)
}

// SubscriptionManager tracks installed filters and WebSocket subscriptions,
// providing incremental polling (eth_getFilterChanges), one-shot queries
// (eth_getLogs), and push-based notifications (eth_subscribe).
type SubscriptionManager struct {
	mu            sync.Mutex
	filters       map[string]*installedFilter
	subscriptions map[string]*Subscription
	backend       Backend
	nextSeq       uint64 // monotonic counter to make IDs unique
}

// NewSubscriptionManager creates a new subscription manager backed by
// the given chain backend.
func NewSubscriptionManager(backend Backend) *SubscriptionManager {
	return &SubscriptionManager{
		filters:       make(map[string]*installedFilter),
		subscriptions: make(map[string]*Subscription),
		backend:       backend,
	}
}

// generateID produces a unique hex filter ID using keccak256 over the
// current sequence number and timestamp.
func (sm *SubscriptionManager) generateID() string {
	sm.nextSeq++
	buf := make([]byte, 16)
	seq := sm.nextSeq
	ts := uint64(time.Now().UnixNano())
	for i := 0; i < 8; i++ {
		buf[i] = byte(seq >> (8 * i))
		buf[8+i] = byte(ts >> (8 * i))
	}
	h := crypto.Keccak256(buf)
	return "0x" + hex.EncodeToString(h[:16])
}

// NewLogFilter installs a log filter and returns its ID.
func (sm *SubscriptionManager) NewLogFilter(query FilterQuery) string {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	id := sm.generateID()

	lastBlock := uint64(0)
	if query.FromBlock != nil {
		if *query.FromBlock > 0 {
			lastBlock = *query.FromBlock - 1
		}
	} else {
		header := sm.backend.CurrentHeader()
		if header != nil {
			lastBlock = header.Number.Uint64()
		}
	}

	sm.filters[id] = &installedFilter{
		typ:       LogFilter,
		query:     query,
		lastPoll:  time.Now(),
		lastBlock: lastBlock,
	}
	return id
}

// NewBlockFilter installs a block filter and returns its ID.
func (sm *SubscriptionManager) NewBlockFilter() string {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	id := sm.generateID()

	lastBlock := uint64(0)
	header := sm.backend.CurrentHeader()
	if header != nil {
		lastBlock = header.Number.Uint64()
	}

	sm.filters[id] = &installedFilter{
		typ:       BlockFilter,
		lastPoll:  time.Now(),
		lastBlock: lastBlock,
	}
	return id
}

// NewPendingTxFilter installs a pending transaction filter and returns its ID.
func (sm *SubscriptionManager) NewPendingTxFilter() string {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	id := sm.generateID()
	sm.filters[id] = &installedFilter{
		typ:      PendingTxFilter,
		lastPoll: time.Now(),
	}
	return id
}

// Uninstall removes a filter. Returns true if the filter existed.
func (sm *SubscriptionManager) Uninstall(id string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	_, ok := sm.filters[id]
	if ok {
		delete(sm.filters, id)
	}
	return ok
}

// GetFilterChanges returns new results since the last poll for the given filter.
func (sm *SubscriptionManager) GetFilterChanges(id string) (interface{}, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	f, ok := sm.filters[id]
	if !ok {
		return nil, false
	}
	f.lastPoll = time.Now()

	switch f.typ {
	case LogFilter:
		return sm.pollLogs(f), true
	case BlockFilter:
		return sm.pollBlocks(f), true
	case PendingTxFilter:
		result := f.hashes
		f.hashes = nil
		if result == nil {
			result = []types.Hash{}
		}
		return result, true
	}
	return nil, false
}

// GetFilterLogs returns all logs matching a log filter's criteria from scratch.
func (sm *SubscriptionManager) GetFilterLogs(id string) ([]*types.Log, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	f, ok := sm.filters[id]
	if !ok || f.typ != LogFilter {
		return nil, false
	}
	f.lastPoll = time.Now()

	return sm.queryLogs(f.query), true
}

// QueryLogs performs a one-shot log query (eth_getLogs) without installing a filter.
func (sm *SubscriptionManager) QueryLogs(query FilterQuery) []*types.Log {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.queryLogs(query)
}

// NotifyNewBlock pushes the block hash to all installed block filters.
func (sm *SubscriptionManager) NotifyNewBlock(hash types.Hash) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, f := range sm.filters {
		if f.typ == BlockFilter {
			f.hashes = append(f.hashes, hash)
		}
	}
}

// NotifyPendingTx pushes a pending tx hash to all pending-tx filters.
func (sm *SubscriptionManager) NotifyPendingTx(hash types.Hash) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, f := range sm.filters {
		if f.typ == PendingTxFilter {
			f.hashes = append(f.hashes, hash)
		}
	}
}

// CleanupStale removes filters that have not been polled within filterTimeout.
func (sm *SubscriptionManager) CleanupStale() int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	removed := 0
	now := time.Now()
	for id, f := range sm.filters {
		if now.Sub(f.lastPoll) > filterTimeout {
			delete(sm.filters, id)
			removed++
		}
	}
	return removed
}

// FilterCount returns the number of currently installed filters.
func (sm *SubscriptionManager) FilterCount() int {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return len(sm.filters)
}

// ---------- internal helpers ----------

func (sm *SubscriptionManager) pollLogs(f *installedFilter) []*types.Log {
	header := sm.backend.CurrentHeader()
	if header == nil {
		return []*types.Log{}
	}
	currentNum := header.Number.Uint64()

	startBlock := f.lastBlock + 1
	endBlock := currentNum

	if f.query.ToBlock != nil && *f.query.ToBlock < endBlock {
		endBlock = *f.query.ToBlock
	}

	var result []*types.Log
	for blockNum := startBlock; blockNum <= endBlock; blockNum++ {
		h := sm.backend.HeaderByNumber(BlockNumber(blockNum))
		if h == nil {
			continue
		}
		blockHash := h.Hash()

		if !bloomMatchesQuery(h.Bloom, f.query) {
			continue
		}

		logs := sm.backend.GetLogs(blockHash)
		for _, log := range logs {
			if MatchFilter(log, f.query) {
				result = append(result, log)
			}
		}
	}

	f.lastBlock = endBlock
	if result == nil {
		result = []*types.Log{}
	}
	return result
}

func (sm *SubscriptionManager) pollBlocks(f *installedFilter) []types.Hash {
	header := sm.backend.CurrentHeader()
	if header == nil {
		return []types.Hash{}
	}
	currentNum := header.Number.Uint64()

	result := f.hashes
	f.hashes = nil

	for blockNum := f.lastBlock + 1; blockNum <= currentNum; blockNum++ {
		h := sm.backend.HeaderByNumber(BlockNumber(blockNum))
		if h == nil {
			continue
		}
		result = append(result, h.Hash())
	}

	f.lastBlock = currentNum
	if result == nil {
		result = []types.Hash{}
	}
	return result
}

func (sm *SubscriptionManager) queryLogs(query FilterQuery) []*types.Log {
	header := sm.backend.CurrentHeader()
	if header == nil {
		return []*types.Log{}
	}
	currentNum := header.Number.Uint64()

	fromBlock := uint64(0)
	toBlock := currentNum

	if query.FromBlock != nil {
		fromBlock = *query.FromBlock
	}
	if query.ToBlock != nil {
		toBlock = *query.ToBlock
	}

	var result []*types.Log
	for blockNum := fromBlock; blockNum <= toBlock; blockNum++ {
		h := sm.backend.HeaderByNumber(BlockNumber(blockNum))
		if h == nil {
			continue
		}
		blockHash := h.Hash()

		if !bloomMatchesQuery(h.Bloom, query) {
			continue
		}

		logs := sm.backend.GetLogs(blockHash)
		for _, log := range logs {
			if MatchFilter(log, query) {
				result = append(result, log)
			}
		}
	}

	if result == nil {
		result = []*types.Log{}
	}
	return result
}
