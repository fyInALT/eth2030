# epbs/bid — Builder bid validation, scoring, and ranking

## Overview

Package `bid` provides the full validation and scoring pipeline for ePBS builder bids. Validation enforces collateral requirements, signature integrity (Keccak-256 commitment over bid hash, builder address, slot, and value), and minimum/maximum bid value bounds. Scoring weighs bid amount, builder reputation, inclusion quality, and submission latency into a composite score used to rank competing bids.

Reputation tracking uses a sliding-window model that accounts for successful reveals, failures, and slashing events. Deterministic tiebreaking by lexicographically smallest bid hash ensures a unique winner even when multiple bids share the same composite score.

## Functionality

**Bid validation** (`bid_validator.go`)
- `BidValidatorConfig{MinCollateral uint64, MinBidValue uint64, MaxBidValue uint64}` — default min collateral: 32 ETH in Gwei
- `NewBidValidator(cfg BidValidatorConfig) *BidValidator`
- `ValidateBidSignature(bid *epbs.BuilderBid, builderAddr Address) error` — verifies Keccak-256 commitment `H(bidHash || builderAddr || slot || value)`
- `ValidateBidCollateral(bid *epbs.BuilderBid, collateral uint64) error`
- `ValidateBidValue(bid *epbs.BuilderBid) error`
- `FullBidValidation(bid *epbs.BuilderBid, builderAddr Address, collateral uint64) error` — chains all checks
- `SelectWinningBid(bids []*epbs.BuilderBid) *epbs.BuilderBid` — highest value; lexicographically smallest bid hash as tiebreak

**Bid scoring** (`bid_scoring.go`)
- `BidScoreConfig{AmountWeight=0.50, ReputationWeight=0.20, InclusionWeight=0.15, LatencyWeight=0.15}`
- `ScoreComponents{BidAmount, ReputationScore, InclusionQuality float64, LatencyMs int64}`
- `NewBidScoreCalculator(cfg BidScoreConfig) *BidScoreCalculator`
- `Score(components ScoreComponents) float64` — normalized weighted sum
- `ReputationTracker` — sliding-window reliability tracker
  - `Register(builderAddr Address)`
  - `RecordBid(addr Address)` / `RecordReveal(addr Address)` / `RecordFailure(addr Address)`
  - `GetScore(addr Address) float64`
- `BidRanker`
  - `NewBidRanker(scorer *BidScoreCalculator) *BidRanker`
  - `Rank(bids []*epbs.BuilderBid, components map[Hash]ScoreComponents) []*epbs.BuilderBid` — score desc, bid hash asc tiebreak
- `TiebreakerRule`
  - `Break(a, b *epbs.BuilderBid) *epbs.BuilderBid` — lexicographic bid hash comparison
- `MinBidEnforcer{minimum uint64}`
  - `Check(bid *epbs.BuilderBid) error`
  - `SetMinimum(v uint64)`
  - `FilterBids(bids []*epbs.BuilderBid) []*epbs.BuilderBid`

Parent package: [`epbs`](../README.md)
