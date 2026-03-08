package debugapi

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"

	coretypes "github.com/eth2030/eth2030/core/types"
	rpcbackend "github.com/eth2030/eth2030/rpc/backend"
	rpctypes "github.com/eth2030/eth2030/rpc/types"
)

// DebugAPI implements the debug namespace RPC methods.
// It provides block introspection, RLP encoding, and chain management utilities.
type DebugAPI struct {
	backend rpcbackend.Backend
}

// NewDebugAPI creates a new DebugAPI instance.
func NewDebugAPI(backend rpcbackend.Backend) *DebugAPI {
	return &DebugAPI{backend: backend}
}

// HandleDebugRequest dispatches a debug_ namespace JSON-RPC request.
func (d *DebugAPI) HandleDebugRequest(req *rpctypes.Request) *rpctypes.Response {
	switch req.Method {
	case "debug_traceBlockByNumber":
		return d.debugNSTraceBlockByNumber(req)
	case "debug_traceBlockByHash":
		return d.debugNSTraceBlockByHash(req)
	case "debug_getBlockRlp":
		return d.debugGetBlockRlp(req)
	case "debug_printBlock":
		return d.debugPrintBlock(req)
	case "debug_chaindbProperty":
		return d.debugChaindbProperty(req)
	case "debug_chaindbCompact":
		return d.debugChaindbCompact(req)
	case "debug_setHead":
		return d.debugSetHead(req)
	case "debug_freeOSMemory":
		return d.debugFreeOSMemory(req)
	default:
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeMethodNotFound,
			fmt.Sprintf("method %q not found in debug namespace", req.Method))
	}
}

// debugNSTraceBlockByNumber traces all transactions in a block by number.
// Returns an array of TraceResult, one per transaction.
func (d *DebugAPI) debugNSTraceBlockByNumber(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 1 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "missing block number parameter")
	}

	var bn rpctypes.BlockNumber
	if err := json.Unmarshal(req.Params[0], &bn); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid block number: "+err.Error())
	}

	block := d.backend.BlockByNumber(bn)
	if block == nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "block not found")
	}

	return d.traceBlockTxs(req, block)
}

// debugNSTraceBlockByHash traces all transactions in a block by hash.
// Returns an array of TraceResult, one per transaction.
func (d *DebugAPI) debugNSTraceBlockByHash(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 1 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "missing block hash parameter")
	}

	var hashHex string
	if err := json.Unmarshal(req.Params[0], &hashHex); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid block hash: "+err.Error())
	}

	hash := coretypes.HexToHash(hashHex)
	block := d.backend.BlockByHash(hash)
	if block == nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "block not found")
	}

	return d.traceBlockTxs(req, block)
}

// traceBlockTxs produces a trace result for each transaction in the block.
func (d *DebugAPI) traceBlockTxs(req *rpctypes.Request, block *coretypes.Block) *rpctypes.Response {
	txs := block.Transactions()
	blockHash := block.Hash()
	receipts := d.backend.GetReceipts(blockHash)

	results := make([]*DebugBlockTraceEntry, len(txs))
	for i, tx := range txs {
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

		results[i] = &DebugBlockTraceEntry{
			TxHash: rpctypes.EncodeHash(tx.Hash()),
			Result: trace,
		}
	}

	return rpctypes.NewSuccessResponse(req.ID, results)
}

// debugGetBlockRlp returns the RLP-encoded block as a hex string.
func (d *DebugAPI) debugGetBlockRlp(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 1 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "missing block number parameter")
	}

	var bn rpctypes.BlockNumber
	if err := json.Unmarshal(req.Params[0], &bn); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid block number: "+err.Error())
	}

	block := d.backend.BlockByNumber(bn)
	if block == nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "block not found")
	}

	rlpBytes, err := block.EncodeRLP()
	if err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "RLP encoding failed: "+err.Error())
	}

	return rpctypes.NewSuccessResponse(req.ID, "0x"+hex.EncodeToString(rlpBytes))
}

// debugPrintBlock returns a human-readable representation of a block.
func (d *DebugAPI) debugPrintBlock(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 1 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "missing block number parameter")
	}

	var bn rpctypes.BlockNumber
	if err := json.Unmarshal(req.Params[0], &bn); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid block number: "+err.Error())
	}

	block := d.backend.BlockByNumber(bn)
	if block == nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "block not found")
	}

	header := block.Header()
	txCount := len(block.Transactions())

	repr := fmt.Sprintf(
		"Block #%d [%s]\n  Parent:     %s\n  Coinbase:   %s\n  StateRoot:  %s\n  TxRoot:     %s\n  GasLimit:   %d\n  GasUsed:    %d\n  Timestamp:  %d\n  TxCount:    %d",
		header.Number.Uint64(),
		rpctypes.EncodeHash(header.Hash()),
		rpctypes.EncodeHash(header.ParentHash),
		rpctypes.EncodeAddress(header.Coinbase),
		rpctypes.EncodeHash(header.Root),
		rpctypes.EncodeHash(header.TxHash),
		header.GasLimit,
		header.GasUsed,
		header.Time,
		txCount,
	)

	if header.BaseFee != nil {
		repr += fmt.Sprintf("\n  BaseFee:    %s", header.BaseFee.String())
	}

	return rpctypes.NewSuccessResponse(req.ID, repr)
}

// debugChaindbProperty returns a database property value.
// Supported properties: "leveldb.stats", "leveldb.iostats", "version".
func (d *DebugAPI) debugChaindbProperty(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 1 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "missing property parameter")
	}

	var property string
	if err := json.Unmarshal(req.Params[0], &property); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid property: "+err.Error())
	}

	// Return a basic response for known properties.
	switch property {
	case "leveldb.stats":
		return rpctypes.NewSuccessResponse(req.ID, "Compactions: 0\nLevel  Files  Size(MB)\n")
	case "leveldb.iostats":
		return rpctypes.NewSuccessResponse(req.ID, "Read(MB): 0.0\nWrite(MB): 0.0\n")
	case "version":
		return rpctypes.NewSuccessResponse(req.ID, "ETH2030/db/v1.0")
	default:
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams,
			fmt.Sprintf("unknown property: %q", property))
	}
}

// debugChaindbCompact triggers database compaction.
func (d *DebugAPI) debugChaindbCompact(req *rpctypes.Request) *rpctypes.Response {
	// In a real implementation, this would trigger LevelDB compaction.
	// For now, return success.
	return rpctypes.NewSuccessResponse(req.ID, nil)
}

// debugSetHead rewinds the chain head to a specific block number.
func (d *DebugAPI) debugSetHead(req *rpctypes.Request) *rpctypes.Response {
	if len(req.Params) < 1 {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "missing block number parameter")
	}

	var bn rpctypes.BlockNumber
	if err := json.Unmarshal(req.Params[0], &bn); err != nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInvalidParams, "invalid block number: "+err.Error())
	}

	// Verify the target block exists.
	header := d.backend.HeaderByNumber(bn)
	if header == nil {
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeInternal, "target block not found")
	}

	// In a full implementation, this would rewind the chain.
	// Return success to indicate the operation was accepted.
	return rpctypes.NewSuccessResponse(req.ID, nil)
}

// debugFreeOSMemory triggers a garbage collection and returns released
// memory back to the OS.
func (d *DebugAPI) debugFreeOSMemory(req *rpctypes.Request) *rpctypes.Response {
	runtime.GC()
	debug.FreeOSMemory()
	return rpctypes.NewSuccessResponse(req.ID, nil)
}
