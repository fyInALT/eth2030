package engine

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"

	"github.com/eth2030/eth2030/bal"
	"github.com/eth2030/eth2030/core/block"
	coreconfig "github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/execution"
	"github.com/eth2030/eth2030/core/gas"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
	enginepayload "github.com/eth2030/eth2030/engine/payload"
	"github.com/eth2030/eth2030/focil"
	"github.com/eth2030/eth2030/log"
	"github.com/eth2030/eth2030/metrics"
)

var backendLog = log.Default().Module("engine/backend")

const (
	// defaultMaxPayloads is the default cap for pending built payloads.
	defaultMaxPayloads = 32

	// defaultMaxILs is the default cap for inclusion lists stored per slot.
	defaultMaxILs = 256
)

// pendingPayload holds a payload being built by the block builder.
type pendingPayload struct {
	block        *types.Block
	receipts     []*types.Receipt
	bal          *bal.BlockAccessList // EIP-7928
	blockValue   *big.Int
	parentHash   types.Hash
	timestamp    uint64
	feeRecipient types.Address
	prevRandao   types.Hash
	withdrawals  []*Withdrawal
}

// EngineBackendConfig holds tunable memory limits for EngineBackend.
type EngineBackendConfig struct {
	// MaxPayloads caps the number of pending built payloads kept in memory.
	// Defaults to defaultMaxPayloads when zero.
	MaxPayloads int

	// MaxILs caps the number of inclusion lists stored per slot.
	// Defaults to defaultMaxILs when zero.
	MaxILs int
}

// EngineBackend is the execution-layer backend that connects the Engine API
// to the block builder and state processor.
type EngineBackend struct {
	mu            sync.RWMutex
	config        *coreconfig.ChainConfig
	statedb       state.StateDB
	processor     *execution.StateProcessor
	blocks        map[types.Hash]*types.Block
	bals          map[types.Hash]*bal.BlockAccessList // stored BALs for getPayloadBodiesV2
	ils           []*types.InclusionList              // received via engine_newInclusionListV1
	headHash      types.Hash
	safeHash      types.Hash
	finalHash     types.Hash
	payloads      map[PayloadID]*pendingPayload
	payloadOrder  []PayloadID // insertion order for payload LRU eviction
	nextPayloadID uint64
	maxPayloads   int // configurable cap; set from EngineBackendConfig
	maxILs        int // configurable cap; set from EngineBackendConfig
}

// NewEngineBackend creates a new Engine API backend with default memory limits.
func NewEngineBackend(config *coreconfig.ChainConfig, statedb state.StateDB, genesis *types.Block) *EngineBackend {
	return NewEngineBackendWithConfig(config, statedb, genesis, EngineBackendConfig{})
}

// NewEngineBackendWithConfig creates a new Engine API backend with explicit
// memory limits. Zero values in cfg fall back to package defaults.
func NewEngineBackendWithConfig(config *coreconfig.ChainConfig, statedb state.StateDB, genesis *types.Block, cfg EngineBackendConfig) *EngineBackend {
	if cfg.MaxPayloads <= 0 {
		cfg.MaxPayloads = defaultMaxPayloads
	}
	if cfg.MaxILs <= 0 {
		cfg.MaxILs = defaultMaxILs
	}
	b := &EngineBackend{
		config:      config,
		statedb:     statedb,
		processor:   execution.NewStateProcessor(config),
		blocks:      make(map[types.Hash]*types.Block),
		bals:        make(map[types.Hash]*bal.BlockAccessList),
		payloads:    make(map[PayloadID]*pendingPayload),
		maxPayloads: cfg.MaxPayloads,
		maxILs:      cfg.MaxILs,
	}
	if genesis != nil {
		h := genesis.Hash()
		b.blocks[h] = genesis
		b.headHash = h
		b.safeHash = h
		b.finalHash = h
	}
	return b
}

// evictOldBlocks removes block entries that are more than 64 blocks behind the
// current head from b.blocks and b.bals. Must be called with b.mu held.
func (b *EngineBackend) evictOldBlocks() {
	head, ok := b.blocks[b.headHash]
	if !ok || head.NumberU64() < 64 {
		return
	}
	cutoff := head.NumberU64() - 64
	for hash, blk := range b.blocks {
		if blk.NumberU64() < cutoff {
			delete(b.blocks, hash)
			delete(b.bals, hash)
		}
	}
}

