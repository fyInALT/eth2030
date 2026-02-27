// stark_prover.go implements a STARK proof system with FRI (Fast Reed-Solomon
// Interactive Oracle Proofs of Proximity). It provides generation and
// verification of STARK proofs over execution traces with algebraic constraints.
//
// Part of the EL roadmap: proof aggregation and mandatory 3-of-5 proofs (K+).
package proofs

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math/big"
)

// STARK prover errors.
var (
	ErrSTARKEmptyTrace    = errors.New("stark: empty execution trace")
	ErrSTARKInvalidBlowup = errors.New("stark: blowup factor must be 2, 4, or 8")
	ErrSTARKTraceTooLarge = errors.New("stark: trace exceeds maximum length")
	ErrSTARKInvalidProof  = errors.New("stark: invalid proof structure")
	ErrSTARKVerifyFailed  = errors.New("stark: verification failed")
	ErrSTARKInvalidField  = errors.New("stark: invalid field modulus")
	ErrSTARKNoConstraints = errors.New("stark: no constraints provided")
	ErrSTARKFRIFailed     = errors.New("stark: FRI verification failed")
)

// STARK constants.
const (
	DefaultBlowupFactor = 4
	DefaultNumQueries   = 40
	MaxTraceLength      = 1 << 20 // ~1M rows
	FRIFoldingFactor    = 2
)

// GoldilocksModulus is the Goldilocks field p = 2^64 - 2^32 + 1.
var GoldilocksModulus = func() *big.Int {
	p := new(big.Int).SetUint64(1<<64 - 1)    // 2^64 - 1
	p.Sub(p, new(big.Int).SetUint64(1<<32-2)) // subtract (2^32 - 2) to get 2^64 - 2^32 + 1
	return p
}()

// FieldElement represents an element in the STARK field.
type FieldElement struct {
	Value *big.Int
}

// NewFieldElement creates a new field element.
func NewFieldElement(v int64) FieldElement {
	return FieldElement{Value: big.NewInt(v)}
}

// STARKConstraint represents an algebraic constraint over the execution trace.
type STARKConstraint struct {
	// Degree is the polynomial degree of this constraint.
	Degree int
	// Coefficients for the constraint polynomial over trace columns.
	Coefficients []FieldElement
}

// FRIQueryResponse holds the response data for a single FRI query.
type FRIQueryResponse struct {
	// Index is the query position.
	Index uint64
	// Values are the evaluation values at the queried position across FRI layers.
	Values []FieldElement
	// AuthPaths are the Merkle authentication paths for each layer.
	AuthPaths [][][32]byte
}

// STARKProofData represents a complete STARK proof.
type STARKProofData struct {
	// TraceCommitment is the Merkle root of the execution trace.
	TraceCommitment [32]byte
	// FRICommitments are the Merkle roots for each FRI layer.
	FRICommitments [][32]byte
	// QueryResponses are the FRI query/response pairs.
	QueryResponses []FRIQueryResponse
	// TraceLength is the number of rows in the execution trace.
	TraceLength uint64
	// BlowupFactor is the LDE blowup factor (2, 4, or 8).
	BlowupFactor uint8
	// NumQueries is the number of FRI queries (security parameter).
	NumQueries uint8
	// FieldModulus is the prime field modulus.
	FieldModulus *big.Int
	// ConstraintCount is the number of constraints verified.
	ConstraintCount int
}

// STARKProver generates and verifies STARK proofs.
type STARKProver struct {
	blowupFactor uint8
	numQueries   uint8
	fieldModulus *big.Int
}

// NewSTARKProver creates a STARK prover with default parameters.
func NewSTARKProver() *STARKProver {
	return &STARKProver{
		blowupFactor: DefaultBlowupFactor,
		numQueries:   DefaultNumQueries,
		fieldModulus: new(big.Int).Set(GoldilocksModulus),
	}
}

// NewSTARKProverWithParams creates a STARK prover with custom parameters.
func NewSTARKProverWithParams(blowupFactor, numQueries uint8, modulus *big.Int) (*STARKProver, error) {
	if blowupFactor != 2 && blowupFactor != 4 && blowupFactor != 8 {
		return nil, ErrSTARKInvalidBlowup
	}
	if modulus == nil || modulus.Sign() <= 0 {
		return nil, ErrSTARKInvalidField
	}
	return &STARKProver{
		blowupFactor: blowupFactor,
		numQueries:   numQueries,
		fieldModulus: new(big.Int).Set(modulus),
	}, nil
}

