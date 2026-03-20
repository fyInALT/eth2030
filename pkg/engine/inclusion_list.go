package engine

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/engine/backendapi"
)

// flexibleUint64 can unmarshal from a JSON number, decimal string, or hex string.
// This handles different serialization formats from CL implementations.
type flexibleUint64 uint64

func (f flexibleUint64) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(uint64(f), 10))
}

func (f *flexibleUint64) UnmarshalJSON(data []byte) error {
	// Try JSON number first
	var num uint64
	if err := json.Unmarshal(data, &num); err == nil {
		*f = flexibleUint64(num)
		return nil
	}

	// Try string
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("flexibleUint64: expected number or string, got %s", string(data))
	}

	// Try hex prefix
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		n := new(big.Int)
		if _, ok := n.SetString(s[2:], 16); !ok {
			return fmt.Errorf("flexibleUint64: invalid hex %q", s)
		}
		*f = flexibleUint64(n.Uint64())
		return nil
	}

	// Try decimal
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return fmt.Errorf("flexibleUint64: invalid decimal %q: %w", s, err)
	}
	*f = flexibleUint64(n)
	return nil
}

// InclusionListV1 is the Engine API representation of an inclusion list.
// Sent from the CL to the EL via engine_newInclusionListV1.
type InclusionListV1 struct {
	Slot           flexibleUint64 `json:"slot"`
	ValidatorIndex flexibleUint64 `json:"validatorIndex"`
	CommitteeRoot  types.Hash     `json:"inclusionListCommitteeRoot"`
	Transactions   [][]byte       `json:"transactions"`
}

// InclusionListStatusV1 is the response to engine_newInclusionListV1.
type InclusionListStatusV1 struct {
	Status string  `json:"status"`
	Error  *string `json:"error,omitempty"`
}

// GetInclusionListResponseV1 is the response to engine_getInclusionListV1.
// Transactions are hex-encoded strings with 0x prefix (JSON-RPC convention).
type GetInclusionListResponseV1 struct {
	Transactions []string `json:"transactions"`
}

// Inclusion list status values.
const (
	ILStatusAccepted = "ACCEPTED"
	ILStatusInvalid  = "INVALID"
)

// ToCore converts the Engine API inclusion list to the core types representation.
func (il *InclusionListV1) ToCore() *types.InclusionList {
	return &types.InclusionList{
		Slot:           uint64(il.Slot),
		ValidatorIndex: uint64(il.ValidatorIndex),
		CommitteeRoot:  il.CommitteeRoot,
		Transactions:   il.Transactions,
	}
}

// InclusionListFromCore converts a core types InclusionList to Engine API format.
func InclusionListFromCore(il *types.InclusionList) *InclusionListV1 {
	return &InclusionListV1{
		Slot:           flexibleUint64(il.Slot),
		ValidatorIndex: flexibleUint64(il.ValidatorIndex),
		CommitteeRoot:  il.CommitteeRoot,
		Transactions:   il.Transactions,
	}
}

// NewInclusionListV1 receives and validates a new inclusion list from the CL.
func (api *EngineAPI) NewInclusionListV1(il InclusionListV1) (InclusionListStatusV1, error) {
	coreIL := il.ToCore()

	// Validate that the backend supports inclusion list handling.
	ilBackend, ok := api.backend.(InclusionListBackend)
	if !ok {
		return InclusionListStatusV1{
			Status: ILStatusInvalid,
			Error:  strPtr("inclusion lists not supported"),
		}, nil
	}

	err := ilBackend.ProcessInclusionList(coreIL)
	if err != nil {
		return InclusionListStatusV1{
			Status: ILStatusInvalid,
			Error:  strPtr(err.Error()),
		}, nil
	}

	return InclusionListStatusV1{Status: ILStatusAccepted}, nil
}

// GetInclusionListV1 returns an inclusion list generated from the EL's mempool.
// Called by CL validators who are inclusion list committee members.
func (api *EngineAPI) GetInclusionListV1() (*GetInclusionListResponseV1, error) {
	ilBackend, ok := api.backend.(InclusionListBackend)
	if !ok {
		return &GetInclusionListResponseV1{Transactions: []string{}}, nil
	}

	il := ilBackend.GetInclusionList()
	// Convert transactions to hex-encoded strings with 0x prefix
	txs := make([]string, len(il.Transactions))
	for i, tx := range il.Transactions {
		txs[i] = fmt.Sprintf("0x%x", tx)
	}
	return &GetInclusionListResponseV1{
		Transactions: txs,
	}, nil
}

// InclusionListBackend is a type alias — canonical definition in engine/backendapi.
type InclusionListBackend = backendapi.InclusionListBackend

// handleNewInclusionListV1 processes an engine_newInclusionListV1 request.
func (api *EngineAPI) handleNewInclusionListV1(params []json.RawMessage) (any, *jsonrpcError) {
	if len(params) != 1 {
		return nil, &jsonrpcError{
			Code:    InvalidParamsCode,
			Message: fmt.Sprintf("expected 1 param, got %d", len(params)),
		}
	}

	var il InclusionListV1
	if err := json.Unmarshal(params[0], &il); err != nil {
		return nil, &jsonrpcError{
			Code:    InvalidParamsCode,
			Message: fmt.Sprintf("invalid inclusion list: %v", err),
		}
	}

	result, err := api.NewInclusionListV1(il)
	if err != nil {
		return nil, engineErrorToRPC(err)
	}
	return result, nil
}

// handleGetInclusionListV1 processes an engine_getInclusionListV1 request.
func (api *EngineAPI) handleGetInclusionListV1(params []json.RawMessage) (any, *jsonrpcError) {
	result, err := api.GetInclusionListV1()
	if err != nil {
		return nil, engineErrorToRPC(err)
	}
	return result, nil
}

func strPtr(s string) *string { return &s }
