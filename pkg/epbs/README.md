# Package epbs

Enshrined Proposer-Builder Separation (EIP-7732).

## Overview

The `epbs` package implements the Execution Layer side of Enshrined Proposer-Builder Separation as specified in EIP-7732. ePBS decouples execution payload validation from consensus block production by introducing an in-protocol builder role. Builders register on-chain, submit sealed bids committing to a payload hash and a payment value, and then reveal the full execution payload after the proposer selects a winning bid.

This package defines the core EL data structures (`BuilderBid`, `PayloadEnvelope`, `PayloadAttestation`), the validation logic for signed objects, and the epoch-level state transitions for builder payments and escrow. The Payload Timeliness Committee (PTC) attests to whether the winning builder revealed their payload on time, and those attestations feed into the next-epoch payment settlement.

The `epbs` package is consumed by the `engine` package (which exposes the ePBS Engine API methods to the CL) and by consensus layer logic that processes builder withdrawals and slashing.

## Table of Contents

- [Functionality](#functionality)
- [Subpackages](#subpackages)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### Core Types

| Type | Description |
|------|-------------|
| `BuilderBid` | A builder's offer for a slot: block hash commitment, fee recipient, gas limit, value (in Gwei), and BLS public key |
| `SignedBuilderBid` | `BuilderBid` with a 96-byte BLS12-381 signature |
| `PayloadEnvelope` | Metadata accompanying the revealed execution payload: payload root, beacon block root, slot, state root, and KZG commitments |
| `SignedPayloadEnvelope` | `PayloadEnvelope` with a BLS signature |
| `PayloadAttestationData` | PTC member attestation: beacon block root, slot, and `PayloadStatus` |
| `PayloadAttestation` | Aggregated PTC attestation with a 512-bit aggregation bitfield |
| `PayloadAttestationMessage` | Single PTC member attestation message |
| `BuilderIndex` | `uint64` registry index identifying a builder |

### Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `PTC_SIZE` | 512 | Payload Timeliness Committee size (2^9) |
| `MAX_PAYLOAD_ATTESTATIONS` | 4 | Maximum payload attestations per beacon block |
| `MAX_BLOB_COMMITMENTS_PER_BLOCK` | 4096 | Maximum blob KZG commitments per block |
| `PayloadAbsent` | 0 | Builder's payload was not observed |
| `PayloadPresent` | 1 | Builder revealed payload on time |
| `PayloadWithheld` | 2 | Builder withheld the payload |

### Validation

`ValidateBuilderBid(signed *SignedBuilderBid) error` — validates a signed builder bid:
- Block hash and parent block hash must be non-zero
- Bid value must be greater than zero
- Slot must be greater than zero
- BLS signature is verified over the bid hash using `crypto.DefaultBLSBackend().Verify()` (skipped when signature or pubkey is the zero value, for testing)

`ValidatePayloadEnvelope(env *PayloadEnvelope) error` — validates payload root, beacon block root, state root, and slot.

`ValidatePayloadAttestationData(data *PayloadAttestationData) error` — validates beacon block root, slot, and that `PayloadStatus` is one of `PayloadAbsent`, `PayloadPresent`, or `PayloadWithheld`.

`ValidateBidEnvelopeConsistency(bid *BuilderBid, env *PayloadEnvelope) error` — cross-checks that the slot and builder index in a bid match those in the corresponding envelope.

### Bid Hashing

`BuilderBid.BidHash() types.Hash` — computes a deterministic Keccak-256 hash of the bid fields (parent hash, block hash, prevRandao, fee recipient, gas limit, builder index, slot, value) for use as the signing message.

`IsPayloadStatusValid(status uint8) bool` — returns true for `PayloadAbsent`, `PayloadPresent`, or `PayloadWithheld`.

### Builder Epoch State and Payment Processing

`BuilderEpochState` tracks per-epoch builder balances and pending payments:

- `SetBuilderBalance(idx, balance)` / `GetBuilderBalance(idx)` — manage builder escrow balances
- `AddPendingPayment(idx, amount)` — queue a payment from builder to proposer
- `ProcessBuilderPendingPayments(state *BuilderEpochState)` — deducts pending payments from builder balances at epoch boundary; if a payment exceeds the available balance it is capped (builder is drained rather than going negative)

### Auction Sub-Package

The `auction` sub-package maintains a `PayloadAuction` that tracks signed builder bids per slot and selects the winner:

- `SubmitBid(signed *SignedBuilderBid) error` — validates and inserts a bid; bids are stored sorted by value descending
- Bids for the same builder on the same slot are replaced by the higher-value bid

`builder_selfbuild.go` defines `BuilderIndexSelfBuild = math.MaxUint64`, the sentinel value used when the proposer builds the payload themselves (no external builder).

### Slashing

`slashing/` implements builder slashing for equivocation and payload withholding:

- Builders who submit conflicting bids for the same slot can be slashed
- Builders whose `PayloadStatus` is `PayloadWithheld` receive slashing penalties

### Escrow

`escrow/` implements the bid escrow and `BuilderWithdrawalRegistry`:

- 64-epoch exit delay for builder stake withdrawals
- Maximum sweep limit of 16384 entries per epoch
- `DomainProposerPreferences = 0x0D000000` signing domain constant

### Commit-Reveal for Builder Bids

`commit/` implements the encrypted commit-reveal scheme used by builders to submit sealed bids:

- Builders commit to an encrypted bid hash in the early part of the slot
- The full bid parameters are revealed after the CL selects a winning commitment
- Prevents MEV extraction by other builders observing bids before the auction closes

### MEV Burn

`mevburn/` implements the EIP-7732 MEV burn mechanism where a portion of the builder's payment is burned rather than forwarded to the proposer, reducing validator extractable value incentives.

## Subpackages

| Subpackage | Description |
|------------|-------------|
| [`auction/`](./auction/) | Per-slot bid tracking with value-sorted storage and winner selection |
| [`bid/`](./bid/) | Builder bid utilities and helper types |
| [`builder/`](./builder/) | Builder registration and state management |
| [`commit/`](./commit/) | Encrypted commit-reveal scheme for sealed bids |
| [`escrow/`](./escrow/) | Bid escrow, withdrawal registry (64-epoch delay, 16384 sweep limit) |
| [`mevburn/`](./mevburn/) | MEV burn implementation per EIP-7732 |
| [`slashing/`](./slashing/) | Builder slashing for equivocation and payload withholding |

## Usage

```go
import "github.com/eth2030/eth2030/epbs"

// Validate an incoming signed builder bid.
if err := epbs.ValidateBuilderBid(signedBid); err != nil {
    return fmt.Errorf("invalid builder bid: %w", err)
}

// Validate a revealed payload envelope.
if err := epbs.ValidatePayloadEnvelope(&env); err != nil {
    return fmt.Errorf("invalid payload envelope: %w", err)
}

// Ensure the bid and envelope are for the same slot and builder.
if err := epbs.ValidateBidEnvelopeConsistency(&bid, &env); err != nil {
    return fmt.Errorf("bid/envelope mismatch: %w", err)
}

// Validate PTC attestation data.
if err := epbs.ValidatePayloadAttestationData(&attestation.Data); err != nil {
    return fmt.Errorf("invalid PTC attestation: %w", err)
}

// Epoch boundary: process builder pending payments.
state := epbs.NewBuilderEpochState()
state.SetBuilderBalance(builderIdx, currentBalance)
state.AddPendingPayment(builderIdx, proposerPayment)
epbs.ProcessBuilderPendingPayments(state)
newBalance := state.GetBuilderBalance(builderIdx)

// Check if a builder index represents a self-build (no external builder).
import "github.com/eth2030/eth2030/epbs/auction"
if bid.BuilderIndex == auction.BuilderIndexSelfBuild {
    // Proposer builds the payload themselves.
}
```

## Documentation References

- [Design Doc](../../docs/DESIGN.md)
- [EIP Spec Implementation](../../docs/EIP_SPEC_IMPL.md)
- [GAP Analysis](../../docs/GAP_ANALYSIS.md)
- [EIP-7732: Enshrined Proposer-Builder Separation](https://eips.ethereum.org/EIPS/eip-7732)
