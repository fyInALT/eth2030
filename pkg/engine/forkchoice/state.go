// Package forkchoice implements fork choice state management for the Engine API.
//
// This provides the execution-layer's view of the consensus fork choice:
//   - Justified and finalized checkpoint tracking
//   - Head block determination with safe/unsafe distinction
//   - Proposer boost accounting per the LMD-GHOST fork choice rule
//   - Reorg detection and notification to subscribers
//
// The ForkchoiceStateManager sits between the Engine API and the block store,
// maintaining a consistent view of the canonical chain head as the CL sends
// forkchoiceUpdated calls. It detects reorgs by comparing the new head against
// the previous head and notifies registered listeners.
//
// Reference: consensus-specs/specs/phase0/fork-choice.md, execution-apis
package forkchoice

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/engine/payload"
)

// ChainReader is the minimal interface into the persistent block store used by
// ForkchoiceStateManager to resolve ancestry when a block is not in the
// in-memory cache. *chain.Blockchain satisfies this interface.
type ChainReader interface {
	// GetBlock returns the full block for the given hash, or nil if unknown.
	GetBlock(hash types.Hash) *types.Block
}

// Fork choice state errors.
var (
	ErrFCStateNilUpdate       = errors.New("forkchoice_state: nil update")
	ErrFCStateZeroHead        = errors.New("forkchoice_state: head block hash is zero")
	ErrFCStateFinalizedAhead  = errors.New("forkchoice_state: finalized hash ahead of justified")
	ErrFCStateHeadNotFound    = errors.New("forkchoice_state: head block not found in store")
	ErrFCStateSafeNotAncestor = errors.New("forkchoice_state: safe block not ancestor of head")
)

// Checkpoint represents a justified or finalized checkpoint.
type Checkpoint struct {
	// Epoch is the checkpoint epoch.
	Epoch uint64

	// Root is the block root at the checkpoint boundary.
	Root types.Hash
}

// ProposerBoost holds the current proposer boost state per LMD-GHOST.
// The proposer of the current slot receives a boost to its fork-choice
// weight for a limited duration to prevent short-range reorgs.
type ProposerBoost struct {
	// Slot is the slot for which the boost is active.
	Slot uint64

	// BlockRoot is the root of the block receiving the boost.
	BlockRoot types.Hash

	// BoostWeight is the additional weight applied (committee weight * 40/100).
	BoostWeight uint64
}

// ReorgEvent describes a chain reorganization detected by the fork choice.
type ReorgEvent struct {
	// Slot is the slot at which the reorg was detected.
	Slot uint64

	// OldHead is the previous head block hash.
	OldHead types.Hash

	// NewHead is the new head block hash after the reorg.
	NewHead types.Hash

	// Depth is the number of blocks reorganized (distance to common ancestor).
	Depth uint64

	// OldHeadNumber is the block number of the old head.
	OldHeadNumber uint64

	// NewHeadNumber is the block number of the new head.
	NewHeadNumber uint64
}

// ReorgListener is a callback invoked when a chain reorg is detected.
type ReorgListener func(event ReorgEvent)

// BlockInfo stores minimal block metadata needed for fork choice.
type BlockInfo struct {
	Hash       types.Hash
	ParentHash types.Hash
	Number     uint64
	Slot       uint64
}

// defaultFCPruneBuffer is the default number of blocks retained behind the
// finalized head before pruning. Covers short reorgs and ancestry walks.
const defaultFCPruneBuffer = 128

// ForkchoiceStateManager manages the fork choice state on the execution layer.
// It tracks justified/finalized checkpoints, maintains the head/safe/finalized
// block distinction, accounts for proposer boost, and detects reorgs.
//
// All public methods are safe for concurrent use.
type ForkchoiceStateManager struct {
	mu sync.RWMutex

	// Current fork choice pointers.
	headHash      types.Hash
	safeHash      types.Hash
	finalizedHash types.Hash

	// Checkpoint tracking.
	justifiedCheckpoint Checkpoint
	finalizedCheckpoint Checkpoint

	// Proposer boost state.
	currentBoost *ProposerBoost

	// Block metadata store for ancestry checks and reorg depth.
	blocks map[types.Hash]*BlockInfo

	// pruneBuffer is the number of blocks behind finality kept in memory.
	// Blocks older than finalized-pruneBuffer are pruned on finality advance.
	pruneBuffer uint64

	// chain is an optional fallback for block lookups not in the in-memory
	// store. When set, isAncestor and reorgDepth walk the persistent DB
	// instead of short-circuiting on a cache miss.
	chain ChainReader

	// Reorg detection.
	reorgListeners []ReorgListener

	// Statistics.
	updateCount atomic.Uint64
	reorgCount  atomic.Uint64
}

