package consensus

// minimmit_compat.go re-exports types from consensus/minimmit for backward compatibility.

import "github.com/eth2030/eth2030/consensus/minimmit"

// Minimmit type aliases.
type (
	FinalityMode   = minimmit.FinalityMode
	MinimmitConfig = minimmit.MinimmitConfig
	MinimmitVote   = minimmit.MinimmitVote
	MinimmitEngine = minimmit.MinimmitEngine
	MinimmitState  = minimmit.MinimmitState
)

// Minimmit constants.
const (
	FinalityModeClassic  = minimmit.FinalityModeClassic
	FinalityModeSSF      = minimmit.FinalityModeSSF
	FinalityModeMinimmit = minimmit.FinalityModeMinimmit
	MinimmitIdle         = minimmit.MinimmitIdle
	MinimmitProposed     = minimmit.MinimmitProposed
	MinimmitVoting       = minimmit.MinimmitVoting
	MinimmitFinalized    = minimmit.MinimmitFinalized
	MinimmitFailed       = minimmit.MinimmitFailed
)

// Minimmit error aliases.
var (
	ErrMinimmitNilConfig     = minimmit.ErrMinimmitNilConfig
	ErrMinimmitZeroStake     = minimmit.ErrMinimmitZeroStake
	ErrMinimmitDuplicateVote = minimmit.ErrMinimmitDuplicateVote
	ErrMinimmitWrongSlot     = minimmit.ErrMinimmitWrongSlot
	ErrMinimmitAlreadyFinal  = minimmit.ErrMinimmitAlreadyFinal
	ErrMinimmitInvalidState  = minimmit.ErrMinimmitInvalidState
	ErrMinimmitEquivocation  = minimmit.ErrMinimmitEquivocation
)

// Minimmit function wrappers.
func DefaultMinimmitConfig() *MinimmitConfig { return minimmit.DefaultMinimmitConfig() }
func NewMinimmitEngine(config *MinimmitConfig) (*MinimmitEngine, error) {
	return minimmit.NewMinimmitEngine(config)
}
