package ethapi

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/eth2030/eth2030/core/eips"
	"github.com/eth2030/eth2030/core/types"
)

type rpcUserOperation struct {
	Sender                        string  `json:"sender"`
	Nonce                         string  `json:"nonce"`
	Factory                       *string `json:"factory,omitempty"`
	FactoryData                   string  `json:"factoryData,omitempty"`
	CallData                      string  `json:"callData"`
	CallGasLimit                  string  `json:"callGasLimit"`
	VerificationGasLimit          string  `json:"verificationGasLimit"`
	PreVerificationGas            string  `json:"preVerificationGas"`
	MaxFeePerGas                  string  `json:"maxFeePerGas"`
	MaxPriorityFeePerGas          string  `json:"maxPriorityFeePerGas"`
	Paymaster                     *string `json:"paymaster,omitempty"`
	PaymasterVerificationGasLimit string  `json:"paymasterVerificationGasLimit,omitempty"`
	PaymasterPostOpGasLimit       string  `json:"paymasterPostOpGasLimit,omitempty"`
	PaymasterData                 string  `json:"paymasterData,omitempty"`
	Signature                     string  `json:"signature,omitempty"`
}

type rpcUserOperationReceipt struct {
	UserOpHash      string `json:"userOpHash"`
	TransactionHash string `json:"transactionHash"`
}

func (api *EthAPI) sendUserOperation(req *Request) *Response {
	if len(req.Params) < 2 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing user operation or entry point")
	}

	var opReq rpcUserOperation
	if err := json.Unmarshal(req.Params[0], &opReq); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	var entryPointHex string
	if err := json.Unmarshal(req.Params[1], &entryPointHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}
	if types.HexToAddress(entryPointHex) != types.AAEntryPoint {
		return errorResponse(req.ID, ErrCodeInvalidParams, fmt.Sprintf("unsupported entry point %q", entryPointHex))
	}

	userOp, err := parseRPCUserOperation(&opReq)
	if err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}
	if err := eips.ValidateUserOp(userOp); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	nonce, ok := userOp.Nonce.Uint64(), userOp.Nonce.BitLen() <= 64
	if !ok {
		return errorResponse(req.ID, ErrCodeInvalidParams, "user operation nonce exceeds 64 bits")
	}

	aatx := &types.AATx{
		ChainID:                api.backend.ChainID(),
		Nonce:                  nonce,
		Sender:                 userOp.Sender,
		SenderValidationData:   userOp.Signature,
		Deployer:               userOp.Factory,
		DeployerData:           userOp.FactoryData,
		Paymaster:              userOp.Paymaster,
		PaymasterData:          userOp.PaymasterData,
		SenderExecutionData:    userOp.CallData,
		MaxPriorityFeePerGas:   userOp.MaxPriorityFeePerGas,
		MaxFeePerGas:           userOp.MaxFeePerGas,
		SenderValidationGas:    userOp.VerificationGasLimit,
		PaymasterValidationGas: userOp.PaymasterVerificationGasLimit,
		SenderExecutionGas:     userOp.CallGasLimit,
		PaymasterPostOpGas:     userOp.PaymasterPostOpGasLimit,
	}
	if err := types.ValidateAATx(aatx); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}

	tx := types.NewTransaction(aatx)
	tx.SetSender(aatx.Sender)
	if err := api.backend.SendTransaction(tx); err != nil {
		return errorResponse(req.ID, ErrCodeInternal, err.Error())
	}

	opHash := eips.UserOpHash(userOp, api.backend.ChainID())
	api.userOpsMu.Lock()
	api.userOps[opHash] = tx.Hash()
	api.userOpsMu.Unlock()
	return successResponse(req.ID, encodeHash(opHash))
}

func (api *EthAPI) getUserOperationReceipt(req *Request) *Response {
	if len(req.Params) < 1 {
		return errorResponse(req.ID, ErrCodeInvalidParams, "missing user operation hash")
	}

	var opHashHex string
	if err := json.Unmarshal(req.Params[0], &opHashHex); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, err.Error())
	}
	opHash := types.HexToHash(opHashHex)

	api.userOpsMu.RLock()
	txHash, ok := api.userOps[opHash]
	api.userOpsMu.RUnlock()
	if !ok {
		return successResponse(req.ID, nil)
	}

	tx, blockNum, _ := api.backend.GetTransaction(txHash)
	if tx == nil {
		return successResponse(req.ID, nil)
	}
	if api.historyPruned(blockNum) {
		return errorResponse(req.ID, ErrCodeHistoryPruned, "historical receipt pruned (EIP-4444)")
	}

	header := api.backend.HeaderByNumber(BlockNumber(blockNum))
	if header == nil {
		return successResponse(req.ID, nil)
	}
	for _, receipt := range api.backend.GetReceipts(header.Hash()) {
		if receipt.TxHash == txHash {
			return successResponse(req.ID, &rpcUserOperationReceipt{
				UserOpHash:      encodeHash(opHash),
				TransactionHash: encodeHash(txHash),
			})
		}
	}
	return successResponse(req.ID, nil)
}

func parseRPCUserOperation(req *rpcUserOperation) (*eips.UserOperation, error) {
	op := &eips.UserOperation{
		Sender:               types.HexToAddress(req.Sender),
		Nonce:                parseHexBigInt(req.Nonce),
		FactoryData:          fromHexBytes(req.FactoryData),
		CallData:             fromHexBytes(req.CallData),
		CallGasLimit:         parseHexUint64(req.CallGasLimit),
		VerificationGasLimit: parseHexUint64(req.VerificationGasLimit),
		PreVerificationGas:   parseHexUint64(req.PreVerificationGas),
		MaxFeePerGas:         parseHexBigInt(req.MaxFeePerGas),
		MaxPriorityFeePerGas: parseHexBigInt(req.MaxPriorityFeePerGas),
		PaymasterData:        fromHexBytes(req.PaymasterData),
		Signature:            fromHexBytes(req.Signature),
	}
	if op.Nonce == nil {
		op.Nonce = new(big.Int)
	}
	if req.Factory != nil {
		factory := types.HexToAddress(*req.Factory)
		op.Factory = &factory
	}
	if req.Paymaster != nil {
		paymaster := types.HexToAddress(*req.Paymaster)
		op.Paymaster = &paymaster
		op.PaymasterVerificationGasLimit = parseHexUint64(req.PaymasterVerificationGasLimit)
		op.PaymasterPostOpGasLimit = parseHexUint64(req.PaymasterPostOpGasLimit)
	}
	return op, nil
}