// NewForkchoiceStateManager creates a new fork choice state manager with
// the default prune buffer. If genesis is non-nil, it seeds head/safe/finalized.
func NewForkchoiceStateManager(genesis *BlockInfo) *ForkchoiceStateManager {
	return NewForkchoiceStateManagerWithBuffer(genesis, defaultFCPruneBuffer)
}

// NewForkchoiceStateManagerWithBuffer creates a fork choice state manager
// that retains pruneBuffer blocks behind the finalized head before pruning.
// Use 0 to disable automatic pruning.
func NewForkchoiceStateManagerWithBuffer(genesis *BlockInfo, pruneBuffer uint64) *ForkchoiceStateManager {
	m := &ForkchoiceStateManager{
		blocks:      make(map[types.Hash]*BlockInfo),
		pruneBuffer: pruneBuffer,
	}
	if genesis != nil {
		m.blocks[genesis.Hash] = genesis
		m.headHash = genesis.Hash
		m.safeHash = genesis.Hash
		m.finalizedHash = genesis.Hash
		m.justifiedCheckpoint = Checkpoint{Root: genesis.Hash}
		m.finalizedCheckpoint = Checkpoint{Root: genesis.Hash}
	}
	return m
}

// SetChain wires a persistent block store so that ancestry walks can fall back
// to the DB when a block is absent from the in-memory cache (e.g. after restart).
// Safe to call concurrently; typically called once during node startup.
func (m *ForkchoiceStateManager) SetChain(chain ChainReader) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.chain = chain
}

// AddBlock registers a block in the fork choice store. This must be called
// for all blocks the node knows about so ancestry lookups work.
func (m *ForkchoiceStateManager) AddBlock(info *BlockInfo) {
	if info == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blocks[info.Hash] = info
}

// ProcessForkchoiceUpdate applies a forkchoice state update from the CL.
// It validates the update, detects reorgs, updates checkpoints, and
// notifies registered listeners of any reorg.
func (m *ForkchoiceStateManager) ProcessForkchoiceUpdate(update payload.ForkchoiceStateV1) error {
	if update.HeadBlockHash == (types.Hash{}) {
		return ErrFCStateZeroHead
	}

	m.mu.Lock()

	m.updateCount.Add(1)

	// Verify the head block is known.
	headInfo, headKnown := m.blocks[update.HeadBlockHash]
	if !headKnown {
		m.mu.Unlock()
		return fmt.Errorf("%w: %s", ErrFCStateHeadNotFound, update.HeadBlockHash.Hex())
	}

	// Detect reorg: if the new head differs from the old head, and the new
	// head is not a direct descendant of the old head, it is a reorg.
	// Note: isAncestorLocked walks the in-memory blocks map. If intermediate
	// blocks are absent (e.g. after restart), the walk short-circuits and
	// returns false — treating the situation as a reorg rather than an
	// extension. Guard: both old and new head must be known in the map.
	oldHead := m.headHash
	var reorgEvent *ReorgEvent
	if oldHead != (types.Hash{}) && oldHead != update.HeadBlockHash {
		oldInfo := m.blocks[oldHead]
		if oldInfo != nil && !m.isAncestorLocked(oldHead, update.HeadBlockHash) {
			depth := m.reorgDepthLocked(oldHead, update.HeadBlockHash)
			reorgEvent = &ReorgEvent{
				Slot:          headInfo.Slot,
				OldHead:       oldHead,
				NewHead:       update.HeadBlockHash,
				Depth:         depth,
				OldHeadNumber: oldInfo.Number,
				NewHeadNumber: headInfo.Number,
			}
			m.reorgCount.Add(1)
		}
	}

	// Update fork choice pointers.
	m.headHash = update.HeadBlockHash
	m.safeHash = update.SafeBlockHash
	m.finalizedHash = update.FinalizedBlockHash

	// Update finalized checkpoint if the finalized hash changed.
	if update.FinalizedBlockHash != (types.Hash{}) {
		if finInfo, ok := m.blocks[update.FinalizedBlockHash]; ok {
			m.finalizedCheckpoint = Checkpoint{
				Epoch: finInfo.Slot / 32, // slots per epoch
				Root:  update.FinalizedBlockHash,
			}
			// Prune blocks well behind finality to bound memory usage.
			// m.pruneBuffer controls how many blocks to keep behind finality.
			if m.pruneBuffer > 0 && finInfo.Number > m.pruneBuffer {
				m.pruneBeforeNumberLocked(finInfo.Number - m.pruneBuffer)
			}
		}
	}

	// Update justified checkpoint from safe hash (safe ~ justified in PoS).
	if update.SafeBlockHash != (types.Hash{}) {
		if safeInfo, ok := m.blocks[update.SafeBlockHash]; ok {
			m.justifiedCheckpoint = Checkpoint{
				Epoch: safeInfo.Slot / 32,
				Root:  update.SafeBlockHash,
			}
		}
	}

	// Snapshot the listener list before releasing the lock. Listeners are
	// called outside the lock to avoid deadlocks: a listener may call back
	// into the txpool, blockchain, or any other subsystem that could
	// re-enter the forkchoice state manager (e.g. via Head() or AddBlock()).
	// This mirrors how geth uses event.Feed: state is committed first, then
	// subscribers are notified asynchronously via channels.
	var listeners []ReorgListener
	if reorgEvent != nil {
		listeners = make([]ReorgListener, len(m.reorgListeners))
		copy(listeners, m.reorgListeners)
	}

	m.mu.Unlock()

	// Invoke reorg listeners outside the lock.
	for _, l := range listeners {
		l(*reorgEvent)
	}

	return nil
}

