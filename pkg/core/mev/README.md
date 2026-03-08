# mev

MEV protection: bundle validation, sandwich/frontrun detection, and fair ordering.

[← core](../README.md)

## Overview

Package `mev` provides MEV (Maximal Extractable Value) protection primitives
for the ETH2030 block builder and transaction pool. It implements Flashbots-style
bundle validation, heuristic sandwich and frontrun detection, and a fair-ordering
enforcer that sorts transactions by arrival time and flags position violations.
This implements the commit-reveal MEV ordering protection from the L1 Strawmap
encrypted mempool track.

## Functionality

### FlashbotsBundle

Represents a bundle of transactions for atomic block inclusion.

```go
type FlashbotsBundle struct {
    Transactions      []*types.Transaction
    BlockNumber       uint64
    MinTimestamp      uint64
    MaxTimestamp      uint64
    RevertingTxHashes []types.Hash
}
```

- `Validate()` — checks non-empty, size ≤ `MaxBundleSize` (32), and timestamp
  consistency.
- `IsValidAtTime(timestamp)` — checks `MinTimestamp`/`MaxTimestamp` bounds.
- `TotalGas()` — sum of gas across all bundle transactions.
- `IsRevertAllowed(txHash)` — checks revert-protection allow-list.

### MEV Detection

- `DetectSandwich(txs []*types.Transaction) []SandwichCandidate` — finds
  `{FrontTx, VictimTx, BackTx, Attacker}` triplets where the same sender
  brackets a victim targeting the same contract with a higher gas price.
- `DetectFrontrun(txs, maxRatio) []FrontrunCandidate` — finds adjacent-pair
  `{Frontrunner, Victim, GasRatio}` entries where the frontrunner's gas price
  exceeds the victim's by more than `maxRatio`.
- `FairOrdering(entries, maxDelay) ([]FairOrderingEntry, []error)` — sorts
  transactions by arrival time and flags any transaction displaced more than
  `maxDelay` positions from its natural order.

### Configuration

```go
type MEVProtectionConfig struct {
    EnableSandwichDetection bool
    EnableFrontrunDetection bool
    EnableFairOrdering      bool
    MaxGasPriceRatio        uint64 // default: 10
    SandwichProfitThreshold *big.Int
    FairOrderMaxDelay       int    // default: 5
}
```

`DefaultMEVProtectionConfig()` — returns production defaults.

### Errors

`ErrEmptyBundle`, `ErrBundleTooLarge`, `ErrSandwichDetected`,
`ErrFrontrunDetected`, `ErrInvalidFairOrder`.

### Supporting Types

- `SandwichCandidate` — `{FrontTx, VictimTx, BackTx, Attacker}`.
- `FrontrunCandidate` — `{Frontrunner, Victim, GasRatio}`.
- `BackrunOpportunity` — `{TriggerTx, BackrunTx, TargetAddress}`.
- `FairOrderingEntry` — `{Transaction, ArrivalTime}`.

## Usage

```go
cfg := mev.DefaultMEVProtectionConfig()

// Bundle validation.
bundle := &mev.FlashbotsBundle{
    Transactions: txs,
    BlockNumber:  targetBlock,
}
if err := bundle.Validate(); err != nil {
    return err
}

// Sandwich detection.
sandwiches := mev.DetectSandwich(pendingTxs)
for _, s := range sandwiches {
    log.Printf("sandwich: attacker=%s", s.Attacker)
}

// Fair ordering.
entries := makeEntries(pendingTxs, arrivalTimes)
ordered, violations := mev.FairOrdering(entries, cfg.FairOrderMaxDelay)
```
