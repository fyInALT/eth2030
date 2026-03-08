// v7.go implements Engine API V7 types and handler for the 2030 roadmap (K+ era).
package api

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/engine/apierrors"
	"github.com/eth2030/eth2030/engine/backendapi"
	"github.com/eth2030/eth2030/engine/payload"
)

// Type aliases — canonical definitions live in engine/payload and engine/backendapi.
type (
	EngineV7Backend      = backendapi.EngineV7Backend
	DALayerConfig        = payload.DALayerConfig
	ProofRequirements    = payload.ProofRequirements
	PayloadAttributesV7  = payload.PayloadAttributesV7
	ExecutionPayloadV7   = payload.ExecutionPayloadV7
	GetPayloadV7Response = payload.GetPayloadV7Response
)

// EngineV7 provides the Engine API V7 methods.
// Thread-safe: all state is protected by a mutex.
type EngineV7 struct {
	mu      sync.Mutex
	backend EngineV7Backend
}

// NewEngineV7 creates a new Engine API V7 handler.
func NewEngineV7(backend EngineV7Backend) *EngineV7 {
	return &EngineV7{
		backend: backend,
	}
}

// HandleNewPayloadV7 validates and executes a 2028-era execution payload.
func (e *EngineV7) HandleNewPayloadV7(p *ExecutionPayloadV7) (*PayloadStatusV1, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if p == nil {
		return nil, apierrors.ErrInvalidParams
	}

	// Validate blob commitments are present when blob gas is used.
	if p.BlobGasUsed > 0 && len(p.BlobCommitments) == 0 {
		errMsg := "blob gas used but no blob commitments provided"
		return &PayloadStatusV1{
			Status:          "INVALID",
			ValidationError: &errMsg,
		}, nil
	}

	// Validate proof submissions format.
	if p.ProofSubmissions == nil {
		return nil, apierrors.ErrInvalidParams
	}

	for i, proof := range p.ProofSubmissions {
		if len(proof) == 0 {
			errMsg := fmt.Sprintf("empty proof submission at index %d", i)
			return &PayloadStatusV1{
				Status:          "INVALID",
				ValidationError: &errMsg,
			}, nil
		}
	}

	return e.backend.NewPayloadV7(p)
}

// HandleForkchoiceUpdatedV7 processes a forkchoice state update with V7
// payload attributes.
func (e *EngineV7) HandleForkchoiceUpdatedV7(
	state *ForkchoiceStateV1,
	attrs *PayloadAttributesV7,
) (*ForkchoiceUpdatedResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if state == nil {
		return nil, apierrors.ErrInvalidForkchoiceState
	}

	if state.HeadBlockHash == (types.Hash{}) {
		return nil, apierrors.ErrInvalidForkchoiceState
	}

	if attrs != nil {
		if attrs.Timestamp == 0 {
			return nil, apierrors.ErrInvalidPayloadAttributes
		}

		if attrs.ProofRequirements != nil {
			if err := attrs.ProofRequirements.Validate(); err != nil {
				return nil, apierrors.ErrInvalidPayloadAttributes
			}
		}

		if attrs.DALayerConfig != nil {
			if attrs.DALayerConfig.SampleCount == 0 {
				return nil, apierrors.ErrInvalidPayloadAttributes
			}
			if attrs.DALayerConfig.ColumnCount == 0 {
				return nil, apierrors.ErrInvalidPayloadAttributes
			}
		}
	}

	return e.backend.ForkchoiceUpdatedV7(state, attrs)
}

// HandleGetPayloadV7 retrieves a previously built V7 payload by its ID.
func (e *EngineV7) HandleGetPayloadV7(payloadID payload.PayloadID) (*ExecutionPayloadV7, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if payloadID == (payload.PayloadID{}) {
		return nil, apierrors.ErrUnknownPayload
	}

	return e.backend.GetPayloadV7(payloadID)
}

// GenerateV7PayloadID creates a V7 payload ID from parent hash and timestamp.
func GenerateV7PayloadID(parentHash types.Hash, timestamp uint64) payload.PayloadID {
	var id payload.PayloadID
	binary.BigEndian.PutUint64(id[:], timestamp)
	for i := 0; i < types.HashLength; i++ {
		id[i%8] ^= parentHash[i]
	}
	return id
}