// Head returns the current head block hash.
func (m *ForkchoiceStateManager) Head() types.Hash {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.headHash
}

// SafeHead returns the current safe (justified) block hash.
func (m *ForkchoiceStateManager) SafeHead() types.Hash {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.safeHash
}

// FinalizedHead returns the current finalized block hash.
func (m *ForkchoiceStateManager) FinalizedHead() types.Hash {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.finalizedHash
}

// HeadInfo returns full block info for the current head, or nil if unknown.
func (m *ForkchoiceStateManager) HeadInfo() *BlockInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	info := m.blocks[m.headHash]
	if info == nil {
		return nil
	}
	cp := *info
	return &cp
}

// JustifiedCheckpoint returns the current justified checkpoint.
func (m *ForkchoiceStateManager) JustifiedCheckpoint() Checkpoint {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.justifiedCheckpoint
}

// FinalizedCheckpoint returns the current finalized checkpoint.
func (m *ForkchoiceStateManager) FinalizedCheckpoint() Checkpoint {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.finalizedCheckpoint
}

// SetProposerBoost sets the proposer boost for the current slot.
// This gives extra fork-choice weight to the timely block.
func (m *ForkchoiceStateManager) SetProposerBoost(slot uint64, blockRoot types.Hash, weight uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentBoost = &ProposerBoost{
		Slot:        slot,
		BlockRoot:   blockRoot,
		BoostWeight: weight,
	}
}

// ClearProposerBoost clears the current proposer boost (e.g., at slot boundary).
func (m *ForkchoiceStateManager) ClearProposerBoost() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentBoost = nil
}

// ProposerBoostFor returns the proposer boost for a given block root, or zero
// if no boost is active or the boost is for a different block.
func (m *ForkchoiceStateManager) ProposerBoostFor(blockRoot types.Hash) uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.currentBoost != nil && m.currentBoost.BlockRoot == blockRoot {
		return m.currentBoost.BoostWeight
	}
	return 0
}

// GetCurrentBoost returns a copy of the current proposer boost, or nil.
func (m *ForkchoiceStateManager) GetCurrentBoost() *ProposerBoost {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.currentBoost == nil {
		return nil
	}
	cp := *m.currentBoost
	return &cp
}

// IsHeadSafe returns true if the current head equals the safe head.
func (m *ForkchoiceStateManager) IsHeadSafe() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.headHash == m.safeHash && m.headHash != (types.Hash{})
}

