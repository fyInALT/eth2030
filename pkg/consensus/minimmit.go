// minimmit.go implements the Minimmit one-round BFT consensus engine.
// Minimmit achieves single-round finality by requiring 2/3+ of total stake
// to vote for a proposed block within a single round, eliminating the
// multi-round message exchange of traditional PBFT/HotStuff protocols.
//
// State machine: Idle -> Proposed -> Voting -> Finalized (or Failed).
package consensus

import (
	"errors"
	"sync"

	"github.com/eth2030/eth2030/core/types"
)

// Minimmit errors.
var (
	ErrMinimmitNilConfig     = errors.New("minimmit: nil config")
	ErrMinimmitZeroStake     = errors.New("minimmit: zero total stake")
	ErrMinimmitDuplicateVote = errors.New("minimmit: duplicate vote")
	ErrMinimmitWrongSlot     = errors.New("minimmit: vote for wrong slot")
	ErrMinimmitAlreadyFinal  = errors.New("minimmit: slot already finalized")
	ErrMinimmitInvalidState  = errors.New("minimmit: invalid state transition")
	ErrMinimmitEquivocation  = errors.New("minimmit: equivocation detected")
)

// FinalityMode indicates the finality engine variant.
type FinalityMode uint8

const (
	// FinalityModeClassic represents traditional Casper FFG finality.
	FinalityModeClassic FinalityMode = iota
	// FinalityModeSSF represents Single-Slot Finality.
	FinalityModeSSF
	// FinalityModeMinimmit represents Minimmit one-round BFT finality.
	FinalityModeMinimmit
)

// String returns a human-readable name for the finality mode.
func (m FinalityMode) String() string {
	switch m {
	case FinalityModeClassic:
		return "Classic"
	case FinalityModeSSF:
		return "SSF"
	case FinalityModeMinimmit:
		return "Minimmit"
	default:
		return "Unknown"
	}
}

// MinimmitState represents the BFT state machine states.
type MinimmitState uint8

const (
	// MinimmitIdle is waiting for a block proposal.
	MinimmitIdle MinimmitState = iota
	// MinimmitProposed means a block has been proposed, awaiting votes.
	MinimmitProposed
	// MinimmitVoting means the engine is collecting validator votes.
	MinimmitVoting
	// MinimmitFinalized means 2/3+ votes reached; the slot is finalized.
	MinimmitFinalized
	// MinimmitFailed means the slot was missed or had insufficient votes.
	MinimmitFailed
)

