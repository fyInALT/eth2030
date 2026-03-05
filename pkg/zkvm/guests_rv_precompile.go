// guests_precompile.go provides RISC-V guest programs for the three highest-frequency
// EVM precompiles (Keccak256, SHA-256, ECRECOVER) and the corresponding ECALL
// handlers that implement them in Go.
//
// Each program is minimal RV32IM machine code that:
//  1. Issues the crypto ECALL (codes 3–5) — the Go handler consumes all InputBuf
//     bytes and appends the digest/address to OutputBuf.
//  2. Issues ECALL 0 (Halt) with exit code 0.
//
// Gas model (EL-2.3): routed through CanonicalGuestPrecompile, which charges
// the standard Go precompile gas cost.  RISC-V cycle overhead is absorbed by
// the surrounding gas estimation; no per-cycle surcharge at this tier.
//
// Part of the I+ roadmap (US-EL-2: RISC-V Precompile Guest Programs).
package zkvm

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/eth2030/eth2030/crypto"
	"golang.org/x/crypto/sha3"
)

// RV32IM machine code for the three precompile guest programs.
// Encoding: little-endian 4-byte instructions.
//
//	addi a7, x0, <ecall_code>   ; set syscall number
//	ecall                       ; invoke crypto handler
//	addi a7, x0, 0              ; syscall = RVEcallHalt
//	addi a0, x0, 0              ; exit code = 0
//	ecall                       ; halt
var (
	// RVKeccak256Program is the Keccak-256 guest (ECALL 3).
	RVKeccak256Program = []byte{
		0x93, 0x08, 0x30, 0x00, // addi a7, x0, 3
		0x73, 0x00, 0x00, 0x00, // ecall
		0x93, 0x08, 0x00, 0x00, // addi a7, x0, 0
		0x13, 0x05, 0x00, 0x00, // addi a0, x0, 0
		0x73, 0x00, 0x00, 0x00, // ecall (halt)
	}

	// RVSHA256Program is the SHA-256 guest (ECALL 4).
	RVSHA256Program = []byte{
		0x93, 0x08, 0x40, 0x00, // addi a7, x0, 4
		0x73, 0x00, 0x00, 0x00, // ecall
		0x93, 0x08, 0x00, 0x00, // addi a7, x0, 0
		0x13, 0x05, 0x00, 0x00, // addi a0, x0, 0
		0x73, 0x00, 0x00, 0x00, // ecall (halt)
	}

	// RVECRecoverProgram is the secp256k1 ECRECOVER guest (ECALL 5).
	// Input: 128 bytes (hash[32] || v[32] || r[32] || s[32]).
	// Output: 32 bytes padded address (left-zero-padded, right-aligned).
	RVECRecoverProgram = []byte{
		0x93, 0x08, 0x50, 0x00, // addi a7, x0, 5
		0x73, 0x00, 0x00, 0x00, // ecall
		0x93, 0x08, 0x00, 0x00, // addi a7, x0, 0
		0x13, 0x05, 0x00, 0x00, // addi a0, x0, 0
		0x73, 0x00, 0x00, 0x00, // ecall (halt)
	}
)

// ecrecover input layout (matches Ethereum precompile convention).
const (
	ecrecoverHashOffset = 0
	ecrecoverVOffset    = 32
	ecrecoverROffset    = 64
	ecrecoverSOffset    = 96
	ecrecoverInputSize  = 128
)

// Keccak256EcallHandler implements ECALL 3: hash all remaining InputBuf bytes
// with Keccak-256 and append the 32-byte digest to OutputBuf.
func Keccak256EcallHandler(cpu *RVCPU) error {
	data := cpu.InputBuf[cpu.inputPos:]
	h := sha3.NewLegacyKeccak256()
	h.Write(data)
	cpu.OutputBuf = h.Sum(cpu.OutputBuf)
	cpu.inputPos = len(cpu.InputBuf)
	return nil
}

// SHA256EcallHandler implements ECALL 4: hash all remaining InputBuf bytes
// with SHA-256 and append the 32-byte digest to OutputBuf.
func SHA256EcallHandler(cpu *RVCPU) error {
	data := cpu.InputBuf[cpu.inputPos:]
	digest := sha256.Sum256(data)
	cpu.OutputBuf = append(cpu.OutputBuf, digest[:]...)
	cpu.inputPos = len(cpu.InputBuf)
	return nil
}

