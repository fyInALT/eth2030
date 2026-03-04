> [← Back to Sprint Index](README.md)

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
