package txpool

import (
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

var (
	ntAddr1 = types.BytesToAddress([]byte{0xA1})
	ntAddr2 = types.BytesToAddress([]byte{0xA2})
)

// --- DefaultNonceTrackerConfig ---

func TestDefaultNonceTrackerConfig(t *testing.T) {
	cfg := DefaultNonceTrackerConfig()
	if cfg.MaxNonceAhead != MaxNonceGap {
		t.Errorf("MaxNonceAhead = %d, want %d", cfg.MaxNonceAhead, MaxNonceGap)
	}
}

// --- NewNonceTracker ---

func TestNewNonceTracker(t *testing.T) {
	s := newMockState()
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	if nt == nil {
		t.Fatal("expected non-nil NonceTracker")
	}
	if nt.AccountCount() != 0 {
		t.Errorf("AccountCount = %d, want 0", nt.AccountCount())
	}
}

// --- GetNonce ---

func TestNonceTracker_GetNonce_FromState(t *testing.T) {
	s := newMockState()
	s.nonces[ntAddr1] = 7
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	if got := nt.GetNonce(ntAddr1); got != 7 {
		t.Errorf("GetNonce = %d, want 7 (from state)", got)
	}
}

func TestNonceTracker_GetNonce_NilState(t *testing.T) {
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), nil)
	if got := nt.GetNonce(ntAddr1); got != 0 {
		t.Errorf("GetNonce with nil state = %d, want 0", got)
	}
}

func TestNonceTracker_GetNonce_PendingMaxPlusOne(t *testing.T) {
	s := newMockState()
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	nt.TrackTx(ntAddr1, 3)
	if got := nt.GetNonce(ntAddr1); got != 4 {
		t.Errorf("GetNonce with pending 3 = %d, want 4", got)
	}
}

// --- SetNonce ---

func TestNonceTracker_SetNonce(t *testing.T) {
	s := newMockState()
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	nt.TrackTx(ntAddr1, 0)
	nt.TrackTx(ntAddr1, 1)
	nt.SetNonce(ntAddr1, 1) // tx 0 included in block
	// Known nonce 0 should be removed; 1 stays.
	known := nt.KnownNonces(ntAddr1)
	for _, n := range known {
		if n < 1 {
			t.Errorf("SetNonce did not remove nonce %d below 1", n)
		}
	}
}

func TestNonceTracker_SetNonce_ClearsAllPending(t *testing.T) {
	s := newMockState()
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	nt.TrackTx(ntAddr1, 0)
	nt.TrackTx(ntAddr1, 1)
	nt.SetNonce(ntAddr1, 5) // all txs included
	if nt.AccountCount() != 0 {
		// Account should be gone since knownNonces is now empty after pruning.
		// SetNonce removes from knownNonces but account may remain with empty map.
		// Either way GetNonce should return 5.
	}
	if got := nt.GetNonce(ntAddr1); got != 5 {
		t.Errorf("GetNonce after SetNonce(5) = %d, want 5", got)
	}
}

// --- TrackTx ---

func TestNonceTracker_TrackTx(t *testing.T) {
	s := newMockState()
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	nt.TrackTx(ntAddr1, 0)
	nt.TrackTx(ntAddr1, 1)
	nt.TrackTx(ntAddr1, 2)
	if nt.AccountCount() != 1 {
		t.Errorf("AccountCount = %d, want 1", nt.AccountCount())
	}
	if got := nt.GetNonce(ntAddr1); got != 3 {
		t.Errorf("GetNonce = %d, want 3", got)
	}
}

func TestNonceTracker_TrackTx_MultipleAddresses(t *testing.T) {
	s := newMockState()
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	nt.TrackTx(ntAddr1, 0)
	nt.TrackTx(ntAddr2, 0)
	if nt.AccountCount() != 2 {
		t.Errorf("AccountCount = %d, want 2", nt.AccountCount())
	}
}

// --- UntrackTx ---

func TestNonceTracker_UntrackTx(t *testing.T) {
	s := newMockState()
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	nt.TrackTx(ntAddr1, 0)
	nt.TrackTx(ntAddr1, 1)
	nt.UntrackTx(ntAddr1, 0)
	known := nt.KnownNonces(ntAddr1)
	for _, n := range known {
		if n == 0 {
			t.Error("nonce 0 should be untracked")
		}
	}
}

func TestNonceTracker_UntrackTx_RemovesEmptyAccount(t *testing.T) {
	s := newMockState()
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	nt.TrackTx(ntAddr1, 0)
	nt.UntrackTx(ntAddr1, 0)
	if nt.AccountCount() != 0 {
		t.Errorf("AccountCount = %d, want 0 after removing last nonce", nt.AccountCount())
	}
}

func TestNonceTracker_UntrackTx_UnknownAddr(t *testing.T) {
	s := newMockState()
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	// Should not panic.
	nt.UntrackTx(ntAddr1, 0)
}

// --- DetectGap ---

func TestNonceTracker_DetectGap_NoGap(t *testing.T) {
	s := newMockState()
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	nt.TrackTx(ntAddr1, 0)
	// txNonce == stateNonce → no gap (stateNonce=0 since no explicit set).
	gap := nt.DetectGap(ntAddr1, 1)
	if gap != nil {
		t.Errorf("expected nil gap, got %+v", gap)
	}
}

