# ETH2030 Sprint User Stories — March 2026

> **Generated from**: `docs/plans/leanroadmap-coverage-2026-03.md` and all docs in `docs/plans/vitalik/`
> **Framework**: Scrum / Agile (INVEST criteria)
> **Date**: 2026-03-04
> **Story Points scale**: Fibonacci (1, 2, 3, 5, 8, 13, 21)

---

## INVEST Compliance Legend

| Criterion | Description |
|-----------|-------------|
| **I** Independent | Story can be worked without depending on another unfinished story |
| **N** Negotiable | Scope and approach can be refined in sprint planning |
| **V** Valuable | Delivers measurable value to protocol, interop, or security |
| **E** Estimable | Enough detail to size the effort |
| **S** Small | Fits in one sprint (≤13 SP) |
| **T** Testable | Has explicit, verifiable acceptance criteria |

---

## Epic Index

| Epic | Title | Stories | Total SP |
|------|-------|---------|----------|
| EP-1 | Account Abstraction & EIP-8141 | US-AA-1 … US-AA-5 | 47 |
| EP-2 | EL State Tree, BLAKE3 & RISC-V VM | US-BL-1, US-EL-2 … US-EL-4 | 45 |
| EP-3 | Post-Quantum Cryptography | US-PQ-2 … US-PQ-6 (incl. 5a/5b split) | 58 |
| EP-4 | leanConsensus & leanroadmap | US-LEAN-1 … US-LEAN-6, US-LEAN-3a/3b, US-LEAN-8 | 69 |
| EP-5 | Vitalik Roadmap Gaps | US-GAP-1 … US-GAP-5, US-GAP-7 | 70 |
| EP-6 | Block Building Pipeline | US-BB-1 … US-BB-2 | 18 |
| EP-7 | EIP Specification Compliance | US-SPEC-1, US-SPEC-3 … US-SPEC-7 | 63 |
| **Total** | | **37 stories** | **370 SP** |

---

## INVEST Refinement Log

The following corrections were applied after cross-checking all stories against their source EIP documents:

| Story | Issue | Fix Applied |
|-------|-------|-------------|
| US-PQ-5 | **S-violation**: labelled 13 SP but tasks total 21 SP | Split into US-PQ-5a (13 SP) + US-PQ-5b (8 SP) |
| US-AA-5 | **I-note**: depends on US-PQ-3 completing NTT address alignment first | Added `Depends on: US-PQ-3` note |
| EP-1 overall | EIP-8141 frame receipt structure, TSTORE cross-frame semantics, SENDER mode enforcement, TXPARAM* table completeness — not covered | Added US-SPEC-1, US-SPEC-2 in EP-7 |
| EP-5/EP-6 | EIP-7732 builder withdrawal lifecycle and epoch processing — not covered | Added US-SPEC-3 in EP-7 |
| EP-4/EP-6 | EIP-7805 IL equivocation detection, satisfaction algorithm, engine API status — not covered | Added US-SPEC-4 in EP-7 |
| EP-5 | EIP-7928 BAL ordering, ITEM_COST=2000 sizing, BlockAccessIndex assignment, retention policy — not covered | Added US-SPEC-5 in EP-7 |
| (missing) | EIP-7706 3D fee vector transaction type — **entirely absent** | Added US-SPEC-6 in EP-7 |
| EP-2 | EIP-7864 get_tree_key correctness, header data packing, code chunking accuracy — not covered by US-BL-1 (which covers hash function) | Added US-SPEC-7 in EP-7 |

### Consolidation Pass (2026-03-04)

| Change | Stories Affected | Result |
|--------|-----------------|--------|
| **Merge A** | US-EL-1 (BLAKE3 trie, 8 SP) + US-PQ-1 (BLAKE3 PQ crypto, 5 SP) | Combined into **US-BL-1** "BLAKE3 Hash Backend Integration" (12 SP). Moved out of EP-2 and EP-3; added as first story of EP-2 (renamed to "EL State Tree, BLAKE3 & RISC-V VM"). |
| **Merge B** | US-SPEC-1 (frame receipt, 8 SP) + US-SPEC-2 (TXPARAM*, 5 SP) | Combined into **US-SPEC-1** "EIP-8141 Frame TX Full Compliance" (13 SP). SPEC-2.1 → SPEC-1.5, SPEC-2.2 → SPEC-1.6. US-SPEC-2 removed from EP-7. |
| **Merge C** | US-GAP-3 (Random attesters, 13 SP) + US-LEAN-7 (Reduced committee fork-choice, 8 SP) | **US-GAP-3** retained at 13 SP; LEAN-7 devnet test content absorbed into GAP-3.4 DoD. LEAN-7.1 and LEAN-7.2 were duplicates of GAP-3.1 and GAP-3.3. US-LEAN-7 removed from EP-4. |
| **Merge D** | US-GAP-5 (Minimmit, 13 SP) + US-GAP-6 (3SF backoff, 5 SP) | **US-GAP-5** retained at 13 SP; GAP-6.2 simulation content absorbed into GAP-5.3 acceptance criteria. GAP-5 description and acceptance criteria updated to mention is_justifiable_slot backoff and 1000-slot simulation. US-GAP-6 removed from EP-5. |
| **Split E** | US-LEAN-3 "Separate PQ Aggregator Role" (labeled 13 SP, tasks summed to 14 SP) | Split into **US-LEAN-3a** "PQ Aggregator Role — Types & Duty Selection" (6 SP, LEAN-3.1 + LEAN-3.2) and **US-LEAN-3b** "PQ Aggregator Role — Collection & Aggregation" (8 SP, LEAN-3.3 + LEAN-3.4). |

### SP Label Fixes

| Story | Old SP Label | Correct SP (task sum) |
|-------|--------------|-----------------------|
| US-AA-3 | 13 | 12 (3+5+1+2+1=12) |
| US-AA-5 | 8 | 9 (2+5+2=9) |
| US-EL-3 | 8 | 10 (2+5+3=10) |
| US-EL-4 | 8 | 10 (5+2+3=10) |
| US-GAP-7 | 8 | 10 (5+2+3=10) |

---

---

# EPIC 1 — Account Abstraction & EIP-8141

**Goal**: Complete the remaining EIP-8141 frame transaction gaps so ETH2030 is Hegotá-ready with DoS-safe paymasters, dual-tier mempool, and full VERIFY simulation.

---

## US-AA-1: Paymaster Staking Registry

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **protocol maintainer**, I want a paymaster staking registry that requires paymasters to post a bond before they are accepted by the txpool, so that malicious paymasters cannot spam the mempool with frame transactions that appear valid but never settle.

**Priority**: P0 | **Story Points**: 13 | **Sprint Target**: Sprint 1

### Tasks

#### Task AA-1.1 — Implement `paymaster_registry.go`
- **Description**: Create `pkg/core/paymaster_registry.go` implementing minimum stake (e.g., 1 ETH), stake deposit/withdrawal logic, and a slashing counter per paymaster address. Model after ERC-4337 EntryPoint staking semantics but at the protocol layer. Must expose `IsApprovedPaymaster(addr common.Address, stateDB StateDB) bool`.
- **Estimated Effort**: 5 SP
- **Assignee**: Go Engineer (core/vm)
- **Testing Method**:
  - Unit tests in `pkg/core/paymaster_registry_test.go`: stake deposits, insufficient stake rejection, slash increments, withdrawal cooldown.
  - Integration test: frame tx with unstaked paymaster → txpool rejects; frame tx with staked paymaster → admitted.
- **Definition of Done**:
  - `IsApprovedPaymaster()` passes all unit tests.
  - `go test ./core/... -run TestPaymasterRegistry` green.
  - Code reviewed and merged to feature branch.
  - ≥ 80% line coverage on new file.

#### Task AA-1.2 — Wire registry into txpool frame tx validation
- **Description**: In `pkg/txpool/txpool.go`, call `IsApprovedPaymaster()` when a frame transaction contains a payer-approval (scope 1 or scope 2 APPROVE) frame. Reject transactions whose paymaster is not in the registry with error `ErrUnstakedPaymaster`. Add config flag `--paymaster-registry=strict|off` (default: `strict`).
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (txpool)
- **Testing Method**:
  - Txpool integration test: submit frame tx with unregistered paymaster → rejected; registered → admitted.
  - Test config flag: `--paymaster-registry=off` disables check.
- **Definition of Done**:
  - `go test ./txpool/... -run TestFrameTxPaymaster` green.
  - Flag documented in CLI help text.
  - No regression in existing txpool tests.

#### Task AA-1.3 — Slashing on invalid paymaster settlement
- **Description**: In `pkg/core/processor.go`, after frame tx execution, if the payer (APPROVE scope 1) fails to cover gas, increment the paymaster's slash counter in the registry. After N slashes (configurable, default 3), mark paymaster as banned.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (core)
- **Testing Method**:
  - Unit test: simulate 3 settlement failures → paymaster banned.
  - State transition test: banned paymaster's subsequent frame txs rejected at txpool.
