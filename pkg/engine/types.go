// Package engine defines types for the Engine API (CL-EL communication).
package engine

import (
	"math/big"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/engine/payload"
)

// Re-exported type aliases for backward compatibility.
// The canonical definitions live in engine/payload.
type (
	PayloadID            = payload.PayloadID
	Withdrawal           = payload.Withdrawal
	ExecutionPayloadV1   = payload.ExecutionPayloadV1
	ExecutionPayloadV2   = payload.ExecutionPayloadV2
	ExecutionPayloadV3   = payload.ExecutionPayloadV3
	ExecutionPayloadV4   = payload.ExecutionPayloadV4
	ExecutionPayloadV5   = payload.ExecutionPayloadV5
	PayloadAttributesV1  = payload.PayloadAttributesV1
	PayloadAttributesV2  = payload.PayloadAttributesV2
	PayloadAttributesV3  = payload.PayloadAttributesV3
	PayloadAttributesV4  = payload.PayloadAttributesV4
	GetPayloadV3Response = payload.GetPayloadV3Response
	GetPayloadV4Response = payload.GetPayloadV4Response
	GetPayloadV6Response = payload.GetPayloadV6Response
	GetPayloadResponse   = payload.GetPayloadResponse
	BlobsBundleV1        = payload.BlobsBundleV1
)

// ForkchoiceStateV1 represents the fork choice state from the consensus layer.
type ForkchoiceStateV1 struct {
	HeadBlockHash      types.Hash `json:"headBlockHash"`
	SafeBlockHash      types.Hash `json:"safeBlockHash"`
	FinalizedBlockHash types.Hash `json:"finalizedBlockHash"`
}

// PayloadStatus values.
const (
	StatusValid            = "VALID"
	StatusInvalid          = "INVALID"
	StatusSyncing          = "SYNCING"
	StatusAccepted         = "ACCEPTED"
	StatusInvalidBlockHash = "INVALID_BLOCK_HASH"
	// StatusInclusionListUnsatisfied is returned when a valid IL tx is absent
	// from the block with sufficient remaining gas (EIP-7805 §engine-api).
	StatusInclusionListUnsatisfied = "INCLUSION_LIST_UNSATISFIED"
)

// PayloadStatusV1 is the response to engine_newPayload.
type PayloadStatusV1 struct {
	Status          string      `json:"status"`
	LatestValidHash *types.Hash `json:"latestValidHash,omitempty"`
	ValidationError *string     `json:"validationError,omitempty"`
}

// ForkchoiceUpdatedResult is the response to engine_forkchoiceUpdated.
type ForkchoiceUpdatedResult struct {
	PayloadStatus PayloadStatusV1 `json:"payloadStatus"`
	PayloadID     *PayloadID      `json:"payloadId,omitempty"`
}

// TransitionConfigurationV1 for Engine API transition configuration exchange.
type TransitionConfigurationV1 struct {
	TerminalTotalDifficulty *big.Int   `json:"terminalTotalDifficulty"`
	TerminalBlockHash       types.Hash `json:"terminalBlockHash"`
	TerminalBlockNumber     uint64     `json:"terminalBlockNumber"`
}
