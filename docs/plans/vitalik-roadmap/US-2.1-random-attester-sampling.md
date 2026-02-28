# US-2.1 — Random Attester Sampling

**Epic:** EP-2 Fast Slots Infrastructure
**Total Story Points:** 13
**Sprint:** 2

> **As a** consensus layer developer,
> **I want** each slot to select 256-1024 random attesters instead of using full committee shuffles,
> **so that** attestation aggregation overhead is eliminated and slot times can decrease below 6 seconds.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

---

## Vitalik's Proposal

> Replace the current aggregation-heavy model (~150K attesters × 128 committees) with random sampling of 256-1024 validators per slot. Individual signatures are published directly (no aggregation). This removes the entire attestation aggregation phase (currently 2s of the 6s slot), enabling slot times as low as 2-4 seconds. Combined with BLS signature aggregation on the receiver side, 256 signatures per slot is manageable.

---

## Tasks

### Task 2.1.1 — Random Attester Selector

| Field | Detail |
|-------|--------|
| **Description** | Implement `RandomAttesterSelector` that selects N validators (configurable 256-1024) per slot using RANDAO-seeded sampling. Uses the existing `ComputeShuffledIndex()` swap-or-not algorithm but samples a fixed-size subset instead of computing full committee assignments. Validators are selected with probability proportional to effective balance. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Consensus Engineer |
| **Testing Method** | (1) Exactly N attesters selected per slot. (2) Deterministic given same RANDAO seed. (3) Different slots produce different sets. (4) Balance-weighted: higher-stake validators selected more often over many slots. (5) No duplicates in selected set. |
| **Definition of Done** | Tests pass; selection is deterministic and balance-weighted; reviewed. |

### Task 2.1.2 — Direct Signature Publishing

| Field | Detail |
|-------|--------|
| **Description** | Selected attesters publish individual BLS signatures directly via gossip (no pre-aggregation). The attestation message is simplified: just `(slot, block_root, signature)` — no committee index needed. Remove committee bits from the gossip message for sampled attesters. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Consensus Engineer |
| **Testing Method** | (1) Attester publishes individual signature on gossip. (2) Message contains no committee index. (3) Receiver can verify BLS signature. (4) Duplicate detection works. |
| **Definition of Done** | Tests pass; individual signatures gossipped; reviewed. |

### Task 2.1.3 — Receiver-Side BLS Aggregation

| Field | Detail |
|-------|--------|
| **Description** | Block proposers and validators aggregate received individual signatures using `FastAggregateVerify`. Integrate with existing `ParallelAggregator` (16 workers, batch size 4096) for efficient verification. Quorum is checked against the sampled set size, not total validator count. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Consensus Engineer |
| **Testing Method** | (1) Proposer aggregates 256 individual signatures. (2) Quorum checked against sampled set (e.g., 171/256 = 2/3). (3) `FastAggregateVerify` succeeds for aggregated signature. (4) Partial aggregation (< quorum) correctly detected. |
| **Definition of Done** | Tests pass; aggregation verified; quorum threshold correct; reviewed. |

### Task 2.1.4 — Phase Timer Update for No-Aggregation Slots

| Field | Detail |
|-------|--------|
| **Description** | With random sampling, the aggregation phase (currently 2s) is eliminated. Update `PhaseTimer` to support 2-phase slots (Proposal + Attestation) when random sampling is active. The freed time can be used for shorter slots or longer attestation windows. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Consensus Engineer |
| **Testing Method** | (1) PhaseTimer with 2 phases: Proposal (2s) + Attestation (4s) = 6s. (2) PhaseTimer with 2 phases: Proposal (1.5s) + Attestation (2.5s) = 4s. (3) Phase events emitted correctly. (4) Backward compatible with 3-phase mode. |
| **Definition of Done** | Tests pass; 2-phase mode works alongside 3-phase; reviewed. |

### Task 2.1.5 — Fork-Choice Integration

