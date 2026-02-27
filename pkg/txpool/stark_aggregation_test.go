package txpool

import (
	"testing"
	"time"

	"github.com/eth2030/eth2030/core/types"
)

func TestNewSTARKAggregator(t *testing.T) {
	agg := NewSTARKAggregator("node-1")
	if agg.peerID != "node-1" {
		t.Errorf("expected peer ID node-1, got %s", agg.peerID)
	}
	if agg.tickInterval != DefaultTickInterval {
		t.Errorf("expected tick interval %v, got %v", DefaultTickInterval, agg.tickInterval)
	}
	if agg.IsRunning() {
		t.Error("aggregator should not be running initially")
	}
}

func TestSTARKAggregatorAddRemove(t *testing.T) {
	agg := NewSTARKAggregator("node-1")

	var hash1, hash2 types.Hash
	hash1[0] = 0x01
	hash2[0] = 0x02

	agg.AddValidatedTx(hash1, []byte("proof1"), 21000)
	agg.AddValidatedTx(hash2, []byte("proof2"), 42000)

	if agg.PendingCount() != 2 {
		t.Errorf("expected 2 pending, got %d", agg.PendingCount())
	}

	agg.RemoveTx(hash1)
	if agg.PendingCount() != 1 {
		t.Errorf("expected 1 pending, got %d", agg.PendingCount())
	}
}

func TestSTARKAggregatorGenerateTick(t *testing.T) {
	agg := NewSTARKAggregator("node-1")

	// Empty tick should fail.
	_, err := agg.GenerateTick()
	if err != ErrAggNoTransactions {
		t.Errorf("expected ErrAggNoTransactions, got %v", err)
	}

	// Add some transactions.
	for i := 0; i < 5; i++ {
		var hash types.Hash
		hash[0] = byte(i + 1)
		agg.AddValidatedTx(hash, []byte("validation-proof"), uint64(21000*(i+1)))
	}

	tick, err := agg.GenerateTick()
	if err != nil {
		t.Fatal(err)
	}

	if len(tick.ValidTxHashes) != 5 {
		t.Errorf("expected 5 tx hashes, got %d", len(tick.ValidTxHashes))
	}
	if tick.AggregateProof == nil {
		t.Error("aggregate proof should not be nil")
	}
	if tick.PeerID != "node-1" {
		t.Errorf("expected peer ID node-1, got %s", tick.PeerID)
	}
	if tick.TickNumber != 1 {
		t.Errorf("expected tick number 1, got %d", tick.TickNumber)
	}
}

func TestSTARKAggregatorMergeTick(t *testing.T) {
	agg1 := NewSTARKAggregator("node-1")
	agg2 := NewSTARKAggregator("node-2")

	// Add tx to agg1.
	var hash1 types.Hash
	hash1[0] = 0x01
	agg1.AddValidatedTx(hash1, []byte("proof1"), 21000)

	// Generate tick from agg1.
	tick, err := agg1.GenerateTick()
	if err != nil {
		t.Fatal(err)
	}

	// Merge into agg2.
	err = agg2.MergeTick(tick)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSTARKAggregatorMergeNil(t *testing.T) {
	agg := NewSTARKAggregator("node-1")
	err := agg.MergeTick(nil)
	if err != ErrAggInvalidTick {
		t.Errorf("expected ErrAggInvalidTick, got %v", err)
	}
}

func TestSTARKAggregatorStartStop(t *testing.T) {
	agg := NewSTARKAggregatorWithInterval("node-1", 50*time.Millisecond)

	err := agg.Start()
	if err != nil {
		t.Fatal(err)
	}
	if !agg.IsRunning() {
		t.Error("aggregator should be running")
	}

	// Double start should fail.
	err = agg.Start()
	if err != ErrAggAlreadyRunning {
		t.Errorf("expected ErrAggAlreadyRunning, got %v", err)
	}

	agg.Stop()
	// Give goroutine time to exit.
	time.Sleep(100 * time.Millisecond)
	if agg.IsRunning() {
		t.Error("aggregator should not be running after stop")
	}
}

func TestSTARKAggregatorDiscardList(t *testing.T) {
	agg := NewSTARKAggregator("node-1")

	var hash1, hash2 types.Hash
	hash1[0] = 0x01
	hash2[0] = 0x02

	agg.AddValidatedTx(hash1, []byte("proof1"), 21000)
	agg.AddValidatedTx(hash2, []byte("proof2"), 42000)

	// Remove one tx before generating tick.
	agg.RemoveTx(hash1)

	tick, err := agg.GenerateTick()
	if err != nil {
		t.Fatal(err)
	}

	// Discard list should contain hash1.
	if len(tick.DiscardList) != 1 {
		t.Errorf("expected 1 discard, got %d", len(tick.DiscardList))
	}
	if tick.DiscardList[0] != hash1 {
		t.Error("discard list should contain hash1")
	}
}

func TestTickHash(t *testing.T) {
	agg := NewSTARKAggregator("node-1")
	var hash1 types.Hash
	hash1[0] = 0x01
	agg.AddValidatedTx(hash1, []byte("proof1"), 21000)

	tick, err := agg.GenerateTick()
	if err != nil {
		t.Fatal(err)
	}

	h1 := TickHash(tick)
	h2 := TickHash(tick)
	if h1 != h2 {
		t.Error("tick hash should be deterministic")
	}

	var zero types.Hash
	if h1 == zero {
		t.Error("tick hash should not be zero")
	}
}