- **Definition of Done**:
  - `go test ./core/... -run TestPaymasterSlashing` green.
  - Slashing logic does not add regression in `pkg/core/eftest` (36,126 state tests still passing).

#### Task AA-1.4 — Write test for end-to-end paymaster registry flow
- **Description**: Integration test covering: register paymaster (stake 1 ETH on-chain), submit frame tx using that paymaster, verify gas deducted from paymaster, verify correct settlement in receipt, then deregister and verify txpool rejection.
- **Estimated Effort**: 2 SP
- **Assignee**: QA Engineer
- **Testing Method**: Go integration test using in-memory chain + StateDB.
- **Definition of Done**: Test in `pkg/core/paymaster_registry_integration_test.go` passes. CI green.

---

## US-AA-2: Dual-Tier Mempool for Frame Transactions

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **node operator**, I want to configure a conservative or aggressive mempool tier for frame transactions, so that my node can choose the level of frame tx complexity it accepts based on its risk tolerance.

**Priority**: P1 | **Story Points**: 8 | **Sprint Target**: Sprint 2

### Tasks

#### Task AA-2.1 — Implement conservative frame ruleset
- **Description**: Create `pkg/txpool/frame_rules.go` with `ConservativeFrameRules`: VERIFY frame must come first, VERIFY frame cannot make external CALL opcodes, VERIFY frame gas limit capped at 50,000. Implement `ValidateFrameTxConservative(tx *types.FrameTx) error`.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (txpool)
- **Testing Method**: Unit tests covering each rule violation (no VERIFY first, external CALL in VERIFY, excess gas).
- **Definition of Done**: `go test ./txpool/... -run TestConservativeFrameRules` green. ≥90% coverage on new file.

#### Task AA-2.2 — Implement aggressive frame ruleset
- **Description**: Create `pkg/txpool/frame_rules_aggressive.go` with `AggressiveFrameRules`: same as conservative but allows external calls from VERIFY frames if paymaster is staked. Implement `ValidateFrameTxAggressive(tx *types.FrameTx, registry *PaymasterRegistry) error`.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (txpool)
- **Testing Method**: Unit tests: paymaster staked + external call in VERIFY → accepted; unstaked → rejected.
- **Definition of Done**: Tests pass. Both rulesets documented in godoc.

#### Task AA-2.3 — CLI flag and config wiring
- **Description**: Add `--frame-mempool=conservative|aggressive` flag (default: `conservative`) to `cmd/eth2030/`. Wire to txpool initialization so the correct ruleset is selected at startup.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (cmd)
- **Testing Method**: Manual: start node with each flag, verify txpool reports the configured tier.
- **Definition of Done**: Flag present in `--help` output. Config field `FrameMempoolTier` documented.

#### Task AA-2.4 — Txpool tier metrics
- **Description**: Add Prometheus counters in `pkg/metrics/`: `frame_tx_rejected_conservative_total`, `frame_tx_rejected_aggressive_total`, `frame_tx_accepted_total`. Expose via `/metrics` endpoint.
- **Estimated Effort**: 1 SP
- **Assignee**: DevOps Engineer
- **Testing Method**: Start node, submit frame txs, verify counters increment via `curl /metrics`.
- **Definition of Done**: Metrics visible in Grafana dashboard. No performance regression.

---

## US-AA-3: Full VERIFY Frame Simulation in Txpool

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **mempool operator**, I want the txpool to execute VERIFY frames in a read-only EVM before admitting frame transactions, so that transactions that never call APPROVE are rejected before wasting block gas.

**Priority**: P0 | **Story Points**: 12 | **Sprint Target**: Sprint 1

### Tasks

#### Task AA-3.1 — Inject StateDB into txpool
- **Description**: Refactor `pkg/txpool/txpool.go` to accept a `StateReader` interface providing `GetCodeSize(addr) int` and `GetBalance(addr) *big.Int`. Wire via `node.go` at startup. The current `txpool.go` has no StateDB reference — this is a structural prerequisite.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (txpool + node)
- **Testing Method**: Unit test with mock StateReader. Verify existing txpool tests still pass.
- **Definition of Done**: `StateReader` interface defined. `txpool.New()` accepts it. No regression.

#### Task AA-3.2 — Implement `simulateVerifyFrame()`
- **Description**: In `pkg/txpool/txpool.go`, add `simulateVerifyFrame(tx *types.FrameTx, stateDB StateDB) error` that runs the first VERIFY frame in a read-only EVM snapshot (no state commits). Returns error if: (a) APPROVE is never called, (b) VERIFY frame reverts, (c) target has no code.
- **Estimated Effort**: 5 SP
- **Assignee**: Go Engineer (txpool + vm)
- **Testing Method**: Unit tests: VERIFY that reverts → rejected; VERIFY that calls APPROVE → accepted; VERIFY targeting EOA (no code) → rejected with message `"frame tx: VERIFY target 0x... has no code (EOA)"`.
- **Definition of Done**: `go test ./txpool/... -run TestVerifyFrameSimulation` passes. Simulation adds < 5ms median latency per frame tx.

#### Task AA-3.3 — Codeless VERIFY target error message
- **Description**: In `pkg/core/frame_execution.go` and `pkg/txpool/txpool.go`, update the error message for VERIFY frames targeting an address with no code from generic caller error to `fmt.Errorf("frame tx: VERIFY target %s has no code (EOA)", addr.Hex())`.
- **Estimated Effort**: 1 SP
- **Assignee**: Go Engineer (core)
- **Testing Method**: Unit test verifies the exact error string.
- **Definition of Done**: Error string matches spec. Test green.

#### Task AA-3.4 — Benchmark VERIFY simulation overhead
- **Description**: Add `BenchmarkVerifyFrameSimulation` in `pkg/txpool/txpool_bench_test.go` measuring simulation overhead for 1, 10, 100 concurrent VERIFY simulations. Target: < 5ms p99 per frame tx.
- **Estimated Effort**: 2 SP
- **Assignee**: QA Engineer
- **Testing Method**: `go test -bench=BenchmarkVerifyFrameSimulation ./txpool/`. Compare to baseline txpool admission latency.
- **Definition of Done**: Benchmark result documented. Simulation overhead < 5ms p99 confirmed.

#### Task AA-3.5 — Refactoring note: StateDB injection pattern
- **Description**: Document the StateDB injection pattern in `pkg/txpool/README.md` (or add godoc) so future developers understand why txpool has a StateReader and how to avoid tight coupling.
- **Estimated Effort**: 1 SP
- **Assignee**: Go Engineer (documentation)
- **Testing Method**: Code review confirms documentation accuracy.
- **Definition of Done**: Doc added. PR reviewed.

---

## US-AA-4: Privacy Pool Reference Flow Documentation

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **DApp developer**, I want documented and tested reference flows showing how to use 2D nonces, ZK paymasters, and frame transactions for a privacy pool, so that I can replace centralized public broadcasters without relying on off-chain relayers.

**Priority**: P2 | **Story Points**: 5 | **Sprint Target**: Sprint 3

### Tasks

#### Task AA-4.1 — Document end-to-end privacy pool flow
- **Description**: Add `docs/guides/privacy-pool-frame-tx.md` documenting the three-step privacy pool pattern: (1) frame tx with nonce key = pool address (not user EOA), (2) VERIFY frame calls 0x0205 precompile with ZK-SNARK, (3) no relayer required. Include annotated code examples.
- **Estimated Effort**: 2 SP
- **Assignee**: Protocol Researcher
- **Testing Method**: Technical review by security engineer for correctness.
- **Definition of Done**: Document reviewed and merged. No broken links.

#### Task AA-4.2 — Integration test for privacy pool flow
- **Description**: Write `pkg/core/privacy_pool_integration_test.go` exercising: (a) submit encrypted frame tx with ZK paymaster calldata, (b) execute VERIFY frame calling AA proof precompile (0x0205), (c) confirm no relayer address in tx origin, (d) confirm payer is pool contract, not user EOA.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (core + proofs)
- **Testing Method**: `go test ./core/... -run TestPrivacyPoolFlow`. Test must pass on in-memory chain.
- **Definition of Done**: Test green. Coverage of privacy flow ≥ 80%. Code reviewed.

---

## US-AA-5: NTT-Accelerated Falcon for VERIFY Frames

**INVEST**: I⚠ N✓ V✓ E✓ S✓ T✓
> **I-note**: This story depends on **US-PQ-3** (NTT precompile addresses aligned to `0x0f–0x12`) completing first. It is independently schedulable in a later sprint but cannot be implemented before US-PQ-3 lands. Within its own sprint (Sprint 3) it is unblocked.

**User Story**:
> As a **wallet developer**, I want Falcon-512 signature verification in VERIFY frames to use the NTT precompile for acceleration, so that PQ-secured accounts have gas costs competitive with ECDSA accounts (~200K gas vs 3000).

**Priority**: P2 | **Story Points**: 9 | **Sprint Target**: Sprint 3

