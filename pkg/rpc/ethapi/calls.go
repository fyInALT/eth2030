package ethapi

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/eth2030/eth2030/core/types"
)

// CallRequest wraps the full parameter set for eth_call and eth_estimateGas,
// including the optional state overrides and block number resolution.
type CallRequest struct {
	Args      CallArgs
	BlockNum  BlockNumber
	Overrides StateOverride
}

// MaxGasEstimateIterations caps the binary search loop in estimateGas
// to prevent unbounded compute. log2(30_000_000) ~ 25, so 64 is generous.
const MaxGasEstimateIterations = 64

// maxGasEstimateIterations is the unexported alias used within the package.
const maxGasEstimateIterations = MaxGasEstimateIterations

// IntrinsicGasFloor is the minimum gas for a simple transfer (21000).
const IntrinsicGasFloor uint64 = 21000

// intrinsicGasFloor is the unexported alias used within the package.
const intrinsicGasFloor = IntrinsicGasFloor

// ErrCodeExecution is the standard error code for EVM execution failures.
const ErrCodeExecution = -32015

// RevertError wraps an EVM revert with the optional decoded reason.
type RevertError struct {
	Reason string
	Data   []byte
}

// Error returns a human-readable message including the revert reason.
func (e *RevertError) Error() string {
	if e.Reason != "" {
		return "execution reverted: " + e.Reason
	}
	return "execution reverted"
}

// ParseCallArgs is the exported version of parseCallArgs.
func ParseCallArgs(args *CallArgs) (from types.Address, to *types.Address, gas uint64, value *big.Int, data []byte) {
	return parseCallArgs(args)
}

// parseCallArgs extracts EVM call parameters from CallArgs.
func parseCallArgs(args *CallArgs) (from types.Address, to *types.Address, gas uint64, value *big.Int, data []byte) {
	if args.From != nil {
		from = types.HexToAddress(*args.From)
	}
	if args.To != nil {
		addr := types.HexToAddress(*args.To)
		to = &addr
	}
	gas = 50_000_000 // default gas limit
	if args.Gas != nil {
		gas = parseHexUint64(*args.Gas)
	}
	value = new(big.Int)
	if args.Value != nil {
		value = parseHexBigInt(*args.Value)
	}
	data = args.GetData()
	return
}

// parseCallRequest extracts a full CallRequest (args, block number,
// optional overrides) from JSON-RPC request params.
func parseCallRequest(params []json.RawMessage) (*CallRequest, *RPCError) {
	if len(params) < 1 {
		return nil, &RPCError{Code: ErrCodeInvalidParams, Message: "missing call arguments"}
	}

	var args CallArgs
	if err := json.Unmarshal(params[0], &args); err != nil {
		return nil, &RPCError{Code: ErrCodeInvalidParams, Message: err.Error()}
	}

	bn := LatestBlockNumber
	if len(params) > 1 {
		if err := json.Unmarshal(params[1], &bn); err != nil {
			return nil, &RPCError{Code: ErrCodeInvalidParams, Message: "invalid block number: " + err.Error()}
		}
	}

	var overrides StateOverride
	if len(params) > 2 {
		if err := json.Unmarshal(params[2], &overrides); err != nil {
			return nil, &RPCError{Code: ErrCodeInvalidParams, Message: "invalid state overrides: " + err.Error()}
		}
	}

	return &CallRequest{Args: args, BlockNum: bn, Overrides: overrides}, nil
}

// resolveBlockNumber translates a symbolic block number (latest, pending,
// earliest, safe, finalized) into a concrete *types.Header, returning nil
// if the requested block is not available.
func resolveBlockNumber(backend Backend, bn BlockNumber) *types.Header {
	return backend.HeaderByNumber(bn)
}

// overrideStateDB is the minimal interface needed to apply state overrides.
type overrideStateDB interface {
	GetBalance(types.Address) *big.Int
	SubBalance(types.Address, *big.Int)
	AddBalance(types.Address, *big.Int)
	SetNonce(types.Address, uint64)
	SetCode(types.Address, []byte)
	SetState(types.Address, types.Hash, types.Hash)
}

