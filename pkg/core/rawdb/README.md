# rawdb

Low-level persistent storage: key-value interfaces, FileDB with WAL, chain accessors, history expiry (EIP-4444), and ancient/freezer storage.

[← core](../README.md)

## Overview

Package `rawdb` provides the storage foundation for the ETH2030 execution
layer. It defines key-value store interfaces, an append-only `FileDB` backed by
a write-ahead log (WAL) for crash safety, a `ChainDB` with LRU caches for
blocks/receipts/headers, and EIP-4444 history pruning. The schema follows
go-ethereum's single-byte key-prefix convention to prevent collisions.

## Functionality

### Interfaces (`database.go`, `key_value_store.go`)

| Interface | Description |
|---|---|
| `KeyValueReader` | `Has(key)`, `Get(key)` |
| `KeyValueWriter` | `Put(key, value)`, `Delete(key)` |
| `KeyValueStore` | `KeyValueReader + KeyValueWriter + Close()` |
| `KeyValueIterator` | `KeyValueStore + NewIterator(prefix)` |
| `Batch` | Write-only batch: `Put`, `Delete`, `ValueSize`, `Write`, `Reset` |
| `Batcher` | `NewBatch() Batch` |
| `Database` | `KeyValueStore + Batcher` (primary interface) |
| `Iterator` | `Next`, `Key`, `Value`, `Release` |

`ErrNotFound` — returned when a key does not exist.

### FileDB (`filedb.go`)

Persistent file-based key-value store with WAL for crash safety.

Layout:
```
<dir>/
  LOCK       — flock exclusive lock (prevents concurrent processes)
  wal        — binary append-only write-ahead log
  data/      — per-key files (filename = hex(key))
```

- `NewFileDB(dir string) (*FileDB, error)` — opens or creates a database.
- Implements the full `Database` interface including `NewBatch()`.
- WAL record types: `walPut` (0x01), `walDelete` (0x02), `walCommit` (0x03).
- In-memory index rebuilt from disk on open; all reads are served from memory.

### MemoryDB (`memorydb.go`)

In-memory key-value store for tests; implements `KeyValueIterator`.
`NewMemoryDB() *MemoryDB`.

### ChainDB (`chaindb.go`)

High-level chain store with LRU caches wrapping raw database accessors.

Cache sizes: `blockCacheSize` (256), `headerCacheSize` (1024),
`receiptCacheSize` (256), `tdCacheSize` (1024).

### Schema (`schema.go`)

Key prefixes (single-byte, big-endian block number encoding):

| Prefix | Key pattern | Stores |
|---|---|---|
| `h` | h + num + hash | Header RLP |
| `H` | H + hash | Block number |
| `b` | b + num + hash | Body RLP |
| `r` | r + num + hash | Receipts RLP |
| `l` | l + txHash | Tx lookup (block num) |
| `c` | c + num | Canonical hash |
| `hh` | literal | Head header hash |
| `hb` | literal | Head block hash |
| `C` | C + codeHash | Contract bytecode |
| `d` | d + num + hash | Total difficulty RLP |

### Accessors (`accessors.go`)

Block/header/receipt/transaction read and write helpers keyed by the schema
above. `WriteBlock`, `ReadBlock`, `WriteReceipts`, `ReadReceipts`,
`WriteTxLookup`, `ReadTxLookup`, `WriteHeadBlock`, `ReadHeadBlock`, etc.

### History Expiry — EIP-4444 (`history.go`)

- `DefaultHistoryRetention` = 3,153,600 blocks (~1 year).
- `BALRetentionSlots` = 113,056 blocks (~15.6 days, EIP-7928 §retention-policy).
- `IsBALRetained(headBlock, blockNum)` — true if BAL is within retention window.
- `WriteHistoryOldest(db, blockNum)`, `ReadHistoryOldest(db)` — persist the
  oldest available block for body/receipt queries.
- `ErrHistoryPruned` — returned for pruned historical data.

### Ancient / Freezer Store (`ancient_store.go`, `freezer.go`, `freezer_table.go`)

Immutable "frozen" segments of the chain moved from the live database to an
efficient read-only store. `AncientStore` implements `Database` and exposes
`HasAncient`, `ReadAncient`, `WriteAncients`, `TruncateAncients`.

### Batch (`batch.go`)

`MemoryBatch` — accumulates writes in memory and flushes atomically via `Write()`.

### Table (`table.go`)

`Table` — namespaced view of a `KeyValueStore` with an automatic key prefix.
Useful for isolating subsystem data within a shared database.

### Chain Iterator (`chain_iterator.go`)

`ChainIterator` — iterates over canonical chain blocks in order for sync and
history pruning.

## Usage

```go
db, err := rawdb.NewFileDB("/var/eth2030/chaindata")
// or for tests:
db = rawdb.NewMemoryDB()

// Write a block.
rawdb.WriteBlock(db, block)
rawdb.WriteReceipts(db, block.Hash(), block.NumberU64(), receipts)
rawdb.WriteHeadBlock(db, block)

// Read back.
blk := rawdb.ReadBlock(db, hash, number)
receipts := rawdb.ReadReceipts(db, hash, number)

// EIP-4444: check if BAL is within retention window.
if rawdb.IsBALRetained(headNumber, queryNumber) {
    // serve BAL data
}
```

## Subpackages

| Subpackage | Description |
|---|---|
| [`cache/`](./cache/) | LRU hash cache mapping block numbers to block hashes |
