# EL State Tree & VM Changes — Line-by-Line Status Analysis

> Source: Vitalik's message on Execution Layer changes (state tree + VM)
> Analyzed: 2026-03-04
> Covers: EIP-7864 binary tree, NTT/vectorized precompile, zkVM/RISC-V roadmap

---

## Executive Summary

| Area | Status | Notes |
|------|--------|-------|
| EIP-7864 binary trie (stem/leaf architecture) | **COMPLETE** | `pkg/trie/bintrie/` — SHA-256, StemNode=256 leaves |
| Binary tree proofs (inclusion + exclusion) | **COMPLETE** | `pkg/trie/bintrie/proof.go`, `proof_verifier.go` |
| Storage "pages" (adjacent slot co-location) | **COMPLETE** | StemNode groups 256 leaves sharing a 31-byte stem |
| MPT → binary tree migration | **COMPLETE** | `pkg/trie/migration.go` — batch, checkpoint, gas |
| State expiry + revival | **COMPLETE** | `pkg/core/state/state_expiry.go`, `expiry_engine.go` |
| SHA-256 for binary tree nodes | **COMPLETE** | `pkg/trie/bintrie/hasher_extended.go` |
| Poseidon hash (ZK-circuit use) | **COMPLETE** | `pkg/zkvm/poseidon.go` — BN254 field, Grain LFSR |
| Blake3 hash | **MISSING** | Not implemented; SHA-256 is the current tree hash |
| NTT / vectorized math precompile (EIP-7885) | **COMPLETE** | `pkg/core/vm/precompile_ntt.go` — BN254 + Goldilocks |
| RISC-V CPU emulator (RV32IM) | **COMPLETE** | `pkg/zkvm/riscv_cpu.go`, `riscv_memory.go` |
| Canonical guest framework | **COMPLETE** | `pkg/zkvm/canonical.go` — `GuestRegistry`, `CanonicalGuestPrecompile` |
| STF executor (state transition proof) | **COMPLETE** | `pkg/zkvm/stf_executor.go` |
| zkISA bridge (EVM→RISC-V) | **COMPLETE** | `pkg/zkvm/zkisa_bridge.go` — 9 op selectors |
| Step 1: Precompiles as RISC-V programs | **PARTIAL** | zkISA bridge maps ops to guests; no auto-conversion of EVM precompiles |
| Step 2: User RISC-V contracts | **PARTIAL** | `CanonicalGuestPrecompile` at 0x0200 accepts arbitrary RISC-V programs |
| Step 3: EVM as RISC-V smart contract | **MISSING** | Long-term; explicitly listed as longer-term/non-consensus |

---

## Line-by-Line Analysis

---

### Section A: Binary Trees

---

### A.1 "4x shorter Merkle branches (binary is 32×log(n) and hexary is 512×log(n)/4), which makes client-side branch verification more viable. This makes Helios, PIR and more 4x cheaper by data bandwidth."

**Status: COMPLETE**

The EIP-7864 binary trie uses a 31-byte stem to navigate binary internal nodes, with the 32nd byte selecting one of 256 values in a StemNode. Each tree level requires only one 32-byte sibling hash (vs 15 siblings per level in hexary MPT), giving the 4x branch-size reduction.

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/trie/bintrie/bintrie.go` | 1–8 | Package doc: "first 31 bytes form the stem, final byte selects one of 256 leaves in a StemNode" |
| `pkg/trie/bintrie/stem_node.go` | 14–19 | `StemNode` struct: Stem (31 bytes), Values (256 slots), parent-less |
| `pkg/trie/bintrie/node.go` | 21 | `StemNodeWidth = 256` |
| `pkg/trie/bintrie/proof.go` | 1–160 | `GenerateProof()`: collects one sibling hash per binary level, yielding O(log₂ n) path |
| `pkg/trie/bintrie/proof_verifier.go` | 1–350 | `VerifyInclusion()`, `VerifyExclusion()`: reconstructs root from binary path |

The light client (`pkg/light/`) can use these short proofs directly. The `CLProofGenerator` in `pkg/light/cl_proofs.go` generates 40-element Merkle paths (vs ~160 for hexary MPT at the same depth).

---

### A.2 "Proving efficiency. 3-4x comes from shorter Merkle branches. On top of that, the hash function change: either blake3 [perhaps 3x vs keccak] or a Poseidon variant [100x, but more security work to be done]."

**Status: PARTIAL** — SHA-256 deployed; Poseidon available for ZK circuits; Blake3 not implemented

#### SHA-256 (Current Production Hash)

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/trie/bintrie/hasher_extended.go` | 85–136 | `BinaryHasher.hashInternal()`: `SHA256(left ‖ right)` for branch nodes |
| `pkg/trie/bintrie/hasher_extended.go` | 222–238 | `HashLeafValue()`: `SHA256(leaf_data)` |
| `pkg/trie/bintrie/hasher_extended.go` | 110–136 | Parallel hashing: goroutines for subtrees exceeding configurable threshold |
| `pkg/trie/bintrie/bintrie.go` | 43–50 | `GetBinaryTreeKey()`: SHA-256 for all tree key derivation |
| `pkg/trie/bintrie/stem_node.go` | 92–123 | `StemNode.Hash()`: SHA-256 Merkle tree over 256 leaf values |

