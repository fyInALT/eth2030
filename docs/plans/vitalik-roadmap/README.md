# Vitalik Roadmap Gap Analysis — Line-by-Line Plans

> **Source:** Vitalik's "Fast Slots, Fast Finality, Scaling" roadmap document (Feb 2026)
> **Scope:** Gaps between Vitalik's proposals and the ETH2030 codebase implementation

---

## Gap Summary

| ID | Epic | Story | Severity | Status |
|----|------|-------|----------|--------|
| EP-1 | Multidimensional Gas | US-1.1 Gas Reservoir Mechanism | MEDIUM | ❌ Not Implemented |
| EP-1 | Multidimensional Gas | US-1.2 SSTORE State Creation Dimension | MEDIUM | ⚠️ Partial (repriced, not separated) |
| EP-2 | Fast Slots | US-2.1 Random Attester Sampling | MEDIUM | ❌ Not Implemented |
| EP-2 | Fast Slots | US-2.2 Intermediate 8s Slot Step | LOW | ❌ Not Implemented |
| EP-3 | Block Propagation | US-3.1 Block-Level Erasure Coding | MEDIUM | ❌ Not Implemented |
| EP-4 | Finality Protocol | US-4.1 Minimmit One-Round BFT | MEDIUM | ❌ Not Implemented |
| EP-4 | Finality Protocol | US-4.2 Poseidon2 Hash Backend | LOW | ⚠️ Partial (Poseidon1 exists) |

---

## Epics

### EP-1 — Multidimensional Gas Reservoir

Vitalik proposes separating "state creation gas" from execution gas with a **reservoir mechanism**: `GAS` returns only execution gas, `CALL` passes reservoir gas alongside regular gas, and `SSTORE` zero→nonzero draws from the reservoir dimension. ETH2030 has 5-dim pricing (`multidim_gas.go`) but no reservoir semantics in the EVM.

- [US-1.1 — Gas Reservoir Mechanism](US-1.1-gas-reservoir-mechanism.md)
- [US-1.2 — SSTORE State Creation Dimension](US-1.2-sstore-state-creation-dimension.md)

### EP-2 — Fast Slots Infrastructure

Vitalik proposes sqrt(2) progressive slot reduction (12→8→6→4→3→2s) and replacing full committee attestation with 256-1024 randomly-sampled attesters per slot. ETH2030 has 6s slots and full committee selection but lacks the 8s intermediate step and random attester sampling.

- [US-2.1 — Random Attester Sampling](US-2.1-random-attester-sampling.md)
- [US-2.2 — Intermediate 8s Slot Step](US-2.2-intermediate-8s-slot-step.md)

### EP-3 — Block-Level Erasure Coding

Vitalik proposes splitting execution blocks into 8 erasure-coded pieces (k-of-n) for faster propagation under gigagas conditions. ETH2030 has blob-level PeerDAS erasure coding (128 columns, GF(2^8)) and block-in-blobs chunking, but no block-level k-of-n sharding.

- [US-3.1 — Block-Level Erasure Coding](US-3.1-block-level-erasure-coding.md)

### EP-4 — Finality Protocol

Vitalik references Minimmit (one-round BFT) and Poseidon2 alternatives. ETH2030 has SSF, endgame pipeline, and Poseidon1 for ZK circuits, but no Minimmit protocol or Poseidon2.

- [US-4.1 — Minimmit One-Round BFT](US-4.1-minimmit-one-round-bft.md)
- [US-4.2 — Poseidon2 Hash Backend](US-4.2-poseidon2-hash-backend.md)

---

## Codebase Overview

| Area | Key Files | LOC | Status |
|------|-----------|-----|--------|
| Multidim Gas | `core/multidim.go`, `core/multidim_gas.go` | ~750 | 5-dim pricing engine complete; no reservoir |
| SSTORE Gas | `core/vm/gas_table.go`, `core/vm/evm_storage_ops.go` | ~1,300 | Zero→nonzero detection complete; not in separate dimension |
| Slot Timing | `consensus/quick_slots.go`, `consensus/phase_timer.go` | ~560 | 6s slots, 4-slot epochs; no 8s step |
| Attester Selection | `consensus/committee_assignment.go`, `consensus/committee_rotation.go` | ~780 | Full committee shuffle; no random sampling |
| Block Propagation | `p2p/block_gossip.go`, `das/erasure/` | ~800 | sqrt(n) fanout + blob RS coding; no block RS |
| Finality | `consensus/ssf.go`, `consensus/endgame_pipeline.go` | ~640 | SSF + endgame; no Minimmit |
| Hash Backends | `crypto/pqc/hash_backend.go`, `zkvm/poseidon.go` | ~377 | Keccak/SHA256/BLAKE3(stub)/Poseidon1; no Poseidon2 |

---

## Execution Order

```
EP-1 (Gas Reservoir + State Creation) ──┐
                                         ├──→ Integration testing
EP-2 (Random Attesters + 8s Slots) ─────┤
                                         │
EP-3 (Block Erasure Coding) ────────────┤
                                         │
EP-4 (Minimmit + Poseidon2) ────────────┘
```

All epics are independent and can be implemented in parallel.
