package consensus

import (
	"sync"
	"testing"
	"time"

	"github.com/eth2030/eth2030/core/types"
)

// makeTestSSFRound creates an SSFRound for testing.
func makeTestSSFRound(slot uint64, blockRoot types.Hash, voteCount int, stakePerVote uint64, finalized bool) *SSFRound {
	round := &SSFRound{
		Slot: slot, Phase: PhaseFinalize, BlockRoot: blockRoot,
		Votes:       make(map[types.Hash]*SSFRoundVote, voteCount),
		StakeByRoot: make(map[types.Hash]uint64),
		Finalized:   finalized, StartedAt: time.Now().Add(-100 * time.Millisecond), FinalizedAt: time.Now(),
	}
	var totalStake uint64
	for i := 0; i < voteCount; i++ {
		var pkHash types.Hash
		pkHash[0] = byte(i)
		pkHash[1] = byte(i >> 8)
		vote := &SSFRoundVote{ValidatorPubkeyHash: pkHash, Slot: slot, BlockRoot: blockRoot, Stake: stakePerVote}
		round.Votes[pkHash] = vote
		totalStake += stakePerVote
	}
	round.StakeByRoot[blockRoot] = totalStake
	round.TotalVoteStake = totalStake
	return round
}

func testBlockRoot(n byte) types.Hash { var h types.Hash; h[0] = n; h[31] = n; return h }
func testStateRoot(n byte) types.Hash { var h types.Hash; h[0] = n; h[15] = n; return h }

func newTestPipeline(t *testing.T, totalStake uint64) *BFTFinalityPipeline {
	t.Helper()
	cfg := DefaultBFTPipelineCfg()
	cfg.TotalStake = totalStake
	p, err := NewBFTFinalityPipeline(cfg)
	if err != nil {
		t.Fatalf("NewBFTFinalityPipeline: %v", err)
	}
	return p
}

func TestBFTPipelineProcessFinalizedRound(t *testing.T) {
	p := newTestPipeline(t, 100*GweiPerETH)
	blockRoot, stateRoot := testBlockRoot(1), testStateRoot(1)
	round := makeTestSSFRound(10, blockRoot, 3, 34*GweiPerETH, true)
	cp, err := p.ProcessRound(round, 5, stateRoot)
	if err != nil {
		t.Fatalf("ProcessRound failed: %v", err)
	}
	if cp == nil {
		t.Fatal("expected non-nil checkpoint")
	}
	if cp.Epoch != 5 || cp.Slot != 10 {
		t.Errorf("epoch=%d slot=%d, want 5/10", cp.Epoch, cp.Slot)
	}
	if cp.BlockRoot != blockRoot || cp.StateRoot != stateRoot {
		t.Error("root mismatch")
	}
	if cp.Proof == nil {
		t.Error("expected non-nil proof")
	}
	if cp.ParticipantCount != 3 {
		t.Errorf("participants = %d, want 3", cp.ParticipantCount)
	}
}

func TestBFTPipelineProcessNonFinalizedRound(t *testing.T) {
	p := newTestPipeline(t, 100*GweiPerETH)
	round := makeTestSSFRound(20, testBlockRoot(2), 3, 34*GweiPerETH, false)
	_, err := p.ProcessRound(round, 10, testStateRoot(2))
	if err != ErrBFTPipelineRoundNotFinalized {
		t.Errorf("expected ErrBFTPipelineRoundNotFinalized, got: %v", err)
	}
}

func TestBFTPipelineLatestCheckpoint(t *testing.T) {
	p := newTestPipeline(t, 100*GweiPerETH)
	if latest := p.LatestCheckpoint(); latest != nil {
		t.Fatal("expected nil for empty pipeline")
	}
	round := makeTestSSFRound(30, testBlockRoot(3), 3, 34*GweiPerETH, true)
	if _, err := p.ProcessRound(round, 15, testStateRoot(3)); err != nil {
		t.Fatalf("ProcessRound failed: %v", err)
	}
	if latest := p.LatestCheckpoint(); latest == nil || latest.Epoch != 15 {
		t.Errorf("latest = %v, want epoch 15", latest)
	}
}

func TestBFTPipelineCheckpointByEpoch(t *testing.T) {
	p := newTestPipeline(t, 100*GweiPerETH)
	round := makeTestSSFRound(40, testBlockRoot(4), 3, 34*GweiPerETH, true)
	if _, err := p.ProcessRound(round, 20, testStateRoot(4)); err != nil {
		t.Fatalf("ProcessRound failed: %v", err)
	}
	if cp, ok := p.CheckpointByEpoch(20); !ok || cp.Epoch != 20 {
		t.Error("expected checkpoint at epoch 20")
	}
	if _, ok := p.CheckpointByEpoch(999); ok {
		t.Error("expected no checkpoint at epoch 999")
	}
}