SHA-256 is ~3x more prover-efficient than Keccak-256 in ZK circuits (fewer constraints), though not as good as Poseidon.

#### Poseidon Hash (ZK-Circuit Use)

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/zkvm/poseidon.go` | 1–405 | Full Poseidon-BN254: SBox (x^5), MDS matrix, Grain LFSR round constants |
| `pkg/zkvm/poseidon.go` | 42–67 | Parameters: t=3 (rate=2, capacity=1), 8 full rounds + 57 partial rounds |
| `pkg/zkvm/poseidon.go` | 152–185 | `PoseidonHash()`: sponge construction for arbitrary-length inputs |
| `pkg/zkvm/poseidon2.go` | — | Poseidon2 variant (additional implementation) |

Poseidon is available for ZK circuits that compose with Ethereum state (e.g., a DApp using state proofs inside a SNARK). The binary tree itself still uses SHA-256, not Poseidon, which means a ZK prover verifying binary-tree paths must use SHA-256 constraints (~100-200 constraints per node) rather than Poseidon (~50 constraints). Switching the tree hash to Poseidon would reduce ZK circuit size by ~4x but requires additional security analysis (Vitalik notes "more security work to be done").

#### Blake3 — NOT IMPLEMENTED

Blake3 is referenced in Vitalik's description as a possible hash ("blake3 [perhaps 3x vs keccak]") but is not implemented in this codebase. The PQC test file references "blake3" as a string constant only.

**TODO (hash upgrade):**
- Add `pkg/trie/bintrie/hasher_blake3.go` using the `github.com/zeebo/blake3` Go library
- Benchmark Blake3 vs SHA-256 for binary tree node hashing
- Add `pkg/trie/bintrie/hasher_poseidon.go` for future Poseidon-backed tree (after security analysis)
- Make hash function configurable via fork parameter in chain config

---

### A.3 "Client-side proving: if you want ZK applications that compose with the ethereum state, instead of making their own tree like today, then the ethereum state tree needs to be prover-friendly."

**Status: COMPLETE** — proofs, stateless execution, VOPS all wired

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/trie/bintrie/proof_verifier.go` | 91–135 | `VerifyInclusion()`: verifies a value at a binary-tree path; usable inside ZK circuits |
| `pkg/witness/state_proof.go` | — | `ExecutionWitness`: binary Merkle proof branches with SHA-256, usable in ZK |
| `pkg/core/vops/` | — | `PartialState`: minimal state witness for stateless execution |
| `pkg/light/cl_proofs.go` | 58–100 | `CLStateProof`, `ValidatorProof`: Merkle proofs about Beacon state |
| `pkg/proofs/aa_proof_circuits.go` | — | Groth16 circuit over AA state — shows DApp-level ZK composition pattern |

A DApp using state proofs inside a SNARK would:
1. Obtain a binary-tree inclusion proof from `bintrie.GenerateProof()`
2. Verify it using `bintrie.VerifyInclusion()` or inline the SHA-256 Merkle path checks in a Groth16 circuit using Poseidon-compatible constraints

---

### A.4 "Cheaper access for adjacent slots: the binary tree design groups together storage slots into 'pages' (e.g. 64–256 slots, so 2–8 kB). This allows storage to get the same efficiency benefits as code in terms of loading and editing lots of it at a time, both in raw execution and in the prover. The block header and the first ~1-4 kB of code and storage live in the same page. Many dapps today already load a lot of data from the first few storage slots, so this could save them >10k gas per tx."

