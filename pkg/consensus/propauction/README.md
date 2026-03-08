# propauction

Sealed-bid Vickrey proposer auctions with VRF fallback (APS, ePBS-compatible).

## Overview

Package `propauction` implements `AuctionedProposerSelection` (APS): validators
submit sealed `AuctionBid` values for a target slot; on close, the highest
bidder wins but pays only the second-highest price (Vickrey / second-price
sealed-bid auction). When no bids are placed, a VRF-derived deterministic
fallback proposer is used. Committee rotation is performed once per epoch using
Fisher-Yates with a Keccak256-based seed.

The package is compatible with the ePBS builder API — the `BlockCommitment`
field in each bid binds the winning builder to a specific payload root before
reveal.

## Functionality

### Types

| Name | Description |
|------|-------------|
| `AuctionedProposerSelection` | Thread-safe auction manager |
| `AuctionedProposerConfig` | `MinBid`, `MaxAuctionSlots`, `FallbackEnabled` |
| `ProposerAuction` | Per-slot sealed-bid auction state |
| `AuctionBid` | Bidder, slot, amount (Gwei), block commitment, BLS signature |
| `AuctionClearing` | Slot, winner, winning bid, Vickrey clearing price, bid count |
| `ProposerScheduleEntry` | Slot, proposer index, `IsAuctioned`, clearing price |
| `CommitteeRotationEntry` | Epoch, shuffled validator list, seed |

### Functions / methods

| Name | Description |
|------|-------------|
| `DefaultAuctionedProposerConfig() AuctionedProposerConfig` | 1 ETH min bid, 32-slot lookahead, fallback enabled |
| `NewAuctionedProposerSelection(config) *AuctionedProposerSelection` | Create manager |
| `(*AuctionedProposerSelection).OpenAuction(slot) error` | Open a new sealed-bid auction |
| `(*AuctionedProposerSelection).SubmitBid(bid) error` | Add a sealed bid |
| `(*AuctionedProposerSelection).CloseAuction(slot) (*AuctionClearing, error)` | Close and compute Vickrey winner |
| `(*AuctionedProposerSelection).FallbackProposer(slot, validators, seed) uint64` | VRF-based fallback selection |
| `(*AuctionedProposerSelection).RotateCommittee(epoch, validators, seed) *CommitteeRotationEntry` | Epoch committee shuffle |
| `(*AuctionedProposerSelection).GetScheduleEntry(slot) (*ProposerScheduleEntry, bool)` | Look up assigned proposer |
| `(*AuctionedProposerSelection).GetClearing(slot) (*AuctionClearing, bool)` | Look up auction result |

### Errors

`ErrAuctionSlotPast`, `ErrAuctionAlreadyOpen`, `ErrAuctionNotOpen`, `ErrAuctionAlreadyClosed`, `ErrAuctionDuplicateBid`, `ErrAuctionZeroBid`, `ErrAuctionNoBids`, `ErrAuctionInvalidCommit`

## Usage

```go
import "github.com/eth2030/eth2030/consensus/propauction"

aps := propauction.NewAuctionedProposerSelection(propauction.DefaultAuctionedProposerConfig())
aps.OpenAuction(slot)
aps.SubmitBid(&propauction.AuctionBid{Bidder: idx, Slot: slot, Amount: bidGwei, ...})
clearing, err := aps.CloseAuction(slot)
// clearing.Winner pays clearing.ClearingPrice (Vickrey)
```

[← consensus](../README.md)
