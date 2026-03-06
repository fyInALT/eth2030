package vm

// ewasm_compat.go re-exports types from core/vm/ewasm for backward compatibility.

import (
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/core/vm/ewasm"
)

// Ewasm type aliases.
type (
	EWASMEngine             = ewasm.EWASMEngine
	WASMOpcode              = ewasm.WASMOpcode
	WASMExecutorConfig      = ewasm.WASMExecutorConfig
	WASMFrame               = ewasm.WASMFrame
	WASMExecutor            = ewasm.WASMExecutor
	WasmValueType           = ewasm.WasmValueType
	WasmValue               = ewasm.WasmValue
	WasmInstruction         = ewasm.WasmInstruction
	EWASMInterpreterConfig  = ewasm.EWASMInterpreterConfig
	EWASMInterpreter        = ewasm.EWASMInterpreter
	WasmOpcode              = ewasm.WasmOpcode
	WasmOp                  = ewasm.WasmOp
	OptimizationPass        = ewasm.OptimizationPass
	OptimizationMetrics     = ewasm.OptimizationMetrics
	ConstantFolding         = ewasm.ConstantFolding
	DeadCodeElimination     = ewasm.DeadCodeElimination
	StackScheduling         = ewasm.StackScheduling
	InliningPass            = ewasm.InliningPass
	LoopUnrolling           = ewasm.LoopUnrolling
	OptimizationPipeline    = ewasm.OptimizationPipeline
	EWASMPrecompile         = ewasm.EWASMPrecompile
	EWASMPrecompileRegistry = ewasm.EWASMPrecompileRegistry
	WasmSection             = ewasm.WasmSection
	WasmModule              = ewasm.WasmModule
	JITCache                = ewasm.JITCache
	WasmGasCalculator       = ewasm.WasmGasCalculator
)

// Ewasm error variables.
var (
	ErrInterpStackOverflow    = ewasm.ErrInterpStackOverflow
	ErrInterpStackUnderflow   = ewasm.ErrInterpStackUnderflow
	ErrInterpDivisionByZero   = ewasm.ErrInterpDivisionByZero
	ErrInterpOutOfGas         = ewasm.ErrInterpOutOfGas
	ErrInterpUnknownOpcode    = ewasm.ErrInterpUnknownOpcode
	ErrInterpLocalOutOfRange  = ewasm.ErrInterpLocalOutOfRange
	ErrInterpUnreachable      = ewasm.ErrInterpUnreachable
	ErrInterpInvalidSelect    = ewasm.ErrInterpInvalidSelect
	ErrWASMStackOverflow      = ewasm.ErrWASMStackOverflow
	ErrWASMStackUnderflow     = ewasm.ErrWASMStackUnderflow
	ErrWASMOutOfMemory        = ewasm.ErrWASMOutOfMemory
	ErrWASMDivisionByZero     = ewasm.ErrWASMDivisionByZero
	ErrWASMUnreachable        = ewasm.ErrWASMUnreachable
	ErrEWASMPrecompileNotFound = ewasm.ErrEWASMPrecompileNotFound
	ErrEWASMExecutionFailed   = ewasm.ErrEWASMExecutionFailed
	ErrEWASMOutOfGas          = ewasm.ErrEWASMOutOfGas
	ErrEWASMAlreadyRegistered = ewasm.ErrEWASMAlreadyRegistered
)

// Ewasm function wrappers.
func NewEWASMEngine() *EWASMEngine                { return ewasm.NewEWASMEngine() }
func NewWASMExecutor(config WASMExecutorConfig) *WASMExecutor {
	return ewasm.NewWASMExecutor(config)
}
func NewEWASMInterpreter(config EWASMInterpreterConfig) *EWASMInterpreter {
	return ewasm.NewEWASMInterpreter(config)
}
func NewI32Const(val uint32) WasmOp               { return ewasm.NewI32Const(val) }
func NewI64Const(val uint64) WasmOp               { return ewasm.NewI64Const(val) }
func NewLocalGet(idx uint32) WasmOp               { return ewasm.NewLocalGet(idx) }
func NewLocalSet(idx uint32) WasmOp               { return ewasm.NewLocalSet(idx) }
func NewCallOp(funcIdx uint32) WasmOp             { return ewasm.NewCallOp(funcIdx) }
func NewOptimizationPipeline(passes ...OptimizationPass) *OptimizationPipeline {
	return ewasm.NewOptimizationPipeline(passes...)
}
func NewEWASMPrecompileRegistry() *EWASMPrecompileRegistry {
	return ewasm.NewEWASMPrecompileRegistry()
}
func NewJITCache(capacity int) *JITCache         { return ewasm.NewJITCache(capacity) }
func DefaultWasmGasCalculator() *WasmGasCalculator { return ewasm.DefaultWasmGasCalculator() }
func ValidateWasmBytecode(code []byte) error     { return ewasm.ValidateWasmBytecode(code) }
func CompileModule(code []byte) (*WasmModule, error) { return ewasm.CompileModule(code) }
func ExecuteExport(module *WasmModule, funcName string, args []byte) ([]byte, error) {
	return ewasm.ExecuteExport(module, funcName, args)
}
func BuildMinimalWasm(exportNames ...string) []byte { return ewasm.BuildMinimalWasm(exportNames...) }

// Ensure core/types import is used (needed for types.Hash in JITCache).
var _ = types.Hash{}
