package consensus

// prequorum_compat.go re-exports types from consensus/prequorum for backward compatibility.

import (
	"github.com/eth2030/eth2030/consensus/prequorum"
	"github.com/eth2030/eth2030/core/types"
)

// Prequorum type aliases.
type (
	PrequorumConfig       = prequorum.PrequorumConfig
	Preconfirmation       = prequorum.Preconfirmation
	PrequorumStatus       = prequorum.PrequorumStatus
	PrequorumEngine       = prequorum.PrequorumEngine
	SecurePrequorumConfig = prequorum.SecurePrequorumConfig
	SecurePrequorumVote   = prequorum.SecurePrequorumVote
	SecureVoteReveal      = prequorum.SecureVoteReveal
	SecureQuorumStatus    = prequorum.SecureQuorumStatus
	SecurePrequorumState  = prequorum.SecurePrequorumState
)

// Prequorum error aliases.
var (
	ErrNilPreconfirmation   = prequorum.ErrNilPreconfirmation
	ErrPrequorumInvalidSlot = prequorum.ErrPrequorumInvalidSlot
	ErrEmptySignature       = prequorum.ErrEmptySignature
	ErrEmptyTxHash          = prequorum.ErrEmptyTxHash
	ErrEmptyCommitment      = prequorum.ErrEmptyCommitment
	ErrSlotFull             = prequorum.ErrSlotFull
	ErrDuplicatePreconf     = prequorum.ErrDuplicatePreconf
	ErrInvalidCommitment    = prequorum.ErrInvalidCommitment
)

// Prequorum constant aliases.
const DefaultQuorumThreshold = prequorum.DefaultQuorumThreshold

// Prequorum function wrappers.
func DefaultPrequorumConfig() PrequorumConfig               { return prequorum.DefaultPrequorumConfig() }
func NewPrequorumEngine(c PrequorumConfig) *PrequorumEngine { return prequorum.NewPrequorumEngine(c) }
func ComputeCommitment(slot, vi uint64, tx types.Hash) types.Hash {
	return prequorum.ComputeCommitment(slot, vi, tx)
}
func DefaultSecurePrequorumConfig() SecurePrequorumConfig {
	return prequorum.DefaultSecurePrequorumConfig()
}
func NewSecurePrequorumState(c SecurePrequorumConfig) *SecurePrequorumState {
	return prequorum.NewSecurePrequorumState(c)
}
func VerifyVoteCommitment(vote *SecurePrequorumVote) bool {
	return prequorum.VerifyVoteCommitment(vote)
}
func ComputeVRFWeight(vrfProof []byte, vss uint64) float64 {
	return prequorum.ComputeVRFWeight(vrfProof, vss)
}
func ComputeSecureCommitment(slot, vi uint64, root types.Hash, vrf []byte) types.Hash {
	return prequorum.ComputeSecureCommitment(slot, vi, root, vrf)
}