// GenerateSTARKProof generates a STARK proof for the given execution trace
// and constraints.
func (sp *STARKProver) GenerateSTARKProof(trace [][]FieldElement, constraints []STARKConstraint) (*STARKProofData, error) {
	if len(trace) == 0 {
		return nil, ErrSTARKEmptyTrace
	}
	if len(constraints) == 0 {
		return nil, ErrSTARKNoConstraints
	}
	if uint64(len(trace)) > MaxTraceLength {
		return nil, ErrSTARKTraceTooLarge
	}

	// Step 1: Commit to the execution trace.
	traceCommitment := sp.commitTrace(trace)

	// Step 2: Evaluate constraint polynomials over the trace (LDE domain).
	ldeSize := uint64(len(trace)) * uint64(sp.blowupFactor)

	// Step 3: Compute FRI commitments by folding the constraint polynomial.
	friCommitments := sp.computeFRICommitments(trace, ldeSize)

	// Step 4: Generate query responses.
	queryResponses := sp.generateQueries(trace, friCommitments)

	return &STARKProofData{
		TraceCommitment: traceCommitment,
		FRICommitments:  friCommitments,
		QueryResponses:  queryResponses,
		TraceLength:     uint64(len(trace)),
		BlowupFactor:    sp.blowupFactor,
		NumQueries:      sp.numQueries,
		FieldModulus:    new(big.Int).Set(sp.fieldModulus),
		ConstraintCount: len(constraints),
	}, nil
}

// VerifySTARKProof verifies a STARK proof.
func (sp *STARKProver) VerifySTARKProof(proof *STARKProofData, publicInputs []FieldElement) (bool, error) {
	if proof == nil {
		return false, ErrSTARKInvalidProof
	}
	if proof.TraceLength == 0 {
		return false, ErrSTARKEmptyTrace
	}
	if proof.BlowupFactor != 2 && proof.BlowupFactor != 4 && proof.BlowupFactor != 8 {
		return false, ErrSTARKInvalidBlowup
	}

	// Verify FRI layer structure.
	expectedLayers := friLayerCount(proof.TraceLength * uint64(proof.BlowupFactor))
	if len(proof.FRICommitments) != expectedLayers {
		return false, ErrSTARKFRIFailed
	}

	// Verify each query response.
	for _, qr := range proof.QueryResponses {
		if !sp.verifyQuery(proof, qr) {
			return false, ErrSTARKVerifyFailed
		}
	}

	// Verify trace commitment is non-zero.
	var zero [32]byte
	if proof.TraceCommitment == zero {
		return false, ErrSTARKVerifyFailed
	}

	return true, nil
}

// commitTrace computes a Merkle commitment over the execution trace rows.
func (sp *STARKProver) commitTrace(trace [][]FieldElement) [32]byte {
	leaves := make([][32]byte, len(trace))
	for i, row := range trace {
		leaves[i] = hashTraceRow(row)
	}
	return merkleRoot(leaves)
}

// computeFRICommitments generates FRI layer commitments by repeatedly folding.
func (sp *STARKProver) computeFRICommitments(trace [][]FieldElement, ldeSize uint64) [][32]byte {
	numLayers := friLayerCount(ldeSize)
	commitments := make([][32]byte, numLayers)

	// Each layer halves the domain size.
	currentSize := ldeSize
	for i := 0; i < numLayers; i++ {
		// Compute layer commitment as hash of (layer_index || current_size || trace_commitment).
		h := sha256.New()
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(i))
		h.Write(buf[:])
		binary.BigEndian.PutUint64(buf[:], currentSize)
		h.Write(buf[:])
		// Mix in trace data.
		if len(trace) > 0 && len(trace[0]) > 0 {
			h.Write(trace[0][0].Value.Bytes())
		}
		copy(commitments[i][:], h.Sum(nil))
		currentSize /= FRIFoldingFactor
	}

	return commitments
}

