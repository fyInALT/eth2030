package focil

import (
	"testing"
)

func TestILEquivocationDetectionIdentical(t *testing.T) {
	store := NewILStore()
	il := &InclusionList{Slot: 5, ProposerIndex: 1}

	// Two identical ILs → no equivocation.
	store.AddIL(1, 5, il)
	store.AddIL(1, 5, il)

	if store.IsEquivocator(1, 5) {
		t.Error("identical ILs should not trigger equivocation")
	}
}

func TestILEquivocationDetectionDifferent(t *testing.T) {
	store := NewILStore()

	il1 := &InclusionList{Slot: 5, ProposerIndex: 1, Entries: []InclusionListEntry{
		{Transaction: []byte{0x01}, Index: 0},
	}}
	il2 := &InclusionList{Slot: 5, ProposerIndex: 1, Entries: []InclusionListEntry{
		{Transaction: []byte{0x02}, Index: 0},
	}}

	store.AddIL(1, 5, il1)
	store.AddIL(1, 5, il2)

	if !store.IsEquivocator(1, 5) {
		t.Error("different ILs for same slot should mark equivocator")
	}
	if count := store.EquivocatorCount(5); count != 1 {
		t.Errorf("equivocator count: got %d, want 1", count)
	}
}

func TestILEquivocationDropsSubsequent(t *testing.T) {
	store := NewILStore()

	il1 := &InclusionList{Slot: 5, ProposerIndex: 1, Entries: []InclusionListEntry{
		{Transaction: []byte{0x01}, Index: 0},
	}}
	il2 := &InclusionList{Slot: 5, ProposerIndex: 1, Entries: []InclusionListEntry{
		{Transaction: []byte{0x02}, Index: 0},
	}}
	il3 := &InclusionList{Slot: 5, ProposerIndex: 1, Entries: []InclusionListEntry{
		{Transaction: []byte{0x03}, Index: 0},
	}}

	store.AddIL(1, 5, il1)             // accepted
	store.AddIL(1, 5, il2)             // equivocation detected
	accepted := store.AddIL(1, 5, il3) // should be dropped
	if accepted {
		t.Error("third IL from equivocator should be dropped")
	}
}
