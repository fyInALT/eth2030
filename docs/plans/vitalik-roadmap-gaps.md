# Vitalik Roadmap Gap Analysis — Fast Slots, Fast Finality, Scaling

> **Source:** Vitalik's "Fast Slots, Fast Finality, Scaling" roadmap document (Feb 2026)
> **Method:** Line-by-line comparison of Vitalik's proposals against ETH2030 codebase
> **Date:** 2026-02-28

---

## Summary

| Category | Matching | Different Approach | Missing | Total |
|----------|----------|-------------------|---------|-------|
| Fast Slots | 3 | 0 | 2 | 5 |
| Fast Finality | 4 | 1 | 1 | 6 |
| Scaling / Gas | 6 | 1 | 1 | 8 |
| **Total** | **13** | **2** | **4** | **19** |

---

## Matching Items (13)

| # | Vitalik's Proposal | ETH2030 Implementation | Key Files |
|---|-------|-----|-----------|
| 1 | 6-second slots | `QuickSlotConfig{SlotDuration: 6s, SlotsPerEpoch: 4}` | `consensus/quick_slots.go:25-40` |
| 2 | 1-epoch finality | SSF + endgame pipeline (<500ms target) | `consensus/ssf.go`, `consensus/endgame_pipeline.go` |
| 3 | ePBS + FOCIL (complex slot structure) | ePBS builder API + FOCIL inclusion lists | `epbs/`, `focil/` |
| 4 | BLS signature aggregation | Parallel BLS (16 workers, 4096 batch) | `consensus/parallel_bls.go:96-185` |
| 5 | PQ crypto for finality | FinalityBLSAdapter with PQ fallback | `consensus/finality_bls_adapter.go:46` |
| 6 | Dilithium attestations | PQ attestations with STARK aggregation | `consensus/pq_attestation.go`, `consensus/stark_sig_aggregation.go` |
| 7 | BALs for parallel execution | Block Access Lists (5,366 LOC) | `bal/` |
| 8 | Multidimensional gas | 5-dim pricing (Compute/Storage/Bandwidth/Blob/Witness) | `core/multidim_gas.go:20-36` |
| 9 | SSTORE repricing | Glamsterdam: Set 5000, Reset 1500 + EIP-8037 state creation | `core/glamsterdam_repricing.go:27-28`, `core/vm/gas_table.go:113-131` |
| 10 | ZK-EVM (RISC-V guest) | zkVM framework with RV32IM executor | `zkvm/` (15,405 LOC) |
| 11 | 3-of-5 mandatory proofs | Prover assignment, submission, verification, penalties | `proofs/mandatory.go` (14,323 LOC) |
| 12 | PeerDAS data availability | 128-column DAS, cell gossip, blob reconstruction | `das/` (43,249 LOC) |
| 13 | Poseidon hash (ZK circuits) | Full Poseidon1 over BN254 with sponge construction | `zkvm/poseidon.go` (293 lines) |

## Different Approach (2)

| # | Vitalik's Proposal | ETH2030 Approach | Gap Severity |
|---|-------|------|------|
| 14 | Block-level erasure coding (8-piece k-of-n for propagation) | Blob-level PeerDAS (128-column GF(2^8) for data availability) + block-in-blobs chunking (no erasure) | MEDIUM — blob-level RS exists, block-level needs separate encoder |
| 15 | Minimmit one-round BFT | SSF (4-phase) + endgame pipeline (3-sub-slot) | MEDIUM — SSF is functionally equivalent but multi-round |

## Missing Items (4)

| # | Vitalik's Proposal | Gap | Severity | Plan |
|---|-------|-----|------|------|
| 16 | Gas reservoir mechanism (GAS returns regular only, CALL forwards reservoir) | No `StateGasReservoir` field in Contract/EVM; all gas is single counter | MEDIUM | [US-1.1](vitalik-roadmap/US-1.1-gas-reservoir-mechanism.md) |
| 17 | SSTORE zero→nonzero in separate gas dimension | Detection exists but charges to single gas counter, not DimStorage | MEDIUM | [US-1.2](vitalik-roadmap/US-1.2-sstore-state-creation-dimension.md) |
| 18 | Random attester sampling (256-1024 per slot) | Full committee shuffles only; no random subset sampling | MEDIUM | [US-2.1](vitalik-roadmap/US-2.1-random-attester-sampling.md) |
| 19 | 8s intermediate slot step (sqrt(2) progression) | Infrastructure supports variable durations but no 8s config | LOW | [US-2.2](vitalik-roadmap/US-2.2-intermediate-8s-slot-step.md) |

## Additional Plans (Different Approaches)

| # | Plan |
|---|------|
| 14 | [US-3.1 Block-Level Erasure Coding](vitalik-roadmap/US-3.1-block-level-erasure-coding.md) |
| 15 | [US-4.1 Minimmit One-Round BFT](vitalik-roadmap/US-4.1-minimmit-one-round-bft.md) |
| — | [US-4.2 Poseidon2 Hash Backend](vitalik-roadmap/US-4.2-poseidon2-hash-backend.md) |

---

## Detailed Gap Analysis

### EP-1: Multidimensional Gas — Reservoir Mechanism

**Vitalik's proposal:** Separate "state creation gas" from regular execution gas with a reservoir mechanism.

