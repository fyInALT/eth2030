package rpc

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/types"
)

// ---------- eth_getBlockReceipts enhanced fields ----------

func TestGetBlockReceipts_EffectiveGasPrice(t *testing.T) {
	mb := newMockBackend()
	blockHash := mb.headers[42].Hash()

	receipt := &types.Receipt{
		Status:            types.ReceiptStatusSuccessful,
		Type:              types.DynamicFeeTxType,
		CumulativeGasUsed: 21000,
		GasUsed:           21000,
		EffectiveGasPrice: big.NewInt(2000000000),
		TxHash:            types.HexToHash("0x1111"),
		BlockHash:         blockHash,
		BlockNumber:       big.NewInt(42),
		TransactionIndex:  0,
		Logs:              []*types.Log{},
	}
	mb.receipts[blockHash] = []*types.Receipt{receipt}

	api := NewEthAPI(mb)
	resp := callRPC(t, api, "eth_getBlockReceipts", "0x2a")
	if resp.Error != nil {
		t.Fatalf("error: %v", resp.Error.Message)
	}

	receipts := resp.Result.([]*RPCReceipt)
	if len(receipts) != 1 {
		t.Fatalf("want 1 receipt, got %d", len(receipts))
	}
	r := receipts[0]
	if r.Type != "0x2" {
		t.Fatalf("want type 0x2, got %v", r.Type)
	}
	if r.EffectiveGasPrice != "0x77359400" {
		t.Fatalf("want effectiveGasPrice 0x77359400, got %v", r.EffectiveGasPrice)
	}
	if r.Status != "0x1" {
		t.Fatalf("want status 0x1, got %v", r.Status)
	}
}

func TestGetBlockReceipts_EIP4844Fields(t *testing.T) {
	mb := newMockBackend()
	blockHash := mb.headers[42].Hash()

	receipt := &types.Receipt{
		Status:            types.ReceiptStatusSuccessful,
		Type:              types.BlobTxType,
		CumulativeGasUsed: 21000,
		GasUsed:           21000,
		EffectiveGasPrice: big.NewInt(1000000000),
		BlobGasUsed:       131072,
		BlobGasPrice:      big.NewInt(1000),
		TxHash:            types.HexToHash("0x3333"),
		BlockHash:         blockHash,
		BlockNumber:       big.NewInt(42),
		TransactionIndex:  0,
		Logs:              []*types.Log{},
	}
	mb.receipts[blockHash] = []*types.Receipt{receipt}

	api := NewEthAPI(mb)
	resp := callRPC(t, api, "eth_getBlockReceipts", "0x2a")
	if resp.Error != nil {
		t.Fatalf("error: %v", resp.Error.Message)
	}

	receipts := resp.Result.([]*RPCReceipt)
	if len(receipts) != 1 {
		t.Fatalf("want 1 receipt, got %d", len(receipts))
	}
	r := receipts[0]
	if r.Type != "0x3" {
		t.Fatalf("want type 0x3, got %v", r.Type)
	}
	if r.BlobGasUsed == nil || *r.BlobGasUsed != "0x20000" {
		t.Fatalf("want blobGasUsed 0x20000, got %v", r.BlobGasUsed)
	}
	if r.BlobGasPrice == nil || *r.BlobGasPrice != "0x3e8" {
		t.Fatalf("want blobGasPrice 0x3e8, got %v", r.BlobGasPrice)
	}
}

func TestGetBlockReceipts_NoBlobFields_LegacyTx(t *testing.T) {
	mb := newMockBackend()
	blockHash := mb.headers[42].Hash()

	receipt := &types.Receipt{
		Status:            types.ReceiptStatusSuccessful,
		Type:              types.LegacyTxType,
		CumulativeGasUsed: 21000,
		GasUsed:           21000,
		EffectiveGasPrice: big.NewInt(1000000000),
		TxHash:            types.HexToHash("0x4444"),
		BlockHash:         blockHash,
		BlockNumber:       big.NewInt(42),
		TransactionIndex:  0,
		Logs:              []*types.Log{},
	}
	mb.receipts[blockHash] = []*types.Receipt{receipt}

	api := NewEthAPI(mb)
	resp := callRPC(t, api, "eth_getBlockReceipts", "0x2a")
	if resp.Error != nil {
		t.Fatalf("error: %v", resp.Error.Message)
	}

	receipts := resp.Result.([]*RPCReceipt)
	r := receipts[0]
	if r.BlobGasUsed != nil {
		t.Fatalf("want nil blobGasUsed for legacy tx, got %v", *r.BlobGasUsed)
	}
	if r.BlobGasPrice != nil {
		t.Fatalf("want nil blobGasPrice for legacy tx, got %v", *r.BlobGasPrice)
	}
	if r.Type != "0x0" {
		t.Fatalf("want type 0x0, got %v", r.Type)
	}
}

