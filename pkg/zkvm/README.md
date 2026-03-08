# zkvm

Zero-knowledge virtual machine framework for proof-carrying Ethereum block execution, implementing the canonical zkVM (M+ roadmap), STF-in-zkISA executor (J+ roadmap), and exposed zkISA bridge (M+ roadmap).

## Overview

The `zkvm` package provides the full zero-knowledge virtual machine stack for ETH2030, enabling verifiable computation over Ethereum state transitions. It spans three interconnected layers: the zxVM (a compact RISC-like bytecode VM with integrated trace generation), the canonical RISC-V guest execution framework, and the zkISA bridge that exposes ZK-circuit operations to EVM contracts.

The package implements the K+/M+ roadmap milestone for mandatory proof-carrying blocks: each block is accompanied by a ZK proof that the state transition function (STF) was applied correctly. Verifiers can check this proof without re-executing the block. The `STFExecutor` encodes a block's transactions and pre-state into a guest program input, executes it on the RISC-V CPU emulator, and generates a proof via the proof backend. The `CanonicalGuestPrecompile` at address `0x0200` exposes this capability to smart contracts.

The zkISA bridge (`zkisa_bridge.go`) implements the M+ "exposed zkISA" feature: EVM contracts call the precompile at `0x20` with a 4-byte operation selector to invoke ZK-circuit-backed operations including Keccak-256, SHA-256, ECDSA recovery, modular exponentiation, BN256 operations, BLS12-381 verification, and custom guest programs. This enables efficient proof-of-computation for EVM-level operations without requiring a full ZK circuit per contract.

## Table of Contents

