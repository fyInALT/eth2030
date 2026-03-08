# slashdetect

Slash detection for proposers (double proposals) and attesters (double votes, surround votes).

## Overview

Package `slashdetect` provides `SlashingDetector`, which implements the beacon
chain phase0 slashing conditions. It maintains two in-memory registries:

- **Block registry**: maps `(proposer, slot)` to a list of seen block roots.
  A second distinct root for the same pair generates `ProposerSlashingEvidence`.
- **Attestation registry**: per-validator list of `AttestationRecord` values.
  Each new attestation is checked for double votes (same target epoch, different
  root) and surround votes (one attestation's span encloses the other's) in
  both directions per the spec.

Evidence is accumulated in internal buffers and returned by `DetectProposerSlashing`
/ `DetectAttesterSlashing` which also clear the buffer. `Peek*` variants
return copies without clearing. Old attestations beyond `AttestationWindow`
(default 256 epochs) are pruned automatically.

## Functionality

### Types

| Name | Description |
|------|-------------|
| `SlashingDetector` | Thread-safe slash detector |
| `SlashingDetectorConfig` | `AttestationWindow uint64` (default 256 epochs) |
| `ProposerSlashingEvidence` | `ProposerIndex`, `Slot`, `Root1`, `Root2` |
| `AttesterSlashingEvidence` | `ValidatorIndex`, `Type` (`"double_vote"` or `"surround_vote"`), two `AttestationRecord` values |
| `BlockRecord` | `ProposerIndex`, `Slot`, `Root` |
| `AttestationRecord` | `ValidatorIndex`, `SourceEpoch`, `TargetEpoch`, `TargetRoot` |

### Functions / methods

| Name | Description |
|------|-------------|
| `DefaultSlashingDetectorConfig() SlashingDetectorConfig` | 256-epoch window |
| `NewSlashingDetector(config) *SlashingDetector` | Create detector |
| `(*SlashingDetector).RegisterBlock(proposer, slot, root)` | Record a block; emit evidence on double proposal |
| `(*SlashingDetector).RegisterAttestation(validator, sourceEpoch, targetEpoch, targetRoot)` | Record an attestation; emit evidence on violation |
| `(*SlashingDetector).DetectProposerSlashing() []*ProposerSlashingEvidence` | Drain proposer evidence buffer |
| `(*SlashingDetector).DetectAttesterSlashing() []*AttesterSlashingEvidence` | Drain attester evidence buffer |
| `(*SlashingDetector).PeekProposerSlashing() []*ProposerSlashingEvidence` | Read without draining |
| `(*SlashingDetector).PeekAttesterSlashing() []*AttesterSlashingEvidence` | Read without draining |
| `(*SlashingDetector).ValidatorsWithAttestations() []ValidatorIndex` | Sorted list of tracked validators |

## Usage

```go
import "github.com/eth2030/eth2030/consensus/slashdetect"

d := slashdetect.NewSlashingDetector(slashdetect.DefaultSlashingDetectorConfig())
d.RegisterBlock(proposer, slot, root)
d.RegisterAttestation(validator, source, target, targetRoot)

for _, ev := range d.DetectProposerSlashing() {
    // submit ev as on-chain proposer slashing
}
```

[← consensus](../README.md)
