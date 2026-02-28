# EIP-8141 Frame Transaction — Sprint Planning Document

> **Branch:** `plan/eip-8141`
> **Created:** 2026-02-27
> **EIP Reference:** [EIP-8141 Frame Transaction](https://ethereum-magicians.org/t/frame-transaction/27617)
> **Requires:** EIP-2718 (typed envelopes), EIP-4844 (blob transactions)

---

## Overview

EIP-8141 introduces a new transaction type (`FRAME_TX_TYPE = 0x06`) that replaces ECDSA-based authentication with arbitrary EVM-defined validation logic. It is the primary on-ramp from ECDSA to post-quantum cryptographic systems and realizes full account abstraction by composing a transaction as a sequential list of **frames**, each with a distinct execution mode, caller identity, gas budget, and data payload.

This document decomposes the EIP into Scrum-ready user stories and actionable engineering tasks. The project is a Go-based Ethereum execution client (`eth2030`). Existing partial implementations exist in `pkg/core/types/tx_frame.go`, `pkg/core/vm/eip8141_opcodes.go`, and `pkg/core/types/frame_receipt.go`; all tasks must integrate with and extend those files.

---

## INVEST Compliance Legend

| Symbol | Criterion |
|--------|-----------|
| I | Independent — can be developed without waiting for another story |
| N | Negotiable — scope/approach can be discussed |
| V | Valuable — delivers observable value to users or the system |
| E | Estimable — team can size it with reasonable confidence |
| S | Small — completable within one sprint |
| T | Testable — has verifiable acceptance criteria |

---

## Epics

| # | Epic | Key Areas Covered |
|---|------|-------------------|
| EP-1 | Transaction Type & RLP Encoding | #1 |
| EP-2 | Frame Structure & Mode Semantics | #2, #3 |
| EP-3 | APPROVE Opcode | #4 |
| EP-4 | TXPARAM* Opcodes | #5 |
| EP-5 | Transaction Execution Engine | #6, #11, #13, #14 |
| EP-6 | Gas Accounting | #7 |
| EP-7 | Receipt Structure | #8 |
| EP-8 | Signature Hash Computation | #9 |
| EP-9 | Static Validity Constraints | #10 |
| EP-10 | Frame Interactions & Cross-Frame State | #11 |
| EP-11 | ORIGIN Opcode Behavior Change | #12 |
| EP-12 | Mempool Validation & DoS Mitigation | #15 |

---

## User Stories

---

### EP-1: Transaction Type & RLP Encoding

---

#### US-1.1 — Frame Transaction RLP Serialization

> **As a** protocol engineer,
> **I want** the `FrameTx` struct to serialize to and deserialize from canonical RLP as defined in EIP-8141,
> **so that** frame transactions can be included in blocks, signed, and transmitted over the wire.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

##### Task 1.1.1 — Finalize `FrameTx` RLP Encode/Decode

| Field | Detail |
|-------|--------|
| **Description** | Implement `EncodeRLP` and `DecodeRLP` for `FrameTx` following the canonical layout: `[chain_id, nonce, sender, frames, max_priority_fee_per_gas, max_fee_per_gas, max_fee_per_blob_gas, blob_versioned_hashes]`. Each frame encodes as `[mode, target, gas_limit, data]`. Null target must encode as empty bytes (`0x80`) and decode back to `nil`. Ensure `blob_versioned_hashes` is an empty list (not nil RLP) and `max_fee_per_blob_gas` is `0` when no blobs are present. Integrate with the existing `FrameTx` in `pkg/core/types/tx_frame.go`. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | Unit tests in `pkg/core/types/tx_frame_test.go`: (1) round-trip encode→decode for transactions with 0, 1, and `MAX_FRAMES` frames; (2) assert null target encodes as `0x80` and decodes to `nil`; (3) assert blob fields zero when empty; (4) fuzz test with random byte inputs to `DecodeRLP`. |
| **Definition of Done** | All unit tests pass; `go fmt ./...` clean; `go vet ./...` clean; no regression in existing `tx_frame_test.go`; code reviewed and merged; coverage ≥ 80% on new encode/decode paths. |

##### Task 1.1.2 — EIP-2718 Transaction Envelope Integration

| Field | Detail |
|-------|--------|
| **Description** | Register `FrameTxType = 0x06` in the EIP-2718 typed transaction envelope dispatcher in `pkg/core/types/transaction.go` (or equivalent dispatch file). Ensure `TypedTxData` switch statements handle `0x06` for `Hash()`, `SigningHash()`, `RawSigningHash()`, `Cost()`, `Gas()`. Confirm that `Transaction.Type()` returns `0x06` for `FrameTx` instances. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | Unit test: create a `Transaction` wrapping a `FrameTx`, call `tx.Type()` and assert `0x06`; call `tx.Hash()` and assert deterministic output; serialize to bytes, deserialize, compare hash. |
| **Definition of Done** | `tx.Type() == 0x06` passes; round-trip hash equality holds; no panic in switch fallthrough; code reviewed; no regressions in existing typed-tx tests. |

##### Task 1.1.3 — `CalcFrameTxGas` Total Gas Computation

| Field | Detail |
|-------|--------|
| **Description** | Implement or complete `CalcFrameTxGas(tx *FrameTx) uint64` in `pkg/core/types/tx_frame.go`. The formula is: `FRAME_TX_INTRINSIC_COST (15000) + calldata_cost(rlp(tx.frames)) + sum(frame.gas_limit for all frames)`. `calldata_cost` uses standard EVM rules: 4 gas per zero byte, 16 gas per non-zero byte, applied to the RLP encoding of the frames list only (not the full transaction). Overflow must be detected and return `math.MaxUint64` or panic with a sentinel error. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | Unit tests: (1) single frame with all-zero data yields `15000 + calldata_bytes*4 + gas_limit`; (2) frame with mixed zero/non-zero data; (3) `MAX_FRAMES` frames with large gas limits to test overflow guard; (4) compare output with hand-computed values from the EIP data-efficiency tables. |
| **Definition of Done** | All tests pass; overflow handled without panic; `go vet` clean; reviewed. |

---

#### US-1.2 — Frame Transaction Fee Calculation

> **As a** block builder,
> **I want** `FrameTx` total fee computation (EIP-1559 + EIP-4844 blob fees) to be correct,
> **so that** the block builder can price transactions and refund the payer accurately.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

##### Task 1.2.1 — `EffectiveGasPrice` and `TotalFee` for FrameTx

| Field | Detail |
|-------|--------|
| **Description** | Implement fee calculation helpers for `FrameTx`: `EffectiveGasTip(baseFee *big.Int) *big.Int` using EIP-1559 capping logic (`min(max_priority_fee_per_gas, max_fee_per_gas - base_fee)`); `TotalFee(baseFee, blobBaseFee *big.Int) *big.Int` = `tx_gas_limit * effective_gas_price + len(blob_versioned_hashes) * GAS_PER_BLOB * blob_base_fee`. Integrate with the existing fee interfaces in `pkg/core/types/`. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | Table-driven unit tests with known baseFee, blob fee, and gas limit values; assert against hand-calculated expected fees. Test blob fee when `blob_versioned_hashes` is empty (must be zero). |
| **Definition of Done** | Tests pass; EIP-1559 capping verified; blob fee zero when no blobs; code reviewed. |

##### Task 1.2.2 — `MaxCost` TXPARAM Parameter (0x06)

| Field | Detail |
|-------|--------|
| **Description** | Implement the `max_cost` computation exposed at TXPARAM index `0x06`. The full formula (worst-case cost, `basefee = max_fee_per_gas`, all gas used, blobs at max price) is: `max_cost = tx_gas_limit * max_fee_per_gas + len(blob_versioned_hashes) * GAS_PER_BLOB * max_fee_per_blob_gas`. This includes the intrinsic cost (baked into `tx_gas_limit`) and all blob costs. Used by `opTxParamLoad` in `pkg/core/vm/eip8141_opcodes.go`. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Unit test with no blobs: assert `max_cost = tx_gas_limit * max_fee_per_gas`. (2) Unit test with 2 blobs at known `max_fee_per_blob_gas`: assert blob component is `2 * GAS_PER_BLOB * max_fee_per_blob_gas`. (3) Overflow guard: very large gas limit and fee do not produce silent truncation. |
| **Definition of Done** | All 3 tests pass; blob component zero when no blobs; implementation matches EIP formula including intrinsic cost in gas limit; reviewed. |

---

### EP-2: Frame Structure & Mode Semantics

---

#### US-2.1 — Frame Mode Definitions and Caller Identity

> **As a** protocol engineer,
> **I want** each frame mode (DEFAULT=0, VERIFY=1, SENDER=2) to set the correct caller and execution constraints,
> **so that** frames execute with the correct identity and state-modification permissions.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

##### Task 2.1.1 — DEFAULT Mode: Caller = ENTRY_POINT

| Field | Detail |
|-------|--------|
| **Description** | In the frame execution engine (`pkg/core/frame_execution.go` + `pkg/core/processor.go`), when a frame has `mode == DEFAULT`, set the `caller` of the EVM call to `ENTRY_POINT` (`address(0xaa)`). Confirm `EntryPointAddress` is defined in `pkg/core/types/tx_frame.go` as `HexToAddress("0x00000000000000000000000000000000000000aa")`. No state-modification restrictions apply. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | Integration test: deploy a contract that records `msg.sender`; execute a DEFAULT frame targeting it; assert `msg.sender == ENTRY_POINT`. |
| **Definition of Done** | Test passes; `EntryPointAddress` constant defined; reviewed. |

##### Task 2.1.2 — VERIFY Mode: STATICCALL Semantics + APPROVE Requirement

| Field | Detail |
|-------|--------|
| **Description** | When a frame has `mode == VERIFY`, the EVM call must behave as a `STATICCALL` — all standard state-modifying opcodes (`SSTORE`, `LOG*`, `CREATE*`, `SELFDESTRUCT`, `CALL` with non-zero value) must result in exceptional halt; the `readOnly` flag must be `true` and propagated through all sub-calls. The frame's caller must be `ENTRY_POINT`. **Critical exception:** the `APPROVE` opcode (`0xaa`) is explicitly permitted in VERIFY frames despite STATICCALL semantics. `APPROVE` modifies **transaction-scoped** state (`sender_approved`, `payer_approved`, nonce, balance), not EVM account/storage state; therefore `opApprove` must bypass the `readOnly` check and is the only state-changing action allowed. After frame completion, detect whether `APPROVE` was called by inspecting whether `sender_approved` or `payer_approved` changed during the frame; if neither changed (APPROVE was never called successfully), mark the entire transaction invalid. Integrate with `pkg/core/frame_execution.go`. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | (1) VERIFY frame that calls `SSTORE` → exceptional halt. (2) VERIFY frame that calls `LOG0` → exceptional halt. (3) VERIFY frame that calls `APPROVE(0x2)` → succeeds despite STATICCALL; both `sender_approved` and `payer_approved` become true. (4) VERIFY frame that calls `APPROVE` followed by `SSTORE` — APPROVE succeeds (exits frame), SSTORE never executes (APPROVE terminates frame like RETURN). (5) Integration test: VERIFY frame completes without any APPROVE → transaction invalid. (6) Assert caller is `ENTRY_POINT` inside VERIFY frame and all sub-calls. (7) Sub-call from within VERIFY frame also sees `readOnly = true`; (8) VERIFY frame that issues a `CALL` opcode with non-zero value → exceptional halt (value transfer is blocked by STATICCALL semantics, identical to how a standard `STATICCALL` restricts value-bearing calls). |
| **Definition of Done** | All 7 tests pass; `readOnly` flag propagated through sub-calls; `APPROVE` bypasses `readOnly` check; missing APPROVE causes tx invalid; reviewed. |

##### Task 2.1.3 — SENDER Mode: Caller = tx.sender + Authorization Guard

| Field | Detail |
|-------|--------|
| **Description** | When a frame has `mode == SENDER`, the EVM call's caller must be `tx.sender`. Before dispatching the call, check that `sender_approved == true` in the `FrameContext`; if not, reject the entire transaction as invalid (not just revert the frame). No state-modification restrictions apply. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | (1) Integration test: SENDER frame after successful APPROVE(0x0) — caller inside call is `tx.sender`. (2) Integration test: SENDER frame without prior sender approval — transaction is rejected. (3) Assert state changes in SENDER mode persist (not static). |
| **Definition of Done** | Tests pass; transaction rejected (not reverted) when `sender_approved == false`; reviewed. |

##### Task 2.1.4 — Null Target Handling (defaults to tx.sender)

| Field | Detail |
|-------|--------|
| **Description** | When `frame.target` is `nil` (null), the call target must be set to `tx.sender`. This applies to all three modes. Implement this null-target substitution in the frame dispatch logic. Confirm RLP encoding encodes null target as `0x80` (per Task 1.1.1). |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | Unit test: frame with `Target == nil` dispatches call to `tx.sender`; assert via a contract that records `address(this)`. |
| **Definition of Done** | Test passes; null target substituted correctly for all three modes; RLP round-trip correct; reviewed. |

---

### EP-3: APPROVE Opcode

---

#### US-3.1 — APPROVE Opcode Core Behavior

> **As an** EVM engineer,
> **I want** the `APPROVE` opcode (`0xaa`) to update transaction-scoped approval state based on `scope`,
> **so that** smart accounts can authorize execution (`sender_approved`) and gas payment (`payer_approved`) in a controlled, non-reentrant manner.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

##### Task 3.1.1 — APPROVE Scope 0x0: Execution Approval

| Field | Detail |
|-------|--------|
| **Description** | Implement `scope == 0x0` branch in `opApprove` (`pkg/core/vm/eip8141_opcodes.go`): set `FrameContext.SenderApproved = true`. Preconditions in order: (1) `CALLER == frame.target` (universal APPROVE guard from Task 3.1.5), else revert; (2) `CALLER == tx.sender`, else revert — scope 0x0 is only valid when `frame.target` equals `tx.sender`; these two conditions together mean the sender contract is calling APPROVE on itself; (3) `SenderApproved` must not already be `true` — approval flags are **monotonic**: once set, they cannot be re-approved (revert on any re-approval attempt). On success: set `SenderApproved = true` and terminate the frame like `RETURN`, returning `memory[offset:offset+length]` as the frame's return data. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | Unit tests: (1) correct caller where `CALLER == frame.target == tx.sender` → `SenderApproved = true`, frame exits like RETURN; (2) `CALLER != frame.target` → revert (remaining gas returned); (3) `CALLER == frame.target` but `CALLER != tx.sender` (i.e., frame.target ≠ tx.sender) → revert; (4) double approval (scope 0x0 called when `SenderApproved` already true) → revert; (5) frame exits like RETURN — return data is `memory[offset:offset+length]`, no further EVM instructions execute after APPROVE; (6) APPROVE with non-zero `offset`/`length` stack arguments — frame return data is correctly `memory[offset:offset+length]` (e.g., offset=64, length=32 returns `memory[64:96]`; offset=0, length=0 returns empty return data), verifying offset and length parameters actually govern the returned slice. |
| **Definition of Done** | All 5 precondition tests pass; `SenderApproved` set exactly once; monotonicity enforced; frame exits correctly; reviewed. |

##### Task 3.1.2 — APPROVE Scope 0x1: Payment Approval

| Field | Detail |
|-------|--------|
| **Description** | Implement `scope == 0x1` branch: preconditions: (1) `CALLER == frame.target`, else revert; (2) `PayerApproved` not already set, else revert; (3) `SenderApproved == true`, else revert; (4) `frame.target` has sufficient balance to cover `tx_fee`, else revert. On success: increment `tx.sender` nonce by 1, deduct `tx_fee` from `frame.target` balance, set `PayerApproved = true`, record `payer = frame.target` in receipt context. Like all `APPROVE` scopes, the opcode then exits the frame like `RETURN` — returning `memory[offset:offset+length]` as the frame's return data, per the EIP: "`APPROVE` is like `RETURN (0xf3)`". |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | (1) Valid call with sufficient balance → nonce incremented, balance deducted, `PayerApproved = true`. (2) Insufficient balance → revert, state unchanged (remaining gas returned). (3) `SenderApproved == false` → revert. (4) Double payment approval → revert. (5) Assert payer address recorded. (6) Successful APPROVE(0x1) exits frame like RETURN — caller of the frame receives `memory[offset:offset+length]` as return data. |
| **Definition of Done** | All 6 tests pass; nonce increment atomic; balance deducted correctly; payer recorded; RETURN-like exit verified; reviewed. |

##### Task 3.1.3 — APPROVE Scope 0x2: Combined Execution + Payment Approval

| Field | Detail |
|-------|--------|
| **Description** | Implement `scope == 0x2` branch: sets both `SenderApproved = true` and `PayerApproved = true` atomically. Preconditions: (1) `CALLER == frame.target`, else revert; (2) `CALLER == tx.sender`, else revert; (3) neither `SenderApproved` nor `PayerApproved` already set, else revert; (4) `frame.target` (= `tx.sender`) has sufficient balance, else revert. On success: increment sender nonce, deduct fee from `tx.sender`, set both flags, record `payer = tx.sender`. Like all `APPROVE` scopes, the opcode exits the frame like `RETURN` — returning `memory[offset:offset+length]` as the frame's return data, per the EIP: "`APPROVE` is like `RETURN (0xf3)`". |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | (1) Valid combined approval → both flags true, nonce up, balance deducted. (2) `CALLER != tx.sender` → revert. (3) `SenderApproved` already true → revert. (4) `PayerApproved` already true → revert. (5) Insufficient balance → revert. (6) Scope 0x2 from non-sender target → revert. (7) Successful APPROVE(0x2) exits frame like RETURN — caller of the frame receives `memory[offset:offset+length]` as return data. |
| **Definition of Done** | All seven test cases pass; atomicity guaranteed (either both set or neither); RETURN-like exit verified; reviewed. |

##### Task 3.1.4 — APPROVE Invalid Scope Guard

| Field | Detail |
|-------|--------|
| **Description** | Any `scope` value outside `{0x0, 0x1, 0x2}` must result in an exceptional halt (`ErrInvalidApproveScope`). Implement this guard at the top of `opApprove`. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | Unit test with scopes `0x3`, `0xff`, `0xdeadbeef` — all must result in exceptional halt. |
| **Definition of Done** | Test passes; existing scope tests unaffected; reviewed. |

##### Task 3.1.5 — APPROVE Caller-Must-Equal-FrameTarget Invariant

| Field | Detail |
|-------|--------|
| **Description** | Enforce the universal precondition: `APPROVE` can only be called when `CALLER == frame.target`. This check applies before scope-specific checks. If `CALLER != frame.target`, the frame **reverts** (remaining gas returned to the per-frame budget) regardless of scope. Note: the EIP-8141 opcode description (line 124) says "exceptional halt" but the authoritative Behavior section (lines 148–160) consistently uses "revert the frame" — revert semantics apply here. Invalid scope values (outside `{0x0, 0x1, 0x2}`) remain an exceptional halt (Task 3.1.4). Implement this as the first check in `opApprove`. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | Unit tests: (1) call `APPROVE` from a contract that is not the frame target — must revert (remaining gas returned), confirm error `ErrCallerNotFrameTarget`; (2) APPROVE executed with fewer than 3 stack items (stack underflow) → exceptional halt, all frame gas consumed (standard EVM stack-underflow behavior, not a revert). |
| **Definition of Done** | Test passes; error type correct; reviewed. |

##### Task 3.1.6 — APPROVE Permitted in VERIFY Frames Despite STATICCALL

| Field | Detail |
|-------|--------|
| **Description** | Clarify and test that `APPROVE` is the **only** action in a VERIFY frame that is exempt from STATICCALL restrictions. `APPROVE` modifies **transaction-scoped** variables (`sender_approved`, `payer_approved`, nonce, payer balance), which are not EVM storage/account state. The `opApprove` implementation in `pkg/core/vm/eip8141_opcodes.go` must explicitly bypass the `vm.EVM.readOnly` flag — i.e., it must not call `vm.StaticCallError()` before executing. Any other opcode that normally requires `readOnly == false` (`SSTORE`, `LOG*`, `CREATE*`, `SELFDESTRUCT`, value-bearing `CALL`) must continue to halt when `readOnly == true`. The approval flags (`sender_approved`, `payer_approved`) and balance changes are tracked outside `StateDB` (in `FrameContext`) precisely so they are not subject to EVM static-call guards. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | (1) VERIFY frame calls `APPROVE(0x0)` → succeeds, `SenderApproved` becomes true, no STATICCALL error raised. (2) VERIFY frame calls `APPROVE(0x2)` → succeeds, both flags set. (3) VERIFY frame calls `SSTORE` → exceptional halt (readOnly enforced). (4) VERIFY frame calls `APPROVE(0x2)` then nothing else (APPROVE exits frame) → frame completes successfully, state-modifying code after APPROVE is unreachable. (5) Non-VERIFY frame with `readOnly == false` calls `APPROVE` → still subject to scope preconditions (not readOnly). (6) Confirm `opApprove` does **not** check `interpreter.readOnly` before executing. |
| **Definition of Done** | All 6 tests pass; `opApprove` documented as exempted from `readOnly` guard; code comment explains why (tx-scoped not EVM state); all other state-modifying opcodes still blocked in VERIFY; reviewed. |

---

### EP-4: TXPARAM* Opcodes

---

#### US-4.1 — TXPARAM Opcode Family: Transaction Parameter Introspection

> **As an** EVM contract developer,
> **I want** `TXPARAMLOAD` (0xb0), `TXPARAMSIZE` (0xb1), and `TXPARAMCOPY` (0xb2) opcodes to expose all 16 transaction parameters,
> **so that** smart account contracts can inspect transaction fields during validation and execution without expensive in-EVM recomputation.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

##### Task 4.1.1a — TXPARAMLOAD (0xb0): Scalar Parameters (0x00–0x10)

| Field | Detail |
|-------|--------|
| **Description** | Implement the scalar-parameter path of `opTxParamLoad` in `pkg/core/vm/eip8141_opcodes.go`. Stack layout: **2 inputs**: `in1` (parameter index), `in2` (must be `0` for all scalar indices). Returns a 32-byte word pushed onto the stack. Implement the **11 scalar indices**: `0x00` tx type, `0x01` nonce, `0x02` sender (left-padded address), `0x03` max_priority_fee_per_gas, `0x04` max_fee_per_gas, `0x05` max_fee_per_blob_gas, `0x06` max_cost, `0x07` blob hash count, `0x08` sig_hash, `0x09` frame count, `0x10` current frame index. Gap indices `0x0a`–`0x0f` are undefined in the EIP table and must result in exceptional halt. Any other undefined `in1` → exceptional halt. For `0x07`, note that per the EIP "blob index" out-of-bounds access results in an exceptional halt — since `in2` must be 0 for `0x07`, any non-zero `in2` is caught by the scalar enforcement rule and effectively represents an OOB blob index. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | Table-driven unit tests: (1) correct return value for each of the 11 valid scalar `in1` values from a known `FrameContext`; (2) TXPARAMLOAD(0x07, 0) returns 0 for a tx with no blobs, returns 2 for a tx with 2 blobs; (3) TXPARAMLOAD(0x10, 0) when executing frame N returns N; (4) `in1` in gap range `0x0a`–`0x0f` → exceptional halt; (5) other undefined `in1` (e.g., `0x16`, `0xff`) → exceptional halt. |
| **Definition of Done** | All 11 scalar indices pass tests; gap-index halts verified; `go vet` clean; reviewed; coverage ≥ 85%. |

##### Task 4.1.1b — TXPARAMLOAD (0xb0): Frame-Indexed Parameters (0x11–0x15) + Error Handling

| Field | Detail |
|-------|--------|
| **Description** | Implement the frame-indexed parameter path of `opTxParamLoad`. Indices `0x11`–`0x15` accept any valid frame index in `in2`; validate `in2 < len(frames)`, else exceptional halt. Implement: `0x11` frame[in2].target (32-byte left-padded address); `0x12` frame[in2].data (first 32 bytes, zero-padded if shorter; returns 32 zero bytes for VERIFY frames — data elided per sig hash rules); `0x13` frame[in2].gas_limit; `0x14` frame[in2].mode; `0x15` frame[in2].status (exceptional halt if queried from the currently-executing frame or a future frame; returns 0 or 1 for past completed frames). Integrate with the same `opTxParamLoad` function from Task 4.1.1a. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | (1) TXPARAMLOAD(0x12, frameIdx, 0) returns first 32 bytes of frame data, zero-padded; (2) TXPARAMLOAD(0x12, verifyFrameIdx, 0) returns 32 zero bytes (data elided for VERIFY frames); (3) out-of-bounds frame index for `0x11`–`0x15` → exceptional halt; (4) `status` (`0x15`) queried from the currently-executing frame → exceptional halt; (5) `status` queried for a future frame → exceptional halt; (6) `status` queried from a later frame on a past completed frame returns 0 or 1 matching the frame receipt. |
| **Definition of Done** | All 5 frame-indexed indices tested; status boundary conditions (current, future, past) tested; VERIFY data elision confirmed; `go vet` clean; reviewed. |

##### Task 4.1.2 — TXPARAMSIZE (0xb1): Dynamic Size Query

| Field | Detail |
|-------|--------|
| **Description** | Implement `opTxParamSize` for the `TXPARAMSIZE` opcode. Follows `CALLDATASIZE` pattern — takes `in1` and `in2`, returns the byte size of the parameter. For fixed-size parameters (all except `0x12` data), returns `32`. For `0x12` (frame data), returns `len(frame[in2].data)`. For VERIFY frames, `data` is elided → size is `0`. Apply same out-of-bounds and invalid-index rules as `TXPARAMLOAD`. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | (1) Fixed-size param → 32; (2) frame data with 100 bytes → 100; (3) VERIFY frame data → 0; (4) out-of-bounds frame index → halt; (5) invalid in1 → halt. |
| **Definition of Done** | Tests pass; VERIFY data elision verified; reviewed. |

##### Task 4.1.3 — TXPARAMCOPY (0xb2): Dynamic Copy

| Field | Detail |
|-------|--------|
| **Description** | Implement `opTxParamCopy` following the `CALLDATACOPY` pattern. Stack: `[in1, in2, destOffset, dataOffset, length]` (5 stack inputs — `in1` and `in2` identify the parameter, then the CALLDATACOPY-equivalent 3). Copies parameter bytes into memory. Apply standard EVM memory expansion gas cost. For `0x12` (frame data), copy `frame[in2].data[dataOffset:dataOffset+length]`; out-of-range bytes are zero-padded. For VERIFY frames, data is treated as zero-length. For fixed-size parameters, treat the 32-byte encoding as the source byte array. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | (1) Copy frame data into memory — assert memory contents match. (2) Offset beyond data length — zero-padded. (3) VERIFY frame data — zero-padded (treated as empty). (4) Memory expansion cost charged. (5) Out-of-bounds frame index → halt. |
| **Definition of Done** | Memory copy correct; zero-padding correct; gas cost verified; reviewed. |

##### Task 4.1.4 — TXPARAM Signature Hash (0x08)

| Field | Detail |
|-------|--------|
| **Description** | Ensure `TXPARAMLOAD(0x08, 0)` returns `compute_sig_hash(tx)` as a 32-byte hash. The sig hash must be pre-computed at transaction execution start and stored in `FrameContext.SigHash`. Implement `ComputeFrameSigHash(tx *FrameTx) common.Hash` in `pkg/core/types/tx_frame.go`: RLP-encode the transaction with all VERIFY frame data elided, then return `keccak256(rlp(tx))`. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Transaction with one VERIFY frame: assert hash differs from full-data hash; (2) transaction with no VERIFY frames: assert hash equals standard RLP hash; (3) two different VERIFY data payloads → same sig hash (data elided); (4) assert `TXPARAMLOAD(0x08)` returns same value as `ComputeFrameSigHash`. |
| **Definition of Done** | Tests pass; VERIFY data elision confirmed; hash deterministic; reviewed. |

##### Task 4.1.5 — TXPARAM Gas Cost

| Field | Detail |
|-------|--------|
| **Description** | All three TXPARAM opcodes follow standard EVM memory expansion costs. For `TXPARAMLOAD`/`TXPARAMSIZE`: base gas cost of 2 (`GasBase`). For `TXPARAMCOPY`: base gas cost of 3 (`GasVerylow`) + memory expansion. Implement gas cost functions in `pkg/core/vm/dynamic_gas.go` or alongside the opcode implementations. Register in the instruction table. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | Unit test: execute TXPARAMCOPY with known length in known memory state; assert gas consumed matches formula. Test memory expansion triggered on large copy. |
| **Definition of Done** | Gas cost matches CALLDATACOPY pattern; expansion cost charged; registered in instruction table; reviewed. |

##### Task 4.1.6 — TXPARAM Scalar Parameter `in2 == 0` Enforcement

| Field | Detail |
|-------|--------|
| **Description** | For all scalar (fixed-size) parameters — `in1` values `0x00` through `0x10` inclusive (11 indices) — the `in2` operand **must** be exactly `0`. Per the EIP table, these parameters have `in2 = "must be 0"`. Any non-zero `in2` for these indices results in an **exceptional halt**. This rule applies identically to `TXPARAMLOAD`, `TXPARAMSIZE`, and `TXPARAMCOPY`. Implement a guard in each opcode handler after identifying the `in1` value as scalar, before reading the parameter. Frame-indexed parameters (`in1 = 0x11`–`0x15`) accept any valid frame index in `in2` and do not require `in2 == 0`. Note: `in1` values `0x0a`–`0x0f` are gap indices and halt unconditionally (covered by Task 4.1.1). |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | Table-driven tests for each of the 11 scalar parameter indices (`0x00`, `0x01`, `0x02`, `0x03`, `0x04`, `0x05`, `0x06`, `0x07`, `0x08`, `0x09`, `0x10`): (1) `TXPARAMLOAD(index, 0)` → succeeds and returns correct value; (2) `TXPARAMLOAD(index, 1)` → exceptional halt; (3) `TXPARAMLOAD(index, 0xff)` → exceptional halt. Also test: (4) `TXPARAMSIZE(index, 1)` → exceptional halt for scalar indices; (5) frame-indexed params (`0x11`, `0x13`) with `in2 > 0` within bounds → succeeds (not halted by this rule). |
| **Definition of Done** | All 11 scalar indices tested for `in2 != 0` rejection; all three opcodes enforce the rule; frame-indexed params unaffected; `go vet` clean; reviewed. |

---

### EP-5: Transaction Execution Engine

---

#### US-5.1 — Frame-by-Frame Execution Orchestrator

> **As a** node operator,
> **I want** the transaction processor to execute each frame sequentially with correct caller, mode, and validation logic,
> **so that** frame transactions are processed according to the EIP-8141 state machine.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

##### Task 5.1.1 — Nonce Validation

| Field | Detail |
|-------|--------|
| **Description** | At the start of frame transaction processing (before any frame executes), check `tx.nonce == state[tx.sender].nonce`. If they differ, reject the transaction (not revert — the transaction is entirely invalid). The nonce increment happens inside `APPROVE(0x1)` or `APPROVE(0x2)`, not at transaction boundary. Implement in `pkg/core/frame_execution.go` or the state processor. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Correct nonce → proceeds to frame execution. (2) Nonce too low → transaction rejected. (3) Nonce too high → transaction rejected. (4) Assert no state changes occur on rejection. |
| **Definition of Done** | Tests pass; no state side-effects on rejection; reviewed. |

##### Task 5.1.2a — Frame Dispatch Loop: Core Setup

| Field | Detail |
|-------|--------|
| **Description** | Implement the skeleton of the frame execution loop in `pkg/core/frame_execution.go`. (1) Initialize `sender_approved = false` and `payer_approved = false` in `FrameContext` before any frame executes. (2) Iterate over `tx.frames`; for each frame: substitute null target with `tx.sender`; set caller per mode (DEFAULT/VERIFY → `ENTRY_POINT`, SENDER → `tx.sender`); dispatch EVM call with `frame.gas_limit`. (3) Continue to next frame regardless of individual frame success/failure. (4) At loop end, proceed to post-loop validation (Task 5.1.3). This task establishes the loop skeleton only — mode-specific guards and status recording are in Task 5.1.2b. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | Integration test: (1) two-frame tx (DEFAULT then SENDER) — both frames are dispatched in order; (2) null target in frame 0 resolves to `tx.sender` as call target; (3) DEFAULT frame caller = `ENTRY_POINT`; (4) SENDER frame caller = `tx.sender`; (5) frame gas limit is passed to the EVM call and not shared with adjacent frames. |
| **Definition of Done** | Loop skeleton dispatches all frames in order; null-target substitution works; correct caller set per mode; tests pass; reviewed. |

##### Task 5.1.2b — Frame Dispatch Loop: Validation State Machine

| Field | Detail |
|-------|--------|
| **Description** | Layer validation logic onto the frame loop from Task 5.1.2a: (1) For **SENDER** frames — before dispatching the EVM call, assert `sender_approved == true`; if false, immediately mark the entire transaction invalid (not just the frame). (2) For **VERIFY** frames — after the EVM call completes, detect whether `APPROVE` was successfully called by checking whether `sender_approved` or `payer_approved` changed during this frame; if neither changed, mark the entire transaction invalid. (3) **Immediately after each frame completes**, record `frame.status` (1 for success/clean return, 0 for revert or exception) in `FrameContext.Frames[i].Status` — this status becomes queryable via `TXPARAM(0x15, i)` for all subsequent frames but results in exceptional halt if queried for the currently-executing or a future frame. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | Integration tests: (1) Example 1 from EIP (VERIFY + SENDER) — succeeds end-to-end; (2) Example 1b (deploy + VERIFY + SENDER) — succeeds; (3) Example 2 (sponsored, 5 frames) — succeeds; (4) VERIFY frame completes without calling APPROVE → transaction invalid; (5) SENDER frame reached with `sender_approved == false` → transaction immediately invalid; (6) frame.status = 1 after a successful frame, 0 after a reverted frame; (7) TXPARAM(0x15) on a just-completed past frame returns correct status. |
| **Definition of Done** | All 7 tests pass; frame statuses recorded at correct timing; invalid-tx conditions return transaction-level error; reviewed. |

##### Task 5.1.3 — payer_approved Final Check and Gas Refund

| Field | Detail |
|-------|--------|
| **Description** | After all frames execute, verify `payer_approved == true`. If not, the entire transaction is invalid (no state changes commit). If true, compute gas refund: `refund = sum(frame.gas_limit) - total_gas_used_across_frames`. Refund this amount to the payer address (the account that called `APPROVE(0x1)` or `APPROVE(0x2)`). Add the refunded gas back to the block gas pool. This refund is separate from EIP-3529 storage refunds. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Transaction with unused gas → payer balance increases by refund × effective_gas_price; (2) all gas used → zero refund; (3) `payer_approved == false` at end → no state commits, no refund; (4) block gas pool correctly updated. |
| **Definition of Done** | Tests pass; payer refunded correctly; block pool updated; state rolled back when payer not approved; reviewed. |

##### Task 5.1.5 — State Atomicity on Invalid Transaction

| Field | Detail |
|-------|--------|
| **Description** | When a frame transaction is determined to be invalid at any point during execution — whether because a VERIFY frame completes without having called APPROVE, a SENDER frame is reached with `sender_approved == false`, or the final post-loop check finds `payer_approved == false` — **all state changes made by all preceding frames must be rolled back**. This includes nonce increments and fee deductions that APPROVE performed. The EIP states "the whole transaction is invalid" in each of these cases, which means no state change can commit. Implement this using a top-level state snapshot taken before the first frame executes; on any invalid condition, revert to that snapshot before returning. Implement in `pkg/core/frame_execution.go`. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Transaction with successful APPROVE(0x2) in frame 0 followed by a VERIFY frame that never calls APPROVE in frame 1: assert sender nonce was NOT incremented, payer balance was NOT deducted, and no other state changes persist. (2) Transaction with APPROVE(0x1) in frame 2 followed by a SENDER frame in frame 3 that has no prior `sender_approved`: assert the nonce increment from frame 2's APPROVE is rolled back. (3) Transaction where `payer_approved` is never set after all frames: assert all state changes (including any SSTORE from DEFAULT frames) are rolled back. (4) Valid transaction: assert state changes ARE committed after successful completion. |
| **Definition of Done** | All 4 test cases pass; snapshot-and-revert mechanism verified; no partial state visible after invalid tx; reviewed. |

##### Task 5.1.4 — ENTRY_POINT Address Constant

| Field | Detail |
|-------|--------|
| **Description** | Confirm `ENTRY_POINT = address(0xaa)` is defined as a constant in `pkg/core/types/tx_frame.go` and referenced consistently throughout the execution engine. The address is `0x00000000000000000000000000000000000000aa`. Ensure it is not confused with the `APPROVE` opcode value (`0xaa`) — these are different concepts sharing the same hex value. Add a comment clarifying this in the code. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | Code review: confirm constant defined; grep codebase for any hardcoded `0xaa` address strings; unit test that `EntryPointAddress == common.HexToAddress("0x00...00aa")`. |
| **Definition of Done** | Constant defined; comment added; no hardcoded address strings; test passes; reviewed. |

---

### EP-6: Gas Accounting

---

#### US-6.1 — Per-Frame Gas Isolation

> **As a** protocol engineer,
> **I want** each frame to have its own independent gas budget with no cross-frame gas spill,
> **so that** one frame cannot consume gas from another frame's allocation.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

##### Task 6.1.1 — Per-Frame Gas Limit Enforcement

| Field | Detail |
|-------|--------|
| **Description** | Each frame executes with exactly `frame.gas_limit` gas. Unused gas from a frame is **not** carried over to the next frame. When the frame's gas is exhausted, execution halts for that frame (out-of-gas exception), but the remaining frames still execute with their own gas limits. The total transaction gas is pre-charged at transaction entry (deducted from payer when `APPROVE(0x1)/(0x2)` is called). Implement gas isolation in the frame dispatch loop in `pkg/core/frame_execution.go`. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Frame A uses 100 of 200 gas limit; Frame B starts fresh with its own gas limit — not 100 extra. (2) Frame A exhausts gas — Frame A fails; Frame B executes normally. (3) Gas refund is sum of unused gas across all frames, not carry-over. |
| **Definition of Done** | Tests pass; no gas leak between frames; out-of-gas in frame A does not halt frame B; reviewed. |

##### Task 6.1.2 — Calldata Cost for Frames List RLP

| Field | Detail |
|-------|--------|
| **Description** | Compute calldata cost for the RLP-encoded frames list using standard EVM calldata cost rules (4 gas/zero byte, 16 gas/non-zero byte). This is added to `FRAME_TX_INTRINSIC_COST` in `CalcFrameTxGas`. Note: the actual implementation computes calldata cost inline within `CalcFrameTxGas` (via `CalldataTokenGas`), not as a separate exported function. Ensure the RLP encoding used matches the wire encoding used for transaction submission. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) All-zero data frame — cost = `4 * bytes`. (2) All non-zero data — cost = `16 * bytes`. (3) Mixed data — correct weighted sum. (4) Compare with values from EIP data-efficiency table (134 bytes for basic tx). |
| **Definition of Done** | Costs match EIP table examples; function deterministic; reviewed. |

