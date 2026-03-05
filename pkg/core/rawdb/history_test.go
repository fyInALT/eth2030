package rawdb

import (
	"bytes"
	"testing"
)

func TestWriteReadHistoryOldest(t *testing.T) {
	db := NewMemoryDB()

	// Before any write, oldest should be 0.
	oldest, err := ReadHistoryOldest(db)
	if err != nil {
		t.Fatalf("ReadHistoryOldest: %v", err)
	}
	if oldest != 0 {
		t.Fatalf("want oldest 0, got %d", oldest)
	}

	// Write oldest.
	if err := WriteHistoryOldest(db, 1000); err != nil {
		t.Fatalf("WriteHistoryOldest: %v", err)
	}

	oldest, err = ReadHistoryOldest(db)
	if err != nil {
		t.Fatalf("ReadHistoryOldest: %v", err)
	}
	if oldest != 1000 {
		t.Fatalf("want oldest 1000, got %d", oldest)
	}
}

func TestHistoryAvailable(t *testing.T) {
	db := NewMemoryDB()

	// No pruning: all blocks should be available.
	avail, err := HistoryAvailable(db, 0)
	if err != nil {
		t.Fatalf("HistoryAvailable: %v", err)
	}
	if !avail {
		t.Fatal("block 0 should be available before pruning")
	}

	// Set oldest to 500.
	WriteHistoryOldest(db, 500)

	// Block 499 should not be available.
	avail, err = HistoryAvailable(db, 499)
	if err != nil {
		t.Fatalf("HistoryAvailable: %v", err)
	}
	if avail {
		t.Fatal("block 499 should not be available after pruning to 500")
	}

	// Block 500 should be available.
	avail, err = HistoryAvailable(db, 500)
	if err != nil {
		t.Fatalf("HistoryAvailable: %v", err)
	}
	if !avail {
		t.Fatal("block 500 should be available")
	}

	// Block 1000 should be available.
	avail, err = HistoryAvailable(db, 1000)
	if err != nil {
		t.Fatalf("HistoryAvailable: %v", err)
	}
	if !avail {
		t.Fatal("block 1000 should be available")
	}
}

func TestPruneHistory(t *testing.T) {
	db := NewMemoryDB()

	// Write 10 blocks with bodies, receipts, and canonical hashes.
	for i := uint64(0); i < 10; i++ {
		hash := [32]byte{byte(i)}
		WriteCanonicalHash(db, i, hash)
		WriteBody(db, i, hash, []byte("body-data"))
		WriteReceipts(db, i, hash, []byte("receipt-data"))
		// Also write headers to verify they are preserved.
		WriteHeader(db, i, hash, []byte("header-data"))
	}

	// Prune with retention=5, head=9 -> threshold = 4 -> prune blocks 0-3.
	pruned, newOldest, err := PruneHistory(db, 9, 5)
	if err != nil {
		t.Fatalf("PruneHistory: %v", err)
	}
	if pruned != 4 {
		t.Fatalf("want 4 pruned, got %d", pruned)
	}
	if newOldest != 4 {
		t.Fatalf("want newOldest 4, got %d", newOldest)
	}

	// Verify pruned blocks have no bodies or receipts.
	for i := uint64(0); i < 4; i++ {
		hash := [32]byte{byte(i)}
		if HasBody(db, i, hash) {
			t.Fatalf("block %d body should be pruned", i)
		}
		_, err := ReadReceipts(db, i, hash)
		if err == nil {
			t.Fatalf("block %d receipts should be pruned", i)
		}
	}

	// Verify headers are still present for pruned blocks.
	for i := uint64(0); i < 4; i++ {
		hash := [32]byte{byte(i)}
		if !HasHeader(db, i, hash) {
			t.Fatalf("block %d header should be preserved after pruning", i)
		}
	}

	// Verify surviving blocks still have bodies and receipts.
	for i := uint64(4); i < 10; i++ {
		hash := [32]byte{byte(i)}
		body, err := ReadBody(db, i, hash)
		if err != nil {
			t.Fatalf("block %d body should exist: %v", i, err)
		}
		if !bytes.Equal(body, []byte("body-data")) {
			t.Fatalf("block %d body data mismatch", i)
		}
		receipt, err := ReadReceipts(db, i, hash)
		if err != nil {
			t.Fatalf("block %d receipts should exist: %v", i, err)
		}
		if !bytes.Equal(receipt, []byte("receipt-data")) {
			t.Fatalf("block %d receipt data mismatch", i)
		}
	}

	// Verify oldest is persisted.
	oldest, err := ReadHistoryOldest(db)
	if err != nil {
		t.Fatalf("ReadHistoryOldest: %v", err)
	}
	if oldest != 4 {
		t.Fatalf("want oldest 4, got %d", oldest)
	}
}

func TestPruneHistory_NoOp(t *testing.T) {
	db := NewMemoryDB()

	// Head < retention: nothing to prune.
	pruned, _, err := PruneHistory(db, 5, 100)
	if err != nil {
		t.Fatalf("PruneHistory: %v", err)
	}
	if pruned != 0 {
		t.Fatalf("want 0 pruned, got %d", pruned)
	}
}

