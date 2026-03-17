package node

import (
	"encoding/json"
	"log/slog"
	"math/big"

	"github.com/eth2030/eth2030/core/block"
	coregas "github.com/eth2030/eth2030/core/gas"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/engine"
	"github.com/eth2030/eth2030/engine/forkchoice"
	epbsmevburn "github.com/eth2030/eth2030/epbs/mevburn"
	"github.com/eth2030/eth2030/rollup"
	rollupproof "github.com/eth2030/eth2030/rollup/proof"
)

// execBlockInternal reconstructs the block from an Engine API payload and inserts it.
// Called only by the processor goroutine (processLoop).
func (b *engineBackend) execBlockInternal(
	payload *engine.ExecutionPayloadV3,
	parentBeaconBlockRoot types.Hash,
	requestsHash *types.Hash,
) (engine.PayloadStatusV1, error) {
	bc := b.node.blockchain

	slog.Debug("engine_newPayload",
		"blockNumber", payload.BlockNumber,
		"blockHash", payload.BlockHash,
		"parentHash", payload.ParentHash,
		"timestamp", payload.Timestamp,
		"txCount", len(payload.Transactions),
	)

	// Short-circuit for already-known blocks.
	if bc.HasBlock(payload.BlockHash) {
		h := payload.BlockHash
		slog.Debug("engine_newPayload: already known, returning VALID",
			"blockNumber", payload.BlockNumber,
			"blockHash", payload.BlockHash,
		)
		b.registerBlockInFCState(payload)
		b.node.txPool.Reset(bc.State())
		return engine.PayloadStatusV1{Status: engine.StatusValid, LatestValidHash: &h}, nil
	}

	// Decode transactions from raw bytes.
	var txs []*types.Transaction
	for _, raw := range payload.Transactions {
		tx, err := types.DecodeTxRLP(raw)
		if err != nil {
			latestValid := payload.ParentHash
			return engine.PayloadStatusV1{
				Status:          engine.StatusInvalid,
				LatestValidHash: &latestValid,
			}, nil
		}
		txs = append(txs, tx)
	}

	// Decode withdrawals.
	var withdrawals []*types.Withdrawal
	if payload.Withdrawals != nil {
		withdrawals = make([]*types.Withdrawal, 0, len(payload.Withdrawals))
		for _, w := range payload.Withdrawals {
			withdrawals = append(withdrawals, &types.Withdrawal{
				Index:          w.Index,
				ValidatorIndex: w.ValidatorIndex,
				Address:        w.Address,
				Amount:         w.Amount,
			})
		}
	}

	// Reconstruct the header.
	blobGasUsed := payload.BlobGasUsed
	excessBlobGas := payload.ExcessBlobGas
	header := &types.Header{
		ParentHash:    payload.ParentHash,
		UncleHash:     types.EmptyUncleHash,
		Coinbase:      payload.FeeRecipient,
		Root:          payload.StateRoot,
		ReceiptHash:   payload.ReceiptsRoot,
		Bloom:         payload.LogsBloom,
		Difficulty:    new(big.Int),
		Number:        new(big.Int).SetUint64(payload.BlockNumber),
		GasLimit:      payload.GasLimit,
		GasUsed:       payload.GasUsed,
		Time:          payload.Timestamp,
		Extra:         payload.ExtraData,
		BaseFee:       payload.BaseFeePerGas,
		MixDigest:     payload.PrevRandao,
		TxHash:        block.DeriveTxsRoot(txs),
		BlobGasUsed:   &blobGasUsed,
		ExcessBlobGas: &excessBlobGas,
	}

	// EIP-4788: set ParentBeaconRoot when provided (Cancun+).
	if parentBeaconBlockRoot != (types.Hash{}) {
		header.ParentBeaconRoot = &parentBeaconBlockRoot
	}

	// EIP-4895: compute WithdrawalsHash.
	if payload.Withdrawals != nil {
		ws := withdrawals
		if ws == nil {
			ws = []*types.Withdrawal{}
		}
		wHash := block.DeriveWithdrawalsRoot(ws)
		header.WithdrawalsHash = &wHash
	}

	// EIP-7685: set RequestsHash for Prague blocks.
	if requestsHash != nil {
		header.RequestsHash = requestsHash
	}

	// EIP-7706: reconstruct CalldataExcessGas/CalldataGasUsed.
	if types.EIP7706HashFields && bc.Config().IsGlamsterdan(payload.Timestamp) {
		parentBlock := bc.GetBlock(payload.ParentHash)
		if parentBlock == nil {
			slog.Warn("engine_newPayload: parent block unavailable for EIP-7706, returning SYNCING",
				"blockNumber", payload.BlockNumber,
				"parentHash", payload.ParentHash,
			)
			return engine.PayloadStatusV1{Status: engine.StatusSyncing}, nil
		}
		ph := parentBlock.Header()
		var pCalldataExcess, pCalldataUsed uint64
		if ph.CalldataExcessGas != nil {
			pCalldataExcess = *ph.CalldataExcessGas
		}
		if ph.CalldataGasUsed != nil {
			pCalldataUsed = *ph.CalldataGasUsed
		}
		calldataExcessGas := coregas.CalcCalldataExcessGas(pCalldataExcess, pCalldataUsed, ph.GasLimit)
		header.CalldataExcessGas = &calldataExcessGas
		var calldataGasUsed uint64
		for _, tx := range txs {
			calldataGasUsed += tx.CalldataGas()
		}
		header.CalldataGasUsed = &calldataGasUsed
	}

	block := types.NewBlock(header, &types.Body{Transactions: txs, Withdrawals: withdrawals})

	// Check if parent is known.
	slog.Debug("engine_newPayload: step2 checking parent",
		"blockNumber", payload.BlockNumber,
		"parentHash", payload.ParentHash,
	)
	if !bc.HasBlock(payload.ParentHash) {
		slog.Debug("engine_newPayload: parent unknown, returning SYNCING",
			"parentHash", payload.ParentHash,
		)
		return engine.PayloadStatusV1{Status: engine.StatusSyncing}, nil
	}

	// Insert the block.
	slog.Debug("engine_newPayload: step3 calling InsertBlock",
		"blockNumber", payload.BlockNumber,
		"blockHash", payload.BlockHash,
	)
	if err := bc.InsertBlock(block); err != nil {
		slog.Warn("engine_newPayload: insert failed", "err", err)
		latestValid := payload.ParentHash
		return engine.PayloadStatusV1{
			Status:          engine.StatusInvalid,
			LatestValidHash: &latestValid,
		}, nil
	}

	// Cache blobs for engine_getBlobsV2.
	b.cacheBlobsFromBlock(block)

	// Update snapshot tree.
	if b.node.snapshotTree != nil {
		if statedb, err := bc.StateAtRoot(payload.StateRoot); err == nil {
			if mdb, ok := statedb.(*state.MemoryStateDB); ok {
				accounts, storage := mdb.SnapshotDiff()
				var parentStateRoot types.Hash
				if parentBlock := bc.GetBlock(payload.ParentHash); parentBlock != nil {
					parentStateRoot = parentBlock.Header().Root
				}
				if uerr := b.node.snapshotTree.Update(payload.StateRoot, parentStateRoot, accounts, storage); uerr == nil {
					if depth := b.node.config.SnapshotCapDepth; depth > 0 {
						b.node.snapshotTree.Cap(payload.StateRoot, depth)
					}
				}
			}
		}
	}

	// Track prunable state roots (I+ EIP-7864).
	if b.node.triePruner != nil && bc.Config().IsIPlus(payload.Timestamp) {
		b.node.triePruner.AddRoot(payload.BlockNumber, payload.StateRoot)
	}

	// MPT→BinaryTrie migration step (I+ EIP-7864).
	if b.node.trieMigrator != nil && bc.Config().IsIPlus(payload.Timestamp) {
		every := uint64(b.node.config.MigrateEveryBlocks)
		if every > 0 && payload.BlockNumber > 0 && payload.BlockNumber%every == 0 {
			if count, complete := b.node.trieMigrator.MigrateBatch(); !complete {
				slog.Debug("trie migration step", "keys", count, "block", payload.BlockNumber)
			}
		}
	}

	// Insert state root into binary announce trie (I+ EIP-7864).
	if bc.Config().IsIPlus(payload.Timestamp) {
		key := payload.BlockHash[:]
		val := payload.StateRoot[:]
		if b.node.trieAnnouncer != nil {
			if err := b.node.trieAnnouncer.Insert(key, val); err != nil {
				slog.Debug("trie announcer insert", "block", payload.BlockNumber, "err", err)
			}
		}
		if b.node.stackTrie != nil {
			if err := b.node.stackTrie.Put(payload.StateRoot, val); err != nil {
				slog.Debug("stack trie put", "block", payload.BlockNumber, "err", err)
			}
		}
	}

	// Request blob data (EIP-7594).
	if b.node.blobSyncMgr != nil && payload.BlobGasUsed > 0 {
		const gasPerBlob = uint64(131072)
		blobCount := (payload.BlobGasUsed + gasPerBlob - 1) / gasPerBlob
		indices := make([]uint64, blobCount)
		for i := range indices {
			indices[i] = uint64(i)
		}
		if err := b.node.blobSyncMgr.RequestBlobs(payload.BlockNumber, indices); err != nil {
			slog.Debug("blob sync request", "block", payload.BlockNumber, "blobs", blobCount, "err", err)
		}
	}

	// Record MEV burn for ePBS (Amsterdam EIP-7732).
	if b.node.epbsMEVBurn != nil && bc.Config().IsAmsterdam(payload.Timestamp) {
		epoch := payload.BlockNumber / 32
		b.node.epbsMEVBurn.RecordBurn(epoch, epbsmevburn.MEVBurnResult{})
	}

	// Confirm L1→L2 bridge deposits (EIP-8079).
	if b.node.rollupBridge != nil {
		if confirmed := b.node.rollupBridge.ConfirmDeposits(payload.BlockNumber); confirmed > 0 {
			slog.Debug("rollup bridge: deposits confirmed", "block", payload.BlockNumber, "count", confirmed)
		}
	}

	// Update portal content-radius.
	if b.node.portalRouter != nil {
		b.node.portalRouter.UpdateRadius(payload.BlockNumber, 1<<32)
	}

	// Advance native rollup anchor state (Amsterdam EIP-8079).
	if b.node.rollupAnchor != nil && bc.Config().IsAmsterdam(payload.Timestamp) {
		execOut := &rollup.ExecuteOutput{
			PostStateRoot: payload.StateRoot,
			ReceiptsRoot:  payload.ReceiptsRoot,
			GasUsed:       payload.GasUsed,
			Success:       true,
		}
		if err := b.node.rollupAnchor.UpdateAfterExecute(execOut, payload.BlockNumber, payload.Timestamp); err != nil {
			slog.Debug("rollup anchor update", "block", payload.BlockNumber, "err", err)
		}
	}

	// Generate cross-layer deposit proof (EIP-8079).
	if b.node.rollupProof != nil && bc.Config().IsAmsterdam(payload.Timestamp) {
		msg := &rollupproof.CrossLayerMessage{
			Source:      rollupproof.LayerL1,
			Destination: rollupproof.LayerL2,
			Nonce:       payload.BlockNumber,
			Sender:      payload.FeeRecipient,
			Target:      payload.FeeRecipient,
			Value:       new(big.Int),
		}
		if _, err := b.node.rollupProof.GenerateDepositProof(msg, payload.StateRoot); err != nil {
			slog.Debug("rollup proof generate", "block", payload.BlockNumber, "err", err)
		}
	}

	// Sync txpool state.
	b.node.txPool.Reset(bc.State())

	// Notify gas oracle.
	if b.node.gasOracle != nil {
		tips := extractBlockTips(txs, payload.BaseFeePerGas)
		b.node.gasOracle.RecordBlock(payload.BlockNumber, payload.BaseFeePerGas, tips)
	}

	// Feed txpool gas-price suggestor.
	b.node.txPool.RecordBlock(header, txs)

	// Record block gas for gigagas tracking.
	if b.node.gasRateTracker != nil {
		b.node.gasRateTracker.RecordBlockGas(payload.BlockNumber, payload.GasUsed, payload.Timestamp)
	}

	// Advance encrypted mempool epoch.
	if b.node.encryptedProtocol != nil {
		b.node.encryptedProtocol.SetEpoch(payload.BlockNumber)
		b.node.encryptedProtocol.ExpireOldCommits(payload.BlockNumber)
	}
	if b.node.encryptedPool != nil {
		b.node.encryptedPool.ExpireCommits(payload.Timestamp)
	}

	// Reset txpool trackers.
	newState := bc.State()
	if b.node.acctTracker != nil {
		b.node.acctTracker.ResetOnReorg(newState)
	}
	if b.node.nonceTracker != nil {
		b.node.nonceTracker.Reset(newState)
	}

	// Chunk payload for streaming.
	if b.node.payloadChunker != nil {
		if encoded, encErr := json.Marshal(payload); encErr == nil {
			if chunks, chunkErr := b.node.payloadChunker.ChunkPayload(encoded); chunkErr == nil {
				slog.Debug("payload chunked", "block", payload.BlockNumber, "chunks", len(chunks))
			}
		}
	}

	// Register in forkchoice state manager.
	if b.node.fcStateManager != nil {
		bi := &forkchoice.BlockInfo{
			Hash:       payload.BlockHash,
			ParentHash: payload.ParentHash,
			Number:     payload.BlockNumber,
			Slot:       payload.BlockNumber,
		}
		b.node.fcStateManager.AddBlock(bi)
		b.node.fcTracker.Reorgs.AddBlock(bi)
	}

	// Announce nonce (EIP-8077).
	if b.node.nonceAnnouncer != nil {
		if err := b.node.nonceAnnouncer.AnnounceNonce("local", payload.BlockHash, payload.BlockNumber); err != nil {
			slog.Debug("nonce announce", "block", payload.BlockNumber, "err", err)
		}
	}

	blockHash := block.Hash()
	slog.Info("engine_newPayload: accepted",
		"blockNumber", payload.BlockNumber,
		"blockHash", blockHash,
	)
	return engine.PayloadStatusV1{
		Status:          engine.StatusValid,
		LatestValidHash: &blockHash,
	}, nil
}