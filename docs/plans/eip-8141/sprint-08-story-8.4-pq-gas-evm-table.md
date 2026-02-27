# Sprint 8, Story 8.4 — PQ Gas Constants in EVM Tables

**Sprint goal:** Mirror PQ signature verification gas costs from the algorithm registry into EVM gas tables.
**Files modified:** `pkg/core/vm/gas.go`, `pkg/core/vm/gas_table.go`, `pkg/crypto/pqc/pq_algorithm_registry.go`
**Files tested:** `pkg/core/vm/gas_table_test.go`, `pkg/crypto/pqc/pq_algorithm_registry_test.go`

## Overview

The PQ algorithm registry (`pkg/crypto/pqc/pq_algorithm_registry.go`) defines gas costs for each post-quantum signature scheme (ML-DSA-44: 3500, ML-DSA-65: 4500, etc.). However, these costs only existed in the registry — they were absent from the EVM gas tables in `pkg/core/vm/`. When the EVM needs to charge gas for a PQ signature verification precompile, it has no way to look up the cost without importing the `pqc` package directly.

## Gap (RISK-PQ2)

**Severity:** LOW
**File:** `pkg/core/vm/gas.go` and `pkg/core/vm/gas_table.go`

**Evidence:** The EVM gas constants (gas.go lines 1-101) define costs for all standard opcodes and EIP-specific operations (EIP-4762 Verkle, EIP-7904 Glamsterdam, EIP-1153 transient storage), but have no PQ verification constants. The gas_table.go provides dynamic gas functions but no PQ lookup. The PQ registry's `GasCostMLDSA44 = 3500` etc. are only accessible via the `pqc` package.

**Impact:** The EVM and PQ registry could drift — if someone changes gas costs in one place but not the other, PQ transactions would be incorrectly priced.

## EIP-8141 / EIP-8051 Spec Reference

> EIP-8051 specifies VERIFY_MLDSA at address 0x12 with gas cost 4500.

EIP-8141 frame transactions support arbitrary verification schemes including PQ signatures. The EVM must know the gas cost for each algorithm to charge correctly during VERIFY frame execution.

## Implement

### Step 1: Add PQ gas constants to gas.go

```go
// Post-quantum signature verification gas costs (EIP-8051).
const (
    GasPQVerifyMLDSA44   uint64 = 3500
    GasPQVerifyMLDSA65   uint64 = 4500
    GasPQVerifyMLDSA87   uint64 = 5500
    GasPQVerifyFalcon512 uint64 = 3000
    GasPQVerifySLHDSA    uint64 = 8000
    GasPQVerifyBase      uint64 = 1000
)
```

### Step 2: Add GasPQVerify() lookup function to gas_table.go

```go
func GasPQVerify(algorithmID uint8) uint64 {
    switch algorithmID {
    case 1: return GasPQVerifyMLDSA44
    case 2: return GasPQVerifyMLDSA65
    case 3: return GasPQVerifyMLDSA87
    case 4: return GasPQVerifyFalcon512
    case 5: return GasPQVerifySLHDSA
    default: return GasPQVerifyBase
    }
}
```

### Step 3: Add ValidateGasCostsMatch() to PQ registry

```go
type EVMGasLookup func(algorithmID uint8) uint64

func (r *PQAlgorithmRegistry) ValidateGasCostsMatch(evmGasLookup EVMGasLookup) error {
    r.mu.RLock()
    defer r.mu.RUnlock()
    for algType, desc := range r.algorithms {
        evmGas := evmGasLookup(uint8(algType))
        if evmGas != desc.GasCost {
            return fmt.Errorf("pq_registry: gas mismatch for algorithm %d (%s): registry=%d, evm=%d",
                algType, desc.Name, desc.GasCost, evmGas)
        }
    }
    return nil
}
```

This method accepts any gas lookup function and cross-checks every registered algorithm's gas cost against it. Tests can pass `vm.GasPQVerify` to verify consistency.

## Tests

- `TestGasPQVerify` — each algorithm ID returns correct gas; unknown IDs return base cost
- `TestPQGasCostConsistency` — `ValidateGasCostsMatch` passes with matching lookup
- `TestPQGasCostMismatch` — `ValidateGasCostsMatch` fails with wrong lookup
- `TestPQGasTable_RegistryConsistency` — integration test using hardcoded EVM values

## Codebase Locations

| File | Line | Purpose |
|------|------|---------|
| `pkg/core/vm/gas.go` | 103 | PQ gas constants |
| `pkg/core/vm/gas_table.go` | 908 | `GasPQVerify()` lookup function |
| `pkg/crypto/pqc/pq_algorithm_registry.go` | 300 | `EVMGasLookup` type |
| `pkg/crypto/pqc/pq_algorithm_registry.go` | 303 | `ValidateGasCostsMatch()` method |
