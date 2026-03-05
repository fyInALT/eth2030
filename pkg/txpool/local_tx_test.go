package txpool

import (
	"math/big"
	"strings"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// makeLocalTx builds a type-0x08 LocalTx with a pre-set sender.
func makeLocalTx(from types.Address, nonce uint64, gas uint64, scopeHint []byte) *types.Transaction {
	to := types.BytesToAddress([]byte{0xde, 0xad})
	tx := types.NewLocalTx(
		big.NewInt(1),
		nonce,
		&to,
		big.NewInt(0),
		gas,
		big.NewInt(1),
		big.NewInt(2),
		nil,
		scopeHint,
	)
	tx.SetSender(from)
	return tx
}

// newPoolWithLocalTx creates a pool with AllowLocalTx set as specified.
func newPoolWithLocalTx(allow bool) *TxPool {
	cfg := DefaultConfig()
	cfg.AllowLocalTx = allow
	st := newMockState()
	st.balances[testSender] = richBalance
	return New(cfg, st)
}

// --- BB-2.2: AllowLocalTx gate ---

// TestLocalTx_RejectedWhenFlagOff verifies the pool rejects type-0x08 by default.
func TestLocalTx_RejectedWhenFlagOff(t *testing.T) {
	pool := newPoolWithLocalTx(false)
	tx := makeLocalTx(testSender, 0, 21000, []byte{0x0a})

	err := pool.AddLocal(tx)
	if err == nil {
		t.Fatal("expected rejection when AllowLocalTx=false, got nil")
	}
	if !strings.Contains(err.Error(), "local tx") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestLocalTx_AcceptedWhenFlagOn verifies the pool accepts type-0x08 when enabled.
func TestLocalTx_AcceptedWhenFlagOn(t *testing.T) {
	pool := newPoolWithLocalTx(true)
	tx := makeLocalTx(testSender, 0, 21000, []byte{0x0a})

	if err := pool.AddLocal(tx); err != nil {
		t.Fatalf("AddLocal() error: %v", err)
	}
	if pool.lookup.Get(tx.Hash()) == nil {
		t.Fatal("LocalTx not found in pool after AddLocal")
	}
}

// TestLocalTx_AddRemoteRejectedWhenFlagOff verifies remote path also gates.
func TestLocalTx_AddRemoteRejectedWhenFlagOff(t *testing.T) {
	pool := newPoolWithLocalTx(false)
	tx := makeLocalTx(testSender, 0, 21000, []byte{0x0a})
	if pool.AddRemote(tx) == nil {
		t.Fatal("expected rejection for remote LocalTx when AllowLocalTx=false")
	}
}

// TestLocalTx_DefaultConfigDisallows verifies DefaultConfig has AllowLocalTx=false.
func TestLocalTx_DefaultConfigDisallows(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.AllowLocalTx {
		t.Error("DefaultConfig.AllowLocalTx should be false")
	}
}

// TestLocalTx_MultipleScopes verifies multiple LocalTxs with different scopes
// can all be added when the flag is on.
func TestLocalTx_MultipleScopes(t *testing.T) {
	pool := newPoolWithLocalTx(true)

	senderA := types.BytesToAddress([]byte{0x0a, 0x00})
	senderB := types.BytesToAddress([]byte{0x0b, 0x00})
	pool.state.(*mockState).balances[senderA] = richBalance
	pool.state.(*mockState).balances[senderB] = richBalance

	// Different gas values ensure distinct tx hashes even if other fields match.
	txA := makeLocalTx(senderA, 0, 21000, []byte{0x0a})
	txB := makeLocalTx(senderB, 0, 22000, []byte{0x0b})

	if err := pool.AddLocal(txA); err != nil {
		t.Fatalf("AddLocal(txA): %v", err)
	}
	if err := pool.AddLocal(txB); err != nil {
		t.Fatalf("AddLocal(txB): %v", err)
	}
}

// TestLocalTx_NormalTxUnaffected verifies legacy txs still work when flag is off.
func TestLocalTx_NormalTxUnaffected(t *testing.T) {
	pool := newPoolWithLocalTx(false)
	tx := makeTx(0, 1, 21000)
	pool.state.(*mockState).balances[testSender] = richBalance
	if err := pool.AddLocal(tx); err != nil {
		t.Fatalf("legacy tx should still be accepted: %v", err)
	}
}

// TestLocalTx_GasCheckedAfterFlag verifies intrinsic gas check still applies.
func TestLocalTx_GasCheckedAfterFlag(t *testing.T) {
	pool := newPoolWithLocalTx(true)
	// Gas below 21000 intrinsic minimum.
	tx := makeLocalTx(testSender, 0, 100, []byte{0x0a})

	err := pool.AddLocal(tx)
	if err == nil {
		t.Fatal("expected intrinsic gas error, got nil")
	}
	if !strings.Contains(err.Error(), "intrinsic") {
		t.Errorf("want intrinsic gas error, got: %v", err)
	}
}
