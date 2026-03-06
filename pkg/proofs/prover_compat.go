package proofs

// prover_compat.go re-exports types from proofs/prover for backward compatibility.

import (
	"math/big"

	"github.com/eth2030/eth2030/proofs/prover"
)

// Prover type aliases.
type (
	FieldElement               = prover.FieldElement
	STARKConstraint            = prover.STARKConstraint
	FRIQueryResponse           = prover.FRIQueryResponse
	STARKProofData             = prover.STARKProofData
	STARKProver                = prover.STARKProver
	ProverCandidate            = prover.ProverCandidate
	AssignmentResult           = prover.AssignmentResult
	ReputationScorer           = prover.ReputationScorer
	ProverPool                 = prover.ProverPool
	ValidationFrameProver      = prover.ValidationFrameProver
	StubValidationFrameProver  = prover.StubValidationFrameProver
	STARKValidationFrameProver = prover.STARKValidationFrameProver
)

// Prover constants.
const (
	DefaultBlowupFactor = prover.DefaultBlowupFactor
	DefaultNumQueries   = prover.DefaultNumQueries
	MaxTraceLength      = prover.MaxTraceLength
	FRIFoldingFactor    = prover.FRIFoldingFactor
)

// Prover error variables.
var (
	ErrSTARKEmptyTrace     = prover.ErrSTARKEmptyTrace
	ErrSTARKInvalidBlowup  = prover.ErrSTARKInvalidBlowup
	ErrSTARKTraceTooLarge  = prover.ErrSTARKTraceTooLarge
	ErrSTARKInvalidProof   = prover.ErrSTARKInvalidProof
	ErrSTARKVerifyFailed   = prover.ErrSTARKVerifyFailed
	ErrSTARKInvalidField   = prover.ErrSTARKInvalidField
	ErrSTARKNoConstraints  = prover.ErrSTARKNoConstraints
	ErrSTARKFRIFailed      = prover.ErrSTARKFRIFailed
	ErrProverAlreadyExists = prover.ErrProverAlreadyExists
	ErrProverNotRegistered = prover.ErrProverNotRegistered
	ErrProverAtCapacity    = prover.ErrProverAtCapacity
	ErrFrameReverted       = prover.ErrFrameReverted
	ErrFrameNilOutput      = prover.ErrFrameNilOutput
	ErrFrameEmptyBatch     = prover.ErrFrameEmptyBatch
)

// GoldilocksModulus is the Goldilocks field modulus.
var GoldilocksModulus = prover.GoldilocksModulus

// Prover function wrappers.
func NewFieldElement(v int64) FieldElement { return prover.NewFieldElement(v) }
func NewSTARKProver() *STARKProver         { return prover.NewSTARKProver() }
func NewSTARKProverWithParams(blowupFactor, numQueries uint8, modulus *big.Int) (*STARKProver, error) {
	return prover.NewSTARKProverWithParams(blowupFactor, numQueries, modulus)
}
func NewReputationScorer() *ReputationScorer { return prover.NewReputationScorer() }
func NewProverPool(minProvers int) *ProverPool { return prover.NewProverPool(minProvers) }
func NewSTARKValidationFrameProver() *STARKValidationFrameProver {
	return prover.NewSTARKValidationFrameProver()
}
func ValidationFrameProofSize(proof *STARKProofData) int {
	return prover.ValidationFrameProofSize(proof)
}
