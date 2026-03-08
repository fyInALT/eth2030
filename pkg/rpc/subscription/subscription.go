// Package rpcsub provides subscription management types and logic for the
// Ethereum JSON-RPC WebSocket subscription API.
package rpcsub

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
	rpcfilter "github.com/eth2030/eth2030/rpc/filter"
	rpctypes "github.com/eth2030/eth2030/rpc/types"
)

// ---------- SubKind / SubRegistry ----------

// SubKind identifies the type of real-time subscription.
type SubKind int

const (
	SubKindNewHeads   SubKind = iota // New block header notifications.
	SubKindLogs                      // Matching log event notifications.
	SubKindPendingTx                 // Pending transaction notifications.
	SubKindSyncStatus                // Sync progress notifications.
)

// subKindNames maps SubKind to its eth_subscribe parameter name.
var subKindNames = map[string]SubKind{
	"newHeads":               SubKindNewHeads,
	"logs":                   SubKindLogs,
	"newPendingTransactions": SubKindPendingTx,
	"syncing":                SubKindSyncStatus,
}

// Subscription manager errors.
var (
	ErrSubManagerClosed     = errors.New("subscription manager: closed")
	ErrSubManagerCapacity   = errors.New("subscription manager: capacity reached")
	ErrSubManagerNotFound   = errors.New("subscription manager: subscription not found")
	ErrSubManagerRateLimit  = errors.New("subscription manager: rate limit exceeded")
	ErrSubManagerInvalidTyp = errors.New("subscription manager: invalid subscription type")
)

// ParseSubKind converts a subscription type name to a SubKind.
// Returns ErrSubManagerInvalidTyp if the name is not recognized.
func ParseSubKind(name string) (SubKind, error) {
	kind, ok := subKindNames[name]
	if !ok {
		return 0, ErrSubManagerInvalidTyp
	}
	return kind, nil
}

// SubEntry is a single registered subscription.
type SubEntry struct {
	ID        string
	Kind      SubKind
	ConnID    string                // connection identifier for grouping
	Query     rpcfilter.FilterQuery // only used for logs subscriptions
	CreatedAt time.Time
	ch        chan interface{}
}

// Channel returns the notification channel.
func (s *SubEntry) Channel() <-chan interface{} {
	return s.ch
}

// SubRateLimitConfig configures per-connection subscription rate limiting.
type SubRateLimitConfig struct {
	MaxSubsPerConn  int           // max subscriptions per connection
	WindowDuration  time.Duration // rate limit window
	MaxEventsPerSec int           // max events per second per subscription
}

// DefaultSubRateLimitConfig returns sensible defaults.
func DefaultSubRateLimitConfig() SubRateLimitConfig {
	return SubRateLimitConfig{
		MaxSubsPerConn:  32,
		WindowDuration:  time.Second,
		MaxEventsPerSec: 1000,
	}
}

// connTracker tracks subscriptions per connection for rate limiting.
type connTracker struct {
	subCount    int
	lastEventAt time.Time
	eventCount  int
}

// SubRegistry manages active subscriptions across multiple connections.
type SubRegistry struct {
	mu           sync.Mutex
	subs         map[string]*SubEntry
	connTrackers map[string]*connTracker
	rateConfig   SubRateLimitConfig
	bufferSize   int
	nextSeq      uint64
	closed       bool
}

// NewSubRegistry creates a new subscription registry.
func NewSubRegistry(rateConfig SubRateLimitConfig, bufferSize int) *SubRegistry {
	if bufferSize <= 0 {
		bufferSize = 128
	}
	return &SubRegistry{
		subs:         make(map[string]*SubEntry),
		connTrackers: make(map[string]*connTracker),
		rateConfig:   rateConfig,
		bufferSize:   bufferSize,
	}
}

