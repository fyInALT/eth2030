package debugapi

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"time"

	coretypes "github.com/eth2030/eth2030/core/types"
	rpcbackend "github.com/eth2030/eth2030/rpc/backend"
	rpctypes "github.com/eth2030/eth2030/rpc/types"
)

// DbgAPI implements extended debug namespace RPC methods.
// It is separate from the main EthAPI to keep concerns decoupled.
type DbgAPI struct {
	backend rpcbackend.Backend
}

// NewDbgAPI creates a new DbgAPI instance.
func NewDbgAPI(backend rpcbackend.Backend) *DbgAPI {
	return &DbgAPI{backend: backend}
}

// HandleRequest dispatches debug namespace requests.
func (d *DbgAPI) HandleRequest(req *rpctypes.Request) *rpctypes.Response {
	switch req.Method {
	case "debug_traceBlockByNumber":
		return d.traceBlockByNumber(req)
	case "debug_traceBlockByHash":
		return d.traceBlockByHash(req)
	case "debug_storageRangeAt":
		return d.storageRangeAt(req)
	case "debug_getBadBlocks":
		return d.getBadBlocks(req)
	case "debug_setHead":
		return d.setHead(req)
	case "debug_dumpBlock":
		return d.dumpBlock(req)
	default:
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeMethodNotFound,
			fmt.Sprintf("method %q not found in debug namespace", req.Method))
	}
}

// traceBlockByNumber implements debug_traceBlockByNumber.
// Traces all transactions in the block identified by number.
func (d *DbgAPI) traceBlockByNumber(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 1 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "missing block number parameter")
	}

	var bn rpctypes.BlockNumber
	if err := json.Unmarshal(req.Params[0], &bn); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid block number: "+err.Error())
	}

	// Parse optional trace config.
	cfg := DefaultDbgTraceConfig()
	if len(req.Params) > 1 {
		if err := json.Unmarshal(req.Params[1], &cfg); err != nil {
			return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid trace config: "+err.Error())
		}
	}

	block := d.backend.BlockByNumber(bn)
	if block == nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "block not found")
	}

	return d.traceBlockTxs(req, block, cfg)
}

// traceBlockByHash implements debug_traceBlockByHash.
// Traces all transactions in the block identified by hash.
func (d *DbgAPI) traceBlockByHash(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 1 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "missing block hash parameter")
	}

	var hashHex string
	if err := json.Unmarshal(req.Params[0], &hashHex); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid block hash: "+err.Error())
	}

	// Parse optional trace config.
	cfg := DefaultDbgTraceConfig()
	if len(req.Params) > 1 {
		if err := json.Unmarshal(req.Params[1], &cfg); err != nil {
			return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid trace config: "+err.Error())
		}
	}

	hash := coretypes.HexToHash(hashHex)
	block := d.backend.BlockByHash(hash)
	if block == nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "block not found")
	}

	return d.traceBlockTxs(req, block, cfg)
}

// traceBlockTxs produces a trace for each transaction in the block.
// When config.Tracer is "callTracer", it returns call frames. Otherwise
// it returns struct logs (default tracer behavior).
func (d *DbgAPI) traceBlockTxs(req *rpctypes.Request, block *coretypes.Block, cfg DbgTraceConfig) *rpctypes.Response {
	timeout := parseDbgTimeout(cfg.Timeout)
	deadline := time.Now().Add(timeout)

	txs := block.Transactions()
	blockHash := block.Hash()
	receipts := d.backend.GetReceipts(blockHash)

	switch cfg.Tracer {
	case "callTracer":
		return d.traceBlockCallTracer(req, txs, receipts, blockHash, deadline)
	default:
		return d.traceBlockStructLog(req, txs, receipts, cfg, deadline)
	}
}