**Status: COMPLETE** — StemNode is the "page"; header + code + first 64 storage slots co-located

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/trie/bintrie/bintrie.go` | 21–37 | `headerStorageOffset=64`, `codeOffset=128`, `nodeWidthLog2=8` |
| `pkg/trie/bintrie/bintrie.go` | 67–83 | `GetBinaryTreeKeyStorageSlot()`: slots 0–63 map to leaf indices 64–127 in the same StemNode as account basic data |
| `pkg/trie/bintrie/bintrie.go` | 85–92 | `GetBinaryTreeKeyCodeChunk()`: first code chunks (offsets 128–255) share the same page |
| `pkg/trie/bintrie/bintrie.go` | 53–58 | `GetBinaryTreeKeyBasicData()`: nonce/balance/code_size at leaf index 0 in same page |
| `pkg/trie/bintrie/bintrie.go` | 94–107 | `StorageIndex()`: main storage (slot ≥ 64) mapped to separate stems |
| `pkg/trie/bintrie/stem_node.go` | 14 | `StemNode`: groups 256 contiguous leaf values under one 31-byte stem |
| `pkg/trie/bintrie/internal_node.go` | 117 | `InsertValuesAtStem()`: inserts all 256 values of a stem in one operation |

**Page layout for a typical DApp contract:**
- Leaf 0: `BasicData` (nonce 8B + balance 16B + code size 3B)
- Leaf 1: `CodeHash` (32 bytes)
- Leaves 64–127: Storage slots 0–63 (first 2 kB of storage, at 32 bytes each)
- Leaves 128–255: First 128 code chunks (first 4 kB of code, at 32 bytes each)

All of these share one `StemNode` (one tree branch walk). Reading all of them requires loading only the single branch path once, not 256 separate branch paths.

---

### A.5 "Reduced variance in access depth (loads from big contracts vs small contracts). Binary trees are simpler. Opportunity to add any metadata bits we end up needing for state expiry."

**Status: COMPLETE**

The binary trie has uniform branch depth (log₂ N for N accounts), no extension nodes, and no nested RLP encoding — significantly simpler than MPT.

State expiry is implemented:

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/core/state/state_expiry.go` | 10–240 | `StateExpiryManager`: epoch-based access tracking, `ExpireStaleState()`, `ReviveAccount()` with witness proof |
| `pkg/core/state/state_expiry.go` | 13–28 | `StateExpiryConfig`: `ExpiryPeriod`, `MaxWitnessSize`, `RevivalGasCost` |
| `pkg/core/state/expiry_engine.go` | 27–165 | `ExpiryEngine`: per-account snapshots at expiry time, state root + epoch recorded |
| `pkg/core/state/expiry.go` | 10–50 | `ExpiryConfig`, `ExpiryRecord` with `StateRoot` for revival proof verification |
| `pkg/core/state/misc_purges.go` | 184–200 | `PurgeExpiredStorage()`: removes stale storage below `cutoffBlock` |

**State expiry metadata in binary tree:** The EIP-7864 StemNode has 256 leaf slots — slots can be reserved for expiry metadata (e.g., `lastAccessEpoch` at a well-known leaf index within the StemNode). This is not yet wired between `StateExpiryManager` and the `bintrie` StemNode layout, but the hooks are in place.

**TODO:** Wire `StateExpiryManager.TouchAccount()` / `TouchStorage()` callbacks into the binary trie read path so that access epoch is updated when a StemNode is read.

---

### A.6 "Binary trees are an 'omnibus' that allows us to take all of our learnings from the past ten years about what makes a good state tree, and actually apply them."

