// glamsterdam.go implements Engine API post-Glamsterdam handler types and logic.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/engine/apierrors"
	"github.com/eth2030/eth2030/engine/payload"
)

// jsonrpcError represents a JSON-RPC 2.0 error object (local to this package).
type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// GlamsterdamBackend defines the backend interface for post-Glamsterdam Engine API.
// Implementations must be safe for concurrent use.
type GlamsterdamBackend interface {
	// NewPayloadV5 validates and executes a post-Glamsterdam payload with
	// blob versioned hashes, parent beacon block root, and execution requests.
	NewPayloadV5(p *payload.ExecutionPayloadV5,
		expectedBlobVersionedHashes []types.Hash,
		parentBeaconBlockRoot types.Hash,
		executionRequests [][]byte) (*PayloadStatusV1, error)

	// ForkchoiceUpdatedV4G processes a forkchoice update with V4 attributes
	// including withdrawals, parent beacon block root, and slot number.
	ForkchoiceUpdatedV4G(state *ForkchoiceStateV1, attrs *GlamsterdamPayloadAttributes) (*ForkchoiceUpdatedResult, error)

	// GetPayloadV5 retrieves a previously built payload by ID.
	// Returns execution payload, block value, blobs bundle, and execution requests.
	GetPayloadV5(id payload.PayloadID) (*GetPayloadV5Response, error)

	// GetBlobsV2 retrieves blobs by versioned hashes from the blob pool.
	// Returns nil for the entire result if any blob is missing (all-or-nothing).
	GetBlobsV2(versionedHashes []types.Hash) ([]*BlobAndProofV2, error)
}

// PayloadStatusV1 is the response to engine_newPayload.
type PayloadStatusV1 struct {
	Status          string      `json:"status"`
	LatestValidHash *types.Hash `json:"latestValidHash,omitempty"`
	ValidationError *string     `json:"validationError,omitempty"`
}

// ForkchoiceStateV1 represents the fork choice state from the consensus layer.
type ForkchoiceStateV1 struct {
	HeadBlockHash      types.Hash `json:"headBlockHash"`
	SafeBlockHash      types.Hash `json:"safeBlockHash"`
	FinalizedBlockHash types.Hash `json:"finalizedBlockHash"`
}

// ForkchoiceUpdatedResult is the response to engine_forkchoiceUpdated.
type ForkchoiceUpdatedResult struct {
	PayloadStatus PayloadStatusV1    `json:"payloadStatus"`
	PayloadID     *payload.PayloadID `json:"payloadId,omitempty"`
}

// BlobAndProofV2 represents a blob with its cell proofs (Osaka spec).
type BlobAndProofV2 struct {
	// Blob is the raw blob data (131072 bytes).
	Blob []byte `json:"blob"`
	// Proofs contains KZG cell proofs for the blob (CELLS_PER_EXT_BLOB proofs).
	Proofs [][]byte `json:"proofs"`
}

// BlobsBundleV2 extends BlobsBundleV1 with cell proofs per EIP-7594.
type BlobsBundleV2 struct {
	Commitments [][]byte `json:"commitments"`
	Proofs      [][]byte `json:"proofs"`
	Blobs       [][]byte `json:"blobs"`
}

// GlamsterdamPayloadAttributes contains attributes for building a post-Glamsterdam
// payload. Extends V3 with targetBlobCount and slot number.
type GlamsterdamPayloadAttributes struct {
	// Timestamp for the new payload.
	Timestamp uint64 `json:"timestamp"`
	// PrevRandao for the new payload.
	PrevRandao types.Hash `json:"prevRandao"`
	// SuggestedFeeRecipient for the new payload.
	SuggestedFeeRecipient types.Address `json:"suggestedFeeRecipient"`
	// Withdrawals to process in this payload.
	Withdrawals []*payload.Withdrawal `json:"withdrawals"`
	// ParentBeaconBlockRoot is the root of the parent beacon block.
	ParentBeaconBlockRoot types.Hash `json:"parentBeaconBlockRoot"`
	// TargetBlobCount is the target number of blobs per block.
	TargetBlobCount uint64 `json:"targetBlobCount"`
	// SlotNumber is the slot for this payload (Amsterdam PayloadAttributesV4).
	SlotNumber uint64 `json:"slotNumber"`
}

