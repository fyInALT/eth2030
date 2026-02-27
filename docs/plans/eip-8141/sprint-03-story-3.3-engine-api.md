# Sprint 3, Story 3.3 — Engine API FrameTx Receipt Handling

**Sprint goal:** Document FrameTx receipt handling in the Engine API.
**Files modified:** `pkg/engine/backend.go`

## Overview

The Engine API's `ProcessBlockV5()` and `GetPayloadV6ByID()` need to account for FrameTx receipts, which contain per-frame gas usage, per-frame status, and per-frame logs.

## Gap (GAP-FRAME5)

**Severity:** IMPORTANT
**File:** `pkg/engine/backend.go`
**Evidence:** No special handling for FrameTx existed. Frame receipt structure was not included in payload responses.

## Implement

Documentation was added to `ProcessBlockV5()` and `GetPayloadV6ByID()` describing the expected FrameTx receipt structure:

```go
// ProcessBlockV5 processes a block and returns validation results.
// For FrameTx (type 0x06), receipts include per-frame gas breakdown
// via the Receipt.FrameReceipts field, which contains:
//   - FrameIndex: the 0-based index of the frame
//   - GasUsed: gas consumed by this frame
//   - Status: success/failure per frame
//   - Logs: events emitted during this frame
// CL clients should validate that sum(frame.GasUsed) <= receipt.GasUsed.
```

**Note:** This is a documentation-only change. The actual receipt structure is already defined in `pkg/core/frame_execution.go:BuildFrameReceipt()`. The gap was that the Engine API had no awareness of this structure.

## Future Work

When CL clients begin processing FrameTx receipts, additional validation should be added:
- Verify per-frame gas sums match total gas used
- Verify APPROVE scope is correctly reported
- Include frame receipt data in ExecutionPayloadV6 response

## Codebase Locations

| File | Purpose |
|------|---------|
| `pkg/engine/backend.go` | Engine API — ProcessBlockV5, GetPayloadV6ByID |
| `pkg/core/frame_execution.go` | BuildFrameReceipt — constructs per-frame receipt data |
