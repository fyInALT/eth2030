# Sprint 8, Story 8.1 — STARK Constraint Evaluation

**Sprint goal:** Fix STARK constraints being accepted but never evaluated.
**Files modified:** `pkg/proofs/stark_prover.go`
**Files tested:** `pkg/proofs/stark_prover_test.go`

## Overview

The ethresear.ch proposal requires that STARK proofs actually verify algebraic constraints over the execution trace. Our prover accepted constraints as input and recorded `ConstraintCount` in the proof, but never computed constraint evaluations. A verifier could accept a proof with fabricated trace data because the constraints were never checked.

## Gap (RISK-PQ1 + RISK-STARK1)

**Severity:** MEDIUM
**File:** `pkg/proofs/stark_prover.go` — `GenerateSTARKProof()` at line 124
**Evidence:** `GenerateSTARKProof()` received `constraints []STARKConstraint` but only stored `len(constraints)` in the proof. No constraint polynomial was ever evaluated over the trace. The `VerifySTARKProof()` function never checked that constraint evaluations matched the trace.

**Impact:** A malicious prover could submit a proof over an arbitrary trace without satisfying the constraint system. For the recursive STARK mempool, this means the "proof of tx validity" asserts nothing about actual transaction data.

## ethresear.ch Spec Reference

> Every tick (eg. 500ms), they generate a recursive STARK proving validity of all still-valid objects they know about.

The validity proof must actually constrain the trace — specifically, each row (representing a validated tx) must satisfy the algebraic constraints.

## Implement

### Step 1: Add ConstraintEvalCommitment field to STARKProofData

```go
// pkg/proofs/stark_prover.go
type STARKProofData struct {
    // ... existing fields ...
    ConstraintCount          int
    // Merkle root of the per-row constraint evaluations.
    ConstraintEvalCommitment [32]byte
}
```

### Step 2: Add evaluateConstraints() method

For each trace row, compute `sum(coeff[i] * trace[row][col]^degree) mod fieldModulus` per constraint, hash the concatenated results:

```go
func (sp *STARKProver) evaluateConstraints(trace [][]FieldElement, constraints []STARKConstraint) [][32]byte {
    rowHashes := make([][32]byte, len(trace))
    for r, row := range trace {
        h := sha256.New()
        for _, c := range constraints {
            eval := new(big.Int)
            for i, coeff := range c.Coefficients {
                if i < len(row) && row[i].Value != nil && coeff.Value != nil {
                    term := new(big.Int).Set(row[i].Value)
                    if c.Degree > 1 {
                        term.Exp(term, big.NewInt(int64(c.Degree)), sp.fieldModulus)
                    }
                    term.Mul(term, coeff.Value)
                    term.Mod(term, sp.fieldModulus)
                    eval.Add(eval, term)
                }
            }
            eval.Mod(eval, sp.fieldModulus)
            h.Write(eval.Bytes())
        }
        copy(rowHashes[r][:], h.Sum(nil))
    }
    return rowHashes
}
```

### Step 3: Commit evaluation hashes via Merkle tree

```go
func commitConstraintEvals(evalHashes [][32]byte) [32]byte {
    return merkleRoot(evalHashes)
}
```

### Step 4: Wire into GenerateSTARKProof

```go
evalHashes := sp.evaluateConstraints(trace, constraints)
constraintEvalCommitment := commitConstraintEvals(evalHashes)
// ... include in returned STARKProofData
```

### Step 5: Reject proofs with missing constraint eval in VerifySTARKProof

```go
if proof.ConstraintCount > 0 {
    var zeroCommitment [32]byte
    if proof.ConstraintEvalCommitment == zeroCommitment {
        return false, ErrSTARKVerifyFailed
    }
}
```

### Step 6: Update ProofSize

Add 32 bytes for the `ConstraintEvalCommitment` field.

## Tests

- `TestSTARKConstraintEvaluation` — different traces produce different constraint commitments
- `TestSTARKAggregator_EndToEnd_WithConstraints` — end-to-end: 2-constraint proof verifies, zero commitment is rejected

## Codebase Locations

| File | Line | Purpose |
|------|------|---------|
| `pkg/proofs/stark_prover.go` | 89 | `ConstraintEvalCommitment` field |
| `pkg/proofs/stark_prover.go` | 231 | `evaluateConstraints()` method |
| `pkg/proofs/stark_prover.go` | 257 | `commitConstraintEvals()` function |
| `pkg/proofs/stark_prover.go` | 140 | `GenerateSTARKProof()` — calls eval + commit |
| `pkg/proofs/stark_prover.go` | 197 | `VerifySTARKProof()` — rejects zero commitment |
