package rpcbatch

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	rpctypes "github.com/eth2030/eth2030/rpc/types"
)

// mockRequestHandler is a minimal stub that satisfies RequestHandler.
// It handles eth_chainId, eth_blockNumber, eth_gasPrice and returns
// method-not-found for everything else.
type mockRequestHandler struct{}

func (m *mockRequestHandler) HandleRequest(req *rpctypes.Request) *rpctypes.Response {
	switch req.Method {
	case "eth_chainId":
		return rpctypes.NewSuccessResponse(req.ID, "0x539")
	case "eth_blockNumber":
		return rpctypes.NewSuccessResponse(req.ID, "0x2a")
	case "eth_gasPrice":
		return rpctypes.NewSuccessResponse(req.ID, "0x3b9aca00")
	default:
		return rpctypes.NewErrorResponse(req.ID, rpctypes.ErrCodeMethodNotFound,
			"method not found: "+req.Method)
	}
}

func newMockHandler() *mockRequestHandler { return &mockRequestHandler{} }

func TestNewBatchHandler(t *testing.T) {
	bh := NewBatchHandler(newMockHandler())
	if bh == nil {
		t.Fatal("NewBatchHandler returned nil")
	}
	if bh.Parallelism() != DefaultParallelism {
		t.Fatalf("expected parallelism %d, got %d", DefaultParallelism, bh.Parallelism())
	}
}

func TestBatchHandler_SetParallelism(t *testing.T) {
	bh := NewBatchHandler(newMockHandler())
	bh.SetParallelism(4)
	if bh.Parallelism() != 4 {
		t.Fatalf("expected parallelism 4, got %d", bh.Parallelism())
	}
	// Minimum is 1.
	bh.SetParallelism(0)
	if bh.Parallelism() != 1 {
		t.Fatalf("expected parallelism 1, got %d", bh.Parallelism())
	}
}

func TestBatchHandler_HandleBatch_SingleRequest(t *testing.T) {
	bh := NewBatchHandler(newMockHandler())

	body := `[{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}]`
	responses, err := bh.HandleBatch([]byte(body))
	if err != nil {
		t.Fatalf("HandleBatch: %v", err)
	}
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	if responses[0].Error != nil {
		t.Fatalf("unexpected error: %v", responses[0].Error.Message)
	}
	if responses[0].Result != "0x539" {
		t.Fatalf("expected 0x539, got %v", responses[0].Result)
	}
}

func TestBatchHandler_HandleBatch_MultipleRequests(t *testing.T) {
	bh := NewBatchHandler(newMockHandler())

	body := `[
		{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1},
		{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":2},
		{"jsonrpc":"2.0","method":"eth_gasPrice","params":[],"id":3}
	]`
	responses, err := bh.HandleBatch([]byte(body))
	if err != nil {
		t.Fatalf("HandleBatch: %v", err)
	}
	if len(responses) != 3 {
		t.Fatalf("expected 3 responses, got %d", len(responses))
	}
	expected := []string{"0x539", "0x2a", "0x3b9aca00"}
	for i, want := range expected {
		got, ok := responses[i].Result.(string)
		if !ok {
			t.Fatalf("response %d: result not string: %T", i, responses[i].Result)
		}
		if got != want {
			t.Fatalf("response %d: got %q, want %q", i, got, want)
		}
	}
}

func TestBatchHandler_HandleBatch_NotArray(t *testing.T) {
	bh := NewBatchHandler(newMockHandler())

	body := `{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}`
	_, err := bh.HandleBatch([]byte(body))
	if err == nil {
		t.Fatal("expected error for non-array body")
	}
}

