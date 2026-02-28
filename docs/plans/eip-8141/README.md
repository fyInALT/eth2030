# EIP-8141 Frame Transaction — Story Index

> **EIP Reference:** [EIP-8141 Frame Transaction](https://ethereum-magicians.org/t/frame-transaction/27617)
> **Requires:** EIP-2718 (typed envelopes), EIP-4844 (blob transactions)
> **Full Plan:** [eip-8141-frame-tx.md](../eip-8141-frame-tx.md)

## Stories

| File | Story | Epic | SP | Status |
|------|-------|------|----|--------|
| [US-1.1-rlp-serialization.md](US-1.1-rlp-serialization.md) | Frame Transaction RLP Serialization | EP-1 | 7 | ✅ Mostly |
| [US-1.2-fee-calculation.md](US-1.2-fee-calculation.md) | Frame Transaction Fee Calculation | EP-1 | 3 | ⚠️ Partial |
| [US-2.1-frame-modes.md](US-2.1-frame-modes.md) | Frame Mode Definitions and Caller Identity | EP-2 | 7 | ⚠️ Partial |
| [US-3.1-approve-opcode.md](US-3.1-approve-opcode.md) | APPROVE Opcode Core Behavior | EP-3 | 12 | ✅ Mostly |
| [US-4.1-txparam-opcodes.md](US-4.1-txparam-opcodes.md) | TXPARAM Opcode Family | EP-4 | 16 | ✅ Mostly |
| [US-5.1-execution-engine.md](US-5.1-execution-engine.md) | Frame-by-Frame Execution Orchestrator | EP-5 | 13 | ❌ Missing |
| [US-6.1-gas-accounting.md](US-6.1-gas-accounting.md) | Per-Frame Gas Isolation | EP-6 | 5 | ⚠️ Partial |
| [US-7.1-receipt-structure.md](US-7.1-receipt-structure.md) | Frame Transaction Receipt | EP-7 | 7 | ✅ Mostly |
| [US-8.1-signature-hash.md](US-8.1-signature-hash.md) | Canonical Signature Hash | EP-8 | 3 | ✅ Done |
| [US-9.1-static-validation.md](US-9.1-static-validation.md) | Static Transaction Validation | EP-9 | 3 | ✅ Mostly |
| [US-10.1-frame-interactions.md](US-10.1-frame-interactions.md) | Cross-Frame State Interactions | EP-10 | 5 | ❌ Missing |
| [US-11.1-origin-opcode.md](US-11.1-origin-opcode.md) | ORIGIN Opcode Behavior Change | EP-11 | 4 | ❌ Missing |
| [US-12.1-mempool-validation.md](US-12.1-mempool-validation.md) | Mempool Validation & DoS Mitigation | EP-12 | 14 | ❌ Missing |
| [US-13.1-integration-tests.md](US-13.1-integration-tests.md) | Integration Test Suite | EP-13 | 23 | ❌ Missing |
| **Total** | | | **~124 SP** | |

> Each story file includes: tasks, EIP reference excerpts, codebase locations, and implementation status.

## Sprint Allocation

| Sprint | Focus | SP |
|--------|-------|----|
| Sprint 1 | Foundations (RLP, fees, static validation, sig hash) | ~24 |
| Sprint 2 | APPROVE + VERIFY/SENDER modes + execution engine | ~29 |
| Sprint 3 | TXPARAM* opcodes + receipt | ~25 |
| Sprint 4 | Mempool + cross-frame state | ~23 |
| Sprint 5 | E2E integration testing | ~23 |