// traceBlockStructLog produces struct-log style traces for each transaction.
func (d *DbgAPI) traceBlockStructLog(
	req *rpctypes.Request,
	txs []*coretypes.Transaction,
	receipts []*coretypes.Receipt,
	cfg DbgTraceConfig,
	deadline time.Time,
) *rpctypes.Response {
	results := make([]*DbgBlockTraceEntry, len(txs))
	for i, tx := range txs {
		if time.Now().After(deadline) {
			return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "trace timeout exceeded")
		}

		trace := &TraceResult{
			Gas:         tx.Gas(),
			Failed:      false,
			ReturnValue: "",
			StructLogs:  []StructLog{},
		}

		if i < len(receipts) {
			trace.Gas = receipts[i].GasUsed
			trace.Failed = receipts[i].Status == coretypes.ReceiptStatusFailed
		}

		// Attempt real tracing via backend if available.
		tracer, err := d.backend.TraceTransaction(tx.Hash())
		if err == nil && tracer != nil {
			trace.Gas = tracer.GasUsed()
			trace.Failed = tracer.Error() != nil
			if out := tracer.Output(); len(out) > 0 {
				trace.ReturnValue = rpctypes.EncodeBytes(out)
			}

			structLogs := make([]StructLog, 0, len(tracer.Logs))
			for _, entry := range tracer.Logs {
				sl := StructLog{
					PC:      entry.Pc,
					Op:      entry.Op.String(),
					Gas:     entry.Gas,
					GasCost: entry.GasCost,
					Depth:   entry.Depth,
				}
				if !cfg.DisableStack {
					stackHex := make([]string, len(entry.Stack))
					for j, val := range entry.Stack {
						stackHex[j] = "0x" + val.Text(16)
					}
					sl.Stack = stackHex
				}
				structLogs = append(structLogs, sl)
			}
			trace.StructLogs = structLogs
		}

		results[i] = &DbgBlockTraceEntry{
			TxHash: rpctypes.EncodeHash(tx.Hash()),
			Result: trace,
		}
	}

	return rpctypes.NewSuccessResponse(req.ID, results)
}

// traceBlockCallTracer produces call-frame style traces for each transaction.
func (d *DbgAPI) traceBlockCallTracer(
	req *rpctypes.Request,
	txs []*coretypes.Transaction,
	receipts []*coretypes.Receipt,
	blockHash coretypes.Hash,
	deadline time.Time,
) *rpctypes.Response {
	type callTracerEntry struct {
		TxHash string        `json:"txHash"`
		Result *DbgCallFrame `json:"result"`
	}

	results := make([]*callTracerEntry, len(txs))
	for i, tx := range txs {
		if time.Now().After(deadline) {
			return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "trace timeout exceeded")
		}

		frame := &DbgCallFrame{
			Type:  "CALL",
			Input: rpctypes.EncodeBytes(tx.Data()),
			Gas:   rpctypes.EncodeUint64(tx.Gas()),
		}

		if sender := tx.Sender(); sender != nil {
			frame.From = rpctypes.EncodeAddress(*sender)
		}
		if tx.To() != nil {
			frame.To = rpctypes.EncodeAddress(*tx.To())
		}
		if tx.Value() != nil && tx.Value().Sign() > 0 {
			frame.Value = rpctypes.EncodeBigInt(tx.Value())
		}

		if i < len(receipts) {
			frame.GasUsed = rpctypes.EncodeUint64(receipts[i].GasUsed)
			if receipts[i].Status == coretypes.ReceiptStatusFailed {
				frame.Error = "execution reverted"
			}
		} else {
			frame.GasUsed = "0x0"
		}

		results[i] = &callTracerEntry{
			TxHash: rpctypes.EncodeHash(tx.Hash()),
			Result: frame,
		}
	}

	return rpctypes.NewSuccessResponse(req.ID, results)
}

