package minimmit

import (
	"errors"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// makeMinimmitRoot creates a test block root from a single byte.
func makeMinimmitRoot(b byte) types.Hash {
	var h types.Hash
	h[0] = b
	return h
}

func TestDefaultMinimmitConfig(t *testing.T) {
	cfg := DefaultMinimmitConfig()

	expectedStake := uint64(32_000_000) * gweiPerETH
	if cfg.TotalStake != expectedStake {
		t.Errorf("TotalStake = %d, want %d", cfg.TotalStake, expectedStake)
	}
	if cfg.FinalityThresholdNum != 2 {
		t.Errorf("FinalityThresholdNum = %d, want 2", cfg.FinalityThresholdNum)
	}
	if cfg.FinalityThresholdDen != 3 {
		t.Errorf("FinalityThresholdDen = %d, want 3", cfg.FinalityThresholdDen)
	}
	if cfg.VoterLimit != 8192 {
		t.Errorf("VoterLimit = %d, want 8192", cfg.VoterLimit)
	}
	if cfg.MissedSlotPenalty != 1*gweiPerETH {
		t.Errorf("MissedSlotPenalty = %d, want %d", cfg.MissedSlotPenalty, 1*gweiPerETH)
	}
}

func TestNewMinimmitEngine_Valid(t *testing.T) {
	cfg := &MinimmitConfig{
		TotalStake:           300,
		FinalityThresholdNum: 2,
		FinalityThresholdDen: 3,
		VoterLimit:           100,
		MissedSlotPenalty:    10,
	}
	engine, err := NewMinimmitEngine(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if engine == nil {
		t.Fatal("engine should not be nil")
	}
	if engine.State() != MinimmitIdle {
		t.Errorf("initial state = %s, want Idle", engine.State())
	}
}

func TestNewMinimmitEngine_NilConfig(t *testing.T) {
	engine, err := NewMinimmitEngine(nil)
	if !errors.Is(err, ErrMinimmitNilConfig) {
		t.Errorf("expected ErrMinimmitNilConfig, got %v", err)
	}
	if engine != nil {
		t.Error("engine should be nil on error")
	}
}

func TestNewMinimmitEngine_ZeroStake(t *testing.T) {
	cfg := &MinimmitConfig{TotalStake: 0}
	engine, err := NewMinimmitEngine(cfg)
	if !errors.Is(err, ErrMinimmitZeroStake) {
		t.Errorf("expected ErrMinimmitZeroStake, got %v", err)
	}
	if engine != nil {
		t.Error("engine should be nil on error")
	}
}

func TestMinimmitProposeBlock(t *testing.T) {
	cfg := &MinimmitConfig{
		TotalStake:           300,
		FinalityThresholdNum: 2,
		FinalityThresholdDen: 3,
	}
	engine, _ := NewMinimmitEngine(cfg)

	root := makeMinimmitRoot(0xAA)
	err := engine.ProposeBlock(1, root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if engine.State() != MinimmitVoting {
		t.Errorf("state = %s, want Voting", engine.State())
	}
}

func TestMinimmitProposeBlock_WrongState(t *testing.T) {
	cfg := &MinimmitConfig{
		TotalStake:           300,
		FinalityThresholdNum: 2,
		FinalityThresholdDen: 3,
	}
	engine, _ := NewMinimmitEngine(cfg)

	root := makeMinimmitRoot(0xAA)
	_ = engine.ProposeBlock(1, root)

	// Try to propose again while in Voting state.
	err := engine.ProposeBlock(2, makeMinimmitRoot(0xBB))
	if !errors.Is(err, ErrMinimmitInvalidState) {
		t.Errorf("expected ErrMinimmitInvalidState, got %v", err)
	}
}

func TestMinimmitCastVote_Basic(t *testing.T) {
	cfg := &MinimmitConfig{
		TotalStake:           300,
		FinalityThresholdNum: 2,
		FinalityThresholdDen: 3,
	}
	engine, _ := NewMinimmitEngine(cfg)
	root := makeMinimmitRoot(0xAA)
	_ = engine.ProposeBlock(1, root)

	vote := MinimmitVote{
		ValidatorIndex: 0,
		Slot:           1,
		BlockRoot:      root,
		Stake:          100,
	}
	err := engine.CastVote(vote)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if engine.VoteCount() != 1 {
		t.Errorf("vote count = %d, want 1", engine.VoteCount())
	}
	if engine.StakeForRoot(root) != 100 {
		t.Errorf("stake for root = %d, want 100", engine.StakeForRoot(root))
	}
}

func TestMinimmitCastVote_WrongSlot(t *testing.T) {
	cfg := &MinimmitConfig{
		TotalStake:           300,
		FinalityThresholdNum: 2,
		FinalityThresholdDen: 3,
	}
	engine, _ := NewMinimmitEngine(cfg)
	root := makeMinimmitRoot(0xAA)
	_ = engine.ProposeBlock(1, root)

	vote := MinimmitVote{
		ValidatorIndex: 0,
		Slot:           2, // wrong slot
		BlockRoot:      root,
		Stake:          100,
	}
	err := engine.CastVote(vote)
	if !errors.Is(err, ErrMinimmitWrongSlot) {
		t.Errorf("expected ErrMinimmitWrongSlot, got %v", err)
	}
}

func TestMinimmitCastVote_Duplicate(t *testing.T) {
	cfg := &MinimmitConfig{
		TotalStake:           300,
		FinalityThresholdNum: 2,
		FinalityThresholdDen: 3,
	}
	engine, _ := NewMinimmitEngine(cfg)
	root := makeMinimmitRoot(0xAA)
	_ = engine.ProposeBlock(1, root)

	vote := MinimmitVote{
		ValidatorIndex: 0,
		Slot:           1,
		BlockRoot:      root,
		Stake:          100,
	}
	_ = engine.CastVote(vote)

	// Same validator, same root = duplicate.
	err := engine.CastVote(vote)
	if !errors.Is(err, ErrMinimmitDuplicateVote) {
		t.Errorf("expected ErrMinimmitDuplicateVote, got %v", err)
	}
}

func TestMinimmitCastVote_Equivocation(t *testing.T) {
	cfg := &MinimmitConfig{
		TotalStake:           300,
		FinalityThresholdNum: 2,
		FinalityThresholdDen: 3,
	}
	engine, _ := NewMinimmitEngine(cfg)
	root := makeMinimmitRoot(0xAA)
	_ = engine.ProposeBlock(1, root)

	vote1 := MinimmitVote{
		ValidatorIndex: 0,
		Slot:           1,
		BlockRoot:      root,
		Stake:          100,
	}
	_ = engine.CastVote(vote1)

	// Same validator, different root = equivocation.
	vote2 := MinimmitVote{
		ValidatorIndex: 0,
		Slot:           1,
		BlockRoot:      makeMinimmitRoot(0xBB),
		Stake:          100,
	}
	err := engine.CastVote(vote2)
	if !errors.Is(err, ErrMinimmitEquivocation) {
		t.Errorf("expected ErrMinimmitEquivocation, got %v", err)
	}
}

func TestMinimmitCastVote_Finality(t *testing.T) {
	cfg := &MinimmitConfig{
		TotalStake:           300,
		FinalityThresholdNum: 2,
		FinalityThresholdDen: 3,
	}
	engine, _ := NewMinimmitEngine(cfg)
	root := makeMinimmitRoot(0xAA)
	_ = engine.ProposeBlock(1, root)

	// Cast votes totaling >= 2/3 of 300 = 200.
	for i := 0; i < 3; i++ {
		vote := MinimmitVote{
			ValidatorIndex: ValidatorIndex(i),
			Slot:           1,
			BlockRoot:      root,
			Stake:          100,
		}
		_ = engine.CastVote(vote)
	}

	if !engine.CheckFinality() {
		t.Error("expected finality after 2/3+ votes")
	}
	if engine.State() != MinimmitFinalized {
		t.Errorf("state = %s, want Finalized", engine.State())
	}
}

func TestMinimmitCheckFinality(t *testing.T) {
	cfg := &MinimmitConfig{
		TotalStake:           300,
		FinalityThresholdNum: 2,
		FinalityThresholdDen: 3,
	}
	engine, _ := NewMinimmitEngine(cfg)

	// Before any proposal, not finalized.
	if engine.CheckFinality() {
		t.Error("should not be finalized in idle state")
	}

	root := makeMinimmitRoot(0xAA)
	_ = engine.ProposeBlock(1, root)

	// After proposal, not yet finalized.
	if engine.CheckFinality() {
		t.Error("should not be finalized before votes")
	}

	// Add one vote (not enough).
	vote := MinimmitVote{
		ValidatorIndex: 0,
		Slot:           1,
		BlockRoot:      root,
		Stake:          100,
	}
	_ = engine.CastVote(vote)
	if engine.CheckFinality() {
		t.Error("should not be finalized with only 100/300 stake")
	}
}

func TestMinimmitMissSlot(t *testing.T) {
	cfg := &MinimmitConfig{
		TotalStake:           300,
		FinalityThresholdNum: 2,
		FinalityThresholdDen: 3,
	}
	engine, _ := NewMinimmitEngine(cfg)

	engine.MissSlot(5)
	if engine.State() != MinimmitFailed {
		t.Errorf("state = %s, want Failed", engine.State())
	}
	missed := engine.MissedSlots()
	if len(missed) != 1 || missed[0] != 5 {
		t.Errorf("missed slots = %v, want [5]", missed)
	}
}

func TestMinimmitReset(t *testing.T) {
	cfg := &MinimmitConfig{
		TotalStake:           300,
		FinalityThresholdNum: 2,
		FinalityThresholdDen: 3,
	}
	engine, _ := NewMinimmitEngine(cfg)
	root := makeMinimmitRoot(0xAA)
	_ = engine.ProposeBlock(1, root)

	engine.Reset()
	if engine.State() != MinimmitIdle {
		t.Errorf("state after reset = %s, want Idle", engine.State())
	}
	if engine.VoteCount() != 0 {
		t.Errorf("vote count after reset = %d, want 0", engine.VoteCount())
	}
}

func TestMinimmitFinalizedHead(t *testing.T) {
	cfg := &MinimmitConfig{
		TotalStake:           300,
		FinalityThresholdNum: 2,
		FinalityThresholdDen: 3,
	}
	engine, _ := NewMinimmitEngine(cfg)

	// Initially zero.
	slot, root := engine.FinalizedHead()
	if slot != 0 {
		t.Errorf("initial finalized slot = %d, want 0", slot)
	}
	if root != (types.Hash{}) {
		t.Errorf("initial finalized root should be zero hash")
	}

	// Finalize a slot.
	blockRoot := makeMinimmitRoot(0xCC)
	_ = engine.ProposeBlock(10, blockRoot)
	for i := 0; i < 3; i++ {
		_ = engine.CastVote(MinimmitVote{
			ValidatorIndex: ValidatorIndex(i),
			Slot:           10,
			BlockRoot:      blockRoot,
			Stake:          100,
		})
	}

	slot, root = engine.FinalizedHead()
	if slot != 10 {
		t.Errorf("finalized slot = %d, want 10", slot)
	}
	if root != blockRoot {
		t.Errorf("finalized root = %x, want %x", root, blockRoot)
	}
}

func TestMinimmitFullRound(t *testing.T) {
	cfg := &MinimmitConfig{
		TotalStake:           900,
		FinalityThresholdNum: 2,
		FinalityThresholdDen: 3,
	}
	engine, err := NewMinimmitEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Phase 1: Propose.
	root := makeMinimmitRoot(0xDD)
	if err := engine.ProposeBlock(42, root); err != nil {
		t.Fatalf("propose: %v", err)
	}
	if engine.State() != MinimmitVoting {
		t.Fatalf("state = %s, want Voting", engine.State())
	}

	// Phase 2: Cast votes (need >= 600 of 900).
	for i := 0; i < 6; i++ {
		vote := MinimmitVote{
			ValidatorIndex: ValidatorIndex(i),
			Slot:           42,
			BlockRoot:      root,
			Stake:          100,
		}
		if err := engine.CastVote(vote); err != nil {
			t.Fatalf("vote %d: %v", i, err)
		}
	}

	// Should be finalized.
	if !engine.CheckFinality() {
		t.Fatal("expected finality after 600/900 stake")
	}
	if engine.VoteCount() != 6 {
		t.Errorf("vote count = %d, want 6", engine.VoteCount())
	}

	slot, froot := engine.FinalizedHead()
	if slot != 42 || froot != root {
		t.Errorf("finalized head = (%d, %x), want (42, %x)", slot, froot, root)
	}

	// Phase 3: Reset for next round.
	engine.Reset()
	if engine.State() != MinimmitIdle {
		t.Errorf("state after reset = %s, want Idle", engine.State())
	}

	// Finalized head should be preserved.
	slot, froot = engine.FinalizedHead()
	if slot != 42 {
		t.Errorf("finalized slot after reset = %d, want 42", slot)
	}
}

func TestMinimmitMeetsThreshold(t *testing.T) {
	tests := []struct {
		name       string
		voteStake  uint64
		totalStake uint64
		num, den   uint64
		want       bool
	}{
		{"exact 2/3", 200, 300, 2, 3, true},
		{"above 2/3", 250, 300, 2, 3, true},
		{"below 2/3", 199, 300, 2, 3, false},
		{"all stake", 300, 300, 2, 3, true},
		{"zero vote", 0, 300, 2, 3, false},
		{"zero total", 100, 0, 2, 3, false},
		{"zero den", 100, 300, 2, 0, false},
		{"1/2 threshold", 150, 300, 1, 2, true},
		{"1/2 threshold below", 149, 300, 1, 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := meetsMinimmitThreshold(tt.voteStake, tt.totalStake, tt.num, tt.den)
			if got != tt.want {
				t.Errorf("meetsMinimmitThreshold(%d, %d, %d, %d) = %v, want %v",
					tt.voteStake, tt.totalStake, tt.num, tt.den, got, tt.want)
			}
		})
	}
}

func TestFinalityMode_String(t *testing.T) {
	tests := []struct {
		mode FinalityMode
		want string
	}{
		{FinalityModeClassic, "Classic"},
		{FinalityModeSSF, "SSF"},
		{FinalityModeMinimmit, "Minimmit"},
		{FinalityMode(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("FinalityMode(%d).String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

func TestMinimmitState_String(t *testing.T) {
	tests := []struct {
		state MinimmitState
		want  string
	}{
		{MinimmitIdle, "Idle"},
		{MinimmitProposed, "Proposed"},
		{MinimmitVoting, "Voting"},
		{MinimmitFinalized, "Finalized"},
		{MinimmitFailed, "Failed"},
		{MinimmitState(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("MinimmitState(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestMinimmitProposeBlock_AlreadyFinalized(t *testing.T) {
	cfg := &MinimmitConfig{
		TotalStake:           300,
		FinalityThresholdNum: 2,
		FinalityThresholdDen: 3,
	}
	engine, _ := NewMinimmitEngine(cfg)

	// Finalize slot 10.
	root := makeMinimmitRoot(0xAA)
	_ = engine.ProposeBlock(10, root)
	for i := 0; i < 3; i++ {
		_ = engine.CastVote(MinimmitVote{
			ValidatorIndex: ValidatorIndex(i),
			Slot:           10,
			BlockRoot:      root,
			Stake:          100,
		})
	}

	engine.Reset()

	// Try to propose for a slot <= finalized slot.
	err := engine.ProposeBlock(10, makeMinimmitRoot(0xBB))
	if !errors.Is(err, ErrMinimmitAlreadyFinal) {
		t.Errorf("expected ErrMinimmitAlreadyFinal, got %v", err)
	}

	// Proposing a later slot should succeed.
	err = engine.ProposeBlock(11, makeMinimmitRoot(0xCC))
	if err != nil {
		t.Errorf("unexpected error for later slot: %v", err)
	}
}

func TestMinimmitMissedSlotsAccumulate(t *testing.T) {
	cfg := &MinimmitConfig{
		TotalStake:           300,
		FinalityThresholdNum: 2,
		FinalityThresholdDen: 3,
	}
	engine, _ := NewMinimmitEngine(cfg)

	engine.MissSlot(1)
	engine.Reset()
	engine.MissSlot(3)
	engine.Reset()
	engine.MissSlot(7)

	missed := engine.MissedSlots()
	if len(missed) != 3 {
		t.Fatalf("missed slots count = %d, want 3", len(missed))
	}
	if missed[0] != 1 || missed[1] != 3 || missed[2] != 7 {
		t.Errorf("missed slots = %v, want [1, 3, 7]", missed)
	}
}

func TestMinimmitCastVote_NotVotingState(t *testing.T) {
	cfg := &MinimmitConfig{
		TotalStake:           300,
		FinalityThresholdNum: 2,
		FinalityThresholdDen: 3,
	}
	engine, _ := NewMinimmitEngine(cfg)

	// Engine is in Idle state, should reject vote.
	vote := MinimmitVote{
		ValidatorIndex: 0,
		Slot:           1,
		BlockRoot:      makeMinimmitRoot(0xAA),
		Stake:          100,
	}
	err := engine.CastVote(vote)
	if !errors.Is(err, ErrMinimmitInvalidState) {
		t.Errorf("expected ErrMinimmitInvalidState, got %v", err)
	}
}
