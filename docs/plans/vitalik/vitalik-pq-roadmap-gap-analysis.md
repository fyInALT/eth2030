# Post-Quantum Resistance Roadmap: Gap Analysis

**Source**: Vitalik Buterin's message on the quantum resistance roadmap
**Date**: 2026-03-04
**Scope**: Four quantum-vulnerable areas in Ethereum, mapped line-by-line against ETH2030 implementation.

---

## Overview

Vitalik identifies four areas that are quantum-vulnerable today:

| # | Area | Vitalik's fix | ETH2030 status |
|---|------|--------------|----------------|
| 1 | Consensus-layer BLS signatures | Hash-based sigs (Winternitz) + STARK aggregation | **Mostly done** — gaps in hash selection & lean-available-chain mode |
| 2 | Data availability (KZG commitments + proofs) | Move to STARKs for erasure coding | **Partial** — PQ blob layer exists; core KZG pipeline not replaced |
| 3 | EOA signatures (ECDSA) | Native AA (EIP-8141) + PQ sig algorithms + vectorized precompiles | **Mostly done** — Blake3 backend is placeholder; gas benchmarks missing |
| 4 | Application-layer ZK proofs (KZG/Groth16) | Protocol-layer recursive proof aggregation via EIP-8141 validation frames | **Mostly done** — end-to-end frame→STARK replacement not wired |

---

## 1. Consensus-Layer BLS Signatures

### What Vitalik says

> Lean consensus includes fully replacing BLS signatures with hash-based signatures (some variant of Winternitz), and using STARKs to do aggregation.
>
> Before lean finality, we stand a good chance of getting the **Lean available chain**. This also involves hash-based signatures, but there are much fewer signatures (256-1024 per slot), so we do **not** need STARKs for aggregation.
>
> One important thing upstream of this is **choosing the hash function**. This may be "Ethereum's last hash function"... Likely options are: Poseidon2 + extra rounds (potentially Monolith non-arithmetic layers), Poseidon1, BLAKE3.

### What is DONE

#### Hash-based signature schemes

| File | What it provides |
|------|----------------|
| `pkg/crypto/pqc/unified_hash_signer.go` | XMSS + WOTS+ unified signer. SHA-256 based. Tree heights H=10/16/20 (1K/64K/1M sigs). Key exhaustion tracking, XMSSKeyManager multi-tree. |
| `pkg/crypto/pqc/l1_hash_sig.go` | Standalone L1 Winternitz OTS tree signer, Keccak256-backed. Configurable height. |
| `pkg/crypto/pqc/l1_hash_sig_v2.go` | V2 with extended leaf format and multi-signer support. |
| `pkg/crypto/pqc/hash_sig.go` | Hash-based signature shares for threshold aggregation. |
| `pkg/crypto/pqc/hash_signature.go` | Core hash signature type definitions. |

#### Pluggable hash backend (for "last hash function" flexibility)

| File | What it provides |
|------|----------------|
| `pkg/crypto/pqc/hash_backend.go` | `HashBackend` interface. Concrete backends: `Keccak256Backend`, `SHA256Backend`, `Blake3Backend` (placeholder — see gaps). |
| `pkg/zkvm/poseidon.go` | Poseidon1 over BN254 scalar field (t=3, full=8, partial=57). |
| `pkg/zkvm/poseidon2.go` | Poseidon2 over BN254 (diagonal MDS, external/internal split rounds). |

#### PQ attestations (using lattice-based PQ signers)

| File | What it provides |
|------|----------------|
| `pkg/consensus/pq_attestation.go` | `PQAttestation` type. Dilithium3/ML-DSA sign+verify with classic BLS fallback. Configurable `MinPQValidators` for transition. |
| `pkg/consensus/pq_chain_security.go` | PQ-aware fork choice: SHA-3 based chain security, quantum-resistant chain selection. |
| `pkg/crypto/pqc/mldsa_signer.go` | ML-DSA-65 (FIPS 204) real lattice signer via `circl`. |
| `pkg/crypto/pqc/dilithium.go`, `dilithium_sign.go` | Dilithium3 real lattice operations. |
| `pkg/crypto/pqc/pq_algorithm_registry.go` | Algorithm registry: Dilithium3, Falcon512, SPHINCS+, ML-DSA-65, XMSS registered. |
| `pkg/crypto/pqc/pubkey_registry.go` | On-chain PQ pubkey registry for validators. |

#### STARK aggregation (replaces BLS aggregate-verify for large validator sets)

| File | What it provides |
|------|----------------|
| `pkg/consensus/stark_sig_aggregation.go` | `STARKSignatureAggregator`. Takes N `PQAttestation`s, produces a single `STARKSignatureAggregation` with `AggregateProof`. CommitteeRoot binds the validator set. Replaces O(N) verify with 1 STARK verify. |
| `pkg/consensus/parallel_bls.go` | Parallel BLS aggregate verify (current-era, pre-PQ). |
| `pkg/consensus/jeanvm_aggregation.go` | jeanVM Groth16 ZK-circuit BLS aggregation (K+ milestone). |
| `pkg/consensus/batch_verifier.go` | Batch signature verifier supporting both classic and PQ paths. |

