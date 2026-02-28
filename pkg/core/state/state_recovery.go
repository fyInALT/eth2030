// state_recovery.go implements state recovery from corruption with checkpoint
// rollback, consistency validation, and corruption detection. This supports
// the sustainability track of the L1 strawmap by ensuring state integrity
// under gigagas workloads where corruption risk increases.
package state

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Recovery errors.
var (
	ErrNoCheckpoints       = errors.New("state_recovery: no checkpoints available")
	ErrNoValidTarget       = errors.New("state_recovery: no valid rollback target before corrupt block")
	ErrCheckpointTooOld    = errors.New("state_recovery: checkpoint too old for rollback")
	ErrMaxRollbackExceeded = errors.New("state_recovery: max rollback depth exceeded")
	ErrRollbackDisabled    = errors.New("state_recovery: auto-repair is disabled")
	ErrDuplicateCheckpoint = errors.New("state_recovery: duplicate checkpoint block number")
)

// CorruptionType categorizes the kind of state corruption detected.
type CorruptionType int

const (
	// CorruptionAccountMismatch indicates account data does not match expected.
	CorruptionAccountMismatch CorruptionType = iota
	// CorruptionStorageMismatch indicates a storage slot does not match expected.
	CorruptionStorageMismatch
	// CorruptionDanglingStorage indicates storage exists for a nonexistent account.
	CorruptionDanglingStorage
)

// CorruptionError is a structured error describing a state inconsistency.
type CorruptionError struct {
	Type    CorruptionType
	Address [20]byte
	Key     [32]byte // only for storage mismatches
	Message string
}

func (e *CorruptionError) Error() string { return e.Message }

// StateRecoveryConfig holds configuration for the recovery subsystem.
type StateRecoveryConfig struct {
	// MaxRollbackDepth is the maximum number of blocks we will roll back.
	MaxRollbackDepth uint64
	// ConsistencyCheckInterval is blocks between consistency checks (0 = disabled).
	ConsistencyCheckInterval uint64
	// AutoRepair enables automatic rollback on corruption detection.
	AutoRepair bool
	// MaxCheckpoints is the maximum number of checkpoints to retain.
	MaxCheckpoints int
}

// DefaultRecoveryConfig returns sensible defaults for state recovery.
func DefaultRecoveryConfig() StateRecoveryConfig {
	return StateRecoveryConfig{
		MaxRollbackDepth:         1024,
		ConsistencyCheckInterval: 256,
		AutoRepair:               true,
		MaxCheckpoints:           64,
	}
}

// RecoveryCheckpoint represents a known-good state snapshot at a specific block.
type RecoveryCheckpoint struct {
	// BlockNumber is the block at which this checkpoint was taken.
	BlockNumber uint64
	// StateRoot is the state trie root at this block.
	StateRoot [32]byte
	// Timestamp is when the checkpoint was created.
	Timestamp time.Time
}

// CorruptionDetector detects inconsistencies in the Ethereum state. It
// accumulates errors found during scanning so callers can inspect all
// issues at once rather than stopping at the first one.
type CorruptionDetector struct {
	mu     sync.Mutex
	errors []*CorruptionError
}

// NewCorruptionDetector creates a new corruption detector.
func NewCorruptionDetector() *CorruptionDetector {
	return &CorruptionDetector{}
}

// CheckAccountConsistency verifies that the actual account data matches
// the expected data. Both are raw encoded bytes; nil expected means the
// account should not exist.
func (cd *CorruptionDetector) CheckAccountConsistency(addr [20]byte, expected, actual []byte) error {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	if bytes.Equal(expected, actual) {
		return nil
	}

	var msg string
	switch {
	case expected == nil && actual != nil:
		msg = fmt.Sprintf("account %x: unexpected account exists (%d bytes)", addr, len(actual))
	case expected != nil && actual == nil:
		msg = fmt.Sprintf("account %x: expected account missing (%d bytes expected)", addr, len(expected))
	default:
		msg = fmt.Sprintf("account %x: data mismatch (expected %d bytes, got %d bytes)", addr, len(expected), len(actual))
	}

	ce := &CorruptionError{
		Type:    CorruptionAccountMismatch,
		Address: addr,
		Message: msg,
	}
	cd.errors = append(cd.errors, ce)
	return ce
}

// CheckStorageConsistency verifies that a storage slot matches its expected value.
func (cd *CorruptionDetector) CheckStorageConsistency(addr [20]byte, key, expected, actual [32]byte) error {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	if expected == actual {
		return nil
	}

	ce := &CorruptionError{
		Type:    CorruptionStorageMismatch,
		Address: addr,
		Key:     key,
		Message: fmt.Sprintf("account %x storage %x: value mismatch", addr, key),
	}
	cd.errors = append(cd.errors, ce)
	return ce
}

// DetectDanglingStorage checks for storage keys that exist under an account
// that should not have storage (e.g., the account has been deleted). Returns
// one error per dangling key found.
func (cd *CorruptionDetector) DetectDanglingStorage(addr [20]byte, storageKeys [][32]byte, accountExists bool) []error {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	if accountExists || len(storageKeys) == 0 {
		return nil
	}

	var errs []error
	for _, key := range storageKeys {
		ce := &CorruptionError{
			Type:    CorruptionDanglingStorage,
			Address: addr,
			Key:     key,
			Message: fmt.Sprintf("account %x: dangling storage key %x (account does not exist)", addr, key),
		}
		cd.errors = append(cd.errors, ce)
		errs = append(errs, ce)
	}
	return errs
}

