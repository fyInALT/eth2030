# Sprint 2, Story 2.3 — Nonce Increment Timing

**Sprint goal:** Fix nonce semantics for EIP-8141 frame transactions.
**Files modified:** `pkg/core/processor.go`

## Overview

Standard transactions increment the sender's nonce before EVM execution. EIP-8141 specifies that for frame transactions, the nonce should increment during the APPROVE scope (scope 0), not before execution begins. If APPROVE fails, the nonce must not be incremented.

## Gap (GAP-FRAME4)

**Severity:** IMPORTANT
**File:** `pkg/core/processor.go:881`
**Evidence:** Nonce increment happened unconditionally before EVM execution for all non-create transactions. For FrameTx, this means the nonce was already incremented even if APPROVE failed.

## Implement

### Step 1: Guard the early nonce increment

```go
// pkg/core/processor.go:881
// Skip nonce increment for FrameTx — APPROVE handles it.
if !isCreate && msg.TxType != types.FrameTxType {
    statedb.SetNonce(msg.From, statedb.GetNonce(msg.From)+1)
}
```

### Step 2: Increment nonce after APPROVE success

```go
// pkg/core/processor.go — after ExecuteFrameTx returns
if frameCtx != nil && frameCtx.SenderApproved {
    statedb.SetNonce(msg.From, statedb.GetNonce(msg.From)+1)
}
```

**Key insight:** This matches EIP-4337 (Account Abstraction) nonce semantics where validation determines whether the nonce should be consumed.

## EIP-8141 Spec Reference

> The sender's nonce is incremented only after the APPROVE scope (scope 0) successfully validates the sender. If APPROVE is not called or execution reverts, the nonce is not incremented.
