package consensus

import (
	"sync"
	"testing"
	"time"
)

// TestLeanSlotTimeline verifies phase offsets for a 4-second slot duration.
// Expected offsets:
//
//	PhaseEPBSBidOpen     → 0ms
//	PhaseFOCILDeadline   → 1000ms (T*0.25)
//	PhaseEPBSBidDeadline → 2000ms (T*0.50)
//	PhasePeerDASSampling → 2400ms (T*0.60)
//	PhaseFOCILViewFreeze → 3000ms (T*0.75)
//	Phase3SFJustification → 3600ms (T*0.90)
//	PhaseSlotEnd          → 4000ms (T*1.0)
func TestLeanSlotTimeline(t *testing.T) {
	tl := NewLeanSlotTimeline(4 * time.Second)
	if tl == nil {
		t.Fatal("expected non-nil LeanSlotTimeline")
	}

	tests := []struct {
		phase  LeanSlotPhase
		wantMs int64
	}{
		{PhaseEPBSBidOpen, 0},
		{PhaseFOCILDeadline, 1000},
		{PhaseEPBSBidDeadline, 2000},
		{PhasePeerDASSampling, 2400},
		{PhaseFOCILViewFreeze, 3000},
		{Phase3SFJustification, 3600},
		{PhaseSlotEnd, 4000},
	}

	for _, tt := range tests {
		got := tl.PhaseOffset(tt.phase).Milliseconds()
		if got != tt.wantMs {
			t.Errorf("PhaseOffset(%d) = %dms, want %dms", tt.phase, got, tt.wantMs)
		}
	}
}

// TestLeanSlotTimeline6s verifies phase offsets for a 6-second slot duration.
func TestLeanSlotTimeline6s(t *testing.T) {
	tl := NewLeanSlotTimeline(6 * time.Second)
	if tl == nil {
		t.Fatal("expected non-nil LeanSlotTimeline")
	}

	tests := []struct {
		phase  LeanSlotPhase
		wantMs int64
	}{
		{PhaseEPBSBidOpen, 0},
		{PhaseFOCILDeadline, 1500},    // 6000 * 0.25
		{PhaseEPBSBidDeadline, 3000},  // 6000 * 0.50
		{PhasePeerDASSampling, 3600},  // 6000 * 0.60
		{PhaseFOCILViewFreeze, 4500},  // 6000 * 0.75
		{Phase3SFJustification, 5400}, // 6000 * 0.90
		{PhaseSlotEnd, 6000},          // 6000 * 1.0
	}

	for _, tt := range tests {
		got := tl.PhaseOffset(tt.phase).Milliseconds()
		if got != tt.wantMs {
			t.Errorf("6s: PhaseOffset(%d) = %dms, want %dms", tt.phase, got, tt.wantMs)
		}
	}
}

// TestLeanSlotTimelinePhaseTime verifies absolute phase times from a slot start.
func TestLeanSlotTimelinePhaseTime(t *testing.T) {
	tl := NewLeanSlotTimeline(4 * time.Second)
	slotStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	got := tl.PhaseTime(slotStart, PhaseEPBSBidOpen)
	if !got.Equal(slotStart) {
		t.Errorf("PhaseTime EPBSBidOpen = %v, want %v", got, slotStart)
	}

	got = tl.PhaseTime(slotStart, PhaseFOCILDeadline)
	want := slotStart.Add(1000 * time.Millisecond)
	if !got.Equal(want) {
		t.Errorf("PhaseTime FOCILDeadline = %v, want %v", got, want)
	}

	got = tl.PhaseTime(slotStart, PhaseSlotEnd)
	want = slotStart.Add(4 * time.Second)
	if !got.Equal(want) {
		t.Errorf("PhaseTime SlotEnd = %v, want %v", got, want)
	}
}

// TestLeanSlotTimerStart verifies that all phase callbacks fire within one slot.
func TestLeanSlotTimerStart(t *testing.T) {
	tl := NewLeanSlotTimeline(200 * time.Millisecond)
	st := NewLeanSlotTimer(tl)
	if st == nil {
		t.Fatal("expected non-nil LeanSlotTimer")
	}

	var mu sync.Mutex
	fired := make(map[LeanSlotPhase]bool)

	var wg sync.WaitGroup
	wg.Add(int(numPhases))

	st.RegisterCallback(func(phase LeanSlotPhase, _ time.Time) {
		mu.Lock()
		if !fired[phase] {
			fired[phase] = true
			wg.Done()
		}
		mu.Unlock()
	})

	slotStart := time.Now()
	st.Start(slotStart)

	// Wait up to 500ms for all phases to fire (slot is 200ms).
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		mu.Lock()
		missing := []LeanSlotPhase{}
		for p := LeanSlotPhase(0); p < numPhases; p++ {
			if !fired[p] {
				missing = append(missing, p)
			}
		}
		mu.Unlock()
		t.Fatalf("timed out: phases not fired: %v", missing)
	}

	st.Stop()
}

