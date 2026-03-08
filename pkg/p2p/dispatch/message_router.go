package dispatch

import (
	"container/heap"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/eth2030/eth2030/p2p/wire"
)

const (
	defaultRateLimit         = 200
	defaultRateBurst         = 50
	defaultOutboundQueueSize = 4096
	routerResponseTimeout    = 30 * time.Second

	// PriorityHigh, PriorityNormal, PriorityLow define outbound queue priorities.
	PriorityHigh   = 0
	PriorityNormal = 1
	PriorityLow    = 2
)

var (
	ErrRouterClosed     = errors.New("p2p: message router closed")
	ErrNoHandler        = errors.New("p2p: no handler for message code")
	ErrRateLimited      = errors.New("p2p: message rate limited")
	ErrResponseTimeout  = errors.New("p2p: response timeout")
	ErrDuplicateHandler = errors.New("p2p: handler already registered for code")
	ErrQueueFull        = errors.New("p2p: outbound queue full")
	ErrPeerNotTracked   = errors.New("p2p: peer not tracked by router")
	ErrNilHandler       = errors.New("p2p: nil message handler")
)

// RouterHandler handles a message received from a peer.
type RouterHandler func(peerID string, msg wire.Msg) error

// MessageRouter demultiplexes protocol messages, tracks request-response
// correlation, enforces per-peer rate limits, and manages outbound priority queue.
type MessageRouter struct {
	mu        sync.RWMutex
	handlers  map[uint64]RouterHandler
	closed    bool
	reqMu     sync.Mutex
	nextReq   atomic.Uint64
	pending   map[uint64]*routerPendingReq
	rateMu    sync.Mutex
	rates     map[string]*rateLimiter
	rateMax   int
	rateBurst int
	outMu     sync.Mutex
	outQueue  priorityQueue
	outMax    int
	outCond   *sync.Cond
	stats     RouterStats
}

// RouterStats tracks message router statistics.
type RouterStats struct {
	Dispatched, Dropped, RateLimited, Sent atomic.Uint64
}

type routerPendingReq struct {
	id      uint64
	code    uint64
	peerID  string
	created time.Time
	respCh  chan wire.Msg
}

type rateLimiter struct {
	tokens     float64
	maxTokens  float64
	refillRate float64
	lastRefill time.Time
}

func newRateLimiter(rate, burst int) *rateLimiter {
	return &rateLimiter{
		tokens:     float64(burst),
		maxTokens:  float64(burst),
		refillRate: float64(rate),
		lastRefill: time.Now(),
	}
}

func (rl *rateLimiter) allow() bool {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.lastRefill = now
	rl.tokens += elapsed * rl.refillRate
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}

	if rl.tokens < 1.0 {
		return false
	}
	rl.tokens--
	return true
}

// RouterConfig configures the message router.
type RouterConfig struct {
	RateLimit   int // Max messages/sec/peer (0 = default)
	RateBurst   int // Burst allowance (0 = default)
	OutboundMax int // Max outbound queue size (0 = default)
}

// NewMessageRouter creates a message router with the given config.
func NewMessageRouter(cfg RouterConfig) *MessageRouter {
	if cfg.RateLimit <= 0 {
		cfg.RateLimit = defaultRateLimit
	}
	if cfg.RateBurst <= 0 {
		cfg.RateBurst = defaultRateBurst
	}
	if cfg.OutboundMax <= 0 {
		cfg.OutboundMax = defaultOutboundQueueSize
	}

	r := &MessageRouter{
		handlers:  make(map[uint64]RouterHandler),
		pending:   make(map[uint64]*routerPendingReq),
		rates:     make(map[string]*rateLimiter),
		rateMax:   cfg.RateLimit,
		rateBurst: cfg.RateBurst,
		outMax:    cfg.OutboundMax,
	}
	r.outCond = sync.NewCond(&r.outMu)
	heap.Init(&r.outQueue)
	return r
}

// RegisterHandler registers a handler for a message code.
func (r *MessageRouter) RegisterHandler(code uint64, handler RouterHandler) error {
	if handler == nil {
		return ErrNilHandler
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.handlers[code]; exists {
		return fmt.Errorf("%w: 0x%02x", ErrDuplicateHandler, code)
	}
	r.handlers[code] = handler
	return nil
}

// SetHandler sets (or replaces) the handler for a message code.
func (r *MessageRouter) SetHandler(code uint64, handler RouterHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if handler == nil {
		delete(r.handlers, code)
	} else {
		r.handlers[code] = handler
	}
}

// UnregisterHandler removes the handler for a message code.
func (r *MessageRouter) UnregisterHandler(code uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.handlers, code)
}

