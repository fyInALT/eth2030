# Sprint 7, Story 7.2 — ValidatePQSignature Bridge

**Sprint goal:** Bridge the PQ algorithm registry to transaction validation.
**Files modified:** `pkg/crypto/pqc/pq_algorithm_registry.go`

## Overview

The PQ algorithm registry at `pkg/crypto/pqc/registry.go` contains 5 algorithms (Dilithium3, Falcon512, SPHINCS+, XMSS, WOTS+) with gas costs, but no integration point existed between the registry and transaction validation.

## Gap (GAP-PQ2)

**Severity:** IMPORTANT
**File:** `pkg/crypto/pqc/registry.go`
**Evidence:** `core/processor.go`'s signature validation path used only secp256k1 ECDSA. No function existed to validate PQ signatures using the registry.

## Implement

```go
// pkg/crypto/pqc/pq_algorithm_registry.go:241
// ValidatePQSignature validates a post-quantum signature using the
// registered algorithm. This bridges the PQ algorithm registry to the
// transaction validation path, enabling PQ-signed transactions.
func ValidatePQSignature(algorithmID uint8, publicKey, message, signature []byte) error {
    reg := DefaultPQAlgorithmRegistry()
    algo, err := reg.GetAlgorithm(algorithmID)
    if err != nil {
        return fmt.Errorf("pqc: unknown algorithm ID %d: %w", algorithmID, err)
    }
    if !algo.Verify(publicKey, message, signature) {
        return fmt.Errorf("pqc: %s signature verification failed", algo.Name)
    }
    return nil
}
```

**Usage from processor.go (future):**

```go
// When PQ tx type 0x07 is introduced:
if tx.Type() == types.PQTxType {
    err := pqc.ValidatePQSignature(tx.PQAlgorithmID(), tx.PQPublicKey(), tx.SigningHash(), tx.PQSignature())
    if err != nil {
        return nil, err
    }
}
```

## Codebase Locations

| File | Line | Purpose |
|------|------|---------|
| `pkg/crypto/pqc/pq_algorithm_registry.go` | 241 | ValidatePQSignature function |
| `pkg/crypto/pqc/registry.go` | 1 | PQ algorithm registry with 5 algorithms |
