// lru_cache.go implements an LRU cache for recently-built execution
// payloads, keyed by PayloadID. It avoids redundant computation when the
// consensus layer requests the same payload multiple times.
//
// The cache is implemented using the actor pattern: all operations are
// processed sequentially through a message channel, eliminating lock contention.
package payload

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/eth2030/eth2030/engine/actor"
)

// Default maximum entries for the payload LRU cache.
const DefaultLRUCacheMaxEntries = 64

// LRUCachePayload represents an execution payload stored in the LRU cache.
type LRUCachePayload struct {
	ParentHash    [32]byte
	FeeRecipient  [20]byte
	StateRoot     [32]byte
	ReceiptsRoot  [32]byte
	BlockNumber   uint64
	GasLimit      uint64
	GasUsed       uint64
	Timestamp     uint64
	BaseFeePerGas uint64
	BlockHash     [32]byte
	Transactions  [][]byte
}

// LRUCacheStats tracks hit/miss/eviction statistics for the payload cache.
type LRUCacheStats struct {
	Hits      uint64
	Misses    uint64
	Evictions uint64
}

// lruEntry wraps a payload with its key and doubly-linked list pointers
// for O(1) LRU tracking.
type lruEntry struct {
	id      PayloadID
	payload *LRUCachePayload
	prev    *lruEntry
	next    *lruEntry
}

// --- Message types for actor communication ---

// lruGetMsg retrieves a payload by ID.
type lruGetMsg struct {
	actor.BaseMessage
	ID PayloadID
}

// lruGetByHashMsg retrieves a payload by block hash.
type lruGetByHashMsg struct {
	actor.BaseMessage
	Hash [32]byte
}

// lruPutMsg stores a payload.
type lruPutMsg struct {
	actor.BaseMessage
	ID      PayloadID
	Payload *LRUCachePayload
}

// lruRemoveMsg removes a payload.
type lruRemoveMsg struct {
	actor.BaseMessage
	ID PayloadID
}

// lruLenMsg returns the cache length.
type lruLenMsg struct {
	actor.BaseMessage
}

// lruClearMsg clears the cache.
type lruClearMsg struct {
	actor.BaseMessage
}

// --- Actor implementation ---

// PayloadLRUCacheActor is the actor-based implementation of the payload LRU cache.
// It processes all operations sequentially through its inbox channel.
type PayloadLRUCacheActor struct {
	maxEntries int
	entries    map[PayloadID]*lruEntry
	hashIndex  map[[32]byte]PayloadID
	head, tail *lruEntry

	// Inbox receives all cache operations.
	inbox chan any

	// Statistics tracked atomically for reads without messaging.
	hits, misses, evictions atomic.Uint64
}

// NewPayloadLRUCacheActor creates a new actor-based LRU cache.
func NewPayloadLRUCacheActor(maxEntries int) *PayloadLRUCacheActor {
	if maxEntries <= 0 {
		maxEntries = DefaultLRUCacheMaxEntries
	}
	return &PayloadLRUCacheActor{
		maxEntries: maxEntries,
		entries:    make(map[PayloadID]*lruEntry),
		hashIndex:  make(map[[32]byte]PayloadID),
		inbox:      make(chan any, 64), // buffered for throughput
	}
}

// Run implements actor.Actor. It processes messages until ctx is cancelled.
func (a *PayloadLRUCacheActor) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-a.inbox:
			a.handleMessage(msg)
		}
	}
}

// Inbox returns the actor's message channel.
func (a *PayloadLRUCacheActor) Inbox() chan<- any {
	return a.inbox
}