// HasHandler returns true if a handler is registered for the code.
func (r *MessageRouter) HasHandler(code uint64) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.handlers[code]
	return ok
}

// Dispatch routes an incoming message, checking rate limits and pending requests.
func (r *MessageRouter) Dispatch(peerID string, msg wire.Msg) error {
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return ErrRouterClosed
	}
	r.mu.RUnlock()

	if !r.checkRateLimit(peerID) {
		r.stats.RateLimited.Add(1)
		r.stats.Dropped.Add(1)
		return ErrRateLimited
	}

	if r.deliverResponse(msg) {
		r.stats.Dispatched.Add(1)
		return nil
	}

	r.mu.RLock()
	handler, ok := r.handlers[msg.Code]
	r.mu.RUnlock()

	if !ok {
		r.stats.Dropped.Add(1)
		return fmt.Errorf("%w: 0x%02x", ErrNoHandler, msg.Code)
	}

	r.stats.Dispatched.Add(1)
	return handler(peerID, msg)
}

// SendRequest sends a request and waits for a correlated response.
func (r *MessageRouter) SendRequest(transport wire.Transport, requestCode, responseCode uint64, payload []byte, peerID string) (wire.Msg, error) {
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return wire.Msg{}, ErrRouterClosed
	}
	r.mu.RUnlock()

	reqID := r.nextReq.Add(1)

	pr := &routerPendingReq{
		id:      reqID,
		code:    responseCode,
		peerID:  peerID,
		created: time.Now(),
		respCh:  make(chan wire.Msg, 1),
	}
	r.reqMu.Lock()
	r.pending[reqID] = pr
	r.reqMu.Unlock()

	reqPayload := make([]byte, 8+len(payload))
	putUint64BE(reqPayload[:8], reqID)
	copy(reqPayload[8:], payload)

	if err := transport.WriteMsg(wire.Msg{
		Code:    requestCode,
		Size:    uint32(len(reqPayload)),
		Payload: reqPayload,
	}); err != nil {
		r.reqMu.Lock()
		delete(r.pending, reqID)
		r.reqMu.Unlock()
		return wire.Msg{}, err
	}

	select {
	case resp := <-pr.respCh:
		return resp, nil
	case <-time.After(routerResponseTimeout):
		r.reqMu.Lock()
		delete(r.pending, reqID)
		r.reqMu.Unlock()
		return wire.Msg{}, fmt.Errorf("%w: code=0x%02x id=%d", ErrResponseTimeout, responseCode, reqID)
	}
}

// deliverResponse checks if the message matches a pending request.
func (r *MessageRouter) deliverResponse(msg wire.Msg) bool {
	if len(msg.Payload) < 8 {
		return false
	}

	reqID := getUint64BE(msg.Payload[:8])

	r.reqMu.Lock()
	pr, ok := r.pending[reqID]
	if ok && pr.code == msg.Code {
		delete(r.pending, reqID)
		r.reqMu.Unlock()
		pr.respCh <- wire.Msg{
			Code:    msg.Code,
			Size:    msg.Size - 8,
			Payload: msg.Payload[8:],
		}
		return true
	}
	r.reqMu.Unlock()
	return false
}

func (r *MessageRouter) checkRateLimit(peerID string) bool {
	r.rateMu.Lock()
	defer r.rateMu.Unlock()

	rl, ok := r.rates[peerID]
	if !ok {
		rl = newRateLimiter(r.rateMax, r.rateBurst)
		r.rates[peerID] = rl
	}
	return rl.allow()
}

// TrackPeer initializes rate limiting for a peer.
func (r *MessageRouter) TrackPeer(peerID string) {
	r.rateMu.Lock()
	defer r.rateMu.Unlock()
	if _, ok := r.rates[peerID]; !ok {
		r.rates[peerID] = newRateLimiter(r.rateMax, r.rateBurst)
	}
}

