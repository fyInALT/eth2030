// batch_handler_ext.go extends batch handling with validation, statistics,
// and notification batching support. Independent types delegate to rpc/batch.
package rpc

import (
	"encoding/json"
	"fmt"
	"sync"

	rpcbatch "github.com/eth2030/eth2030/rpc/batch"
)

// Re-export independent batch types from rpc/batch.
type (
	BatchStats          = rpcbatch.BatchStats
	BatchStatsSnapshot  = rpcbatch.BatchStatsSnapshot
	BatchItemResult     = rpcbatch.BatchItemResult
	BatchRequestSummary = rpcbatch.BatchRequestSummary
)

// Re-export extended batch constants from rpc/batch.
const (
	MinBatchSize             = rpcbatch.MinBatchSize
	MaxNotificationBatchSize = rpcbatch.MaxNotificationBatchSize
	DefaultBatchTimeout      = rpcbatch.DefaultBatchTimeout
)

// Re-export batch utility functions from rpc/batch.
var (
	SummarizeBatch = rpcbatch.SummarizeBatch
	SplitBatch     = rpcbatch.SplitBatch
)

// BatchValidator checks the structural validity of a batch request.
type BatchValidator struct {
	maxSize int
}

// NewBatchValidator creates a validator with the given maximum batch size.
func NewBatchValidator(maxSize int) *BatchValidator {
	if maxSize <= 0 {
		maxSize = MaxBatchSize
	}
	return &BatchValidator{maxSize: maxSize}
}

// Validate checks the structural validity of a parsed batch. Returns a
// slice of per-item validation errors (nil entries mean the item is valid).
func (v *BatchValidator) Validate(requests []BatchRequest) []error {
	errs := make([]error, len(requests))
	for i, req := range requests {
		if req.JSONRPC != "2.0" {
			errs[i] = fmt.Errorf("invalid jsonrpc version: %q", req.JSONRPC)
			continue
		}
		if req.Method == "" {
			errs[i] = fmt.Errorf("method is required")
			continue
		}
	}
	return errs
}

// ValidateBatchSize checks the batch size constraints.
func (v *BatchValidator) ValidateBatchSize(count int) error {
	if count < MinBatchSize {
		return ErrBatchEmpty
	}
	if count > v.maxSize {
		return fmt.Errorf("rpc: batch exceeds maximum size of %d", v.maxSize)
	}
	return nil
}

// NotificationBatch accumulates subscription notifications and flushes
// them as a batch when the batch is full or a flush interval is reached.
type NotificationBatch struct {
	mu    sync.Mutex
	items []json.RawMessage
	limit int
}

// NewNotificationBatch creates a new batch with the given flush limit.
func NewNotificationBatch(limit int) *NotificationBatch {
	if limit <= 0 {
		limit = MaxNotificationBatchSize
	}
	return &NotificationBatch{
		items: make([]json.RawMessage, 0, limit),
		limit: limit,
	}
}

// Add appends a notification to the batch. Returns the serialized batch
// if the limit is reached, or nil if more items can be added.
func (nb *NotificationBatch) Add(notification interface{}) []byte {
	data, err := json.Marshal(notification)
	if err != nil {
		return nil
	}
	nb.mu.Lock()
	defer nb.mu.Unlock()
	nb.items = append(nb.items, json.RawMessage(data))
	if len(nb.items) >= nb.limit {
		return nb.flushLocked()
	}
	return nil
}

// Flush forces the batch to be serialized and returned, even if not full.
func (nb *NotificationBatch) Flush() []byte {
	nb.mu.Lock()
	defer nb.mu.Unlock()
	return nb.flushLocked()
}

func (nb *NotificationBatch) flushLocked() []byte {
	if len(nb.items) == 0 {
		return nil
	}
	data, _ := json.Marshal(nb.items)
	nb.items = nb.items[:0]
	return data
}

// Len returns the number of buffered notifications.
func (nb *NotificationBatch) Len() int {
	nb.mu.Lock()
	defer nb.mu.Unlock()
	return len(nb.items)
}

// ExtendedBatchHandler extends BatchHandler with validation, statistics,
// and notification batching support.
type ExtendedBatchHandler struct {
	api         *EthAPI
	parallelism int
	stats       BatchStats
	validator   *BatchValidator
}

// NewExtendedBatchHandler creates an extended batch handler with the given API.
func NewExtendedBatchHandler(api *EthAPI) *ExtendedBatchHandler {
	return &ExtendedBatchHandler{
		api:         api,
		parallelism: DefaultParallelism,
		validator:   NewBatchValidator(MaxBatchSize),
	}
}

// SetParallelism sets the concurrency limit for parallel execution.
func (bh *ExtendedBatchHandler) SetParallelism(n int) {
	if n < 1 {
		n = 1
	}
	bh.parallelism = n
}

// HandleBatchValidated parses, validates, and executes a batch request.
func (bh *ExtendedBatchHandler) HandleBatchValidated(body []byte) ([]BatchResponse, error) {
	requests, err := parseBatchRequests(body)
	if err != nil {
		return nil, err
	}
	if err := bh.validator.ValidateBatchSize(len(requests)); err != nil {
		return nil, err
	}

	itemErrors := bh.validator.Validate(requests)

	bh.stats.TotalBatches.Add(1)
	bh.stats.TotalRequests.Add(uint64(len(requests)))
	current := uint64(len(requests))
	for {
		largest := bh.stats.LargestBatch.Load()
		if current <= largest {
			break
		}
		if bh.stats.LargestBatch.CompareAndSwap(largest, current) {
			break
		}
	}

	return bh.executeWithValidation(requests, itemErrors), nil
}

// executeWithValidation runs valid batch items in parallel and returns
// pre-built error responses for invalid ones.
func (bh *ExtendedBatchHandler) executeWithValidation(
	requests []BatchRequest,
	itemErrors []error,
) []BatchResponse {
	n := len(requests)
	responses := make([]BatchResponse, n)

	sem := make(chan struct{}, bh.parallelism)
	var wg sync.WaitGroup

	for i, req := range requests {
		if itemErrors[i] != nil {
			responses[i] = BatchResponse{
				JSONRPC: "2.0",
				Error:   &RPCError{Code: ErrCodeInvalidRequest, Message: itemErrors[i].Error()},
				ID:      req.ID,
			}
			bh.stats.TotalErrors.Add(1)
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, r BatchRequest) {
			defer wg.Done()
			defer func() { <-sem }()

			apiReq := &Request{
				JSONRPC: r.JSONRPC,
				Method:  r.Method,
				Params:  r.Params,
				ID:      r.ID,
			}
			resp := bh.api.HandleRequest(apiReq)
			responses[idx] = BatchResponse{
				JSONRPC: resp.JSONRPC,
				Result:  resp.Result,
				Error:   resp.Error,
				ID:      resp.ID,
			}
			if resp.Error != nil {
				bh.stats.TotalErrors.Add(1)
			}
		}(i, req)
	}

	wg.Wait()
	bh.stats.ParallelBatches.Add(1)
	return responses
}

// Stats returns the current batch processing statistics.
func (bh *ExtendedBatchHandler) Stats() BatchStatsSnapshot {
	return bh.stats.Snapshot()
}