// IsHeadFinalized returns true if the current head equals the finalized head.
func (m *ForkchoiceStateManager) IsHeadFinalized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.headHash == m.finalizedHash && m.headHash != (types.Hash{})
}

// OnReorg registers a listener that is called whenever a chain reorg is detected.
func (m *ForkchoiceStateManager) OnReorg(listener ReorgListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reorgListeners = append(m.reorgListeners, listener)
}

// Stats returns fork choice statistics.
func (m *ForkchoiceStateManager) Stats() (updateCount, reorgCount uint64) {
	return m.updateCount.Load(), m.reorgCount.Load()
}

// GetForkchoiceState returns the current fork choice state as a
// ForkchoiceStateV1 suitable for Engine API responses.
func (m *ForkchoiceStateManager) GetForkchoiceState() payload.ForkchoiceStateV1 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return payload.ForkchoiceStateV1{
		HeadBlockHash:      m.headHash,
		SafeBlockHash:      m.safeHash,
		FinalizedBlockHash: m.finalizedHash,
	}
}

// PruneBeforeNumber removes block metadata for blocks with number < n.
// Does not remove blocks that are currently referenced as head/safe/finalized.
func (m *ForkchoiceStateManager) PruneBeforeNumber(n uint64) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.pruneBeforeNumberLocked(n)
}

// pruneBeforeNumberLocked is the lock-free inner implementation.
// Caller must hold m.mu (write).
func (m *ForkchoiceStateManager) pruneBeforeNumberLocked(n uint64) int {
	pruned := 0
	for hash, info := range m.blocks {
		if info.Number < n {
			// Do not prune blocks that are currently referenced.
			if hash == m.headHash || hash == m.safeHash || hash == m.finalizedHash {
				continue
			}
			delete(m.blocks, hash)
			pruned++
		}
	}
	return pruned
}

// BlockCount returns the number of blocks in the store.
func (m *ForkchoiceStateManager) BlockCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.blocks)
}

// --- internal helpers ---

// parentOf returns the parent hash of the given block, consulting the in-memory
// store first and then falling back to the persistent chain DB.
// Returns (zero, false) if the parent cannot be resolved.
// Caller must hold m.mu (read or write).
func (m *ForkchoiceStateManager) parentOf(hash types.Hash) (types.Hash, bool) {
	if info, ok := m.blocks[hash]; ok {
		return info.ParentHash, true
	}
	if m.chain != nil {
		if blk := m.chain.GetBlock(hash); blk != nil {
			return blk.Header().ParentHash, true
		}
	}
	return types.Hash{}, false
}

// isAncestorLocked checks if ancestorHash is an ancestor of descendantHash
// by walking the parent chain. Falls back to the persistent DB for blocks not
// in the in-memory store so a restart does not produce false-positive reorgs.
// Caller must hold at least m.mu read lock.
func (m *ForkchoiceStateManager) isAncestorLocked(ancestorHash, descendantHash types.Hash) bool {
	current := descendantHash
	for i := 0; i < 1024; i++ {
		if current == ancestorHash {
			return true
		}
		parent, ok := m.parentOf(current)
		if !ok || parent == current {
			// Block unknown or self-referencing (genesis / broken chain).
			return false
		}
		current = parent
	}
	return false
}

// reorgDepthLocked computes the depth of a reorg by finding the common
// ancestor between oldHead and newHead. Falls back to the persistent DB for
// blocks not in the in-memory store. Returns 0 if no common ancestor is found
// within 1024 blocks. Caller must hold at least m.mu read lock.
func (m *ForkchoiceStateManager) reorgDepthLocked(oldHead, newHead types.Hash) uint64 {
	// Collect ancestors of oldHead with their distance.
	oldAncestors := make(map[types.Hash]uint64)
	current := oldHead
	for dist := uint64(0); dist < 1024; dist++ {
		oldAncestors[current] = dist
		parent, ok := m.parentOf(current)
		if !ok || parent == current {
			break
		}
		current = parent
	}

	// Walk newHead's ancestors to find the first shared entry.
	current = newHead
	for dist := uint64(0); dist < 1024; dist++ {
		if oldDist, found := oldAncestors[current]; found {
			if dist > oldDist {
				return dist
			}
			return oldDist
		}
		parent, ok := m.parentOf(current)
		if !ok || parent == current {
			break
		}
		current = parent
	}

	return 0
}
