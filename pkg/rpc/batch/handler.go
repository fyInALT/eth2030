// handler.go implements BatchHandler and ExtendedBatchHandler for parallel
// JSON-RPC batch request execution.
package rpcbatch

import (
	"encoding/json"
	"fmt"
	"sync"

	rpctypes "github.com/eth2030/eth2030/rpc/types"
)

// RequestHandler dispatches a single JSON-RPC request and returns a response.
// *ethapi.EthAPI satisfies this interface.
type RequestHandler interface {
	HandleRequest(req *rpctypes.Request) *rpctypes.Response
}

// AdminRequestHandler dispatches an admin_ JSON-RPC request.
// *rpc.AdminDispatchAPI satisfies this interface.
type AdminRequestHandler interface {
	HandleAdminRequest(req *rpctypes.Request) *rpctypes.Response
}

// NetRequestHandler dispatches net_ namespace JSON-RPC requests.
// *netapi.API satisfies this interface.
type NetRequestHandler interface {
	HandleNetRequest(req *rpctypes.Request) *rpctypes.Response
}

// BeaconRequestHandler dispatches beacon_ namespace JSON-RPC requests.
// *beaconapi.BeaconAPI satisfies this interface.
type BeaconRequestHandler interface {
	HandleBeaconRequest(req *rpctypes.Request) *rpctypes.Response
}

// BatchResponse represents a single response within a JSON-RPC batch.
type BatchResponse struct {
	JSONRPC string             `json:"jsonrpc"`
	Result  interface{}        `json:"result,omitempty"`
	Error   *rpctypes.RPCError `json:"error,omitempty"`
	ID      json.RawMessage    `json:"id"`
}

// BatchHandler processes JSON-RPC batch requests by dispatching each item
// to the underlying RequestHandler, optionally also serving admin_, net_,
// and beacon_ methods.
type BatchHandler struct {
	api          RequestHandler
	adminAPI     AdminRequestHandler
	netAPI       NetRequestHandler
	beaconAPI    BeaconRequestHandler
	parallelism  int
	maxBatchSize int
}

// NewBatchHandler creates a new batch handler for the given RequestHandler.
func NewBatchHandler(api RequestHandler) *BatchHandler {
	return &BatchHandler{
		api:          api,
		parallelism:  DefaultParallelism,
		maxBatchSize: MaxBatchSize,
	}
}

// SetAdminHandler wires an AdminRequestHandler so admin_ methods are dispatched.
func (bh *BatchHandler) SetAdminHandler(h AdminRequestHandler) {
	bh.adminAPI = h
}

// SetNetHandler wires a NetRequestHandler so net_ methods are dispatched.
func (bh *BatchHandler) SetNetHandler(h NetRequestHandler) {
	bh.netAPI = h
}

// SetBeaconHandler wires a BeaconRequestHandler so beacon_ methods are dispatched.
func (bh *BatchHandler) SetBeaconHandler(h BeaconRequestHandler) {
	bh.beaconAPI = h
}

// SetMaxBatchSize sets the maximum number of items in a single batch.
func (bh *BatchHandler) SetMaxBatchSize(n int) {
	if n < 1 {
		n = 1
	}
	bh.maxBatchSize = n
}

// SetParallelism sets the number of goroutines for parallel execution.
func (bh *BatchHandler) SetParallelism(n int) {
	if n < 1 {
		n = 1
	}
	bh.parallelism = n
}

// Parallelism returns the current concurrency limit.
func (bh *BatchHandler) Parallelism() int {
	return bh.parallelism
}

// HandleBatch parses body as a JSON-RPC batch and executes it.
func (bh *BatchHandler) HandleBatch(body []byte) ([]BatchResponse, error) {
	requests, err := ParseBatchRequests(body)
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

// ExecuteParallel executes batch items in parallel and returns results in order.
func (bh *BatchHandler) ExecuteParallel(requests []BatchRequest) []BatchResponse {
	n := len(requests)
	responses := make([]BatchResponse, n)
	sem := make(chan struct{}, bh.parallelism)
	var wg sync.WaitGroup
	for i, req := range requests {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, r BatchRequest) {
			defer wg.Done()
			defer func() { <-sem }()
			responses[idx] = bh.executeOne(r)
		}(i, req)
	}
	wg.Wait()
	return responses
}

