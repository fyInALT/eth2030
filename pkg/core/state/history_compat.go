package state

// history_compat.go re-exports types from core/state/history for backward compatibility.

import "github.com/eth2030/eth2030/core/state/history"

// History type aliases.
type (
	AccountHistoryEntry = history.AccountHistoryEntry
	StorageHistoryEntry = history.StorageHistoryEntry
	HistoryRange        = history.HistoryRange
	StateHistoryReader  = history.StateHistoryReader
)

// History error variables.
var (
	ErrBlockNotInRange    = history.ErrBlockNotInRange
	ErrNoHistoryAvailable = history.ErrNoHistoryAvailable
	ErrHistoryPruned      = history.ErrHistoryPruned
	ErrInvalidPruneRange  = history.ErrInvalidPruneRange
)

// History function wrappers.
func NewStateHistoryReader(retentionWindow uint64) *StateHistoryReader {
	return history.NewStateHistoryReader(retentionWindow)
}
