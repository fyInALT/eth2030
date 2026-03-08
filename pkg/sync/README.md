# sync

Blockchain synchronization: full sync, snap sync, beam sync, beacon sync, trie healing, and checkpoint sync.

## Overview

The `sync` package provides the complete block synchronization pipeline for the ETH2030 execution client. It implements three complementary sync modes — full, snap, and beam — and orchestrates their interaction through a single `Syncer` entry point.

In full sync mode the syncer downloads headers in batches, validates the chain (parent hash linkage, block number sequence, timestamp ordering), fetches block bodies, and inserts assembled blocks sequentially. In snap sync mode the syncer first downloads the world state at a pivot block (accounts, storage, bytecodes) via the Snap/1 protocol, runs concurrent trie healing to fill gaps, then switches to full block sync for the remaining blocks. Snap sync falls back to full sync automatically on errors.

Beam sync (`sync/beam`) extends stateless execution: rather than downloading all state upfront, it fetches individual account and storage slots on-demand from peers as the EVM needs them during block execution. The beacon sync subpackage handles synchronization of beacon chain blocks and blob sidecars, with a blob recovery mechanism for partial availability scenarios.

## Table of Contents

- [Functionality](#functionality)
- [Subpackages](#subpackages)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### Sync Modes and Stages

```go
const (
    ModeSnap = "snap"
    ModeFull = "full"
)

const (
    StageNone          uint32 = 0
    StageHeaders       uint32 = 1  // downloading headers
    StageSnapAccounts  uint32 = 2  // snap: downloading accounts
    StageSnapStorage   uint32 = 3  // snap: downloading storage
    StageSnapBytecodes uint32 = 4  // snap: downloading bytecodes
    StageSnapHealing   uint32 = 5  // snap: healing trie
    StageBlocks        uint32 = 6  // downloading and inserting blocks
    StageCaughtUp      uint32 = 7  // sync complete
)
```

`StageName(stage uint32) string` returns a human-readable stage label.

### Syncer

`Syncer` is the central orchestrator:

```go
type Syncer struct { ... }

func NewSyncer(config *Config) *Syncer
func (s *Syncer) SetFetchers(hf HeaderSource, bf BodySource, ins BlockInserter)
func (s *Syncer) SetSnapSync(peer SnapPeer, writer StateWriter)
func (s *Syncer) SetCallbacks(insertHeaders, insertBlocks, currentHeight, hasBlock) // legacy API
func (s *Syncer) RunSync(targetBlock uint64) error
func (s *Syncer) Start(targetHeight uint64) error
func (s *Syncer) Cancel()
func (s *Syncer) State() uint32         // StateIdle / StateSyncing / StateDone
func (s *Syncer) Stage() uint32         // current SyncStage constant
func (s *Syncer) Mode() string
func (s *Syncer) GetProgress() Progress
func (s *Syncer) IsSyncing() bool
func (s *Syncer) MarkDone()
```

`RunSync(targetBlock uint64)` drives the pipeline: snap then full (with automatic fallback), or full-only if snap components are absent.

Default configuration (`DefaultConfig`): snap mode, 192 headers per batch, 16 max pending, 128 bodies per batch.

### Interfaces

```go
type BlockInserter interface {
    InsertChain(blocks []*types.Block) (int, error)
    CurrentBlock() *types.Block
}

type HeaderSource interface {
    FetchHeaders(from uint64, count int) ([]*types.Header, error)
}

type BodySource interface {
    FetchBodies(hashes []types.Hash) ([]*types.Body, error)
}
```

### Progress Tracking

```go
type Progress struct {
    StartingBlock uint64
    CurrentBlock  uint64
    HighestBlock  uint64
    PulledHeaders uint64
    PulledBodies  uint64
    Stage         uint32
    Mode          string
    SnapProgress  *SnapProgress // non-nil during snap sync
}

func (p Progress) Percentage() float64
```

### Header Validation

`ValidateHeaderChain(headers []*types.Header, parent *types.Header) error` checks:
- Sequential block numbers
- Parent hash linkage
- Timestamp not more than 15 seconds in the future (`maxFutureTimestamp`)
- Timestamp at or after parent timestamp

### Utility Functions

`AssembleBlocks(headers []*types.Header, bodies []*types.Body) ([]*types.Block, error)` pairs headers with bodies to construct complete `types.Block` values.

### Snap Sync

The `snap` subpackage implements the Snap/1 protocol client. `SnapSyncer` drives four sequential phases:

1. **Accounts** (`PhaseAccounts`) — range requests for account data in the state trie.
2. **Storage** (`PhaseStorage`) — range requests for storage slots of each account.
3. **Bytecode** (`PhaseBytecode`) — retrieval of contract bytecodes referenced by accounts.
4. **Healing** (`PhaseHealing`) — concurrent trie healing to fill any gaps left by range sync.

`SelectPivot(targetBlock uint64) (uint64, error)` selects a pivot block sufficiently behind the target to avoid reorg risk.

`SnapProgress` tracks per-phase account/slot counts, bytecode counts, and healing request counts.

### Beam Sync

`BeamSync` (`sync/beam`) enables stateless block execution:

```go
type BeamSync struct { ... }
type BeamStateFetcher interface {
    FetchAccount(addr types.Address) (*BeamAccountData, error)
    FetchStorage(addr types.Address, key types.Hash) (types.Hash, error)
}
```

`BeamSync` maintains a local account/storage cache. On cache miss it calls the `BeamStateFetcher` to retrieve data from peers in real time. A `BeamPrefetcher` speculatively loads likely-accessed state based on transaction call patterns.

Statistics: `fetchCount`, `cacheHits`, `cacheMisses` (atomic counters).

### Beacon Sync

The `beacon` subpackage syncs beacon chain blocks and blob sidecars:

```go
type BeaconSyncer struct { ... }
type BeaconBlock struct {
    Slot, ProposerIndex uint64
    ParentRoot, StateRoot [32]byte
    Body []byte
}
type BlobSidecar struct {
    Index       uint64
    BlobData    []byte
    Commitment  [48]byte
    Proof       [48]byte
}
```

`BeaconSyncer` fetches blocks slot-by-slot, validates each block against its parent, and stores blob sidecars. `BlobRecovery` reconstructs missing blobs from partial data using data availability sampling concepts when fewer than the required sidecars are available.

### Checkpoint Sync

The `checkpoint` subpackage provides trusted checkpoint-based sync: downloads a pre-validated state snapshot at a known checkpoint hash, then continues with full sync from that point forward.

### State Sync

The `statesync` subpackage implements concurrent state trie synchronization for full state download, distinct from snap sync's account-range approach.

### Downloader

The `downloader` subpackage provides `HeaderSource` and `BodySource` implementations over the ETH wire protocol, including request batching, peer scoring, and timeout handling.

### Inserter

The `inserter` subpackage wraps the core blockchain's `InsertChain` and provides retry logic and fork-choice integration.

### Range Proof

The `rangeproof` subpackage verifies Merkle range proofs returned by snap peers to ensure account/storage data integrity during snap sync.

### Healer

The `healer` subpackage implements concurrent trie healing: after snap sync completes account and storage ranges, the healer identifies missing trie nodes and fetches them from peers.

## Subpackages

| Subpackage | Description |
|------------|-------------|
| [`beacon/`](./beacon/) | Beacon block and blob sidecar sync; blob recovery |
| [`beam/`](./beam/) | On-demand stateless state fetching for beam sync |
| [`checkpoint/`](./checkpoint/) | Trusted checkpoint-based sync |
| [`checksync/`](./checksync/) | Chain integrity verification during sync |
| [`downloader/`](./downloader/) | ETH wire protocol header and body fetching |
| [`healer/`](./healer/) | Concurrent trie node healing after snap sync |
| [`inserter/`](./inserter/) | Block insertion with retry and fork-choice integration |
| [`rangeproof/`](./rangeproof/) | Merkle range proof verification for snap sync |
| [`snap/`](./snap/) | Snap/1 protocol: account, storage, bytecode, healing phases |
| [`statesync/`](./statesync/) | Full state trie download for non-snap sync |
| [`support/`](./support/) | Shared sync utilities and peer abstraction |

## Usage

```go
// Create a syncer in snap mode.
cfg := sync.DefaultConfig()
cfg.Mode = sync.ModeSnap
s := sync.NewSyncer(cfg)

// Wire fetchers and the block inserter.
s.SetFetchers(headerFetcher, bodyFetcher, blockInserter)

// Optionally configure snap sync.
s.SetSnapSync(snapPeer, stateWriter)

// Run synchronization to block 1,000,000.
if err := s.RunSync(1_000_000); err != nil {
    log.Error("sync failed", "err", err)
}

// Monitor progress.
prog := s.GetProgress()
log.Info("sync", "stage", sync.StageName(prog.Stage),
    "current", prog.CurrentBlock,
    "highest", prog.HighestBlock,
    "pct", prog.Percentage())

// Cancel if needed.
s.Cancel()
```

## Documentation References

- [Roadmap](../../docs/ROADMAP.md)
- [Snap/1 Protocol](https://github.com/ethereum/devp2p/blob/master/caps/snap.md)
- [ETH Wire Protocol](https://github.com/ethereum/devp2p/blob/master/caps/eth.md)
- I+ roadmap: beam sync, stateless execution
- J+ roadmap: light client, beam sync improvements
