package node

import (
	"log/slog"
	"sync"

	"github.com/eth2030/eth2030/core/block"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/engine"
	"github.com/eth2030/eth2030/engine/forkchoice"
	"github.com/eth2030/eth2030/engine/vhash"
)

// pendingPayload stores a built payload for later retrieval by getPayload.
type pendingPayload struct {
	block    *types.Block
	receipts []*types.Receipt
	err      error
	done     chan struct{}
}

// blockProcReq is a work item for the block processor goroutine.
type blockProcReq struct {
	payload               *engine.ExecutionPayloadV3
	parentBeaconBlockRoot types.Hash
	requestsHash          *types.Hash
	replyCh               chan blockProcResp
}

// blockProcResp carries the result of a processed block.
type blockProcResp struct {
	status engine.PayloadStatusV1
	err    error
}

const fcuCacheSize = 8

// fcuCacheEntry records a (head,safe,finalized) triple from a processed FCU.
type fcuCacheEntry struct {
	head      types.Hash
	safe      types.Hash
	finalized types.Hash
}

// postFCUWork carries slow state-update work deferred from ForkchoiceUpdated.
type postFCUWork struct {
	fcState    engine.ForkchoiceStateV1
	headBlock  *types.Block
	finalBlock *types.Block
	safeBlock  *types.Block
	hasAttrs   bool
}

// engineBackend adapts the Node to the engine.Backend interface.
type engineBackend struct {
	node *Node

	mu           sync.Mutex
	payloads     map[engine.PayloadID]*pendingPayload
	payloadOrder []engine.PayloadID
	maxPayloads  int
	builder      *block.BlockBuilder
	buildMu      sync.Mutex

	fcMu          sync.RWMutex
	safeHash      types.Hash
	finalizedHash types.Hash

	processCh chan blockProcReq
	procChCap int
	stopCh    chan struct{}

	fcuCache   [fcuCacheSize]fcuCacheEntry
	fcuCacheWr int
	fcuCacheMu sync.Mutex

	postFCUCh chan postFCUWork

	blobCache   map[types.Hash][]byte
	blobCacheMu sync.RWMutex
}

// newEngineBackend creates and starts the engine backend.
func newEngineBackend(n *Node) *engineBackend {
	pool := &txPoolAdapter{node: n}
	builder := block.NewBlockBuilder(n.blockchain.Config(), n.blockchain, pool)
	maxPayloads := n.config.CacheEnginePayloads
	if maxPayloads <= 0 {
		maxPayloads = 32
	}
	b := &engineBackend{
		node:        n,
		payloads:    make(map[engine.PayloadID]*pendingPayload),
		maxPayloads: maxPayloads,
		builder:     builder,
		processCh:   make(chan blockProcReq, 8),
		procChCap:   8,
		stopCh:      make(chan struct{}),
		postFCUCh:   make(chan postFCUWork, 1),
		blobCache:   make(map[types.Hash][]byte),
	}
	go b.processLoop()
	go b.postFCULoop()
	return b
}

// processLoop is the dedicated block processor goroutine.
func (b *engineBackend) processLoop() {
	for {
		select {
		case <-b.stopCh:
			return
		case req := <-b.processCh:
			status, err := b.execBlockInternal(req.payload, req.parentBeaconBlockRoot, req.requestsHash)
			req.replyCh <- blockProcResp{status: status, err: err}
		}
	}
}

// Close stops the processor goroutine.
func (b *engineBackend) Close() {
	close(b.stopCh)
}

func (b *engineBackend) GetHeadHash() types.Hash {
	if blk := b.node.blockchain.CurrentBlock(); blk != nil {
		return blk.Hash()
	}
	return types.Hash{}
}

func (b *engineBackend) GetSafeHash() types.Hash {
	if blk := b.node.blockchain.CurrentSafeBlock(); blk != nil {
		return blk.Hash()
	}
	b.fcMu.RLock()
	defer b.fcMu.RUnlock()
	return b.safeHash
}

func (b *engineBackend) GetFinalizedHash() types.Hash {
	if blk := b.node.blockchain.CurrentFinalBlock(); blk != nil {
		return blk.Hash()
	}
	b.fcMu.RLock()
	defer b.fcMu.RUnlock()
	return b.finalizedHash
}

func (b *engineBackend) ProcessBlock(
	payload *engine.ExecutionPayloadV3,
	expectedBlobVersionedHashes []types.Hash,
	parentBeaconBlockRoot types.Hash,
) (engine.PayloadStatusV1, error) {
	if len(expectedBlobVersionedHashes) > 0 {
		if err := vhash.VerifyAllBlobVersionBytes(expectedBlobVersionedHashes); err != nil {
			latestValid := payload.ParentHash
			slog.Warn("engine_newPayload: invalid blob versioned hash version byte", "err", err)
			return engine.PayloadStatusV1{
				Status:          engine.StatusInvalid,
				LatestValidHash: &latestValid,
			}, nil
		}
	}
	return b.processBlockInternal(payload, parentBeaconBlockRoot, nil)
}

func (b *engineBackend) processBlockInternal(
	payload *engine.ExecutionPayloadV3,
	parentBeaconBlockRoot types.Hash,
	requestsHash *types.Hash,
) (engine.PayloadStatusV1, error) {
	if depth := len(b.processCh); depth > b.procChCap/2 {
		slog.Warn("engine_newPayload: block processor queue growing",
			"depth", depth, "capacity", b.procChCap)
	}
	replyCh := make(chan blockProcResp, 1)
	b.processCh <- blockProcReq{
		payload:               payload,
		parentBeaconBlockRoot: parentBeaconBlockRoot,
		requestsHash:          requestsHash,
		replyCh:               replyCh,
	}
	resp := <-replyCh
	return resp.status, resp.err
}

// registerBlockInFCState adds the block to forkchoice state manager.
func (b *engineBackend) registerBlockInFCState(payload *engine.ExecutionPayloadV3) {
	if b.node.fcStateManager != nil {
		bi := &forkchoice.BlockInfo{
			Hash:       payload.BlockHash,
			ParentHash: payload.ParentHash,
			Number:     payload.BlockNumber,
			Slot:       payload.BlockNumber,
		}
		b.node.fcStateManager.AddBlock(bi)
		if b.node.fcTracker != nil {
			b.node.fcTracker.Reorgs.AddBlock(bi)
		}
	}
}