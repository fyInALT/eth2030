package engine

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// ---- helpers ---------------------------------------------------------------

func v4makeDepositPayload(n int) []byte {
	return make([]byte, n*DepositRequestSize)
}

func v4makeWithdrawalPayload(n int) []byte {
	return make([]byte, n*WithdrawalRequestSize)
}

func v4makeConsolidationPayload(n int) []byte {
	return make([]byte, n*ConsolidationRequestSize)
}

// ---- DecodeDepositRequests -------------------------------------------------

func TestDecodeDepositRequests_Empty(t *testing.T) {
	deps, err := DecodeDepositRequests(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deps != nil {
		t.Errorf("expected nil, got %v", deps)
	}
}

func TestDecodeDepositRequests_InvalidLength(t *testing.T) {
	_, err := DecodeDepositRequests(make([]byte, 10))
	if err == nil {
		t.Error("expected error for invalid length")
	}
}

func TestDecodeDepositRequests_One(t *testing.T) {
	buf := v4makeDepositPayload(1)
	buf[0] = 0xAB  // pubkey byte 0
	binary.LittleEndian.PutUint64(buf[80:], 32_000_000_000)  // amount
	binary.LittleEndian.PutUint64(buf[184:], 42)              // index

	deps, err := DecodeDepositRequests(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deps) != 1 {
		t.Fatalf("len = %d, want 1", len(deps))
	}
	if deps[0].Pubkey[0] != 0xAB {
		t.Errorf("pubkey[0] = 0x%02x, want 0xAB", deps[0].Pubkey[0])
	}
	if deps[0].Amount != 32_000_000_000 {
		t.Errorf("Amount = %d, want 32_000_000_000", deps[0].Amount)
	}
	if deps[0].Index != 42 {
		t.Errorf("Index = %d, want 42", deps[0].Index)
	}
}

func TestDecodeDepositRequests_Two(t *testing.T) {
	buf := v4makeDepositPayload(2)
	binary.LittleEndian.PutUint64(buf[184:], 7)
	binary.LittleEndian.PutUint64(buf[DepositRequestSize+184:], 8)

	deps, err := DecodeDepositRequests(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deps) != 2 {
		t.Fatalf("len = %d, want 2", len(deps))
	}
	if deps[0].Index != 7 || deps[1].Index != 8 {
		t.Errorf("indices = %d, %d; want 7, 8", deps[0].Index, deps[1].Index)
	}
}

// ---- EncodeDepositRequest / round-trip -------------------------------------

func TestEncodeDecodeDepositRequest(t *testing.T) {
	orig := DepositRequest{
		Amount: 32_000_000_000,
		Index:  99,
	}
	orig.Pubkey[0] = 0xCC
	orig.WithdrawalCredentials[0] = 0xDD
	orig.Signature[0] = 0xEE

	encoded := EncodeDepositRequest(&orig)
	if len(encoded) != DepositRequestSize {
		t.Fatalf("encoded length = %d, want %d", len(encoded), DepositRequestSize)
	}

	deps, err := DecodeDepositRequests(encoded)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(deps) != 1 {
		t.Fatalf("len = %d, want 1", len(deps))
	}
	got := deps[0]
	if got.Pubkey[0] != orig.Pubkey[0] {
		t.Errorf("Pubkey[0] = 0x%02x, want 0x%02x", got.Pubkey[0], orig.Pubkey[0])
	}
	if got.WithdrawalCredentials[0] != orig.WithdrawalCredentials[0] {
		t.Errorf("WithdrawalCreds[0] mismatch")
	}
	if got.Signature[0] != orig.Signature[0] {
		t.Errorf("Signature[0] mismatch")
	}
	if got.Amount != orig.Amount {
		t.Errorf("Amount = %d, want %d", got.Amount, orig.Amount)
	}
	if got.Index != orig.Index {
		t.Errorf("Index = %d, want %d", got.Index, orig.Index)
	}
}

// ---- DecodeWithdrawalRequests ----------------------------------------------

func TestDecodeWithdrawalRequests_Empty(t *testing.T) {
	reqs, err := DecodeWithdrawalRequests(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reqs != nil {
		t.Errorf("expected nil, got %v", reqs)
	}
}

func TestDecodeWithdrawalRequests_InvalidLength(t *testing.T) {
	_, err := DecodeWithdrawalRequests(make([]byte, 5))
	if err == nil {
		t.Error("expected error for invalid length")
	}
}

func TestEncodeDecodeWithdrawalRequest(t *testing.T) {
	orig := WithdrawalRequest{
		Amount: 16_000_000_000,
	}
	orig.SourceAddress[0] = 0xAA
	orig.ValidatorPubkey[0] = 0xBB

	encoded := EncodeWithdrawalRequest(&orig)
	if len(encoded) != WithdrawalRequestSize {
		t.Fatalf("encoded length = %d, want %d", len(encoded), WithdrawalRequestSize)
	}

	reqs, err := DecodeWithdrawalRequests(encoded)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(reqs) != 1 {
		t.Fatalf("len = %d, want 1", len(reqs))
	}
	got := reqs[0]
	if got.SourceAddress[0] != orig.SourceAddress[0] {
		t.Errorf("SourceAddress[0] mismatch")
	}
	if got.ValidatorPubkey[0] != orig.ValidatorPubkey[0] {
		t.Errorf("ValidatorPubkey[0] mismatch")
	}
	if got.Amount != orig.Amount {
		t.Errorf("Amount = %d, want %d", got.Amount, orig.Amount)
	}
}

// ---- DecodeConsolidationRequests -------------------------------------------

func TestDecodeConsolidationRequests_Empty(t *testing.T) {
	reqs, err := DecodeConsolidationRequests(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reqs != nil {
		t.Errorf("expected nil, got %v", reqs)
	}
}

func TestDecodeConsolidationRequests_InvalidLength(t *testing.T) {
	_, err := DecodeConsolidationRequests(make([]byte, 5))
	if err == nil {
		t.Error("expected error for invalid length")
	}
}

func TestEncodeDecodeConsolidationRequest(t *testing.T) {
	orig := ConsolidationRequest{}
	orig.SourceAddress[0] = 0x11
	orig.SourcePubkey[0] = 0x22
	orig.TargetPubkey[0] = 0x33

	encoded := EncodeConsolidationRequest(&orig)
	if len(encoded) != ConsolidationRequestSize {
		t.Fatalf("encoded length = %d, want %d", len(encoded), ConsolidationRequestSize)
	}

	reqs, err := DecodeConsolidationRequests(encoded)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(reqs) != 1 {
		t.Fatalf("len = %d, want 1", len(reqs))
	}
	got := reqs[0]
	if got.SourceAddress[0] != orig.SourceAddress[0] {
		t.Errorf("SourceAddress[0] mismatch")
	}
	if got.SourcePubkey[0] != orig.SourcePubkey[0] {
		t.Errorf("SourcePubkey[0] mismatch")
	}
	if got.TargetPubkey[0] != orig.TargetPubkey[0] {
		t.Errorf("TargetPubkey[0] mismatch")
	}
}

// ---- ValidateExecutionRequests ---------------------------------------------

func TestValidateExecutionRequests_Nil(t *testing.T) {
	err := ValidateExecutionRequests(nil)
	if !errors.Is(err, ErrV4MissingRequests) {
		t.Errorf("expected ErrV4MissingRequests, got %v", err)
	}
}

func TestValidateExecutionRequests_EmptySlice(t *testing.T) {
	// Empty slice (not nil) is valid.
	if err := ValidateExecutionRequests([][]byte{}); err != nil {
		t.Errorf("unexpected error for empty slice: %v", err)
	}
}

func TestValidateExecutionRequests_EmptyRequest(t *testing.T) {
	err := ValidateExecutionRequests([][]byte{{}})
	if !errors.Is(err, ErrV4RequestTypeMismatch) {
		t.Errorf("expected ErrV4RequestTypeMismatch, got %v", err)
	}
}

func TestValidateExecutionRequests_OrderViolation(t *testing.T) {
	// Deposit (0x00) then type 0x02 is fine, but 0x02 then 0x01 is wrong.
	req0 := append([]byte{ConsolidationRequestType}, v4makeConsolidationPayload(1)...)
	req1 := append([]byte{WithdrawalRequestType}, v4makeWithdrawalPayload(1)...)
	err := ValidateExecutionRequests([][]byte{req0, req1})
	if !errors.Is(err, ErrV4InvalidRequestOrder) {
		t.Errorf("expected ErrV4InvalidRequestOrder, got %v", err)
	}
}

func TestValidateExecutionRequests_DuplicateType(t *testing.T) {
	req0 := append([]byte{DepositRequestType}, v4makeDepositPayload(1)...)
	req1 := append([]byte{DepositRequestType}, v4makeDepositPayload(1)...)
	err := ValidateExecutionRequests([][]byte{req0, req1})
	if !errors.Is(err, ErrV4DuplicateRequestType) {
		t.Errorf("expected ErrV4DuplicateRequestType, got %v", err)
	}
}

func TestValidateExecutionRequests_ValidOrdering(t *testing.T) {
	dep := append([]byte{DepositRequestType}, v4makeDepositPayload(1)...)
	wd := append([]byte{WithdrawalRequestType}, v4makeWithdrawalPayload(1)...)
	cons := append([]byte{ConsolidationRequestType}, v4makeConsolidationPayload(1)...)
	if err := ValidateExecutionRequests([][]byte{dep, wd, cons}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateExecutionRequests_UnknownType(t *testing.T) {
	// Unknown types should be accepted without validation.
	req := []byte{0xFF, 0x01, 0x02}
	if err := ValidateExecutionRequests([][]byte{req}); err != nil {
		t.Errorf("unexpected error for unknown type: %v", err)
	}
}

func TestValidateExecutionRequests_InvalidDepositSize(t *testing.T) {
	// Payload that is not a multiple of DepositRequestSize.
	payload := make([]byte, DepositRequestSize+10)
	req := append([]byte{DepositRequestType}, payload...)
	err := ValidateExecutionRequests([][]byte{req})
	if !errors.Is(err, ErrV4RequestTooLarge) {
		t.Errorf("expected ErrV4RequestTooLarge, got %v", err)
	}
}

// ---- ClassifyExecutionRequests ---------------------------------------------

func TestClassifyExecutionRequests_Empty(t *testing.T) {
	result, err := ClassifyExecutionRequests(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if len(result.Deposits) != 0 || len(result.Withdrawals) != 0 || len(result.Consolidations) != 0 {
		t.Errorf("expected empty slices, got %+v", result)
	}
}

func TestClassifyExecutionRequests_AllTypes(t *testing.T) {
	dep := DepositRequest{Amount: 1}
	dep.Pubkey[0] = 0xAA
	wd := WithdrawalRequest{Amount: 2}
	wd.SourceAddress[0] = 0xBB
	cons := ConsolidationRequest{}
	cons.SourceAddress[0] = 0xCC

	requests := BuildExecutionRequestsList(
		[]DepositRequest{dep},
		[]WithdrawalRequest{wd},
		[]ConsolidationRequest{cons},
	)

	result, err := ClassifyExecutionRequests(requests)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Deposits) != 1 {
		t.Errorf("Deposits = %d, want 1", len(result.Deposits))
	}
	if len(result.Withdrawals) != 1 {
		t.Errorf("Withdrawals = %d, want 1", len(result.Withdrawals))
	}
	if len(result.Consolidations) != 1 {
		t.Errorf("Consolidations = %d, want 1", len(result.Consolidations))
	}
	if result.Deposits[0].Amount != 1 {
		t.Errorf("deposit amount = %d, want 1", result.Deposits[0].Amount)
	}
	if result.Withdrawals[0].Amount != 2 {
		t.Errorf("withdrawal amount = %d, want 2", result.Withdrawals[0].Amount)
	}
}

func TestClassifyExecutionRequests_EmptyRequest(t *testing.T) {
	_, err := ClassifyExecutionRequests([][]byte{{}})
	if err == nil {
		t.Error("expected error for empty request")
	}
}

func TestClassifyExecutionRequests_UnknownType(t *testing.T) {
	req := []byte{0xFF, 0x01}
	result, err := ClassifyExecutionRequests([][]byte{req})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Unknown types are silently skipped.
	if len(result.Deposits) != 0 || len(result.Withdrawals) != 0 || len(result.Consolidations) != 0 {
		t.Errorf("expected empty results for unknown type, got %+v", result)
	}
}

// ---- BuildExecutionRequestsList -------------------------------------------

func TestBuildExecutionRequestsList_OnlyDeposits(t *testing.T) {
	dep := DepositRequest{Amount: 100}
	list := BuildExecutionRequestsList([]DepositRequest{dep}, nil, nil)
	if len(list) != 1 {
		t.Fatalf("len = %d, want 1", len(list))
	}
	if list[0][0] != DepositRequestType {
		t.Errorf("type byte = 0x%02x, want 0x%02x", list[0][0], DepositRequestType)
	}
	// Payload length = DepositRequestSize.
	if len(list[0]) != 1+DepositRequestSize {
		t.Errorf("len = %d, want %d", len(list[0]), 1+DepositRequestSize)
	}
}

func TestBuildExecutionRequestsList_OnlyWithdrawals(t *testing.T) {
	wd := WithdrawalRequest{Amount: 50}
	list := BuildExecutionRequestsList(nil, []WithdrawalRequest{wd}, nil)
	if len(list) != 1 {
		t.Fatalf("len = %d, want 1", len(list))
	}
	if list[0][0] != WithdrawalRequestType {
		t.Errorf("type byte = 0x%02x, want 0x%02x", list[0][0], WithdrawalRequestType)
	}
}

func TestBuildExecutionRequestsList_OnlyConsolidations(t *testing.T) {
	cons := ConsolidationRequest{}
	list := BuildExecutionRequestsList(nil, nil, []ConsolidationRequest{cons})
	if len(list) != 1 {
		t.Fatalf("len = %d, want 1", len(list))
	}
	if list[0][0] != ConsolidationRequestType {
		t.Errorf("type byte = 0x%02x, want 0x%02x", list[0][0], ConsolidationRequestType)
	}
}

func TestBuildExecutionRequestsList_None(t *testing.T) {
	list := BuildExecutionRequestsList(nil, nil, nil)
	if len(list) != 0 {
		t.Errorf("len = %d, want 0", len(list))
	}
}

func TestBuildExecutionRequestsList_RoundTrip(t *testing.T) {
	dep := DepositRequest{Amount: 999}
	dep.Pubkey[5] = 0xDE
	wd := WithdrawalRequest{Amount: 888}
	wd.ValidatorPubkey[3] = 0xAD
	cons := ConsolidationRequest{}
	cons.TargetPubkey[10] = 0xBE

	list := BuildExecutionRequestsList(
		[]DepositRequest{dep},
		[]WithdrawalRequest{wd},
		[]ConsolidationRequest{cons},
	)
	if len(list) != 3 {
		t.Fatalf("len = %d, want 3", len(list))
	}

	result, err := ClassifyExecutionRequests(list)
	if err != nil {
		t.Fatalf("classify: %v", err)
	}
	if result.Deposits[0].Amount != 999 {
		t.Errorf("deposit amount = %d, want 999", result.Deposits[0].Amount)
	}
	if result.Deposits[0].Pubkey[5] != 0xDE {
		t.Errorf("pubkey[5] = 0x%02x, want 0xDE", result.Deposits[0].Pubkey[5])
	}
	if result.Withdrawals[0].Amount != 888 {
		t.Errorf("withdrawal amount = %d, want 888", result.Withdrawals[0].Amount)
	}
	if result.Consolidations[0].TargetPubkey[10] != 0xBE {
		t.Errorf("target pubkey[10] = 0x%02x, want 0xBE", result.Consolidations[0].TargetPubkey[10])
	}
}

// ---- ExecutionRequestsHash -------------------------------------------------

func TestV4ExecutionRequestsHash_Empty(t *testing.T) {
	h := ExecutionRequestsHash(nil)
	var zero [32]byte
	if h != zero {
		t.Errorf("expected zero hash for nil requests, got %x", h)
	}
}

func TestV4ExecutionRequestsHash_Deterministic(t *testing.T) {
	reqs := [][]byte{{0x00, 0x01}, {0x01, 0x02}}
	h1 := ExecutionRequestsHash(reqs)
	h2 := ExecutionRequestsHash(reqs)
	if h1 != h2 {
		t.Error("not deterministic")
	}
}

// ---- errors_extended: IsClientError / IsServerError / IsEngineError --------

func TestIsClientError(t *testing.T) {
	// IsClientError: code >= -32699 && code <= -32600
	// ParseErrorCode (-32700) is outside this range per implementation.
	tests := []struct {
		code int
		want bool
	}{
		{ParseErrorCode, false},         // -32700 is outside [-32699,-32600]
		{InvalidRequestCode, true},      // -32600
		{MethodNotFoundCode, true},      // -32601
		{InvalidParamsCode, true},       // -32602
		{InternalErrorCode, true},       // -32603
		{UnknownPayloadCode, false},     // -38001 is an Engine error
		{ServerBusyCode, false},         // -32005 is outside [-32699,-32600]
		{0, false},
	}
	for _, tc := range tests {
		got := IsClientError(tc.code)
		if got != tc.want {
			t.Errorf("IsClientError(%d) = %v, want %v", tc.code, got, tc.want)
		}
	}
}

func TestIsServerError(t *testing.T) {
	tests := []struct {
		code int
		want bool
	}{
		{-32000, true},
		{-32099, true},
		{-32005, true},
		{-32006, true},
		{InternalErrorCode, false}, // -32603 outside [-32099,-32000]
		{0, false},
	}
	for _, tc := range tests {
		got := IsServerError(tc.code)
		if got != tc.want {
			t.Errorf("IsServerError(%d) = %v, want %v", tc.code, got, tc.want)
		}
	}
}

func TestIsEngineError(t *testing.T) {
	tests := []struct {
		code int
		want bool
	}{
		{UnknownPayloadCode, true},
		{InvalidForkchoiceStateCode, true},
		{InvalidPayloadAttributeCode, true},
		{TooLargeRequestCode, true},
		{UnsupportedForkCode, true},
		{InternalErrorCode, false},
		{0, false},
	}
	for _, tc := range tests {
		got := IsEngineError(tc.code)
		if got != tc.want {
			t.Errorf("IsEngineError(%d) = %v, want %v", tc.code, got, tc.want)
		}
	}
}

// ---- errors_extended: ErrorResponse ----------------------------------------

func TestErrorResponse(t *testing.T) {
	id := json.RawMessage(`1`)
	resp := ErrorResponse(id, InvalidParamsCode, "bad params")

	var out struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Error   struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(resp, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.JSONRPC != "2.0" {
		t.Errorf("jsonrpc = %q, want 2.0", out.JSONRPC)
	}
	if out.Error.Code != InvalidParamsCode {
		t.Errorf("code = %d, want %d", out.Error.Code, InvalidParamsCode)
	}
	if out.Error.Message != "bad params" {
		t.Errorf("message = %q, want %q", out.Error.Message, "bad params")
	}
}

// ---- errors_extended: EngineError ------------------------------------------

func TestEngineError_Error(t *testing.T) {
	e := NewEngineError(InternalErrorCode, "something went wrong")
	if e.Error() != "something went wrong" {
		t.Errorf("Error() = %q", e.Error())
	}
}

func TestEngineError_WithCause(t *testing.T) {
	cause := errors.New("root cause")
	e := WrapEngineError(InternalErrorCode, "wrapped", cause)
	if !strings.Contains(e.Error(), "root cause") {
		t.Errorf("Error() = %q, want to contain 'root cause'", e.Error())
	}
	if !errors.Is(e, cause) {
		t.Error("errors.Is should find cause via Unwrap")
	}
}

func TestEngineError_MarshalJSON(t *testing.T) {
	e := NewEngineError(InvalidParamsCode, "test error")
	b, err := e.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	var out struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Code != InvalidParamsCode {
		t.Errorf("code = %d, want %d", out.Code, InvalidParamsCode)
	}
	if out.Message != "test error" {
		t.Errorf("message = %q, want %q", out.Message, "test error")
	}
}

func TestErrorCodeFromError(t *testing.T) {
	tests := []struct {
		err      error
		wantCode int
	}{
		{nil, 0},
		{ErrUnknownPayload, UnknownPayloadCode},
		{ErrPayloadNotBuilding, UnknownPayloadCode},
		{ErrInvalidForkchoiceState, InvalidForkchoiceStateCode},
		{ErrInvalidPayloadAttributes, InvalidPayloadAttributeCode},
		{ErrTooLargeRequest, TooLargeRequestCode},
		{ErrRequestTooLarge, TooLargeRequestCode},
		{ErrUnsupportedFork, UnsupportedForkCode},
		{ErrServerBusy, ServerBusyCode},
		{ErrRequestTimeout, RequestTimeoutCode},
		{errors.New("unknown"), InternalErrorCode},
	}
	for _, tc := range tests {
		got := ErrorCodeFromError(tc.err)
		if got != tc.wantCode {
			t.Errorf("ErrorCodeFromError(%v) = %d, want %d", tc.err, got, tc.wantCode)
		}
	}
}

func TestErrorCodeFromError_EngineError(t *testing.T) {
	e := NewEngineError(ServerBusyCode, "busy")
	if got := ErrorCodeFromError(e); got != ServerBusyCode {
		t.Errorf("got %d, want %d", got, ServerBusyCode)
	}
}

func TestValidatePayloadVersion(t *testing.T) {
	tests := []struct {
		version     int
		withdrawals bool
		requests    bool
		bal         bool
		wantNil     bool
	}{
		{1, false, false, false, true},
		{2, true, false, false, true},
		{2, false, false, false, false}, // missing withdrawals
		{4, true, true, false, true},
		{4, true, false, false, false}, // missing execution requests
		{5, true, true, true, true},
		{5, true, true, false, false}, // missing BAL
	}
	for _, tc := range tests {
		err := ValidatePayloadVersion(tc.version, tc.withdrawals, tc.requests, tc.bal)
		if tc.wantNil && err != nil {
			t.Errorf("version=%d: unexpected error: %v", tc.version, err)
		}
		if !tc.wantNil && err == nil {
			t.Errorf("version=%d: expected error, got nil", tc.version)
		}
	}
}

func TestErrorName(t *testing.T) {
	tests := []struct {
		code int
		want string
	}{
		{ParseErrorCode, "ParseError"},
		{InvalidRequestCode, "InvalidRequest"},
		{MethodNotFoundCode, "MethodNotFound"},
		{InvalidParamsCode, "InvalidParams"},
		{InternalErrorCode, "InternalError"},
		{UnknownPayloadCode, "UnknownPayload"},
		{InvalidForkchoiceStateCode, "InvalidForkchoiceState"},
		{InvalidPayloadAttributeCode, "InvalidPayloadAttributes"},
		{TooLargeRequestCode, "TooLargeRequest"},
		{UnsupportedForkCode, "UnsupportedFork"},
		{ServerBusyCode, "ServerBusy"},
		{RequestTimeoutCode, "RequestTimeout"},
		{9999, "Unknown(9999)"},
	}
	for _, tc := range tests {
		got := ErrorName(tc.code)
		if got != tc.want {
			t.Errorf("ErrorName(%d) = %q, want %q", tc.code, got, tc.want)
		}
	}
}