// GetPayloadV5Response is the response for engine_getPayloadV5 (Osaka spec).
type GetPayloadV5Response struct {
	ExecutionPayload  *payload.ExecutionPayloadV3 `json:"executionPayload"`
	BlockValue        []byte                      `json:"blockValue"`
	BlobsBundle       *BlobsBundleV2              `json:"blobsBundle"`
	Override          bool                        `json:"shouldOverrideBuilder"`
	ExecutionRequests [][]byte                    `json:"executionRequests"`
}

// ClientVersionV2 extends ClientVersionV1 with additional fields.
type ClientVersionV2 struct {
	Code         string   `json:"code"`
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Commit       string   `json:"commit"`
	OS           string   `json:"os"`
	Language     string   `json:"language"`
	Capabilities []string `json:"capabilities"`
}

// ClientVersionV1 represents the client version information.
type ClientVersionV1 struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

// EngineGlamsterdam provides the post-Glamsterdam Engine API methods.
// Thread-safe: all state is protected by a mutex.
type EngineGlamsterdam struct {
	mu      sync.Mutex
	backend GlamsterdamBackend
}

// NewEngineGlamsterdam creates a new post-Glamsterdam Engine API handler.
func NewEngineGlamsterdam(backend GlamsterdamBackend) *EngineGlamsterdam {
	return &EngineGlamsterdam{
		backend: backend,
	}
}

// HandleNewPayloadV5 validates and executes a post-Glamsterdam execution payload.
func (e *EngineGlamsterdam) HandleNewPayloadV5(
	p *payload.ExecutionPayloadV5,
	expectedBlobVersionedHashes []types.Hash,
	parentBeaconBlockRoot types.Hash,
	executionRequests [][]byte,
) (*PayloadStatusV1, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if p == nil {
		return nil, apierrors.ErrInvalidParams
	}

	// EIP-4788: parentBeaconBlockRoot must be provided (non-zero).
	if parentBeaconBlockRoot == (types.Hash{}) {
		return nil, apierrors.ErrInvalidParams
	}

	// EIP-7685: executionRequests must be provided (can be empty, not nil).
	if executionRequests == nil {
		return nil, apierrors.ErrInvalidParams
	}

	// Validate execution request ordering.
	if err := validateExecutionRequestsGlamsterdam(executionRequests); err != nil {
		return nil, apierrors.ErrInvalidParams
	}

	// Block access list must be present for Amsterdam payloads.
	if p.BlockAccessList == nil {
		return nil, apierrors.ErrInvalidParams
	}

	return e.backend.NewPayloadV5(p, expectedBlobVersionedHashes, parentBeaconBlockRoot, executionRequests)
}

// HandleForkchoiceUpdatedV4 processes a forkchoice state update with
// post-Glamsterdam payload attributes.
func (e *EngineGlamsterdam) HandleForkchoiceUpdatedV4(
	state *ForkchoiceStateV1,
	attrs *GlamsterdamPayloadAttributes,
) (*ForkchoiceUpdatedResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if state == nil {
		return nil, apierrors.ErrInvalidForkchoiceState
	}

	// Head block hash must be non-zero.
	if state.HeadBlockHash == (types.Hash{}) {
		return nil, apierrors.ErrInvalidForkchoiceState
	}

	// Validate attributes if provided.
	if attrs != nil {
		if attrs.Timestamp == 0 {
			return nil, apierrors.ErrInvalidPayloadAttributes
		}
		// ParentBeaconBlockRoot must be provided for V4.
		if attrs.ParentBeaconBlockRoot == (types.Hash{}) {
			return nil, apierrors.ErrInvalidPayloadAttributes
		}
	}

	return e.backend.ForkchoiceUpdatedV4G(state, attrs)
}

// HandleGetPayloadV5 retrieves a previously built payload by ID.
func (e *EngineGlamsterdam) HandleGetPayloadV5(payloadID payload.PayloadID) (*GetPayloadV5Response, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if payloadID == (payload.PayloadID{}) {
		return nil, apierrors.ErrUnknownPayload
	}

	return e.backend.GetPayloadV5(payloadID)
}