**Status: COMPLETE — MPT→binary migration pipeline**

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/trie/migration.go` | 29–66 | `BatchConverter`: configurable batch size for MPT→binary key conversion |
| `pkg/trie/migration.go` | 80–151 | `AddressSpaceSplitter`: splits 256-bit space into N ranges for parallel migration |
| `pkg/trie/migration.go` | 153–193 | `StateProofGenerator`: generates and caches MPT proofs during migration |
| `pkg/trie/migration.go` | 194–206 | `MigrationCheckpoint`: (keys_migrated, last_key_hash, source_root, dest_root, batch_number) |
| `pkg/trie/migration.go` | 244–314 | `GasAccountant`: per-read (200), per-write (5000), per-proof (3000) gas with budget enforcement |
| `pkg/trie/migration.go` | 316–444 | `MPTToBinaryTrieMigrator`: full pipeline with atomic batches, checkpointing |

---

### Section B: NTT / Vectorized Math Precompile

---

### B.1 "A vectorized math precompile (basically, do 32-bit or potentially 64-bit operations on lists of numbers at the same time; in principle this could accelerate many hashes, STARK validation, FHE, lattice-based quantum-resistance signatures, and more by 8-64x); think 'the GPU for the EVM'."

**Status: COMPLETE** — EIP-7885 NTT precompile over BN254 and Goldilocks fields

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/core/vm/precompile_ntt.go` | 1–60 | EIP-7885 NTT precompile at address `0x15` (I+ fork) |
| `pkg/core/vm/precompile_ntt.go` | 27–45 | BN254 scalar field: `p = 21888...617`; Goldilocks: `p = 2^64 − 2^32 + 1` |
| `pkg/core/vm/precompile_ntt.go` | 47–52 | Op types: 0=ForwardBN254, 1=InverseBN254, 2=ForwardGoldilocks, 3=InverseGoldilocks |
| `pkg/core/vm/precompile_ntt.go` | 54–59 | Gas: 1000 base + 10 × log₂(n) per element; max degree 2^16 = 65536 |
| `pkg/core/vm/precompile_ntt.go` | 19–21 | Use cases: "Lattice-based crypto (Falcon, Dilithium), STARK verification, ZK polynomial operations" |

**Accelerated use cases:**
- **Falcon-512 signatures**: NTT is the core operation of Falcon's polynomial arithmetic; on-chain Falcon verification becomes viable
- **STARK verification**: Goldilocks field NTT matches the field used in most STARK provers (e.g., Plonky2, SP1)
- **ZK polynomial ops**: BN254 NTT accelerates KZG commitment polynomial evaluation
- **FHE**: Ring-LWE and other lattice cryptosystems use NTT over various moduli

The NII precompiles (`pkg/core/vm/` — modexp/field-mul/field-inv/batch-verify at 4 addresses) complement the NTT by providing field arithmetic without requiring full program execution.

---

### Section C: VM Changes (EVM → RISC-V)

---

### C.1 "One reason why the protocol gets uglier over time with more special cases is that people have a certain latent fear of 'using the EVM'. If the EVM is not good enough to actually meet the needs of that generality, then we should tackle the problem head-on, and make a better VM."

**Status: COMPLETE** — RISC-V zkVM framework fully implemented

#### RISC-V CPU Emulator (RV32IM)

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/zkvm/riscv_cpu.go` | — | RV32IM interpreter: all base integer instructions + M extension (multiply/divide) |
| `pkg/zkvm/riscv_memory.go` | 1–120 | Sparse page-based memory: 4KB pages, 32-bit address space, MMIO at `0xF0000000+` |
| `pkg/zkvm/riscv_memory.go` | 16–26 | `RVPageSize=4096`, max 16384 pages = 64 MiB |
| `pkg/zkvm/riscv_encode.go` | — | RISC-V instruction encoding |
| `pkg/zkvm/riscv_witness.go` | — | Witness collection during RISC-V execution |

#### Guest Program Framework

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/zkvm/canonical.go` | 67–295 | `RiscVGuest`, `GuestExecution`, `GuestRegistry`, `CanonicalGuestPrecompile` |
| `pkg/zkvm/canonical.go` | 41–48 | Config: 16M cycle limit, 256 MiB memory, "stark" proof system |
| `pkg/zkvm/canonical.go` | 214–267 | `GuestRegistry.RegisterGuest()`: registers RISC-V program by hash; `GetGuest()` for retrieval |
| `pkg/zkvm/canonical_executor.go` | 101–260 | `CanonicalExecutor`: runs registered guest, collects witness, generates ZK proof |
| `pkg/zkvm/guest.go` | — | Guest program abstractions |

---

### C.2 "More efficient than EVM in raw execution, to the point where most precompiles become unnecessary."

**Status: PARTIAL** — framework exists; "most precompiles as RISC-V" not yet done

The RISC-V interpreter is more efficient than the EVM for compute-heavy workloads. Today's EVM precompiles (BLS, KZG, PQC, etc.) were added because EVM interpretation is too slow. With a native RISC-V execution path, these operations can be implemented as RISC-V guest programs instead.