### GAPS / TODO

#### GAP 1.1 — Hash function not finalized

Vitalik says this is "Ethereum's last hash function" — must choose wisely before committing the hash-based signature scheme.

- `Blake3Backend` in `pkg/crypto/pqc/hash_backend.go` is explicitly marked a **structural placeholder** ("production use should integrate `lukechampine.com/blake3` or `zeebo/blake3`"). It uses two rounds of SHA-256 with domain labels, not real BLAKE3.
- **Monolith** (non-arithmetic layer mixed into Poseidon2) is not implemented anywhere.
- **Poseidon2 + extra rounds** variant is not present — `poseidon2.go` uses default t=3, external=8, internal=56. No "extra rounds for conservatism" variant.
- No comparative benchmarks between Poseidon1 / Poseidon2 / BLAKE3 / Monolith in a circuit context.

**Action needed**:
1. Integrate real BLAKE3 (`lukechampine.com/blake3`) into `hash_backend.go`.
2. Add `Poseidon2ExtraRoundsParams` (e.g., external=12, internal=84) to `pkg/zkvm/poseidon2.go`.
3. Add `MonolithBackend` stub: Poseidon2 permutation with one non-arithmetic S-box layer (lookup table).
4. Add `pkg/crypto/pqc/hash_bench_test.go` benchmarking all four candidates at the same security level.
5. Document the chosen hash and rationale in `docs/` once EF finalizes.

#### GAP 1.2 — Lean available chain is not a distinct mode

Vitalik describes a **pre-lean-finality** phase where only 256–1024 signatures per slot need hash-based sigs, without STARK aggregation. ETH2030 only has the full "STARKs for aggregation" path.

**Action needed**:
- Add `LeanAvailableChainMode` flag to `pkg/consensus/config.go`.
- When enabled: use hash-based sigs for attestation, skip STARK aggregation (direct aggregate via Merkle tree of WOTS+ keys instead).
- Wire the mode into `pq_attestation.go` and `stark_sig_aggregation.go`.

#### GAP 1.3 — STARK aggregation uses placeholder prover

`stark_sig_aggregation.go` delegates to `proofs.NewSTARKProver()`. That prover (`pkg/proofs/stark_prover.go`) is a working FRI-based implementation over Goldilocks field, but is not production-grade (no recursive STARK, no circuit specialization for Winternitz verification).

**Action needed** (longer-term, K+ milestone):
Integrate a real STARK prover for Winternitz signature verification circuits (e.g., using `gnark` STARK extension or `refs/research/circlestark` as reference).

---

## 2. Data Availability (KZG → PQ-safe)

### What Vitalik says

> Today, we rely pretty heavily on KZG for erasure coding. We could move to STARKs, but:
> 1. **2D DAS + linearity**: "our current thinking is that it should be sufficient to just max out 1D DAS (PeerDAS)."
> 2. **Erasure correctness proofs**: KZG does this for free. STARKs can substitute but a STARK is bigger than a blob — needs **recursive STARKs** (or alternatives).
>
> Summary: manageable, but a lot of engineering work.

### What is DONE

#### 1D PeerDAS (EIP-7594) — complete

| File | What it provides |
|------|----------------|
| `pkg/das/sampling.go`, `sampling_scheduler.go` | DAS sampling rounds per EIP-7594. |
| `pkg/das/column_custody.go`, `custody_manager.go` | Column custody, subnet management. |
| `pkg/das/reed_solomon_encode.go` | Reed-Solomon blob encoding. |
| `pkg/das/erasure/reed_solomon.go`, `reed_solomon_encoder.go` | Full RS encoder/decoder over GF(2^8) and GF extension fields. |
| `pkg/das/reconstruction.go`, `reconstruction_pipeline.go` | Blob reconstruction from partial columns. |
| `pkg/das/cell_gossip.go`, `cell_messages.go` | Cell-level gossip (EIP-7594). |
| `pkg/das/blob_validator.go`, `blob_reconstruct.go` | Blob validation and reconstruction. |
| `pkg/das/varblob.go`, `variable_blobs.go` | Variable-size blobs. |

2D DAS is explicitly **not pursued** — consistent with Vitalik's 1D-sufficient stance.

#### PQ blob layer (L+ milestone)

