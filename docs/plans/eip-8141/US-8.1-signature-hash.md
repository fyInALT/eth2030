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

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/core/types/tx_frame.go:175-220` | `ComputeFrameSigHash` — elides VERIFY data, RLP encodes, keccak256 |
| `pkg/core/vm/eip8141_opcodes.go:51` | `FrameContext.SigHash` field (pre-computed at tx entry) |

## Implementation Status

**✅ Implemented**

- ✅ `ComputeFrameSigHash` implemented correctly
- ✅ VERIFY frame data elided (set to empty bytes)
- ✅ Uses `keccak256(0x06 || rlp(modified_tx))`
- ✅ Does not mutate original transaction (builds copy)

---

## EIP-8141 Reference Excerpts

### Specification → Signature Hash

> With the frame transaction, the signature may be at an arbitrary location in the frame list. In the canonical signature hash any frame with mode `VERIFY` will have its data elided:
>
> ```python
> def compute_sig_hash(tx: FrameTx) -> Hash:
>     for i, frame in enumerate(tx.frames):
>         if frame.mode == VERIFY:
>             tx.frames[i].data = Bytes()
>     return keccak(rlp(tx))
> ```

### Rationale → Canonical signature hash

> The canonical signature hash is provided in `TXPARAMLOAD` to simplify the development of smart accounts.
>
> Computing the signature hash in EVM is complicated and expensive. While using the canonical signature hash is not mandatory, it is strongly recommended. Creating a bespoke signature requires precise commitment to the underlying transaction data. Without this, it's possible that some elements can be manipulated in-the-air while the transaction is pending and have unexpected effects. This is known as transaction malleability. Using the canonical signature hash avoids malleability of the frames other than `VERIFY`.
>
> The `frame.data` of `VERIFY` frames is elided from the signature hash. This is done for two reasons:
>
> 1. It contains the signature so by definition it cannot be part of the signature hash.
> 2. In the future it may be desired to aggregate the cryptographic operations for data and compute efficiency reasons. If the data was introspectable, it would not be possible to aggregate the verify frames in the future.
> 3. For gas sponsoring workflows, we also recommend using a `VERIFY` frame to approve the gas payment. Here, the input data to the sponsor is intentionally left malleable so it can be added onto the transaction after the `sender` has made its signature. Notably, the `frame.target` of `VERIFY` frames is covered by the signature hash, i.e. the `sender` chooses the sponsor address explicitly.
