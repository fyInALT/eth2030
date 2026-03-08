# txpool/journal — Persistent transaction journal for crash recovery

## Overview

Package `journal` provides a write-ahead log for pending transactions, enabling the pool to survive process restarts. Transactions are RLP-encoded and appended as newline-delimited JSON records. The journal supports rotation (compaction to only active transactions) and crash-safe recovery via atomic file rename. A second implementation (`TxJrnl`) provides an alternative compact binary journal for the same purpose.

## Functionality

**Types**
- `TxJournal` — append-only journal; `Insert(tx, local)`, `InsertBatch`, `Rotate(pending)`, `Close`, `Count`, `Exists`
- `JournalEntry{TxRLP, Sender, Timestamp, Local, Hash}` — on-disk record format

**Functions**
- `NewTxJournal(path)` — create or open journal file in append mode
- `Load(path)` — read and decode all entries, skipping corrupt lines; returns `([]*types.Transaction, []JournalEntry, error)`

**Errors**
- `ErrJournalClosed` — write after close
- `ErrJournalCorrupt` — unparseable file
- `ErrJournalNotFound` — file missing on `Load`

## Usage

```go
j, err := journal.NewTxJournal("/var/eth2030/pool/txpool.journal")
j.Insert(tx, true /* local */)

// on startup:
txs, _, err := journal.Load("/var/eth2030/pool/txpool.journal")
// re-add txs to pool

// periodic compaction:
j.Rotate(pool.Pending())
```

[← txpool](../README.md)