// generateSubID creates a unique hex subscription ID.
func (r *SubRegistry) generateSubID() string {
	r.nextSeq++
	buf := make([]byte, 16)
	seq := r.nextSeq
	ts := uint64(time.Now().UnixNano())
	for i := 0; i < 8; i++ {
		buf[i] = byte(seq >> (8 * i))
		buf[8+i] = byte(ts >> (8 * i))
	}
	h := crypto.Keccak256(buf)
	return "0x" + hex.EncodeToString(h[:16])
}

// NewHeadsSub creates a new heads subscription for the given connection.
func (r *SubRegistry) NewHeadsSub(connID string) (string, error) {
	return r.addSub(connID, SubKindNewHeads, rpcfilter.FilterQuery{})
}

// LogsSub creates a logs subscription with the given filter for a connection.
func (r *SubRegistry) LogsSub(connID string, query rpcfilter.FilterQuery) (string, error) {
	return r.addSub(connID, SubKindLogs, query)
}

// PendingTxSub creates a pending transaction subscription for a connection.
func (r *SubRegistry) PendingTxSub(connID string) (string, error) {
	return r.addSub(connID, SubKindPendingTx, rpcfilter.FilterQuery{})
}

// SyncStatusSub creates a sync status subscription for a connection.
func (r *SubRegistry) SyncStatusSub(connID string) (string, error) {
	return r.addSub(connID, SubKindSyncStatus, rpcfilter.FilterQuery{})
}

// addSub adds a subscription after checking rate limits.
func (r *SubRegistry) addSub(connID string, kind SubKind, query rpcfilter.FilterQuery) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return "", ErrSubManagerClosed
	}

	// Check per-connection limit.
	tracker := r.connTrackers[connID]
	if tracker == nil {
		tracker = &connTracker{}
		r.connTrackers[connID] = tracker
	}
	if tracker.subCount >= r.rateConfig.MaxSubsPerConn {
		return "", ErrSubManagerRateLimit
	}

	id := r.generateSubID()
	entry := &SubEntry{
		ID:        id,
		Kind:      kind,
		ConnID:    connID,
		Query:     query,
		CreatedAt: time.Now(),
		ch:        make(chan interface{}, r.bufferSize),
	}
	r.subs[id] = entry
	tracker.subCount++

	return id, nil
}

// Unsubscribe removes a subscription by ID.
func (r *SubRegistry) Unsubscribe(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.subs[id]
	if !ok {
		return ErrSubManagerNotFound
	}

	close(entry.ch)
	delete(r.subs, id)

	if tracker := r.connTrackers[entry.ConnID]; tracker != nil {
		tracker.subCount--
		if tracker.subCount <= 0 {
			delete(r.connTrackers, entry.ConnID)
		}
	}
	return nil
}

// GetSub returns a subscription by ID, or nil if not found.
func (r *SubRegistry) GetSub(id string) *SubEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.subs[id]
}

// DisconnectConn removes all subscriptions for a given connection,
// cleaning up channels and tracker state.
func (r *SubRegistry) DisconnectConn(connID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	removed := 0
	for id, entry := range r.subs {
		if entry.ConnID == connID {
			close(entry.ch)
			delete(r.subs, id)
			removed++
		}
	}
	delete(r.connTrackers, connID)
	return removed
}

// Count returns the total number of active subscriptions.
func (r *SubRegistry) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.subs)
}

// CountByKind returns the number of subscriptions of a given kind.
func (r *SubRegistry) CountByKind(kind SubKind) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, entry := range r.subs {
		if entry.Kind == kind {
			count++
		}
	}
	return count
}

// ConnSubCount returns the number of subscriptions for a connection.
func (r *SubRegistry) ConnSubCount(connID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	if tracker := r.connTrackers[connID]; tracker != nil {
		return tracker.subCount
	}
	return 0
}

// NotifyNewHead sends a new header to all newHeads subscribers.
// The header is formatted as *rpctypes.RPCBlock before sending.
func (r *SubRegistry) NotifyNewHead(header *types.Header) {
	r.mu.Lock()
	defer r.mu.Unlock()

	formatted := rpctypes.FormatHeader(header)
	for _, entry := range r.subs {
		if entry.Kind == SubKindNewHeads {
			select {
			case entry.ch <- formatted:
			default:
				// Drop if full.
			}
		}
	}
}