// Errors returns all accumulated corruption errors.
func (cd *CorruptionDetector) Errors() []*CorruptionError {
	cd.mu.Lock()
	defer cd.mu.Unlock()
	result := make([]*CorruptionError, len(cd.errors))
	copy(result, cd.errors)
	return result
}

// HasErrors returns true if any corruption has been detected.
func (cd *CorruptionDetector) HasErrors() bool {
	cd.mu.Lock()
	defer cd.mu.Unlock()
	return len(cd.errors) > 0
}

// Reset clears all accumulated errors.
func (cd *CorruptionDetector) Reset() {
	cd.mu.Lock()
	defer cd.mu.Unlock()
	cd.errors = nil
}

// ErrorCount returns the number of errors detected.
func (cd *CorruptionDetector) ErrorCount() int {
	cd.mu.Lock()
	defer cd.mu.Unlock()
	return len(cd.errors)
}

// RollbackManager manages checkpoints and rolling back state to a previous
// consistent point. Checkpoints are ordered by block number and pruned
// when the maximum count is exceeded.
type RollbackManager struct {
	mu          sync.Mutex
	cfg         StateRecoveryConfig
	checkpoints []RecoveryCheckpoint
}

// NewRollbackManager creates a new rollback manager with the given config.
func NewRollbackManager(cfg StateRecoveryConfig) *RollbackManager {
	return &RollbackManager{
		cfg:         cfg,
		checkpoints: make([]RecoveryCheckpoint, 0),
	}
}

// AddCheckpoint adds a checkpoint. If a checkpoint at the same block already
// exists, it is updated. The list is kept sorted by block number. If adding
// exceeds MaxCheckpoints, the oldest checkpoint is pruned.
func (rm *RollbackManager) AddCheckpoint(cp RecoveryCheckpoint) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Check for duplicate block number and update if found.
	for i, existing := range rm.checkpoints {
		if existing.BlockNumber == cp.BlockNumber {
			rm.checkpoints[i] = cp
			return
		}
	}

	rm.checkpoints = append(rm.checkpoints, cp)

	// Keep sorted by block number ascending.
	sort.Slice(rm.checkpoints, func(i, j int) bool {
		return rm.checkpoints[i].BlockNumber < rm.checkpoints[j].BlockNumber
	})

	// Prune if we exceed max checkpoints.
	if rm.cfg.MaxCheckpoints > 0 && len(rm.checkpoints) > rm.cfg.MaxCheckpoints {
		excess := len(rm.checkpoints) - rm.cfg.MaxCheckpoints
		rm.checkpoints = rm.checkpoints[excess:]
	}
}

// FindRollbackTarget finds the most recent checkpoint before the given corrupt
// block number that is within the max rollback depth.
func (rm *RollbackManager) FindRollbackTarget(corruptBlock uint64) (*RecoveryCheckpoint, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if len(rm.checkpoints) == 0 {
		return nil, ErrNoCheckpoints
	}

	// Search backwards for the latest checkpoint before the corrupt block.
	var best *RecoveryCheckpoint
	for i := len(rm.checkpoints) - 1; i >= 0; i-- {
		cp := rm.checkpoints[i]
		if cp.BlockNumber < corruptBlock {
			best = &cp
			break
		}
	}

	if best == nil {
		return nil, ErrNoValidTarget
	}

	// Check rollback depth.
	depth := corruptBlock - best.BlockNumber
	if rm.cfg.MaxRollbackDepth > 0 && depth > rm.cfg.MaxRollbackDepth {
		return nil, fmt.Errorf("%w: depth %d exceeds max %d", ErrMaxRollbackExceeded, depth, rm.cfg.MaxRollbackDepth)
	}

	return best, nil
}

// PruneOldCheckpoints removes all checkpoints at or before the given block
// number. Returns the number of checkpoints removed.
func (rm *RollbackManager) PruneOldCheckpoints(keepAfter uint64) int {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	removed := 0
	kept := make([]RecoveryCheckpoint, 0, len(rm.checkpoints))
	for _, cp := range rm.checkpoints {
		if cp.BlockNumber <= keepAfter {
			removed++
		} else {
			kept = append(kept, cp)
		}
	}
	rm.checkpoints = kept
	return removed
}

// LatestCheckpoint returns the most recent checkpoint, or nil if none exist.
func (rm *RollbackManager) LatestCheckpoint() *RecoveryCheckpoint {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if len(rm.checkpoints) == 0 {
		return nil
	}
	cp := rm.checkpoints[len(rm.checkpoints)-1]
	return &cp
}

// OldestCheckpoint returns the oldest checkpoint, or nil if none exist.
func (rm *RollbackManager) OldestCheckpoint() *RecoveryCheckpoint {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if len(rm.checkpoints) == 0 {
		return nil
	}
	cp := rm.checkpoints[0]
	return &cp
}

// CheckpointCount returns the number of stored checkpoints.
func (rm *RollbackManager) CheckpointCount() int {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return len(rm.checkpoints)
}

// CheckpointAtBlock returns the checkpoint at the given block, or nil.
func (rm *RollbackManager) CheckpointAtBlock(block uint64) *RecoveryCheckpoint {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for _, cp := range rm.checkpoints {
		if cp.BlockNumber == block {
			return &cp
		}
	}
	return nil
}

// ShouldCheck returns true if a consistency check is due at the given block.
func (rm *RollbackManager) ShouldCheck(blockNumber uint64) bool {
	if rm.cfg.ConsistencyCheckInterval == 0 {
		return false
	}
	return blockNumber%rm.cfg.ConsistencyCheckInterval == 0
}

// Config returns the recovery configuration.
func (rm *RollbackManager) Config() StateRecoveryConfig {
	return rm.cfg
}
