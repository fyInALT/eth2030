# US-4.1 — TXPARAM Opcode Family: Transaction Parameter Introspection

**Epic:** EP-4 TXPARAM* Opcodes
**Total Story Points:** 16
**Sprint:** 3

> **As an** EVM contract developer,
> **I want** `TXPARAMLOAD` (0xb0), `TXPARAMSIZE` (0xb1), and `TXPARAMCOPY` (0xb2) opcodes to expose all 16 transaction parameters,
> **so that** smart account contracts can inspect transaction fields during validation and execution without expensive in-EVM recomputation.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

---

## Tasks

### Task 4.1.1a — TXPARAMLOAD (0xb0): Scalar Parameters (0x00–0x10)

| Field | Detail |
|-------|--------|
| **Description** | Implement the scalar-parameter path of `opTxParamLoad` in `pkg/core/vm/eip8141_opcodes.go`. Stack layout: `in1` (parameter index), `in2` (must be `0` for all scalar indices), `byte_offset`. Returns a 32-byte word at `byte_offset`, zero-padded when offset exceeds 32. Implement the **11 scalar indices**: `0x00` tx type, `0x01` nonce, `0x02` sender, `0x03` max_priority_fee_per_gas, `0x04` max_fee_per_gas, `0x05` max_fee_per_blob_gas, `0x06` max_cost, `0x07` blob hash count, `0x08` sig_hash, `0x09` frame count, `0x10` current frame index. Gap indices `0x0a`–`0x0f` → exceptional halt. Any other undefined `in1` → exceptional halt. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | Table-driven unit tests: (1) correct return value for each of 11 scalar indices; (2) `byte_offset=16` returns trailing zero-padded slice; (3) blob hash count = 0 with no blobs; (4) current frame index returns N; (5) gap range `0x0a`–`0x0f` → halt; (6) undefined `in1` → halt. |
| **Definition of Done** | All 11 scalar indices pass; gap-index halts verified; coverage ≥ 85%; reviewed. |

### Task 4.1.1b — TXPARAMLOAD (0xb0): Frame-Indexed Parameters (0x11–0x15)

| Field | Detail |
|-------|--------|
| **Description** | Implement frame-indexed parameters: `0x11` target, `0x12` data (VERIFY → size 0), `0x13` gas_limit, `0x14` mode, `0x15` status (halt if current/future frame). Validate `in2 < len(frames)`, else exceptional halt. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | (1) frame data returns first 32 bytes; (2) VERIFY frame data → 32 zero bytes; (3) OOB frame index → halt; (4) status of current frame → halt; (5) status of future frame → halt; (6) status of past frame → 0 or 1. |
| **Definition of Done** | All 5 indices tested; boundary conditions verified; reviewed. |

### Task 4.1.2 — TXPARAMSIZE (0xb1): Dynamic Size Query

| Field | Detail |
|-------|--------|
| **Description** | Implement `opTxParamSize`. Fixed-size params → 32. Frame data (`0x12`) → `len(frame[in2].data)`. VERIFY frame data → 0. Same error rules as TXPARAMLOAD. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | (1) Fixed-size param → 32; (2) frame data 100 bytes → 100; (3) VERIFY data → 0; (4) OOB frame index → halt; (5) invalid in1 → halt. |
| **Definition of Done** | Tests pass; VERIFY elision verified; reviewed. |

### Task 4.1.3 — TXPARAMCOPY (0xb2): Dynamic Copy

| Field | Detail |
|-------|--------|
| **Description** | Implement `opTxParamCopy` following CALLDATACOPY pattern. Stack: `[mem_dest, src_offset, length, in1, in2]`. Copies parameter bytes into memory. Standard EVM memory expansion gas cost. VERIFY frame data treated as zero-length. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | (1) Copy frame data → memory matches; (2) offset beyond data → zero-padded; (3) VERIFY data → zero-padded; (4) memory expansion cost charged; (5) OOB frame index → halt. |
| **Definition of Done** | Memory copy correct; zero-padding correct; gas cost verified; reviewed. |

### Task 4.1.4 — TXPARAM Signature Hash (0x08)

| Field | Detail |
|-------|--------|
| **Description** | Ensure `TXPARAMLOAD(0x08, 0)` returns `compute_sig_hash(tx)`. Pre-computed in `FrameContext.SigHash`. Implement `ComputeSigHash(tx *FrameTx) common.Hash`: deep-copy tx, elide VERIFY frame data, RLP encode, keccak256. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) VERIFY frame → hash differs from full-data hash; (2) no VERIFY → equals standard hash; (3) different VERIFY data → same hash; (4) TXPARAMLOAD(0x08) returns same value; (5) frame.target of VERIFY NOT elided. |
| **Definition of Done** | Tests pass; immutability of input verified; reviewed. |

### Task 4.1.5 — TXPARAM Gas Cost

| Field | Detail |
|-------|--------|
| **Description** | Gas costs: TXPARAMLOAD/SIZE = 3 (like CALLDATALOAD/SIZE). TXPARAMCOPY = `3 + ceil(length/32) * 3` + memory expansion. Register in instruction table. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | Unit test: TXPARAMCOPY with known length, assert gas consumed matches formula. Test memory expansion. |
| **Definition of Done** | Gas matches CALLDATACOPY pattern; registered; reviewed. |

### Task 4.1.6 — TXPARAM Scalar Parameter `in2 == 0` Enforcement

| Field | Detail |
|-------|--------|
| **Description** | For all 11 scalar parameter indices (`0x00`–`0x10`), `in2` must be exactly `0`. Non-zero `in2` → exceptional halt. Applies to all three TXPARAM opcodes. Frame-indexed params (`0x11`–`0x15`) accept any valid frame index in `in2`. |
| **Estimated Effort** | 1 story point |
| **Assignee/Role** | EVM Engineer |
| **Testing Method** | Table-driven: (1) `TXPARAMLOAD(index, 0, 0)` → succeeds; (2) `TXPARAMLOAD(index, 1, 0)` → halt; (3) `TXPARAMLOAD(index, 0xff, 0)` → halt; (4) `TXPARAMSIZE(index, 1)` → halt; (5) frame-indexed with `in2 > 0` within bounds → succeeds. |
| **Definition of Done** | All 11 scalar indices tested for `in2 != 0` rejection; reviewed. |
