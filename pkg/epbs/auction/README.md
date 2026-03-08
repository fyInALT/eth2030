# epbs/auction — ePBS payload auction engine and self-build support

## Overview

Package `auction` implements the builder auction layer for Enshrined Proposer-Builder Separation (EIP-7732). It exposes two complementary components: `PayloadAuction` provides a lightweight per-slot bid store with sorted insertion for fast winner lookup, while `AuctionEngine` manages the full four-state auction lifecycle (Open → BiddingClosed → WinnerSelected → Finalized) with equivocation and non-delivery slashing records.

The package also implements the EIP-7732 self-build special case: when the proposer builds its own block, a sentinel `BuilderIndexSelfBuild = UINT64_MAX` is used to bypass the auction path entirely, and the domain constant `DomainProposerPreferences = 0x0D000000` scopes preference messages to the proposer preferences domain.

## Functionality

**PayloadAuction** (`auction.go`)
- `NewPayloadAuction() *PayloadAuction`
- `SubmitBid(bid *epbs.BuilderBid) error` — inserts bid in descending-value order
- `GetWinningBid(slot uint64) (*epbs.BuilderBid, bool)` — returns highest-value bid
- `GetBidsForSlot(slot uint64) []*epbs.BuilderBid`
- `BidCount(slot uint64) int`
- `PruneSlot(slot uint64)` / `PruneBefore(slot uint64) int`

**AuctionEngine** (`auction_engine.go`)
- `AuctionBid{BuilderPubkey [48]byte, Slot, Value *big.Int, PayloadHash Hash, Timestamp time.Time, Signature [96]byte}`
- `SlashingViolation{BuilderPubkey [48]byte, Slot uint64, Reason string, DetectedAt time.Time}`
- State constants: `AuctionStateOpen`, `AuctionStateBiddingClosed`, `AuctionStateWinnerSelected`, `AuctionStateFinalized`
- `NewAuctionEngine() *AuctionEngine`
- `OpenAuction(slot uint64) error`
- `SubmitBid(bid *AuctionBid) error`
- `CloseBidding(slot uint64) error`
- `SelectWinner(slot uint64) (*AuctionBid, error)` — highest value; earliest timestamp as tiebreak
- `FinalizeAuction(slot uint64) error`
- `RecordViolation(v SlashingViolation)`
- `Violations() []SlashingViolation`
- `History(slot uint64) []*AuctionBid`

**Self-build support** (`builder_selfbuild.go`)
- `BuilderIndexSelfBuild = epbs.BuilderIndex(^uint64(0))` — UINT64_MAX sentinel per EIP-7732
- `DomainProposerPreferences uint32 = 0x0D000000`
- `ProcessBidWithSelfBuild(engine *AuctionEngine, bid *AuctionBid) error` — bypasses auction when builder index equals self-build sentinel

Parent package: [`epbs`](../README.md)
