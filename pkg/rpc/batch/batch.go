// Package rpcbatch provides JSON-RPC batch request/response types,
// validation, statistics, and utility functions for the Ethereum JSON-RPC API.
package rpcbatch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

// Batch processing errors.
var (
	ErrBatchEmpty    = errors.New("rpc: empty batch")
	ErrBatchTooLarge = fmt.Errorf("rpc: batch exceeds maximum size of %d", MaxBatchSize)
	ErrNotBatch      = errors.New("rpc: request is not a JSON array")
)

// Batch processing constants.
const (
	// MaxBatchSize is the maximum number of requests in a single batch.
	MaxBatchSize = 100

	// DefaultParallelism is the default number of goroutines for parallel execution.
	DefaultParallelism = 16

	// MinBatchSize is the minimum number of requests for a valid batch.
	MinBatchSize = 1

	// MaxNotificationBatchSize is the max number of notifications batched together.
	MaxNotificationBatchSize = 50

	// DefaultBatchTimeout is the default timeout per batch item in milliseconds.
	DefaultBatchTimeout = 5000
)

// BatchRequest represents a single request within a JSON-RPC batch.
type BatchRequest struct {
	JSONRPC string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
	ID      json.RawMessage   `json:"id"`
}

// IsBatchRequest returns true if the request body is a JSON array (batch).
func IsBatchRequest(body []byte) bool {
	trimmed := bytes.TrimSpace(body)
	return len(trimmed) > 0 && trimmed[0] == '['
}

// ParseBatchRequests decodes a JSON array of requests.
func ParseBatchRequests(body []byte) ([]BatchRequest, error) {
	if !IsBatchRequest(body) {
		return nil, ErrNotBatch
	}
	var reqs []BatchRequest
	if err := json.Unmarshal(body, &reqs); err != nil {
		return nil, fmt.Errorf("rpcbatch: invalid batch JSON: %w", err)
	}
	if len(reqs) == 0 {
		return nil, ErrBatchEmpty
	}
	if len(reqs) > MaxBatchSize {
		return nil, ErrBatchTooLarge
	}
	return reqs, nil
}

// TrimWhitespace returns body with leading whitespace removed.
func TrimWhitespace(b []byte) []byte {
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\t' || b[0] == '\r' || b[0] == '\n') {
		b = b[1:]
	}
	return b
}

// BatchStats tracks statistics about batch processing for diagnostics.
type BatchStats struct {
	TotalBatches    atomic.Uint64
	TotalRequests   atomic.Uint64
	TotalErrors     atomic.Uint64
	LargestBatch    atomic.Uint64
	ParallelBatches atomic.Uint64
}

// Snapshot returns a point-in-time snapshot of the batch statistics.
func (s *BatchStats) Snapshot() BatchStatsSnapshot {
	return BatchStatsSnapshot{
		TotalBatches:    s.TotalBatches.Load(),
		TotalRequests:   s.TotalRequests.Load(),
		TotalErrors:     s.TotalErrors.Load(),
		LargestBatch:    s.LargestBatch.Load(),
		ParallelBatches: s.ParallelBatches.Load(),
	}
}

// BatchStatsSnapshot is a non-atomic copy of BatchStats for serialization.
type BatchStatsSnapshot struct {
	TotalBatches    uint64 `json:"totalBatches"`
	TotalRequests   uint64 `json:"totalRequests"`
	TotalErrors     uint64 `json:"totalErrors"`
	LargestBatch    uint64 `json:"largestBatch"`
	ParallelBatches uint64 `json:"parallelBatches"`
}

// BatchItemResult contains the result of executing a single batch item.
type BatchItemResult struct {
	Index int
	Error bool
}

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

// MaxSize returns the configured maximum batch size.
func (v *BatchValidator) MaxSize() int {
	return v.maxSize
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

// BatchRequestSummary provides a human-readable summary of a batch request.
type BatchRequestSummary struct {
	Count   int      `json:"count"`
	Methods []string `json:"methods"`
}

// SummarizeBatch extracts method names from a batch for diagnostic logging.
func SummarizeBatch(requests []BatchRequest) BatchRequestSummary {
	methods := make([]string, len(requests))
	for i, req := range requests {
		methods[i] = req.Method
	}
	return BatchRequestSummary{
		Count:   len(requests),
		Methods: methods,
	}
}

// SplitBatch splits a large batch into smaller sub-batches of at most chunkSize.
func SplitBatch(requests []BatchRequest, chunkSize int) [][]BatchRequest {
	if chunkSize <= 0 {
		chunkSize = MaxBatchSize
	}
	var chunks [][]BatchRequest
	for i := 0; i < len(requests); i += chunkSize {
		end := i + chunkSize
		if end > len(requests) {
			end = len(requests)
		}
		chunks = append(chunks, requests[i:end])
	}
	return chunks
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
// Returns nil if the batch is empty.
func (nb *NotificationBatch) Flush() []byte {
	nb.mu.Lock()
	defer nb.mu.Unlock()
	return nb.flushLocked()
}

// flushLocked serializes all accumulated items as a JSON array and resets.
// Caller must hold nb.mu.
func (nb *NotificationBatch) flushLocked() []byte {
	if len(nb.items) == 0 {
		return nil
	}
	data, _ := json.Marshal(nb.items)
	nb.items = nb.items[:0]
	return data
}

// Limit returns the flush limit configured for this batch.
func (nb *NotificationBatch) Limit() int {
	return nb.limit
}

// Len returns the number of buffered notifications.
func (nb *NotificationBatch) Len() int {
	nb.mu.Lock()
	defer nb.mu.Unlock()
	return len(nb.items)
}
