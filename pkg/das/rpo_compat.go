package das

// rpo_compat.go re-exports types from das/rpo for backward compatibility.

import "github.com/eth2030/eth2030/das/rpo"

// RPO type aliases.
type (
	RPOConfig        = rpo.RPOConfig
	ThroughputEstimate = rpo.ThroughputEstimate
	RPOSchedule      = rpo.RPOSchedule
	RPOHistoryEntry  = rpo.RPOHistoryEntry
	RPOManager       = rpo.RPOManager
)

// RPO error variables.
var (
	ErrRPOBelowMin       = rpo.ErrRPOBelowMin
	ErrRPOAboveMax       = rpo.ErrRPOAboveMax
	ErrRPOStepTooLarge   = rpo.ErrRPOStepTooLarge
	ErrRPONotIncreasing  = rpo.ErrRPONotIncreasing
	ErrRPOScheduleEmpty  = rpo.ErrRPOScheduleEmpty
	ErrRPOScheduleOrder  = rpo.ErrRPOScheduleOrder
	ErrRPOScheduleValues = rpo.ErrRPOScheduleValues
)

// RPO function wrappers.
func DefaultRPOConfig() RPOConfig              { return rpo.DefaultRPOConfig() }
func NewRPOManager(config RPOConfig) *RPOManager { return rpo.NewRPOManager(config) }
func ValidateBlobSchedule(schedule []*RPOSchedule, config RPOConfig) error {
	return rpo.ValidateBlobSchedule(schedule, config)
}
func BPO3Schedule() []*RPOSchedule                          { return rpo.BPO3Schedule() }
func BPO4Schedule() []*RPOSchedule                          { return rpo.BPO4Schedule() }
func MergeBPOSchedules(phases ...[]*RPOSchedule) ([]*RPOSchedule, error) {
	return rpo.MergeBPOSchedules(phases...)
}
