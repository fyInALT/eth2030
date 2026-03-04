> [← Back to Sprint Index](README.md)

# EPIC 4 — leanConsensus & leanroadmap

**Goal**: Implement the critical features proven by pq-devnets: 4-second slots, leanSig key format alignment, separate aggregator role, Gossipsub V2, and the integrated 3SF+ePBS+FOCIL+PeerDAS slot structure.

---

## US-LEAN-1: 4-Second Slot Configuration

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **consensus protocol engineer**, I want to configure 4-second slots (proven viable by pq-devnet-0), so that ETH2030 can participate in leanroadmap multi-client interop tests using the faster slot timing.

**Priority**: P0 | **Story Points**: 8 | **Sprint Target**: Sprint 1

### Tasks

#### Task LEAN-1.1 — Add 4-second slot config
- **Description**: In `pkg/consensus/quick_slots.go`, add `QuickSlotConfig{SlotDuration: 4*time.Second, SlotsPerEpoch: 4}` alongside the existing 6-second config. Add CLI flag `--slot-duration=4s|6s` (default `6s`). Add fork activation `IsQuick4s(blockNum)`.
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Unit test: 4s config → slot timer fires at 4s; 6s config → 6s. `go test ./consensus/... -run TestQuickSlot4s`.
- **Definition of Done**: Config works. Timer accuracy ± 50ms. Existing 6s tests unchanged.

#### Task LEAN-1.2 — Validate finality timing at 4-second slots
- **Description**: With 4-second slots and 4-slot epochs, 1-epoch finality = 16 seconds. Validate that `pkg/consensus/ssf.go` 4-phase state machine completes within 4s per phase. Add timing assertions in `pkg/consensus/endgame_pipeline.go`.
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Benchmark test: simulate 100 slots at 4s duration, measure SSF completion time per slot. Must be < 3.5s (leaving 500ms buffer).
- **Definition of Done**: `go test -bench=BenchmarkSSF4sSlot ./consensus/` shows < 3.5s p95. No finality failures in simulation.

#### Task LEAN-1.3 — 4s slot devnet validation
- **Description**: Create Kurtosis devnet config `pkg/devnet/kurtosis/configs/4s-slots-test.yaml` with `--slot-duration=4s`. Run for 200 slots. Verify: chain advances, no missed slots > 5%, finality achieved every epoch.
- **Estimated Effort**: 3 SP
- **Assignee**: DevOps Engineer
- **Testing Method**: `./scripts/run-devnet.sh 4s-slots-test`. Check CL logs: `finalized_slot` progresses. Check EL block count matches expected.
- **Definition of Done**: 200-slot devnet test passes. Chain progresses without stalls. Finality achieved. Logs archived.

---

## US-LEAN-2: leanSig Public Key Format Alignment

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **multi-client interoperability engineer**, I want ETH2030's XMSS public key format to use the leanSig 50-byte format (8-element tree root + 5-element randomiser), so that validators using the Rust leanSig client and the Go ETH2030 client can verify each other's attestations.

**Priority**: P0 | **Story Points**: 8 | **Sprint Target**: Sprint 2

### Tasks

#### Task LEAN-2.1 — Implement leanSig 50-byte pubkey serialization
- **Description**: In `pkg/crypto/pqc/unified_hash_signer.go`, add `SerializeLeanSigPubKey() ([]byte, error)` serializing the XMSS public key as 50 bytes: first 40 bytes = 8 root elements (5 bytes each), last 10 bytes = 5 randomiser elements (2 bytes each). Add `DeserializeLeanSigPubKey([]byte) (*XMSSPublicKey, error)`.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (pqc)
- **Testing Method**: Round-trip test: serialize → deserialize → identical key. Cross-check with `refs/leanSig` Rust test vectors (extract from `refs/leanSig/tests/`). `go test ./crypto/pqc/... -run TestLeanSigPubKeyFormat`.
- **Definition of Done**: Serialization produces 50-byte output. Round-trip passes. Test vectors from `refs/leanSig` match.

#### Task LEAN-2.2 — Update `pq_attestation.go` to use leanSig key format
- **Description**: In `pkg/consensus/pq_attestation.go`, when encoding a `PQAttestation` for gossip, use `SerializeLeanSigPubKey()` for the public key field. When decoding incoming attestations, try `DeserializeLeanSigPubKey()` first, fall back to internal format. Log a warning if internal format detected from a peer.
- **Estimated Effort**: 2 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Encode attestation from ETH2030 → decode on simulated leanSig peer using `refs/leanSig` test vectors → verify correct. `go test ./consensus/... -run TestPQAttestationLeanSigFormat`.
- **Definition of Done**: Attestation encodes to leanSig-compatible format. Cross-client test vector passes.

