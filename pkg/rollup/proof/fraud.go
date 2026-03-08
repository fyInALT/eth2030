// Package proof implements fraud proofs, state proofs, trace dispute resolution,
// and cross-layer message proofs for native rollups (EIP-8079).
// fraud.go provides fraud proof generation and verification for optimistic rollups.
package proof

import (
	"encoding/binary"
	"errors"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/crypto"
)

// FraudProofType identifies the kind of fraud being proven.
type FraudProofType uint8

const (
	// InvalidStateRoot indicates the block's post-state root is incorrect.
	InvalidStateRoot FraudProofType = iota + 1

	// InvalidReceipt indicates a transaction receipt is incorrect.
	InvalidReceipt

	// InvalidTransaction indicates a transaction within the block is invalid.
	InvalidTransaction
)

// Fraud proof errors.
var (
	ErrFraudProofNil           = errors.New("fraud_proof: nil fraud proof")
	ErrFraudProofTypeUnknown   = errors.New("fraud_proof: unknown proof type")
	ErrFraudProofPreStateZero  = errors.New("fraud_proof: pre-state root is zero")
	ErrFraudProofPostStateZero = errors.New("fraud_proof: post-state root is zero")
	ErrFraudProofDataEmpty     = errors.New("fraud_proof: proof data is empty")
	ErrFraudProofInvalid       = errors.New("fraud_proof: proof verification failed")
	ErrFraudProofRootsMatch    = errors.New("fraud_proof: pre and post state roots match (no fraud)")
	ErrFraudBlockNumberZero    = errors.New("fraud_proof: block number must be non-zero")
	ErrFraudNilStateReader     = errors.New("fraud_proof: nil state reader function")
	ErrFraudNilTxExecutor      = errors.New("fraud_proof: nil transaction executor function")
	ErrFraudNilStateVerifier   = errors.New("fraud_proof: nil state verifier function")
	ErrFraudTxEmpty            = errors.New("fraud_proof: transaction data is empty")
	ErrBisectionNilClaim       = errors.New("fraud_proof: nil bisection claim")
	ErrBisectionStepIndexMatch = errors.New("fraud_proof: bisection step indices match")
	ErrBisectionConverged      = errors.New("fraud_proof: bisection has converged to single step")
)

// FraudProof represents a proof that a rollup block contains an invalid
// state transition.
type FraudProof struct {
	// Type identifies what kind of fraud is being proven.
	Type FraudProofType

	// BlockNumber is the L2 block number containing the fraud.
	BlockNumber uint64

	// StepIndex is the transaction index within the block where fraud occurs.
	StepIndex uint64

	// PreStateRoot is the state root before the fraudulent step.
	PreStateRoot [32]byte

	// PostStateRoot is the claimed (incorrect) state root after the step.
	PostStateRoot [32]byte

	// ExpectedRoot is the correct post-state root (computed by the challenger).
	ExpectedRoot [32]byte

	// Proof contains the encoded proof data (state witness, tx data, etc.).
	Proof []byte
}

// StateReaderFunc reads state given a root hash, returning the state data.
type StateReaderFunc func(root [32]byte) ([]byte, error)

// TxExecutorFunc executes a transaction against a pre-state, returning
// the resulting post-state root.
type TxExecutorFunc func(preState [32]byte, tx []byte) ([32]byte, error)

// StateVerifierFunc verifies a state transition, returning true if the
// transition from preState to postState is valid given the proof.
type StateVerifierFunc func(preState, postState [32]byte, proof []byte) bool

// FraudProofGenerator generates fraud proofs by comparing expected and actual
// state roots after executing transactions.
type FraudProofGenerator struct {
	stateReader StateReaderFunc
	txExecutor  TxExecutorFunc
}

// NewFraudProofGenerator creates a new fraud proof generator.
func NewFraudProofGenerator(
	stateReader StateReaderFunc,
	txExecutor TxExecutorFunc,
) (*FraudProofGenerator, error) {
	if stateReader == nil {
		return nil, ErrFraudNilStateReader
	}
	if txExecutor == nil {
		return nil, ErrFraudNilTxExecutor
	}
	return &FraudProofGenerator{
		stateReader: stateReader,
		txExecutor:  txExecutor,
	}, nil
}

