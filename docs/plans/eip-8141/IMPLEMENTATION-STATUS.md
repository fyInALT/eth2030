# EIP-8141 Implementation Status Assessment

**Date:** 2026-02-27
**Assessed against:** `refs/EIPs/EIPS/eip-8141.md`

## Codebase Locations

| File | Contents | Lines |
|------|----------|-------|
| `pkg/core/types/tx_frame.go` | `FrameTx`, `Frame` structs, RLP encode/decode, `CalcFrameTxGas`, `ComputeFrameSigHash`, `ValidateFrameTx` | 290 |
| `pkg/core/types/tx_frame_test.go` | Unit tests for tx_frame.go | — |
| `pkg/core/types/frame_receipt.go` | `FrameResult`, `FrameTxReceipt` structs | 35 |
| `pkg/core/types/frame_receipt_extended.go` | Extended receipt functionality | 319 |
| `pkg/core/types/frame_receipt_extended_test.go` | Tests for extended receipts | — |
| `pkg/core/types/transaction.go` | Transaction envelope dispatch | 580 |
| `pkg/core/vm/eip8141_opcodes.go` | `opApprove`, `opTxParamLoad`, `opTxParamSize`, `opTxParamCopy`, `FrameContext`, `txParamValue`, `txParamSize` | 383 |
| `pkg/core/vm/eip8141_opcodes_test.go` | Tests for EIP-8141 opcodes | — |
| `pkg/core/vm/opcodes.go` | Opcode constants: `APPROVE=0xaa`, `TXPARAMLOAD=0xb0`, `TXPARAMSIZE=0xb1`, `TXPARAMCOPY=0xb2` | 245 |
| `pkg/core/vm/aa_executor.go` | EIP-7701 AA executor (NOT EIP-8141 frame executor — needs new implementation) | 641 |
| `pkg/txpool/validation_pipeline.go` | Transaction pool validation | — |

## Per-Story Implementation Status

### US-1.1 — RLP Serialization ✅ Mostly Implemented

**EIP Reference (Specification → New Transaction Type):**
> `[chain_id, nonce, sender, frames, max_priority_fee_per_gas, max_fee_per_gas, max_fee_per_blob_gas, blob_versioned_hashes]`
> `frames = [[mode, target, gas_limit, data], ...]`

**Status:**
- ✅ `FrameTx` struct defined with all fields
- ✅ `EncodeFrameTx` / `DecodeFrameTx` implemented with correct RLP layout
- ✅ Null target encodes as empty bytes, decodes to nil
- ✅ Empty blob_versioned_hashes encoded as empty list (not nil)
- ⚠️ **Gap:** `CalcFrameTxGas` does not check for overflow (returns silently truncated uint64)
- ⚠️ **Gap:** Not yet integrated into `Transaction` wrapper's `TypedTxData` dispatch (Task 1.1.2)

**Relevant code:** `pkg/core/types/tx_frame.go:107-170` (encode/decode), `pkg/core/types/tx_frame.go:251-268` (gas calc)

### US-1.2 — Fee Calculation ⚠️ Partially Implemented

**EIP Reference (Specification → Gas Accounting):**
> `tx_fee = tx_gas_limit * effective_gas_price + blob_fees`
> `blob_fees = len(blob_versioned_hashes) * GAS_PER_BLOB * blob_base_fee`

**Status:**
- ✅ `gasFeeCap()` and `gasTipCap()` return correct fields
- ❌ **Missing:** `EffectiveGasTip(baseFee)` helper not implemented
- ❌ **Missing:** `TotalFee(baseFee, blobBaseFee)` helper not implemented
- ❌ **Missing:** `MaxCost` computation for TXPARAM 0x06 (exists in FrameContext but not calculated)

### US-2.1 — Frame Modes ⚠️ Partially Implemented

**EIP Reference (Specification → Modes):**
> DEFAULT: Execute frame as regular call where the caller address is `ENTRY_POINT`.
> VERIFY: Identifies the frame as a validation frame... must call `APPROVE`... behaves the same as `STATICCALL`.
> SENDER: Execute frame as regular call where the caller address is `sender`. This mode effectively acts on behalf of the transaction sender and can only be used after explicitly approved.