**What exists:**

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/zkvm/zkisa_bridge.go` | 36–100 | `ZKISABridge`: maps 9 EVM operation selectors to RISC-V guest programs |
| `pkg/zkvm/zkisa_bridge.go` | 36–46 | Op 0x01=Keccak256, 0x02=SHA256, 0x03=ECDSARecovery, 0x04=ModExp, 0x05-0x07=BN256, 0x08=BLS12-381, 0xFF=Custom |
| `pkg/zkvm/zkisa_bridge.go` | 202–245 | `NewZKISABridge()`: auto-registers built-in guest programs for each op |
| `pkg/zkvm/stf_executor.go` | 82–180 | `RealSTFExecutor`: runs the entire Ethereum state transition as a RISC-V guest |

**What's PARTIAL / TODO:**
- The 9 operation selectors are wired, but the individual precompile EVM contracts are NOT auto-converted to RISC-V programs — a developer must manually write the RISC-V guest binary
- No EIP exists yet for "precompile X is now a RISC-V program at address Y" automatic mapping
- The NTT precompile (0x15) and BLS precompiles currently run as Go functions, not as RISC-V guests

**TODO (Step 1 of deployment roadmap):**
- Create `pkg/zkvm/guests/keccak256.s` — RISC-V assembly for Keccak-256 (replaces precompile 0x01)
- Create `pkg/zkvm/guests/sha256.s` — SHA-256 (replaces precompile 0x02)
- Wire `pkg/core/vm/precompiles.go` to route to RISC-V guest when available
- Add fork activation flag: `IsPrecompileRISCV()` → use RISC-V path

---

### C.3 "More prover-efficient than EVM (today, provers are written in RISC-V, hence my proposal to just make the new VM be RISC-V)."

**Status: COMPLETE** — STF in RISC-V with STARK proofs

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/zkvm/stf_executor.go` | 55–180 | `RealSTFOutput`: valid flag, post_root, gas_used, cycle_count, proof_data, verification_key, trace_commitment, public_inputs_hash |
| `pkg/zkvm/proof_backend.go` | — | Proof backend: STARK proof generation from RISC-V witness trace |
| `pkg/zkvm/riscv_witness.go` | — | Witness collection: captures register + memory state at each instruction |
| `pkg/zkvm/circuit_builder.go` | — | Circuit construction from RISC-V execution trace |
| `pkg/zkvm/verifier.go` | — | Proof verification |

The state transition function runs on RISC-V, and the witness is fed into a STARK circuit builder, producing a proof of correct execution. This is the key primitive for mandatory proof-carrying blocks (K+ roadmap).

---

### C.4 "Client-side-prover friendly. You should be able to, client-side, make ZK-proofs about e.g. what happens if your account gets called with a certain piece of data."

**Status: COMPLETE** — VOPS + light client proofs + stateless execution

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/core/vops/` | — | Partial state validator: prove state transition from partial witness |
| `pkg/light/cl_proofs.go` | 58–100 | `CLProofGenerator`: proves slot, state root, validator balance, attestation |
| `pkg/light/proof_generator_test.go` | — | `GenerateStateRootProof()`: 40-element Merkle paths |
| `pkg/proofs/aa_proof_circuits.go` | — | Groth16 circuit: proves AA validation passed (nonce, sig, gas) |
| `pkg/witness/state_proof.go` | — | Binary Merkle proof generation for execution witnesses |

A user can generate a proof that "if address X is called with calldata Y, the resulting state change is Z" using the stateless execution path (`PartialState` + VOPS) without syncing the full chain.

---

### C.5 "Maximum simplicity. A RISC-V interpreter is only a couple hundred lines of code, it's what a blockchain VM 'should feel like'."

**Status: COMPLETE** — RV32IM interpreter is minimal by design

The RISC-V CPU in `pkg/zkvm/riscv_cpu.go` implements only the RV32IM instruction set (47 instructions). Compare to the EVM which has 164+ opcodes, complex gas tables, and context-dependent semantics.

---

### C.6 Deployment Roadmap — Three Steps

**Step 1: RISC-V only for precompiles (80% of today's precompiles become RISC-V blobs)**

**Status: PARTIAL**

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/zkvm/zkisa_bridge.go` | 36–100 | Maps EVM precompile semantics to RISC-V operation selectors |
| `pkg/zkvm/zkisa_bridge.go` | 202–245 | `NewZKISABridge()`: registers built-in guest stubs for each op |

The framework exists but no existing EVM precompile has been implemented as a RISC-V binary yet.

**TODO (Step 1):**
- Implement `pkg/zkvm/guests/`: RISC-V binaries for Keccak256, SHA256, ECRECOVER, ModExp, BN256Add/Mul/Pair
- Wire precompile router to use RISC-V path when `IsI+()` fork active
- Benchmark: RISC-V guest gas vs current Go function gas

---

**Step 2: Users can deploy RISC-V contracts**

