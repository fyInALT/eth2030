// engine_api_v4.go re-exports Engine API V4 symbols from engine/api sub-package
// and provides backward-compatible function aliases.
package engine

import (
	"github.com/eth2030/eth2030/core/types"
	engapi "github.com/eth2030/eth2030/engine/api"
)

// Re-export constants from the api sub-package.
const (
	DepositRequestType       = engapi.DepositRequestType
	WithdrawalRequestType    = engapi.WithdrawalRequestType
	ConsolidationRequestType = engapi.ConsolidationRequestType

	MaxDepositRequests       = engapi.MaxDepositRequests
	MaxWithdrawalRequests    = engapi.MaxWithdrawalRequests
	MaxConsolidationRequests = engapi.MaxConsolidationRequests
	DepositRequestSize       = engapi.DepositRequestSize
	WithdrawalRequestSize    = engapi.WithdrawalRequestSize
	ConsolidationRequestSize = engapi.ConsolidationRequestSize
)

// Re-export error variables from the api sub-package.
var (
	ErrV4NilPayload           = engapi.ErrV4NilPayload
	ErrV4RequestTypeMismatch  = engapi.ErrV4RequestTypeMismatch
	ErrV4RequestTooLarge      = engapi.ErrV4RequestTooLarge
	ErrV4MissingRequests      = engapi.ErrV4MissingRequests
	ErrV4InvalidRequestOrder  = engapi.ErrV4InvalidRequestOrder
	ErrV4DuplicateRequestType = engapi.ErrV4DuplicateRequestType
)

// Re-export functions from the api sub-package (backward-compatible aliases).

// NewEngV4 creates a new EngV4 instance.
// The backend must satisfy engapi.V4Backend.
func NewEngV4(backend engapi.V4Backend) *EngV4 {
	return engapi.NewEngV4(backend)
}

// ValidateExecutionRequests checks that the execution requests byte slices are well-formed.
func ValidateExecutionRequests(requests [][]byte) error {
	return engapi.ValidateExecutionRequests(requests)
}

// DecodeDepositRequests parses deposit request objects from the raw bytes.
func DecodeDepositRequests(payload []byte) ([]DepositRequest, error) {
	return engapi.DecodeDepositRequests(payload)
}

// DecodeWithdrawalRequests parses withdrawal request objects from the raw bytes.
func DecodeWithdrawalRequests(payload []byte) ([]WithdrawalRequest, error) {
	return engapi.DecodeWithdrawalRequests(payload)
}

// DecodeConsolidationRequests parses consolidation request objects from raw bytes.
func DecodeConsolidationRequests(payload []byte) ([]ConsolidationRequest, error) {
	return engapi.DecodeConsolidationRequests(payload)
}

// EncodeDepositRequest serializes a deposit request to bytes.
func EncodeDepositRequest(d *DepositRequest) []byte {
	return engapi.EncodeDepositRequest(d)
}

// EncodeWithdrawalRequest serializes a withdrawal request to bytes.
func EncodeWithdrawalRequest(w *WithdrawalRequest) []byte {
	return engapi.EncodeWithdrawalRequest(w)
}

// EncodeConsolidationRequest serializes a consolidation request to bytes.
func EncodeConsolidationRequest(c *ConsolidationRequest) []byte {
	return engapi.EncodeConsolidationRequest(c)
}

// BuildExecutionRequestsList constructs the ordered execution requests byte slice list.
func BuildExecutionRequestsList(
	deposits []DepositRequest,
	withdrawals []WithdrawalRequest,
	consolidations []ConsolidationRequest,
) [][]byte {
	return engapi.BuildExecutionRequestsList(deposits, withdrawals, consolidations)
}

// ExecutionRequestsHash computes the hash of the execution requests list.
func ExecutionRequestsHash(requests [][]byte) types.Hash {
	return engapi.ExecutionRequestsHash(requests)
}

// ClassifyExecutionRequests separates raw request byte slices into typed request structures.
func ClassifyExecutionRequests(requests [][]byte) (*ExecutionRequestsV4, error) {
	return engapi.ClassifyExecutionRequests(requests)
}
