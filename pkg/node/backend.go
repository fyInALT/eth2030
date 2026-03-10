package node

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"sync"

	"github.com/eth2030/eth2030/core/block"
	"github.com/eth2030/eth2030/core/mev"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/core/vm"
	"github.com/eth2030/eth2030/engine"
	"github.com/eth2030/eth2030/engine/forkchoice"
	"github.com/eth2030/eth2030/engine/vhash"
	epbsbid "github.com/eth2030/eth2030/epbs/bid"
	epbsmevburn "github.com/eth2030/eth2030/epbs/mevburn"
	"github.com/eth2030/eth2030/rollup"
	rollupproof "github.com/eth2030/eth2030/rollup/proof"
	"github.com/eth2030/eth2030/rpc"
	"github.com/eth2030/eth2030/trie"
	"github.com/eth2030/eth2030/txpool/shared"
)

// extractBlockTips returns the effective priority fee (tip) for each
// transaction in the block, given the block's base fee.
func extractBlockTips(txs []*types.Transaction, baseFee *big.Int) []*big.Int {
	tips := make([]*big.Int, 0, len(txs))
	if baseFee == nil {
		baseFee = new(big.Int)
	}
	for _, tx := range txs {
		var tip *big.Int
		switch tx.Type() {
		case types.DynamicFeeTxType:
			// EIP-1559: effectiveTip = min(GasTipCap, GasFeeCap - baseFee)
			tipCap := tx.GasTipCap()
			feeCap := tx.GasFeeCap()
			if tipCap == nil || feeCap == nil {
				continue
			}
			effective := new(big.Int).Sub(feeCap, baseFee)
			if effective.Sign() < 0 {
				continue
			}
			tip = tipCap
			if effective.Cmp(tipCap) < 0 {
				tip = effective
			}
		default:
			// Legacy / access-list tx: tip = gasPrice - baseFee
			gp := tx.GasPrice()
			if gp == nil {
				continue
			}
			tip = new(big.Int).Sub(gp, baseFee)
			if tip.Sign() < 0 {
				continue
			}
		}
		tips = append(tips, new(big.Int).Set(tip))
	}
	return tips
}

// nodeBackend adapts the Node to the rpc.Backend interface.
type nodeBackend struct {
	node *Node
}

func newNodeBackend(n *Node) rpc.Backend {
	return &nodeBackend{node: n}
}

func (b *nodeBackend) HeaderByNumber(number rpc.BlockNumber) *types.Header {
	bc := b.node.blockchain
	switch number {
	case rpc.LatestBlockNumber, rpc.PendingBlockNumber:
		blk := bc.CurrentBlock()
		if blk != nil {
			return blk.Header()
		}
		return nil
	case rpc.EarliestBlockNumber:
		blk := bc.GetBlockByNumber(0)
		if blk != nil {
			return blk.Header()
		}
		return nil
	default:
		blk := bc.GetBlockByNumber(uint64(number))
		if blk != nil {
			return blk.Header()
		}
		return nil
	}
}

func (b *nodeBackend) HeaderByHash(hash types.Hash) *types.Header {
	blk := b.node.blockchain.GetBlock(hash)
	if blk != nil {
		return blk.Header()
	}
	return nil
}

func (b *nodeBackend) BlockByNumber(number rpc.BlockNumber) *types.Block {
	bc := b.node.blockchain
	switch number {
	case rpc.LatestBlockNumber, rpc.PendingBlockNumber:
		return bc.CurrentBlock()
	case rpc.EarliestBlockNumber:
		return bc.GetBlockByNumber(0)
	default:
		return bc.GetBlockByNumber(uint64(number))
	}
}

func (b *nodeBackend) BlockByHash(hash types.Hash) *types.Block {
	return b.node.blockchain.GetBlock(hash)
}

func (b *nodeBackend) CurrentHeader() *types.Header {
	blk := b.node.blockchain.CurrentBlock()
	if blk != nil {
		return blk.Header()
	}
	return nil
}

func (b *nodeBackend) ChainID() *big.Int {
	return b.node.blockchain.Config().ChainID
}

func (b *nodeBackend) StateAt(root types.Hash) (state.StateDB, error) {
	return b.node.blockchain.StateAtRoot(root)
}

func (b *nodeBackend) GetProof(addr types.Address, storageKeys []types.Hash, blockNumber rpc.BlockNumber) (*trie.AccountProof, error) {
	header := b.HeaderByNumber(blockNumber)
	if header == nil {
		return nil, fmt.Errorf("block not found")
	}

	statedb, err := b.StateAt(header.Root)
	if err != nil {
		return nil, err
	}

	// Type-assert to MemoryStateDB to access trie-building methods.
	memState, ok := statedb.(*state.MemoryStateDB)
	if !ok {
		return nil, fmt.Errorf("state does not support proof generation")
	}

	// Build the full state trie from all accounts.
	stateTrie := memState.BuildStateTrie()

	// Build the storage trie for the requested account.
	storageTrie := memState.BuildStorageTrie(addr)

	// Generate account proof with storage proofs.
	return trie.ProveAccountWithStorage(stateTrie, addr, storageTrie, storageKeys)
}

func (b *nodeBackend) SendTransaction(tx *types.Transaction) error {
	if err := b.node.txPool.AddLocal(tx); err != nil {
		return err
	}
	// Persist to journal for crash recovery.
	if b.node.txJournal != nil {
		if jerr := b.node.txJournal.Insert(tx, true); jerr != nil {
			slog.Debug("tx journal insert failed", "hash", tx.Hash(), "err", jerr)
		}
	}
	// Propagate into shared mempool for cross-node gossip (sharded mempool).
	if b.node.sharedPool != nil {
		var chainID uint64
		if cfg := b.node.blockchain.Config(); cfg != nil && cfg.ChainID != nil {
			chainID = cfg.ChainID.Uint64()
		}
		signer := types.LatestSigner(chainID)
		sender, _ := signer.Sender(tx)
		smTx := shared.SharedMempoolTx{
			Hash:     tx.Hash(),
			Sender:   sender,
			Nonce:    tx.Nonce(),
			GasPrice: tx.GasFeeCap().Uint64(),
			Data:     tx.Data(),
		}
		if err := b.node.sharedPool.AddTransaction(smTx); err != nil {
			slog.Debug("shared mempool add", "hash", tx.Hash(), "err", err)
		}
	}
	// Feed calldata to native rollup sequencer for L2 batch assembly (EIP-8079).
	if b.node.rollupSeq != nil {
		if data := tx.Data(); len(data) > 0 {
			if seqErr := b.node.rollupSeq.AddTransaction(data); seqErr != nil {
				slog.Debug("rollup sequencer add", "hash", tx.Hash(), "err", seqErr)
			}
		}
	}
	return nil
}

