# Lean4 Verification Check-list

## Ongoing
- [x] Extend EVM formal core with jump-family opcodes (`JUMP`, `JUMPI`, `JUMPDEST`, `POP`).
- [x] Expand compiler decoding for jump-family opcodes.
- [x] Add execution and run-level correctness lemmas for the expanded opcodes.
- [x] Add bytecode-offset interpreter (`runBytecode`) and decoding helpers (`decodeAt`, `decodeNoImmediate`, `decodePush`).
- [x] Document offset-vs-index jump mismatch using an executable counterexample theorem pair (`run` vs `runBytecode`).
- [ ] Add a full refinement theorem from `compile` to bytecode-offset execution (`compile`↔`runBytecode`).
- [ ] Add machine-state fields for `pc`/memory/calldata/accounting as separate invariants.
- [ ] Prove compilation soundness for a larger instruction subset with controlled jump targets.

## Backlog
- `goals`:
  - define a bytecode operational semantics that respects byte-offset jumps,
  - prove compile-then-run agreement against bytecode-offset semantics,
  - then lift jump semantics from op-index simulation to byte-offset simulation,
  - prove end-to-end `compile`→`run` correctness against bytecode semantics,
  - begin RISC-V interpreter/simulation proof chain once VM core stabilizes.