##### Task 6.1.3 — Intrinsic Cost: FRAME_TX_INTRINSIC_COST = 15000

| Field | Detail |
|-------|--------|
| **Description** | Ensure `FRAME_TX_INTRINSIC_COST = 15000` is applied as a base cost to every frame transaction, separate from standard intrinsic costs (which do not apply to frame transactions — frame transactions have no `to` field in the traditional sense). This constant is already defined in `pkg/core/types/tx_frame.go`; verify it is applied in `CalcFrameTxGas` and not double-counted with legacy intrinsic cost calculation. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | Unit test: `CalcFrameTxGas` with zero-data frames and zero gas limits returns exactly `15000 + calldata_cost`. |
| **Definition of Done** | Test passes; legacy intrinsic cost not applied to frame txs; reviewed. |

---

### EP-7: Receipt Structure

---

#### US-7.1 — Frame Transaction Receipt

> **As a** block explorer developer,
> **I want** frame transaction receipts to include `payer`, `cumulative_gas_used`, and per-frame receipts,
> **so that** users and tools can trace which account paid fees and inspect per-frame outcomes.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

##### Task 7.1.1 — FrameReceipt Struct and RLP Encoding

| Field | Detail |
|-------|--------|
| **Description** | Define (or verify) the `FrameReceipt` struct in `pkg/core/types/frame_receipt.go`: `[status uint64, gas_used uint64, logs []*Log]`. The top-level receipt payload for frame transactions is: `[cumulative_gas_used, payer Address, frame_receipts []FrameReceipt]`. Implement `EncodeRLP` and `DecodeRLP` for both structs. Integrate with the EIP-2718 receipt dispatch so that transactions of type `0x06` use this receipt format. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Round-trip encode/decode for receipt with 3 frame receipts. (2) Assert `payer` field is the address that called `APPROVE(0x1)` or `APPROVE(0x2)`. (3) Assert `status` is 0 or 1 per frame. (4) Assert logs are correctly assigned to frames. (5) `cumulative_gas_used` accumulates across blocks correctly. |
| **Definition of Done** | Round-trip tests pass; payer populated from execution context; logs per-frame; coverage ≥ 80%; reviewed. |

