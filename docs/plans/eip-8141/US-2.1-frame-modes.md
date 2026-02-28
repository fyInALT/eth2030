# US-2.1 — Frame Mode Definitions and Caller Identity

**Epic:** EP-2 Frame Structure & Mode Semantics
**Total Story Points:** 7
**Sprint:** 1-2 (Tasks 2.1.1/2.1.4 in Sprint 1, Tasks 2.1.2/2.1.3 in Sprint 2)

> **As a** protocol engineer,
> **I want** each frame mode (DEFAULT=0, VERIFY=1, SENDER=2) to set the correct caller and execution constraints,
> **so that** frames execute with the correct identity and state-modification permissions.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

---

## Tasks

### Task 2.1.1 — DEFAULT Mode: Caller = ENTRY_POINT

| Field | Detail |
|-------|--------|
| **Description** | In the frame execution engine (`pkg/core/frame_execution.go` or equivalent), when a frame has `mode == DEFAULT`, set the `caller` of the EVM call to `ENTRY_POINT` (`address(0xaa)`). Confirm `EntryPointAddress` is defined in `pkg/core/types/tx_frame.go` as `HexToAddress("0x00000000000000000000000000000000000000aa")`. No state-modification restrictions apply. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | Integration test: deploy a contract that records `msg.sender`; execute a DEFAULT frame targeting it; assert `msg.sender == ENTRY_POINT`. |
| **Definition of Done** | Test passes; `EntryPointAddress` constant defined; reviewed. |

### Task 2.1.2 — VERIFY Mode: STATICCALL Semantics + APPROVE Requirement

| Field | Detail |
|-------|--------|
| **Description** | When a frame has `mode == VERIFY`, the EVM call must behave as a `STATICCALL` — all standard state-modifying opcodes (`SSTORE`, `LOG*`, `CREATE*`, `SELFDESTRUCT`, `CALL` with non-zero value) must result in exceptional halt; the `readOnly` flag must be `true` and propagated through all sub-calls. The frame's caller must be `ENTRY_POINT`. **Critical exception:** the `APPROVE` opcode (`0xaa`) is explicitly permitted in VERIFY frames despite STATICCALL semantics. `APPROVE` modifies **transaction-scoped** state (`sender_approved`, `payer_approved`, nonce, balance), not EVM account/storage state; therefore `opApprove` must bypass the `readOnly` check and is the only state-changing action allowed. After frame completion, detect whether `APPROVE` was called by inspecting whether `sender_approved` or `payer_approved` changed during the frame; if neither changed (APPROVE was never called successfully), mark the entire transaction invalid. Integrate with `pkg/core/frame_execution.go`. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | (1) VERIFY frame that calls `SSTORE` → exceptional halt. (2) VERIFY frame that calls `LOG0` → exceptional halt. (3) VERIFY frame that calls `APPROVE(0x2)` → succeeds despite STATICCALL; both `sender_approved` and `payer_approved` become true. (4) VERIFY frame that calls `APPROVE` followed by `SSTORE` — APPROVE succeeds (exits frame), SSTORE never executes (APPROVE terminates frame like RETURN). (5) Integration test: VERIFY frame completes without any APPROVE → transaction invalid. (6) Assert caller is `ENTRY_POINT` inside VERIFY frame and all sub-calls. (7) Sub-call from within VERIFY frame also sees `readOnly = true`; (8) VERIFY frame that issues a `CALL` opcode with non-zero value → exceptional halt (value transfer is blocked by STATICCALL semantics). |
| **Definition of Done** | All 8 tests pass; `readOnly` flag propagated through sub-calls; `APPROVE` bypasses `readOnly` check; missing APPROVE causes tx invalid; reviewed. |

### Task 2.1.3 — SENDER Mode: Caller = tx.sender + Authorization Guard

| Field | Detail |
|-------|--------|
| **Description** | When a frame has `mode == SENDER`, the EVM call's caller must be `tx.sender`. Before dispatching the call, check that `sender_approved == true` in the `FrameContext`; if not, reject the entire transaction as invalid (not just revert the frame). No state-modification restrictions apply. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | (1) Integration test: SENDER frame after successful APPROVE(0x0) — caller inside call is `tx.sender`. (2) Integration test: SENDER frame without prior sender approval — transaction is rejected. (3) Assert state changes in SENDER mode persist (not static). |
| **Definition of Done** | Tests pass; transaction rejected (not reverted) when `sender_approved == false`; reviewed. |

