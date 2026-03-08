# fastconfirm

Fast confirmation pre-finality signal for the CL latency track (Glamsterdam).

## Overview

Package `fastconfirm` provides optimistic pre-finality confirmation of blocks.
Validators attest to a block; once a configurable quorum threshold (default
67%) of the total validator set has attested to the same `(slot, blockRoot)`
pair, the block is considered "fast confirmed" — not finalized, but with a
high probability of canonical inclusion. This reduces perceived latency for
users and applications without weakening finality guarantees.

`FastConfirmTracker` is the central thread-safe store. It tracks per-slot
attestation sets, deduplicates votes by `ValidatorIndex`, automatically
confirms when quorum is met, and prunes stale slots to bound memory usage.

## Functionality

### Types

| Name | Description |
|------|-------------|
| `FastConfirmConfig` | Threshold fraction, min attesters, confirm timeout, max tracked slots, total validator count |
| `FastConfirmTracker` | Thread-safe tracker; collects attestations and checks quorum |
| `FastConfirmation` | Result snapshot: slot, block root, attestation count, confirmed flag, timestamp |

### Functions / methods

| Name | Description |
|------|-------------|
| `DefaultFastConfirmConfig() *FastConfirmConfig` | 67% threshold, 64 min attesters, 4 s timeout, 64 tracked slots |
| `NewFastConfirmTracker(cfg) *FastConfirmTracker` | Create a tracker |
| `(*FastConfirmTracker).AddAttestation(slot, blockRoot, attesterIndex) error` | Record an attestation; auto-confirms on quorum |
| `(*FastConfirmTracker).GetConfirmation(slot) (*FastConfirmation, error)` | Retrieve confirmation state for a slot |
| `(*FastConfirmTracker).IsConfirmed(slot, blockRoot) bool` | True if slot+root pair has been fast confirmed |
| `(*FastConfirmTracker).AttestationCount(slot) int` | Number of attestations received for a slot |
| `(*FastConfirmTracker).PruneExpired(now time.Time) int` | Remove expired slot entries |
| `ValidateConfirmation(fc, cfg) error` | Validate a `FastConfirmation` struct |
| `ValidateFastConfirmConfig(cfg) error` | Validate config fields |

### Errors

`ErrFCSlotZero`, `ErrFCBlockRootEmpty`, `ErrFCDuplicateAttester`, `ErrFCSlotExpired`, `ErrFCNotFound`

## Usage

```go
import "github.com/eth2030/eth2030/consensus/fastconfirm"

tracker := fastconfirm.NewFastConfirmTracker(fastconfirm.DefaultFastConfirmConfig())

// Feed incoming attestations.
tracker.AddAttestation(slot, blockRoot, validatorIdx)

// Check if fast-confirmed (non-blocking, no finality guarantee).
if tracker.IsConfirmed(slot, blockRoot) {
    // Signal application layer.
}
```

[← consensus](../README.md)
