# US-1.1 — Gas Reservoir Mechanism

**Epic:** EP-1 Multidimensional Gas Reservoir
**Total Story Points:** 13
**Sprint:** 1

> **As a** smart contract developer,
> **I want** the EVM to enforce a separate gas reservoir for state-creation operations,
> **so that** SSTORE zero→nonzero and CREATE operations consume a dedicated budget that cannot be circumvented via GAS/CALL forwarding.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

---

## Vitalik's Proposal

> Separate "state creation gas" from regular gas. The GAS opcode returns regular gas only (excluding reservoir). CALL forwards the entire reservoir alongside regular gas. SSTORE zero→nonzero draws from the reservoir first, then regular gas if reservoir is exhausted. This prevents contracts from avoiding state creation costs by forwarding all gas to child calls.

**Reference:** EIP-8037 spec in `specs/eips/eip-8037.md` lines 98-117:

```
intrinsic_gas = intrinsic_regular_gas + intrinsic_state_gas
execution_gas = tx.gas - intrinsic_gas
regular_gas_budget = TX_MAX_GAS_LIMIT - intrinsic_regular_gas
gas_left = min(regular_gas_budget, execution_gas)
state_gas_reservoir = execution_gas - gas_left
```

---

## Tasks

### Task 1.1.1 — Add Reservoir Field to Contract and EVM

| Field | Detail |
|-------|--------|
| **Description** | Add `StateGasReservoir uint64` field to `Contract` struct and `EVM` struct. The reservoir is initialized from the transaction's state gas budget (`execution_gas - regular_gas_budget`). During execution, `GAS` returns only `contract.Gas` (regular gas), not reservoir. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core EVM Engineer |
| **Testing Method** | (1) GAS opcode returns regular gas only (excludes reservoir). (2) Contract with reservoir>0 has separate tracking. (3) Reservoir initialized correctly from tx gas. |
| **Definition of Done** | Tests pass; GAS opcode behavior verified; reviewed. |

### Task 1.1.2 — Reservoir Forwarding in CALL/CALLCODE/DELEGATECALL

| Field | Detail |
|-------|--------|
| **Description** | When a CALL-family opcode creates a child context, pass the **entire** reservoir to the child (no 63/64 rule for reservoir). Regular gas follows the existing 63/64 rule. On return, any unused reservoir flows back to the parent. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core EVM Engineer |
| **Testing Method** | (1) Child receives full reservoir. (2) Regular gas follows 63/64 rule as before. (3) Unused reservoir returns to parent. (4) Exceptional halt in child zeros reservoir (not returned). |
| **Definition of Done** | Tests pass; CALL behavior verified with reservoir tracking; reviewed. |

### Task 1.1.3 — SSTORE Reservoir Draw

| Field | Detail |
|-------|--------|
| **Description** | When SSTORE detects zero→nonzero (state creation), deduct state creation cost from `StateGasReservoir` first. If reservoir is insufficient, fall back to regular gas. If both are insufficient, out-of-gas. For nonzero→nonzero (reset) and nonzero→zero (clear), charge regular gas only. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core EVM Engineer |
| **Testing Method** | (1) Zero→nonzero with sufficient reservoir: reservoir decreases, regular gas unchanged. (2) Zero→nonzero with insufficient reservoir: spills to regular gas. (3) Zero→nonzero with both insufficient: OOG. (4) Nonzero→nonzero: regular gas only. (5) Nonzero→zero (clear): regular gas + refund. |
| **Definition of Done** | All 5 tests pass; state creation costs correctly drawn from reservoir; reviewed. |

### Task 1.1.4 — CREATE/CREATE2 Reservoir Draw

| Field | Detail |
|-------|--------|
| **Description** | CREATE and CREATE2 also create state (new account). Charge the state creation portion (`GasCreateGlamsterdam = 83,144` per EIP-8037) from reservoir first, then regular gas. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core EVM Engineer |
| **Testing Method** | (1) CREATE with sufficient reservoir: reservoir decreases. (2) CREATE with insufficient reservoir: spills to regular gas. (3) CREATE2 same behavior. |
| **Definition of Done** | Tests pass; CREATE family uses reservoir for state creation costs; reviewed. |

### Task 1.1.5 — Transaction Intrinsic Gas Split

