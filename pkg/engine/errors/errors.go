// Package errors defines Engine API error variables, codes, and utilities.
package errors

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
)

// Standard Engine API error sentinels per the execution-apis spec.
var (
	// ErrInvalidParams is returned when the request parameters are invalid.
	ErrInvalidParams = stderrors.New("invalid params")

	// ErrUnknownPayload is returned when the requested payload is not found.
	ErrUnknownPayload = stderrors.New("unknown payload")

	// ErrInvalidForkchoiceState is returned when the forkchoice state is invalid.
	ErrInvalidForkchoiceState = stderrors.New("invalid forkchoice state")

	// ErrInvalidPayloadAttributes is returned when payload attributes are invalid.
	ErrInvalidPayloadAttributes = stderrors.New("invalid payload attributes")

	// ErrTooLargeRequest is returned when the request size exceeds limits.
	ErrTooLargeRequest = stderrors.New("too large request")

	// ErrUnsupportedFork is returned when the requested fork is not supported.
	ErrUnsupportedFork = stderrors.New("unsupported fork")

	// ErrInvalidBlockHash is returned when the block hash in the payload
	// does not match the computed block hash.
	ErrInvalidBlockHash = stderrors.New("invalid block hash")

	// ErrInvalidBlobHashes is returned when the blob versioned hashes
	// in the payload do not match the expected hashes from the CL.
	ErrInvalidBlobHashes = stderrors.New("invalid blob versioned hashes")

	// ErrMissingBeaconRoot is returned when the parent beacon block root
	// is missing (zero) in a V3+ newPayload call.
	ErrMissingBeaconRoot = stderrors.New("missing parent beacon block root")

	// ErrRequestTooLarge is returned when a request exceeds the maximum allowed size.
	ErrRequestTooLarge = stderrors.New("request too large")

	// ErrServerBusy is returned when the server is too busy to process the request.
	ErrServerBusy = stderrors.New("server busy")

	// ErrRequestTimeout is returned when request processing times out.
	ErrRequestTimeout = stderrors.New("request timeout")

	// ErrPayloadNotBuilding is returned when getPayload is called but no
	// payload is being built for the given ID.
	ErrPayloadNotBuilding = stderrors.New("payload not building")

	// ErrInvalidTerminalBlock is returned when the terminal block does not
	// satisfy the terminal total difficulty condition.
	ErrInvalidTerminalBlock = stderrors.New("invalid terminal block")

	// ErrPayloadTimestamp is returned when the payload timestamp does not
	// advance beyond the parent block's timestamp.
	ErrPayloadTimestamp = stderrors.New("invalid payload timestamp")

	// ErrMissingWithdrawals is returned when withdrawals are expected but
	// not provided in V2+ payloads.
	ErrMissingWithdrawals = stderrors.New("missing withdrawals")

	// ErrMissingExecutionRequests is returned when execution requests are
	// expected but not provided in V4+ payloads.
	ErrMissingExecutionRequests = stderrors.New("missing execution requests")

	// ErrMissingBlockAccessList is returned when the block access list is
	// expected but not provided in V5+ payloads.
	ErrMissingBlockAccessList = stderrors.New("missing block access list")
)

// Payload status strings per the execution-apis spec.
const (
	StatusValid            = "VALID"
	StatusInvalid          = "INVALID"
	StatusSyncing          = "SYNCING"
	StatusAccepted         = "ACCEPTED"
	StatusInvalidBlockHash = "INVALID_BLOCK_HASH"
)

// Standard JSON-RPC 2.0 error codes.
const (
	ParseErrorCode     = -32700
	InvalidRequestCode = -32600
	MethodNotFoundCode = -32601
	InvalidParamsCode  = -32602
	InternalErrorCode  = -32603
)

// Engine API specific error codes (per execution-apis spec).
const (
	UnknownPayloadCode          = -38001
	InvalidForkchoiceStateCode  = -38002
	InvalidPayloadAttributeCode = -38003
	TooLargeRequestCode         = -38004
	UnsupportedForkCode         = -38005
)

// Extended error codes for additional server conditions.
const (
	ServerBusyCode     = -32005
	RequestTimeoutCode = -32006
)

// EngineError is a structured error that carries an Engine API error code
// for proper JSON-RPC encoding.
type EngineError struct {
	Code    int
	Message string
	Cause   error
}

// Error implements the error interface.
func (e *EngineError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying cause for errors.Is/errors.As.
func (e *EngineError) Unwrap() error {
	return e.Cause
}

// MarshalJSON encodes the error as a JSON-RPC error object.
func (e *EngineError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{
		Code:    e.Code,
		Message: e.Error(),
	})
}