// HandleGetBlobsV2 retrieves blobs by versioned hashes from the blob pool.
func (e *EngineGlamsterdam) HandleGetBlobsV2(versionedHashes []types.Hash) ([]*BlobAndProofV2, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if versionedHashes == nil {
		return nil, apierrors.ErrInvalidParams
	}

	// Spec: MUST support at least 128, MUST return TooLargeRequest if exceeded.
	if len(versionedHashes) > 128 {
		return nil, apierrors.ErrTooLargeRequest
	}

	return e.backend.GetBlobsV2(versionedHashes)
}

// HandleGetClientVersionV2 returns extended client version information.
func (e *EngineGlamsterdam) HandleGetClientVersionV2(_ *ClientVersionV1) []ClientVersionV2 {
	e.mu.Lock()
	defer e.mu.Unlock()

	return []ClientVersionV2{
		{
			Code:     "ET",
			Name:     "ETH2030",
			Version:  "v0.2.0",
			Commit:   "000000",
			OS:       "linux",
			Language: "go",
			Capabilities: []string{
				"engine_newPayloadV5",
				"engine_forkchoiceUpdatedV4",
				"engine_getPayloadV5",
				"engine_getBlobsV2",
				"engine_getClientVersionV2",
			},
		},
	}
}

// HandleJSONRPC dispatches a JSON-RPC request to the appropriate handler method.
func (e *EngineGlamsterdam) HandleJSONRPC(method string, params []json.RawMessage) (any, *jsonrpcError) {
	switch method {
	case "engine_newPayloadV5":
		return e.handleNewPayloadV5RPC(params)
	case "engine_forkchoiceUpdatedV4":
		return e.handleForkchoiceUpdatedV4RPC(params)
	case "engine_getPayloadV5":
		return e.handleGetPayloadV5RPC(params)
	case "engine_getBlobsV2":
		return e.handleGetBlobsV2RPC(params)
	case "engine_getClientVersionV2":
		return e.handleGetClientVersionV2RPC(params)
	default:
		return nil, &jsonrpcError{
			Code:    apierrors.MethodNotFoundCode,
			Message: fmt.Sprintf("method %q not found", method),
		}
	}
}

func (e *EngineGlamsterdam) handleNewPayloadV5RPC(params []json.RawMessage) (any, *jsonrpcError) {
	if len(params) != 4 {
		return nil, &jsonrpcError{
			Code:    apierrors.InvalidParamsCode,
			Message: fmt.Sprintf("expected 4 params, got %d", len(params)),
		}
	}

	var p payload.ExecutionPayloadV5
	if err := json.Unmarshal(params[0], &p); err != nil {
		return nil, &jsonrpcError{Code: apierrors.InvalidParamsCode, Message: fmt.Sprintf("invalid payload: %v", err)}
	}

	var hashes []types.Hash
	if err := json.Unmarshal(params[1], &hashes); err != nil {
		return nil, &jsonrpcError{Code: apierrors.InvalidParamsCode, Message: fmt.Sprintf("invalid blob hashes: %v", err)}
	}

	var root types.Hash
	if err := json.Unmarshal(params[2], &root); err != nil {
		return nil, &jsonrpcError{Code: apierrors.InvalidParamsCode, Message: fmt.Sprintf("invalid beacon root: %v", err)}
	}

	var requests [][]byte
	if err := json.Unmarshal(params[3], &requests); err != nil {
		return nil, &jsonrpcError{Code: apierrors.InvalidParamsCode, Message: fmt.Sprintf("invalid requests: %v", err)}
	}

	result, err := e.HandleNewPayloadV5(&p, hashes, root, requests)
	if err != nil {
		return nil, glamsterdamErrorToRPC(err)
	}
	return result, nil
}

