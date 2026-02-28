# US-3.1 — Block-Level Erasure Coding

**Epic:** EP-3 Block Propagation
**Total Story Points:** 13
**Sprint:** 3

> **As a** node operator,
> **I want** execution blocks to be split into k-of-n erasure-coded pieces for propagation,
> **so that** block propagation latency is reduced by allowing reconstruction from any k pieces rather than waiting for the full block.

**INVEST:** I ✓ | N ✓ | V ✓ | E ✓ | S ✓ | T ✓

---

## Vitalik's Proposal

> For gigagas blocks (>10 MB), split the block into 8 erasure-coded pieces using a k-of-n scheme. Each piece is ~1/k of the block size plus parity overhead. Validators and peers propagate individual pieces, and any node receiving k pieces can reconstruct the full block. This reduces worst-case propagation latency from O(block_size) to O(block_size/k) since pieces propagate in parallel via different peers.

---

## Tasks

### Task 3.1.1 — Block Erasure Encoder

| Field | Detail |
|-------|--------|
| **Description** | Implement `BlockErasureEncoder` that splits a serialized block into `k` data pieces and `m` parity pieces (total `n = k + m`). Default: k=4, m=4, n=8 (can reconstruct from any 4 of 8). Use the existing `RSEncoderGF256` from `das/erasure/reed_solomon_encoder.go` for GF(2^8) MDS encoding. Each piece includes a header: `(block_hash, piece_index, total_pieces, data_pieces, piece_size)`. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Block encoded into 8 pieces. (2) Each piece has correct header. (3) Pieces have uniform size (±1 byte for rounding). (4) Block hash in header matches original. (5) Encoder handles blocks up to 10 MB. |
| **Definition of Done** | Tests pass; encoding produces correct piece count; reviewed. |

### Task 3.1.2 — Block Erasure Decoder

| Field | Detail |
|-------|--------|
| **Description** | Implement `BlockErasureDecoder` that reconstructs the full block from any k pieces (out of n). Validates: (a) all pieces have the same block hash, (b) piece indices are valid, (c) no duplicate indices, (d) reconstructed block matches expected hash. Use `RSEncoderGF256.Decode()` for recovery. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Reconstruction from exactly k=4 data pieces. (2) Reconstruction from k=4 mixed (2 data + 2 parity). (3) Reconstruction from k=4 parity-only (if applicable). (4) Fewer than k pieces → error. (5) Mismatched block hashes → error. (6) Duplicate piece indices → error. (7) Reconstructed block hash matches original. |
| **Definition of Done** | All 7 tests pass; reconstruction verified for all piece combinations; reviewed. |

### Task 3.1.3 — Block Piece Gossip Protocol

| Field | Detail |
|-------|--------|
| **Description** | Add `BlockPiece` gossip topic and routing. When a node produces a block (or receives a full block), it encodes it into 8 pieces and sends different pieces to different peer subsets. Use the existing sqrt(n) fanout from `block_gossip.go` but route each piece to a different peer subset. Add `BlockPieceMessage` type with piece data + header. |
| **Estimated Effort** | 3 story points |
| **Assignee/Role** | P2P Engineer |
| **Testing Method** | (1) Block encoded and pieces distributed to different peers. (2) Each peer receives ~1 piece initially. (3) Peers re-gossip received pieces to their own peers. (4) Block piece validation rejects malformed pieces. (5) Deduplication prevents processing same piece twice. |
| **Definition of Done** | Tests pass; pieces propagate via gossip; reviewed. |

### Task 3.1.4 — Block Assembly Manager

| Field | Detail |
|-------|--------|
| **Description** | Implement `BlockAssemblyManager` that collects incoming block pieces, tracks which pieces have been received per block hash, and triggers reconstruction when k pieces are available. Handles concurrent piece arrival from multiple peers. Cleans up completed assemblies after a timeout. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Pieces arrive out of order → assembly completes when k-th piece arrives. (2) Assembly timeout: incomplete assemblies cleaned up. (3) Concurrent pieces from multiple peers handled. (4) Already-complete block hash → new pieces ignored. (5) Reconstruction failure → assembly marked failed. |
| **Definition of Done** | Tests pass; concurrent assembly works; timeout cleanup verified; reviewed. |

### Task 3.1.5 — Integration with Block Processing Pipeline