##### Task 7.1.2 — Frame Status Tracking During Execution

| Field | Detail |
|-------|--------|
| **Description** | During frame execution, track each frame's outcome: `status = 1` (success) if the EVM call returns without exception; `status = 0` if it reverts or has an exception. Record `gas_used` as the gas consumed by that frame (not the limit). Record `logs` emitted by that frame. Store these in `FrameContext.Frames[i].Status` and populate into `FrameReceipt` objects after execution. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Successful frame → status=1, gas_used < gas_limit. (2) Reverted frame → status=0. (3) Log emitted in frame 1 appears only in frame_receipts[1].logs. (4) TXPARAM `0x15` (status) on a completed frame returns 0 or 1 matching the receipt. |
| **Definition of Done** | Tests pass; status/gas/logs correctly isolated per frame; TXPARAM status consistent with receipt; reviewed. |

##### Task 7.1.3 — JSON-RPC Receipt Serialization

| Field | Detail |
|-------|--------|
| **Description** | Extend `eth_getTransactionReceipt` JSON-RPC response for type-`0x06` transactions to include: `"payer": "0x..."`, `"frameReceipts": [{"status": "0x1", "gasUsed": "0x...", "logs": [...]}]`. The `payer` field is critical because it cannot be determined statically from the transaction. Update the JSON marshaling in `pkg/rpc/` (or equivalent). |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | RPC Engineer |
| **Testing Method** | Integration test: submit a frame transaction, fetch receipt via `eth_getTransactionReceipt`, assert `payer` and `frameReceipts` fields present and correct. Manual validation with `curl` against devnet. |
| **Definition of Done** | JSON includes `payer` and `frameReceipts`; values correct; existing receipt tests unaffected; reviewed. |