### Tasks

#### Task AA-5.1 — Audit Falcon CVEs against Go bindings
- **Description**: Review `pkg/crypto/pqc/falcon_signer.go` Go C-bindings against the three CVEs in `refs/ethfalcon/`: CVETH-2025-080201 (CRITICAL: salt size), CVETH-2025-080202 (MEDIUM: signature malleability), CVETH-2025-080203 (LOW: domain separation). Document findings in `docs/security/falcon-cve-assessment-2026.md`.
- **Estimated Effort**: 2 SP
- **Assignee**: Security Engineer
- **Testing Method**: Code audit + test: attempt Falcon verify with truncated salt, verify rejection.
- **Definition of Done**: CVE assessment doc complete. Critical CVE either confirmed not-applicable or patched. Security review signed off.

#### Task AA-5.2 — Implement NTT-aware Falcon precompile wrapper
- **Description**: Create `pkg/core/vm/precompile_falcon.go` at address `0x0206`: takes Falcon-512 public key + message hash + signature as input, internally calls NTT precompile (addresses `0x0f`–`0x12` per spec, see NTT alignment task US-LEAN-3) for polynomial operations, returns 1 (valid) or 0 (invalid). Target: < 200K gas total.
- **Estimated Effort**: 5 SP
- **Assignee**: Go Engineer (vm + pqc)
- **Testing Method**: Unit test: verify valid Falcon-512 sig → precompile returns 1. Invalid sig → returns 0. Gas benchmark: `go test -bench=BenchmarkFalconPrecompile ./core/vm/`.
- **Definition of Done**: `go test ./core/vm/... -run TestFalconPrecompile` green. Gas ≤ 200K confirmed. CVE audit done (Task AA-5.1) before deployment.

#### Task AA-5.3 — Gas benchmark: Falcon vs ECDSA in VERIFY frame
- **Description**: Add `pkg/core/vm/pq_precompile_gas_test.go` benchmarking: ECDSA (baseline 3000 gas), Falcon-512 via NTT precompile, ML-DSA-65 via EVM, SPHINCS+ via EVM. Output comparative table. Target: Falcon ≤ 200K gas.
- **Estimated Effort**: 2 SP
- **Assignee**: QA Engineer
- **Testing Method**: `go test -bench=BenchmarkPQGas ./core/vm/`. Output logged to CI artifact.
- **Definition of Done**: Benchmark documented. All results within expected ranges per Vitalik's ~200K estimate.

---

---

# EPIC 2 — EL State Tree, BLAKE3 & RISC-V VM

**Goal**: Integrate BLAKE3 as a shared hash backend for both the binary trie and PQ crypto layer, align the binary trie with EIP-7864 spec, implement RISC-V precompile guests, and wire the RISC-V execution path for production deployment.

---

## US-BL-1: BLAKE3 Hash Backend Integration

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **protocol engineer**, I want BLAKE3 integrated as a shared hash backend for both the binary trie (EIP-7864) and the PQ crypto layer (hash-based signatures), so that both subsystems benefit from the same vetted dependency (lukechampine.com/blake3) without duplicated work.

**Priority**: P0 | **Story Points**: 12 | **Sprint Target**: Sprint 1

### Tasks

#### Task BL-1.1 — Implement binary trie BLAKE3 hasher
- **Description**: Create `pkg/trie/bintrie/hasher_blake3.go` using `lukechampine.com/blake3`. Implement `BinaryHasher.hashBLAKE3(left, right []byte)` following the EIP-7864 hash rule: `hash(left||right)=BLAKE3-256`. Add `HashFunctionBlake3` constant. Add `lukechampine.com/blake3` to `pkg/go.mod`.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (trie)
- **Testing Method**: Unit test: known 4-leaf tree BLAKE3 root matches reference vector. `go test ./trie/bintrie/... -run TestBlake3Hasher`.
- **Definition of Done**: `go test ./trie/bintrie/...` green. Dependency in `go.mod`.

#### Task BL-1.2 — Implement PQ crypto BLAKE3 backend
- **Description**: In `pkg/crypto/pqc/hash_backend.go`, replace `Blake3Backend.Hash()` stub (SHA-256 approximation) with `blake3.Sum256(data)` from `lukechampine.com/blake3`.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (pqc)
- **Testing Method**: Unit test: BLAKE3 of empty string = `AF1349B9...` (known vector). Compare with `refs/hash-sig/`. `go test ./crypto/pqc/... -run TestBlake3Backend` green.
- **Definition of Done**: `go test ./crypto/pqc/... -run TestBlake3Backend` green.

#### Task BL-1.3 — Fork-configurable hash function for trie
- **Description**: Add `BinaryTreeHashFunc` field to chain config (default `SHA256`, `BLAKE3` when `EIP7864FinalHash` fork active). Wire into `pkg/trie/bintrie/bintrie.go` `NewBinaryTrie()`. No existing tree roots change unless fork explicitly activated.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (trie + core)
- **Testing Method**: Fork switch test; verify 36,126 EF state tests still pass.
- **Definition of Done**: Fork gate working. `go test ./core/eftest/...` still 36,126/36,126. Config documented.

#### Task BL-1.4 — Wire state expiry epoch into StemNode
- **Description**: In `pkg/core/state/state_expiry.go` `TouchAccount()` and `TouchStorage()`, call `bintrie.UpdateLeafMetadata(stem, subindex=2, epoch)` to record last-access epoch in the reserved metadata slot (subindex 2–63).
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (state + trie)
- **Testing Method**: Unit test: read account → `TouchAccount` called → StemNode subindex 2 updated. `go test ./core/state/... -run TestStateExpiryEpochTracking` green.
- **Definition of Done**: `go test ./core/state/... -run TestStateExpiryEpochTracking` green. No regression in existing state tests.

#### Task BL-1.5 — BLAKE3 benchmark and hash recommendation doc
- **Description**: Add `pkg/trie/bintrie/hasher_bench_test.go` and `pkg/crypto/pqc/hash_bench_test.go` benchmarking BLAKE3 vs SHA-256 for trie hashing and PQ signing respectively. Document results in `docs/plans/blake3-benchmark-2026-03.md`.
- **Estimated Effort**: 2 SP
- **Assignee**: QA Engineer
- **Testing Method**: `go test -bench=. ./trie/bintrie/ ./crypto/pqc/`.
- **Definition of Done**: BLAKE3 ≥ 2× SHA-256 or deviation explained. Benchmark doc created.

---

## US-EL-2: RISC-V Precompile Guest Programs (Step 1)

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **zkVM protocol developer**, I want the 7 most common EVM precompiles (Keccak256, SHA256, ECRECOVER, ModExp, BN256Add, BN256Mul, BN256Pair) implemented as RISC-V guest programs, so that 80% of today's precompile usage runs through the prover-efficient RISC-V path.

**Priority**: P0 | **Story Points**: 13 | **Sprint Target**: Sprint 2

### Tasks

#### Task EL-2.1 — Implement Keccak256 and SHA256 RISC-V guests
- **Description**: Create `pkg/zkvm/guests/keccak256.s` and `pkg/zkvm/guests/sha256.s` — RISC-V RV32IM assembly for Keccak-256 and SHA-256. Register both in `pkg/zkvm/canonical.go` `GuestRegistry` at startup. These are the highest-frequency precompiles.
- **Estimated Effort**: 5 SP
- **Assignee**: ZK Engineer (zkvm)
- **Testing Method**: Round-trip test: compute hash via Go native → compute via RISC-V guest → compare outputs. `go test ./zkvm/... -run TestKeccakRISCV` and `TestSHA256RISCV`.
- **Definition of Done**: Both guests registered. Round-trip test passes. Gas cost compared to current Go function gas (documented).

#### Task EL-2.2 — Implement ECRECOVER RISC-V guest
- **Description**: Create `pkg/zkvm/guests/ecrecover.s` implementing secp256k1 ECDSA recovery in RISC-V. Use existing `pkg/crypto/` secp256k1 logic as reference. Register at op selector `0x03` in `zkisa_bridge.go`.
- **Estimated Effort**: 5 SP
- **Assignee**: ZK Engineer (zkvm + crypto)
- **Testing Method**: Regression test: 100 random ECRECOVER test vectors from Ethereum test suite, compare RISC-V output with go-ethereum output.
- **Definition of Done**: All 100 vectors pass. `go test ./zkvm/... -run TestECRECOVERRISCV` green.

#### Task EL-2.3 — Wire precompile router to RISC-V path
- **Description**: In `pkg/core/vm/precompiles.go`, add a fork check `IsPrecompileRISCV(blockNum)` (active at I+ fork). When active, route Keccak256/SHA256/ECRECOVER calls through `pkg/zkvm/canonical.go` `CanonicalGuestPrecompile` instead of Go native functions. Other precompiles still use Go path.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (vm + zkvm)
- **Testing Method**: Toggle fork: before I+ → Go path; after I+ → RISC-V path. Both must produce identical results. EF state tests: 36,126/36,126 still passing.
- **Definition of Done**: Fork gate working. `go test ./core/eftest/...` green. No performance regression > 20%.

