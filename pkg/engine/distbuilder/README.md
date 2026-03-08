# engine/distbuilder — Distributed block builder network

Manages a network of registered distributed builders that submit competitive
bids for block construction rights. Implements the "distributed block building"
item from the L+ roadmap.

## Overview

`BuilderNetwork` maintains a registry of `DistBuilder` entries (each with ID,
address, stake, and active status) and a per-slot bid store. Builders submit
`BuilderBid` messages containing a block hash, payload bytes, and `*big.Int`
value. `GetWinningBid` selects the highest-value bid for a slot.

This package differs from `engine/auction` in using `*big.Int` values (for
arbitrary-precision bids) and not enforcing the Vickrey second-price mechanic —
it is the lower-level P2P builder network registry, while `engine/auction`
provides the on-chain auction logic.

`PruneStaleBids` removes bids older than a given slot to bound memory growth.

## Functionality

**Types**
- `BuilderConfig` — `MaxBuilders`, `BuilderTimeout`, `MinBid`, `SlotAuctionDuration`
- `DistBuilder` — `ID`, `Address`, `Stake`, `Active`, `LastSeen`
- `BuilderBid` — `BuilderID`, `Slot`, `BlockHash`, `Value *big.Int`, `Payload`, `Timestamp`
- `BuilderNetwork` — thread-safe registry and bid store

**Functions**
- `DefaultBuilderConfig() *BuilderConfig`
- `NewBuilderNetwork(config) *BuilderNetwork`
- `(*BuilderNetwork).RegisterBuilder(id, address, stake) error`
- `(*BuilderNetwork).UnregisterBuilder(id) error`
- `(*BuilderNetwork).SubmitBid(bid) error`
- `(*BuilderNetwork).GetWinningBid(slot) *BuilderBid`
- `(*BuilderNetwork).ActiveBuilders() int`
- `(*BuilderNetwork).PruneStaleBids(beforeSlot uint64)`

## Usage

```go
net := distbuilder.NewBuilderNetwork(distbuilder.DefaultBuilderConfig())
net.RegisterBuilder(builderID, builderAddr, stake)

net.SubmitBid(&distbuilder.BuilderBid{
    BuilderID: builderID, Slot: 12345,
    Value: big.NewInt(1e9), Payload: payloadBytes,
})
winner := net.GetWinningBid(12345)
```

[← engine](../README.md)
