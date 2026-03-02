# Lean + RISC-V implementation strategy for eth2030

This is the first concrete slice to bridge toward a full Lean-verified VM.

## Lean 4 strategy and implementation plan

The formalization is intentionally staged to keep the trusted base small and mechanically checkable:

1. Keep the existing Go implementation as the executable behavior reference.
2. Define pure Lean operational semantics for each slice first (currently: `VM` and subset `EVM` kernels).
3. Prove byte/stack/gas invariants and compile-time safety properties before adding concurrency or I/O.
4. Strengthen cross-module correctness theorems from bytecode decode → AST `Program` → run state.
5. Extend features incrementally (`jump`, memory accounting, calldata, storage, receipts).
6. Lift verified semantics into the RISC-V witness pipeline and prove a full refinement bridge.

## Prerequisites for Lean verification work

- Install Lean 4.28.0 and Lake:
  - `curl -sSf https://raw.githubusercontent.com/leanprover/elan/master/elan-init.sh | sh`
  - `elan toolchain install 4.28.0`
  - `cd formal/lean && elan install` (if required by your environment)
- Run Lean checks:
  - `cd formal/lean && lake build`
  - `cd formal/lean && lake test`

## Milestone 1 (today)

- Introduce `formal/lean` as a Lean 4 proof workspace.
- Define a complete formal model for the current `pkg/zkvm/leanvm` instruction set.
- Prove core invariants about stack size, gas monotonicity, and byte-level operations.

Current progress from this commit:
- Added `formal/lean/lakefile.lean` and `formal/lean/lean-toolchain`.
- Added `formal/lean/Lean2030/VM/Spec.lean` with executable Lean semantics.
- Added `formal/lean/Lean2030/VM/Compile.lean` matching `pkg/zkvm/leanvm.go` translation.
- Added `formal/lean/Lean2030/VM/Proofs.lean` with initial theorem skeletons and invariants.

## Recommended next milestones

1. **Core EVM VM kernel extraction**
   - Port `pkg/core/vm` instruction semantics for pure logic-first components.
   - Separate pure gas/memory/accounting layer first, leaving I/O and RPC out.
   - For each opcode, prove:
     - stack discipline
     - gas monotonicity
     - deterministic state transition
     - invalid-state erroring instead of panicking

2. **RISC-V front-end in Lean**
   - Formalize a small RTL-level machine (`riscv_cpu`, `riscv_memory`, `riscv_encode`).
   - Prove a compiler/correctness theorem from EVM bytecode IR → RISC-V micro-ops.

3. **Witness and constraint bridge**
   - Mirror `pkg/zkvm/riscv_witness.go` and `pkg/zkvm/constraint_compiler.go` in Lean structures.
   - Show witness extraction is complete and cycle-aligned.

4. **Cross-language refinement bridge (Go ↔ Lean)**
   - Add a Go test harness that emits deterministic state-transition traces.
   - Add Lean parser over these traces and prove trace acceptance implies Go-state equivalence for the modeled slice.

5. **Proof-carrying production path**
   - Replace trusted Go subset with Lean-certified kernel via generated/statically checked artifacts.
   - Keep the rest of the Go client intact and progressively reduce trusted attack surface.