| File | What it provides |
|------|----------------|
| `pkg/das/pq_blobs.go` | `PQBlobCommitment` using lattice-based commitment (dim=256, modulus=12289 — Kyber-like). Commit, open, verify, batch verify. |
| `pkg/das/pq_blob_signer.go` | Signs blob commitments with PQ key. |
| `pkg/das/pq_blob_validator.go` | Validates PQ-signed blob commitments. |
| `pkg/das/pq_blob_integrity.go` | Integrity checks on PQ blob data. |
| `pkg/das/lattice_blob_commit.go` (via `pq_blobs.go`) | Lattice polynomial commitment over Rq = Zq[X]/(X^n+1). |

#### STARK DA type scaffold

| File | What it provides |
|------|----------------|
| `pkg/das/types.go:114` | `STARKCommitment` struct: `TraceCommitment []byte`, `ProofSize int`, `UseSTARKDA bool`. |
| `pkg/proofs/stark_prover.go` | Full STARK prover with FRI over Goldilocks field — can in principle prove erasure coding correctness. |

### GAPS / TODO

#### GAP 2.1 — Core blob commitment pipeline still uses KZG

The production blob pipeline (`pkg/das/`, `pkg/core/`, `pkg/geth/`) commits blobs via KZG (EIP-4844). The `PQBlobCommitment` in `pq_blobs.go` is an **additional** layer (blob signed with PQ key), not a replacement for KZG polynomial commitments.

Replacing KZG with STARKs for the erasure coding commitment is the hard engineering Vitalik describes — not yet done.

**Action needed**:
1. Design `STARKBlobCommitter` in `pkg/das/`: takes blob data, produces a STARK proof that the RS-encoded columns are correct, replacing KZG opening proofs.
2. Wire `STARKCommitment` (currently a stub struct in `types.go`) into `blob_validator.go` as an alternative to KZG verification.
3. Handle "STARK is bigger than a blob" — implement recursive STARK composition in `pkg/proofs/recursive_prover.go` so the erasure-correctness STARK folds into a smaller proof.
4. Distributed blob selection with STARK proofs — design doc needed before implementation.

#### GAP 2.2 — No erasure-correctness STARK proof generation

KZG provides erasure correctness "for free" through its polynomial binding property. A STARK substitute must explicitly prove: "these K columns are the correct RS encoding of this blob." There is no such circuit or prover today.

**Action needed**: Add `pkg/das/erasure_stark_prover.go` — a STARK circuit that takes the blob polynomial coefficients and proves the RS evaluation is correct.

#### GAP 2.3 — `STARKCommitment` is a struct stub only

`das/types.go:114–120` defines the type but there is no constructor, no proof generation, no verification logic, and no integration with the sampling pipeline.

**Action needed**: Promote `STARKCommitment` from struct to working commitment type with `Commit()`, `Open()`, `Verify()` methods backed by `pkg/proofs/stark_prover.go`.

---

## 3. EOA Signatures (ECDSA → Native AA + PQ Algorithms)

### What Vitalik says

> The answer is clear: add **native AA** (EIP-8141) so we get first-class accounts that can use any signature algorithm.
>
> We need quantum-resistant signature algorithms to actually be viable:
> - ECDSA: 3000 gas
> - Hash-based PQ sigs: **~200k gas** range
> - Lattice-based sigs: extremely inefficient today, but **vectorized math precompiles** (NTT / butterfly permutations, +, *, %, dot product) could reduce to a similar range
>
> Long-term fix: **protocol-layer recursive signature and proof aggregation** → gas overhead near-zero.

### What is DONE

#### Native AA via EIP-8141

| File | What it provides |
|------|----------------|
| `pkg/core/vm/eip8141_opcodes.go` | `APPROVE` (0xaa) and `TXPARAM*` opcodes. `FrameContext` with sender/payer approval, 2D nonce, frame list. |
| `pkg/core/vm/aa_executor.go` | AA frame executor. |
| `pkg/core/vm/call_frame.go` | Call frame lifecycle for frame transactions. |
| `pkg/core/types/tx_frame.go` | `FrameTx` type with validation frame support. |
| `pkg/core/vm/precompile_aa_proof.go` | AA proof precompile (0x0205): code hash, storage proof, validation result verification. |

#### Vectorized math precompiles (for lattice ops)

| File | What it provides |
|------|----------------|
| `pkg/core/vm/precompile_ntt.go` | NTT precompile (EIP-7885, addr 0x15). BN254 scalar field + Goldilocks field. Forward/inverse NTT. Gas: 1000 base + 10/element. Max degree 65536. |
| `pkg/core/vm/precompile_field.go` | Field arithmetic precompiles: modexp, field-mul, field-inv, batch-verify. |
| `pkg/core/vm/nii_precompile.go` | NII batch Merkle inclusion proof precompile. |
| `pkg/crypto/pqc/kyber_ntt.go` | Kyber NTT (in-place, modulus 3329). Used by lattice ops. |
| `pkg/crypto/pqc/poly_ring.go` | Polynomial ring arithmetic Rq = Zq[X]/(X^n+1). |

#### PQ signature algorithms

