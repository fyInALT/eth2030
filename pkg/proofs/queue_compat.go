package proofs

// queue_compat.go re-exports types from proofs/queue for backward compatibility.

import (
	"time"

	"github.com/eth2030/eth2030/proofs/queue"
)

// Queue type aliases.
type (
	QueueProofType        = queue.QueueProofType
	ProofResult           = queue.ProofResult
	ProofQueueConfig      = queue.ProofQueueConfig
	ProofQueue            = queue.ProofQueue
	MandatoryProofTracker = queue.MandatoryProofTracker
	ProofDeadline         = queue.ProofDeadline
)

// Queue constants.
const MandatoryThreshold = queue.MandatoryThreshold

// Queue proof type constants.
const (
	QueueStateProof     = queue.StateProof
	QueueStorageProof   = queue.StorageProof
	QueueExecutionTrace = queue.ExecutionTrace
	QueueWitnessProof   = queue.WitnessProof
	QueueReceiptProof   = queue.ReceiptProof
)

// Queue error variables.
var (
	ErrQueueClosed           = queue.ErrQueueClosed
	ErrQueueFull             = queue.ErrQueueFull
	ErrProofDataEmpty        = queue.ErrProofDataEmpty
	ErrBlockHashZero         = queue.ErrBlockHashZero
	ErrDeadlineExceeded      = queue.ErrDeadlineExceeded
	ErrInvalidQueueProofType = queue.ErrInvalidQueueProofType
)

// Queue variables.
var AllQueueProofTypes = queue.AllQueueProofTypes

// Queue function wrappers.
func DefaultProofQueueConfig() ProofQueueConfig        { return queue.DefaultProofQueueConfig() }
func NewProofQueue(config ProofQueueConfig) *ProofQueue { return queue.NewProofQueue(config) }
func NewMandatoryProofTracker() *MandatoryProofTracker { return queue.NewMandatoryProofTracker() }
func NewProofDeadline(duration time.Duration) *ProofDeadline {
	return queue.NewProofDeadline(duration)
}
func MakeValidProof(blockHash [32]byte, proofType QueueProofType) []byte {
	return queue.MakeValidProof(blockHash, proofType)
}
