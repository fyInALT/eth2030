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
| Receipt Generation | ✅ Complete | - |
| Receipt Integration | ✅ Complete | - |
| RPC Enhancement | ✅ Complete | - |

---

## 3. Identified Gaps

### 3.1 Receipt Integration Gap (RESOLVED ✅)

**Status**: Complete. Added conversion methods to bridge FrameTxReceipt and standard Receipt.

**Implementation**:
- Added `ToReceipt()` and `ToReceiptWithPayer()` methods to `FrameTxReceipt` in `pkg/core/types/frame_receipt.go`
- Added `DeriveStatus()` to compute overall status from frame results
- Added `ComputeBloom()` to compute bloom filter from all logs
- Updated `receipt_generation.go` to handle FrameTx via `FrameReceipt` field in `TxExecutionOutcome`

**Files Modified**:
- `pkg/core/types/frame_receipt.go` - Added conversion methods
- `pkg/core/execution/receipt_generation.go` - Added FrameReceipt field and generateFrameReceipt method
- `pkg/core/types/frame_receipt_rlp_test.go` - Added tests for conversion methods

---

### 3.2 RPC Response Enhancement (RESOLVED ✅)

**Status**: Complete. Extended RPCReceipt with EIP-8141 frame-specific fields.

**Implementation**:
- Added `Payer` and `FrameResults` fields to `RPCReceipt` in `pkg/rpc/types/types.go`
- Added `RPCFrameResult` type for frame result representation
- Updated `FormatReceipt()` to populate Payer for FrameTx (type 0x06)
- For FrameTx receipts, `ContractAddress` field is repurposed to hold payer address

**Files Modified**:
- `pkg/rpc/types/types.go` - Extended RPCReceipt type
- `pkg/rpc/subscription/manager_test.go` - Added FrameTx receipt formatting tests

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

### Phase 1: Receipt Integration (COMPLETE ✅)

1. **Add conversion method to FrameTxReceipt**
   ```go
   // pkg/core/types/frame_receipt.go
   func (r *FrameTxReceipt) ToReceipt(txHash, blockHash Hash, blockNumber *big.Int, txIndex uint, effectiveGasPrice *big.Int) *Receipt
   func (r *FrameTxReceipt) ToReceiptWithPayer(txHash, blockHash Hash, blockNumber *big.Int, txIndex uint, effectiveGasPrice *big.Int) *Receipt
   func (r *FrameTxReceipt) DeriveStatus() uint64
   func (r *FrameTxReceipt) ComputeBloom() Bloom
   ```

2. **Update receipt generation**
   ```go
   // pkg/core/execution/receipt_generation.go
   type TxExecutionOutcome struct {
       // ... existing fields
       FrameReceipt *types.FrameTxReceipt // For EIP-8141
   }
   ```

3. **Add tests for conversion**
   - `pkg/core/types/frame_receipt_rlp_test.go` - Added tests for ToReceipt, ToReceiptWithPayer, DeriveStatus, ComputeBloom

### Phase 2: RPC Enhancement (COMPLETE ✅)

1. **Extended RPCReceipt type**
   ```go
   // pkg/rpc/types/types.go
   type RPCReceipt struct {
       // ... existing fields
       Payer        *string           `json:"payer,omitempty"`
       FrameResults *[]RPCFrameResult `json:"frameResults,omitempty"`
   }
   
   type RPCFrameResult struct {
       Status  string    `json:"status"`
       GasUsed string    `json:"gasUsed"`
       Logs    []*RPCLog `json:"logs,omitempty"`
   }
   ```

2. **Updated FormatReceipt**
   - FrameTx type detection (type == 0x06)
   - Populate Payer field for FrameTx receipts

3. **Added tests for RPC responses**
   - `pkg/rpc/subscription/manager_test.go` - Added TestFormatReceipt_FrameTx, TestFormatReceipt_FrameTx_NoPayer, TestFormatReceipt_NonFrameTx_NoPayer

### Phase 3: Devnet Testing (PENDING)

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
- [x] FrameTxReceipt stored correctly
- [x] RPC returns payer address
- [x] Bloom filter computed correctly
- [x] Logs aggregated from all frames

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