// GenerateStateRootProof generates a fraud proof for an invalid state root.
func (g *FraudProofGenerator) GenerateStateRootProof(
	blockNumber uint64,
	expectedRoot, actualRoot [32]byte,
) (*FraudProof, error) {
	if blockNumber == 0 {
		return nil, ErrFraudBlockNumberZero
	}
	if expectedRoot == ([32]byte{}) {
		return nil, ErrFraudProofPreStateZero
	}
	if actualRoot == ([32]byte{}) {
		return nil, ErrFraudProofPostStateZero
	}
	if expectedRoot == actualRoot {
		return nil, ErrFraudProofRootsMatch
	}

	stateData, err := g.stateReader(expectedRoot)
	if err != nil {
		return nil, err
	}

	p := buildStateRootProofData(expectedRoot, actualRoot, stateData)

	return &FraudProof{
		Type:          InvalidStateRoot,
		BlockNumber:   blockNumber,
		StepIndex:     0,
		PreStateRoot:  expectedRoot,
		PostStateRoot: actualRoot,
		ExpectedRoot:  expectedRoot,
		Proof:         p,
	}, nil
}

// GenerateSingleStepProof generates a fraud proof for a single transaction
// step within a block.
func (g *FraudProofGenerator) GenerateSingleStepProof(
	blockNumber uint64,
	txIndex uint64,
	preState, postState [32]byte,
	txData []byte,
) (*FraudProof, error) {
	if blockNumber == 0 {
		return nil, ErrFraudBlockNumberZero
	}
	if preState == ([32]byte{}) {
		return nil, ErrFraudProofPreStateZero
	}
	if postState == ([32]byte{}) {
		return nil, ErrFraudProofPostStateZero
	}
	if len(txData) == 0 {
		return nil, ErrFraudTxEmpty
	}

	expectedRoot, err := g.txExecutor(preState, txData)
	if err != nil {
		return nil, err
	}

	if expectedRoot == postState {
		return nil, ErrFraudProofRootsMatch
	}

	p := buildSingleStepProofData(preState, postState, expectedRoot, txData)

	return &FraudProof{
		Type:          InvalidStateRoot,
		BlockNumber:   blockNumber,
		StepIndex:     txIndex,
		PreStateRoot:  preState,
		PostStateRoot: postState,
		ExpectedRoot:  expectedRoot,
		Proof:         p,
	}, nil
}

// FraudProofVerifier verifies fraud proofs submitted by challengers.
type FraudProofVerifier struct {
	stateVerifier StateVerifierFunc
}

// NewFraudProofVerifier creates a new verifier with the given state verifier.
func NewFraudProofVerifier(verifier StateVerifierFunc) (*FraudProofVerifier, error) {
	if verifier == nil {
		return nil, ErrFraudNilStateVerifier
	}
	return &FraudProofVerifier{
		stateVerifier: verifier,
	}, nil
}

// VerifyFraudProof checks whether a fraud proof is valid.
// Returns true if fraud is confirmed (proof is valid), false otherwise.
func (v *FraudProofVerifier) VerifyFraudProof(p *FraudProof) (bool, error) {
	if p == nil {
		return false, ErrFraudProofNil
	}
	if p.Type < InvalidStateRoot || p.Type > InvalidTransaction {
		return false, ErrFraudProofTypeUnknown
	}
	if p.PreStateRoot == ([32]byte{}) {
		return false, ErrFraudProofPreStateZero
	}
	if p.PostStateRoot == ([32]byte{}) {
		return false, ErrFraudProofPostStateZero
	}
	if len(p.Proof) == 0 {
		return false, ErrFraudProofDataEmpty
	}

	if !verifyProofIntegrity(p) {
		return false, ErrFraudProofInvalid
	}

	// The claimed transition (preState -> postState) should be INVALID.
	if v.stateVerifier(p.PreStateRoot, p.PostStateRoot, p.Proof) {
		return false, nil
	}

	return true, nil
}

// ComputeStateTransition executes a single transaction against a pre-state
// root using a deterministic Keccak256 computation.
func ComputeStateTransition(preState [32]byte, tx []byte) ([32]byte, error) {
	if preState == ([32]byte{}) {
		return [32]byte{}, ErrFraudProofPreStateZero
	}
	if len(tx) == 0 {
		return [32]byte{}, ErrFraudTxEmpty
	}

	hash := crypto.Keccak256(preState[:], tx)
	var result [32]byte
	copy(result[:], hash)
	return result, nil
}

// InteractiveVerification implements a multi-round bisection protocol for
// narrowing down the exact step where a fraud occurred.
type InteractiveVerification struct {
	blockNumber     uint64
	startStep       uint64
	endStep         uint64
	claimerRoots    map[uint64][32]byte
	challengerRoots map[uint64][32]byte
	converged       bool
	disputedStep    uint64
}

// NewInteractiveVerification creates a new bisection protocol instance.
func NewInteractiveVerification(blockNumber, startStep, endStep uint64) *InteractiveVerification {
	return &InteractiveVerification{
		blockNumber:     blockNumber,
		startStep:       startStep,
		endStep:         endStep,
		claimerRoots:    make(map[uint64][32]byte),
		challengerRoots: make(map[uint64][32]byte),
	}
}

