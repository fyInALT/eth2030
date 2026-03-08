package rpc

// eth_api.go re-exports EthDirectAPI from rpc/ethapi and provides
// package-level helpers for test files in this package.

import (
	"math/big"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/rpc/ethapi"
)

// EthDirectAPI is re-exported from rpc/ethapi.
type EthDirectAPI = ethapi.EthDirectAPI

// EthDirectAPI errors re-exported from ethapi.
var (
	ErrNoCurrentBlock   = ethapi.ErrNoCurrentBlock
	ErrBlockNotFound    = ethapi.ErrBlockNotFound
	ErrStateUnavailable = ethapi.ErrStateUnavailable
	ErrTxNotFound       = ethapi.ErrTxNotFound
	ErrReceiptNotFound  = ethapi.ErrReceiptNotFound
	ErrEmptyTxData      = ethapi.ErrEmptyTxData
	ErrExecutionFailed  = ethapi.ErrExecutionFailed
)

// NewEthDirectAPI is re-exported from rpc/ethapi.
var NewEthDirectAPI = ethapi.NewEthDirectAPI

// parseBlockNumber is a package-level helper (delegating to ethapi) so that
// test files in package rpc can call it by the original unexported name.
func parseBlockNumber(s string) BlockNumber {
	return ethapi.ParseBlockNumber(s)
}

// apiSubs returns the *SubscriptionManager underlying an *EthAPI.
// The top-level NewEthAPI always passes a *SubscriptionManager, so this
// type assertion is safe. Used only by tests in package rpc.
func apiSubs(api *EthAPI) *SubscriptionManager {
	sm, _ := api.Subs().(*SubscriptionManager)
	return sm
}

// extractCallArgs is a package-level wrapper delegating to rpc/ethapi.
func extractCallArgs(args map[string]interface{}) (from types.Address, to *types.Address, gas uint64, value *big.Int, data []byte) {
	return ethapi.ExtractCallArgs(args)
}

// formatHeaderMap is a package-level wrapper delegating to rpc/ethapi.
func formatHeaderMap(h *types.Header) map[string]interface{} {
	return ethapi.FormatHeaderMap(h)
}

// formatTxAsMap is a package-level wrapper delegating to rpc/ethapi.
func formatTxAsMap(tx *types.Transaction) map[string]interface{} {
	return ethapi.FormatTxAsMap(tx)
}
