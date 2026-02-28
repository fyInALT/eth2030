package state

import (
	"testing"
	"time"
)

// --- CorruptionDetector tests ---

func TestCorruptionDetectorAccountMatch(t *testing.T) {
	cd := NewCorruptionDetector()
	addr := [20]byte{0x01}
	data := []byte{0xaa, 0xbb}
	if err := cd.CheckAccountConsistency(addr, data, data); err != nil {
		t.Fatalf("expected no error for matching accounts, got %v", err)
	}
	if cd.HasErrors() {
		t.Fatal("expected no errors")
	}
}

func TestCorruptionDetectorAccountMismatch(t *testing.T) {
	cd := NewCorruptionDetector()
	addr := [20]byte{0x02}
	expected := []byte{0x01}
	actual := []byte{0x02}
	err := cd.CheckAccountConsistency(addr, expected, actual)
	if err == nil {
		t.Fatal("expected error for mismatched accounts")
	}
	ce, ok := err.(*CorruptionError)
	if !ok {
		t.Fatalf("expected *CorruptionError, got %T", err)
	}
	if ce.Type != CorruptionAccountMismatch {
		t.Fatalf("expected CorruptionAccountMismatch, got %d", ce.Type)
	}
	if cd.ErrorCount() != 1 {
		t.Fatalf("expected 1 error, got %d", cd.ErrorCount())
	}
}

func TestCorruptionDetectorAccountMissing(t *testing.T) {
	cd := NewCorruptionDetector()
	addr := [20]byte{0x03}
	expected := []byte{0xaa, 0xbb, 0xcc}
	err := cd.CheckAccountConsistency(addr, expected, nil)
	if err == nil {
		t.Fatal("expected error for missing account")
	}
	ce := err.(*CorruptionError)
	if ce.Type != CorruptionAccountMismatch {
		t.Fatalf("expected CorruptionAccountMismatch, got %d", ce.Type)
	}
}

func TestCorruptionDetectorUnexpectedAccount(t *testing.T) {
	cd := NewCorruptionDetector()
	addr := [20]byte{0x04}
	actual := []byte{0xdd}
	err := cd.CheckAccountConsistency(addr, nil, actual)
	if err == nil {
		t.Fatal("expected error for unexpected account")
	}
}

func TestCorruptionDetectorStorageMatch(t *testing.T) {
	cd := NewCorruptionDetector()
	addr := [20]byte{0x05}
	key := [32]byte{0x01}
	val := [32]byte{0xab}
	if err := cd.CheckStorageConsistency(addr, key, val, val); err != nil {
		t.Fatalf("expected no error for matching storage, got %v", err)
	}
}

func TestCorruptionDetectorStorageMismatch(t *testing.T) {
	cd := NewCorruptionDetector()
	addr := [20]byte{0x06}
	key := [32]byte{0x01}
	expected := [32]byte{0xaa}
	actual := [32]byte{0xbb}
	err := cd.CheckStorageConsistency(addr, key, expected, actual)
	if err == nil {
		t.Fatal("expected error for storage mismatch")
	}
	ce := err.(*CorruptionError)
	if ce.Type != CorruptionStorageMismatch {
		t.Fatalf("expected CorruptionStorageMismatch, got %d", ce.Type)
	}
	if ce.Key != key {
		t.Fatal("corruption error should contain the storage key")
	}
}

func TestCorruptionDetectorDanglingStorageNoAccount(t *testing.T) {
	cd := NewCorruptionDetector()
	addr := [20]byte{0x07}
	keys := [][32]byte{{0x01}, {0x02}, {0x03}}
	errs := cd.DetectDanglingStorage(addr, keys, false)
	if len(errs) != 3 {
		t.Fatalf("expected 3 dangling errors, got %d", len(errs))
	}
	if cd.ErrorCount() != 3 {
		t.Fatalf("expected 3 accumulated errors, got %d", cd.ErrorCount())
	}
}

func TestCorruptionDetectorDanglingStorageAccountExists(t *testing.T) {
	cd := NewCorruptionDetector()
	addr := [20]byte{0x08}
	keys := [][32]byte{{0x01}}
	errs := cd.DetectDanglingStorage(addr, keys, true)
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors when account exists, got %d", len(errs))
	}
}

func TestCorruptionDetectorDanglingStorageEmpty(t *testing.T) {
	cd := NewCorruptionDetector()
	addr := [20]byte{0x09}
	errs := cd.DetectDanglingStorage(addr, nil, false)
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors for empty keys, got %d", len(errs))
	}
}

func TestCorruptionDetectorReset(t *testing.T) {
	cd := NewCorruptionDetector()
	cd.CheckAccountConsistency([20]byte{}, []byte{1}, []byte{2})
	if cd.ErrorCount() == 0 {
		t.Fatal("expected errors before reset")
	}
	cd.Reset()
	if cd.ErrorCount() != 0 {
		t.Fatalf("expected 0 errors after reset, got %d", cd.ErrorCount())
	}
}

