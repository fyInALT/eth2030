package rawdb

import (
	"errors"
	"os"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// newTestAncientStore creates a temporary AncientStore and returns it with a
// cleanup function.
func newTestAncientStore(t *testing.T) (*AncientStore, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "ancientstore-test-*")
	if err != nil {
		t.Fatal(err)
	}
	as, err := NewAncientStore(AncientStoreConfig{DataDir: dir})
	if err != nil {
		os.RemoveAll(dir)
		t.Fatal(err)
	}
	return as, func() {
		as.Close()
		os.RemoveAll(dir)
	}
}

func makeAncientHash(n byte) types.Hash {
	var h types.Hash
	h[0] = n
	return h
}

// freezeTestBlock appends a single test block to the store.
func freezeTestBlock(t *testing.T, as *AncientStore, number uint64) {
	t.Helper()
	h := makeAncientHash(byte(number + 1))
	err := as.FreezeBlock(number, h,
		[]byte("header"),
		[]byte("body"),
		[]byte("receipts"),
	)
	if err != nil {
		t.Fatalf("FreezeBlock %d: %v", number, err)
	}
}

// --- NewAncientStore ---

func TestNewAncientStore_CreatesFiles(t *testing.T) {
	as, cleanup := newTestAncientStore(t)
	defer cleanup()

	if as.Frozen() != 0 {
		t.Fatalf("expected Frozen()=0, got %d", as.Frozen())
	}
}

func TestNewAncientStore_ReadOnlyFlag(t *testing.T) {
	_, cleanup := newTestAncientStore(t)
	cleanup() // close and remove

	// Verify read-only enforced on writes.
	dir, _ := os.MkdirTemp("", "ro-test-*")
	defer os.RemoveAll(dir)
	as2, err := NewAncientStore(AncientStoreConfig{DataDir: dir, ReadOnly: true})
	if err != nil {
		t.Skip("read-only store creation failed (platform may not support it)")
	}
	defer as2.Close()

	err = as2.FreezeBlock(0, makeAncientHash(1), []byte("h"), []byte("b"), []byte("r"))
	if !errors.Is(err, ErrAncientReadOnly) {
		t.Fatalf("expected ErrAncientReadOnly, got %v", err)
	}
}

// --- FreezeBlock & reads ---

func TestAncientStore_FreezeAndRead(t *testing.T) {
	as, cleanup := newTestAncientStore(t)
	defer cleanup()

	h := makeAncientHash(0xAB)
	header := []byte("header-data")
	body := []byte("body-data")
	receipts := []byte("receipts-data")

	if err := as.FreezeBlock(0, h, header, body, receipts); err != nil {
		t.Fatal(err)
	}

	gotHeader, err := as.ReadHeader(0)
	if err != nil {
		t.Fatalf("ReadHeader: %v", err)
	}
	if string(gotHeader) != string(header) {
		t.Fatalf("header mismatch: got %q, want %q", gotHeader, header)
	}

	gotBody, err := as.ReadBody(0)
	if err != nil {
		t.Fatalf("ReadBody: %v", err)
	}
	if string(gotBody) != string(body) {
		t.Fatalf("body mismatch: got %q, want %q", gotBody, body)
	}

	gotReceipts, err := as.ReadReceipts(0)
	if err != nil {
		t.Fatalf("ReadReceipts: %v", err)
	}
	if string(gotReceipts) != string(receipts) {
		t.Fatalf("receipts mismatch: got %q, want %q", gotReceipts, receipts)
	}

	gotHash, err := as.ReadHash(0)
	if err != nil {
		t.Fatalf("ReadHash: %v", err)
	}
	if gotHash != h {
		t.Fatalf("hash mismatch: got %x, want %x", gotHash, h)
	}
}

func TestAncientStore_HasBlock(t *testing.T) {
	as, cleanup := newTestAncientStore(t)
	defer cleanup()

	if as.HasBlock(0) {
		t.Fatal("expected HasBlock(0)=false before freeze")
	}
	freezeTestBlock(t, as, 0)
	if !as.HasBlock(0) {
		t.Fatal("expected HasBlock(0)=true after freeze")
	}
	if as.HasBlock(1) {
		t.Fatal("expected HasBlock(1)=false (not frozen)")
	}
}

func TestAncientStore_BlockRange(t *testing.T) {
	as, cleanup := newTestAncientStore(t)
	defer cleanup()

	freezeTestBlock(t, as, 0)
	freezeTestBlock(t, as, 1)
	freezeTestBlock(t, as, 2)

	tail, frozen := as.BlockRange()
	if tail != 0 {
		t.Fatalf("tail: got %d, want 0", tail)
	}
	if frozen != 3 {
		t.Fatalf("frozen: got %d, want 3", frozen)
	}
}

