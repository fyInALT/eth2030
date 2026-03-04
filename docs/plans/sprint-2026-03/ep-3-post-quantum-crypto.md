> [← Back to Sprint Index](README.md)

# EPIC 3 — Post-Quantum Cryptography

**Goal**: Complete the PQ cryptography stack — Lean Available Chain mode, NTT precompile alignment, STARK mempool wiring — so ETH2030 is ready for the PQ-available chain milestone. BLAKE3 backend is implemented in US-BL-1 (EP-2).

---

## US-PQ-2: Lean Available Chain Mode

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **validator node operator**, I want a "Lean Available Chain" mode where only 256–1024 validators per slot use hash-based signatures without STARK aggregation, so that the transition to PQ attestations can begin before full STARK aggregation infrastructure is production-ready.

**Priority**: P1 | **Story Points**: 8 | **Sprint Target**: Sprint 2

### Tasks

#### Task PQ-2.1 — Add `LeanAvailableChainMode` config flag
- **Description**: Add `LeanAvailableChainMode bool` and `LeanAvailableChainValidators int` (default 512, range 256–1024) to `pkg/consensus/config.go`. Add CLI flag `--lean-available-chain` and `--lean-available-validators=512`.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (consensus + cmd)
- **Testing Method**: Unit test: config parse with and without flag. Verify default values.
- **Definition of Done**: Config struct updated. Flag in CLI help. No regression.

#### Task PQ-2.2 — Wire mode into `pq_attestation.go`
- **Description**: In `pkg/consensus/pq_attestation.go`, when `LeanAvailableChainMode` is enabled, limit PQ attestors to the configured `LeanAvailableChainValidators` count using random subset sampling (simple Fisher-Yates shuffle on validator indices, seed = `keccak(slot || epoch_seed)`). Remaining validators continue using BLS.
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Unit test: with 1000 validators and `LeanAvailableValidators=512`, verify exactly 512 are selected as PQ attestors per slot. Verify deterministic selection (same seed → same set).
- **Definition of Done**: `go test ./consensus/... -run TestLeanAvailableChainMode` green. No STARK prover called in lean mode.

#### Task PQ-2.3 — Skip STARK aggregation in lean mode
- **Description**: In `pkg/consensus/stark_sig_aggregation.go`, when `LeanAvailableChainMode` is enabled, skip STARK proof generation and instead use direct Merkle tree of WOTS+ public keys to aggregate (O(N log N) verification, acceptable for ≤1024 validators). Wire mode check before calling `proofs.NewSTARKProver()`.
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Integration test: 512-validator lean mode → no STARK proof generated; full mode → STARK proof generated. Verify aggregate verification succeeds in both cases.
- **Definition of Done**: `go test ./consensus/... -run TestLeanAvailableAggregation` green. STARK prover not invoked in lean mode.

---

## US-PQ-3: NTT Precompile Address Alignment (Fork-Breaking Fix)

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **multi-client interoperability engineer**, I want the NTT precompile addresses to match the ntt-eip spec (`0x0f–0x12`) instead of the current ETH2030 address (`0x15`), so that Falcon on-chain verification contracts deployed on other clients can also run on ETH2030 without modification.

**Priority**: P0 | **Story Points**: 13 | **Sprint Target**: Sprint 1

### Tasks

