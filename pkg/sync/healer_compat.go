package sync

// healer_compat.go re-exports types from sync/healer for backward compatibility.

import (
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/sync/healer"
)

// Healer type aliases.
type (
	HealingTask            = healer.HealingTask
	HealingProgress        = healer.HealingProgress
	StateHealer            = healer.StateHealer
	TrieSync               = healer.TrieSync
	HealPriority           = healer.HealPriority
	HealTask               = healer.HealTask
	GapResult              = healer.GapResult
	GapFinder              = healer.GapFinder
	NodeVerifier           = healer.NodeVerifier
	HealScheduler          = healer.HealScheduler
	ConcurrentHealProgress = healer.ConcurrentHealProgress
	ConcurrentHealConfig   = healer.ConcurrentHealConfig
	ConcurrentTrieHealer   = healer.ConcurrentTrieHealer
	TrieHealNode           = healer.TrieHealNode
	TrieHealCheckpoint     = healer.TrieHealCheckpoint
	TrieHealProgress       = healer.TrieHealProgress
	TrieHealConfig         = healer.TrieHealConfig
	TrieHealer             = healer.TrieHealer
)

// Healer constants.
const (
	PriorityHigh   = healer.PriorityHigh
	PriorityMedium = healer.PriorityMedium
	PriorityLow    = healer.PriorityLow
)

// Healer error variables.
var (
	ErrHealerClosed        = healer.ErrHealerClosed
	ErrHealerRunning       = healer.ErrHealerRunning
	ErrNoGapsFound         = healer.ErrNoGapsFound
	ErrHealBatchEmpty      = healer.ErrHealBatchEmpty
	ErrHealNodeInvalid     = healer.ErrHealNodeInvalid
	ErrConcHealerClosed    = healer.ErrConcHealerClosed
	ErrConcHealerRunning   = healer.ErrConcHealerRunning
	ErrNodeVerifyFailed    = healer.ErrNodeVerifyFailed
	ErrGapFinderNoRoot     = healer.ErrGapFinderNoRoot
	ErrAlreadyProcessed    = healer.ErrAlreadyProcessed
	ErrNotRequested        = healer.ErrNotRequested
	ErrHashMismatch        = healer.ErrHashMismatch
	ErrTrieHealerClosed    = healer.ErrTrieHealerClosed
	ErrTrieHealerRunning   = healer.ErrTrieHealerRunning
	ErrTrieHealNoPeer      = healer.ErrTrieHealNoPeer
	ErrTrieHealCheckpoint  = healer.ErrTrieHealCheckpoint
	ErrTrieHealInvalidNode = healer.ErrTrieHealInvalidNode
)

// Healer function wrappers.
func NewStateHealer(root types.Hash, writer healer.StateWriter) *StateHealer {
	return healer.NewStateHealer(root, writer)
}
func NewGapFinder(writer healer.StateWriter) *GapFinder {
	return healer.NewGapFinder(writer)
}
func NewNodeVerifier() *NodeVerifier   { return healer.NewNodeVerifier() }
func NewHealScheduler() *HealScheduler { return healer.NewHealScheduler() }
func DefaultConcurrentHealConfig() ConcurrentHealConfig {
	return healer.DefaultConcurrentHealConfig()
}
func NewConcurrentTrieHealer(config ConcurrentHealConfig, root types.Hash, writer healer.StateWriter) *ConcurrentTrieHealer {
	return healer.NewConcurrentTrieHealer(config, root, writer)
}
func DefaultTrieHealConfig() TrieHealConfig { return healer.DefaultTrieHealConfig() }
func NewTrieHealer(config TrieHealConfig, root types.Hash, writer healer.StateWriter) *TrieHealer {
	return healer.NewTrieHealer(config, root, writer)
}
func FormatETA(d interface{}) string { return "" }
