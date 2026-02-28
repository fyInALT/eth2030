// trace_dispute.go implements execution trace building and dispute resolution
// for fraud proofs in native rollups. Complements fraud_proof.go with
// step-level tracing and binary-search dispute resolution.
package rollup

import (
	"encoding/binary"
	"errors"
	"sort"

	"github.com/eth2030/eth2030/crypto"
)

// Trace dispute errors.
var (
	ErrTraceEmpty        = errors.New("trace_dispute: execution trace is empty")
	ErrTraceStepNil      = errors.New("trace_dispute: nil execution step")
	ErrTraceHashMismatch = errors.New("trace_dispute: step output does not match next step input")
	ErrDisputeNoInvalid  = errors.New("trace_dispute: no invalid step found")
	ErrDisputeMaxRounds  = errors.New("trace_dispute: exceeded maximum resolution rounds")
	ErrDisputeConverged  = errors.New("trace_dispute: dispute already resolved")
	ErrDisputeIdxRange   = errors.New("trace_dispute: index out of trace range")
	ErrTraceBlockZero    = errors.New("trace_dispute: block number must be non-zero")
	ErrTraceRootMismatch = errors.New("trace_dispute: pre/post state root mismatch in proof")
	ErrStepVerifyFailed  = errors.New("trace_dispute: step verification failed")
	ErrBisectSingleStep  = errors.New("trace_dispute: cannot bisect a single-step trace")
)

// ExecutionStep captures a single opcode execution with state transition hashes.
type ExecutionStep struct {
	// Index is the position of this step in the trace (0-based).
	Index uint64

	// Opcode is the EVM opcode executed at this step.
	Opcode byte

	// GasUsed is the gas consumed by this step.
	GasUsed uint64

	// InputStateHash is the state hash before this step executes.
	InputStateHash [32]byte

	// OutputStateHash is the state hash after this step executes.
	OutputStateHash [32]byte

	// StackHash is a hash of the EVM stack at this step (for debugging).
	StackHash [32]byte
}

// Hash computes a deterministic hash of the execution step.
func (s *ExecutionStep) Hash() [32]byte {
	if s == nil {
		return [32]byte{}
	}
	var data []byte
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], s.Index)
	data = append(data, buf[:]...)
	data = append(data, s.Opcode)
	binary.BigEndian.PutUint64(buf[:], s.GasUsed)
	data = append(data, buf[:]...)
	data = append(data, s.InputStateHash[:]...)
	data = append(data, s.OutputStateHash[:]...)
	data = append(data, s.StackHash[:]...)
	h := crypto.Keccak256(data)
	var result [32]byte
	copy(result[:], h)
	return result
}

// IsValid checks whether the step has non-zero input and output state hashes.
func (s *ExecutionStep) IsValid() bool {
	if s == nil {
		return false
	}
	return s.InputStateHash != ([32]byte{}) && s.OutputStateHash != ([32]byte{})
}

// ExecutionTrace is an ordered list of execution steps for a rollup block.
type ExecutionTrace struct {
	// BlockNumber is the rollup block this trace belongs to.
	BlockNumber uint64

	// Steps contains the ordered execution steps.
	Steps []ExecutionStep

	// PreStateRoot is the state root before trace execution begins.
	PreStateRoot [32]byte

	// PostStateRoot is the state root after trace execution completes.
	PostStateRoot [32]byte
}

// NewExecutionTrace creates a new execution trace from the given steps.
func NewExecutionTrace(blockNum uint64, steps []ExecutionStep, preRoot, postRoot [32]byte) *ExecutionTrace {
	stepsCopy := make([]ExecutionStep, len(steps))
	copy(stepsCopy, steps)
	return &ExecutionTrace{
		BlockNumber:   blockNum,
		Steps:         stepsCopy,
		PreStateRoot:  preRoot,
		PostStateRoot: postRoot,
	}
}

// Len returns the number of steps in the trace.
func (et *ExecutionTrace) Len() int {
	if et == nil {
		return 0
	}
	return len(et.Steps)
}