**Status:**
- ✅ Mode constants defined (`ModeDefault=0`, `ModeVerify=1`, `ModeSender=2`)
- ✅ `EntryPointAddress` defined as `0x00...00aa`
- ❌ **Missing:** Frame execution dispatcher — `aa_executor.go` handles EIP-7701 (AA type 0x04), NOT EIP-8141 frame transactions. A new `FrameExecutor` is needed.
- ❌ **Missing:** VERIFY mode STATICCALL enforcement (readOnly flag)
- ❌ **Missing:** SENDER mode sender_approved guard
- ❌ **Missing:** Null target → tx.sender substitution in execution logic

**Refactoring needed:** `aa_executor.go` is for EIP-7701. Need a separate `frame_executor.go` that implements the EIP-8141 frame dispatch loop.

### US-3.1 — APPROVE Opcode ✅ Mostly Implemented

**EIP Reference (Specification → APPROVE opcode):**
> `APPROVE` is like `RETURN (0xf3)`. It exits the current context successfully and updates the transaction-scoped approval context based on the `scope` operand. It can only be called when `CALLER == frame.target`.

**Status:**
- ✅ `opApprove` implemented with all 3 scopes (0x0, 0x1, 0x2)
- ✅ `CALLER == frame.target` check (via `contract.CallerAddress != contract.Address`)
- ✅ Scope 0x0: `SenderApproved` + `CALLER == tx.sender` check
- ✅ Scope 0x1: `SenderApproved` prerequisite + balance check + `PayerApproved`
- ✅ Scope 0x2: Combined check + both flags
- ✅ Double-approval prevention (monotonic flags)
- ✅ Returns memory[offset:offset+length] like RETURN
- ⚠️ **Gap:** Scope 0x1 does NOT increment nonce or deduct fee from balance — only checks balance
- ⚠️ **Gap:** Scope 0x2 does NOT increment nonce or deduct fee
- ⚠️ **Gap:** Error semantics: code returns `error` (exceptional halt) for caller mismatch, but EIP says "revert" (remaining gas returned)
- ⚠️ **Gap:** `contract.CallerAddress != contract.Address` check is NOT equivalent to `CALLER == frame.target` — it checks caller==code address, which may differ if the contract was called via DELEGATECALL

### US-4.1 — TXPARAM* Opcodes ✅ Mostly Implemented

**EIP Reference (Specification → TXPARAM* opcodes):**
> The `TXPARAMLOAD` (`0xb0`), `TXPARAMSIZE` (`0xb1`), and `TXPARAMCOPY` (`0xb2`) opcodes follow the pattern of `CALLDATA*` / `RETURNDATA*` opcode families.

**Status:**
- ✅ All 16 parameter indices implemented in `txParamValue`
- ✅ `in2 == 0` enforcement for scalar parameters
- ✅ Frame index bounds checking
- ✅ VERIFY frame data returns nil/empty
- ✅ Status (0x15) blocks current/future frame access
- ✅ `opTxParamLoad`, `opTxParamSize`, `opTxParamCopy` all implemented
- ⚠️ **Gap:** `opTxParamLoad` has 2 stack inputs (in1, in2) but EIP specifies 3 (in1, in2, byte_offset) — missing byte_offset support
- ⚠️ **Gap:** Gap indices 0x0a–0x0f not explicitly handled (falls through to default → halt, which is correct)
- ⚠️ **Gap:** Gas costs not registered in instruction table

### US-5.1 — Execution Engine ❌ Not Implemented

**EIP Reference (Specification → Behavior):**
> When processing a frame transaction, perform the following steps... Then for each call frame: Execute a call with the specified `mode`, `target`, `gas_limit`, and `data`.

**Status:**
- ❌ **Missing:** No `FrameExecutor` or frame dispatch loop exists
- ❌ **Missing:** Nonce validation before frame execution
- ❌ **Missing:** `sender_approved`/`payer_approved` state machine
- ❌ **Missing:** Post-loop `payer_approved` check
- ❌ **Missing:** Gas refund to payer
- ❌ **Missing:** State atomicity (snapshot/revert on invalid tx)
- The existing `aa_executor.go` handles EIP-7701, not EIP-8141

**Action needed:** Create `pkg/core/vm/frame_executor.go` implementing the complete frame dispatch loop.

### US-6.1 — Gas Accounting ⚠️ Partially Implemented

**Status:**
- ✅ `CalcFrameTxGas` implemented (intrinsic + calldata + sum of frame limits)
- ✅ `FrameTxIntrinsicCost = 15000` defined
- ❌ **Missing:** Per-frame gas isolation in execution (no executor yet)
- ❌ **Missing:** Overflow detection in CalcFrameTxGas

