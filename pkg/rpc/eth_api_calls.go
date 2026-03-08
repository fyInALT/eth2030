package rpc

// eth_api_calls.go exposes package-level helpers for tests in this package
// that need access to call-handling internals now living in rpc/ethapi.

import (
	"encoding/json"
	"math/big"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/rpc/ethapi"
)

// intrinsicGasFloor and maxGasEstimateIterations mirror the ethapi constants
// so that test files in this package can access them by the original name.
const intrinsicGasFloor uint64 = ethapi.IntrinsicGasFloor
const maxGasEstimateIterations = ethapi.MaxGasEstimateIterations

// parseCallRequest delegates to the exported ethapi function.
func parseCallRequest(params []json.RawMessage) (*CallRequest, *RPCError) {
	return ethapi.ParseCallRequest(params)
}

// parseCallArgs delegates to the exported ethapi function.
func parseCallArgs(args *CallArgs) (from types.Address, to *types.Address, gas uint64, value *big.Int, data []byte) {
	return ethapi.ParseCallArgs(args)
}

// decodeRevertReason delegates to the exported ethapi function.
func decodeRevertReason(data []byte) string {
	return ethapi.DecodeRevertReason(data)
}

// resolveBlockNumber delegates to the exported ethapi function.
func resolveBlockNumber(b Backend, num BlockNumber) *types.Header {
	return ethapi.ResolveBlockNumber(b, num)
}

// estimateGasBinarySearch is a package-level helper that wraps the ethapi
// method so tests in package rpc can call it without defining a new method on
// the type alias (which is not permitted in Go).
func estimateGasBinarySearch(
	api *EthAPI,
	from types.Address,
	to *types.Address,
	data []byte,
	value *big.Int,
	lo, hi uint64,
	blockNum BlockNumber,
) (uint64, error) {
	return ethapi.EstimateGasBinarySearch(api, from, to, data, value, lo, hi, blockNum)
}
