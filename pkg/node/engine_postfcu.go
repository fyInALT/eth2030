package node

import (
	"log/slog"

	"github.com/eth2030/eth2030/engine/forkchoice"
	epbsbid "github.com/eth2030/eth2030/epbs/bid"
)

// doPostFCUWork executes slow operations deferred from ForkchoiceUpdated.
func (b *engineBackend) doPostFCUWork(work postFCUWork) {
	bc := b.node.blockchain
	headBlock := work.headBlock
	headNum := headBlock.NumberU64()

	// Persist finalized block to rawdb.
	if work.finalBlock != nil {
		bc.SetFinalized(work.finalBlock)
		slog.Debug("engine_forkchoiceUpdated: finalized persisted",
			"hash", work.finalBlock.Hash(),
			"number", work.finalBlock.NumberU64(),
		)
		if b.node.txPool != nil {
			b.node.txPool.Reset(bc.State())
		}
	}

	// Persist safe block to rawdb.
	if work.safeBlock != nil {
		bc.SetSafe(work.safeBlock)
	}

	// Forkchoice state manager (reorg detection).
	if b.node.fcStateManager != nil {
		headInfo := &forkchoice.BlockInfo{
			Hash:       headBlock.Hash(),
			ParentHash: headBlock.Header().ParentHash,
			Number:     headBlock.NumberU64(),
			Slot:       headBlock.NumberU64(),
		}
		b.node.fcStateManager.AddBlock(headInfo)
		if b.node.fcTracker != nil {
			b.node.fcTracker.Reorgs.AddBlock(headInfo)
		}
		if err := b.node.fcStateManager.ProcessForkchoiceUpdate(work.fcState); err != nil {
			slog.Debug("fcStateManager update", "err", err)
		}
	}

	// High-level tracker: conflict detection, FCU history, reorg analytics.
	if b.node.fcTracker != nil {
		safeNum := uint64(0)
		finalNum := uint64(0)
		if work.safeBlock != nil {
			safeNum = work.safeBlock.NumberU64()
		}
		if work.finalBlock != nil {
			finalNum = work.finalBlock.NumberU64()
		}
		conflict, reason, reorg := b.node.fcTracker.ProcessUpdate(
			work.fcState, work.hasAttrs, headNum, safeNum, finalNum,
		)
		if conflict {
			slog.Warn("forkchoice conflict detected", "reason", reason)
		}
		if reorg != nil {
			slog.Warn("forkchoice tracker: reorg",
				"depth", reorg.Depth,
				"oldHead", reorg.OldHead,
				"newHead", reorg.NewHead,
			)
		}
	}

	// ePBS auction lifecycle (Amsterdam EIP-7732).
	if bc.Config().IsAmsterdam(headBlock.Time()) {
		if b.node.epbsAuction != nil {
			if err := b.node.epbsAuction.OpenAuction(headNum); err != nil {
				slog.Debug("epbs: open auction", "slot", headNum, "err", err)
			}
		}
		if headNum > 32 {
			if b.node.epbsBuilder != nil {
				b.node.epbsBuilder.PruneBefore(headNum - 32)
			}
			if b.node.epbsEscrow != nil {
				b.node.epbsEscrow.PruneBefore(headNum - 32)
			}
			if b.node.epbsCommit != nil {
				b.node.epbsCommit.PruneSlot(headNum - 32)
			}
		}
		if b.node.epbsBid != nil {
			components := epbsbid.ScoreComponents{
				BidAmount:        0,
				ReputationScore:  50.0,
				InclusionQuality: 1.0,
				LatencyMs:        0,
			}
			score := b.node.epbsBid.ComputeScore(components)
			slog.Debug("epbs: bid baseline score", "slot", headNum, "score", score)
		}
		if b.node.engineAuction != nil {
			if result, aErr := b.node.engineAuction.RunAuction(headNum); aErr != nil {
				slog.Debug("engine auction run", "slot", headNum, "err", aErr)
			} else {
				slog.Debug("engine auction result", "slot", headNum, "bids", result.TotalBids)
			}
		}
	}

	// Trie pruner (I+ EIP-7864).
	if b.node.triePruner != nil && bc.Config().IsIPlus(headBlock.Time()) {
		if work.finalBlock != nil {
			pruned := b.node.triePruner.Prune(128)
			if len(pruned) > 0 {
				slog.Debug("trie pruner: pruned stale roots", "count", len(pruned))
			}
		}
	}

	// State healer (trie gap detection).
	if b.node.stateHealer != nil {
		if n, err := b.node.stateHealer.DetectGaps(); err == nil && n > 0 {
			slog.Debug("state healer: trie gaps detected", "count", n)
		}
	}

	// Snap-sync pivot.
	if b.node.stateSyncSched != nil && work.finalBlock != nil {
		b.node.stateSyncSched.SetPivot(work.finalBlock.Header())
	}
}