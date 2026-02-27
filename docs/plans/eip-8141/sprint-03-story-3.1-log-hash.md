# Sprint 3, Story 3.1 — Log Hash Attribution

**Sprint goal:** Fix log retrieval in FrameTx callFn.
**Files modified:** `pkg/core/processor.go`, `pkg/core/message.go`

## Overview

After each frame executes, the callFn collects logs emitted during execution via `statedb.GetLogs()`. The logs are keyed by transaction hash in the StateDB, but the callFn was passing an empty hash.

## Gap (AUDIT-2)

**Severity:** HIGH
**File:** `pkg/core/processor.go:1118`
**Evidence:** `statedb.GetLogs(types.Hash{})` used an empty hash. The StateDB stores logs keyed by the transaction hash set via `SetTxContext()`. An empty hash would never match, returning no logs.

## Implement

### Step 1: Add TxHash to Message struct

```go
// pkg/core/message.go:25
type Message struct {
    // ... existing fields ...
    TxHash types.Hash // transaction hash for log attribution
}
```

### Step 2: Populate TxHash in TransactionToMessage

```go
// pkg/core/message.go:41
msg := Message{
    // ... existing fields ...
    TxHash: tx.Hash(),
}
```

### Step 3: Use TxHash in callFn

```go
// pkg/core/processor.go:1118
// Collect logs emitted during this frame using the actual tx hash
// so they match the key set by SetTxContext.
logs := statedb.GetLogs(msg.TxHash)
```

**Note:** `TxHash` is populated for all transaction types, not just FrameTx, since it's a generally useful field.

## Codebase Locations

| File | Line | Purpose |
|------|------|---------|
| `pkg/core/message.go` | 25 | TxHash field in Message struct |
| `pkg/core/message.go` | 41 | TxHash populated from tx.Hash() |
| `pkg/core/processor.go` | 1118 | GetLogs uses msg.TxHash |
