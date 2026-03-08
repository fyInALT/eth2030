// Package engine defines types for the Engine API (CL-EL communication).
package engine

import (
	"math/big"

	"github.com/eth2030/eth2030/core/types"
	engapi "github.com/eth2030/eth2030/engine/api"
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

// Re-exported type aliases from engine/api sub-package.
type (
	// From api/glamsterdam.go (core types used widely in server.go/handler.go)
	PayloadStatusV1         = engapi.PayloadStatusV1
	ForkchoiceStateV1       = engapi.ForkchoiceStateV1
	ForkchoiceUpdatedResult = engapi.ForkchoiceUpdatedResult
	// From api/glamsterdam.go
	BlobAndProofV2               = engapi.BlobAndProofV2
	BlobsBundleV2                = engapi.BlobsBundleV2
	GlamsterdamPayloadAttributes = engapi.GlamsterdamPayloadAttributes
	GetPayloadV5Response         = engapi.GetPayloadV5Response
	ClientVersionV2              = engapi.ClientVersionV2
	EngineGlamsterdam            = engapi.EngineGlamsterdam
	// From api/v4.go
	DepositRequest       = engapi.DepositRequest
	WithdrawalRequest    = engapi.WithdrawalRequest
	ConsolidationRequest = engapi.ConsolidationRequest
	ExecutionRequestsV4  = engapi.ExecutionRequestsV4
	GetPayloadV4Result   = engapi.GetPayloadV4Result
	EngV4                = engapi.EngV4
	// From api/uncoupled.go
	InclusionProof           = engapi.InclusionProof
	UncoupledPayloadEnvelope = engapi.UncoupledPayloadEnvelope
	UncoupledPayloadHandler  = engapi.UncoupledPayloadHandler
	// From api/v7.go
	DALayerConfig        = engapi.DALayerConfig
	ProofRequirements    = engapi.ProofRequirements
	PayloadAttributesV7  = engapi.PayloadAttributesV7
	ExecutionPayloadV7   = engapi.ExecutionPayloadV7
	GetPayloadV7Response = engapi.GetPayloadV7Response
	// Note: EngineV7 is NOT aliased here; engine_v7.go defines engineV7Wrapper
	// which wraps engapi.EngineV7 and exposes backend for package-internal tests.
	// From api/epbs.go
	GetPayloadHeaderV1Response   = engapi.GetPayloadHeaderV1Response
	SubmitBlindedBlockV1Request  = engapi.SubmitBlindedBlockV1Request
	SubmitBlindedBlockV1Response = engapi.SubmitBlindedBlockV1Response
)

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

// TransitionConfigurationV1 for Engine API transition configuration exchange.
type TransitionConfigurationV1 struct {
	TerminalTotalDifficulty *big.Int   `json:"terminalTotalDifficulty"`
	TerminalBlockHash       types.Hash `json:"terminalBlockHash"`
	TerminalBlockNumber     uint64     `json:"terminalBlockNumber"`
}