---

## US-EL-3: RVCREATE Opcode for User RISC-V Contracts (Step 2)

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **DApp developer**, I want to deploy RISC-V bytecode to a contract address using a `RVCREATE` opcode, so that users can write high-performance contracts in RISC-V without going through the precompile registry.

**Priority**: P1 | **Story Points**: 10 | **Sprint Target**: Sprint 2

### Tasks

#### Task EL-3.1 — Add `RVCREATE` opcode definition
- **Description**: Add `RVCREATE = 0xF6` to `pkg/core/vm/opcodes.go`. Define gas model: 32,000 base (same as CREATE) + 200 × code_size for RISC-V programs (prover overhead). Add to jump table at I+ fork level.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (vm)
- **Testing Method**: Unit test: `RVCREATE` in jump table at I+ block. Gas calculation test.
- **Definition of Done**: Opcode defined. Gas formula verified. No regression.

#### Task EL-3.2 — Implement `RVCREATE` execution handler
- **Description**: Implement `opRVCreate` in `pkg/core/vm/instructions.go`: validates initcode starts with RISC-V magic bytes `0xFE 0x52 0x56`, deploys to CREATE2-style deterministic address, stores program in `GuestRegistry`. On call, routes `CALL` to RISC-V executor when code magic matches.
- **Estimated Effort**: 5 SP
- **Assignee**: Go Engineer (vm + zkvm)
- **Testing Method**: Integration test: deploy RISC-V program with `RVCREATE`, call it via standard `CALL`, verify output matches expected. Test: invalid magic bytes → revert.
- **Definition of Done**: `go test ./core/vm/... -run TestRVCREATE` green. `go test ./core/eftest/...` 36,126/36,126 still passing.

#### Task EL-3.3 — RISC-V contract call routing via magic bytes
- **Description**: In `pkg/core/vm/evm.go` `Call()` method, before delegating to EVM bytecode execution, check if code starts with RISC-V magic `0xFE 0x52 0x56`. If so, route to `CanonicalGuestPrecompile` instead. Apply gas model: 1 RISC-V cycle = 1 EVM gas (configurable via chain config `RVCycleGasRatio`).
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (vm)
- **Testing Method**: Test: deploy RISC-V program → call it → RISC-V executor invoked (verify via trace). Call EVM contract → EVM executor invoked.
- **Definition of Done**: `go test ./core/vm/... -run TestRVCallRouting` green. No ABI-breaking change to existing contracts.

---

## US-EL-4: Production KZG Backend Upgrade

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **validator node operator**, I want the KZG backend to use the official Ethereum trusted setup (from `go-eth-kzg`) instead of the placeholder test SRS (s=42), so that blob commitments, proofs, and PeerDAS cell operations are cryptographically valid on mainnet.

**Priority**: P0 | **Story Points**: 10 | **Sprint Target**: Sprint 1

### Tasks

#### Task EL-4.1 — Replace `PlaceholderKZGBackend` with `go-eth-kzg`
- **Description**: Replace `PlaceholderKZGBackend` in the relevant package with `goethkzg.NewContext4096Secure()` from `github.com/crate-crypto/go-eth-kzg` (already in `refs/`). Implement `BlobToKZGCommitment`, `ComputeBlobKZGProof`, `VerifyBlobKZGProof`, `VerifyBlobKZGProofBatchPar`, and `VerifyCellKZGProofBatch` using the production API.
- **Estimated Effort**: 5 SP
- **Assignee**: Go Engineer (das + crypto)
- **Testing Method**: KZG test suite: commit 10 blobs, compute proofs, verify batch. EIP-7594 cell proof test: `VerifyCellKZGProofBatch` with known cell/commitment pairs. All must pass with production SRS.
- **Definition of Done**: `PlaceholderKZGBackend` fully replaced. `go test ./das/... -run TestKZG` green. Blob sizes confirmed: `Blob=[131072]byte`, `KZGCommitment=[48]byte`, `KZGProof=[48]byte`, `Cell=[2048]byte`.

#### Task EL-4.2 — Verify `custody_verify.go` uses `VerifyCellKZGProofBatch`
- **Description**: Update `pkg/das/custody_verify.go` to use `go-eth-kzg`'s `VerifyCellKZGProofBatch` instead of the previous cell verification path. This is the key path for PeerDAS 8 MB/sec throughput target.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (das)
- **Testing Method**: Throughput test: verify 128-column batch at simulated 8 MB/sec. Compare latency before/after upgrade.
- **Definition of Done**: Custody verify uses production API. Throughput benchmark passes. CI green.

#### Task EL-4.3 — PeerDAS throughput devnet benchmark
- **Description**: Add Kurtosis devnet config `pkg/devnet/kurtosis/configs/peerdas-throughput.yaml` that sends 6 blobs/slot (current max) and verifies all cell proofs within slot time. Goal: confirm 8 MB/sec target is achievable.
- **Estimated Effort**: 3 SP
- **Assignee**: DevOps Engineer
- **Testing Method**: `cd pkg/devnet/kurtosis && ./scripts/run-devnet.sh peerdas-throughput`. Verify EL logs show `custody_verify: batch OK` within 2s of slot start.
- **Definition of Done**: Devnet test passes. Throughput benchmark result documented.

---

---

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

---

---

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

---

---

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

---

---

# EPIC 6 — Block Building Pipeline

**Goal**: Complete the remaining block building pipeline gaps identified in the Vitalik analysis: real mixnet transport and passive serverless order-matching research.

---

## US-BB-1: Real Mixnet Integration (Tor/Nym)

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **privacy-sensitive user**, I want ETH2030's transaction ingress to optionally route transactions through a real anonymizing network (Tor or Nym), so that sender IP addresses are not visible to block builders even in the Ethereum network layer.

**Priority**: Medium | **Story Points**: 13 | **Sprint Target**: Sprint 5

### Tasks

#### Task BB-1.1 — Define `ExternalMixnetTransport` interface
- **Description**: In `pkg/p2p/anonymous_transport.go`, add `ExternalMixnetTransport` interface with `SendViaExternalMixnet(tx []byte, endpoint string) error`. Add `MixnetTransportMode` enum: `Simulated` (current), `TorSocks5`, `NymSocks5`. Add CLI flag `--mixnet=simulated|tor|nym` (default `simulated`).
- **Estimated Effort**: 2 SP
- **Assignee**: P2P Engineer
- **Testing Method**: Config test: mode enum parsed correctly. Interface present but implementations gated behind build tag.
- **Definition of Done**: Interface defined. CLI flag present. No regression.

#### Task BB-1.2 — Implement Tor SOCKS5 transport
- **Description**: Create `pkg/p2p/tor_transport.go` implementing `ExternalMixnetTransport` using SOCKS5 proxy at `127.0.0.1:9050` (Tor default). Submit transactions as HTTP POST to the node's own RPC via Tor, ensuring sender IP is obscured at the network layer.
- **Estimated Effort**: 5 SP
- **Assignee**: P2P Engineer
- **Testing Method**: Integration test (requires Tor daemon): connect via SOCKS5, submit a transaction, verify receipt. Mock test without Tor: `MockTorTransport` that simulates SOCKS5 protocol. `go test ./p2p/... -run TestTorTransport`.
- **Definition of Done**: SOCKS5 connection working. Mock test passes without Tor daemon. Real Tor test passes in CI with Tor installed.

#### Task BB-1.3 — Transport selection and fallback
- **Description**: In `pkg/p2p/anonymous_transport.go` `TransportManager`, add priority: `Tor > Nym > Simulated`. If Tor is not reachable within 500ms, fall back to Nym; if Nym unavailable, fall back to Simulated. Log transport selection at startup.
- **Estimated Effort**: 3 SP
- **Assignee**: P2P Engineer
- **Testing Method**: Test: Tor unavailable → Nym attempted → Nym unavailable → Simulated used. `go test ./p2p/... -run TestTransportFallback`.
- **Definition of Done**: Fallback chain working. Startup log shows selected transport. Tests pass.

#### Task BB-1.4 — Kohaku interface alignment
- **Description**: Update `TransportManager` API to align with the kohaku protocol interface (once spec is published at the referenced `@ncsgy` repo). Add `KohakuCompatible bool` flag to `TransportConfig` — when `true`, use kohaku wire format for transport control messages.
- **Estimated Effort**: 3 SP
- **Assignee**: P2P Engineer
- **Testing Method**: Config test: `KohakuCompatible=true` → cohaku format messages sent. `go test ./p2p/... -run TestKohakuCompatibility`.
- **Definition of Done**: Kohaku flag present. Wire format conditional. Tests pass. TODO note: update once kohaku spec finalizes.

---

## US-BB-2: Distributed Block Building — New Local Tx Type (Research Spike)

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **protocol researcher**, I want a documented design for a "less global" transaction type that is cheaper and more amenable to distributed building (as described by Vitalik), so that the ETH2030 team can evaluate the design before it becomes an EIP.