func (a *PayloadLRUCacheActor) handleMessage(msg any) {
	switch m := msg.(type) {
	case *lruGetMsg:
		entry, ok := a.entries[m.ID]
		if ok {
			a.hits.Add(1)
			a.moveToFront(entry)
			m.Reply() <- actor.Reply{Result: entry.payload}
		} else {
			a.misses.Add(1)
			m.Reply() <- actor.Reply{}
		}

	case *lruGetByHashMsg:
		id, ok := a.hashIndex[m.Hash]
		if !ok {
			a.misses.Add(1)
			m.Reply() <- actor.Reply{}
			return
		}
		entry, ok := a.entries[id]
		if !ok {
			// Stale index entry; clean up.
			delete(a.hashIndex, m.Hash)
			a.misses.Add(1)
			m.Reply() <- actor.Reply{}
			return
		}
		a.hits.Add(1)
		a.moveToFront(entry)
		m.Reply() <- actor.Reply{Result: entry.payload}

	case *lruPutMsg:
		err := a.put(m.ID, m.Payload)
		m.Reply() <- actor.Reply{Error: err}

	case *lruRemoveMsg:
		ok := a.remove(m.ID)
		m.Reply() <- actor.Reply{Result: ok}

	case *lruLenMsg:
		m.Reply() <- actor.Reply{Result: len(a.entries)}

	case *lruClearMsg:
		a.clear()
		m.Reply() <- actor.Reply{}
	}
}

// put stores a payload (internal, no lock).
func (a *PayloadLRUCacheActor) put(id PayloadID, p *LRUCachePayload) error {
	if p == nil {
		return errors.New("payload_lru_cache: nil payload")
	}

	// If key already exists, update in place and move to front.
	if entry, ok := a.entries[id]; ok {
		// Remove old block hash index entry if hash changed.
		if entry.payload.BlockHash != p.BlockHash {
			delete(a.hashIndex, entry.payload.BlockHash)
		}
		entry.payload = p
		a.hashIndex[p.BlockHash] = id
		a.moveToFront(entry)
		return nil
	}

	// Evict if at capacity.
	if len(a.entries) >= a.maxEntries {
		a.evictLRU()
	}

	entry := &lruEntry{id: id, payload: p}
	a.entries[id] = entry
	a.hashIndex[p.BlockHash] = id
	a.pushFront(entry)
	return nil
}

// remove removes a payload (internal, no lock).
func (a *PayloadLRUCacheActor) remove(id PayloadID) bool {
	entry, ok := a.entries[id]
	if !ok {
		return false
	}
	a.removeEntry(entry)
	return true
}

// clear removes all entries (internal, no lock).
func (a *PayloadLRUCacheActor) clear() {
	a.entries = make(map[PayloadID]*lruEntry)
	a.hashIndex = make(map[[32]byte]PayloadID)
	a.head = nil
	a.tail = nil
	a.hits.Store(0)
	a.misses.Store(0)
	a.evictions.Store(0)
}

// pushFront inserts an entry at the front (most-recently-used) of the list.
func (a *PayloadLRUCacheActor) pushFront(entry *lruEntry) {
	entry.prev = nil
	entry.next = a.head
	if a.head != nil {
		a.head.prev = entry
	}
	a.head = entry
	if a.tail == nil {
		a.tail = entry
	}
}

// moveToFront moves an existing entry to the front of the list.
func (a *PayloadLRUCacheActor) moveToFront(entry *lruEntry) {
	if a.head == entry {
		return // Already at front.
	}
	a.detach(entry)
	a.pushFront(entry)
}

// detach removes an entry from the linked list without deleting it from the map.
func (a *PayloadLRUCacheActor) detach(entry *lruEntry) {
	if entry.prev != nil {
		entry.prev.next = entry.next
	} else {
		a.head = entry.next
	}
	if entry.next != nil {
		entry.next.prev = entry.prev
	} else {
		a.tail = entry.prev
	}
	entry.prev = nil
	entry.next = nil
}

// removeEntry removes an entry from both the linked list and the maps.
func (a *PayloadLRUCacheActor) removeEntry(entry *lruEntry) {
	a.detach(entry)
	delete(a.entries, entry.id)
	delete(a.hashIndex, entry.payload.BlockHash)
}

// evictLRU removes the least-recently-used entry (the tail).
func (a *PayloadLRUCacheActor) evictLRU() {
	if a.tail == nil {
		return
	}
	a.removeEntry(a.tail)
	a.evictions.Add(1)
}

// Stats returns cache statistics (atomic, no messaging needed).
func (a *PayloadLRUCacheActor) Stats() *LRUCacheStats {
	return &LRUCacheStats{
		Hits:      a.hits.Load(),
		Misses:    a.misses.Load(),
		Evictions: a.evictions.Load(),
	}
}

