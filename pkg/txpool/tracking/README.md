# txpool/tracking — Per-account nonce and balance tracking

## Overview

Package `tracking` provides two complementary account-state trackers for the transaction pool. `NonceTracker` manages expected next-nonce state, detects nonce gaps in pending transactions, and enforces a configurable maximum nonce lookahead to reject transactions that are too far ahead of the on-chain state. `AcctTrack` adds balance reservation tracking: it records the cost of each pending transaction (gas * gasPrice + value + blob cost), computes available balance, and detects balance deficits and nonce gaps across all tracked accounts.

Both types lazily load state from an underlying `StateDB` interface and support batch refresh after chain reorganizations.

## Functionality

**Types**
- `NonceTracker` — `TrackTx`, `UntrackTx`, `GetNonce`, `SetNonce`, `DetectGap`, `IsTooFarAhead`, `AllGaps`, `KnownNonces`, `Reset`
- `NonceGap{Address, Expected, TxNonce}` — gap descriptor
- `AcctTrack` — `Track`, `Untrack`, `AddPendingTx`, `RemovePendingTx`, `ReplacePendingTx`, `GetPendingNonce`, `GetInfo`, `CheckBalanceDeficit`, `DetectNonceGaps`, `ResetOnReorg`, `RefreshBatch`, `MarkDirty`, `AccountsWithDeficit`, `AccountsWithGaps`
- `AcctInfo{StateNonce, PendingNonce, StateBalance, ReservedBalance, PendingTxs}` — `AvailableBalance`, `HasBalanceDeficit`, `NonceGaps`

**Interfaces**
- `NonceStateReader` — `GetNonce(addr)`
- `AccountStateReader` — `GetNonce(addr)`, `GetBalance(addr)`

## Usage

```go
tracker := tracking.NewAcctTrack(stateDB)
tracker.AddPendingTx(sender, tx)
if err := tracker.CheckBalanceDeficit(sender); err != nil { /* evict */ }
invalidated := tracker.ResetOnReorg(newState)
```

[← txpool](../README.md)
