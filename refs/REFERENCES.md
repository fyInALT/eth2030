# ETH2030 Reference Submodules

All submodules in `refs/` are **read-only** reference repositories. They serve as upstream
sources, research material, and implementation references for achieving the ETH2030 goals
(Fast L1, Gigagas L1, Teragas L2, Post-Quantum L1, Private L1).

---

## Categories

| # | Category | Submodules |
|---|----------|-----------|
| A | [Ethereum Specifications](#a-ethereum-specifications) | consensus-specs, execution-specs, execution-apis, beacon-APIs, builder-specs, EIPs, ERCs |
| B | [Test Vectors & Fixtures](#b-test-vectors--fixtures) | consensus-spec-tests, execution-spec-tests, lean-spec-tests |
| C | [Reference Clients](#c-reference-clients) | go-ethereum, lighthouse, prysm, ream |
| D | [Research & Roadmap](#d-research--roadmap) | research, leanroadmap, ream-study-group, pm, iptf-pocs, ethereum-org-website |
| E | [Cryptography Libraries](#e-cryptography-libraries) | blst, circl, gnark, gnark-crypto, go-eth-kzg, c-kzg-4844, go-ipa, go-verkle, hash-sig, ntt-eip, ethfalcon |
| F | [Lean Ethereum (Audited)](#f-lean-ethereum-audited) | leanSpec, leanSig, leanMultisig, fiat-shamir, lean-spec-tests |
| G | [Devops & Testing Infrastructure](#g-devops--testing-infrastructure) | ethereum-package, benchmarkoor, benchmarkoor-tests, spamoor, xatu, execution-processor, consensoor, erigone |
| H | [Utilities & Tooling](#h-utilities--tooling) | eth-utils, web3.py, eip-review-bot |

---

## A. Ethereum Specifications

### consensus-specs
- **URL**: https://github.com/ethereum/consensus-specs
- **Language**: Python (executable specs)
- **Purpose**: Official Ethereum Proof-of-Stake consensus specification. Covers Phase0
  through Electra, including beacon chain state transition, fork choice (LMD-GHOST + Casper
  FFG), p2p networking, light client sync, and honest validator guide.
- **ETH2030 relevance**:
  - Authoritative reference for 3SF (3-slot finality), 4-slot epochs, 1-epoch finality
  - `ePBS` (EIP-7732) spec lives here under `eips/eip-7732`
  - `FOCIL` (EIP-7805) inclusion list spec
  - `EIP-7251` max effective balance (2048 ETH)
  - `EIP-7549` attestation committee index changes
  - Post-quantum attestation design (Dilithium, ML-DSA)
  - Endgame finality BLS adapter and APS committee selection
  - **Use**: Cross-check all CL logic in `pkg/consensus/` against the Python reference

### execution-specs
- **URL**: https://github.com/ethereum/execution-specs
- **Language**: Python
- **Purpose**: Executable Ethereum execution layer spec. Covers all historical hardforks
  from Frontier through Cancun/Prague as Python modules, plus EIP prototypes.
- **ETH2030 relevance**:
  - Reference for all EVM opcode semantics, gas accounting, and state transition
  - EOF (EIP-3540) container format reference
  - EIP-7702 SetCode, EIP-4844 blob tx, EIP-7685 requests
  - Basis for verifying `pkg/core/vm/` opcode implementations
  - **Use**: Diff ETH2030 opcodes against canonical Python before any gas/state changes

### execution-apis
- **URL**: https://github.com/ethereum/execution-apis
- **Language**: OpenAPI / JSON schemas
- **Purpose**: JSON-RPC and Engine API specifications. Defines all `eth_*`, `engine_*`,
  `debug_*`, `net_*` method schemas including Engine API V1–V4.
- **ETH2030 relevance**:
  - Ground truth for `pkg/rpc/` method signatures and types
  - Engine API `forkchoiceUpdated`, `newPayload`, `getPayload` V3/V4/V6 request/response shapes
  - EIP-7898 uncoupled execution payload extension
  - **Use**: Validate all RPC handler input/output types in `pkg/rpc/`

### beacon-APIs
- **URL**: https://github.com/ethereum/beacon-APIs
- **Language**: OpenAPI YAML
- **Purpose**: REST API specification for Ethereum consensus layer (Beacon API). Covers
  16+ endpoints including `/eth/v1/beacon/`, `/eth/v2/validator/`, `/eth/v1/config/`,
  and event streams.
- **ETH2030 relevance**:
  - Reference for the 16 Beacon API endpoints in `pkg/rpc/`
  - SSZ vs JSON response negotiation (`Accept: application/octet-stream`)
  - Lean Consensus API extensions (new endpoints for Lean spec)
  - **Use**: Validate endpoint schemas and response types in `pkg/rpc/`

### builder-specs
- **URL**: https://github.com/ethereum/builder-specs
- **Language**: OpenAPI YAML
- **Purpose**: MEV-Boost / PBS builder API specification. Defines `RegisterValidator`,
  `GetHeader`, `GetPayload`, `SubmitBlindedBlock` interactions between proposer and builder.
- **ETH2030 relevance**:
  - Foundation for `pkg/epbs/` ePBS builder bid and payload envelope
  - Vickrey auction extension (second-price sealed-bid + slashing in `pkg/engine/`)
  - Distributed block builder registration and bid protocol
  - **Use**: Align ePBS builder protocol with upstream PBS spec before diverging

### EIPs
- **URL**: https://github.com/ethereum/EIPs
- **Language**: Markdown
- **Purpose**: Ethereum Improvement Proposals repository. Canonical source for all EIP
  specifications (Core, Networking, Interface, ERC, Meta).
- **ETH2030 relevance**:
  - Primary reference for all 58 complete EIPs implemented in ETH2030
  - Key EIPs: 7702, 7701, 7706, 7742, 7807, 7928, 8141, 8077, 8079, etc.
  - Proposed but not yet final EIPs referenced by roadmap items
  - **Use**: Any time an EIP is implemented, verify against the canonical EIP text here

### ERCs
- **URL**: https://github.com/ethereum/ERCs
- **Language**: Markdown
- **Purpose**: Ethereum Request for Comments — application-layer standards (token standards,
  wallet interfaces, etc.).
- **ETH2030 relevance**:
  - ERC-4337 native AA reference (alongside EIP-7701/7702)
  - Token and wallet interface standards for devnet compatibility testing
  - **Use**: Reference for AA-adjacent wallet integration standards

---

## B. Test Vectors & Fixtures

### consensus-spec-tests
- **URL**: https://github.com/ethereum/consensus-spec-tests
- **Language**: YAML / SSZ binary fixtures
- **Purpose**: Official test vectors generated from the executable consensus-specs. Covers
  state transitions, fork choice, p2p, operations (deposits, withdrawals, slashings,
  attestations) across all hard forks.
- **ETH2030 relevance**:
  - Ground-truth test suite for all CL state transition functions in `pkg/consensus/`
  - Attestation, deposit, withdrawal, slashing operation vectors
  - Fork choice (LMD-GHOST) vectors for `pkg/consensus/` fork choice
  - BLS12-381 signature aggregate/verify vectors
  - **Use**: Add a test runner in `pkg/consensus/` that loads these YAML fixtures

### execution-spec-tests
- **URL**: https://github.com/ethereum/execution-spec-tests
- **Language**: Python (pytest-based fixture generator) + JSON fixtures
- **Purpose**: EL execution test suite. Generates JSON state test fixtures via `fill`
  command, covering EVM edge cases, precompile behaviour, and each EIP.
- **ETH2030 relevance**:
  - Complements the 36,126 EF state tests already passing in `pkg/core/eftest/`
  - New EIP fixture generators for EIP-7702, EIP-7702, EIP-8141, etc.
  - EOF container test vectors
  - **Use**: Generate new fixtures when implementing new EIPs; feed into `pkg/core/eftest/`

### lean-spec-tests
- **URL**: https://github.com/ReamLabs/lean-spec-tests
- **Language**: JSON test vectors
- **Purpose**: Test vectors for the Lean Consensus specification (beam chain). Currently
  contains devnet2 (43 vectors) and devnet3 (58 vectors) covering `fork_choice`,
  `state_transition`, `verify_signatures`, and `ssz` categories.
- **ETH2030 relevance**:
  - **Direct test coverage** for Lean Consensus features in `pkg/consensus/`
  - `verify_signatures` vectors: validate PQ attestation and BLS aggregate verify
  - `state_transition` vectors: validate 3SF finality, quick slots state transitions
  - `ssz` vectors: validate SSZ encoding of Lean Consensus types
  - `fork_choice` vectors: validate Lean fork choice (LMD-GHOST with 3SF)
  - **Use**: Add `pkg/consensus/lean_spec_test.go` that loads devnet2+devnet3 vectors;
    run alongside `pkg/core/eftest/` as the CL counterpart

---

## C. Reference Clients

### go-ethereum
- **URL**: https://github.com/ethereum/go-ethereum
- **Language**: Go
- **Version imported**: v1.17.0 (as library in `pkg/geth/`)
- **Purpose**: Reference EL client. Full Ethereum execution layer implementation including
  EVM, state trie, p2p (RLPx/devp2p), snap sync, Pebble DB, JSON-RPC.
- **ETH2030 relevance**:
  - **Directly embedded** as library: `pkg/geth/` adapter wraps go-ethereum for real
    mainnet/testnet sync (`cmd/eth2030-geth/` binary)
  - 13 custom precompiles injected at Glamsterdam/Hegotá/I+ fork levels
  - `pkg/core/eftest/` uses go-ethereum's state transition as EF test backend
  - Idiomatic Go patterns, error handling, and interface design to emulate
  - **Use**: Check go-ethereum implementations before writing EVM logic from scratch

### lighthouse
- **URL**: https://github.com/sigp/lighthouse
- **Language**: Rust
- **Purpose**: Production Ethereum consensus layer client. Used as the CL in the ETH2030
  Kurtosis devnet (`cl-1-lighthouse-geth` service). Implements all beacon chain specs.
- **ETH2030 relevance**:
  - **Active devnet partner**: ETH2030 EL paired with Lighthouse CL for all devnet tests
  - Engine API compatibility target: all `engine_*` calls come from Lighthouse
  - Reference for Engine API edge cases (forkchoiceUpdated timing, newPayload validation)
  - **Use**: When devnet produces errors, cross-check Engine API request logs from Lighthouse

### prysm
- **URL**: https://github.com/OffchainLabs/prysm (OffchainLabs fork)
- **Language**: Go
- **Purpose**: Go-based Ethereum consensus layer client. Alternative CL for devnet
  multi-client testing.
- **ETH2030 relevance**:
  - Secondary CL client for multi-client devnet configurations
  - Go implementation to compare CL logic patterns with ETH2030's `pkg/consensus/`
  - Beacon API server implementation reference (Go, same language as ETH2030)
  - **Use**: Multi-client devnet second CL; reference Go patterns for CL subsystems

### ream
- **URL**: https://github.com/ReamLabs/ream
- **Language**: Rust
- **Purpose**: First implementation of the Lean Consensus specification (beam chain /
  beacon chain 2.0). Modular, contributor-friendly Rust client targeting fast finality,
  snarkification, and post-quantum security.
- **Crate structure**:
  - `crates/common/consensus/{beacon,lean}` — Beacon and Lean consensus state machines
  - `crates/common/fork_choice/{beacon,lean}` — Fork choice for both specs
  - `crates/common/execution/engine` — Engine API client
  - `crates/crypto/{bls,post_quantum,merkle,keystore}` — Crypto primitives
  - `crates/networking/` — libp2p gossipsub, req/resp
  - `crates/common/polynomial_commitments` — KZG commitments
  - `crates/common/light_client` — Light client sync
- **ETH2030 relevance**:
  - **Primary specification reference** for all Lean Consensus features in `pkg/consensus/`
  - Lean fork choice logic: directly maps to 3SF, 4-slot epochs, 1-epoch finality
  - PQ crypto crate: Dilithium/ML-DSA patterns for `pkg/crypto/pqc/`
  - Lean API types (`crates/common/api_types/lean`) define new Beacon API extensions
  - Polynomial commitments crate: KZG patterns for PeerDAS cells
  - **Use**: When implementing Lean Consensus features, diff against Ream's Rust
    implementation for algorithmic correctness

---

## D. Research & Roadmap

### research
- **URL**: https://github.com/ethereum/research
- **Language**: Python (primarily), some Rust/C
- **Author**: Vitalik Buterin et al.
- **Purpose**: Vitalik's research prototypes and proofs-of-concept. Contains early
  implementations of almost every major Ethereum research direction.
- **Key directories**:

  | Directory | Relevance |
  |-----------|-----------|
  | `3sf-mini/` | 3-slot finality (3SF) mini-spec prototype |
  | `erasure_code/` | 2D Reed-Solomon, EC256/EC65536 — PeerDAS blob reconstruction |
  | `kzg_data_availability/` | KZG proofs, FK20 multi/single — EIP-4844 / EIP-7594 |
  | `verkle/`, `verkle_trie/` | Verkle trie proofs — binary tree migration reference |
  | `zksnark/` | R1CS, QAP — zkVM/proof system foundations |
  | `zkstark/` | STARK protocols — proof aggregation research |
  | `beacon_chain_impl/` | Early PoS beacon chain Python impl |
  | `ssz_research/`, `newssz/` | SSZ encoding experiments |
  | `polynomial_reconstruction/` | Lagrange interpolation for DAS |
  | `binary_fft/` | Binary field FFT — Binius-related research |
  | `binius/` | Binius proof system research |
  | `circlestark/` | Circle STARK research |
  | `bulletproofs/` | Bulletproof range proofs |
  | `py_plonk/` | PLONK proof system |
  | `whisk_csidh/` | WHISK secret leader election / CSIDH isogeny |
  | `rollup_compression/` | Rollup data compression research |
  | `fast_cross_shard_execution/` | Cross-shard execution model |
  | `papers/` | LaTeX sources for research papers |

- **ETH2030 relevance**:
  - `3sf-mini/`: Algorithmic reference for `pkg/consensus/` 3SF implementation
  - `erasure_code/`: Reed-Solomon reconstruction for `pkg/das/erasure/`
  - `kzg_data_availability/`: FK20 proof generation for `pkg/das/`
  - `zksnark/` + `zkstark/`: Proof system foundations for `pkg/zkvm/` and `pkg/proofs/`
  - **Use**: When a crypto primitive or protocol needs a proof-of-concept, check here first

### leanroadmap
- **URL**: https://github.com/ReamLabs/leanroadmap
- **Language**: TypeScript / Next.js (website)
- **Live site**: https://leanroadmap.org
- **Purpose**: Lean Consensus research progress tracker website. Tracks all workstreams,
  milestones, client team progress, devnets, benchmarks, and learning resources.
- **Key data files** (`data/`):

  | File | Contents |
  |------|---------|
  | `research-tracks.tsx` | Research tracks with milestones: Poseidon cryptanalysis, PQ signatures, zkVM, light client, fork choice, networking, aggregation |
  | `timeline.tsx` | Phase timeline: Pilling (2024-2025), Speccing (2025-2026), Building (2026-2027), Testing (2027-2029) |
  | `client-teams.tsx` | CL client teams and their Lean Consensus implementations |
  | `devnets.ts` | Devnet schedules and configurations |
  | `benchmarks/` | Performance benchmark data |
  | `lean-calls.ts` | Lean Consensus call meeting notes |
  | `learning-resources.ts` | Curated learning materials |
  | `call-to-actions.ts` | Contribution opportunities |

- **ETH2030 relevance**:
  - **Research track status**: Know which Lean Consensus research areas are active/blocked
  - **Milestone tracking**: Align ETH2030 implementation priorities with Lean spec milestones
  - **Devnet schedule**: ETH2030 devnet configurations should target leanroadmap devnet specs
  - **Use**: Check `data/research-tracks.tsx` to understand current research focus areas
    and priority order for implementing Lean Consensus features

### ream-study-group
- **URL**: https://github.com/ReamLabs/ream-study-group
- **Language**: Markdown (study materials + meeting notes)
- **Purpose**: Curated learning materials and weekly meeting notes covering the Beacon chain,
  Beam chain (Lean Consensus), cryptography, and zkVM topics. 60+ weekly session notes
  (Nov 2024 – Feb 2026).
- **Key directories**:

  | Directory | Contents |
  |-----------|---------|
  | `beacon-chain.md` | Deep-dive on Beacon chain Phase0 through Deneb: data types, state transition, fork choice, p2p, validator guide |
  | `cryptography/` | Study notes on BLS, hash functions, PQ signatures, merkle trees |
  | `zkvm/` | Study notes on zkVM design, RISC-V, proof systems |
  | `meeting-notes/` | 60 weekly sessions covering Lean spec discussions, Ream development, PQ crypto, aggregation, fork choice |
  | `figures/` | Diagrams for beacon chain and Lean Consensus |

- **ETH2030 relevance**:
  - `beacon-chain.md`: Comprehensive reference for all Phase0–Deneb CL concepts
  - Meeting notes track evolving Lean Consensus design decisions (what changed and why)
  - `cryptography/` covers BLS aggregation, Poseidon, ML-DSA — directly maps to
    `pkg/crypto/pqc/` and `pkg/consensus/` design
  - `zkvm/` study notes inform `pkg/zkvm/` RISC-V and zkISA bridge design
  - **Use**: Before implementing a Lean Consensus feature, read the relevant meeting notes
    for historical design rationale; check `beacon-chain.md` for canonical CL definitions

### pm
- **URL**: https://github.com/ethereum/pm
- **Language**: Markdown
- **Purpose**: Ethereum Protocol Meetings repository. Records of All Core Devs (ACD)
  meetings, EIP discussion notes, and protocol governance decisions.
- **ETH2030 relevance**:
  - Track which EIPs are accepted/rejected for upcoming forks
  - Decision rationale for Glamsterdam, Hegotá, and I+ upgrade EIP selection
  - **Use**: Cross-reference EIP implementation priority against ACD decisions

### iptf-pocs
- **URL**: https://github.com/ethereum/iptf-pocs
- **Language**: Various
- **Purpose**: IPTF (Interoperability, Performance, Testing, and Features) proof-of-concept
  implementations from the EF research team.
- **ETH2030 relevance**:
  - Prototype implementations for novel EIPs before they land in production clients
  - **Use**: Check for early PoC implementations of new roadmap EIPs

### ethereum-org-website
- **URL**: https://github.com/ethereum/ethereum-org-website
- **Language**: TypeScript / Next.js
- **Purpose**: Official ethereum.org website source. Contains developer documentation,
  upgrade history, roadmap explanations, and educational content.
- **ETH2030 relevance**:
  - Roadmap narrative and upgrade history context
  - Developer documentation for EIPs ETH2030 implements
  - **Use**: Reference for public-facing upgrade descriptions and EIP motivations

---

## E. Cryptography Libraries

### blst
- **URL**: https://github.com/supranational/blst
- **Language**: C / assembly (Go bindings)
- **License**: Apache-2.0
- **CGO**: Yes
- **Purpose**: Production BLS12-381 signature library. Industry-standard, audited, used
  by Lighthouse, Prysm, and go-ethereum.
- **ETH2030 relevance**:
  - Upgrade target for `pkg/consensus/` `PureGoBLSBackend` → `blst` for production performance
  - Aggregate verify for 1M attestations/slot (K+ milestone)
  - jeanVM aggregation (Groth16 ZK-circuit BLS) performance baseline
  - **Use**: Replace PureGoBLSBackend with blst bindings when production performance needed

### circl
- **URL**: https://github.com/cloudflare/circl
- **Language**: Go
- **License**: BSD-3
- **CGO**: No
- **Purpose**: Cloudflare's cryptographic library. Includes ML-DSA (FIPS 204), ML-KEM,
  SLH-DSA (SPHINCS+), X25519, and many other algorithms.
- **ETH2030 relevance**:
  - **Already used**: `pkg/crypto/pqc/` uses circl for ML-DSA-65 (real FIPS 204 lattice ops)
  - Dilithium3 and SPHINCS+ implementations
  - ML-KEM for encrypted mempool threshold crypto
  - **Use**: Primary PQC library; expand usage for PQ blob commitments and PQ attestations

### gnark
- **URL**: https://github.com/Consensys/gnark
- **Language**: Go
- **License**: Apache-2.0
- **CGO**: No
- **Purpose**: ZK-SNARK circuit framework. Supports Groth16 and PLONK over BN254, BLS12-381,
  BLS12-377 curves. High-level Go DSL for circuit writing.
- **ETH2030 relevance**:
  - `pkg/proofs/` proof aggregation framework upgrade target
  - jeanVM aggregation (Groth16 ZK-circuit BLS in `pkg/consensus/`)
  - AA proof circuits (nonce/sig/gas constraints) in `pkg/proofs/`
  - STF in zkISA proof generation in `pkg/zkvm/`
  - **Use**: Implement real Groth16 proof generation replacing placeholder in `pkg/proofs/`

### gnark-crypto
- **URL**: https://github.com/Consensys/gnark-crypto
- **Language**: Go
- **License**: Apache-2.0
- **CGO**: No
- **Purpose**: Low-level elliptic curve arithmetic. BN254, BLS12-381, BLS12-377,
  BLS24-315 field ops, pairings, hash-to-curve, KZG.
- **ETH2030 relevance**:
  - BN254 field arithmetic for `pkg/crypto/` Pedersen commitments (private L1)
  - BLS12-381 pairing for endgame finality BLS adapter
  - Hash-to-curve for BLS signatures
  - **Use**: Low-level curve ops when `gnark` circuit DSL is too high-level

### go-eth-kzg
- **URL**: https://github.com/crate-crypto/go-eth-kzg
- **Language**: Go
- **License**: Apache-2.0
- **CGO**: No
- **Purpose**: Pure-Go KZG commitment library with the Ethereum trusted setup. Implements
  EIP-4844 and EIP-7594 PeerDAS KZG operations.
- **ETH2030 relevance**:
  - Upgrade target for `pkg/das/` `PlaceholderKZGBackend` → real trusted setup
  - Cell-level KZG proofs for PeerDAS (EIP-7594) in `pkg/das/`
  - Blob polynomial evaluation for `pkg/core/` blob gas pricing
  - **Use**: Replace placeholder KZG with `go-eth-kzg` for production PeerDAS

### c-kzg-4844
- **URL**: https://github.com/ethereum/c-kzg-4844
- **Language**: C (Go bindings)
- **License**: Apache-2.0
- **CGO**: Yes
- **Purpose**: C implementation of KZG for EIP-4844. Audited, faster than pure-Go.
  Used by go-ethereum in production.
- **ETH2030 relevance**:
  - Alternative to `go-eth-kzg` for higher performance KZG in `pkg/das/`
  - EIP-4844 blob verification for the `pkg/geth/` adapter
  - **Use**: Production blob verification when CGO is acceptable

### go-ipa
- **URL**: https://github.com/crate-crypto/go-ipa
- **Language**: Go
- **License**: Apache-2.0
- **Purpose**: Inner Product Argument (IPA) proofs over the Banderwagon curve (Verkle
  trie proofs).
- **ETH2030 relevance**:
  - `pkg/crypto/` IPA proof implementation reference
  - Verkle trie proof verification for state migration (`pkg/trie/`)
  - **Use**: Reference for Banderwagon IPA operations in `pkg/crypto/`

### go-verkle
- **URL**: https://github.com/ethereum/go-verkle
- **Language**: Go
- **Purpose**: Verkle trie implementation for Ethereum state. Used alongside `go-ipa`.
- **ETH2030 relevance**:
  - Verkle → binary trie migration reference (I+ milestone)
  - Partial statelessness witness format for `pkg/core/vops/`
  - **Use**: Reference for verkle proof structures in witness generation

### hash-sig
- **URL**: https://github.com/b-wagn/hash-sig
- **Language**: Rust
- **Purpose**: Hash-based multi-signature schemes research implementation. Covers
  XMSS, WOTS+, and novel hash-based multi-sig constructions.
- **ETH2030 relevance**:
  - Reference for `pkg/crypto/pqc/` unified hash signer (XMSS/WOTS+) in M+ milestone
  - PQ L1 hash-based signatures (M+ roadmap item)
  - **Use**: Algorithm reference when implementing XMSS/WOTS+ in `pkg/crypto/pqc/`

### ntt-eip
- **URL**: https://github.com/ZKNoxHQ/NTT
- **Language**: Go / Solidity
- **Purpose**: Number Theoretic Transform (NTT) precompile reference implementation for
  EIP-7885. Fast polynomial multiplication in finite fields.
- **ETH2030 relevance**:
  - **Directly referenced** for the NTT precompile in `pkg/core/vm/`
  - Post-quantum signature verification acceleration (Falcon, Dilithium use NTT)
  - Polynomial commitment fast multiplication
  - **Use**: NTT precompile correctness verification in `pkg/core/vm/`

### ethfalcon
- **URL**: https://github.com/ZKNoxHQ/ETHFALCON
- **Language**: Go / Solidity
- **Purpose**: Falcon512 signature scheme on EVM using NTT precompile. Reference for
  PQ signature verification as an EVM precompile.
- **ETH2030 relevance**:
  - Reference for Falcon512 implementation in `pkg/crypto/pqc/`
  - EVM precompile design for PQ signature verification
  - Integration with NTT precompile for fast lattice operations
  - **Use**: Falcon512 correctness reference and NTT-based EVM precompile design

---

## F. Lean Ethereum (Audited)

These four repositories are from the [leanEthereum](https://github.com/leanEthereum) org
and were subject to an 18-finding security audit (9 PRs submitted to upstream).

### leanSpec
- **URL**: https://github.com/leanEthereum/leanSpec
- **Language**: Python
- **Purpose**: Executable Lean Consensus specification. The canonical Lean Consensus state
  machine in Python, analogous to how `consensus-specs` is the reference for the current
  beacon chain. Covers fork choice, state transition, SSZ types, and signature verification.
- **ETH2030 relevance**:
  - **Primary spec source** for all Lean Consensus features in `pkg/consensus/`
  - 3SF state transition, Lean fork choice rules
  - Lean SSZ types and serialization format
  - `lean-spec-tests` vectors are generated from this spec
  - **Use**: When implementing any Lean Consensus feature, align with this Python spec first

### leanSig
- **URL**: https://github.com/leanEthereum/leanSig
- **Language**: Python / Solidity
- **Purpose**: Lean signature scheme reference. Covers BLS aggregate signatures, threshold
  signatures, and PQ signature integration for the Lean Consensus.
- **ETH2030 relevance**:
  - PQ attestation signature scheme reference for `pkg/consensus/` and `pkg/crypto/pqc/`
  - Threshold BLS for encrypted mempool decryption in `pkg/txpool/encrypted/`
  - **Use**: Signature scheme correctness reference; cross-check with circl PQ implementations

### leanMultisig
- **URL**: https://github.com/leanEthereum/leanMultisig
- **Language**: Python / Solidity
- **Purpose**: Lean multi-signature scheme. Reference for aggregated validator signatures
  in the Lean Consensus, covering key aggregation and multi-party signing.
- **ETH2030 relevance**:
  - jeanVM aggregation (Groth16 ZK-circuit BLS) reference in `pkg/consensus/`
  - 1M attestations/slot aggregate verification (K+ milestone)
  - **Use**: Multi-sig aggregation correctness; basis for 1M attestation scalability

### fiat-shamir
- **URL**: https://github.com/leanEthereum/fiat-shamir
- **Language**: Python
- **Purpose**: Fiat-Shamir transform reference implementation for non-interactive proofs.
  Used in the leanEthereum audit as a test for ZK proof soundness.
- **ETH2030 relevance**:
  - ZK proof non-interactivity in `pkg/proofs/` and `pkg/zkvm/`
  - Reference for correct Fiat-Shamir heuristic application in Groth16/PLONK circuits
  - **Use**: Verify Fiat-Shamir transform correctness in ZK proof generation code

---

## G. Devops & Testing Infrastructure

### ethereum-package
- **URL**: https://github.com/ethpandaops/ethereum-package
- **Language**: Starlark (Kurtosis)
- **Purpose**: Kurtosis package for spinning up full Ethereum devnets with multiple EL/CL
  client pairs, assertoor, spamoor, monitoring, and tooling.
- **ETH2030 relevance**:
  - **Directly used**: ETH2030 Kurtosis devnet configs in `pkg/devnet/kurtosis/` extend
    this package
  - Configuration reference for multi-client devnet topology
  - Assertoor integration for automated consensus checking
  - **Use**: Upstream changes to this package affect devnet configs; sync periodically

### spamoor
- **URL**: https://github.com/ethpandaops/spamoor
- **Language**: Go
- **Purpose**: Transaction spammer for Ethereum devnets. Supports multiple tx types
  (EOA transfers, ERC20, blob txs) with configurable rate and concurrency.
- **ETH2030 relevance**:
  - Devnet stress testing: send high-volume blob txs and EIP-7702 SetCode txs
  - Gigagas throughput testing in `pkg/devnet/kurtosis/`
  - **Use**: Include in devnet configs for throughput benchmarks

### benchmarkoor
- **URL**: https://github.com/ethpandaops/benchmarkoor
- **Language**: Go
- **Purpose**: Ethereum client benchmark framework. Measures EVM execution throughput,
  state access performance, and block processing latency.
- **ETH2030 relevance**:
  - Benchmark `pkg/core/vm/` EVM throughput against go-ethereum baseline
  - Track Gigagas L1 progress (1 Ggas/sec target)
  - **Use**: Run benchmarks after significant EVM or state trie changes

### benchmarkoor-tests
- **URL**: https://github.com/ethpandaops/benchmarkoor-tests
- **Language**: Go
- **Purpose**: Test suite for benchmarkoor. Contains benchmark scenarios and regression
  test vectors.
- **ETH2030 relevance**:
  - Benchmark scenario reference for ETH2030-specific EVM opcodes
  - Performance regression detection for repricing changes
  - **Use**: Add ETH2030-specific benchmark scenarios when testing new EIPs

### xatu
- **URL**: https://github.com/ethpandaops/xatu
- **Language**: Go
- **Purpose**: Ethereum network data collection and analysis platform. Aggregates beacon
  chain events, slot timing, blob propagation, and peer data.
- **ETH2030 relevance**:
  - Network-level analysis of blob propagation (EIP-4844/7594 PeerDAS)
  - Slot timing distribution reference for quick slots design
  - **Use**: Reference for p2p metrics collection in `pkg/metrics/`

### execution-processor
- **URL**: https://github.com/ethpandaops/execution-processor
- **Language**: Go
- **Purpose**: Post-processes execution layer data (blocks, traces, receipts) for analysis
  and indexing.
- **ETH2030 relevance**:
  - Block and receipt indexing patterns for `pkg/core/rawdb/`
  - Trace format reference for debug API implementation
  - **Use**: Reference for block processing pipeline and trace output format

### consensoor
- **URL**: https://github.com/ethpandaops/consensoor
- **Language**: Go
- **Purpose**: Consensus layer monitoring and analysis tool. Tracks validator performance,
  missed slots, and finality events.
- **ETH2030 relevance**:
  - Devnet consensus health monitoring in `pkg/devnet/kurtosis/scripts/`
  - Finality detection for 3SF and 1-epoch finality validation
  - **Use**: Integrate into devnet check scripts for automated consensus verification

### erigone
- **URL**: https://github.com/ethpandaops/erigone
- **Language**: Go
- **Purpose**: Ethereum client based on Erigon architecture, managed by ethpandaops for
  devnet testing.
- **ETH2030 relevance**:
  - Alternative EL client for multi-client devnet configurations
  - Architecture reference for staged sync pipeline
  - **Use**: Secondary EL in multi-client devnets to test Engine API compatibility

---

## H. Utilities & Tooling

### eth-utils
- **URL**: https://github.com/ethereum/eth-utils
- **Language**: Python
- **Purpose**: Python utility library for Ethereum. Covers address encoding, hex
  conversion, ABI encoding, and type coercions.
- **ETH2030 relevance**:
  - Utility reference for Python test scripts in `pkg/devnet/kurtosis/scripts/`
  - ABI encoding reference for precompile call data construction
  - **Use**: Python devnet scripts use eth-utils for address/hex handling

### web3.py
- **URL**: https://github.com/ethereum/web3.py
- **Language**: Python
- **Purpose**: Python Ethereum client library. Full JSON-RPC, contract interaction,
  event filtering, and signing support.
- **ETH2030 relevance**:
  - Devnet smoke-test scripts (`scripts/test-precompiles.sh`) use web3.py for RPC calls
  - Reference for JSON-RPC method parameter encoding
  - **Use**: Precompile and EIP smoke-test scripts in devnet

### eip-review-bot
- **URL**: https://github.com/ethereum/eip-review-bot
- **Language**: TypeScript
- **Purpose**: GitHub bot that automates EIP PR review: validates formatting, assigns
  reviewers, checks EIP number uniqueness.
- **ETH2030 relevance**:
  - EIP lifecycle tracking (which EIPs are Draft → Review → Final)
  - Automation reference for internal PR governance
  - **Use**: Monitor EIP status changes for roadmap EIPs

---

## ETH2030 Implementation Priority Matrix

The table below maps roadmap goals to the most directly relevant reference submodules:

| ETH2030 Goal | Primary Refs | Secondary Refs |
|---|---|---|
| **Fast L1** (3SF, finality in seconds) | `leanSpec`, `ream`, `consensus-specs` | `lean-spec-tests`, `research/3sf-mini`, `ream-study-group` |
| **Gigagas L1** (1 Ggas/sec) | `go-ethereum`, `execution-specs` | `benchmarkoor`, `benchmarkoor-tests`, `execution-spec-tests` |
| **Teragas L2** (PeerDAS, blob streaming) | `go-eth-kzg`, `c-kzg-4844` | `research/kzg_data_availability`, `research/erasure_code` |
| **Post-Quantum L1** | `circl`, `hash-sig`, `ntt-eip`, `ethfalcon` | `leanSpec`, `ream`, `leanSig`, `research` |
| **Private L1** (shielded transfers) | `gnark`, `gnark-crypto` | `research/bulletproofs`, `research/zksnark` |
| **ePBS / MEV** | `builder-specs`, `consensus-specs` | `leanroadmap`, `ream-study-group` |
| **ZK / Proof Aggregation** | `gnark`, `gnark-crypto` | `research/zksnark`, `research/py_plonk`, `fiat-shamir` |
| **Native Rollups** | `execution-specs`, `EIPs` | `ream-study-group` (beam chain proposals) |
| **1M Attestations** | `blst`, `leanSpec`, `ream` | `leanMultisig`, `leanSig` |
| **Devnet / CI** | `ethereum-package`, `spamoor` | `benchmarkoor`, `consensoor`, `xatu` |
| **EIP Conformance** | `EIPs`, `execution-spec-tests` | `consensus-spec-tests`, `lean-spec-tests` |

---

*Last updated: 2026-03-04. Total submodules: 47.*