func TestBFTPipelinePruneOlderThan(t *testing.T) {
	p := newTestPipeline(t, 100*GweiPerETH)
	for i := uint64(1); i <= 5; i++ {
		round := makeTestSSFRound(i*10, testBlockRoot(byte(i)), 3, 34*GweiPerETH, true)
		if _, err := p.ProcessRound(round, i, testStateRoot(byte(i))); err != nil {
			t.Fatalf("epoch %d: %v", i, err)
		}
	}
	if pruned := p.PruneOlderThan(3); pruned != 2 {
		t.Errorf("pruned = %d, want 2", pruned)
	}
	if p.CheckpointCount() != 3 {
		t.Errorf("count = %d, want 3", p.CheckpointCount())
	}
	if _, ok := p.CheckpointByEpoch(1); ok {
		t.Error("epoch 1 should be pruned")
	}
	if _, ok := p.CheckpointByEpoch(3); !ok {
		t.Error("epoch 3 should exist")
	}
}

func TestBFTPipelineSequentialRounds(t *testing.T) {
	p := newTestPipeline(t, 100*GweiPerETH)
	for i := uint64(1); i <= 10; i++ {
		round := makeTestSSFRound(i*10, testBlockRoot(byte(i)), 3, 34*GweiPerETH, true)
		cp, err := p.ProcessRound(round, i, testStateRoot(byte(i)))
		if err != nil {
			t.Fatalf("epoch %d: %v", i, err)
		}
		if cp.Epoch != i {
			t.Errorf("epoch = %d, want %d", cp.Epoch, i)
		}
	}
	if p.ProcessedCount() != 10 {
		t.Errorf("processed = %d, want 10", p.ProcessedCount())
	}
	if p.LatestEpoch() != 10 {
		t.Errorf("latest epoch = %d, want 10", p.LatestEpoch())
	}
}

func TestBFTPipelineConcurrentProcessing(t *testing.T) {
	p := newTestPipeline(t, 100*GweiPerETH)
	var wg sync.WaitGroup
	for i := uint64(1); i <= 20; i++ {
		wg.Add(1)
		go func(epoch uint64) {
			defer wg.Done()
			round := makeTestSSFRound(epoch*10, testBlockRoot(byte(epoch)), 3, 34*GweiPerETH, true)
			_, _ = p.ProcessRound(round, epoch, testStateRoot(byte(epoch)))
		}(i)
	}
	wg.Wait()
	if p.ProcessedCount() != 20 {
		t.Errorf("concurrent processed = %d, want 20", p.ProcessedCount())
	}
}

func TestBFTPipelineNilRound(t *testing.T) {
	p := newTestPipeline(t, 100*GweiPerETH)
	if _, err := p.ProcessRound(nil, 1, testStateRoot(1)); err != ErrBFTPipelineNilRound {
		t.Errorf("expected ErrBFTPipelineNilRound, got: %v", err)
	}
}

func TestBFTPipelineEmptyVotes(t *testing.T) {
	cfg := DefaultBFTPipelineCfg()
	cfg.TotalStake = 0
	p, _ := NewBFTFinalityPipeline(cfg)
	round := &SSFRound{
		Slot: 10, Phase: PhaseFinalize, BlockRoot: testBlockRoot(1),
		Votes: make(map[types.Hash]*SSFRoundVote), StakeByRoot: make(map[types.Hash]uint64),
		Finalized: true, StartedAt: time.Now(), FinalizedAt: time.Now(),
	}
	if _, err := p.ProcessRound(round, 1, testStateRoot(1)); err == nil {
		t.Fatal("expected error for empty votes")
	}
}

func TestBFTPipelineConfigValidation(t *testing.T) {
	if _, err := NewBFTFinalityPipeline(nil); err != ErrBFTPipelineNilConfig {
		t.Errorf("nil config: got %v", err)
	}
	cfg := DefaultBFTPipelineCfg()
	cfg.ConfirmationDepth = 0
	if _, err := NewBFTFinalityPipeline(cfg); err != ErrBFTPipelineZeroConfirmation {
		t.Errorf("zero depth: got %v", err)
	}
	cfg = DefaultBFTPipelineCfg()
	cfg.CheckpointInterval = 0
	if _, err := NewBFTFinalityPipeline(cfg); err != ErrBFTPipelineZeroInterval {
		t.Errorf("zero interval: got %v", err)
	}
	cfg = DefaultBFTPipelineCfg()
	cfg.MaxPending = 0
	if _, err := NewBFTFinalityPipeline(cfg); err != ErrBFTPipelineZeroMaxPending {
		t.Errorf("zero max pending: got %v", err)
	}
}

func TestBFTPipelineMultipleCheckpointsSameEpoch(t *testing.T) {
	p := newTestPipeline(t, 100*GweiPerETH)
	round1 := makeTestSSFRound(10, testBlockRoot(1), 3, 34*GweiPerETH, true)
	if _, err := p.ProcessRound(round1, 5, testStateRoot(1)); err != nil {
		t.Fatal(err)
	}
	blockRoot2 := testBlockRoot(2)
	round2 := makeTestSSFRound(11, blockRoot2, 3, 34*GweiPerETH, true)
	if _, err := p.ProcessRound(round2, 5, testStateRoot(2)); err != nil {
		t.Fatal(err)
	}
	cp, ok := p.CheckpointByEpoch(5)
	if !ok || cp.BlockRoot != blockRoot2 {
		t.Error("expected replaced checkpoint with second root")
	}
	if p.CheckpointCount() != 1 {
		t.Errorf("count = %d, want 1", p.CheckpointCount())
	}
}

