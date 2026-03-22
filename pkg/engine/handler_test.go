package engine

import (
	"encoding/json"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// newHandlerTestAPI creates an EngineAPI with a mock backend for handler tests.
func newHandlerTestAPI() *EngineAPI {
	return NewEngineAPI(&handlerMockBackend{})
}

// handlerMockBackend implements engine.Backend for handler tests.
type handlerMockBackend struct{}

func (m *handlerMockBackend) ProcessBlock(payload *ExecutionPayloadV3, _ []types.Hash, _ types.Hash) (PayloadStatusV1, error) {
	return PayloadStatusV1{Status: StatusValid}, nil
}
func (m *handlerMockBackend) ProcessBlockV4(payload *ExecutionPayloadV3, _ []types.Hash, _ types.Hash, _ [][]byte) (PayloadStatusV1, error) {
	return PayloadStatusV1{Status: StatusValid}, nil
}
func (m *handlerMockBackend) ProcessBlockV5(payload *ExecutionPayloadV5, _ []types.Hash, _ types.Hash, _ [][]byte) (PayloadStatusV1, error) {
	return PayloadStatusV1{Status: StatusValid}, nil
}
func (m *handlerMockBackend) ForkchoiceUpdated(state ForkchoiceStateV1, attrs *PayloadAttributesV3) (ForkchoiceUpdatedResult, error) {
	return ForkchoiceUpdatedResult{PayloadStatus: PayloadStatusV1{Status: StatusValid}}, nil
}
func (m *handlerMockBackend) ForkchoiceUpdatedV4(state ForkchoiceStateV1, attrs *PayloadAttributesV4) (ForkchoiceUpdatedResult, error) {
	return ForkchoiceUpdatedResult{PayloadStatus: PayloadStatusV1{Status: StatusValid}}, nil
}
func (m *handlerMockBackend) GetPayloadByID(id PayloadID) (*GetPayloadResponse, error) {
	return nil, ErrUnknownPayload
}
func (m *handlerMockBackend) GetPayloadV4ByID(id PayloadID) (*GetPayloadV4Response, error) {
	return nil, ErrUnknownPayload
}
func (m *handlerMockBackend) GetPayloadV6ByID(id PayloadID) (*GetPayloadV6Response, error) {
	return nil, ErrUnknownPayload
}
func (m *handlerMockBackend) GetHeadTimestamp() uint64              { return 1000 }
func (m *handlerMockBackend) GetBlockTimestamp(_ types.Hash) uint64 { return 0 }
func (m *handlerMockBackend) IsCancun(ts uint64) bool               { return true }
func (m *handlerMockBackend) IsPrague(ts uint64) bool               { return true }
func (m *handlerMockBackend) IsAmsterdam(ts uint64) bool            { return true }
func (m *handlerMockBackend) GetHeadHash() types.Hash               { return types.Hash{} }
func (m *handlerMockBackend) GetSafeHash() types.Hash               { return types.Hash{} }
func (m *handlerMockBackend) GetFinalizedHash() types.Hash          { return types.Hash{} }

// TestHandler_ParseError tests that invalid JSON returns a parse error.
func TestHandler_ParseError(t *testing.T) {
	api := newHandlerTestAPI()
	resp := api.HandleRequest([]byte(`not-json`))

	var rpcResp jsonrpcResponse
	if err := json.Unmarshal(resp, &rpcResp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if rpcResp.Error == nil {
		t.Fatal("expected parse error")
	}
	if rpcResp.Error.Code != ParseErrorCode {
		t.Fatalf("want code %d, got %d", ParseErrorCode, rpcResp.Error.Code)
	}
}

// TestHandler_InvalidJSONRPCVersion tests that wrong version returns error.
func TestHandler_InvalidJSONRPCVersion(t *testing.T) {
	api := newHandlerTestAPI()
	req := `{"jsonrpc":"1.0","method":"engine_exchangeCapabilities","params":[[]],"id":1}`
	resp := api.HandleRequest([]byte(req))

	var rpcResp jsonrpcResponse
	if err := json.Unmarshal(resp, &rpcResp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if rpcResp.Error == nil {
		t.Fatal("expected error for wrong version")
	}
}

// TestHandler_MethodNotFound tests that an unknown method returns the
// correct error code.
func TestHandler_MethodNotFound(t *testing.T) {
	api := newHandlerTestAPI()
	req := `{"jsonrpc":"2.0","method":"engine_nonexistent","params":[],"id":1}`
	resp := api.HandleRequest([]byte(req))

	var rpcResp jsonrpcResponse
	if err := json.Unmarshal(resp, &rpcResp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if rpcResp.Error == nil {
		t.Fatal("expected method not found error")
	}
	if rpcResp.Error.Code != MethodNotFoundCode {
		t.Fatalf("want code %d, got %d", MethodNotFoundCode, rpcResp.Error.Code)
	}
}

// TestHandler_ExchangeCapabilities tests the engine_exchangeCapabilities method.
func TestHandler_ExchangeCapabilities(t *testing.T) {
	api := newHandlerTestAPI()
	req := `{"jsonrpc":"2.0","method":"engine_exchangeCapabilities","params":[["engine_newPayloadV3"]],"id":1}`
	resp := api.HandleRequest([]byte(req))

	var rpcResp jsonrpcResponse
	if err := json.Unmarshal(resp, &rpcResp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if rpcResp.Error != nil {
		t.Fatalf("error: code=%d msg=%s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	resultJSON, _ := json.Marshal(rpcResp.Result)
	var caps []string
	if err := json.Unmarshal(resultJSON, &caps); err != nil {
		t.Fatalf("unmarshal caps: %v", err)
	}
	if len(caps) == 0 {
		t.Fatal("expected non-empty capabilities")
	}
	found := false
	for _, c := range caps {
		if c == "engine_newPayloadV3" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected engine_newPayloadV3 in capabilities")
	}
}

// TestHandler_GetClientVersion tests the engine_getClientVersionV1 method.
func TestHandler_GetClientVersion(t *testing.T) {
	api := newHandlerTestAPI()
	req := `{"jsonrpc":"2.0","method":"engine_getClientVersionV1","params":[{"code":"GE","name":"geth","version":"v1.0.0","commit":"abc123"}],"id":1}`
	resp := api.HandleRequest([]byte(req))

	var rpcResp jsonrpcResponse
	if err := json.Unmarshal(resp, &rpcResp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if rpcResp.Error != nil {
		t.Fatalf("error: %s", rpcResp.Error.Message)
	}

	resultJSON, _ := json.Marshal(rpcResp.Result)
	var versions []ClientVersionV1
	if err := json.Unmarshal(resultJSON, &versions); err != nil {
		t.Fatalf("unmarshal versions: %v", err)
	}
	if len(versions) == 0 {
		t.Fatal("expected non-empty client versions")
	}
	if versions[0].Code != "ET" {
		t.Fatalf("want code ET, got %s", versions[0].Code)
	}
	if versions[0].Name != "ETH2030" {
		t.Fatalf("want name ETH2030, got %s", versions[0].Name)
	}
}

// TestHandler_DispatchAllMethods tests that dispatch routes all known methods
// without returning method-not-found errors.
func TestHandler_DispatchAllMethods(t *testing.T) {
	api := newHandlerTestAPI()

	methods := []struct {
		name   string
		params string
	}{
		{"engine_exchangeCapabilities", `[[]]`},
		{"engine_getClientVersionV1", `[{}]`},
	}

	for _, m := range methods {
		t.Run(m.name, func(t *testing.T) {
			var params []json.RawMessage
			json.Unmarshal([]byte(m.params), &params)
			_, rpcErr := api.dispatch(m.name, params)
			if rpcErr != nil && rpcErr.Code == MethodNotFoundCode {
				t.Fatalf("dispatch should recognize %s", m.name)
			}
		})
	}
}

// TestHandler_NewPayloadV3_MissingParams tests newPayloadV3 with wrong param count.
func TestHandler_NewPayloadV3_MissingParams(t *testing.T) {
	api := newHandlerTestAPI()
	req := `{"jsonrpc":"2.0","method":"engine_newPayloadV3","params":[],"id":1}`
	resp := api.HandleRequest([]byte(req))

	var rpcResp jsonrpcResponse
	if err := json.Unmarshal(resp, &rpcResp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rpcResp.Error == nil {
		t.Fatal("expected error for missing params")
	}
	if rpcResp.Error.Code != InvalidParamsCode {
		t.Fatalf("want code %d, got %d", InvalidParamsCode, rpcResp.Error.Code)
	}
}

// TestHandler_ForkchoiceUpdatedV3_MissingParams tests with wrong param count.
func TestHandler_ForkchoiceUpdatedV3_MissingParams(t *testing.T) {
	api := newHandlerTestAPI()
	req := `{"jsonrpc":"2.0","method":"engine_forkchoiceUpdatedV3","params":[],"id":1}`
	resp := api.HandleRequest([]byte(req))

	var rpcResp jsonrpcResponse
	if err := json.Unmarshal(resp, &rpcResp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rpcResp.Error == nil {
		t.Fatal("expected error for missing params")
	}
}

// TestHandler_GetPayloadV3_MissingParams tests getPayloadV3 with wrong param count.
func TestHandler_GetPayloadV3_MissingParams(t *testing.T) {
	api := newHandlerTestAPI()
	req := `{"jsonrpc":"2.0","method":"engine_getPayloadV3","params":[],"id":1}`
	resp := api.HandleRequest([]byte(req))

	var rpcResp jsonrpcResponse
	if err := json.Unmarshal(resp, &rpcResp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rpcResp.Error == nil {
		t.Fatal("expected error for missing params")
	}
}

// TestHandler_IDPropagation verifies request ID is preserved in the response.
func TestHandler_IDPropagation(t *testing.T) {
	api := newHandlerTestAPI()

	tests := []string{`1`, `"abc"`, `42`}
	for _, id := range tests {
		t.Run(id, func(t *testing.T) {
			req := `{"jsonrpc":"2.0","method":"engine_exchangeCapabilities","params":[[]],"id":` + id + `}`
			resp := api.HandleRequest([]byte(req))

			var rpcResp jsonrpcResponse
			if err := json.Unmarshal(resp, &rpcResp); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if string(rpcResp.ID) != id {
				t.Fatalf("want ID %s, got %s", id, string(rpcResp.ID))
			}
		})
	}
}

// TestHandler_EngineErrorToRPC tests the error code mapping for known engine errors.
func TestHandler_EngineErrorToRPC(t *testing.T) {
	tests := []struct {
		err  error
		code int
	}{
		{ErrUnknownPayload, UnknownPayloadCode},
		{ErrInvalidForkchoiceState, InvalidForkchoiceStateCode},
		{ErrInvalidPayloadAttributes, InvalidPayloadAttributeCode},
		{ErrInvalidParams, InvalidParamsCode},
		{ErrTooLargeRequest, TooLargeRequestCode},
		{ErrUnsupportedFork, UnsupportedForkCode},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			rpcErr := engineErrorToRPC(tt.err)
			if rpcErr.Code != tt.code {
				t.Fatalf("want code %d for %v, got %d", tt.code, tt.err, rpcErr.Code)
			}
		})
	}
}

// TestHandler_InclusionListV1 tests the inclusion list handler.
func TestHandler_InclusionListV1(t *testing.T) {
	api := newHandlerTestAPI()

	il := InclusionListV1{
		Slot:           5,
		ValidatorIndex: 1,
		CommitteeRoot:  types.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
		Transactions:   []hexBytes{{0x01, 0x02}},
	}
	ilJSON, _ := json.Marshal(il)
	params := []json.RawMessage{ilJSON}

	result, rpcErr := api.handleNewInclusionListV1(params)
	if rpcErr != nil {
		t.Fatalf("error: code=%d msg=%s", rpcErr.Code, rpcErr.Message)
	}
	status, ok := result.(InclusionListStatusV1)
	if !ok {
		t.Fatalf("result type: %T", result)
	}
	// Backend doesn't implement InclusionListBackend, so it should return INVALID.
	if status.Status != ILStatusInvalid {
		t.Fatalf("want INVALID, got %s", status.Status)
	}
}

// TestHandler_GetInclusionListV1 tests getting an inclusion list.
func TestHandler_GetInclusionListV1(t *testing.T) {
	api := newHandlerTestAPI()

	result, rpcErr := api.handleGetInclusionListV1(nil)
	if rpcErr != nil {
		t.Fatalf("error: code=%d msg=%s", rpcErr.Code, rpcErr.Message)
	}
	resp, ok := result.(*GetInclusionListResponseV1)
	if !ok {
		t.Fatalf("result type: %T", result)
	}
	if len(resp.Transactions) != 0 {
		t.Fatalf("want 0 transactions, got %d", len(resp.Transactions))
	}
}

// TestHandler_NewPayloadV4_MissingParams tests newPayloadV4 with wrong param count.
func TestHandler_NewPayloadV4_MissingParams(t *testing.T) {
	api := newHandlerTestAPI()
	req := `{"jsonrpc":"2.0","method":"engine_newPayloadV4","params":[],"id":1}`
	resp := api.HandleRequest([]byte(req))

	var rpcResp jsonrpcResponse
	if err := json.Unmarshal(resp, &rpcResp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rpcResp.Error == nil {
		t.Fatal("expected error for missing params")
	}
}

// TestHandler_NewPayloadV5_MissingParams tests newPayloadV5 with wrong param count.
func TestHandler_NewPayloadV5_MissingParams(t *testing.T) {
	api := newHandlerTestAPI()
	req := `{"jsonrpc":"2.0","method":"engine_newPayloadV5","params":[],"id":1}`
	resp := api.HandleRequest([]byte(req))

	var rpcResp jsonrpcResponse
	if err := json.Unmarshal(resp, &rpcResp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rpcResp.Error == nil {
		t.Fatal("expected error for missing params")
	}
}

// TestHandler_GetPayloadV4_MissingParams tests getPayloadV4 with wrong param count.
func TestHandler_GetPayloadV4_MissingParams(t *testing.T) {
	api := newHandlerTestAPI()
	req := `{"jsonrpc":"2.0","method":"engine_getPayloadV4","params":[],"id":1}`
	resp := api.HandleRequest([]byte(req))

	var rpcResp jsonrpcResponse
	if err := json.Unmarshal(resp, &rpcResp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rpcResp.Error == nil {
		t.Fatal("expected error for missing params")
	}
}

// TestHandler_GetPayloadV6_MissingParams tests getPayloadV6 with wrong param count.
func TestHandler_GetPayloadV6_MissingParams(t *testing.T) {
	api := newHandlerTestAPI()
	req := `{"jsonrpc":"2.0","method":"engine_getPayloadV6","params":[],"id":1}`
	resp := api.HandleRequest([]byte(req))

	var rpcResp jsonrpcResponse
	if err := json.Unmarshal(resp, &rpcResp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rpcResp.Error == nil {
		t.Fatal("expected error for missing params")
	}
}

// TestHandler_ForkchoiceUpdatedV4_MissingParams tests forkchoiceUpdatedV4 with wrong param count.
func TestHandler_ForkchoiceUpdatedV4_MissingParams(t *testing.T) {
	api := newHandlerTestAPI()
	req := `{"jsonrpc":"2.0","method":"engine_forkchoiceUpdatedV4","params":[],"id":1}`
	resp := api.HandleRequest([]byte(req))

	var rpcResp jsonrpcResponse
	if err := json.Unmarshal(resp, &rpcResp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rpcResp.Error == nil {
		t.Fatal("expected error for missing params")
	}
}

// TestHandler_GetPayloadHeaderV1_MissingParams tests getPayloadHeaderV1 with wrong param count.
func TestHandler_GetPayloadHeaderV1_MissingParams(t *testing.T) {
	api := newHandlerTestAPI()
	_, rpcErr := api.handleGetPayloadHeaderV1(nil)
	if rpcErr == nil {
		t.Fatal("expected error for nil params")
	}
	if rpcErr.Code != InvalidParamsCode {
		t.Fatalf("want code %d, got %d", InvalidParamsCode, rpcErr.Code)
	}
}

// TestHandler_SubmitBlindedBlockV1_MissingParams tests submitBlindedBlockV1 with wrong param count.
func TestHandler_SubmitBlindedBlockV1_MissingParams(t *testing.T) {
	api := newHandlerTestAPI()
	_, rpcErr := api.handleSubmitBlindedBlockV1(nil)
	if rpcErr == nil {
		t.Fatal("expected error for nil params")
	}
	if rpcErr.Code != InvalidParamsCode {
		t.Fatalf("want code %d, got %d", InvalidParamsCode, rpcErr.Code)
	}
}

// TestHexBytes_JSON tests hexBytes JSON marshaling/unmarshaling.
func TestHexBytes_JSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []byte
	}{
		{
			name:     "hex with 0x prefix",
			input:    `"0x0102"`,
			expected: []byte{0x01, 0x02},
		},
		{
			name:     "hex with 0X prefix",
			input:    `"0X0a0b"`,
			expected: []byte{0x0a, 0x0b},
		},
		{
			name:     "hex without prefix",
			input:    `"deadbeef"`,
			expected: []byte{0xde, 0xad, 0xbe, 0xef},
		},
		{
			name:     "odd length hex",
			input:    `"0x1"`,
			expected: []byte{0x01},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var h hexBytes
			if err := json.Unmarshal([]byte(tt.input), &h); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}
			if string(h) != string(tt.expected) {
				t.Fatalf("want %x, got %x", tt.expected, h)
			}
		})
	}
}

// TestHexBytes_Marshal tests hexBytes marshals to hex string with 0x prefix.
func TestHexBytes_Marshal(t *testing.T) {
	h := hexBytes{0x01, 0x02, 0x03}
	data, err := json.Marshal(h)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	expected := `"0x010203"`
	if string(data) != expected {
		t.Fatalf("want %s, got %s", expected, string(data))
	}
}

// TestInclusionListV1_JSON tests InclusionListV1 JSON with hex-encoded transactions.
func TestInclusionListV1_JSON(t *testing.T) {
	// Simulate what Lighthouse sends (hex-encoded strings with 0x prefix)
	jsonInput := `{
		"slot": "123",
		"validatorIndex": "456",
		"inclusionListCommitteeRoot": "0x1111111111111111111111111111111111111111111111111111111111111111",
		"transactions": ["0x0102", "0xdeadbeef"]
	}`

	var il InclusionListV1
	if err := json.Unmarshal([]byte(jsonInput), &il); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if uint64(il.Slot) != 123 {
		t.Fatalf("want slot 123, got %d", il.Slot)
	}
	if uint64(il.ValidatorIndex) != 456 {
		t.Fatalf("want validatorIndex 456, got %d", il.ValidatorIndex)
	}
	if len(il.Transactions) != 2 {
		t.Fatalf("want 2 transactions, got %d", len(il.Transactions))
	}
	if string(il.Transactions[0]) != string([]byte{0x01, 0x02}) {
		t.Fatalf("want first tx [01 02], got %x", il.Transactions[0])
	}
	if string(il.Transactions[1]) != string([]byte{0xde, 0xad, 0xbe, 0xef}) {
		t.Fatalf("want second tx [de ad be ef], got %x", il.Transactions[1])
	}
}