**Status: PARTIAL** — call mechanism exists; no user-facing contract deployment

| File | Lines | What it does |
|------|-------|-------------|
| `pkg/zkvm/canonical.go` | 295–342 | `CanonicalGuestPrecompile` at `0x0200`: accepts any registered RISC-V program |
| `pkg/zkvm/canonical.go` | 228–250 | `RegisterGuest(program []byte)`: hashes and stores a RISC-V program |

Currently users can call `0x0200` with a pre-registered program hash + input. But there is no user-facing flow for:
- Deploying a RISC-V program to a contract address (like `CREATE` for EVM contracts)
- Calling a RISC-V contract address directly without going through the `0x0200` precompile

**TODO (Step 2):**
- Add `RVCREATE` opcode or `CREATE3` variant that deploys RISC-V bytecode to an address
- Route `CALL` to RISC-V executor when code starts with `0xFE 0x52 0x56` (RISC-V magic bytes)
- Gas model: RISC-V cycles → EVM gas conversion (currently 1 cycle = 1 gas at `pkg/zkvm/canonical.go:48`)

---

**Step 3: EVM is retired and turns into a smart contract written in RISC-V**

**Status: MISSING / Long-term**

This step is explicitly described by Vitalik as "longer-term and still more non-consensus." No implementation exists, and none is expected in the near term.

**Design note:** The existing EVM interpreter in `pkg/core/vm/evm.go` would need to be compiled to a RISC-V binary and deployed as a smart contract at a well-known address. All `CALL` instructions to EVM-format contracts would route through this RISC-V EVM interpreter.

**TODO (Step 3 — tracking only):**
- Research: compile the EVM interpreter to RISC-V (go → RISC-V via `GOARCH=riscv64`)
- Performance estimate: how many RISC-V cycles does EVM bytecode interpretation take?
- Backwards compatibility: EVM users experience no visible change except gas cost shifts

---

## Complete Items Reference Table

| Item | File(s) | Key types/functions |
|------|---------|---------------------|
| EIP-7864 binary trie | `pkg/trie/bintrie/bintrie.go` | `GetBinaryTreeKey`, `ChunkifyCode`, `GetBinaryTreeKeyStorageSlot` |
| StemNode (256-leaf page) | `pkg/trie/bintrie/stem_node.go` | `StemNode`, `StemNodeWidth=256` |
| SHA-256 tree hasher | `pkg/trie/bintrie/hasher_extended.go` | `BinaryHasher`, `hashInternal`, parallel threshold |
| Binary Merkle proofs | `pkg/trie/bintrie/proof.go` | `GenerateProof` |
| Proof verification | `pkg/trie/bintrie/proof_verifier.go` | `VerifyInclusion`, `VerifyExclusion` |
| MPT→binary migration | `pkg/trie/migration.go` | `MPTToBinaryTrieMigrator`, `MigrationCheckpoint`, `GasAccountant` |
| State expiry | `pkg/core/state/state_expiry.go` | `StateExpiryManager`, `ExpireStaleState`, `ReviveAccount` |
| Expiry engine | `pkg/core/state/expiry_engine.go` | `ExpiryEngine`, `ExpireAccountWithRoot` |
| Storage purge | `pkg/core/state/misc_purges.go` | `PurgeExpiredStorage` |
| Poseidon hash | `pkg/zkvm/poseidon.go` | `PoseidonHash`, `PoseidonSponge`, Grain LFSR |
| Poseidon2 | `pkg/zkvm/poseidon2.go` | Poseidon2 variant |
| NTT precompile (EIP-7885) | `pkg/core/vm/precompile_ntt.go` | Address 0x15, BN254 + Goldilocks fields |
| RISC-V CPU | `pkg/zkvm/riscv_cpu.go` | RV32IM interpreter |
| RISC-V memory | `pkg/zkvm/riscv_memory.go` | 4KB pages, 64 MiB max, MMIO |
| Canonical guest framework | `pkg/zkvm/canonical.go` | `GuestRegistry`, `CanonicalGuestPrecompile` at 0x0200 |
| Canonical executor | `pkg/zkvm/canonical_executor.go` | `CanonicalExecutor`, witness + proof |
| STF executor | `pkg/zkvm/stf_executor.go` | `RealSTFExecutor`, STARK proof of ETH state transition |
| zkISA bridge | `pkg/zkvm/zkisa_bridge.go` | `ZKISABridge`, 9 op selectors |
| ZK circuit builder | `pkg/zkvm/circuit_builder.go` | Circuit from RISC-V trace |
| RISC-V witness | `pkg/zkvm/riscv_witness.go` | Witness collection per instruction |
| Proof backend | `pkg/zkvm/proof_backend.go` | STARK proof generation |

