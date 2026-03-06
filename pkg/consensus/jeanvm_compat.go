package consensus

// jeanvm_compat.go re-exports types from consensus/jeanvm for backward compatibility.

import (
	"github.com/eth2030/eth2030/consensus/jeanvm"
	"github.com/eth2030/eth2030/core/types"
)

// JeanVM type aliases.
type (
	AggregationCircuit      = jeanvm.AggregationCircuit
	JeanVMAggregationProof  = jeanvm.JeanVMAggregationProof
	JeanVMAggregator        = jeanvm.JeanVMAggregator
	JeanVMAttestationInput  = jeanvm.JeanVMAttestationInput
	BatchAggregationCircuit = jeanvm.BatchAggregationCircuit
	JeanVMBatchProof        = jeanvm.JeanVMBatchProof
)

// JeanVM error variables.
var (
	ErrJeanVMNoAttestations    = jeanvm.ErrJeanVMNoAttestations
	ErrJeanVMInvalidProof      = jeanvm.ErrJeanVMInvalidProof
	ErrJeanVMCommitteeMismatch = jeanvm.ErrJeanVMCommitteeMismatch
	ErrJeanVMCircuitFailed     = jeanvm.ErrJeanVMCircuitFailed
	ErrJeanVMBatchEmpty        = jeanvm.ErrJeanVMBatchEmpty
)

// JeanVM function wrappers.
func NewAggregationCircuit(committeeSize int, message types.Hash) *AggregationCircuit {
	return jeanvm.NewAggregationCircuit(committeeSize, message)
}
func NewJeanVMAggregator() *JeanVMAggregator { return jeanvm.NewJeanVMAggregator() }
func NewBatchAggregationCircuit(committees [][]JeanVMAttestationInput, messages []types.Hash) *BatchAggregationCircuit {
	return jeanvm.NewBatchAggregationCircuit(committees, messages)
}
func ValidateAggregationProof(proof *JeanVMAggregationProof) error {
	return jeanvm.ValidateAggregationProof(proof)
}
func ValidateBatchAggregationProof(proof *JeanVMBatchProof) error {
	return jeanvm.ValidateBatchAggregationProof(proof)
}