| File | What it provides |
|------|----------------|
| `pkg/crypto/pqc/mldsa_signer.go` | ML-DSA-65 (FIPS 204), real lattice signer via `circl`. |
| `pkg/crypto/pqc/dilithium.go`, `dilithium_sign.go` | Dilithium3 full lattice sign/verify. |
| `pkg/crypto/pqc/falcon_signer.go`, `falcon.go` | Falcon512 signer. |
| `pkg/crypto/pqc/sphincs_signer.go`, `sphincs_sign.go` | SPHINCS+SHA256 stateless hash-based. |
| `pkg/crypto/pqc/pq_tx_signer.go` | Transaction-level PQ signer: wraps Dilithium3/ML-DSA for signing `FrameTx` and `AATx`. |
| `pkg/crypto/pqc/pq_signing_pipeline.go` | Full signing pipeline: select algorithm → sign → attach to tx. |
| `pkg/crypto/pqc/hybrid.go`, `hybrid_threshold.go` | Hybrid classic+PQ signer for transition period. |

#### Pubkey registry

| File | What it provides |
|------|----------------|
| `pkg/crypto/pqc/pubkey_registry.go` | On-chain registry: register PQ pubkey, look up by address, migration from ECDSA. |

### GAPS / TODO

#### GAP 3.1 — Blake3 backend is a structural placeholder

`hash_backend.go` `Blake3Backend.Hash()` uses two rounds of SHA-256 with domain strings. Real BLAKE3 (tree hashing, parallel lanes, keyed mode) is not integrated. Hash-based signatures at the "~200k gas" range Vitalik mentions assume an efficient hash — BLAKE3 is the leading candidate for non-ZK contexts.

**Action needed**:
```
cd pkg && go get lukechampine.com/blake3
```
Replace `Blake3Backend.Hash()` with `blake3.Sum256(data)`.

#### GAP 3.2 — No gas cost benchmarks for PQ sig verification

Vitalik gives a concrete baseline: ECDSA = 3000 gas, hash-based PQ ≈ 200k gas. ETH2030 has no gas benchmarks or on-chain cost model for:
- Falcon512 verify (EVM execution cost)
- Dilithium3/ML-DSA verify (EVM execution cost)
- SPHINCS+ verify
- Hash-based WOTS+ verify via EIP-8141 validation frame

**Action needed**:
Add `pkg/core/vm/pq_precompile_gas_test.go` that benchmarks each PQ algorithm's verification gas cost via the EVM, comparing against the 3000-gas ECDSA baseline.

#### GAP 3.3 — NTT precompile missing vector dot-product and butterfly ops

Vitalik mentions: "+, *, %, **dot product**, also **NTT / butterfly permutations**". The NTT precompile covers NTT forward/inverse. But `precompile_ntt.go` does not expose:
- Vector dot product (inner product of two coefficient vectors)
- Butterfly permutation (index bit-reversal)
- Vectorized modular multiply-accumulate

These are needed to make Dilithium/Falcon verification fast enough in the EVM.

**Action needed**:
Extend `pkg/core/vm/precompile_ntt.go` with additional op codes:
- `NTTOpDotProduct = 4` — inner product of two n-length vectors mod q
- `NTTOpButterfly = 5` — Cooley-Tukey butterfly permutation
- `NTTOpVecMulAcc = 6` — vector multiply-accumulate mod q

#### GAP 3.4 — No end-to-end gas reduction path documented

Protocol-layer recursive aggregation (see Area 4) should reduce PQ sig gas to near-zero. The connection between EIP-8141 validation frames, STARK replacement, and the resulting gas saving is not documented or tested end-to-end.

---

## 4. Application-Layer ZK Proofs (Groth16/KZG → Recursive STARK Aggregation)

### What Vitalik says

> Today: ZK-SNARK costs ~300-500k gas; quantum-resistant STARK is ~10M gas. The latter is **unacceptable** for privacy protocols, L2s, and other proof users.
>
> Solution: **protocol-layer recursive signature and proof aggregation**.
>
> In EIP-8141, transactions have a "validation frame" during which signature verifications happen. Validation frames **cannot access the outside world** — only calldata in, return value out. This is designed so it's possible to **replace any validation frame (and its calldata) with a STARK** that verifies it.
>
> A block could "contain" a thousand validation frames (each 3kB–256kB), but they never come onchain — a single STARK verifying all of them does.
>
> This proving could happen at the **mempool layer**: every 500ms, each node passes along new valid transactions plus a proof verifying them. Overhead is static: **one proof per 500ms**.
>
> Reference: https://ethresear.ch/t/recursive-stark-based-bandwidth-efficient-mempool/23838

### What is DONE

#### EIP-8141 validation frames (the substrate)

