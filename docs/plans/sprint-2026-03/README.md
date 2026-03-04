# ETH2030 Sprint User Stories — March 2026

> **Generated from**: `docs/plans/leanroadmap-coverage-2026-03.md` and all docs in `docs/plans/vitalik/`
> **Framework**: Scrum / Agile (INVEST criteria)
> **Date**: 2026-03-04
> **Story Points scale**: Fibonacci (1, 2, 3, 5, 8, 13, 21)

---

## INVEST Compliance Legend

| Criterion | Description |
|-----------|-------------|
| **I** Independent | Story can be worked without depending on another unfinished story |
| **N** Negotiable | Scope and approach can be refined in sprint planning |
| **V** Valuable | Delivers measurable value to protocol, interop, or security |
| **E** Estimable | Enough detail to size the effort |
| **S** Small | Fits in one sprint (≤13 SP) |
| **T** Testable | Has explicit, verifiable acceptance criteria |

---

## Epic Index

| Epic | Title | Stories | Total SP |
|------|-------|---------|----------|
| EP-1 | Account Abstraction & EIP-8141 | [US-AA-1 … US-AA-5](ep-1-account-abstraction.md) | 47 |
| EP-2 | EL State Tree, BLAKE3 & RISC-V VM | [US-BL-1, US-EL-2 … US-EL-4](ep-2-el-state-blake3-riscv.md) | 45 |
| EP-3 | Post-Quantum Cryptography | [US-PQ-2 … US-PQ-6 (incl. 5a/5b split)](ep-3-post-quantum-crypto.md) | 58 |
| EP-4 | leanConsensus & leanroadmap | [US-LEAN-1 … US-LEAN-6, US-LEAN-3a/3b, US-LEAN-8](ep-4-lean-consensus.md) | 69 |
| EP-5 | Vitalik Roadmap Gaps | [US-GAP-1 … US-GAP-5, US-GAP-7](ep-5-vitalik-roadmap-gaps.md) | 70 |
| EP-6 | Block Building Pipeline | [US-BB-1 … US-BB-2](ep-6-block-building-pipeline.md) | 18 |
| EP-7 | EIP Specification Compliance | [US-SPEC-1, US-SPEC-3 … US-SPEC-7](ep-7-eip-spec-compliance.md) | 63 |
| **Total** | | **37 stories** | **370 SP** |

---

## INVEST Refinement Log

The following corrections were applied after cross-checking all stories against their source EIP documents:

| Story | Issue | Fix Applied |
|-------|-------|-------------|
| US-PQ-5 | **S-violation**: labelled 13 SP but tasks total 21 SP | Split into US-PQ-5a (13 SP) + US-PQ-5b (8 SP) |
| US-AA-5 | **I-note**: depends on US-PQ-3 completing NTT address alignment first | Added `Depends on: US-PQ-3` note |
| EP-1 overall | EIP-8141 frame receipt structure, TSTORE cross-frame semantics, SENDER mode enforcement, TXPARAM* table completeness — not covered | Added US-SPEC-1, US-SPEC-2 in EP-7 |
| EP-5/EP-6 | EIP-7732 builder withdrawal lifecycle and epoch processing — not covered | Added US-SPEC-3 in EP-7 |
| EP-4/EP-6 | EIP-7805 IL equivocation detection, satisfaction algorithm, engine API status — not covered | Added US-SPEC-4 in EP-7 |
| EP-5 | EIP-7928 BAL ordering, ITEM_COST=2000 sizing, BlockAccessIndex assignment, retention policy — not covered | Added US-SPEC-5 in EP-7 |
| (missing) | EIP-7706 3D fee vector transaction type — **entirely absent** | Added US-SPEC-6 in EP-7 |
| EP-2 | EIP-7864 get_tree_key correctness, header data packing, code chunking accuracy — not covered by US-BL-1 (which covers hash function) | Added US-SPEC-7 in EP-7 |

### Consolidation Pass (2026-03-04)

| Change | Stories Affected | Result |
|--------|-----------------|--------|
| **Merge A** | US-EL-1 (BLAKE3 trie, 8 SP) + US-PQ-1 (BLAKE3 PQ crypto, 5 SP) | Combined into **US-BL-1** "BLAKE3 Hash Backend Integration" (12 SP). Moved out of EP-2 and EP-3; added as first story of EP-2 (renamed to "EL State Tree, BLAKE3 & RISC-V VM"). |
| **Merge B** | US-SPEC-1 (frame receipt, 8 SP) + US-SPEC-2 (TXPARAM*, 5 SP) | Combined into **US-SPEC-1** "EIP-8141 Frame TX Full Compliance" (13 SP). SPEC-2.1 → SPEC-1.5, SPEC-2.2 → SPEC-1.6. US-SPEC-2 removed from EP-7. |
| **Merge C** | US-GAP-3 (Random attesters, 13 SP) + US-LEAN-7 (Reduced committee fork-choice, 8 SP) | **US-GAP-3** retained at 13 SP; LEAN-7 devnet test content absorbed into GAP-3.4 DoD. LEAN-7.1 and LEAN-7.2 were duplicates of GAP-3.1 and GAP-3.3. US-LEAN-7 removed from EP-4. |
| **Merge D** | US-GAP-5 (Minimmit, 13 SP) + US-GAP-6 (3SF backoff, 5 SP) | **US-GAP-5** retained at 13 SP; GAP-6.2 simulation content absorbed into GAP-5.3 acceptance criteria. GAP-5 description and acceptance criteria updated to mention is_justifiable_slot backoff and 1000-slot simulation. US-GAP-6 removed from EP-5. |
| **Split E** | US-LEAN-3 "Separate PQ Aggregator Role" (labeled 13 SP, tasks summed to 14 SP) | Split into **US-LEAN-3a** "PQ Aggregator Role — Types & Duty Selection" (6 SP, LEAN-3.1 + LEAN-3.2) and **US-LEAN-3b** "PQ Aggregator Role — Collection & Aggregation" (8 SP, LEAN-3.3 + LEAN-3.4). |

