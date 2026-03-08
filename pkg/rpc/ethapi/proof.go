package ethapi

import (
	"encoding/hex"
	"encoding/json"
	"math/big"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/rlp"
	"github.com/eth2030/eth2030/rpc/debugapi"
)

// Re-export trace types from rpc/debugapi.
type (
	// StructLog is re-exported from rpc/debugapi.
	StructLog = debugapi.StructLog
	// TraceResult is re-exported from rpc/debugapi.
	TraceResult = debugapi.TraceResult
	// BlockTraceResult is re-exported from rpc/debugapi.
	BlockTraceResult = debugapi.BlockTraceResult
)

// getProof implements eth_getProof (EIP-1186).
// Returns the account and storage values along with Merkle proofs.
func (api *EthAPI) getProof(req *Request) *Response {
	if len(req.Params) < 3 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing address, storageKeys, or block number")
	}

	var addrHex string
	if err := json.Unmarshal(req.Params[0], &addrHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid address: "+err.Error())
	}

	var storageKeysHex []string
	if err := json.Unmarshal(req.Params[1], &storageKeysHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid storageKeys: "+err.Error())
	}

	var bn BlockNumber
	if err := json.Unmarshal(req.Params[2], &bn); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid block number: "+err.Error())
	}

	addr := types.HexToAddress(addrHex)

	// Convert storage key hex strings to types.Hash.
	storageKeys := make([]types.Hash, len(storageKeysHex))
	for i, keyHex := range storageKeysHex {
		storageKeys[i] = types.HexToHash(keyHex)
	}

	// Generate real MPT proofs via the backend.
	proof, err := api.backend.GetProof(addr, storageKeys, bn)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}

	// Convert trie.StorageProof to rpc.StorageProof with hex encoding.
	rpcStorageProofs := make([]StorageProof, len(proof.StorageProof))
	for i, sp := range proof.StorageProof {
		rpcStorageProofs[i] = StorageProof{
			Key:   storageKeysHex[i],
			Value: encodeBigInt(sp.Value),
			Proof: encodeProofNodes(sp.Proof),
		}
	}

	result := &AccountProof{
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

// rlpAccountForProof is the RLP-serializable account struct matching the Yellow Paper
// definition: [nonce, balance, storageRoot, codeHash].
type rlpAccountForProof struct {
	Nonce    uint64
	Balance  *big.Int
	Root     []byte
	CodeHash []byte
}

// encodeAccountRLP encodes an account as RLP per the Yellow Paper:
// RLP([nonce, balance, storageRoot, codeHash]).
func encodeAccountRLP(nonce uint64, balance *big.Int, storageRoot, codeHash types.Hash) []byte {
	acc := rlpAccountForProof{
		Nonce:    nonce,
		Balance:  balance,
		Root:     storageRoot[:],
		CodeHash: codeHash[:],
	}
	encoded, err := rlp.EncodeToBytes(acc)
	if err != nil {
		return nil
	}
	return encoded
}

// encodeProofNodes converts raw proof node bytes to 0x-prefixed hex strings.
func encodeProofNodes(nodes [][]byte) []string {
	result := make([]string, len(nodes))
	for i, node := range nodes {
		result[i] = "0x" + hex.EncodeToString(node)
	}
	return result
}
