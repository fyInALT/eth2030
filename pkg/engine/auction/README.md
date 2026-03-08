# engine/auction — Vickrey block-builder auction

Implements a second-price sealed-bid (Vickrey) auction for distributed block
construction rights. Builders register with stake, submit bids per slot, and
the highest bidder wins while paying the second-highest price.

## Overview

`BuilderAuction` maintains a registry of staked builders and a per-slot bid
store. Builders that misbehave can be slashed; slashed builders cannot submit
further bids. `RunAuction` sorts bids descending by value, selects the winner,
and returns the second-highest bid price as the clearing price.

All methods are safe for concurrent use via `sync.RWMutex`.

## Functionality

**Types**
- `AuctionConfig` — `MinBid`, `MaxBid`, `AuctionDeadline`, `MinStake`
- `AuctionBid` — `BuilderID`, `Slot`, `Value`, `GasLimit`, `Payload`, `Signature`
- `AuctionResult` — `WinnerID`, `WinningValue`, `SecondPrice`, `TotalBids`
- `BuilderAuction` — thread-safe auction manager

**Functions**
- `DefaultAuctionConfig() AuctionConfig`
- `NewBuilderAuction(config) *BuilderAuction`
- `(*BuilderAuction).RegisterBuilder(builderID, stake) error`
- `(*BuilderAuction).SlashBuilder(builderID, reason) error`
- `(*BuilderAuction).ValidateBid(bid) error`
- `(*BuilderAuction).SubmitBid(bid) error`
- `(*BuilderAuction).GetWinningBid(slot) (*AuctionBid, error)`
- `(*BuilderAuction).RunAuction(slot) (*AuctionResult, error)`
- `(*BuilderAuction).GetBidHistory(slot) []*AuctionBid`
- `(*AuctionBid).Hash() types.Hash`

## Usage

```go
auction := auction.NewBuilderAuction(auction.DefaultAuctionConfig())
auction.RegisterBuilder(builderID, stake)
auction.SubmitBid(&auction.AuctionBid{
    BuilderID: builderID, Slot: 12345, Value: 100,
    GasLimit: 30_000_000, Payload: payloadBytes, Signature: sig,
})
result, _ := auction.RunAuction(12345)
fmt.Println("winner:", result.WinnerID, "price:", result.SecondPrice)
```

[← engine](../README.md)