// Hash computes a Merkle root of all step hashes in the trace.
func (et *ExecutionTrace) Hash() [32]byte {
	if et == nil || len(et.Steps) == 0 {
		return [32]byte{}
	}

	leaves := make([][32]byte, len(et.Steps))
	for i := range et.Steps {
		leaves[i] = et.Steps[i].Hash()
	}

	// Build binary Merkle tree.
	for len(leaves) > 1 {
		var next [][32]byte
		for i := 0; i < len(leaves); i += 2 {
			if i+1 < len(leaves) {
				h := crypto.Keccak256(leaves[i][:], leaves[i+1][:])
				var node [32]byte
				copy(node[:], h)
				next = append(next, node)
			} else {
				// Odd leaf: hash with itself.
				h := crypto.Keccak256(leaves[i][:], leaves[i][:])
				var node [32]byte
				copy(node[:], h)
				next = append(next, node)
			}
		}
		leaves = next
	}
	return leaves[0]
}

// VerifyChain checks that each step's output hash matches the next step's
// input hash, forming a valid chain. Also verifies that the first step's
// input matches PreStateRoot and the last step's output matches PostStateRoot.
func (et *ExecutionTrace) VerifyChain() error {
	if et == nil || len(et.Steps) == 0 {
		return ErrTraceEmpty
	}

	// Verify first step links to pre-state root.
	if et.Steps[0].InputStateHash != et.PreStateRoot {
		return ErrTraceRootMismatch
	}

	// Verify step chain continuity.
	for i := 0; i < len(et.Steps)-1; i++ {
		if et.Steps[i].OutputStateHash != et.Steps[i+1].InputStateHash {
			return ErrTraceHashMismatch
		}
	}

	// Verify last step links to post-state root.
	if et.Steps[len(et.Steps)-1].OutputStateHash != et.PostStateRoot {
		return ErrTraceRootMismatch
	}

	return nil
}

// SubTrace returns a sub-trace from index start (inclusive) to end (exclusive).
func (et *ExecutionTrace) SubTrace(start, end int) (*ExecutionTrace, error) {
	if et == nil || len(et.Steps) == 0 {
		return nil, ErrTraceEmpty
	}
	if start < 0 || end > len(et.Steps) || start >= end {
		return nil, ErrDisputeIdxRange
	}

	subSteps := make([]ExecutionStep, end-start)
	copy(subSteps, et.Steps[start:end])

	preRoot := subSteps[0].InputStateHash
	postRoot := subSteps[len(subSteps)-1].OutputStateHash

	return &ExecutionTrace{
		BlockNumber:   et.BlockNumber,
		Steps:         subSteps,
		PreStateRoot:  preRoot,
		PostStateRoot: postRoot,
	}, nil
}

// TraceFraudProof pinpoints the exact invalid step in an execution trace.
type TraceFraudProof struct {
	// BlockNumber is the L2 block containing the fraud.
	BlockNumber uint64

	// PreStateRoot is the state root before trace execution.
	PreStateRoot [32]byte

	// ClaimedPostRoot is the (incorrect) claimed post-state root.
	ClaimedPostRoot [32]byte

	// ActualPostRoot is the correct post-state root.
	ActualPostRoot [32]byte

	// InvalidStepIndex is the index of the first invalid step.
	InvalidStepIndex int

	// InvalidStep is a copy of the invalid execution step.
	InvalidStep ExecutionStep

	// TraceHash is the Merkle root of the full execution trace.
	TraceHash [32]byte
}

// TraceProofBuilder generates fraud proofs from execution traces.
type TraceProofBuilder struct {
	// verifyStep is a function that verifies a single execution step.
	// Returns true if the step's OutputStateHash is correct.
	verifyStep func(step *ExecutionStep) bool
}

// NewTraceProofBuilder creates a new builder with the given step verifier.
// If verifyStep is nil, a default hash-based verifier is used.
func NewTraceProofBuilder(verifyStep func(*ExecutionStep) bool) *TraceProofBuilder {
	if verifyStep == nil {
		verifyStep = defaultStepVerifier
	}
	return &TraceProofBuilder{verifyStep: verifyStep}
}