---

## TODO List

| Priority | Item | File to create/modify | Notes |
|----------|------|----------------------|-------|
| P0 | Blake3 hash for binary tree | `pkg/trie/bintrie/hasher_blake3.go` | ~3x proving speedup vs Keccak; compare to SHA-256 baseline |
| P0 | RISC-V precompile guests (Step 1) | `pkg/zkvm/guests/` | Keccak256, SHA256, ECRECOVER as RISC-V binaries |
| P0 | Wire precompile router to RISC-V path | `pkg/core/vm/precompiles.go` | Fork flag `IsPrecompileRISCV()` |
| P1 | Poseidon binary tree option | `pkg/trie/bintrie/hasher_poseidon.go` | After Poseidon security review; 100x ZK proving |
| P1 | State expiry wired to bintrie access | `pkg/core/state/state_expiry.go` + bintrie | Touch epoch updated on StemNode read |
| P1 | RVCREATE opcode for user RISC-V contracts (Step 2) | `pkg/core/vm/evm.go` | `CREATE`-style for RISC-V code deployment |
| P2 | CALL routing to RISC-V executor | `pkg/core/vm/evm.go` | RISC-V magic bytes detection |
| P3 | EVM interpreter compiled to RISC-V (Step 3) | Research | `GOARCH=riscv64` compile experiment |

---

## Spec References (from `refs/`)

### EIP-7864 — Binary Tree Exact Spec

Source: `refs/EIPs/EIPS/eip-7864.md`

**Key structure**: `32 bytes = stem (31 bytes) || subindex (1 byte)`

**Node types and hash rules**:
```python
# Internal node
internal_node_hash = hash(left_hash || right_hash)

# Stem node
stem_node_hash = hash(stem || 0x00 || hash(left_hash || right_hash))

# Leaf node
leaf_node_hash = hash(value)    # value is 32 bytes

# Empty node (subtree optimization)
empty_node_hash = bytes32(0)
hash([0x00] * 64) = bytes32(0)  # empty 64-byte input → zero
```

**Hash function**: BLAKE3 (draft/experimental). Final choice is TBD — candidates are Keccak and Poseidon2. ETH2030 currently uses SHA-256 (`pkg/trie/bintrie/hasher_extended.go`), which is not the spec value.

**Proof size vs MPT**: binary tree needs log₂(N) ≈ 36 levels for N=2³² accounts (31 bytes × 8 bits = 248 levels to stem + 8 bits within StemNode). MPT needs ≈ 120 nodes (15 siblings × log₁₆(2³²) = 15 × 8). **4x reduction matches Vitalik's claim.**

**ETH2030 gap**: Hash function is SHA-256 (not BLAKE3). To align with the spec draft:
```go
// pkg/trie/bintrie/hasher_blake3.go  (to create)
import blake3 "lukechampine.com/blake3"

func (h *BinaryHasher) hashInternalBLAKE3(left, right []byte) []byte {
    var input [64]byte
    copy(input[:32], left)
    copy(input[32:], right)
    sum := blake3.Sum256(input[:])
    return sum[:]
}
```

**StemNode page layout** (from EIP-7864, confirmed in `pkg/trie/bintrie/bintrie.go`):
```
Subindex 0:    BasicData (nonce 8B + balance 16B + code_size 3B)
Subindex 1:    CodeHash (32 bytes)
Subindex 64–127:  Storage slots 0–63  (first 2 KB of storage)
Subindex 128–255: Code chunks 0–127   (first 4 KB of code)
Subindex 2–63: reserved / expiry metadata
```

---

### NTT Precompile — Exact Addresses

Source: `refs/ntt-eip/EIP/EIPNTT.md`

| Precompile | Spec Address | ETH2030 Address | Gas |
|---|---|---|---|
| `NTT_FW` | `0x0f` | `0x15` (combined) | 600 flat |
| `NTT_INV` | `0x10` | `0x15` (combined) | 600 flat |
| `NTT_VECMULMOD` | `0x11` | not separate | `k * log₂(n) / 8` |
| `NTT_VECADDMOD` | `0x12` | not separate | `k * log₂(n) / 32` |

**ETH2030 discrepancy**: Our NTT is at address `0x15` with op-type byte dispatch. The spec defines 4 separate precompile addresses. This is a fork-choice-breaking difference that needs resolving before Hegotá testnet.