| File | What it provides |
|------|----------------|
| `pkg/core/vm/eip8141_opcodes.go` | Full `FrameContext`, `APPROVE`/`TXPARAM*` opcodes. Validation frame isolation enforced: frames have their own calldata, cannot access world state. |
| `pkg/core/vm/call_frame.go` | Frame execution with isolation boundary. |
| `pkg/core/types/tx_frame.go` | `FrameTx` wire type with frames array. |

#### Proof infrastructure

| File | What it provides |
|------|----------------|
| `pkg/proofs/stark_prover.go` | STARK prover with FRI over Goldilocks field. `Prove(trace, constraints)` → `STARKProofData`. `Verify()`. |
| `pkg/proofs/recursive_prover.go` | Binary-tree recursive proof composition. `ComposeRecursive(proofs)` builds Merkle tree of proof roots. |
| `pkg/proofs/recursive_aggregator.go` | Multi-strategy aggregation: Sequential, Parallel, Recursive. Per-type verifiers for SNARK/STARK/IPA/KZG. |
| `pkg/proofs/groth16_verifier.go` | Groth16 proof size validation (placeholder for full `gnark` circuit proving). |
| `pkg/proofs/mandatory.go`, `mandatory_proofs.go` | Mandatory 3-of-5 proof system (prover assignment, submission, penalty). |

#### Mempool STARK aggregation ticks (directly from ethresear.ch post)

| File | What it provides |
|------|----------------|
| `pkg/txpool/stark_aggregation.go` | `MempoolSTARKAggregator`. Runs every `DefaultTickInterval = 500ms`. Each tick: collects validated txs, generates `STARKProofData` over all their validation proofs, produces `MempoolAggregationTick`. `MaxTickSize = 128KB` per ethresear.ch spec. Peer sharing of ticks. |

This is the direct implementation of Vitalik's ethresear.ch proposal.

#### AA proof circuits

| File | What it provides |
|------|----------------|
| `pkg/proofs/aa_proof_circuits.go` | AA proof circuit types: nonce constraint, sig constraint, gas constraint. |
| `pkg/proofs/aa_proofs.go` | AA proof generation and verification. |
| `pkg/core/vm/precompile_aa_proof.go` | On-chain AA proof precompile (0x0205). |

### GAPS / TODO

#### GAP 4.1 — Validation frame → STARK replacement not wired end-to-end

The most critical missing piece: the block builder/validator must be able to:
1. Receive a block with N validation frames
2. Verify all frames are correct locally
3. Replace the frames + their calldata with a single STARK proof in the block header

Steps 1–2 are done (EIP-8141 execution). Step 3 — the actual **replacement of validation frame calldata with a STARK** — is not implemented. No code removes validation frames from the block and substitutes a proof.

**Action needed**:
Add `pkg/core/vm/frame_stark_replacer.go`:
- `ReplaceValidationFrames(block, starkProver) (*Block, *STARKProof, error)` — extracts all validation frames, generates a STARK proving their collective validity, returns a stripped block + proof.
- Wire into block sealing in `pkg/engine/` and block import in `pkg/core/`.

#### GAP 4.2 — STARK prover not yet circuit-specialized for validation frames

`pkg/proofs/stark_prover.go` proves arbitrary algebraic traces. A validation frame is specifically: "run this calldata through the EVM, output must be non-zero." The circuit for this (mini-EVM trace over Goldilocks field) does not exist.

**Action needed**:
Add `pkg/proofs/validation_frame_circuit.go`:
- Define the EVM trace constraint system for validation frame execution.
- Expose `ProveValidationFrame(frameCalldata, output []byte) (*STARKProofData, error)`.
- Batch version: `ProveAllValidationFrames(frames [][]byte) (*STARKProofData, error)`.

#### GAP 4.3 — Groth16 verifier is a placeholder

`pkg/proofs/groth16_verifier.go` validates proof size and structure but does not perform real pairing-based verification. Real Groth16 requires `gnark` circuit proving backend (refs: `refs/gnark/`).

**Action needed**: Integrate `github.com/consensys/gnark` for real Groth16 proving/verification. Current placeholder is acceptable for structure but cannot produce/verify production proofs.

#### GAP 4.4 — STARK proof size vs. blob size problem unresolved

Vitalik notes: "a STARK is bigger than a blob." For the DA case this requires **recursive STARKs**. `pkg/proofs/recursive_prover.go` implements recursive *aggregation* (Merkle root over proof hashes) but not true recursive STARK composition (a STARK proving another STARK is valid, reducing proof size).

**Action needed** (research phase first):
- Reference `refs/research/circlestark/` for Circle STARK recursive composition.
- Design `RecursiveSTARKProver` that produces constant-size proofs regardless of inner proof count.
- Evaluate proof size: target < 128KB per EIP-7594 blob size for DA erasure proofs.

#### GAP 4.5 — Mempool STARK aggregation not connected to p2p gossip