---

### EP-8: Signature Hash Computation

---

#### US-8.1 — Canonical Signature Hash with VERIFY Frame Elision

> **As a** smart account developer,
> **I want** a canonical signature hash that elides VERIFY frame data,
> **so that** accounts can sign a hash that is stable across gas sponsor changes and cannot be malleated by modifying VERIFY frame data.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

##### Task 8.1.1 — ComputeFrameSigHash Implementation

| Field | Detail |
|-------|--------|
| **Description** | Implement `ComputeFrameSigHash(tx *FrameTx) common.Hash` in `pkg/core/types/tx_frame.go`. Algorithm: (1) deep-copy `tx`; (2) for each frame where `mode == VERIFY`, set `frame.data = []byte{}`; (3) RLP-encode the modified transaction; (4) return `keccak256(rlp_bytes)`. The function must not mutate the original transaction. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Transaction with VERIFY frames — sig hash differs from full-data keccak; (2) replacing VERIFY frame data with different bytes → same sig hash; (3) non-VERIFY frames — changing their data changes sig hash; (4) `frame.target` of VERIFY frames is **not** elided — changing it changes sig hash; (5) no mutation of original tx. |
| **Definition of Done** | Tests pass; immutability of input verified; reviewed. |

##### Task 8.1.2 — Sig Hash Pre-Computation at Tx Entry

