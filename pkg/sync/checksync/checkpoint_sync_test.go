package checksync

import (
	"sync"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

func makeTestCheckpoint(epoch, blockNum uint64) Checkpoint {
	return Checkpoint{
		Epoch:       epoch,
		BlockNumber: blockNum,
		BlockHash:   types.HexToHash("0xaa11bb22cc33dd44ee55ff6677889900aabbccdd11223344556677889900aabb"),
		StateRoot:   types.HexToHash("0x1122334455667788990011223344556677889900aabbccddeeff001122334455"),
	}
}

func TestChecksync_NewCheckpointSyncer(t *testing.T) {
	cs := NewCheckpointSyncer(DefaultCheckpointConfig())
	if cs == nil {
		t.Fatal("expected non-nil syncer")
	}
	if cs.config.MaxHeaderBatch != 64 {
		t.Errorf("expected MaxHeaderBatch=64, got %d", cs.config.MaxHeaderBatch)
	}
	if cs.GetCheckpoint() != nil {
		t.Error("expected nil checkpoint initially")
	}
}

func TestChecksync_SetAndGetCheckpoint(t *testing.T) {
	cs := NewCheckpointSyncer(DefaultCheckpointConfig())
	cp := makeTestCheckpoint(10, 320)
	if err := cs.SetCheckpoint(cp); err != nil {
		t.Fatalf("SetCheckpoint: %v", err)
	}
	got := cs.GetCheckpoint()
	if got == nil || got.BlockNumber != 320 {
		t.Error("checkpoint not stored correctly")
	}
}

func TestChecksync_SetCheckpoint_Errors(t *testing.T) {
	cs := NewCheckpointSyncer(DefaultCheckpointConfig())

	if err := cs.SetCheckpoint(Checkpoint{
		Epoch: 1, BlockNumber: 32, StateRoot: types.HexToHash("0x1234"),
	}); err != ErrCheckpointZeroHash {
		t.Errorf("expected ErrCheckpointZeroHash, got %v", err)
	}

	if err := cs.SetCheckpoint(Checkpoint{
		Epoch: 1, BlockNumber: 32, BlockHash: types.HexToHash("0xabcd"),
	}); err != ErrCheckpointZeroState {
		t.Errorf("expected ErrCheckpointZeroState, got %v", err)
	}

	if err := cs.SetCheckpoint(Checkpoint{
		BlockHash: types.HexToHash("0xabcd"), StateRoot: types.HexToHash("0x1234"),
	}); err != ErrCheckpointZeroBlock {
		t.Errorf("expected ErrCheckpointZeroBlock, got %v", err)
	}
}

func TestChecksync_Hash(t *testing.T) {
	cp1 := makeTestCheckpoint(10, 320)
	cp2 := makeTestCheckpoint(10, 320)
	cp3 := makeTestCheckpoint(11, 352)

	if cp1.Hash() != cp2.Hash() {
		t.Error("same checkpoints should produce same hash")
	}
	if cp1.Hash() == cp3.Hash() {
		t.Error("different checkpoints should produce different hashes")
	}
	if cp1.Hash().IsZero() {
		t.Error("checkpoint hash should not be zero")
	}
}

func TestChecksync_VerifyCheckpoint(t *testing.T) {
	cs := NewCheckpointSyncer(DefaultCheckpointConfig())
	cp := makeTestCheckpoint(10, 320)
	valid, err := cs.VerifyCheckpoint(cp)
	if err != nil {
		t.Fatalf("VerifyCheckpoint: %v", err)
	}
	if !valid {
		t.Error("expected valid checkpoint")
	}
	if cs.VerifiedCheckpoints() != 1 {
		t.Errorf("expected 1 verified, got %d", cs.VerifiedCheckpoints())
	}
	if !cs.IsVerified(cp) {
		t.Error("checkpoint should be marked verified")
	}
}

func TestChecksync_SyncFromCheckpoint(t *testing.T) {
	cp := makeTestCheckpoint(10, 320)
	cs := NewCheckpointSyncer(CheckpointConfig{
		TrustedCheckpoint: &cp,
		MaxHeaderBatch:    64,
	})
	cs.SetTarget(1000)
	if err := cs.SyncFromCheckpoint(); err != nil {
		t.Fatalf("SyncFromCheckpoint: %v", err)
	}
	p := cs.Progress()
	if p.CurrentBlock != 320 {
		t.Errorf("expected CurrentBlock=320, got %d", p.CurrentBlock)
	}
	if p.TargetBlock != 1000 {
		t.Errorf("expected TargetBlock=1000, got %d", p.TargetBlock)
	}
	if cs.IsComplete() {
		t.Error("sync should not be complete yet")
	}
}

func TestChecksync_UpdateProgressAndReset(t *testing.T) {
	cp := makeTestCheckpoint(10, 320)
	cs := NewCheckpointSyncer(CheckpointConfig{
		TrustedCheckpoint: &cp, MaxHeaderBatch: 64,
	})
	cs.SetTarget(1320)
	_ = cs.SyncFromCheckpoint()

	cs.UpdateProgress(820)
	p := cs.Progress()
	if p.Percentage < 49.9 || p.Percentage > 50.1 {
		t.Errorf("expected ~50%%, got %.2f", p.Percentage)
	}

	cs.Reset()
	if cs.IsComplete() {
		t.Error("expected not complete after reset")
	}
}

func TestChecksync_ConcurrentAccess(t *testing.T) {
	cp := makeTestCheckpoint(10, 320)
	cs := NewCheckpointSyncer(CheckpointConfig{
		TrustedCheckpoint: &cp, MaxHeaderBatch: 64,
	})
	cs.SetTarget(10000)
	_ = cs.SyncFromCheckpoint()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		blockNum := uint64(320 + i*10)
		go func(bn uint64) {
			defer wg.Done()
			cs.UpdateProgress(bn)
		}(blockNum)
		go func() {
			defer wg.Done()
			_ = cs.Progress()
		}()
	}
	wg.Wait()
}