`pkg/txpool/stark_aggregation.go` generates 500ms ticks with STARK proofs but the tick distribution over p2p (gossip or req-resp) is not wired. `pkg/p2p/gossip_topics.go` has the topic infrastructure but no mempool-tick topic.

**Action needed**:
Add gossip topic `mempool-stark-tick/1` in `pkg/p2p/gossip_topics.go`. Wire `stark_aggregation.go` tick output to gossip broadcast and tick ingestion from peers into local tx validity cache.

---

## Summary: Prioritized TODO List

### P0 — Prerequisite decisions (must resolve before implementation)

| ID | Task | Area |
|----|------|------|
| P0-A | Choose "Ethereum's last hash function": benchmark Poseidon1 vs Poseidon2+rounds vs BLAKE3 vs Monolith in ZK circuit and EVM contexts | Area 1 |
| P0-B | Confirm 1D PeerDAS (no 2D) as the final DA architecture — document in `docs/` | Area 2 |

### P1 — Near-term (high leverage, relatively contained)

| ID | Task | File to create/modify | Area |
|----|------|-----------------------|------|
| P1-A | Integrate real BLAKE3 into `hash_backend.go` | `pkg/crypto/pqc/hash_backend.go` | 1, 3 |
| P1-B | Add Poseidon2 extra-rounds variant | `pkg/zkvm/poseidon2.go` | 1 |
| P1-C | Add `LeanAvailableChainMode` (hash sigs, no STARK aggregation, 256-1024 sigs/slot) | `pkg/consensus/config.go`, `pq_attestation.go` | 1 |
| P1-D | Gas benchmarks for PQ sig algorithms via EVM | `pkg/core/vm/pq_precompile_gas_test.go` | 3 |
| P1-E | NTT precompile: add dot-product, butterfly, vec-mul-acc ops | `pkg/core/vm/precompile_ntt.go` | 3 |
| P1-F | Wire mempool STARK ticks to p2p gossip | `pkg/p2p/gossip_topics.go`, `pkg/txpool/stark_aggregation.go` | 4 |

### P2 — Medium-term (significant engineering, K+/L+ milestones)

| ID | Task | File to create | Area |
|----|------|---------------|------|
| P2-A | `frame_stark_replacer.go`: replace validation frames with STARK in block sealing | `pkg/core/vm/frame_stark_replacer.go` | 4 |
| P2-B | `validation_frame_circuit.go`: EVM mini-trace STARK circuit for frame verification | `pkg/proofs/validation_frame_circuit.go` | 4 |
| P2-C | Promote `STARKCommitment` from stub to working DA commitment type | `pkg/das/stark_commitment.go` | 2 |
| P2-D | `erasure_stark_prover.go`: STARK proving RS encoding correctness | `pkg/das/erasure_stark_prover.go` | 2 |
| P2-E | Integrate `gnark` for real Groth16 proving | `pkg/proofs/groth16_verifier.go` | 4 |
| P2-F | Add `MonolithBackend` (Poseidon2 + lookup non-arithmetic layer) | `pkg/crypto/pqc/hash_backend.go` | 1 |

### P3 — Long-term research (M+ milestone, 2029+)

| ID | Task | Area |
|----|------|------|
| P3-A | Recursive STARK composition (constant-size proof regardless of inner count) — reference `refs/research/circlestark/` | 2, 4 |
| P3-B | STARK aggregation circuit specialized for Winternitz signature verification (1M attestations) | 1 |
| P3-C | Distributed blob selection logistics with STARK erasure proofs | 2 |
| P3-D | "Near-zero gas" protocol-layer PQ sig aggregation — full pipeline test | 3, 4 |

---

## Code Reference Quick-Map

```
Area 1 — Consensus BLS → Hash+STARK
  pkg/crypto/pqc/unified_hash_signer.go      ← XMSS/WOTS+ (done)
  pkg/crypto/pqc/hash_backend.go             ← pluggable hash (Blake3: placeholder)
  pkg/crypto/pqc/l1_hash_sig.go              ← L1 Winternitz OTS (done)
  pkg/zkvm/poseidon.go                        ← Poseidon1 (done)
  pkg/zkvm/poseidon2.go                       ← Poseidon2 (done, no extra-rounds variant)
  pkg/consensus/pq_attestation.go            ← PQ attestations (done)
  pkg/consensus/stark_sig_aggregation.go     ← STARK aggregation (done, prover placeholder)
  pkg/consensus/pq_chain_security.go         ← SHA-3 fork choice (done)

Area 2 — DA KZG → STARK
  pkg/das/                                    ← full 1D PeerDAS (done)
  pkg/das/erasure/                            ← Reed-Solomon (done)
  pkg/das/pq_blobs.go                        ← PQ blob commitments (done, not replacing KZG)
  pkg/das/types.go:114                        ← STARKCommitment stub (gap)
  pkg/proofs/stark_prover.go                  ← STARK prover (done, not wired to DA)

Area 3 — EOA → Native AA + PQ sigs
  pkg/core/vm/eip8141_opcodes.go             ← EIP-8141 frames (done)
  pkg/core/vm/precompile_ntt.go              ← NTT precompile (done, missing dot-product/butterfly)
  pkg/crypto/pqc/mldsa_signer.go             ← ML-DSA-65 (done)
  pkg/crypto/pqc/falcon_signer.go            ← Falcon512 (done)
  pkg/crypto/pqc/pq_tx_signer.go             ← PQ tx signing (done)
  pkg/crypto/pqc/hash_backend.go             ← Blake3 placeholder (gap)

Area 4 — ZK proofs → Recursive STARK via EIP-8141
  pkg/core/vm/eip8141_opcodes.go             ← validation frames (done, no replacement step)
  pkg/proofs/stark_prover.go                  ← STARK prover (done, no frame circuit)
  pkg/proofs/recursive_prover.go             ← recursive composition (done)
  pkg/txpool/stark_aggregation.go            ← 500ms mempool ticks (done, p2p not wired)
  pkg/proofs/groth16_verifier.go             ← Groth16 (placeholder, needs gnark)
```