// GenerateProof builds a fraud proof by finding the first invalid step
// in the trace and recording the context.
func (b *TraceProofBuilder) GenerateProof(
	blockNum uint64, preStateRoot, claimedPostRoot, actualPostRoot [32]byte,
	trace *ExecutionTrace,
) (*TraceFraudProof, error) {
	if blockNum == 0 {
		return nil, ErrTraceBlockZero
	}
	if trace == nil || len(trace.Steps) == 0 {
		return nil, ErrTraceEmpty
	}
	if claimedPostRoot == actualPostRoot {
		return nil, ErrFraudProofRootsMatch
	}

	invalidIdx, err := b.FindInvalidStep(trace)
	if err != nil {
		return nil, err
	}

	return &TraceFraudProof{
		BlockNumber:      blockNum,
		PreStateRoot:     preStateRoot,
		ClaimedPostRoot:  claimedPostRoot,
		ActualPostRoot:   actualPostRoot,
		InvalidStepIndex: invalidIdx,
		InvalidStep:      trace.Steps[invalidIdx],
		TraceHash:        trace.Hash(),
	}, nil
}

// ValidateProof checks a TraceFraudProof for internal consistency.
func (b *TraceProofBuilder) ValidateProof(proof *TraceFraudProof) bool {
	if proof == nil {
		return false
	}
	if proof.BlockNumber == 0 {
		return false
	}
	if proof.ClaimedPostRoot == proof.ActualPostRoot {
		return false
	}
	if proof.PreStateRoot == ([32]byte{}) {
		return false
	}
	if proof.TraceHash == ([32]byte{}) {
		return false
	}
	// Verify the invalid step itself is flagged as invalid by the verifier.
	return !b.verifyStep(&proof.InvalidStep)
}

// FindInvalidStep uses a linear scan to find the first step whose output
// does not match the verifier's expectation.
func (b *TraceProofBuilder) FindInvalidStep(trace *ExecutionTrace) (int, error) {
	if trace == nil || len(trace.Steps) == 0 {
		return -1, ErrTraceEmpty
	}

	for i := range trace.Steps {
		if !b.verifyStep(&trace.Steps[i]) {
			return i, nil
		}
	}
	return -1, ErrDisputeNoInvalid
}

// DisputeResolver binary-searches for the first invalid execution step.
type DisputeResolver struct {
	maxRounds  int
	verifyStep func(step *ExecutionStep) bool
	resolved   bool
	resolvedAt int
}

// NewDisputeResolver creates a resolver with the given max rounds and verifier.
func NewDisputeResolver(maxRounds int, verifyStep func(*ExecutionStep) bool) *DisputeResolver {
	if maxRounds <= 0 {
		maxRounds = 64
	}
	if verifyStep == nil {
		verifyStep = defaultStepVerifier
	}
	return &DisputeResolver{
		maxRounds:  maxRounds,
		verifyStep: verifyStep,
		resolvedAt: -1,
	}
}

// BisectTrace splits a trace into two halves around the midpoint.
// Returns the left and right sub-traces and the midpoint index.
func (dr *DisputeResolver) BisectTrace(
	trace *ExecutionTrace, claimedInvalidIdx int,
) (left, right *ExecutionTrace, midIdx int, err error) {
	if trace == nil || len(trace.Steps) == 0 {
		return nil, nil, 0, ErrTraceEmpty
	}
	if len(trace.Steps) < 2 {
		return nil, nil, 0, ErrBisectSingleStep
	}

	midIdx = len(trace.Steps) / 2

	left, err = trace.SubTrace(0, midIdx)
	if err != nil {
		return nil, nil, 0, err
	}
	right, err = trace.SubTrace(midIdx, len(trace.Steps))
	if err != nil {
		return nil, nil, 0, err
	}

	return left, right, midIdx, nil
}

// VerifyStep checks a single execution step using the verifier function.
func (dr *DisputeResolver) VerifyStep(step *ExecutionStep) bool {
	if step == nil {
		return false
	}
	return dr.verifyStep(step)
}

