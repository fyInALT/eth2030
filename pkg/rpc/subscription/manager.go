// manager.go implements SubscriptionManager which tracks installed filters
// and WebSocket subscriptions for the JSON-RPC server.
package rpcsub

import (
	"encoding/hex"
	"sync"
	"time"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
	rpcbackend "github.com/eth2030/eth2030/rpc/backend"
	"github.com/eth2030/eth2030/rpc/ethapi"
	rpcfilter "github.com/eth2030/eth2030/rpc/filter"
	rpctypes "github.com/eth2030/eth2030/rpc/types"
)

// SubType distinguishes the kind of WebSocket subscription.
type SubType = ethapi.SubType

// Re-export SubType constants from ethapi.
const (
	SubNewHeads  SubType = ethapi.SubNewHeads
	SubLogs      SubType = ethapi.SubLogs
	SubPendingTx SubType = ethapi.SubPendingTx
)

// Subscription represents an active WebSocket subscription.
type Subscription struct {
	ID    string
	Type  SubType
	Query rpcfilter.FilterQuery
	ch    chan interface{}
}

// Channel returns the notification channel for this subscription.
func (s *Subscription) Channel() <-chan interface{} {
	return s.ch
}

// subscriptionBufferSize is the channel buffer for subscription notifications.
const subscriptionBufferSize = 128

// filterTimeout is how long a filter lives without being polled.
const filterTimeout = 5 * time.Minute

// installedFilter is a stateful filter installed via eth_newFilter and friends.
type installedFilter struct {
	typ       rpcfilter.FilterType
	query     rpcfilter.FilterQuery
	lastPoll  time.Time
	blockLogs []*types.Log
	hashes    []types.Hash
	lastBlock uint64
}

// SubscriptionManager tracks installed filters and WebSocket subscriptions,
// providing incremental polling (eth_getFilterChanges), one-shot queries
// (eth_getLogs), and push-based notifications (eth_subscribe).
type SubscriptionManager struct {
	mu            sync.Mutex
	filters       map[string]*installedFilter
	subscriptions map[string]*Subscription
	backend       rpcbackend.Backend
	nextSeq       uint64
}

// NewSubscriptionManager creates a new subscription manager backed by
// the given chain backend.
func NewSubscriptionManager(backend rpcbackend.Backend) *SubscriptionManager {
	return &SubscriptionManager{
		filters:       make(map[string]*installedFilter),
		subscriptions: make(map[string]*Subscription),
		backend:       backend,
	}
}

