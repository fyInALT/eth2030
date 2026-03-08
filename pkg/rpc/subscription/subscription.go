// Package rpcsub provides subscription management types and logic for the
// Ethereum JSON-RPC WebSocket subscription API.
package rpcsub

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
)

// SubKind distinguishes the kind of WebSocket subscription.
type SubKind int

const (
	// SubKindNewHeads watches for new block headers.
	SubKindNewHeads SubKind = iota
	// SubKindLogs watches for matching log events.
	SubKindLogs
	// SubKindPendingTx watches for new pending transaction hashes.
	SubKindPendingTx
	// SubKindSyncStatus watches for sync status changes.
	SubKindSyncStatus
)

// SubRateLimitConfig configures per-subscription rate limiting.
type SubRateLimitConfig struct {
	// MaxNotificationsPerSecond limits notifications sent per second (0 = unlimited).
	MaxNotificationsPerSecond int
	// BurstSize is the allowed burst above the rate limit.
	BurstSize int
}

// SubEntry represents an active WebSocket subscription in a registry.
type SubEntry struct {
	ID        string
	ConnID    string
	Kind      SubKind
	CreatedAt time.Time
	LastRecv  time.Time
}

// SubRegistry manages WebSocket subscriptions across multiple connections.
type SubRegistry struct {
	mu   sync.RWMutex
	subs map[string]*SubEntry // id → entry
	seq  uint64
}

// NewSubRegistry creates a new subscription registry.
func NewSubRegistry() *SubRegistry {
	return &SubRegistry{subs: make(map[string]*SubEntry)}
}

// generateID creates a unique subscription ID.
func (r *SubRegistry) generateID() string {
	r.seq++
	var buf [16]byte
	for i := 0; i < 8; i++ {
		buf[i] = byte(r.seq >> (8 * i))
		buf[8+i] = byte(uint64(time.Now().UnixNano()) >> (8 * i))
	}
	h := crypto.Keccak256(buf[:])
	return "0x" + hex.EncodeToString(h[:16])
}

// HeadsSub registers a newHeads subscription for connID.
func (r *SubRegistry) HeadsSub(connID string) (string, error) {
	return r.addSub(connID, SubKindNewHeads)
}

// LogsSub registers a logs subscription for connID.
func (r *SubRegistry) LogsSub(connID string) (string, error) {
	return r.addSub(connID, SubKindLogs)
}

// PendingTxSub registers a pendingTransactions subscription for connID.
func (r *SubRegistry) PendingTxSub(connID string) (string, error) {
	return r.addSub(connID, SubKindPendingTx)
}

// SyncSub registers a syncing subscription for connID.
func (r *SubRegistry) SyncSub(connID string) (string, error) {
	return r.addSub(connID, SubKindSyncStatus)
}

func (r *SubRegistry) addSub(connID string, kind SubKind) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.generateID()
	r.subs[id] = &SubEntry{
		ID:        id,
		ConnID:    connID,
		Kind:      kind,
		CreatedAt: time.Now(),
	}
	return id, nil
}

// Unsubscribe removes a subscription by ID. Returns true if it existed.
func (r *SubRegistry) Unsubscribe(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.subs[id]
	if ok {
		delete(r.subs, id)
	}
	return ok
}

// SubsForConn returns all subscription IDs for a given connection.
func (r *SubRegistry) SubsForConn(connID string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var ids []string
	for id, entry := range r.subs {
		if entry.ConnID == connID {
			ids = append(ids, id)
		}
	}
	return ids
}

// RemoveConn removes all subscriptions for a connection.
func (r *SubRegistry) RemoveConn(connID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for id, entry := range r.subs {
		if entry.ConnID == connID {
			delete(r.subs, id)
			count++
		}
	}
	return count
}

// Count returns total subscription count.
func (r *SubRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.subs)
}

// SubscriptionConfig configures a WebSocket subscription manager.
type SubscriptionConfig struct {
	// MaxSubscriptions is the maximum number of concurrent subscriptions.
	MaxSubscriptions int
	// TTL is the maximum lifetime for an idle subscription.
	TTL time.Duration
	// MaxNotifyBuffer is the notification channel buffer size.
	MaxNotifyBuffer int
}