// evictOldestPayload removes the oldest pending payload when the map exceeds
// b.maxPayloads. Must be called with b.mu held.
func (b *EngineBackend) evictOldestPayload() {
	for len(b.payloads) > b.maxPayloads {
		if len(b.payloadOrder) == 0 {
			break
		}
		oldest := b.payloadOrder[0]
		b.payloadOrder = b.payloadOrder[1:]
		delete(b.payloads, oldest)
	}
}

// ProcessBlock validates and executes a new payload from the consensus layer.
func (b *EngineBackend) ProcessBlock(
	payload *ExecutionPayloadV3,
	expectedBlobVersionedHashes []types.Hash,
	parentBeaconBlockRoot types.Hash,
) (PayloadStatusV1, error) {
	metrics.EngineNewPayload.Inc()
	backendLog.Debug("payload_received",
		"event", "payload_received",
		"blockHash", payload.BlockHash.Hex(),
		"blockNum", payload.BlockNumber,
		"parentHash", payload.ParentHash.Hex(),
		"txCount", len(payload.Transactions),
		"gasUsed", payload.GasUsed,
		"gasLimit", payload.GasLimit,
		"timestamp", payload.Timestamp,
	)

	blk, err := payloadToBlock(payload)
	if err != nil {
		errMsg := err.Error()
		backendLog.Warn("payload_invalid",
			"event", "payload_invalid",
			"blockHash", payload.BlockHash.Hex(),
			"blockNum", payload.BlockNumber,
			"reason", "payload decode failed",
			"error", errMsg,
		)
		return PayloadStatusV1{
			Status:          StatusInvalid,
			ValidationError: &errMsg,
		}, nil
	}

	// Validate block hash: the hash computed from the header fields must match
	// the blockHash provided in the payload.
	computedHash := blk.Hash()
	if payload.BlockHash != (types.Hash{}) && computedHash != payload.BlockHash {
		errMsg := fmt.Sprintf("block hash mismatch: computed %s, payload %s", computedHash, payload.BlockHash)
		backendLog.Warn("payload_invalid",
			"event", "payload_invalid",
			"blockHash", payload.BlockHash.Hex(),
			"blockNum", payload.BlockNumber,
			"reason", "block hash mismatch",
			"computedHash", computedHash.Hex(),
		)
		return PayloadStatusV1{
			Status:          StatusInvalidBlockHash,
			LatestValidHash: &computedHash,
			ValidationError: &errMsg,
		}, nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Check that the parent exists.
	parentHash := blk.ParentHash()
	if _, ok := b.blocks[parentHash]; !ok {
		backendLog.Debug("payload_syncing",
			"event", "payload_syncing",
			"blockHash", payload.BlockHash.Hex(),
			"blockNum", payload.BlockNumber,
			"parentHash", parentHash.Hex(),
		)
		return PayloadStatusV1{Status: StatusSyncing}, nil
	}

	// Validate timestamp progression: block timestamp must be > parent timestamp.
	parentBlock := b.blocks[parentHash]
	if parentBlock != nil && payload.Timestamp <= parentBlock.Header().Time {
		errMsg := fmt.Sprintf("invalid timestamp: block %d <= parent %d", payload.Timestamp, parentBlock.Header().Time)
		backendLog.Warn("payload_invalid",
			"event", "payload_invalid",
			"blockHash", payload.BlockHash.Hex(),
			"blockNum", payload.BlockNumber,
			"reason", "timestamp not advancing",
			"blockTs", payload.Timestamp,
			"parentTs", parentBlock.Header().Time,
		)
		return PayloadStatusV1{
			Status:          StatusInvalid,
			LatestValidHash: &parentHash,
			ValidationError: &errMsg,
		}, nil
	}

	// Run through the state processor.
	stateCopy := b.statedb.Dup()
	_, err = b.processor.Process(blk, stateCopy)
	if err != nil {
		errMsg := fmt.Sprintf("state processing failed: %v", err)
		backendLog.Error("payload_exec_fail",
			"event", "payload_exec_fail",
			"blockHash", payload.BlockHash.Hex(),
			"blockNum", payload.BlockNumber,
			"txCount", len(payload.Transactions),
			"error", err,
		)
		return PayloadStatusV1{
			Status:          StatusInvalid,
			ValidationError: &errMsg,
		}, nil
	}

	// Store the block and update state (evict old blocks to bound memory).
	blockHash := blk.Hash()
	b.blocks[blockHash] = blk
	b.evictOldBlocks()
	b.statedb = stateCopy

	backendLog.Info("payload_valid",
		"event", "payload_valid",
		"blockHash", blockHash.Hex(),
		"blockNum", payload.BlockNumber,
		"txCount", len(payload.Transactions),
		"gasUsed", payload.GasUsed,
	)

	return PayloadStatusV1{
		Status:          StatusValid,
		LatestValidHash: &blockHash,
	}, nil
}

// ProcessBlockV4 validates and executes a Prague payload with execution requests.
func (b *EngineBackend) ProcessBlockV4(
	payload *ExecutionPayloadV3,
	expectedBlobVersionedHashes []types.Hash,
	parentBeaconBlockRoot types.Hash,
	executionRequests [][]byte,
) (PayloadStatusV1, error) {
	// Delegate to ProcessBlock for core validation; execution requests are
	// stored alongside the block but validated at a higher level.
	return b.ProcessBlock(payload, expectedBlobVersionedHashes, parentBeaconBlockRoot)
}

// GetHeadTimestamp returns the timestamp of the current head block.
func (b *EngineBackend) GetHeadTimestamp() uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if headBlock, ok := b.blocks[b.headHash]; ok {
		return headBlock.Header().Time
	}
	return 0
}

