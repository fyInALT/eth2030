# Sprint 2, Story 2.2 — Transient Storage Isolation Between Frames

**Sprint goal:** Ensure EIP-1153 transient storage is isolated per frame.
**Files modified:** `pkg/core/processor.go`, `pkg/core/frame_execution.go`

## Overview

EIP-8141 specifies that each frame in a frame transaction has isolated transient storage (EIP-1153). TSTORE/TLOAD values from frame N must not be visible in frame N+1.

## Gap (GAP-FRAME3)

**Severity:** IMPORTANT
**File:** `pkg/core/frame_execution.go:70`
**Evidence:** `ExecuteFrameTx()` loops through frames but never calls `statedb.ClearTransientStorage()` between them.

## Implement

Transient storage clearing is performed in the callFn closure in processor.go, not in frame_execution.go, because only the processor has access to the statedb:

```go
// pkg/core/processor.go — inside callFn closure
callFn := func(frameIndex int, target types.Address, ...) (...) {
    // Clear transient storage between frames (EIP-1153 isolation).
    if frameIndex > 0 {
        statedb.ClearTransientStorage()
    }
    // ... execute frame ...
}
```

A comment was added to `frame_execution.go` noting that transient storage isolation is handled by the callFn callback:

```go
// pkg/core/frame_execution.go
// Note: transient storage (EIP-1153) isolation between frames is handled
// by the callFn callback in processor.go, which calls
// statedb.ClearTransientStorage() before each frame (except the first).
```

## EIP-8141 Spec Reference

> Each frame executes in an isolated transient storage context. TSTORE writes in frame i are not visible to TLOAD reads in frame i+1.
