# US-3.1 — APPROVE Opcode Core Behavior

**Epic:** EP-3 APPROVE Opcode
**Total Story Points:** 12
**Sprint:** 1-2 (Guards in Sprint 1, Scopes in Sprint 2)

> **As an** EVM engineer,
> **I want** the `APPROVE` opcode (`0xaa`) to update transaction-scoped approval state based on `scope`,
> **so that** smart accounts can authorize execution (`sender_approved`) and gas payment (`payer_approved`) in a controlled, non-reentrant manner.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

---

## Tasks

### Task 3.1.1 — APPROVE Scope 0x0: Execution Approval

| Field | Detail |
|-------|--------|
| **Description** | Implement `scope == 0x0` branch in `opApprove` (`pkg/core/vm/eip8141_opcodes.go`): set `FrameContext.SenderApproved = true`. Preconditions in order: (1) `CALLER == frame.target` (universal APPROVE guard from Task 3.1.5), else revert; (2) `CALLER == tx.sender`, else revert — scope 0x0 is only valid when `frame.target` equals `tx.sender`; (3) `SenderApproved` must not already be `true` — approval flags are **monotonic** (revert on re-approval). On success: set `SenderApproved = true` and terminate the frame like `RETURN`, returning `memory[offset:offset+length]` as the frame's return data. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | Unit tests: (1) correct caller where `CALLER == frame.target == tx.sender` → `SenderApproved = true`, frame exits like RETURN; (2) `CALLER != frame.target` → revert (remaining gas returned); (3) `CALLER == frame.target` but `CALLER != tx.sender` → revert; (4) double approval → revert; (5) frame exits like RETURN — return data is `memory[offset:offset+length]`; (6) APPROVE with non-zero `offset`/`length` stack arguments — return data correctly `memory[offset:offset+length]`. |
| **Definition of Done** | All 6 tests pass; `SenderApproved` set exactly once; monotonicity enforced; frame exits correctly; reviewed. |

### Task 3.1.2 — APPROVE Scope 0x1: Payment Approval

| Field | Detail |
|-------|--------|
| **Description** | Implement `scope == 0x1` branch: preconditions: (1) `CALLER == frame.target`, else revert; (2) `PayerApproved` not already set, else revert; (3) `SenderApproved == true`, else revert; (4) `frame.target` has sufficient balance to cover `tx_fee`, else revert. On success: increment `tx.sender` nonce by 1, deduct `tx_fee` from `frame.target` balance, set `PayerApproved = true`, record `payer = frame.target` in receipt context. Exits frame like `RETURN`. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | (1) Valid call with sufficient balance → nonce incremented, balance deducted, `PayerApproved = true`. (2) Insufficient balance → revert, state unchanged. (3) `SenderApproved == false` → revert. (4) Double payment approval → revert. (5) Assert payer address recorded. (6) Exits frame like RETURN. |
| **Definition of Done** | All 6 tests pass; nonce increment atomic; balance deducted correctly; payer recorded; reviewed. |

### Task 3.1.3 — APPROVE Scope 0x2: Combined Execution + Payment Approval

| Field | Detail |
|-------|--------|
| **Description** | Implement `scope == 0x2` branch: sets both `SenderApproved = true` and `PayerApproved = true` atomically. Preconditions: (1) `CALLER == frame.target`, else revert; (2) `CALLER == tx.sender`, else revert; (3) neither flag already set, else revert; (4) sufficient balance, else revert. On success: increment sender nonce, deduct fee, set both flags, record `payer = tx.sender`. Exits frame like `RETURN`. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | (1) Valid combined approval → both flags true, nonce up, balance deducted. (2) `CALLER != tx.sender` → revert. (3) `SenderApproved` already true → revert. (4) `PayerApproved` already true → revert. (5) Insufficient balance → revert. (6) Scope 0x2 from non-sender target → revert. (7) Exits frame like RETURN. |
| **Definition of Done** | All 7 tests pass; atomicity guaranteed; reviewed. |

### Task 3.1.4 — APPROVE Invalid Scope Guard

| Field | Detail |
|-------|--------|
| **Description** | Any `scope` value outside `{0x0, 0x1, 0x2}` must result in an exceptional halt (`ErrInvalidApproveScope`). Implement this guard at the top of `opApprove`. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | Unit test with scopes `0x3`, `0xff`, `0xdeadbeef` — all must result in exceptional halt. |
| **Definition of Done** | Test passes; existing scope tests unaffected; reviewed. |

