# US-4.1 — Minimmit One-Round BFT

**Epic:** EP-4 Finality Protocol
**Total Story Points:** 13
**Sprint:** 4

> **As a** consensus layer developer,
> **I want** to implement the Minimmit one-round BFT protocol as an alternative finality mechanism,
> **so that** blocks can achieve finality in a single message round (proposal + vote → finalize) instead of multi-phase consensus.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

---

## Vitalik's Proposal

> Minimmit is a one-round BFT consensus protocol designed for single-slot finality. Unlike PBFT/HotStuff which require multiple message rounds, Minimmit achieves finality in one round: the proposer broadcasts a block, validators vote, and if 2/3+ stake votes for the same block, it's final. No prepare/commit phases. The trade-off is that conflicting proposals cannot be resolved in the same slot — a missed slot results in a gap.

---

## Tasks

### Task 4.1.1 — Minimmit Protocol State Machine

| Field | Detail |
|-------|--------|
| **Description** | Implement `MinimmitEngine` with a single-round finality state machine. States: `Idle → Proposed → Voting → Finalized/Failed`. On block proposal: transition to `Voting`. On receiving 2/3+ stake votes for same block: transition to `Finalized`. On timeout without quorum: transition to `Failed` (slot missed). No prepare/commit distinction. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Consensus Engineer |
| **Testing Method** | (1) Proposal → Voting state. (2) 2/3+ votes → Finalized. (3) Timeout without quorum → Failed. (4) Conflicting votes (equivocation) detected. (5) State transitions are monotonic (no backwards). |
| **Definition of Done** | Tests pass; state machine correctly handles all transitions; reviewed. |

### Task 4.1.2 — One-Round Vote Aggregation

| Field | Detail |
|-------|--------|
| **Description** | Validators cast a single vote per slot (no committee phases). Vote is `(slot, block_root, validator_index, BLS_signature)`. Integrate with existing `FinalityBLSAdapter` for vote digest computation and signature verification. Aggregate votes using `AggregateVoteSignatures()`. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Consensus Engineer |
| **Testing Method** | (1) Single vote per validator per slot. (2) Duplicate votes rejected. (3) BLS signature verified via `FinalityBLSAdapter.VerifyVote()`. (4) Aggregated signature verified via `FastAggregateVerify`. (5) Equivocating validators detected (two votes, different roots). |
| **Definition of Done** | Tests pass; vote aggregation works with BLS; reviewed. |

### Task 4.1.3 — Finality Proof Generation

| Field | Detail |
|-------|--------|
| **Description** | When 2/3+ stake votes for a block, generate a `MinimmitFinalityProof` containing: aggregate BLS signature, participant bitfield, total voting stake, slot, block root, state root. Use existing `FinalityBLSAdapter.GenerateFinalityProof()` pattern. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Consensus Engineer |
| **Testing Method** | (1) Proof generated when quorum met. (2) Proof contains correct bitfield. (3) Proof verifiable by any node with validator pubkeys. (4) Proof serialization/deserialization round-trips. |
| **Definition of Done** | Tests pass; proofs generated and verifiable; reviewed. |

### Task 4.1.4 — Missed Slot Handling

| Field | Detail |
|-------|--------|
| **Description** | When a slot fails to finalize (no quorum within timeout), the next proposer starts a new slot. Implement skip-slot logic: if slot N has no finality proof, slot N+1 proposer builds on the last finalized block (not on the unfinalized slot N block). Handle the edge case where slot N's block was valid but quorum was not met — the block is discarded, not orphaned. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Consensus Engineer |
| **Testing Method** | (1) Missed slot N → slot N+1 builds on last finalized. (2) Slot N block is not included in chain. (3) No state changes from missed slot persist. (4) Consecutive missed slots handled. |
| **Definition of Done** | Tests pass; missed slot handling is clean; reviewed. |

### Task 4.1.5 — Integration with Existing Finality Infrastructure

