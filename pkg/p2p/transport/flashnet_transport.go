package transport

import (
	"sync"
	"sync/atomic"

	"github.com/eth2030/eth2030/core/types"
)

// FlashnetConfig configures the Flashnet broadcast transport.
type FlashnetConfig struct {
	// MaxPeers is the maximum number of Flashnet peers.
	MaxPeers int
	// MaxPending is the max pending transactions in the channel.
	MaxPending int
}

// DefaultFlashnetConfig returns default Flashnet configuration.
func DefaultFlashnetConfig() *FlashnetConfig {
	return &FlashnetConfig{
		MaxPeers:   64,
		MaxPending: 512,
	}
}

// FlashnetTransport implements AnonymousTransport with encrypted broadcast.
// Transactions are broadcast to all known Flashnet peers simultaneously
// with ephemeral DH key per message. Faster than Mixnet but weaker anonymity.
type FlashnetTransport struct {
	mu      sync.Mutex
	config  *FlashnetConfig
	ch      chan *types.Transaction
	running bool

	// Metrics.
	submitted atomic.Uint64
	broadcast atomic.Uint64
}

// NewFlashnetTransport creates a new Flashnet transport.
func NewFlashnetTransport(config *FlashnetConfig) *FlashnetTransport {
	if config == nil {
		config = DefaultFlashnetConfig()
	}
	return &FlashnetTransport{
		config: config,
		ch:     make(chan *types.Transaction, config.MaxPending),
	}
}

// Name returns "flashnet".
func (f *FlashnetTransport) Name() string { return "flashnet" }

// Submit broadcasts a transaction to all Flashnet peers.
// Each broadcast uses an ephemeral encryption key (simulated).
func (f *FlashnetTransport) Submit(tx *types.Transaction) error {
	if tx == nil {
		return ErrAnonTransportNilTx
	}

	f.mu.Lock()
	if !f.running {
		f.mu.Unlock()
		return ErrAnonTransportClosed
	}
	f.mu.Unlock()

	f.submitted.Add(1)

	// Broadcast immediately (no delay unlike mixnet).
	select {
	case f.ch <- tx:
		f.broadcast.Add(1)
	default:
		// Channel full.
	}
	return nil
}

// Receive returns the channel of received Flashnet transactions.
func (f *FlashnetTransport) Receive() <-chan *types.Transaction { return f.ch }

// Start initializes the Flashnet transport.
func (f *FlashnetTransport) Start() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.running = true
	return nil
}

// Stop shuts down the Flashnet transport.
func (f *FlashnetTransport) Stop() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.running = false
	return nil
}

// Stats returns the number of submitted and broadcast transactions.
func (f *FlashnetTransport) Stats() (submitted, broadcast uint64) {
	return f.submitted.Load(), f.broadcast.Load()
}
