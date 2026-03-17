package node

import (
	"fmt"
	"log/slog"

	"github.com/eth2030/eth2030/core/block"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/core/vm"
	"github.com/eth2030/eth2030/engine"
)

func (b *engineBackend) ForkchoiceUpdated(
	fcState engine.ForkchoiceStateV1,
	payloadAttributes *engine.PayloadAttributesV3,
) (engine.ForkchoiceUpdatedResult, error) {
	bc := b.node.blockchain

	slog.Debug("engine_forkchoiceUpdated",
		"headBlockHash", fcState.HeadBlockHash,
		"safeBlockHash", fcState.SafeBlockHash,
		"finalizedBlockHash", fcState.FinalizedBlockHash,
		"hasPayloadAttrs", payloadAttributes != nil,
		"genesisHash", bc.Genesis().Hash(),
	)

	// Step 1: look up the forkchoice head block.
	headBlock := bc.GetBlock(fcState.HeadBlockHash)
	if headBlock == nil {
		slog.Warn("engine_forkchoiceUpdated: unknown head block, returning SYNCING",
			"headBlockHash", fcState.HeadBlockHash,
			"genesisHash", bc.Genesis().Hash(),
			"currentHead", bc.CurrentBlock().Hash(),
		)
		return engine.ForkchoiceUpdatedResult{
			PayloadStatus: engine.PayloadStatusV1{Status: engine.StatusSyncing},
		}, nil
	}

	headHash := headBlock.Hash()
	payloadStatus := engine.PayloadStatusV1{
		Status:          engine.StatusValid,
		LatestValidHash: &headHash,
	}

	// Step 2: update in-memory safe/finalized hashes.
	b.fcMu.Lock()
	finalHashChanged := fcState.FinalizedBlockHash != (types.Hash{}) &&
		fcState.FinalizedBlockHash != b.finalizedHash
	safeHashChanged := fcState.SafeBlockHash != (types.Hash{}) &&
		fcState.SafeBlockHash != b.safeHash
	if fcState.FinalizedBlockHash != (types.Hash{}) {
		b.finalizedHash = fcState.FinalizedBlockHash
	}
	if fcState.SafeBlockHash != (types.Hash{}) {
		b.safeHash = fcState.SafeBlockHash
	}
	b.fcMu.Unlock()

	if finalHashChanged {
		slog.Info("engine_forkchoiceUpdated: finalized block advanced",
			"finalizedHash", fcState.FinalizedBlockHash,
			"headNum", headBlock.NumberU64(),
		)
	}

	// Step 3: FCU cache check.
	triple := fcuCacheEntry{
		head:      headHash,
		safe:      fcState.SafeBlockHash,
		finalized: fcState.FinalizedBlockHash,
	}
	if payloadAttributes == nil && !finalHashChanged && !safeHashChanged {
		b.fcuCacheMu.Lock()
		hit := b.fcuCacheContains(triple)
		b.fcuCacheMu.Unlock()
		if hit {
			slog.Debug("engine_forkchoiceUpdated: cache hit, returning immediately",
				"headBlockHash", headHash,
			)
			return engine.ForkchoiceUpdatedResult{PayloadStatus: payloadStatus}, nil
		}
	}

	// Step 4: resolve finalized/safe blocks.
	var finalBlock, safeBlock *types.Block
	if finalHashChanged {
		if finalBlock = bc.GetBlock(fcState.FinalizedBlockHash); finalBlock == nil {
			slog.Warn("engine_forkchoiceUpdated: finalized block unknown, skipping",
				"hash", fcState.FinalizedBlockHash,
			)
		}
	}
	if safeHashChanged {
		if safeBlock = bc.GetBlock(fcState.SafeBlockHash); safeBlock == nil {
			slog.Warn("engine_forkchoiceUpdated: safe block unknown, skipping",
				"hash", fcState.SafeBlockHash,
			)
		}
	}

	// Step 5: push slow work to background goroutine.
	b.sendPostFCUWork(postFCUWork{
		fcState:    fcState,
		headBlock:  headBlock,
		finalBlock: finalBlock,
		safeBlock:  safeBlock,
		hasAttrs:   payloadAttributes != nil,
	})

	// Step 6: store triple in FCU cache.
	b.fcuCacheMu.Lock()
	b.fcuCache[b.fcuCacheWr] = triple
	b.fcuCacheWr = (b.fcuCacheWr + 1) % fcuCacheSize
	b.fcuCacheMu.Unlock()

	// Step 7: no payload attributes — return immediately.
	if payloadAttributes == nil {
		slog.Debug("engine_forkchoiceUpdated: no attrs, done",
			"headNum", headBlock.NumberU64(),
		)
		return engine.ForkchoiceUpdatedResult{PayloadStatus: payloadStatus}, nil
	}

	// Step 7a: check if we can quickly obtain the parent state.
	const stateThreshold = 32
	if !bc.CanQuicklyGetState(headBlock, stateThreshold) {
		slog.Warn("engine_forkchoiceUpdated: parent state needs deep re-execution, warming and returning SYNCING",
			"headNum", headBlock.NumberU64(),
			"headHash", headBlock.Hash(),
		)
		go func() {
			_, err := bc.StateAtBlock(headBlock)
			if err != nil {
				slog.Warn("engine_forkchoiceUpdated: background state warming failed",
					"headNum", headBlock.NumberU64(),
					"err", err,
				)
			} else {
				slog.Info("engine_forkchoiceUpdated: background state warming completed",
					"headNum", headBlock.NumberU64(),
				)
			}
		}()
		return engine.ForkchoiceUpdatedResult{
			PayloadStatus: engine.PayloadStatusV1{Status: engine.StatusSyncing},
		}, nil
	}

	// Step 7b: payload attributes provided — start async block build.
	slog.Debug("engine_forkchoiceUpdated: building payload",
		"parentNum", headBlock.NumberU64(),
		"parentHash", headBlock.Hash(),
		"timestamp", payloadAttributes.Timestamp,
		"feeRecipient", payloadAttributes.SuggestedFeeRecipient,
	)
	parentHeader := headBlock.Header()

	// Convert engine withdrawals to core types.
	var withdrawals []*types.Withdrawal
	for _, w := range payloadAttributes.Withdrawals {
		withdrawals = append(withdrawals, &types.Withdrawal{
			Index:          w.Index,
			ValidatorIndex: w.ValidatorIndex,
			Address:        w.Address,
			Amount:         w.Amount,
		})
	}

	beaconRoot := payloadAttributes.ParentBeaconBlockRoot
	attrs := &block.BuildBlockAttributes{
		Timestamp:    payloadAttributes.Timestamp,
		FeeRecipient: payloadAttributes.SuggestedFeeRecipient,
		Random:       payloadAttributes.PrevRandao,
		Withdrawals:  withdrawals,
		BeaconRoot:   &beaconRoot,
		GasLimit:     parentHeader.GasLimit,
	}

	// Generate payload ID.
	payloadID := generatePayloadID(parentHeader.Hash(), attrs)

	// Idempotency check.
	b.mu.Lock()
	if existing, ok := b.payloads[payloadID]; ok && existing != nil {
		b.mu.Unlock()
		slog.Debug("engine_forkchoiceUpdated: payload already building, reusing slot",
			"payloadID", payloadID,
		)
		return engine.ForkchoiceUpdatedResult{
			PayloadStatus: payloadStatus,
			PayloadID:     &payloadID,
		}, nil
	}

	// Register in-progress slot.
	pp := &pendingPayload{done: make(chan struct{})}
	b.payloads[payloadID] = pp
	b.payloadOrder = append(b.payloadOrder, payloadID)
	for len(b.payloads) > b.maxPayloads && len(b.payloadOrder) > 0 {
		oldest := b.payloadOrder[0]
		b.payloadOrder = b.payloadOrder[1:]
		delete(b.payloads, oldest)
	}
	b.mu.Unlock()

	// Build the block in the background.
	go func() {
		parentBlock := b.node.blockchain.GetBlock(parentHeader.Hash())
		var statedb state.StateDB
		if parentBlock != nil {
			var err error
			statedb, err = b.node.blockchain.StateAtBlock(parentBlock)
			if err != nil {
				slog.Warn("engine_forkchoiceUpdated: state fetch failed",
					"parentNum", parentHeader.Number,
					"parentHash", parentHeader.Hash(),
					"err", err,
				)
				pp.err = fmt.Errorf("state fetch: %w", err)
				close(pp.done)
				return
			}
		}
		slog.Debug("engine_forkchoiceUpdated: state ready",
			"parentNum", parentHeader.Number,
			"parentHash", parentHeader.Hash(),
		)

		b.buildMu.Lock()
		defer b.buildMu.Unlock()

		if statedb != nil {
			b.builder.SetState(statedb)
		}

		slog.Debug("engine_forkchoiceUpdated: calling BuildBlock",
			"parentNum", parentHeader.Number,
			"parentHash", parentHeader.Hash(),
		)
		builtBlock, receipts, err := b.builder.BuildBlock(parentHeader, attrs)
		if err != nil {
			slog.Warn("engine_forkchoiceUpdated: build block failed",
				"parentNum", parentHeader.Number,
				"err", err,
			)
			pp.err = fmt.Errorf("build block: %w", err)
			close(pp.done)
			return
		}

		// Replace VERIFY frame calldata with STARK proof when enabled.
		if prover := b.node.starkFrameProver; prover != nil {
			if sealed, _, serr := vm.ReplaceValidationFrames(builtBlock, prover); serr != nil {
				slog.Warn("frame stark replacement failed", "err", serr)
			} else {
				builtBlock = sealed
			}
		}

		slog.Debug("engine_forkchoiceUpdated: BuildBlock done",
			"blockNum", builtBlock.NumberU64(),
			"blockHash", builtBlock.Hash(),
			"txCount", len(builtBlock.Transactions()),
		)

		pp.block = builtBlock
		pp.receipts = receipts

		b.cacheBlobsFromBlock(builtBlock)

		if insertErr := b.node.blockchain.InsertBlock(builtBlock); insertErr != nil {
			slog.Warn("engine_forkchoiceUpdated: pre-insert built block failed",
				"blockNum", builtBlock.NumberU64(),
				"blockHash", builtBlock.Hash(),
				"err", insertErr,
			)
		}

		close(pp.done)

		slog.Info("engine_forkchoiceUpdated: built payload",
			"payloadID", payloadID,
			"blockNumber", builtBlock.NumberU64(),
			"blockHash", builtBlock.Hash(),
			"txCount", len(builtBlock.Transactions()),
		)
	}()

	return engine.ForkchoiceUpdatedResult{
		PayloadStatus: payloadStatus,
		PayloadID:     &payloadID,
	}, nil
}

// fcuCacheContains reports whether e is present in the FCU cache.
func (b *engineBackend) fcuCacheContains(e fcuCacheEntry) bool {
	for _, c := range b.fcuCache {
		if c == e {
			return true
		}
	}
	return false
}

// sendPostFCUWork dispatches work to the background goroutine non-blocking.
func (b *engineBackend) sendPostFCUWork(work postFCUWork) {
	select {
	case b.postFCUCh <- work:
	default:
		select {
		case <-b.postFCUCh:
		default:
		}
		select {
		case b.postFCUCh <- work:
		default:
		}
	}
}

// postFCULoop processes deferred FCU work in the background.
func (b *engineBackend) postFCULoop() {
	for {
		select {
		case <-b.stopCh:
			return
		case work := <-b.postFCUCh:
			b.doPostFCUWork(work)
		}
	}
}