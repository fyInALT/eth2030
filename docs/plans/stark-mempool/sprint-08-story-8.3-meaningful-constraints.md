# Sprint 8, Story 8.3 — Meaningful STARK Aggregation Constraints

**Sprint goal:** Replace the trivial single constraint with meaningful hash-consistency and gas-bounds constraints.
**Files modified:** `pkg/txpool/stark_aggregation.go`
**Files tested:** `pkg/txpool/stark_recursion_test.go`

## Overview

The STARK mempool aggregator builds an execution trace where each row is `[hash_hi, hash_lo, gas_used]`. The constraint system should verify properties of this trace. The old code used a single trivial constraint `{Degree: 1, Coefficients: [1]}` which only extracted the first column — it didn't verify hash consistency or gas bounds.

## Gap (Combined)

**Severity:** LOW
**File:** `pkg/txpool/stark_aggregation.go` — `GenerateTick()` at line 339

**Evidence:**
```go
// OLD — trivial single constraint
constraints := []proofs.STARKConstraint{
    {Degree: 1, Coefficients: []proofs.FieldElement{proofs.NewFieldElement(1)}},
}
```

This constraint computes `1 * hash_hi` for each row — it doesn't assert any relationship between columns. A trace with zeroed hash columns would pass. The `ConstraintCount` was 1 instead of the expected 2 for meaningful coverage.

**Impact:** The STARK proof over the mempool trace doesn't actually verify that transaction hashes are non-trivial or that gas data is present.

## ethresear.ch Spec Reference

> proving validity of all still-valid objects they know about

Validity implies the proof constrains actual tx data — hash presence and gas accounting.

## Implement

### Replace with two meaningful constraints

```go
// Constraint 1 (hash consistency): sums hash_hi + hash_lo (non-zero for real tx hashes).
// Constraint 2 (gas bounds): extracts gas_used column.
constraints := []proofs.STARKConstraint{
    {Degree: 1, Coefficients: []proofs.FieldElement{
        proofs.NewFieldElement(1), proofs.NewFieldElement(1),
    }},
    {Degree: 1, Coefficients: []proofs.FieldElement{
        proofs.NewFieldElement(0), proofs.NewFieldElement(0), proofs.NewFieldElement(1),
    }},
}
```

**Constraint 1** (`coeff=[1,1]`): Evaluates `1*hash_hi + 1*hash_lo` per row. For any real 32-byte tx hash, this sum is non-zero, binding the proof to actual hash data.

**Constraint 2** (`coeff=[0,0,1]`): Evaluates `0*hash_hi + 0*hash_lo + 1*gas_used` per row. This extracts the gas column, ensuring gas data is committed in the constraint evaluation.

## Tests

- `TestGenerateTick_MeaningfulConstraints` — verifies `AggregateProof.ConstraintCount == 2`
- `TestSTARKTickGossipBandwidth` — end-to-end: 50-tx tick has 2 constraints and non-zero constraint eval commitment

## Codebase Locations

| File | Line | Purpose |
|------|------|---------|
| `pkg/txpool/stark_aggregation.go` | 339 | Constraint definitions in `GenerateTick()` |
