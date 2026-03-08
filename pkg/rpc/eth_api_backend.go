package rpc

// eth_api_backend.go re-exports EthAPIBackend from rpc/ethapi.

import "github.com/eth2030/eth2030/rpc/ethapi"

// EthAPIBackend is re-exported from rpc/ethapi.
type EthAPIBackend = ethapi.EthAPIBackend

// EthAPIBackend errors re-exported from ethapi.
var (
	ErrAPIBackendNoBlock    = ethapi.ErrAPIBackendNoBlock
	ErrAPIBackendNoState    = ethapi.ErrAPIBackendNoState
	ErrAPIBackendNoTx       = ethapi.ErrAPIBackendNoTx
	ErrAPIBackendNoReceipt  = ethapi.ErrAPIBackendNoReceipt
	ErrAPIBackendNoLogs     = ethapi.ErrAPIBackendNoLogs
	ErrAPIBackendEstimate   = ethapi.ErrAPIBackendEstimate
	ErrAPIBackendNoPending  = ethapi.ErrAPIBackendNoPending
	ErrAPIBackendInvalidArg = ethapi.ErrAPIBackendInvalidArg
)

// TxWithReceipt is re-exported from rpc/ethapi.
type TxWithReceipt = ethapi.TxWithReceipt

// LogFilterParams is re-exported from rpc/ethapi.
type LogFilterParams = ethapi.LogFilterParams

// GasEstimateArgs is re-exported from rpc/ethapi.
type GasEstimateArgs = ethapi.GasEstimateArgs

// PendingTxInfo is re-exported from rpc/ethapi.
type PendingTxInfo = ethapi.PendingTxInfo

// NewEthAPIBackend is re-exported from rpc/ethapi.
var NewEthAPIBackend = ethapi.NewEthAPIBackend
