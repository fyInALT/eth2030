package txpool

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/eth2030/eth2030/core/types"
)

func newTestJrnl(t *testing.T) (*TxJrnl, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "tx.jrnl")
	cfg := JrnlConfig{
		Path:       path,
		FlushCount: 0, // flush on every write
	}
	j, err := NewTxJrnl(cfg)
	if err != nil {
		t.Fatalf("NewTxJrnl: %v", err)
	}
	return j, path
}

// --- DefaultJrnlConfig ---

func TestDefaultJrnlConfig(t *testing.T) {
	cfg := DefaultJrnlConfig()
	if cfg.Path == "" {
		t.Error("Path should not be empty")
	}
	if cfg.FlushCount <= 0 {
		t.Errorf("FlushCount = %d, want >0", cfg.FlushCount)
	}
	if cfg.RotateCount <= 0 {
		t.Errorf("RotateCount = %d, want >0", cfg.RotateCount)
	}
}

// --- NewTxJrnl ---

func TestNewTxJrnl_Basic(t *testing.T) {
	j, path := newTestJrnl(t)
	defer j.Close()
	if j.Path() != path {
		t.Errorf("Path = %q, want %q", j.Path(), path)
	}
	if j.IsClosed() {
		t.Error("journal should not be closed after creation")
	}
	if !j.Exists() {
		t.Error("journal file should exist after creation")
	}
}

func TestNewTxJrnl_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "tx.jrnl")
	cfg := JrnlConfig{Path: path}
	j, err := NewTxJrnl(cfg)
	if err != nil {
		t.Fatalf("NewTxJrnl with nested dir: %v", err)
	}
	defer j.Close()
	if !j.Exists() {
		t.Error("journal file should exist")
	}
}

// --- Write ---

func TestTxJrnl_Write_Basic(t *testing.T) {
	j, _ := newTestJrnl(t)
	defer j.Close()

	tx := makeTx(0, 1e9, 21000)
	if err := j.Write(tx, true); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if j.Metrics.TotalWrites.Load() != 1 {
		t.Errorf("TotalWrites = %d, want 1", j.Metrics.TotalWrites.Load())
	}
	if j.JrnlSize() == 0 {
		t.Error("JrnlSize should be >0 after write")
	}
}

func TestTxJrnl_Write_AfterClose(t *testing.T) {
	j, _ := newTestJrnl(t)
	j.Close()
	err := j.Write(makeTx(0, 1e9, 21000), false)
	if err != ErrJrnlClosed {
		t.Errorf("expected ErrJrnlClosed, got %v", err)
	}
}

func TestTxJrnl_Write_MultipleTxs(t *testing.T) {
	j, _ := newTestJrnl(t)
	defer j.Close()

	for i := range uint64(5) {
		if err := j.Write(makeTx(i, 1e9, 21000), true); err != nil {
			t.Fatalf("Write %d: %v", i, err)
		}
	}
	if j.Metrics.TotalWrites.Load() != 5 {
		t.Errorf("TotalWrites = %d, want 5", j.Metrics.TotalWrites.Load())
	}
}

// --- WriteBatch ---

func TestTxJrnl_WriteBatch(t *testing.T) {
	j, _ := newTestJrnl(t)
	defer j.Close()

	batch := []*types.Transaction{makeTx(0, 1e9, 21000), makeTx(1, 1e9, 21000)}
	if err := j.WriteBatch(batch, false); err != nil {
		t.Fatalf("WriteBatch: %v", err)
	}
	if j.Metrics.TotalWrites.Load() != 2 {
		t.Errorf("TotalWrites = %d, want 2", j.Metrics.TotalWrites.Load())
	}
}

// --- Flush ---

func TestTxJrnl_Flush_AfterWrite(t *testing.T) {
	dir := t.TempDir()
	cfg := JrnlConfig{Path: filepath.Join(dir, "tx.jrnl"), FlushCount: 100} // don't auto-flush
	j, _ := NewTxJrnl(cfg)
	defer j.Close()

	j.Write(makeTx(0, 1e9, 21000), true)
	if err := j.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if j.Metrics.FlushCount.Load() != 1 {
		t.Errorf("FlushCount = %d, want 1", j.Metrics.FlushCount.Load())
	}
}

func TestTxJrnl_Flush_NoWrites_NoOp(t *testing.T) {
	j, _ := newTestJrnl(t)
	defer j.Close()
	// No writes, flush should be a no-op.
	if err := j.Flush(); err != nil {
		t.Errorf("Flush with no writes: %v", err)
	}
	if j.Metrics.FlushCount.Load() != 0 {
		t.Errorf("FlushCount = %d, want 0", j.Metrics.FlushCount.Load())
	}
}

