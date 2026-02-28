// bft_finality_pipeline.go connects the Minimmit 1-round BFT output to a
// finality tracking pipeline. It orchestrates the flow from completed SSF
// rounds through BLS signature aggregation, finality proof generation via
// FinalityBLSAdapter, and checkpoint storage for external consumers.
//
// This bridges the SSFRoundEngine and FinalityBLSAdapter into a cohesive
// pipeline that tracks finalized epochs/slots with their proofs.
package consensus

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/eth2030/eth2030/core/types"
)

// BFT finality pipeline errors.
var (
	ErrBFTPipelineNilConfig        = errors.New("bft_pipeline: nil config")
	ErrBFTPipelineNilRound         = errors.New("bft_pipeline: nil round")
	ErrBFTPipelineRoundNotFinalized = errors.New("bft_pipeline: round is not finalized")
	ErrBFTPipelineNoVotes          = errors.New("bft_pipeline: round has no votes")
	ErrBFTPipelineInvalidEpoch     = errors.New("bft_pipeline: invalid epoch")
	ErrBFTPipelineStoreCapacity    = errors.New("bft_pipeline: checkpoint store at capacity")
	ErrBFTPipelineInsufficientStake = errors.New("bft_pipeline: insufficient stake in round")
	ErrBFTPipelineZeroConfirmation = errors.New("bft_pipeline: confirmation depth must be > 0")
	ErrBFTPipelineZeroInterval     = errors.New("bft_pipeline: checkpoint interval must be > 0")
	ErrBFTPipelineZeroMaxPending   = errors.New("bft_pipeline: max pending must be > 0")
)

// BFTPipelineCheckpoint is a finalized checkpoint produced by the BFT
// finality pipeline. It records the epoch, slot, block root, state root,
// and the corresponding finality proof with aggregate BLS signature.
type BFTPipelineCheckpoint struct {
	// Epoch is the epoch of this checkpoint.
	Epoch uint64

	// Slot is the slot within the epoch that was finalized.
	Slot uint64

	// BlockRoot is the finalized block root.
	BlockRoot types.Hash

	// StateRoot is the post-state root at finalization.
	StateRoot types.Hash

	// Proof is the finality proof containing the aggregate BLS signature,
	// participant bitfield, and stake totals.
	Proof *FinalityProof

	// ParticipantCount is the number of validators that voted.
	ParticipantCount uint64

	// TotalStake is the total stake that voted for this block root.
	TotalStake uint64
}

// BFTPipelineCfg configures the BFT finality pipeline.
type BFTPipelineCfg struct {
	// ConfirmationDepth is the minimum number of finalized rounds
	// required before a checkpoint is considered confirmed.
	ConfirmationDepth uint64

	// CheckpointInterval determines how frequently (in epochs) the
	// pipeline produces stored checkpoints. E.g., 1 = every epoch.
	CheckpointInterval uint64

	// MaxPending is the maximum number of checkpoints to retain
	// in the store before pruning the oldest ones.
	MaxPending int

	// MinStakeThresholdNum and MinStakeThresholdDen express the
	// minimum fraction of total stake that must be present in the
	// round for the pipeline to accept it (default 2/3).
	MinStakeThresholdNum uint64
	MinStakeThresholdDen uint64

	// TotalStake is the total active validator stake for threshold checks.
	TotalStake uint64
}

// DefaultBFTPipelineCfg returns production defaults for the pipeline.
func DefaultBFTPipelineCfg() *BFTPipelineCfg {
	return &BFTPipelineCfg{
		ConfirmationDepth:    1,
		CheckpointInterval:   1,
		MaxPending:           256,
		MinStakeThresholdNum: 2,
		MinStakeThresholdDen: 3,
		TotalStake:           32_000_000 * GweiPerETH,
	}
}

// ValidateBFTPipelineCfg checks the config for correctness.
func ValidateBFTPipelineCfg(cfg *BFTPipelineCfg) error {
	if cfg == nil {
		return ErrBFTPipelineNilConfig
	}
	if cfg.ConfirmationDepth == 0 {
		return ErrBFTPipelineZeroConfirmation
	}
	if cfg.CheckpointInterval == 0 {
		return ErrBFTPipelineZeroInterval
	}
	if cfg.MaxPending == 0 {
		return ErrBFTPipelineZeroMaxPending
	}
	return nil
}

