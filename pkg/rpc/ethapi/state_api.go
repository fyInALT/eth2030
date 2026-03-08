// state_api.go implements state query RPC methods for reading contract
// storage, bytecode, Merkle proofs (EIP-1186), balances, and nonces.
package ethapi

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/eth2030/eth2030/core/types"
)

// StateAPI implements state-querying RPC methods.
type StateAPI struct {
	backend Backend
}

// NewStateAPI creates a new StateAPI instance.
func NewStateAPI(backend Backend) *StateAPI {
	return &StateAPI{backend: backend}
}

// HandleRequest dispatches state namespace requests.
func (s *StateAPI) HandleRequest(req *Request) *Response {
	switch req.Method {
	case "eth_getStorageAt":
		return s.getStorageAt(req)
	case "eth_getCode":
		return s.getCode(req)
	case "eth_getProof":
		return s.getProof(req)
	case "eth_getBalance":
		return s.getBalance(req)
	case "eth_getTransactionCount":
		return s.getTransactionCount(req)
	default:
		return errorResponse(req.ID, ErrCodeMethodNotFound,
			fmt.Sprintf("method %q not found", req.Method))
	}
}

// resolveBlockNum translates a BlockNumber tag into a concrete header.
func (s *StateAPI) resolveBlockNum(bn BlockNumber) *types.Header {
	return s.backend.HeaderByNumber(bn)
}

// parseBlockNumberParam extracts and validates a block number parameter
// from the given JSON-RPC params at the specified index.
func parseBlockNumberParam(params []json.RawMessage, idx int) (BlockNumber, *RPCError) {
	if idx >= len(params) {
		return LatestBlockNumber, nil
	}
	var bn BlockNumber
	if err := json.Unmarshal(params[idx], &bn); err != nil {
		return 0, &RPCError{
			Code:    ErrCodeInvalidParams,
			Message: "invalid block number: " + err.Error(),
		}
	}
	return bn, nil
}

// getStorageAt implements eth_getStorageAt.
func (s *StateAPI) getStorageAt(req *Request) *Response {
	if len(req.Params) < 3 {
		return errorResponse(req.ID, ErrCodeInvalidParams,
			"expected 3 params: address, slot, blockNumber")
	}

	var addrHex, slotHex string
	if err := json.Unmarshal(req.Params[0], &addrHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid address: "+err.Error())
	}
	if err := json.Unmarshal(req.Params[1], &slotHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid slot: "+err.Error())
	}

	bn, rpcErr := parseBlockNumberParam(req.Params, 2)
	if rpcErr != nil {
		return &Response{JSONRPC: "2.0", Error: rpcErr, ID: req.ID}
	}

	header := s.resolveBlockNum(bn)
	if header == nil {
		return errorResponse(req.ID, ErrCodeInternal, "block not found")
	}

	statedb, err := s.backend.StateAt(header.Root)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, "state unavailable: "+err.Error())
	}

	addr := types.HexToAddress(addrHex)
	slot := types.HexToHash(slotHex)
	value := statedb.GetState(addr, slot)

	return successResponse(req.ID, encodeHash(value))
}

// getCode implements eth_getCode.
func (s *StateAPI) getCode(req *Request) *Response {
	if len(req.Params) < 2 {
		return errorResponse(req.ID, ErrCodeInvalidParams,
			"expected 2 params: address, blockNumber")
	}

	var addrHex string
	if err := json.Unmarshal(req.Params[0], &addrHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid address: "+err.Error())
	}

	bn, rpcErr := parseBlockNumberParam(req.Params, 1)
	if rpcErr != nil {
		return &Response{JSONRPC: "2.0", Error: rpcErr, ID: req.ID}
	}

	header := s.resolveBlockNum(bn)
	if header == nil {
		return errorResponse(req.ID, ErrCodeInternal, "block not found")
	}

	statedb, err := s.backend.StateAt(header.Root)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, "state unavailable: "+err.Error())
	}

	addr := types.HexToAddress(addrHex)
	code := statedb.GetCode(addr)

	return successResponse(req.ID, encodeBytes(code))
}

// StateAccountProof is the EIP-1186 response for eth_getProof.
type StateAccountProof struct {
	Address      string              `json:"address"`
	AccountProof []string            `json:"accountProof"`
	Balance      string              `json:"balance"`
	CodeHash     string              `json:"codeHash"`
	Nonce        string              `json:"nonce"`
	StorageHash  string              `json:"storageHash"`
	StorageProof []StateStorageProof `json:"storageProof"`
}