func TestCorruptionDetectorMultipleErrors(t *testing.T) {
	cd := NewCorruptionDetector()
	cd.CheckAccountConsistency([20]byte{0x01}, []byte{1}, []byte{2})
	cd.CheckStorageConsistency([20]byte{0x02}, [32]byte{1}, [32]byte{1}, [32]byte{2})
	cd.DetectDanglingStorage([20]byte{0x03}, [][32]byte{{1}}, false)

	errs := cd.Errors()
	if len(errs) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(errs))
	}
	// Verify types.
	types := map[CorruptionType]bool{}
	for _, e := range errs {
		types[e.Type] = true
	}
	if !types[CorruptionAccountMismatch] || !types[CorruptionStorageMismatch] || !types[CorruptionDanglingStorage] {
		t.Fatal("expected all three corruption types")
	}
}

// --- RollbackManager tests ---

func TestRollbackManagerAddAndLatest(t *testing.T) {
	cfg := DefaultRecoveryConfig()
	rm := NewRollbackManager(cfg)

	cp := RecoveryCheckpoint{BlockNumber: 100, StateRoot: [32]byte{0x01}, Timestamp: time.Now()}
	rm.AddCheckpoint(cp)

	latest := rm.LatestCheckpoint()
	if latest == nil {
		t.Fatal("expected a checkpoint")
	}
	if latest.BlockNumber != 100 {
		t.Fatalf("expected block 100, got %d", latest.BlockNumber)
	}
}

func TestRollbackManagerOrdering(t *testing.T) {
	cfg := DefaultRecoveryConfig()
	rm := NewRollbackManager(cfg)

	rm.AddCheckpoint(RecoveryCheckpoint{BlockNumber: 300})
	rm.AddCheckpoint(RecoveryCheckpoint{BlockNumber: 100})
	rm.AddCheckpoint(RecoveryCheckpoint{BlockNumber: 200})

	oldest := rm.OldestCheckpoint()
	if oldest.BlockNumber != 100 {
		t.Fatalf("expected oldest block 100, got %d", oldest.BlockNumber)
	}
	latest := rm.LatestCheckpoint()
	if latest.BlockNumber != 300 {
		t.Fatalf("expected latest block 300, got %d", latest.BlockNumber)
	}
}

func TestRollbackManagerFindTarget(t *testing.T) {
	cfg := DefaultRecoveryConfig()
	cfg.MaxRollbackDepth = 500
	rm := NewRollbackManager(cfg)

	rm.AddCheckpoint(RecoveryCheckpoint{BlockNumber: 100})
	rm.AddCheckpoint(RecoveryCheckpoint{BlockNumber: 200})
	rm.AddCheckpoint(RecoveryCheckpoint{BlockNumber: 300})

	target, err := rm.FindRollbackTarget(350)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.BlockNumber != 300 {
		t.Fatalf("expected rollback to block 300, got %d", target.BlockNumber)
	}
}

func TestRollbackManagerFindTargetNoCheckpoints(t *testing.T) {
	cfg := DefaultRecoveryConfig()
	rm := NewRollbackManager(cfg)

	_, err := rm.FindRollbackTarget(100)
	if err != ErrNoCheckpoints {
		t.Fatalf("expected ErrNoCheckpoints, got %v", err)
	}
}

func TestRollbackManagerFindTargetNoValidTarget(t *testing.T) {
	cfg := DefaultRecoveryConfig()
	rm := NewRollbackManager(cfg)

	rm.AddCheckpoint(RecoveryCheckpoint{BlockNumber: 200})

	_, err := rm.FindRollbackTarget(100) // corrupt at 100, no checkpoint before it
	if err != ErrNoValidTarget {
		t.Fatalf("expected ErrNoValidTarget, got %v", err)
	}
}

func TestRollbackManagerFindTargetDepthExceeded(t *testing.T) {
	cfg := DefaultRecoveryConfig()
	cfg.MaxRollbackDepth = 10
	rm := NewRollbackManager(cfg)

	rm.AddCheckpoint(RecoveryCheckpoint{BlockNumber: 100})

	_, err := rm.FindRollbackTarget(200) // depth = 100, max = 10
	if err == nil {
		t.Fatal("expected error for exceeded depth")
	}
}

func TestRollbackManagerPrune(t *testing.T) {
	cfg := DefaultRecoveryConfig()
	rm := NewRollbackManager(cfg)

	rm.AddCheckpoint(RecoveryCheckpoint{BlockNumber: 100})
	rm.AddCheckpoint(RecoveryCheckpoint{BlockNumber: 200})
	rm.AddCheckpoint(RecoveryCheckpoint{BlockNumber: 300})

	removed := rm.PruneOldCheckpoints(200)
	if removed != 2 {
		t.Fatalf("expected 2 removed, got %d", removed)
	}
	if rm.CheckpointCount() != 1 {
		t.Fatalf("expected 1 remaining, got %d", rm.CheckpointCount())
	}
	if rm.LatestCheckpoint().BlockNumber != 300 {
		t.Fatal("remaining checkpoint should be block 300")
	}
}

