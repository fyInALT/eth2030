# US-4.2 — Poseidon2 Hash Backend

**Epic:** EP-4 Finality Protocol
**Total Story Points:** 8
**Sprint:** 4

> **As a** ZK infrastructure developer,
> **I want** a Poseidon2 hash backend implementation alongside the existing Poseidon1,
> **so that** ZK circuit proofs (Groth16, PLONK, STARKs) can use the more efficient Poseidon2 permutation, reducing proving time and circuit complexity.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

---

## Vitalik's Proposal

> Poseidon2 offers better performance than Poseidon1 for ZK circuits due to: (a) external/internal round separation with simpler MDS in internal rounds, (b) fewer constraints per hash in Groth16/PLONK, (c) better resistance analysis. For L1 finality proofs and CL proof circuits, Poseidon2 reduces proving time. ETH2030 already has Poseidon1 for ZK — adding Poseidon2 as an option improves performance. Alternatives: increased Poseidon1 rounds, or BLAKE3/SHA-256 for non-ZK paths.

---

## Tasks

### Task 4.2.1 — Poseidon2 Permutation

| Field | Detail |
|-------|--------|
| **Description** | Implement the Poseidon2 permutation function. Key difference from Poseidon1: internal rounds use a diagonal MDS matrix (`M_I = I + D` where D is a diagonal matrix) instead of the full Cauchy MDS used in external rounds. This reduces multiplication cost in internal rounds. S-box remains `x^5` over BN254. Parameters: t=3, 4 external rounds (2 before + 2 after), 56 internal rounds. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | ZK Cryptography Engineer |
| **Testing Method** | (1) Known test vectors for Poseidon2 over BN254. (2) Permutation is deterministic. (3) Different inputs produce different outputs. (4) Round count matches specification. (5) Internal rounds use diagonal MDS. |
| **Definition of Done** | Tests pass; known test vectors match; reviewed. |

### Task 4.2.2 — Poseidon2 Sponge and Hash

| Field | Detail |
|-------|--------|
| **Description** | Wrap the Poseidon2 permutation in a sponge construction (identical to Poseidon1 sponge in `poseidon.go:187-251`). Expose `Poseidon2Hash(params, inputs...)` function and `Poseidon2Sponge` type with `Absorb()`/`Squeeze()`. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | ZK Cryptography Engineer |
| **Testing Method** | (1) Single input → deterministic hash. (2) Multiple inputs → deterministic hash. (3) Absorb/Squeeze streaming matches batch hash. (4) Different inputs produce different hashes. |
| **Definition of Done** | Tests pass; sponge works; reviewed. |

### Task 4.2.3 — HashBackend Integration

| Field | Detail |
|-------|--------|
| **Description** | Add `Poseidon2Backend` implementing the `HashBackend` interface from `crypto/pqc/hash_backend.go`. Register it in `HashBackendByName("poseidon2")`. This allows any component using the pluggable hash interface to opt into Poseidon2. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | ZK Cryptography Engineer |
| **Testing Method** | (1) `HashBackendByName("poseidon2")` returns non-nil. (2) Backend produces 32-byte hashes. (3) Backend name is "poseidon2". (4) Block size matches BN254 field element size. |
| **Definition of Done** | Tests pass; backend registered; reviewed. |

### Task 4.2.4 — Circuit Builder Integration