| Field | Detail |
|-------|--------|
| **Description** | Split `IntrinsicGasGlamsterdam()` into regular and state components. State component covers access list storage key creation costs. Compute `state_gas_reservoir` at the transaction boundary in `applyMessage()` before EVM execution begins. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Transaction with access list: state gas deducted from reservoir. (2) Simple transfer: zero state gas. (3) Reservoir correctly propagated to first EVM call. (4) Block gas pool accounting correct. |
| **Definition of Done** | Tests pass; intrinsic gas split is correct; reservoir flows into EVM; reviewed. |

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/vm/instructions.go:490-492` | `opGas()` — returns `contract.Gas` to stack. Must continue returning only regular gas (no change needed if reservoir is a separate field). |
| `pkg/core/vm/instructions.go:752-808` | `opCall()` — creates child context with `evm.callGasTemp`. Must also forward reservoir. |
| `pkg/core/vm/gas_table.go:234-291` | `SstoreGas()` — detects zero→nonzero via `isZero(original)` at line 253. Must draw from reservoir for creation case. |
| `pkg/core/vm/gas_table.go:324-331` | `isZero(val [32]byte)` — helper for zero detection. Used by SSTORE logic. |
| `pkg/core/vm/evm_storage_ops.go:97-162` | `SstoreGasCost()` — high-level SSTORE gas. Checks `isAllZero(original)` at line 119 for slot creation. |
| `pkg/core/vm/dynamic_gas.go:160-229` | `CalcSStoreGas()` — configurable SSTORE. Checks `dgIsZero(original)` at line 186. |
| `pkg/core/vm/gas_table.go:113-131` | EIP-8037 State Creation Gas constants: `CostPerStateByte=662`, `GasCreateGlamsterdam=83,144`, `GasSstoreSetGlamsterdam=24,084` |
| `pkg/core/multidim_gas.go:20-36` | `GasDimension` enum with 5 dimensions (Compute, Storage, Bandwidth, Blob, Witness). Storage dimension exists but is not wired as a reservoir. |
| `pkg/core/multidim_gas.go:128-169` | `DefaultMultidimGasConfig()` — Storage: 5M target, 10M max, elasticity 2. This is the per-block budget; reservoir is per-tx. |
| `pkg/core/glamsterdam_repricing.go:166-191` | `IntrinsicGasGlamsterdam()` — computes total intrinsic gas. Needs split into regular + state components. |
| `pkg/core/vm/gas.go:27-28` | Pre-Glamsterdam SSTORE constants: `GasSstoreSet=20000`, `GasSstoreReset=2900` |

---

## Implementation Status

**❌ Not Implemented**

### What Exists
- ✅ 5-dim gas pricing engine (`multidim_gas.go`) with Storage dimension — but this is block-level base fee pricing, not per-tx reservoir
- ✅ SSTORE zero→nonzero detection in 3 independent gas calculators (`gas_table.go:253`, `evm_storage_ops.go:119`, `dynamic_gas.go:186`)
- ✅ EIP-8037 spec defined in `specs/eips/eip-8037.md` with reservoir semantics
- ✅ State Creation Gas constants (`GasSstoreSetGlamsterdam=24,084`, `GasCreateGlamsterdam=83,144`) at `gas_table.go:113-131`
- ✅ GAS opcode (`instructions.go:490`) returns `contract.Gas` — already returns only the contract's gas (would naturally exclude a separate reservoir field)

### What's Missing
- ❌ `StateGasReservoir` field on `Contract` or `EVM` struct
- ❌ Reservoir initialization from transaction gas budget
- ❌ Reservoir forwarding in CALL/CALLCODE/DELEGATECALL (full pass-through, no 63/64 rule)
- ❌ Reservoir draw in SSTORE zero→nonzero path
- ❌ Reservoir draw in CREATE/CREATE2 path
- ❌ Intrinsic gas split (regular vs. state) at transaction boundary
- ❌ Reservoir zeroing on exceptional halt

### Proposed Solution

1. Add `StateGasReservoir uint64` to `Contract` struct (alongside existing `Gas uint64`)
2. In `applyMessage()` (or equivalent tx entry point), compute reservoir per EIP-8037 formula
3. In `opCall()` at line 774: forward full reservoir to child (set child's `Contract.StateGasReservoir`)
4. In `SstoreGas()` at line 253: when `isZero(original)`, try `contract.StateGasReservoir` first
5. GAS opcode needs no change — already returns `contract.Gas` only

---

## EIP-8037 Reference Excerpts

> **State gas reservoir**: When a transaction specifies gas, the gas is split into regular execution gas and state creation gas. State-intensive operations (SSTORE set, CREATE) first consume from the state reservoir. The GAS opcode MUST return only the regular gas budget. CALL-family opcodes MUST forward the entire state reservoir to child contexts without applying the 63/64 retention rule.
>
> **Rationale**: Without a reservoir, contracts can avoid state creation costs by checking remaining gas and forwarding minimal gas to state-creating calls. The reservoir ensures state creation has a dedicated budget that cannot be gamed.
