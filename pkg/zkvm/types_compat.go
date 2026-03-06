package zkvm

// types_compat.go re-exports types from zkvm/zkvmtypes for backward compatibility.

import "github.com/eth2030/eth2030/zkvm/zkvmtypes"

// zkVM type aliases.
type (
	GuestProgram    = zkvmtypes.GuestProgram
	VerificationKey = zkvmtypes.VerificationKey
	Proof           = zkvmtypes.Proof
	ProverBackend   = zkvmtypes.ProverBackend
	ExecutionResult = zkvmtypes.ExecutionResult
	GuestInput      = zkvmtypes.GuestInput
)
