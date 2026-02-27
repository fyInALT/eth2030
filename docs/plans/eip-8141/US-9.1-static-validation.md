# US-9.1 — Static Transaction Validation

**Epic:** EP-9 Static Validity Constraints
**Total Story Points:** 3
**Sprint:** 1

> **As a** node operator,
> **I want** statically invalid frame transactions to be rejected before entering the execution pipeline,
> **so that** the node wastes no execution resources on structurally malformed transactions.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

---

## Tasks

### Task 9.1.1 — Implement Static Validity Check Function

| Field | Detail |
|-------|--------|
| **Description** | Implement `ValidateFrameTxStatic(tx *FrameTx) error`: (1) `tx.chain_id < 2^256` (check nil); (2) `tx.nonce < 2^64`; (3) `1 <= len(tx.frames) <= MAX_FRAMES (1000)`; (4) `len(tx.sender) == 20`; (5) for each frame: `frame.mode < 3`; (6) for each frame: `frame.target == nil || len(frame.target) == 20`. Return descriptive error per violation. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | Table-driven: one test per constraint violation. Boundary: `nonce = 2^64 - 1` (valid), `nonce = 2^64` (invalid); `len(frames) = 0` (invalid), `1` (valid), `1000` (valid), `1001` (invalid). |
| **Definition of Done** | All 6 constraint categories tested; descriptive errors; called at tx pool ingress; reviewed. |

### Task 9.1.2 — chain_id Validation

| Field | Detail |
|-------|--------|
| **Description** | Add chain ID matching: `tx.chain_id` must equal network's chain ID. Prevents cross-chain replay. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Correct chain ID → passes. (2) Wrong chain ID → `ErrInvalidChainID`. |
| **Definition of Done** | Test passes; integrated with chain config; reviewed. |

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/types/tx_frame.go:227-248` | `ValidateFrameTx` — static checks: frame count, chain_id sign, mode < 3, target length, blob consistency |

## Implementation Status

**✅ Mostly Implemented**

- ✅ `ValidateFrameTx` checks: frame count (0 and >MAX_FRAMES), chain_id sign, mode < 3, target length, blob field consistency
- ⚠️ **Gap:** Does not check `nonce < 2^64` (Go `uint64` naturally caps this — acceptable)
- ⚠️ **Gap:** Does not check `len(sender) == 20` (`Address` type is always 20 bytes — acceptable)

---

## EIP-8141 Reference Excerpts

### Specification → Constraints

> Some validity constraints can be determined statically. They are outlined below:
>
> ```python
> assert tx.chain_id < 2**256
> assert tx.nonce < 2**64
> assert len(tx.frames) > 0 and len(tx.frames) <= MAX_FRAMES
> assert len(tx.sender) == 20
> assert tx.frames[n].mode < 3
> assert len(tx.frames[n].target) == 20 or tx.frames[n].target is None
> ```