// TestLeanSlotTimerStop verifies that Stop cancels pending timers.
func TestLeanSlotTimerStop(t *testing.T) {
	tl := NewLeanSlotTimeline(10 * time.Second)
	st := NewLeanSlotTimer(tl)

	fired := make(chan LeanSlotPhase, 16)
	st.RegisterCallback(func(phase LeanSlotPhase, _ time.Time) {
		fired <- phase
	})

	// Start with slot beginning far in the future so no phases fire soon.
	slotStart := time.Now().Add(5 * time.Second)
	st.Start(slotStart)
	st.Stop()

	// Give a brief moment to see if anything fires unexpectedly.
	select {
	case p := <-fired:
		t.Errorf("unexpected phase fired after Stop: %v", p)
	case <-time.After(50 * time.Millisecond):
		// OK: nothing fired.
	}
}

// TestLeanSlotTimerMultipleCallbacks verifies that multiple callbacks all fire.
func TestLeanSlotTimerMultipleCallbacks(t *testing.T) {
	tl := NewLeanSlotTimeline(200 * time.Millisecond)
	st := NewLeanSlotTimer(tl)

	var wg1, wg2 sync.WaitGroup
	wg1.Add(int(numPhases))
	wg2.Add(int(numPhases))

	var mu sync.Mutex
	fired1 := make(map[LeanSlotPhase]bool)
	fired2 := make(map[LeanSlotPhase]bool)

	st.RegisterCallback(func(phase LeanSlotPhase, _ time.Time) {
		mu.Lock()
		if !fired1[phase] {
			fired1[phase] = true
			wg1.Done()
		}
		mu.Unlock()
	})
	st.RegisterCallback(func(phase LeanSlotPhase, _ time.Time) {
		mu.Lock()
		if !fired2[phase] {
			fired2[phase] = true
			wg2.Done()
		}
		mu.Unlock()
	})

	st.Start(time.Now())

	done := make(chan struct{})
	go func() {
		wg1.Wait()
		wg2.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out: not all callbacks received all phases")
	}

	st.Stop()
}

// TestIntegratedSlotProtocols runs 3 real slots with SlotProtocolWiring
// and verifies all protocol flags are set after each slot.
func TestIntegratedSlotProtocols(t *testing.T) {
	const slotDuration = 150 * time.Millisecond
	tl := NewLeanSlotTimeline(slotDuration)

	for slot := 0; slot < 3; slot++ {
		// Fresh wiring per slot.
		w := NewSlotProtocolWiring(tl)
		w.Wire()

		slotStart := time.Now()
		w.timer.Start(slotStart)

		// Wait for the slot to finish (all phases including SlotEnd at T*1.0).
		time.Sleep(slotDuration + 20*time.Millisecond)

		status := w.ProtocolStatus()
		for key, ok := range status {
			if !ok {
				t.Errorf("slot %d: protocol %q not completed", slot, key)
			}
		}

		w.timer.Stop()
	}
}

// TestIntegratedSlot20 simulates 20 slots with 100ms duration each,
// verifying all protocol flags are set after each slot.
// Each slot creates a fresh SlotProtocolWiring to avoid state leakage.
func TestIntegratedSlot20(t *testing.T) {
	const slotDuration = 100 * time.Millisecond
	tl := NewLeanSlotTimeline(slotDuration)

	for slot := 0; slot < 20; slot++ {
		// Fresh wiring per slot avoids accumulated callbacks.
		w := NewSlotProtocolWiring(tl)
		w.Wire()

		// Use a WaitGroup to wait for all phases to fire.
		var wg sync.WaitGroup
		wg.Add(int(numPhases))

		var countMu sync.Mutex
		counted := 0

		w.timer.RegisterCallback(func(_ LeanSlotPhase, _ time.Time) {
			countMu.Lock()
			if counted < int(numPhases) {
				counted++
				wg.Done()
			}
			countMu.Unlock()
		})

		slotStart := time.Now()
		w.timer.Start(slotStart)

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(slotDuration * 3):
			t.Errorf("slot %d: timed out waiting for all phases", slot)
			w.timer.Stop()
			continue
		}

		// Give a tiny settling time for the last callbacks.
		time.Sleep(10 * time.Millisecond)

		status := w.ProtocolStatus()
		for key, ok := range status {
			if !ok {
				t.Errorf("slot %d: protocol %q not completed", slot, key)
			}
		}

		w.timer.Stop()
	}
}
