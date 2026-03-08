package consensus

// voting_compat.go re-exports types from consensus/voting for backward compatibility.

import (
	"github.com/eth2030/eth2030/consensus/voting"
	"github.com/eth2030/eth2030/core/types"
)

// Voting type aliases.
type (
	Vote                = voting.Vote
	VotingRound         = voting.VotingRound
	VotingManagerConfig = voting.VotingManagerConfig
	VotingManager       = voting.VotingManager
)

// Voting error aliases.
var (
	ErrVotingRoundExists    = voting.ErrVotingRoundExists
	ErrVotingRoundNotFound  = voting.ErrVotingRoundNotFound
	ErrVotingAlreadyVoted   = voting.ErrVotingAlreadyVoted
	ErrVotingRoundFinalized = voting.ErrVotingRoundFinalized
	ErrVotingWrongProposal  = voting.ErrVotingWrongProposal
)

// NewVotingManager creates a new voting manager with the given config.
func NewVotingManager(config VotingManagerConfig) *VotingManager {
	return voting.NewVotingManager(config)
}

// StartVotingRound begins a new voting round (wrapper kept for compatibility).
func StartVotingRound(vm *VotingManager, slot uint64, proposalHash types.Hash, threshold float64) error {
	return vm.StartRound(slot, proposalHash, threshold)
}
