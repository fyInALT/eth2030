# state/history — Historical state reader with EIP-4444 pruning

[← state](../README.md)

## Overview

Package `history` implements time-travel reads over historical account and storage state. It maintains a configurable retention window (in blocks) and supports pruning entries older than a given block height, aligned with EIP-4444 history expiry semantics.

State snapshots are stored per address (accounts) and per address/slot pair (storage). Point-in-time queries return the most recent entry at or before the requested block number.

## Functionality

### Types

```go
type AccountHistoryEntry struct {
    BlockNumber uint64
    Address     types.Address
    Nonce       uint64
    Balance     []byte
    CodeHash    types.Hash
    StorageRoot types.Hash
    Proof       []byte
}

type StorageHistoryEntry struct {
    BlockNumber uint64
    Address     types.Address
    Slot        types.Hash
    Value       types.Hash
}

type HistoryRange struct {
    MinBlock uint64
    MaxBlock uint64
}
```

### StateHistoryReader

```go
func NewStateHistoryReader(retentionWindow uint64) *StateHistoryReader
func (r *StateHistoryReader) AddAccountEntry(entry AccountHistoryEntry)
func (r *StateHistoryReader) AddStorageEntry(entry StorageHistoryEntry)
func (r *StateHistoryReader) GetAccountAt(addr, blockNum) (*AccountHistoryEntry, error)
func (r *StateHistoryReader) GetStorageAt(addr, slot, blockNum) (*StorageHistoryEntry, error)
func (r *StateHistoryReader) GetAccountHistory(addr) []AccountHistoryEntry
func (r *StateHistoryReader) GetStorageHistory(addr, slot) []StorageHistoryEntry
func (r *StateHistoryReader) PruneHistory(beforeBlock uint64) (int, error)
func (r *StateHistoryReader) Range() HistoryRange
```

### Errors

`ErrBlockNotInRange`, `ErrNoHistoryAvailable`, `ErrHistoryPruned`, `ErrInvalidPruneRange`

## Usage

```go
reader := history.NewStateHistoryReader(128) // keep 128 blocks

reader.AddAccountEntry(history.AccountHistoryEntry{
    BlockNumber: 1000, Address: addr, Nonce: 5,
})

entry, err := reader.GetAccountAt(addr, 1000)
// entry.Nonce == 5

pruned, err := reader.PruneHistory(900) // remove blocks < 900
```
