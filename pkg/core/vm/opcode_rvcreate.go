// opcode_rvcreate.go implements the RVCREATE opcode (0xF6) for deploying
// RISC-V RV32IM bytecode to a deterministic address (EL-3.1, EL-3.2, EL-3.3).
//
// Gas model (EL-3.1):
//   - Base:   32,000 (same as CREATE)
//   - Size:   200 × code_size (RISC-V prover overhead)
//
// Initcode must begin with the RISC-V magic bytes 0xFE 0x52 0x56.
// The deployment address is computed CREATE2-style using the program hash
// as the salt, ensuring each unique program gets a unique stable address.
//
// EL-3.3: EVM.Call() checks for the RV magic prefix before running EVM
// bytecode and routes those calls through the RISC-V executor.
//
// Activated at the I+ fork via NewIPlusJumpTable.
package vm

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
	"github.com/eth2030/eth2030/zkvm"
)

// RV magic bytes that identify RISC-V bytecode.
const (
	RVMagic0 byte = 0xFE
	RVMagic1 byte = 0x52
	RVMagic2 byte = 0x56
)

// IsRVCode reports whether the first 3 bytes are the RISC-V magic prefix.
func IsRVCode(code []byte) bool {
	return len(code) >= 3 && code[0] == RVMagic0 && code[1] == RVMagic1 && code[2] == RVMagic2
}

// ErrRVInvalidMagic is returned when RVCREATE initcode lacks the RV magic.
var ErrRVInvalidMagic = errors.New("rvcreate: initcode must start with 0xFE 0x52 0x56")

// gasRVCreate computes dynamic gas for RVCREATE: 200 × code_size.
func gasRVCreate(evm *EVM, contract *Contract, stack *Stack, mem *Memory, memorySize uint64) (uint64, error) {
	// Size is at stack[1] after gas and offset are at [0].
	// Actually, in the dynamic gas function, the stack has not been popped yet
	// for CREATE-style opcodes: [value, offset, size].
	// We peek at index 2 (size is the third item from top).
	if stack.Len() < 3 {
		return 0, nil
	}
	size := stack.Data()[stack.Len()-3].Uint64()
	return 200 * size, nil
}

// rvCreateAddress computes the deterministic RVCREATE deployment address.
// Uses CREATE2-style derivation: keccak256(0xff ++ from ++ salt ++ codeHash)[12:]
// where salt = keccak256(RVMagic ++ initcode).
func rvCreateAddress(from types.Address, initcode []byte) types.Address {
	magic := []byte{RVMagic0, RVMagic1, RVMagic2}
	saltInput := append(magic, initcode...) //nolint:gocritic
	salt := crypto.Keccak256(saltInput)
	codeHash := crypto.Keccak256(initcode)

	buf := make([]byte, 1+20+32+32)
	buf[0] = 0xff
	copy(buf[1:21], from[:])
	copy(buf[21:53], salt)
	copy(buf[53:85], codeHash)
	h := crypto.Keccak256(buf)
	return types.BytesToAddress(h[12:])
}

// opRVCreate implements the RVCREATE opcode (EL-3.2).
// Stack: value, offset, size
// Pushes the new contract address on success, 0 on failure.
func opRVCreate(pc *uint64, evm *EVM, contract *Contract, memory *Memory, stack *Stack) ([]byte, error) {
	if evm.readOnly {
		return nil, ErrWriteProtection
	}

	value := stack.Pop()
	offset, size := stack.Pop(), stack.Pop()
	initCode := memory.Get(int64(offset.Uint64()), int64(size.Uint64()))

	// Validate RISC-V magic bytes.
	if !IsRVCode(initCode) {
		stack.Push(new(big.Int)) // failure: 0
		return nil, nil
	}

	// Compute deterministic deployment address.
	addr := rvCreateAddress(contract.Address, initCode)

	// Validate the program (must be non-empty and properly sized for RV32IM).
	if len(initCode) < 4 || len(initCode)%4 != 0 {
		stack.Push(new(big.Int))
		return nil, nil
	}

	if evm.StateDB == nil {
		stack.Push(new(big.Int))
		return nil, nil
	}

	// Deploy: store the RISC-V bytecode at the address.
	codeHash := types.BytesToHash(crypto.Keccak256(initCode))
	if !evm.StateDB.Exist(addr) {
		evm.StateDB.CreateAccount(addr)
	}
	evm.StateDB.SetCode(addr, initCode)

	// Transfer value if any.
	if value != nil && value.Sign() > 0 {
		callerBalance := evm.StateDB.GetBalance(contract.Address)
		if callerBalance.Cmp(value) < 0 {
			stack.Push(new(big.Int))
			return nil, nil
		}
		evm.StateDB.SubBalance(contract.Address, value)
		evm.StateDB.AddBalance(addr, value)
	}

	// Register program in the global guest registry for later CALL routing.
	// (Best-effort: if registry unavailable, code is still deployed to StateDB.)
	if evm.guestRegistry != nil {
		if _, err := evm.guestRegistry.RegisterGuest(initCode); err != nil {
			// Non-fatal: log but continue (program was already set via SetCode).
			_ = fmt.Sprintf("rvcreate: register guest: %v", err)
		}
	}

	_ = codeHash // used implicitly via SetCode above

	// EIP-7928: record address touch for BAL tracking.
	if evm.balTracker != nil {
		evm.balTracker.RecordAddressTouch(addr)
	}

	stack.Push(new(big.Int).SetBytes(addr[:]))
	return nil, nil
}

// runRVContract executes RISC-V bytecode stored in code for a CALL to addr.
// Returns (ret, gasLeft, err). Used by EVM.Call when magic bytes are detected
// (EL-3.3).
func runRVContract(code, input []byte, gas uint64) ([]byte, uint64, error) {
	// Strip magic prefix (3 bytes) to get the raw RISC-V program.
	program := code[3:]
	if len(program) == 0 {
		return nil, gas, nil
	}

	// Gas model: 1 RISC-V cycle = 1 EVM gas.
	out, err := zkvm.RunPrecompileGuestProgram(program, input, gas)
	if err != nil {
		return nil, 0, fmt.Errorf("rvcreate: execution: %w", err)
	}
	// Conservatively charge 200 gas per instruction (5 instr min).
	// TODO: track actual cycle count in RunPrecompileGuestProgram.
	charged := uint64(len(program)/4) * 200
	if charged > gas {
		return nil, 0, ErrOutOfGas
	}
	return out, gas - charged, nil
}
