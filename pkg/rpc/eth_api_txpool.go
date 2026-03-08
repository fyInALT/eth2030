package rpc

// eth_api_txpool.go re-exports TxPoolAPI from rpc/ethapi.

import (
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/rpc/ethapi"
)

// TxPoolBackend is re-exported from rpc/ethapi.
type TxPoolBackend = ethapi.TxPoolBackend

// TxPoolAPI is re-exported from rpc/ethapi.
type TxPoolAPI = ethapi.TxPoolAPI

// TxPoolStatusResult is re-exported from rpc/ethapi.
type TxPoolStatusResult = ethapi.TxPoolStatusResult

// TxPoolContentResult is re-exported from rpc/ethapi.
type TxPoolContentResult = ethapi.TxPoolContentResult

// TxPoolInspectResult is re-exported from rpc/ethapi.
type TxPoolInspectResult = ethapi.TxPoolInspectResult

// NewTxPoolAPI is re-exported from rpc/ethapi.
var NewTxPoolAPI = ethapi.NewTxPoolAPI

// EffectiveGasPrice is re-exported from rpc/ethapi.
var EffectiveGasPrice = ethapi.EffectiveGasPrice

// decodeRawTransaction is a package-level wrapper delegating to rpc/ethapi.
func decodeRawTransaction(rawHex string) (*types.Transaction, []byte, error) {
	return ethapi.DecodeRawTransaction(rawHex)
}

// formatTxPoolMap is a package-level wrapper delegating to rpc/ethapi.
func formatTxPoolMap(txMap map[types.Address][]*types.Transaction) map[string]map[string]*RPCTransaction {
	return ethapi.FormatTxPoolMap(txMap)
}
