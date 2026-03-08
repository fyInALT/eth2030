# chain

Canonical chain manager, fork choice, header chain, and state cache.

[← core](../README.md)

## Overview

Package `chain` manages the EL canonical chain of blocks. It stores blocks and
receipts in a raw database, maintains a canonical number-to-hash mapping,
handles chain reorganizations, and exposes a `ForkChoice` adapter that receives
`engine_forkchoiceUpdated` calls from the consensus layer via the Engine API.

The package avoids circular imports by depending on `block`, `execution`, and
`rawdb` subpackages while providing a `BlockchainReader` interface back to
callers in `block/`.

## Functionality

### Types

- `Blockchain` — full chain manager. Constructed with
  `NewBlockchain(config, genesis, statedb, db)`.
  Key methods: `InsertBlock(block, receipts)`, `GetBlock(hash)`,
  `GetBlockByNumber(num)`, `GetReceipts(hash)`, `StateAtBlock(block)`,
  `CurrentBlock()`, `Config()`, `Genesis()`.
- `ForkChoice` — tracks head/safe/finalized pointers. Constructed with
  `NewForkChoice(bc)`. Key methods: `Head()`, `Safe()`, `Finalized()`,
  `ForkchoiceUpdate(headHash, safeHash, finalizedHash)`.
- `HeaderChain` — header-only chain management for light clients and sync.
- `StateCache` — LRU state cache keyed by block hash to avoid re-executing
  from genesis on every Engine API call.
- `TxLookupEntry` — `{BlockHash, BlockNumber, TxIndex}` for transaction
  location indexing.

### Errors

| Error | Meaning |
|---|---|
| `ErrNoGenesis` | genesis block not provided |
| `ErrGenesisExists` | genesis already initialized |
| `ErrBlockNotFound` | requested block not in chain |
| `ErrInvalidChain` | blocks are not contiguous |
| `ErrStateNotFound` | state unavailable for block |
| `ErrFinalizedBlockUnknown` | finalized hash refers to unknown block |
| `ErrSafeBlockUnknown` | safe hash refers to unknown block |
| `ErrReorgPastFinalized` | reorg would revert past finalized block |
| `ErrInvalidFinalizedChain` | finalized not in head's ancestry |
| `ErrInvalidSafeChain` | safe not in head's ancestry |
| `ErrSafeNotFinalized` | safe block below finalized number |

### Chain Readers

`reader.go` and `reader_ext.go` provide read-only accessors (receipts, tx
lookups, header by number, block by hash) used by the RPC and Engine API
layers without requiring a write lock.

## Usage

```go
db, _ := rawdb.NewFileDB("/data/chain")
genesis := config.DefaultGenesisBlock()
genesisBlock := genesis.ToBlock()
bc, err := chain.NewBlockchain(chainConfig, genesisBlock, genesisState, db)

fc := chain.NewForkChoice(bc)
err = fc.ForkchoiceUpdate(headHash, safeHash, finalizedHash)

head := fc.Head()
finalized := fc.Finalized()
```
