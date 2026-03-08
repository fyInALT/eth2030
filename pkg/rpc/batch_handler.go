// batch_handler.go implements BatchHandler for processing JSON-RPC batch
// requests. Independent types are re-exported from rpc/batch.
package rpc

import (
	"encoding/json"
	"fmt"
	"sync"

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

// Re-export batch request type.
type BatchRequest = rpcbatch.BatchRequest

// BatchResponse represents a single response within a JSON-RPC batch response.
type BatchResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
	ID      json.RawMessage `json:"id"`
}

// BatchHandler processes JSON-RPC batch requests.
type BatchHandler struct {
	api          *EthAPI
	adminAPI     *AdminDispatchAPI
	parallelism  int
	maxBatchSize int
}

// NewBatchHandler creates a new batch handler that dispatches to the given API.
func NewBatchHandler(api *EthAPI) *BatchHandler {
	return &BatchHandler{
		api:          api,
		parallelism:  DefaultParallelism,
		maxBatchSize: MaxBatchSize,
	}
}

// SetAdminBackend wires an AdminBackend for admin_ method dispatch in batches.
func (bh *BatchHandler) SetAdminBackend(b AdminBackend) {
	bh.adminAPI = NewAdminDispatchAPI(b)
}

// SetMaxBatchSize sets the maximum supported batch size.
func (bh *BatchHandler) SetMaxBatchSize(n int) {
	if n < 1 {
		n = 1
	}
	bh.maxBatchSize = n
}

// SetParallelism sets the maximum number of goroutines used for parallel
// batch execution.
func (bh *BatchHandler) SetParallelism(n int) {
	if n < 1 {
		n = 1
	}
	bh.parallelism = n
}

// HandleBatch parses a raw JSON body as a batch request and returns the
// batch response.
func (bh *BatchHandler) HandleBatch(body []byte) ([]BatchResponse, error) {
	requests, err := parseBatchRequests(body)
	if err != nil {
		return nil, err
	}
	if len(requests) == 0 {
		return nil, ErrBatchEmpty
	}
	if len(requests) > bh.maxBatchSize {
		if bh.maxBatchSize == MaxBatchSize {
			return nil, ErrBatchTooLarge
		}
		return nil, fmt.Errorf("rpc: batch exceeds maximum size of %d", bh.maxBatchSize)
	}
	return bh.ExecuteParallel(requests), nil
}

// ExecuteParallel executes a slice of batch requests in parallel with bounded
// concurrency. Results are returned in the same order as the input requests.
func (bh *BatchHandler) ExecuteParallel(requests []BatchRequest) []BatchResponse {
	n := len(requests)
	responses := make([]BatchResponse, n)

	sem := make(chan struct{}, bh.parallelism)
	var wg sync.WaitGroup

	for i, req := range requests {
		wg.Add(1)
		sem <- struct{}{} // acquire
		go func(idx int, r BatchRequest) {
			defer wg.Done()
			defer func() { <-sem }() // release
			responses[idx] = bh.executeOne(r)
		}(i, req)
	}
	wg.Wait()
	return responses
}

// executeOne dispatches a single BatchRequest to the API and wraps the
// result into a BatchResponse.
func (bh *BatchHandler) executeOne(req BatchRequest) BatchResponse {
	if req.JSONRPC != "2.0" {
		return BatchResponse{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: ErrCodeInvalidRequest, Message: "invalid jsonrpc version"},
			ID:      req.ID,
		}
	}
	if req.Method == "" {
		return BatchResponse{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: ErrCodeInvalidRequest, Message: "method is required"},
			ID:      req.ID,
		}
	}

	apiReq := &Request{
		JSONRPC: req.JSONRPC,
		Method:  req.Method,
		Params:  req.Params,
		ID:      req.ID,
	}
	var resp *Response
	if isAdminMethod(req.Method) && bh.adminAPI != nil {
		resp = bh.adminAPI.HandleAdminRequest(apiReq)
	} else {
		resp = bh.api.HandleRequest(apiReq)
	}

	return BatchResponse{
		JSONRPC: resp.JSONRPC,
		Result:  resp.Result,
		Error:   resp.Error,
		ID:      resp.ID,
	}
}

// MarshalBatchResponse serializes a batch of responses to JSON.
func MarshalBatchResponse(responses []BatchResponse) ([]byte, error) {
	return json.Marshal(responses)
}

// parseBatchRequests parses a JSON byte slice as an array of BatchRequest.
func parseBatchRequests(body []byte) ([]BatchRequest, error) {
	trimmed := rpcbatch.TrimWhitespace(body)
	if len(trimmed) == 0 || trimmed[0] != '[' {
		return nil, ErrNotBatch
	}

	var requests []BatchRequest
	if err := json.Unmarshal(body, &requests); err != nil {
		return nil, fmt.Errorf("rpc: invalid JSON in batch: %w", err)
	}
	return requests, nil
}

// trimWhitespace returns body with leading whitespace removed.
// Kept for internal backward compatibility.
func trimWhitespace(b []byte) []byte {
	return rpcbatch.TrimWhitespace(b)
}

// IsBatchRequest checks whether a JSON body is a batch request.
var IsBatchRequest = rpcbatch.IsBatchRequest
