package focil

import (
	"testing"
)

// --- Extended ILStore tests ---

func TestILStore_MultipleValidatorsEquivocating(t *testing.T) {
	store := NewILStore()

	// Two validators each equivocate on the same slot.
	for v := uint64(1); v <= 2; v++ {
		il1 := &InclusionList{Slot: 7, ProposerIndex: v, Entries: []InclusionListEntry{
			{Transaction: []byte{byte(v), 0x01}, Index: 0},
		}}
		il2 := &InclusionList{Slot: 7, ProposerIndex: v, Entries: []InclusionListEntry{
			{Transaction: []byte{byte(v), 0x02}, Index: 0},
		}}
		store.AddIL(v, 7, il1)
		store.AddIL(v, 7, il2)
	}

	if count := store.EquivocatorCount(7); count != 2 {
		t.Errorf("equivocator count slot 7: got %d, want 2", count)
	}
	// Different slot: no equivocation.
	if count := store.EquivocatorCount(8); count != 0 {
		t.Errorf("equivocator count slot 8: got %d, want 0", count)
	}
}

func TestILStore_EquivocationOnlyOnSpecificSlot(t *testing.T) {
	store := NewILStore()

	// Validator 1 equivocates on slot 5 but is clean on slot 6.
	il_a := &InclusionList{Slot: 5, ProposerIndex: 1, Entries: []InclusionListEntry{{Transaction: []byte{0xAA}, Index: 0}}}
	il_b := &InclusionList{Slot: 5, ProposerIndex: 1, Entries: []InclusionListEntry{{Transaction: []byte{0xBB}, Index: 0}}}
	il_c := &InclusionList{Slot: 6, ProposerIndex: 1, Entries: []InclusionListEntry{{Transaction: []byte{0xCC}, Index: 0}}}

	store.AddIL(1, 5, il_a)
	store.AddIL(1, 5, il_b)
	store.AddIL(1, 6, il_c) // different slot, first IL

	if !store.IsEquivocator(1, 5) {
		t.Error("validator 1 should be equivocator on slot 5")
	}
	if store.IsEquivocator(1, 6) {
		t.Error("validator 1 should NOT be equivocator on slot 6")
	}
}

// --- Extended IL satisfaction tests ---

func TestILSatisfaction_EmptyILs(t *testing.T) {
	block := makeBlockWithTxs()
	// Empty IL list → always satisfied.
	result := CheckILSatisfaction(block, nil, nil, 1_000_000)
	if result != ILSatisfied {
		t.Errorf("empty ILs: expected ILSatisfied, got %v", result)
	}
}

func TestILSatisfaction_NilPostState(t *testing.T) {
	// nil postState → no state exemption → absent tx with gas causes unsatisfied.
	tx1 := makeMinimalTx(1)
	block := makeBlockWithTxs() // empty
	il := makeILWithRawTxs(tx1)

	// With nil postState, state check is skipped entirely; absent tx → unsatisfied.
	result := CheckILSatisfaction(block, []*InclusionList{il}, nil, 1_000_000)
	if result != ILUnsatisfied {
		t.Errorf("nil postState: expected ILUnsatisfied for absent tx, got %v", result)
	}
}

func TestILSatisfaction_AllTxsInBlock(t *testing.T) {
	tx1 := makeMinimalTx(1)
	tx2 := makeMinimalTx(2)
	block := makeBlockWithTxs(tx1, tx2)
	il := makeILWithRawTxs(tx1, tx2)

	result := CheckILSatisfaction(block, []*InclusionList{il}, nil, 1_000_000)
	if result != ILSatisfied {
		t.Errorf("all txs in block: expected ILSatisfied, got %v", result)
	}
}

func TestILSatisfaction_PartialPresence(t *testing.T) {
	// tx1 in block, tx2 not in block but gas available → unsatisfied.
	tx1 := makeMinimalTx(1)
	tx2 := makeMinimalTx(2)
	block := makeBlockWithTxs(tx1) // only tx1
	il := makeILWithRawTxs(tx1, tx2)

	result := CheckILSatisfaction(block, []*InclusionList{il}, nil, 1_000_000)
	if result != ILUnsatisfied {
		t.Errorf("partial: expected ILUnsatisfied, got %v", result)
	}
}

func TestILSatisfaction_MultipleILsOneSatisfied(t *testing.T) {
	tx1 := makeMinimalTx(1)
	tx2 := makeMinimalTx(2)
	block := makeBlockWithTxs(tx1) // only tx1
	il1 := makeILWithRawTxs(tx1)   // satisfied
	il2 := makeILWithRawTxs(tx2)   // unsatisfied

	result := CheckILSatisfaction(block, []*InclusionList{il1, il2}, nil, 1_000_000)
	if result != ILUnsatisfied {
		t.Errorf("second IL unsatisfied: expected ILUnsatisfied, got %v", result)
	}
}

func TestILSatisfaction_InvalidTxSkipped(t *testing.T) {
	// IL entry with invalid RLP → skipped (not a violation).
	il := &InclusionList{Slot: 1, ProposerIndex: 1, Entries: []InclusionListEntry{
		{Transaction: []byte{0xDE, 0xAD, 0xBE, 0xEF}, Index: 0}, // invalid RLP
	}}
	block := makeBlockWithTxs()
	result := CheckILSatisfaction(block, []*InclusionList{il}, nil, 1_000_000)
	if result != ILSatisfied {
		t.Errorf("invalid RLP tx: expected ILSatisfied (skipped), got %v", result)
	}
}