// String returns a human-readable name for the Minimmit state.
func (s MinimmitState) String() string {
	switch s {
	case MinimmitIdle:
		return "Idle"
	case MinimmitProposed:
		return "Proposed"
	case MinimmitVoting:
		return "Voting"
	case MinimmitFinalized:
		return "Finalized"
	case MinimmitFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// MinimmitConfig configures the Minimmit engine.
type MinimmitConfig struct {
	// TotalStake is the total active stake in Gwei.
	TotalStake uint64

	// FinalityThresholdNum is the numerator of the finality fraction (default: 2).
	FinalityThresholdNum uint64

	// FinalityThresholdDen is the denominator of the finality fraction (default: 3).
	FinalityThresholdDen uint64

	// VoterLimit is the maximum number of voters per round.
	VoterLimit uint64

	// MissedSlotPenalty is the penalty in Gwei for missing a slot.
	MissedSlotPenalty uint64
}

// MinimmitVote is a single validator vote in a Minimmit round.
type MinimmitVote struct {
	ValidatorIndex ValidatorIndex
	Slot           uint64
	BlockRoot      types.Hash
	Signature      [96]byte
	Stake          uint64
}

// MinimmitEngine implements one-round BFT finality.
type MinimmitEngine struct {
	mu     sync.RWMutex
	config *MinimmitConfig

	// Current round state.
	state        MinimmitState
	currentSlot  uint64
	proposedRoot types.Hash

	// Vote tracking.
	votes          map[ValidatorIndex]*MinimmitVote
	stakeByRoot    map[types.Hash]uint64
	totalVoteStake uint64

	// Equivocation tracking: validator -> first block root voted for.
	firstVotes map[ValidatorIndex]types.Hash

	// Finality results.
	finalizedSlot uint64
	finalizedRoot types.Hash

	// Missed slot tracking.
	missedSlots []uint64
}

// DefaultMinimmitConfig returns production defaults for Minimmit.
// Uses 32M ETH total stake, 2/3 supermajority threshold, 8192 voter limit,
// and 1 ETH missed slot penalty.
func DefaultMinimmitConfig() *MinimmitConfig {
	return &MinimmitConfig{
		TotalStake:           32_000_000 * GweiPerETH,
		FinalityThresholdNum: 2,
		FinalityThresholdDen: 3,
		VoterLimit:           8192,
		MissedSlotPenalty:    1 * GweiPerETH,
	}
}

// NewMinimmitEngine creates a new Minimmit engine with the given config.
// Returns an error if config is nil or has zero total stake.
func NewMinimmitEngine(config *MinimmitConfig) (*MinimmitEngine, error) {
	if config == nil {
		return nil, ErrMinimmitNilConfig
	}
	if config.TotalStake == 0 {
		return nil, ErrMinimmitZeroStake
	}
	return &MinimmitEngine{
		config:      config,
		state:       MinimmitIdle,
		votes:       make(map[ValidatorIndex]*MinimmitVote),
		stakeByRoot: make(map[types.Hash]uint64),
		firstVotes:  make(map[ValidatorIndex]types.Hash),
	}, nil
}

// ProposeBlock proposes a block for the given slot. Transitions the engine
// from Idle to Voting (via Proposed). Returns an error if the engine is not
// in the Idle state or the slot is already finalized.
func (e *MinimmitEngine) ProposeBlock(slot uint64, blockRoot types.Hash) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state != MinimmitIdle {
		return ErrMinimmitInvalidState
	}
	if e.finalizedSlot > 0 && slot <= e.finalizedSlot {
		return ErrMinimmitAlreadyFinal
	}

	e.currentSlot = slot
	e.proposedRoot = blockRoot
	e.state = MinimmitProposed

	// Initialize vote tracking for this round.
	e.votes = make(map[ValidatorIndex]*MinimmitVote)
	e.stakeByRoot = make(map[types.Hash]uint64)
	e.firstVotes = make(map[ValidatorIndex]types.Hash)
	e.totalVoteStake = 0

	// Immediately transition to Voting.
	e.state = MinimmitVoting
	return nil
}

// CastVote records a validator's vote for the current round.
// Validates that the engine is in Voting state, the vote targets the current
// slot, and that the validator has not already voted or equivocated.
// If the finality threshold is met after this vote, transitions to Finalized.
func (e *MinimmitEngine) CastVote(vote MinimmitVote) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.state != MinimmitVoting {
		return ErrMinimmitInvalidState
	}
	if vote.Slot != e.currentSlot {
		return ErrMinimmitWrongSlot
	}

	// Check for equivocation: same validator, different block root.
	if prevRoot, exists := e.firstVotes[vote.ValidatorIndex]; exists {
		if prevRoot != vote.BlockRoot {
			return ErrMinimmitEquivocation
		}
		// Same root = duplicate vote.
		return ErrMinimmitDuplicateVote
	}

	// Record the vote.
	voteCopy := vote
	e.votes[vote.ValidatorIndex] = &voteCopy
	e.firstVotes[vote.ValidatorIndex] = vote.BlockRoot
	e.stakeByRoot[vote.BlockRoot] += vote.Stake
	e.totalVoteStake += vote.Stake

	// Check if the finality threshold is met for the voted root.
	if meetsMinimmitThreshold(
		e.stakeByRoot[vote.BlockRoot],
		e.config.TotalStake,
		e.config.FinalityThresholdNum,
		e.config.FinalityThresholdDen,
	) {
		e.state = MinimmitFinalized
		e.finalizedSlot = e.currentSlot
		e.finalizedRoot = vote.BlockRoot
	}

	return nil
}

// CheckFinality returns true if the current round has reached finality.
func (e *MinimmitEngine) CheckFinality() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.state == MinimmitFinalized
}

// MissSlot marks the current slot as missed and transitions the engine
// to the Failed state.
func (e *MinimmitEngine) MissSlot(slot uint64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.state = MinimmitFailed
	e.missedSlots = append(e.missedSlots, slot)
}

// Reset resets the engine to the Idle state for the next round.
// Finalized slot/root are preserved.
func (e *MinimmitEngine) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.state = MinimmitIdle
	e.currentSlot = 0
	e.proposedRoot = types.Hash{}
	e.votes = make(map[ValidatorIndex]*MinimmitVote)
	e.stakeByRoot = make(map[types.Hash]uint64)
	e.firstVotes = make(map[ValidatorIndex]types.Hash)
	e.totalVoteStake = 0
}

// State returns the current engine state.
func (e *MinimmitEngine) State() MinimmitState {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.state
}

// FinalizedHead returns the most recently finalized slot and block root.
func (e *MinimmitEngine) FinalizedHead() (slot uint64, root types.Hash) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.finalizedSlot, e.finalizedRoot
}

// VoteCount returns the number of votes recorded in the current round.
func (e *MinimmitEngine) VoteCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.votes)
}

// StakeForRoot returns the accumulated stake for a given block root
// in the current round.
func (e *MinimmitEngine) StakeForRoot(root types.Hash) uint64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.stakeByRoot[root]
}

// MissedSlots returns a copy of the missed slot numbers.
func (e *MinimmitEngine) MissedSlots() []uint64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	result := make([]uint64, len(e.missedSlots))
	copy(result, e.missedSlots)
	return result
}

// meetsMinimmitThreshold returns true if voteStake meets or exceeds the
// supermajority threshold. Uses integer arithmetic:
// voteStake * den >= totalStake * num.
func meetsMinimmitThreshold(voteStake, totalStake, num, den uint64) bool {
	if totalStake == 0 || den == 0 {
		return false
	}
	return voteStake*den >= totalStake*num
}