// --- Compatibility wrapper ---

// PayloadLRUCache is a thread-safe LRU cache for execution payloads.
// It wraps PayloadLRUCacheActor for API compatibility with existing code.
type PayloadLRUCache struct {
	actor   *PayloadLRUCacheActor
	ctx     context.Context
	cancel  context.CancelFunc
	timeout time.Duration
}

// NewPayloadLRUCache creates a new payload LRU cache with actor backend.
func NewPayloadLRUCache(maxEntries int) *PayloadLRUCache {
	ctx, cancel := context.WithCancel(context.Background())
	act := NewPayloadLRUCacheActor(maxEntries)
	go act.Run(ctx)

	return &PayloadLRUCache{
		actor:   act,
		ctx:     ctx,
		cancel:  cancel,
		timeout: actor.DefaultTimeout,
	}
}

// Put stores a payload in the cache under the given PayloadID.
func (c *PayloadLRUCache) Put(id PayloadID, p *LRUCachePayload) error {
	msg, replyCh := actor.NewBaseMessage()
	if err := actor.Send[any](c.actor.Inbox(), &lruPutMsg{BaseMessage: msg, ID: id, Payload: p}, c.timeout); err != nil {
		return err
	}
	_, err := actor.CallResult[struct{}](replyCh, c.timeout)
	return err
}

// Get retrieves a payload by PayloadID.
func (c *PayloadLRUCache) Get(id PayloadID) (*LRUCachePayload, bool) {
	msg, replyCh := actor.NewBaseMessage()
	if err := actor.Send[any](c.actor.Inbox(), &lruGetMsg{BaseMessage: msg, ID: id}, c.timeout); err != nil {
		return nil, false
	}
	result, err := actor.CallResult[*LRUCachePayload](replyCh, c.timeout)
	if err != nil || result == nil {
		return nil, false
	}
	return result, true
}

// GetByBlockHash retrieves a payload by its BlockHash field.
func (c *PayloadLRUCache) GetByBlockHash(hash [32]byte) (*LRUCachePayload, bool) {
	msg, replyCh := actor.NewBaseMessage()
	if err := actor.Send[any](c.actor.Inbox(), &lruGetByHashMsg{BaseMessage: msg, Hash: hash}, c.timeout); err != nil {
		return nil, false
	}
	result, err := actor.CallResult[*LRUCachePayload](replyCh, c.timeout)
	if err != nil || result == nil {
		return nil, false
	}
	return result, true
}

// Remove removes a payload by PayloadID.
func (c *PayloadLRUCache) Remove(id PayloadID) bool {
	msg, replyCh := actor.NewBaseMessage()
	if err := actor.Send[any](c.actor.Inbox(), &lruRemoveMsg{BaseMessage: msg, ID: id}, c.timeout); err != nil {
		return false
	}
	result, err := actor.CallResult[bool](replyCh, c.timeout)
	if err != nil {
		return false
	}
	return result
}

// Len returns the number of entries currently in the cache.
func (c *PayloadLRUCache) Len() int {
	msg, replyCh := actor.NewBaseMessage()
	if err := actor.Send[any](c.actor.Inbox(), &lruLenMsg{BaseMessage: msg}, c.timeout); err != nil {
		return 0
	}
	result, err := actor.CallResult[int](replyCh, c.timeout)
	if err != nil {
		return 0
	}
	return result
}

// Clear removes all entries from the cache and resets statistics.
func (c *PayloadLRUCache) Clear() {
	msg, replyCh := actor.NewBaseMessage()
	if err := actor.Send[any](c.actor.Inbox(), &lruClearMsg{BaseMessage: msg}, c.timeout); err != nil {
		return
	}
	actor.CallResult[struct{}](replyCh, c.timeout)
}

// Stats returns a snapshot of cache hit/miss/eviction statistics.
func (c *PayloadLRUCache) Stats() *LRUCacheStats {
	return c.actor.Stats()
}

// MaxEntries returns the maximum number of entries the cache can hold.
func (c *PayloadLRUCache) MaxEntries() int {
	return c.actor.maxEntries
}

// Close stops the actor goroutine.
func (c *PayloadLRUCache) Close() {
	c.cancel()
}