// DefaultSubscriptionConfig returns sensible defaults.
func DefaultSubscriptionConfig() SubscriptionConfig {
	return SubscriptionConfig{
		MaxSubscriptions: 1000,
		TTL:              10 * time.Minute,
		MaxNotifyBuffer:  256,
	}
}

// WSNotification is a JSON-RPC 2.0 subscription notification sent over WebSocket.
type WSNotification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// WSSubscriptionResult wraps the subscription ID and result for notifications.
type WSSubscriptionResult struct {
	Subscription string      `json:"subscription"`
	Result       interface{} `json:"result"`
}

// FormatWSNotification creates a JSON-RPC subscription notification message.
func FormatWSNotification(subID string, result interface{}) *WSNotification {
	params := WSSubscriptionResult{
		Subscription: subID,
		Result:       result,
	}
	raw, _ := json.Marshal(params)
	return &WSNotification{
		JSONRPC: "2.0",
		Method:  "eth_subscription",
		Params:  raw,
	}
}

// Dispatcher errors.
var (
	ErrDispatcherClosed    = errors.New("subscription dispatcher is closed")
	ErrDispatcherFull      = errors.New("subscription dispatcher topic buffer full")
	ErrDispatcherUnknown   = errors.New("unknown subscriber ID")
	ErrDispatcherDuplicate = errors.New("duplicate subscriber ID")
)

// SubDispatchKind categorizes the kind of subscription for dispatch.
type SubDispatchKind int

const (
	SubDispatchNewHeads SubDispatchKind = iota
	SubDispatchLogs
	SubDispatchPendingTx
	SubDispatchSyncStatus
)

// SubMessage holds a single notification payload for a subscriber.
type SubMessage struct {
	SubscriptionID string
	Topic          SubDispatchKind
	Data           interface{}
	Timestamp      time.Time
}

// DispatchStats tracks dispatcher statistics.
type DispatchStats struct {
	TotalSent    uint64
	TotalDropped uint64
	ActiveSubs   int
}

// WSSubEntry is an entry in the WebSocket subscription manager.
type WSSubEntry struct {
	ID          string
	Kind        SubKind
	ConnID      string
	CreatedAt   time.Time
	LastNotify  time.Time
	NotifyCount uint64
}

// SubscriptionDispatcher dispatches events to multiple subscribers.
type SubscriptionDispatcher struct {
	mu      sync.Mutex
	subs    map[string]*subDispEntry
	closed  bool
	stats   DispatchStats
	nextSeq uint64
}

type subDispEntry struct {
	id       string
	kind     SubDispatchKind
	ch       chan SubMessage
	created  time.Time
	lastRecv time.Time
}

// NewSubscriptionDispatcher creates a new dispatcher.
func NewSubscriptionDispatcher() *SubscriptionDispatcher {
	return &SubscriptionDispatcher{
		subs: make(map[string]*subDispEntry),
	}
}

// Subscribe registers a subscriber for the given topic and returns the subscription ID.
func (d *SubscriptionDispatcher) Subscribe(kind SubDispatchKind, bufSize int) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return "", ErrDispatcherClosed
	}
	d.nextSeq++
	var buf [8]byte
	for i := 0; i < 8; i++ {
		buf[i] = byte(d.nextSeq >> (8 * i))
	}
	h := crypto.Keccak256(buf[:])
	id := "0x" + hex.EncodeToString(h[:8])

	if _, exists := d.subs[id]; exists {
		return "", ErrDispatcherDuplicate
	}
	d.subs[id] = &subDispEntry{
		id:      id,
		kind:    kind,
		ch:      make(chan SubMessage, bufSize),
		created: time.Now(),
	}
	d.stats.ActiveSubs++
	return id, nil
}

// Unsubscribe removes a subscriber.
func (d *SubscriptionDispatcher) Unsubscribe(id string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	entry, ok := d.subs[id]
	if !ok {
		return false
	}
	close(entry.ch)
	delete(d.subs, id)
	d.stats.ActiveSubs--
	return true
}