// BFTCheckpointStore is an in-memory store for recent finalized checkpoints
// produced by the BFT finality pipeline. It supports epoch-based lookup
// and automatic pruning when capacity is exceeded. Thread-safe.
type BFTCheckpointStore struct {
	mu          sync.RWMutex
	checkpoints map[uint64]*BFTPipelineCheckpoint // keyed by epoch
	epochs      []uint64                          // sorted ascending
	maxSize     int
}

// NewBFTCheckpointStore creates a store with the given maximum capacity.
func NewBFTCheckpointStore(maxSize int) *BFTCheckpointStore {
	if maxSize <= 0 {
		maxSize = 256
	}
	return &BFTCheckpointStore{
		checkpoints: make(map[uint64]*BFTPipelineCheckpoint),
		epochs:      make([]uint64, 0),
		maxSize:     maxSize,
	}
}

// Put stores a checkpoint, replacing any existing one for the same epoch.
// If the store exceeds capacity, the oldest checkpoint is pruned.
func (s *BFTCheckpointStore) Put(cp *BFTPipelineCheckpoint) {
	if cp == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.checkpoints[cp.Epoch]; !exists {
		s.epochs = append(s.epochs, cp.Epoch)
		sort.Slice(s.epochs, func(i, j int) bool { return s.epochs[i] < s.epochs[j] })
	}

	cpCopy := *cp
	if cp.Proof != nil {
		proofCopy := *cp.Proof
		cpCopy.Proof = &proofCopy
	}
	s.checkpoints[cp.Epoch] = &cpCopy

	// Prune oldest if over capacity.
	for len(s.epochs) > s.maxSize {
		oldest := s.epochs[0]
		s.epochs = s.epochs[1:]
		delete(s.checkpoints, oldest)
	}
}

// Get retrieves the checkpoint for the given epoch.
func (s *BFTCheckpointStore) Get(epoch uint64) (*BFTPipelineCheckpoint, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp, ok := s.checkpoints[epoch]
	if !ok {
		return nil, false
	}
	cpCopy := *cp
	return &cpCopy, true
}

// Latest returns the most recent checkpoint, or nil if empty.
func (s *BFTCheckpointStore) Latest() *BFTPipelineCheckpoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.epochs) == 0 {
		return nil
	}
	latest := s.epochs[len(s.epochs)-1]
	cp := s.checkpoints[latest]
	cpCopy := *cp
	return &cpCopy
}

// Count returns the number of stored checkpoints.
func (s *BFTCheckpointStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.checkpoints)
}

// PruneOlderThan removes all checkpoints with epoch < cutoff.
// Returns the number of checkpoints pruned.
func (s *BFTCheckpointStore) PruneOlderThan(cutoff uint64) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	pruned := 0
	remaining := make([]uint64, 0, len(s.epochs))
	for _, ep := range s.epochs {
		if ep < cutoff {
			delete(s.checkpoints, ep)
			pruned++
		} else {
			remaining = append(remaining, ep)
		}
	}
	s.epochs = remaining
	return pruned
}

// Epochs returns all stored checkpoint epochs in ascending order.
func (s *BFTCheckpointStore) Epochs() []uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]uint64, len(s.epochs))
	copy(result, s.epochs)
	return result
}

// BFTFinalityPipeline orchestrates the flow from completed SSFRound output
// through BLS signature aggregation (via FinalityBLSAdapter), finality proof
// generation, and checkpoint storage. Thread-safe.
type BFTFinalityPipeline struct {
	mu      sync.Mutex
	config  *BFTPipelineCfg
	adapter *FinalityBLSAdapter
	store   *BFTCheckpointStore

	// processedCount tracks total rounds processed.
	processedCount uint64

	// latestEpoch tracks the most recently finalized epoch.
	latestEpoch uint64
}

// NewBFTFinalityPipeline creates a new BFT finality pipeline with the
// given config. Uses the standard FinalityBLSAdapter for proof generation.
func NewBFTFinalityPipeline(cfg *BFTPipelineCfg) (*BFTFinalityPipeline, error) {
	if err := ValidateBFTPipelineCfg(cfg); err != nil {
		return nil, err
	}
	return &BFTFinalityPipeline{
		config:  cfg,
		adapter: NewFinalityBLSAdapter(),
		store:   NewBFTCheckpointStore(cfg.MaxPending),
	}, nil
}

