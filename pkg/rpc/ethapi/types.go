// Package ethapi implements the eth_ namespace JSON-RPC API.
package ethapi

import (
	"encoding/json"
	"math/big"

	"github.com/eth2030/eth2030/core/types"
	rpcbackend "github.com/eth2030/eth2030/rpc/backend"
	rpctypes "github.com/eth2030/eth2030/rpc/types"
)

// Re-export backend and RPC types used throughout ethapi.
type (
	Backend          = rpcbackend.Backend
	BlockNumber      = rpctypes.BlockNumber
	Request          = rpctypes.Request
	Response         = rpctypes.Response
	RPCError         = rpctypes.RPCError
	RPCBlock         = rpctypes.RPCBlock
	RPCBlockWithTxs  = rpctypes.RPCBlockWithTxs
	RPCTransaction   = rpctypes.RPCTransaction
	RPCReceipt       = rpctypes.RPCReceipt
	RPCLog           = rpctypes.RPCLog
	RPCWithdrawal    = rpctypes.RPCWithdrawal
	RPCAccessTuple   = rpctypes.RPCAccessTuple
	RPCAuthorization = rpctypes.RPCAuthorization
	CallArgs         = rpctypes.CallArgs
	FilterCriteria   = rpctypes.FilterCriteria
)

// Re-export block number constants.
const (
	LatestBlockNumber    = rpctypes.LatestBlockNumber
	PendingBlockNumber   = rpctypes.PendingBlockNumber
	EarliestBlockNumber  = rpctypes.EarliestBlockNumber
	SafeBlockNumber      = rpctypes.SafeBlockNumber
	FinalizedBlockNumber = rpctypes.FinalizedBlockNumber

	ErrCodeParse          = rpctypes.ErrCodeParse
	ErrCodeInvalidRequest = rpctypes.ErrCodeInvalidRequest
	ErrCodeMethodNotFound = rpctypes.ErrCodeMethodNotFound
	ErrCodeInvalidParams  = rpctypes.ErrCodeInvalidParams
	ErrCodeInternal       = rpctypes.ErrCodeInternal
	ErrCodeHistoryPruned  = rpctypes.ErrCodeHistoryPruned
)

// Re-export formatting functions.
var (
	FormatBlock       = rpctypes.FormatBlock
	FormatHeader      = rpctypes.FormatHeader
	FormatTransaction = rpctypes.FormatTransaction
	FormatReceipt     = rpctypes.FormatReceipt
	FormatLog         = rpctypes.FormatLog
)

// Unexported helpers delegating to rpctypes.
func encodeHash(h types.Hash) string       { return rpctypes.EncodeHash(h) }
func encodeAddress(a types.Address) string { return rpctypes.EncodeAddress(a) }
func encodeBytes(b []byte) string          { return rpctypes.EncodeBytes(b) }
func encodeBloom(b types.Bloom) string     { return rpctypes.EncodeBloom(b) }
func encodeUint64(n uint64) string         { return rpctypes.EncodeUint64(n) }
func encodeBigInt(n *big.Int) string       { return rpctypes.EncodeBigInt(n) }
func fromHexBytes(s string) []byte         { return rpctypes.FromHexBytes(s) }
func parseHexUint64(s string) uint64       { return rpctypes.ParseHexUint64(s) }
func parseHexBigInt(s string) *big.Int     { return rpctypes.ParseHexBigInt(s) }

// successResponse is a helper that builds a JSON-RPC success response.
func successResponse(id json.RawMessage, result interface{}) *Response {
	return rpctypes.NewSuccessResponse(id, result)
}

// errorResponse is a helper that builds a JSON-RPC error response.
func errorResponse(id json.RawMessage, code int, msg string) *Response {
	return rpctypes.NewErrorResponse(id, code, msg)
}

// AccountProof is the EIP-1186 response for eth_getProof.
type AccountProof struct {
	Address      string         `json:"address"`
	AccountProof []string       `json:"accountProof"`
	Balance      string         `json:"balance"`
	CodeHash     string         `json:"codeHash"`
	Nonce        string         `json:"nonce"`
	StorageHash  string         `json:"storageHash"`
	StorageProof []StorageProof `json:"storageProof"`
}

// StorageProof is a single storage slot proof within eth_getProof.
type StorageProof struct {
	Key   string   `json:"key"`
	Value string   `json:"value"`
	Proof []string `json:"proof"`
}

// TxWithReceipt pairs a transaction with its receipt for lookup results.
type TxWithReceipt struct {
	Tx          *types.Transaction
	Receipt     *types.Receipt
	BlockNumber uint64
	BlockHash   types.Hash
	TxIndex     uint64
}

// LogFilterParams holds parameters for GetLogs queries.
type LogFilterParams struct {
	FromBlock uint64
	ToBlock   uint64
	Addresses []types.Address
	Topics    [][]types.Hash
}

// GasEstimateArgs holds arguments for gas estimation.
type GasEstimateArgs struct {
	From   types.Address
	To     *types.Address
	Gas    uint64
	Value  *big.Int
	Data   []byte
	GasTip *big.Int // EIP-1559 max priority fee
	GasCap *big.Int // EIP-1559 max fee per gas
}

// PendingTxInfo holds pending transaction information.
type PendingTxInfo struct {
	Tx     *types.Transaction
	Sender types.Address
}

// StateOverride maps addresses to account overrides for eth_call.
type StateOverride map[string]AccountOverride

// AccountOverride specifies replacement state for a single account.
type AccountOverride struct {
	Balance   *string           `json:"balance,omitempty"`
	Nonce     *string           `json:"nonce,omitempty"`
	Code      *string           `json:"code,omitempty"`
	State     map[string]string `json:"state,omitempty"`
	StateDiff map[string]string `json:"stateDiff,omitempty"`
}
