package txpool

import (
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

var (
	acctAddr1 = types.BytesToAddress([]byte{0x01})
	acctAddr2 = types.BytesToAddress([]byte{0x02})
)

// --- AcctInfo helpers ---

func TestAcctInfo_AvailableBalance_Basic(t *testing.T) {
	ai := &AcctInfo{
		StateBalance:    big.NewInt(1000),
		ReservedBalance: big.NewInt(300),
	}
	avail := ai.AvailableBalance()
	if avail.Cmp(big.NewInt(700)) != 0 {
		t.Errorf("AvailableBalance = %s, want 700", avail)
	}
}

func TestAcctInfo_AvailableBalance_NoNegative(t *testing.T) {
	ai := &AcctInfo{
		StateBalance:    big.NewInt(100),
		ReservedBalance: big.NewInt(500),
	}
	avail := ai.AvailableBalance()
	if avail.Sign() < 0 {
		t.Errorf("AvailableBalance should not be negative, got %s", avail)
	}
	if avail.Sign() != 0 {
		t.Errorf("AvailableBalance = %s, want 0", avail)
	}
}

func TestAcctInfo_AvailableBalance_NilStateBalance(t *testing.T) {
	ai := &AcctInfo{StateBalance: nil, ReservedBalance: big.NewInt(100)}
	avail := ai.AvailableBalance()
	if avail.Sign() != 0 {
		t.Errorf("expected 0 for nil balance, got %s", avail)
	}
}

func TestAcctInfo_HasBalanceDeficit_True(t *testing.T) {
	ai := &AcctInfo{
		StateBalance:    big.NewInt(100),
		ReservedBalance: big.NewInt(200),
	}
	if !ai.HasBalanceDeficit() {
		t.Error("expected HasBalanceDeficit=true")
	}
}

func TestAcctInfo_HasBalanceDeficit_False(t *testing.T) {
	ai := &AcctInfo{
		StateBalance:    big.NewInt(1000),
		ReservedBalance: big.NewInt(200),
	}
	if ai.HasBalanceDeficit() {
		t.Error("expected HasBalanceDeficit=false")
	}
}

func TestAcctInfo_HasBalanceDeficit_NilStateBalance(t *testing.T) {
	ai := &AcctInfo{StateBalance: nil, ReservedBalance: big.NewInt(100)}
	if !ai.HasBalanceDeficit() {
		t.Error("expected deficit when balance is nil")
	}
}

func TestAcctInfo_HasBalanceDeficit_NilReserved(t *testing.T) {
	ai := &AcctInfo{StateBalance: nil, ReservedBalance: nil}
	if ai.HasBalanceDeficit() {
		t.Error("expected no deficit when both nil")
	}
}

func TestAcctInfo_NonceGaps_NoGaps(t *testing.T) {
	ai := &AcctInfo{
		StateNonce:   2,
		PendingNonce: 5,
		PendingTxs:   map[uint64]*big.Int{2: big.NewInt(1), 3: big.NewInt(1), 4: big.NewInt(1)},
	}
	gaps := ai.NonceGaps()
	if len(gaps) != 0 {
		t.Errorf("expected no gaps, got %v", gaps)
	}
}

func TestAcctInfo_NonceGaps_WithGap(t *testing.T) {
	ai := &AcctInfo{
		StateNonce:   0,
		PendingNonce: 3,
		PendingTxs:   map[uint64]*big.Int{0: big.NewInt(1), 2: big.NewInt(1)}, // missing nonce 1
	}
	gaps := ai.NonceGaps()
	if len(gaps) != 1 || gaps[0] != 1 {
		t.Errorf("expected gap at nonce 1, got %v", gaps)
	}
}

// --- NewAcctTrack / Track / Untrack ---

func TestNewAcctTrack(t *testing.T) {
	s := newMockState()
	at := NewAcctTrack(s)
	if at.TrackedCount() != 0 {
		t.Errorf("TrackedCount = %d, want 0", at.TrackedCount())
	}
}

func TestAcctTrack_Track(t *testing.T) {
	s := newMockState()
	s.nonces[acctAddr1] = 5
	s.balances[acctAddr1] = big.NewInt(1000)
	at := NewAcctTrack(s)
	at.Track(acctAddr1)
	if at.TrackedCount() != 1 {
		t.Errorf("TrackedCount = %d, want 1", at.TrackedCount())
	}
	nonce := at.GetPendingNonce(acctAddr1)
	if nonce != 5 {
		t.Errorf("PendingNonce = %d, want 5", nonce)
	}
}

func TestAcctTrack_Track_Idempotent(t *testing.T) {
	s := newMockState()
	at := NewAcctTrack(s)
	at.Track(acctAddr1)
	at.Track(acctAddr1)
	if at.TrackedCount() != 1 {
		t.Errorf("TrackedCount = %d, want 1 (idempotent)", at.TrackedCount())
	}
}

func TestAcctTrack_Untrack(t *testing.T) {
	s := newMockState()
	at := NewAcctTrack(s)
	at.Track(acctAddr1)
	at.Track(acctAddr2)
	at.Untrack(acctAddr1)
	if at.TrackedCount() != 1 {
		t.Errorf("TrackedCount = %d, want 1", at.TrackedCount())
	}
}

// --- AddPendingTx ---

func TestAcctTrack_AddPendingTx(t *testing.T) {
	s := newMockState()
	s.balances[acctAddr1] = big.NewInt(1e18)
	at := NewAcctTrack(s)

	tx := makeTxFrom(acctAddr1, 0, 1e9, 21000)
	if err := at.AddPendingTx(acctAddr1, tx); err != nil {
		t.Fatalf("AddPendingTx: %v", err)
	}
	if at.GetPendingNonce(acctAddr1) != 1 {
		t.Errorf("PendingNonce = %d, want 1", at.GetPendingNonce(acctAddr1))
	}
}

func TestAcctTrack_AddPendingTx_MultipleContiguous(t *testing.T) {
	s := newMockState()
	s.balances[acctAddr1] = big.NewInt(1e18)
	at := NewAcctTrack(s)

	for i := range uint64(5) {
		tx := makeTxFrom(acctAddr1, i, 1e9, 21000)
		at.AddPendingTx(acctAddr1, tx)
	}
	if at.GetPendingNonce(acctAddr1) != 5 {
		t.Errorf("PendingNonce = %d, want 5", at.GetPendingNonce(acctAddr1))
	}
}

func TestAcctTrack_AddPendingTx_GapDoesNotAdvancePendingNonce(t *testing.T) {
	s := newMockState()
	s.balances[acctAddr1] = big.NewInt(1e18)
	at := NewAcctTrack(s)

	// Add nonce 0 and nonce 2 — gap at 1.
	tx0 := makeTxFrom(acctAddr1, 0, 1e9, 21000)
	tx2 := makeTxFrom(acctAddr1, 2, 1e9, 21000)
	at.AddPendingTx(acctAddr1, tx0)
	at.AddPendingTx(acctAddr1, tx2)

	// PendingNonce should stop at 1 (gap).
	if at.GetPendingNonce(acctAddr1) != 1 {
		t.Errorf("PendingNonce = %d, want 1 (gap at 1)", at.GetPendingNonce(acctAddr1))
	}
}

// --- RemovePendingTx ---

func TestAcctTrack_RemovePendingTx(t *testing.T) {
	s := newMockState()
	s.balances[acctAddr1] = big.NewInt(1e18)
	at := NewAcctTrack(s)
	tx := makeTxFrom(acctAddr1, 0, 1e9, 21000)
	at.AddPendingTx(acctAddr1, tx)
	if err := at.RemovePendingTx(acctAddr1, 0); err != nil {
		t.Fatalf("RemovePendingTx: %v", err)
	}
	// Account should be removed when empty.
	if at.TrackedCount() != 0 {
		t.Errorf("TrackedCount = %d, want 0 after removing last tx", at.TrackedCount())
	}
}

func TestAcctTrack_RemovePendingTx_NotTracked(t *testing.T) {
	s := newMockState()
	at := NewAcctTrack(s)
	err := at.RemovePendingTx(acctAddr1, 0)
	if err != ErrAcctNotTracked {
		t.Errorf("expected ErrAcctNotTracked, got %v", err)
	}
}

func TestAcctTrack_RemovePendingTx_NonExistentNonce(t *testing.T) {
	s := newMockState()
	s.balances[acctAddr1] = big.NewInt(1e18)
	at := NewAcctTrack(s)
	tx := makeTxFrom(acctAddr1, 0, 1e9, 21000)
	at.AddPendingTx(acctAddr1, tx)
	// Remove nonce 99 which doesn't exist — no error expected.
	if err := at.RemovePendingTx(acctAddr1, 99); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- ReplacePendingTx ---

func TestAcctTrack_ReplacePendingTx(t *testing.T) {
	s := newMockState()
	s.balances[acctAddr1] = big.NewInt(1e18)
	at := NewAcctTrack(s)

	old := makeTxFrom(acctAddr1, 0, 1e9, 21000)
	at.AddPendingTx(acctAddr1, old)

	info1 := at.GetInfo(acctAddr1)
	oldReserved := new(big.Int).Set(info1.ReservedBalance)

	replacement := makeTxFrom(acctAddr1, 0, 2e9, 21000) // higher gas price
	at.ReplacePendingTx(acctAddr1, replacement)

	info2 := at.GetInfo(acctAddr1)
	if info2.ReservedBalance.Cmp(oldReserved) <= 0 {
		t.Error("expected higher reserved balance after replacement with higher fee")
	}
}

// --- GetInfo ---

func TestAcctTrack_GetInfo_NotTracked(t *testing.T) {
	s := newMockState()
	at := NewAcctTrack(s)
	info := at.GetInfo(acctAddr1)
	if info != nil {
		t.Errorf("expected nil for untracked addr, got %+v", info)
	}
}

func TestAcctTrack_GetInfo_ReturnsCopy(t *testing.T) {
	s := newMockState()
	s.balances[acctAddr1] = big.NewInt(1000)
	at := NewAcctTrack(s)
	at.Track(acctAddr1)
	info := at.GetInfo(acctAddr1)
	// Modifying the copy should not affect the tracker.
	info.StateBalance.SetInt64(0)
	info2 := at.GetInfo(acctAddr1)
	if info2.StateBalance.Cmp(big.NewInt(1000)) != 0 {
		t.Errorf("GetInfo should return an independent copy, got %s", info2.StateBalance)
	}
}

// --- CheckBalanceDeficit ---

func TestAcctTrack_CheckBalanceDeficit_OK(t *testing.T) {
	s := newMockState()
	s.balances[acctAddr1] = big.NewInt(1e18)
	at := NewAcctTrack(s)
	at.Track(acctAddr1)
	if err := at.CheckBalanceDeficit(acctAddr1); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAcctTrack_CheckBalanceDeficit_Deficit(t *testing.T) {
	s := newMockState()
	s.balances[acctAddr1] = big.NewInt(100) // very low
	at := NewAcctTrack(s)

	// Add tx that costs more than balance.
	tx := makeTxFrom(acctAddr1, 0, 1e10, 1_000_000)
	at.AddPendingTx(acctAddr1, tx)

	if err := at.CheckBalanceDeficit(acctAddr1); err != ErrAcctInsufficientBal {
		t.Errorf("expected ErrAcctInsufficientBal, got %v", err)
	}
}

// --- DetectNonceGaps ---

func TestAcctTrack_DetectNonceGaps_NoGaps(t *testing.T) {
	s := newMockState()
	s.balances[acctAddr1] = big.NewInt(1e18)
	at := NewAcctTrack(s)
	tx0 := makeTxFrom(acctAddr1, 0, 1e9, 21000)
	tx1 := makeTxFrom(acctAddr1, 1, 1e9, 21000)
	at.AddPendingTx(acctAddr1, tx0)
	at.AddPendingTx(acctAddr1, tx1)
	gaps := at.DetectNonceGaps(acctAddr1)
	if len(gaps) != 0 {
		t.Errorf("expected no gaps, got %v", gaps)
	}
}

func TestAcctTrack_DetectNonceGaps_WithGap(t *testing.T) {
	s := newMockState()
	s.balances[acctAddr1] = big.NewInt(1e18)
	at := NewAcctTrack(s)
	// Directly set up AcctInfo with PendingNonce beyond a missing nonce.
	at.mu.Lock()
	at.accts[acctAddr1] = &AcctInfo{
		Address:         acctAddr1,
		StateNonce:      0,
		PendingNonce:    3, // claims nonces 0,1,2 are pending
		StateBalance:    big.NewInt(1e18),
		ReservedBalance: new(big.Int),
		PendingTxs:      map[uint64]*big.Int{0: big.NewInt(1), 2: big.NewInt(1)}, // missing 1
	}
	at.mu.Unlock()
	gaps := at.DetectNonceGaps(acctAddr1)
	if len(gaps) != 1 || gaps[0] != 1 {
		t.Errorf("expected gap [1], got %v", gaps)
	}
}

// --- ResetOnReorg ---

func TestAcctTrack_ResetOnReorg(t *testing.T) {
	s := newMockState()
	s.balances[acctAddr1] = big.NewInt(1e18)
	at := NewAcctTrack(s)

	// Add txs with nonces 0,1,2.
	for i := range uint64(3) {
		at.AddPendingTx(acctAddr1, makeTxFrom(acctAddr1, i, 1e9, 21000))
	}

	// Reorg: new state has nonce 2 (txs 0,1 included in block).
	s2 := newMockState()
	s2.nonces[acctAddr1] = 2
	s2.balances[acctAddr1] = big.NewInt(1e18)
	invalidated := at.ResetOnReorg(s2)

	if len(invalidated) == 0 {
		t.Error("expected at least one invalidated account")
	}
}

func TestAcctTrack_ResetOnReorg_AllTxsIncluded(t *testing.T) {
	s := newMockState()
	s.balances[acctAddr1] = big.NewInt(1e18)
	at := NewAcctTrack(s)

	at.AddPendingTx(acctAddr1, makeTxFrom(acctAddr1, 0, 1e9, 21000))

	s2 := newMockState()
	s2.nonces[acctAddr1] = 1 // tx 0 included
	s2.balances[acctAddr1] = big.NewInt(1e18)
	at.ResetOnReorg(s2)

	// Account should be cleaned up.
	if at.TrackedCount() != 0 {
		t.Errorf("TrackedCount = %d, want 0 after all txs included", at.TrackedCount())
	}
}

// --- RefreshBatch / MarkDirty ---

func TestAcctTrack_RefreshBatch(t *testing.T) {
	s := newMockState()
	s.nonces[acctAddr1] = 3
	s.balances[acctAddr1] = big.NewInt(1000)
	at := NewAcctTrack(s)
	at.Track(acctAddr1)

	// Update state nonce externally.
	s.nonces[acctAddr1] = 7
	at.RefreshBatch([]types.Address{acctAddr1})

	nonce := at.GetPendingNonce(acctAddr1)
	if nonce != 7 {
		t.Errorf("PendingNonce after RefreshBatch = %d, want 7", nonce)
	}
}

func TestAcctTrack_RefreshBatch_Untracked(t *testing.T) {
	s := newMockState()
	at := NewAcctTrack(s)
	// Should not panic for untracked addr.
	at.RefreshBatch([]types.Address{acctAddr1})
}

func TestAcctTrack_MarkDirty(t *testing.T) {
	s := newMockState()
	s.nonces[acctAddr1] = 1
	s.balances[acctAddr1] = big.NewInt(100)
	at := NewAcctTrack(s)
	at.Track(acctAddr1)

	// Mark dirty then update state.
	at.MarkDirty(acctAddr1)
	s.nonces[acctAddr1] = 5
	s.balances[acctAddr1] = big.NewInt(999)

	// Trigger reload via RefreshBatch (which calls computePendingNonce).
	at.RefreshBatch([]types.Address{acctAddr1})
	nonce := at.GetPendingNonce(acctAddr1)
	if nonce != 5 {
		t.Errorf("expected reloaded nonce 5 after MarkDirty+RefreshBatch, got %d", nonce)
	}
}

// --- TrackedAddresses ---

func TestAcctTrack_TrackedAddresses(t *testing.T) {
	s := newMockState()
	at := NewAcctTrack(s)
	at.Track(acctAddr1)
	at.Track(acctAddr2)
	addrs := at.TrackedAddresses()
	if len(addrs) != 2 {
		t.Errorf("TrackedAddresses len = %d, want 2", len(addrs))
	}
}

// --- AccountsWithDeficit / AccountsWithGaps ---

func TestAcctTrack_AccountsWithDeficit(t *testing.T) {
	s := newMockState()
	s.balances[acctAddr1] = big.NewInt(50)  // not enough
	s.balances[acctAddr2] = big.NewInt(1e18) // plenty
	at := NewAcctTrack(s)

	at.AddPendingTx(acctAddr1, makeTxFrom(acctAddr1, 0, 1e10, 1_000_000)) // costs more than balance
	at.AddPendingTx(acctAddr2, makeTxFrom(acctAddr2, 0, 1e9, 21000))

	deficit := at.AccountsWithDeficit()
	if len(deficit) != 1 || deficit[0] != acctAddr1 {
		t.Errorf("AccountsWithDeficit = %v, want [addr1]", deficit)
	}
}

func TestAcctTrack_AccountsWithGaps(t *testing.T) {
	s := newMockState()
	s.balances[acctAddr1] = big.NewInt(1e18)
	s.balances[acctAddr2] = big.NewInt(1e18)
	at := NewAcctTrack(s)

	// addr1: directly set PendingNonce=3 with a hole at nonce 1.
	at.mu.Lock()
	at.accts[acctAddr1] = &AcctInfo{
		Address:         acctAddr1,
		StateNonce:      0,
		PendingNonce:    3,
		StateBalance:    big.NewInt(1e18),
		ReservedBalance: new(big.Int),
		PendingTxs:      map[uint64]*big.Int{0: big.NewInt(1), 2: big.NewInt(1)},
	}
	// addr2: contiguous nonces 0,1,2 — no gaps.
	at.accts[acctAddr2] = &AcctInfo{
		Address:         acctAddr2,
		StateNonce:      0,
		PendingNonce:    3,
		StateBalance:    big.NewInt(1e18),
		ReservedBalance: new(big.Int),
		PendingTxs:      map[uint64]*big.Int{0: big.NewInt(1), 1: big.NewInt(1), 2: big.NewInt(1)},
	}
	at.mu.Unlock()

	gaps := at.AccountsWithGaps()
	if len(gaps) != 1 || gaps[0] != acctAddr1 {
		t.Errorf("AccountsWithGaps = %v, want [addr1]", gaps)
	}
}

// --- SortedPendingNonces ---

func TestAcctTrack_SortedPendingNonces(t *testing.T) {
	s := newMockState()
	s.balances[acctAddr1] = big.NewInt(1e18)
	at := NewAcctTrack(s)

	for _, n := range []uint64{3, 1, 0, 2} {
		at.AddPendingTx(acctAddr1, makeTxFrom(acctAddr1, n, 1e9, 21000))
	}

	nonces := at.SortedPendingNonces(acctAddr1)
	for i, want := range []uint64{0, 1, 2, 3} {
		if nonces[i] != want {
			t.Errorf("nonces[%d] = %d, want %d", i, nonces[i], want)
		}
	}
}

func TestAcctTrack_SortedPendingNonces_NotTracked(t *testing.T) {
	s := newMockState()
	at := NewAcctTrack(s)
	nonces := at.SortedPendingNonces(acctAddr1)
	if nonces != nil {
		t.Errorf("expected nil for untracked addr, got %v", nonces)
	}
}
