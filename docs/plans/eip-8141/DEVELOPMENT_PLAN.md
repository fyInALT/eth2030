# EIP-8141 Development Plan

## Overview

This document outlines the development plan for EIP-8141 (Frame Transaction) in eth2030, based on the implementation analysis documented in `IMPLEMENTATION_REPORT.md`.

---

## 2. Current Implementation Status Summary

| Component | Status | Priority |
|-----------|--------|----------|
| Transaction Type (`0x06`) | ✅ Complete | - |
| RLP Encoding/Decoding | ✅ Complete | - |
| Signature Hash (VERIFY elision) | ✅ Complete | - |
| APPROVE Opcode (`0xaa`) | ✅ Complete | - |
| TXPARAM* Opcodes (`0xb0-0xb2`) | ✅ Complete | - |
| ORIGIN Opcode Modification | ✅ Complete | - |
| Frame Execution (DEFAULT/VERIFY/SENDER) | ✅ Complete | - |
| State Transition | ✅ Complete | - |
| Transaction Pool Validation | ✅ Complete | - |
| Receipt Generation | ⚠️ Needs Integration | High |

---

## 3. Identified Gaps

### 3.1 Receipt Integration Gap (HIGH PRIORITY)

**Issue**: `FrameTxReceipt` and `ExtendedFrameTxReceipt` are separate types from the standard `Receipt`. This may cause issues with:

1. **Database Storage**: The `rawdb` layer stores `[]*types.Receipt`, not `FrameTxReceipt`
2. **RPC Compatibility**: `eth_getTransactionReceipt` returns `types.Receipt`, not `FrameTxReceipt`
3. **Bloom Filter**: Standard receipts compute bloom filters; frame receipts have separate logic

**Current State**:
- `FrameTxReceipt` defined in `pkg/core/types/frame_receipt.go`
- `Receipt` defined in `pkg/core/types/receipt.go`
- No conversion layer found between the two types

**Required Actions**:

```go
// Option A: Convert FrameTxReceipt to standard Receipt
func (r *FrameTxReceipt) ToReceipt(txHash Hash, blockHash Hash, blockNumber *big.Int, txIndex uint) *Receipt {
    return &Receipt{
        Type:              FrameTxType,
        Status:            deriveStatusFromFrames(r.FrameResults),
        CumulativeGasUsed: r.CumulativeGasUsed,
        Logs:              r.AllLogs(),
        // ... other fields
    }
}

// Option B: Store FrameTxReceipt separately and retrieve on demand
```

**Files to Modify**:
- `pkg/core/types/frame_receipt.go` - Add conversion method
- `pkg/core/execution/receipt_generation.go` - Handle FrameTx type
- `pkg/core/rawdb/chaindb.go` - Support FrameTxReceipt storage/retrieval
- `pkg/rpc/types/types.go` - Extend `RPCReceipt` for frame fields

---

### 3.2 RPC Response Enhancement (MEDIUM PRIORITY)

**Issue**: The RPC receipt response should include frame-specific fields for better UX.

**EIP-8141 Spec**:
```
ReceiptPayload = [cumulative_gas_used, payer, [frame_receipt, ...]]
```

**Current RPCReceipt**:
```go
type RPCReceipt struct {
    TransactionHash   string   `json:"transactionHash"`
    // ... standard fields
    // Missing: payer, frameResults
}
```

**Required Actions**:
- Add `Payer` field to `RPCReceipt`
- Add optional `FrameResults` for FrameTx receipts
- Ensure backward compatibility for non-frame receipts

---

### 3.3 Devnet Integration Testing (HIGH PRIORITY)

**Issue**: Need end-to-end testing with real Frame transactions on devnet.

**Test Scenarios**:

| Test ID | Scenario | Expected Result |
|---------|----------|-----------------|
| FT-01 | Simple transfer (VERIFY + SENDER frames) | Success, correct payer |
| FT-02 | Sponsored transaction (different payer) | Payer charged, not sender |
| FT-03 | Failed VERIFY frame (no APPROVE) | Transaction invalid |
| FT-04 | SENDER frame without approval | Transaction invalid |
| FT-05 | Multiple VERIFY frames | All must APPROVE |
| FT-06 | Max frames (1000) | Performance test |
| FT-07 | Blob attachment | Correct blob gas |
| FT-08 | Receipt retrieval via RPC | All frame data present |