func TestTxJrnl_Flush_AfterClose(t *testing.T) {
	j, _ := newTestJrnl(t)
	j.Close()
	if err := j.Flush(); err != ErrJrnlClosed {
		t.Errorf("expected ErrJrnlClosed, got %v", err)
	}
}

// --- Replay / ReplayJrnl ---

func TestTxJrnl_Replay_Empty(t *testing.T) {
	j, _ := newTestJrnl(t)
	defer j.Close()
	txs, err := j.Replay()
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if len(txs) != 0 {
		t.Errorf("expected 0 txs, got %d", len(txs))
	}
}

func TestTxJrnl_Replay_AfterWrites(t *testing.T) {
	j, _ := newTestJrnl(t)

	for i := range uint64(3) {
		j.Write(makeTx(i, 1e9, 21000), true)
	}
	j.Flush()
	j.Close()

	// Re-open and replay.
	j2, path := newTestJrnl(t)
	j2.Close()
	_ = path

	// Replay the original journal directly.
	dir := t.TempDir()
	jpath := filepath.Join(dir, "test.jrnl")
	cfg := JrnlConfig{Path: jpath}
	jw, _ := NewTxJrnl(cfg)
	for i := range uint64(3) {
		jw.Write(makeTx(i, 1e9, 21000), true)
	}
	jw.Flush()
	jw.Close()

	txs, err := ReplayJrnl(jpath, nil)
	if err != nil {
		t.Fatalf("ReplayJrnl: %v", err)
	}
	if len(txs) != 3 {
		t.Errorf("expected 3 txs, got %d", len(txs))
	}
}

func TestReplayJrnl_NonExistent(t *testing.T) {
	txs, err := ReplayJrnl("/nonexistent/path/tx.jrnl", nil)
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}
	if txs != nil {
		t.Error("expected nil txs for missing file")
	}
}

func TestReplayJrnl_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.jrnl")
	os.WriteFile(path, []byte{}, 0644)
	txs, err := ReplayJrnl(path, nil)
	if err != nil {
		t.Fatalf("expected nil error for empty file, got %v", err)
	}
	if len(txs) != 0 {
		t.Errorf("expected 0 txs, got %d", len(txs))
	}
}

func TestReplayJrnl_CorruptData_Skipped(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "corrupt.jrnl")
	// Write garbage that looks like a valid length prefix but has bad payload.
	garbage := make([]byte, 20)
	garbage[0] = 0x00
	garbage[1] = 0x00
	garbage[2] = 0x00
	garbage[3] = 0x0A // length=10
	// 10 bytes of random data (not valid RLP).
	for i := 4; i < 14; i++ {
		garbage[i] = 0xFF
	}
	os.WriteFile(path, garbage, 0644)

	var m JrnlMetrics
	txs, _ := ReplayJrnl(path, &m)
	_ = txs
	if m.CorruptionCount.Load() == 0 {
		t.Error("expected corruption count > 0")
	}
}

// --- NeedsRotation ---

func TestTxJrnl_NeedsRotation_ByCount(t *testing.T) {
	dir := t.TempDir()
	cfg := JrnlConfig{Path: filepath.Join(dir, "tx.jrnl"), RotateCount: 2}
	j, _ := NewTxJrnl(cfg)
	defer j.Close()

	j.Write(makeTx(0, 1e9, 21000), true)
	if j.NeedsRotation() {
		t.Error("should not need rotation after 1 write (threshold=2)")
	}
	j.Write(makeTx(1, 1e9, 21000), true)
	if !j.NeedsRotation() {
		t.Error("should need rotation after 2 writes (threshold=2)")
	}
}

func TestTxJrnl_NeedsRotation_Closed(t *testing.T) {
	j, _ := newTestJrnl(t)
	j.Close()
	if j.NeedsRotation() {
		t.Error("closed journal should not need rotation")
	}
}

// --- Rotate ---

func TestTxJrnl_Rotate(t *testing.T) {
	dir := t.TempDir()
	cfg := JrnlConfig{Path: filepath.Join(dir, "tx.jrnl"), RotateCount: 10}
	j, _ := NewTxJrnl(cfg)
	defer j.Close()

	for i := range uint64(5) {
		j.Write(makeTx(i, 1e9, 21000), true)
	}

	pending := map[types.Address][]*types.Transaction{
		testSender: {makeTx(5, 1e9, 21000)},
	}
	if err := j.Rotate(pending); err != nil {
		t.Fatalf("Rotate: %v", err)
	}
	if j.Metrics.RotationCount.Load() != 1 {
		t.Errorf("RotationCount = %d, want 1", j.Metrics.RotationCount.Load())
	}
	if j.EntriesSinceRotate() != 1 {
		t.Errorf("EntriesSinceRotate = %d, want 1", j.EntriesSinceRotate())
	}
}

