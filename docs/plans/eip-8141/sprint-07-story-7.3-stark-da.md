# Sprint 7, Story 7.3 — STARK Commitment Type in DAS

**Sprint goal:** Add PQ-safe commitment type for data availability.
**Files modified:** `pkg/das/types.go`

## Overview

Vitalik's PQ roadmap specifies migrating DA from KZG to STARK-based commitments for quantum safety. This story adds the `STARKCommitment` type alongside `KZGCommitment` in the DAS pipeline.

## Gap (GAP-PQ4)

**Severity:** IMPORTANT
**File:** `pkg/das/types.go`
**Evidence:** DAS used KZG commitments throughout with no STARK commitment path.

## Implement

```go
// pkg/das/types.go
// STARKCommitment represents a STARK-based data availability commitment
// that is post-quantum secure. It replaces KZG commitments for PQ safety.
type STARKCommitment struct {
    Root        [32]byte // Merkle root of the FRI commitment
    ProofSize   uint32   // expected proof size in bytes
    BlowupFactor uint8   // FRI blowup factor (typically 4 or 8)
}
```

**Future integration points:**
- `DASConfig.UseSTARKDA bool` — flag to switch commitment type
- `das/erasure/` — STARK-based Reed-Solomon encoding
- `das/sampling.go` — verify STARK commitments instead of KZG

## Codebase Locations

| File | Purpose |
|------|---------|
| `pkg/das/types.go` | STARKCommitment type definition |
| `pkg/das/sampling.go` | Future: STARK commitment verification |
| `pkg/das/erasure/` | Future: STARK-based erasure coding |
