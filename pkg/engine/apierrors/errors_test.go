package apierrors

import (
	"errors"
	"testing"
)

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
			if err != other && errors.Is(err, other) {
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
