package engine

import (
	"encoding/json"
	"fmt"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/engine/backendapi"
	enginepayload "github.com/eth2030/eth2030/engine/payload"
)

// ExecutionPayloadBodyV1 is the response body for engine_getPayloadBodiesByHash/RangeV1.
type ExecutionPayloadBodyV1 = enginepayload.ExecutionPayloadBodyV1

// ExecutionPayloadBodyV2 is the response body for engine_getPayloadBodiesByHash/RangeV2.
// It extends V1 (transactions + withdrawals) with a blockAccessList field (EIP-7928).
type ExecutionPayloadBodyV2 = enginepayload.ExecutionPayloadBodyV2

// GetPayloadBodiesByHashV1 returns payload bodies for the given block hashes (Shanghai).
func (api *EngineAPI) GetPayloadBodiesByHashV1(hashes []types.Hash) ([]*ExecutionPayloadBodyV1, error) {
	v2s, err := api.GetPayloadBodiesByHashV2(hashes)
	if err != nil {
		return nil, err
	}
	return enginepayload.V2SliceToV1(v2s), nil
}

// GetPayloadBodiesByRangeV1 returns payload bodies for a range of block numbers (Shanghai).
func (api *EngineAPI) GetPayloadBodiesByRangeV1(start, count uint64) ([]*ExecutionPayloadBodyV1, error) {
	v2s, err := api.GetPayloadBodiesByRangeV2(start, count)
	if err != nil {
		return nil, err
	}
	return enginepayload.V2SliceToV1(v2s), nil
}

// GetPayloadBodiesByHashV2 returns payload bodies for the given block hashes,
// including the Block Access List per EIP-7928 §engine-api.
func (api *EngineAPI) GetPayloadBodiesByHashV2(hashes []types.Hash) ([]*ExecutionPayloadBodyV2, error) {
	pb, ok := api.backend.(backendapi.PayloadBodiesBackend)
	if !ok {
		return nil, fmt.Errorf("payload bodies not supported by this backend")
	}
	return pb.GetPayloadBodiesByHash(hashes)
}

// GetPayloadBodiesByRangeV2 returns payload bodies for a range of block numbers,
// including the Block Access List per EIP-7928 §engine-api.
func (api *EngineAPI) GetPayloadBodiesByRangeV2(start, count uint64) ([]*ExecutionPayloadBodyV2, error) {
	if count == 0 || count > 1024 {
		return nil, fmt.Errorf("count must be in [1, 1024], got %d", count)
	}
	pb, ok := api.backend.(backendapi.PayloadBodiesBackend)
	if !ok {
		return nil, fmt.Errorf("payload bodies not supported by this backend")
	}
	return pb.GetPayloadBodiesByRange(start, count)
}

// blockToPayloadBodyV2 is a package-level wrapper for tests that access this
// function directly (in the engine package).
func blockToPayloadBodyV2(block *types.Block) *ExecutionPayloadBodyV2 {
	return enginepayload.BlockToPayloadBodyV2(block)
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

// handleGetPayloadBodiesByHashV1 processes engine_getPayloadBodiesByHashV1.
func (api *EngineAPI) handleGetPayloadBodiesByHashV1(params []json.RawMessage) (any, *jsonrpcError) {
	if len(params) != 1 {
		return nil, &jsonrpcError{Code: InvalidParamsCode, Message: "expected 1 param"}
	}
	var hashes []types.Hash
	if err := json.Unmarshal(params[0], &hashes); err != nil {
		return nil, &jsonrpcError{Code: InvalidParamsCode, Message: fmt.Sprintf("invalid hashes: %v", err)}
	}
	result, err := api.GetPayloadBodiesByHashV1(hashes)
	if err != nil {
		return nil, engineErrorToRPC(err)
	}
	return result, nil
}

// handleGetPayloadBodiesByRangeV1 processes engine_getPayloadBodiesByRangeV1.
func (api *EngineAPI) handleGetPayloadBodiesByRangeV1(params []json.RawMessage) (any, *jsonrpcError) {
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
	result, err := api.GetPayloadBodiesByRangeV1(start, count)
	if err != nil {
		return nil, engineErrorToRPC(err)
	}
	return result, nil
}