func TestAncientStore_Frozen(t *testing.T) {
	as, cleanup := newTestAncientStore(t)
	defer cleanup()

	if as.Frozen() != 0 {
		t.Fatalf("expected 0 before freeze")
	}
	freezeTestBlock(t, as, 0)
	if as.Frozen() != 1 {
		t.Fatalf("expected 1 after freezing block 0")
	}
}

// --- Closed store rejects operations ---

func TestAncientStore_ClosedRejectsOperations(t *testing.T) {
	as, cleanup := newTestAncientStore(t)
	cleanup() // closes and removes

	as.mu.Lock()
	as.closed = true
	as.mu.Unlock()

	if _, err := as.ReadHeader(0); !errors.Is(err, ErrAncientClosed) {
		t.Fatalf("ReadHeader: expected ErrAncientClosed, got %v", err)
	}
	if _, err := as.ReadBody(0); !errors.Is(err, ErrAncientClosed) {
		t.Fatalf("ReadBody: expected ErrAncientClosed, got %v", err)
	}
	if _, err := as.ReadReceipts(0); !errors.Is(err, ErrAncientClosed) {
		t.Fatalf("ReadReceipts: expected ErrAncientClosed, got %v", err)
	}
	if _, err := as.ReadHash(0); !errors.Is(err, ErrAncientClosed) {
		t.Fatalf("ReadHash: expected ErrAncientClosed, got %v", err)
	}
	if err := as.FreezeBlock(0, types.Hash{}, nil, nil, nil); !errors.Is(err, ErrAncientClosed) {
		t.Fatalf("FreezeBlock: expected ErrAncientClosed, got %v", err)
	}
}

// --- Close idempotent ---

func TestAncientStore_CloseIdempotent(t *testing.T) {
	as, cleanup := newTestAncientStore(t)
	defer cleanup()

	as.Close()
	if err := as.Close(); err != nil {
		t.Fatalf("second Close returned error: %v", err)
	}
}

// --- Stats ---

func TestAncientStore_Stats(t *testing.T) {
	as, cleanup := newTestAncientStore(t)
	defer cleanup()

	freezeTestBlock(t, as, 0)
	freezeTestBlock(t, as, 1)

	stats := as.Stats()
	if stats.Frozen != 2 {
		t.Fatalf("expected Frozen=2, got %d", stats.Frozen)
	}
	if len(stats.Tables) != 4 {
		t.Fatalf("expected 4 tables, got %d", len(stats.Tables))
	}
	if stats.TotalSize == 0 {
		t.Fatal("expected non-zero TotalSize")
	}
}

// --- Verify ---

func TestAncientStore_Verify(t *testing.T) {
	as, cleanup := newTestAncientStore(t)
	defer cleanup()

	freezeTestBlock(t, as, 0)
	freezeTestBlock(t, as, 1)

	count, err := as.Verify(0, 1)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 verified, got %d", count)
	}
}

func TestAncientStore_Verify_ClosedReturnsError(t *testing.T) {
	as, cleanup := newTestAncientStore(t)
	defer cleanup()
	as.Close()

	as.mu.Lock()
	as.closed = true
	as.mu.Unlock()

	_, err := as.Verify(0, 0)
	if !errors.Is(err, ErrAncientClosed) {
		t.Fatalf("expected ErrAncientClosed, got %v", err)
	}
}

// --- TruncateBlocks & PruneTail ---

func TestAncientStore_TruncateBlocks(t *testing.T) {
	as, cleanup := newTestAncientStore(t)
	defer cleanup()

	for i := uint64(0); i < 5; i++ {
		freezeTestBlock(t, as, i)
	}
	if as.Frozen() != 5 {
		t.Fatalf("expected 5, got %d", as.Frozen())
	}
	if err := as.TruncateBlocks(3); err != nil {
		t.Fatal(err)
	}
	if as.Frozen() != 3 {
		t.Fatalf("expected 3 after truncate, got %d", as.Frozen())
	}
}

func TestAncientStore_PruneTail(t *testing.T) {
	as, cleanup := newTestAncientStore(t)
	defer cleanup()

	for i := uint64(0); i < 4; i++ {
		freezeTestBlock(t, as, i)
	}
	if err := as.PruneTail(2); err != nil {
		t.Fatal(err)
	}
	if !as.HasBlock(3) {
		t.Fatal("block 3 should still be accessible")
	}
}