### Task 2.1.4 — Null Target Handling (defaults to tx.sender)

| Field | Detail |
|-------|--------|
| **Description** | When `frame.target` is `nil` (null), the call target must be set to `tx.sender`. This applies to all three modes. Implement this null-target substitution in the frame dispatch logic. Confirm RLP encoding encodes null target as `0x80` (per Task 1.1.1). |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | Unit test: frame with `Target == nil` dispatches call to `tx.sender`; assert via a contract that records `address(this)`. |
| **Definition of Done** | Test passes; null target substituted correctly for all three modes; RLP round-trip correct; reviewed. |

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/types/tx_frame.go:18-22` | Mode constants: `ModeDefault=0`, `ModeVerify=1`, `ModeSender=2` |
| `pkg/core/types/tx_frame.go:25` | `EntryPointAddress = 0x00...00aa` |
| `pkg/core/vm/aa_executor.go` | EIP-7701 AA executor — handles type 0x04, NOT 0x06. **Not reusable for frame txs.** |
| `pkg/core/vm/eip8141_opcodes.go:30-35` | `FrameModeDefault/Verify/Sender` constants (duplicated from types) |

## Implementation Status

**✅ Mostly Implemented**

- ✅ Mode constants defined (`ModeDefault=0`, `ModeVerify=1`, `ModeSender=2`)
- ✅ `EntryPointAddress` defined as `0x00...00aa`
- ✅ Frame execution dispatcher in `pkg/core/frame_execution.go`: `ExecuteFrameTx` handles all three modes (DEFAULT/VERIFY → `ENTRY_POINT` caller, SENDER → `tx.sender` caller)
- ✅ SENDER mode `sender_approved` guard (`ErrFrameSenderNotApproved`)
- ✅ VERIFY mode APPROVE requirement enforced (`ErrFrameVerifyNoApprove`)
- ✅ Null target → `tx.sender` substitution in frame_execution.go lines 66-69
- ⚠️ **Gap:** VERIFY mode STATICCALL enforcement (readOnly flag) — not set in processor.go's `callFn` callback

---

## EIP-8141 Reference Excerpts

### Specification → Modes

> There are three modes:
>
> | Mode | Name           | Summary                                                   |
> | ---- | -------------- | --------------------------------------------------------- |
> |    0 | `DEFAULT`      | Execute frame as `ENTRY_POINT`                            |
> |    1 | `VERIFY`       | Frame identifies as transaction validation                |
> |    2 | `SENDER`       | Execute frame as `sender`                                 |
>
> ##### `DEFAULT` Mode
>
> Frame executes as regular call where the caller address is `ENTRY_POINT`.
>
> ##### `VERIFY` Mode
>
> Identifies the frame as a validation frame. Its purpose is to *verify* that a sender and/or payer authorized the transaction. It must call `APPROVE` during execution. Failure to do so will result in the whole transaction being invalid.
>
> The execution behaves the same as `STATICCALL`, state cannot be modified.
>
> Frames in this mode will have their data elided from signature hash calculation and from introspection by other frames.
>
> ##### `SENDER` Mode
>
> Frame executes as regular call where the caller address is `sender`. This mode effectively acts on behalf of the transaction sender and can only be used after explicitly approved.

### Specification → Behavior (frame dispatch)

> Then for each call frame:
>
> 1. Execute a call with the specified `mode`, `target`, `gas_limit`, and `data`.
>    - If `target` is null, set the call target to `tx.sender`.
>    - If mode is `SENDER`:
>        - `sender_approved` must be `true`. If not, the transaction is invalid.
>        - Set `caller` as `tx.sender`.
>    - If mode is `DEFAULT` or `VERIFY`:
>        - Set the `caller` to `ENTRY_POINT`.
>    - The `ORIGIN` opcode returns frame `caller` throughout all call depths.
> 2. If frame has mode `VERIFY` and the frame did not successfully call `APPROVE`, the transaction is invalid.
