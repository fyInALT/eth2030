# Project Status

- Formal Lean workspace is present under `formal/lean`.
- Initial EVM/VM model (`Lean2030/VM`) and richer `Lean2030/EVM` model are now linked in `Lean2030/Lean2030.lean`.
- EVM model now includes:
  - arithmetic and bitwise ops,
  - DUP/SWAP,
  - POP/JUMP/JUMPI/JUMPDEST,
  - compiler and theorem suites.
- Current status: still a semantic subset only (toy control-flow, no memory/tracing, no Go cross-checking).
- Added bytecode-offset execution module `formal/lean/Lean2030/EVM/Bytecode.lean` with `decodeAt`/`decodePush`/`runBytecode`.
- Added executable mismatch lemmas showing current op-index `run` diverges from byte-offset EVM behavior for `PUSH`-preceded jumps.
- Next major deliverable: full `compile`→`runBytecode` refinement theorem and stronger instruction/stack invariants.
