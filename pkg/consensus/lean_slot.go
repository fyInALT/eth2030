// Package consensus implements the Ethereum consensus layer.
// This file defines LeanSlotTimeline and LeanSlotTimer: a sub-slot
// phase timeline that integrates ePBS, FOCIL, PeerDAS, and 3SF
// checkpoints into a unified per-slot schedule.
package consensus

import (
	"sync"
	"time"
)

// LeanSlotPhase identifies a sub-protocol phase within a slot.
type LeanSlotPhase int

const (
	// PhaseEPBSBidOpen is when the ePBS bid window opens (t=0).
	PhaseEPBSBidOpen LeanSlotPhase = iota
	// PhaseFOCILDeadline is the FOCIL inclusion list deadline (t=T*0.25).
	PhaseFOCILDeadline
	// PhaseEPBSBidDeadline is the ePBS bid deadline (t=T*0.50).
	PhaseEPBSBidDeadline
	// PhasePeerDASSampling is the PeerDAS sampling target (t=T*0.60).
	PhasePeerDASSampling
	// PhaseFOCILViewFreeze is the FOCIL view freeze (t=T*0.75).
	PhaseFOCILViewFreeze
	// Phase3SFJustification is the 3SF justification vote (t=T*0.90).
	Phase3SFJustification
	// PhaseSlotEnd is the slot boundary (t=T).
	PhaseSlotEnd

	numPhases
)

// phasefractions holds the fractional offsets for each phase.
// Indices correspond to LeanSlotPhase constants.
var phasefractions = [numPhases]float64{
	0.00, // PhaseEPBSBidOpen
	0.25, // PhaseFOCILDeadline
	0.50, // PhaseEPBSBidDeadline
	0.60, // PhasePeerDASSampling
	0.75, // PhaseFOCILViewFreeze
	0.90, // Phase3SFJustification
	1.00, // PhaseSlotEnd
}

// LeanSlotTimeline holds timing offsets for each phase within a slot.
type LeanSlotTimeline struct {
	SlotDuration time.Duration
	// Offsets maps each phase to its offset from slot start.
	Offsets [numPhases]time.Duration
}

// NewLeanSlotTimeline creates a timeline for the given slot duration.
// Phase offsets are computed as the specified fractions of SlotDuration:
//
//	t=0, T*0.25, T*0.5, T*0.6, T*0.75, T*0.9, T*1.0
func NewLeanSlotTimeline(slotDuration time.Duration) *LeanSlotTimeline {
	tl := &LeanSlotTimeline{SlotDuration: slotDuration}
	for i := LeanSlotPhase(0); i < numPhases; i++ {
		ns := int64(float64(slotDuration.Nanoseconds()) * phasefractions[i])
		tl.Offsets[i] = time.Duration(ns)
	}
	return tl
}

// PhaseOffset returns the time offset of the given phase from slot start.
func (tl *LeanSlotTimeline) PhaseOffset(phase LeanSlotPhase) time.Duration {
	if phase < 0 || phase >= numPhases {
		return 0
	}
	return tl.Offsets[phase]
}

// PhaseTime returns the absolute time for a phase given the slot start time.
func (tl *LeanSlotTimeline) PhaseTime(slotStart time.Time, phase LeanSlotPhase) time.Time {
	return slotStart.Add(tl.PhaseOffset(phase))
}

// PhaseCallback is called when a timeline phase fires.
type PhaseCallback func(phase LeanSlotPhase, slotStart time.Time)

// LeanSlotTimer manages phase timers for a single slot.
type LeanSlotTimer struct {
	timeline  *LeanSlotTimeline
	mu        sync.Mutex
	callbacks []PhaseCallback
	timers    [numPhases]*time.Timer
}

// NewLeanSlotTimer creates a timer for the given timeline.
func NewLeanSlotTimer(tl *LeanSlotTimeline) *LeanSlotTimer {
	return &LeanSlotTimer{
		timeline: tl,
	}
}

// RegisterCallback registers a callback for all phase events.
func (st *LeanSlotTimer) RegisterCallback(cb PhaseCallback) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.callbacks = append(st.callbacks, cb)
}

// Start fires all phase timers for a slot starting at slotStart.
// Any previously running timers are stopped first.
func (st *LeanSlotTimer) Start(slotStart time.Time) {
	st.Stop()
	now := time.Now()

	st.mu.Lock()
	cbs := make([]PhaseCallback, len(st.callbacks))
	copy(cbs, st.callbacks)
	st.mu.Unlock()

	for i := LeanSlotPhase(0); i < numPhases; i++ {
		phase := i
		fireAt := slotStart.Add(st.timeline.Offsets[i])
		delay := fireAt.Sub(now)
		if delay < 0 {
			delay = 0
		}

		timer := time.AfterFunc(delay, func() {
			for _, cb := range cbs {
				cb(phase, slotStart)
			}
		})

		st.mu.Lock()
		st.timers[phase] = timer
		st.mu.Unlock()
	}
}

// Stop cancels all pending timers.
func (st *LeanSlotTimer) Stop() {
	st.mu.Lock()
	defer st.mu.Unlock()
	for i := range st.timers {
		if st.timers[i] != nil {
			st.timers[i].Stop()
			st.timers[i] = nil
		}
	}
}
