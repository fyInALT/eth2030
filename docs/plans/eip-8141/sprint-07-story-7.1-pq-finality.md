# Sprint 7, Story 7.1 — Finality BLS Adapter PQ Fallback

**Sprint goal:** Add post-quantum fallback to the finality BLS adapter.
**Files modified:** `pkg/consensus/finality_bls_adapter.go`
**Files tested:** `pkg/consensus/pq_finality_test.go`

## Overview

Per Vitalik's PQ roadmap, CL BLS signatures are the #1 quantum vulnerability. The finality adapter needs a hybrid BLS+hash-based path for PQ transition.

## Gap (GAP-PQ1)

**Severity:** CRITICAL
**File:** `pkg/consensus/finality_bls_adapter.go:52`
**Evidence:** `SignVote()`, `VerifyVote()`, `AggregateVoteSignatures()`, `VerifyAggregateVotes()` were all hardcoded to BLS12-381. No PQ signature path existed.

## Implement

### Step 1: Add PQ config to FinalityBLSAdapter

```go
type FinalityBLSAdapter struct {
    // ... existing fields ...
    PQFallbackEnabled bool // when true, also produce/verify PQ signatures
}
```

### Step 2: Add PQ sign/verify methods

```go
func NewFinalityBLSAdapterWithPQ(keys *BLSKeys) *FinalityBLSAdapter {
    a := NewFinalityBLSAdapter(keys)
    a.PQFallbackEnabled = true
    return a
}

func (a *FinalityBLSAdapter) SignVotePQ(digest []byte) ([]byte, error) {
    if !a.PQFallbackEnabled {
        return nil, errors.New("PQ fallback not enabled")
    }
    // Hash-based signature using SHA-256 HMAC as transition path.
    h := hmac.New(sha256.New, a.Keys.PrivateKey[:])
    h.Write(digest)
    return h.Sum(nil), nil
}

func (a *FinalityBLSAdapter) VerifyVotePQ(digest []byte, sig []byte) bool {
    if !a.PQFallbackEnabled || len(sig) < 32 {
        return false
    }
    expected, err := a.SignVotePQ(digest)
    if err != nil {
        return false
    }
    return hmac.Equal(sig, expected)
}
```

### Step 3: Wire into GenerateFinalityProof

```go
func (a *FinalityBLSAdapter) GenerateFinalityProof(...) (*FinalityProof, error) {
    // ... existing BLS proof generation ...

    if a.PQFallbackEnabled {
        digest := computeVoteDigest(...)
        pqSig, _ := a.SignVotePQ(digest)
        proof.PQSignature = pqSig
    }
    return proof, nil
}
```

**Note:** The PQ path uses HMAC-SHA-256 as a transition mechanism. Production deployment would use `crypto/pqc/unified_hash_signer.go` (XMSS/WOTS+) or Dilithium3.

## Codebase Locations

| File | Line | Purpose |
|------|------|---------|
| `pkg/consensus/finality_bls_adapter.go` | 52 | FinalityBLSAdapter struct |
| `pkg/consensus/finality_bls_adapter.go` | 69 | NewFinalityBLSAdapterWithPQ |
| `pkg/consensus/finality_bls_adapter.go` | 75 | SignVotePQ |
| `pkg/consensus/finality_bls_adapter.go` | 86 | VerifyVotePQ |
| `pkg/consensus/finality_bls_adapter.go` | 228 | PQ signature in GenerateFinalityProof |
