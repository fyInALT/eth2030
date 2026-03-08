// batch_handler_ext.go re-exports extended batch types from rpc/batch for
// backward compatibility.
package rpc

import (
	rpcbatch "github.com/eth2030/eth2030/rpc/batch"
)

// Re-export extended batch types from rpc/batch.
type (
	BatchStats           = rpcbatch.BatchStats
	BatchStatsSnapshot   = rpcbatch.BatchStatsSnapshot
	BatchItemResult      = rpcbatch.BatchItemResult
	BatchRequestSummary  = rpcbatch.BatchRequestSummary
	BatchValidator       = rpcbatch.BatchValidator
	NotificationBatch    = rpcbatch.NotificationBatch
	ExtendedBatchHandler = rpcbatch.ExtendedBatchHandler
)

// Re-export extended batch constants from rpc/batch.
const (
	MinBatchSize             = rpcbatch.MinBatchSize
	MaxNotificationBatchSize = rpcbatch.MaxNotificationBatchSize
	DefaultBatchTimeout      = rpcbatch.DefaultBatchTimeout
)

// Re-export batch utility functions from rpc/batch.
var (
	SummarizeBatch       = rpcbatch.SummarizeBatch
	SplitBatch           = rpcbatch.SplitBatch
	NewBatchValidator    = rpcbatch.NewBatchValidator
	NewNotificationBatch = rpcbatch.NewNotificationBatch
)

// NewExtendedBatchHandler creates an extended batch handler.
// EthAPI implements rpcbatch.RequestHandler.
func NewExtendedBatchHandler(api *EthAPI) *ExtendedBatchHandler {
	return rpcbatch.NewExtendedBatchHandler(api)
}
