# secretproposer

VRF-based secret proposer election with commit-reveal (L+ roadmap).

## Overview

Package `secretproposer` implements the secret proposer selection mechanism
that prevents MEV-based proposer manipulation. Validators commit to a
proposer secret `LookaheadSlots` (default 32) in advance:

```
CommitHash = Keccak256(validatorIndex || slot || secret)
```

The commitment conceals the proposer's identity. At reveal time, the validator
publishes the secret; if it matches the commitment, the validator is confirmed
as proposer. `ValidateCommitReveal` enforces the timing and hash binding.

`DetermineProposer` provides a RANDAO-based deterministic fallback when no
commitment exists: `hash(slot || randaoMix) mod validatorCount`.

## Functionality

### Types

| Name | Description |
|------|-------------|
| `SecretProposerSelector` | Thread-safe commit-reveal manager keyed by slot |
| `SecretProposerConfig` | `LookaheadSlots` (32), `CommitmentPeriod` (2), `RevealPeriod` (1) |
| `ProposerCommitment` | Slot, validator index, `CommitHash`, revealed secret (after reveal) |

### Functions / methods

| Name | Description |
|------|-------------|
| `DefaultSecretProposerConfig() *SecretProposerConfig` | 32-slot lookahead, 2-slot commitment, 1-slot reveal |
| `NewSecretProposerSelector(config, seed) *SecretProposerSelector` | Create selector |
| `(*SecretProposerSelector).CommitProposer(validatorIndex, slot, secret) (*ProposerCommitment, error)` | Store commitment |
| `(*SecretProposerSelector).RevealProposer(slot, secret) (uint64, error)` | Verify and reveal |
| `(*SecretProposerSelector).IsCommitted(slot) bool` | True if a commitment exists |
| `(*SecretProposerSelector).GetCommitment(slot) *ProposerCommitment` | Retrieve commitment |
| `DetermineProposer(slot, validatorCount, randaoMix) uint64` | RANDAO-based fallback proposer |
| `ValidateCommitReveal(commitment, secret, currentSlot) error` | Validate a commit-reveal pair |
| `ValidateSecretProposerConfig(cfg) error` | Config sanity checks |

### Errors

`ErrSPNoCommitment`, `ErrSPWrongSecret`, `ErrSPAlreadyRevealed`, `ErrSPZeroValidators`

## Usage

```go
import "github.com/eth2030/eth2030/consensus/secretproposer"

sel := secretproposer.NewSecretProposerSelector(
    secretproposer.DefaultSecretProposerConfig(), randaoSeed,
)

// 32 slots ahead: commit.
sel.CommitProposer(validatorIndex, targetSlot, mySecret)

// At proposal time: reveal.
idx, err := sel.RevealProposer(targetSlot, mySecret)
```

[← consensus](../README.md)
