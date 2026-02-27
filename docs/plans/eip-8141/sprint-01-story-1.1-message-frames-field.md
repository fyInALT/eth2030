# Sprint 1, Story 1.1 — Message Struct: Add Frames Field

**Sprint goal:** Wire EIP-8141 frame transactions into the execution pipeline.
**Files modified:** `pkg/core/message.go`
**Files tested:** `pkg/core/frame_processor_test.go`

## Overview

The `Message` struct in `pkg/core/message.go` is the bridge between the transaction envelope (`types.Transaction`) and the EVM execution path (`applyMessage`). Without a `Frames` field, all frame data is discarded during `TransactionToMessage()` conversion, making FrameTx (type 0x06) execution impossible.

## Gap (GAP-FRAME2)

**Severity:** CRITICAL
**File:** `pkg/core/message.go:10-24`
**Evidence:** The `Message` struct had fields for From, To, Nonce, Value, GasLimit, GasPrice, GasFeeCap, GasTipCap, Data, AccessList, BlobHashes, AuthList, TxType — but **NO** `Frames` field.

`TransactionToMessage()` at line 32 converted transactions but discarded all frame structure. Even if `processor.go` dispatched FrameTx, the frame data would be lost.

## Write Failing Tests

```go
// pkg/core/frame_processor_test.go
func TestTransactionToMessage_FrameTxPopulatesFrames(t *testing.T) {
    frameTx := &types.FrameTx{
        ChainID: 1,
        Nonce:   0,
        Sender:  types.Address{0x01},
        Frames: []types.Frame{
            {Target: types.Address{0xAA}, Mode: types.ModeDefault, Data: []byte{0x01}},
        },
    }
    tx := types.NewTransaction(frameTx)
    msg := TransactionToMessage(tx)

    if len(msg.Frames) != 1 {
        t.Fatalf("expected 1 frame, got %d", len(msg.Frames))
    }
    if msg.FrameSender != (types.Address{0x01}) {
        t.Fatalf("expected frame sender 0x01, got %x", msg.FrameSender)
    }
}
```

```bash
cd pkg && go test ./core/ -run TestTransactionToMessage_FrameTxPopulatesFrames -v
# Expected: PASS
```

## Implement

Add three fields to `Message` and populate them in `TransactionToMessage()`:

```go
// pkg/core/message.go:10
type Message struct {
    // ... existing fields ...
    TxType      uint8
    Frames      []types.Frame   // EIP-8141 frame transaction frames
    FrameSender types.Address   // EIP-8141 frame tx sender (from FrameTx.Sender)
    TxHash      types.Hash      // transaction hash for log attribution
}

// pkg/core/message.go:54
func TransactionToMessage(tx *types.Transaction) Message {
    msg := Message{
        // ... existing field population ...
        TxHash: tx.Hash(),
    }
    // EIP-8141: populate frame data for FrameTx type.
    if tx.Type() == types.FrameTxType {
        msg.Frames = tx.Frames()
        msg.FrameSender = tx.FrameSender()
    }
    return msg
}
```

**Key decisions:**
- `TxHash` added for all tx types (needed by log retrieval in callFn — see Story 3.1)
- `FrameSender` extracted from `FrameTx.Sender` field (not `tx.From()` which may differ)

## Format & Commit

```bash
git add pkg/core/message.go pkg/core/frame_processor_test.go
git commit -m "core: add Frames/FrameSender/TxHash to Message for EIP-8141 frame tx"
```

## EIP-8141 Spec Reference

> A frame transaction contains an ordered list of frames. Each frame specifies a target address, execution mode, and calldata. The transaction envelope carries the frame list alongside standard fields.

## Codebase Locations

| File | Purpose |
|------|---------|
| `pkg/core/message.go:10` | Message struct definition |
| `pkg/core/message.go:32` | TransactionToMessage conversion |
| `pkg/core/types/tx_frame.go:40` | FrameTx struct with Frames field |
| `pkg/core/types/transaction.go` | Frames() and FrameSender() accessors |
