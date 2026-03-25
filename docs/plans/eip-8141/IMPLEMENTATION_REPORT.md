# EIP-8141 Frame Transaction Implementation Report

## Overview

EIP-8141 introduces a new transaction type that enables native account abstraction through frame-based execution. This document provides a comprehensive analysis of the current implementation status in eth2030.

**EIP Status**: Draft  
**Transaction Type**: `0x06` (`FRAME_TX_TYPE`)  
**Entry Point**: `0x00000000000000000000000000000000000000aa`

---

## 1. Specification Summary

### 1.1 Constants

| Name | Value | Description |
|------|-------|-------------|
| `FRAME_TX_TYPE` | `0x06` | Transaction type identifier |
| `FRAME_TX_INTRINSIC_COST` | `15000` | Base gas cost for frame transactions |
| `ENTRY_POINT` | `0xaa` | Canonical caller address for frames |
| `MAX_FRAMES` | `1000` | Maximum number of frames per transaction |

### 1.2 Transaction Structure

```
[chain_id, nonce, sender, frames, max_priority_fee_per_gas, max_fee_per_gas, 
 max_fee_per_blob_gas, blob_versioned_hashes]

frames = [[mode, target, gas_limit, data], ...]
```

### 1.3 Frame Modes

| Mode | Name | Caller | Description |
|------|------|--------|-------------|
| 0 | DEFAULT | `ENTRY_POINT` | Regular call from entry point |
| 1 | VERIFY | `ENTRY_POINT` | Validation frame (StaticCall, must APPROVE) |
| 2 | SENDER | `tx.sender` | Execution on behalf of sender (requires approval) |

### 1.4 Opcodes

| Opcode | Value | Description |
|--------|-------|-------------|
| `APPROVE` | `0xaa` | Approve execution/payment, terminates frame |
| `TXPARAMLOAD` | `0xb0` | Load transaction parameter to stack |
| `TXPARAMSIZE` | `0xb1` | Get size of transaction parameter |
| `TXPARAMCOPY` | `0xb2` | Copy transaction parameter to memory |

---

## 2. Implementation Status

### 2.1 Transaction Type Definition

**Status**: ✅ FULLY IMPLEMENTED

**File**: `pkg/core/types/tx_frame.go`

```go
type FrameTx struct {
    ChainID              *big.Int
    Nonce                *big.Int    // 256-bit: upper 192 bits = key, lower 64 bits = sequence
    Sender               Address
    Frames               []Frame
    MaxPriorityFeePerGas *big.Int
    MaxFeePerGas         *big.Int
    MaxFeePerBlobGas     *big.Int
    BlobVersionedHashes  []Hash
}

type Frame struct {
    Mode     uint8
    Target   *Address  // nil defaults to sender
    GasLimit uint64
    Data     []byte
}
```

**Features**:
- [x] All fields per EIP-8141 spec
- [x] 2D nonce support (`NonceKey()` and `NonceSeq()` methods)
- [x] TxData interface implementation

### 2.2 RLP Encoding/Decoding

**Status**: ✅ FULLY IMPLEMENTED

**File**: `pkg/core/types/tx_frame.go`

| Function | Status | Description |
|----------|--------|-------------|
| `EncodeFrameTx()` | ✅ | Encodes to `0x06 \|\| RLP([...])` |
| `DecodeFrameTx()` | ✅ | Decodes RLP payload to FrameTx |
| `decodeFrameTxWrapped()` | ✅ | Integration with `DecodeTxRLP()` |

**RPC Integration**:
- `eth_sendRawTransaction` supports FrameTx via `DecodeTxRLP()`
- Automatic type detection from first byte

### 2.3 Signature Hash Computation

**Status**: ✅ FULLY IMPLEMENTED

**File**: `pkg/core/types/tx_frame.go:221-260`

```go
func ComputeFrameSigHash(tx *FrameTx) Hash {
    // VERIFY frames have their data elided (set to empty) before hashing
    for i, f := range tx.Frames {
        if f.Mode == ModeVerify {
            frames[i].Data = []byte{}  // Elide signature data
        } else {
            frames[i].Data = f.Data
        }
    }
    return keccak256(0x06 || rlp(tx))
}
```

**Features**:
- [x] VERIFY frame data elision for signature malleability protection
- [x] Accessible via `TXPARAMLOAD` index `0x08`

### 2.4 APPROVE Opcode (0xaa)

**Status**: ✅ FULLY IMPLEMENTED

**File**: `pkg/core/vm/eip8141_opcodes.go:69-149`

| Scope | Function | Status |
|-------|----------|--------|
| 0 | Execution approval | ✅ Sets `SenderApproved = true` |
| 1 | Payment approval | ✅ Sets `PayerApproved = true` (requires sender approval first) |
| 2 | Combined approval | ✅ Sets both flags |