// GetBlockTimestamp returns the timestamp of the block with the given hash,
// or 0 if the block is not known.
func (b *EngineBackend) GetBlockTimestamp(hash types.Hash) uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if blk, ok := b.blocks[hash]; ok {
		return blk.Header().Time
	}
	return 0
}

// IsCancun returns true if the given timestamp falls within the Cancun fork.
func (b *EngineBackend) IsCancun(timestamp uint64) bool {
	return b.config.IsCancun(timestamp)
}

// ForkchoiceUpdated processes a forkchoice state update from the CL.
func (b *EngineBackend) ForkchoiceUpdated(
	fcState ForkchoiceStateV1,
	attrs *PayloadAttributesV3,
) (ForkchoiceUpdatedResult, error) {
	metrics.EngineFCU.Inc()
	backendLog.Debug("fcu_received",
		"event", "fcu_received",
		"head", fcState.HeadBlockHash.Hex(),
		"safe", fcState.SafeBlockHash.Hex(),
		"finalized", fcState.FinalizedBlockHash.Hex(),
		"hasAttrs", attrs != nil,
	)

	b.mu.Lock()
	defer b.mu.Unlock()

	// Validate head block exists.
	if fcState.HeadBlockHash != (types.Hash{}) {
		if _, ok := b.blocks[fcState.HeadBlockHash]; !ok {
			backendLog.Debug("fcu_syncing",
				"event", "fcu_syncing",
				"head", fcState.HeadBlockHash.Hex(),
			)
			return ForkchoiceUpdatedResult{
				PayloadStatus: PayloadStatusV1{Status: StatusSyncing},
			}, nil
		}
	}

	// Update forkchoice pointers.
	b.headHash = fcState.HeadBlockHash
	b.safeHash = fcState.SafeBlockHash
	b.finalHash = fcState.FinalizedBlockHash

	headNum := uint64(0)
	if headBlock, ok := b.blocks[b.headHash]; ok {
		headNum = headBlock.NumberU64()
	}
	backendLog.Info("fcu_updated",
		"event", "fcu_updated",
		"head", b.headHash.Hex(),
		"headNum", headNum,
		"safe", b.safeHash.Hex(),
		"finalized", b.finalHash.Hex(),
		"hasAttrs", attrs != nil,
	)

	status := PayloadStatusV1{
		Status:          StatusValid,
		LatestValidHash: &b.headHash,
	}

	result := ForkchoiceUpdatedResult{PayloadStatus: status}

	// If payload attributes provided, start building a new payload.
	if attrs != nil {
		if attrs.Timestamp == 0 {
			return ForkchoiceUpdatedResult{}, ErrInvalidPayloadAttributes
		}

		parentBlock := b.blocks[fcState.HeadBlockHash]
		if parentBlock == nil {
			return ForkchoiceUpdatedResult{}, ErrInvalidForkchoiceState
		}

		// Validate timestamp progression: must be greater than parent block.
		if attrs.Timestamp <= parentBlock.Header().Time {
			return ForkchoiceUpdatedResult{}, ErrInvalidPayloadAttributes
		}

		id := b.generatePayloadID(fcState.HeadBlockHash, attrs.Timestamp)

		backendLog.Debug("payload_build_start",
			"event", "payload_build_start",
			"payloadID", fmt.Sprintf("%x", id),
			"parentHash", fcState.HeadBlockHash.Hex(),
			"parentNum", parentBlock.NumberU64(),
			"timestamp", attrs.Timestamp,
			"feeRecipient", attrs.SuggestedFeeRecipient.Hex(),
		)

		// Build an empty block (no pending transactions from txpool yet).
		builder := block.NewBlockBuilder(b.config, nil, nil)
		builder.SetState(b.statedb.Dup())
		parentHeader := parentBlock.Header()

		blk, receipts, err := builder.BuildBlock(parentHeader, &block.BuildBlockAttributes{
			Timestamp:    attrs.Timestamp,
			FeeRecipient: attrs.SuggestedFeeRecipient,
			Random:       attrs.PrevRandao,
			GasLimit:     parentHeader.GasLimit,
			Withdrawals:  WithdrawalsToCore(attrs.Withdrawals),
		})
		if err != nil {
			backendLog.Error("payload_build_fail",
				"event", "payload_build_fail",
				"payloadID", fmt.Sprintf("%x", id),
				"parentHash", fcState.HeadBlockHash.Hex(),
				"error", err,
			)
			return ForkchoiceUpdatedResult{}, fmt.Errorf("payload build failed: %w", err)
		}

		backendLog.Debug("payload_build_done",
			"event", "payload_build_done",
			"payloadID", fmt.Sprintf("%x", id),
			"blockHash", blk.Hash().Hex(),
			"blockNum", blk.NumberU64(),
			"txCount", len(blk.Transactions()),
			"gasUsed", blk.Header().GasUsed,
			"gasLimit", blk.Header().GasLimit,
		)

		b.payloads[id] = &pendingPayload{
			block:        blk,
			receipts:     receipts,
			blockValue:   new(big.Int),
			parentHash:   fcState.HeadBlockHash,
			timestamp:    attrs.Timestamp,
			feeRecipient: attrs.SuggestedFeeRecipient,
			prevRandao:   attrs.PrevRandao,
			withdrawals:  attrs.Withdrawals,
		}
		b.payloadOrder = append(b.payloadOrder, id)
		b.evictOldestPayload()

		result.PayloadID = &id
	}

	return result, nil
}