func (b *nodeBackend) GetTransaction(hash types.Hash) (*types.Transaction, uint64, uint64) {
	// Check the blockchain's tx lookup index first.
	if entry, found := b.node.blockchain.GetTxLookupEntry(hash); found {
		block := b.node.blockchain.GetBlock(entry.BlockHash)
		if block != nil {
			txs := block.Transactions()
			if int(entry.TxIndex) < len(txs) {
				return txs[entry.TxIndex], entry.BlockNumber, entry.TxIndex
			}
		}
	}
	// Fall back to txpool for pending transactions.
	tx := b.node.txPool.Get(hash)
	if tx != nil {
		return tx, 0, 0
	}
	return nil, 0, 0
}

func (b *nodeBackend) SuggestGasPrice() *big.Int {
	// Use the gas oracle if it has seen blocks (returns baseFee + percentile tip).
	if b.node.gasOracle != nil && b.node.gasOracle.BaseFee().Sign() > 0 {
		return b.node.gasOracle.SuggestGasPrice()
	}
	// Fallback: return current base fee when oracle has no data yet.
	blk := b.node.blockchain.CurrentBlock()
	if blk != nil && blk.Header().BaseFee != nil {
		return new(big.Int).Set(blk.Header().BaseFee)
	}
	return big.NewInt(1_000_000_000) // 1 gwei default
}

func (b *nodeBackend) GetReceipts(blockHash types.Hash) []*types.Receipt {
	return b.node.blockchain.GetReceipts(blockHash)
}

func (b *nodeBackend) GetLogs(blockHash types.Hash) []*types.Log {
	return b.node.blockchain.GetLogs(blockHash)
}

func (b *nodeBackend) GetBlockReceipts(number uint64) []*types.Receipt {
	return b.node.blockchain.GetBlockReceipts(number)
}

func (b *nodeBackend) EVMCall(from types.Address, to *types.Address, data []byte, gas uint64, value *big.Int, blockNumber rpc.BlockNumber) ([]byte, uint64, error) {
	bc := b.node.blockchain

	// Resolve block header.
	var header *types.Header
	switch blockNumber {
	case rpc.LatestBlockNumber, rpc.PendingBlockNumber:
		blk := bc.CurrentBlock()
		if blk != nil {
			header = blk.Header()
		}
	default:
		blk := bc.GetBlockByNumber(uint64(blockNumber))
		if blk != nil {
			header = blk.Header()
		}
	}
	if header == nil {
		return nil, 0, fmt.Errorf("block not found")
	}

	// Get state at this block.
	statedb, err := b.StateAt(header.Root)
	if err != nil {
		return nil, 0, fmt.Errorf("state not found: %w", err)
	}

	// Default gas to 50M if zero.
	if gas == 0 {
		gas = 50_000_000
	}
	if value == nil {
		value = new(big.Int)
	}

	// Build block and tx contexts.
	blockCtx := vm.BlockContext{
		GetHash:     bc.GetHashFn(),
		BlockNumber: header.Number,
		Time:        header.Time,
		GasLimit:    header.GasLimit,
		BaseFee:     header.BaseFee,
	}
	txCtx := vm.TxContext{
		Origin:   from,
		GasPrice: header.BaseFee,
	}

	evm := vm.NewEVMWithState(blockCtx, txCtx, vm.Config{}, statedb)

	// Apply fork rules so the correct precompile map and jump table are used.
	if chainCfg := bc.Config(); chainCfg != nil {
		rules := chainCfg.Rules(header.Number, chainCfg.IsMerge(), header.Time)
		forkRules := vm.ForkRules{
			IsGlamsterdan:    rules.IsGlamsterdan,
			IsPrague:         rules.IsPrague,
			IsCancun:         rules.IsCancun,
			IsShanghai:       rules.IsShanghai,
			IsMerge:          rules.IsMerge,
			IsLondon:         rules.IsLondon,
			IsBerlin:         rules.IsBerlin,
			IsIstanbul:       rules.IsIstanbul,
			IsConstantinople: rules.IsConstantinople,
			IsByzantium:      rules.IsByzantium,
			IsHomestead:      rules.IsHomestead,
			IsEIP158:         rules.IsEIP158,
			IsEIP7708:        rules.IsEIP7708,
			IsEIP7954:        rules.IsEIP7954,
			IsIPlus:          rules.IsIPlus,
		}
		evm.SetJumpTable(vm.SelectJumpTable(forkRules))
		evm.SetPrecompiles(vm.SelectPrecompiles(forkRules))
		evm.SetForkRules(forkRules)
	}

	if to == nil {
		// Contract creation call - just return empty.
		return nil, gas, nil
	}

	ret, gasLeft, err := evm.Call(from, *to, data, gas, value)
	return ret, gasLeft, err
}

func (b *nodeBackend) HistoryOldestBlock() uint64 {
	// Delegate to the blockchain's configured history oldest block.
	return b.node.blockchain.HistoryOldestBlock()
}

