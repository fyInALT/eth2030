// batch_handler.go re-exports BatchHandler and related types from rpc/batch
// for backward compatibility.
package rpc

import (
	rpcbatch "github.com/eth2030/eth2030/rpc/batch"
)

// Re-export batch errors.
var (
	ErrBatchEmpty    = rpcbatch.ErrBatchEmpty
	ErrBatchTooLarge = rpcbatch.ErrBatchTooLarge
	ErrNotBatch      = rpcbatch.ErrNotBatch
)

// Re-export batch constants.
const (
	MaxBatchSize       = rpcbatch.MaxBatchSize
	DefaultParallelism = rpcbatch.DefaultParallelism
)

// Re-export batch types.
type (
	BatchRequest  = rpcbatch.BatchRequest
	BatchResponse = rpcbatch.BatchResponse
	BatchHandler  = rpcbatch.BatchHandler
)

// NewBatchHandler creates a new batch handler that dispatches to the given API.
// EthAPI implements rpcbatch.RequestHandler.
func NewBatchHandler(api *EthAPI) *BatchHandler {
	return rpcbatch.NewBatchHandler(api)
}

// MarshalBatchResponse serializes a batch of responses to JSON.
var MarshalBatchResponse = rpcbatch.MarshalBatchResponse

// IsBatchRequest checks whether a JSON body is a batch request.
var IsBatchRequest = rpcbatch.IsBatchRequest
