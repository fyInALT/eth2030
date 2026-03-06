package zkvm

// riscv_compat.go re-exports types from zkvm/riscv for backward compatibility.

import "github.com/eth2030/eth2030/zkvm/riscv"

// RISCV type aliases.
type (
	EcallHandler       = riscv.EcallHandler
	RVCPU              = riscv.RVCPU
	MemOp              = riscv.MemOp
	RVMemory           = riscv.RVMemory
	RVWitnessStep      = riscv.RVWitnessStep
	RVWitnessCollector = riscv.RVWitnessCollector
)

// RISCV constants.
const (
	RVRegCount       = riscv.RVRegCount
	RVPageSize       = riscv.RVPageSize
	RVPageShift      = riscv.RVPageShift
	RVMMIOBase       = riscv.RVMMIOBase
	RVMaxPages       = riscv.RVMaxPages
	RVEcallHalt      = riscv.RVEcallHalt
	RVEcallOutput    = riscv.RVEcallOutput
	RVEcallInput     = riscv.RVEcallInput
	RVEcallKeccak256 = riscv.RVEcallKeccak256
	RVEcallSHA256    = riscv.RVEcallSHA256
	RVEcallECRecover = riscv.RVEcallECRecover
)

// RISCV error variables.
var (
	ErrRVInvalidInstruction = riscv.ErrRVInvalidInstruction
	ErrRVGasExhausted       = riscv.ErrRVGasExhausted
	ErrRVHalted             = riscv.ErrRVHalted
	ErrRVMemoryFault        = riscv.ErrRVMemoryFault
	ErrRVEmptyProgram       = riscv.ErrRVEmptyProgram
	ErrRVMemPageLimit       = riscv.ErrRVMemPageLimit
	ErrRVMemUnaligned       = riscv.ErrRVMemUnaligned
	ErrRVMemMMIOWrite       = riscv.ErrRVMemMMIOWrite
	ErrRVMemSegOverlap      = riscv.ErrRVMemSegOverlap
	ErrRVMemSegEmpty        = riscv.ErrRVMemSegEmpty
)

// RISCV function wrappers.
func NewRVCPU(gasLimit uint64) *RVCPU { return riscv.NewRVCPU(gasLimit) }
func ValidateCPUConfig(gasLimit uint64, maxMemoryPages int) error {
	return riscv.ValidateCPUConfig(gasLimit, maxMemoryPages)
}
func NewRVMemory() *RVMemory                     { return riscv.NewRVMemory() }
func NewRVWitnessCollector() *RVWitnessCollector { return riscv.NewRVWitnessCollector() }
func DeserializeWitness(data []byte) (*RVWitnessCollector, error) {
	return riscv.DeserializeWitness(data)
}

// Encode wrappers.
func EncodeRType(opcode, rd, funct3, rs1, rs2, funct7 uint32) uint32 {
	return riscv.EncodeRType(opcode, rd, funct3, rs1, rs2, funct7)
}
func EncodeIType(opcode, rd, funct3, rs1 uint32, imm int32) uint32 {
	return riscv.EncodeIType(opcode, rd, funct3, rs1, imm)
}
func EncodeSType(opcode, funct3, rs1, rs2 uint32, imm int32) uint32 {
	return riscv.EncodeSType(opcode, funct3, rs1, rs2, imm)
}
func EncodeBType(opcode, funct3, rs1, rs2 uint32, imm int32) uint32 {
	return riscv.EncodeBType(opcode, funct3, rs1, rs2, imm)
}
func EncodeUType(opcode, rd uint32, imm uint32) uint32 { return riscv.EncodeUType(opcode, rd, imm) }
func EncodeJType(opcode, rd uint32, imm int32) uint32  { return riscv.EncodeJType(opcode, rd, imm) }