**Validation**:
- [x] `CALLER == frame.target` check
- [x] Balance check for payment
- [x] Prevents double approval
- [x] Returns memory data like `RETURN`

### 2.5 TXPARAM* Opcodes

**Status**: ✅ FULLY IMPLEMENTED

**File**: `pkg/core/vm/eip8141_opcodes.go:151-395`

| Index | Parameter | Status | Notes |
|-------|-----------|--------|-------|
| 0x00 | tx_type | ✅ | Returns `0x06` |
| 0x01 | nonce | ✅ | 256-bit nonce |
| 0x02 | sender | ✅ | 20-byte address |
| 0x03 | max_priority_fee_per_gas | ✅ | |
| 0x04 | max_fee_per_gas | ✅ | |
| 0x05 | max_fee_per_blob_gas | ✅ | |
| 0x06 | max_cost | ✅ | Calculated dynamically |
| 0x07 | len(blob_versioned_hashes) | ✅ | |
| 0x08 | sig_hash | ✅ | Canonical signature hash |
| 0x09 | len(frames) | ✅ | |
| 0x10 | current_frame_index | ✅ | |
| 0x11 | frame_target[idx] | ✅ | |
| 0x12 | frame_data[idx] | ✅ | Elided for VERIFY frames |
| 0x13 | frame_gas_limit[idx] | ✅ | |
| 0x14 | frame_mode[idx] | ✅ | |
| 0x15 | frame_status[idx] | ✅ | Exceptional halt for current/future |

### 2.6 ORIGIN Opcode Modification

**Status**: ✅ FULLY IMPLEMENTED

**File**: `pkg/core/vm/instructions.go:358-366`

```go
func opOrigin(...) {
    // EIP-8141: inside a frame transaction, ORIGIN returns the frame caller
    if evm.FrameCtx != nil {
        stack.Push(new(big.Int).SetBytes(evm.FrameCtx.Sender[:]))
        return nil, nil
    }
    stack.Push(new(big.Int).SetBytes(evm.TxContext.Origin[:]))
}
```

### 2.7 Frame Execution

**Status**: ✅ FULLY IMPLEMENTED

**File**: `pkg/core/eips/frame_execution.go`

```go
func ExecuteFrameTx(evm *EVM, tx *types.FrameTx) (*types.FrameTxReceipt, error) {
    // 1. Validate nonce
    // 2. Initialize frame context (SenderApproved, PayerApproved = false)
    // 3. For each frame:
    //    - DEFAULT: Call(ENTRY_POINT, target)
    //    - VERIFY: StaticCall(ENTRY_POINT, target) + APPROVE check
    //    - SENDER: Call(sender, target) [requires SenderApproved]
    // 4. Verify PayerApproved == true
    // 5. Settle gas payment
}
```

**Features**:
- [x] Frame isolation (transient storage cleared between frames)
- [x] VERIFY mode uses StaticCall (no state modification)
- [x] SENDER mode requires prior approval
- [x] Proper error handling for missing approvals

### 2.8 State Transition

**Status**: ✅ FULLY IMPLEMENTED

**File**: `pkg/core/execution/processor.go:1437-1580`

**Features**:
- [x] Frame context setup in EVM before execution
- [x] Gas accounting across frames
- [x] Sponsored gas settlement (payer charged, not sender)
- [x] Nonce increment after successful approval
- [x] Balance deduction from payer

### 2.9 Transaction Pool Validation

**Status**: ✅ FULLY IMPLEMENTED

**File**: `pkg/txpool/txpool.go:554-625`

**Validation Rules**:
- [x] Minimum gas check (`>= FRAME_TX_INTRINSIC_COST`)
- [x] Frame count (`1 <= len(frames) <= MAX_FRAMES`)
- [x] Mode validity (0-2)
- [x] SENDER frames require VERIFY frame
- [x] VERIFY frame code check via simulation

**File**: `pkg/txpool/frametx/frame_rules.go`

| Rule Set | First Frame | Gas Limit |
|----------|-------------|-----------|
| Conservative | Must be VERIFY | ≤ 50K |
| Aggressive | Any | ≤ 200K (staked paymasters) |

### 2.10 Receipt Generation

**Status**: ✅ FULLY IMPLEMENTED

**File**: `pkg/core/types/frame_receipt.go`

```go
type FrameTxReceipt struct {
    CumulativeGasUsed uint64
    Payer             Address      // Who paid for the transaction
    FrameResults      []FrameResult
}

type FrameResult struct {
    Status  uint64
    GasUsed uint64
    Logs    []*Log
}
```

**Features**:
- [x] RLP encoding/decoding with `0x06` type prefix
- [x] Per-frame status and gas tracking
- [x] Payer address in receipt (not determinable statically)

---

## 3. Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        FrameTx (0x06) Transaction Flow                       │
└─────────────────────────────────────────────────────────────────────────────┘