func isAdminMethod(method string) bool {
	return len(method) > 6 && method[:6] == "admin_"
}

func isNetMethod(method string) bool {
	return len(method) > 4 && method[:4] == "net_"
}

func isBeaconMethod(method string) bool {
	return len(method) > 7 && method[:7] == "beacon_"
}

func (bh *BatchHandler) executeOne(req BatchRequest) BatchResponse {
	if req.JSONRPC != "2.0" {
		return BatchResponse{
			JSONRPC: "2.0",
			Error:   &rpctypes.RPCError{Code: rpctypes.ErrCodeInvalidRequest, Message: "invalid jsonrpc version"},
			ID:      req.ID,
		}
	}
	if req.Method == "" {
		return BatchResponse{
			JSONRPC: "2.0",
			Error:   &rpctypes.RPCError{Code: rpctypes.ErrCodeInvalidRequest, Message: "method is required"},
			ID:      req.ID,
		}
	}
	apiReq := &rpctypes.Request{
		JSONRPC: req.JSONRPC,
		Method:  req.Method,
		Params:  req.Params,
		ID:      req.ID,
	}
	var resp *rpctypes.Response
	switch {
	case isAdminMethod(req.Method) && bh.adminAPI != nil:
		resp = bh.adminAPI.HandleAdminRequest(apiReq)
	case isNetMethod(req.Method) && bh.netAPI != nil:
		resp = bh.netAPI.HandleNetRequest(apiReq)
	case isBeaconMethod(req.Method) && bh.beaconAPI != nil:
		resp = bh.beaconAPI.HandleBeaconRequest(apiReq)
	default:
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

// ExtendedBatchHandler adds validation, statistics, and parallel execution
// on top of basic batch handling.
type ExtendedBatchHandler struct {
	api         RequestHandler
	parallelism int
	stats       BatchStats
	validator   *BatchValidator
}

// NewExtendedBatchHandler creates an extended batch handler.
func NewExtendedBatchHandler(api RequestHandler) *ExtendedBatchHandler {
	return &ExtendedBatchHandler{
		api:         api,
		parallelism: DefaultParallelism,
		validator:   NewBatchValidator(MaxBatchSize),
	}
}

// SetParallelism sets the concurrency limit.
func (bh *ExtendedBatchHandler) SetParallelism(n int) {
	if n < 1 {
		n = 1
	}
	bh.parallelism = n
}

// Parallelism returns the current concurrency limit.
func (bh *ExtendedBatchHandler) Parallelism() int {
	return bh.parallelism
}

// HandleBatchValidated parses, validates, and executes a batch.
func (bh *ExtendedBatchHandler) HandleBatchValidated(body []byte) ([]BatchResponse, error) {
	requests, err := ParseBatchRequests(body)
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

func (bh *ExtendedBatchHandler) executeWithValidation(requests []BatchRequest, itemErrors []error) []BatchResponse {
	n := len(requests)
	responses := make([]BatchResponse, n)
	sem := make(chan struct{}, bh.parallelism)
	var wg sync.WaitGroup
	for i, req := range requests {
		if itemErrors[i] != nil {
			responses[i] = BatchResponse{
				JSONRPC: "2.0",
				Error:   &rpctypes.RPCError{Code: rpctypes.ErrCodeInvalidRequest, Message: itemErrors[i].Error()},
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
			apiReq := &rpctypes.Request{
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

// Stats returns a snapshot of current batch statistics.
func (bh *ExtendedBatchHandler) Stats() BatchStatsSnapshot {
	return bh.stats.Snapshot()
}
