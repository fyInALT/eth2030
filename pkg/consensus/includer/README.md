# includer

1 ETH includer selection and duty management (L+ roadmap).

## Overview

Package `includer` implements the 1 ETH includer mechanism that democratises
transaction inclusion. Any participant who stakes exactly 1 ETH (`OneETH =
1e18 wei`) registers as an includer. Each slot, a single active includer is
selected pseudorandomly using `H(slot || randomSeed) mod len(active)`. The
selected includer builds a signed `IncluderDuty` listing the transactions
that the block proposer must include (FOCIL-adjacent).

`IncluderPool` is the thread-safe registry. Slashing reduces stake by
`SlashPenaltyPercent` (10%) and marks the includer as inactive. The
deterministic ordering is maintained by sorting active addresses
lexicographically.

## Functionality

### Types

| Name | Description |
|------|-------------|
| `IncluderPool` | Thread-safe registry of registered 1 ETH includers |
| `IncluderRecord` | Per-includer state: address, stake, status, slash reason |
| `IncluderDuty` | Signed per-slot duty: includer address, tx list hash, deadline |
| `IncluderStatus` | `IncluderActive`, `IncluderSlashed`, `IncluderExited` |

### Constants

| Name | Value | Description |
|------|-------|-------------|
| `BaseIncluderReward` | 10,000 Gwei | Base reward per correctly fulfilled duty |
| `IncluderRewardDecay` | 100 Gwei | Per-slot decay applied to reward |
| `SlashPenaltyPercent` | 10 | Percent of stake slashed for misbehaviour |

### Functions / methods

| Name | Description |
|------|-------------|
| `NewIncluderPool() *IncluderPool` | Create an empty pool |
| `(*IncluderPool).RegisterIncluder(addr, stake) error` | Register with exactly 1 ETH stake |
| `(*IncluderPool).SelectIncluder(slot, randomSeed) (Address, error)` | Deterministic per-slot selection |
| `(*IncluderPool).SlashIncluder(addr, reason) error` | Slash and deactivate an includer |
| `(*IncluderDuty).Hash() types.Hash` | Canonical duty hash for signing |
| `VerifyIncluderSignature(duty, sig) bool` | ECDSA signature verification |
| `IncluderReward(slot) uint64` | Compute slot reward with decay |
| `ValidateIncluderRegistration(addr, stake, pool) error` | Pre-registration checks |
| `ValidateIncluderDuty(duty) error` | Duty well-formedness checks |

## Usage

```go
import "github.com/eth2030/eth2030/consensus/includer"

pool := includer.NewIncluderPool()
pool.RegisterIncluder(addr, includer.OneETH)

selected, err := pool.SelectIncluder(slot, randaoMix)
duty := &includer.IncluderDuty{Slot: slot, Includer: selected, ...}
```

[← consensus](../README.md)