// applyOverrides modifies the state database according to the provided
// state overrides.
func applyOverrides(statedb overrideStateDB, overrides StateOverride) {
	for addrHex, ov := range overrides {
		addr := types.HexToAddress(addrHex)

		if ov.Balance != nil {
			bal := parseHexBigInt(*ov.Balance)
			current := statedb.GetBalance(addr)
			diff := new(big.Int).Sub(bal, current)
			if diff.Sign() > 0 {
				statedb.AddBalance(addr, diff)
			} else if diff.Sign() < 0 {
				statedb.SubBalance(addr, new(big.Int).Neg(diff))
			}
		}
		if ov.Nonce != nil {
			statedb.SetNonce(addr, parseHexUint64(*ov.Nonce))
		}
		if ov.Code != nil {
			statedb.SetCode(addr, fromHexBytes(*ov.Code))
		}
		// "state" replaces all storage keys with the given map.
		if ov.State != nil {
			for keyHex, valHex := range ov.State {
				statedb.SetState(addr, types.HexToHash(keyHex), types.HexToHash(valHex))
			}
		}
		// "stateDiff" merges provided keys, leaving others intact.
		if ov.StateDiff != nil {
			for keyHex, valHex := range ov.StateDiff {
				statedb.SetState(addr, types.HexToHash(keyHex), types.HexToHash(valHex))
			}
		}
	}
}

// decodeRevertReason attempts to extract a human-readable revert reason
// from EVM return data.
func decodeRevertReason(data []byte) string {
	// Minimum length: 4 (selector) + 32 (offset) + 32 (length) + 0 (data)
	if len(data) < 68 {
		return ""
	}
	// Check selector: 0x08c379a2 = keccak("Error(string)")[:4]
	selector := [4]byte{0x08, 0xc3, 0x79, 0xa2}
	if data[0] != selector[0] || data[1] != selector[1] ||
		data[2] != selector[2] || data[3] != selector[3] {
		return ""
	}

	// ABI decode the string: offset at [4:36], length at [36:68], data at [68:]
	offset := new(big.Int).SetBytes(data[4:36]).Uint64()
	if offset != 32 {
		return "" // non-standard encoding
	}
	length := new(big.Int).SetBytes(data[36:68]).Uint64()
	if uint64(len(data)) < 68+length {
		return ""
	}
	return string(data[68 : 68+length])
}

// ethCallWithOverrides executes an EVM call with state overrides applied.
func (api *EthAPI) ethCallWithOverrides(
	from types.Address,
	to *types.Address,
	data []byte,
	gas uint64,
	value *big.Int,
	bn BlockNumber,
	overrides StateOverride,
) ([]byte, uint64, error) {
	_ = overrides // backend EVMCall does not yet accept overrides
	return api.backend.EVMCall(from, to, data, gas, value, bn)
}

// ethCallFull handles eth_call with full parsing, overrides, and error decoding.
func (api *EthAPI) ethCallFull(req *Request) *Response {
	cr, rpcErr := parseCallRequest(req.Params)
	if rpcErr != nil {
		return &Response{JSONRPC: "2.0", Error: rpcErr, ID: req.ID}
	}

	from, to, gas, value, data := parseCallArgs(&cr.Args)
	result, _, err := api.ethCallWithOverrides(from, to, data, gas, value, cr.BlockNum, cr.Overrides)
	if err != nil {
		reason := decodeRevertReason(result)
		msg := "execution error: " + err.Error()
		if reason != "" {
			msg = fmt.Sprintf("execution reverted: %s", reason)
		}
		return errorResponse(req.ID, ErrCodeExecution, msg)
	}

	return successResponse(req.ID, encodeBytes(result))
}

// estimateGasBinarySearch performs a bounded binary search to find the
// minimum gas needed to execute a call.
func (api *EthAPI) estimateGasBinarySearch(
	from types.Address,
	to *types.Address,
	data []byte,
	value *big.Int,
	lo, hi uint64,
	bn BlockNumber,
) (uint64, error) {
	// Verify the upper bound works.
	_, _, err := api.backend.EVMCall(from, to, data, hi, value, bn)
	if err != nil {
		return 0, err
	}

	// Quick check: if lo works, return immediately.
	_, _, errLo := api.backend.EVMCall(from, to, data, lo, value, bn)
	if errLo == nil {
		return lo, nil
	}

	// Binary search between lo and hi with bounded iterations.
	iterations := 0
	for lo+1 < hi && iterations < maxGasEstimateIterations {
		mid := lo + (hi-lo)/2
		_, _, err := api.backend.EVMCall(from, to, data, mid, value, bn)
		if err != nil {
			lo = mid
		} else {
			hi = mid
		}
		iterations++
	}

	return hi, nil
}

