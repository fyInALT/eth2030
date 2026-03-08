# cltypes

Shared consensus-layer primitive types used across all consensus subpackages.

## Overview

Package `cltypes` defines the fundamental typed integers and structs — `Slot`,
`Epoch`, `ValidatorIndex`, `Checkpoint`, `JustificationBits`, and
`BeaconState` — that are shared by all CL packages. Keeping them in a
dedicated leaf package prevents import cycles between the larger consensus
subpackages.

`JustificationBits` provides a compact 8-bit bitfield where bit 0 represents
the current epoch and higher bits represent progressively older epochs, exactly
matching the beacon chain spec layout.

## Functionality

### Types

| Name | Description |
|------|-------------|
| `Epoch` | Consensus-layer epoch number (`uint64`) |
| `Slot` | Consensus-layer slot number (`uint64`) |
| `ValidatorIndex` | Beacon-chain validator index (`uint64`) |
| `Checkpoint` | Finality checkpoint: `{Epoch, Root types.Hash}` |
| `JustificationBits` | 8-bit bitfield tracking justified epochs (bit 0 = current) |
| `BeaconState` | Minimal beacon state: slot, epoch, finalized/justified checkpoints, and justification bits |

### Methods on `JustificationBits`

| Method | Description |
|--------|-------------|
| `IsJustified(offset uint) bool` | True when the epoch at offset is justified |
| `Set(offset uint)` | Marks the epoch at offset as justified |
| `Shift(n uint)` | Ages the bitfield by n positions (bit 0 cleared) |

### Functions

| Name | Description |
|------|-------------|
| `SlotToEpoch(slot Slot, slotsPerEpoch uint64) Epoch` | Returns the epoch containing the given slot |
| `EpochStartSlot(epoch Epoch, slotsPerEpoch uint64) Slot` | First slot of an epoch |

## Usage

```go
import "github.com/eth2030/eth2030/consensus/cltypes"

epoch := cltypes.SlotToEpoch(128, 32) // epoch 4

var bits cltypes.JustificationBits
bits.Set(0) // current epoch justified
bits.Set(1) // previous epoch justified
bits.Shift(1) // advance: previous becomes two-epochs-ago
```

[← consensus](../README.md)
