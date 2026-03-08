# trie/migrate — MPT-to-binary-trie migration

Converts an existing Merkle Patricia Trie (MPT) state into a binary Merkle trie (EIP-7864), with batch processing, parallel address-space splitting, crash-recovery checkpointing, and gas accounting.

[← trie](../README.md)

## Overview

`MigrateFromMPT` is the simple one-shot converter: it iterates all key-value pairs from the source MPT, rehashes each key with Keccak256 (matching binary trie key derivation), and inserts them into a new `mpt.BinaryTrie`.

`MPTToBinaryTrieMigrator` provides the production migrator. It collects all MPT pairs once, then exposes `MigrateBatch()` which processes one batch per call, charges gas via a `GasAccountant`, and saves a `MigrationCheckpoint` after each batch for crash recovery. The migrator is composed of:

- `BatchConverter` — inserts up to `batchSize` pairs per call into the destination trie.
- `AddressSpaceSplitter` — divides the 256-bit key space into N equal ranges for parallel migration workers.
- `StateProofGenerator` — caches MPT inclusion proofs generated during migration for cross-verification.
- `MigrationCheckpointer` — stores a stack of `MigrationCheckpoint` values with `Latest()` for resume.
- `GasAccountant` — tracks read/write/proof costs against a gas budget.

`OverlayMigrator` (`overlay.go`) and `MigrationPlanner` (`migration_planner.go`) provide overlay-based incremental migration and planning primitives.

## Functionality

### Types

| Type | Purpose |
|------|---------|
| `MPTToBinaryTrieMigrator` | Full-featured orchestrator (batch, gas, checkpoint, proofs) |
| `BatchConverter` | Converts fixed-size key-value batches into the destination trie |
| `AddressSpaceSplitter` | Splits 256-bit space into N `AddressRange` slices |
| `MigrationCheckpointer` | Stores/retrieves `MigrationCheckpoint` for crash recovery |
| `GasAccountant` | Tracks per-op gas against a budget |
| `StateProofGenerator` | Generates and caches MPT proofs during migration |

### Key Functions

- `MigrateFromMPT(source)` — simple one-shot migration
- `NewMPTToBinaryTrieMigrator(source, batchSize, gasBudget)` / `MigrateBatch()` / `Destination()` / `KeysMigrated()`
- `NewBatchConverter(batchSize)` / `ConvertBatch(pairs, dest)`
- `NewAddressSpaceSplitter(n)` / `Ranges()` / `InRange(key, r)`
- `NewMigrationCheckpointer()` / `Save(cp)` / `Latest()`
- `NewGasAccountant(budget, perRead, perWrite, perProof)` / `ChargeRead()` / `ChargeWrite()` / `ChargeProof()`

## Usage

```go
migrator, _ := migrate.NewMPTToBinaryTrieMigrator(sourceTrie, 1000, 1_000_000)
for {
    count, done, err := migrator.MigrateBatch()
    if err != nil || done { break }
    _ = count
}
binaryTrie := migrator.Destination()
```
