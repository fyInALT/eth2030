# prequorum

Pre-quorum voting phase for pre-finality transaction preconfirmations.

## Overview

Package `prequorum` implements the Secure Prequorum engine for pre-finality
confirmations as described in the CL Cryptography track milestone. Validators
submit signed `Preconfirmation` messages binding themselves to the inclusion
of specific transactions in upcoming slots. When enough unique validators
have preconfirmed for a given slot, prequorum is reached, giving users
high-confidence that their transactions will be included before full finality
completes.

`PrequorumEngine` is thread-safe and manages per-slot `slotData` internally.
Each preconfirmation is validated against a commitment hash:
`Keccak256(slot || validatorIndex || txHash)`.

`SecurePrequorumEngine` (in `secure_prequorum.go`) extends the base with
cryptographic signature verification.

## Functionality

### Types

| Name | Description |
|------|-------------|
| `PrequorumConfig` | `QuorumThreshold` (default 0.67), `MaxPreconfsPerSlot`, `ValidatorSetSize` |
| `PrequorumEngine` | Thread-safe per-slot preconfirmation store |
| `Preconfirmation` | Validator preconf: slot, validator index, tx hash, commitment, signature |
| `PrequorumStatus` | Snapshot: total preconfs, unique validators, quorum reached, confidence |

### Functions / methods

| Name | Description |
|------|-------------|
| `DefaultPrequorumConfig() PrequorumConfig` | 67% threshold, 10k max/slot, 1k validator set |
| `NewPrequorumEngine(config) *PrequorumEngine` | Create engine |
| `(*PrequorumEngine).SubmitPreconfirmation(preconf) error` | Validate and store a preconfirmation |
| `(*PrequorumEngine).ValidatePreconfirmation(preconf) error` | Validate without storing |
| `(*PrequorumEngine).CheckPrequorum(slot) *PrequorumStatus` | Check quorum status for a slot |
| `(*PrequorumEngine).GetPreconfirmations(slot) []*Preconfirmation` | Return all preconfs for a slot |
| `(*PrequorumEngine).GetConfirmedTxs(slot) []types.Hash` | Set of preconfirmed tx hashes |
| `(*PrequorumEngine).PurgeSlot(slot)` | Free memory for old slots |
| `ComputeCommitment(slot, validatorIndex, txHash) types.Hash` | Derive the expected commitment hash |

### Errors

`ErrNilPreconfirmation`, `ErrPrequorumInvalidSlot`, `ErrEmptySignature`, `ErrEmptyTxHash`, `ErrEmptyCommitment`, `ErrSlotFull`, `ErrDuplicatePreconf`, `ErrInvalidCommitment`

## Usage

```go
import "github.com/eth2030/eth2030/consensus/prequorum"

engine := prequorum.NewPrequorumEngine(prequorum.DefaultPrequorumConfig())

commitment := prequorum.ComputeCommitment(slot, validatorIndex, txHash)
engine.SubmitPreconfirmation(&prequorum.Preconfirmation{
    Slot: slot, ValidatorIndex: validatorIndex,
    TxHash: txHash, Commitment: commitment, Signature: sig,
})

status := engine.CheckPrequorum(slot)
if status.QuorumReached {
    // Pre-confirm transaction to user.
}
```

[← consensus](../README.md)