| Field | Detail |
|-------|--------|
| **Description** | Wire `MinimmitEngine` as an alternative to `SSFState` and `EndgamePipeline`. Add a `FinalityMode` enum (`{SSF, Endgame, Minimmit}`) to consensus config. The fork-choice rule uses whichever finality mode is active. Ensure PQ attestation fallback (`pq_attestation.go`) works with Minimmit votes. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Consensus Engineer |
| **Testing Method** | (1) Config selects Minimmit mode. (2) Block processing uses Minimmit engine. (3) Fork-choice respects Minimmit finality proofs. (4) PQ attestations work with Minimmit. (5) Switching from SSF to Minimmit at fork boundary. |
| **Definition of Done** | Tests pass; Minimmit integrates with existing consensus; reviewed. |

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/consensus/ssf.go:95-119` | `SSFState` — current SSF implementation with vote tracking, stake accumulation, and finality check. Minimmit replaces this with a simpler one-round state machine. |
| `pkg/consensus/ssf.go:124-158` | `CastVote()` — validates and records votes. Reusable for Minimmit (same vote format). |
| `pkg/consensus/ssf.go:162-177` | `CheckFinality()` — checks 2/3 supermajority. Same threshold applies to Minimmit. |
| `pkg/consensus/ssf_round_engine.go:13-41` | `SSFRoundPhase` enum — 4 phases (Propose, Attest, Aggregate, Finalize). Minimmit has only 2 effective phases (Propose, Vote→Finalize). |
| `pkg/consensus/endgame_pipeline.go:184-235` | `ProcessBlock()` — validate → execute → collect votes → verify BLS → finalize. Minimmit skips aggregation phase. |
| `pkg/consensus/endgame_pipeline.go:240-321` | `AttemptFastFinality()` — accumulates stake, checks quorum, verifies BLS. Core logic reusable for Minimmit. |
| `pkg/consensus/finality_bls_adapter.go:94-107` | `VoteDigest()` — computes `domain || slot || blockRoot` (44 bytes). Same digest format for Minimmit votes. |
| `pkg/consensus/finality_bls_adapter.go:132-152` | `AggregateVoteSignatures()` — BLS aggregation with bitfield. Directly reusable. |
| `pkg/consensus/finality_bls_adapter.go:192-245` | `GenerateFinalityProof()` — generates proof from finalized round. Pattern reusable. |
| `pkg/consensus/finality_bls_adapter.go:250-285` | `VerifyFinalityProof()` — verifies aggregate signature. Directly reusable. |
| `pkg/consensus/block_finalization_engine.go:205-227` | `ProposeBlock()` — starts finalization timer. Reusable for Minimmit proposal. |
| `pkg/consensus/block_finalization_engine.go:230-264` | `ReceiveVote()` — accumulates votes, checks threshold. Core logic reusable. |
| `pkg/consensus/pq_attestation.go:79-131` | `VerifyAttestation()` — PQ + classic verification. Must work with Minimmit vote format. |

---

## Implementation Status

**❌ Not Implemented**

### What Exists
- ✅ SSF with 4-phase state machine (`ssf.go`, `ssf_round_engine.go`) — Minimmit simplifies this to 1 round
- ✅ Endgame pipeline with <500ms finality target (`endgame_pipeline.go`) — similar goal, different mechanism
- ✅ BLS vote signing/verification/aggregation (`finality_bls_adapter.go`) — directly reusable
- ✅ Block finalization engine with stake-weighted voting (`block_finalization_engine.go`) — pattern reusable
- ✅ PQ attestation support with Dilithium + STARK aggregation (`pq_attestation.go`, `stark_sig_aggregation.go`)
- ✅ Quorum checking via integer arithmetic (2/3 threshold) — same for Minimmit

### What's Missing
- ❌ `MinimmitEngine` — no one-round BFT state machine (0 references to "Minimmit" in codebase)
- ❌ No missed-slot handling specific to one-round BFT (current SSF retries, Minimmit skips)
- ❌ No `FinalityMode` enum for selecting between SSF/Endgame/Minimmit
- ❌ No fork-transition logic for switching finality protocols

### Proposed Solution

1. Create `pkg/consensus/minimmit.go` with `MinimmitEngine` implementing a simplified state machine
2. Reuse `FinalityBLSAdapter` for vote digest, signing, verification, and proof generation
3. Reuse `ReceiveVote()` stake accumulation pattern from `block_finalization_engine.go`
4. Add `FinalityMode` to `ConsensusConfig` (or equivalent)
5. Missed slots: proposer builds on `FinalizedSlot` (from `SSFState.FinalizedSlot` or equivalent)

### Key Differences from SSF

| Aspect | SSF (Current) | Minimmit (Proposed) |
|--------|---------------|---------------------|
| Phases | 4 (Propose, Attest, Aggregate, Finalize) | 2 (Propose, Vote→Finalize) |
| Aggregation | Explicit aggregation phase | Receiver-side only |
| Missed slots | Retry/timeout | Skip to next slot |
| Latency | ~6-12s | ~2-4s (shorter voting window) |
| Conflict resolution | Multi-round | None (missed slot) |
| Code reuse | - | Reuses BLS adapter, vote format, quorum check |

---

## Spec Reference

> **Vitalik:**
> Minimmit achieves finality in one round of communication. The proposer sends a block, validators vote, and if 2/3+ of stake votes within the voting window, the block is final. There's no separate prepare/commit phase. The simplicity of the protocol makes it well-suited for very short slot times (2-4s) where multi-round protocols would be too slow.
>
> **Trade-off:** Minimmit cannot recover from equivocation within the same slot. If the proposer equivocates (sends different blocks to different validators), the slot is missed. This is acceptable because: (a) equivocation is slashable, (b) missed slots are rare with honest proposers, (c) the next slot can proceed immediately.
