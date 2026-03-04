> [← Back to Sprint Index](README.md)

# EPIC 5 — Vitalik Roadmap Gaps

**Goal**: Implement the 7 user stories corresponding to Missing and Different-Approach items from the Vitalik Fast Slots / Fast Finality / Scaling gap analysis.

---

## US-GAP-1: Gas Reservoir Mechanism (EP-1, US-1.1)

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **EVM contract developer**, I want a gas reservoir dimension where state-creation operations (SSTORE zero→nonzero, CREATE) draw from a separate reservoir budget, so that large contract deployments do not exhaust the regular execution gas cap.

**Priority**: Medium | **Story Points**: 13 | **Sprint Target**: Sprint 3

### Tasks

#### Task GAP-1.1 — Add `StateGasReservoir` to `Contract` struct
- **Description**: In `pkg/core/vm/interpreter.go`, add `StateGasReservoir uint64` to the `Contract` struct. Initialize from `ReservoirConfig.InitReservoir()` at frame start. GAS opcode (`pkg/core/vm/instructions.go:490`, `opGas`) must return `Contract.Gas` (execution gas only, not reservoir). Reservoir is separate.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (vm)
- **Testing Method**: Unit test: GAS opcode in a frame with 1000 execution gas + 500 reservoir gas → returns 1000. `go test ./core/vm/... -run TestGasOpcodeExcludesReservoir`.
- **Definition of Done**: `Contract.StateGasReservoir` field added. GAS opcode returns execution gas only. EF state tests (36,126) unaffected.

#### Task GAP-1.2 — CALL forwards reservoir
- **Description**: In `pkg/core/vm/instructions.go:752` `opCall`, when forwarding gas via 63/64 rule, also forward the full `StateGasReservoir` to the sub-call (not scaled by 63/64). After sub-call returns, restore parent reservoir = parent reservoir - (child initial reservoir - child final reservoir).
- **Estimated Effort**: 5 SP
- **Assignee**: Go Engineer (vm)
- **Testing Method**: Unit test: parent has 1000 reservoir, CALLs child → child has 1000 reservoir available. Child spends 300 → parent has 700 reservoir after CALL returns. `go test ./core/vm/... -run TestCallForwardsReservoir`.
- **Definition of Done**: Reservoir forwarding correct. Tests green. No regression in EF state tests.

