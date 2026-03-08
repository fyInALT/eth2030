// v7.go implements Engine API V7 types and handler for the 2030 roadmap (K+ era).
package api

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/engine/apierrors"
	"github.com/eth2030/eth2030/engine/payload"
)

// EngineV7Backend defines the backend interface for Engine API V7.
type EngineV7Backend interface {
	// NewPayloadV7 validates and executes a 2028-era payload.
	NewPayloadV7(p *ExecutionPayloadV7) (*PayloadStatusV1, error)

	// ForkchoiceUpdatedV7 processes a forkchoice update with V7 attributes.
	ForkchoiceUpdatedV7(state *ForkchoiceStateV1, attrs *PayloadAttributesV7) (*ForkchoiceUpdatedResult, error)

	// GetPayloadV7 retrieves a previously built V7 payload by ID.
	GetPayloadV7(id payload.PayloadID) (*ExecutionPayloadV7, error)
}

// DALayerConfig configures the data availability layer for the 2030 roadmap.
type DALayerConfig struct {
	SampleCount       uint64 `json:"sampleCount"`
	ColumnCount       uint64 `json:"columnCount"`
	RecoveryThreshold uint64 `json:"recoveryThreshold"`
}

// ProofRequirements specifies the mandatory proof parameters per the K+ era.
type ProofRequirements struct {
	MinProofs    uint64   `json:"minProofs"`
	TotalProofs  uint64   `json:"totalProofs"`
	AllowedTypes []string `json:"allowedTypes"`
}

// Validate checks that proof requirements are internally consistent.
func (pr *ProofRequirements) Validate() error {
	if pr.TotalProofs == 0 {
		return errors.New("engine: totalProofs must be > 0")
	}
	if pr.MinProofs == 0 {
		return errors.New("engine: minProofs must be > 0")
	}
	if pr.MinProofs > pr.TotalProofs {
		return fmt.Errorf("engine: minProofs (%d) > totalProofs (%d)", pr.MinProofs, pr.TotalProofs)
	}
	return nil
}

// PayloadAttributesV7 extends V3 attributes with 2030 roadmap features.
type PayloadAttributesV7 struct {
	payload.PayloadAttributesV3

	DALayerConfig     *DALayerConfig     `json:"daLayerConfig,omitempty"`
	ProofRequirements *ProofRequirements `json:"proofRequirements,omitempty"`
	ShieldedTxs       [][]byte           `json:"shieldedTxs,omitempty"`
}

// ExecutionPayloadV7 extends V3 with 2030 roadmap fields.
type ExecutionPayloadV7 struct {
	payload.ExecutionPayloadV3

	BlobCommitments  []types.Hash `json:"blobCommitments"`
	ProofSubmissions [][]byte     `json:"proofSubmissions"`
	ShieldedResults  []types.Hash `json:"shieldedResults"`
}

// GetPayloadV7Response is the response for engine_getPayloadV7.
type GetPayloadV7Response struct {
	ExecutionPayload *ExecutionPayloadV7    `json:"executionPayload"`
	BlockValue       []byte                 `json:"blockValue"`
	BlobsBundle      *payload.BlobsBundleV1 `json:"blobsBundle"`
	Override         bool                   `json:"shouldOverrideBuilder"`
}

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
