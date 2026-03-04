package core

import (
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/eth2030/eth2030/core/types"
)

// Paymaster registry constants.
const (
	DefaultMinStakeWei   = 1_000_000_000_000_000_000 // 1 ETH in wei
	DefaultMaxSlashes    = 3
	DefaultWithdrawDelay = 256 // blocks
)

var (
	ErrUnstakedPaymaster  = errors.New("paymaster: insufficient stake")
	ErrBannedPaymaster    = errors.New("paymaster: banned due to excessive slashing")
	ErrWithdrawCooldown   = errors.New("paymaster: withdrawal cooldown not elapsed")
	ErrInvalidWithdrawAmt = errors.New("paymaster: withdrawal amount exceeds stake")
	ErrInvalidDeposit     = errors.New("paymaster: deposit amount must be positive")
)

// PaymasterEntry tracks a single paymaster's stake and slash state.
type PaymasterEntry struct {
	Stake            *big.Int
	SlashCount       int
	Banned           bool
	WithdrawReqBlock uint64 // 0 = no pending request
}

// RegistryConfig configures the PaymasterRegistry.
type RegistryConfig struct {
	MinStake      *big.Int // minimum required stake (wei)
	MaxSlashes    int      // slashes before ban
	WithdrawDelay uint64   // cooldown in blocks
}

// DefaultRegistryConfig returns production defaults.
func DefaultRegistryConfig() RegistryConfig {
	return RegistryConfig{
		MinStake:      new(big.Int).SetUint64(DefaultMinStakeWei),
		MaxSlashes:    DefaultMaxSlashes,
		WithdrawDelay: DefaultWithdrawDelay,
	}
}

// PaymasterRegistry tracks stakes and slash counts for EIP-8141 paymasters.
type PaymasterRegistry struct {
	mu         sync.RWMutex
	entries    map[types.Address]*PaymasterEntry
	minStake   *big.Int
	maxSlashes int
	delay      uint64
}

// NewPaymasterRegistry creates a registry with the given config.
func NewPaymasterRegistry(cfg RegistryConfig) *PaymasterRegistry {
	minStake := cfg.MinStake
	if minStake == nil || minStake.Sign() <= 0 {
		minStake = new(big.Int).SetUint64(DefaultMinStakeWei)
	}
	maxSlashes := cfg.MaxSlashes
	if maxSlashes <= 0 {
		maxSlashes = DefaultMaxSlashes
	}
	return &PaymasterRegistry{
		entries:    make(map[types.Address]*PaymasterEntry),
		minStake:   new(big.Int).Set(minStake),
		maxSlashes: maxSlashes,
		delay:      cfg.WithdrawDelay,
	}
}

// Deposit adds stake for addr. blockNum is recorded for future use (e.g. events).
func (r *PaymasterRegistry) Deposit(addr types.Address, amount *big.Int, blockNum uint64) error {
	if amount == nil || amount.Sign() <= 0 {
		return ErrInvalidDeposit
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	e := r.entry(addr)
	e.Stake = new(big.Int).Add(e.Stake, amount)
	return nil
}

// RequestWithdraw records a withdrawal intent at blockNum, starting the cooldown timer.
func (r *PaymasterRegistry) RequestWithdraw(addr types.Address, blockNum uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if e, ok := r.entries[addr]; ok {
		e.WithdrawReqBlock = blockNum
	}
}

// Withdraw removes amount of stake from addr after the cooldown has elapsed.
// Returns ErrWithdrawCooldown if called too early, ErrInvalidWithdrawAmt if
// amount exceeds the current stake, or ErrUnstakedPaymaster if addr has no stake.
func (r *PaymasterRegistry) Withdraw(addr types.Address, amount *big.Int, currentBlock uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	e, ok := r.entries[addr]
	if !ok || e.Stake.Sign() == 0 {
		return ErrUnstakedPaymaster
	}
	if e.WithdrawReqBlock == 0 || currentBlock < e.WithdrawReqBlock+r.delay {
		return ErrWithdrawCooldown
	}
	if amount.Cmp(e.Stake) > 0 {
		return ErrInvalidWithdrawAmt
	}
	e.Stake = new(big.Int).Sub(e.Stake, amount)
	e.WithdrawReqBlock = 0
	return nil
}

// Slash increments the slash counter for addr. Once SlashCount reaches MaxSlashes
// the paymaster is banned and IsApprovedPaymaster will return false.
func (r *PaymasterRegistry) Slash(addr types.Address) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e := r.entry(addr)
	e.SlashCount++
	if e.SlashCount >= r.maxSlashes {
		e.Banned = true
	}
}

// IsApprovedPaymaster returns true iff addr has stake >= MinStake and is not banned.
func (r *PaymasterRegistry) IsApprovedPaymaster(addr types.Address) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[addr]
	if !ok {
		return false
	}
	return !e.Banned && e.Stake.Cmp(r.minStake) >= 0
}

// Entry returns a copy of the PaymasterEntry for addr, or nil if not registered.
func (r *PaymasterRegistry) Entry(addr types.Address) *PaymasterEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[addr]
	if !ok {
		return nil
	}
	cp := *e
	cp.Stake = new(big.Int).Set(e.Stake)
	return &cp
}

// entry returns (creating if needed) the mutable PaymasterEntry for addr.
// Caller must hold r.mu (write lock).
func (r *PaymasterRegistry) entry(addr types.Address) *PaymasterEntry {
	e, ok := r.entries[addr]
	if !ok {
		e = &PaymasterEntry{Stake: new(big.Int)}
		r.entries[addr] = e
	}
	return e
}

// SlashOnBadSettlement is called by the frame processor when a paymaster-approved
// frame tx fails to cover gas. It increments the slash counter and, once the
// paymaster is banned, returns a wrapped ErrBannedPaymaster.
func (r *PaymasterRegistry) SlashOnBadSettlement(addr types.Address) error {
	r.Slash(addr)
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[addr]
	if ok && e.Banned {
		return fmt.Errorf("%w: %s", ErrBannedPaymaster, addr.Hex())
	}
	return nil
}