| Field | Detail |
|-------|--------|
| **Description** | Pre-compute `ComputeFrameSigHash` once when the frame transaction enters the execution pipeline and store the result in `FrameContext.SigHash`. This avoids recomputing it on every `TXPARAMLOAD(0x08)` call. Implement in the transaction pre-processing step, before frame dispatch. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | Assert `FrameContext.SigHash` is non-zero after pre-processing; assert `TXPARAMLOAD(0x08)` returns the same value as `ComputeFrameSigHash` called independently. |
| **Definition of Done** | Pre-computation verified; no recomputation in opcode handler; reviewed. |

---

### EP-9: Static Validity Constraints

---

#### US-9.1 — Static Transaction Validation

> **As a** node operator,
> **I want** statically invalid frame transactions to be rejected before entering the execution pipeline,
> **so that** the node wastes no execution resources on structurally malformed transactions.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

##### Task 9.1.1 — Implement Static Validity Check Function

| Field | Detail |
|-------|--------|
| **Description** | Implement `ValidateFrameTx(tx *FrameTx) error` in `pkg/core/types/tx_frame.go` (or a validator file). Enforce all static constraints from the EIP: (1) `tx.chain_id < 2^256` (always true for `*big.Int`, check for nil); (2) `tx.nonce < 2^64`; (3) `1 <= len(tx.frames) <= MAX_FRAMES (1000)`; (4) `len(tx.sender) == 20` (address type — check for zero address separately if needed); (5) for each frame: `frame.mode < 3`; (6) for each frame: `frame.target == nil || len(frame.target) == 20`. Return a descriptive error for each violation. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | Table-driven tests: one test case per constraint violation. Also test a valid transaction passes all checks. Test boundary conditions: `nonce = 2^64 - 1` (valid), `nonce = 2^64` (invalid); `len(frames) = 1` (valid), `len(frames) = 0` (invalid), `len(frames) = 1000` (valid), `len(frames) = 1001` (invalid). |
| **Definition of Done** | All 6 constraint categories tested with valid and invalid inputs; descriptive error messages; function called during tx pool ingress; reviewed. |