// GetPayloadByID retrieves a previously built payload by its ID.
func (b *EngineBackend) GetPayloadByID(id PayloadID) (*GetPayloadResponse, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	pending, ok := b.payloads[id]
	if !ok {
		backendLog.Warn("payload_get_miss",
			"event", "payload_get_miss",
			"payloadID", fmt.Sprintf("%x", id),
		)
		return nil, ErrUnknownPayload
	}

	ep := enginepayload.BlockToPayload(pending.block, pending.prevRandao, pending.withdrawals)

	backendLog.Debug("payload_get",
		"event", "payload_get",
		"payloadID", fmt.Sprintf("%x", id),
		"blockHash", pending.block.Hash().Hex(),
		"blockNum", pending.block.NumberU64(),
		"txCount", len(pending.block.Transactions()),
		"blockValue", pending.blockValue.String(),
	)

	return &GetPayloadResponse{
		ExecutionPayload: ep,
		BlockValue:       new(big.Int).Set(pending.blockValue),
		BlobsBundle:      collectBlobsBundle(pending.block.Transactions()),
	}, nil
}

// generatePayloadID creates a unique payload ID from parent hash and timestamp.
func (b *EngineBackend) generatePayloadID(parentHash types.Hash, timestamp uint64) PayloadID {
	b.nextPayloadID++
	var id PayloadID
	binary.BigEndian.PutUint64(id[:], b.nextPayloadID)
	return id
}

