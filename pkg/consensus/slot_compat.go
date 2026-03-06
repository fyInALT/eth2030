package consensus

// slot_compat.go re-exports types from consensus/slot for backward compatibility.

import (
	"time"

	"github.com/eth2030/eth2030/consensus/slot"
)

// Slot timing type aliases.
type (
	QuickSlotConfig         = slot.QuickSlotConfig
	QuickSlotScheduler      = slot.QuickSlotScheduler
	ValidatorDuties         = slot.ValidatorDuties
	ProgressiveSlotSchedule = slot.ProgressiveSlotSchedule
	ProgressiveSlotEntry    = slot.ProgressiveSlotEntry
	PhaseTimerConfig        = slot.PhaseTimerConfig
	PhaseTimer              = slot.PhaseTimer
	SlotTimerConfig         = slot.SlotTimerConfig
	SlotTimer               = slot.SlotTimer
	LeanSlotPhase           = slot.LeanSlotPhase
	LeanSlotTimeline        = slot.LeanSlotTimeline
	LeanSlotTimer           = slot.LeanSlotTimer
	SlotProtocolWiring      = slot.SlotProtocolWiring
)

// Slot timing constants.
const (
	PhaseEPBSBidOpen      = slot.PhaseEPBSBidOpen
	PhaseFOCILDeadline    = slot.PhaseFOCILDeadline
	PhaseEPBSBidDeadline  = slot.PhaseEPBSBidDeadline
	PhasePeerDASSampling  = slot.PhasePeerDASSampling
	PhaseFOCILViewFreeze  = slot.PhaseFOCILViewFreeze
	Phase3SFJustification = slot.Phase3SFJustification
	PhaseSlotEnd          = slot.PhaseSlotEnd
)

// Slot timing function wrappers.
func DefaultQuickSlotConfig() *QuickSlotConfig { return slot.DefaultQuickSlotConfig() }
func QuickSlot4sConfig() *QuickSlotConfig      { return slot.QuickSlot4sConfig() }
func IsQuick4s(cfg *QuickSlotConfig) bool      { return slot.IsQuick4s(cfg) }
func NewQuickSlotScheduler(cfg *QuickSlotConfig, genesisTime time.Time) *QuickSlotScheduler {
	return slot.NewQuickSlotScheduler(cfg, genesisTime)
}
func NewProgressiveSlotSchedule(entries []ProgressiveSlotEntry) (*ProgressiveSlotSchedule, error) {
	return slot.NewProgressiveSlotSchedule(entries)
}
func DefaultProgressiveSlotSchedule() *ProgressiveSlotSchedule {
	return slot.DefaultProgressiveSlotSchedule()
}
func DefaultPhaseTimerConfig() *PhaseTimerConfig      { return slot.DefaultPhaseTimerConfig() }
func NewPhaseTimer(cfg *PhaseTimerConfig) *PhaseTimer { return slot.NewPhaseTimer(cfg) }
func DefaultSlotTimerConfig() SlotTimerConfig         { return slot.DefaultSlotTimerConfig() }
func NewSlotTimer(cfg SlotTimerConfig) *SlotTimer {
	return slot.NewSlotTimer(cfg)
}
func NewLeanSlotTimer(timeline *LeanSlotTimeline) *LeanSlotTimer {
	return slot.NewLeanSlotTimer(timeline)
}
func NewSlotProtocolWiring(timeline *LeanSlotTimeline) *SlotProtocolWiring {
	return slot.NewSlotProtocolWiring(timeline)
}