##### Task 9.1.2 — chain_id Validation

| Field | Detail |
|-------|--------|
| **Description** | Add chain ID matching to `ValidateFrameTx`: if `tx.chain_id` does not equal the current network's chain ID, reject the transaction. This prevents cross-chain replay. Integrate with the chain config in `pkg/core/`. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Correct chain ID → passes. (2) Wrong chain ID → rejected with `ErrInvalidChainID`. |
| **Definition of Done** | Test passes; integrated with chain config; reviewed. |

---

### EP-10: Frame Interactions & Cross-Frame State

---

#### US-10.1 — Shared Warm/Cold State Journal Across Frames

> **As a** protocol engineer,
> **I want** the access list warm/cold state journal to persist across all frames in a transaction,
> **so that** gas accounting for repeated state accesses is consistent within a transaction.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

##### Task 10.1.1 — Cross-Frame Access List Journal

| Field | Detail |
|-------|--------|
| **Description** | In the frame execution loop, maintain a single `AccessList` / warm-state journal across all frames. When Frame N warms a storage slot or account, Frame N+1 should see it as already warm. This is the opposite of transient storage (which is reset between frames). Implement by passing or sharing the state journal in `pkg/core/frame_execution.go`. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Frame 0 reads `address(X)` storage slot → cold access, pays 2100 gas. (2) Frame 1 reads same slot → warm access, pays 100 gas. (3) Assert gas charged correctly in each frame. (4) Confirm this applies to accounts as well as storage slots. |
| **Definition of Done** | Tests pass; warm state not reset between frames; gas charges correct; reviewed. |

##### Task 10.1.2 — Transient Storage (TSTORE/TLOAD) Isolation Between Frames

| Field | Detail |
|-------|--------|
| **Description** | Transient storage (EIP-1153 `TSTORE`/`TLOAD`) must be cleared between frames. At the end of each frame (before executing the next), discard all transient storage writes from that frame. A value written by `TSTORE` in Frame N is not visible to Frame N+1. Implement this in the frame dispatch loop by clearing the transient storage state after each frame completes. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Frame 0 executes `TSTORE(key, value)`. (2) Frame 1 executes `TLOAD(key)` → must return 0. (3) Within a single frame, `TSTORE` then `TLOAD` works normally. (4) Sub-calls within a frame share transient storage normally (only cleared at frame boundary). |
| **Definition of Done** | Tests pass; isolation confirmed; sub-call behavior unaffected; reviewed. |

---

### EP-11: ORIGIN Opcode Behavior Change

---

#### US-11.1 — ORIGIN Returns Frame Caller

> **As a** protocol engineer,
> **I want** the `ORIGIN` opcode to return the frame's caller (ENTRY_POINT or tx.sender) for frame transactions,
> **so that** contracts that use ORIGIN for identification get a consistent, meaningful value.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

##### Task 11.1.1 — ORIGIN Opcode Override for Frame Transactions

| Field | Detail |
|-------|--------|
| **Description** | Modify `opOrigin` in `pkg/core/vm/instructions.go` (or equivalent): when executing within a frame transaction context, return the frame's caller address (not the traditional `tx.origin`). For DEFAULT and VERIFY frames, this is `ENTRY_POINT`. For SENDER frames, this is `tx.sender`. The override applies throughout all call depths within that frame (sub-calls also see the frame caller as ORIGIN). Implement by checking `FrameContext != nil` in the EVM context when executing `ORIGIN`. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | (1) DEFAULT frame: contract at any call depth calls `ORIGIN` → returns `ENTRY_POINT`. (2) SENDER frame: `ORIGIN` → returns `tx.sender`. (3) VERIFY frame: `ORIGIN` → returns `ENTRY_POINT`. (4) Non-frame transaction: `ORIGIN` behavior unchanged. (5) Regression test: existing tests that use `ORIGIN` must still pass for legacy tx types. |
| **Definition of Done** | Tests pass for all three modes; non-frame txs unaffected; regression tests pass; reviewed; change documented with backward-compatibility note. |

##### Task 11.1.2 — Backward Compatibility Documentation

| Field | Detail |
|-------|--------|
| **Description** | Add a comment in `opOrigin` and in the execution engine noting the behavior change for frame transactions, consistent with the EIP's backward compatibility section. Reference EIP-7702's precedent. Add a test that documents the `ORIGIN == CALLER` anti-pattern and how it behaves under frame transactions. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | Code review of comment accuracy; test asserting `ORIGIN != tx.origin` for frame transactions. |
| **Definition of Done** | Comment merged; test documents behavior change; reviewed. |

---

### EP-12: Mempool Validation & DoS Mitigation

---

#### US-12.1 — Mempool Validation Policies for Frame Transactions

> **As a** node operator,
> **I want** the transaction pool to apply ERC-7562-inspired restrictions on VERIFY frames,
> **so that** the node is protected from mass-invalidation DoS attacks.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

##### Task 12.1.1 — Validation Frame Opcode Restrictions

| Field | Detail |
|-------|--------|
| **Description** | **Only VERIFY frames are executed during mempool simulation.** DEFAULT and SENDER frames are NOT run during mempool validation — they execute only at block inclusion time. This isolation prevents arbitrary EVM execution during propagation. In the mempool validation pipeline (`pkg/txpool/`), simulate each VERIFY frame with the following ERC-7562-inspired opcode restrictions to prevent mass invalidation DoS: (a) block-context opcodes forbidden: `TIMESTAMP`, `NUMBER`, `BASEFEE`, `DIFFICULTY`/`PREVRANDAO`, `GASLIMIT`, `COINBASE`, `BLOBHASH`; (b) external account introspection forbidden: `BALANCE`, `EXTCODESIZE`, `EXTCODECOPY`, `EXTCODEHASH`, `SELFBALANCE`; (c) cross-account storage access forbidden: `SLOAD`/`SSTORE` on accounts other than the VERIFY frame's own `frame.target`; (d) `DELEGATECALL` and `CALLCODE` forbidden. The simulation runs until `APPROVE` is successfully called (which exits the frame naturally via RETURN semantics), at which point the transaction is accepted into the mempool for propagation. These restrictions apply only during mempool simulation; block execution enforces no such limits. Implement using the opcode tracer / hook infrastructure in `pkg/core/vm/`. |
| **Estimated Effort** | 5 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) VERIFY frame using `TIMESTAMP` → rejected by mempool. (2) VERIFY frame using `BALANCE` of an external address → rejected by mempool. (3) VERIFY frame reading its own contract's storage (`SLOAD` on `frame.target`) → accepted. (4) VERIFY frame without restricted opcodes, calls `APPROVE(0x2)` → accepted into mempool. (5) Block processing (not mempool) allows all opcodes in VERIFY frames. (6) DEFAULT and SENDER frames are NOT executed during mempool simulation — assert no EVM execution for these modes during validation phase. |
| **Definition of Done** | Tests pass; VERIFY-only execution confirmed; restricted opcode list documented in code and config; block processing unaffected; reviewed. |

