package epbs

import (
	"errors"
	"sync"
)

var errBuilderNotFound = errors.New("builder withdrawal: builder not registered")

// EIP-7732 §builder-withdrawal constants.
const (
	// MinBuilderWithdrawabilityDelay is the minimum epoch delay before a
	// builder withdrawal is processed (64 epochs per EIP-7732).
	MinBuilderWithdrawabilityDelay uint64 = 64

	// MaxBuildersPerWithdrawalsSweep is the maximum number of builders swept
	// per epoch for withdrawal processing (EIP-7732 §withdrawal-sweep).
	MaxBuildersPerWithdrawalsSweep = 16384

	// BuilderWithdrawalPrefix is the withdrawal credentials prefix for
	// builders (0x03 per EIP-7732 §withdrawal-prefix).
	BuilderWithdrawalPrefix byte = 0x03
)

// builderWithdrawalEntry holds withdrawal state for a registered builder.
type builderWithdrawalEntry struct {
	Address          [20]byte
	Balance          uint64
	WithdrawableAt   uint64 // epoch when withdrawal is allowed (0 = not requested)
	WithdrawalAmount uint64
}

// BuilderWithdrawalRegistry tracks builder balances and pending withdrawals.
type BuilderWithdrawalRegistry struct {
	mu      sync.Mutex
	entries map[[20]byte]*builderWithdrawalEntry
	order   [][20]byte // insertion order for deterministic sweep
}

// NewBuilderWithdrawalRegistry creates a new registry.
func NewBuilderWithdrawalRegistry() *BuilderWithdrawalRegistry {
	return &BuilderWithdrawalRegistry{
		entries: make(map[[20]byte]*builderWithdrawalEntry),
	}
}

// AddBuilder registers a builder with an initial balance.
func (r *BuilderWithdrawalRegistry) AddBuilder(addr [20]byte, balance uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.entries[addr]; !ok {
		r.order = append(r.order, addr)
	}
	r.entries[addr] = &builderWithdrawalEntry{Address: addr, Balance: balance}
}

// RequestWithdrawal schedules a builder withdrawal with the 64-epoch delay.
// Returns an error if the builder is not registered.
func (r *BuilderWithdrawalRegistry) RequestWithdrawal(addr [20]byte, amount uint64, currentEpoch uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.entries[addr]
	if !ok {
		return errBuilderNotFound
	}
	e.WithdrawableAt = currentEpoch + MinBuilderWithdrawabilityDelay
	e.WithdrawalAmount = amount
	return nil
}

// WithdrawableBuilders returns up to MaxBuildersPerWithdrawalsSweep builders
// whose withdrawal delay has elapsed at the given epoch.
func (r *BuilderWithdrawalRegistry) WithdrawableBuilders(atEpoch uint64) []*builderWithdrawalEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	var ready []*builderWithdrawalEntry
	for _, addr := range r.order {
		e := r.entries[addr]
		if e.WithdrawableAt > 0 && atEpoch >= e.WithdrawableAt {
			ready = append(ready, e)
			if len(ready) >= MaxBuildersPerWithdrawalsSweep {
				break
			}
		}
	}
	return ready
}
