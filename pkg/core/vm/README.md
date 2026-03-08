# core/vm — EVM interpreter and precompiles

[← core](../README.md)

## Overview

Package `vm` implements the Ethereum Virtual Machine used during state transition. It provides a stack-based interpreter for 164+ opcodes (legacy EVM plus EOF, EIP-8141 frame opcodes, EIP-7701 AA opcodes, and miscellaneous purge candidates), a dynamic gas table, 24 precompiled contracts, and parallel execution support targeting gigagas throughput.

The interpreter is configurable via `Config` and operates against any `StateDB`-compatible backend. EOF containers (EIP-3540) are validated by `EOFValidator` before execution. The `PrecompileRegistry` allows fork-based dynamic registration of precompiles.

## Functionality

### Interpreter and EVM

```go
type Config struct { ... }

// Primary entry point for contract execution.
type EVM struct { ... }

// Call executes a message call to addr with the given input and gas.
func (evm *EVM) Call(caller, addr types.Address, input []byte, gas uint64, value *big.Int) ([]byte, uint64, error)

// Create deploys a new contract.
func (evm *EVM) Create(caller types.Address, code []byte, gas uint64, value *big.Int) ([]byte, types.Address, uint64, error)
```

### Opcode categories

| Category | Opcodes |
|----------|---------|
| Arithmetic | ADD, MUL, SUB, DIV, MOD, EXP, CLZ, ... |
| Stack | POP, PUSH1..PUSH32, DUP1..DUP16, SWAP1..SWAP16, DUPN, SWAPN, EXCHANGE |
| Memory | MLOAD, MSTORE, MSTORE8, MSIZE, MCOPY |
| Storage | SLOAD, SSTORE, TLOAD, TSTORE (EIP-1153) |
| Control | JUMP, JUMPI, PC, STOP, RETURN, REVERT |
| EOF | EXTCALL, EXTDELEGATECALL, EXTSTATICCALL, RETURNDATALOAD, EOFCREATE, RETURNCONTRACT, DATALOAD, DATALOADN, DATASIZE, DATACOPY |
| Frame (EIP-8141) | APPROVE, TXPARAM* |
| AA (EIP-7701) | CURRENT_ROLE, ACCEPT_ROLE |

### Precompiles (24 total)

- 0x01–0x09: Standard precompiles (ecRecover, SHA-256, RIPEMD-160, identity, modexp, BN254 add/mul/pairing, blake2f)
- 0x0a: Point evaluation (EIP-4844 KZG)
- BLS12-381 precompiles (9): EIP-2537 (G1/G2 add, mul, MSM, pairing, map-to-G1/G2)
- NTT precompile (EIP-7885): number-theoretic transform
- NII precompiles (4): modexp, field-mul, field-inv, batch-verify
- Field precompiles (4): additional field arithmetic
- 7702 precompile: EIP-7702 SetCode helper
- AA proof precompile: EIP-7701 nonce/sig/gas constraint verification

```go
type PrecompileRegistry struct { ... }
func NewPrecompileRegistry() *PrecompileRegistry
func (r *PrecompileRegistry) Register(info PrecompileInfo) error
func (r *PrecompileRegistry) Lookup(addr types.Address) (*PrecompileInfo, bool)
func (r *PrecompileRegistry) ActiveAt(fork string) []PrecompileInfo
```

### Gas

- `dynamic_gas.go` — per-opcode dynamic gas functions
- `gas_eip2929.go` — EIP-2929 warm/cold access costs
- `gas_verkle.go` — EIP-4762 stateless gas (Verkle)
- `gas_table_ext.go` — Glamsterdam repricing extensions
- `gas_cache.go` — gas cost memoisation
- `gas_pool_tracker.go` — cross-block gas pool tracking
- `gas_scheduler.go` — parallel execution gas scheduling
- `gas_futures_long.go` — long-dated gas futures (M+)

### Parallel execution

```go
type ParallelExecutor struct { ... }
func NewParallelExecutor(workers int) *ParallelExecutor
func (pe *ParallelExecutor) ExecuteBatch(txs []*types.Transaction, state StateDB, header *types.Header) ([]*ExecutionResult, error)
```

### EOF (EIP-3540)

```go
type EOFContainer struct { ... }
func ParseEOF(code []byte) (*EOFContainer, error)

type EOFValidator struct { ... }
func (v *EOFValidator) Validate(container *EOFContainer) error
```

### Logging

```go
type Logger interface {
    CaptureStart(evm *EVM, from, to types.Address, create bool, input []byte, gas uint64, value *big.Int)
    CaptureState(pc uint64, op OpCode, gas, cost uint64, scope *ScopeContext, rData []byte, depth int, err error)
    CaptureEnd(output []byte, gasUsed uint64, err error)
}
```

## Subpackages

- [`abi/`](abi/README.md) — Solidity ABI encoder/decoder
- [`ewasm/`](ewasm/README.md) — eWASM interpreter for canonical guest roadmap
