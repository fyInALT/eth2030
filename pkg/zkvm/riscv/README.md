# zkvm/riscv — RISC-V RV32IM CPU emulator with execution witness collection

## Overview

`zkvm/riscv` implements a RISC-V RV32IM processor emulator and a companion execution witness collector for ZK proof generation. The CPU supports the complete RV32I base integer ISA plus the M extension (multiply/divide). Memory is sparse and page-based (4 KB pages, up to 64 MiB), with a memory-mapped I/O region at `0xF0000000` for host communication. Gas metering (1 gas per instruction) bounds execution time.

Every executed instruction is optionally recorded by an `RVWitnessCollector` — capturing PC, instruction word, register state before and after, and any memory operations. The resulting trace can be serialized, deserialized, and committed to a SHA-256 Merkle root for use by proof backends. This implements the K+ roadmap item for canonical RISC-V guest execution.

## Functionality

**CPU**

- `RVCPU` — 32 general-purpose registers, `PC`, `GasLimit`/`GasUsed`, `Steps`, `InputBuf`/`OutputBuf`, optional `*RVWitnessCollector`, and a map of custom `EcallHandler` functions.
- `NewRVCPU(gasLimit uint64) *RVCPU`
- `LoadProgram(code []byte, base, entryPoint uint32) error` — writes bytes into memory and sets PC.
- `Run() error` — run until halt or gas exhaustion.
- `Step() error` — execute one instruction.
- `RegisterEcallHandler(code uint32, h EcallHandler)` — install a custom ECALL handler.
- `ValidateCPUConfig(gasLimit uint64, maxMemoryPages int) error`

Built-in ECALL codes (placed in `a7`/`x17`): `RVEcallHalt=0`, `RVEcallOutput=1`, `RVEcallInput=2`, `RVEcallKeccak256=3`, `RVEcallSHA256=4`, `RVEcallECRecover=5`.

**Memory (`RVMemory`)**

- `NewRVMemory() *RVMemory` — sparse 4 KB page map, default max 16384 pages (64 MiB).
- `ReadByteAt / WriteByteAt`, `ReadHalfword / WriteHalfword`, `ReadWord / WriteWord`
- `LoadSegment(base uint32, data []byte) error` — load an ELF program segment.
- `SetMMIO(read, write func)` — register MMIO handlers for `0xF0000000+`.
- `SetMaxPages(n int)`, `PageCount() int`, `Reset()`

**Witness (`RVWitnessCollector`)**

- `NewRVWitnessCollector() *RVWitnessCollector`
- `RecordStep(pc, instr, regsBefore, regsAfter, memOps)` — appends one `RVWitnessStep`.
- `Serialize() []byte` / `DeserializeWitness(data []byte)` — binary round-trip.
- `ComputeTraceCommitment() [32]byte` — SHA-256 Merkle root over all steps.
- `StepCount() int`, `Reset()`

## Usage

```go
cpu := riscv.NewRVCPU(1_000_000)

// Attach witness collector before loading.
cpu.Witness = riscv.NewRVWitnessCollector()

if err := cpu.LoadProgram(elfBytes, 0x10000, 0x10000); err != nil {
    log.Fatal(err)
}
cpu.InputBuf = inputBytes

if err := cpu.Run(); err != nil && !errors.Is(err, riscv.ErrRVHalted) {
    log.Fatal(err)
}

output := cpu.OutputBuf
commitment := cpu.Witness.ComputeTraceCommitment()
traceBytes := cpu.Witness.Serialize()
```

---

Parent package: [`zkvm`](../)