// estimateGasFull handles eth_estimateGas with full parsing and binary search.
func (api *EthAPI) estimateGasFull(req *Request) *Response {
	cr, rpcErr := parseCallRequest(req.Params)
	if rpcErr != nil {
		return &Response{JSONRPC: "2.0", Error: rpcErr, ID: req.ID}
	}

	from, to, _, value, data := parseCallArgs(&cr.Args)

	header := api.backend.HeaderByNumber(cr.BlockNum)
	if header == nil {
		return errorResponse(req.ID, ErrCodeInternal, "block not found")
	}

	// Upper bound: block gas limit (or user-specified gas if lower).
	hi := header.GasLimit
	if cr.Args.Gas != nil {
		userGas := parseHexUint64(*cr.Args.Gas)
		if userGas > 0 && userGas < hi {
			hi = userGas
		}
	}

	// Lower bound: intrinsic gas floor.
	lo := intrinsicGasFloor

	// If data is provided, add calldata cost to the floor estimate.
	if len(data) > 0 {
		calldataGas := uint64(0)
		for _, b := range data {
			if b == 0 {
				calldataGas += 4
			} else {
				calldataGas += 16
			}
		}
		lo += calldataGas
	}

	// Ensure lo does not exceed hi.
	if lo > hi {
		lo = hi
	}

	estimated, err := api.estimateGasBinarySearch(from, to, data, value, lo, hi, cr.BlockNum)
	if err != nil {
		return errorResponse(req.ID, ErrCodeExecution, "gas estimation failed: "+err.Error())
	}

	return successResponse(req.ID, encodeUint64(estimated))
}

// BlockNumberOrHashParam can hold either a block number or a block hash,
// following the BlockNumberOrHash pattern from the Ethereum JSON-RPC spec.
type BlockNumberOrHashParam struct {
	BlockNumber *BlockNumber `json:"blockNumber,omitempty"`
	BlockHash   *string      `json:"blockHash,omitempty"`
}

// ResolveHeader resolves a BlockNumberOrHashParam to a header using the backend.
func (p *BlockNumberOrHashParam) ResolveHeader(backend Backend) *types.Header {
	if p.BlockHash != nil {
		hash := types.HexToHash(*p.BlockHash)
		return backend.HeaderByHash(hash)
	}
	if p.BlockNumber != nil {
		return backend.HeaderByNumber(*p.BlockNumber)
	}
	return backend.CurrentHeader()
}

// StateOverrideApplier applies state overrides to a state database
// for simulated calls.
type StateOverrideApplier struct {
	Overrides StateOverride
}

// NewStateOverrideApplier creates a new applier from the given overrides.
func NewStateOverrideApplier(overrides StateOverride) *StateOverrideApplier {
	if overrides == nil {
		return &StateOverrideApplier{Overrides: make(StateOverride)}
	}
	return &StateOverrideApplier{Overrides: overrides}
}

// HasOverrides returns true if any overrides are configured.
func (a *StateOverrideApplier) HasOverrides() bool {
	return len(a.Overrides) > 0
}

// Apply applies the state overrides to the given state database.
func (a *StateOverrideApplier) Apply(statedb overrideStateDB) {
	applyOverrides(statedb, a.Overrides)
}

// AccountCount returns the number of accounts with overrides.
func (a *StateOverrideApplier) AccountCount() int {
	return len(a.Overrides)
}

// ParseCallRequest is the exported version of parseCallRequest,
// used by tests and dependent packages.
func ParseCallRequest(params []json.RawMessage) (*CallRequest, *RPCError) {
	return parseCallRequest(params)
}

// DecodeRevertReason is the exported version of decodeRevertReason.
func DecodeRevertReason(data []byte) string {
	return decodeRevertReason(data)
}

// ResolveBlockNumber is the exported version of resolveBlockNumber.
func ResolveBlockNumber(backend Backend, bn BlockNumber) *types.Header {
	return resolveBlockNumber(backend, bn)
}

// EstimateGasBinarySearch is the exported version of estimateGasBinarySearch,
// allowing dependent packages to call it without access to unexported methods.
func EstimateGasBinarySearch(
	api *EthAPI,
	from types.Address,
	to *types.Address,
	data []byte,
	value *big.Int,
	lo, hi uint64,
	bn BlockNumber,
) (uint64, error) {
	return api.estimateGasBinarySearch(from, to, data, value, lo, hi, bn)
}