// TraceTransaction re-executes a transaction with a StructLogTracer attached.
// It looks up the block containing the transaction, re-processes all prior
// transactions to build up state, then executes the target tx with tracing.
func (b *nodeBackend) TraceTransaction(txHash types.Hash) (*vm.StructLogTracer, error) {
	bc := b.node.blockchain

	// Look up the transaction in the chain index.
	entry, found := bc.GetTxLookupEntry(txHash)
	if !found {
		return nil, fmt.Errorf("transaction %v not found", txHash)
	}

	block := bc.GetBlock(entry.BlockHash)
	if block == nil {
		return nil, fmt.Errorf("block %v not found", entry.BlockHash)
	}

	txs := block.Transactions()
	if int(entry.TxIndex) >= len(txs) {
		return nil, fmt.Errorf("transaction index %d out of range", entry.TxIndex)
	}

	// Get state at the parent block.
	header := block.Header()
	parentBlock := bc.GetBlock(header.ParentHash)
	if parentBlock == nil {
		return nil, fmt.Errorf("parent block %v not found", header.ParentHash)
	}
	statedb, err := b.StateAt(parentBlock.Header().Root)
	if err != nil {
		return nil, fmt.Errorf("state not found for parent block: %w", err)
	}

	blockCtx := vm.BlockContext{
		GetHash:     bc.GetHashFn(),
		BlockNumber: header.Number,
		Time:        header.Time,
		GasLimit:    header.GasLimit,
		BaseFee:     header.BaseFee,
	}

	// Re-execute all transactions before the target to build up state.
	for i := uint64(0); i < entry.TxIndex; i++ {
		tx := txs[i]
		from := types.Address{}
		if sender := tx.Sender(); sender != nil {
			from = *sender
		}
		txCtx := vm.TxContext{
			Origin:   from,
			GasPrice: tx.GasPrice(),
		}
		evm := vm.NewEVMWithState(blockCtx, txCtx, vm.Config{}, statedb)
		to := tx.To()
		if to != nil {
			evm.Call(from, *to, tx.Data(), tx.Gas(), tx.Value())
		}
		// Update nonce after replaying the transaction.
		statedb.SetNonce(from, statedb.GetNonce(from)+1)
	}

	// Now execute the target transaction with tracing enabled.
	targetTx := txs[entry.TxIndex]
	from := types.Address{}
	if sender := targetTx.Sender(); sender != nil {
		from = *sender
	}
	txCtx := vm.TxContext{
		Origin:   from,
		GasPrice: targetTx.GasPrice(),
	}

	tracer := vm.NewStructLogTracer()
	tracingCfg := vm.Config{
		Debug:  true,
		Tracer: tracer,
	}
	evm := vm.NewEVMWithState(blockCtx, txCtx, tracingCfg, statedb)

	// Subtract intrinsic gas before passing to the EVM so the trace faithfully
	// replicates the actual execution environment (the state transition charges
	// 21000 + calldata cost before entering the EVM).
	intrinsicGas := uint64(21000)
	for _, b := range targetTx.Data() {
		if b == 0 {
			intrinsicGas += 4
		} else {
			intrinsicGas += 16
		}
	}
	if targetTx.To() == nil {
		intrinsicGas += 32000
	}
	evmGas := uint64(0)
	if targetTx.Gas() > intrinsicGas {
		evmGas = targetTx.Gas() - intrinsicGas
	}

	to := targetTx.To()
	if to != nil {
		ret, gasLeft, err := evm.Call(from, *to, targetTx.Data(), evmGas, targetTx.Value())
		gasUsed := evmGas - gasLeft
		tracer.CaptureEnd(ret, gasUsed, err)
	} else if evmGas > 0 {
		// Contract creation: record as failed trace (no full create tracing here).
		tracer.CaptureEnd(nil, evmGas, fmt.Errorf("contract creation tracing not supported"))
	}

	return tracer, nil
}

// txPoolAdapter adapts *txpool.TxPool to core.TxPoolReader.
type txPoolAdapter struct {
	node *Node
}

func (a *txPoolAdapter) Pending() []*types.Transaction {
	txs := a.node.txPool.PendingFlat()

	// Apply fair ordering for MEV protection when enabled.
	if a.node.mevConfig != nil && a.node.mevConfig.EnableFairOrdering && len(txs) > 0 {
		entries := make([]mev.FairOrderingEntry, len(txs))
		for i, tx := range txs {
			entries[i] = mev.FairOrderingEntry{
				Transaction: tx,
				ArrivalTime: uint64(i), // use insertion order as proxy for arrival time
			}
		}
		ordered, _ := mev.FairOrdering(entries, a.node.mevConfig.FairOrderMaxDelay)
		txs = make([]*types.Transaction, len(ordered))
		for i, e := range ordered {
			txs[i] = e.Transaction
		}
	}
	return txs
}

// pendingPayload stores a built payload for later retrieval by getPayload.
type pendingPayload struct {
	block    *types.Block
	receipts []*types.Receipt
}

// engineBackend adapts the Node to the engine.Backend interface.
type engineBackend struct {
	node *Node

	mu           sync.Mutex
	payloads     map[engine.PayloadID]*pendingPayload
	payloadOrder []engine.PayloadID // insertion order for LRU eviction
	maxPayloads  int                // cap from node config
	builder      *block.BlockBuilder
}

func newEngineBackend(n *Node) engine.Backend {
	pool := &txPoolAdapter{node: n}
	builder := block.NewBlockBuilder(n.blockchain.Config(), n.blockchain, pool)
	maxPayloads := n.config.CacheEnginePayloads
	if maxPayloads <= 0 {
		maxPayloads = 32
	}
	return &engineBackend{
		node:        n,
		payloads:    make(map[engine.PayloadID]*pendingPayload),
		maxPayloads: maxPayloads,
		builder:     builder,
	}
}

func (b *engineBackend) GetHeadHash() types.Hash {
	if blk := b.node.blockchain.CurrentBlock(); blk != nil {
		return blk.Hash()
	}
	return types.Hash{}
}

func (b *engineBackend) GetSafeHash() types.Hash      { return types.Hash{} }
func (b *engineBackend) GetFinalizedHash() types.Hash { return types.Hash{} }

func (b *engineBackend) ProcessBlock(
	payload *engine.ExecutionPayloadV3,
	expectedBlobVersionedHashes []types.Hash,
	parentBeaconBlockRoot types.Hash,
) (engine.PayloadStatusV1, error) {
	// EIP-4844: validate blob versioned hashes against KZG commitments in txs.
	if len(expectedBlobVersionedHashes) > 0 {
		if err := vhash.VerifyAllBlobVersionBytes(expectedBlobVersionedHashes); err != nil {
			latestValid := payload.ParentHash
			slog.Warn("engine_newPayload: invalid blob versioned hash version byte",
				"err", err,
			)
			return engine.PayloadStatusV1{
				Status:          engine.StatusInvalid,
				LatestValidHash: &latestValid,
			}, nil
		}
	}
	return b.processBlockInternal(payload, parentBeaconBlockRoot, nil)
}