### US-7.1 — Receipt Structure ✅ Mostly Implemented

**EIP Reference (Specification → Receipt):**
> `[cumulative_gas_used, payer, [frame_receipt, ...]]`
> `frame_receipt = [status, gas_used, logs]`

**Status:**
- ✅ `FrameResult` struct: `{Status, GasUsed, Logs}`
- ✅ `FrameTxReceipt` struct: `{CumulativeGasUsed, Payer, FrameResults}`
- ✅ `TotalGasUsed()` and `AllLogs()` helpers
- ⚠️ **Gap:** RLP encode/decode for `FrameTxReceipt` not visible (may be in extended file)
- ⚠️ **Gap:** JSON-RPC serialization for `eth_getTransactionReceipt` not implemented

### US-8.1 — Signature Hash ✅ Implemented

**EIP Reference (Specification → Signature Hash):**
> ```python
> def compute_sig_hash(tx: FrameTx) -> Hash:
>     for i, frame in enumerate(tx.frames):
>         if frame.mode == VERIFY:
>             tx.frames[i].data = Bytes()
>     return keccak(rlp(tx))
> ```

**Status:**
- ✅ `ComputeFrameSigHash` implemented correctly
- ✅ VERIFY frame data elided (set to empty bytes)
- ✅ Uses keccak256(0x06 || rlp(modified_tx))
- ✅ Does not mutate original transaction (builds copy)

### US-9.1 — Static Validation ✅ Mostly Implemented

**Status:**
- ✅ `ValidateFrameTx` checks: frame count (0 and >MAX_FRAMES), chain_id sign, mode < 3, target length, blob field consistency
- ⚠️ **Gap:** Does not check `nonce < 2^64` (Go uint64 naturally caps this)
- ⚠️ **Gap:** Does not check `len(sender) == 20` (Address type is always 20 bytes)

### US-10.1 — Frame Interactions ❌ Not Implemented

**Status:**
- ❌ No frame executor → no warm/cold journal sharing
- ❌ No transient storage clearing between frames

### US-11.1 — ORIGIN Opcode ❌ Not Implemented

**Status:**
- ❌ `opOrigin` not modified for frame transactions
- ❌ No `FrameContext` check in ORIGIN handler

### US-12.1 — Mempool Validation ❌ Not Implemented

**Status:**
- ❌ No VERIFY-only mempool simulation
- ❌ No opcode restrictions for validation frames
- ❌ No deployer factory allowlist
- ❌ No single-pending-tx-per-sender for frame txs

### US-13.1 — Integration Tests ❌ Not Implemented

**Status:**
- ❌ No E2E integration tests for frame transactions

---

## Summary

| Category | Status | Count |
|----------|--------|-------|
| ✅ Implemented | Sig hash, static validation (mostly), opcodes defined | 3 |
| ⚠️ Partially implemented | RLP, APPROVE, TXPARAM*, receipt, gas calc, fee calc | 6 |
| ❌ Not implemented | Frame executor, ORIGIN, mempool, frame interactions, integration tests | 5 |

## Key Architectural Gap

**The biggest missing piece is the Frame Executor (`frame_executor.go`)** — the frame dispatch loop that:
1. Validates nonce
2. Iterates frames with correct caller/mode
3. Enforces sender_approved/payer_approved state machine
4. Handles VERIFY frame APPROVE detection
5. Per-frame gas isolation
6. Post-loop payer check + refund
7. State atomicity via snapshot/revert

The existing `aa_executor.go` handles EIP-7701 (type 0x04) and cannot be reused for EIP-8141 (type 0x06). A new executor is needed.

## Refactoring Recommendations

1. **Create `pkg/core/vm/frame_executor.go`** — new file for EIP-8141 frame transaction execution
2. **Fix APPROVE error semantics** — caller mismatch should revert (return gas), not exceptional halt
3. **Add nonce increment + fee deduction** to APPROVE scope 0x1 and 0x2
4. **Fix APPROVE caller check** — use `frame.target` from FrameContext, not `contract.Address`
5. **Add byte_offset support** to TXPARAMLOAD (3rd stack input)
6. **Register TXPARAM gas costs** in the instruction table
7. **Integrate FrameTx** into Transaction envelope dispatch (type switch)
8. **Add ORIGIN override** in `opOrigin` checking for FrameContext
