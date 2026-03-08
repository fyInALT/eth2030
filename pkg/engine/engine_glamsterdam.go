package engine

import (
	"errors"
	"fmt"

	"github.com/eth2030/eth2030/core/types"
	engapi "github.com/eth2030/eth2030/engine/api"
)

// GlamsterdamBackend defines the backend interface for post-Glamsterdam Engine API.
// This is an alias of the interface defined in engine/api — redeclared here
// so callers in package engine can implement it without importing engine/api.
type GlamsterdamBackend interface {
	NewPayloadV5(payload *ExecutionPayloadV5,
		expectedBlobVersionedHashes []types.Hash,
		parentBeaconBlockRoot types.Hash,
		executionRequests [][]byte) (*engapi.PayloadStatusV1, error)

	ForkchoiceUpdatedV4G(state *engapi.ForkchoiceStateV1, attrs *GlamsterdamPayloadAttributes) (*engapi.ForkchoiceUpdatedResult, error)

	GetPayloadV5(id PayloadID) (*GetPayloadV5Response, error)

	GetBlobsV2(versionedHashes []types.Hash) ([]*BlobAndProofV2, error)
}

// glamsterdamBridge adapts the engine-package GlamsterdamBackend to the
// engapi.GlamsterdamBackend interface required by EngineGlamsterdam.
type glamsterdamBridge struct {
	inner GlamsterdamBackend
}

func (b *glamsterdamBridge) NewPayloadV5(
	p *ExecutionPayloadV5,
	hashes []types.Hash,
	root types.Hash,
	reqs [][]byte,
) (*engapi.PayloadStatusV1, error) {
	return b.inner.NewPayloadV5(p, hashes, root, reqs)
}

func (b *glamsterdamBridge) ForkchoiceUpdatedV4G(
	state *engapi.ForkchoiceStateV1,
	attrs *GlamsterdamPayloadAttributes,
) (*engapi.ForkchoiceUpdatedResult, error) {
	return b.inner.ForkchoiceUpdatedV4G(state, attrs)
}

func (b *glamsterdamBridge) GetPayloadV5(id PayloadID) (*GetPayloadV5Response, error) {
	return b.inner.GetPayloadV5(id)
}

func (b *glamsterdamBridge) GetBlobsV2(hashes []types.Hash) ([]*BlobAndProofV2, error) {
	return b.inner.GetBlobsV2(hashes)
}

// NewEngineGlamsterdam creates a new post-Glamsterdam Engine API handler.
func NewEngineGlamsterdam(backend GlamsterdamBackend) *EngineGlamsterdam {
	return engapi.NewEngineGlamsterdam(&glamsterdamBridge{inner: backend})
}

// validateExecutionRequests checks that execution requests are well-formed per EIP-7685.
// This unexported function is kept here so package-internal tests can call it directly.
func validateExecutionRequests(requests [][]byte) error {
	if len(requests) == 0 {
		return nil
	}
	var lastType byte
	for i, req := range requests {
		if len(req) <= 1 {
			return fmt.Errorf("request at index %d too short", i)
		}
		reqType := req[0]
		if i > 0 && reqType <= lastType {
			return fmt.Errorf("request types not ascending at index %d", i)
		}
		lastType = reqType
	}
	return nil
}

// glamsterdamErrorToRPC maps engine errors to JSON-RPC error responses.
// This unexported function is kept here so package-internal tests can call it directly.
func glamsterdamErrorToRPC(err error) *jsonrpcError {
	switch {
	case errors.Is(err, ErrUnknownPayload):
		return &jsonrpcError{Code: UnknownPayloadCode, Message: err.Error()}
	case errors.Is(err, ErrInvalidForkchoiceState):
		return &jsonrpcError{Code: InvalidForkchoiceStateCode, Message: err.Error()}
	case errors.Is(err, ErrInvalidPayloadAttributes):
		return &jsonrpcError{Code: InvalidPayloadAttributeCode, Message: err.Error()}
	case errors.Is(err, ErrInvalidParams):
		return &jsonrpcError{Code: InvalidParamsCode, Message: err.Error()}
	case errors.Is(err, ErrTooLargeRequest):
		return &jsonrpcError{Code: TooLargeRequestCode, Message: err.Error()}
	case errors.Is(err, ErrUnsupportedFork):
		return &jsonrpcError{Code: UnsupportedForkCode, Message: err.Error()}
	default:
		return &jsonrpcError{Code: InternalErrorCode, Message: err.Error()}
	}
}