// NewEngineError creates a new EngineError with the given code and message.
func NewEngineError(code int, message string) *EngineError {
	return &EngineError{Code: code, Message: message}
}

// WrapEngineError wraps an existing error with an Engine API error code.
func WrapEngineError(code int, message string, cause error) *EngineError {
	return &EngineError{Code: code, Message: message, Cause: cause}
}

// ErrorCodeFromError maps known engine errors to their corresponding
// JSON-RPC error codes.
func ErrorCodeFromError(err error) int {
	if err == nil {
		return 0
	}

	var engineErr *EngineError
	if stderrors.As(err, &engineErr) {
		return engineErr.Code
	}

	switch {
	case stderrors.Is(err, ErrUnknownPayload), stderrors.Is(err, ErrPayloadNotBuilding):
		return UnknownPayloadCode
	case stderrors.Is(err, ErrInvalidForkchoiceState):
		return InvalidForkchoiceStateCode
	case stderrors.Is(err, ErrInvalidPayloadAttributes):
		return InvalidPayloadAttributeCode
	case stderrors.Is(err, ErrTooLargeRequest), stderrors.Is(err, ErrRequestTooLarge):
		return TooLargeRequestCode
	case stderrors.Is(err, ErrUnsupportedFork):
		return UnsupportedForkCode
	case stderrors.Is(err, ErrInvalidParams):
		return InvalidParamsCode
	case stderrors.Is(err, ErrInvalidBlockHash):
		return InvalidParamsCode
	case stderrors.Is(err, ErrInvalidBlobHashes):
		return InvalidParamsCode
	case stderrors.Is(err, ErrMissingBeaconRoot):
		return InvalidParamsCode
	case stderrors.Is(err, ErrServerBusy):
		return ServerBusyCode
	case stderrors.Is(err, ErrRequestTimeout):
		return RequestTimeoutCode
	default:
		return InternalErrorCode
	}
}

// IsClientError returns true if the error code indicates a client-side error.
func IsClientError(code int) bool {
	return code >= -32699 && code <= -32600
}

// IsServerError returns true if the error code indicates a server-side error.
func IsServerError(code int) bool {
	return code >= -32099 && code <= -32000
}

// IsEngineError returns true if the error code is an Engine API specific error.
func IsEngineError(code int) bool {
	return code >= -38005 && code <= -38001
}

// ErrorResponse constructs a full JSON-RPC error response as raw JSON bytes.
func ErrorResponse(id json.RawMessage, code int, message string) []byte {
	resp := struct {
		JSONRPC string          `json:"jsonrpc"`
		Error   interface{}     `json:"error"`
		ID      json.RawMessage `json:"id"`
	}{
		JSONRPC: "2.0",
		Error: struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{Code: code, Message: message},
		ID: id,
	}
	out, _ := json.Marshal(resp)
	return out
}

// ErrorName returns a human-readable name for an Engine API error code.
func ErrorName(code int) string {
	switch code {
	case ParseErrorCode:
		return "ParseError"
	case InvalidRequestCode:
		return "InvalidRequest"
	case MethodNotFoundCode:
		return "MethodNotFound"
	case InvalidParamsCode:
		return "InvalidParams"
	case InternalErrorCode:
		return "InternalError"
	case UnknownPayloadCode:
		return "UnknownPayload"
	case InvalidForkchoiceStateCode:
		return "InvalidForkchoiceState"
	case InvalidPayloadAttributeCode:
		return "InvalidPayloadAttributes"
	case TooLargeRequestCode:
		return "TooLargeRequest"
	case UnsupportedForkCode:
		return "UnsupportedFork"
	case ServerBusyCode:
		return "ServerBusy"
	case RequestTimeoutCode:
		return "RequestTimeout"
	default:
		return fmt.Sprintf("Unknown(%d)", code)
	}
}

// ValidatePayloadVersion checks that required fields are present for the
// given payload version.
func ValidatePayloadVersion(version int, hasWithdrawals, hasExecutionRequests, hasBlockAccessList bool) *EngineError {
	if version >= 2 && !hasWithdrawals {
		return WrapEngineError(InvalidParamsCode, "withdrawals required for V2+ payload", ErrMissingWithdrawals)
	}
	if version >= 4 && !hasExecutionRequests {
		return WrapEngineError(InvalidParamsCode, "execution requests required for V4+ payload", ErrMissingExecutionRequests)
	}
	if version >= 5 && !hasBlockAccessList {
		return WrapEngineError(InvalidParamsCode, "block access list required for V5+ payload", ErrMissingBlockAccessList)
	}
	return nil
}
