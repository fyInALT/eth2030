# Sprint 3, Story 3.2 — Transaction Pool Frame Validation

**Sprint goal:** Add FrameTx-specific validation to the transaction pool.
**Files modified:** `pkg/txpool/txpool.go`

## Overview

The transaction pool's `validateTx()` checks gas, nonce, and balance for standard transactions but had no FrameTx-specific validation. Invalid frame transactions could enter the pool and waste block space.

## Gap (GAP-FRAME6 + AUDIT-3 + AUDIT-7)

**Severity:** IMPORTANT → HIGH (upgraded after second audit)
**File:** `pkg/txpool/txpool.go:358`

**Round 1 evidence:** No FrameTx validation existed beyond standard checks.
**Round 2 evidence:** Only intrinsic gas was checked, not frame structure (count, modes, targets).

## Implement

```go
// pkg/txpool/txpool.go:390
if tx.Type() == types.FrameTxType {
    // Check intrinsic gas minimum.
    if tx.Gas() < types.FrameTxIntrinsicCost {
        return ErrIntrinsicGas
    }
    // Validate frame structure (count, modes, targets, blob consistency).
    frames := tx.Frames()
    if len(frames) == 0 {
        return errors.New("frame tx: must have at least one frame")
    }
    if len(frames) > types.MaxFrames {
        return errors.New("frame tx: too many frames")
    }
    for i, f := range frames {
        if f.Mode > types.ModeSender {
            return fmt.Errorf("frame tx: invalid mode %d in frame %d", f.Mode, i)
        }
    }
}
```

**Checks added:**
1. `tx.Gas() < types.FrameTxIntrinsicCost` — minimum gas for FrameTx
2. `len(frames) == 0` — must have at least one frame
3. `len(frames) > types.MaxFrames` — enforce max frame count (1000)
4. `f.Mode > types.ModeSender` — validate mode is Default(0), Verify(1), or Sender(2)

## EIP-8141 Spec Reference

> Frame transactions MUST contain at least one frame and at most MAX_FRAMES (1000) frames. Each frame's mode MUST be one of: Default (0), Verify (1), or Sender (2). The transaction MUST provide sufficient gas for the intrinsic frame cost.

## Codebase Locations

| File | Line | Purpose |
|------|------|---------|
| `pkg/txpool/txpool.go` | 390 | FrameTx validation in validateTx() |
| `pkg/core/types/tx_frame.go` | 10 | MaxFrames constant (1000) |
| `pkg/core/types/tx_frame.go` | 15 | FrameTxIntrinsicCost constant |
| `pkg/core/types/tx_frame.go` | 25 | ModeDefault/ModeVerify/ModeSender constants |