// payloadToBlock converts an ExecutionPayloadV3 to a types.Block.
func payloadToBlock(payload *ExecutionPayloadV3) (*types.Block, error) {
	// Use the existing PayloadToHeader helper (which takes V4).
	v4 := &ExecutionPayloadV4{ExecutionPayloadV3: *payload}
	header := PayloadToHeader(v4)

	// Decode transactions.
	txs := make([]*types.Transaction, len(payload.Transactions))
	for i, enc := range payload.Transactions {
		tx, err := types.DecodeTxRLP(enc)
		if err != nil {
			return nil, fmt.Errorf("invalid transaction %d: %w", i, err)
		}
		txs = append(txs[:i], tx)
	}

	// Convert withdrawals.
	var withdrawals []*types.Withdrawal
	if payload.Withdrawals != nil {
		withdrawals = WithdrawalsToCore(payload.Withdrawals)
	}

	block := types.NewBlock(header, &types.Body{
		Transactions: txs,
		Withdrawals:  withdrawals,
	})
	return block, nil
}

// restoreCalldataGasFields recomputes CalldataGasUsed and CalldataExcessGas
// for a block received via the engine API. These fields are part of the
// EIP-7706 header but are not transmitted in ExecutionPayloadV5. By deriving
// them from block transactions and parent state we restore the original block
// hash so the parent chain tracks calldata excess gas correctly (SPEC-6.4).
// If the recomputed block hash matches payloadBlockHash the augmented block is
// returned; otherwise the original block is returned unchanged.
func restoreCalldataGasFields(block *types.Block, parent *types.Block, payloadBlockHash types.Hash) *types.Block {
	// Sum calldata gas across all transactions.
	calldataGasUsed := uint64(0)
	for _, tx := range block.Transactions() {
		calldataGasUsed += tx.CalldataGas()
	}
	// Derive excess gas from parent state (defaults to 0 if parent lacks fields).
	parentExcess, parentUsed := uint64(0), uint64(0)
	if parent != nil {
		if parent.Header().CalldataExcessGas != nil {
			parentExcess = *parent.Header().CalldataExcessGas
		}
		if parent.Header().CalldataGasUsed != nil {
			parentUsed = *parent.Header().CalldataGasUsed
		}
	}
	calldataExcessGas := gas.CalcCalldataExcessGas(parentExcess, parentUsed, block.Header().GasLimit)

	// Rebuild the block with the calldata gas fields injected into the header.
	hdr := block.Header()
	hdr.CalldataGasUsed = &calldataGasUsed
	hdr.CalldataExcessGas = &calldataExcessGas
	augmented := types.NewBlock(hdr, block.Body())

	// Only use the augmented block if its hash matches what the builder produced.
	// This protects against incorrect recomputation.
	if payloadBlockHash == (types.Hash{}) || augmented.Hash() == payloadBlockHash {
		return augmented
	}
	return block
}

