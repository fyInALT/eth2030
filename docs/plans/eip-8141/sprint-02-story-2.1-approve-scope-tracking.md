# Sprint 2, Story 2.1 — APPROVE Scope Tracking

**Sprint goal:** Fix APPROVE scope detection in the processor's callFn.
**Files modified:** `pkg/core/vm/eip8141_opcodes.go`, `pkg/core/processor.go`
**Files tested:** `pkg/core/frame_processor_test.go`

## Overview

The APPROVE opcode (0xAA) in EIP-8141 supports 3 scopes:
- **Scope 0**: Sender approval — the sender's validation contract approves the frame
- **Scope 1**: Payer approval — the payer's validation contract approves gas payment
- **Scope 2**: Combined sender+payer approval — both approvals in a single APPROVE call

The callFn in processor.go must correctly report which scope was used so ExecuteFrameTx can track approval state.

## Gap (AUDIT-1)

**Severity:** CRITICAL
**File:** `pkg/core/processor.go:1097` and `pkg/core/vm/eip8141_opcodes.go:44`

**Evidence:** The original callFn inferred APPROVE scope by checking boolean flags `SenderApproved`/`PayerApproved` after frame execution:

```go
// WRONG — old code (removed in commit 9c84089)
if evm.FrameCtx.SenderApproved && !frameTx.Sender.IsZero() {
    approved = true
    if evm.FrameCtx.PayerApproved {
        approveScope = 2  // combined
    } else {
        approveScope = 0  // sender only
    }
} else if evm.FrameCtx.PayerApproved {
    approved = true
    approveScope = 1  // payer only
}
```

**Problem:** This logic couldn't distinguish APPROVE(2) (single combined call) from APPROVE(0) followed by APPROVE(1) (two separate calls). Both would set `SenderApproved=true` and `PayerApproved=true`, but should report different scope values.

**Impact:** Frame execution would incorrectly track approval state, leading to wrong nonce/payer semantics.

## Write Failing Tests

```go
func TestApproveScope_DistinguishesCombinedFromSeparate(t *testing.T) {
    fc := &vm.FrameContext{}

    // Simulate APPROVE(2) — combined scope.
    fc.ApproveCalledThisFrame = true
    fc.LastApproveScope = 2
    if fc.LastApproveScope != 2 {
        t.Fatal("expected scope 2 for combined approval")
    }

    // Reset and simulate APPROVE(0) + APPROVE(1) — separate calls.
    fc.ApproveCalledThisFrame = true
    fc.LastApproveScope = 1 // last call was scope 1
    if fc.LastApproveScope != 1 {
        t.Fatal("expected scope 1 for last separate call")
    }
}
```

## Implement

### Step 1: Add tracking fields to FrameContext

```go
// pkg/core/vm/eip8141_opcodes.go:46
type FrameContext struct {
    // ... existing fields ...

    // APPROVE tracking: records the most recent APPROVE call within the
    // current frame so the callFn can distinguish APPROVE(2) from
    // APPROVE(0)+APPROVE(1) without relying solely on boolean flags.
    LastApproveScope       uint8
    ApproveCalledThisFrame bool

    // ... remaining fields ...
}
```

### Step 2: Set fields in opApprove

```go
// pkg/core/vm/eip8141_opcodes.go:89
func opApprove(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
    // ... scope validation ...

    // Track the APPROVE call for the callFn to read.
    fc.ApproveCalledThisFrame = true
    fc.LastApproveScope = uint8(scopeVal)

    // ... existing scope switch ...
}
```

### Step 3: Simplify callFn in processor.go

```go
// pkg/core/processor.go:1097
// Reset per-frame APPROVE tracking before each frame call.
evm.FrameCtx.ApproveCalledThisFrame = false

// ... frame execution via evm.Call() ...

// Check if APPROVE was called during this frame using explicit tracking
// fields set by opApprove. This correctly distinguishes APPROVE(2) from
// separate APPROVE(0)+APPROVE(1) calls.
approved := false
var approveScope uint8
if evm.FrameCtx != nil && evm.FrameCtx.ApproveCalledThisFrame {
    approved = true
    approveScope = evm.FrameCtx.LastApproveScope
}
```

**Key insight:** By reading `LastApproveScope` directly from FrameContext (set by opApprove at EVM level), the callFn doesn't need to infer scope from boolean state. The opApprove function is the source of truth.

## Format & Commit

```bash
git add pkg/core/vm/eip8141_opcodes.go pkg/core/processor.go
git commit -m "eip-8141: fix APPROVE scope tracking with explicit LastApproveScope field"
```

## EIP-8141 Spec Reference

> APPROVE(scope): scope ∈ {0, 1, 2}
> - 0: sender approval
> - 1: payer approval
> - 2: combined sender + payer approval
>
> The opcode must be called from within the frame's target contract. CALLER must equal frame.target.

## Codebase Locations

| File | Line | Purpose |
|------|------|---------|
| `pkg/core/vm/eip8141_opcodes.go` | 44 | FrameContext struct — added LastApproveScope, ApproveCalledThisFrame |
| `pkg/core/vm/eip8141_opcodes.go` | 72 | opApprove — sets tracking fields before scope switch |
| `pkg/core/processor.go` | 1078 | callFn — resets ApproveCalledThisFrame before each frame |
| `pkg/core/processor.go` | 1100 | callFn — reads LastApproveScope instead of inferring |