// StateStorageProof is a single storage slot proof within the EIP-1186 response.
type StateStorageProof struct {
	Key   string   `json:"key"`
	Value string   `json:"value"`
	Proof []string `json:"proof"`
}

// getProof implements eth_getProof (EIP-1186).
func (s *StateAPI) getProof(req *Request) *Response {
	if len(req.Params) < 3 {
		return errorResponse(req.ID, ErrCodeInvalidParams,
			"expected 3 params: address, storageKeys, blockNumber")
	}

	var addrHex string
	if err := json.Unmarshal(req.Params[0], &addrHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid address: "+err.Error())
	}

	var storageKeysHex []string
	if err := json.Unmarshal(req.Params[1], &storageKeysHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid storageKeys: "+err.Error())
	}

	bn, rpcErr := parseBlockNumberParam(req.Params, 2)
	if rpcErr != nil {
		return &Response{JSONRPC: "2.0", Error: rpcErr, ID: req.ID}
	}

	addr := types.HexToAddress(addrHex)
	storageKeys := make([]types.Hash, len(storageKeysHex))
	for i, keyHex := range storageKeysHex {
		storageKeys[i] = types.HexToHash(keyHex)
	}

	proof, err := s.backend.GetProof(addr, storageKeys, bn)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}

	rpcStorageProofs := make([]StateStorageProof, len(proof.StorageProof))
	for i, sp := range proof.StorageProof {
		rpcStorageProofs[i] = StateStorageProof{
			Key:   storageKeysHex[i],
			Value: encodeBigInt(sp.Value),
			Proof: encodeProofNodes(sp.Proof),
		}
	}

	result := &StateAccountProof{
		Address:      encodeAddress(proof.Address),
		AccountProof: encodeProofNodes(proof.AccountProof),
		Balance:      encodeBigInt(proof.Balance),
		CodeHash:     encodeHash(proof.CodeHash),
		Nonce:        encodeUint64(proof.Nonce),
		StorageHash:  encodeHash(proof.StorageHash),
		StorageProof: rpcStorageProofs,
	}

	return successResponse(req.ID, result)
}

// getBalance implements eth_getBalance with block number resolution.
func (s *StateAPI) getBalance(req *Request) *Response {
	if len(req.Params) < 2 {
		return errorResponse(req.ID, ErrCodeInvalidParams,
			"expected 2 params: address, blockNumber")
	}

	var addrHex string
	if err := json.Unmarshal(req.Params[0], &addrHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid address: "+err.Error())
	}

	bn, rpcErr := parseBlockNumberParam(req.Params, 1)
	if rpcErr != nil {
		return &Response{JSONRPC: "2.0", Error: rpcErr, ID: req.ID}
	}

	header := s.resolveBlockNum(bn)
	if header == nil {
		return errorResponse(req.ID, ErrCodeInternal, "block not found")
	}

	statedb, err := s.backend.StateAt(header.Root)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, "state unavailable: "+err.Error())
	}

	addr := types.HexToAddress(addrHex)
	balance := statedb.GetBalance(addr)
	if balance == nil {
		balance = new(big.Int)
	}

	return successResponse(req.ID, encodeBigInt(balance))
}

// getTransactionCount implements eth_getTransactionCount (nonce).
func (s *StateAPI) getTransactionCount(req *Request) *Response {
	if len(req.Params) < 2 {
		return errorResponse(req.ID, ErrCodeInvalidParams,
			"expected 2 params: address, blockNumber")
	}

	var addrHex string
	if err := json.Unmarshal(req.Params[0], &addrHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid address: "+err.Error())
	}

	bn, rpcErr := parseBlockNumberParam(req.Params, 1)
	if rpcErr != nil {
		return &Response{JSONRPC: "2.0", Error: rpcErr, ID: req.ID}
	}

	header := s.resolveBlockNum(bn)
	if header == nil {
		return errorResponse(req.ID, ErrCodeInternal, "block not found")
	}

	statedb, err := s.backend.StateAt(header.Root)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, "state unavailable: "+err.Error())
	}

	addr := types.HexToAddress(addrHex)
	nonce := statedb.GetNonce(addr)

	return successResponse(req.ID, encodeUint64(nonce))
}