func TestGetBlockReceipts_LogIndexing(t *testing.T) {
	mb := newMockBackend()
	blockHash := mb.headers[42].Hash()

	contractAddr := types.HexToAddress("0xcccc")
	topic := types.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

	receipt1 := &types.Receipt{
		Status:            types.ReceiptStatusSuccessful,
		CumulativeGasUsed: 21000,
		GasUsed:           21000,
		TxHash:            types.HexToHash("0x1111"),
		BlockHash:         blockHash,
		BlockNumber:       big.NewInt(42),
		TransactionIndex:  0,
		Logs: []*types.Log{
			{Address: contractAddr, Topics: []types.Hash{topic}, Data: []byte{0x01}, BlockNumber: 42, BlockHash: blockHash, TxIndex: 0, Index: 0},
			{Address: contractAddr, Topics: []types.Hash{topic}, Data: []byte{0x02}, BlockNumber: 42, BlockHash: blockHash, TxIndex: 0, Index: 1},
		},
	}
	receipt2 := &types.Receipt{
		Status:            types.ReceiptStatusSuccessful,
		CumulativeGasUsed: 42000,
		GasUsed:           21000,
		TxHash:            types.HexToHash("0x2222"),
		BlockHash:         blockHash,
		BlockNumber:       big.NewInt(42),
		TransactionIndex:  1,
		Logs: []*types.Log{
			{Address: contractAddr, Topics: []types.Hash{topic}, Data: []byte{0x03}, BlockNumber: 42, BlockHash: blockHash, TxIndex: 1, Index: 2},
		},
	}
	mb.receipts[blockHash] = []*types.Receipt{receipt1, receipt2}

	api := NewEthAPI(mb)
	resp := callRPC(t, api, "eth_getBlockReceipts", "0x2a")
	if resp.Error != nil {
		t.Fatalf("error: %v", resp.Error.Message)
	}

	receipts := resp.Result.([]*RPCReceipt)
	if len(receipts) != 2 {
		t.Fatalf("want 2 receipts, got %d", len(receipts))
	}
	if len(receipts[0].Logs) != 2 {
		t.Fatalf("want 2 logs in receipt 0, got %d", len(receipts[0].Logs))
	}
	if receipts[0].Logs[0].LogIndex != "0x0" {
		t.Fatalf("want logIndex 0x0, got %v", receipts[0].Logs[0].LogIndex)
	}
	if receipts[1].Logs[0].LogIndex != "0x2" {
		t.Fatalf("want logIndex 0x2, got %v", receipts[1].Logs[0].LogIndex)
	}
}

// ---------- JSON serialization round-trip ----------

func TestRPCReceipt_JSON_WithBlobFields(t *testing.T) {
	bgu := "0x20000"
	bgp := "0x3e8"
	receipt := &RPCReceipt{
		TransactionHash:   "0x1111",
		TransactionIndex:  "0x0",
		BlockHash:         "0x2222",
		BlockNumber:       "0x2a",
		GasUsed:           "0x5208",
		CumulativeGasUsed: "0x5208",
		Status:            "0x1",
		LogsBloom:         "0x00",
		Type:              "0x3",
		EffectiveGasPrice: "0x3b9aca00",
		BlobGasUsed:       &bgu,
		BlobGasPrice:      &bgp,
		Logs:              []*RPCLog{},
	}

	data, err := json.Marshal(receipt)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if decoded["blobGasUsed"] != "0x20000" {
		t.Fatalf("want blobGasUsed 0x20000, got %v", decoded["blobGasUsed"])
	}
	if decoded["type"] != "0x3" {
		t.Fatalf("want type 0x3, got %v", decoded["type"])
	}
}

func TestRPCReceipt_JSON_NoBlobFields(t *testing.T) {
	receipt := &RPCReceipt{
		TransactionHash:   "0x1111",
		TransactionIndex:  "0x0",
		BlockHash:         "0x2222",
		BlockNumber:       "0x2a",
		GasUsed:           "0x5208",
		CumulativeGasUsed: "0x5208",
		Status:            "0x1",
		LogsBloom:         "0x00",
		Type:              "0x0",
		EffectiveGasPrice: "0x3b9aca00",
		Logs:              []*RPCLog{},
	}

	data, _ := json.Marshal(receipt)
	var decoded map[string]interface{}
	json.Unmarshal(data, &decoded)

	if _, exists := decoded["blobGasUsed"]; exists {
		t.Fatal("blobGasUsed should be omitted for non-blob tx")
	}
	if _, exists := decoded["blobGasPrice"]; exists {
		t.Fatal("blobGasPrice should be omitted for non-blob tx")
	}
}

// ---------- Dispatcher routing ----------

func TestDispatcher_DebugTraceBlockByNumber(t *testing.T) {
	mb := newMockBackend()
	block := types.NewBlock(mb.headers[42], nil)
	mb.blocks[42] = block

	api := NewEthAPI(mb)
	resp := callRPC(t, api, "debug_traceBlockByNumber", "0x2a")
	if resp.Error != nil {
		t.Fatalf("expected routing to succeed, got error: %v", resp.Error.Message)
	}
}

func TestDispatcher_DebugTraceBlockByHash(t *testing.T) {
	mb := newMockBackend()
	block := types.NewBlock(mb.headers[42], nil)
	mb.blocks[42] = block

	api := NewEthAPI(mb)
	resp := callRPC(t, api, "debug_traceBlockByHash", encodeHash(block.Hash()))
	if resp.Error != nil {
		t.Fatalf("expected routing to succeed, got error: %v", resp.Error.Message)
	}
}