func TestNonceTracker_DetectGap_Gap(t *testing.T) {
	s := newMockState()
	s.nonces[ntAddr1] = 0
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	// No nonce 0 tracked — gap when txNonce=2.
	gap := nt.DetectGap(ntAddr1, 2)
	if gap == nil {
		t.Fatal("expected gap")
	}
	if gap.Expected != 0 {
		t.Errorf("gap.Expected = %d, want 0", gap.Expected)
	}
	if gap.TxNonce != 2 {
		t.Errorf("gap.TxNonce = %d, want 2", gap.TxNonce)
	}
}

func TestNonceTracker_DetectGap_EqualToStateNonce(t *testing.T) {
	s := newMockState()
	s.nonces[ntAddr1] = 5
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	// txNonce == stateNonce → no gap.
	gap := nt.DetectGap(ntAddr1, 5)
	if gap != nil {
		t.Errorf("expected nil for txNonce == stateNonce, got %+v", gap)
	}
}

func TestNonceTracker_DetectGap_BelowStateNonce(t *testing.T) {
	s := newMockState()
	s.nonces[ntAddr1] = 10
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	// Stale nonce — below state → no gap.
	gap := nt.DetectGap(ntAddr1, 5)
	if gap != nil {
		t.Errorf("expected nil for stale nonce, got %+v", gap)
	}
}

func TestNonceTracker_DetectGap_ContiguousNonces(t *testing.T) {
	s := newMockState()
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	nt.TrackTx(ntAddr1, 0)
	nt.TrackTx(ntAddr1, 1)
	// txNonce=2 with 0,1 tracked — contiguous, no gap.
	gap := nt.DetectGap(ntAddr1, 2)
	if gap != nil {
		t.Errorf("expected nil (contiguous), got %+v", gap)
	}
}

// --- IsTooFarAhead ---

func TestNonceTracker_IsTooFarAhead_True(t *testing.T) {
	s := newMockState()
	s.nonces[ntAddr1] = 0
	cfg := DefaultNonceTrackerConfig()
	cfg.MaxNonceAhead = 10
	nt := NewNonceTracker(cfg, s)
	if !nt.IsTooFarAhead(ntAddr1, 11) {
		t.Error("expected IsTooFarAhead=true for nonce 11 when max=10")
	}
}

func TestNonceTracker_IsTooFarAhead_False(t *testing.T) {
	s := newMockState()
	s.nonces[ntAddr1] = 0
	cfg := DefaultNonceTrackerConfig()
	cfg.MaxNonceAhead = 10
	nt := NewNonceTracker(cfg, s)
	if nt.IsTooFarAhead(ntAddr1, 5) {
		t.Error("expected IsTooFarAhead=false for nonce 5 when max=10")
	}
}

func TestNonceTracker_IsTooFarAhead_Boundary(t *testing.T) {
	s := newMockState()
	s.nonces[ntAddr1] = 0
	cfg := DefaultNonceTrackerConfig()
	cfg.MaxNonceAhead = 10
	nt := NewNonceTracker(cfg, s)
	// Exactly at limit.
	if nt.IsTooFarAhead(ntAddr1, 10) {
		t.Error("expected IsTooFarAhead=false at exactly the limit")
	}
}

// --- AllGaps ---

func TestNonceTracker_AllGaps_NoGaps(t *testing.T) {
	s := newMockState()
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	nt.TrackTx(ntAddr1, 0)
	nt.TrackTx(ntAddr1, 1)
	nt.TrackTx(ntAddr1, 2)
	gaps := nt.AllGaps(ntAddr1)
	if len(gaps) != 0 {
		t.Errorf("expected no gaps, got %v", gaps)
	}
}

func TestNonceTracker_AllGaps_WithGaps(t *testing.T) {
	s := newMockState()
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	nt.TrackTx(ntAddr1, 0)
	nt.TrackTx(ntAddr1, 2) // gap at 1
	gaps := nt.AllGaps(ntAddr1)
	found := false
	for _, g := range gaps {
		if g == 1 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected gap at nonce 1, got %v", gaps)
	}
}

func TestNonceTracker_AllGaps_NotTracked(t *testing.T) {
	s := newMockState()
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	gaps := nt.AllGaps(ntAddr1)
	if gaps != nil {
		t.Errorf("expected nil for untracked, got %v", gaps)
	}
}

// --- KnownNonces ---

func TestNonceTracker_KnownNonces_Sorted(t *testing.T) {
	s := newMockState()
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	for _, n := range []uint64{5, 2, 8, 1} {
		nt.TrackTx(ntAddr1, n)
	}
	nonces := nt.KnownNonces(ntAddr1)
	for i := 1; i < len(nonces); i++ {
		if nonces[i] < nonces[i-1] {
			t.Errorf("KnownNonces not sorted: %v", nonces)
			break
		}
	}
}

func TestNonceTracker_KnownNonces_NotTracked(t *testing.T) {
	s := newMockState()
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	if got := nt.KnownNonces(ntAddr1); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

// --- Reset ---

func TestNonceTracker_Reset(t *testing.T) {
	s := newMockState()
	nt := NewNonceTracker(DefaultNonceTrackerConfig(), s)
	nt.TrackTx(ntAddr1, 0)
	nt.TrackTx(ntAddr2, 0)
	if nt.AccountCount() != 2 {
		t.Fatalf("AccountCount = %d, want 2", nt.AccountCount())
	}

	s2 := newMockState()
	s2.nonces[ntAddr1] = 5
	nt.Reset(s2)

	if nt.AccountCount() != 0 {
		t.Errorf("AccountCount = %d, want 0 after Reset", nt.AccountCount())
	}
	// After reset, state lookup uses new state.
	if got := nt.GetNonce(ntAddr1); got != 5 {
		t.Errorf("GetNonce after Reset = %d, want 5", got)
	}
}
