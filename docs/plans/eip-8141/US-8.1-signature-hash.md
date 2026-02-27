# US-8.1 — Canonical Signature Hash with VERIFY Frame Elision

**Epic:** EP-8 Signature Hash Computation
**Total Story Points:** 3
**Sprint:** 1

> **As a** smart account developer,
> **I want** a canonical signature hash that elides VERIFY frame data,
> **so that** accounts can sign a hash that is stable across gas sponsor changes and cannot be malleated by modifying VERIFY frame data.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

---

## Tasks

### Task 8.1.1 — ComputeSigHash Implementation

| Field | Detail |
|-------|--------|
| **Description** | Implement `ComputeSigHash(tx *FrameTx) common.Hash`: (1) deep-copy `tx`; (2) for each frame where `mode == VERIFY`, set `frame.data = []byte{}`; (3) RLP-encode the modified transaction; (4) return `keccak256(rlp_bytes)`. Must not mutate the original transaction. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) VERIFY frames → hash differs from full-data hash; (2) no VERIFY frames → hash equals standard RLP hash; (3) different VERIFY data → same sig hash; (4) `frame.target` of VERIFY NOT elided — changing it changes hash; (5) no mutation of original tx. |
| **Definition of Done** | Tests pass; immutability verified; reviewed. |

### Task 8.1.2 — Sig Hash Pre-Computation at Tx Entry

| Field | Detail |
|-------|--------|
| **Description** | Pre-compute `ComputeSigHash` once at tx entry and store in `FrameContext.SigHash`. Avoids recomputation on every `TXPARAMLOAD(0x08)`. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | Assert `FrameContext.SigHash` is non-zero after pre-processing; `TXPARAMLOAD(0x08)` returns same value as independent `ComputeSigHash`. |
| **Definition of Done** | Pre-computation verified; no recomputation in opcode handler; reviewed. |