#### Task LEAN-2.3 — Interop test with leanSig test vectors
- **Description**: Add `pkg/consensus/leansig_interop_test.go` loading test vectors from `refs/lean-spec-tests/` (or manually from `refs/leanSig/tests/`) and verifying that ETH2030 produces identical public keys and signatures for the same seed.
- **Estimated Effort**: 3 SP
- **Assignee**: QA Engineer
- **Testing Method**: `go test ./consensus/... -run TestLeanSigInterop`. All test vectors must match.
- **Definition of Done**: All available test vectors pass. Any failures documented with root cause.

---

## US-LEAN-3a: PQ Aggregator Role — Types & Duty Selection

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **consensus protocol engineer**, I want the PQ aggregator role types and duty selection logic defined, so that we can run the pq-devnet-3 configuration which requires this separation.

**Priority**: P1 | **Story Points**: 6 | **Sprint Target**: Sprint 3

### Tasks

#### Task LEAN-3.1 — Define `PQAggregatorRole` interface and types
- **Description**: Create `pkg/consensus/pq_aggregator.go` with: `PQAggregatorRole` type (duty assignment, slot range), `XMSSSignatureBundle` type (validator index + XMSS sig), `AggregateRequest` type (slot + message hash + expected validators). Define `PQAggregator` interface with `CollectSignatures()`, `ProduceAggregate()`, `PropagateAggregate()`.
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Unit test: type construction, validation, serialization round-trip.
- **Definition of Done**: Types defined. Serialization matches leanSpec wire format (from `refs/leanSpec/`). `go test ./consensus/... -run TestPQAggregatorTypes`.

#### Task LEAN-3.2 — Implement aggregator duty selection
- **Description**: In `pkg/consensus/pq_aggregator.go`, implement `SelectAggregators(epoch, beaconState) []PQAggregatorDuty` — deterministically select 1–4 aggregators per slot using `keccak(slot || epoch_randao) % num_validators`. Aggregators are separate from block proposers.
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Unit test: 1000 validators, 100 slots → verify each slot has 1–4 aggregators, proposer ≠ aggregator in all cases. Verify determinism: same seed → same selection.
- **Definition of Done**: `go test ./consensus/... -run TestAggregatorDutySelection` green. Determinism verified.

---

## US-LEAN-3b: PQ Aggregator Role — Collection & Aggregation

**INVEST**: I⚠ N✓ V✓ E✓ S✓ T✓
> **I-note**: Depends on **US-LEAN-3a** for the types and interfaces (can use stubs during parallel development). Within Sprint 3 it is schedulable after LEAN-3a's types land.

**User Story**:
> As a **consensus protocol engineer**, I want the PQ aggregator to collect per-validator XMSS signatures and produce a STARK aggregate, so that the aggregation phase is fully decoupled from block production as required by pq-devnet-3.

**Priority**: P1 | **Story Points**: 8 | **Sprint Target**: Sprint 3

### Tasks

#### Task LEAN-3.3 — Implement `XMSSSignatureBundle` collection protocol
- **Description**: In `pkg/consensus/pq_aggregator.go`, implement `CollectSignatures(ctx, slot, validators) ([]XMSSSignatureBundle, error)`: broadcasts an aggregation request on P2P topic `pq-agg-request/1`, collects XMSS signature bundles from validators via response, waits up to `t=3s` of slot. Wire to `pkg/p2p/gossip_topics.go`.
- **Estimated Effort**: 5 SP
- **Assignee**: Consensus Engineer + P2P Engineer
- **Testing Method**: Simulation test: 10 validators, aggregator collects all 10 bundles within 3s. Test partial collection: 7/10 bundles → aggregator proceeds with available set.
- **Definition of Done**: Collection protocol working. `go test ./consensus/... -run TestAggregatorCollection` green.

#### Task LEAN-3.4 — Aggregate and propagate
- **Description**: In `pkg/consensus/pq_aggregator.go`, implement `ProduceAggregate(bundles []XMSSSignatureBundle) (*STARKSignatureAggregation, error)` by calling `stark_sig_aggregation.go` `STARKSignatureAggregator`. Propagate result on `pq-agg-result/1` P2P topic. Proposer waits for aggregate before building block.
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Integration test: aggregator collects → produces STARK aggregate → proposer uses aggregate in block. Block with valid STARK aggregate accepted by peers.
- **Definition of Done**: `go test ./consensus/... -run TestAggregatorEndToEnd` green. Aggregate size ≤ 500KB (current leanMultisig limit; track reduction separately).

---

## US-LEAN-4: Gossipsub V2.0 Implementation

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **P2P network engineer**, I want ETH2030 to support Gossipsub V2.0 features (improved scoring, opportunistic grafting, message prioritization), so that 1M validator attestation propagation and 4-second slot timing are achievable.

**Priority**: P1 | **Story Points**: 13 | **Sprint Target**: Sprint 3

### Tasks

