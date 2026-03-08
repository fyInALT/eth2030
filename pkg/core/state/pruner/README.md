# state/pruner — Bloom-filter state pruner

[← state](../README.md)

## Overview

Package `pruner` removes stale state data from the key-value store using a two-phase mark-and-sweep algorithm. During the mark phase, all keys reachable from the target state root are added to a bloom filter. During the sweep phase, the database is iterated and any key whose prefix belongs to a prunable namespace but is absent from the bloom filter is deleted in batched writes.

The bloom filter uses three FNV-1a hash probes over a configurable bit array (default 256 MiB, ~2 billion bits) to keep false positive rates low.

## Functionality

```go
type PrunerConfig struct {
    BloomSize uint64  // default: DefaultBloomSize (256 MiB)
    Datadir   string
}

const DefaultBloomSize = 256 * 1024 * 1024

func NewPruner(config PrunerConfig, db prunerDB) *Pruner

// Prune performs mark+sweep from the given root. Returns (deletedCount, error).
func (p *Pruner) Prune(root types.Hash) (int, error)

// PruneByKeys deletes all keys under prefixes that are not in the keep set.
func (p *Pruner) PruneByKeys(keep map[string]struct{}, prefixes [][]byte) (int, error)
```

Prunable key prefixes: `"sa"` (snapshot accounts), `"ss"` (snapshot storage), `"t"` (trie nodes).

The `prunerDB` interface requires `rawdb.KeyValueStore` plus `NewIterator(prefix)` and `NewBatch()`.

## Usage

```go
cfg := pruner.PrunerConfig{BloomSize: pruner.DefaultBloomSize}
p := pruner.NewPruner(cfg, db)
deleted, err := p.Prune(stateRoot)
fmt.Printf("pruned %d entries\n", deleted)
```
