package slot

import (
	"math"
	"testing"
	"time"
)

func TestDefaultProgressiveSlotSchedule(t *testing.T) {
	s := DefaultProgressiveSlotSchedule()
	entries := s.Entries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Name != "genesis" {
		t.Fatalf("expected first entry name 'genesis', got %q", entries[0].Name)
	}
	if entries[1].Name != "fast-slots" {
		t.Fatalf("expected second entry name 'fast-slots', got %q", entries[1].Name)
	}
	if entries[2].Name != "quick-slots" {
		t.Fatalf("expected third entry name 'quick-slots', got %q", entries[2].Name)
	}
}

func TestEightSecondSlotConfig(t *testing.T) {
	cfg := EightSecondSlotConfig()
	if cfg.SlotDuration != 8*time.Second {
		t.Fatalf("expected 8s slot duration, got %v", cfg.SlotDuration)
	}
	if cfg.SlotsPerEpoch != 8 {
		t.Fatalf("expected 8 slots/epoch, got %d", cfg.SlotsPerEpoch)
	}
}

func TestEightSecondPhaseTimerConfig(t *testing.T) {
	cfg := EightSecondPhaseTimerConfig()
	total := cfg.ProposalPhaseMs + cfg.AttestationPhaseMs + cfg.AggregationPhaseMs
	if total != 8000 {
		t.Fatalf("expected phase durations sum to 8000ms, got %d", total)
	}
	if cfg.SlotDurationMs != 8000 {
		t.Fatalf("expected SlotDurationMs=8000, got %d", cfg.SlotDurationMs)
	}
	if cfg.ProposalPhaseMs != 3000 {
		t.Fatalf("expected ProposalPhaseMs=3000, got %d", cfg.ProposalPhaseMs)
	}
	if cfg.AttestationPhaseMs != 3000 {
		t.Fatalf("expected AttestationPhaseMs=3000, got %d", cfg.AttestationPhaseMs)
	}
	if cfg.AggregationPhaseMs != 2000 {
		t.Fatalf("expected AggregationPhaseMs=2000, got %d", cfg.AggregationPhaseMs)
	}
	if cfg.SlotsPerEpoch != 8 {
		t.Fatalf("expected SlotsPerEpoch=8, got %d", cfg.SlotsPerEpoch)
	}
}

func TestGetSlotDuration_Genesis(t *testing.T) {
	s := DefaultProgressiveSlotSchedule()
	d := s.GetSlotDuration(0)
	if d != 12*time.Second {
		t.Fatalf("expected 12s at epoch 0, got %v", d)
	}
}

func TestGetSlotDuration_FastSlots(t *testing.T) {
	s := DefaultProgressiveSlotSchedule()
	d := s.GetSlotDuration(100000)
	if d != 8*time.Second {
		t.Fatalf("expected 8s at epoch 100000, got %v", d)
	}
}

func TestGetSlotDuration_QuickSlots(t *testing.T) {
	s := DefaultProgressiveSlotSchedule()
	d := s.GetSlotDuration(200000)
	if d != 6*time.Second {
		t.Fatalf("expected 6s at epoch 200000, got %v", d)
	}
}

func TestGetSlotDuration_Between(t *testing.T) {
	s := DefaultProgressiveSlotSchedule()
	// Epoch 50000 is before the first transition at 100000.
	d := s.GetSlotDuration(50000)
	if d != 12*time.Second {
		t.Fatalf("expected 12s at epoch 50000, got %v", d)
	}

	// Epoch 150000 is between fast-slots and quick-slots.
	d = s.GetSlotDuration(150000)
	if d != 8*time.Second {
		t.Fatalf("expected 8s at epoch 150000, got %v", d)
	}

	// Epoch 300000 is after quick-slots.
	d = s.GetSlotDuration(300000)
	if d != 6*time.Second {
		t.Fatalf("expected 6s at epoch 300000, got %v", d)
	}
}

func TestGetSlotsPerEpoch_Genesis(t *testing.T) {
	s := DefaultProgressiveSlotSchedule()
	spe := s.GetSlotsPerEpoch(0)
	if spe != 32 {
		t.Fatalf("expected 32 at epoch 0, got %d", spe)
	}
}