// Dispatch sends a message to all subscribers of the given topic.
func (d *SubscriptionDispatcher) Dispatch(kind SubDispatchKind, data interface{}) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return
	}

	msg := SubMessage{
		Topic:     kind,
		Data:      data,
		Timestamp: time.Now(),
	}

	for _, entry := range d.subs {
		if entry.kind != kind {
			continue
		}
		msg.SubscriptionID = entry.id
		select {
		case entry.ch <- msg:
			d.stats.TotalSent++
			entry.lastRecv = time.Now()
		default:
			d.stats.TotalDropped++
		}
	}
}

// Channel returns the notification channel for the given subscription ID.
func (d *SubscriptionDispatcher) Channel(id string) (<-chan SubMessage, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	entry, ok := d.subs[id]
	if !ok {
		return nil, ErrDispatcherUnknown
	}
	return entry.ch, nil
}

// Stats returns dispatcher statistics.
func (d *SubscriptionDispatcher) Stats() DispatchStats {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.stats
}

// Close shuts down the dispatcher.
func (d *SubscriptionDispatcher) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return
	}
	d.closed = true
	for _, entry := range d.subs {
		close(entry.ch)
	}
	d.subs = make(map[string]*subDispEntry)
}

// WSSubscriptionManager manages active WebSocket subscriptions with TTL support.
type WSSubscriptionManager struct {
	mu     sync.RWMutex
	subs   map[string]*WSSubEntry
	config SubscriptionConfig
	seq    uint64
}

// NewWSSubscriptionManager creates a new WebSocket subscription manager.
func NewWSSubscriptionManager(config SubscriptionConfig) *WSSubscriptionManager {
	return &WSSubscriptionManager{
		subs:   make(map[string]*WSSubEntry),
		config: config,
	}
}

// Add creates a new subscription and returns its ID.
func (m *WSSubscriptionManager) Add(connID string, kind SubKind) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config.MaxSubscriptions > 0 && len(m.subs) >= m.config.MaxSubscriptions {
		return "", errors.New("subscription limit reached")
	}

	m.seq++
	var buf [8]byte
	for i := 0; i < 8; i++ {
		buf[i] = byte(m.seq >> (8 * i))
	}
	h := crypto.Keccak256(buf[:])
	id := "0x" + hex.EncodeToString(h[:8])

	m.subs[id] = &WSSubEntry{
		ID:        id,
		Kind:      kind,
		ConnID:    connID,
		CreatedAt: time.Now(),
	}
	return id, nil
}

// Remove removes a subscription by ID.
func (m *WSSubscriptionManager) Remove(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.subs[id]
	if ok {
		delete(m.subs, id)
	}
	return ok
}

// Get returns a subscription entry by ID.
func (m *WSSubscriptionManager) Get(id string) (*WSSubEntry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.subs[id]
	return e, ok
}

// Count returns the total subscription count.
func (m *WSSubscriptionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.subs)
}

// PruneExpired removes subscriptions that have exceeded their TTL.
func (m *WSSubscriptionManager) PruneExpired() int {
	if m.config.TTL <= 0 {
		return 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := time.Now().Add(-m.config.TTL)
	count := 0
	for id, entry := range m.subs {
		if entry.LastNotify.Before(cutoff) && entry.CreatedAt.Before(cutoff) {
			delete(m.subs, id)
			count++
		}
	}
	return count
}

// ConnsForKind returns all subscription IDs of the given kind.
func (m *WSSubscriptionManager) ConnsForKind(kind SubKind) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var ids []string
	for id, entry := range m.subs {
		if entry.Kind == kind {
			ids = append(ids, id)
		}
	}
	return ids
}

// RecordNotify updates the LastNotify timestamp and increments count for a subscription.
func (m *WSSubscriptionManager) RecordNotify(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if entry, ok := m.subs[id]; ok {
		entry.LastNotify = time.Now()
		entry.NotifyCount++
	}
}

// GetConfig returns the subscription config.
func (m *WSSubscriptionManager) GetConfig() SubscriptionConfig {
	return m.config
}

// typesHashEncode encodes a Hash as hex string (for notifications).
func typesHashEncode(h types.Hash) string {
	return h.Hex()
}