| Field | Detail |
|-------|--------|
| **Description** | Wire `BlockAssemblyManager` into the block processing pipeline. When a node receives block pieces (not full blocks), assembly is attempted. Once assembled, the block enters the normal validation/execution pipeline. Full blocks (from peers that don't use erasure coding) bypass assembly. Add config flag to enable/disable block erasure coding. |
| **Estimated Effort** | 2 story points |
| **Assignee/Role** | Core Protocol Engineer |
| **Testing Method** | (1) Block from pieces → enters validation pipeline after assembly. (2) Full block from legacy peer → enters pipeline directly. (3) Config flag disables erasure coding (legacy mode). (4) Metrics: assembly latency, piece arrival rate. |
| **Definition of Done** | Tests pass; end-to-end pipeline works with both modes; reviewed. |

---

## Codebase Locations

| File | Relevance |
|------|-----------|
| `pkg/das/erasure/reed_solomon_encoder.go:37-54` | `RSEncoderGF256` — production-ready GF(2^8) Reed-Solomon encoder. k-of-n MDS encoding. Can be reused for block pieces (up to 255 total shards per GF(2^8) field order). |
| `pkg/das/erasure/reed_solomon_encoder.go:56-85` | `NewRSEncoderGF256()` — creates encoder with precomputed evaluation points. For blocks: `NewRSEncoderGF256(4, 4)` = 4 data + 4 parity. |
| `pkg/das/erasure/galois_field.go:23-81` | GF(2^8) field arithmetic — `gf256Modulus=0x11D`, pre-computed log/exp/mul/inv tables. Used by `RSEncoderGF256`. |
| `pkg/das/erasure/reed_solomon.go:1-171` | XOR-based Reed-Solomon (simple demo). Not suitable for block-level encoding — use `RSEncoderGF256` instead. |
| `pkg/das/block_in_blob.go:106-171` | `BlobBlockEncoder.EncodeBlock()` — sequential chunking into blobs. NOT erasure-coded. This is the block-in-blobs feature (K+ roadmap). Block erasure coding is different: it's for propagation, not blob storage. |
| `pkg/das/reconstruction.go:160-233` | `ReconstructBlob()` — Lagrange interpolation for blob reconstruction from 64/128 cells. Pattern is relevant (k-of-n recovery) but uses BLS12-381 field, not GF(2^8). |
| `pkg/p2p/block_gossip.go:117-153` | `PropagateBlock()` — sqrt(n) fanout at line 133. Currently sends full blocks. Must be extended to send individual pieces to different peer subsets. |
| `pkg/p2p/block_gossip.go:26-42` | `BlockGossipConfig` — MaxBlockSize=10 MiB. With 8 pieces, each piece is ~1.25 MiB + overhead. |
| `pkg/p2p/gossip_topics.go:19` | `BeaconBlock` topic — global block propagation. Need to add `BlockPiece` topic. |
| `pkg/p2p/gossip.go:116-165` | `PublishMessage()` — topic-based pub/sub. Block pieces can use the same mechanism. |
| `pkg/das/cell_gossip.go:54-72` | `GossipRouter` — routes cells to custody subnets. Pattern reusable for routing block pieces to peer subsets. |

---

## Implementation Status

**❌ Not Implemented**

### What Exists
- ✅ `RSEncoderGF256` — production GF(2^8) Reed-Solomon encoder supporting k-of-n MDS encoding (`das/erasure/reed_solomon_encoder.go`)
- ✅ GF(2^8) field arithmetic with pre-computed lookup tables (`das/erasure/galois_field.go`)
- ✅ sqrt(n) block fanout gossip (`p2p/block_gossip.go:133`)
- ✅ Gossip topic system with per-topic routing (`p2p/gossip.go`, `p2p/gossip_topics.go`)
- ✅ Block-in-blobs sequential chunking (`das/block_in_blob.go`) — related but different purpose
- ✅ Cell-level gossip routing with subnet assignment (`das/cell_gossip.go`) — pattern reusable
- ✅ Blob reconstruction from partial data (`das/reconstruction.go`) — pattern reusable

### What's Missing
- ❌ `BlockErasureEncoder` — no block-level encoding into erasure-coded pieces
- ❌ `BlockErasureDecoder` — no block-level reconstruction from k-of-n pieces
- ❌ `BlockPiece` gossip topic — no per-piece propagation
- ❌ `BlockAssemblyManager` — no piece collection and assembly
- ❌ Block processing pipeline integration for piece-based block reception
- ❌ Per-piece peer routing (currently sends full blocks to sqrt(n) peers)

### Proposed Solution

1. Create `pkg/das/block_erasure.go` with `BlockErasureEncoder` and `BlockErasureDecoder` wrapping `RSEncoderGF256`
2. Create `pkg/p2p/block_piece_gossip.go` with `BlockPiece` topic, piece routing, and assembly
3. Extend `PropagateBlock()` to optionally encode and send pieces instead of full blocks
4. Add `BlockAssemblyManager` with concurrent piece tracking and reconstruction trigger
5. Config flag: `UseBlockErasureCoding bool` (default false until gigagas era)

### Key Design Decision

Block erasure coding is for **propagation latency** (gigagas L1), while blob erasure coding (PeerDAS) is for **data availability**. They serve different purposes:
- Block erasure: split EL block → pieces → reconstruct at receiver → execute
- Blob erasure: split blob data → columns → sample for availability → reconstruct on demand

---

## Spec Reference

> **Vitalik:**
> For very large blocks (gigagas), propagation latency becomes the bottleneck. We can use erasure coding to split the block into 8 pieces. Each node only needs to download ~4 pieces to reconstruct the full block. Pieces propagate in parallel through different paths in the P2P network, reducing effective latency by ~2x. Combined with streaming execution, this enables blocks up to 100 MB with sub-second propagation.
