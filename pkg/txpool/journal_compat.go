package txpool

// journal_compat.go re-exports types from txpool/journal for backward compatibility.

import (
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/txpool/journal"
)

// Journal type aliases.
type (
	JournalEntry = journal.JournalEntry
	TxJournal    = journal.TxJournal
	JrnlConfig   = journal.JrnlConfig
	JrnlMetrics  = journal.JrnlMetrics
	JrnlError    = journal.JrnlError
	TxJrnl       = journal.TxJrnl
)

// Journal error variables.
var (
	ErrJournalClosed   = journal.ErrJournalClosed
	ErrJournalCorrupt  = journal.ErrJournalCorrupt
	ErrJournalNotFound = journal.ErrJournalNotFound
	ErrJrnlClosed      = journal.ErrJrnlClosed
	ErrJrnlEmpty       = journal.ErrJrnlEmpty
)

// Journal function wrappers.
func NewTxJournal(path string) (*TxJournal, error) { return journal.NewTxJournal(path) }
func Load(path string) ([]*types.Transaction, []JournalEntry, error) {
	return journal.Load(path)
}
func DefaultJrnlConfig() JrnlConfig                    { return journal.DefaultJrnlConfig() }
func NewTxJrnl(config JrnlConfig) (*TxJrnl, error)    { return journal.NewTxJrnl(config) }
func ReplayJrnl(path string, metrics *JrnlMetrics) ([]*types.Transaction, error) {
	return journal.ReplayJrnl(path, metrics)
}
func ValidateJournal(path string) (total int, corrupt int, firstErr error) {
	return journal.ValidateJournal(path)
}