- [Functionality](#functionality)
- [Subpackages](#subpackages)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### zxVM — Canonical Zero-Knowledge VM

`ZxVMInstance` is a stack-based VM with a RISC-like instruction set, designed for proof-friendly execution. Programs are loaded as bytecode and produce verifiable execution traces.

Opcodes:
| Opcode | Byte | Description |
|--------|------|-------------|
| `ZxOpPUSH` | `0x09` | Push immediate uint64 (8 LE bytes follow) |
| `ZxOpADD` | `0x01` | Pop a, b; push a+b |
| `ZxOpSUB` | `0x02` | Pop a, b; push a-b |
| `ZxOpMUL` | `0x03` | Pop a, b; push a*b |
| `ZxOpLOAD` | `0x04` | Pop addr; push mem[addr] |
| `ZxOpSTORE` | `0x05` | Pop addr, val; mem[addr] = val |
| `ZxOpJUMP` | `0x06` | Pop target; pc = target |
| `ZxOpHASH` | `0x08` | Pop val; push keccak256(val as 8 LE bytes)[0:8] |
| `ZxOpHALT` | `0x07` | Stop execution |

Key methods on `ZxVMInstance`:
- `LoadProgram(prog *ZxProgram) error` — reset VM state and load a program
- `Execute() (*ZxExecutionResult, error)` — run to completion, return output + gas + `ProofCommitment`
- `GenerateTrace() (*ZxTrace, error)` — record every step as a `ZxStep` (PC, opcode, stack hash, memory hash)
- `VerifyTrace(trace *ZxTrace) (bool, error)` — replay execution and verify step hashes
- `EstimateGas() (uint64, error)` — dry-run without side effects

`ZxExecutionResult` contains: `Output []byte`, `GasUsed uint64`, `ProofCommitment []byte` (Keccak256 of code || output || gas || cycles), `CycleCount uint64`, `MemoryPeak uint64`.

Default configuration (`DefaultZxVMConfig`): 4M cycles, 512 KiB memory, STARK proof system, 1024-entry stack.

`BuildZxPush(val uint64) []byte` builds a correctly encoded PUSH instruction.

### RISC-V Guest Execution (Canonical zkVM)

`RiscVGuest` wraps a RISC-V RV32IM binary program and input buffer for execution on the `RVCPU` emulator. `Execute() (*GuestExecution, error)` runs the program and:
1. Loads the program into the CPU at address 0 with entry point 0
2. Sets the stack pointer (`x2/sp`) to `0x80000000`
3. Attaches an `RVWitnessCollector` to record the execution trace
4. After execution, generates proof data via `ProveExecution` using the witness trace

`GuestExecution` returns: `Output []byte`, `Cycles uint64`, `ProofData []byte`, `Success bool`.

`VerifyExecution(execution, program, input)` verifies that proof data has valid Groth16 structure size.

### Guest Registry

`GuestRegistry` maintains a map from program hash to program bytes, allowing contracts to invoke registered guest programs by hash:
- `RegisterGuest(program []byte) (types.Hash, error)` — register and return the keccak256 hash
- `GetGuest(hash types.Hash) ([]byte, error)` — retrieve a registered program
- `ValidateGuestProgram(program, expectedHash, registry)` — hash match and registration check

### Canonical Guest Precompile

`CanonicalGuestPrecompile` at address `0x0200` (K+ precompile range):
- Input: `programHash(32) || guestInput(variable)`
- Looks up the registered program, executes it on `RiscVGuest`, returns output bytes
- Gas: `GuestPrecompileBaseGas = 100,000` + 1 gas per input byte beyond the hash

### STF Executor (J+ Roadmap)

`RealSTFExecutor` encodes a full Ethereum state transition (block header + transactions) as a guest input, executes it on the RISC-V CPU, and generates a ZK proof of correctness:
- `RealSTFOutput` contains `Valid bool`, `PostRoot types.Hash`, `GasUsed uint64`, `CycleCount uint64`, `ProofData []byte`, `VerificationKey []byte`, `TraceCommitment [32]byte`, `PublicInputsHash [32]byte`
- Default config: `GasLimit = 1<<24` cycles, `MaxWitnessSize = 16 MiB`, STARK proof system

### zkISA Bridge (M+ Roadmap)

`ZKISABridge` at precompile address `0x20` exposes 9 ZK-circuit-backed operations to EVM contracts:
| Selector | Operation | Gas |
|----------|-----------|-----|
| `0x01` | Keccak-256 | 3,000 |
| `0x02` | SHA-256 | 3,000 |
| `0x03` | ECDSA recovery | 5,000 |
| `0x04` | Modular exponentiation | 10,000 |
| `0x05` | BN256 point addition | 2,000 |
| `0x06` | BN256 scalar multiplication | 8,000 |
| `0x07` | BN256 pairing check | 50,000 |
| `0x08` | BLS12-381 signature verify | 12,000 |
| `0xFF` | Custom guest program | 100,000 + 8/byte |

Input format: `selector(4) || operationInput(variable)`.

### Proof Backend

`ProofBackend` (`proof_backend.go`) manages the pluggable proof system. `ProofRequest` carries the witness trace, public inputs, and program hash. `ProofResult` returns `ProofBytes`, `PublicInputsHash [32]byte`, and a `TraceCommitment`. `ProveExecution(req)` dispatches to the configured proof system (STARK / Groth16 / PLONK).

### Supporting Components

- `circuit_builder.go` / `constraint_compiler.go` — arithmetic circuit construction and R1CS constraint compilation
- `proof_aggregator.go` — aggregates multiple execution proofs into a single combined proof
- `verifier.go` — standalone proof verification without full re-execution
- `stf.go` — high-level STF proof interface (wraps `RealSTFExecutor`)
- `leanvm.go` / `ewasm.go` — LeanVM and eWASM guest adapters
- `types_compat.go` — shared type aliases across subpackages

## Subpackages

| Subpackage | Description |
|------------|-------------|
| [`riscv/`](./riscv/) | RISC-V RV32IM CPU emulator: `RVCPU`, memory, witness collector, encoding |
| [`poseidon/`](./poseidon/) | Poseidon hash function for ZK-circuit-friendly hashing |
| [`r1cs/`](./r1cs/) | R1CS (rank-1 constraint system) types and builder |
| [`zkvmtypes/`](./zkvmtypes/) | Shared types: proof systems, program identifiers, circuit parameters |

## Usage

```go
import "github.com/eth2030/eth2030/zkvm"

// Execute a zxVM program and get a proof commitment.
prog := &zkvm.ZxProgram{
    Code:     append(zkvm.BuildZxPush(42), zkvm.BuildZxPush(58)...),
    GasLimit: 10_000,
}
prog.Code = append(prog.Code, zkvm.ZxOpADD, zkvm.ZxOpHALT)

vm := zkvm.NewZxVM(zkvm.DefaultZxVMConfig())
vm.LoadProgram(prog)
result, err := vm.Execute()
// result.Output contains the answer; result.ProofCommitment is the commitment hash.
```

```go
// Register and invoke a RISC-V guest program.
registry := zkvm.NewGuestRegistry()
programHash, err := registry.RegisterGuest(riscvBinary)

guest := zkvm.NewRiscVGuest(riscvBinary, inputData, zkvm.DefaultGuestConfig())
execution, err := guest.Execute()
if execution.Success {
    _ = execution.Output    // output bytes
    _ = execution.ProofData // Groth16 proof
}
```

```go
// Use the canonical guest precompile interface.
precompile := &zkvm.CanonicalGuestPrecompile{
    Registry: registry,
    Config:   zkvm.DefaultGuestConfig(),
}
// Input: programHash(32) || guestInput
output, err := precompile.Run(input)
```

```go
// Prove a state transition.
stfConfig := zkvm.DefaultRealSTFConfig()
executor := zkvm.NewRealSTFExecutor(stfConfig, registry)
output, err := executor.Execute(stfInput)
if output.Valid {
    _ = output.PostRoot  // verified post-state root
    _ = output.ProofData // STARK proof bytes
}
```

## Documentation References

- [EIP-8025: Execution Witness and Proofs](https://eips.ethereum.org/EIPS/eip-8025)
- [L1 Strawmap — J+: STF in zkISA framework](../../docs/ROADMAP.md)
- [L1 Strawmap — K+: Canonical guest (RISC-V CPU)](../../docs/ROADMAP.md)
- [L1 Strawmap — M+: Canonical zkVM, exposed zkISA](../../docs/ROADMAP.md)
