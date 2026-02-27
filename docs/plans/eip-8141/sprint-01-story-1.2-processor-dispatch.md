# Sprint 1, Story 1.2 — processor.go: FrameTx Dispatch

**Sprint goal:** Wire EIP-8141 frame transactions into the execution pipeline.
**Files modified:** `pkg/core/processor.go`
**Files tested:** `pkg/core/frame_processor_test.go`

## Overview

The `applyMessage()` function in `pkg/core/processor.go` is the central dispatch point for all transaction types. Without an explicit FrameTx branch, type 0x06 transactions silently execute as regular calls — frames are ignored and APPROVE scopes never run.

## Gap (GAP-FRAME1)

**Severity:** CRITICAL
**File:** `pkg/core/processor.go:796` — `applyMessage()`
**Evidence:** `applyMessage()` dispatches all transaction types (create vs call) but had zero references to `FrameTxType`, `ExecuteFrameTx`, or `FrameContext` anywhere in the function. The complete `ExecuteFrameTx()` at `pkg/core/frame_execution.go:47` was never called.

## Write Failing Tests

```go
// pkg/core/frame_processor_test.go
func TestApplyMessage_FrameTxCallsExecuteFrameTx(t *testing.T) {
    // Create a FrameTx with one frame targeting a contract.
    frameTx := &types.FrameTx{
        ChainID:  1,
        Nonce:    0,
        GasLimit: 100000,
        Sender:   types.Address{0x01},
        Frames: []types.Frame{
            {Target: types.Address{0xAA}, Mode: types.ModeDefault, Data: []byte{0x01}},
        },
    }
    tx := types.NewTransaction(frameTx)
    msg := TransactionToMessage(tx)

    // msg.TxType should be FrameTxType
    if msg.TxType != types.FrameTxType {
        t.Fatalf("expected TxType %d, got %d", types.FrameTxType, msg.TxType)
    }
    if len(msg.Frames) != 1 {
        t.Fatalf("expected 1 frame in message")
    }
}
```

## Implement

In `applyMessage()`, add a FrameTx branch at line 1024 that:

1. Constructs a `FrameContext` with transaction parameters
2. Creates a `callFn` closure that wraps `evm.Call()` for each frame
3. Calls `ExecuteFrameTx()` to process all frames sequentially
4. Handles nonce increment after APPROVE (not before)

```go
// pkg/core/processor.go:1024
if msg.TxType == types.FrameTxType && len(msg.Frames) > 0 {
    frameTx := &types.FrameTx{
        ChainID:  config.ChainID.Uint64(),
        Nonce:    msg.Nonce,
        GasLimit: msg.GasLimit,
        Sender:   msg.FrameSender,
        Frames:   msg.Frames,
    }

    // Initialize FrameContext in the EVM for APPROVE/TXPARAM opcodes.
    evm.FrameCtx = &vm.FrameContext{
        TxType:         uint64(types.FrameTxType),
        Nonce:          msg.Nonce,
        // ... populate all TXPARAM fields ...
    }

    callFn := func(frameIndex int, target types.Address, data []byte, ...) (...) {
        // Clear transient storage between frames (EIP-1153 isolation).
        if frameIndex > 0 {
            statedb.ClearTransientStorage()
        }
        // Reset per-frame APPROVE tracking.
        evm.FrameCtx.ApproveCalledThisFrame = false

        // Execute frame via evm.Call().
        ret, remainGas, callErr := evm.Call(...)

        // Read APPROVE state from FrameCtx (set by opApprove).
        approved := evm.FrameCtx.ApproveCalledThisFrame
        approveScope := evm.FrameCtx.LastApproveScope

        logs := statedb.GetLogs(msg.TxHash)
        return status, gasUsed, logs, approved, approveScope, callErr
    }

    frameCtx, frameErr = ExecuteFrameTx(frameTx, stateNonce, callFn)

    // Nonce increment: only after successful APPROVE (scope 0).
    if frameCtx != nil && frameCtx.SenderApproved {
        statedb.SetNonce(msg.From, statedb.GetNonce(msg.From)+1)
    }
}
```

**Key design decisions:**
- Nonce guard at line 881 changed from `if !isCreate` to `if !isCreate && msg.TxType != types.FrameTxType` to skip early nonce increment for FrameTx
- Transient storage cleared between frames via `statedb.ClearTransientStorage()` in callFn
- APPROVE tracking uses dedicated FrameContext fields (see Story 2.1)

## Format & Commit

```bash
git add pkg/core/processor.go
git commit -m "core: dispatch FrameTx (type 0x06) through ExecuteFrameTx in applyMessage"
```

## EIP-8141 Spec Reference

> Frame transactions are executed by processing each frame sequentially. The transaction sender's nonce is incremented only after the APPROVE scope validates the sender. If APPROVE is not called or fails, the nonce is not incremented and the entire transaction is reverted.

## Codebase Locations

| File | Purpose |
|------|---------|
| `pkg/core/processor.go:796` | applyMessage entry point |
| `pkg/core/processor.go:881` | Nonce guard (skip for FrameTx) |
| `pkg/core/processor.go:1024` | FrameTx dispatch branch |
| `pkg/core/frame_execution.go:47` | ExecuteFrameTx called by processor |
| `pkg/core/vm/eip8141_opcodes.go:44` | FrameContext struct |