// ProcessBlockV5 validates and executes an Amsterdam payload with BAL validation.
// NOTE: EIP-8141 FrameTx receipts are included in the standard receipt array.
// The FrameTxReceipt type (core/types/frame_receipt.go) provides per-frame
// results; however, gas accounting flows through the standard Receipt structure
// so no special handling is needed here.
func (b *EngineBackend) ProcessBlockV5(
	payload *ExecutionPayloadV5,
	expectedBlobVersionedHashes []types.Hash,
	parentBeaconBlockRoot types.Hash,
	executionRequests [][]byte,
) (PayloadStatusV1, error) {
	metrics.EngineNewPayload.Inc()
	backendLog.Debug("payload_received",
		"event", "payload_received",
		"version", "V5",
		"blockHash", payload.BlockHash.Hex(),
		"blockNum", payload.BlockNumber,
		"parentHash", payload.ParentHash.Hex(),
		"txCount", len(payload.Transactions),
		"gasUsed", payload.GasUsed,
		"timestamp", payload.Timestamp,
	)

	// First, process the block through the standard path.
	blk, err := payloadToBlock(&payload.ExecutionPayloadV3)
	if err != nil {
		errMsg := err.Error()
		backendLog.Warn("payload_invalid",
			"event", "payload_invalid",
			"version", "V5",
			"blockHash", payload.BlockHash.Hex(),
			"reason", "payload decode failed",
			"error", errMsg,
		)
		return PayloadStatusV1{
			Status:          StatusInvalid,
			ValidationError: &errMsg,
		}, nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Check that the parent exists.
	parentHash := blk.ParentHash()
	if _, ok := b.blocks[parentHash]; !ok {
		backendLog.Debug("payload_syncing",
			"event", "payload_syncing",
			"version", "V5",
			"blockHash", payload.BlockHash.Hex(),
			"parentHash", parentHash.Hex(),
		)
		return PayloadStatusV1{Status: StatusSyncing}, nil
	}

	// SPEC-6.4: restore CalldataGasUsed/CalldataExcessGas stripped by the
	// engine API wire format. These fields are part of the header hash
	// (EIP-7706) but not included in ExecutionPayloadV5. We recompute them
	// from the block's transactions and the parent's calldata gas state.
	if b.config != nil && b.config.IsGlamsterdan(blk.Header().Time) {
		blk = restoreCalldataGasFields(blk, b.blocks[parentHash], payload.BlockHash)
	}

	// Run through the state processor with BAL computation.
	stateCopy := b.statedb.Dup()
	result, err := b.processor.ProcessWithBAL(blk, stateCopy)
	if err != nil {
		errMsg := fmt.Sprintf("state processing failed: %v", err)
		backendLog.Error("payload_exec_fail",
			"event", "payload_exec_fail",
			"version", "V5",
			"blockHash", payload.BlockHash.Hex(),
			"blockNum", payload.BlockNumber,
			"error", err,
		)
		return PayloadStatusV1{
			Status:          StatusInvalid,
			ValidationError: &errMsg,
		}, nil
	}

	// Validate the BAL by comparing the computed BAL with the provided one.
	if payload.BlockAccessList != nil {
		computedBAL := result.BlockAccessList
		if computedBAL == nil {
			computedBAL = bal.NewBlockAccessList()
		}
		computedEncoded, _ := computedBAL.EncodeRLP()

		var providedBALBytes []byte
		if err := json.Unmarshal(payload.BlockAccessList, &providedBALBytes); err != nil {
			// If the blockAccessList isn't valid JSON bytes, it may be null.
			if string(payload.BlockAccessList) != "null" {
				errMsg := fmt.Sprintf("invalid blockAccessList encoding: %v", err)
				backendLog.Warn("payload_bal_invalid",
					"event", "payload_bal_invalid",
					"blockHash", payload.BlockHash.Hex(),
					"reason", "BAL JSON decode failed",
					"error", errMsg,
				)
				return PayloadStatusV1{
					Status:          StatusInvalid,
					ValidationError: &errMsg,
				}, nil
			}
		} else if !bytes.Equal(computedEncoded, providedBALBytes) {
			errMsg := "blockAccessList mismatch: computed BAL does not match provided BAL"
			backendLog.Warn("payload_bal_mismatch",
				"event", "payload_bal_mismatch",
				"blockHash", payload.BlockHash.Hex(),
				"blockNum", payload.BlockNumber,
			)
			return PayloadStatusV1{
				Status:          StatusInvalid,
				ValidationError: &errMsg,
			}, nil
		}
	}

	// EIP-7805: check IL satisfaction against block and stored ILs.
	if len(b.ils) > 0 {
		ils := b.ilsAsFocil()
		gasRemaining := blk.GasLimit() - blk.GasUsed()
		if result := focilCheckILSatisfaction(blk, ils, gasRemaining); !result {
			errMsg := focil.InclusionListUnsatisfied
			backendLog.Warn("payload_il_unsatisfied",
				"event", "payload_il_unsatisfied",
				"blockHash", payload.BlockHash.Hex(),
				"blockNum", payload.BlockNumber,
			)
			return PayloadStatusV1{
				Status:          StatusInclusionListUnsatisfied,
				ValidationError: &errMsg,
			}, nil
		}
	}

	// Store the block and update state (evict old blocks to bound memory).
	blockHash := blk.Hash()
	b.blocks[blockHash] = blk
	// Store BAL for engine_getPayloadBodiesByHashV2.
	if result.BlockAccessList != nil {
		b.bals[blockHash] = result.BlockAccessList
	}
	b.evictOldBlocks()
	b.statedb = stateCopy

	// ILs are slot-scoped: once a block is accepted, the ILs for that slot
	// are consumed. Clear them to prevent unbounded growth across slots.
	b.ils = b.ils[:0]

	backendLog.Info("payload_valid",
		"event", "payload_valid",
		"version", "V5",
		"blockHash", blockHash.Hex(),
		"blockNum", payload.BlockNumber,
		"txCount", len(payload.Transactions),
		"gasUsed", payload.GasUsed,
	)

	return PayloadStatusV1{
		Status:          StatusValid,
		LatestValidHash: &blockHash,
	}, nil
}

// ForkchoiceUpdatedV4 processes a forkchoice update with V4 payload attributes (Amsterdam).
func (b *EngineBackend) ForkchoiceUpdatedV4(
	fcState ForkchoiceStateV1,
	attrs *PayloadAttributesV4,
) (ForkchoiceUpdatedResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Validate head block exists.
	if fcState.HeadBlockHash != (types.Hash{}) {
		if _, ok := b.blocks[fcState.HeadBlockHash]; !ok {
			return ForkchoiceUpdatedResult{
				PayloadStatus: PayloadStatusV1{Status: StatusSyncing},
			}, nil
		}
	}

	// Update forkchoice pointers (per spec: must NOT be rolled back on attribute errors).
	b.headHash = fcState.HeadBlockHash
	b.safeHash = fcState.SafeBlockHash
	b.finalHash = fcState.FinalizedBlockHash

	status := PayloadStatusV1{
		Status:          StatusValid,
		LatestValidHash: &b.headHash,
	}

	result := ForkchoiceUpdatedResult{PayloadStatus: status}

	if attrs != nil {
		if attrs.Timestamp == 0 {
			return ForkchoiceUpdatedResult{}, ErrInvalidPayloadAttributes
		}

		id := b.generatePayloadID(fcState.HeadBlockHash, attrs.Timestamp)

		parentBlock := b.blocks[fcState.HeadBlockHash]
		if parentBlock == nil {
			return ForkchoiceUpdatedResult{}, ErrInvalidForkchoiceState
		}

		// Validate that timestamp is greater than parent.
		if attrs.Timestamp <= parentBlock.Header().Time {
			return ForkchoiceUpdatedResult{}, ErrInvalidPayloadAttributes
		}

		builder := block.NewBlockBuilder(b.config, nil, nil)
		builder.SetState(b.statedb.Dup())
		parentHeader := parentBlock.Header()

		blk, receipts, err := builder.BuildBlock(parentHeader, &block.BuildBlockAttributes{
			Timestamp:        attrs.Timestamp,
			FeeRecipient:     attrs.SuggestedFeeRecipient,
			Random:           attrs.PrevRandao,
			GasLimit:         parentHeader.GasLimit,
			Withdrawals:      WithdrawalsToCore(attrs.Withdrawals),
			InclusionListTxs: attrs.InclusionListTransactions,
		})
		if err != nil {
			return ForkchoiceUpdatedResult{}, fmt.Errorf("payload build failed: %w", err)
		}

		// Compute BAL for the built block (EIP-7928).
		var blockBAL *bal.BlockAccessList
		if b.config.IsAmsterdam(attrs.Timestamp) {
			stateCopy2 := b.statedb.Dup()
			balResult, err := b.processor.ProcessWithBAL(blk, stateCopy2)
			if err == nil && balResult != nil {
				blockBAL = balResult.BlockAccessList
			}
		}

		b.payloads[id] = &pendingPayload{
			block:        blk,
			receipts:     receipts,
			bal:          blockBAL,
			blockValue:   new(big.Int),
			parentHash:   fcState.HeadBlockHash,
			timestamp:    attrs.Timestamp,
			feeRecipient: attrs.SuggestedFeeRecipient,
			prevRandao:   attrs.PrevRandao,
			withdrawals:  attrs.Withdrawals,
		}
		b.payloadOrder = append(b.payloadOrder, id)
		b.evictOldestPayload()

		result.PayloadID = &id
	}

	return result, nil
}

// GetPayloadV4ByID retrieves a previously built payload for getPayloadV4 (Prague).
func (b *EngineBackend) GetPayloadV4ByID(id PayloadID) (*GetPayloadV4Response, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	pending, ok := b.payloads[id]
	if !ok {
		return nil, ErrUnknownPayload
	}

	ep4 := enginepayload.BlockToPayload(pending.block, pending.prevRandao, pending.withdrawals)

	return &GetPayloadV4Response{
		ExecutionPayload:  &ep4.ExecutionPayloadV3,
		BlockValue:        new(big.Int).Set(pending.blockValue),
		BlobsBundle:       collectBlobsBundle(pending.block.Transactions()),
		ExecutionRequests: [][]byte{},
	}, nil
}

// GetPayloadV6ByID retrieves a previously built payload for getPayloadV6 (Amsterdam).
// NOTE: EIP-8141 FrameTx receipts use the standard Receipt structure for gas
// accounting. Per-frame results are available via FrameTxReceipt if needed by
// downstream consumers (e.g., block explorers).
func (b *EngineBackend) GetPayloadV6ByID(id PayloadID) (*GetPayloadV6Response, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	pending, ok := b.payloads[id]
	if !ok {
		return nil, ErrUnknownPayload
	}

	ep5 := enginepayload.BlockToPayloadV5(pending.block, pending.prevRandao, pending.withdrawals, pending.bal)

	return &GetPayloadV6Response{
		ExecutionPayload:  ep5,
		BlockValue:        new(big.Int).Set(pending.blockValue),
		BlobsBundle:       collectBlobsBundle(pending.block.Transactions()),
		ExecutionRequests: [][]byte{},
	}, nil
}

// collectBlobsBundle builds a BlobsBundleV1 from the blob sidecar data attached
// to the given transactions. Returns an empty bundle if no blob transactions are present.
func collectBlobsBundle(txs []*types.Transaction) *BlobsBundleV1 {
	bundle := &BlobsBundleV1{}
	for _, tx := range txs {
		sc := tx.BlobSidecar()
		if sc == nil {
			continue
		}
		bundle.Blobs = append(bundle.Blobs, sc.Blobs...)
		bundle.Commitments = append(bundle.Commitments, sc.Commitments...)
		bundle.Proofs = append(bundle.Proofs, sc.Proofs...)
	}
	return bundle
}

// IsPrague returns true if the given timestamp falls within the Prague fork.
func (b *EngineBackend) IsPrague(timestamp uint64) bool {
	return b.config.IsPrague(timestamp)
}

// IsAmsterdam returns true if the given timestamp falls within the Amsterdam fork.
func (b *EngineBackend) IsAmsterdam(timestamp uint64) bool {
	return b.config.IsAmsterdam(timestamp)
}

// ProcessInclusionList validates and stores a new inclusion list from the CL.
// Implements InclusionListBackend.
func (b *EngineBackend) ProcessInclusionList(il *types.InclusionList) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ils = append(b.ils, il)
	// Bound the IL slice: drop oldest entries beyond b.maxILs.
	if len(b.ils) > b.maxILs {
		b.ils = b.ils[len(b.ils)-b.maxILs:]
	}
	return nil
}