**Supported fields** (from spec, relevant to ZK proving):
- **M31** (`q = 2³¹−1`): Circle STARKs (StarkWare), from `refs/research/circlestark/`
- **Goldilocks** (`q = 2⁶⁴−2³²+1`): Plonky2, SP1
- **BN254 scalar** (`q = 21888...617`): gnark Groth16, KZG

---

### Groth16 Proving — gnark API

Source: `refs/gnark/backend/groth16/`

**To upgrade `pkg/proofs/groth16_verifier.go` from placeholder to real gnark proving**:

```go
import (
    "github.com/consensys/gnark/backend/groth16"
    "github.com/consensys/gnark/frontend"
    bn254 "github.com/consensys/gnark-crypto/ecc/bn254"
)

// Prove a circuit
func Prove(r1cs *cs.R1CS, pk *ProvingKey, fullWitness witness.Witness,
    opts ...backend.ProverOption) (*Proof, error)

// Verify a proof
func Verify(proof Proof, vk VerifyingKey, publicWitness witness.Witness,
    opts ...backend.VerifierOption) error

// Proof structure (BLS12-381)
type Proof struct {
    Ar, Krs      curve.G1Affine   // Main proof elements
    Bs           curve.G2Affine
    Commitments  []curve.G1Affine // Pedersen commitments (BSB22)
    CommitmentPok curve.G1Affine  // Batched PoK
}
```

Supported curves: BN254, BLS12-381, BLS12-377, BW6-761.

**ETH2030 next step**: Replace `pkg/proofs/groth16_verifier.go` placeholder with `gnark/backend/groth16.Verify()` call. The `AAValidationCircuit` in `pkg/proofs/aa_proof_circuits.go` needs to compile to R1CS via `frontend.Compile()` first.

---

### KZG — go-eth-kzg API

Source: `refs/go-eth-kzg/api.go`, `prove.go`, `verify.go`

**To upgrade `PlaceholderKZGBackend` (test SRS s=42) to production trusted setup**:

```go
import "github.com/crate-crypto/go-eth-kzg"

// Production context (uses official Ethereum trusted setup JSON)
ctx, err := goethkzg.NewContext4096Secure()

// Blob commitment
commitment, err := ctx.BlobToKZGCommitment(blob, numGoRoutines)

// Single-point proof (EIP-4844)
proof, err := ctx.ComputeBlobKZGProof(blob, commitment, numGoRoutines)

// Verify single blob
err = ctx.VerifyBlobKZGProof(blob, commitment, proof)

// Batch verify (parallel goroutines)
err = ctx.VerifyBlobKZGProofBatchPar(blobs, commitments, proofs)

// EIP-7594 cell proofs
err = ctx.VerifyCellKZGProofBatch(commitments, cellIndices, cells, proofs)
```

**Types**:
```go
type Blob [4096 * 32]byte   // 131,072 bytes
type KZGCommitment [48]byte // G1 point
type KZGProof [48]byte      // G1 quotient commitment
type Cell [64 * 32]byte     // PeerDAS cell (2,048 bytes)
```

---

### Circle STARK — Poseidon (M31 field)

Source: `refs/research/circlestark/poseidon.py`

This is the recursive STARK variant relevant to `GAP P3-A` (recursive STARK composition in `vitalik-pq-roadmap-gap-analysis.md`).

**Poseidon parameters for Circle STARKs**:
- Field: `M31 = 2³¹ − 1` (fast modular arithmetic, fits in 32-bit)
- State width: 16 elements
- Rounds: 64 total (4 full + 56 partial + 4 full)
- Full rounds: SBox on all 16 elements + full MDS
- Partial rounds: SBox on `state[0]` only + inner-diagonal MDS (much cheaper)
- Output: `state[8:16] + input[8:16]` (8-element hash output)
- STARK trace: 192 columns for 32-leaf Merkle branch

**FRI parameters** (from `refs/research/circlestark/fast_fri.py`):
- Base case: 64 evaluations
- Folding: 3 folds per round (8× degree reduction per step)
- Security: 80 random queries
- Proof size: O(log²(degree)) via Merkle trees

**ETH2030 current Poseidon** (`pkg/zkvm/poseidon.go`) uses BN254 scalar field with t=3 (rate=2, capacity=1). The Circle STARK uses M31 field with t=16 — these are different variants. For recursive STARK DA proofs, the M31 Poseidon from `refs/research/circlestark/` is the reference to follow.
