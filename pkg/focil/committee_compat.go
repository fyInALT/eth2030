package focil

// committee_compat.go re-exports types from focil/committee for backward compatibility.

import "github.com/eth2030/eth2030/focil/committee"

// Committee type aliases.
type (
	CommitteeDuty            = committee.CommitteeDuty
	SlotCommittee            = committee.SlotCommittee
	QuorumStatus             = committee.QuorumStatus
	CommitteeTrackerConfig   = committee.CommitteeTrackerConfig
	CommitteeTracker         = committee.CommitteeTracker
	SubmissionRecord         = committee.SubmissionRecord
	CommitteeVotingConfig    = committee.CommitteeVotingConfig
	CommitteeVoting          = committee.CommitteeVoting
	SelectionProof           = committee.SelectionProof
	RotationRecord           = committee.RotationRecord
	CommitteeSelectionConfig = committee.CommitteeSelectionConfig
	CommitteeSelector        = committee.CommitteeSelector
)

// Committee constants.
const (
	IL_COMMITTEE_SIZE_COMPAT = committee.IL_COMMITTEE_SIZE
	SlotsPerEpoch            = committee.SlotsPerEpoch
	QuorumNumerator          = committee.QuorumNumerator
	QuorumDenominator        = committee.QuorumDenominator
)

// Committee error variables.
var (
	ErrTrackerNoValidators     = committee.ErrTrackerNoValidators
	ErrTrackerSlotZero         = committee.ErrTrackerSlotZero
	ErrTrackerNotCommittee     = committee.ErrTrackerNotCommittee
	ErrTrackerQuorumNotReached = committee.ErrTrackerQuorumNotReached
	ErrTrackerDuplicateList    = committee.ErrTrackerDuplicateList
	ErrTrackerSlotMismatch     = committee.ErrTrackerSlotMismatch
	ErrVotingSlotZero          = committee.ErrVotingSlotZero
	ErrVotingNoValidators      = committee.ErrVotingNoValidators
	ErrVotingNotMember         = committee.ErrVotingNotMember
	ErrVotingDuplicateSubmit   = committee.ErrVotingDuplicateSubmit
	ErrVotingEpochZero         = committee.ErrVotingEpochZero
)

// Committee function wrappers.
func DefaultCommitteeTrackerConfig() CommitteeTrackerConfig {
	return committee.DefaultCommitteeTrackerConfig()
}
func NewCommitteeTracker(config CommitteeTrackerConfig, validators []uint64) *CommitteeTracker {
	return committee.NewCommitteeTracker(config, validators)
}
func DefaultCommitteeVotingConfig() CommitteeVotingConfig {
	return committee.DefaultCommitteeVotingConfig()
}
func NewCommitteeVoting(config CommitteeVotingConfig, validatorCount uint64) *CommitteeVoting {
	return committee.NewCommitteeVoting(config, validatorCount)
}
func DefaultCommitteeSelectionConfig() CommitteeSelectionConfig {
	return committee.DefaultCommitteeSelectionConfig()
}
func NewCommitteeSelector(config CommitteeSelectionConfig, validators []uint64) *CommitteeSelector {
	return committee.NewCommitteeSelector(config, validators)
}
func VerifySelectionProof(proof *SelectionProof, seed [32]byte) bool {
	return committee.VerifySelectionProof(proof, seed)
}