#### Task GAP-1.3 — SSTORE draws from reservoir for zero→nonzero
- **Description**: In `pkg/core/vm/gas_table.go:234` `SstoreGas`, when the SSTORE changes zero→nonzero (state creation), charge the extra cost to `Contract.StateGasReservoir` first. If reservoir is depleted, charge remainder to `Contract.Gas`. If both depleted, out-of-gas.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (vm)
- **Testing Method**: Unit test: SSTORE zero→nonzero with reservoir 10K, execution gas 5K → reservoir charged first; SSTORE zero→nonzero with reservoir 0 → execution gas charged. EF state tests: still 36,126/36,126 (reservoir does not change gas accounting in test fixtures which don't use reservoir config).
- **Definition of Done**: Charge logic correct. EF tests unaffected. `go test ./core/vm/... -run TestSSTOREReservoir`.

#### Task GAP-1.4 — Glamsterdam repricing: SSTORE zero→nonzero cost increase
- **Description**: In `pkg/core/glamsterdam_repricing.go:27-28`, update SSTORE zero→nonzero cost from 20,000 to 60,000 (per Vitalik's scaling message). Ensure the increase is charged to the reservoir dimension (Task GAP-1.3). Add unit test with the new cost.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (core)
- **Testing Method**: Unit test: SSTORE zero→nonzero at Glamsterdam fork → 60,000 gas charged. Pre-Glamsterdam → 20,000 gas. `go test ./core/... -run TestGlamsterdamSSTORE`.
- **Definition of Done**: Cost updated. Fork gate correct. EF state tests unaffected (EF tests don't depend on Glamsterdam SSTORE cost).

---

## US-GAP-2: SSTORE State Creation as Separate Gas Dimension (EP-1, US-1.2)

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **EVM protocol engineer**, I want SSTORE zero→nonzero operations to charge to the `DimStorage` gas dimension (not to the compute/regular gas counter), so that state creation and regular execution are independently metered and capped.

**Priority**: Medium | **Story Points**: 8 | **Sprint Target**: Sprint 3

### Tasks

#### Task GAP-2.1 — Route SSTORE state-creation to `DimStorage`
- **Description**: In `pkg/core/vm/gas_table.go`, `pkg/core/vm/evm_storage_ops.go`, and `pkg/core/vm/dynamic_gas.go`, when SSTORE changes zero→nonzero, charge the creation premium to `GasDimension.DimStorage` in the `MultidimensionalGasEngine` (via `pkg/core/multidim_gas.go`) rather than to the main gas counter. The base SSTORE cost (5000) still goes to DimCompute.
- **Estimated Effort**: 5 SP
- **Assignee**: Go Engineer (vm + core)
- **Testing Method**: Unit test: SSTORE zero→nonzero → `DimStorage` counter incremented; DimCompute counter unchanged beyond base cost. `go test ./core/vm/... -run TestSSTOREDimensionRouting`.
- **Definition of Done**: Dimension routing correct. `go test ./core/...` green. No regression in EF state tests.

#### Task GAP-2.2 — Per-dimension block cap enforcement
- **Description**: In `pkg/core/multidim_gas.go`, add block-level cap for `DimStorage`: 4M storage gas per block (separate from the 30M compute gas block limit). Block building rejects txs that would exceed the DimStorage cap even if compute gas is available.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (core)
- **Testing Method**: Unit test: fill block to DimStorage cap → next SSTORE zero→nonzero rejected even though DimCompute gas available. `go test ./core/... -run TestDimStorageCap`.
- **Definition of Done**: Cap enforced. Block builder respects cap. `go test ./core/...` green.

---

## US-GAP-3: Random Attester Sampling (EP-2, US-2.1)

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **consensus protocol engineer**, I want a `RandomAttesterSelector` that samples 256–1024 attesters per slot (instead of the full committee), so that attestation bandwidth scales sub-linearly with the total validator count.

**Priority**: Medium | **Story Points**: 13 | **Sprint Target**: Sprint 2

_Note: US-LEAN-7 (Reduced-Committee Fork Choice) has been merged into this story. GAP-3.3 covers fork-choice weight for sampled attesters (previously LEAN-7.2), and GAP-3.4 covers the config flag and devnet validation (previously LEAN-7.3). LEAN-7.1 was a duplicate of GAP-3.1._

### Tasks

#### Task GAP-3.1 — Implement `RandomAttesterSelector`
- **Description**: Create `pkg/consensus/random_attester_selector.go` with `RandomAttesterSelector{SampleSize int}` implementing `SelectAttesters(slot, randao []byte, validators []ValidatorIndex) []ValidatorIndex`. Use Fisher-Yates shuffle, seed = `keccak(slot || randao)`, select first `SampleSize` elements.
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Distribution test: 10K samples of 256 from 10K validators → uniform distribution (each validator selected ~2.56% of the time, within 95% CI). Determinism test. `go test ./consensus/... -run TestRandomAttesterSelector`.
- **Definition of Done**: Correct distribution. Determinism verified. Tests green.

#### Task GAP-3.2 — Committee-less attestation format
- **Description**: In `pkg/consensus/attestation.go`, add `SampledAttestation` type without `CommitteeBits` field (unnecessary when all attesters are explicit). Wire `RandomAttesterSelector` to produce `SampledAttestation` objects. Add aggregate: N `SampledAttestation`s → single `SampledAggregate`.
- **Estimated Effort**: 5 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Round-trip test: create `SampledAttestation`, serialize, deserialize. Aggregate 256 `SampledAttestation`s → `SampledAggregate`. Verify aggregate signature. `go test ./consensus/... -run TestSampledAttestation`.
- **Definition of Done**: Types defined. Serialization correct. Aggregation works. Tests green.

#### Task GAP-3.3 — Fork-choice weight for sampled attesters
- **Description**: In `pkg/consensus/ssf.go`, when processing `SampledAggregate`, scale vote weight by `full_committee_size / sample_size`. Wire `phase_timer.go` 2-phase slot mode (Proposal + Attestation phases only, aggregation phase eliminated when in sampled mode).
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Fork-choice test: 256 sampled attesters, all vote for head → head has full effective weight. `go test ./consensus/... -run TestSampledForkChoiceWeight`.
- **Definition of Done**: Weight scaling correct. 2-phase slot mode working. Tests green.

#### Task GAP-3.4 — Config flag, backward-compatibility, and devnet validation
- **Description**: Add `--attester-sample-size=0|256|512|1024` flag (0 = full committee, default). Attestation format detection: full committee mode uses existing `Attestation` type; sampled mode uses `SampledAttestation`. Both must coexist during transition period. Add Kurtosis devnet config `4s-slots-reduced-committee.yaml` with 256-validator sample, run 100 slots, verify no finality failures.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (consensus + cmd)
- **Testing Method**: Test: node in full committee mode sends attestations → peer in sampled mode accepts and weights correctly. `go test ./consensus/... -run TestAttesterModeInterop`. Devnet: 100 slots, 4s, 256-validator sample — `cast bn` shows advancing block number, CL logs show `finalized_slot` progressing every 3 slots.
- **Definition of Done**: Both modes work. Interop test passes. Devnet test passes with finality achieved consistently.

---

## US-GAP-4: Block-Level Erasure Coding (EP-3, US-3.1)

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **P2P block propagation engineer**, I want execution blocks split into 8 erasure-coded pieces for propagation (requiring only k pieces to reconstruct), so that block propagation time is reduced and validators can begin processing blocks before downloading all pieces.

**Priority**: Medium | **Story Points**: 13 | **Sprint Target**: Sprint 4

### Tasks

#### Task GAP-4.1 — Implement `BlockErasureEncoder`
- **Description**: Create `pkg/das/block_erasure.go` with `BlockErasureEncoder` wrapping the existing `RSEncoderGF256` from `pkg/das/erasure/reed_solomon_encoder.go`. Parameters: 8 data pieces + 8 parity pieces (k=8, n=16). Input: serialized block bytes. Output: 8 `BlockPiece` objects with RS-encoded data.
- **Estimated Effort**: 5 SP
- **Assignee**: Go Engineer (das)
- **Testing Method**: Round-trip test: encode block → decode from any 8 of 16 pieces → identical block. Adversarial test: remove 8 random pieces → reconstruction succeeds. `go test ./das/... -run TestBlockErasureEncoder`.
- **Definition of Done**: Encode/decode correct. Any 8-of-16 reconstruction works. `go test ./das/...` green.

#### Task GAP-4.2 — Implement `BlockPiece` gossip topic
- **Description**: Create `pkg/p2p/block_piece_gossip.go` with gossip topic `block_piece/{slot}/{piece_index}` and `BlockPiece` message type (containing piece index, piece data, RS commitment). Add peer-level routing: each peer is assigned custody of 2 of 16 pieces based on `keccak(peer_id || slot) % 16`.
- **Estimated Effort**: 5 SP
- **Assignee**: P2P Engineer
- **Testing Method**: Simulation: 16 peers, each holding 2 pieces → any peer can request missing pieces and reconstruct block. `go test ./p2p/... -run TestBlockPieceGossip`.
- **Definition of Done**: Topic registered. Piece routing correct. Reconstruction from partial set works.

#### Task GAP-4.3 — Implement `BlockAssemblyManager`
- **Description**: Create `BlockAssemblyManager` in `pkg/das/block_erasure.go`: maintains a per-slot map of received pieces, triggers reconstruction when ≥ k=8 pieces received, integrates with `pkg/engine/block_pipeline.go` `StageIngress`. Once block is assembled, clear piece cache for that slot.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (das + engine)
- **Testing Method**: Integration test: simulate receiving 8 pieces one by one → on 8th piece, assembly triggers and reconstructed block is available. Timeout test: if only 7 pieces received in 2s → fallback to full block download. `go test ./das/... -run TestBlockAssemblyManager`.
- **Definition of Done**: Assembly working. Timeout fallback implemented. Tests green.

---

## US-GAP-5: Minimmit One-Round BFT + 3SF Backoff (EP-4, US-4.1)

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **consensus protocol researcher**, I want a `MinimmitEngine` implementing one-round BFT with the `is_justifiable_slot` backoff algorithm from 3SF-mini integrated, so that ETH2030 has an alternative to the 4-phase SSF when sub-1-second finality is the priority, with the bounded-finality guarantee preserved via the square/oblong slot progression.

**Priority**: Low | **Story Points**: 13 | **Sprint Target**: Sprint 5

### Tasks

#### Task GAP-5.1 — Implement `MinimmitEngine` core
- **Description**: Create `pkg/consensus/minimmit.go` with `MinimmitEngine` implementing one-round BFT: proposer broadcasts block + pre-vote; validators send single combined vote (pre-vote + commit); finality achieved after 2/3 votes received in a single round. Reference: PBFT family, simplified for Ethereum's ~256 validator committee in one-round mode.
- **Estimated Effort**: 8 SP
- **Assignee**: Consensus Engineer (senior)
- **Testing Method**: Simulation: 100 validators, all honest → finality in 1 round. BFT safety: 33 validators equivocate → no finality but no conflicting finals. Liveness: 67 honest validators → eventual finality. `go test ./consensus/... -run TestMinimmitEngine`.
- **Definition of Done**: Safety and liveness tests pass. `MinimmitEngine` satisfies `FinalityEngine` interface. Code reviewed.

#### Task GAP-5.2 — `FinalityMode` enum and engine selection
- **Description**: Add `FinalityMode` enum to `pkg/consensus/config.go`: `FinalityModeSSF` (default, 4-phase), `FinalityModeMinimmit` (one-round). Add CLI flag `--finality-mode=ssf|minimmit`. Wire `FinalityBLSAdapter` to delegate to the selected engine.
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Test: switch finality mode via config → correct engine instantiated. `go test ./consensus/... -run TestFinalityModeSelection`.
- **Definition of Done**: Mode selection working. Default SSF mode unchanged. Tests pass.

#### Task GAP-5.3 — Minimmit + 3SF integration and bounded-finality simulation
- **Description**: Wire `MinimmitEngine` to use the 3SF `is_justifiable_slot()` backoff algorithm from `refs/research/3sf-mini/consensus.py`: only trigger finality vote at slots where `delta <= 5 || is_perfect_square(delta) || is_oblong(delta)`. This prevents premature finality votes during network partitions. Also add `isJustifiableSlot(delta uint64) bool` to `pkg/consensus/ssf.go` replacing the always-on justification trigger. Validate with a 1000-slot simulation: worst-case honest network must achieve finality within `delta ≤ 25` slots (oblong/square progression), validating the bounded-finality guarantee.
- **Estimated Effort**: 2 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Unit test: 100 consecutive slots, verify Minimmit only fires at justifiable slots (delta ≤5, 1,4,9,16,25,2,6,12,20). `go test ./consensus/... -run TestMinimmit3SFBackoff`. Simulation: `pkg/consensus/ssf_simulation_test.go`, 1000 slots, 10 runs — all must achieve finality. `go test ./consensus/... -run TestSSFBackoffBoundedFinality`.
- **Definition of Done**: Backoff algorithm correct. Timing matches `refs/research/3sf-mini/consensus.py` output. Bounded finality simulation passes 10 runs without failures.

---

## US-GAP-7: Production BLS Backend Upgrade

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **mainnet validator operator**, I want the BLS backend to use the `blst` library (supranational/blst) instead of the pure-Go backend, so that aggregate BLS verification reaches production-grade throughput (10K+ aggregates/sec needed for 1M attestations/slot).

**Priority**: P0 | **Story Points**: 10 | **Sprint Target**: Sprint 1

### Tasks

#### Task GAP-7.1 — Integrate `blst` BLS backend
- **Description**: Add `github.com/supranational/blst` to `pkg/go.mod`. Create `pkg/crypto/blst_backend.go` implementing the `BLSBackend` interface: `Sign`, `Verify`, `FastAggregateVerify`, `AggregatePublicKeys`, `G2Aggregate`. Use `blst`'s `P1Affine` (G1) and `P2Affine` (G2) types directly.
- **Estimated Effort**: 5 SP
- **Assignee**: Go Engineer (crypto)
- **Testing Method**: Test all 5 BLS operations against known test vectors from `refs/consensus-specs/tests/bls/`. Run `go test ./crypto/... -run TestBLSTBackend`. Benchmark: `go test -bench=BenchmarkBLSTAggregateVerify ./crypto/`. Expect > 10K ops/sec.
- **Definition of Done**: All BLS test vectors pass. Benchmark > 10K agg-verify/sec. `PureGoBLSBackend` still available as fallback behind config flag.

#### Task GAP-7.2 — Wire `blst` as default BLS backend
- **Description**: In the consensus and engine layers, replace `PureGoBLSBackend` with `BLSTBackend` as the default. Add config flag `--bls-backend=blst|pure-go` (default `blst`). Ensure CGO is properly enabled for the `blst` C library.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (consensus + build)
- **Testing Method**: `go test ./consensus/... -run TestBLSAttestation` — must pass with blst backend. `go build ./...` with CGO enabled.
- **Definition of Done**: blst is default. Build succeeds. Tests pass. CI has CGO enabled.

#### Task GAP-7.3 — Parallel BLS aggregate verify for 1M attestations
- **Description**: Update `pkg/consensus/parallel_bls.go` to use `blst`'s native parallel batch verification `blst.P2AggregateVerify()`. Target: 1M attestations/slot verified in < 500ms using 16 workers. Benchmark against current pure-Go path.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (consensus + crypto)
- **Testing Method**: `go test -bench=BenchmarkParallelBLSAggregateVerify ./consensus/`. 1M-attestation simulation must complete in < 500ms on 16-core machine.
- **Definition of Done**: 1M attestations verified < 500ms. Benchmark archived. CI throughput regression alert if > 20% slower.
