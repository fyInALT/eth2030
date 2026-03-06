package zkvm

// r1cs_compat.go re-exports types from zkvm/r1cs for backward compatibility.

import (
	"math/big"

	"github.com/eth2030/eth2030/zkvm/r1cs"
)

// R1CS type aliases.
type (
	SparseTerm       = r1cs.SparseTerm
	SparseConstraint = r1cs.SparseConstraint
	R1CSSystem       = r1cs.R1CSSystem
	R1CSStats        = r1cs.R1CSStats
)

// R1CS error variables.
var (
	ErrR1CSNoVariables       = r1cs.ErrR1CSNoVariables
	ErrR1CSPublicExceedsVars = r1cs.ErrR1CSPublicExceedsVars
	ErrR1CSWitnessSize       = r1cs.ErrR1CSWitnessSize
	ErrR1CSConstraintFailed  = r1cs.ErrR1CSConstraintFailed
	ErrR1CSNoConstraints     = r1cs.ErrR1CSNoConstraints
	ErrR1CSIndexOOB          = r1cs.ErrR1CSIndexOOB
	ErrR1CSPublicInputSize   = r1cs.ErrR1CSPublicInputSize
	ErrR1CSSolveUnsupported  = r1cs.ErrR1CSSolveUnsupported
)

// R1CS function wrappers.
func NewR1CSSystem(numVars, numPublic int) (*R1CSSystem, error) {
	return r1cs.NewR1CSSystem(numVars, numPublic)
}
func NewR1CSSystemWithField(numVars, numPublic int, field *big.Int) (*R1CSSystem, error) {
	return r1cs.NewR1CSSystemWithField(numVars, numPublic, field)
}
