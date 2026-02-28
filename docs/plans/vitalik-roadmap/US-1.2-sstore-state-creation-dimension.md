# US-1.2 — SSTORE State Creation Dimension

**Epic:** EP-1 Multidimensional Gas Reservoir
**Total Story Points:** 8
**Sprint:** 1

> **As a** protocol developer,
> **I want** SSTORE zero→nonzero to consume gas from a dedicated "State Creation" dimension in the multidimensional gas pricing engine,
> **so that** state growth has an independent EIP-1559 base fee that adjusts based on actual state creation demand per block.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

---

## Vitalik's Proposal

> SSTORE zero→nonzero should be in a separate gas dimension from execution gas. This means it has its own base fee that adjusts independently via EIP-1559 dynamics. Blocks with heavy state creation see rising state creation fees without affecting compute or calldata pricing.

---

## Tasks

### Task 1.2.1 — Wire SSTORE Set to Storage Dimension

| Field | Detail |
|-------|--------|
| **Description** | When SSTORE detects zero→nonzero (slot creation), charge the cost against `DimStorage` in the multidimensional gas engine, not `DimCompute`. The existing `MultidimGasPool` tracks per-dimension usage; wire the SSTORE set path to increment `DimStorage` usage. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core EVM Engineer |
| **Testing Method** | (1) SSTORE zero→nonzero increments DimStorage usage in block tracking. (2) SSTORE nonzero→nonzero remains in DimCompute. (3) Block gas pool tracks dimensions separately. (4) Base fee adjustment uses per-dimension targets. |
| **Definition of Done** | Tests pass; SSTORE set costs accrue to DimStorage; reviewed. |

### Task 1.2.2 — Wire CREATE to Storage Dimension

| Field | Detail |
|-------|--------|
| **Description** | CREATE/CREATE2 account creation cost (`GasCreateGlamsterdam=83,144`) charges against `DimStorage` dimension. The execution portion of CREATE (code copying, initialization) remains in `DimCompute`. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core EVM Engineer |
| **Testing Method** | (1) CREATE charges state creation portion to DimStorage. (2) Code execution inside CREATE charges DimCompute. (3) Failed CREATE still charges creation gas. |
| **Definition of Done** | Tests pass; CREATE costs split across dimensions; reviewed. |

### Task 1.2.3 — Block Gas Accounting for Storage Dimension

| Field | Detail |
|-------|--------|
| **Description** | Update block building and validation to track `DimStorage` usage separately. The block header must include per-dimension gas used (or at minimum, the storage dimension base fee must adjust per block based on actual state creation in that block). |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Block with heavy SSTORE creates: DimStorage usage high, DimCompute normal. (2) Next block's storage base fee increases. (3) Block with no state creation: DimStorage base fee decreases toward floor. (4) Base fees for other dimensions unaffected. |
| **Definition of Done** | Tests pass; per-dimension base fee adjustment works; reviewed. |

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/multidim_gas.go:20-36` | `GasDimension` enum — `DimStorage(1)` already defined. Currently used for block-level base fee pricing but not wired to specific opcodes. |
| `pkg/core/multidim_gas.go:128-169` | `DefaultMultidimGasConfig()` — Storage dimension config: 5M target, 10M max, elasticity 2, denom 8, min 1 wei. |
| `pkg/core/multidim_gas.go:254-292` | `UpdateBaseFees()` — adjusts all 5 base fees based on block usage. Storage base fee will naturally adjust if DimStorage usage is tracked. |
| `pkg/core/multidim_gas.go:294-333` | `adjustDimBaseFee()` — per-dimension EIP-1559 formula: `newBase = oldBase * (1 + (used - target) / target / denom)`. |
| `pkg/core/multidim_gas.go:360-378` | `TotalGasCost()` — sums `usage[dim] * baseFee[dim]` across all dimensions. |
| `pkg/core/multidim.go:202-241` | `MultidimGasPool` — tracks gas across dimensions during block execution. `SubGas()` at line 217 deducts atomically. |
| `pkg/core/vm/gas_table.go:234-291` | `SstoreGas()` — where zero→nonzero detection happens. At line 253, `isZero(original)` determines slot creation. This is where DimStorage charge should be injected. |
| `pkg/core/vm/gas_table.go:113-131` | EIP-8037 constants — `GasSstoreSetGlamsterdam=24,084`. This is the state creation cost that should accrue to DimStorage. |
| `pkg/core/vm/evm_storage_ops.go:119-121` | `SstoreGasCost()` — checks `isAllZero(original)` for slot creation detection. |
| `pkg/core/glamsterdam_repricing.go:27-28` | Glamsterdam SSTORE costs: Set 5000, Reset 1500 (pre-EIP-8037 values). |

---

## Implementation Status

**⚠️ Partial**

### What Exists
- ✅ `DimStorage` dimension defined in `GasDimension` enum (`multidim_gas.go:21`)
- ✅ Storage dimension config with target/max/elasticity/denom (`multidim_gas.go:145-151`)
- ✅ Per-dimension base fee adjustment via EIP-1559 formula (`multidim_gas.go:294-333`)
- ✅ `MultidimGasPool` for tracking per-dimension usage during block execution (`multidim.go:202-241`)
- ✅ SSTORE zero→nonzero detection in 3 independent calculators
- ✅ EIP-8037 state creation gas constants defined (`gas_table.go:113-131`)

### What's Missing
- ❌ SSTORE zero→nonzero cost not charged to `DimStorage` — all SSTORE costs currently go to a single gas counter
- ❌ CREATE/CREATE2 account creation cost not charged to `DimStorage`
- ❌ No per-dimension gas usage tracking at the EVM execution level (only at block-level `MultidimGasPool`)
- ❌ No mechanism to pass dimension-specific gas from EVM opcode execution to `MultidimGasPool`

### Proposed Solution

1. Add `DimensionUsage [5]uint64` tracking to `Contract` or `EVM` struct
2. In `SstoreGas()` at line 253: when `isZero(original)`, tag the cost as `DimStorage`
3. Propagate dimension usage from EVM execution back to `MultidimGasPool` in block processor
4. `UpdateBaseFees()` at line 254 will automatically adjust storage base fee based on actual DimStorage usage

### Dependencies

- US-1.1 (Gas Reservoir) can be implemented independently — reservoir is per-tx, dimension pricing is per-block. But they complement each other: reservoir ensures per-tx budget, dimension ensures per-block base fee.