// ProcessRound takes a completed SSFRound, generates a finality proof via
// the BLS adapter, and stores the resulting checkpoint. The round must be
// finalized (round.Finalized == true). Returns the checkpoint or an error.
func (fp *BFTFinalityPipeline) ProcessRound(
	round *SSFRound,
	epoch uint64,
	stateRoot types.Hash,
) (*BFTPipelineCheckpoint, error) {
	if round == nil {
		return nil, ErrBFTPipelineNilRound
	}
	if !round.Finalized {
		return nil, ErrBFTPipelineRoundNotFinalized
	}
	if len(round.Votes) == 0 {
		return nil, ErrBFTPipelineNoVotes
	}

	// Verify the round meets the minimum stake threshold.
	if fp.config.TotalStake > 0 && fp.config.MinStakeThresholdDen > 0 {
		if !roundMeetsStakeThreshold(
			round.TotalVoteStake,
			fp.config.TotalStake,
			fp.config.MinStakeThresholdNum,
			fp.config.MinStakeThresholdDen,
		) {
			return nil, fmt.Errorf("%w: vote stake %d, total %d, threshold %d/%d",
				ErrBFTPipelineInsufficientStake,
				round.TotalVoteStake,
				fp.config.TotalStake,
				fp.config.MinStakeThresholdNum,
				fp.config.MinStakeThresholdDen,
			)
		}
	}

	// Generate finality proof via the BLS adapter.
	proof, err := fp.adapter.GenerateFinalityProof(round, epoch, stateRoot)
	if err != nil {
		return nil, fmt.Errorf("bft_pipeline: proof generation failed: %w", err)
	}

	cp := &BFTPipelineCheckpoint{
		Epoch:            epoch,
		Slot:             round.Slot,
		BlockRoot:        round.BlockRoot,
		StateRoot:        stateRoot,
		Proof:            proof,
		ParticipantCount: proof.ParticipantCount,
		TotalStake:       proof.TotalStake,
	}

	fp.mu.Lock()
	defer fp.mu.Unlock()

	fp.store.Put(cp)
	fp.processedCount++
	if epoch > fp.latestEpoch {
		fp.latestEpoch = epoch
	}

	return cp, nil
}

// LatestCheckpoint returns the most recently stored checkpoint,
// or nil if no checkpoints have been produced.
func (fp *BFTFinalityPipeline) LatestCheckpoint() *BFTPipelineCheckpoint {
	fp.mu.Lock()
	defer fp.mu.Unlock()
	return fp.store.Latest()
}

// CheckpointByEpoch returns the checkpoint for the given epoch, if it exists.
func (fp *BFTFinalityPipeline) CheckpointByEpoch(epoch uint64) (*BFTPipelineCheckpoint, bool) {
	fp.mu.Lock()
	defer fp.mu.Unlock()
	return fp.store.Get(epoch)
}

// PruneOlderThan removes all checkpoints with epoch < cutoff.
// Returns the number of checkpoints pruned.
func (fp *BFTFinalityPipeline) PruneOlderThan(epoch uint64) int {
	fp.mu.Lock()
	defer fp.mu.Unlock()
	return fp.store.PruneOlderThan(epoch)
}

// ProcessedCount returns the total number of rounds processed.
func (fp *BFTFinalityPipeline) ProcessedCount() uint64 {
	fp.mu.Lock()
	defer fp.mu.Unlock()
	return fp.processedCount
}

// LatestEpoch returns the most recently finalized epoch.
func (fp *BFTFinalityPipeline) LatestEpoch() uint64 {
	fp.mu.Lock()
	defer fp.mu.Unlock()
	return fp.latestEpoch
}

// CheckpointCount returns the number of stored checkpoints.
func (fp *BFTFinalityPipeline) CheckpointCount() int {
	fp.mu.Lock()
	defer fp.mu.Unlock()
	return fp.store.Count()
}

// StoredEpochs returns all stored checkpoint epochs in ascending order.
func (fp *BFTFinalityPipeline) StoredEpochs() []uint64 {
	fp.mu.Lock()
	defer fp.mu.Unlock()
	return fp.store.Epochs()
}

// Adapter returns the underlying BLS adapter for external use (e.g., proof
// verification).
func (fp *BFTFinalityPipeline) Adapter() *FinalityBLSAdapter {
	return fp.adapter
}

// roundMeetsStakeThreshold checks if a round's total vote stake meets the
// supermajority requirement: voteStake * den >= totalStake * num.
func roundMeetsStakeThreshold(voteStake, totalStake, num, den uint64) bool {
	if totalStake == 0 || den == 0 {
		return false
	}
	return voteStake*den >= totalStake*num
}