func TestBFTPipelineWithBLSAdapter(t *testing.T) {
	p := newTestPipeline(t, 100*GweiPerETH)
	if p.Adapter() == nil {
		t.Fatal("expected non-nil adapter")
	}
	blockRoot := testBlockRoot(7)
	round := makeTestSSFRound(70, blockRoot, 3, 34*GweiPerETH, true)
	cp, err := p.ProcessRound(round, 35, testStateRoot(7))
	if err != nil {
		t.Fatal(err)
	}
	if cp.Proof == nil || cp.Proof.Epoch != 35 || cp.Proof.BlockRoot != blockRoot {
		t.Error("proof mismatch")
	}
}

func TestBFTPipelineStoreCapacityLimits(t *testing.T) {
	cfg := DefaultBFTPipelineCfg()
	cfg.TotalStake = 100 * GweiPerETH
	cfg.MaxPending = 5
	p, _ := NewBFTFinalityPipeline(cfg)
	for i := uint64(1); i <= 10; i++ {
		round := makeTestSSFRound(i*10, testBlockRoot(byte(i)), 3, 34*GweiPerETH, true)
		if _, err := p.ProcessRound(round, i, testStateRoot(byte(i))); err != nil {
			t.Fatalf("epoch %d: %v", i, err)
		}
	}
	if p.CheckpointCount() != 5 {
		t.Errorf("count = %d, want 5", p.CheckpointCount())
	}
	for i := uint64(1); i <= 5; i++ {
		if _, ok := p.CheckpointByEpoch(i); ok {
			t.Errorf("epoch %d should be auto-pruned", i)
		}
	}
	for i := uint64(6); i <= 10; i++ {
		if _, ok := p.CheckpointByEpoch(i); !ok {
			t.Errorf("epoch %d should exist", i)
		}
	}
}

func TestBFTPipelineInsufficientStake(t *testing.T) {
	p := newTestPipeline(t, 100*GweiPerETH)
	round := makeTestSSFRound(90, testBlockRoot(9), 1, 10*GweiPerETH, true)
	if _, err := p.ProcessRound(round, 45, testStateRoot(9)); err == nil {
		t.Fatal("expected error for insufficient stake")
	}
}

func TestBFTPipelineStoredEpochs(t *testing.T) {
	p := newTestPipeline(t, 100*GweiPerETH)
	for i := uint64(5); i >= 1; i-- {
		round := makeTestSSFRound(i*10, testBlockRoot(byte(i)), 3, 34*GweiPerETH, true)
		if _, err := p.ProcessRound(round, i, testStateRoot(byte(i))); err != nil {
			t.Fatalf("epoch %d: %v", i, err)
		}
	}
	epochs := p.StoredEpochs()
	if len(epochs) != 5 {
		t.Fatalf("len = %d, want 5", len(epochs))
	}
	for i := 1; i < len(epochs); i++ {
		if epochs[i] <= epochs[i-1] {
			t.Errorf("epochs not sorted: %v", epochs)
			break
		}
	}
}

func TestBFTCheckpointStoreDirectPruning(t *testing.T) {
	store := NewBFTCheckpointStore(100)
	for i := uint64(1); i <= 10; i++ {
		store.Put(&BFTPipelineCheckpoint{Epoch: i, Slot: i * 10, BlockRoot: testBlockRoot(byte(i))})
	}
	if store.Count() != 10 {
		t.Fatalf("count = %d, want 10", store.Count())
	}
	if pruned := store.PruneOlderThan(6); pruned != 5 {
		t.Errorf("pruned = %d, want 5", pruned)
	}
	if store.Count() != 5 {
		t.Errorf("remaining = %d, want 5", store.Count())
	}
}

func TestBFTValidateBFTPipelineCfg(t *testing.T) {
	if err := ValidateBFTPipelineCfg(nil); err != ErrBFTPipelineNilConfig {
		t.Errorf("nil: %v", err)
	}
	cases := []struct {
		name string
		cfg  BFTPipelineCfg
		want error
	}{
		{"zero depth", BFTPipelineCfg{ConfirmationDepth: 0, CheckpointInterval: 1, MaxPending: 1}, ErrBFTPipelineZeroConfirmation},
		{"zero interval", BFTPipelineCfg{ConfirmationDepth: 1, CheckpointInterval: 0, MaxPending: 1}, ErrBFTPipelineZeroInterval},
		{"zero max", BFTPipelineCfg{ConfirmationDepth: 1, CheckpointInterval: 1, MaxPending: 0}, ErrBFTPipelineZeroMaxPending},
	}
	for _, tc := range cases {
		if err := ValidateBFTPipelineCfg(&tc.cfg); err != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, err, tc.want)
		}
	}
	if err := ValidateBFTPipelineCfg(DefaultBFTPipelineCfg()); err != nil {
		t.Errorf("valid config: %v", err)
	}
}
