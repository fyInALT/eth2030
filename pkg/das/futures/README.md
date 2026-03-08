# das/futures — Short-dated blob availability futures market

[← das](../README.md)

## Overview

This package implements short-dated blob futures contracts for the Data Layer throughput roadmap track (Hegotá: "short-dated blob futures"). A `BlobFuture` represents a bet that a specific blob will be available by a given expiry slot. Futures are priced in wei and tracked by a `FuturesMarket` that maps futures by both ID and expiry slot for efficient settlement.

`blob_futures.go` ties futures to the actual blob sampling result: when a slot is finalized, the market settles all futures expiring at that slot based on observed blob availability. `futures_market.go` provides the primary on-chain settlement model with volume tracking.

## Functionality

**Types**
- `BlobFuture` — `ID types.Hash`, `ExpirySlot uint64`, `BlobHash types.Hash`, `Price *big.Int`, `Creator types.Address`, `Settled bool`
- `FuturesMarket` — manages `active` futures and `byExpiry` index; tracks `totalVolume`

**Market operations**
- `NewFuturesMarket(currentSlot uint64) *FuturesMarket`
- `(m *FuturesMarket) CreateFuture(creator types.Address, blobHash types.Hash, expirySlot uint64, price *big.Int) (*BlobFuture, error)`
- `(m *FuturesMarket) GetFuture(id types.Hash) (*BlobFuture, bool)`
- `(m *FuturesMarket) SettleFuture(id types.Hash, blobAvailable bool) error`
- `(m *FuturesMarket) SettleSlot(slot uint64, availableBlobs map[types.Hash]bool) (settled int, err error)`
- `(m *FuturesMarket) AdvanceSlot(slot uint64)` — updates current slot, expires outstanding futures
- `(m *FuturesMarket) Stats() (active, total int, volume *big.Int)`

**Errors**
- `ErrFutureNotFound`, `ErrFutureExpired`, `ErrFutureSettled`, `ErrInvalidExpiry`, `ErrInvalidPrice`

## Usage

```go
market := futures.NewFuturesMarket(currentSlot)

fut, err := market.CreateFuture(creator, blobHash, currentSlot+32, price)
// ... slot advances ...
settled, err := market.SettleSlot(slot, availableBlobs)
```