#### Task PQ-3.1 — Split NTT precompile into 4 separate addresses
- **Description**: Refactor `pkg/core/vm/precompile_ntt.go` from a single dispatch precompile at `0x15` into 4 separate precompile registrations: `NTT_FW=0x0f` (forward NTT, 600 gas flat), `NTT_INV=0x10` (inverse NTT, 600 gas flat), `NTT_VECMULMOD=0x11` (element-wise mul, `k*log₂(n)/8` gas), `NTT_VECADDMOD=0x12` (element-wise add, `k*log₂(n)/32` gas). Remove `0x15` registration.
- **Estimated Effort**: 5 SP
- **Assignee**: Go Engineer (vm)
- **Testing Method**: Unit test each of the 4 addresses: correct operation, correct gas deduction. Cross-check with ntt-eip spec gas formulas. EF state tests: 36,126/36,126 still pass (NTT precompile wasn't active in EF tests).
- **Definition of Done**: 4 precompiles registered at correct addresses. `go test ./core/vm/... -run TestNTTPrecompile` green. Old `0x15` address removed from code.

#### Task PQ-3.2 — Add vector dot-product and butterfly ops
- **Description**: Add `NTT_DOTPRODUCT=0x13` (inner product mod q) and `NTT_BUTTERFLY=0x14` (Cooley-Tukey bit-reversal permutation) as additional addresses per Vitalik's suggestion. These are the missing ops for Dilithium/Falcon fast verification in EVM (see GAP 3.3 in PQ doc).
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (vm + pqc)
- **Testing Method**: Unit test: dot product of `[1,2,3]` and `[4,5,6]` mod `12289` = `32`. Butterfly permutation of `[0,1,2,3,4,5,6,7]` = `[0,4,2,6,1,5,3,7]`. Gas formula verified.
- **Definition of Done**: Tests pass. Gas benchmarks documented.

#### Task PQ-3.3 — Update chain config fork registration
- **Description**: Move NTT precompile activation from I+ fork to a configurable `NTTPrecompileFork`. Update chain config `glamsterdam.go`/`hegota.go` as appropriate. Ensure the transition from `0x15` to `0x0f–0x12` is gated at a specific fork to avoid breaking existing test deployments.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (core + cmd)
- **Testing Method**: Fork switch test: before fork → `0x15` active; at fork → `0x0f–0x12` active; `0x15` returns `0x0` (not-found) after fork.
- **Definition of Done**: Fork gate correct. `go test ./core/... -run TestNTTForkActivation` green. CLAUDE.md EIP table updated.

#### Task PQ-3.4 — Integration test: Falcon on-chain verify via NTT precompile
- **Description**: End-to-end test: sign a message with Falcon-512 (`pkg/crypto/pqc/falcon_signer.go`), submit a VERIFY frame tx that calls the NTT precompile addresses `0x0f–0x12` to verify the signature, confirm APPROVE is called. Measure total gas.
- **Estimated Effort**: 2 SP
- **Assignee**: QA Engineer (security + vm)
- **Testing Method**: `go test ./core/vm/... -run TestFalconNTTOnChain`. Verify gas ≤ 2M (EPERVIER-equivalent without precompile is 1.5M; with precompile target is ~1500 gas for NTT ops).
- **Definition of Done**: Test green. Gas report documented. CVE check from US-AA-5.1 must be complete before this test is deployed to testnet.

---

## US-PQ-4: Mempool STARK Ticks P2P Wiring

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **full node operator**, I want mempool STARK aggregation ticks (generated every 500ms) to be propagated to peers via P2P gossip, so that the bandwidth-efficient mempool from Vitalik's ethresear.ch proposal is fully operational.

**Priority**: P1 | **Story Points**: 8 | **Sprint Target**: Sprint 2

### Tasks

#### Task PQ-4.1 — Add `mempool-stark-tick/1` gossip topic
- **Description**: In `pkg/p2p/gossip_topics.go`, register topic `mempool-stark-tick/1` with message type `MempoolAggregationTick`. Add `GossipMempoolStarkTick(tick *MempoolAggregationTick) error` method. Score function: penalize nodes that send ticks > 2× per interval or with invalid proofs.
- **Estimated Effort**: 3 SP
- **Assignee**: P2P Engineer
- **Testing Method**: Unit test: publish tick → subscriber receives it. Scoring: invalid proof → peer score decremented. `go test ./p2p/... -run TestMempoolStarkTopic`.
- **Definition of Done**: Topic registered. Subscription works. Tests pass.

#### Task PQ-4.2 — Wire `stark_aggregation.go` output to gossip
- **Description**: In `pkg/txpool/stark_aggregation.go`, after each 500ms tick generation, call `p2p.GossipMempoolStarkTick(tick)`. Wire `MempoolSTARKAggregator` to accept a `P2PBroadcaster` interface (to avoid circular imports). Add `MaxTickSize = 128KB` enforcement before broadcast.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (txpool + p2p)
- **Testing Method**: Integration test: two nodes connected, STARK tick generated on node A → received on node B. Tick size ≤ 128KB verified. `go test ./txpool/... -run TestStarkTickPropagation`.
- **Definition of Done**: Tests pass. Tick propagation visible in devnet logs.

#### Task PQ-4.3 — Ingest peer ticks into local validity cache
- **Description**: On receiving a `MempoolAggregationTick` from a peer: (a) verify the STARK proof in the tick; (b) for each tx in the tick, mark it as "STARK-validated by peer"; (c) reduce local validation cost for these txs. Cache TTL = 2 slots.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (txpool)
- **Testing Method**: Unit test: receive tick with 5 txs → all 5 marked as peer-validated. Cache eviction after 2 slots.
- **Definition of Done**: Cache working. `go test ./txpool/... -run TestPeerTickCache` green.

---

## US-PQ-5a: EVM Mini-Trace Validation Frame Circuit

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓
> **S-note**: Original US-PQ-5 was 21 SP across 3 tasks and violated the Small criterion. Split into US-PQ-5a (8 SP, circuit) + US-PQ-5b (13 SP, replacer + wiring). US-PQ-5b depends on US-PQ-5a's circuit interface but can be started with a stub.

**User Story**:
> As a **ZK proof system developer**, I want a STARK circuit that proves the validity of a single EVM VERIFY frame execution (calldata in → non-zero return), so that block builders have a concrete prover to call when stripping validation frames from blocks.

**Priority**: P2 | **Story Points**: 8 | **Sprint Target**: Sprint 4

### Acceptance Criteria
- `ProveValidationFrame(frameCalldata, output []byte)` returns a valid `STARKProofData`
- `ProveAllValidationFrames(frames [][]byte)` batches up to 100 frames into a single proof
- Proof size for a batch of 100 frames ≤ 128 KB
- Invalid calldata (reverts) causes proof generation to return an error, not a false proof

### Tasks

#### Task PQ-5a.1 — Implement `validation_frame_circuit.go`
- **Description**: Create `pkg/proofs/validation_frame_circuit.go`. Define the EVM trace constraint system for VERIFY frame execution: a mini-EVM over the Goldilocks field with constraints on CALL stack depth, APPROVE opcode presence, and return value non-zero. Expose `ProveValidationFrame` and `ProveAllValidationFrames`.
- **Estimated Effort**: 5 SP
- **Assignee**: ZK Engineer (proofs)
- **Testing Method**: Circuit test: valid VERIFY calldata → proof generated and verified. Invalid calldata (reverts) → error returned. Batch test: 1, 10, 100 frames. `go test ./proofs/... -run TestValidationFrameCircuit`. Proof size assertion: ≤ 128 KB for batch of 100.
- **Definition of Done**: Circuit produces valid proof for correct frames. Proof size ≤ 128 KB. Invalid frames never produce a valid proof. Code reviewed. ≥ 80% coverage.

#### Task PQ-5a.2 — Expose stub interface for concurrent US-PQ-5b development
- **Description**: Define `type ValidationFrameProver interface { ProveAllValidationFrames(frames [][]byte) (*STARKProofData, error); Verify(*STARKProofData) bool }` in `pkg/proofs/`. Add `StubValidationFrameProver` that returns a fixed-size dummy proof for use by US-PQ-5b during development.
- **Estimated Effort**: 1 SP
- **Assignee**: ZK Engineer (proofs)
- **Testing Method**: Unit test: stub prover returns a proof, real prover returns a real proof. Both satisfy the interface.
- **Definition of Done**: Interface and stub defined. US-PQ-5b can compile and test against the stub.

#### Task PQ-5a.3 — Circuit benchmarks and proof size report
- **Description**: Add `BenchmarkValidationFrameCircuit` in `pkg/proofs/validation_frame_bench_test.go`: measure prove time and proof size for 1, 10, 50, 100 frames. Document results in `docs/plans/stark-frame-circuit-perf-2026.md`.
- **Estimated Effort**: 2 SP
- **Assignee**: QA Engineer
- **Testing Method**: `go test -bench=BenchmarkValidationFrameCircuit ./proofs/`. Results logged to CI artifact.
- **Definition of Done**: Benchmark documented. Proof size ≤ 128 KB for 100 frames confirmed.

---

## US-PQ-5b: Frame STARK Replacer & Block Sealing Integration

**INVEST**: I⚠ N✓ V✓ E✓ S✓ T✓
> **I-note**: Depends on **US-PQ-5a** providing the `ValidationFrameProver` interface (can use stub during development). **S-note**: This story is 13 SP, within the sprint limit.

**User Story**:
> As a **block builder**, I want to replace all VERIFY frames in a block with a single STARK proof of their collective validity, so that validation frame calldata does not consume block bandwidth (per the recursive STARK mempool proposal from ethresear.ch/t/recursive-stark-based-bandwidth-efficient-mempool/23838).

**Priority**: P2 | **Story Points**: 13 | **Sprint Target**: Sprint 4

### Acceptance Criteria
- `ReplaceValidationFrames(block, prover)` returns a stripped block where all VERIFY frame calldata has been removed and replaced by a `STARKValidationProof` header field
- A peer node importing the stripped block can verify the STARK proof and reach the same state root as a node that executed the frames directly
- Feature is gated behind `--stark-validation-frames=on|off` (default `off`) until stability is proven

### Tasks

#### Task PQ-5b.1 — Implement `frame_stark_replacer.go`
- **Description**: Create `pkg/core/vm/frame_stark_replacer.go` with `ReplaceValidationFrames(block *types.Block, prover ValidationFrameProver) (*types.Block, *proofs.STARKProofData, error)`. Extract all VERIFY frame calldata from block txs, call `prover.ProveAllValidationFrames()`, return stripped block + proof. If prover fails, return the original block unchanged.
- **Estimated Effort**: 5 SP
- **Assignee**: Go Engineer (vm + proofs)
- **Testing Method**: Unit test using `StubValidationFrameProver`: block with 10 frame txs → stripped block has 0 VERIFY calldata bytes, proof returned. `go test ./core/vm/... -run TestFrameStarkReplacer`. Full test with real circuit after US-PQ-5a lands.
- **Definition of Done**: Block correctly stripped. STARK proof field populated. Original block returned if prover errors. Code reviewed.

#### Task PQ-5b.2 — Wire into block sealing and import
- **Description**: In `pkg/engine/` block sealing: when `--stark-validation-frames=on`, call `ReplaceValidationFrames()` after building the block. In `pkg/core/` block import: if block header has `STARKValidationProof` field, call `prover.Verify()` instead of re-executing VERIFY frames. Both paths must reach the same state root.
- **Estimated Effort**: 5 SP
- **Assignee**: Go Engineer (engine + core)
- **Testing Method**: Integration test: build block (sealing), import on peer node (import). Verify both reach identical state root. EF state tests: 36,126/36,126 still pass (STARK path only activates for frame txs with the flag on). `go test ./core/... -run TestSTARKFrameImport`.
- **Definition of Done**: Sealing + import work end-to-end. EF tests unaffected. Feature flag working.

#### Task PQ-5b.3 — Devnet smoke test for STARK frame replacement
- **Description**: Add Kurtosis devnet config `pkg/devnet/kurtosis/configs/stark-frames.yaml` with `--stark-validation-frames=on` and 5% of txs as frame txs. Run for 50 slots. Verify: chain advances, no state root mismatches between nodes, block sizes smaller than without STARK replacement.
- **Estimated Effort**: 3 SP
- **Assignee**: DevOps Engineer
- **Testing Method**: `./scripts/run-devnet.sh stark-frames`. CL logs: no state root failures. Block size report: compare with/without STARK replacement enabled.
- **Definition of Done**: 50-slot devnet passes. Block sizes smaller by ≥ 10%. No state divergence.

---

## US-PQ-6: Production Groth16 Backend (gnark)

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **ZK proof system developer**, I want the Groth16 verifier to use real pairing-based verification from `gnark` instead of proof-size-only validation, so that AA proof circuits, mandatory 3-of-5 proofs, and jeanVM aggregation can produce and verify production proofs.

**Priority**: P2 | **Story Points**: 8 | **Sprint Target**: Sprint 3

### Tasks

#### Task PQ-6.1 — Integrate `gnark` Groth16 prove and verify
- **Description**: Replace proof-size-only validation in `pkg/proofs/groth16_verifier.go` with real `gnark/backend/groth16.Verify()`. Import `github.com/consensys/gnark` (already in `refs/`). Compile `AAValidationCircuit` to R1CS via `frontend.Compile()`. Add `gnark` to `pkg/go.mod`.
- **Estimated Effort**: 5 SP
- **Assignee**: ZK Engineer (proofs)
- **Testing Method**: Circuit test: generate a real Groth16 proof with `gnark/backend/groth16.Prove()`, verify with new verifier. Test: tampered proof → verify returns error. `go test ./proofs/... -run TestGroth16RealVerify`.
- **Definition of Done**: Real pairing verification works. `go test ./proofs/...` green. `PlaceholderGroth16` stub removed.

#### Task PQ-6.2 — Update `AAValidationCircuit` R1CS compilation
- **Description**: In `pkg/proofs/aa_proof_circuits.go`, add `CompileAACircuit() (*cs.R1CS, error)` that calls `frontend.Compile(ecc.BLS12_381.ScalarField(), r1cs.NewBuilder, &AAValidationCircuit{})`. Cache the compiled R1CS on startup. Expose `SetupKeys() (groth16.ProvingKey, groth16.VerifyingKey, error)`.
- **Estimated Effort**: 3 SP
- **Assignee**: ZK Engineer (proofs)
- **Testing Method**: Unit test: compile circuit → R1CS has expected number of constraints. Setup → proving key and verifying key produced. `go test ./proofs/... -run TestAACircuitCompile`.
- **Definition of Done**: Circuit compiles. Keys generated. Setup takes < 60s (acceptable for startup). Code reviewed.
