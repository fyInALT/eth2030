package engine

import (
	"encoding/json"
	"fmt"

	"github.com/eth2030/eth2030/core/rawdb"
	"github.com/eth2030/eth2030/core/types"
)

// ExecutionPayloadBodyV2 is the response body for engine_getPayloadBodiesByHash/RangeV2.
// It extends V1 (transactions + withdrawals) with a blockAccessList field (EIP-7928).
type ExecutionPayloadBodyV2 struct {
	Transactions    [][]byte      `json:"transactions"`
	Withdrawals     []*Withdrawal `json:"withdrawals"`
	BlockAccessList []byte        `json:"blockAccessList,omitempty"`
}

// GetPayloadBodiesByHashV2 returns payload bodies for the given block hashes,
// including the Block Access List per EIP-7928 §engine-api.
func (api *EngineAPI) GetPayloadBodiesByHashV2(hashes []types.Hash) ([]*ExecutionPayloadBodyV2, error) {
	backend, ok := api.backend.(*EngineBackend)
	if !ok {
		return nil, fmt.Errorf("payload bodies not supported by this backend")
	}
	backend.mu.RLock()
	defer backend.mu.RUnlock()

	// SPEC-5.5: determine head block number for retention window check.
	headNum := uint64(0)
	if head, ok := backend.blocks[backend.headHash]; ok {
		headNum = head.NumberU64()
	}

	results := make([]*ExecutionPayloadBodyV2, len(hashes))
	for i, h := range hashes {
		block, found := backend.blocks[h]
		if !found || !rawdb.IsBALRetained(headNum, block.NumberU64()) {
			results[i] = nil
			continue
		}
		body := blockToPayloadBodyV2(block)
		// Attach stored BAL if available.
		if bal, ok := backend.bals[h]; ok {
			balBytes, _ := json.Marshal(bal)
			body.BlockAccessList = balBytes
		}
		results[i] = body
	}
	return results, nil
}

// GetPayloadBodiesByRangeV2 returns payload bodies for a range of block numbers,
// including the Block Access List per EIP-7928 §engine-api.
func (api *EngineAPI) GetPayloadBodiesByRangeV2(start, count uint64) ([]*ExecutionPayloadBodyV2, error) {
	backend, ok := api.backend.(*EngineBackend)
	if !ok {
		return nil, fmt.Errorf("payload bodies not supported by this backend")
	}
	if count == 0 || count > 1024 {
		return nil, fmt.Errorf("count must be in [1, 1024], got %d", count)
	}
	backend.mu.RLock()
	defer backend.mu.RUnlock()

	// SPEC-5.5: determine head block number for retention window check.
	headNum := uint64(0)
	if head, ok := backend.blocks[backend.headHash]; ok {
		headNum = head.NumberU64()
	}

	results := make([]*ExecutionPayloadBodyV2, count)
	for i := uint64(0); i < count; i++ {
		num := start + i
		// Find block by number in stored blocks.
		var found *types.Block
		for _, b := range backend.blocks {
			if b.NumberU64() == num {
				found = b
				break
			}
		}
		if found == nil || !rawdb.IsBALRetained(headNum, num) {
			results[i] = nil
			continue
		}
		body := blockToPayloadBodyV2(found)
		if bal, ok := backend.bals[found.Hash()]; ok {
			balBytes, _ := json.Marshal(bal)
			body.BlockAccessList = balBytes
		}
		results[i] = body
	}
	return results, nil
}

func blockToPayloadBodyV2(block *types.Block) *ExecutionPayloadBodyV2 {
	txs := make([][]byte, 0, len(block.Transactions()))
	for _, tx := range block.Transactions() {
		enc, err := tx.EncodeRLP()
		if err == nil {
			txs = append(txs, enc)
		}
	}
	ws := make([]*Withdrawal, 0, len(block.Withdrawals()))
	for _, w := range block.Withdrawals() {
		if w != nil {
			ws = append(ws, &Withdrawal{
				Index:          w.Index,
				ValidatorIndex: w.ValidatorIndex,
				Address:        w.Address,
				Amount:         w.Amount,
			})
		}
	}
	return &ExecutionPayloadBodyV2{
		Transactions: txs,
		Withdrawals:  ws,
	}
}

// handleGetPayloadBodiesByHashV2 processes engine_getPayloadBodiesByHashV2.
func (api *EngineAPI) handleGetPayloadBodiesByHashV2(params []json.RawMessage) (any, *jsonrpcError) {
	if len(params) != 1 {
		return nil, &jsonrpcError{Code: InvalidParamsCode, Message: "expected 1 param"}
	}
	var hashes []types.Hash
	if err := json.Unmarshal(params[0], &hashes); err != nil {
		return nil, &jsonrpcError{Code: InvalidParamsCode, Message: fmt.Sprintf("invalid hashes: %v", err)}
	}
	result, err := api.GetPayloadBodiesByHashV2(hashes)
	if err != nil {
		return nil, engineErrorToRPC(err)
	}
	return result, nil
}

// handleGetPayloadBodiesByRangeV2 processes engine_getPayloadBodiesByRangeV2.
func (api *EngineAPI) handleGetPayloadBodiesByRangeV2(params []json.RawMessage) (any, *jsonrpcError) {
	if len(params) != 2 {
		return nil, &jsonrpcError{Code: InvalidParamsCode, Message: "expected 2 params"}
	}
	var start, count uint64
	if err := json.Unmarshal(params[0], &start); err != nil {
		return nil, &jsonrpcError{Code: InvalidParamsCode, Message: fmt.Sprintf("invalid start: %v", err)}
	}
	if err := json.Unmarshal(params[1], &count); err != nil {
		return nil, &jsonrpcError{Code: InvalidParamsCode, Message: fmt.Sprintf("invalid count: %v", err)}
	}
	result, err := api.GetPayloadBodiesByRangeV2(start, count)
	if err != nil {
		return nil, engineErrorToRPC(err)
	}
	return result, nil
}
