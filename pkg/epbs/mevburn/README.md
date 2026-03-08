# epbs/mevburn — MEV-burn computation and smoothing for ePBS

## Overview

Package `mevburn` implements the MEV-burn mechanism described in EIP-7732: a configurable fraction of each winning builder bid is burned rather than paid to the proposer, reducing MEV-driven proposer income variance and improving protocol fairness. The core computation is `ComputeMEVBurn`, which applies a fixed burn fraction to a bid value and returns the split between burn and proposer payment.

`MEVBurnTracker` accumulates per-epoch burn statistics and maintains an exponential moving average (EMA) over recent bids so that callers can estimate a smoothed burn rate for fee market analysis and policy tuning. A companion validator function guards against misconfigured parameters before they reach the burn path.

## Functionality

**Configuration and results**
- `MEVBurnConfig{BurnFraction=0.50, SmoothingFactor=0.10, MinBurnThreshold=100 Gwei, Tolerance=0.01}`
- `MEVBurnResult{BidValue, BurnAmount, ProposerPayment uint64}`
- `ValidateMEVBurnConfig(cfg MEVBurnConfig) error` — checks fractions in (0,1), threshold > 0

**Core burn computation**
- `ComputeMEVBurn(bidValue uint64, cfg MEVBurnConfig) (MEVBurnResult, error)`
  - `burnAmount = floor(bidValue * burnFraction)`
  - `proposerPayment = bidValue - burnAmount`
  - Returns error if `burnAmount < MinBurnThreshold`

**Smoothing**
- `EstimateSmoothedBurn(recentBids []uint64, cfg MEVBurnConfig) (uint64, error)` — EMA over the provided bid values using `SmoothingFactor`; returns estimated burn for the next bid

**Burn validation**
- `ValidateBurnAmount(actual, expected uint64, tolerance float64) error` — relative tolerance check: `|actual-expected|/expected <= tolerance`

**MEVBurnTracker** (per-epoch statistics)
- `EpochBurnStats{TotalBurned, TotalBidValue uint64, BidCount int}`
- `NewMEVBurnTracker(cfg MEVBurnConfig) *MEVBurnTracker`
- `RecordBid(epoch, slot, bidValue uint64) (MEVBurnResult, error)` — burns and records
- `EpochStats(epoch uint64) (EpochBurnStats, bool)`
- `CurrentEMA() uint64` — smoothed burn amount across all recorded bids
- `TotalBurned() uint64`

Parent package: [`epbs`](../README.md)