func (e *EngineGlamsterdam) handleForkchoiceUpdatedV4RPC(params []json.RawMessage) (any, *jsonrpcError) {
	if len(params) < 1 || len(params) > 2 {
		return nil, &jsonrpcError{
			Code:    apierrors.InvalidParamsCode,
			Message: fmt.Sprintf("expected 1-2 params, got %d", len(params)),
		}
	}

	var state ForkchoiceStateV1
	if err := json.Unmarshal(params[0], &state); err != nil {
		return nil, &jsonrpcError{Code: apierrors.InvalidParamsCode, Message: fmt.Sprintf("invalid state: %v", err)}
	}

	var attrs *GlamsterdamPayloadAttributes
	if len(params) == 2 && string(params[1]) != "null" {
		attrs = new(GlamsterdamPayloadAttributes)
		if err := json.Unmarshal(params[1], attrs); err != nil {
			return nil, &jsonrpcError{Code: apierrors.InvalidParamsCode, Message: fmt.Sprintf("invalid attrs: %v", err)}
		}
	}

	result, err := e.HandleForkchoiceUpdatedV4(&state, attrs)
	if err != nil {
		return nil, glamsterdamErrorToRPC(err)
	}
	return result, nil
}

func (e *EngineGlamsterdam) handleGetPayloadV5RPC(params []json.RawMessage) (any, *jsonrpcError) {
	if len(params) != 1 {
		return nil, &jsonrpcError{
			Code:    apierrors.InvalidParamsCode,
			Message: fmt.Sprintf("expected 1 param, got %d", len(params)),
		}
	}

	var payloadID payload.PayloadID
	if err := json.Unmarshal(params[0], &payloadID); err != nil {
		return nil, &jsonrpcError{Code: apierrors.InvalidParamsCode, Message: fmt.Sprintf("invalid ID: %v", err)}
	}

	result, err := e.HandleGetPayloadV5(payloadID)
	if err != nil {
		return nil, glamsterdamErrorToRPC(err)
	}
	return result, nil
}

func (e *EngineGlamsterdam) handleGetBlobsV2RPC(params []json.RawMessage) (any, *jsonrpcError) {
	if len(params) != 1 {
		return nil, &jsonrpcError{
			Code:    apierrors.InvalidParamsCode,
			Message: fmt.Sprintf("expected 1 param, got %d", len(params)),
		}
	}

	var hashes []types.Hash
	if err := json.Unmarshal(params[0], &hashes); err != nil {
		return nil, &jsonrpcError{Code: apierrors.InvalidParamsCode, Message: fmt.Sprintf("invalid hashes: %v", err)}
	}

	result, err := e.HandleGetBlobsV2(hashes)
	if err != nil {
		return nil, glamsterdamErrorToRPC(err)
	}
	return result, nil
}

func (e *EngineGlamsterdam) handleGetClientVersionV2RPC(params []json.RawMessage) (any, *jsonrpcError) {
	var peerVersion *ClientVersionV1
	if len(params) > 0 {
		peerVersion = new(ClientVersionV1)
		if err := json.Unmarshal(params[0], peerVersion); err != nil {
			return nil, &jsonrpcError{Code: apierrors.InvalidParamsCode, Message: fmt.Sprintf("invalid version: %v", err)}
		}
	}
	return e.HandleGetClientVersionV2(peerVersion), nil
}

// validateExecutionRequestsGlamsterdam checks that execution requests are well-formed per EIP-7685.
func validateExecutionRequestsGlamsterdam(requests [][]byte) error {
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
func glamsterdamErrorToRPC(err error) *jsonrpcError {
	code := apierrors.InternalErrorCode
	switch {
	case errors.Is(err, apierrors.ErrUnknownPayload):
		code = apierrors.UnknownPayloadCode
	case errors.Is(err, apierrors.ErrInvalidForkchoiceState):
		code = apierrors.InvalidForkchoiceStateCode
	case errors.Is(err, apierrors.ErrInvalidPayloadAttributes):
		code = apierrors.InvalidPayloadAttributeCode
	case errors.Is(err, apierrors.ErrInvalidParams):
		code = apierrors.InvalidParamsCode
	case errors.Is(err, apierrors.ErrTooLargeRequest):
		code = apierrors.TooLargeRequestCode
	case errors.Is(err, apierrors.ErrUnsupportedFork):
		code = apierrors.UnsupportedForkCode
	}
	return &jsonrpcError{Code: code, Message: err.Error()}
}