---

## Spec References (from `refs/`)

### NTT Precompile — Exact Addresses and Gas (ntt-eip)

Source: `refs/ntt-eip/EIP/EIPNTT.md`

| Precompile | Address | Gas | Field support |
|---|---|---|---|
| `NTT_FW` | `0x0f` | 600 gas flat | Falcon q=12289, Dilithium q=8380417, Goldilocks, M31, BN254... |
| `NTT_INV` | `0x10` | 600 gas flat | Same as FW |
| `NTT_VECMULMOD` | `0x11` | `k * log₂(n) / 8` | Element-wise modular multiply |
| `NTT_VECADDMOD` | `0x12` | `k * log₂(n) / 32` | Element-wise modular add |

Polynomial multiplication formula (from spec): `f × g = NTT_INV(NTT_VECMULMOD(NTT_FW(f), NTT_FW(g)))`

**GAP 3.3 resolution**: Dot-product and butterfly ops are addressed by `NTT_VECMULMOD` (inner product of NTT-domain vectors) and `NTT_VECADDMOD`. The spec does not define a separate "butterfly permutation" precompile — butterfly is internal to NTT_FW/NTT_INV. ETH2030's missing GAP 3.3 ops (dot-product, butterfly, vec-mul-acc) map to `0x11` and `0x12` in the ntt-eip spec.

**Falcon gas benchmarks** (from `refs/ethfalcon/doc/benchmarks.md`):
| Implementation | Gas Cost | Notes |
|---|---|---|
| Tetration (recursive NTT, Solidity) | 24M gas | Baseline |
| ZKNOX recursive NTT (Solidity) | 8.3M gas | Optimized |
| ZKNOX iterative Yul | 3.9M (NIST) / 1.6M (EVM-friendly) | With precomputed PK |
| EPERVIER (recovery + hint, Yul) | **1.5M gas** | Best known |
| With NTT precompile (0x0f–0x12) | **~1,500 gas** | 1000× speedup |

**Known Falcon CVEs** (from `refs/ethfalcon/`):
- `CVETH-2025-080201` **CRITICAL**: Salt size not checked during verification — allows bypass
- `CVETH-2025-080202` **MEDIUM**: Signature malleability on coefficient signs
- `CVETH-2025-080203` **LOW**: Missing domain separation in XOF

These vulnerabilities affect the `ethfalcon` EPERVIER implementation. Our `pkg/crypto/pqc/falcon_signer.go` uses Go bindings to the C Falcon reference implementation — verify these issues don't apply to the C bindings (`refs/ethfalcon/falcon/falcon.go`).

---

### ML-DSA-65 / Dilithium3 — Exact Sizes

Source: `refs/circl/sign/mldsa/mldsa65/internal/params.go`, `refs/circl/sign/dilithium/mode3/internal/params.go`

| Algorithm | PubKey | PrivKey | Sig | K | L | Eta | Tau |
|---|---|---|---|---|---|---|---|
| ML-DSA-65 (FIPS 204) | 1,952 B | 4,000 B | 3,309 B | 6 | 5 | 4 | 49 |
| Dilithium3 (pre-NIST) | 1,952 B | 4,000 B | 3,309 B | 6 | 5 | 4 | 49 |
| ML-DSA-44 | 1,312 B | 2,560 B | 2,420 B | 4 | 4 | 2 | 39 |
| ML-DSA-87 | 2,592 B | 4,896 B | 4,627 B | 8 | 7 | 2 | 60 |

**ML-DSA vs Dilithium3**: Same parameters (K=6, L=5), different standards. ML-DSA-65 adds optional context string and randomized signing mode; Dilithium3 uses 32-byte hash, ML-DSA uses 64-byte hash (CTildeSize=48).