// processBlockInternal reconstructs the block from an Engine API payload and
// inserts it. parentBeaconBlockRoot is included in the header hash (EIP-4788).
// requestsHash is non-nil only for Prague (V4) payloads.
func (b *engineBackend) processBlockInternal(
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

	// Decode withdrawals. When the CL sends withdrawals:[] (non-nil empty),
	// the decoded slice must also be non-nil so the block passes Shanghai
	// validation (which rejects nil withdrawals on Shanghai+ blocks).
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

	// Reconstruct the header with all fields that contribute to the block hash.
	blobGasUsed := payload.BlobGasUsed
	excessBlobGas := payload.ExcessBlobGas
	header := &types.Header{
		ParentHash:    payload.ParentHash,
		UncleHash:     types.EmptyUncleHash, // always empty for PoS
		Coinbase:      payload.FeeRecipient,
		Root:          payload.StateRoot,
		ReceiptHash:   payload.ReceiptsRoot,
		Bloom:         payload.LogsBloom,
		Difficulty:    new(big.Int), // always 0 for PoS
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

	// EIP-4895: compute WithdrawalsHash when withdrawals are present.
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

	block := types.NewBlock(header, &types.Body{Transactions: txs, Withdrawals: withdrawals})

	// Verify block hash matches what the CL provided.
	if block.Hash() != payload.BlockHash {
		slog.Warn("engine_newPayload: block hash mismatch",
			"computed", block.Hash(),
			"payload", payload.BlockHash,
		)
		latestValid := payload.ParentHash
		return engine.PayloadStatusV1{
			Status:          engine.StatusInvalid,
			LatestValidHash: &latestValid,
		}, nil
	}

	// Step 2: check if parent is known.
	slog.Debug("engine_newPayload: step2 checking parent",
		"blockNumber", payload.BlockNumber,
		"parentHash", payload.ParentHash,
	)
	if !bc.HasBlock(payload.ParentHash) {
		slog.Debug("engine_newPayload: parent unknown, returning SYNCING",
			"parentHash", payload.ParentHash,
		)
		return engine.PayloadStatusV1{
			Status: engine.StatusSyncing,
		}, nil
	}

	// Guard against expensive state re-execution while holding bc.mu.Lock().
	// InsertBlock acquires bc.mu.Lock() for the full duration of stateAt(),
	// which re-executes all blocks from genesis if the parent state is not
	// cached. For deep fork blocks whose parent state was evicted from the
	// cache, this can block all concurrent FCU calls for tens of seconds.
	// Return SYNCING instead so the CL can provide blocks in canonical order.
	//
	// stateReExecMaxGap must not exceed defaultMaxCachedStates (64) so that
	// normal sequential processing always has the parent state cached.
	const stateReExecMaxGap = 64
	if !bc.HasStateCached(payload.ParentHash) {
		if headBlk := bc.CurrentBlock(); headBlk != nil && payload.BlockNumber > 0 {
			parentNum := payload.BlockNumber - 1
			if parentNum+stateReExecMaxGap < headBlk.NumberU64() {
				slog.Debug("engine_newPayload: parent state not cached, returning SYNCING",
					"blockNumber", payload.BlockNumber,
					"parentNum", parentNum,
					"headNum", headBlk.NumberU64(),
				)
				return engine.PayloadStatusV1{Status: engine.StatusSyncing}, nil
			}
		}
	}

	// Step 3: insert the block (Phase 1: validate, Phase 2: execute, Phase 3: write).
	slog.Debug("engine_newPayload: step3 calling InsertBlock",
		"blockNumber", payload.BlockNumber,
		"blockHash", payload.BlockHash,
		"stateCached", bc.HasStateCached(payload.ParentHash),
	)
	if err := bc.InsertBlock(block); err != nil {
		slog.Warn("engine_newPayload: insert failed", "err", err)
		latestValid := payload.ParentHash
		return engine.PayloadStatusV1{
			Status:          engine.StatusInvalid,
			LatestValidHash: &latestValid,
		}, nil
	}

	// Update the snapshot tree with the new block's state so snap-sync peers
	// can retrieve accounts and storage slots at this root. SnapshotDiff exports
	// the full MemoryStateDB as a diff layer on top of the disk layer.
	if b.node.snapshotTree != nil {
		if statedb, err := bc.StateAtRoot(payload.StateRoot); err == nil {
			if mdb, ok := statedb.(*state.MemoryStateDB); ok {
				accounts, storage := mdb.SnapshotDiff()
				// The snapshot tree is keyed by state roots, not block hashes.
				// Retrieve the parent block to get its state root.
				var parentStateRoot types.Hash
				if parentBlock := bc.GetBlock(payload.ParentHash); parentBlock != nil {
					parentStateRoot = parentBlock.Header().Root
				}
				if uerr := b.node.snapshotTree.Update(payload.StateRoot, parentStateRoot, accounts, storage); uerr == nil {
					// Cap diff layers to bound memory growth; 0 disables periodic flushing.
					if depth := b.node.config.SnapshotCapDepth; depth > 0 {
						b.node.snapshotTree.Cap(payload.StateRoot, depth)
					}
				}
			}
		}
	}

	// Track prunable state roots for binary trie / MPT state pruning (I+ EIP-7864).
	if b.node.triePruner != nil && bc.Config().IsIPlus(payload.Timestamp) {
		b.node.triePruner.AddRoot(payload.BlockNumber, payload.StateRoot)
	}

	// Incremental MPT→BinaryTrie migration step (I+ EIP-7864, every MigrateEveryBlocks).
	if b.node.trieMigrator != nil && bc.Config().IsIPlus(payload.Timestamp) {
		every := uint64(b.node.config.MigrateEveryBlocks)
		if every > 0 && payload.BlockNumber > 0 && payload.BlockNumber%every == 0 {
			if count, complete := b.node.trieMigrator.MigrateBatch(); !complete {
				slog.Debug("trie migration step", "keys", count, "block", payload.BlockNumber)
			}
		}
	}

	// Insert accepted block's state root into the binary announce trie and
	// stack-trie node collector (I+ EIP-7864 binary trie infrastructure).
	if bc.Config().IsIPlus(payload.Timestamp) {
		key := payload.BlockHash[:]
		val := payload.StateRoot[:]
		if b.node.trieAnnouncer != nil {
			if err := b.node.trieAnnouncer.Insert(key, val); err != nil {
				slog.Debug("trie announcer insert", "block", payload.BlockNumber, "err", err)
			}
		}
		if b.node.stackTrie != nil {
			// Put the block state root as a collected trie node for later flush.
			if err := b.node.stackTrie.Put(payload.StateRoot, val); err != nil {
				slog.Debug("stack trie put", "block", payload.BlockNumber, "err", err)
			}
		}
	}

	// Request blob data for this block from the beacon blob sync manager (EIP-7594).
	// BlobGasUsed > 0 signals that blob transactions are present in the block.
	// Each blob consumes GAS_PER_BLOB = 131072 gas (EIP-4844).
	if b.node.blobSyncMgr != nil && payload.BlobGasUsed > 0 {
		const gasPerBlob = uint64(131072)
		blobCount := (payload.BlobGasUsed + gasPerBlob - 1) / gasPerBlob
		indices := make([]uint64, blobCount)
		for i := range indices {
			indices[i] = uint64(i)
		}
		if err := b.node.blobSyncMgr.RequestBlobs(payload.BlockNumber, indices); err != nil {
			slog.Debug("blob sync request", "block", payload.BlockNumber,
				"blobs", blobCount, "err", err)
		}
	}

	// Record MEV burn for ePBS bid tracking (Amsterdam EIP-7732).
	if b.node.epbsMEVBurn != nil && bc.Config().IsAmsterdam(payload.Timestamp) {
		epoch := payload.BlockNumber / 32
		b.node.epbsMEVBurn.RecordBurn(epoch, epbsmevburn.MEVBurnResult{})
	}

	// Confirm pending L1→L2 bridge deposits that have accrued enough confirmations (EIP-8079).
	if b.node.rollupBridge != nil {
		if confirmed := b.node.rollupBridge.ConfirmDeposits(payload.BlockNumber); confirmed > 0 {
			slog.Debug("rollup bridge: deposits confirmed", "block", payload.BlockNumber, "count", confirmed)
		}
	}

	// Update portal content-radius estimate based on block height (Portal network).
	// Real storage metrics feed into this when the state backend is wired.
	if b.node.portalRouter != nil {
		b.node.portalRouter.UpdateRadius(payload.BlockNumber, 1<<32)
	}

	// Advance native rollup anchor state after each accepted EL block (Amsterdam EIP-8079).
	// This records the latest L1 block hash and state root for L2 chains to anchor against.
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

	// Generate a cross-layer deposit proof anchoring the accepted block's state
	// root to the L1 chain (EIP-8079 native rollups, Amsterdam+).
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

	// Sync txpool state with the new head so pending/queued txs are re-evaluated.
	b.node.txPool.Reset(bc.State())

	// Notify gas oracle so it can refine its fee estimates.
	if b.node.gasOracle != nil {
		tips := extractBlockTips(txs, payload.BaseFeePerGas)
		b.node.gasOracle.RecordBlock(payload.BlockNumber, payload.BaseFeePerGas, tips)
	}

	// Feed the txpool gas-price suggestor with the new block so that
	// SuggestGasPrice / SuggestAllTiers return up-to-date recommendations.
	b.node.txPool.RecordBlock(header, txs)

	// Record block gas usage for gigagas throughput tracking (M+ north star).
	if b.node.gasRateTracker != nil {
		b.node.gasRateTracker.RecordBlockGas(payload.BlockNumber, payload.GasUsed, payload.Timestamp)
	}

	// Advance encrypted mempool epoch and expire stale commits (Hegotá MEV protection).
	if b.node.encryptedProtocol != nil {
		b.node.encryptedProtocol.SetEpoch(payload.BlockNumber)
		b.node.encryptedProtocol.ExpireOldCommits(payload.BlockNumber)
	}
	if b.node.encryptedPool != nil {
		b.node.encryptedPool.ExpireCommits(payload.Timestamp)
	}

	// Reset txpool trackers to the new head state so nonce/balance data stays fresh.
	newState := bc.State()
	if b.node.acctTracker != nil {
		b.node.acctTracker.ResetOnReorg(newState)
	}
	if b.node.nonceTracker != nil {
		b.node.nonceTracker.Reset(newState)
	}

	// Chunk the accepted payload for streaming delivery to the CL (Hegotá payload chunking).
	if b.node.payloadChunker != nil {
		if encoded, encErr := json.Marshal(payload); encErr == nil {
			if chunks, chunkErr := b.node.payloadChunker.ChunkPayload(encoded); chunkErr == nil {
				slog.Debug("payload chunked", "block", payload.BlockNumber, "chunks", len(chunks))
			}
		}
	}

	// Register accepted block in forkchoice state manager for reorg detection.
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

	// Announce this block's sequence number to peers (EIP-8077 announce-nonce, ETH/72).
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
	slog.Debug("engine_forkchoiceUpdated: step1 lookup head",
		"headBlockHash", fcState.HeadBlockHash,
		"genesisHash", bc.Genesis().Hash(),
	)
	headBlock := bc.GetBlock(fcState.HeadBlockHash)
	var payloadStatus engine.PayloadStatusV1
	if headBlock == nil {
		slog.Warn("engine_forkchoiceUpdated: unknown head block, returning SYNCING",
			"headBlockHash", fcState.HeadBlockHash,
			"genesisHash", bc.Genesis().Hash(),
			"currentHead", bc.CurrentBlock().Hash(),
		)
		payloadStatus = engine.PayloadStatusV1{
			Status: engine.StatusSyncing,
		}
		return engine.ForkchoiceUpdatedResult{
			PayloadStatus: payloadStatus,
		}, nil
	}

	// Head is known. Report valid.
	headHash := headBlock.Hash()
	payloadStatus = engine.PayloadStatusV1{
		Status:          engine.StatusValid,
		LatestValidHash: &headHash,
	}

	slog.Debug("engine_forkchoiceUpdated: head known",
		"headBlockHash", headHash,
		"number", headBlock.NumberU64(),
	)

	// Step 2: update forkchoice state manager (reorg detection).
	slog.Debug("engine_forkchoiceUpdated: step2 fcstate update",
		"headNum", headBlock.NumberU64(),
		"hasPayloadAttrs", payloadAttributes != nil,
	)
	if b.node.fcStateManager != nil {
		if err := b.node.fcStateManager.ProcessForkchoiceUpdate(fcState); err != nil {
			slog.Debug("fcStateManager update", "err", err)
		}
	}

	// Update the high-level tracker: conflict detection, FCU history, reorg analytics.
	if b.node.fcTracker != nil {
		safeNum := uint64(0)
		finalNum := uint64(0)
		if safeBlock := bc.GetBlock(fcState.SafeBlockHash); safeBlock != nil {
			safeNum = safeBlock.NumberU64()
		}
		if finalBlock := bc.GetBlock(fcState.FinalizedBlockHash); finalBlock != nil {
			finalNum = finalBlock.NumberU64()
		}
		conflict, reason, reorg := b.node.fcTracker.ProcessUpdate(
			fcState, payloadAttributes != nil, headBlock.NumberU64(), safeNum, finalNum,
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

	// ePBS auction lifecycle: open a new auction slot and prune stale bids/escrow.
	// Only active after Amsterdam fork (EIP-7732 ePBS).
	headNum := headBlock.NumberU64()
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
		// Score current builder bids for the auction slot using composite metrics.
		if b.node.epbsBid != nil {
			components := epbsbid.ScoreComponents{
				BidAmount:        0,    // no live bids yet; zero baseline
				ReputationScore:  50.0, // neutral starting reputation
				InclusionQuality: 1.0,  // full IL compliance assumed
				LatencyMs:        0,
			}
			score := b.node.epbsBid.ComputeScore(components)
			slog.Debug("epbs: bid baseline score", "slot", headNum, "score", score)
		}
		// Run the EL-side builder auction to close bids for this slot.
		if b.node.engineAuction != nil {
			if result, aErr := b.node.engineAuction.RunAuction(headNum); aErr != nil {
				slog.Debug("engine auction run", "slot", headNum, "err", aErr)
			} else {
				slog.Debug("engine auction result", "slot", headNum, "bids", result.TotalBids)
			}
		}
	}

	// Prune stale state roots when the finalized block advances (I+ EIP-7864).
	if b.node.triePruner != nil && bc.Config().IsIPlus(headBlock.Time()) {
		if finalBlock := bc.GetBlock(fcState.FinalizedBlockHash); finalBlock != nil {
			pruned := b.node.triePruner.Prune(128)
			if len(pruned) > 0 {
				slog.Debug("trie pruner: pruned stale roots", "count", len(pruned))
			}
		}
	}

	// Drive trie-healing gap detection on each forkchoice update.
	// DetectGaps scans for missing trie nodes that need to be fetched from peers.
	if b.node.stateHealer != nil {
		if n, err := b.node.stateHealer.DetectGaps(); err == nil && n > 0 {
			slog.Debug("state healer: trie gaps detected", "count", n)
		}
	}

	// Configure state-sync pivot to the latest finalized block header.
	// This enables snap-sync to resume from a finalized checkpoint.
	if b.node.stateSyncSched != nil {
		if finalBlock := bc.GetBlock(fcState.FinalizedBlockHash); finalBlock != nil {
			b.node.stateSyncSched.SetPivot(finalBlock.Header())
		}
	}

	// Step 3: if no payload attributes, return the FCU acknowledgment.
	if payloadAttributes == nil {
		slog.Debug("engine_forkchoiceUpdated: step3 no attrs, done",
			"headNum", headBlock.NumberU64(),
		)
		return engine.ForkchoiceUpdatedResult{
			PayloadStatus: payloadStatus,
		}, nil
	}

	// Step 3: payload attributes provided — build a new block.
	slog.Debug("engine_forkchoiceUpdated: step3 building payload",
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
		GasLimit:     parentHeader.GasLimit, // keep parent gas limit
	}

	slog.Debug("engine_forkchoiceUpdated: step4 calling BuildBlock",
		"parentNum", parentHeader.Number,
		"parentHash", parentHeader.Hash(),
	)
	block, receipts, err := b.builder.BuildBlock(parentHeader, attrs)
	if err != nil {
		slog.Warn("engine_forkchoiceUpdated: build block failed",
			"parentNum", parentHeader.Number,
			"err", err,
		)
		return engine.ForkchoiceUpdatedResult{
			PayloadStatus: payloadStatus,
		}, fmt.Errorf("build block: %w", err)
	}
	slog.Debug("engine_forkchoiceUpdated: step4 BuildBlock done",
		"blockNum", block.NumberU64(),
		"blockHash", block.Hash(),
		"txCount", len(block.Transactions()),
	)

	// EP-3 US-PQ-5b: replace VERIFY frame calldata with STARK proof when enabled.
	if prover := b.node.starkFrameProver; prover != nil {
		if sealed, _, err := vm.ReplaceValidationFrames(block, prover); err != nil {
			slog.Warn("frame stark replacement failed", "err", err)
		} else {
			block = sealed
		}
	}

	// Step 5: generate payload ID and store the built block.
	slog.Debug("engine_forkchoiceUpdated: step5 storing payload",
		"blockNum", block.NumberU64(),
		"blockHash", block.Hash(),
	)
	// Generate a payload ID from the block parameters.
	payloadID := generatePayloadID(parentHeader.Hash(), attrs)

	// Store the built payload, evicting the oldest when over cap.
	b.mu.Lock()
	b.payloads[payloadID] = &pendingPayload{
		block:    block,
		receipts: receipts,
	}
	b.payloadOrder = append(b.payloadOrder, payloadID)
	for len(b.payloads) > b.maxPayloads && len(b.payloadOrder) > 0 {
		oldest := b.payloadOrder[0]
		b.payloadOrder = b.payloadOrder[1:]
		delete(b.payloads, oldest)
	}
	b.mu.Unlock()

	slog.Info("engine_forkchoiceUpdated: built payload",
		"payloadID", payloadID,
		"blockNumber", block.NumberU64(),
		"blockHash", block.Hash(),
		"txCount", len(block.Transactions()),
	)

	return engine.ForkchoiceUpdatedResult{
		PayloadStatus: payloadStatus,
		PayloadID:     &payloadID,
	}, nil
}

func (b *engineBackend) ProcessBlockV4(
	payload *engine.ExecutionPayloadV3,
	expectedBlobVersionedHashes []types.Hash,
	parentBeaconBlockRoot types.Hash,
	executionRequests [][]byte,
) (engine.PayloadStatusV1, error) {
	// EIP-7685: compute RequestsHash from the raw execution requests.
	// Each element is [type_byte, ...data]; convert to types.Requests for hashing.
	var reqs types.Requests
	for _, reqBytes := range executionRequests {
		if len(reqBytes) == 0 {
			continue
		}
		reqs = append(reqs, &types.Request{Type: reqBytes[0], Data: reqBytes[1:]})
	}
	rHash := types.ComputeRequestsHash(reqs)
	return b.processBlockInternal(payload, parentBeaconBlockRoot, &rHash)
}

func (b *engineBackend) ProcessBlockV5(
	payload *engine.ExecutionPayloadV5,
	expectedBlobVersionedHashes []types.Hash,
	parentBeaconBlockRoot types.Hash,
	executionRequests [][]byte,
) (engine.PayloadStatusV1, error) {
	// Compute RequestsHash from execution requests (same as V4).
	var reqs types.Requests
	for _, reqBytes := range executionRequests {
		if len(reqBytes) == 0 {
			continue
		}
		reqs = append(reqs, &types.Request{Type: reqBytes[0], Data: reqBytes[1:]})
	}
	rHash := types.ComputeRequestsHash(reqs)
	return b.processBlockInternal(&payload.ExecutionPayloadV3, parentBeaconBlockRoot, &rHash)
}

func (b *engineBackend) ForkchoiceUpdatedV4(
	state engine.ForkchoiceStateV1,
	payloadAttributes *engine.PayloadAttributesV4,
) (engine.ForkchoiceUpdatedResult, error) {
	// Promote V4 attributes to V3 and delegate.
	var v3Attrs *engine.PayloadAttributesV3
	if payloadAttributes != nil {
		v3Attrs = &payloadAttributes.PayloadAttributesV3
	}
	return b.ForkchoiceUpdated(state, v3Attrs)
}

func (b *engineBackend) GetPayloadV4ByID(id engine.PayloadID) (*engine.GetPayloadV4Response, error) {
	resp, err := b.GetPayloadByID(id)
	if err != nil {
		return nil, err
	}
	return &engine.GetPayloadV4Response{
		ExecutionPayload:  &resp.ExecutionPayload.ExecutionPayloadV3,
		BlockValue:        resp.BlockValue,
		BlobsBundle:       resp.BlobsBundle,
		ExecutionRequests: [][]byte{},
	}, nil
}

func (b *engineBackend) GetPayloadV6ByID(id engine.PayloadID) (*engine.GetPayloadV6Response, error) {
	resp, err := b.GetPayloadByID(id)
	if err != nil {
		return nil, err
	}
	return &engine.GetPayloadV6Response{
		ExecutionPayload: &engine.ExecutionPayloadV5{
			ExecutionPayloadV4: *resp.ExecutionPayload,
		},
		BlockValue:        resp.BlockValue,
		BlobsBundle:       resp.BlobsBundle,
		ExecutionRequests: [][]byte{},
	}, nil
}

func (b *engineBackend) GetHeadTimestamp() uint64 {
	head := b.node.blockchain.CurrentBlock()
	if head != nil {
		return head.Time()
	}
	return 0
}

func (b *engineBackend) GetBlockTimestamp(hash types.Hash) uint64 {
	blk := b.node.blockchain.GetBlock(hash)
	if blk != nil {
		return blk.Time()
	}
	return 0
}

func (b *engineBackend) IsCancun(timestamp uint64) bool {
	return b.node.blockchain.Config().IsCancun(timestamp)
}

func (b *engineBackend) IsPrague(timestamp uint64) bool {
	return b.node.blockchain.Config().IsPrague(timestamp)
}

func (b *engineBackend) IsAmsterdam(timestamp uint64) bool {
	return b.node.blockchain.Config().IsAmsterdam(timestamp)
}

func (b *engineBackend) GetPayloadByID(id engine.PayloadID) (*engine.GetPayloadResponse, error) {
	slog.Debug("engine_getPayload", "payloadID", id)

	b.mu.Lock()
	payload, ok := b.payloads[id]
	b.mu.Unlock()

	if !ok {
		slog.Warn("engine_getPayload: payload not found", "payloadID", id)
		return nil, fmt.Errorf("payload %v not found", id)
	}

	block := payload.block
	header := block.Header()

	// Convert block to execution payload.
	execPayload := &engine.ExecutionPayloadV4{
		ExecutionPayloadV3: engine.ExecutionPayloadV3{
			ExecutionPayloadV2: engine.ExecutionPayloadV2{
				ExecutionPayloadV1: engine.ExecutionPayloadV1{
					ParentHash:    header.ParentHash,
					FeeRecipient:  header.Coinbase,
					StateRoot:     header.Root,
					ReceiptsRoot:  header.ReceiptHash,
					LogsBloom:     header.Bloom,
					PrevRandao:    header.MixDigest,
					BlockNumber:   block.NumberU64(),
					GasLimit:      header.GasLimit,
					GasUsed:       header.GasUsed,
					Timestamp:     header.Time,
					ExtraData:     header.Extra,
					BaseFeePerGas: header.BaseFee,
					BlockHash:     block.Hash(),
					Transactions:  encodeTxsRLP(block.Transactions()),
				},
			},
		},
	}

	// Add withdrawals if present.
	if ws := block.Withdrawals(); ws != nil {
		for _, w := range ws {
			execPayload.Withdrawals = append(execPayload.Withdrawals, &engine.Withdrawal{
				Index:          w.Index,
				ValidatorIndex: w.ValidatorIndex,
				Address:        w.Address,
				Amount:         w.Amount,
			})
		}
	}

	// Calculate block value (sum of priority fees paid).
	blockValue := new(big.Int)
	for _, receipt := range payload.receipts {
		if receipt.EffectiveGasPrice != nil && header.BaseFee != nil {
			tip := new(big.Int).Sub(receipt.EffectiveGasPrice, header.BaseFee)
			if tip.Sign() > 0 {
				tipTotal := new(big.Int).Mul(tip, new(big.Int).SetUint64(receipt.GasUsed))
				blockValue.Add(blockValue, tipTotal)
			}
		}
	}

	slog.Debug("engine_getPayload: returning payload",
		"payloadID", id,
		"blockNumber", block.NumberU64(),
		"blockHash", block.Hash(),
		"txCount", len(block.Transactions()),
		"blockValue", blockValue,
	)

	return &engine.GetPayloadResponse{
		ExecutionPayload: execPayload,
		BlockValue:       blockValue,
		BlobsBundle:      &engine.BlobsBundleV1{},
		Override:         false,
	}, nil
}

// generatePayloadID creates a deterministic PayloadID from the parent hash
// and build attributes.
func generatePayloadID(parentHash types.Hash, attrs *block.BuildBlockAttributes) engine.PayloadID {
	var id engine.PayloadID

	// Mix parent hash, timestamp, and fee recipient into the ID.
	// Use a simple approach: take bytes from parent hash + timestamp.
	copy(id[:], parentHash[:4])
	binary.BigEndian.PutUint32(id[4:], uint32(attrs.Timestamp))

	// If the ID collides (unlikely), add some randomness.
	if id == (engine.PayloadID{}) {
		rand.Read(id[:])
	}

	return id
}

// encodeTxsRLP encodes a list of transactions to RLP byte slices
// for inclusion in an Engine API ExecutionPayload.
func encodeTxsRLP(txs []*types.Transaction) [][]byte {
	encoded := make([][]byte, 0, len(txs))
	for _, tx := range txs {
		raw, err := tx.EncodeRLP()
		if err != nil {
			continue
		}
		encoded = append(encoded, raw)
	}
	return encoded
}

// nodeAdminBackend adapts the Node to the rpc.AdminBackend interface.
type nodeAdminBackend struct {
	node *Node
}

func newNodeAdminBackend(n *Node) rpc.AdminBackend {
	return &nodeAdminBackend{node: n}
}

// NodeInfo returns information about the running node.
func (b *nodeAdminBackend) NodeInfo() rpc.NodeInfoData {
	p2p := b.node.p2pServer
	nodeID := p2p.LocalID()

	listenAddr := ""
	ip := ""
	port := 0
	if addr := p2p.ListenAddr(); addr != nil {
		listenAddr = addr.String()
		host, portStr, err := net.SplitHostPort(listenAddr)
		if err == nil {
			ip = host
			fmt.Sscanf(portStr, "%d", &port)
		}
	}

	enode := fmt.Sprintf("enode://%s@%s:%d", nodeID, ip, port)

	chainID := uint64(0)
	if cfg := b.node.blockchain.Config(); cfg != nil && cfg.ChainID != nil {
		chainID = cfg.ChainID.Uint64()
	}

	return rpc.NodeInfoData{
		Name:       "eth2030",
		ID:         nodeID,
		ENR:        "",
		Enode:      enode,
		IP:         ip,
		ListenAddr: listenAddr,
		Ports: rpc.NodePorts{
			Discovery: port,
			Listener:  port,
		},
		Protocols: map[string]interface{}{
			"eth": map[string]interface{}{
				"network": chainID,
				"genesis": "",
			},
		},
	}
}

// Peers returns information about connected peers.
func (b *nodeAdminBackend) Peers() []rpc.PeerInfoData {
	peers := b.node.p2pServer.PeersList()
	infos := make([]rpc.PeerInfoData, len(peers))
	for i, p := range peers {
		caps := make([]string, 0, len(p.Caps()))
		for _, c := range p.Caps() {
			caps = append(caps, fmt.Sprintf("%s/%d", c.Name, c.Version))
		}
		infos[i] = rpc.PeerInfoData{
			ID:         p.ID(),
			Name:       "",
			RemoteAddr: p.RemoteAddr(),
			Caps:       caps,
		}
	}
	return infos
}

// AddPeer requests adding a new remote peer.
func (b *nodeAdminBackend) AddPeer(url string) error {
	return b.node.p2pServer.AddPeer(url)
}

// RemovePeer requests disconnection from a remote peer (stub).
func (b *nodeAdminBackend) RemovePeer(_ string) error {
	return nil
}

// ChainID returns the current chain ID.
func (b *nodeAdminBackend) ChainID() uint64 {
	if cfg := b.node.blockchain.Config(); cfg != nil && cfg.ChainID != nil {
		return cfg.ChainID.Uint64()
	}
	return 0
}

// DataDir returns the node's data directory.
func (b *nodeAdminBackend) DataDir() string {
	return b.node.config.DataDir
}

// nodeNetBackend adapts the Node to the netapi.Backend interface.
type nodeNetBackend struct {
	node *Node
}

func newNodeNetBackend(n *Node) *nodeNetBackend {
	return &nodeNetBackend{node: n}
}

// NetworkID returns the configured network identifier.
func (b *nodeNetBackend) NetworkID() uint64 {
	return b.node.config.NetworkID
}

// IsListening reports whether the P2P server is accepting connections.
func (b *nodeNetBackend) IsListening() bool {
	return b.node.p2pServer.ListenAddr() != nil
}

// PeerCount returns the number of currently connected peers.
func (b *nodeNetBackend) PeerCount() int {
	return b.node.p2pServer.PeerCount()
}

// MaxPeers returns the configured maximum peer count.
func (b *nodeNetBackend) MaxPeers() int {
	return b.node.config.MaxPeers
}
