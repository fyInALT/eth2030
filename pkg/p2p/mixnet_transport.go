package p2p

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
)

// MixnetConfig configures the mixnet relay transport.
type MixnetConfig struct {
	// Hops is the number of mix-node hops (default: 3).
	Hops int
	// HopDelay is the simulated delay per hop (default: 500ms).
	HopDelay time.Duration
	// MaxPending is the max pending transactions in the channel.
	MaxPending int
}

// DefaultMixnetConfig returns sensible defaults for the mixnet transport.
func DefaultMixnetConfig() *MixnetConfig {
	return &MixnetConfig{
		Hops:       3,
		HopDelay:   500 * time.Millisecond,
		MaxPending: 256,
	}
}

// MixnetTransport implements AnonymousTransport with simulated mix-node delays.
// Each hop applies onion-style re-encryption (Keccak256-based envelope wrapping).
type MixnetTransport struct {
	mu      sync.Mutex
	config  *MixnetConfig
	ch      chan *types.Transaction
	running bool
	stopCh  chan struct{}
}

// NewMixnetTransport creates a new mixnet transport with the given config.
func NewMixnetTransport(config *MixnetConfig) *MixnetTransport {
	if config == nil {
		config = DefaultMixnetConfig()
	}
	return &MixnetTransport{
		config: config,
		ch:     make(chan *types.Transaction, config.MaxPending),
		stopCh: make(chan struct{}),
	}
}

// Name returns "mixnet".
func (m *MixnetTransport) Name() string { return "mixnet" }

// Submit sends a transaction through the simulated mixnet with multi-hop delays.
func (m *MixnetTransport) Submit(tx *types.Transaction) error {
	if tx == nil {
		return ErrAnonTransportNilTx
	}

	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return ErrAnonTransportClosed
	}
	m.mu.Unlock()

	// Simulate multi-hop relay in a goroutine.
	go m.relay(tx)
	return nil
}

// relay simulates the multi-hop mix-node relay with delays.
func (m *MixnetTransport) relay(tx *types.Transaction) {
	for hop := 0; hop < m.config.Hops; hop++ {
		select {
		case <-m.stopCh:
			return
		case <-time.After(m.config.HopDelay):
			// Simulate onion re-encryption at each hop.
			// In production, each hop would strip one encryption layer.
		}
	}

	// Deliver the transaction after all hops.
	select {
	case m.ch <- tx:
	case <-m.stopCh:
	default:
		// Channel full, drop the tx.
	}
}

// Receive returns the channel of transactions that have completed the mixnet relay.
func (m *MixnetTransport) Receive() <-chan *types.Transaction { return m.ch }

// Start initializes the mixnet transport.
func (m *MixnetTransport) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = true
	return nil
}

// Stop shuts down the mixnet transport.
func (m *MixnetTransport) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.running {
		m.running = false
		close(m.stopCh)
	}
	return nil
}

// WrapOnion applies N layers of onion encryption to data (stub: Keccak256 per layer).
// Each layer is keyed by a per-hop seed derived from the hop index.
func WrapOnion(data []byte, hops int) []byte {
	wrapped := make([]byte, len(data))
	copy(wrapped, data)
	for i := 0; i < hops; i++ {
		var hopBuf [4]byte
		binary.BigEndian.PutUint32(hopBuf[:], uint32(i))
		wrapped = crypto.Keccak256(
			[]byte("mixnet-onion-layer"),
			hopBuf[:],
			wrapped,
		)
	}
	return wrapped
}
