package cltypes

import (
	"testing"
)

func TestJustificationBits(t *testing.T) {
	var j JustificationBits

	if j.IsJustified(0) {
		t.Error("expected bit 0 to be unset initially")
	}

	j.Set(0)
	if !j.IsJustified(0) {
		t.Error("expected bit 0 to be set after Set(0)")
	}

	j.Shift(1)
	if j.IsJustified(0) {
		t.Error("expected bit 0 to be cleared after Shift(1)")
	}
	if !j.IsJustified(1) {
		t.Error("expected bit 1 to be set after Shift(1)")
	}

	// Out-of-range offset.
	if j.IsJustified(8) {
		t.Error("out-of-range offset should return false")
	}
}

func TestSlotToEpoch(t *testing.T) {
	tests := []struct {
		slot          Slot
		slotsPerEpoch uint64
		want          Epoch
	}{
		{0, 32, 0},
		{31, 32, 0},
		{32, 32, 1},
		{64, 32, 2},
		{0, 0, 0}, // degenerate: slotsPerEpoch=0
	}
	for _, tt := range tests {
		got := SlotToEpoch(tt.slot, tt.slotsPerEpoch)
		if got != tt.want {
			t.Errorf("SlotToEpoch(%d, %d) = %d, want %d", tt.slot, tt.slotsPerEpoch, got, tt.want)
		}
	}
}

func TestEpochStartSlot(t *testing.T) {
	if got := EpochStartSlot(3, 32); got != 96 {
		t.Errorf("EpochStartSlot(3,32) = %d, want 96", got)
	}
}