| Field | Detail |
|-------|--------|
| **Description** | Update fork-choice rule to weight attestations from randomly sampled attesters. Each sampled attester's vote carries `total_stake / sample_size` weight (or actual balance). Integrate with existing `SSFState.CastVote()` and `FastConfirmTracker.AddAttestation()`. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Consensus Engineer |
| **Testing Method** | (1) 256-attester sample with 2/3 voting → finality. (2) Weight calculation correct. (3) Fork-choice selects block with most sampled attester weight. (4) Existing committee-based attestations still work in parallel. |
| **Definition of Done** | Tests pass; fork-choice respects sampled attestations; reviewed. |

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/consensus/committee_assignment.go:122-168` | `SwapOrNotShuffle()` — 90-round shuffle algorithm used for committee selection. Random attester sampling can reuse this for subset selection. |
| `pkg/consensus/committee_assignment.go:206-265` | `ComputeBeaconCommittees()` — computes all committees per epoch. Random sampling replaces this with a simpler N-element selection. |
| `pkg/consensus/committee_selection.go:39-86` | `ComputeShuffledIndex()` — swap-or-not shuffle. Can be used to sample N indices from validator set. |
| `pkg/consensus/committee_selection.go:191-235` | `ComputeProposerIndex()` — balance-weighted selection via sampling loop. Pattern reusable for balance-weighted attester sampling. |
| `pkg/consensus/committee_rotation.go:175-264` | `ComputeEpochCommittees()` — enforces 128K attester cap, builds committee structure. Random sampling bypasses committee structure entirely. |
| `pkg/consensus/quick_slots.go:177-249` | `GetDuties()` — assigns validator duties per slot using Fisher-Yates shuffle. Must be extended with random attester sampling mode. |
| `pkg/consensus/phase_timer.go:36-44` | Phase durations: 2s proposal + 2s attestation + 2s aggregation. Aggregation phase eliminated with random sampling. |
| `pkg/consensus/attestation.go:73-98` | `CreateAttestation()` — creates attestation with committee bits. Must support committee-less attestation format. |
| `pkg/consensus/attestation_aggregator.go:135-158` | `AddAttestation()` — adds to aggregation pool. With random sampling, individual signatures go directly to proposer. |
| `pkg/consensus/parallel_bls.go:96-185` | `Aggregate()` — parallel batch BLS aggregation with 16 workers. Reusable for receiver-side aggregation of individual signatures. |
| `pkg/consensus/ssf.go:124-158` | `CastVote()` — records vote with stake. Must accept sampled attester votes with appropriate weight. |
| `pkg/consensus/fast_confirm.go:100-143` | `AddAttestation()` — tracks attestation count for fast confirmation. Quorum threshold changes with random sampling. |
| `pkg/consensus/attestation_scaler.go:14-36` | Constants: max buffer 2M, scale thresholds. With 256-1024 attesters, scaling requirements change dramatically. |

---

## Implementation Status

**❌ Not Implemented**

### What Exists
- ✅ Full committee shuffle infrastructure (`SwapOrNotShuffle`, `ComputeShuffledIndex`) — 90-round swap-or-not algorithm
- ✅ Balance-weighted proposer selection (`ComputeProposerIndex`) — sampling loop pattern reusable
- ✅ Parallel BLS aggregation (`ParallelAggregator`) — 16 workers, 4096 batch size
- ✅ Fast confirmation tracker with configurable quorum threshold
- ✅ SSF vote casting with stake accumulation
- ✅ Phase timer with configurable phase count and durations
- ✅ 128K attester cap in committee rotation

### What's Missing
- ❌ `RandomAttesterSelector` — no random subset sampling (only full committee shuffles exist)
- ❌ Committee-less attestation format (all attestations carry committee bits)
- ❌ No 2-phase slot mode (aggregation phase is always present)
- ❌ No weight scaling for sampled attesters in fork-choice
- ❌ No direct signature publishing (all go through aggregation pipeline)

### Proposed Solution

1. Add `RandomAttesterSelector` using `ComputeShuffledIndex()` to pick N indices from validator list
2. Use balance-weighted sampling (similar to `ComputeProposerIndex()` pattern at `committee_selection.go:214`)
3. Add `SampledAttestation` type without committee bits — just `(slot, root, signature, validator_index)`
4. Add 2-phase mode to `PhaseTimer` with `SamplingEnabled` config flag
5. In `CastVote()`, weight sampled attester votes by `total_stake / sample_size`

---

## Spec Reference

> **Vitalik's Proposal:**
> With random sampling, we select ~256-1024 validators per slot. Each publishes their own attestation. No aggregation subnets, no committee index. The proposer collects individual signatures and aggregates them using BLS. Quorum = 2/3 of sampled set. This removes the entire aggregation phase from the slot structure, enabling faster slots.
>
> **Security:** With 256 random attesters and 2/3 threshold (~171), an attacker needs >1/3 of total stake to prevent finality in any given slot. The security properties match full committee attestation because the sampling is random and weighted by stake.