#### Task LEAN-4.1 — Implement improved score functions
- **Description**: Create `pkg/p2p/gossip_v2.go` with `GossipV2ScoreParams` extending current scoring with: message-delivery rate reward, first-message-delivery bonus, invalid-message penalty with exponential decay. Reference `refs/libp2p/specs/pull/653` for parameters.
- **Estimated Effort**: 5 SP
- **Assignee**: P2P Engineer
- **Testing Method**: Simulation: 100 peers, inject 10% dishonest peers sending duplicate messages → dishonest peers scored down within 5 slots. `go test ./p2p/... -run TestGossipV2Scoring`.
- **Definition of Done**: Scoring simulation passes. Dishonest peers de-scored. Honest peers maintain connectivity.

#### Task LEAN-4.2 — Opportunistic grafting
- **Description**: In `pkg/p2p/gossip_v2.go`, add `OpportunisticGraft(topic, targetMeshSize int)`: when a topic's mesh is under-connected (< `D_low`), opportunistically graft the highest-scoring non-mesh peers. Fires every 60 seconds (gossipsub heartbeat).
- **Estimated Effort**: 3 SP
- **Assignee**: P2P Engineer
- **Testing Method**: Simulation: start with under-connected mesh → after heartbeat → mesh expands to target D. `go test ./p2p/... -run TestOpportunisticGraft`.
- **Definition of Done**: Mesh recovery working. Test passes.

#### Task LEAN-4.3 — Message prioritization by type
- **Description**: In `pkg/p2p/gossip_topics.go`, add priority tiers: HIGH (block proposals, FOCIL ILs), MEDIUM (attestations, aggregations), LOW (mempool ticks). Implement `PrioritizedGossipRouter` that drains HIGH-priority queue before MEDIUM, MEDIUM before LOW.
- **Estimated Effort**: 3 SP
- **Assignee**: P2P Engineer
- **Testing Method**: Simulation: flood LOW-priority messages, verify HIGH-priority messages still propagate within 200ms. `go test ./p2p/... -run TestMessagePrioritization`.
- **Definition of Done**: HIGH-priority messages never blocked by LOW-priority flood. Test passes. P99 HIGH latency < 200ms under load.

#### Task LEAN-4.4 — Generalized gossipsub configuration
- **Description**: Add `GossipParamsByTopic` map in chain config: per-topic D/D_low/D_high/D_score/D_lazy values (different for blocks vs attestations vs cell gossip vs PQ aggregation). Replace hardcoded gossip parameters.
- **Estimated Effort**: 2 SP
- **Assignee**: P2P Engineer
- **Testing Method**: Config test: default params match existing behavior. Custom params apply per-topic.
- **Definition of Done**: Per-topic params configurable. Existing devnet tests still pass.

---

## US-LEAN-5: Rateless Set Reconciliation

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **P2P network engineer**, I want rateless set reconciliation for mempool and attestation sets, so that nodes only transmit the elements their peers are missing (O(difference) communication), improving propagation efficiency at scale.

**Priority**: P1 | **Story Points**: 8 | **Sprint Target**: Sprint 4

### Tasks

#### Task LEAN-5.1 — Implement IBLT (Invertible Bloom Lookup Table)
- **Description**: Create `pkg/p2p/iblt.go` implementing an IBLT per arXiv:2402.02668. IBLT encodes a set into a fixed-size sketch that allows recovery of the symmetric difference between two sets. Operations: `Insert(item)`, `Delete(item)`, `Subtract(other IBLT) IBLT`, `Decode() (inserted, deleted []Item, ok bool)`.
- **Estimated Effort**: 5 SP
- **Assignee**: P2P Engineer
- **Testing Method**: Unit test: set A = {1,2,3,4,5}, set B = {3,4,5,6,7}, diff = {1,2} missing in B, {6,7} missing in A. IBLT subtraction recovers exactly this diff. Test failure rate < 1% for sets up to 1000 elements. `go test ./p2p/... -run TestIBLT`.
- **Definition of Done**: IBLT decode failure rate < 1%. `go test ./p2p/... -run TestIBLT` green. ≥ 80% coverage.

#### Task LEAN-5.2 — Implement `set_reconciliation.go` protocol
- **Description**: Create `pkg/p2p/set_reconciliation.go` with `SetReconciliationProtocol`: peer A sends its IBLT sketch of mempool tx hashes, peer B subtracts its own IBLT, decodes the diff, fetches only the missing txs. Round-trip: 1 IBLT exchange + targeted tx fetches.
- **Estimated Effort**: 3 SP
- **Assignee**: P2P Engineer
- **Testing Method**: Integration test: two nodes with 90% mempool overlap, reconciliation recovers the 10% diff in 1 round-trip. `go test ./p2p/... -run TestSetReconciliation`.
- **Definition of Done**: Reconciliation completes in 1 round-trip for ≤1000-tx mempool diffs. Test green.

