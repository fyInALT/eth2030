# sync/snap — Snap sync world-state downloader

Downloads the full Ethereum world state at a recent pivot block without replaying historical transactions, implementing the four-phase snap sync protocol.

[← sync](../README.md)

## Overview

Snap sync skips transaction replay by fetching state trie leaves directly from peers. `SnapSyncer` orchestrates four sequential phases:

1. **Accounts** — downloads account trie leaves in parallel key-space ranges (`AccountRangeRequest`), queuing accounts with non-empty storage and non-empty code for the following phases.
2. **Storage** — fetches storage trie leaves for each contract account that has non-trivial storage.
3. **Bytecode** — downloads contract bytecodes by code hash, verifying each via Keccak256.
4. **Healing** — queries `StateWriter.MissingTrieNodes` to identify interior trie nodes that were not covered by the leaf-level downloads, then fetches them in batches.

The pivot block is selected `PivotOffset` (64) blocks behind the chain head via `SelectPivot`. All data is persisted through the `StateWriter` interface. The sync is cancellable at any phase.

## Functionality

### Types

| Type | Purpose |
|------|---------|
| `SnapSyncer` | Four-phase state downloader |
| `SnapPeer` | Interface: account/storage/bytecode/trie-node requests |
| `StateWriter` | Interface: write accounts, storage, bytecodes, trie nodes |
| `SnapProgress` | Per-phase counters, bytes downloaded, ETA |
| `AccountData` / `StorageData` / `BytecodeData` | Downloaded state items |
| `AccountRangeRequest` / `StorageRangeRequest` / `BytecodeRequest` | Typed wire requests |

### Key Functions

- `NewSnapSyncer(writer)` / `SetPivot(header)` / `Start(peer)` / `Cancel()`
- `SelectPivot(headNumber)` / `Phase()` / `Progress()`
- `VerifyAccountRange(root, accounts, proof)` / `SplitAccountRange(origin, limit, n)`
- `MergeAccountRanges(a, b)` / `DetectHealingNeeded(writer, root)`

## Usage

```go
syncer := snap.NewSnapSyncer(myStateWriter)
pivot, _ := snap.SelectPivot(headNumber)
syncer.SetPivot(pivotHeader)
if err := syncer.Start(mySnapPeer); err != nil {
    log.Fatal(err)
}
p := syncer.Progress()
fmt.Printf("accounts: %d, storage: %d, bytecodes: %d\n",
    p.AccountsDone, p.StorageTotal, p.BytecodesTotal)
```
