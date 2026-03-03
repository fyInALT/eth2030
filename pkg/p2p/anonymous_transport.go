package p2p

import (
	"errors"
	"sync"

	"github.com/eth2030/eth2030/core/types"
)

// Anonymous transport errors.
var (
	ErrAnonTransportClosed   = errors.New("anon_transport: transport is closed")
	ErrAnonTransportExists   = errors.New("anon_transport: transport already registered")
	ErrAnonTransportNotFound = errors.New("anon_transport: transport not found")
	ErrAnonTransportNilTx    = errors.New("anon_transport: nil transaction")
)

// AnonymousTransport is the interface for anonymous transaction submission.
// Implementations hide the sender's IP address from the P2P network.
type AnonymousTransport interface {
	// Name returns the transport identifier (e.g., "tor", "mixnet", "flashnet").
	Name() string
	// Submit sends a transaction through the anonymous transport.
	Submit(tx *types.Transaction) error
	// Receive returns a channel of transactions received via this transport.
	Receive() <-chan *types.Transaction
	// Start initializes the transport.
	Start() error
	// Stop shuts down the transport.
	Stop() error
}

// TransportStats holds statistics for an anonymous transport.
type TransportStats struct {
	Name      string
	Submitted uint64
	Received  uint64
	Errors    uint64
	Running   bool
}

// TransportManager manages multiple anonymous transports and provides
// a unified interface for anonymous transaction submission.
type TransportManager struct {
	mu         sync.RWMutex
	transports map[string]AnonymousTransport
	stats      map[string]*TransportStats
	closed     bool
}

// NewTransportManager creates a new transport manager.
func NewTransportManager() *TransportManager {
	return &TransportManager{
		transports: make(map[string]AnonymousTransport),
		stats:      make(map[string]*TransportStats),
	}
}

// RegisterTransport adds an anonymous transport to the manager.
func (tm *TransportManager) RegisterTransport(t AnonymousTransport) error {
	if t == nil {
		return ErrAnonTransportNilTx
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.closed {
		return ErrAnonTransportClosed
	}

	name := t.Name()
	if _, exists := tm.transports[name]; exists {
		return ErrAnonTransportExists
	}

	tm.transports[name] = t
	tm.stats[name] = &TransportStats{Name: name}
	return nil
}

// UnregisterTransport removes a transport from the manager.
func (tm *TransportManager) UnregisterTransport(name string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	t, exists := tm.transports[name]
	if !exists {
		return ErrAnonTransportNotFound
	}

	_ = t.Stop()
	delete(tm.transports, name)
	delete(tm.stats, name)
	return nil
}

// SubmitAnonymous submits a tx via all registered transports.
// Returns the number of successful submissions and any errors.
func (tm *TransportManager) SubmitAnonymous(tx *types.Transaction) (int, []error) {
	if tx == nil {
		return 0, []error{ErrAnonTransportNilTx}
	}

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var errs []error
	submitted := 0
	for name, t := range tm.transports {
		if err := t.Submit(tx); err != nil {
			errs = append(errs, err)
			if s := tm.stats[name]; s != nil {
				s.Errors++
			}
		} else {
			submitted++
			if s := tm.stats[name]; s != nil {
				s.Submitted++
			}
		}
	}
	return submitted, errs
}

// StartAll starts all registered transports.
func (tm *TransportManager) StartAll() []error {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var errs []error
	for name, t := range tm.transports {
		if err := t.Start(); err != nil {
			errs = append(errs, err)
		} else if s := tm.stats[name]; s != nil {
			s.Running = true
		}
	}
	return errs
}

// StopAll stops all registered transports.
func (tm *TransportManager) StopAll() []error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.closed = true
	var errs []error
	for name, t := range tm.transports {
		if err := t.Stop(); err != nil {
			errs = append(errs, err)
		}
		if s := tm.stats[name]; s != nil {
			s.Running = false
		}
	}
	return errs
}

// GetStats returns a copy of the stats for all transports.
func (tm *TransportManager) GetStats() []TransportStats {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var result []TransportStats
	for _, s := range tm.stats {
		result = append(result, *s)
	}
	return result
}

// TransportCount returns the number of registered transports.
func (tm *TransportManager) TransportCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.transports)
}
