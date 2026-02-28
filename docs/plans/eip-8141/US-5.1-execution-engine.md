# US-5.1 — Frame-by-Frame Execution Orchestrator

**Epic:** EP-5 Transaction Execution Engine
**Total Story Points:** 13
**Sprint:** 2

> **As a** node operator,
> **I want** the transaction processor to execute each frame sequentially with correct caller, mode, and validation logic,
> **so that** frame transactions are processed according to the EIP-8141 state machine.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

---

## Tasks

### Task 5.1.1 — Nonce Validation

| Field | Detail |
|-------|--------|
| **Description** | Before any frame executes, check `tx.nonce == state[tx.sender].nonce`. If they differ, reject the transaction (entirely invalid). The nonce increment happens inside `APPROVE(0x1)` or `APPROVE(0x2)`, not at transaction boundary. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Correct nonce → proceeds. (2) Nonce too low → rejected. (3) Nonce too high → rejected. (4) No state changes on rejection. |
| **Definition of Done** | Tests pass; no state side-effects on rejection; reviewed. |

### Task 5.1.2a — Frame Dispatch Loop: Core Setup

| Field | Detail |
|-------|--------|
| **Description** | Implement frame execution loop skeleton: (1) Initialize `sender_approved = false`, `payer_approved = false`. (2) For each frame: substitute null target → `tx.sender`; set caller per mode; dispatch EVM call with `frame.gas_limit`. (3) Continue regardless of individual frame outcome. (4) Proceed to post-loop validation. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Two-frame tx dispatched in order; (2) null target resolves to `tx.sender`; (3) DEFAULT caller = `ENTRY_POINT`; (4) SENDER caller = `tx.sender`; (5) frame gas limit not shared. |
| **Definition of Done** | Loop dispatches all frames; null-target works; correct caller per mode; reviewed. |

### Task 5.1.2b — Frame Dispatch Loop: Validation State Machine

| Field | Detail |
|-------|--------|
| **Description** | Layer validation onto frame loop: (1) SENDER frames — assert `sender_approved == true`, else tx invalid. (2) VERIFY frames — after completion, check if APPROVE was called (flag change detection); if not, tx invalid. (3) Record `frame.status` immediately after each frame completes (queryable via `TXPARAM(0x15, i)` by subsequent frames). |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) EIP Example 1 (VERIFY + SENDER) end-to-end; (2) Example 1b (deploy + VERIFY + SENDER); (3) Example 2 (sponsored, 5 frames); (4) VERIFY without APPROVE → invalid; (5) SENDER without sender_approved → invalid; (6) frame.status = 1 (success) / 0 (revert); (7) TXPARAM(0x15) on past frame returns correct status. |
| **Definition of Done** | All 7 tests pass; frame statuses recorded correctly; reviewed. |

### Task 5.1.3 — payer_approved Final Check and Gas Refund

| Field | Detail |
|-------|--------|
| **Description** | After all frames: verify `payer_approved == true`, else tx invalid (no state commits). If true: refund = `sum(frame.gas_limit) - total_gas_used`. Refund to payer address. Add refunded gas to block gas pool. Separate from EIP-3529 storage refunds. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Unused gas → payer balance increases; (2) all gas used → zero refund; (3) `payer_approved == false` → no commits; (4) block gas pool updated. |
| **Definition of Done** | Tests pass; payer refunded; block pool updated; state rolled back when payer not approved; reviewed. |

### Task 5.1.4 — ENTRY_POINT Address Constant

| Field | Detail |
|-------|--------|
| **Description** | Confirm `ENTRY_POINT = address(0xaa)` = `0x00000000000000000000000000000000000000aa`. Not to be confused with the `APPROVE` opcode value (`0xaa`). Add clarifying comment. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | Code review; grep for hardcoded `0xaa` strings; unit test for constant value. |
| **Definition of Done** | Constant defined; comment added; test passes; reviewed. |

### Task 5.1.5 — State Atomicity on Invalid Transaction

| Field | Detail |
|-------|--------|
| **Description** | When a frame transaction is invalid at any point (VERIFY without APPROVE, SENDER without sender_approved, payer_approved never set), **all state changes from all preceding frames must be rolled back**, including nonce increments and fee deductions from APPROVE. Use a top-level state snapshot before the first frame; revert to it on any invalid condition. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) APPROVE(0x2) in frame 0 + VERIFY without APPROVE in frame 1 → nonce NOT incremented, balance NOT deducted. (2) APPROVE(0x1) in frame 2 + SENDER without sender_approved in frame 3 → rollback. (3) payer_approved never set → all state rolled back. (4) Valid tx → state committed. |
| **Definition of Done** | All 4 tests pass; snapshot-revert verified; no partial state on invalid tx; reviewed. |

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/frame_execution.go` | `ExecuteFrameTx` — complete frame dispatch loop with nonce validation, mode dispatch, SENDER/VERIFY guards, payer check, gas refund, receipt building |
| `pkg/core/processor.go:1038-1143` | Wires `ExecuteFrameTx` into state processor: sets up `FrameContext`, builds `callFn` callback, transient storage clearing, nonce increment |
| `pkg/core/vm/aa_executor.go` | EIP-7701 AA executor — handles type 0x04 only. **Not used for EIP-8141.** |
| `pkg/core/vm/eip8141_opcodes.go:38-56` | `FrameContext` struct (approval flags, tx params, frame list) |
| `pkg/core/types/tx_frame.go:243-268` | `ValidateFrameTx` static checks |

## Implementation Status

**✅ Complete**

- ✅ `ExecuteFrameTx` in `pkg/core/frame_execution.go`: nonce validation, frame dispatch loop, mode dispatch (DEFAULT/VERIFY/SENDER), SENDER guard (`sender_approved`), VERIFY guard (APPROVE required), post-loop `payer_approved` check
- ✅ `CalcFrameRefund` computes gas refund (`sum(frame.gas_limit) - total_gas_used`)
- ✅ `MaxFrameTxCost` computes max ETH cost (wired into `FrameContext.MaxCost` in processor.go)
- ✅ `BuildFrameReceipt` constructs receipt from execution context
- ✅ `processApprove` handles APPROVE scope 0/1/2 with monotonic flags
- ✅ `processor.go` wires everything: `FrameContext` setup, `ComputeFrameSigHash` pre-computation, transient storage clearing between frames, APPROVE tracking via `ApproveCalledThisFrame`/`LastApproveScope`, frame status recording, nonce increment after execution
- ⚠️ **Gap:** State atomicity (snapshot/revert on invalid tx) — `ExecuteFrameTx` returns an error on invalid conditions, but processor.go does not take a snapshot before frame execution to roll back partial state changes

---

## EIP-8141 Reference Excerpts

### Specification → Behavior

> When processing a frame transaction, perform the following steps.
>
> Perform stateful validation check:
>
> - Ensure `tx.nonce == state[tx.sender].nonce`
>
> Initialize with transaction-scoped variables:
>
> - `payer_approved = false`
> - `sender_approved = false`
>
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
>
> After executing all frames, verify that `payer_approved == true`. If it is, refund any unpaid gas to the gas payer. If it is not, the whole transaction is invalid.
>
> Note: it is implied by the handling that the sender must approve the transaction *before* the payer and that once `sender_approved` or `payer_approved` become `true` they cannot be re-approved or reverted.