func TestPruneHistory_Idempotent(t *testing.T) {
	db := NewMemoryDB()

	for i := uint64(0); i < 10; i++ {
		hash := [32]byte{byte(i)}
		WriteCanonicalHash(db, i, hash)
		WriteBody(db, i, hash, []byte("body"))
		WriteReceipts(db, i, hash, []byte("receipt"))
	}

	// First prune.
	pruned1, oldest1, err := PruneHistory(db, 9, 5)
	if err != nil {
		t.Fatalf("PruneHistory: %v", err)
	}
	if pruned1 != 4 {
		t.Fatalf("first prune: want 4, got %d", pruned1)
	}

	// Second prune with same parameters: should be a no-op.
	pruned2, oldest2, err := PruneHistory(db, 9, 5)
	if err != nil {
		t.Fatalf("PruneHistory: %v", err)
	}
	if pruned2 != 0 {
		t.Fatalf("second prune: want 0 (idempotent), got %d", pruned2)
	}
	if oldest2 != oldest1 {
		t.Fatalf("oldest should not change: want %d, got %d", oldest1, oldest2)
	}
}

// --- BAL retention policy (SPEC-5.5) ---

func TestBALRetentionConstants(t *testing.T) {
	if BALRetentionEpochs != 3_533 {
		t.Errorf("BALRetentionEpochs = %d, want 3533 (EIP-7928 weak-subjectivity period)", BALRetentionEpochs)
	}
	if SlotsPerEpoch != 32 {
		t.Errorf("SlotsPerEpoch = %d, want 32", SlotsPerEpoch)
	}
	want := uint64(3_533 * 32) // 113056
	if BALRetentionSlots != want {
		t.Errorf("BALRetentionSlots = %d, want %d", BALRetentionSlots, want)
	}
}

func TestIsBALRetained_ChainTooShort(t *testing.T) {
	// When the chain has not grown past BALRetentionSlots, everything is retained.
	cases := []struct {
		head     uint64
		blockNum uint64
	}{
		{0, 0},
		{1000, 0},
		{BALRetentionSlots - 1, 0},
		{BALRetentionSlots - 1, BALRetentionSlots - 1},
	}
	for _, tc := range cases {
		if !IsBALRetained(tc.head, tc.blockNum) {
			t.Errorf("head=%d block=%d: expected retained (chain too short)", tc.head, tc.blockNum)
		}
	}
}

func TestIsBALRetained_WindowBoundary(t *testing.T) {
	head := BALRetentionSlots + 1000
	// First block inside the retention window.
	firstRetained := head - BALRetentionSlots
	if !IsBALRetained(head, firstRetained) {
		t.Errorf("block at window start (%d) should be retained (head=%d)", firstRetained, head)
	}
	// Block one step before the window is NOT retained.
	if IsBALRetained(head, firstRetained-1) {
		t.Errorf("block just before window (%d) should NOT be retained (head=%d)", firstRetained-1, head)
	}
	// Latest block is always retained.
	if !IsBALRetained(head, head) {
		t.Errorf("head block should always be retained")
	}
}

func TestIsBALRetained_ExactWindowEdge(t *testing.T) {
	// At exactly head == BALRetentionSlots, block 0 is at the boundary.
	head := BALRetentionSlots
	// head - BALRetentionSlots == 0, so block 0 should be retained.
	if !IsBALRetained(head, 0) {
		t.Errorf("block 0 at exact boundary should be retained (head=%d)", head)
	}
}

func TestIsBALRetained_OldBlock(t *testing.T) {
	// A block far in the past should not be retained.
	head := uint64(1_000_000)
	oldBlock := head - BALRetentionSlots - 500
	if IsBALRetained(head, oldBlock) {
		t.Errorf("old block %d should not be retained (head=%d, window=%d)", oldBlock, head, BALRetentionSlots)
	}
}

func TestIsBALRetained_FutureBlock(t *testing.T) {
	// A block number greater than head is always >= head-BALRetentionSlots, so retained.
	head := uint64(200_000)
	futureBlock := head + 100
	if !IsBALRetained(head, futureBlock) {
		t.Errorf("future block %d should be retained (head=%d)", futureBlock, head)
	}
}

func TestPruneHistory_IncrementalPrune(t *testing.T) {
	db := NewMemoryDB()

	for i := uint64(0); i < 20; i++ {
		hash := [32]byte{byte(i)}
		WriteCanonicalHash(db, i, hash)
		WriteBody(db, i, hash, []byte("body"))
		WriteReceipts(db, i, hash, []byte("receipt"))
	}

	// First prune: head=14, retention=10 -> prune blocks 0-3.
	pruned, oldest, err := PruneHistory(db, 14, 10)
	if err != nil {
		t.Fatalf("PruneHistory: %v", err)
	}
	if pruned != 4 {
		t.Fatalf("want 4 pruned, got %d", pruned)
	}
	if oldest != 4 {
		t.Fatalf("want oldest 4, got %d", oldest)
	}

	// Second prune: head=19, retention=10 -> prune blocks 4-8.
	pruned, oldest, err = PruneHistory(db, 19, 10)
	if err != nil {
		t.Fatalf("PruneHistory: %v", err)
	}
	if pruned != 5 {
		t.Fatalf("want 5 pruned, got %d", pruned)
	}
	if oldest != 9 {
		t.Fatalf("want oldest 9, got %d", oldest)
	}

	// Verify blocks 9-19 still have bodies.
	for i := uint64(9); i < 20; i++ {
		hash := [32]byte{byte(i)}
		if !HasBody(db, i, hash) {
			t.Fatalf("block %d body should exist", i)
		}
	}
}
