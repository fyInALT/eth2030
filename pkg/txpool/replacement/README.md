# txpool/replacement — Replace-by-fee policy and transaction ordering helpers

## Overview

Package `replacement` contains two complementary layers. `RBFPolicyEngine` enforces replace-by-fee rules: it validates minimum fee bumps (10% default for EIP-1559, 100% for blob fee caps on EIP-4844 txs), tracks per-account replacement chains, and caps replacement depth to prevent spam. `ReplacementPolicy` is a simpler stateless helper used by the queue and pending layers for direct price-bump checks.

The package also exports standalone functions for EIP-1559-aware gas price ordering, tip computation, nonce grouping, and promotable transaction selection that are reused across the txpool.

## Functionality

**Types**
- `RBFPolicyEngine` — `ValidateReplacement(sender, existing, newTx)`, `ReplacementCount`, `AccountReplacementDepth`, `ReplacementChain`, `ClearAccount`, `ClearNonce`, `Stats`, `MinFeeBumpRequired`, `MinBlobFeeBumpRequired`
- `RBFStats` — counters for attempts, accepted, rejected, fee/tip/blob/spam/chain rejects
- `ReplacementPolicy` — `CanReplace(existing, newTx)`, `ComputePriceBump`
- `AccountPending{Nonce, Transactions}` — `Executable`, `Len`

**Functions**
- `EffectiveGasPrice(tx, baseFee)`, `EffectiveTip(tx, baseFee)` — EIP-1559-aware price/tip
- `SortByPrice(txs, baseFee)` — sort descending by effective price
- `GetPromotable(pending, baseFee)` — all executable txs sorted by price
- `BestByNonce(txs, baseFee)` — deduplicate by nonce, keeping highest-priced
- `FilterByMinTip(txs, baseFee, minTip)`, `GroupByNonce(txs)`

## Usage

```go
engine := replacement.NewRBFPolicyEngine(replacement.DefaultRBFPolicyConfig())
if err := engine.ValidateReplacement(sender, existing, newTx); err != nil {
    return err // ErrRBFInsufficientFeeBump, ErrRBFMaxReplacements, etc.
}
engine.ClearNonce(sender, minedNonce)
```

[← txpool](../README.md)