**Priority**: Low | **Story Points**: 5 | **Sprint Target**: Sprint 6

### Tasks

#### Task BB-2.1 — Research spike: define "local tx" semantics
- **Description**: Write `docs/research/local-tx-design-2026-03.md` exploring: (1) what makes a tx "less global" (limited state access, predeclared BAL), (2) how local txs could be built by distributed builders without ordering coordination, (3) gas discount model (50–80% cheaper), (4) mempool routing (sharded mempool per sender prefix).
- **Estimated Effort**: 3 SP
- **Assignee**: Protocol Researcher
- **Testing Method**: Document peer-reviewed by senior engineer and consensus researcher. No code.
- **Definition of Done**: Design document written and reviewed. Key trade-offs documented. EIP sketch (optional).

#### Task BB-2.2 — Prototype `LocalTx` type (gated behind flag)
- **Description**: Create `pkg/core/types/tx_local.go` with `LocalTx` type (tx type `0x08`): must declare BAL in tx body, state access limited to declared keys, gas price discount configurable. Gated behind `--experimental-local-tx` flag. This is a proof-of-concept, not production-ready.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (core/types)
- **Testing Method**: Unit test: `LocalTx` struct construction, BAL validation (declared vs actual access), gas discount calculation. `go test ./core/types/... -run TestLocalTx`.
- **Definition of Done**: `LocalTx` type defined. BAL check works. Gas discount applied. Gated behind experimental flag. Design document (Task BB-2.1) complete first.

---

---

# EPIC 7 — EIP Specification Compliance

**Goal**: Close the gaps found during cross-referencing of user stories against the six source EIP documents (EIP-8141, EIP-7732, EIP-7805, EIP-7928, EIP-7706, EIP-7864). Every story in this epic corresponds to a spec requirement that was absent from Epics 1–6.

---

## US-SPEC-1: EIP-8141 Frame TX Full Compliance

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As an **EVM engineer**, I want frame transaction receipts to use the correct 3-layer structure, TSTORE/TLOAD discarded between frames, and all 16 TXPARAM* parameter indices correctly implemented, so that ETH2030 is fully EIP-8141 spec-compliant.

**Priority**: P0 | **Story Points**: 13 | **Sprint Target**: Sprint 2

### Acceptance Criteria
- Frame tx receipt RLP encodes as `[cumulative_gas_used, payer, [[status, gas_used, logs], ...]]`
- `TSTORE` written in frame N is not visible to frame N+1 (`TLOAD` returns zero)
- `ORIGIN` opcode inside a frame returns the frame caller address, not the traditional tx sender
- Frame mode SENDER (0x02) reverts immediately if `sender_approved == false` at entry
- All 16 parameter indices from the EIP-8141 spec table are implemented and testable
- `TXPARAMLOAD(0x08, 0)` returns `compute_sig_hash(tx)` — 32-byte signing hash
- `TXPARAMLOAD(0x09, 0)` returns `len(tx.frames)`
- `TXPARAMLOAD(0x10, 0)` returns currently executing frame index
- `TXPARAMLOAD(0x15, frame_index)` returns frame execution status (0=fail, 1=success)
- `TXPARAMSIZE` returns correct byte length for each parameter

### Tasks

#### Task SPEC-1.1 — Implement frame tx receipt 3-layer encoding
- **Description**: In `pkg/core/types/receipt.go`, add `FrameReceipt` struct `{Status uint64, GasUsed uint64, Logs []*Log}` and update `Receipt` for frame txs to contain `Payer common.Address` and `FrameReceipts []FrameReceipt`. RLP encoder must produce `[cumulative_gas_used, payer, [frame_receipt, ...]]` exactly per EIP-8141 spec §receipt.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (core/types)
- **Testing Method**: Unit test: execute frame tx with 3 frames → receipt has 3 `FrameReceipts` entries, each with correct status/gas/logs. RLP round-trip test. `go test ./core/types/... -run TestFrameTxReceipt`.
- **Definition of Done**: `FrameReceipt` type defined. RLP encoding correct. Round-trip passes. ≥ 80% coverage. EF state tests unaffected.

#### Task SPEC-1.2 — Enforce TSTORE/TLOAD cross-frame discard
- **Description**: In `pkg/core/vm/evm.go` frame execution loop, after each frame completes (success or revert), call `stateDB.ClearTransientStorage()` before starting the next frame. Per EIP-8141 spec: warm/cold state journal is shared across frames, but transient storage (`TSTORE`/`TLOAD`) is discarded at frame boundary.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (vm)
- **Testing Method**: Test: frame 0 does `TSTORE(key, 42)`, frame 1 does `TLOAD(key)` → returns 0. Contrast: `SLOAD` set in frame 0 is visible to frame 1 (warm journal shared). `go test ./core/vm/... -run TestFrameTSTORECrossFrame`.
- **Definition of Done**: `ClearTransientStorage()` called at each frame boundary. Test passes. EF state tests (36,126) unaffected.

#### Task SPEC-1.3 — ORIGIN opcode returns frame caller in frame context
- **Description**: In `pkg/core/vm/instructions.go` `opOrigin`, when executing inside a FRAME context (detect via `FrameContext.IsActive`), return `FrameContext.Caller` instead of `tx.From`. Outside frame context, ORIGIN continues to return `tx.From` as usual.
- **Estimated Effort**: 1 SP
- **Assignee**: Go Engineer (vm)
- **Testing Method**: Unit test: ORIGIN inside a frame → returns frame caller. ORIGIN outside frame → returns `tx.From`. `go test ./core/vm/... -run TestOriginInFrameContext`.
- **Definition of Done**: ORIGIN correct in both contexts. Test green. EF state tests unaffected.

#### Task SPEC-1.4 — SENDER frame mode enforces sender_approved precondition
- **Description**: In `pkg/core/vm/aa_executor.go` frame execution dispatch, when frame mode = `0x02` (SENDER), check `FrameContext.SenderApproved == true` before executing. If false, immediately revert with error `"frame: SENDER mode requires prior sender_approved"`. This is a stateful precondition not validated at tx admission.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (vm)
- **Testing Method**: Test: frame tx with SENDER frame before a VERIFY+APPROVE frame → SENDER frame reverts. After VERIFY+APPROVE, SENDER frame executes. `go test ./core/vm/... -run TestFrameSenderModePrecondition`.
- **Definition of Done**: Precondition enforced. Error message matches spec text. Test green.

#### Task SPEC-1.5 — Audit and implement missing TXPARAM indices
- **Description**: In `pkg/core/vm/eip8141_opcodes.go`, compare the existing `TXPARAMLOAD` switch against the 16-entry spec table. Add any missing cases: particularly `in1=0x06` (max cost), `in1=0x07` (blob hash count), `in1=0x08` (`compute_sig_hash`), `in1=0x09` (`len(frames)`), `in1=0x10` (current frame index), `in1=0x15` (frame status). Implement `compute_sig_hash(tx)` as `keccak256(RLP(chain_id, nonce, sender, frames, ...))`.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (vm)
- **Testing Method**: Table-driven test in `pkg/core/vm/eip8141_txparam_test.go` covering all 16 parameter indices. Each test case: construct a frame tx with known fields, execute `TXPARAMLOAD(in1, in2)`, assert expected value. `go test ./core/vm/... -run TestTXPARAMAllIndices`.
- **Definition of Done**: All 16 parameter indices covered. Table-driven test passes. EF state tests unaffected.

#### Task SPEC-1.6 — TXPARAMCOPY and TXPARAMSIZE completeness
- **Description**: In `pkg/core/vm/eip8141_opcodes.go`, verify `TXPARAMCOPY` (0xb2) correctly handles variable-size parameters (frame data via `in1=0x12`). Verify `TXPARAMSIZE` (0xb1) returns 32 for all fixed-size params and the correct dynamic size for `in1=0x12` (frame data) and blob hashes. Add tests for dynamic-size copy.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (vm)
- **Testing Method**: Unit test: `TXPARAMCOPY` copies frame data into memory at correct offset; memory matches frame data exactly. `TXPARAMSIZE` for fixed params returns 32; for dynamic returns actual length. `go test ./core/vm/... -run TestTXPARAMCOPY`.
- **Definition of Done**: Copy and size ops correct for all param types. Tests green.

---

## US-SPEC-3: EIP-7732 ePBS Builder Withdrawal & Epoch Processing

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **builder node operator**, I want the builder withdrawal mechanism (64-epoch delay, withdrawal prefix `0x03`, batch sweep of 16,384 builders/epoch) and `process_builder_pending_payments` epoch processing correctly implemented, so that builder balances are managed safely and predictably across epoch boundaries.

**Priority**: P1 | **Story Points**: 8 | **Sprint Target**: Sprint 2