---

## US-LEAN-6: 3SF + ePBS + FOCIL + PeerDAS Integrated Slot Structure

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **consensus protocol researcher**, I want an explicit slot structure that integrates 3SF justification, ePBS builder bids, FOCIL inclusion lists, and PeerDAS sampling into a single coherent timeline, so that each sub-protocol's timing constraints are enforced together and tested as a unit.

**Priority**: P1 | **Story Points**: 13 | **Sprint Target**: Sprint 4

### Tasks

#### Task LEAN-6.1 — Define integrated slot timeline
- **Description**: Create `pkg/consensus/lean_slot.go` with `LeanSlotTimeline` specifying the timeline for a 4-second or 6-second slot: `t=0s` (ePBS bid window opens), `t=T*0.25` (FOCIL IL deadline), `t=T*0.5` (ePBS bid deadline), `t=T*0.6` (PeerDAS sampling target), `t=T*0.75` (FOCIL view freeze), `t=T*0.9` (3SF justification vote), `t=T` (slot end). Parameterized by `SlotDuration`.
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Unit test: 4s slot → each phase starts at correct millisecond offset. 6s slot → same. `go test ./consensus/... -run TestLeanSlotTimeline`.
- **Definition of Done**: Timeline struct defined. All existing phase timers (ePBS, FOCIL, PeerDAS) updated to use `LeanSlotTimeline`. Tests pass.

#### Task LEAN-6.2 — Wire all sub-protocols to unified timeline
- **Description**: Refactor `pkg/consensus/phase_timer.go` to use `LeanSlotTimeline` as the single source of timing truth. Update `pkg/epbs/`, `pkg/focil/`, `pkg/das/sampling_scheduler.go`, and `pkg/consensus/ssf.go` to register phase callbacks with the unified timer.
- **Estimated Effort**: 5 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Integration test: run 10 slots with all 4 sub-protocols active. Verify: ePBS bid collected before `t=T*0.5`, FOCIL frozen at `t=T*0.75`, DAS sampling complete before `t=T*0.9`, 3SF vote at `t=T*0.9`. `go test ./consensus/... -run TestIntegratedSlotProtocols`.
- **Definition of Done**: All sub-protocols use unified timer. Timing assertions pass. No regression in individual protocol tests.

#### Task LEAN-6.3 — Integration test: 20-slot run with all protocols
- **Description**: Write a comprehensive integration test in `pkg/consensus/integrated_slot_test.go`: 20 slots, 4s each, ePBS + FOCIL + PeerDAS + 3SF all active. Verify: finality after 3 slots, no IL violations, all blob samples verified, ePBS builder wins auction every slot.
- **Estimated Effort**: 5 SP
- **Assignee**: QA Engineer (consensus)
- **Testing Method**: `go test ./consensus/... -run TestIntegratedSlot20 -timeout 120s`. All 20 slots must produce finalized blocks.
- **Definition of Done**: Test passes consistently (run 5 times, all pass). Test added to CI pipeline.

---

## US-LEAN-8: Exit Queue Flexibility (Minslack)

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **validator operator**, I want faster exits when the exit queue is short (Minslack), so that validators can exit Ethereum within hours rather than waiting for a static multi-week queue period.

**Priority**: P2 | **Story Points**: 5 | **Sprint Target**: Sprint 5

### Tasks

#### Task LEAN-8.1 — Implement Minslack exit queue logic
- **Description**: Create `pkg/consensus/exit_queue.go` with `MinslackExitQueue`: when `current_epoch - exit_epoch < CHURN_LIMIT / 2`, reduce exit delay by 50%. When queue is empty, exit delay = `MIN_VALIDATOR_WITHDRAWABILITY_DELAY` (256 epochs). Maximum delay = current `MAX_SEED_LOOKAHEAD` (4 epochs). Reference: ethresear.ch 2025-04 Minslack post.
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Unit test: empty queue → fast exit (256 epochs). Full queue → standard exit. Half-full queue → 50% reduced delay. `go test ./consensus/... -run TestMinslackExitQueue`.
- **Definition of Done**: Minslack formula correct. Tests green. Edge cases (0 validators, 1 validator) handled.

#### Task LEAN-8.2 — Wire into validator lifecycle
- **Description**: In `pkg/consensus/` validator processing, replace static exit delay with `MinslackExitQueue.ComputeExitDelay(epoch, queueSize, maxChurn)`. Ensure EIP-7251 (max effective balance) compatibility.
- **Estimated Effort**: 2 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Integration test: submit 10 simultaneous exit requests, verify queue processes them in Minslack order. EIP-7251 max EB still respected.
- **Definition of Done**: `go test ./consensus/... -run TestValidatorExitMinslack` green. EIP-7251 test cases unaffected.
