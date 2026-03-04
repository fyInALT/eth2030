# leanroadmap Coverage Analysis

**Source**: `refs/leanroadmap/` — https://leanroadmap.org (leanEthereum / Justin Drake)
**Date**: 2026-03-04
**Method**: Line-by-line read of all data files in `refs/leanroadmap/data/`; cross-checked against ETH2030 codebase.

leanroadmap tracks 8 research tracks, 4 pq-devnets (3 completed, 1 active, 1 planned), and 2 benchmark categories. ETH2030 is an EL + CL hybrid client, while leanroadmap focuses on the CL (Lean Consensus). This analysis maps each leanroadmap feature to ETH2030 coverage.

---

## Summary

| Track | leanroadmap progress | ETH2030 status | Gap severity |
|---|---|---|---|
| Hash-Based Multi-Signatures | 70% | PARTIAL | HIGH |
| PQ Sig Aggregation (zkVMs) | 50% | PARTIAL | HIGH |
| Poseidon Cryptanalysis | 50% | PARTIAL | MEDIUM |
| Falcon Signatures | 10% | PARTIAL | LOW |
| Formal Verification (Lean 4) | 40% | MISSING | MEDIUM |
| P2P Networking (Gossipsub v2) | 30% | PARTIAL | HIGH |
| Attester-Proposer Separation | 20% | PARTIAL | MEDIUM |
| Faster Finality (3SF) | 50% | PARTIAL | MEDIUM |

**Overall**: ETH2030 covers the cryptographic primitives and basic consensus well but is missing the leanEthereum-specific protocol designs: exact leanSig/leanMultisig API format, separate aggregator role, leanVm recursive aggregation, Gossipsub V2, rateless set reconciliation, grid topology, exit queue Minslack, and 4-second slots.

---

## Track 1: Hash-Based Multi-Signatures