func TestTxJrnl_Rotate_AfterClose(t *testing.T) {
	j, _ := newTestJrnl(t)
	j.Close()
	err := j.Rotate(nil)
	if err != ErrJrnlClosed {
		t.Errorf("expected ErrJrnlClosed, got %v", err)
	}
}

// --- Close ---

func TestTxJrnl_Close_Idempotent(t *testing.T) {
	j, _ := newTestJrnl(t)
	j.Close()
	if err := j.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}
}

// --- IsClosed ---

func TestTxJrnl_IsClosed(t *testing.T) {
	j, _ := newTestJrnl(t)
	if j.IsClosed() {
		t.Error("should not be closed immediately after creation")
	}
	j.Close()
	if !j.IsClosed() {
		t.Error("should be closed after Close()")
	}
}

// --- EntriesSinceRotate ---

func TestTxJrnl_EntriesSinceRotate(t *testing.T) {
	j, _ := newTestJrnl(t)
	defer j.Close()
	if j.EntriesSinceRotate() != 0 {
		t.Errorf("EntriesSinceRotate = %d, want 0", j.EntriesSinceRotate())
	}
	j.Write(makeTx(0, 1e9, 21000), true)
	if j.EntriesSinceRotate() != 1 {
		t.Errorf("EntriesSinceRotate = %d, want 1", j.EntriesSinceRotate())
	}
}

// --- ValidateJournal ---

func TestValidateJournal_ValidEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "valid.jrnl")
	cfg := JrnlConfig{Path: path}
	j, _ := NewTxJrnl(cfg)
	for i := range uint64(3) {
		j.Write(makeTx(i, 1e9, 21000), true)
	}
	j.Close()

	total, corrupt, err := ValidateJournal(path)
	if err != nil {
		t.Fatalf("ValidateJournal: %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if corrupt != 0 {
		t.Errorf("corrupt = %d, want 0", corrupt)
	}
}

func TestValidateJournal_NonExistent(t *testing.T) {
	_, _, err := ValidateJournal("/nonexistent/journal.jrnl")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

// --- JrnlSize ---

func TestTxJrnl_JrnlSize_AfterWrite(t *testing.T) {
	j, _ := newTestJrnl(t)
	defer j.Close()
	if j.JrnlSize() != 0 {
		t.Errorf("JrnlSize = %d, want 0 before write", j.JrnlSize())
	}
	j.Write(makeTx(0, 1e9, 21000), true)
	if j.JrnlSize() == 0 {
		t.Error("JrnlSize should be >0 after write")
	}
}

// --- itoa ---

func TestItoa(t *testing.T) {
	cases := []struct {
		in  int64
		out string
	}{
		{0, "0"},
		{1, "1"},
		{-1, "-1"},
		{12345, "12345"},
		{-9876, "-9876"},
	}
	for _, tc := range cases {
		if got := itoa(tc.in); got != tc.out {
			t.Errorf("itoa(%d) = %q, want %q", tc.in, got, tc.out)
		}
	}
}

// --- JrnlError ---

func TestJrnlError(t *testing.T) {
	inner := os.ErrInvalid
	e := &JrnlError{Offset: 42, Err: inner}
	if e.Error() == "" {
		t.Error("JrnlError.Error() should not be empty")
	}
	if e.Unwrap() != inner {
		t.Error("JrnlError.Unwrap() should return inner error")
	}
}

// --- Flush interval (background goroutine) ---

func TestTxJrnl_FlushInterval(t *testing.T) {
	dir := t.TempDir()
	cfg := JrnlConfig{
		Path:          filepath.Join(dir, "tx.jrnl"),
		FlushInterval: 10 * time.Millisecond,
		FlushCount:    1000, // don't auto-flush by count
	}
	j, err := NewTxJrnl(cfg)
	if err != nil {
		t.Fatalf("NewTxJrnl: %v", err)
	}
	j.Write(makeTx(0, 1e9, 21000), true)
	time.Sleep(50 * time.Millisecond) // let background goroutine flush
	j.Close()
	// Just verify no panic and clean close.
}
