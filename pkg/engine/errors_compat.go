package engine

// errors_compat.go re-exports errors and codes from engine/errors for backward compatibility.

import engerrors "github.com/eth2030/eth2030/engine/errors"

// Engine API errors.
var (
	ErrInvalidParams            = engerrors.ErrInvalidParams
	ErrUnknownPayload           = engerrors.ErrUnknownPayload
	ErrInvalidForkchoiceState   = engerrors.ErrInvalidForkchoiceState
	ErrInvalidPayloadAttributes = engerrors.ErrInvalidPayloadAttributes
	ErrTooLargeRequest          = engerrors.ErrTooLargeRequest
	ErrUnsupportedFork          = engerrors.ErrUnsupportedFork
	ErrInvalidBlockHash         = engerrors.ErrInvalidBlockHash
	ErrInvalidBlobHashes        = engerrors.ErrInvalidBlobHashes
	ErrMissingBeaconRoot        = engerrors.ErrMissingBeaconRoot
)

// Standard JSON-RPC 2.0 error codes.
const (
	ParseErrorCode     = engerrors.ParseErrorCode
	InvalidRequestCode = engerrors.InvalidRequestCode
	MethodNotFoundCode = engerrors.MethodNotFoundCode
	InvalidParamsCode  = engerrors.InvalidParamsCode
	InternalErrorCode  = engerrors.InternalErrorCode
)

// Engine API specific error codes.
const (
	UnknownPayloadCode          = engerrors.UnknownPayloadCode
	InvalidForkchoiceStateCode  = engerrors.InvalidForkchoiceStateCode
	InvalidPayloadAttributeCode = engerrors.InvalidPayloadAttributeCode
	TooLargeRequestCode         = engerrors.TooLargeRequestCode
	UnsupportedForkCode         = engerrors.UnsupportedForkCode
)