### Task 3.1.5 — APPROVE Caller-Must-Equal-FrameTarget Invariant

| Field | Detail |
|-------|--------|
| **Description** | Enforce the universal precondition: `APPROVE` can only be called when `CALLER == frame.target`. If `CALLER != frame.target`, the frame **reverts** (remaining gas returned). Note: the EIP Behavior section uses "revert the frame" semantics here. Invalid scope values remain an exceptional halt (Task 3.1.4). Implement as the first check in `opApprove`. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | Unit tests: (1) `APPROVE` from non-target contract → revert (gas returned); (2) APPROVE with < 3 stack items → exceptional halt (standard stack underflow). |
| **Definition of Done** | Test passes; error type correct; reviewed. |

### Task 3.1.6 — APPROVE Permitted in VERIFY Frames Despite STATICCALL

| Field | Detail |
|-------|--------|
| **Description** | Ensure `APPROVE` is the **only** action in a VERIFY frame exempt from STATICCALL restrictions. `APPROVE` modifies **transaction-scoped** variables, not EVM state. `opApprove` must explicitly bypass `vm.EVM.readOnly`. All other state-modifying opcodes must continue to halt when `readOnly == true`. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | (1) VERIFY frame calls `APPROVE(0x0)` → succeeds. (2) VERIFY frame calls `APPROVE(0x2)` → succeeds. (3) VERIFY frame calls `SSTORE` → exceptional halt. (4) APPROVE exits frame, code after it is unreachable. (5) Non-VERIFY frame APPROVE → still subject to scope preconditions. (6) Confirm `opApprove` does not check `interpreter.readOnly`. |
| **Definition of Done** | All 6 tests pass; `opApprove` documented as exempted from `readOnly`; reviewed. |

---

## EIP-8141 Reference Excerpts

### Specification → APPROVE opcode (`0xaa`)

> The `APPROVE` opcode is like `RETURN (0xf3)`. It exits the current context successfully and updates the transaction-scoped approval context based on the `scope` operand. It can only be called when `CALLER == frame.target`, otherwise it results in an exceptional halt.
>
> ##### Stack
>
> | Stack      | Value        |
> | ---------- | ------------ |
> | `top - 0`  | `offset`     |
> | `top - 1`  | `length`     |
> | `top - 2`  | `scope`      |
>
> ##### Scope Operand
>
> The scope operand must be one of the following values:
>
> 1. `0x0`: Approval of execution - the sender contract approves future frames calling on its behalf. Only valid when `frame.target` equals `tx.sender`.
> 2. `0x1`: Approval of payment - the contract approves paying the total gas cost for the transaction.
> 3. `0x2`: Approval of execution and payment - combines both `0x0` and `0x1`.
>
> Any other value results in an exceptional halt.
>
> ##### Behavior
>
> The behavior of `APPROVE` is defined as follows:
>
> - If `APPROVE` is called when `CALLER != frame.target`, revert.
> - For scopes `0`,`1`, and `2`, execute the following:
>     - `0x0`: Set `sender_approved = true`.
>         - If `sender_approved` was already set, revert the frame.
>         - If `CALLER` != `tx.sender`, revert the frame.
>     - `0x1`: Increment the sender's nonce, collect the total gas cost of the transaction from the account, and set `payer_approved = true`.
>         - If `payer_approved` was already set, revert the frame.
>         - If `frame.target` has insufficient balance, revert the frame.
>         - If `sender_approved == false`, revert the frame.
>    - `0x2`: `sender_approved = true`, increment the sender's nonce, collect the total gas cost of the transaction from `frame.target`, and set `payer_approved = true`.
>         - If `sender_approved` or `payer_approved` was already set, revert the frame.
>         - If `CALLER` != `tx.sender`, revert the frame.
>         - If `frame.target` has insufficient balance, revert the frame.

### Rationale → APPROVE calling convention

> Originally `APPROVE` was meant to extend the space of return statuses from 0 and 1 today to 0 to 4. However, this would mean smart accounts deployed today would not be able to modify their contract code to return with a different value at the top level. For this reason, we've chosen behavior above: `APPROVE` terminates the executing frame successfully like `RETURN`, but it actually updates the transaction scoped values `sender_approved` and `payer_approved` during execution. It is still required that only the sender can toggle the `sender_approved` to `true`. Only the `frame.target` can call `APPROVE` generally, because it can allow the transaction pool and other frames to better reason about `VERIFY` mode frames.