func TestBatchHandler_HandleBatch_InvalidJSON(t *testing.T) {
	bh := NewBatchHandler(newMockHandler())

	body := `[{invalid json}]`
	_, err := bh.HandleBatch([]byte(body))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestBatchHandler_HandleBatch_EmptyArray(t *testing.T) {
	bh := NewBatchHandler(newMockHandler())

	body := `[]`
	_, err := bh.HandleBatch([]byte(body))
	if err == nil {
		t.Fatal("expected error for empty batch")
	}
}

func TestBatchHandler_HandleBatch_InvalidVersion(t *testing.T) {
	bh := NewBatchHandler(newMockHandler())

	body := `[{"jsonrpc":"1.0","method":"eth_chainId","params":[],"id":1}]`
	responses, err := bh.HandleBatch([]byte(body))
	if err != nil {
		t.Fatalf("HandleBatch: %v", err)
	}
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	if responses[0].Error == nil {
		t.Fatal("expected error for invalid jsonrpc version")
	}
	if responses[0].Error.Code != rpctypes.ErrCodeInvalidRequest {
		t.Fatalf("expected code %d, got %d", rpctypes.ErrCodeInvalidRequest, responses[0].Error.Code)
	}
}

func TestBatchHandler_HandleBatch_EmptyMethod(t *testing.T) {
	bh := NewBatchHandler(newMockHandler())

	body := `[{"jsonrpc":"2.0","method":"","params":[],"id":1}]`
	responses, err := bh.HandleBatch([]byte(body))
	if err != nil {
		t.Fatalf("HandleBatch: %v", err)
	}
	if responses[0].Error == nil {
		t.Fatal("expected error for empty method")
	}
}

func TestBatchHandler_HandleBatch_MixedSuccess(t *testing.T) {
	bh := NewBatchHandler(newMockHandler())

	body := `[
		{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1},
		{"jsonrpc":"2.0","method":"nonexistent_method","params":[],"id":2},
		{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":3}
	]`
	responses, err := bh.HandleBatch([]byte(body))
	if err != nil {
		t.Fatalf("HandleBatch: %v", err)
	}
	if len(responses) != 3 {
		t.Fatalf("expected 3 responses, got %d", len(responses))
	}
	if responses[0].Error != nil {
		t.Fatalf("response 0: unexpected error: %v", responses[0].Error.Message)
	}
	if responses[1].Error == nil {
		t.Fatal("response 1: expected error for unknown method")
	}
	if responses[2].Error != nil {
		t.Fatalf("response 2: unexpected error: %v", responses[2].Error.Message)
	}
}

func TestBatchHandler_ExecuteParallel_OrderPreserved(t *testing.T) {
	bh := NewBatchHandler(newMockHandler())

	n := 20
	requests := make([]BatchRequest, n)
	for i := 0; i < n; i++ {
		requests[i] = BatchRequest{
			JSONRPC: "2.0",
			Method:  "eth_chainId",
			Params:  nil,
			ID:      json.RawMessage(`"` + strings.Repeat("x", i+1) + `"`),
		}
	}
	responses := bh.ExecuteParallel(requests)
	if len(responses) != n {
		t.Fatalf("expected %d responses, got %d", n, len(responses))
	}
	for i := 0; i < n; i++ {
		expectedID := `"` + strings.Repeat("x", i+1) + `"`
		if string(responses[i].ID) != expectedID {
			t.Fatalf("response %d: ID mismatch: got %s, want %s", i, responses[i].ID, expectedID)
		}
	}
}

func TestMarshalBatchResponse(t *testing.T) {
	responses := []BatchResponse{
		{JSONRPC: "2.0", Result: "0x539", ID: json.RawMessage(`1`)},
		{JSONRPC: "2.0", Result: "0x2a", ID: json.RawMessage(`2`)},
	}
	data, err := MarshalBatchResponse(responses)
	if err != nil {
		t.Fatalf("MarshalBatchResponse: %v", err)
	}
	var parsed []json.RawMessage
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(parsed) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(parsed))
	}
}

func TestMarshalBatchResponse_Empty(t *testing.T) {
	data, err := MarshalBatchResponse([]BatchResponse{})
	if err != nil {
		t.Fatalf("MarshalBatchResponse: %v", err)
	}
	if string(data) != "[]" {
		t.Fatalf("expected [], got %s", data)
	}
}

func TestIsBatchRequest(t *testing.T) {
	tests := []struct {
		body string
		want bool
	}{
		{`[{"jsonrpc":"2.0"}]`, true},
		{`  [  {"jsonrpc":"2.0"}]`, true},
		{"\t[{}]", true},
		{"\n[{}]", true},
		{`{"jsonrpc":"2.0"}`, false},
		{``, false},
		{` `, false},
	}
	for _, tt := range tests {
		got := IsBatchRequest([]byte(tt.body))
		if got != tt.want {
			t.Errorf("IsBatchRequest(%q) = %v, want %v", tt.body, got, tt.want)
		}
	}
}