// NotifyLogEvents sends matching logs to all logs subscribers.
// Logs are formatted as *rpctypes.RPCLog before sending.
func (r *SubRegistry) NotifyLogEvents(logs []*types.Log) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, entry := range r.subs {
		if entry.Kind != SubKindLogs {
			continue
		}
		for _, log := range logs {
			if rpcfilter.MatchFilter(log, entry.Query) {
				formatted := rpctypes.FormatLog(log)
				select {
				case entry.ch <- formatted:
				default:
				}
			}
		}
	}
}

// NotifyPendingTxHash sends a pending transaction hash to all pendingTx subs.
func (r *SubRegistry) NotifyPendingTxHash(txHash types.Hash) {
	r.mu.Lock()
	defer r.mu.Unlock()

	hashStr := txHash.Hex()
	for _, entry := range r.subs {
		if entry.Kind == SubKindPendingTx {
			select {
			case entry.ch <- hashStr:
			default:
			}
		}
	}
}

// NotifySyncStatus sends sync status to all syncing subscribers.
func (r *SubRegistry) NotifySyncStatus(syncing bool, currentBlock, highestBlock uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result interface{}
	if !syncing {
		result = false
	} else {
		result = map[string]string{
			"currentBlock": fmt.Sprintf("0x%x", currentBlock),
			"highestBlock": fmt.Sprintf("0x%x", highestBlock),
		}
	}

	for _, entry := range r.subs {
		if entry.Kind == SubKindSyncStatus {
			select {
			case entry.ch <- result:
			default:
			}
		}
	}
}

// Close shuts down all subscriptions and prevents new ones.
func (r *SubRegistry) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.closed = true
	for id, entry := range r.subs {
		close(entry.ch)
		delete(r.subs, id)
	}
	r.connTrackers = make(map[string]*connTracker)
}

// IsClosed returns whether the registry is closed.
func (r *SubRegistry) IsClosed() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.closed
}

// CheckRateLimit checks whether a connection has exceeded its event rate.
// Returns true if the event should be allowed, false if rate limited.
func (r *SubRegistry) CheckRateLimit(connID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	tracker := r.connTrackers[connID]
	if tracker == nil {
		return true
	}

	now := time.Now()
	if now.Sub(tracker.lastEventAt) >= r.rateConfig.WindowDuration {
		tracker.eventCount = 0
		tracker.lastEventAt = now
	}
	tracker.eventCount++
	return tracker.eventCount <= r.rateConfig.MaxEventsPerSec
}

// ---------- SubscriptionConfig / WSSubscription / WSSubscriptionManager ----------

// SubscriptionConfig configures the WebSocket subscription manager.
type SubscriptionConfig struct {
	// MaxSubscriptions is the maximum number of concurrent subscriptions.
	MaxSubscriptions int
	// BufferSize is the channel buffer size for each subscription.
	BufferSize int
	// CleanupInterval is the number of seconds between automatic cleanup runs.
	CleanupInterval int64
}

// DefaultSubscriptionConfig returns a SubscriptionConfig with sensible defaults.
func DefaultSubscriptionConfig() SubscriptionConfig {
	return SubscriptionConfig{
		MaxSubscriptions: 256,
		BufferSize:       128,
		CleanupInterval:  300, // 5 minutes
	}
}

// supportedSubTypes lists the valid subscription types.
var supportedSubTypes = map[string]bool{
	"newHeads":               true,
	"logs":                   true,
	"newPendingTransactions": true,
	"syncing":                true,
}

// WSSubscription represents an active WebSocket subscription with
// string-based type and flexible filter criteria.
type WSSubscription struct {
	ID             string
	Type           string
	FilterCriteria map[string]interface{}
	CreatedAt      int64
	Active         bool
	ch             chan interface{}
}