### Acceptance Criteria
- `MIN_BUILDER_WITHDRAWABILITY_DELAY = 64` epochs enforced; withdrawal request before delay → rejected
- `MAX_BUILDERS_PER_WITHDRAWALS_SWEEP = 16,384` per epoch sweep
- `process_builder_pending_payments()` runs in epoch processing and correctly deducts from beacon chain
- `ProposerPreferences` P2P gossip topic (`DOMAIN_PROPOSER_PREFERENCES = 0x0D000000`) registered and handled
- Builder self-build flag (`BUILDER_INDEX_SELF_BUILD = UINT64_MAX`) accepted without auction

### Tasks

#### Task SPEC-3.1 — Implement `process_builder_pending_payments` in epoch processing
- **Description**: In `pkg/consensus/epoch_processing.go` (or equivalent), add `processBuilderPendingPayments(state)` that iterates `state.builder_pending_payments`, deducts amounts from the beacon chain, and updates builder balances. Must run after `processWithdrawals` and before `processFinalUpdates`. Per EIP-7732 §epoch-processing.
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Unit test: add 3 pending payments, run epoch processing, verify all 3 deducted correctly. Edge case: payment amount > builder balance → capped at balance. `go test ./consensus/... -run TestBuilderPendingPayments`.
- **Definition of Done**: Epoch processing function runs. Payments deducted. Edge cases handled. Tests green.

#### Task SPEC-3.2 — Implement builder withdrawal with 64-epoch delay
- **Description**: In `pkg/epbs/builder_registry.go`, implement `RequestBuilderWithdrawal(builderIdx, amount)` that sets `builder.withdrawable_epoch = current_epoch + MIN_BUILDER_WITHDRAWABILITY_DELAY` (64 epochs). In `pkg/consensus/epoch_processing.go` withdrawal sweep, process up to `MAX_BUILDERS_PER_WITHDRAWALS_SWEEP = 16,384` builders per epoch, advancing `state.next_withdrawal_builder_index`. Enforce `BUILDER_WITHDRAWAL_PREFIX = 0x03` on execution addresses.
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Unit test: request withdrawal at epoch 0 → not processed until epoch 64. Batch test: 20,000 builders requesting withdrawal → only 16,384 processed per epoch sweep. `go test ./epbs/... -run TestBuilderWithdrawal`.
- **Definition of Done**: 64-epoch delay enforced. Batch sweep limit correct. Prefix validated. Tests green.

#### Task SPEC-3.3 — ProposerPreferences P2P topic and self-build support
- **Description**: In `pkg/p2p/gossip_topics.go`, register gossip topic for `ProposerPreferences` messages using domain `DOMAIN_PROPOSER_PREFERENCES = 0x0D000000`. In `pkg/epbs/auction_engine.go`, when a bid has `builder_index = UINT64_MAX` (`BUILDER_INDEX_SELF_BUILD`), skip the auction and set the proposer as payload builder directly.
- **Estimated Effort**: 2 SP
- **Assignee**: P2P Engineer + Consensus Engineer
- **Testing Method**: P2P test: publish `ProposerPreferences` → topic subscriber receives it. Self-build test: proposer submits bid with `BUILDER_INDEX_SELF_BUILD` → auction skipped, proposer builds payload. `go test ./epbs/... -run TestBuilderSelfBuild`.
- **Definition of Done**: Topic registered. Self-build works. Tests green.

---

## US-SPEC-4: EIP-7805 FOCIL IL Equivocation Detection & Satisfaction Algorithm

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **validator**, I want IL equivocation detection (rejecting members who publish two conflicting ILs) and the correct O(n) IL satisfaction check (validating nonce + balance against post-execution state), so that the FOCIL protocol cannot be gamed by malicious IL committee members.

**Priority**: P1 | **Story Points**: 8 | **Sprint Target**: Sprint 2

### Acceptance Criteria
- A validator who sends 2 different ILs for the same slot is marked as equivocator; subsequent ILs from them are ignored
- The IL satisfaction algorithm correctly evaluates each tx in ILs against the post-execution state (nonce + balance)
- `engine_getInclusionListV1` Engine API endpoint is implemented and returns the current IL for the given slot
- `INCLUSION_LIST_UNSATISFIED` is returned by `engine_newPayload` when a tx in the IL is valid but absent from the block

### Tasks

#### Task SPEC-4.1 — Implement IL equivocation detection
- **Description**: In `pkg/focil/il_store.go`, maintain per-validator-per-slot IL store. On receiving a second `SignedInclusionList` from the same validator for the same slot: if the ILs differ, mark validator as equivocator via `il_store.MarkEquivocator(validatorIdx, slot)`. Subsequent ILs from the equivocator are silently dropped. Per EIP-7805 spec §equivocation.
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Unit test: receive 2 identical ILs → no equivocation. Receive 2 different ILs → equivocator flagged, 3rd IL dropped. Verify count: `il_store.EquivocatorCount(slot) == 1`. `go test ./focil/... -run TestILEquivocationDetection`.
- **Definition of Done**: Equivocation detection correct. Equivocator's ILs dropped. Tests green. ≥ 80% coverage.

#### Task SPEC-4.2 — Implement EIP-7805 O(n) IL satisfaction algorithm
- **Description**: In `pkg/focil/il_validator.go`, implement `CheckILSatisfaction(block, ils, postState) bool` per EIP-7805 spec §satisfaction: for each tx T in ILs, if T is in block → skip. If gas remaining < T's gas limit → skip (insufficient gas is not a violation). Else validate T's nonce and balance against `postState` (state after all prior txs). If nonce/balance valid but T is absent → return `INCLUSION_LIST_UNSATISFIED`. Replace any ad-hoc current check with this canonical algorithm.
- **Estimated Effort**: 3 SP
- **Assignee**: Consensus Engineer
- **Testing Method**: Unit test cases: (1) all ILs txs in block → satisfied. (2) IL tx absent, gas available, nonce/balance valid → unsatisfied. (3) IL tx absent, insufficient gas → satisfied (gas exemption). (4) IL tx absent, invalid nonce → satisfied (state-invalid exemption). `go test ./focil/... -run TestILSatisfactionAlgorithm`.
- **Definition of Done**: Algorithm matches EIP-7805 spec text exactly. All 4 test cases pass.

#### Task SPEC-4.3 — Add `engine_getInclusionListV1` and `INCLUSION_LIST_UNSATISFIED` status
- **Description**: In `pkg/engine/`, implement `engine_getInclusionListV1(slot, committee_index) -> SignedInclusionList`. In `engine_newPayload` handler, call `CheckILSatisfaction()` and return `{status: "INCLUSION_LIST_UNSATISFIED"}` if the check fails. Add `INCLUSION_LIST_UNSATISFIED = "INCLUSION_LIST_UNSATISFIED"` constant per EIP-7805 spec §engine-api.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (engine)
- **Testing Method**: API test: call `engine_getInclusionListV1` with valid slot → returns IL. `engine_newPayload` with block missing a required IL tx → returns `INCLUSION_LIST_UNSATISFIED`. `go test ./engine/... -run TestEngineILSatisfied` and `TestEngineILUnsatisfied`.
- **Definition of Done**: Both endpoints implemented. Status constant defined. Tests green.

---

## US-SPEC-5: EIP-7928 BAL Ordering, Sizing Constraint & Retention Policy

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As an **EL client developer**, I want the BAL to enforce correct account ordering (lexicographic by address), `ITEM_COST=2000` sizing constraint, correct `BlockAccessIndex` assignment (0 for pre-tx system calls, 1..n for txs, n+1 for post-tx), early rejection of malicious oversized BALs, and a retention period of ≥ 3,533 epochs, so that ETH2030's BAL is fully EIP-7928-compliant and interoperable with other clients.

**Priority**: P1 | **Story Points**: 13 | **Sprint Target**: Sprint 2

### Acceptance Criteria
- Block validation rejects a BAL whose accounts are not in lexicographic order by address
- Block building rejects txs that would push `bal_items > block_gas_limit // ITEM_COST` (ITEM_COST=2000)
- Pre-execution system contract calls assigned `BlockAccessIndex=0`; txs `1..n`; post-execution `n+1`
- `G_remaining >= R_remaining * 2000` feasibility check runs every 8 txs during execution
- BAL storage layer retains BALs for at least 3,533 epochs before pruning
- `engine_getPayloadBodiesByHashV2` and `engine_getPayloadBodiesByRangeV2` return BAL data

### Tasks

#### Task SPEC-5.1 — BAL account ordering validation in block validation path
- **Description**: In `pkg/bal/validator.go` (create if missing), add `ValidateBALOrdering(bal BlockAccessList) error` that iterates all `AccountChanges` entries and verifies: (a) accounts in strict ascending lexicographic order by address, (b) storage_changes within each account in ascending order by key, (c) changes within each key in ascending order by `BlockAccessIndex`. Return error with first violation found.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (bal)
- **Testing Method**: Unit test: correctly ordered BAL → valid. Out-of-order address → error returned with offending address. Out-of-order storage key → error. `go test ./bal/... -run TestBALOrdering`.
- **Definition of Done**: Ordering validation integrated into block validation (`pkg/core/processor.go`). Tests green. EF state tests unaffected.

