// Package payload: status types shared across Engine API handlers.
package payload

import "github.com/eth2030/eth2030/core/types"

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
	PayloadStatus PayloadStatusV1 `json:"payloadStatus"`
	PayloadID     *PayloadID      `json:"payloadId,omitempty"`
}