// Channel returns the notification channel for this subscription.
func (s *WSSubscription) Channel() <-chan interface{} {
	return s.ch
}

// WSSubscriptionManager manages WebSocket subscriptions for real-time
// event streaming. It is thread-safe.
type WSSubscriptionManager struct {
	mu      sync.Mutex
	config  SubscriptionConfig
	subs    map[string]*WSSubscription
	nextSeq uint64
	closed  bool
}

// NewWSSubscriptionManager creates a new subscription manager with
// the given configuration.
func NewWSSubscriptionManager(config SubscriptionConfig) *WSSubscriptionManager {
	if config.MaxSubscriptions <= 0 {
		config.MaxSubscriptions = 256
	}
	if config.BufferSize <= 0 {
		config.BufferSize = 128
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = 300
	}
	return &WSSubscriptionManager{
		config: config,
		subs:   make(map[string]*WSSubscription),
	}
}

// generateSubID produces a unique hex subscription ID.
func (m *WSSubscriptionManager) generateSubID() string {
	m.nextSeq++
	buf := make([]byte, 16)
	seq := m.nextSeq
	ts := uint64(time.Now().UnixNano())
	for i := 0; i < 8; i++ {
		buf[i] = byte(seq >> (8 * i))
		buf[8+i] = byte(ts >> (8 * i))
	}
	h := crypto.Keccak256(buf)
	return "0x" + hex.EncodeToString(h[:16])
}

// Subscribe creates a new subscription of the given type with optional
// filter criteria. Returns the subscription ID or an error.
func (m *WSSubscriptionManager) Subscribe(subType string, criteria map[string]interface{}) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return "", errors.New("subscription manager is closed")
	}
	if !supportedSubTypes[subType] {
		return "", errors.New("unsupported subscription type: " + subType)
	}
	if len(m.subs) >= m.config.MaxSubscriptions {
		return "", errors.New("maximum subscription count reached")
	}

	id := m.generateSubID()
	sub := &WSSubscription{
		ID:             id,
		Type:           subType,
		FilterCriteria: criteria,
		CreatedAt:      time.Now().Unix(),
		Active:         true,
		ch:             make(chan interface{}, m.config.BufferSize),
	}
	m.subs[id] = sub
	return id, nil
}

// Unsubscribe removes a subscription by ID and closes its channel.
// Returns an error if the subscription does not exist.
func (m *WSSubscriptionManager) Unsubscribe(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sub, ok := m.subs[id]
	if !ok {
		return errors.New("subscription not found: " + id)
	}
	sub.Active = false
	close(sub.ch)
	delete(m.subs, id)
	return nil
}

// GetSubscription returns subscription details by ID, or nil if not found.
func (m *WSSubscriptionManager) GetSubscription(id string) *WSSubscription {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.subs[id]
}

// ActiveSubscriptions returns a snapshot of all active subscriptions.
func (m *WSSubscriptionManager) ActiveSubscriptions() []WSSubscription {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]WSSubscription, 0, len(m.subs))
	for _, sub := range m.subs {
		if sub.Active {
			result = append(result, WSSubscription{
				ID:             sub.ID,
				Type:           sub.Type,
				FilterCriteria: sub.FilterCriteria,
				CreatedAt:      sub.CreatedAt,
				Active:         sub.Active,
			})
		}
	}
	return result
}

// PublishEvent sends an event to all subscriptions matching the given type.
func (m *WSSubscriptionManager) PublishEvent(eventType string, data interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, sub := range m.subs {
		if !sub.Active || sub.Type != eventType {
			continue
		}
		select {
		case sub.ch <- data:
		default:
			// Drop if buffer is full; subscriber is too slow.
		}
	}
}

// SubscriberCount returns the number of active subscribers for the given event type.
func (m *WSSubscriptionManager) SubscriberCount(eventType string) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for _, sub := range m.subs {
		if sub.Active && sub.Type == eventType {
			count++
		}
	}
	return count
}