##### Task 12.1.2 — Validation Frame Gas Limit

| Field | Detail |
|-------|--------|
| **Description** | Enforce a maximum gas limit for VERIFY frames during mempool validation. Transactions where any VERIFY frame has `gas_limit` exceeding a configurable threshold (e.g., `MAX_VALIDATION_GAS = 100_000`) should be rejected by the mempool. The threshold should be configurable via node config. This limits the compute cost of mempool validation. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) VERIFY frame with `gas_limit = 50_000` → accepted. (2) VERIFY frame with `gas_limit = 200_000` → rejected by mempool. (3) Non-VERIFY frames not subject to this limit. (4) Configurable threshold respected. |
| **Definition of Done** | Tests pass; threshold configurable; non-verify frames unaffected; reviewed. |

##### Task 12.1.3 — Deployer Factory Allowlist

| Field | Detail |
|-------|--------|
| **Description** | When a frame transaction's first frame has `mode == DEFAULT` and targets a deployer contract (e.g., EIP-7997 deployer), the mempool must only accept known, allowlisted deployer factory addresses as `frame.target`. This ensures that account deployment is deterministic and chain-state-independent. Implement a configurable allowlist in the mempool config. Reject transactions whose first frame targets an unknown deployer address. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Transaction with allowlisted deployer → accepted. (2) Transaction with unknown deployer address as first DEFAULT frame → rejected. (3) Non-deployment first frames (e.g., first frame is VERIFY) → not subject to this check. (4) Allowlist configurable at startup. |
| **Definition of Done** | Tests pass; allowlist configurable; documented in node config; reviewed. |

##### Task 12.1.4 — Single Pending Transaction per Sender

| Field | Detail |
|-------|--------|
| **Description** | Implement the constraint that at most one frame transaction can be pending in the mempool per sender address, consistent with EIP-7702 relay restrictions. If a second frame transaction from the same sender is submitted while one is already pending, the mempool should reject or replace (per standard gas bump rules). Implement in `pkg/txpool/`. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Submit two frame transactions from same sender → second rejected (or replaces first per gas bump). (2) Mixed legacy + frame txs from same sender: legacy tx behavior unchanged. (3) Two frame txs from different senders → both accepted. |
| **Definition of Done** | Tests pass; no mempool bloat per sender; reviewed. |

##### Task 12.1.5 — First VERIFY Frame Enforcement

| Field | Detail |
|-------|--------|
| **Description** | For mempool relay (not block inclusion), enforce the recommendation that the first substantive validation frame must be a VERIFY frame (or be preceded only by deployer DEFAULT frames per the allowlist). If the first non-deployer frame is not VERIFY, the mempool may reject the transaction. This is a soft policy (mempool only) to simplify validation and DoS resistance. Make this behavior configurable. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Standard pattern (VERIFY first) → accepted. (2) SENDER first frame → rejected by strict mempool policy. (3) Deployer DEFAULT then VERIFY → accepted. (4) Policy can be disabled in config. |
| **Definition of Done** | Tests pass; policy configurable; block inclusion not affected; reviewed. |

---

### EP-13: Integration Testing & End-to-End Validation

---

#### US-13.1 — EIP-8141 Integration Test Suite

> **As a** QA engineer,
> **I want** a comprehensive integration test suite covering all EIP-8141 examples and edge cases,
> **so that** the implementation is validated end-to-end before mainnet activation.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

##### Task 13.1.1 — Example 1: Simple Smart Account Transaction

| Field | Detail |
|-------|--------|
| **Description** | Write an integration test for EIP-8141 Example 1: Frame 0 (VERIFY, target=sender, APPROVE(0x2)), Frame 1 (SENDER, target=destination, call data). Deploy a mock smart account contract that verifies a signature in its VERIFY frame and calls APPROVE(0x2). Execute the frame transaction and assert: frame 0 succeeds, sender_approved and payer_approved both true, frame 1 executes as sender, state changes committed, receipt correct. |
| **Estimated Effort** | 5 story points |
| **Assignee/Role** | QA Engineer / Core Protocol Engineer |
| **Testing Method** | End-to-end test using devnet or test EVM: deploy contracts, submit frame tx, assert receipt payer, assert state changes, assert logs. |
| **Definition of Done** | Test passes on devnet; all assertions hold; reviewed. |

##### Task 13.1.2 — Example 1b: Account Deployment Flow

| Field | Detail |
|-------|--------|
| **Description** | Write an integration test for Example 1b: Frame 0 (DEFAULT, deployer, initcode+salt), Frame 1 (VERIFY, null=sender, APPROVE(0x2)), Frame 2 (SENDER, null=sender, transfer). Verify that deployment happens before validation, sender address has code after frame 0, VERIFY in frame 1 succeeds, SENDER executes correctly. |
| **Estimated Effort** | 5 story points |
| **Assignee/Role** | QA Engineer |
| **Testing Method** | E2E test: assert `code(sender) != empty` after tx; assert transfer executed; assert receipt has 3 frame receipts. |
| **Definition of Done** | Test passes; deployment deterministic; reviewed. |

##### Task 13.1.3 — Example 2: Sponsored Transaction (ERC-20 Fee Payment)

| Field | Detail |
|-------|--------|
| **Description** | Write an integration test for Example 2 (5-frame sponsored transaction): Frame 0 (VERIFY sender, APPROVE(0x0)), Frame 1 (VERIFY sponsor, APPROVE(0x1)), Frame 2 (SENDER, ERC-20 transfer to sponsor), Frame 3 (SENDER, intended call), Frame 4 (DEFAULT, sponsor post-op). Verify: sender pays in ERC-20, payer is sponsor address, sponsor balance changes correct, post-op frame executes. |
| **Estimated Effort** | 8 story points |
| **Assignee/Role** | QA Engineer |
| **Testing Method** | E2E test with mock ERC-20 contract and sponsor contract; assert all 5 frame receipts; assert ERC-20 balance changes; assert payer = sponsor. |
| **Definition of Done** | Test passes; payer field verified; ERC-20 accounting correct; reviewed. |

##### Task 13.1.4 — Negative Test Cases

| Field | Detail |
|-------|--------|
| **Description** | Write tests for all invalid transaction scenarios: (1) VERIFY without APPROVE → invalid tx; (2) SENDER without prior sender_approved → invalid tx; (3) payer_approved never set → invalid tx; (4) double APPROVE(0x0) → VERIFY frame reverts, tx invalid; (5) APPROVE from non-target caller → revert; (6) insufficient balance for payment → VERIFY frame reverts; (7) static constraint violations (wrong nonce, bad mode, too many frames) → rejected before execution. |
| **Estimated Effort** | 5 story points |
| **Assignee/Role** | QA Engineer |
| **Testing Method** | Each scenario is an independent test; assert transaction-level rejection vs frame-level revert; assert no state changes on rejection. |
| **Definition of Done** | All 7 scenarios tested; pass/fail distinction between invalid tx and reverted frame correct; reviewed. |

---

## Story Point Summary

| Epic | User Story | Total SP | Notes |
|------|-----------|----------|-------|
| EP-1 | US-1.1, US-1.2 | 11 | |
| EP-2 | US-2.1 | 7 | |
| EP-3 | US-3.1 | 12 | +2 SP: Task 3.1.6 (APPROVE in VERIFY) |
| EP-4 | US-4.1 | 16 | 3+2+2+3+3+2+1 = 16; split 4.1.1→4.1.1a(3)+4.1.1b(2) |
| EP-5 | US-5.1 | 13 | |
| EP-6 | US-6.1 | 5 | |
| EP-7 | US-7.1 | 7 | |
| EP-8 | US-8.1 | 3 | |
| EP-9 | US-9.1 | 3 | |
| EP-10 | US-10.1 | 5 | |
| EP-11 | US-11.1 | 4 | |
| EP-12 | US-12.1 | 14 | |
| EP-13 | US-13.1 | 23 | |
| **Total** | | **~124 SP** | |

---

## Sprint Allocation (Suggested)

### Sprint 1 (Foundations) — ~24 SP
- US-1.1 (RLP encoding, envelope, gas calc) — 7 SP
- US-1.2 (Fee calculation) — 3 SP
- US-9.1 (Static validity) — 3 SP
- US-8.1 (Sig hash) — 3 SP
- US-2.1 Tasks 2.1.1 + 2.1.4 (DEFAULT mode + null target) — 2 SP
- US-5.1 Task 5.1.4 (ENTRY_POINT constant) — 1 SP
- US-11.2 (ORIGIN documentation) — 1 SP
- US-6.1 Task 6.1.3 (Intrinsic cost) — 1 SP
- US-3.1 Tasks 3.1.4, 3.1.5 (APPROVE guards) — 2 SP

