# epbs/builder — Builder registry, reputation, withdrawal, and marketplace

## Overview

Package `builder` implements the full builder management stack for ePBS (EIP-7732). It provides a `BuilderRegistry` for tracking registrations and per-slot bid history, a `BuilderReputationTracker` for sliding-window reliability scoring with slashing-event detection, a `BuilderWithdrawalRegistry` enforcing the 64-epoch withdrawal delay and 16 384-builder sweep limit defined by EIP-7732, and a `BuilderMarket` that runs Vickrey (second-price sealed-bid) auctions and maintains long-running builder scores based on delivery history.

All four components are independently constructible and thread-safe.

## Functionality

**BuilderRegistry** (`builder_registry.go`)
- `BuilderInfo{Address, Pubkey [48]byte, FeeRecipient Address, GasLimit, RegisteredAt time.Time, Active bool, Stake uint64}`
- `BuilderBidRecord{Slot, Value, GasLimit uint64, Timestamp time.Time, Won bool}`
- `BuilderStats{TotalBids, WonBids uint64, TotalValue uint64, WinRate, AvgBidValue float64}`
- `NewBuilderRegistry() *BuilderRegistry`
- `RegisterBuilder(info BuilderInfo) error` / `DeregisterBuilder(addr Address) error`
- `GetBuilder(addr Address) (*BuilderInfo, bool)`
- `ActiveBuilders() []BuilderInfo`
- `RecordBid(addr Address, record BuilderBidRecord)`
- `GetBuilderStats(addr Address) (BuilderStats, error)`
- `TopBuilders(n int) []BuilderInfo` — ranked by win rate
- `PruneInactive(before time.Time) int`

**BuilderReputationTracker** (`builder_reputation.go`)
- `ReputationConfig{WindowSize=100, MinDeliveries=5, SlashingPenalty=0.15, DecayFactor=0.98, DefaultScore=0.5}`
- `ReputationRecord{Deliveries, Failures, TotalLatencyMs int64, SlashingCount int, CurrentScore float64}`
- `NewBuilderReputationTracker(cfg ReputationConfig) *BuilderReputationTracker`
- `RecordDelivery(addr Address, latencyMs int64)` / `RecordFailure(addr Address)`
- `GetScore(addr Address) float64`
- `DecayScores()` — exponential decay toward `DefaultScore`
- `SlashingDetector` — detects `SlashEventLateDelivery`, `SlashEventInvalidPayload`, `SlashEventEquivocation`
  - `CheckLateDelivery`, `CheckEquivocation`, `CheckInvalidPayload`

**BuilderWithdrawalRegistry** (`builder_withdrawal.go`)
- EIP-7732 constants: `MinBuilderWithdrawabilityDelay = 64` epochs, `MaxBuildersPerWithdrawalsSweep = 16384`, `BuilderWithdrawalPrefix = 0x03`
- `NewBuilderWithdrawalRegistry() *BuilderWithdrawalRegistry`
- `AddBuilder(addr Address, stake uint64)`
- `RequestWithdrawal(addr Address, currentEpoch uint64) error` — sets `withdrawableAt = currentEpoch + 64`
- `WithdrawableBuilders(currentEpoch uint64) []Address` — up to `MaxBuildersPerWithdrawalsSweep`

**BuilderMarket** (`builder_market.go`)
- `BuilderMarketConfig{ReservePrice, MaxBidsPerSlot=256, MaxConsecutiveMisses=3, ScoreDecayFactor=0.95, DeliveryBonus=10.0, MissPenalty=25.0}`
- `MarketBid{Bid epbs.BuilderBid, BuilderAddr Address, ReceivedAt time.Time}`
- `BuilderProfile{TotalBids, TotalWins, TotalDeliveries, TotalMisses, ConsecutiveMisses uint64, Score float64, Banned bool}`
- `NewBuilderMarket(cfg BuilderMarketConfig) *BuilderMarket`
- `RegisterBuilder(addr Address) *BuilderProfile`
- `ValidateBid(bid *MarketBid) error` / `SubmitBid(bid *MarketBid) error`
- `SelectWinner(slot uint64) (*MarketBid, uint64, error)` — returns winner and Vickrey clearing price
- `RecordDelivery(addr Address) error` / `RecordMiss(addr Address) error` — updates score; bans on streak
- `UnbanBuilder(addr Address) error`
- `PruneBefore(slot uint64) int`

Parent package: [`epbs`](../README.md)