// Cleanup removes subscriptions that are no longer active or have stale buffers.
func (m *WSSubscriptionManager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().Unix()
	for id, sub := range m.subs {
		if !sub.Active {
			close(sub.ch)
			delete(m.subs, id)
			continue
		}
		age := now - sub.CreatedAt
		if age > m.config.CleanupInterval && len(sub.ch) == cap(sub.ch) {
			sub.Active = false
			close(sub.ch)
			delete(m.subs, id)
		}
	}
}

// Close shuts down all active subscriptions and prevents new ones.
func (m *WSSubscriptionManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	for id, sub := range m.subs {
		sub.Active = false
		close(sub.ch)
		delete(m.subs, id)
	}
}

// ---------- WSNotification / WSSubscriptionResult ----------

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

// ---------- SubscriptionDispatcher ----------

// Dispatcher errors.
var (
	ErrDispatcherClosed       = errors.New("dispatcher: closed")
	ErrDispatcherClientLimit  = errors.New("dispatcher: client subscription limit exceeded")
	ErrDispatcherEventLimit   = errors.New("dispatcher: event rate limit exceeded")
	ErrDispatcherSubNotFound  = errors.New("dispatcher: subscription not found")
	ErrDispatcherInvalidTopic = errors.New("dispatcher: invalid subscription topic")
	ErrDispatcherDuplicate    = errors.New("dispatcher: duplicate subscription")
)

// SubscriptionTopic identifies a category of real-time events.
type SubscriptionTopic string

const (
	TopicNewHeads   SubscriptionTopic = "newHeads"
	TopicLogs       SubscriptionTopic = "logs"
	TopicPendingTxs SubscriptionTopic = "newPendingTransactions"
	TopicSyncing    SubscriptionTopic = "syncing"
)

// validTopics is the set of recognized subscription topics.
var validTopics = map[SubscriptionTopic]bool{
	TopicNewHeads:   true,
	TopicLogs:       true,
	TopicPendingTxs: true,
	TopicSyncing:    true,
}

// IsValidTopic returns whether the topic is recognized.
func IsValidTopic(topic SubscriptionTopic) bool {
	return validTopics[topic]
}

// DispatchSubscription represents a single active subscription managed
// by the dispatcher.
type DispatchSubscription struct {
	ID        string
	ClientID  string
	Topic     SubscriptionTopic
	Filter    interface{} // Topic-specific filter (e.g., log filter criteria).
	Created   time.Time
	LastEvent time.Time
	Events    uint64 // Total events delivered.
	ch        chan interface{}
}

// Channel returns the read-only notification channel.
func (ds *DispatchSubscription) Channel() <-chan interface{} {
	return ds.ch
}

// DispatcherConfig configures the subscription dispatcher.
type DispatcherConfig struct {
	MaxSubsPerClient int           // Max subscriptions per client ID.
	MaxEventsPerSec  int           // Max events per second per client.
	RateWindow       time.Duration // Rate limit window duration.
	BufferSize       int           // Channel buffer size per subscription.
	MaxTotalSubs     int           // Global maximum subscriptions (0 = unlimited).
}

// DefaultDispatcherConfig returns sensible defaults.
func DefaultDispatcherConfig() DispatcherConfig {
	return DispatcherConfig{
		MaxSubsPerClient: 32,
		MaxEventsPerSec:  1000,
		RateWindow:       time.Second,
		BufferSize:       128,
		MaxTotalSubs:     4096,
	}
}

// clientState tracks per-client rate limiting and subscription counts.
type clientState struct {
	subCount    int
	eventCount  int
	windowStart time.Time
}

// SubStats holds per-topic counts and global totals.
type SubStats struct {
	NewHeads   int
	Logs       int
	PendingTxs int
	Syncing    int
	Total      int
	Clients    int
}

// SubscriptionDispatcher manages active subscriptions across multiple
// WebSocket clients with rate limiting and lifecycle tracking.
// All methods are safe for concurrent use.
type SubscriptionDispatcher struct {
	mu      sync.Mutex
	config  DispatcherConfig
	subs    map[string]*DispatchSubscription // Keyed by subscription ID.
	clients map[string]*clientState          // Keyed by client ID.
	nextSeq uint64
	closed  bool
}

