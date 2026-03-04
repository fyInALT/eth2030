> [← Back to Sprint Index](README.md)

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
