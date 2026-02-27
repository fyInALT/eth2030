# US-1.2 — Frame Transaction Fee Calculation

**Epic:** EP-1 Transaction Type & RLP Encoding
**Total Story Points:** 3
**Sprint:** 1 (Foundations)

> **As a** block builder,
> **I want** `FrameTx` total fee computation (EIP-1559 + EIP-4844 blob fees) to be correct,
> **so that** the block builder can price transactions and refund the payer accurately.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

---

## Tasks

### Task 1.2.1 — `EffectiveGasPrice` and `TotalFee` for FrameTx

| Field | Detail |
|-------|--------|
| **Description** | Implement fee calculation helpers for `FrameTx`: `EffectiveGasTip(baseFee *big.Int) *big.Int` using EIP-1559 capping logic (`min(max_priority_fee_per_gas, max_fee_per_gas - base_fee)`); `TotalFee(baseFee, blobBaseFee *big.Int) *big.Int` = `tx_gas_limit * effective_gas_price + len(blob_versioned_hashes) * GAS_PER_BLOB * blob_base_fee`. Integrate with the existing fee interfaces in `pkg/core/types/`. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | Table-driven unit tests with known baseFee, blob fee, and gas limit values; assert against hand-calculated expected fees. Test blob fee when `blob_versioned_hashes` is empty (must be zero). |
| **Definition of Done** | Tests pass; EIP-1559 capping verified; blob fee zero when no blobs; code reviewed. |

### Task 1.2.2 — `MaxCost` TXPARAM Parameter (0x06)

| Field | Detail |
|-------|--------|
| **Description** | Implement the `max_cost` computation exposed at TXPARAM index `0x06`. The full formula (worst-case cost, `basefee = max_fee_per_gas`, all gas used, blobs at max price) is: `max_cost = tx_gas_limit * max_fee_per_gas + len(blob_versioned_hashes) * GAS_PER_BLOB * max_fee_per_blob_gas`. This includes the intrinsic cost (baked into `tx_gas_limit`) and all blob costs. Used by `opTxParamLoad` in `pkg/core/vm/eip8141_opcodes.go`. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Unit test with no blobs: assert `max_cost = tx_gas_limit * max_fee_per_gas`. (2) Unit test with 2 blobs at known `max_fee_per_blob_gas`: assert blob component is `2 * GAS_PER_BLOB * max_fee_per_blob_gas`. (3) Overflow guard: very large gas limit and fee do not produce silent truncation. |
| **Definition of Done** | All 3 tests pass; blob component zero when no blobs; implementation matches EIP formula including intrinsic cost in gas limit; reviewed. |

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/types/tx_frame.go:55-58` | `gasFeeCap()`, `gasTipCap()` return `MaxFeePerGas` / `MaxPriorityFeePerGas` |
| `pkg/core/vm/eip8141_opcodes.go:52` | `FrameContext.MaxCost` field (used by TXPARAM 0x06, but not calculated) |

## Implementation Status

**⚠️ Partially Implemented**

- ✅ `gasFeeCap()` and `gasTipCap()` return correct fields
- ❌ **Missing:** `EffectiveGasTip(baseFee)` helper not implemented
- ❌ **Missing:** `TotalFee(baseFee, blobBaseFee)` helper not implemented
- ❌ **Missing:** `MaxCost` computation for TXPARAM 0x06 (field exists in `FrameContext` but never calculated)

---

## EIP-8141 Reference Excerpts

### Specification → Gas Accounting (fee formula)

> The total fee is defined as:
>
> ```
> tx_fee = tx_gas_limit * effective_gas_price + blob_fees
> blob_fees = len(blob_versioned_hashes) * GAS_PER_BLOB * blob_base_fee
> ```
>
> The `effective_gas_price` is calculated per EIP-1559 and `blob_fees` is calculated as per EIP-4844.

### Specification → TXPARAM* opcodes (0x06 max cost)

> | `in1` | `in2`       | Return value                         | Size    |
> | ----- | ----------- | ------------------------------------ | ------- |
> | 0x06  | must be 0   | max cost (basefee=max, all gas used, includes blob cost and intrinsic cost) | 32 |