func TestBatchHandler_HandleBatch_IDPreserved(t *testing.T) {
	bh := NewBatchHandler(newMockHandler())

	body := `[
		{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":"alpha"},
		{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":42}
	]`
	responses, err := bh.HandleBatch([]byte(body))
	if err != nil {
		t.Fatalf("HandleBatch: %v", err)
	}
	if string(responses[0].ID) != `"alpha"` {
		t.Fatalf("response 0: ID mismatch: got %s, want %q", responses[0].ID, "alpha")
	}
	if string(responses[1].ID) != `42` {
		t.Fatalf("response 1: ID mismatch: got %s, want 42", responses[1].ID)
	}
}

func TestBatchHandler_HandleBatch_NullID(t *testing.T) {
	bh := NewBatchHandler(newMockHandler())

	body := `[{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":null}]`
	responses, err := bh.HandleBatch([]byte(body))
	if err != nil {
		t.Fatalf("HandleBatch: %v", err)
	}
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	if string(responses[0].ID) != `null` {
		t.Fatalf("ID: got %s, want null", responses[0].ID)
	}
}

func TestBatchHandler_HandleBatch_ParallelExecution(t *testing.T) {
	bh := NewBatchHandler(newMockHandler())
	bh.SetParallelism(4)

	var reqs []string
	for i := 0; i < 16; i++ {
		reqs = append(reqs, fmt.Sprintf(
			`{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":%d}`, i))
	}
	body := "[" + strings.Join(reqs, ",") + "]"

	responses, err := bh.HandleBatch([]byte(body))
	if err != nil {
		t.Fatalf("HandleBatch: %v", err)
	}
	if len(responses) != 16 {
		t.Fatalf("expected 16 responses, got %d", len(responses))
	}
	for i, resp := range responses {
		if resp.Error != nil {
			t.Fatalf("response %d: unexpected error: %v", i, resp.Error.Message)
		}
	}
}

func TestBatchHandler_HandleBatch_TooLarge(t *testing.T) {
	bh := NewBatchHandler(newMockHandler())

	var reqs []string
	for i := 0; i <= MaxBatchSize; i++ {
		reqs = append(reqs, `{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}`)
	}
	body := "[" + strings.Join(reqs, ",") + "]"

	_, err := bh.HandleBatch([]byte(body))
	if err == nil {
		t.Fatal("expected error for batch too large")
	}
}

func TestBatchHandler_ExecuteParallel_ConcurrencyBound(t *testing.T) {
	bh := NewBatchHandler(newMockHandler())
	bh.SetParallelism(2)

	requests := make([]BatchRequest, 10)
	for i := range requests {
		requests[i] = BatchRequest{
			JSONRPC: "2.0",
			Method:  "eth_blockNumber",
			Params:  nil,
			ID:      json.RawMessage(`1`),
		}
	}
	responses := bh.ExecuteParallel(requests)
	if len(responses) != 10 {
		t.Fatalf("expected 10 responses, got %d", len(responses))
	}
	for i, resp := range responses {
		if resp.Error != nil {
			t.Fatalf("response %d error: %v", i, resp.Error.Message)
		}
	}
}

func TestBatchHandler_HandleBatch_WhitespacePrefix(t *testing.T) {
	bh := NewBatchHandler(newMockHandler())

	body := "\n\t[{\"jsonrpc\":\"2.0\",\"method\":\"eth_chainId\",\"params\":[],\"id\":1}]"
	responses, err := bh.HandleBatch([]byte(body))
	if err != nil {
		t.Fatalf("HandleBatch: %v", err)
	}
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
}

func TestBatchHandler_HandleBatch_JSONRPCField(t *testing.T) {
	bh := NewBatchHandler(newMockHandler())

	body := `[{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}]`
	responses, err := bh.HandleBatch([]byte(body))
	if err != nil {
		t.Fatalf("HandleBatch: %v", err)
	}
	if responses[0].JSONRPC != "2.0" {
		t.Fatalf("expected jsonrpc 2.0, got %s", responses[0].JSONRPC)
	}
}