### Sprint 2 (APPROVE + VERIFY Mode) — ~29 SP
- US-2.1 Tasks 2.1.2 + 2.1.3 (VERIFY + SENDER modes) — 5 SP
- US-3.1 Tasks 3.1.1 + 3.1.2 + 3.1.3 (APPROVE scopes) — 8 SP
- US-3.1 Task 3.1.6 (APPROVE in VERIFY / tx-scoped state) — 2 SP
- US-6.1 Tasks 6.1.1 + 6.1.2 (Per-frame gas) — 4 SP
- US-5.1 Tasks 5.1.1 + 5.1.2 (Nonce check + frame loop) — 6 SP
- US-5.1 Task 5.1.5 (State atomicity on invalid tx) — 3 SP
- US-11.1 Task 11.1.1 (ORIGIN override) — 3 SP

### Sprint 3 (TXPARAM* + Receipt) — ~25 SP
- US-4.1 Tasks 4.1.1–4.1.6 (All TXPARAM opcodes + in2 guard) — 17 SP
- US-7.1 Tasks 7.1.1 + 7.1.2 (Receipt struct + tracking) — 5 SP
- US-5.1 Task 5.1.3 (Refund) — 3 SP

### Sprint 4 (Mempool + Cross-Frame) — ~23 SP
- US-10.1 (Warm/cold + transient storage) — 5 SP
- US-7.1 Task 7.1.3 (JSON-RPC receipt) — 2 SP
- US-12.1 Tasks 12.1.1–12.1.5 (Mempool policies) — 14 SP
- US-12.1 remaining: deployer allowlist — 3 SP (overlap)

### Sprint 5 (E2E Testing) — ~23 SP
- US-13.1 Tasks 13.1.1–13.1.4 (All integration tests) — 23 SP

---

## Architecture Notes

### Key Files (Existing)
| File | Role |
|------|------|
| `pkg/core/types/tx_frame.go` | `FrameTx`, `Frame` structs, constants, `CalcFrameTxGas`, `ComputeFrameSigHash` |
| `pkg/core/types/frame_receipt.go` | `FrameReceipt` struct |
| `pkg/core/vm/eip8141_opcodes.go` | `opApprove`, `opTxParamLoad`, `opTxParamSize`, `opTxParamCopy` |
| `pkg/core/vm/opcodes.go` | Opcode constant registration |
| `pkg/core/frame_execution.go` | Frame execution engine (`ExecuteFrameTx`, `CalcFrameRefund`, `MaxFrameTxCost`, `BuildFrameReceipt`) |
| `pkg/core/processor.go` | Wires frame execution into the state processor (lines 1038-1143) |
| `pkg/txpool/` | Mempool validation policies |
| `pkg/rpc/` | JSON-RPC receipt serialization |

### New Constants (Required)
```go
// pkg/core/types/tx_frame.go
const (
    FrameTxType          byte   = 0x06
    FrameTxIntrinsicCost uint64 = 15000
    MaxFrames            int    = 1000
    ModeDefault          uint8  = 0
    ModeVerify           uint8  = 1
    ModeSender           uint8  = 2
)

var EntryPointAddress = common.HexToAddress(
    "0x00000000000000000000000000000000000000aa",
)
```

### New Opcodes (Required)
```go
// pkg/core/vm/opcodes.go
APPROVE     OpCode = 0xaa
TXPARAMLOAD OpCode = 0xb0
TXPARAMSIZE OpCode = 0xb1
TXPARAMCOPY OpCode = 0xb2
```

### Refactoring Considerations
1. **FrameContext injection:** The `FrameContext` struct must be passed through the EVM context (`vm.BlockContext` or `vm.TxContext`) so opcode handlers can access it. Consider adding a `FrameCtx *FrameContext` field to `vm.EVM`.
2. **STATICCALL flag reuse:** VERIFY mode can reuse the existing `vm.EVM.readOnly` flag — set it to `true` before VERIFY frame calls.
3. **Access list sharing:** The EVM's `StateDB` access list journal must be shared across frame calls; reset only at transaction commit, not between frames.
4. **Transient storage:** EIP-1153 transient storage lives in `StateDB`; add a `ClearTransientStorage()` method called between frames.
5. **Gas metering:** Frame gas limits must be enforced at the `Contract` level within each EVM call; the outer gas pool tracks the block-level budget.

---

## Definition of Done (Global)

A story or task is considered **Done** when all of the following are true:

1. Code passes `go fmt ./...` with no diffs.
2. Code passes `go vet ./...` with no warnings.
3. All new code has corresponding unit or integration tests.
4. Test coverage for new code is ≥ 80%.
5. All existing tests (regression suite) continue to pass.
6. Code has been reviewed and approved by at least one peer reviewer.
7. The implementation matches the EIP-8141 specification exactly.
8. The commit message follows Conventional Commits format and is ≤ 40 characters.
9. No `Co-Authored-By` or AI attribution in commits.
10. Changes merged to the feature branch; the feature branch builds cleanly.

---

*This document covers all 15 key areas specified in EIP-8141: transaction type (§1), frame structure (§2), frame modes (§3), APPROVE opcode (§4), TXPARAM* opcodes (§5), validation flow (§6), gas accounting (§7), receipt structure (§8), signature hash (§9), static constraints (§10), frame interactions (§11), ORIGIN opcode change (§12), ENTRY_POINT address (§13), null target handling (§14), and mempool security (§15). Review refinements applied in this pass: (a) Task 3.1.1: scope 0x0 now explicitly requires `frame.target == tx.sender`, approval flags documented as monotonic; (b) Tasks 3.1.2/3.1.3: RETURN-like frame termination formally specified in description and testing; (c) Task 3.1.6 (new): APPROVE exempted from STATICCALL guard because it modifies tx-scoped not EVM state; (d) Task 2.1.2: APPROVE exception to readOnly documented, detection mechanism (checking flag changes) specified; (e) Task 4.1.1: 2 stack inputs (in1, in2) confirmed per implementation, all 16 indices tested, gap indices 0x0a–0x0f and current-frame status halt added; (f) Task 4.1.6 (new): explicit in2==0 enforcement for all 11 scalar parameters; (g) Task 5.1.2: frame.status finalization timing clarified — status is recorded immediately after frame completes and is queryable only by subsequent frames; (h) Task 12.1.1: VERIFY-only mempool simulation rule stated explicitly, specific forbidden opcodes enumerated per ERC-7562 pattern.*

---

## Cross-Review Notes (iFlow + Claude verification rounds)

### Coverage Verification (all 10 checkpoints confirmed ✅)

| # | Checkpoint | Status | Story/Task |
|---|-----------|--------|------------|
| 1 | APPROVE scope 0x0 CALLER==tx.sender, scope 0x1 sender_approved prerequisite, scope 0x2 combined, double-approval prevention | ✅ Covered | US-3.1 Tasks 3.1.1–3.1.5 |
| 2 | TXPARAM* all 16 param types, OOB frame index → halt, invalid in1 → halt, VERIFY data returns size 0 | ✅ Covered | US-4.1 Tasks 4.1.1–4.1.6 |
| 3 | VERIFY is STATICCALL (no state mod); SENDER requires sender_approved else tx invalid | ✅ Covered | US-2.1 Tasks 2.1.2, 2.1.3 |
| 4 | Per-frame isolated gas; unused gas NOT available to subsequent frames; refund to payer | ✅ Covered | US-6.1, US-5.1 Task 5.1.3 |
| 5 | Receipt: payer field, per-frame [status, gas_used, logs] | ✅ Covered | US-7.1 Tasks 7.1.1–7.1.3 |
| 6 | Signature hash: VERIFY frame data elided | ✅ Covered | US-8.1 Task 8.1.1 |
| 7 | Frame interactions: shared warm/cold journal, TSTORE/TLOAD discarded between frames | ✅ Covered | US-10.1 Tasks 10.1.1, 10.1.2 |
| 8 | ORIGIN returns frame caller throughout all call depths | ✅ Covered | US-11.1 Task 11.1.1 |
| 9 | Null target defaults to tx.sender | ✅ Covered | US-2.1 Task 2.1.4 |
| 10 | Mempool: opcode restrictions in validation, deployer factory allowlisting, single pending tx per account | ✅ Covered | US-12.1 Tasks 12.1.1–12.1.5 |

### Terminology Clarification

- **APPROVE error semantics**: EIP-8141 uses "revert" (not "exceptional halt") for APPROVE caller/scope failures. Task 3.1.5 already correctly distinguishes: `CALLER != frame.target` → **revert** (gas returned); invalid scope → **exceptional halt** (gas consumed). STATICCALL violations (SSTORE, LOG, etc. in VERIFY mode) remain **exceptional halt** per standard EVM behavior.
- **TXPARAM* errors**: Invalid `in1`, OOB frame index, non-zero `in2` for scalar params → all correctly remain **exceptional halt** per EIP specification.

### Suggested Edge Test Additions (incorporated above)

1. APPROVE with `offset`/`length` stack params: verify return data is `memory[offset:offset+length]` (Task 3.1.1 testing)
2. APPROVE with stack underflow (< 3 items) → exceptional halt (Task 3.1.5 testing, item 2)
3. CALL with non-zero value in VERIFY frame → exceptional halt (Task 2.1.2 testing, item 8)
4. MAX_FRAMES (1000) performance integration test (Task 9.1.1 suggested extension)
