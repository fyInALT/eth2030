package consensus

import (
	"sync"
	"time"
)

// SlotProtocolWiring connects sub-protocol callbacks to the lean slot timer.
// It tracks which sub-protocols have completed their phase for the current slot.
type SlotProtocolWiring struct {
	timer *LeanSlotTimer

	mu                 sync.Mutex
	epbsBidCollected   bool
	focilILDeadlineMet bool
	dasSamplingDone    bool
	focilViewFrozen    bool
	ssfVoteCast        bool
}

// NewSlotProtocolWiring creates a wiring for the given slot timeline.
func NewSlotProtocolWiring(tl *LeanSlotTimeline) *SlotProtocolWiring {
	return &SlotProtocolWiring{
		timer: NewLeanSlotTimer(tl),
	}
}

// Wire registers all sub-protocol callbacks with the timeline.
// It maps each LeanSlotPhase to the appropriate protocol state update:
//
//	PhaseEPBSBidOpen      → opens the bid window (resets epbsBidCollected)
//	PhaseEPBSBidDeadline  → marks ePBS bid as collected
//	PhaseFOCILDeadline    → marks FOCIL IL deadline met
//	PhasePeerDASSampling  → marks DAS sampling done
//	PhaseFOCILViewFreeze  → marks FOCIL view frozen
//	Phase3SFJustification → marks 3SF vote cast
func (w *SlotProtocolWiring) Wire() {
	w.timer.RegisterCallback(func(phase LeanSlotPhase, _ time.Time) {
		w.mu.Lock()
		defer w.mu.Unlock()

		switch phase {
		case PhaseEPBSBidOpen:
			// Bid window opened; reset collection flag for this slot.
			w.epbsBidCollected = false
		case PhaseEPBSBidDeadline:
			w.epbsBidCollected = true
		case PhaseFOCILDeadline:
			w.focilILDeadlineMet = true
		case PhasePeerDASSampling:
			w.dasSamplingDone = true
		case PhaseFOCILViewFreeze:
			w.focilViewFrozen = true
		case Phase3SFJustification:
			w.ssfVoteCast = true
		}
	})
}

// ProtocolStatus returns which protocols have completed their phase for the
// current slot. Keys match the flag names for easy assertion in tests.
func (w *SlotProtocolWiring) ProtocolStatus() map[string]bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return map[string]bool{
		"epbsBidCollected":   w.epbsBidCollected,
		"focilILDeadlineMet": w.focilILDeadlineMet,
		"dasSamplingDone":    w.dasSamplingDone,
		"focilViewFrozen":    w.focilViewFrozen,
		"ssfVoteCast":        w.ssfVoteCast,
	}
}

// Reset clears all protocol completion flags for a new slot.
func (w *SlotProtocolWiring) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.epbsBidCollected = false
	w.focilILDeadlineMet = false
	w.dasSamplingDone = false
	w.focilViewFrozen = false
	w.ssfVoteCast = false
}
