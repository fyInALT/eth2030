# US-6.1 — Per-Frame Gas Isolation

**Epic:** EP-6 Gas Accounting
**Total Story Points:** 5
**Sprint:** 2

> **As a** protocol engineer,
> **I want** each frame to have its own independent gas budget with no cross-frame gas spill,
> **so that** one frame cannot consume gas from another frame's allocation.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

---

## Tasks

### Task 6.1.1 — Per-Frame Gas Limit Enforcement

| Field | Detail |
|-------|--------|
| **Description** | Each frame executes with exactly `frame.gas_limit` gas. Unused gas is NOT carried over. Out-of-gas in one frame does not halt subsequent frames. Total gas pre-charged at tx entry via APPROVE. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Frame A uses 100/200 gas; Frame B starts fresh with its own limit. (2) Frame A exhausts gas → fails; Frame B executes normally. (3) Refund = sum of unused gas across all frames. |
| **Definition of Done** | Tests pass; no gas leak between frames; reviewed. |

### Task 6.1.2 — Calldata Cost for Frames List RLP

| Field | Detail |
|-------|--------|
| **Description** | Implement `CalldataCostFrames(frames []Frame) uint64`: RLP-encode frames list, apply 4 gas/zero byte + 16 gas/non-zero byte. Added to `FRAME_TX_INTRINSIC_COST` in `CalcFrameTxGas`. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) All-zero data → 4 * bytes. (2) All non-zero → 16 * bytes. (3) Mixed → correct weighted sum. (4) Compare with EIP data-efficiency table (134 bytes). |
| **Definition of Done** | Costs match EIP examples; deterministic; reviewed. |

### Task 6.1.3 — Intrinsic Cost: FRAME_TX_INTRINSIC_COST = 15000

| Field | Detail |
|-------|--------|
| **Description** | Ensure `FRAME_TX_INTRINSIC_COST = 15000` is base cost for every frame tx. Verify not double-counted with legacy intrinsic cost. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | Unit test: `CalcFrameTxGas` with zero-data frames and zero gas limits = exactly `15000 + calldata_cost`. |
| **Definition of Done** | Test passes; legacy intrinsic not applied to frame txs; reviewed. |

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/types/tx_frame.go:251-268` | `CalcFrameTxGas` (intrinsic + calldata + sum of frame limits) |
| `pkg/core/types/tx_frame.go:13` | `FrameTxIntrinsicCost = 15000` |

## Implementation Status

**⚠️ Partially Implemented**

- ✅ `CalcFrameTxGas` implemented (intrinsic + calldata + sum of frame limits)
- ✅ `FrameTxIntrinsicCost = 15000` defined
- ❌ **Missing:** Per-frame gas isolation in execution (no executor yet)
- ❌ **Missing:** Overflow detection in `CalcFrameTxGas`

---

## EIP-8141 Reference Excerpts

### Specification → Gas Accounting

> The total gas limit of the transaction is:
>
> ```
> tx_gas_limit = FRAME_TX_INTRINSIC_COST + calldata_cost(rlp(tx.frames)) + sum(frame.gas_limit for all frames)
> ```
>
> Where `calldata_cost` is calculated per standard EVM rules (4 gas per zero byte, 16 gas per non-zero byte).
>
> The total fee is defined as:
>
> ```
> tx_fee = tx_gas_limit * effective_gas_price + blob_fees
> blob_fees = len(blob_versioned_hashes) * GAS_PER_BLOB * blob_base_fee
> ```
>
> The `effective_gas_price` is calculated per EIP-1559 and `blob_fees` is calculated as per EIP-4844.
>
> Each frame has its own `gas_limit` allocation. Unused gas from a frame is **not** available to subsequent frames. After all frames execute, the gas refund is calculated as:
>
> ```
> refund = sum(frame.gas_limit for all frames) - total_gas_used
> ```
>
> This refund is returned to the gas payer (the `target` that called `APPROVE(0x1)` or `APPROVE(0x2)`) and added back to the block gas pool. *Note: This refund mechanism is separate from EIP-3529 storage refunds.*