// NewSubscriptionDispatcher creates a new dispatcher with the given config.
func NewSubscriptionDispatcher(config DispatcherConfig) *SubscriptionDispatcher {
	if config.BufferSize <= 0 {
		config.BufferSize = 128
	}
	if config.MaxSubsPerClient <= 0 {
		config.MaxSubsPerClient = 32
	}
	if config.RateWindow <= 0 {
		config.RateWindow = time.Second
	}
	return &SubscriptionDispatcher{
		config:  config,
		subs:    make(map[string]*DispatchSubscription),
		clients: make(map[string]*clientState),
	}
}

// generateID produces a unique hex subscription ID.
func (d *SubscriptionDispatcher) generateID() string {
	d.nextSeq++
	buf := make([]byte, 16)
	seq := d.nextSeq
	ts := uint64(time.Now().UnixNano())
	for i := 0; i < 8; i++ {
		buf[i] = byte(seq >> (8 * i))
		buf[8+i] = byte(ts >> (8 * i))
	}
	h := crypto.Keccak256(buf)
	return "0x" + hex.EncodeToString(h[:16])
}

// Subscribe creates a new subscription for the given client and topic.
// Returns the subscription or an error if limits are exceeded.
func (d *SubscriptionDispatcher) Subscribe(clientID string, topic SubscriptionTopic, filter interface{}) (*DispatchSubscription, error) {
	if !IsValidTopic(topic) {
		return nil, fmt.Errorf("%w: %s", ErrDispatcherInvalidTopic, topic)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil, ErrDispatcherClosed
	}

	// Check global limit.
	if d.config.MaxTotalSubs > 0 && len(d.subs) >= d.config.MaxTotalSubs {
		return nil, ErrDispatcherClientLimit
	}

	// Check per-client limit.
	cs := d.clients[clientID]
	if cs == nil {
		cs = &clientState{windowStart: time.Now()}
		d.clients[clientID] = cs
	}
	if cs.subCount >= d.config.MaxSubsPerClient {
		return nil, ErrDispatcherClientLimit
	}

	id := d.generateID()
	sub := &DispatchSubscription{
		ID:       id,
		ClientID: clientID,
		Topic:    topic,
		Filter:   filter,
		Created:  time.Now(),
		ch:       make(chan interface{}, d.config.BufferSize),
	}
	d.subs[id] = sub
	cs.subCount++

	return sub, nil
}

// Unsubscribe removes a subscription by ID and closes its channel.
func (d *SubscriptionDispatcher) Unsubscribe(subID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	sub, ok := d.subs[subID]
	if !ok {
		return ErrDispatcherSubNotFound
	}

	close(sub.ch)
	delete(d.subs, subID)

	if cs := d.clients[sub.ClientID]; cs != nil {
		cs.subCount--
		if cs.subCount <= 0 {
			delete(d.clients, sub.ClientID)
		}
	}
	return nil
}

// Broadcast sends data to all subscriptions matching the given topic.
func (d *SubscriptionDispatcher) Broadcast(topic SubscriptionTopic, data interface{}) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return
	}

	now := time.Now()
	for _, sub := range d.subs {
		if sub.Topic != topic {
			continue
		}

		// Check client rate limit.
		cs := d.clients[sub.ClientID]
		if cs != nil && d.config.MaxEventsPerSec > 0 {
			if now.Sub(cs.windowStart) >= d.config.RateWindow {
				cs.eventCount = 0
				cs.windowStart = now
			}
			cs.eventCount++
			if cs.eventCount > d.config.MaxEventsPerSec {
				continue // Rate limited; skip this event.
			}
		}

		select {
		case sub.ch <- data:
			sub.LastEvent = now
			sub.Events++
		default:
			// Buffer full; drop event.
		}
	}
}