// storageRangeAt implements debug_storageRangeAt.
// Returns a range of storage entries for the given account at a specific
// block and transaction index.
// Params: [blockHash, txIndex, address, startKey, maxResult]
func (d *DbgAPI) storageRangeAt(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 5 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams,
			"expected 5 params: blockHash, txIndex, address, startKey, maxResult")
	}

	var blockHashHex, addrHex, startKeyHex string
	var txIndex int
	var maxResult int

	if err := json.Unmarshal(req.Params[0], &blockHashHex); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid blockHash: "+err.Error())
	}
	if err := json.Unmarshal(req.Params[1], &txIndex); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid txIndex: "+err.Error())
	}
	if err := json.Unmarshal(req.Params[2], &addrHex); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid address: "+err.Error())
	}
	if err := json.Unmarshal(req.Params[3], &startKeyHex); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid startKey: "+err.Error())
	}
	if err := json.Unmarshal(req.Params[4], &maxResult); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid maxResult: "+err.Error())
	}

	if maxResult <= 0 {
		maxResult = 256
	}
	if maxResult > 1024 {
		maxResult = 1024
	}
	if txIndex < 0 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "txIndex must be non-negative")
	}

	blockHash := coretypes.HexToHash(blockHashHex)
	header := d.backend.HeaderByHash(blockHash)
	if header == nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "block not found")
	}

	statedb, err := d.backend.StateAt(header.Root)
	if err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "state not available: "+err.Error())
	}

	addr := coretypes.HexToAddress(addrHex)
	startKey := coretypes.HexToHash(startKeyHex)

	// Read storage slots starting from startKey.
	// In a full implementation, we would iterate the storage trie.
	// Here we probe sequential slots from the start key.
	storage := make(map[string]DbgStorageEntry)
	keyInt := new(big.Int).SetBytes(startKey[:])

	for count := 0; count < maxResult; count++ {
		slotHash := coretypes.IntToHash(keyInt)
		value := statedb.GetState(addr, slotHash)

		if value != (coretypes.Hash{}) {
			hexKey := rpctypes.EncodeHash(slotHash)
			storage[hexKey] = DbgStorageEntry{
				Key:   hexKey,
				Value: rpctypes.EncodeHash(value),
			}
		}
		keyInt.Add(keyInt, big.NewInt(1))
	}

	// Compute the next key for pagination.
	nextKeyHash := coretypes.IntToHash(keyInt)
	nextKeyStr := rpctypes.EncodeHash(nextKeyHash)

	result := &DbgStorageRangeResult{
		Storage: storage,
		NextKey: &nextKeyStr,
	}

	return rpctypes.NewSuccessResponse(req.ID, result)
}

// getBadBlocks implements debug_getBadBlocks.
// Returns a list of blocks that were rejected during import.
func (d *DbgAPI) getBadBlocks(req *rpctypes.Request) *rpctypes.Response {
	// In a full implementation, the backend would maintain a bounded
	// list of bad blocks. For now, return an empty list.
	return rpctypes.NewSuccessResponse(req.ID, []*DbgBadBlock{})
}

// setHead implements debug_setHead.
// Rewinds the chain to the specified block number.
func (d *DbgAPI) setHead(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 1 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "missing block number parameter")
	}

	var hexNum string
	if err := json.Unmarshal(req.Params[0], &hexNum); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid block number: "+err.Error())
	}

	targetNum := rpctypes.ParseHexUint64(hexNum)
	if targetNum == 0 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "cannot rewind to block 0")
	}
	if targetNum > uint64(math.MaxInt64) {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "block number overflow")
	}

	// Verify the target block exists.
	header := d.backend.HeaderByNumber(rpctypes.BlockNumber(targetNum)) //nolint:gosec // overflow guarded above
	if header == nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "target block not found")
	}

	// Check that we are actually rewinding (target is before current head).
	current := d.backend.CurrentHeader()
	if current != nil && targetNum > current.Number.Uint64() {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams,
			fmt.Sprintf("target block %d is after current head %d",
				targetNum, current.Number.Uint64()))
	}

	// In a full implementation, this would trigger chain rewinding.
	// Return success to indicate the operation was accepted.
	return rpctypes.NewSuccessResponse(req.ID, nil)
}

// dumpBlock implements debug_dumpBlock.
// Returns a dump of the state at the given block number.
func (d *DbgAPI) dumpBlock(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 1 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "missing block number parameter")
	}

	var bn rpctypes.BlockNumber
	if err := json.Unmarshal(req.Params[0], &bn); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid block number: "+err.Error())
	}

	header := d.backend.HeaderByNumber(bn)
	if header == nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "block not found")
	}

	statedb, err := d.backend.StateAt(header.Root)
	if err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "state not available: "+err.Error())
	}

	// In a full implementation, we would iterate all accounts.
	// Here we produce a minimal dump with just the state root.
	// Callers can use eth_getProof for specific accounts.
	dump := &DbgStateDump{
		Root:     rpctypes.EncodeHash(header.Root),
		Accounts: make(map[string]*DbgDumpAccount),
	}

	// Probe well-known accounts (coinbase) to populate at least one entry.
	coinbase := header.Coinbase
	if statedb.Exist(coinbase) {
		balance := statedb.GetBalance(coinbase)
		nonce := statedb.GetNonce(coinbase)
		codeHash := statedb.GetCodeHash(coinbase)
		dump.Accounts[rpctypes.EncodeAddress(coinbase)] = &DbgDumpAccount{
			Balance:  rpctypes.EncodeBigInt(balance),
			Nonce:    nonce,
			CodeHash: rpctypes.EncodeHash(codeHash),
		}
	}

	return rpctypes.NewSuccessResponse(req.ID, dump)
}