RPC Layer (eth_sendRawTransaction)
    │
    ▼
┌───────────────────┐
│ DecodeTxRLP()     │  ─── Decode 0x06 prefix ───▶ decodeFrameTxWrapped()
│ pkg/rpc/ethapi    │
└───────────────────┘
    │
    ▼
┌───────────────────┐
│ TxPool.Validate() │  ─── ValidateFrameTx()
│ pkg/txpool        │       - Frame count, modes, gas limits
└───────────────────┘       - SimulateVerifyFrame() for code check
    │
    ▼
┌───────────────────┐
│ Block Building    │  ─── processor.ExecuteFrameTx()
│ pkg/core          │       - Nonce validation
└───────────────────┘       - Frame context initialization
    │
    ▼
┌───────────────────────────────────────────────────────────────┐
│ Frame Execution Loop                                          │
│                                                               │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐       │
│  │ DEFAULT     │    │ VERIFY      │    │ SENDER      │       │
│  │ mode=0      │    │ mode=1      │    │ mode=2      │       │
│  │             │    │             │    │             │       │
│  │ caller:     │    │ caller:     │    │ caller:     │       │
│  │ ENTRY_POINT │    │ ENTRY_POINT │    │ tx.sender   │       │
│  │             │    │             │    │             │       │
│  │ Call()      │    │ StaticCall()│    │ Call()      │       │
│  │             │    │             │    │             │       │
│  │             │    │ APPROVE     │    │ (requires   │       │
│  │             │    │ required    │    │ approval)   │       │
│  └─────────────┘    └─────────────┘    └─────────────┘       │
│                                                               │
│  FrameContext tracks: SenderApproved, PayerApproved           │
│  TSTORE/TLOAD cleared between frames                          │
└───────────────────────────────────────────────────────────────┘
    │
    ▼
┌───────────────────┐
│ Receipt Building  │  ─── FrameTxReceipt
│                   │       - Payer address
│                   │       - Per-frame results
└───────────────────┘
```

---

## 4. File Reference

| Component | File Path |
|-----------|-----------|
| FrameTx type | `pkg/core/types/tx_frame.go` |
| Frame receipt | `pkg/core/types/frame_receipt.go` |
| Transaction RLP decode | `pkg/core/types/transaction_rlp.go:405` |
| APPROVE/TXPARAM opcodes | `pkg/core/vm/eip8141_opcodes.go` |
| ORIGIN opcode | `pkg/core/vm/instructions.go:358` |
| Jump table setup | `pkg/core/vm/jump_table.go:695` |
| Frame execution | `pkg/core/eips/frame_execution.go` |
| State transition | `pkg/core/execution/processor.go:1437` |
| TxPool validation | `pkg/txpool/txpool.go:554` |
| Frame rules | `pkg/txpool/frametx/frame_rules.go` |
| Verify simulation | `pkg/txpool/frametx/verify_simulation.go` |
| RPC integration | `pkg/rpc/ethapi/eth_api.go:1259` |

---

## 5. Test Coverage

| Test File | Coverage |
|-----------|----------|
| `pkg/core/types/tx_frame_test.go` | RLP roundtrip, gas calculation, signing |
| `pkg/core/types/frame_receipt_test.go` | Receipt encoding/decoding |
| `pkg/core/vm/eip8141_opcodes_test.go` | All opcodes, all scopes |
| `pkg/core/vm/eip8141_spec_test.go` | TXPARAM indices |
| `pkg/core/eips/frame_execution_test.go` | Frame execution modes |
| `pkg/txpool/frametx/frame_rules_test.go` | Pool validation rules |

---

## 6. Implementation Gaps

**None identified.** All EIP-8141 requirements are implemented.

### 6.1 Additional Features (Beyond Spec)

1. **Paymaster Registry** (`pkg/txpool/frametx/paymaster_registry.go`)
   - Security measure for mempool protection
   - Staked paymasters get higher gas limits

2. **STARK Proof Support** (`pkg/core/vm/frame_stark_replacer.go`)
   - Future optimization for proof aggregation
   - Not required by base EIP

3. **Conservative/Aggressive Rules** (`pkg/txpool/frametx/frame_rules.go`)
   - Configurable mempool validation strictness

---

## 7. Conclusion

**EIP-8141 Frame Transaction is FULLY IMPLEMENTED in eth2030.**

All specification requirements are satisfied:
- Transaction type with RLP encoding/decoding
- Four new opcodes (APPROVE, TXPARAMLOAD, TXPARAMSIZE, TXPARAMCOPY)
- Three frame modes (DEFAULT, VERIFY, SENDER)
- Signature hash with VERIFY data elision
- ORIGIN opcode modification
- Transaction pool validation
- State transition logic
- Receipt generation

The implementation is production-ready for devnet testing.