func TestAncientStore_PruneTail_ReadOnly(t *testing.T) {
	dir, _ := os.MkdirTemp("", "ro-prune-*")
	defer os.RemoveAll(dir)
	as, err := NewAncientStore(AncientStoreConfig{DataDir: dir, ReadOnly: true})
	if err != nil {
		t.Skip("read-only store creation failed")
	}
	defer as.Close()

	if err := as.PruneTail(0); !errors.Is(err, ErrAncientReadOnly) {
		t.Fatalf("expected ErrAncientReadOnly, got %v", err)
	}
}

// --- Compact ---

func TestAncientStore_Compact_ReadOnly(t *testing.T) {
	dir, _ := os.MkdirTemp("", "ro-compact-*")
	defer os.RemoveAll(dir)
	as, err := NewAncientStore(AncientStoreConfig{DataDir: dir, ReadOnly: true})
	if err != nil {
		t.Skip("read-only store creation failed")
	}
	defer as.Close()

	if err := as.Compact(); !errors.Is(err, ErrAncientReadOnly) {
		t.Fatalf("expected ErrAncientReadOnly, got %v", err)
	}
}

func TestAncientStore_Compact_Basic(t *testing.T) {
	as, cleanup := newTestAncientStore(t)
	defer cleanup()

	for i := uint64(0); i < 4; i++ {
		freezeTestBlock(t, as, i)
	}
	// Prune tail to make compaction meaningful.
	as.PruneTail(2)

	if err := as.Compact(); err != nil {
		t.Fatalf("Compact: %v", err)
	}
	// Data after compaction should still be readable.
	if !as.HasBlock(3) {
		t.Fatal("block 3 should be readable after compaction")
	}
}

func TestAncientStore_Compact_DoublePreventsRace(t *testing.T) {
	as, cleanup := newTestAncientStore(t)
	defer cleanup()
	freezeTestBlock(t, as, 0)

	// Manually set compact=true to simulate in-progress compaction.
	as.mu.Lock()
	as.compact = true
	as.mu.Unlock()

	if err := as.Compact(); !errors.Is(err, ErrCompactionPending) {
		t.Fatalf("expected ErrCompactionPending, got %v", err)
	}

	// Clean up.
	as.mu.Lock()
	as.compact = false
	as.mu.Unlock()
}

// --- Sync ---

func TestAncientStore_Sync(t *testing.T) {
	as, cleanup := newTestAncientStore(t)
	defer cleanup()

	freezeTestBlock(t, as, 0)
	if err := as.Sync(); err != nil {
		t.Fatalf("Sync: %v", err)
	}
}

func TestAncientStore_Sync_Closed(t *testing.T) {
	as, cleanup := newTestAncientStore(t)
	defer cleanup()
	as.Close()

	as.mu.Lock()
	as.closed = true
	as.mu.Unlock()

	if err := as.Sync(); !errors.Is(err, ErrAncientClosed) {
		t.Fatalf("expected ErrAncientClosed, got %v", err)
	}
}

// --- MigrateFromDB ---

func TestAncientStore_MigrateFromDB_Basic(t *testing.T) {
	as, cleanup := newTestAncientStore(t)
	defer cleanup()

	db := NewMemoryDB()
	h0 := makeAncientHash(0x01)
	WriteCanonicalHash(db, 0, h0)
	WriteHeader(db, 0, h0, []byte("header0"))
	WriteBody(db, 0, h0, []byte("body0"))
	WriteReceipts(db, 0, h0, []byte("receipts0"))

	count, err := as.MigrateFromDB(db, 0, 0)
	if err != nil {
		t.Fatalf("MigrateFromDB: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 migrated, got %d", count)
	}
	if as.Frozen() != 1 {
		t.Fatalf("expected Frozen=1, got %d", as.Frozen())
	}
}

func TestAncientStore_MigrateFromDB_InvalidRange(t *testing.T) {
	as, cleanup := newTestAncientStore(t)
	defer cleanup()

	db := NewMemoryDB()
	_, err := as.MigrateFromDB(db, 5, 3)
	if !errors.Is(err, ErrMigrationRange) {
		t.Fatalf("expected ErrMigrationRange, got %v", err)
	}
}

func TestAncientStore_MigrateFromDB_WrongStart(t *testing.T) {
	as, cleanup := newTestAncientStore(t)
	defer cleanup()

	db := NewMemoryDB()
	// Freezer is at 0, but we try to start from 5.
	_, err := as.MigrateFromDB(db, 1, 3)
	if !errors.Is(err, ErrMigrationRange) {
		t.Fatalf("expected ErrMigrationRange for wrong start, got %v", err)
	}
}