// generateManagerID produces a unique hex filter ID using keccak256 over the
// current sequence number and timestamp.
func (sm *SubscriptionManager) generateManagerID() string {
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
func (sm *SubscriptionManager) NewLogFilter(query rpcfilter.FilterQuery) string {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	id := sm.generateManagerID()

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
		typ:       rpcfilter.LogFilter,
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

	id := sm.generateManagerID()

	lastBlock := uint64(0)
	header := sm.backend.CurrentHeader()
	if header != nil {
		lastBlock = header.Number.Uint64()
	}

	sm.filters[id] = &installedFilter{
		typ:       rpcfilter.BlockFilter,
		lastPoll:  time.Now(),
		lastBlock: lastBlock,
	}
	return id
}

// NewPendingTxFilter installs a pending transaction filter and returns its ID.
func (sm *SubscriptionManager) NewPendingTxFilter() string {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	id := sm.generateManagerID()
	sm.filters[id] = &installedFilter{
		typ:      rpcfilter.PendingTxFilter,
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
	case rpcfilter.LogFilter:
		return sm.pollLogs(f), true
	case rpcfilter.BlockFilter:
		return sm.pollBlocks(f), true
	case rpcfilter.PendingTxFilter:
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
	if !ok || f.typ != rpcfilter.LogFilter {
		return nil, false
	}
	f.lastPoll = time.Now()

	return sm.queryLogs(f.query), true
}

// QueryLogs performs a one-shot log query (eth_getLogs) without installing a filter.
func (sm *SubscriptionManager) QueryLogs(query rpcfilter.FilterQuery) []*types.Log {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.queryLogs(query)
}

// NotifyNewBlock pushes the block hash to all installed block filters.
func (sm *SubscriptionManager) NotifyNewBlock(hash types.Hash) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, f := range sm.filters {
		if f.typ == rpcfilter.BlockFilter {
			f.hashes = append(f.hashes, hash)
		}
	}
}

// NotifyPendingTx pushes a pending tx hash to all pending-tx filters.
func (sm *SubscriptionManager) NotifyPendingTx(hash types.Hash) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, f := range sm.filters {
		if f.typ == rpcfilter.PendingTxFilter {
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

// bloomMatchesQuery wraps rpcfilter.BloomMatchesFilterQuery for internal use.
func bloomMatchesQuery(bloom types.Bloom, query rpcfilter.FilterQuery) bool {
	return rpcfilter.BloomMatchesFilterQuery(bloom, query)
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
		h := sm.backend.HeaderByNumber(rpctypes.BlockNumber(blockNum))
		if h == nil {
			continue
		}
		blockHash := h.Hash()

		if !bloomMatchesQuery(h.Bloom, f.query) {
			continue
		}

		logs := sm.backend.GetLogs(blockHash)
		for _, log := range logs {
			if rpcfilter.MatchFilter(log, f.query) {
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
		h := sm.backend.HeaderByNumber(rpctypes.BlockNumber(blockNum))
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

func (sm *SubscriptionManager) queryLogs(query rpcfilter.FilterQuery) []*types.Log {
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
		h := sm.backend.HeaderByNumber(rpctypes.BlockNumber(blockNum))
		if h == nil {
			continue
		}
		blockHash := h.Hash()

		if !bloomMatchesQuery(h.Bloom, query) {
			continue
		}

		logs := sm.backend.GetLogs(blockHash)
		for _, log := range logs {
			if rpcfilter.MatchFilter(log, query) {
				result = append(result, log)
			}
		}
	}

	if result == nil {
		result = []*types.Log{}
	}
	return result
}

// ---------- WebSocket subscription methods ----------

// Subscribe creates a new WebSocket subscription and returns its ID.
func (sm *SubscriptionManager) Subscribe(subType SubType, query rpcfilter.FilterQuery) string {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	id := sm.generateManagerID()
	sub := &Subscription{
		ID:    id,
		Type:  subType,
		Query: query,
		ch:    make(chan interface{}, subscriptionBufferSize),
	}
	sm.subscriptions[id] = sub
	return id
}

// Unsubscribe removes a subscription by ID. Returns true if it existed.
func (sm *SubscriptionManager) Unsubscribe(id string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sub, ok := sm.subscriptions[id]
	if ok {
		close(sub.ch)
		delete(sm.subscriptions, id)
	}
	return ok
}

// GetSubscription returns a subscription by ID, or nil if not found.
func (sm *SubscriptionManager) GetSubscription(id string) *Subscription {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.subscriptions[id]
}

// SubscriptionCount returns the number of active subscriptions.
func (sm *SubscriptionManager) SubscriptionCount() int {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return len(sm.subscriptions)
}

// NotifyNewHead broadcasts a new block header to all "newHeads" subscribers.
func (sm *SubscriptionManager) NotifyNewHead(header *types.Header) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	formatted := rpctypes.FormatHeader(header)
	for _, sub := range sm.subscriptions {
		if sub.Type == SubNewHeads {
			select {
			case sub.ch <- formatted:
			default:
				// Drop if buffer is full.
			}
		}
	}
}

// NotifyLogs broadcasts matching logs to "logs" subscribers.
func (sm *SubscriptionManager) NotifyLogs(logs []*types.Log) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, sub := range sm.subscriptions {
		if sub.Type != SubLogs {
			continue
		}
		for _, log := range logs {
			if rpcfilter.MatchFilter(log, sub.Query) {
				formatted := rpctypes.FormatLog(log)
				select {
				case sub.ch <- formatted:
				default:
					// Drop if buffer is full.
				}
			}
		}
	}
}

// NotifyPendingTxHash broadcasts a pending transaction hash to all
// "newPendingTransactions" subscribers.
func (sm *SubscriptionManager) NotifyPendingTxHash(txHash types.Hash) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	hashStr := rpctypes.EncodeHash(txHash)
	for _, sub := range sm.subscriptions {
		if sub.Type == SubPendingTx {
			select {
			case sub.ch <- hashStr:
			default:
				// Drop if buffer is full.
			}
		}
	}
}