**What ETH2030 has:**
- 5-dimensional gas pricing engine (`multidim_gas.go:20-36`): Compute, Storage, Bandwidth, Blob, Witness
- Per-dimension EIP-1559 base fee adjustment (`multidim_gas.go:294-333`)
- SSTORE zero→nonzero detection in 3 gas calculators (`gas_table.go:253`, `evm_storage_ops.go:119`, `dynamic_gas.go:186`)
- EIP-8037 state creation constants defined (`gas_table.go:113-131`)

**What's missing:**
- `StateGasReservoir` field on `Contract` — all gas is a single `uint64` counter
- GAS opcode returns total gas (should exclude reservoir)
- CALL forwards gas via 63/64 rule only (should pass full reservoir)
- SSTORE charges to single counter (should draw from reservoir for zero→nonzero)

**Files to modify:** `pkg/core/vm/interpreter.go` (Contract struct), `pkg/core/vm/instructions.go:490` (opGas), `pkg/core/vm/instructions.go:752` (opCall), `pkg/core/vm/gas_table.go:234` (SstoreGas)

---

### EP-2: Fast Slots — Random Attester Sampling

**Vitalik's proposal:** Replace full committee attestation with 256-1024 random attesters per slot.

**What ETH2030 has:**
- Full committee shuffle (90-round swap-or-not): `committee_assignment.go:122-168`
- 128K attester cap with epoch committee rotation: `committee_rotation.go:175-264`
- Parallel BLS aggregation (16 workers): `parallel_bls.go:96-185`
- 3-phase slots (Proposal 2s + Attestation 2s + Aggregation 2s): `phase_timer.go:36-44`

**What's missing:**
- `RandomAttesterSelector` for N-element subset sampling
- Committee-less attestation format (no CommitteeBits)
- 2-phase slot mode (aggregation phase eliminated)
- Weight scaling in fork-choice for sampled attesters

**Files to modify:** `pkg/consensus/committee_assignment.go`, `pkg/consensus/attestation.go`, `pkg/consensus/phase_timer.go`, `pkg/consensus/ssf.go`

---

### EP-3: Block Propagation — Erasure Coding

**Vitalik's proposal:** Split execution blocks into 8 erasure-coded pieces for faster propagation.

**What ETH2030 has:**
- `RSEncoderGF256` — production GF(2^8) Reed-Solomon for blob columns: `das/erasure/reed_solomon_encoder.go:37-94`
- sqrt(n) block fanout: `p2p/block_gossip.go:133`
- Block-in-blobs sequential chunking: `das/block_in_blob.go:127-171`
- Blob reconstruction from 64/128 cells: `das/reconstruction.go:160-233`

**What's missing:**
- `BlockErasureEncoder` / `BlockErasureDecoder` wrapping RSEncoderGF256
- `BlockPiece` gossip topic with per-piece routing
- `BlockAssemblyManager` for concurrent piece collection and reconstruction
- Pipeline integration for piece-based block reception

**Files to create:** `pkg/das/block_erasure.go`, `pkg/p2p/block_piece_gossip.go`

---

### EP-4: Finality Protocol — Minimmit + Poseidon2

**Vitalik's proposal:** Minimmit one-round BFT for faster finality; Poseidon2 for ZK circuits.

**What ETH2030 has:**
- SSF with 4-phase state machine: `ssf.go` (244 lines), `ssf_round_engine.go`
- Endgame pipeline with <500ms target: `endgame_pipeline.go` (395 lines)
- BLS adapter with PQ fallback: `finality_bls_adapter.go` (330 lines)
- Poseidon1 over BN254: `zkvm/poseidon.go` (293 lines)
- HashBackend interface: `crypto/pqc/hash_backend.go:9-19`

**What's missing:**
- `MinimmitEngine` — no one-round BFT (0 references in codebase)
- Poseidon2 permutation (external/internal round separation)
- `Poseidon2Backend` implementing HashBackend
- `FinalityMode` enum for protocol selection

**Files to create:** `pkg/consensus/minimmit.go`, `pkg/zkvm/poseidon2.go`

---

## Story Point Summary

| Epic | Story | SP |
|------|-------|----|
| EP-1 | US-1.1 Gas Reservoir | 13 |
| EP-1 | US-1.2 SSTORE Dimension | 8 |
| EP-2 | US-2.1 Random Attesters | 13 |
| EP-2 | US-2.2 8s Slot Step | 5 |
| EP-3 | US-3.1 Block Erasure | 13 |
| EP-4 | US-4.1 Minimmit BFT | 13 |
| EP-4 | US-4.2 Poseidon2 | 8 |
| **Total** | | **73** |

---

## Plan Files

All per-story plans are in [`docs/plans/vitalik-roadmap/`](vitalik-roadmap/):

```
docs/plans/vitalik-roadmap/
├── README.md
├── US-1.1-gas-reservoir-mechanism.md
├── US-1.2-sstore-state-creation-dimension.md
├── US-2.1-random-attester-sampling.md
├── US-2.2-intermediate-8s-slot-step.md
├── US-3.1-block-level-erasure-coding.md
├── US-4.1-minimmit-one-round-bft.md
└── US-4.2-poseidon2-hash-backend.md
```

Each story file follows the same format as `docs/plans/eip-7928/` and `docs/plans/eip-8141/`:
- User story (INVEST format)
- Tasks with effort estimates
- Codebase locations table (actual file paths and line numbers)
- Implementation status (verified against code)
- Gap analysis with proposed solutions
- Spec reference excerpts