#### Task SPEC-5.2 — ITEM_COST=2000 BAL sizing constraint
- **Description**: In `pkg/bal/tracker.go`, add running counter `ItemCount`. After each transaction is tracked, check `ItemCount > block_gas_limit // ITEM_COST` (ITEM_COST=2000 per EIP-7928 §constants). If exceeded, return `ErrBALSizeExceeded` and exclude the offending tx from the block. In block building (`pkg/engine/`), enforce this constraint before including a tx.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (bal + engine)
- **Testing Method**: Unit test: block gas limit 30M → max 15,000 BAL items (30M/2000). Add 15,001st item → `ErrBALSizeExceeded`. `go test ./bal/... -run TestBALItemCostLimit`.
- **Definition of Done**: Constraint enforced. Block builder respects it. Tests green.

#### Task SPEC-5.3 — BlockAccessIndex 0 / 1..n / n+1 assignment
- **Description**: In `pkg/bal/tracker.go`, update `BlockAccessIndex` assignment: before any user tx executes, set index to 0 for system contract calls (EIP-6110 deposits, EIP-7002 withdrawals, EIP-7685 requests). For user txs, assign `1..n` in execution order. After all user txs, post-execution system calls get `n+1`. Wire index counter into `pkg/core/processor.go` `Process()` loop.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (bal + core)
- **Testing Method**: Integration test: block with 3 txs + pre/post system calls → verify BAL entries have correct indices: 0 for pre, 1/2/3 for txs, 4 for post. `go test ./bal/... -run TestBlockAccessIndexAssignment`.
- **Definition of Done**: Index assignment correct for all 3 categories. Tests green. Existing parallel execution tests unaffected.

#### Task SPEC-5.4 — Early rejection of malicious oversized BALs
- **Description**: In `pkg/core/processor.go`, implement the EIP-7928 §early-rejection feasibility check every 8 txs: `G_remaining >= R_remaining * 2000` where `R_remaining` is the number of undeclared storage reads not yet accessed and `G_remaining` is remaining block gas. If check fails, return `ErrBALFeasibilityViolated` and reject the block.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (core + bal)
- **Testing Method**: Test: construct adversarial block that declares 10,000 storage reads but has only 1M gas remaining → feasibility check fires and block rejected. Normal block → check passes every 8 txs without rejection. `go test ./core/... -run TestBALEarlyRejection`.
- **Definition of Done**: Feasibility check runs every 8 txs. Adversarial block rejected. Normal blocks unaffected.

#### Task SPEC-5.5 — BAL retention policy and `engine_getPayloadBodies` V2
- **Description**: In `pkg/core/rawdb/`, implement `RetainBALFor(epochs uint64)` that prevents BAL pruning before `3533` epochs (the weak subjectivity period). Add `engine_getPayloadBodiesByHashV2` and `engine_getPayloadBodiesByRangeV2` endpoints in `pkg/engine/` that return `ExecutionPayloadBodyV2` including the `blockAccessList` field.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (engine + rawdb)
- **Testing Method**: Retention test: store BAL, advance 3532 epochs → BAL still present. At 3533 epochs → eligible for pruning. Engine API test: `engine_getPayloadBodiesByHashV2` returns BAL in response body. `go test ./engine/... -run TestPayloadBodiesV2`.
- **Definition of Done**: Retention policy enforced. Both engine API methods return BAL. Tests green.

---

## US-SPEC-6: EIP-7706 Multidimensional Fee Vector Transaction Type

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As an **EVM developer**, I want a new EIP-7706 transaction type with 3-element fee vectors `[execution, blob, calldata]` for `max_fees_per_gas` and `priority_fees_per_gas`, a calldata gas calculation function, and updated block headers with 3D `gas_limits/gas_used/excess_gas` vectors, so that calldata is priced independently from execution gas (preventing calldata from crowding out computation within a block).

**Priority**: P1 | **Story Points**: 13 | **Sprint Target**: Sprint 3

### Acceptance Criteria
- New tx type (EIP-7706) accepted by txpool and included in blocks with `max_fees_per_gas` as 3-element vector
- `get_calldata_gas(calldata)` correctly computes: `tokens = zero_bytes + non_zero_bytes * 4; return tokens * 4` (CALLDATA_GAS_PER_TOKEN=4, TOKENS_PER_NONZERO_BYTE=4)
- Block header fields `gas_limits`, `gas_used`, `excess_gas` are 3-element vectors; `gas_limits[2] = gas_limits[0] // 4` (CALLDATA_GAS_LIMIT_RATIO=4)
- Per-dimension base fee updates via `fake_exponential(MIN_BASE_FEE=1, excess, target * 8)` (BASE_FEE_UPDATE_FRACTION=8)
- Calldata gas cap: tx rejected if calldata gas > `block_gas_limits[2]`

### Tasks

#### Task SPEC-6.1 — Implement EIP-7706 3D fee vector transaction type
- **Description**: In `pkg/core/types/`, add `MultiDimFeeTx` implementing `TypedTransaction` with type byte `EIP7706TxType`. Fields: `chain_id, nonce, gas_limit, to, value, data, access_list, blob_versioned_hashes, max_fees_per_gas [3]uint64, priority_fees_per_gas [3]uint64, y_parity, r, s`. Implement `RLP` encode/decode, `Cost()`, `EffectiveGasTip()` for 3D fees. Register in tx type switch.
- **Estimated Effort**: 5 SP
- **Assignee**: Go Engineer (core/types)
- **Testing Method**: Round-trip RLP test. Txpool admission test: valid 3D fee tx → admitted; fee vector length ≠ 3 → rejected. `go test ./core/types/... -run TestMultiDimFeeTx`. EF state tests unaffected.
- **Definition of Done**: Tx type defined, encoded, decoded. Txpool admits it. Tests green. ≥ 80% coverage.

#### Task SPEC-6.2 — Implement `get_calldata_gas()` and calldata cap
- **Description**: In `pkg/core/gas_utils.go` (create if missing), implement `GetCalldataGas(calldata []byte) uint64`: count zero bytes, multiply non-zero by `TOKENS_PER_NONZERO_BYTE=4`, multiply total tokens by `CALLDATA_GAS_PER_TOKEN=4`. In tx admission (`pkg/txpool/txpool.go` and block building), enforce: `GetCalldataGas(tx.Data) <= block.GasLimits[2]`.
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (core + txpool)
- **Testing Method**: Unit test: `GetCalldataGas([]byte{0x00, 0xff, 0x00})` = `(1 zero + 1 nonzero*4) * 4` + `(1 zero) * 4` ... let me compute: tokens = 2 zeros * 1 + 1 nonzero * 4 = 2+4=6; gas = 6*4=24. Test this. Block cap test: tx with calldata gas > limit → rejected. `go test ./core/... -run TestCalldataGas`.
- **Definition of Done**: `GetCalldataGas()` correct for all byte patterns (zero, nonzero, mixed). Cap enforced. Tests green.

#### Task SPEC-6.3 — Update block header with 3D gas vector fields
- **Description**: In `pkg/core/types/block.go`, add 3-element vector fields `GasLimits [3]uint64`, `GasUsed [3]uint64`, `ExcessGas [3]uint64` to `Header`. Implement `SetCallDataGasLimit()` to enforce `GasLimits[2] = GasLimits[0] / CALLDATA_GAS_LIMIT_RATIO` (ratio=4). Add fork check: before EIP-7706 fork, fields absent (use existing scalar `GasLimit/GasUsed`); after fork, vectors present.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (core/types)
- **Testing Method**: Unit test: header with `GasLimits[0]=30_000_000` → `GasLimits[2]=7_500_000`. Header RLP round-trip. Fork check: pre-fork header has no vector fields; post-fork has them. `go test ./core/types/... -run TestHeaderGasVectors`.
- **Definition of Done**: Header fields defined. Ratio constraint enforced. Fork gate correct. EF state tests unaffected.

#### Task SPEC-6.4 — 3D base fee update formula
- **Description**: In `pkg/core/multidim_gas.go`, extend the per-dimension EIP-1559 base fee update to include `DimCalldata` as the third dimension: `get_base_fee[i] = fake_exponential(MIN_BASE_FEE_PER_GAS=1, excess_gas[i], target_gas[i] * BASE_FEE_UPDATE_FRACTION=8)`. Wire calldata gas tracking into per-tx accounting: deduct `GetCalldataGas(tx.data)` from `GasUsed[2]` for each tx. Update `pkg/core/processor.go` to track all 3 dimensions.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (core)
- **Testing Method**: Unit test: block at target calldata usage → base fee unchanged. Over target → base fee increases. Under target → base fee decreases. `go test ./core/... -run TestCalldata3DBaseFee`.
- **Definition of Done**: `DimCalldata` tracked. Base fee updates correctly for all 3 dimensions. Tests green. No regression.

---

## US-SPEC-7: EIP-7864 Binary Trie Key Generation & Data Layout