### SP Label Fixes

| Story | Old SP Label | Correct SP (task sum) |
|-------|--------------|-----------------------|
| US-AA-3 | 13 | 12 (3+5+1+2+1=12) |
| US-AA-5 | 8 | 9 (2+5+2=9) |
| US-EL-3 | 8 | 10 (2+5+3=10) |
| US-EL-4 | 8 | 10 (5+2+3=10) |
| US-GAP-7 | 8 | 10 (5+2+3=10) |

---

---

# Sprint Planning Summary

## Suggested Sprint Breakdown

| Sprint | Focus | Stories | SP |
|--------|-------|---------|-----|
| Sprint 1 | P0 blockers | US-AA-1(13), US-AA-3(12), US-BL-1(12), US-EL-4(10), US-PQ-3(13), US-GAP-7(10), US-LEAN-1(8) | 78 |
| Sprint 2 | Interop + core spec | US-AA-2(8), US-EL-2(13), US-EL-3(10), US-PQ-2(8), US-LEAN-2(8), US-GAP-3(13), US-SPEC-1(13) | 73 |
| Sprint 3 | Spec compliance + gaps | US-PQ-4(8), US-SPEC-3(8), US-SPEC-4(8), US-SPEC-5(13), US-SPEC-7(8), US-GAP-1(13), US-GAP-2(8) | 66 |
| Sprint 4 | PQ hardening + 3D gas | US-AA-4(5), US-AA-5(9), US-PQ-6(8), US-GAP-5(13), US-SPEC-6(13), US-LEAN-3a(6), US-LEAN-3b(8) | 62 |
| Sprint 5 | STARK + networking | US-PQ-5a(8), US-PQ-5b(13), US-LEAN-5(8), US-LEAN-6(13), US-GAP-4(13) | 55 |
| Sprint 6 | Research + finality + P2P | US-LEAN-4(13), US-LEAN-8(5), US-BB-1(13), US-BB-2(5) | 36 |

> **Note**: Sprint 1 (78 SP) contains the hard P0 blockers: US-PQ-3 (NTT address fix), US-EL-4 (KZG backend), US-BL-1 (BLAKE3 backend), US-GAP-7 (BLS backend). These should be tackled by senior engineers first. US-SPEC-1 (frame TX full compliance, 13 SP) moved to Sprint 2 — it was previously a Sprint 1 entry but was deprioritised to keep Sprint 1 focused on infrastructure. US-GAP-6 and US-LEAN-7 have been merged into US-GAP-5 and US-GAP-3 respectively and are no longer separate sprint items.

---

## Role Allocation Guide

| Role | Primary Stories |
|------|----------------|
| **Go Engineer (vm/core)** | US-GAP-1, US-GAP-2, US-EL-2, US-EL-3, US-PQ-3, US-AA-3, US-BB-2, US-BL-1 |
| **Consensus Engineer** | US-LEAN-1, US-LEAN-2, US-LEAN-3a, US-LEAN-3b, US-LEAN-6, US-LEAN-8, US-GAP-3, US-GAP-5 |
| **ZK Engineer** | US-PQ-5a, US-PQ-5b, US-PQ-6, US-EL-2 |
| **P2P Engineer** | US-LEAN-4, US-LEAN-5, US-GAP-4, US-PQ-4, US-BB-1 |
| **Security Engineer** | US-AA-5, US-PQ-3, US-AA-1 |
| **Protocol Researcher** | US-BB-2, US-LEAN-6 |
| **QA / DevOps Engineer** | US-EL-4, US-LEAN-1, US-GAP-7, US-PQ-4, all benchmark tasks |

---

## Definition of Done — Global Criteria

All stories must meet the following DoD in addition to story-specific criteria:

1. **Compilation**: `cd pkg && go build ./...` passes with zero errors.
2. **Tests**: `cd pkg && go test ./...` passes with zero failures.
3. **Coverage**: New code has ≥ 80% line coverage via `go test -cover`.
4. **Format**: `go fmt ./...` produces no diff.
5. **EF State Tests**: 36,126/36,126 still passing after any EVM/state changes.
6. **Code Review**: At least one engineer other than the implementor reviews and approves.
7. **Docs**: Any new CLI flags documented in `cmd/eth2030/main.go` help text.
8. **Commit hygiene**: Each commit is atomic, < 40-char subject, passes CI.
9. **No Claude attribution**: No `Co-Authored-By` AI lines in commits.

---

*Generated: 2026-03-04. Sources: docs/plans/leanroadmap-coverage-2026-03.md, docs/plans/vitalik/\*.*
