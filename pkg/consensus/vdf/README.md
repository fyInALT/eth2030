# vdf

VDF beacon randomness integrated into the consensus layer (K+/M+ roadmap).

## Overview

Package `vdf` wires the Wesolowski VDF (`crypto.VDFv2`) into the consensus
layer to provide unbiasable epoch-level randomness. Each epoch has a
three-phase lifecycle:

1. **BeginEpoch**: seed (typically the previous epoch's RANDAO mix) is stored
   and the VDF computation is initiated.
2. **RevealOutput**: each participating validator submits their VDF output,
   which is verified by re-running `VDFv2.Evaluate` against the epoch seed with
   domain separation.
3. **FinalizeEpoch**: once `MinParticipation` (default 50%) of the reveal window
   has submitted, the outputs are sorted by validator ID and Keccak256-combined
   to produce the final unbiasable randomness.

## Functionality

### Types

| Name | Description |
|------|-------------|
| `VDFConsensus` | Thread-safe VDF randomness engine |
| `VDFConsensusConfig` | `VDFDifficulty`, `EpochLength`, `RevealWindow`, `MinParticipation` |
| `EpochRandomness` | Finalized randomness: epoch number, VDF output, seed, reveal count |

### Functions / methods

| Name | Description |
|------|-------------|
| `DefaultVDFConsensusConfig() VDFConsensusConfig` | Difficulty=10, epoch=32, window=8, participation=0.5 |
| `NewVDFConsensus(config) *VDFConsensus` | Create engine |
| `(*VDFConsensus).BeginEpoch(epochNum, seed) error` | Start epoch with given seed |
| `(*VDFConsensus).RevealOutput(epochNum, validatorID, output) error` | Submit and verify a VDF reveal |
| `(*VDFConsensus).FinalizeEpoch(epochNum) (*EpochRandomness, error)` | Combine reveals; error if below threshold |
| `(*VDFConsensus).GetRandomness(epochNum) ([]byte, error)` | Retrieve finalized randomness |
| `(*VDFConsensus).IsEpochFinalized(epochNum) bool` | True if epoch has been finalized |
| `(*VDFConsensus).RevealCount(epochNum) int` | Number of reveals received |
| `(*VDFConsensus).CurrentEpoch() uint64` | Highest started epoch |

### Errors

`ErrVDFEpochNotStarted`, `ErrVDFEpochAlreadyFinalized`, `ErrVDFInsufficientReveals`, `ErrVDFInvalidOutput`

## Usage

```go
import "github.com/eth2030/eth2030/consensus/vdf"

vc := vdf.NewVDFConsensus(vdf.DefaultVDFConsensusConfig())
vc.BeginEpoch(epoch, randaoMix)

// Each validator runs VDF and submits output.
vc.RevealOutput(epoch, validatorID, vdfOutput)

randomness, err := vc.FinalizeEpoch(epoch)
```

[← consensus](../README.md)
