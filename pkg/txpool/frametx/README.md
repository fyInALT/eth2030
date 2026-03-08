# txpool/frametx — Frame transaction validation rules and metrics

## Overview

Package `frametx` implements pool-side validation rules for EIP-8141 frame transactions (`FrameTx`). Frame transactions compose multiple execution frames with per-frame gas limits and execution modes. The package provides two rulesets: a conservative ruleset that caps VERIFY-mode frame gas at 50K, and an aggressive ruleset that raises the cap to 200K when a frame targets a registered staked paymaster. It also provides Prometheus counters for accepted and rejected frame transactions.

## Functionality

**Types**
- `ConservativeFrameRules` — strict ruleset; `Validate(tx *types.FrameTx) error`
- `AggressiveFrameRules{Registry PaymasterApprover}` — relaxed ruleset with paymaster detection; `Validate`
- `FrameRuleError{FrameIndex, Reason}` — structured per-frame error
- `FrameTxMetrics` — Prometheus counters; `IncAccepted`, `IncRejectedConservative`, `IncRejectedAggressive`
- `PaymasterApprover` — interface for paymaster registry (`IsApprovedPaymaster`)

**Functions**
- `ValidateFrameTxConservative(tx)` — validate under conservative rules (50K VERIFY cap)
- `ValidateFrameTxAggressive(tx, registry)` — validate under aggressive rules (200K cap if staked paymaster)
- `NewFrameTxMetrics()` — create and register Prometheus counters

**Constants**
- `ConservativeVerifyGasLimit = 50_000`
- `AggressiveVerifyGasLimit = 200_000`

## Usage

```go
if err := frametx.ValidateFrameTxConservative(frameTx); err != nil {
    metrics.IncRejectedConservative()
    return err
}
metrics.IncAccepted()
```

[← txpool](../README.md)