// ECRecoverEcallHandler implements ECALL 5: recover a secp256k1 public-key
// address from a signature and append the 32-byte (12-byte zero pad + 20-byte
// address) result to OutputBuf.
//
// Input layout: hash[32] || v[32] || r[32] || s[32] (128 bytes total).
// v must be 27 or 28 (Ethereum convention); the handler converts to 0/1
// for the recovery index internally.
func ECRecoverEcallHandler(cpu *RVCPU) error {
	data := cpu.InputBuf[cpu.inputPos:]
	if len(data) < ecrecoverInputSize {
		return fmt.Errorf("riscv ecrecover: need %d bytes, got %d", ecrecoverInputSize, len(data))
	}
	hash := data[ecrecoverHashOffset : ecrecoverHashOffset+32]
	v := data[ecrecoverVOffset : ecrecoverVOffset+32]
	r := data[ecrecoverROffset : ecrecoverROffset+32]
	s := data[ecrecoverSOffset : ecrecoverSOffset+32]

	// Build the 65-byte compact signature: r[32] || s[32] || v_byte[1].
	// v is big-endian 32-byte; the recovery byte is the last byte.
	vByte := v[31]
	if vByte < 27 {
		return errors.New("riscv ecrecover: v must be 27 or 28")
	}
	vByte -= 27

	sig := make([]byte, 65)
	copy(sig[0:32], r)
	copy(sig[32:64], s)
	sig[64] = vByte

	// Recover public key bytes (uncompressed: 0x04 prefix + 64 bytes).
	pub, err := crypto.Ecrecover(hash, sig)
	if err != nil || len(pub) < 65 {
		// No address recovered; output 32 zero bytes (Ethereum convention).
		cpu.OutputBuf = append(cpu.OutputBuf, make([]byte, 32)...)
		cpu.inputPos += ecrecoverInputSize
		return nil
	}
	// Derive address: Keccak256(pubkey[1:])[12:32].
	addrHash := crypto.Keccak256(pub[1:])
	var out [32]byte
	copy(out[12:], addrHash[12:])
	cpu.OutputBuf = append(cpu.OutputBuf, out[:]...)
	cpu.inputPos += ecrecoverInputSize
	return nil
}

// DefaultPrecompileEcallHandlers returns the standard map of crypto ECALL
// handlers (codes 3–5). Wire these into an RVCPU before running a precompile
// guest program.
func DefaultPrecompileEcallHandlers() map[uint32]EcallHandler {
	return map[uint32]EcallHandler{
		RVEcallKeccak256: Keccak256EcallHandler,
		RVEcallSHA256:    SHA256EcallHandler,
		RVEcallECRecover: ECRecoverEcallHandler,
	}
}

// NewPrecompileGuestRegistry creates a GuestRegistry pre-loaded with the
// three standard precompile guest programs (Keccak256, SHA256, ECRecover).
func NewPrecompileGuestRegistry() (*GuestRegistry, error) {
	r := NewGuestRegistry()
	for _, prog := range [][]byte{RVKeccak256Program, RVSHA256Program, RVECRecoverProgram} {
		if _, err := r.RegisterGuest(prog); err != nil {
			return nil, fmt.Errorf("zkvm: register precompile guest: %w", err)
		}
	}
	return r, nil
}

// RunPrecompileGuestProgram executes a raw RISC-V program (arbitrary bytecode,
// not identified by an ECALL precompile code) with the default crypto handlers.
// The RV magic prefix (0xFE 0x52 0x56) must already be stripped by the caller.
func RunPrecompileGuestProgram(program, input []byte, gasLimit uint64) ([]byte, error) {
	if len(program) == 0 {
		return nil, nil
	}
	cpu := NewRVCPU(gasLimit)
	for code, h := range DefaultPrecompileEcallHandlers() {
		cpu.RegisterEcallHandler(code, h)
	}
	cpu.InputBuf = append([]byte(nil), input...)
	if err := cpu.LoadProgram(program, 0, 0); err != nil {
		return nil, fmt.Errorf("riscv: load program: %w", err)
	}
	if err := cpu.Run(); err != nil && cpu.ExitCode != 0 {
		return nil, fmt.Errorf("riscv: execution error (exit %d): %w", cpu.ExitCode, err)
	}
	return cpu.OutputBuf, nil
}

// RunPrecompileGuest executes a precompile guest program (identified by its
// ECALL code) on the given input and returns the output bytes.
// It wires the three default crypto handlers automatically.
func RunPrecompileGuest(ecallCode uint32, input []byte, gasLimit uint64) ([]byte, error) {
	var prog []byte
	switch ecallCode {
	case RVEcallKeccak256:
		prog = RVKeccak256Program
	case RVEcallSHA256:
		prog = RVSHA256Program
	case RVEcallECRecover:
		prog = RVECRecoverProgram
	default:
		return nil, fmt.Errorf("riscv: unknown precompile ecall code %d", ecallCode)
	}

	cpu := NewRVCPU(gasLimit)
	for code, h := range DefaultPrecompileEcallHandlers() {
		cpu.RegisterEcallHandler(code, h)
	}
	cpu.InputBuf = append([]byte(nil), input...)

	if err := cpu.LoadProgram(prog, 0, 0); err != nil {
		return nil, fmt.Errorf("riscv: load program: %w", err)
	}
	if err := cpu.Run(); err != nil && cpu.ExitCode != 0 {
		return nil, fmt.Errorf("riscv: execution error (exit %d): %w", cpu.ExitCode, err)
	}
	return cpu.OutputBuf, nil
}
