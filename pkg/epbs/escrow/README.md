# epbs/escrow — ePBS builder collateral escrow and payment processing

## Overview

Package `escrow` manages the two-layer financial settlement system for ePBS builder bids (EIP-7732). `BidEscrow` tracks per-builder collateral deposits and per-slot bid locking: when a builder places a bid, collateral equal to the bid value is locked; on successful payload delivery the collateral is released, and on failure it is slashed. `PaymentProcessor` sits on top of `BidEscrow` and implements the full EIP-7732 payment flow with configurable slash and burn fractions, settlement deadlines, and an immutable audit trail.

Both components are safe for concurrent use and support pruning of settled or expired entries.

## Functionality

**BidEscrow** (`bid_escrow.go`)
- `EscrowBidState` enum: `EscrowBidPending`, `EscrowBidRevealed`, `EscrowBidSettledSuccess`, `EscrowBidSettledSlashed`
- `SettlementResult{Slot, BuilderID, AmountReleased, AmountSlashed uint64, Success bool, Reason string, SettledAt time.Time}`
- `NewBidEscrow(maxResults int) *BidEscrow`
- `Deposit(builderID string, amount uint64) error`
- `PlaceBid(bid *epbs.BuilderBid) error` — locks collateral equal to bid value; one bid per slot
- `RevealPayload(slot uint64, builderID string, payload *epbs.PayloadEnvelope) error` — validates slot, builder index, and `payload.PayloadRoot == bid.BlockHash`
- `SettleBid(slot uint64) (*SettlementResult, error)` — releases on `Revealed`; slashes on `Pending`
- `SlashBuilder(builderID string, amount uint64, reason string) error` — drains available then locked balance
- `WithdrawBalance(builderID string, amount uint64) error`
- `GetBalance(builderID string) uint64` / `GetLockedBalance(builderID string) uint64`
- `GetBidState(slot uint64) (EscrowBidState, bool)` / `GetBid(slot uint64) *epbs.BuilderBid`
- `ActiveBidCount() int`
- `SettlementHistory(n int) []*SettlementResult`
- `PruneBefore(slot uint64) int` — only settled bids are pruned

**PaymentProcessor** (`payment.go`)
- `PaymentState` enum: `Pending`, `Escrowed`, `Released`, `Slashed`, `Refunded`
- `PaymentConfig{SlashFraction=5000 bps (50%), BurnFraction=5000 bps, SettlementDeadline=32 slots}`
- `EscrowRecord` — full audit record with committed hash, delivered hash, timestamps, and final state
- `NewPaymentProcessor(cfg PaymentConfig) *PaymentProcessor`
- `Escrow(slot uint64, builderID string, amount uint64, committedHash Hash) error`
- `ReleasePayment(slot uint64, deliveredHash Hash) error` — verifies `deliveredHash == committedHash`
- `SlashBuilder(slot uint64, reason string) error` — slash fraction burned, remainder to proposer
- `RefundEscrow(slot uint64, currentSlot uint64) error` — deadline-gated refund
- `History(n int) []*EscrowRecord`
- `PruneBefore(slot uint64) int`

Parent package: [`epbs`](../README.md)