func TestGetSlotsPerEpoch_FastSlots(t *testing.T) {
	s := DefaultProgressiveSlotSchedule()
	spe := s.GetSlotsPerEpoch(100000)
	if spe != 8 {
		t.Fatalf("expected 8 at epoch 100000, got %d", spe)
	}
}

func TestGetSlotsPerEpoch_QuickSlots(t *testing.T) {
	s := DefaultProgressiveSlotSchedule()
	spe := s.GetSlotsPerEpoch(200000)
	if spe != 4 {
		t.Fatalf("expected 4 at epoch 200000, got %d", spe)
	}
}

func TestSlotToTime_SingleFork(t *testing.T) {
	// Create a schedule with only genesis: 12s, 32 slots/epoch.
	s, err := NewProgressiveSlotSchedule([]ProgressiveSlotEntry{
		{ForkEpoch: 0, SlotDuration: 12 * time.Second, SlotsPerEpoch: 32, Name: "genesis"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	genesis := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	// Slot 0 should be at genesis.
	t0 := s.SlotToTime(0, genesis)
	if !t0.Equal(genesis) {
		t.Fatalf("slot 0 should be genesis, got %v", t0)
	}

	// Slot 10 should be 120s after genesis.
	t10 := s.SlotToTime(10, genesis)
	expected := genesis.Add(120 * time.Second)
	if !t10.Equal(expected) {
		t.Fatalf("slot 10 expected %v, got %v", expected, t10)
	}

	// Slot 32 (first slot of epoch 1) should be 384s after genesis.
	t32 := s.SlotToTime(32, genesis)
	expected32 := genesis.Add(384 * time.Second)
	if !t32.Equal(expected32) {
		t.Fatalf("slot 32 expected %v, got %v", expected32, t32)
	}
}

func TestSlotToTime_AcrossForks(t *testing.T) {
	// Two forks: epoch 0 with 12s and 4 slots/epoch, epoch 2 with 6s and 4 slots/epoch.
	s, err := NewProgressiveSlotSchedule([]ProgressiveSlotEntry{
		{ForkEpoch: 0, SlotDuration: 12 * time.Second, SlotsPerEpoch: 4, Name: "genesis"},
		{ForkEpoch: 2, SlotDuration: 6 * time.Second, SlotsPerEpoch: 4, Name: "fast"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	genesis := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	// Epoch 0-1 has 2 epochs * 4 slots = 8 slots at 12s each = 96s.
	// Slot 8 starts the fast fork.
	// Slot 4 (start of epoch 1) = 48s.
	t4 := s.SlotToTime(4, genesis)
	expected4 := genesis.Add(48 * time.Second)
	if !t4.Equal(expected4) {
		t.Fatalf("slot 4 expected %v, got %v", expected4, t4)
	}

	// Slot 8 (start of epoch 2, fast fork) = 96s.
	t8 := s.SlotToTime(8, genesis)
	expected8 := genesis.Add(96 * time.Second)
	if !t8.Equal(expected8) {
		t.Fatalf("slot 8 expected %v, got %v", expected8, t8)
	}

	// Slot 10 = 96s + 2*6s = 108s.
	t10 := s.SlotToTime(10, genesis)
	expected10 := genesis.Add(108 * time.Second)
	if !t10.Equal(expected10) {
		t.Fatalf("slot 10 expected %v, got %v", expected10, t10)
	}
}

func TestNewProgressiveSlotSchedule_Empty(t *testing.T) {
	_, err := NewProgressiveSlotSchedule([]ProgressiveSlotEntry{})
	if err != ErrSSNoEntries {
		t.Fatalf("expected ErrSSNoEntries, got %v", err)
	}

	_, err = NewProgressiveSlotSchedule(nil)
	if err != ErrSSNoEntries {
		t.Fatalf("expected ErrSSNoEntries for nil, got %v", err)
	}
}

func TestNewProgressiveSlotSchedule_Sorted(t *testing.T) {
	// Provide entries out of order.
	s, err := NewProgressiveSlotSchedule([]ProgressiveSlotEntry{
		{ForkEpoch: 200, SlotDuration: 6 * time.Second, SlotsPerEpoch: 4, Name: "c"},
		{ForkEpoch: 0, SlotDuration: 12 * time.Second, SlotsPerEpoch: 32, Name: "a"},
		{ForkEpoch: 100, SlotDuration: 8 * time.Second, SlotsPerEpoch: 8, Name: "b"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := s.Entries()
	if entries[0].ForkEpoch != 0 || entries[0].Name != "a" {
		t.Fatalf("first entry should be epoch 0, got %d %q", entries[0].ForkEpoch, entries[0].Name)
	}
	if entries[1].ForkEpoch != 100 || entries[1].Name != "b" {
		t.Fatalf("second entry should be epoch 100, got %d %q", entries[1].ForkEpoch, entries[1].Name)
	}
	if entries[2].ForkEpoch != 200 || entries[2].Name != "c" {
		t.Fatalf("third entry should be epoch 200, got %d %q", entries[2].ForkEpoch, entries[2].Name)
	}
}

func TestNewProgressiveSlotSchedule_Overlapping(t *testing.T) {
	_, err := NewProgressiveSlotSchedule([]ProgressiveSlotEntry{
		{ForkEpoch: 0, SlotDuration: 12 * time.Second, SlotsPerEpoch: 32, Name: "a"},
		{ForkEpoch: 0, SlotDuration: 8 * time.Second, SlotsPerEpoch: 8, Name: "b"},
	})
	if err != ErrSSOverlapping {
		t.Fatalf("expected ErrSSOverlapping, got %v", err)
	}
}

func TestComputeProgressiveDuration_Step0(t *testing.T) {
	base := 12 * time.Second
	d := ComputeProgressiveDuration(base, 0)
	if d != base {
		t.Fatalf("step 0 expected %v, got %v", base, d)
	}
}

func TestComputeProgressiveDuration_Step1(t *testing.T) {
	base := 12 * time.Second
	d := ComputeProgressiveDuration(base, 1)
	// base/sqrt(2) ~ 8.485s
	expected := float64(base.Nanoseconds()) / math.Sqrt2
	if math.Abs(float64(d.Nanoseconds())-expected) > float64(time.Millisecond.Nanoseconds()) {
		t.Fatalf("step 1 expected ~%.3fs, got %v", expected/1e9, d)
	}
}

func TestComputeProgressiveDuration_Step2(t *testing.T) {
	base := 12 * time.Second
	d := ComputeProgressiveDuration(base, 2)
	// base / sqrt(2)^2 = base/2 = 6s
	expected := 6 * time.Second
	if math.Abs(float64(d.Nanoseconds()-expected.Nanoseconds())) > float64(time.Millisecond.Nanoseconds()) {
		t.Fatalf("step 2 expected %v, got %v", expected, d)
	}
}

func TestComputeProgressiveDuration_NegativeStep(t *testing.T) {
	base := 12 * time.Second
	d := ComputeProgressiveDuration(base, -1)
	if d != base {
		t.Fatalf("negative step expected %v, got %v", base, d)
	}
}

func TestGetEntry_BoundaryEpochs(t *testing.T) {
	s := DefaultProgressiveSlotSchedule()

	// Exact boundary: epoch 0 -> genesis entry.
	e := s.GetEntry(0)
	if e == nil || e.Name != "genesis" {
		t.Fatalf("epoch 0 expected genesis entry")
	}

	// Just before fast-slots boundary.
	e = s.GetEntry(99999)
	if e == nil || e.Name != "genesis" {
		t.Fatalf("epoch 99999 expected genesis entry, got %q", e.Name)
	}

	// Exact fast-slots boundary.
	e = s.GetEntry(100000)
	if e == nil || e.Name != "fast-slots" {
		t.Fatalf("epoch 100000 expected fast-slots entry, got %q", e.Name)
	}

	// Just before quick-slots boundary.
	e = s.GetEntry(199999)
	if e == nil || e.Name != "fast-slots" {
		t.Fatalf("epoch 199999 expected fast-slots entry, got %q", e.Name)
	}

	// Exact quick-slots boundary.
	e = s.GetEntry(200000)
	if e == nil || e.Name != "quick-slots" {
		t.Fatalf("epoch 200000 expected quick-slots entry, got %q", e.Name)
	}

	// Well past quick-slots.
	e = s.GetEntry(999999)
	if e == nil || e.Name != "quick-slots" {
		t.Fatalf("epoch 999999 expected quick-slots entry, got %q", e.Name)
	}
}