func TestRollbackManagerMaxCheckpoints(t *testing.T) {
	cfg := DefaultRecoveryConfig()
	cfg.MaxCheckpoints = 3
	rm := NewRollbackManager(cfg)

	for i := uint64(1); i <= 5; i++ {
		rm.AddCheckpoint(RecoveryCheckpoint{BlockNumber: i * 100})
	}

	if rm.CheckpointCount() != 3 {
		t.Fatalf("expected 3 checkpoints, got %d", rm.CheckpointCount())
	}
	// Oldest should be pruned; newest 3 remain (300, 400, 500).
	oldest := rm.OldestCheckpoint()
	if oldest.BlockNumber != 300 {
		t.Fatalf("expected oldest block 300, got %d", oldest.BlockNumber)
	}
}

func TestRollbackManagerDuplicateUpdate(t *testing.T) {
	cfg := DefaultRecoveryConfig()
	rm := NewRollbackManager(cfg)

	rm.AddCheckpoint(RecoveryCheckpoint{BlockNumber: 100, StateRoot: [32]byte{0x01}})
	rm.AddCheckpoint(RecoveryCheckpoint{BlockNumber: 100, StateRoot: [32]byte{0x02}})

	if rm.CheckpointCount() != 1 {
		t.Fatalf("expected 1 checkpoint after duplicate, got %d", rm.CheckpointCount())
	}
	cp := rm.CheckpointAtBlock(100)
	if cp == nil {
		t.Fatal("expected checkpoint at block 100")
	}
	if cp.StateRoot != [32]byte{0x02} {
		t.Fatal("expected updated state root")
	}
}

func TestRollbackManagerShouldCheck(t *testing.T) {
	cfg := DefaultRecoveryConfig()
	cfg.ConsistencyCheckInterval = 256
	rm := NewRollbackManager(cfg)

	if !rm.ShouldCheck(256) {
		t.Fatal("expected check at block 256")
	}
	if !rm.ShouldCheck(512) {
		t.Fatal("expected check at block 512")
	}
	if rm.ShouldCheck(100) {
		t.Fatal("did not expect check at block 100")
	}
}

func TestRollbackManagerShouldCheckDisabled(t *testing.T) {
	cfg := DefaultRecoveryConfig()
	cfg.ConsistencyCheckInterval = 0
	rm := NewRollbackManager(cfg)

	if rm.ShouldCheck(256) {
		t.Fatal("should not trigger check when interval is 0")
	}
}

func TestRollbackManagerCheckpointAtBlock(t *testing.T) {
	cfg := DefaultRecoveryConfig()
	rm := NewRollbackManager(cfg)

	rm.AddCheckpoint(RecoveryCheckpoint{BlockNumber: 500})

	cp := rm.CheckpointAtBlock(500)
	if cp == nil {
		t.Fatal("expected checkpoint at block 500")
	}
	cp = rm.CheckpointAtBlock(999)
	if cp != nil {
		t.Fatal("expected nil for nonexistent block")
	}
}

func TestCheckpointPruneAll(t *testing.T) {
	cfg := DefaultRecoveryConfig()
	rm := NewRollbackManager(cfg)

	rm.AddCheckpoint(RecoveryCheckpoint{BlockNumber: 10})
	rm.AddCheckpoint(RecoveryCheckpoint{BlockNumber: 20})

	removed := rm.PruneOldCheckpoints(100)
	if removed != 2 {
		t.Fatalf("expected 2 removed, got %d", removed)
	}
	if rm.LatestCheckpoint() != nil {
		t.Fatal("expected nil latest after pruning all")
	}
}

func TestCorruptionErrorMessage(t *testing.T) {
	ce := &CorruptionError{
		Type:    CorruptionAccountMismatch,
		Address: [20]byte{0xff},
		Message: "test error message",
	}
	if ce.Error() != "test error message" {
		t.Fatalf("unexpected error message: %s", ce.Error())
	}
}

func TestDefaultRecoveryConfig(t *testing.T) {
	cfg := DefaultRecoveryConfig()
	if cfg.MaxRollbackDepth == 0 {
		t.Fatal("expected non-zero MaxRollbackDepth")
	}
	if cfg.ConsistencyCheckInterval == 0 {
		t.Fatal("expected non-zero ConsistencyCheckInterval")
	}
	if !cfg.AutoRepair {
		t.Fatal("expected AutoRepair to be true by default")
	}
	if cfg.MaxCheckpoints == 0 {
		t.Fatal("expected non-zero MaxCheckpoints")
	}
}