**INVEST**: I✓ N✓ V✓ E✓ S✓ T✓

**User Story**:
> As a **state transition engineer**, I want the binary trie's key generation functions (`get_tree_key`, `get_tree_key_for_basic_data`, `get_tree_key_for_code_chunk`, `get_tree_key_for_storage_slot`) and the `BASIC_DATA_LEAF_KEY` 32-byte header packing verified against the EIP-7864 spec, so that all tooling (block explorers, light clients, provers) that reads the binary trie gets the correct layout.

**Priority**: P1 | **Story Points**: 8 | **Sprint Target**: Sprint 2

### Acceptance Criteria
- `get_tree_key(address32, tree_index, sub_index)` returns `blake3(address32 || tree_index_le32)[:31] || sub_index_byte`
- `BASIC_DATA_LEAF_KEY` 32-byte value packs: `version(1B) | reserved(4B) | code_size(3B) | nonce(8B) | balance(16B)` at exact byte offsets
- Code chunks use 31-byte chunks with leading PUSHDATA-count byte; chunk boundaries are tracked across PUSH data ranges
- `MAIN_STORAGE_OFFSET = 256^31` — main storage slots are keyed at tree_index ≥ `MAIN_STORAGE_OFFSET // 256`
- Empty leaf hash: if `value == [0x00]*64`, `hash = [0x00]*32`

### Tasks

#### Task SPEC-7.1 — Implement and test all `get_tree_key*` functions
- **Description**: In `pkg/trie/bintrie/keys.go` (create if missing), implement the four key generation functions from EIP-7864 spec §key-generation: `GetTreeKey(addr Address32, treeIndex int, subIndex int)`, `GetTreeKeyForBasicData(addr)`, `GetTreeKeyForCodeChunk(addr, chunkID)`, `GetTreeKeyForStorageSlot(addr, storageKey)`. Use BLAKE3 (from US-EL-1 / US-SPEC dependency on `lukechampine.com/blake3`). Inline storage: slots 0–63 at subindex 64–127; code at 128–255; main storage at `MAIN_STORAGE_OFFSET + slot`.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (trie)
- **Testing Method**: Table-driven tests against EIP-7864 spec examples. Key uniqueness test: 1000 random (address, slot) pairs → all distinct keys. `get_tree_key_for_storage_slot(addr, 0)` vs `get_tree_key_for_storage_slot(addr, 64)` → different subindices (64 vs 128). `go test ./trie/bintrie/... -run TestTreeKeyGeneration`.
- **Definition of Done**: All 4 key functions implemented. Spec example vectors pass. Key uniqueness verified. ≥ 80% coverage.

#### Task SPEC-7.2 — Verify BASIC_DATA_LEAF_KEY header packing
- **Description**: In `pkg/trie/bintrie/account.go` (or equivalent), implement `PackBasicDataLeaf(version uint8, codeSize uint32, nonce uint64, balance *big.Int) [32]byte` and `UnpackBasicDataLeaf([32]byte) (version, codeSize, nonce, balance)` following EIP-7864 spec §header-layout: offset 0: version (1B), offsets 1-4: reserved (4B, zero), offsets 5-7: code_size (3B big-endian), offsets 8-15: nonce (8B big-endian), offsets 16-31: balance (16B big-endian).
- **Estimated Effort**: 2 SP
- **Assignee**: Go Engineer (trie)
- **Testing Method**: Round-trip test: pack known (version=1, codeSize=100, nonce=5, balance=1 ETH) → unpack → identical values. Offset test: verify each field is at the exact byte offset specified. `go test ./trie/bintrie/... -run TestBasicDataLeafPacking`.
- **Definition of Done**: Pack/unpack correct. Byte offsets verified. Round-trip passes.

#### Task SPEC-7.3 — Code chunking: 31-byte chunks with PUSHDATA boundary tracking
- **Description**: In `pkg/trie/bintrie/code_chunker.go` (create if missing), implement `ChunkifyCode(code []byte) [][32]byte`: split code into 31-byte chunks, prepend each chunk with a 1-byte `leadingPUSHDATABytes` count. The leading byte counts how many bytes at the start of the chunk are PUSHDATA (not opcodes) — tracking PUSH1–PUSH32 instruction ranges across chunk boundaries. This follows EIP-7864 §code-chunking exactly.
- **Estimated Effort**: 3 SP
- **Assignee**: Go Engineer (trie)
- **Testing Method**: Test 1: `PUSH1 0x60 ADD` → chunk 0 has `leadingPUSHDATABytes=0` (PUSH1 is an opcode). Test 2: code with `PUSH32` spanning a chunk boundary → next chunk has `leadingPUSHDATABytes=N`. Regression test: re-chunking the same code always produces identical output. `go test ./trie/bintrie/... -run TestCodeChunker`.
- **Definition of Done**: PUSH boundary tracking correct for PUSH1–PUSH32. Re-chunking is deterministic. `go test ./trie/bintrie/...` fully green.

---

---

# Sprint Planning Summary

## Suggested Sprint Breakdown

| Sprint | Focus | Stories | SP |
|--------|-------|---------|-----|
| Sprint 1 | P0 blockers | US-AA-1(13), US-AA-3(12), US-BL-1(12), US-EL-4(10), US-PQ-3(13), US-GAP-7(10), US-LEAN-1(8) | 78 |
| Sprint 2 | Interop + core spec | US-AA-2(8), US-EL-2(13), US-EL-3(10), US-PQ-2(8), US-LEAN-2(8), US-GAP-3(13), US-SPEC-1(13) | 73 |
| Sprint 3 | Spec compliance + gaps | US-PQ-4(8), US-SPEC-3(8), US-SPEC-4(8), US-SPEC-5(13), US-SPEC-7(8), US-GAP-1(13), US-GAP-2(8) | 66 |
| Sprint 4 | PQ hardening + 3D gas | US-AA-4(5), US-AA-5(9), US-PQ-6(8), US-GAP-5(13), US-SPEC-6(13), US-LEAN-3a(6), US-LEAN-3b(8) | 62 |
| Sprint 5 | STARK + networking | US-PQ-5a(8), US-PQ-5b(13), US-LEAN-5(8), US-LEAN-6(13), US-GAP-4(13) | 55 |
| Sprint 6 | Research + finality + P2P | US-LEAN-4(13), US-LEAN-8(5), US-BB-1(13), US-BB-2(5) | 36 |

> **Note**: Sprint 1 (78 SP) contains the hard P0 blockers: US-PQ-3 (NTT address fix), US-EL-4 (KZG backend), US-BL-1 (BLAKE3 backend), US-GAP-7 (BLS backend). These should be tackled by senior engineers first. US-SPEC-1 (frame TX full compliance, 13 SP) moved to Sprint 2 — it was previously a Sprint 1 entry but was deprioritised to keep Sprint 1 focused on infrastructure. US-GAP-6 and US-LEAN-7 have been merged into US-GAP-5 and US-GAP-3 respectively and are no longer separate sprint items.

---

## Role Allocation Guide

| Role | Primary Stories |
|------|----------------|
| **Go Engineer (vm/core)** | US-GAP-1, US-GAP-2, US-EL-2, US-EL-3, US-PQ-3, US-AA-3, US-BB-2, US-BL-1 |
| **Consensus Engineer** | US-LEAN-1, US-LEAN-2, US-LEAN-3a, US-LEAN-3b, US-LEAN-6, US-LEAN-8, US-GAP-3, US-GAP-5 |
| **ZK Engineer** | US-PQ-5a, US-PQ-5b, US-PQ-6, US-EL-2 |
| **P2P Engineer** | US-LEAN-4, US-LEAN-5, US-GAP-4, US-PQ-4, US-BB-1 |
| **Security Engineer** | US-AA-5, US-PQ-3, US-AA-1 |
| **Protocol Researcher** | US-BB-2, US-LEAN-6 |
| **QA / DevOps Engineer** | US-EL-4, US-LEAN-1, US-GAP-7, US-PQ-4, all benchmark tasks |

---

## Definition of Done — Global Criteria

All stories must meet the following DoD in addition to story-specific criteria:

1. **Compilation**: `cd pkg && go build ./...` passes with zero errors.
2. **Tests**: `cd pkg && go test ./...` passes with zero failures.
3. **Coverage**: New code has ≥ 80% line coverage via `go test -cover`.
4. **Format**: `go fmt ./...` produces no diff.
5. **EF State Tests**: 36,126/36,126 still passing after any EVM/state changes.
6. **Code Review**: At least one engineer other than the implementor reviews and approves.
7. **Docs**: Any new CLI flags documented in `cmd/eth2030/main.go` help text.
8. **Commit hygiene**: Each commit is atomic, < 40-char subject, passes CI.
9. **No Claude attribution**: No `Co-Authored-By` AI lines in commits.

---

*Generated: 2026-03-04. Sources: docs/plans/leanroadmap-coverage-2026-03.md, docs/plans/vitalik/\*.*