// Resolve binary-searches the trace to find the first invalid step.
func (dr *DisputeResolver) Resolve(trace *ExecutionTrace) (int, error) {
	if trace == nil || len(trace.Steps) == 0 {
		return -1, ErrTraceEmpty
	}
	if dr.resolved {
		return dr.resolvedAt, ErrDisputeConverged
	}

	lo := 0
	hi := len(trace.Steps)
	rounds := 0

	for lo < hi && rounds < dr.maxRounds {
		if hi-lo == 1 {
			// Converged to a single step.
			if !dr.verifyStep(&trace.Steps[lo]) {
				dr.resolved = true
				dr.resolvedAt = lo
				return lo, nil
			}
			return -1, ErrDisputeNoInvalid
		}

		mid := (lo + hi) / 2

		// Check if the left half contains an invalid step by checking
		// whether any step in [lo, mid) fails verification.
		leftHasInvalid := false
		for i := lo; i < mid; i++ {
			if !dr.verifyStep(&trace.Steps[i]) {
				leftHasInvalid = true
				break
			}
		}

		if leftHasInvalid {
			hi = mid
		} else {
			lo = mid
		}
		rounds++
	}

	if rounds >= dr.maxRounds {
		return -1, ErrDisputeMaxRounds
	}
	return -1, ErrDisputeNoInvalid
}

// IsResolved returns whether the dispute has been resolved.
func (dr *DisputeResolver) IsResolved() bool {
	return dr.resolved
}

// ResolvedIndex returns the index of the resolved invalid step, or -1.
func (dr *DisputeResolver) ResolvedIndex() int {
	return dr.resolvedAt
}

// --- Helpers ---

// defaultStepVerifier checks output = Keccak256(input || opcode || gas).
func defaultStepVerifier(step *ExecutionStep) bool {
	if step == nil {
		return false
	}
	expected := computeExpectedOutput(step)
	return step.OutputStateHash == expected
}

// computeExpectedOutput derives the expected output state hash from step params.
func computeExpectedOutput(step *ExecutionStep) [32]byte {
	var data []byte
	data = append(data, step.InputStateHash[:]...)
	data = append(data, step.Opcode)
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], step.GasUsed)
	data = append(data, buf[:]...)
	h := crypto.Keccak256(data)
	var result [32]byte
	copy(result[:], h)
	return result
}

// BuildValidStep creates a valid step with correctly derived output hash.
func BuildValidStep(index uint64, opcode byte, gas uint64, inputState [32]byte) ExecutionStep {
	step := ExecutionStep{
		Index:          index,
		Opcode:         opcode,
		GasUsed:        gas,
		InputStateHash: inputState,
		StackHash:      crypto.Keccak256Hash(inputState[:], []byte{opcode}),
	}
	step.OutputStateHash = computeExpectedOutput(&step)
	return step
}

// BuildValidTrace constructs a valid trace of n chained steps.
func BuildValidTrace(blockNum uint64, n int, initialState [32]byte) *ExecutionTrace {
	if n <= 0 {
		return nil
	}
	steps := make([]ExecutionStep, n)
	currentState := initialState
	for i := 0; i < n; i++ {
		opcode := byte(i % 256)
		gas := uint64(21000 + i*100)
		steps[i] = BuildValidStep(uint64(i), opcode, gas, currentState)
		currentState = steps[i].OutputStateHash
	}
	return NewExecutionTrace(blockNum, steps, initialState, currentState)
}

// SortStepsByGas returns a copy of steps sorted by gas used ascending.
func SortStepsByGas(steps []ExecutionStep) []ExecutionStep {
	sorted := make([]ExecutionStep, len(steps))
	copy(sorted, steps)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GasUsed < sorted[j].GasUsed
	})
	return sorted
}

// TotalGasUsed returns the total gas consumed across all steps in a trace.
func TotalGasUsed(trace *ExecutionTrace) uint64 {
	if trace == nil {
		return 0
	}
	var total uint64
	for i := range trace.Steps {
		total += trace.Steps[i].GasUsed
	}
	return total
}
