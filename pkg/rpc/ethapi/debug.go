package ethapi

import (
	"encoding/json"
	"math/big"

	"github.com/eth2030/eth2030/core/types"
)

// debugTraceTransaction implements debug_traceTransaction.
func (api *EthAPI) debugTraceTransaction(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing transaction hash")
	}

	var txHashHex string
	if err := json.Unmarshal(req.Params[0], &txHashHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid tx hash: "+err.Error())
	}

	txHash := types.HexToHash(txHashHex)

	tracer, err := api.backend.TraceTransaction(txHash)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}

	structLogs := make([]StructLog, len(tracer.Logs))
	for i, entry := range tracer.Logs {
		stackHex := make([]string, len(entry.Stack))
		for j, val := range entry.Stack {
			stackHex[j] = "0x" + val.Text(16)
		}
		structLogs[i] = StructLog{
			PC:      entry.Pc,
			Op:      entry.Op.String(),
			Gas:     entry.Gas,
			GasCost: entry.GasCost,
			Depth:   entry.Depth,
			Stack:   stackHex,
		}
	}

	failed := tracer.Error() != nil
	retVal := ""
	if out := tracer.Output(); len(out) > 0 {
		retVal = encodeBytes(out)
	}

	result := &TraceResult{
		Gas:         tracer.GasUsed(),
		Failed:      failed,
		ReturnValue: retVal,
		StructLogs:  structLogs,
	}

	return successResponse(req.ID, result)
}

// getAccountRange implements debug_getAccountRange (for snap sync debugging).
func (api *EthAPI) getAccountRange(req *Request) *Response {
	return errorResponse(req.ID, ErrCodeMethodNotFound, "debug_getAccountRange not yet implemented")
}

// debugTraceBlockByNumber implements debug_traceBlockByNumber.
func (api *EthAPI) debugTraceBlockByNumber(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing block number parameter")
	}

	var bn BlockNumber
	if err := json.Unmarshal(req.Params[0], &bn); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid block number: "+err.Error())
	}

	block := api.backend.BlockByNumber(bn)
	if block == nil {
		return errorResponse(req.ID, ErrCodeInternal, "block not found")
	}

	return api.traceBlock(req, block)
}

// debugTraceBlockByHash implements debug_traceBlockByHash.
func (api *EthAPI) debugTraceBlockByHash(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing block hash parameter")
	}

	var hashHex string
	if err := json.Unmarshal(req.Params[0], &hashHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid block hash: "+err.Error())
	}

	hash := types.HexToHash(hashHex)
	block := api.backend.BlockByHash(hash)
	if block == nil {
		return errorResponse(req.ID, ErrCodeInternal, "block not found")
	}

	return api.traceBlock(req, block)
}

// traceBlock produces a trace result for each transaction in the block.
func (api *EthAPI) traceBlock(req *Request, block *types.Block) *Response {
	txs := block.Transactions()
	blockHash := block.Hash()

	receipts := api.backend.GetReceipts(blockHash)

	results := make([]*BlockTraceResult, len(txs))
	for i, tx := range txs {
		trace := &TraceResult{
			Gas:         tx.Gas(),
			Failed:      false,
			ReturnValue: "",
			StructLogs:  []StructLog{},
		}

		// If receipts are available, use actual gas used and status.
		if i < len(receipts) {
			trace.Gas = receipts[i].GasUsed
			trace.Failed = receipts[i].Status == types.ReceiptStatusFailed
		}

		results[i] = &BlockTraceResult{
			TxHash: encodeHash(tx.Hash()),
			Result: trace,
		}
	}

	return successResponse(req.ID, results)
}

// debugTraceCall implements debug_traceCall.
func (api *EthAPI) debugTraceCall(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing call arguments")
	}

	var args CallArgs
	if err := json.Unmarshal(req.Params[0], &args); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid call arguments: "+err.Error())
	}

	bn := LatestBlockNumber
	if len(req.Params) > 1 {
		if err := json.Unmarshal(req.Params[1], &bn); err != nil {
			return errorResponse(req.ID, ErrCodeInvalidParams, "invalid block number: "+err.Error())
		}
	}

	from := types.Address{}
	if args.From != nil {
		from = types.HexToAddress(*args.From)
	}
	var to *types.Address
	if args.To != nil {
		addr := types.HexToAddress(*args.To)
		to = &addr
	}
	gas := uint64(50_000_000)
	if args.Gas != nil {
		gas = parseHexUint64(*args.Gas)
	}
	value := new(big.Int)
	if args.Value != nil {
		value = parseHexBigInt(*args.Value)
	}
	data := args.GetData()

	result, gasUsed, err := api.backend.EVMCall(from, to, data, gas, value, bn)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, "execution error: "+err.Error())
	}

	trace := &TraceResult{
		Gas:         gasUsed,
		Failed:      false,
		ReturnValue: encodeBytes(result),
		StructLogs:  []StructLog{},
	}

	return successResponse(req.ID, trace)
}