**leanroadmap**: Winternitz XMSS as PQ replacement for BLS. Reference impl: [leanSig (Rust)](https://github.com/leanEthereum/leanSig).
**Lead**: Benedikt Wagner. Progress: 70%.

### leanroadmap milestones

| Milestone | Status |
|---|---|
| Paper publication (EPRINT 2025/055) and prototype | DONE (2025-01) |
| Efficiency analysis of hash-based sig candidates | DONE (2025-03) |
| Further optimizations exploration | DONE |
| Identification of alternatives for PQ multi-sigs | DONE |
| **Fixing parameters (key lifetime)** | **PENDING** |

### leanSig exact benchmarks (from `refs/leanroadmap/data/benchmarks.ts`)

| Metric | Target | Current best | Status |
|---|---|---|---|
| Public key size | — | **50 bytes** | (8-element root + 5 randomiser) |
| Signature size | — | **3 KiB** | |
| Key generation time | — | **3.5 hours** | 10-core M1, 8-year lifetime |
| Signing time | **500 μs** | 535.2 μs (Hashing Optimized) | 107% — just over target |
| Verification time | **500 μs** | 193.42 μs (Hashing Optimized) | 39% — well within target ✓ |

### ETH2030 coverage

**Has:**
- `pkg/crypto/pqc/unified_hash_signer.go` — XMSS + WOTS+ (SHA-256 backed, tree heights 10/16/20)
- `pkg/crypto/pqc/l1_hash_sig.go` / `l1_hash_sig_v2.go` — Winternitz OTS tree signer (Keccak256)
- `pkg/crypto/pqc/hash_backend.go` — pluggable hash backend (SHA-256 / Keccak / Blake3-placeholder)
- `pkg/consensus/pq_attestation.go` — PQ attestations with Dilithium3/ML-DSA fallback
- Key exhaustion tracking, XMSSKeyManager multi-tree

**Missing:**

| Gap | Description | Priority |
|---|---|---|
| **leanSig exact public key format** | leanSig uses 50-byte pubkey (8 root elements + 5 randomiser). ETH2030 XMSS key size is not aligned with this format. | HIGH |
| **Key lifetime parameter finalization** | leanroadmap has "Fixing parameters such as key lifetime" as PENDING. ETH2030 hardcodes tree heights (H=10/16/20). Need to align with leanSpec's final lifetime choice. | HIGH |
| **Signing time target (500 μs)** | Current leanSig best is 535 μs (barely over). ETH2030 has no signing performance target or benchmark. Add `pkg/crypto/pqc/leansig_bench_test.go`. | MEDIUM |
| **leanSig Rust interop** | `refs/leanSig` is the actively maintained Rust implementation. ETH2030 Go implementation needs to produce identical signatures for multi-client interop (pq-devnet-1 showed client interop is possible). | HIGH |

---

## Track 2: Post-Quantum Signature Aggregation with zkVMs

**leanroadmap**: Minimal zkVMs for aggregating hash-based signatures. Options: Binius M3, SP1, KRU, STU, Jolt, OpenVM. leanMultisig is the reference implementation.
**Lead**: Thomas Coratger. Progress: 50%.

### leanroadmap milestones

| Milestone | Status |
|---|---|
| Benchmark hash in SNARK (Plonky3, STwo, Binius, Hashcaster) | DONE (2025-02) |
| Hashcaster exploration | DONE (2025-02) |
| Snarkify hash-based sig agg with SP1 & OpenVM | DONE (2025-02) |
| **Explore GKR-style provers** | **PENDING** |
| **Explore WHIR** | **PENDING** |
| **More explorations over binary field techniques** | **PENDING** |

### leanMultisig benchmarks (from `refs/leanroadmap/data/benchmarks/xmss-aggregation.ts`)

| Metric | Target | M4 Max (Efficient) | M4 Max (Simple) | i9-12900H |
|---|---|---|---|---|
| XMSS aggregated/sec | **1,000/sec** | ~970/sec (97% ✓) | ~815/sec (82%) | ~380/sec (38%) |
| Aggregate size | **128 KiB** | ~400–500 KiB (312–391% ✗) | ~300 KiB (234% ✗) | — |

**The aggregate size target (128 KiB) is NOT yet met** — current sizes are 2–4× over target. This is the critical blocker for pq-devnet-4.

### pq-devnet goals requiring aggregation (from `refs/leanroadmap/data/devnets.ts`)

| Devnet | Goal | Status |
|---|---|---|
| pq-devnet-2 (Jan 2026) | Integrate leanMultisig aggregation | COMPLETED |
| pq-devnet-3 (Feb 2026) | **Separate aggregator role** from block production | ACTIVE |
| pq-devnet-4 (Mar 2026) | **Recursive PQ sig aggregation using leanVm**; single aggregate per message | PLANNED |

### ETH2030 coverage

**Has:**
- `pkg/consensus/stark_sig_aggregation.go` — `STARKSignatureAggregator`: N PQ attestations → single STARK aggregate proof
- `pkg/consensus/jeanvm_aggregation.go` — jeanVM Groth16 ZK-circuit BLS aggregation
- `pkg/proofs/recursive_prover.go` — binary-tree recursive proof composition
- `pkg/zkvm/` — full RISC-V zkVM framework (22 files)

**Missing:**

| Gap | Description | Priority |
|---|---|---|
| **leanMultisig protocol** | leanMultisig is XMSS-signature-specific aggregation (not STARK over general PQ proofs). ETH2030's `stark_sig_aggregation.go` aggregates proof validity; leanMultisig aggregates the XMSS signatures themselves via a zkVM. These are architecturally different. | HIGH |
| **Separate aggregator role** (pq-devnet-3) | Lean Consensus decouples aggregation from block production: there is a distinct "aggregator" network role that collects per-validator XMSS sigs and produces a single aggregate, then propagates it. ETH2030 has no such role. | HIGH |
| **Recursive aggregation (leanVm)** (pq-devnet-4) | leanVm is the minimal zkVM used to coalesce multiple aggregates for the same message into one final aggregate. ETH2030's `jeanvm_aggregation.go` exists but is Groth16/BLS-based, not XMSS-specific recursive aggregation. | HIGH |
| **Aggregate size reduction** | Current leanMultisig is 300–500 KiB; target is 128 KiB. ETH2030 has no aggregate size budget enforcement in the consensus layer. | MEDIUM |
| **GKR-style provers / WHIR** | These are advanced proving systems for binary-field hash-based SNARK aggregation (still pending in leanroadmap). ETH2030 has Groth16 + STARK but not GKR or WHIR. | LOW |

---

## Track 3: Poseidon Cryptanalysis Initiative

**leanroadmap**: EF-sponsored security analysis of the Poseidon hash function before committing to it as the binary tree hash.
**Lead**: Dmitry Khovratovich. Progress: 50%.

### leanroadmap milestones

| Milestone | Status |
|---|---|
| Bounties established ($66k already earned) | DONE (2025-01) |
| Research grants (three recipients chosen) | DONE (2025-02) |
| Workshop at FSE 2025 | DONE (2025-03) |
| **Workshop at Algebraic Hash Cryptanalysis Days** | **PENDING (2025-05)** |
| **Groebner basis explorations** | **PENDING** |

**Status**: Multiple attack papers published (EPRINT 2025/937, 2025/950, 2025/954). The Groebner basis / Graeffe transform attacks are active research. Final decision on "Ethereum's last hash function" is pending this initiative.

### ETH2030 coverage

**Has:**
- `pkg/zkvm/poseidon.go` — Poseidon1 over BN254 (t=3, 8 full + 57 partial rounds, Grain LFSR)
- `pkg/zkvm/poseidon2.go` — Poseidon2 over BN254 (diagonal MDS, external/internal rounds)

**Missing:**

| Gap | Description | Priority |
|---|---|---|
| **Hash function decision tracking** | ETH2030 commits to SHA-256 for the binary tree (`pkg/trie/bintrie/`). The Poseidon cryptanalysis initiative is still ongoing (EPRINT 2025/937–954 attacks). ETH2030 should defer finalizing the binary tree hash until the initiative concludes. | HIGH — design risk |
| **Poseidon2 extra-rounds conservatism** | leanroadmap cites "Poseidon2 + extra rounds (Monolith non-arithmetic layers)" as a conservative option. ETH2030's `poseidon2.go` uses default round counts — no extra-rounds variant. Add `Poseidon2ExtraRoundsParams`. | MEDIUM |
| **Monolith hash backend** | Poseidon2 + lookup non-arithmetic layer (Monolith). Not implemented. | LOW |
| **Groebner basis benchmark** | No cryptanalysis test suite for verifying Poseidon's security against algebraic attacks. | LOW |

---

## Track 4: Falcon Signatures

**leanroadmap**: Lattice-based alternative to hash-based sigs. Advantage: ~5× more validators (smaller sigs). Currently **inactive** (10% progress).
**Lead**: Josh Beal.

### leanroadmap milestones

| Milestone | Status |
|---|---|
| Proposal for Falcon signature aggregation | DONE (2025-02) |
| **Combining Falcon with code-based SNARKs** | **PENDING** |

### ETH2030 coverage

**Has:**
- `pkg/crypto/pqc/falcon_signer.go` — Falcon512 signer
- `pkg/core/vm/precompile_ntt.go` — NTT precompile (accelerates Falcon on-chain)
- Gas cost model: `pkg/core/types/pq_transaction.go` (`Falcon=12000 gas`)
- 3 security CVEs in refs `ethfalcon/` (CRITICAL, MEDIUM, LOW — see `vitalik-pq-roadmap-gap-analysis.md`)

**Missing:**

| Gap | Description | Priority |
|---|---|---|
| **Falcon signature aggregation** | Aggregating multiple Falcon signatures into a compact aggregate (via code-based SNARKs / LaBRADOR). leanroadmap ref: EPRINT 2024/311, ethresear.ch lattice-based sig aggregation. No such protocol in ETH2030. | LOW (track inactive) |
| **Falcon CVE fixes** | `CVETH-2025-080201` (CRITICAL): salt size not checked. Must verify ETH2030 Go C-bindings are not affected before using Falcon in production. | HIGH |
| **NTT precompile address alignment** | ETH2030 NTT is at `0x15`; ntt-eip spec defines `0x0f–0x12`. This breaks Falcon on-chain verification compatibility with the spec. | HIGH |

---

## Track 5: Formal Verification (Lean 4)

**leanroadmap**: Mathematically prove security of FRI, STIR, WHIR using Lean 4 framework.
**Lead**: Alex Hicks. Progress: 40%.

### leanroadmap milestones

| Milestone | Status |
|---|---|
| zkEVM formal verification project initiation | DONE (2025-01) |
| **Lean 4 framework implementation** | **PENDING (2025-03)** |
| **FRI proof system specification** | **PENDING** |
| **STIR proof system specification** | **PENDING** |
| **WHIR proof system specification** | **PENDING** |

**Resources**: EF's [verified-zkevm.org](https://verified-zkevm.org/), ArkLib Lean blueprint.

### ETH2030 coverage

**Has:**
- `pkg/zkvm/constraint_compiler.go` — R1CS constraint generation
- `pkg/zkvm/r1cs_solver.go` — R1CS solving
- `pkg/proofs/stark_prover.go` — STARK proof generation (FRI-based over Goldilocks)

**Missing:**

| Gap | Description | Priority |
|---|---|---|
| **Lean 4 formal verification** | ETH2030 has no Lean 4 proofs of any proof system correctness. leanroadmap prioritizes formal proofs of FRI, STIR, WHIR. This is a research-level activity, not Go implementation. | LOW — research |
| **STIR / WHIR proof systems** | STIR and WHIR are next-generation polynomial IOP systems with better proof sizes than FRI. ETH2030 only has FRI-based STARK. WHIR reference: https://gfenzi.io/papers/whir/ (2024-09). | MEDIUM |
| **ArkLib integration** | ArkLib is the formally verified SNARK library for Lean 4. No integration path with ETH2030 Go code exists. | LOW |

---

## Track 6: P2P Networking

**leanroadmap**: Next-gen P2P for 4-second slots, 1M+ validators, Gossipsub v2.0.
**Lead**: Raúl Kripalani. Progress: 30%.

### leanroadmap milestones

| Milestone | Status |
|---|---|
| Practical Rateless Set Reconciliation research | DONE (2024-02) |
| **Generalized Gossipsub specification** | **PENDING** |
| **Gossipsub V2 specification** | **PENDING** |
| **Grid Topology research** | **PENDING** |
| **libp2p in C development** | **PENDING** |
| **libp2p in Zig development** | **PENDING** |

### ETH2030 coverage

**Has:**
- `pkg/p2p/gossip_topics.go` + gossip protocol (pub/sub, scoring, banning, dedup)
- `pkg/p2p/discovery_v5.go` — Kademlia V5 ENR discovery
- `pkg/p2p/mixnet_transport.go` — 3-hop simulated mixnet
- `pkg/p2p/flashnet_transport.go` — broadcast transport
- `pkg/p2p/portal_network.go` — content DHT

**Missing:**

| Gap | Description | Priority |
|---|---|---|
| **Gossipsub V2.0** | Spec at `refs/libp2p/specs/pull/653`. V2 includes improved score functions, opportunistic grafting, and message prioritization. ETH2030 uses Gossipsub v1.x equivalent. Required for 1M validators and 4s slots. | HIGH |
| **Rateless Set Reconciliation** | Paper: arXiv 2402.02668. Allows nodes to sync mempool / attestation sets without knowing what their peer has, in O(difference) communication. Essential for efficient propagation at scale. ETH2030 has no set reconciliation protocol. | HIGH |
| **Grid Topology** | Structured grid mesh instead of random topology. Reduces propagation hops and improves message delivery guarantees. HackMD reference in leanroadmap. No grid topology in ETH2030. | MEDIUM |
| **QUIC transport** | pq-devnet-0 used QUIC for P2P. ETH2030 uses TCP/RLPx. QUIC provides 0-RTT reconnection and multiplexing critical for 4s slots. | MEDIUM |
| **Generalized Gossipsub** | Parameterized gossipsub design covering different message types (blocks, attestations, ILs, aggregates) with different D values. Spec at libp2p/specs/pull/664. | MEDIUM |
| **libp2p in C / Zig** | These are for lean consensus client diversity (Lantern / Zeam). Not directly an ETH2030 issue but needed for interop. | LOW |

---

## Track 7: Attester-Proposer Separation (APS)

**leanroadmap**: Separates proposers and attesters to reduce centralization. Rainbow Staking model (1 ETH includers).
**Lead**: TBD. Progress: 20%.

### leanroadmap milestones

| Milestone | Status |
|---|---|
| **Exploratory research** | **PENDING** |

**Resources**: APS Tracker (notion), "Unbundling Staking: Towards Rainbow Staking" (ethresear.ch 2024-02), "Rainbow Roles Incentives: ABPS, FOCIL + AS" (2025-02).

### ETH2030 coverage

**Has:**
- `pkg/consensus/` — APS mentioned in CLAUDE.md: "APS (committee selection)" at L+ milestone
- `pkg/focil/` — FOCIL + ePBS covers partial proposer outsourcing
- 1 ETH includers mentioned in L+ roadmap

**Missing:**

| Gap | Description | Priority |
|---|---|---|
| **Rainbow Staking model** | The full Rainbow Staking design: heavy stakers (32 ETH, full attesters) vs light stakers (1 ETH, only block inclusion). ETH2030 only has `MIN_ACTIVATION_BALANCE` changes; no distinct staker-tier protocol. | MEDIUM |
| **Separate proposer selection mechanism** | "Toward a General Model for Proposer Selection Mechanism Design" (ethresear.ch 2025-02) proposes multi-party proposer selection. ETH2030 uses standard proposer duties from beacon state. | MEDIUM |
| **Exit queue flexibility (Minslack)** | "Adding Flexibility to Ethereum's Exit Queue" (ethresear.ch 2025-04): Minslack proposal allows faster exits when queue is short. ETH2030 has no custom exit queue logic beyond standard EIP-7251 max effective balance. | MEDIUM |
| **ABPS (Attested Block Proposer Separation)** | Specific variant of APS where block builders must have their blocks attested separately. Covered by ePBS + FOCIL combined but not explicitly as ABPS. | LOW |

---

## Track 8: Faster Finality (3SF)

**leanroadmap**: 3-slot finality (3SF) — reduce Ethereum finality from ~15 minutes to seconds.
**Lead**: Barnabé Monnot. Progress: 50%.

### leanroadmap milestones

| Milestone | Status |
|---|---|
| **Exploratory research (3SF integration with ePBS, FOCIL, PeerDAS)** | **PENDING** |

**Resources**: "3-Slot Finality: SSF Is Not About Single Slot" (ethresear.ch 2024-11), "Integrating 3SF with ePBS, FOCIL, and PeerDAS" (ethresear.ch 2025-08), "LMD GHOST with ~256 Validators" (ethresear.ch 2025-08).

### ETH2030 coverage

**Has:**
- `pkg/consensus/ssf.go` — 4-phase SSF with `latest_justified_slot` → `latest_finalized_slot`
- `pkg/consensus/endgame_pipeline.go` — <500ms finality target
- `pkg/consensus/quick_slots.go` — `QuickSlotConfig{SlotDuration: 6s, SlotsPerEpoch: 4}`
- `pkg/consensus/finality_bls_adapter.go` — BLS adapter with PQ fallback

**Missing:**

| Gap | Description | Priority |
|---|---|---|
| **4-second slots** | pq-devnet-0 confirmed 4s slots in multi-client interop. ETH2030 has `SlotDuration: 6s`. leanroadmap targets 4s. Need to add `SlotDuration: 4s` config and validate finality timing. | HIGH |
| **3SF-mini backoff algorithm** | The `is_justifiable_slot()` backoff (allow justification at delta ≤5, perfect squares, oblong numbers) from `refs/research/3sf-mini/consensus.py`. ETH2030 uses a different 4-phase SSF finality protocol. | MEDIUM |
| **LMD GHOST with 256 validators** | leanroadmap milestone references "LMD GHOST with ~256 validators and a fast-following finality gadget" (2025-08). This is a reduced-committee fork-choice variant for 4s slots. ETH2030 uses full committee shuffle. | MEDIUM |
| **3SF + ePBS + FOCIL + PeerDAS integration** | Full integration of all four sub-protocols. Each is implemented separately in ETH2030 but the combined slot structure (ePBS bid → FOCIL IL → PeerDAS sampling → 3SF justification) is not explicitly specified. | HIGH |

---

## pq-devnet Feature Gaps

These are concrete features proven by working devnets that ETH2030 does not have:

| Devnet | Proven feature | ETH2030 status |
|---|---|---|
| pq-devnet-0 (completed) | 4s slots, QUIC, Gossipsub v1.0, 3SF-mini, multi-client interop | 6s slots, no QUIC, partial 3SF |
| pq-devnet-0 (completed) | leanSpec framework | no leanSpec alignment |
| pq-devnet-1 (completed) | leanSig sign+verify in CL client | incompatible key format |
| pq-devnet-2 (completed) | leanMultisig aggregation | different aggregation protocol |
| pq-devnet-3 (active) | **Separate aggregator role, aggregation propagation protocol** | MISSING |
| pq-devnet-4 (planned) | **Recursive PQ sig agg via leanVm** | MISSING |

---

## Complete TODO List (Prioritized)

### P0 — Critical blockers for leanEthereum interop

| ID | Item | File to create/modify | Notes |
|---|---|---|---|
| P0-A | 4-second slot configuration | `pkg/consensus/quick_slots.go` | Add `SlotDuration: 4s` config; validate finality timing |
| P0-B | leanSig public key format (50 bytes: 8-element root + 5 randomiser) | `pkg/crypto/pqc/unified_hash_signer.go` | Align with leanSig Rust ref impl key serialization |
| P0-C | NTT precompile address alignment (0x0f–0x12 per ntt-eip spec vs 0x15) | `pkg/core/vm/precompile_ntt.go` | Fork-choice-breaking incompatibility |
| P0-D | Falcon CVE-2025-080201 fix (CRITICAL: salt size not checked) | `pkg/crypto/pqc/falcon_signer.go` | Verify C bindings are not affected |

### P1 — High-priority leanConsensus features

| ID | Item | File to create | Notes |
|---|---|---|---|
| P1-A | Separate PQ aggregator role | `pkg/consensus/pq_aggregator.go` | Distinct network role: collect per-validator XMSS sigs → produce aggregate; decouple from block production (pq-devnet-3) |
| P1-B | leanMultisig aggregation protocol | `pkg/consensus/lean_multisig.go` | XMSS-specific aggregation via zkVM; different from current STARK aggregation |
| P1-C | Gossipsub V2.0 | `pkg/p2p/gossip_v2.go` | Score functions, opportunistic grafting; refs/libp2p specs PR #653 |
| P1-D | Rateless Set Reconciliation | `pkg/p2p/set_reconciliation.go` | arXiv 2402.02668; O(difference) sync for mempool/attestations |
| P1-E | 3SF + ePBS + FOCIL + PeerDAS integrated slot structure | `pkg/consensus/lean_slot.go` | Explicit slot structure combining all 4 sub-protocols per ethresear.ch 2025-08 |

### P2 — Medium-priority gaps

| ID | Item | File to create | Notes |
|---|---|---|---|
| P2-A | Recursive PQ sig aggregation (leanVm) | `pkg/consensus/lean_vm_agg.go` | Coalesce multiple aggregates → single final aggregate (pq-devnet-4) |
| P2-B | WHIR proof system | `pkg/proofs/whir_prover.go` | Better proof size than FRI; refs: https://gfenzi.io/papers/whir/ |
| P2-C | Grid topology P2P | `pkg/p2p/grid_topology.go` | Structured mesh; reduces propagation hops |
| P2-D | QUIC transport | `pkg/p2p/quic_transport.go` | 0-RTT reconnect; required for 4s slots |
| P2-E | Key lifetime parameter finalization | `pkg/crypto/pqc/` | Track leanSpec final lifetime decision; update tree heights |
| P2-F | LMD GHOST with ~256 validators | `pkg/consensus/lmd_ghost.go` | Reduced-committee fork-choice for 4s slots (ethresear.ch 2025-08) |
| P2-G | Exit queue flexibility (Minslack) | `pkg/consensus/exit_queue.go` | Faster exits when queue short; ethresear.ch 2025-04 |
| P2-H | leanSig signing time benchmark | `pkg/crypto/pqc/leansig_bench_test.go` | Verify ≤500 μs signing on representative hardware |

### P3 — Low-priority / long-term

| ID | Item | Notes |
|---|---|---|
| P3-A | Poseidon2 extra-rounds variant | Await cryptanalysis initiative conclusion before committing |
| P3-B | Rainbow Staking model | Full 2-tier staking (heavy 32 ETH / light 1 ETH); awaits APS spec |
| P3-C | Falcon signature aggregation (code-based SNARKs) | Track is inactive; implement when leanroadmap activates it |
| P3-D | Lean 4 formal verification | Research-level; not a Go implementation task |
| P3-E | GKR-style provers / WHIR for binary-field sig agg | Pending in leanroadmap itself |
| P3-F | Aggregate size reduction to 128 KiB | Current leanMultisig is 300–500 KiB; leanroadmap hasn't solved this yet |
| P3-G | leanSpec alignment | Track leanSpec releases and ensure ETH2030 consensus types match |

---

## Benchmark Targets (from leanroadmap)

ETH2030 should implement the following performance targets from leanroadmap:

| Metric | Target | Source |
|---|---|---|
| leanSig signing time | **≤ 500 μs** (single core) | `benchmarks/leansig-timing.ts` |
| leanSig verification time | **≤ 500 μs** (single core) | `benchmarks/leansig-timing.ts` |
| leanMultisig XMSS aggregation | **≥ 1,000 XMSS/sec** | `benchmarks/xmss-aggregation.ts` |
| Aggregate size | **≤ 128 KiB** | `benchmarks/xmss-aggregation.ts` |

Current state as of 2026-03 (from leanroadmap data):
- leanSig signing: 535 μs on M1 (7% over target) — verification: 193 μs ✓
- leanMultisig throughput: 970/sec on M4 Max (97% of target) — nearly there
- Aggregate size: 300–500 KiB (234–391% over target) — **significant gap**

---

## Key Refs in ETH2030 Project

These `refs/` submodules are directly relevant to leanroadmap features:

| Ref | Relevant to |
|---|---|
| `refs/leanSig` | Track 2 — exact leanSig Rust API; use for Go interop alignment |
| `refs/leanMultisig` | Track 2 — leanMultisig XMSS aggregation protocol |
| `refs/leanSpec` | All tracks — official leanConsensus specification |
| `refs/lean-spec-tests` | All tracks — test vectors for leanSpec conformance |
| `refs/ream` | P2P interop — reference Lean Consensus client (Rust) |
| `refs/hash-sig` | Track 1 — original XMSS/Winternitz Rust prototype |
| `refs/circl` | Track 1 — ML-DSA-65 (FIPS 204) Go implementation |

---

*Generated: 2026-03-04. Based on line-by-line analysis of `refs/leanroadmap/data/`.*
