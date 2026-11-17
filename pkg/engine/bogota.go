package engine

import (
	"encoding/json"
	"fmt"

	"github.com/eth2030/eth2030/core/types"
)

type bogotaBackend interface {
	IsHogota(timestamp uint64) bool
	CheckPayloadInclusionList(types.Hash, [][]byte) (bool, error)
	PayloadInclusionListSatisfied(types.Hash) bool
}

type bogotaExecutionBackend interface {
	ProcessBlockV6(*ExecutionPayloadV5, types.Hash, [][]byte, [][]byte) (PayloadStatusV1, bool, error)
	ForkchoiceUpdatedBogota(ForkchoiceStateV1, *PayloadAttributesV4) (ForkchoiceUpdatedResult, bool, error)
}

type bogotaPayloadStatus struct {
	Status                 string      `json:"status"`
	LatestValidHash        *types.Hash `json:"latestValidHash"`
	ValidationError        *string     `json:"validationError"`
	InclusionListSatisfied *bool       `json:"inclusionListSatisfied"`
}

func payloadStatusV2(status PayloadStatusV1, satisfied bool) bogotaPayloadStatus {
	result := bogotaPayloadStatus{
		Status: status.Status, LatestValidHash: status.LatestValidHash,
		ValidationError: status.ValidationError,
	}
	if status.Status == StatusValid {
		result.InclusionListSatisfied = &satisfied
	} else if status.Status == StatusInvalidBlockHash {
		result.Status = StatusInvalid
	}
	return result
}

func (api *EngineAPI) isBogota(timestamp uint64) bool {
	backend, ok := api.backend.(bogotaBackend)
	return ok && backend.IsHogota(timestamp)
}

func (api *EngineAPI) handleNewPayloadV6(params []json.RawMessage) (any, *jsonrpcError) {
	if len(params) != 5 {
		return nil, &jsonrpcError{Code: InvalidParamsCode, Message: "engine_newPayloadV6 requires five parameters"}
	}
	var p ExecutionPayloadV5
	var hashes []types.Hash
	var parent types.Hash
	for i, target := range []any{&p, &hashes, &parent} {
		if err := json.Unmarshal(params[i], target); err != nil {
			return nil, &jsonrpcError{Code: InvalidParamsCode, Message: fmt.Sprintf("invalid parameter %d: %v", i, err)}
		}
	}
	if hashes == nil {
		return nil, &jsonrpcError{Code: InvalidParamsCode, Message: "expectedBlobVersionedHashes must be an array"}
	}
	requests, err := parseExecutionRequests(params[3])
	if err != nil {
		return nil, &jsonrpcError{Code: InvalidParamsCode, Message: "invalid execution requests: " + err.Error()}
	}
	transactions, err := parseExecutionRequests(params[4])
	if err != nil || transactions == nil {
		return nil, &jsonrpcError{Code: InvalidParamsCode, Message: "inclusionListTransactions must be an array of DATA"}
	}
	if !api.isBogota(p.Timestamp) || !api.backend.IsAmsterdam(p.Timestamp) {
		return nil, engineErrorToRPC(ErrUnsupportedFork)
	}
	if backend, ok := api.backend.(bogotaExecutionBackend); ok {
		var satisfied bool
		status, err := api.processAmsterdamPayloadWith(p, hashes, parent, requests,
			func(payload *ExecutionPayloadV5, _ []types.Hash, root types.Hash, requests [][]byte) (PayloadStatusV1, error) {
				status, verdict, err := backend.ProcessBlockV6(payload, root, requests, transactions)
				satisfied = verdict
				return status, err
			})
		if err != nil {
			return nil, engineErrorToRPC(err)
		}
		return payloadStatusV2(status, satisfied), nil
	}
	status, err := api.processAmsterdamPayload(p, hashes, parent, requests)
	if err != nil {
		return nil, engineErrorToRPC(err)
	}
	var satisfied bool
	if status.Status == StatusValid {
		satisfied, err = api.backend.(bogotaBackend).CheckPayloadInclusionList(p.BlockHash, transactions)
		if err != nil {
			return nil, engineErrorToRPC(err)
		}
	}
	return payloadStatusV2(status, satisfied), nil
}

func validateBogotaAttributes(raw json.RawMessage) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	for _, field := range []string{
		"timestamp", "prevRandao", "suggestedFeeRecipient", "withdrawals",
		"parentBeaconBlockRoot", "slotNumber", "targetGasLimit", "inclusionListTransactions",
	} {
		if value := fields[field]; len(value) == 0 || string(value) == "null" {
			return fmt.Errorf("missing %s", field)
		}
	}
	var typed PayloadAttributesV4
	err := json.Unmarshal(raw, &typed)
	return err
}

func (api *EngineAPI) forkchoiceUpdatedBogota(
	state ForkchoiceStateV1,
	attrs *PayloadAttributesV4,
) (any, *jsonrpcError) {
	var result ForkchoiceUpdatedResult
	var err error
	if backend, ok := api.backend.(bogotaExecutionBackend); ok {
		badFork := attrs != nil && (!api.isBogota(attrs.Timestamp) || !api.backend.IsAmsterdam(attrs.Timestamp))
		if badFork {
			attrs = nil
		}
		result, satisfied, err := backend.ForkchoiceUpdatedBogota(state, attrs)
		if badFork {
			result, err = forkchoiceAttributeError(result, err, ErrUnsupportedFork)
		}
		if err != nil {
			return nil, engineErrorToRPC(err)
		}
		return bogotaForkchoiceResponse(result, satisfied), nil
	}
	if attrs != nil && (!api.isBogota(attrs.Timestamp) || !api.backend.IsAmsterdam(attrs.Timestamp)) {
		result, err = api.backend.ForkchoiceUpdatedV4(state, nil)
		result, err = forkchoiceAttributeError(result, err, ErrUnsupportedFork)
	} else {
		result, err = api.backend.ForkchoiceUpdatedV4(state, attrs)
	}
	if err != nil {
		return nil, engineErrorToRPC(err)
	}
	return api.bogotaForkchoiceResult(state, result), nil
}

func (api *EngineAPI) bogotaForkchoiceResult(state ForkchoiceStateV1, result ForkchoiceUpdatedResult) any {
	satisfied := result.PayloadStatus.Status == StatusValid &&
		api.backend.(bogotaBackend).PayloadInclusionListSatisfied(state.HeadBlockHash)
	return bogotaForkchoiceResponse(result, satisfied)
}

func bogotaForkchoiceResponse(result ForkchoiceUpdatedResult, satisfied bool) any {
	return struct {
		PayloadStatus bogotaPayloadStatus `json:"payloadStatus"`
		PayloadID     *PayloadID          `json:"payloadId"`
	}{payloadStatusV2(result.PayloadStatus, satisfied), result.PayloadID}
}