// IsConverged returns true when the bisection has narrowed to a single step.
func (iv *InteractiveVerification) IsConverged() bool {
	return iv.converged
}

// DisputedStep returns the step index where the dispute was localized.
func (iv *InteractiveVerification) DisputedStep() uint64 {
	return iv.disputedStep
}

// BlockNumber returns the block being disputed.
func (iv *InteractiveVerification) BlockNumber() uint64 {
	return iv.blockNumber
}

// BisectionStep performs one round of the bisection protocol.
func (iv *InteractiveVerification) BisectionStep(
	claimerRoot, challengerRoot [32]byte,
) (uint64, uint64, error) {
	if iv.converged {
		return iv.disputedStep, iv.disputedStep, ErrBisectionConverged
	}

	if iv.endStep <= iv.startStep+1 {
		iv.converged = true
		iv.disputedStep = iv.startStep
		return iv.startStep, iv.endStep, ErrBisectionConverged
	}

	mid := (iv.startStep + iv.endStep) / 2

	iv.claimerRoots[mid] = claimerRoot
	iv.challengerRoots[mid] = challengerRoot

	if claimerRoot == challengerRoot {
		iv.startStep = mid
	} else {
		iv.endStep = mid
	}

	if iv.endStep <= iv.startStep+1 {
		iv.converged = true
		iv.disputedStep = iv.startStep
	}

	return iv.startStep, iv.endStep, nil
}

// GenerateBisectionProof generates a fraud proof once the bisection protocol
// has converged to a single step.
func (iv *InteractiveVerification) GenerateBisectionProof() (*FraudProof, error) {
	if !iv.converged {
		return nil, errors.New("fraud_proof: bisection not yet converged")
	}

	claimerRoot := iv.claimerRoots[iv.disputedStep]
	challengerRoot := iv.challengerRoots[iv.disputedStep]

	var proofData []byte
	var stepBuf [8]byte
	binary.BigEndian.PutUint64(stepBuf[:], iv.disputedStep)
	proofData = append(proofData, stepBuf[:]...)
	proofData = append(proofData, claimerRoot[:]...)
	proofData = append(proofData, challengerRoot[:]...)

	var blockBuf [8]byte
	binary.BigEndian.PutUint64(blockBuf[:], iv.blockNumber)
	commitment := crypto.Keccak256(blockBuf[:], stepBuf[:], claimerRoot[:], challengerRoot[:])
	proofData = append(proofData, commitment...)

	return &FraudProof{
		Type:          InvalidStateRoot,
		BlockNumber:   iv.blockNumber,
		StepIndex:     iv.disputedStep,
		PreStateRoot:  claimerRoot,
		PostStateRoot: challengerRoot,
		Proof:         proofData,
	}, nil
}

// ComputeProofHash computes the Keccak256 hash of a fraud proof.
func ComputeProofHash(p *FraudProof) types.Hash {
	if p == nil {
		return types.Hash{}
	}
	var data []byte
	data = append(data, byte(p.Type))
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], p.BlockNumber)
	data = append(data, buf[:]...)
	binary.BigEndian.PutUint64(buf[:], p.StepIndex)
	data = append(data, buf[:]...)
	data = append(data, p.PreStateRoot[:]...)
	data = append(data, p.PostStateRoot[:]...)
	data = append(data, p.Proof...)
	return crypto.Keccak256Hash(data)
}

// --- Internal helpers ---

func buildStateRootProofData(expectedRoot, actualRoot [32]byte, stateData []byte) []byte {
	stateDataHash := crypto.Keccak256(stateData)
	commitment := crypto.Keccak256(expectedRoot[:], actualRoot[:], stateDataHash)

	result := make([]byte, 0, 128)
	result = append(result, expectedRoot[:]...)
	result = append(result, actualRoot[:]...)
	result = append(result, stateDataHash...)
	result = append(result, commitment...)
	return result
}

func buildSingleStepProofData(preState, postState, expectedRoot [32]byte, txData []byte) []byte {
	txHash := crypto.Keccak256(txData)
	commitment := crypto.Keccak256(preState[:], postState[:], expectedRoot[:], txHash)

	result := make([]byte, 0, 160)
	result = append(result, preState[:]...)
	result = append(result, postState[:]...)
	result = append(result, expectedRoot[:]...)
	result = append(result, txHash...)
	result = append(result, commitment...)
	return result
}

func verifyProofIntegrity(p *FraudProof) bool {
	if len(p.Proof) < 128 {
		return false
	}
	stateDataHash := p.Proof[64:96]
	commitment := p.Proof[96:128]
	recomputed := crypto.Keccak256(p.Proof[0:32], p.Proof[32:64], stateDataHash)
	for i := range commitment {
		if commitment[i] != recomputed[i] {
			return false
		}
	}
	return true
}
