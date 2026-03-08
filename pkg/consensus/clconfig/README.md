# clconfig

Consensus-layer configuration constants and fork parameters.

## Overview

Package `clconfig` holds the `ConsensusConfig` struct that drives slot timing,
epoch sizing, and finality parameters across all consensus-layer components.
It provides two ready-made configs — mainnet defaults and the K+ quick-slots
regime — plus a `Validate` helper that enforces invariants before the node
boots.

The package is intentionally kept minimal so that other packages can depend on
it without pulling in the broader consensus package.

## Functionality

### Types

| Name | Description |
|------|-------------|
| `ConsensusConfig` | Holds all CL tunable parameters (slot duration, epoch size, finality epochs, lean-chain mode) |

### Functions

| Name | Description |
|------|-------------|
| `DefaultConfig() *ConsensusConfig` | Standard Ethereum mainnet: 12 s slots, 32 slots/epoch, 2-epoch Casper FFG finality |
| `QuickSlotsConfig() *ConsensusConfig` | K+ regime: 6 s slots, 4 slots/epoch, 1-epoch finality |
| `(*ConsensusConfig).Validate() error` | Returns an error if any field is out of range |
| `(*ConsensusConfig).EpochDuration() uint64` | Returns `SecondsPerSlot * SlotsPerEpoch` |
| `(*ConsensusConfig).IsSingleEpochFinality() bool` | True when `EpochsForFinality == 1` |

### Key fields

| Field | Default | Notes |
|-------|---------|-------|
| `SecondsPerSlot` | 12 | 6 for quick-slots (K+) |
| `SlotsPerEpoch` | 32 | 4 for quick-slots (K+) |
| `EpochsForFinality` | 2 | 1 enables single-epoch finality |
| `LeanAvailableChainMode` | false | Enables PQ subset attestation |
| `LeanAvailableChainValidators` | 512 | Valid range: [256, 1024] |

## Usage

```go
import "github.com/eth2030/eth2030/consensus/clconfig"

cfg := clconfig.DefaultConfig()
if err := cfg.Validate(); err != nil {
    log.Fatal(err)
}

// K+ quick-slots upgrade
kpCfg := clconfig.QuickSlotsConfig()
fmt.Println(kpCfg.EpochDuration()) // 24 seconds
```

[← consensus](../README.md)
