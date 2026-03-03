package node

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"sync"

	"github.com/eth2030/eth2030/core"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/core/vm"
	"github.com/eth2030/eth2030/engine"
	"github.com/eth2030/eth2030/rpc"
	"github.com/eth2030/eth2030/trie"
)

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
	return b.node.txPool.AddLocal(tx)
}

func (b *nodeBackend) GetTransaction(hash types.Hash) (*types.Transaction, uint64, uint64) {
	// Check the blockchain's tx lookup index first.
	blockHash, blockNum, txIndex, found := b.node.blockchain.GetTransactionLookup(hash)
	if found {
		block := b.node.blockchain.GetBlock(blockHash)
		if block != nil {
			txs := block.Transactions()
			if int(txIndex) < len(txs) {
				return txs[txIndex], blockNum, txIndex
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
	// Return current base fee as a simple gas price suggestion.
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
	blockHash, _, txIndex, found := bc.GetTransactionLookup(txHash)
	if !found {
		return nil, fmt.Errorf("transaction %v not found", txHash)
	}

	block := bc.GetBlock(blockHash)
	if block == nil {
		return nil, fmt.Errorf("block %v not found", blockHash)
	}

	txs := block.Transactions()
	if int(txIndex) >= len(txs) {
		return nil, fmt.Errorf("transaction index %d out of range", txIndex)
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
	for i := uint64(0); i < txIndex; i++ {
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
	targetTx := txs[txIndex]
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

	to := targetTx.To()
	if to != nil {
		ret, gasLeft, err := evm.Call(from, *to, targetTx.Data(), targetTx.Gas(), targetTx.Value())
		gasUsed := targetTx.Gas() - gasLeft
		tracer.CaptureEnd(ret, gasUsed, err)
	}

	return tracer, nil
}

// txPoolAdapter adapts *txpool.TxPool to core.TxPoolReader.
type txPoolAdapter struct {
	node *Node
}

func (a *txPoolAdapter) Pending() []*types.Transaction {
	return a.node.txPool.PendingFlat()
}

// pendingPayload stores a built payload for later retrieval by getPayload.
type pendingPayload struct {
	block    *types.Block
	receipts []*types.Receipt
}

// engineBackend adapts the Node to the engine.Backend interface.
type engineBackend struct {
	node *Node

	mu       sync.Mutex
	payloads map[engine.PayloadID]*pendingPayload
	builder  *core.BlockBuilder
}

func newEngineBackend(n *Node) engine.Backend {
	pool := &txPoolAdapter{node: n}
	builder := core.NewBlockBuilder(n.blockchain.Config(), n.blockchain, pool)
	return &engineBackend{
		node:     n,
		payloads: make(map[engine.PayloadID]*pendingPayload),
		builder:  builder,
	}
}

func (b *engineBackend) ProcessBlock(
	payload *engine.ExecutionPayloadV3,
	expectedBlobVersionedHashes []types.Hash,
	parentBeaconBlockRoot types.Hash,
) (engine.PayloadStatusV1, error) {
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
		TxHash:        core.DeriveTxsRoot(txs),
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
		wHash := core.DeriveWithdrawalsRoot(ws)
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

	// Check if parent is known.
	if !bc.HasBlock(payload.ParentHash) {
		slog.Debug("engine_newPayload: parent unknown, returning SYNCING",
			"parentHash", payload.ParentHash,
		)
		return engine.PayloadStatusV1{
			Status: engine.StatusSyncing,
		}, nil
	}

	// Insert the block.
	if err := bc.InsertBlock(block); err != nil {
		slog.Warn("engine_newPayload: insert failed", "err", err)
		latestValid := payload.ParentHash
		return engine.PayloadStatusV1{
			Status:          engine.StatusInvalid,
			LatestValidHash: &latestValid,
		}, nil
	}

	// Sync txpool state with the new head so pending/queued txs are re-evaluated.
	b.node.txPool.Reset(bc.State())

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

	// Check if we know the head block.
	headBlock := bc.GetBlock(fcState.HeadBlockHash)
	var payloadStatus engine.PayloadStatusV1
	if headBlock == nil {
		slog.Warn("engine_forkchoiceUpdated: unknown head block, returning SYNCING",
			"headBlockHash", fcState.HeadBlockHash,
			"genesisHash", bc.Genesis().Hash(),
			"currentHead", bc.CurrentBlock().Hash(),
		)
		// We don't know this block yet; report syncing.
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

	// If no payload attributes, just return the forkchoice acknowledgment.
	if payloadAttributes == nil {
		return engine.ForkchoiceUpdatedResult{
			PayloadStatus: payloadStatus,
		}, nil
	}

	// Payload attributes provided: build a new block.
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
	attrs := &core.BuildBlockAttributes{
		Timestamp:    payloadAttributes.Timestamp,
		FeeRecipient: payloadAttributes.SuggestedFeeRecipient,
		Random:       payloadAttributes.PrevRandao,
		Withdrawals:  withdrawals,
		BeaconRoot:   &beaconRoot,
		GasLimit:     parentHeader.GasLimit, // keep parent gas limit
	}

	block, receipts, err := b.builder.BuildBlock(parentHeader, attrs)
	if err != nil {
		slog.Warn("engine_forkchoiceUpdated: build block failed", "err", err)
		return engine.ForkchoiceUpdatedResult{
			PayloadStatus: payloadStatus,
		}, fmt.Errorf("build block: %w", err)
	}

	// Generate a payload ID from the block parameters.
	payloadID := generatePayloadID(parentHeader.Hash(), attrs)

	// Store the built payload.
	b.mu.Lock()
	b.payloads[payloadID] = &pendingPayload{
		block:    block,
		receipts: receipts,
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
func generatePayloadID(parentHash types.Hash, attrs *core.BuildBlockAttributes) engine.PayloadID {
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