| Field | Detail |
|-------|--------|
| **Description** | Update `zkvm/circuit_builder.go` to support Poseidon2 as a hash option for ZK circuits. Add `CircuitHashMode` enum (`{Poseidon1, Poseidon2, SHA256}`) and wire Poseidon2 permutation into circuit constraint generation. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | ZK Cryptography Engineer |
| **Testing Method** | (1) Circuit with Poseidon2 hash compiles. (2) Proof generated with Poseidon2 verifies. (3) Poseidon2 circuit has fewer constraints than Poseidon1 circuit for same input. (4) Default remains Poseidon1 for backward compatibility. |
| **Definition of Done** | Tests pass; Poseidon2 reduces constraint count; reviewed. |

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/zkvm/poseidon.go:13-67` | `PoseidonParams` — Poseidon1 parameters. Poseidon2 uses similar struct but with additional `InternalMDS` field for diagonal matrix in internal rounds. |
| `pkg/zkvm/poseidon.go:42-67` | `DefaultPoseidonParams()` — t=3, fullRounds=8, partialRounds=57, BN254 field. Poseidon2: t=3, externalRounds=4, internalRounds=56. |
| `pkg/zkvm/poseidon.go:69-78` | `SBox(x)` — `x^5 mod field`. Same S-box for Poseidon2. |
| `pkg/zkvm/poseidon.go:80-94` | `MDSMul(state, mds)` — full matrix multiplication. Poseidon2 uses this for external rounds only. Internal rounds use `DiagMDSMul(state, diag)` which is cheaper. |
| `pkg/zkvm/poseidon.go:96-150` | `poseidonPermutation()` — full/partial round structure. Poseidon2 replaces this with external/internal round structure. |
| `pkg/zkvm/poseidon.go:152-185` | `PoseidonHash()` — sponge construction. Reusable for Poseidon2 (same sponge, different permutation). |
| `pkg/zkvm/poseidon.go:187-251` | `PoseidonSponge` — streaming sponge. Reusable for Poseidon2. |
| `pkg/zkvm/poseidon.go:254-292` | `generateRoundConstants()` and `generateMDS()` — parameter generation. Poseidon2 needs different constants and additional diagonal matrix generation. |
| `pkg/crypto/pqc/hash_backend.go:9-19` | `HashBackend` interface — `Hash()`, `Name()`, `BlockSize()`. Poseidon2 must implement this. |
| `pkg/crypto/pqc/hash_backend.go:67-82` | `HashBackendByName()` — lookup by name. Must add "poseidon2" case. |
| `pkg/zkvm/circuit_builder.go` | Circuit builder — uses Poseidon hash for ZK circuit constraints. Must support Poseidon2 as alternative. |

---

## Implementation Status

**⚠️ Partial (Poseidon1 exists, Poseidon2 does not)**

### What Exists
- ✅ Full Poseidon1 implementation (`zkvm/poseidon.go`, 293 lines) — BN254 field, sponge construction, S-box, MDS
- ✅ `HashBackend` interface (`crypto/pqc/hash_backend.go`) — pluggable hash with Keccak256, SHA256, BLAKE3 (stub)
- ✅ `HashBackendByName()` lookup — supports "keccak256", "sha256", "blake3"
- ✅ `PoseidonSponge` streaming interface — Absorb/Squeeze pattern
- ✅ Circuit builder integration point (`zkvm/circuit_builder.go`)

### What's Missing
- ❌ Poseidon2 permutation — no external/internal round separation
- ❌ Diagonal MDS matrix for internal rounds (`DiagMDSMul`)
- ❌ Poseidon2-specific round constants and parameters
- ❌ `Poseidon2Backend` implementing `HashBackend`
- ❌ "poseidon2" not registered in `HashBackendByName()`
- ❌ Circuit builder does not distinguish Poseidon1 vs Poseidon2

### Proposed Solution

1. Create `pkg/zkvm/poseidon2.go` with:
   - `Poseidon2Params` (extends `PoseidonParams` with `InternalMDS` diagonal matrix)
   - `poseidon2Permutation()` — external rounds use full MDS, internal rounds use `I + D` diagonal
   - `Poseidon2Hash()` and `Poseidon2Sponge` (same sponge, different permutation)
2. Add `Poseidon2Backend` to `hash_backend.go`
3. Register in `HashBackendByName("poseidon2")`
4. Add `CircuitHashMode` to circuit builder

### Key Differences: Poseidon1 vs Poseidon2

| Aspect | Poseidon1 | Poseidon2 |
|--------|-----------|-----------|
| Full rounds | 8 (all use full MDS) | 4 external (full MDS) |
| Partial rounds | 57 (S-box on element 0 only) | 56 internal (diagonal MDS) |
| MDS in partial/internal | Full Cauchy matrix | `I + D` (identity + diagonal) |
| S-box | `x^5` | `x^5` (same) |
| Constraint cost | Higher (full MDS in partial rounds) | Lower (diagonal = cheaper) |
| Field | BN254 | BN254 (same) |

---

## Spec Reference

> **Vitalik:**
> For ZK-friendly hashing in finality proofs and CL proof circuits, we can use Poseidon2 instead of Poseidon1. Poseidon2 reduces the number of multiplications per hash by ~30% due to the cheaper internal round structure. Alternatives include: (a) using more rounds with Poseidon1 for extra security margin, (b) using BLAKE3 for non-ZK paths, (c) using SHA-256 where ZK-friendliness is not required.
>
> **Security note:** Poseidon2 has undergone extensive cryptanalysis. The external/internal round design provides better security margins per constraint than Poseidon1's full/partial design. For BN254, the recommended parameters (t=3, R_E=4, R_I=56) provide at least 128-bit security.
