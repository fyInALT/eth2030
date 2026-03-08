# vm/ewasm — eWASM interpreter for canonical guest roadmap

[← vm](../README.md)

## Overview

Package `ewasm` provides a stack-based WebAssembly interpreter and execution engine targeting the J+/K+ roadmap items "more precompiles in eWASM", "STF in eRISC", and "canonical guest". It supports a subset of the WebAssembly MVP instruction set (i32 arithmetic, control flow, locals, linear memory) with gas metering and memory bounds checking.

The package contains four cooperating components: `Interpreter` (single-instruction dispatch), `Executor` (structured block/loop/branch control flow, function calls), `Engine` (module-level execution with host function hooks), `Optimizer` (peephole and dead-code elimination passes), and `Precompiles` (WASM-callable Ethereum host functions).

## Functionality

### Core execution

```go
// Interpreter: flat bytecode dispatch.
type WasmInterpreter struct { ... }
func NewWasmInterpreter(maxStack, maxMemory int) *WasmInterpreter
func (wi *WasmInterpreter) Execute(bytecode []byte, locals []int32, gas uint64) ([]int32, uint64, error)

// Executor: structured control flow with block/loop/call.
type WASMExecutor struct { ... }
func NewWASMExecutor(maxStack, maxMemory, maxCallDepth int) *WASMExecutor
func (we *WASMExecutor) Execute(bytecode []byte, locals []int32, gasLimit uint64) ([]int32, uint64, error)

// Engine: module-level with host function registry.
type WasmEngine struct { ... }
func NewWasmEngine(maxStack, maxMemory, maxCallDepth int) *WasmEngine
func (e *WasmEngine) RegisterHostFunc(name string, fn HostFunction)
func (e *WasmEngine) Execute(module WasmModule, funcName string, args []int32, gas uint64) ([]int32, uint64, error)
```

### Optimizer

```go
type WasmOptimizer struct { ... }
func NewWasmOptimizer() *WasmOptimizer
func (opt *WasmOptimizer) Optimize(bytecode []byte) ([]byte, error)
// Passes: dead code elimination, constant folding, nop removal.
```

### Host precompiles

```go
type EthPrecompiles struct { ... }
func NewEthPrecompiles() *EthPrecompiles
// Registered host functions: eth_keccak256, eth_address, eth_balance,
//   eth_caller, eth_callvalue, eth_sload, eth_sstore, eth_return, eth_revert
```

### JIT compilation stub

```go
type WasmJIT struct { ... }
func NewWasmJIT(maxStack, maxMemory int) *WasmJIT
func (j *WasmJIT) Compile(bytecode []byte) ([]byte, error)
func (j *WasmJIT) Execute(compiled []byte, locals []int32, gas uint64) ([]int32, uint64, error)
```

### Supported opcodes

`unreachable`, `nop`, `block`, `loop`, `br`, `br_if`, `return`, `call`, `end`, `drop`, `select`, `local.get`, `local.set`, `i32.load`, `i32.store`, `i32.const`, `i32.eqz`, `i32.eq`, `i32.lt_u`, `i32.gt_u`, `i32.add`, `i32.sub`, `i32.mul`, `i32.div_u`, `i32.rem_u`, `i32.and`, `i32.or`, `i32.xor`, `i32.shl`, `i32.shr_u`
