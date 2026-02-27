# Post-Quantum Implementation Report: ETH2030

> Full-stack post-quantum cryptography for Ethereum, aligned with Vitalik's quantum resistance roadmap and the EF Strawmap.
> Generated: February 27, 2026

---

## Context: Why Post-Quantum Now?

### Vitalik's Quantum Resistance Roadmap (Feb 26, 2026)

On February 26, 2026, Vitalik Buterin published a [comprehensive quantum resistance roadmap](https://x.com/VitalikButerin/status/2027075026378543132) identifying **four pillars** of Ethereum vulnerable to quantum computing:

> "Today, four things in Ethereum are quantum-vulnerable: consensus-layer BLS signatures;
> data availability (KZG commitments+proofs); EOA signatures (ECDSA); Application-layer
> ZK proofs (KZG or groth16)."

Key points from the thread:
- **CL Signatures**: Replace BLS with hash-based (Winternitz variants), aggregate via STARKs
- **Hash function**: "This may be 'Ethereum's last hash function,' so it's important to choose wisely" — candidates: Poseidon2, Poseidon1, BLAKE3
- **DA**: STARKs replace KZG but lack linearity for 2D DAS; 1D PeerDAS may suffice
- **EOA**: ECDSA ~3K gas → hash-based PQ ~200K gas; EIP-8141 frame transactions enable alternative sig algorithms
- **ZK Proofs**: ZK-SNARKs 300-500K gas → quantum-resistant STARKs ~10M gas
- **Solution**: Protocol-layer recursive STARK aggregation — one proof per block

### Earlier Key Posts

| Date | Source | Key Quote |
|------|--------|-----------|
| Mar 2024 | [ethresear.ch](https://ethresear.ch/t/how-to-hard-fork-to-save-most-users-funds-in-a-quantum-emergency/18901) | "The infrastructure needed to implement such a hard fork could in principle start to be built tomorrow." |
| Oct 2024 | [Blog: The Merge](https://vitalik.eth.limo/general/2024/10/14/futures1.html) | "Each piece of Ethereum that currently depends on elliptic curves will need some hash-based or quantum-resistant replacement." |
| Oct 2024 | [Blog: The Verge](https://vitalik.eth.limo/general/2024/10/23/futures4.html) | Skip Verkle trees → binary tree + STARKs (Polygon: 1.7M Poseidon hashes/sec with circle STARKs) |
| Jan 2026 | [Walkaway Test](https://www.coindesk.com/tech/2026/01/12/vitalik-buterin-lays-out-walkaway-test-for-a-quantum-safe-ethereum) | Quantum resistance is a "top priority" — protocol must be safe for decades |
| Jan 2026 | [EF PQ Team formed](https://x.com/drakefjustin/status/2014791629408784816) | Led by Thomas Coratger; $1M prizes for Poseidon hardening and PQ proximity problems |

### EF Strawmap (strawmap.org, Feb 25, 2026)

**"Post-Quantum L1"** is one of the five North Stars: "centuries-long cryptographic security, via hash-based schemes."

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────────────────┐
│                         ETH2030 Post-Quantum Stack                       │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐  │
│  │  Consensus   │  │  Data Layer  │  │  Execution   │  │  Application│  │
│  │   Layer      │  │              │  │   Layer      │  │   Layer     │  │
│  ├─────────────┤  ├──────────────┤  ├──────────────┤  ├─────────────┤  │
│  │ PQ Attest.  │  │ Lattice Blob │  │ PQ Tx Signer │  │ STARK Proofs│  │
│  │ STARK Agg.  │  │ Commitments  │  │ NTT Precomp. │  │ Recursive   │  │
│  │ jeanVM Agg. │  │ Custody Proof│  │ NII Precomp. │  │ Aggregation │  │
│  │ PQ Chain    │  │ Merkle-based │  │ Encrypted    │  │ AA Proofs   │  │
│  │ Security    │  │ PQ Blobs     │  │ Mempool      │  │             │  │
│  └──────┬──────┘  └──────┬───────┘  └──────┬───────┘  └──────┬──────┘  │
│         │                │                  │                 │          │
│  ┌──────┴────────────────┴──────────────────┴─────────────────┴──────┐  │
│  │                     crypto/pqc/ (33 files)                        │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐ │  │
│  │  │ ML-DSA-65│ │Falcon-512│ │ SPHINCS+ │ │XMSS/WOTS+│ │ Hybrid │ │  │
│  │  │ FIPS 204 │ │ NTRU     │ │ Stateless│ │ Stateful │ │ECDSA+PQ│ │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └────────┘ │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────────┐│  │
│  │  │  NTT     │ │ Lattice  │ │  Hash    │ │ Algorithm Registry   ││  │
│  │  │Arithmetic│ │  Commit  │ │ Backends │ │ (EIP-7932 dispatch)  ││  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────────────────┘│  │
│  └───────────────────────────────────────────────────────────────────┘  │
│                                                                          │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                    go-ethereum Integration                        │  │
│  │  geth/extensions.go → PrecompileAdapter → gethvm.SetPrecompiles  │  │
│  │  cmd/eth2030-geth/ → --override.iplus=0 → InjectIntoGethPrague  │  │
│  └───────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## Pillar 1: Consensus Layer — PQ Attestations

**Threat**: BLS12-381 signatures broken by Shor's algorithm.

```
┌─────────────────────────────────────────────────────────────┐
│                   PQ Attestation Flow                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Validator 1 ──Dilithium3──┐                                │
│  Validator 2 ──Dilithium3──┤                                │
│  Validator 3 ──Dilithium3──┼──→ STARK Aggregator ──→ Block  │
│       ...                  │       │                        │
│  Validator N ──Dilithium3──┘       │                        │
│                                    ▼                        │
│                            Single STARK proof               │
│                            (replaces N BLS sigs)            │
│                                                              │
│  Fallback: classic ECDSA/BLS during transition period       │
└─────────────────────────────────────────────────────────────┘
```

### Implementation

| Component | File | Description |
|-----------|------|-------------|
| PQ Attestations | `consensus/pq_attestation.go` | Dilithium3 signatures with classic fallback |
| STARK Sig Aggregation | `consensus/stark_sig_aggregation.go` | STARK proof over N Dilithium signatures |
| jeanVM Aggregation | `consensus/jeanvm_aggregation.go` | Groth16 ZK-circuit BLS (transition) |
| PQ Chain Security | `consensus/pq_chain_security.go` | SHA-3 fork choice, enforcement thresholds |
| CL Proof Circuits | `consensus/cl_proof_circuits.go` | SHA-256 Merkle proof generation |

### PQ Chain Security Model

```
Security Level:    Optional → Preferred → Required
                       │          │           │
PQ Validator %:      <33%      33-67%       >67%
                       │          │           │
Enforcement:        Accept    Warn on      Reject
                    both      classic      classic
```

---

## Pillar 2: Data Availability — Lattice Blob Commitments

**Threat**: KZG commitments use elliptic curve pairings (broken by Shor's).

```
┌──────────────────────────────────────────────────────────────┐
│              PQ Blob Commitment Schemes                       │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  Scheme 1: Merkle-Tree Based (hash-only)                     │
│  ┌─────────────────────────────┐                             │
│  │       Merkle Root            │                             │
│  │        /      \              │                             │
│  │     H(01)    H(23)          │  Hash: Keccak-256           │
│  │     / \      / \            │  Chunks: 32 bytes each      │
│  │   H(0) H(1) H(2) H(3)     │  Proof: authentication path │
│  │    │    │    │    │        │                               │
│  │   [blob chunk data...]     │                               │
│  └─────────────────────────────┘                             │
│                                                               │
│  Scheme 2: Module-LWE Lattice (algebraic)                    │
│  ┌─────────────────────────────┐                             │
│  │  c = A·s + e + m            │  A: public matrix           │
│  │                              │  s: secret vector           │
│  │  Parameters:                 │  e: error vector (small)    │
│  │    k=2, η=2 (Kyber Level 1)│  m: message (blob data)     │
│  │    q=3329                    │                             │
│  │                              │  Binding: MLWE hardness     │
│  │  Verify: recompute c,       │  Hiding: statistical        │
│  │          check equality     │                              │
│  └─────────────────────────────┘                             │
│                                                               │
│  Batch Verification: parallel Merkle root checks             │
│  Custody Proofs: lattice-based (EIP-7594 PeerDAS)            │
└──────────────────────────────────────────────────────────────┘
```

### Implementation

| Component | File | Description |
|-----------|------|-------------|
| Merkle Blob Commit | `crypto/pqc/blob_commitment.go` | Hash-based PQ blob commitment |
| Lattice Commit | `crypto/pqc/lattice_commit.go` | MLWE-based binding commitment |
| Batch Verify | `crypto/pqc/batch_blob_verify.go` | Parallel Merkle proof verification |
| Custody Proofs | `crypto/pqc/custody_replacer.go` | Lattice custody for PeerDAS |
| Lattice Blobs | `crypto/pqc/lattice_blob.go` | Full lattice blob pipeline |

---

## Pillar 3: Execution Layer — PQ Transaction Signatures

**Threat**: ECDSA broken by Shor's algorithm; public key exposed after first transaction.

```
┌──────────────────────────────────────────────────────────────────┐
│              PQ Signature Algorithm Landscape                     │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Algorithm       │ Type      │ Sig Size │ PK Size  │ Gas Cost    │
│  ────────────────┼───────────┼──────────┼──────────┼─────────────│
│  ECDSA (current) │ Elliptic  │   65 B   │   64 B   │   3,000     │
│  ML-DSA-65       │ Lattice   │ 1,376 B  │ 1,568 B  │   4,500     │
│  Dilithium-3     │ Lattice   │ 3,293 B  │ 1,952 B  │   4,500     │
│  Falcon-512      │ NTRU      │   690 B  │   897 B  │   3,000     │
│  SPHINCS+        │ Hash      │49,216 B  │    32 B  │   8,000     │
│  XMSS/WOTS+     │ Hash tree │ ~2.5 KB  │    64 B  │   6,000     │
│                                                                   │
│  EIP-7932 Algorithm Registry: unified dispatch by algorithm ID   │
│  EIP-8141 Frame Transactions: per-tx signature scheme override   │
│  Hybrid Mode: ECDSA + PQ dual-signing for transition period      │
└──────────────────────────────────────────────────────────────────┘
```

### NTT Precompile (EIP-7885, address 0x15)

The NTT (Number Theoretic Transform) precompile enables efficient polynomial evaluation for lattice signature verification and STARK proofs.

```
┌──────────────────────────────────────────────────────────────┐
│                  NTT Precompile (0x15)                        │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  Input: [1 byte op] [N × 32 byte coefficients]              │
│                                                               │
│  Operations:                                                  │
│  ┌──────┬─────────────────────────────────────────────────┐  │
│  │ 0x00 │ Forward NTT — BN254 scalar field                │  │
│  │      │ p = 21888...495617 (254-bit prime)              │  │
│  ├──────┼─────────────────────────────────────────────────┤  │
│  │ 0x01 │ Inverse NTT — BN254 scalar field                │  │
│  ├──────┼─────────────────────────────────────────────────┤  │
│  │ 0x02 │ Forward NTT — Goldilocks field                  │  │
│  │      │ p = 2^64 - 2^32 + 1 (STARK-friendly)           │  │
│  ├──────┼─────────────────────────────────────────────────┤  │
│  │ 0x03 │ Inverse NTT — Goldilocks field                  │  │
│  └──────┴─────────────────────────────────────────────────┘  │
│                                                               │
│  Gas: base(1000) + n × log2(n) × 10                         │
│  Max degree: 65,536 (2^16)                                   │
│                                                               │
│  Use cases:                                                   │
│  • Lattice sig verification (Falcon/Dilithium NTT domains)   │
│  • STARK proof verification (polynomial evaluation)          │
│  • ZK circuit polynomial multiplication                      │
└──────────────────────────────────────────────────────────────┘
```

### NII Field Precompiles (0x0201–0x0208)

```
┌──────────────────────────────────────────────────────────────┐
│            Number-Theoretic Integer Precompiles               │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  Address │ Name            │ Purpose                          │
│  ────────┼─────────────────┼──────────────────────────────── │
│  0x0201  │ niiModExp       │ Modular exponentiation           │
│  0x0202  │ niiFieldMul     │ Field multiplication             │
│  0x0203  │ niiFieldInv     │ Field inversion                  │
│  0x0204  │ niiBatchVerify  │ Batch verification               │
│  0x0205  │ fieldMulExt     │ Extended field multiplication    │
│  0x0206  │ fieldInvExt     │ Extended field inversion          │
│  0x0207  │ fieldExp        │ Field exponentiation             │
│  0x0208  │ batchFieldVerify│ Batch field verification         │
│                                                               │
│  Total: 13 custom precompiles (4 repriced + 1 NTT + 8 NII)  │
└──────────────────────────────────────────────────────────────┘
```

---

## Pillar 4: Application Layer — STARK Proof Aggregation

**Threat**: Groth16/KZG proofs broken by quantum computers.

```
┌──────────────────────────────────────────────────────────────────┐
│              Recursive STARK Aggregation Pipeline                  │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Block N transactions:                                            │
│                                                                   │
│  TX 1 ─→ [PQ sig + frame] ──┐                                   │
│  TX 2 ─→ [PQ sig + frame] ──┤                                   │
│  TX 3 ─→ [STARK proof]   ──┼──→ Mempool STARK Aggregator        │
│  TX 4 ─→ [PQ sig + frame] ──┤         │                          │
│       ...                    │         ▼                          │
│  TX N ─→ [PQ sig + frame] ──┘   Recursive STARK                 │
│                                        │                          │
│                                        ▼                          │
│                              ┌─────────────────┐                 │
│                              │  Single STARK    │                 │
│                              │  proof per block │                 │
│                              │  (~100 KB)       │                 │
│                              └─────────────────┘                 │
│                                                                   │
│  Aggregation cadence: every 500ms (recursive composition)        │
│  Proof systems: ZKSTARK, ZKSNARK (Groth16), IPA, KZG             │
│  Mandatory 3-of-5: at least 3 proof types must verify            │
└──────────────────────────────────────────────────────────────────┘
```

### Implementation

| Component | File | Description |
|-----------|------|-------------|
| STARK Prover | `proofs/stark_prover.go` | STARK proof generation with Goldilocks field |
| Recursive Prover | `proofs/recursive_prover.go` | Recursive STARK composition |
| Proof Aggregator | `proofs/aggregator.go` | Multi-proof-system registry |
| 3-of-5 System | `proofs/mandatory_proofs.go` | Mandatory multi-proof verification |
| Mempool Aggregation | `txpool/stark_aggregation.go` | STARK aggregation in mempool |
| AA Proofs | `proofs/aa_proof_circuits.go` | Account abstraction proof circuits |

---

## Fork Activation & Devnet Testing

### Fork Timeline

```
┌──────────────────────────────────────────────────────────────────┐
│                    ETH2030 Fork Activation                        │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Prague ──→ Glamsterdam ──→ Hogota ──→ I+ ──→ J+ ──→ K+ ──→ ...│
│  (geth)    (repricing)   (BPO blobs) (PQ)                        │
│                                                                   │
│  PQ Components by Fork:                                           │
│                                                                   │
│  Glamsterdam (2026):                                              │
│    ├── Gas repricing (ecAdd, ecPairing, blake2f, pointEval)      │
│    ├── EIP-8141 frame transactions (APPROVE, TXPARAM opcodes)    │
│    └── Encrypted mempool (commit-reveal)                          │
│                                                                   │
│  Hogota (2026-2027):                                              │
│    ├── NTT precompile preparation                                 │
│    ├── BPO blob schedules                                         │
│    └── Binary tree (SHA-256)                                      │
│                                                                   │
│  I+ (2027):                                                       │
│    ├── NTT precompile (0x15) — BN254 + Goldilocks                │
│    ├── NII precompiles (0x0201–0x0208) — field arithmetic        │
│    ├── PQ transaction signing (ML-DSA-65, algorithm registry)    │
│    ├── PQ pubkey registry                                         │
│    └── Native rollups + zkVM framework                            │
│                                                                   │
│  L+ (2029):                                                       │
│    ├── PQ attestations (Dilithium + STARK aggregation)           │
│    ├── PQ chain security enforcement                              │
│    ├── Lattice blob commitments (MLWE)                            │
│    └── jeanVM aggregation (Groth16 ZK-circuit BLS)               │
│                                                                   │
│  M+ (2029+):                                                      │
│    ├── PQ L1 hash-based signer (XMSS/WOTS+ unified)             │
│    ├── STARK recursive proof aggregation                          │
│    └── PQ blob commitments (MLWE lattice)                         │
└──────────────────────────────────────────────────────────────────┘
```

### go-ethereum Integration

```
┌──────────────────────────────────────────────────────────────────┐
│              Precompile Injection into go-ethereum                 │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  eth2030-geth binary startup:                                     │
│                                                                   │
│  1. Parse CLI flags:                                              │
│     --override.glamsterdam=0                                      │
│     --override.hogota=0                                           │
│     --override.iplus=0                                            │
│                                                                   │
│  2. Create precompileInjector(glam, hogota, iplus)               │
│                                                                   │
│  3. injector.InjectIntoGethPrecompiles()                          │
│     ├── Determine max fork level at genesis                       │
│     ├── For each custom precompile where maxLevel >= minFork:    │
│     │   └── gethvm.PrecompiledContractsPrague[addr] = adapter    │
│     └── Result: eth_call + block processing see custom precomps  │
│                                                                   │
│  4. geth.New() starts Ethereum service                            │
│     └── Uses patched PrecompiledContractsPrague map               │
│                                                                   │
│  ┌─────────────────────────────────────────────────┐             │
│  │  Fork Level    │ Custom Precompiles             │             │
│  │  ──────────────┼────────────────────────────── │             │
│  │  Prague (0)    │ 0 (standard go-ethereum only)  │             │
│  │  Glamsterdam(1)│ 4 (repriced ecAdd/pairing/...) │             │
│  │  Hogota (2)    │ 4 (same as Glamsterdam)        │             │
│  │  I+ (3)        │ 13 (+ NTT + 8 NII/field)      │             │
│  └─────────────────────────────────────────────────┘             │
└──────────────────────────────────────────────────────────────────┘
```

### Kurtosis Devnet Configuration

```yaml
# pq-crypto.yaml — I+ fork at genesis for PQ testing
participants:
  - el_type: geth
    el_image: eth2030:local
    el_extra_params:
      - "--override.glamsterdam=0"
      - "--override.hogota=0"
      - "--override.iplus=0"
    cl_type: lighthouse
    cl_image: sigp/lighthouse:latest
    count: 2
```

Devnet verification tests:
- Block production (chain operational)
- NTT BN254 forward/inverse round-trip
- NTT Goldilocks forward/inverse round-trip
- State evolution (stateRoot changes across blocks)
- Transaction pool status (STARK aggregation infrastructure)

---

## Cryptographic Security Assumptions

```
┌──────────────────────────────────────────────────────────────────┐
│              Security Assumptions vs Quantum Threat                │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Scheme        │ Assumption     │ Classical │ Quantum │ Status   │
│  ──────────────┼────────────────┼───────────┼─────────┼──────── │
│  ECDSA         │ ECDLP          │ 2^128     │ broken  │ Replace │
│  BLS12-381     │ DLP + pairing  │ 2^128     │ broken  │ Replace │
│  KZG           │ DLP + pairing  │ 2^128     │ broken  │ Replace │
│  Groth16       │ DLP + pairing  │ 2^128     │ broken  │ Replace │
│  ──────────────┼────────────────┼───────────┼─────────┼──────── │
│  ML-DSA-65     │ Module-LWE     │ 2^192     │ 2^128   │ ✓ Safe  │
│  Falcon-512    │ NTRU lattice   │ 2^128     │ 2^86    │ ✓ Safe  │
│  SPHINCS+      │ Hash collision │ 2^256     │ 2^128   │ ✓ Safe  │
│  XMSS/WOTS+   │ Hash collision │ 2^256     │ 2^128   │ ✓ Safe  │
│  SHA-3/Keccak  │ Sponge         │ 2^256     │ 2^128   │ ✓ Safe  │
│  MLWE Commit   │ Module-LWE     │ 2^192     │ 2^128   │ ✓ Safe  │
│  Merkle Tree   │ Hash collision │ 2^256     │ 2^128   │ ✓ Safe  │
│  STARK         │ Hash + FRI     │ 2^128     │ 2^128   │ ✓ Safe  │
└──────────────────────────────────────────────────────────────────┘
```

---

## Implementation Statistics

```
┌──────────────────────────────────────────────────────────────────┐
│                   PQ Implementation Metrics                       │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  crypto/pqc/:     33 source files, 31 test files (~21K LOC)     │
│  consensus/:       4 PQ-specific files (attestations, security)  │
│  core/vm/:         NTT precompile + 8 NII/field precompiles     │
│  proofs/:          STARK prover + recursive composition          │
│  txpool/:          STARK mempool aggregation                     │
│  geth/:            13 custom precompile adapters                 │
│                                                                   │
│  Signature Algorithms:  6 (ML-DSA, Dilithium, Falcon,            │
│                            SPHINCS+, XMSS/WOTS+, Hybrid)        │
│  Commitment Schemes:    2 (Merkle hash-based, MLWE lattice)     │
│  Proof Systems:         4 (STARK, SNARK/Groth16, IPA, KZG)      │
│  EVM Precompiles:      13 (4 repriced + 1 NTT + 8 NII)         │
│  EIPs Referenced:      12+ (7885, 7932, 8051, 8141, 7594, ...)  │
│                                                                   │
│  Roadmap Coverage:                                                │
│    CL BLS Signatures    → Dilithium + STARK aggregation    [✓]  │
│    DA KZG Commitments   → Merkle + MLWE lattice            [✓]  │
│    EOA ECDSA Signatures → ML-DSA-65 + algorithm registry   [✓]  │
│    App-layer ZK Proofs  → STARK recursive aggregation      [✓]  │
└──────────────────────────────────────────────────────────────────┘
```

---

## References

- Vitalik Buterin, [Quantum Resistance Roadmap](https://x.com/VitalikButerin/status/2027075026378543132), Feb 26, 2026
- Vitalik Buterin, [How to hard-fork to save most users' funds in a quantum emergency](https://ethresear.ch/t/how-to-hard-fork-to-save-most-users-funds-in-a-quantum-emergency/18901), Mar 2024
- Vitalik Buterin, [Possible futures of the Ethereum protocol](https://vitalik.eth.limo/general/2024/10/14/futures1.html) (6-part series), Oct 2024
- EF Architecture Team, [Ethereum Strawmap](https://strawmap.org), Feb 25, 2026
- Justin Drake, [EF Post-Quantum Team announcement](https://x.com/drakefjustin/status/2014791629408784816), Jan 23, 2026
- NIST, [FIPS 204: Module-Lattice-Based Digital Signature Standard (ML-DSA)](https://csrc.nist.gov/pubs/fips/204/final), 2024
- NIST, [FIPS 205: Stateless Hash-Based Digital Signature Standard (SLH-DSA)](https://csrc.nist.gov/pubs/fips/205/final), 2024
