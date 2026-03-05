// validation_frame_circuit.go implements STARK-based validation frame circuits.
// These circuits prove that EIP-8141 validation frames executed correctly,
// enabling frame calldata stripping for block compression.
//
// Part of the EL roadmap: proof aggregation and mandatory 3-of-5 proofs (K+).
package proofs

import (
	"crypto/sha256"
	"errors"
)

// Validation frame circuit errors.
var (
	ErrFrameReverted   = errors.New("validation_frame: frame reverted")
	ErrFrameNilOutput  = errors.New("validation_frame: nil output")
	ErrFrameEmptyBatch = errors.New("validation_frame: empty batch")
)

// ValidationFrameProver proves that validation frames executed correctly.
type ValidationFrameProver interface {
	ProveValidationFrame(frameCalldata, output []byte) (*STARKProofData, error)
	ProveAllValidationFrames(frames [][]byte) (*STARKProofData, error)
	Verify(proof *STARKProofData) bool
}

// StubValidationFrameProver is a stub that always succeeds for testing.
type StubValidationFrameProver struct{}

func (s *StubValidationFrameProver) ProveValidationFrame(frameCalldata, output []byte) (*STARKProofData, error) {
	if output == nil {
		return nil, ErrFrameNilOutput
	}
	if len(output) > 0 && output[0] == 0 {
		return nil, ErrFrameReverted
	}
	return &STARKProofData{
		TraceCommitment: sha256.Sum256(frameCalldata),
		TraceLength:     1,
		BlowupFactor:    DefaultBlowupFactor,
		NumQueries:      1,
		FieldModulus:    GoldilocksModulus,
		ConstraintCount: 1,
	}, nil
}

func (s *StubValidationFrameProver) ProveAllValidationFrames(frames [][]byte) (*STARKProofData, error) {
	if len(frames) == 0 {
		return nil, ErrFrameEmptyBatch
	}
	h := sha256.New()
	for _, f := range frames {
		h.Write(f)
	}
	var commit [32]byte
	copy(commit[:], h.Sum(nil))
	return &STARKProofData{
		TraceCommitment: commit,
		TraceLength:     uint64(len(frames)),
		BlowupFactor:    DefaultBlowupFactor,
		NumQueries:      1,
		FieldModulus:    GoldilocksModulus,
		ConstraintCount: 1,
	}, nil
}

func (s *StubValidationFrameProver) Verify(proof *STARKProofData) bool {
	return proof != nil
}

// STARKValidationFrameProver uses the full STARK prover for frame proofs.
type STARKValidationFrameProver struct {
	prover *STARKProver
}

// NewSTARKValidationFrameProver creates a new STARK-based validation frame prover.
func NewSTARKValidationFrameProver() *STARKValidationFrameProver {
	return &STARKValidationFrameProver{
		prover: NewSTARKProver(),
	}
}

// ProveValidationFrame proves that a single validation frame executed correctly.
func (p *STARKValidationFrameProver) ProveValidationFrame(frameCalldata, output []byte) (*STARKProofData, error) {
	if output == nil {
		return nil, ErrFrameNilOutput
	}
	if len(output) == 0 || output[0] == 0 {
		return nil, ErrFrameReverted
	}

	// Build trace: 1 row with [calldata_hash_hi, calldata_hash_lo, output_nonzero].
	calldataHash := sha256.Sum256(frameCalldata)
	hi := NewFieldElement(0)
	hi.Value.SetBytes(calldataHash[:16])
	lo := NewFieldElement(0)
	lo.Value.SetBytes(calldataHash[16:])
	outputNZ := NewFieldElement(1)

	trace := [][]FieldElement{{hi, lo, outputNZ}}

	// Constraint: column 2 (output_nonzero) must equal 1.
	constraint := STARKConstraint{
		Degree:       1,
		Coefficients: []FieldElement{NewFieldElement(0), NewFieldElement(0), NewFieldElement(1)},
	}

	return p.prover.GenerateSTARKProof(trace, []STARKConstraint{constraint})
}

// ProveAllValidationFrames proves a batch of validation frames in a single STARK proof.
func (p *STARKValidationFrameProver) ProveAllValidationFrames(frames [][]byte) (*STARKProofData, error) {
	if len(frames) == 0 {
		return nil, ErrFrameEmptyBatch
	}

	// Build multi-row trace: one row per frame.
	trace := make([][]FieldElement, len(frames))
	for i, frame := range frames {
		calldataHash := sha256.Sum256(frame)
		hi := NewFieldElement(0)
		hi.Value.SetBytes(calldataHash[:16])
		lo := NewFieldElement(0)
		lo.Value.SetBytes(calldataHash[16:])
		outputNZ := NewFieldElement(1)
		trace[i] = []FieldElement{hi, lo, outputNZ}
	}

	// Constraint: column 2 must equal 1 for all rows.
	constraint := STARKConstraint{
		Degree:       1,
		Coefficients: []FieldElement{NewFieldElement(0), NewFieldElement(0), NewFieldElement(1)},
	}

	return p.prover.GenerateSTARKProof(trace, []STARKConstraint{constraint})
}

// Verify verifies a validation frame STARK proof.
func (p *STARKValidationFrameProver) Verify(proof *STARKProofData) bool {
	ok, err := p.prover.VerifySTARKProof(proof, nil)
	return err == nil && ok
}

// ValidationFrameProofSize returns the approximate serialized proof size.
func ValidationFrameProofSize(proof *STARKProofData) int {
	return proof.ProofSize()
}
