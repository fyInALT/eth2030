# epbs/slashing — Builder slashing conditions and penalty engine for ePBS

## Overview

Package `slashing` defines the three builder slashing conditions for Enshrined Proposer-Builder Separation (EIP-7732) and provides a thread-safe `SlashingEngine` to evaluate them against bid-payload pairs. A builder is slashable for non-delivery (won the auction but did not reveal a payload by the deadline), invalid payload (revealed payload does not match the committed bid), or equivocation (submitted two conflicting bids for the same slot).

Penalties are expressed as basis-point multipliers of the bid value: 2× for non-delivery, 3× for invalid payload, and 5× for equivocation. Each slashing event produces a `SlashingRecord` with a deterministic evidence hash (Keccak-256 of condition type, bid hash, and builder address) for on-chain auditability.

## Functionality

**SlashingCondition interface**
- `Type() SlashingConditionType`
- `Check(bid *epbs.BuilderBid, payload *epbs.PayloadEnvelope) (violated bool, reason string)`

**Condition implementations**
- `NonDeliverySlashing{DeadlineSlots, CurrentSlot uint64}` — triggered when `payload == nil` and `currentSlot > bid.Slot + deadlineSlots`
- `InvalidPayloadSlashing{}` — triggered on slot mismatch, builder index mismatch, or `bid.BlockHash != payload.PayloadRoot`
- `EquivocationSlashing{Evidence *EquivocationEvidence}` — triggered when `BidA` and `BidB` share slot and builder but have different block hashes

**Penalty configuration**
- `PenaltyMultipliers{NonDelivery=20000 bps, InvalidPayload=30000 bps, Equivocation=50000 bps}` — `DefaultPenaltyMultipliers()`
- `ComputePenalty(condType SlashingConditionType, bidValue uint64, multipliers PenaltyMultipliers) (uint64, error)`
  - `penalty = (bidValue / 10000) * mult + (bidValue % 10000) * mult / 10000`

**SlashingRecord**
- Fields: `BuilderIndex`, `BuilderAddr`, `Slot`, `ConditionType`, `Reason`, `BidValue`, `PenaltyGwei`, `EvidenceHash`, `Timestamp`
- `ComputeEvidenceHash(condType, bid, builderAddr) Hash` — Keccak-256(condType || bidHash || builderAddr)

**SlashingEngine**
- `NewSlashingEngine(multipliers PenaltyMultipliers, maxRecords int) *SlashingEngine`
- `RegisterCondition(cond SlashingCondition)`
- `ConditionCount() int`
- `EvaluateAll(bid *epbs.BuilderBid, payload *epbs.PayloadEnvelope, builderAddr Address) ([]*SlashingRecord, error)` — evaluates all conditions; retains up to `maxRecords`
- `Records() []*SlashingRecord`
- `RecordCount() int`
- `RecordsForBuilder(builderAddr Address) []*SlashingRecord`
- `TotalPenaltyForBuilder(builderAddr Address) uint64`

**Error sentinels**
- `ErrSlashingNilBid`, `ErrSlashingNilEvidence`, `ErrSlashingNoConditions`, `ErrSlashingNilPayload`, `ErrSlashingInvalidPenalty`

Parent package: [`epbs`](../README.md)