// UntrackPeer removes rate limiting state for a peer.
func (r *MessageRouter) UntrackPeer(peerID string) {
	r.rateMu.Lock()
	defer r.rateMu.Unlock()
	delete(r.rates, peerID)
}

// OutboundMsg wraps an outbound message with priority and metadata.
type OutboundMsg struct {
	Msg      wire.Msg
	PeerID   string
	Priority int // 0 = highest priority.
	Enqueued time.Time
	index    int // heap index
}

// Enqueue adds a message to the outbound priority queue.
func (r *MessageRouter) Enqueue(msg wire.Msg, peerID string, priority int) error {
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return ErrRouterClosed
	}
	r.mu.RUnlock()

	r.outMu.Lock()
	defer r.outMu.Unlock()

	if r.outQueue.Len() >= r.outMax {
		return ErrQueueFull
	}

	item := &OutboundMsg{
		Msg:      msg,
		PeerID:   peerID,
		Priority: priority,
		Enqueued: time.Now(),
	}
	heap.Push(&r.outQueue, item)
	r.outCond.Signal()
	return nil
}

// Dequeue blocks until a message is available or the router is closed.
func (r *MessageRouter) Dequeue() (*OutboundMsg, error) {
	r.outMu.Lock()
	defer r.outMu.Unlock()

	for r.outQueue.Len() == 0 {
		r.mu.RLock()
		closed := r.closed
		r.mu.RUnlock()
		if closed {
			return nil, ErrRouterClosed
		}
		r.outCond.Wait()
	}

	item := heap.Pop(&r.outQueue).(*OutboundMsg)
	r.stats.Sent.Add(1)
	return item, nil
}

// DequeueNonBlocking returns the next outbound message or nil if the queue is empty.
func (r *MessageRouter) DequeueNonBlocking() *OutboundMsg {
	r.outMu.Lock()
	defer r.outMu.Unlock()

	if r.outQueue.Len() == 0 {
		return nil
	}
	item := heap.Pop(&r.outQueue).(*OutboundMsg)
	r.stats.Sent.Add(1)
	return item
}

// QueueLen returns the number of pending outbound messages.
func (r *MessageRouter) QueueLen() int {
	r.outMu.Lock()
	defer r.outMu.Unlock()
	return r.outQueue.Len()
}

// PendingRequests returns the number of in-flight requests.
func (r *MessageRouter) PendingRequests() int {
	r.reqMu.Lock()
	defer r.reqMu.Unlock()
	return len(r.pending)
}

// ExpireRequests cancels pending requests older than timeout.
func (r *MessageRouter) ExpireRequests(timeout time.Duration) int {
	r.reqMu.Lock()
	defer r.reqMu.Unlock()

	now := time.Now()
	expired := 0
	for id, pr := range r.pending {
		if now.Sub(pr.created) > timeout {
			delete(r.pending, id)
			close(pr.respCh)
			expired++
		}
	}
	return expired
}

// Stats returns router statistics.
func (r *MessageRouter) Stats() (dispatched, dropped, rateLimited, sent uint64) {
	return r.stats.Dispatched.Load(), r.stats.Dropped.Load(),
		r.stats.RateLimited.Load(), r.stats.Sent.Load()
}

// HandlerCount returns the number of registered handlers.
func (r *MessageRouter) HandlerCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.handlers)
}

// Close shuts down the router.
func (r *MessageRouter) Close() {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	r.closed = true
	r.mu.Unlock()

	r.outCond.Broadcast()
}

type priorityQueue []*OutboundMsg

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	if pq[i].Priority != pq[j].Priority {
		return pq[i].Priority < pq[j].Priority
	}
	return pq[i].Enqueued.Before(pq[j].Enqueued)
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*OutboundMsg)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[:n-1]
	return item
}

func putUint64BE(b []byte, v uint64) {
	b[0] = byte(v >> 56)
	b[1] = byte(v >> 48)
	b[2] = byte(v >> 40)
	b[3] = byte(v >> 32)
	b[4] = byte(v >> 24)
	b[5] = byte(v >> 16)
	b[6] = byte(v >> 8)
	b[7] = byte(v)
}

func getUint64BE(b []byte) uint64 {
	return uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 |
		uint64(b[3])<<32 | uint64(b[4])<<24 | uint64(b[5])<<16 |
		uint64(b[6])<<8 | uint64(b[7])
}
