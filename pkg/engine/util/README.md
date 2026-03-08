# engine/util — Payload shrinking, builder coordination, and MEV-burn utilities

## Overview

Package `util` collects standalone utilities used across the Engine API layer. It covers three distinct concerns: compressing and field-pruning oversized execution payloads before transmission (`PayloadShrinker`), coordinating the ePBS builder bid lifecycle with second-price auction settlement and reputation tracking (`BuilderCoordinator`), and computing the MEV-burn amount that a proposer must contribute under protocol rules (`MEVBurnCalculator`).

Each component is self-contained and stateless or lightly stateful, making them straightforward to unit-test and reuse in higher-level engine handlers without circular imports.

## Functionality

**PayloadShrinker** (`payload_shrinking.go`)
- `ShrinkConfig{MaxPayloadSize int64, CompressionLevel int, PrunableFields []string}`
- `ShrinkStats{OriginalSize, ShrunkSize int64, CompressionRatio float64, FieldsPruned int, Strategy string}`
- `NewPayloadShrinker(cfg ShrinkConfig) *PayloadShrinker`
- `ShrinkPayload(payload []byte) ([]byte, ShrinkStats, error)` — applies best-fit strategy
- `EstimateShrinkage(payload []byte) ShrinkStats` — dry-run without modifying payload
- `ApplyStrategy(payload []byte, strategy string) ([]byte, error)` — strategies: `compress` (DEFLATE), `prune_zeros`, `dedup` (consecutive 8-byte chunks), `combined`

**BuilderCoordinator** (`builder_coordinator.go`)
- `CoordinatorConfig{MaxBuilders int, BidTimeout time.Duration, MinBidIncrement uint64, ReputationDecay float64}`
- `AuctionSettlement{WinnerID string, WinnerBid uint64, SettlePrice uint64, RunnerUpID string}` — `SettlePrice` is the second-highest bid (Vickrey)
- `NewBuilderCoordinator(cfg CoordinatorConfig) *BuilderCoordinator`
- `RegisterBuilder(id string) error`
- `SubmitBid(builderID string, slot uint64, value uint64) error`
- `SettleAuction(slot uint64) (*AuctionSettlement, error)`
- `RecordDelivery(builderID string, delivered bool)`
- `BuilderScore(builderID string) (float64, error)`
- `GetReputation(builderID string) (float64, bool)`

**MEVBurnCalculator** (`builder_coordinator.go`)
- `MEVBurnCalculator{BurnRate float64}`
- `Calculate(bidValue uint64) uint64` — `floor(bidValue * burnRate)`
- `CalculateWithDetails(bidValue uint64) (burnAmount, proposerPayment uint64)`

Parent package: [`engine`](../README.md)