// GetSubscriptions returns all active subscriptions for the given client.
func (d *SubscriptionDispatcher) GetSubscriptions(clientID string) []*DispatchSubscription {
	d.mu.Lock()
	defer d.mu.Unlock()

	var result []*DispatchSubscription
	for _, sub := range d.subs {
		if sub.ClientID == clientID {
			cp := &DispatchSubscription{
				ID:        sub.ID,
				ClientID:  sub.ClientID,
				Topic:     sub.Topic,
				Filter:    sub.Filter,
				Created:   sub.Created,
				LastEvent: sub.LastEvent,
				Events:    sub.Events,
			}
			result = append(result, cp)
		}
	}
	return result
}

// GetSubscription returns a single subscription by ID, or nil.
func (d *SubscriptionDispatcher) GetSubscription(subID string) *DispatchSubscription {
	d.mu.Lock()
	defer d.mu.Unlock()

	sub, ok := d.subs[subID]
	if !ok {
		return nil
	}
	return sub
}

// CleanupStale removes subscriptions that have not received an event
// within the given maxAge since their creation. Returns the count removed.
func (d *SubscriptionDispatcher) CleanupStale(maxAge time.Duration) int {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	removed := 0
	for id, sub := range d.subs {
		age := now.Sub(sub.Created)
		if age < maxAge {
			continue
		}
		lastActivity := sub.LastEvent
		if lastActivity.IsZero() {
			lastActivity = sub.Created
		}
		if now.Sub(lastActivity) >= maxAge {
			close(sub.ch)
			delete(d.subs, id)
			if cs := d.clients[sub.ClientID]; cs != nil {
				cs.subCount--
				if cs.subCount <= 0 {
					delete(d.clients, sub.ClientID)
				}
			}
			removed++
		}
	}
	return removed
}

// DisconnectClient removes all subscriptions for the given client ID.
func (d *SubscriptionDispatcher) DisconnectClient(clientID string) int {
	d.mu.Lock()
	defer d.mu.Unlock()

	removed := 0
	for id, sub := range d.subs {
		if sub.ClientID == clientID {
			close(sub.ch)
			delete(d.subs, id)
			removed++
		}
	}
	delete(d.clients, clientID)
	return removed
}

// SubscriptionStats returns per-topic counts and overall statistics.
func (d *SubscriptionDispatcher) SubscriptionStats() *SubStats {
	d.mu.Lock()
	defer d.mu.Unlock()

	stats := &SubStats{
		Total:   len(d.subs),
		Clients: len(d.clients),
	}
	for _, sub := range d.subs {
		switch sub.Topic {
		case TopicNewHeads:
			stats.NewHeads++
		case TopicLogs:
			stats.Logs++
		case TopicPendingTxs:
			stats.PendingTxs++
		case TopicSyncing:
			stats.Syncing++
		}
	}
	return stats
}

// TotalSubscriptions returns the total number of active subscriptions.
func (d *SubscriptionDispatcher) TotalSubscriptions() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.subs)
}

// ClientCount returns the number of distinct clients with subscriptions.
func (d *SubscriptionDispatcher) ClientCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.clients)
}

// Close shuts down all subscriptions and prevents new ones.
func (d *SubscriptionDispatcher) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.closed = true
	for id, sub := range d.subs {
		close(sub.ch)
		delete(d.subs, id)
	}
	d.clients = make(map[string]*clientState)
}

// IsClosed returns whether the dispatcher has been closed.
func (d *SubscriptionDispatcher) IsClosed() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.closed
}

// CheckClientRateLimit returns true if the client is within the event rate limit.
func (d *SubscriptionDispatcher) CheckClientRateLimit(clientID string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	cs := d.clients[clientID]
	if cs == nil {
		return true
	}

	now := time.Now()
	if now.Sub(cs.windowStart) >= d.config.RateWindow {
		cs.eventCount = 0
		cs.windowStart = now
	}
	return cs.eventCount < d.config.MaxEventsPerSec
}