**Files to Create**:
- `pkg/devnet/kurtosis/scripts/features/verify-frame-tx.sh`

---

## 4. Development Tasks

### Phase 1: Receipt Integration (Est. 2-3 days)

1. **Add conversion method to FrameTxReceipt**
   ```go
   // pkg/core/types/frame_receipt.go
   func (r *FrameTxReceipt) ToReceipt(txHash, blockHash Hash, blockNumber *big.Int, txIndex uint, effectiveGasPrice *big.Int) *Receipt
   ```

2. **Update receipt generation**
   ```go
   // pkg/core/execution/receipt_generation.go
   func (g *ReceiptGenerator) GenerateReceipt(outcome *TxExecutionOutcome, txIndex uint, frameReceipt *types.FrameTxReceipt) *types.Receipt
   ```

3. **Add tests for conversion**
   - `pkg/core/types/frame_receipt_test.go` - Add `TestFrameTxReceiptToReceipt`

### Phase 2: RPC Enhancement (Est. 1-2 days)

1. **Extend RPCReceipt type**
   ```go
   // pkg/rpc/types/types.go
   type RPCReceipt struct {
       // ... existing fields
       Payer        *string        `json:"payer,omitempty"`
       FrameResults *[]FrameResult `json:"frameResults,omitempty"`
   }
   ```

2. **Update FormatReceipt**
   - Handle FrameTx type detection
   - Populate frame-specific fields

3. **Add tests for RPC responses**
   - `pkg/rpc/api_test.go` - Add `TestGetTransactionReceipt_FrameTx`

### Phase 3: Devnet Testing (Est. 2-3 days)

1. **Create verification script**
   - Build on existing `verify-native-aa.sh` pattern
   - Test all scenarios in section 3.3

2. **Update spamoor scenarios**
   - Add FrameTx generation
   - Test sponsored transactions

3. **Run full-feature devnet**
   - Verify block production
   - Check CL/EL synchronization

---

## 5. Code Review Checklist

Before marking EIP-8141 as complete, verify:

### Transaction Flow
- [ ] `eth_sendRawTransaction` accepts FrameTx (type `0x06`)
- [ ] Transaction pool validates FrameTx correctly
- [ ] VERIFY frames require code at target
- [ ] SENDER frames require prior approval

### Execution Flow
- [ ] DEFAULT mode: caller = ENTRY_POINT
- [ ] VERIFY mode: StaticCall, APPROVE required
- [ ] SENDER mode: caller = sender, approval required
- [ ] Transient storage cleared between frames
- [ ] Gas accounted correctly per frame

### Opcodes
- [ ] APPROVE(0) sets SenderApproved
- [ ] APPROVE(1) sets PayerApproved (requires sender approval)
- [ ] APPROVE(2) sets both
- [ ] TXPARAM* returns correct values
- [ ] ORIGIN returns frame caller

### Receipts
- [ ] FrameTxReceipt stored correctly
- [ ] RPC returns payer address
- [ ] Bloom filter computed correctly
- [ ] Logs aggregated from all frames

### Edge Cases
- [ ] Invalid scope values rejected
- [ ] Double approval rejected
- [ ] Insufficient payer balance handled
- [ ] Max frames enforced

---

## 6. Related EIPs

| EIP | Relationship |
|-----|--------------|
| EIP-2718 | Typed transaction envelope |
| EIP-4844 | Blob support in FrameTx |
| EIP-7701 | Alternative AA approach (Stagnant) |
| EIP-7702 | SetCode for account code delegation |

---

## 7. References

- [EIP-8141 Specification](../../refs/EIPs/EIPS/eip-8141.md)
- [Implementation Report](./IMPLEMENTATION_REPORT.md)
- [Frame Transaction Tests](../../pkg/core/types/tx_frame_test.go)
- [Opcode Tests](../../pkg/core/vm/eip8141_opcodes_test.go)