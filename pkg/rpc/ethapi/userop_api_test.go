package ethapi

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/eips"
	"github.com/eth2030/eth2030/core/types"
	testutil "github.com/eth2030/eth2030/rpc/internal/testutil"
)

func TestEthAPI_SendUserOperation_SubmitsAATx(t *testing.T) {
	mb := testutil.NewMockBackend()
	api := NewEthAPI(mb, nil)

	op := map[string]any{
		"sender":               "0x0000000000000000000000000000000000000011",
		"nonce":                "0x7",
		"callData":             "0x0102",
		"callGasLimit":         "0x5208",
		"verificationGasLimit": "0xc350",
		"preVerificationGas":   "0x7530",
		"maxFeePerGas":         "0x4a817c800",
		"maxPriorityFeePerGas": "0x77359400",
	}

	resp := callRPC(t, api, "eth_sendUserOperation", op, "0x0000000000000000000000000000000000007701")
	if resp.Error != nil {
		t.Fatalf("send user operation error: %v", resp.Error.Message)
	}

	gotHash, ok := resp.Result.(string)
	if !ok {
		t.Fatalf("result type = %T, want string", resp.Result)
	}
	if len(mb.SentTxs) != 1 {
		t.Fatalf("sent tx count = %d, want 1", len(mb.SentTxs))
	}

	tx := mb.SentTxs[0]
	if tx.Type() != types.AATxType {
		t.Fatalf("tx type = %d, want %d", tx.Type(), types.AATxType)
	}
	aatx, ok := tx.Inner().(*types.AATx)
	if !ok {
		t.Fatalf("inner type = %T, want *types.AATx", tx.Inner())
	}
	if aatx.Sender != types.HexToAddress("0x0000000000000000000000000000000000000011") {
		t.Fatalf("sender = %s", aatx.Sender.Hex())
	}
	if aatx.Nonce != 7 {
		t.Fatalf("nonce = %d, want 7", aatx.Nonce)
	}
	if aatx.SenderExecutionGas != 0x5208 {
		t.Fatalf("call gas = %d, want %d", aatx.SenderExecutionGas, uint64(0x5208))
	}
	if tx.Sender() == nil || *tx.Sender() != aatx.Sender {
		t.Fatalf("cached sender = %v, want %s", tx.Sender(), aatx.Sender.Hex())
	}

	wantHash := encodeHash(eips.UserOpHash(&eips.UserOperation{
		Sender:               aatx.Sender,
		Nonce:                new(big.Int).SetUint64(aatx.Nonce),
		CallData:             aatx.SenderExecutionData,
		CallGasLimit:         aatx.SenderExecutionGas,
		VerificationGasLimit: aatx.SenderValidationGas,
		PreVerificationGas:   0x7530,
		MaxFeePerGas:         aatx.MaxFeePerGas,
		MaxPriorityFeePerGas: aatx.MaxPriorityFeePerGas,
	}, mb.ChainID()))
	if gotHash != wantHash {
		t.Fatalf("user op hash = %s, want %s", gotHash, wantHash)
	}
}

func TestEthAPI_SendUserOperation_InvalidEntrypoint(t *testing.T) {
	mb := testutil.NewMockBackend()
	api := NewEthAPI(mb, nil)

	resp := callRPC(t, api, "eth_sendUserOperation", map[string]any{
		"sender":               "0x0000000000000000000000000000000000000011",
		"nonce":                "0x0",
		"callData":             "0x",
		"callGasLimit":         "0x5208",
		"verificationGasLimit": "0xc350",
		"preVerificationGas":   "0x7530",
		"maxFeePerGas":         "0x4a817c800",
		"maxPriorityFeePerGas": "0x77359400",
	}, "0x000000000000000000000000000000000000dead")
	if resp.Error == nil {
		t.Fatal("expected error for invalid entry point")
	}
}

func TestEthAPI_GetUserOperationReceipt(t *testing.T) {
	mb := testutil.NewMockBackend()
	api := NewEthAPI(mb, nil)

	op := map[string]any{
		"sender":               "0x0000000000000000000000000000000000000011",
		"nonce":                "0x7",
		"callData":             "0x0102",
		"callGasLimit":         "0x5208",
		"verificationGasLimit": "0xc350",
		"preVerificationGas":   "0x7530",
		"maxFeePerGas":         "0x4a817c800",
		"maxPriorityFeePerGas": "0x77359400",
	}
	submitResp := callRPC(t, api, "eth_sendUserOperation", op, "0x0000000000000000000000000000000000007701")
	if submitResp.Error != nil {
		t.Fatalf("send user operation error: %v", submitResp.Error.Message)
	}
	opHashHex, ok := submitResp.Result.(string)
	if !ok {
		t.Fatalf("submit result type = %T, want string", submitResp.Result)
	}

	tx := mb.SentTxs[0]
	txHash := tx.Hash()
	mb.Transactions[txHash] = &testutil.MockTxInfo{Tx: tx, BlockNum: 42, Index: 0}
	blockHash := mb.Headers[42].Hash()
	mb.Receipts[blockHash] = []*types.Receipt{{
		Status:            types.ReceiptStatusSuccessful,
		CumulativeGasUsed: tx.Gas(),
		GasUsed:           tx.Gas(),
		TxHash:            txHash,
		BlockHash:         blockHash,
		BlockNumber:       big.NewInt(42),
		TransactionIndex:  0,
		Type:              tx.Type(),
		Logs:              []*types.Log{},
	}}

	resp := callRPC(t, api, "eth_getUserOperationReceipt", opHashHex)
	if resp.Error != nil {
		t.Fatalf("get user operation receipt error: %v", resp.Error.Message)
	}

	raw, err := json.Marshal(resp.Result)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	var receipt struct {
		UserOpHash      string `json:"userOpHash"`
		TransactionHash string `json:"transactionHash"`
	}
	if err := json.Unmarshal(raw, &receipt); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if receipt.UserOpHash != opHashHex {
		t.Fatalf("userOpHash = %s, want %s", receipt.UserOpHash, opHashHex)
	}
	if receipt.TransactionHash != encodeHash(txHash) {
		t.Fatalf("transactionHash = %s, want %s", receipt.TransactionHash, encodeHash(txHash))
	}
}