**Go API** (production-ready from `refs/circl/sign/mldsa/mldsa65/`):
```go
// Randomized signing with context (ML-DSA-65 FIPS 204)
func SignTo(sk *PrivateKey, msg, ctx []byte, randomized bool, sig []byte) error

// Verification
func Verify(pk *PublicKey, msg, ctx []byte, sig []byte) bool

// Key generation from seed (deterministic)
func NewKeyFromSeed(seed *[32]byte) (*PublicKey, *PrivateKey)
```

**GAP 3.2** (gas benchmarks): Using the sig sizes above + ECDSA baseline of 3,000 gas:
- 3,309-byte ML-DSA sig vs 65-byte ECDSA sig = ~51× larger calldata
- At 16 gas/byte calldata: 3,309 × 16 = **52,944 gas** just for calldata
- Plus verification EVM cost: **~200,000 gas** total (consistent with Vitalik's "~200k range")
- With NTT precompile: reduce verification cost to ~5,000–10,000 gas

---

### Hash-Based Signatures — XMSS/WOTS+ Parameters

Source: `refs/hash-sig/src/signature/generalized_xmss/instantiations_sha.rs`, `refs/hash-sig/README.md`

This repo transitioned to [leanSig](https://github.com/leanEthereum/leanSig) in November 2025 (preserved status). ETH2030 implements `pkg/crypto/pqc/unified_hash_signer.go` (XMSS + WOTS+) and `pkg/crypto/pqc/l1_hash_sig.go` (Winternitz OTS).

**Winternitz parameter sets** (lifetime 2¹⁸, from spec):

| W | CHUNK_SIZE | NUM_CHUNKS | HASH_LEN | Winternitz_w |
|---|---|---|---|---|
| W=1 | 1 bit | 144 | 25 bytes | 8 |
| W=2 | 2 bits | 72 | 25 bytes | 4 |
| W=4 | 4 bits | 36 | 26 bytes | 3 |
| W=8 | 8 bits | 18 | 28 bytes | 2 |

**XMSS tree heights** (= log₂(lifetime)):
- Lifetime 2⁸ (mini): height 8
- Lifetime 2¹⁸: height 18 (1M signatures)
- Lifetime 2³²: height 32

**Hash backends** available in hash-sig repo:
- SHA-based: `ShaMessageHash`, `ShaTweakHash`, `ShaPRF`
- Poseidon-based: `PoseidonMessageHash`, `PoseidonTweakHash`, `PoseidonPRF`

**Key exhaustion**: Each (sk, epoch) pair must be used exactly once. `sk.advance_preparation()` generates new key material in background. This maps to `XMSSKeyManager` in `pkg/crypto/pqc/unified_hash_signer.go`.

---

### 3SF Research — Spec Reference

Source: `refs/research/3sf-mini/consensus.py` (Vitalik's reference implementation, 164 lines)

**Key functions**:
```python
# Justification backoff algorithm
def is_justifiable_slot(delta):
    # Allows justification if:
    # delta <= 5, OR delta is a perfect square, OR delta is oblong (x²+x)
    return delta <= 5 or is_perfect_square(delta) or is_oblong(delta)

# Two-phase finalization
def finalize_slot(state, slot):
    # latest_justified_slot → latest_finalized_slot
    # Requires no other justifiable slots in between

# Fork choice
def get_fork_choice_head(store):
    # LMD GHOST with vote weighting
    # 2/3 threshold for justification
```

**P2P simulation** (`refs/research/3sf-mini/p2p.py`, 224 lines):
- Safe target computed at `t=2/4` of slot time
- View merge: attestation accepted at `t≤1/4` OR inside latest block
- Dependencies queue for orphaned blocks/votes

**ETH2030 alignment**: `pkg/consensus/ssf.go` implements 4-phase SSF; `pkg/consensus/endgame_pipeline.go` implements <500ms target. The 3SF `is_justifiable_slot()` backoff algorithm is distinct from our SSF 4-phase state machine. For GAP "Different Approach #15" in `vitalik-roadmap-gaps.md` — our SSF is functionally equivalent but uses a different (multi-round) finality mechanism.

---

### CircleSTARK — Recursive STARK Reference

Source: `refs/research/circlestark/`

**FRI parameters** for P3-A (recursive STARK composition):
- Base case: 64 evaluations
- Folds per step: 3 (8× degree reduction)
- Security queries: 80
- Proof size: O(log²(degree))

For `GAP 4.4` (STARK is bigger than a blob): the Circle STARK approach achieves sub-linear proof sizes via recursive STARK composition using Poseidon over M31. The `refs/research/circlestark/` directory contains the full Python reference implementation that should guide `pkg/proofs/recursive_prover.go` improvements.

---

*Generated: 2026-03-04. Based on Vitalik Buterin's quantum resistance roadmap message.*