// GetInclusionList generates an inclusion list from the mempool (stub: returns empty IL).
// Implements InclusionListBackend.
func (b *EngineBackend) GetInclusionList() *types.InclusionList {
	return &types.InclusionList{Transactions: [][]byte{}}
}

// ilsAsFocil converts stored types.InclusionList entries to focil.InclusionList format.
func (b *EngineBackend) ilsAsFocil() []*focil.InclusionList {
	result := make([]*focil.InclusionList, len(b.ils))
	for i, il := range b.ils {
		entries := make([]focil.InclusionListEntry, len(il.Transactions))
		for j, tx := range il.Transactions {
			entries[j] = focil.InclusionListEntry{Transaction: tx, Index: uint64(j)}
		}
		result[i] = &focil.InclusionList{
			Slot:          il.Slot,
			ProposerIndex: il.ValidatorIndex,
			CommitteeRoot: il.CommitteeRoot,
			Entries:       entries,
		}
	}
	return result
}

// SetSlasher wires a PaymasterSlasher into the state processor so the
// processor can slash paymasters on bad settlement (AA EIP-7701).
func (b *EngineBackend) SetSlasher(s coreconfig.PaymasterSlasher) {
	b.processor.SetSlasher(s)
}

// focilCheckILSatisfaction wraps focil.CheckILSatisfaction for use in engine_newPayload.
func focilCheckILSatisfaction(block *types.Block, ils []*focil.InclusionList, gasRemaining uint64) bool {
	return focil.CheckILSatisfaction(block, ils, nil, gasRemaining) == focil.ILSatisfied
}

// Verify interface compliance at compile time.
var _ Backend = (*EngineBackend)(nil)
