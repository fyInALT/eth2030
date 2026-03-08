package errors

import (
	"encoding/json"
	stderrors "errors"
	"strings"
	"testing"
)

// ---- error sentinel tests (from apierrors) ---------------------------------

func TestErrors(t *testing.T) {
	errs := []error{
		ErrInvalidParams,
		ErrUnknownPayload,
		ErrInvalidForkchoiceState,
		ErrInvalidPayloadAttributes,
		ErrTooLargeRequest,
		ErrUnsupportedFork,
		ErrInvalidBlockHash,
		ErrInvalidBlobHashes,
		ErrMissingBeaconRoot,
	}
	for _, err := range errs {
		if err == nil {
			t.Fatal("error var is nil")
		}
		for _, other := range errs {
			if err != other && stderrors.Is(err, other) {
				t.Errorf("%v unexpectedly matches %v", err, other)
			}
		}
	}
}

func TestErrorCodes(t *testing.T) {
	if ParseErrorCode != -32700 {
		t.Errorf("ParseErrorCode = %d, want -32700", ParseErrorCode)
	}
	if InvalidRequestCode != -32600 {
		t.Errorf("InvalidRequestCode = %d, want -32600", InvalidRequestCode)
	}
	if MethodNotFoundCode != -32601 {
		t.Errorf("MethodNotFoundCode = %d, want -32601", MethodNotFoundCode)
	}
	if InvalidParamsCode != -32602 {
		t.Errorf("InvalidParamsCode = %d, want -32602", InvalidParamsCode)
	}
	if InternalErrorCode != -32603 {
		t.Errorf("InternalErrorCode = %d, want -32603", InternalErrorCode)
	}
	if UnknownPayloadCode != -38001 {
		t.Errorf("UnknownPayloadCode = %d, want -38001", UnknownPayloadCode)
	}
	if InvalidForkchoiceStateCode != -38002 {
		t.Errorf("InvalidForkchoiceStateCode = %d, want -38002", InvalidForkchoiceStateCode)
	}
	if InvalidPayloadAttributeCode != -38003 {
		t.Errorf("InvalidPayloadAttributeCode = %d, want -38003", InvalidPayloadAttributeCode)
	}
	if TooLargeRequestCode != -38004 {
		t.Errorf("TooLargeRequestCode = %d, want -38004", TooLargeRequestCode)
	}
	if UnsupportedForkCode != -38005 {
		t.Errorf("UnsupportedForkCode = %d, want -38005", UnsupportedForkCode)
	}
}

// ---- IsClientError / IsServerError / IsEngineError -------------------------

func TestIsClientError(t *testing.T) {
	tests := []struct {
		code int
		want bool
	}{
		{ParseErrorCode, false},
		{InvalidRequestCode, true},
		{MethodNotFoundCode, true},
		{InvalidParamsCode, true},
		{InternalErrorCode, true},
		{UnknownPayloadCode, false},
		{ServerBusyCode, false},
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
		{InternalErrorCode, false},
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

// ---- ErrorResponse ---------------------------------------------------------

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

// ---- EngineError -----------------------------------------------------------

func TestEngineError_Error(t *testing.T) {
	e := NewEngineError(InternalErrorCode, "something went wrong")
	if e.Error() != "something went wrong" {
		t.Errorf("Error() = %q", e.Error())
	}
}

func TestEngineError_WithCause(t *testing.T) {
	cause := stderrors.New("root cause")
	e := WrapEngineError(InternalErrorCode, "wrapped", cause)
	if !strings.Contains(e.Error(), "root cause") {
		t.Errorf("Error() = %q, want to contain 'root cause'", e.Error())
	}
	if !stderrors.Is(e, cause) {
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

// ---- ErrorCodeFromError ----------------------------------------------------

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
		{stderrors.New("unknown"), InternalErrorCode},
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

// ---- ValidatePayloadVersion ------------------------------------------------

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
		{2, false, false, false, false},
		{4, true, true, false, true},
		{4, true, false, false, false},
		{5, true, true, true, true},
		{5, true, true, false, false},
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

// ---- ErrorName -------------------------------------------------------------

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
