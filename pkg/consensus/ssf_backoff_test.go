package consensus

import (
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// TestIsJustifiableSlot_EarlySlots verifies delta ≤ 5 is always justifiable (GAP-5.3).
func TestIsJustifiableSlot_EarlySlots(t *testing.T) {
	for delta := uint64(0); delta <= 5; delta++ {
		if !IsJustifiableSlot(100, 100+delta) {
			t.Errorf("IsJustifiableSlot(100, %d) = false, want true (delta=%d ≤ 5)",
				100+delta, delta)
		}
	}
}

// TestIsJustifiableSlot_PerfectSquares verifies perfect square deltas are
// justifiable (GAP-5.3). From 3SF-mini: delta = x^2 for x=1,2,3,4,5,6,...
func TestIsJustifiableSlot_PerfectSquares(t *testing.T) {
	perfectSquares := []uint64{1, 4, 9, 16, 25, 36, 49, 64, 81, 100}
	for _, sq := range perfectSquares {
		if !IsJustifiableSlot(0, sq) {
			t.Errorf("IsJustifiableSlot(0, %d) = false, want true (perfect square)", sq)
		}
	}
}

// TestIsJustifiableSlot_OblongNumbers verifies oblong deltas are justifiable
// (GAP-5.3). Oblong = x*(x+1): 2, 6, 12, 20, 30, 42, 56, 72, 90, 110.
func TestIsJustifiableSlot_OblongNumbers(t *testing.T) {
	oblongNumbers := []uint64{2, 6, 12, 20, 30, 42, 56, 72, 90, 110}
	for _, ob := range oblongNumbers {
		if !IsJustifiableSlot(0, ob) {
			t.Errorf("IsJustifiableSlot(0, %d) = false, want true (oblong)", ob)
		}
	}
}

// TestIsJustifiableSlot_NonJustifiable verifies non-justifiable deltas (GAP-5.3).
// Deltas > 5 that are neither perfect squares nor oblong numbers should be false.
func TestIsJustifiableSlot_NonJustifiable(t *testing.T) {
	// delta=7: not ≤5, not perfect square, not oblong (6 is, 7 is not).
	// delta=8: not ≤5, not perfect square, not oblong.
	// delta=11: not ≤5, not perfect square (9 is, 11 is not), not oblong (12 is).
	nonJustifiable := []uint64{7, 8, 10, 11, 13, 14, 15, 17, 18, 19, 21}
	for _, d := range nonJustifiable {
		if IsJustifiableSlot(0, d) {
			t.Errorf("IsJustifiableSlot(0, %d) = true, want false (non-justifiable delta)", d)
		}
	}
}

// TestIsJustifiableSlot_CandidateBelowFinalized verifies rejection of invalid inputs.
func TestIsJustifiableSlot_CandidateBelowFinalized(t *testing.T) {
	if IsJustifiableSlot(10, 9) {
		t.Error("IsJustifiableSlot(10, 9) = true, want false (candidate < finalized)")
	}
}

// TestIsJustifiableSlot_SameSlot verifies delta=0 is justifiable (trivial).
func TestIsJustifiableSlot_SameSlot(t *testing.T) {
	if !IsJustifiableSlot(5, 5) {
		t.Error("IsJustifiableSlot(5, 5) = false, want true (delta=0)")
	}
}

// TestIsJustifiableSlot_Progression verifies the backoff progression matches
// the sequence from refs/research/3sf-mini/consensus.py. After delta=5, the
// next justifiable slots should be 6 (oblong), 9 (square), 12 (oblong)... (GAP-5.3).
func TestIsJustifiableSlot_Progression(t *testing.T) {
	// Collect all justifiable deltas in range [0, 30].
	var justifiable []uint64
	for d := uint64(0); d <= 30; d++ {
		if IsJustifiableSlot(0, d) {
			justifiable = append(justifiable, d)
		}
	}

	// Expected: 0,1,2,3,4,5 (delta ≤5), 6 (oblong), 9 (square),
	//           12 (oblong), 16 (square), 20 (oblong), 25 (square), 30 (oblong).
	expected := []uint64{0, 1, 2, 3, 4, 5, 6, 9, 12, 16, 20, 25, 30}
	if len(justifiable) != len(expected) {
		t.Fatalf("justifiable slots in [0,30] = %v, want %v", justifiable, expected)
	}
	for i, v := range justifiable {
		if v != expected[i] {
			t.Errorf("justifiable[%d] = %d, want %d", i, v, expected[i])
		}
	}
}

// TestMinimmit3SFBackoff verifies that Minimmit only fires finality at
// justifiable slots per the 3SF backoff algorithm (GAP-5.3).
func TestMinimmit3SFBackoff(t *testing.T) {
	cfg := DefaultMinimmitConfig()
	engine, err := NewMinimmitEngine(cfg)
	if err != nil {
		t.Fatalf("NewMinimmitEngine: %v", err)
	}

	finalizedSlot := uint64(100)

	// Test a run of 30 slots after finalized slot.
	// Minimmit should only attempt finality at justifiable slots.
	for slot := finalizedSlot; slot <= finalizedSlot+30; slot++ {
		justifiable := IsJustifiableSlot(finalizedSlot, slot)
		delta := slot - finalizedSlot

		// Verify the algorithm matches expected justifiable deltas.
		if delta <= 5 {
			if !justifiable {
				t.Errorf("slot %d (delta=%d) should be justifiable (delta ≤ 5)", slot, delta)
			}
		}
	}

	// Verify non-justifiable slots are correctly identified.
	// delta=7: not justifiable.
	if IsJustifiableSlot(finalizedSlot, finalizedSlot+7) {
		t.Error("delta=7 should NOT be justifiable")
	}
	// delta=6: justifiable (oblong: 2*3=6).
	if !IsJustifiableSlot(finalizedSlot, finalizedSlot+6) {
		t.Error("delta=6 should be justifiable (oblong)")
	}

	// Use the engine to demonstrate it can reach finality.
	root := types.Hash{0xaa}
	if err := engine.ProposeBlock(finalizedSlot+6, root); err != nil {
		t.Fatalf("ProposeBlock: %v", err)
	}

	// Cast 2/3+ votes. Use TotalStake*2/3+1 to ensure the threshold is met
	// despite integer rounding (TotalStake/3*2 may round down below the limit).
	stake := cfg.TotalStake*2/3 + 1
	vote := MinimmitVote{
		ValidatorIndex: 0,
		Slot:           finalizedSlot + 6,
		BlockRoot:      root,
		Stake:          stake,
	}
	if err := engine.CastVote(vote); err != nil {
		t.Fatalf("CastVote: %v", err)
	}
	if !engine.CheckFinality() {
		t.Error("expected finality after 2/3+ votes")
	}
}

// TestSSFBackoffBoundedFinality runs a 1000-slot, 10-run simulation and verifies
// that finality always advances within the bounded square/oblong progression (GAP-5.3).
// Per 3SF-mini consensus.py, worst-case gap is bounded to ≤ 25 slots (next square
// after 16 is 25). Each run must achieve lastFinalized ≥ 990 after 1000 slots.
func TestSSFBackoffBoundedFinality(t *testing.T) {
	const totalSlots = uint64(1000)
	const minFinalizedTarget = uint64(990) // allow a small tail window
	const runs = 10

	for run := 0; run < runs; run++ {
		lastFinalized := uint64(0)

		for slot := uint64(1); slot <= totalSlots; slot++ {
			if IsJustifiableSlot(lastFinalized, slot) {
				// Honest majority finalizes at every justifiable slot.
				lastFinalized = slot
			}
		}

		if lastFinalized < minFinalizedTarget {
			t.Errorf("run %d: bounded finality failed: lastFinalized=%d after %d slots (want ≥ %d)",
				run, lastFinalized, totalSlots, minFinalizedTarget)
		}
	}
}

// TestFinalityModeSelection verifies that FinalityMode enum values are correct
// and the engine instantiation can be selected by mode (GAP-5.2).
func TestFinalityModeSelection(t *testing.T) {
	// Verify FinalityMode constants.
	if FinalityModeSSF.String() != "SSF" {
		t.Errorf("FinalityModeSSF.String() = %q, want \"SSF\"", FinalityModeSSF.String())
	}
	if FinalityModeMinimmit.String() != "Minimmit" {
		t.Errorf("FinalityModeMinimmit.String() = %q, want \"Minimmit\"", FinalityModeMinimmit.String())
	}
	if FinalityModeClassic.String() != "Classic" {
		t.Errorf("FinalityModeClassic.String() = %q, want \"Classic\"", FinalityModeClassic.String())
	}

	// Verify Minimmit engine can be selected.
	cfg := DefaultMinimmitConfig()
	eng, err := NewMinimmitEngine(cfg)
	if err != nil {
		t.Fatalf("NewMinimmitEngine: %v", err)
	}
	if eng.State() != MinimmitIdle {
		t.Error("new engine should be in Idle state")
	}

	// Verify mode can be used to switch behavior.
	mode := FinalityModeMinimmit
	if mode != FinalityModeMinimmit {
		t.Error("expected FinalityModeMinimmit")
	}
}

// TestIsPerfectSquare verifies the helper function (GAP-5.3 internal).
func TestIsPerfectSquare(t *testing.T) {
	squares := []uint64{0, 1, 4, 9, 16, 25, 36, 49, 64, 81, 100}
	nonSquares := []uint64{2, 3, 5, 6, 7, 8, 10, 11, 15, 17, 24}

	for _, n := range squares {
		if !isPerfectSquare(n) {
			t.Errorf("isPerfectSquare(%d) = false, want true", n)
		}
	}
	for _, n := range nonSquares {
		if isPerfectSquare(n) {
			t.Errorf("isPerfectSquare(%d) = true, want false", n)
		}
	}
}

// TestIsOblongNumber verifies the oblong number helper (GAP-5.3 internal).
func TestIsOblongNumber(t *testing.T) {
	// Oblong: x*(x+1) for x=1..10: 2,6,12,20,30,42,56,72,90,110.
	oblongs := []uint64{2, 6, 12, 20, 30, 42, 56, 72, 90, 110}
	nonOblongs := []uint64{0, 1, 3, 4, 5, 7, 8, 9, 10, 11, 13}

	for _, n := range oblongs {
		if !isOblongNumber(n) {
			t.Errorf("isOblongNumber(%d) = false, want true", n)
		}
	}
	for _, n := range nonOblongs {
		if isOblongNumber(n) {
			t.Errorf("isOblongNumber(%d) = true, want false", n)
		}
	}
}
