# minimmit

Minimmit one-round BFT consensus engine for single-round slot finality.

## Overview

Package `minimmit` implements the Minimmit one-round BFT finality engine.
Finality is achieved in a single message round: once accumulated stake for a
proposed block root exceeds 2/3 of `TotalStake`, the slot is finalized
immediately, eliminating the multi-round exchange of traditional PBFT/HotStuff.

The engine follows a strict state machine: `Idle → Proposed → Voting →
Finalized` (or `Failed` on a missed slot). Equivocation — the same validator
voting for two different roots in the same slot — is detected and rejected with
`ErrMinimmitEquivocation`.

## Functionality

### Types

| Name | Description |
|------|-------------|
| `MinimmitEngine` | Thread-safe one-round BFT engine |
| `MinimmitConfig` | `TotalStake`, threshold numerator/denominator, voter limit, missed-slot penalty |
| `MinimmitVote` | Validator vote: index, slot, block root, 96-byte BLS signature, stake |
| `MinimmitState` | `Idle`, `Proposed`, `Voting`, `Finalized`, `Failed` |
| `FinalityMode` | `Classic` (Casper FFG), `SSF`, or `Minimmit` |

### Functions / methods

| Name | Description |
|------|-------------|
| `DefaultMinimmitConfig() *MinimmitConfig` | 32M ETH stake, 2/3 threshold, 8192 voter limit, 1 ETH penalty |
| `NewMinimmitEngine(config) (*MinimmitEngine, error)` | Create engine |
| `(*MinimmitEngine).ProposeBlock(slot, blockRoot) error` | Transition `Idle → Voting` |
| `(*MinimmitEngine).CastVote(vote) error` | Record a vote; auto-finalizes on threshold |
| `(*MinimmitEngine).CheckFinality() bool` | True if current round is finalized |
| `(*MinimmitEngine).MissSlot(slot)` | Mark slot missed, transition to `Failed` |
| `(*MinimmitEngine).Reset()` | Return to `Idle` for next round |
| `(*MinimmitEngine).FinalizedHead() (slot uint64, root Hash)` | Most recently finalized slot+root |
| `(*MinimmitEngine).StakeForRoot(root) uint64` | Accumulated stake for a block root |

### Errors

`ErrMinimmitNilConfig`, `ErrMinimmitZeroStake`, `ErrMinimmitDuplicateVote`, `ErrMinimmitWrongSlot`, `ErrMinimmitAlreadyFinal`, `ErrMinimmitInvalidState`, `ErrMinimmitEquivocation`

## Usage

```go
import "github.com/eth2030/eth2030/consensus/minimmit"

engine, _ := minimmit.NewMinimmitEngine(minimmit.DefaultMinimmitConfig())
engine.ProposeBlock(slot, blockRoot)

for _, vote := range incomingVotes {
    engine.CastVote(vote)
    if engine.CheckFinality() {
        break
    }
}
finalizedSlot, finalizedRoot := engine.FinalizedHead()
```

[← consensus](../README.md)