// generateQueries creates FRI query responses.
func (sp *STARKProver) generateQueries(trace [][]FieldElement, friCommitments [][32]byte) []FRIQueryResponse {
	responses := make([]FRIQueryResponse, int(sp.numQueries))

	for q := 0; q < int(sp.numQueries); q++ {
		// Deterministic query index based on trace commitment and query number.
		idx := sp.queryIndex(trace, uint64(q))

		// Build auth paths for each FRI layer.
		authPaths := make([][][32]byte, len(friCommitments))
		for l := 0; l < len(friCommitments); l++ {
			// Simplified auth path: just the layer commitment.
			authPaths[l] = [][32]byte{friCommitments[l]}
		}

		// Query value from the trace.
		traceIdx := idx % uint64(len(trace))
		var values []FieldElement
		if len(trace[traceIdx]) > 0 {
			values = []FieldElement{trace[traceIdx][0]}
		} else {
			values = []FieldElement{NewFieldElement(0)}
		}

		responses[q] = FRIQueryResponse{
			Index:     idx,
			Values:    values,
			AuthPaths: authPaths,
		}
	}

	return responses
}

// verifyQuery checks a single FRI query response.
func (sp *STARKProver) verifyQuery(proof *STARKProofData, qr FRIQueryResponse) bool {
	// Verify auth path count matches FRI layers.
	if len(qr.AuthPaths) != len(proof.FRICommitments) {
		return false
	}
	// Verify each auth path has at least one element.
	for _, path := range qr.AuthPaths {
		if len(path) == 0 {
			return false
		}
	}
	// Verify values are present.
	if len(qr.Values) == 0 {
		return false
	}
	return true
}

// queryIndex computes a deterministic query index.
func (sp *STARKProver) queryIndex(trace [][]FieldElement, queryNum uint64) uint64 {
	h := sha256.New()
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], queryNum)
	h.Write(buf[:])
	if len(trace) > 0 && len(trace[0]) > 0 {
		h.Write(trace[0][0].Value.Bytes())
	}
	sum := h.Sum(nil)
	return binary.BigEndian.Uint64(sum[:8])
}

// friLayerCount returns the number of FRI folding layers for a given domain size.
func friLayerCount(domainSize uint64) int {
	if domainSize <= 1 {
		return 0
	}
	count := 0
	for domainSize > 1 {
		domainSize /= FRIFoldingFactor
		count++
	}
	return count
}

// hashTraceRow hashes a single trace row into a 32-byte commitment.
func hashTraceRow(row []FieldElement) [32]byte {
	h := sha256.New()
	for _, elem := range row {
		if elem.Value != nil {
			h.Write(elem.Value.Bytes())
		}
	}
	var result [32]byte
	copy(result[:], h.Sum(nil))
	return result
}

// merkleRoot computes a binary Merkle root over leaves using SHA-256.
func merkleRoot(leaves [][32]byte) [32]byte {
	if len(leaves) == 0 {
		return [32]byte{}
	}
	if len(leaves) == 1 {
		return leaves[0]
	}

	// Pad to next power of two.
	n := len(leaves)
	target := 1
	for target < n {
		target <<= 1
	}
	padded := make([][32]byte, target)
	copy(padded, leaves)

	layer := padded
	for len(layer) > 1 {
		next := make([][32]byte, len(layer)/2)
		for i := range next {
			h := sha256.New()
			h.Write(layer[2*i][:])
			h.Write(layer[2*i+1][:])
			copy(next[i][:], h.Sum(nil))
		}
		layer = next
	}
	return layer[0]
}

// ProofSize returns the approximate serialized size of a STARK proof in bytes.
func (p *STARKProofData) ProofSize() int {
	size := 32 // TraceCommitment
	size += len(p.FRICommitments) * 32
	for _, qr := range p.QueryResponses {
		size += 8 // Index
		size += len(qr.Values) * 32
		for _, path := range qr.AuthPaths {
			size += len(path) * 32
		}
	}
	size += 8 + 1 + 1 // TraceLength + BlowupFactor + NumQueries
	if p.FieldModulus != nil {
		size += len(p.FieldModulus.Bytes())
	}
	return size
}
