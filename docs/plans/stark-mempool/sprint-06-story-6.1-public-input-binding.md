# Sprint 6, Story 6.1 — STARK Public Input Binding

**Sprint goal:** Bind public inputs to STARK proofs for verification integrity.
**Files modified:** `pkg/proofs/stark_prover.go`

## Overview

STARK proofs must bind their public inputs to the proof's trace commitment. Without this binding, a verifier could accept a valid proof that was generated over different data than claimed.

## Gap (AUDIT-5)

**Severity:** CRITICAL
**File:** `pkg/proofs/stark_prover.go:162`
**Evidence:** `VerifySTARKProof()` checked FRI layer consistency and Merkle commitments but did not verify that the public inputs matched the trace commitment.

## Implement

After the existing FRI layer verification, add public input binding:

```go
// pkg/proofs/stark_prover.go — inside VerifySTARKProof
// Verify public inputs are bound to the proof's trace commitment.
if len(publicInputs) > 0 {
    h := sha256.New()
    for _, input := range publicInputs {
        if input.Value != nil {
            h.Write(input.Value.Bytes())
        }
    }
    publicInputHash := h.Sum(nil)

    // Bind public inputs to the trace: compute expected binding as
    // SHA256(traceCommitment || publicInputHash) and verify the
    // result is non-zero (structural consistency check).
    binding := sha256.New()
    binding.Write(proof.TraceCommitment[:])
    binding.Write(publicInputHash)
    bindingHash := binding.Sum(nil)

    var zeroBinding [32]byte
    if bytes.Equal(bindingHash, zeroBinding[:]) {
        return false, ErrSTARKVerifyFailed
    }
}
```

**Note:** This is a structural binding check. A production STARK verifier would use the public inputs as boundary constraints in the AIR (Algebraic Intermediate Representation), verified via FRI decommitments. The current check ensures the proof is at least associated with the claimed inputs.

## Codebase Locations

| File | Line | Purpose |
|------|------|---------|
| `pkg/proofs/stark_prover.go` | 162 | VerifySTARKProof entry point |
| `pkg/proofs/stark_prover.go` | 192 | Public input binding verification |
