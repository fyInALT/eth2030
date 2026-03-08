// Package payload: Engine API V7 payload types (K+ era).
package payload

import (
	"errors"
	"fmt"

	"github.com/eth2030/eth2030/core/types"
)

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
	PayloadAttributesV3

	DALayerConfig     *DALayerConfig     `json:"daLayerConfig,omitempty"`
	ProofRequirements *ProofRequirements `json:"proofRequirements,omitempty"`
	ShieldedTxs       [][]byte           `json:"shieldedTxs,omitempty"`
}

// ExecutionPayloadV7 extends V3 with 2030 roadmap fields.
type ExecutionPayloadV7 struct {
	ExecutionPayloadV3

	BlobCommitments  []types.Hash `json:"blobCommitments"`
	ProofSubmissions [][]byte     `json:"proofSubmissions"`
	ShieldedResults  []types.Hash `json:"shieldedResults"`
}

// GetPayloadV7Response is the response for engine_getPayloadV7.
type GetPayloadV7Response struct {
	ExecutionPayload *ExecutionPayloadV7 `json:"executionPayload"`
	BlockValue       []byte              `json:"blockValue"`
	BlobsBundle      *BlobsBundleV1      `json:"blobsBundle"`
	Override         bool                `json:"shouldOverrideBuilder"`
}
