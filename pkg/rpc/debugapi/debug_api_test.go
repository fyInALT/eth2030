package debugapi

import (
	"encoding/json"
	"math/big"
	"testing"

	coretypes "github.com/eth2030/eth2030/core/types"
	testutil "github.com/eth2030/eth2030/rpc/internal/testutil"
	rpctypes "github.com/eth2030/eth2030/rpc/types"
)

// encodeHash converts a Hash to its 0x-prefixed hex string.
func encodeHash(h coretypes.Hash) string { return rpctypes.EncodeHash(h) }

// callDebugRPC dispatches a request through the DebugAPI HandleDebugRequest handler.
func callDebugRPC(t *testing.T, api *DebugAPI, method string, params ...interface{}) *rpctypes.Response {
	t.Helper()
	var rawParams []json.RawMessage
	for _, p := range params {
		b, _ := json.Marshal(p)
		rawParams = append(rawParams, json.RawMessage(b))
	}
	req := &rpctypes.Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  rawParams,
		ID:      json.RawMessage(`1`),
	}
	return api.HandleDebugRequest(req)
}

// ---------- debug_traceBlockByNumber ----------

func TestDebugTraceBlockByNumber(t *testing.T) {
	mb := testutil.NewMockBackend()

	// Create two transactions in block 42.
	to := coretypes.HexToAddress("0xbbbb")
	tx1 := coretypes.NewTransaction(&coretypes.LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(1000000000),
		Gas:      21000,
		To:       &to,
		Value:    big.NewInt(100),
	})
	sender := coretypes.HexToAddress("0xaaaa")
	tx1.SetSender(sender)

	tx2 := coretypes.NewTransaction(&coretypes.LegacyTx{
		Nonce:    1,
		GasPrice: big.NewInt(1000000000),
		Gas:      42000,
		To:       &to,
		Value:    big.NewInt(200),
	})
	tx2.SetSender(sender)

	// Build the block with these transactions.
	block := coretypes.NewBlock(mb.Headers[42], &coretypes.Body{
		Transactions: []*coretypes.Transaction{tx1, tx2},
	})
	mb.Blocks[42] = block

	// Add receipts.
	blockHash := block.Hash()
	mb.Receipts[blockHash] = []*coretypes.Receipt{
		{
			Status:            coretypes.ReceiptStatusSuccessful,
			GasUsed:           21000,
			CumulativeGasUsed: 21000,
			TxHash:            tx1.Hash(),
			BlockHash:         blockHash,
			BlockNumber:       big.NewInt(42),
			TransactionIndex:  0,
			Logs:              []*coretypes.Log{},
		},
		{
			Status:            coretypes.ReceiptStatusFailed,
			GasUsed:           15000,
			CumulativeGasUsed: 36000,
			TxHash:            tx2.Hash(),
			BlockHash:         blockHash,
			BlockNumber:       big.NewInt(42),
			TransactionIndex:  1,
			Logs:              []*coretypes.Log{},
		},
	}

	api := NewDebugAPI(mb)
	resp := callDebugRPC(t, api, "debug_traceBlockByNumber", "0x2a")

	if resp.Error != nil {
		t.Fatalf("error: %v", resp.Error.Message)
	}

	results, ok := resp.Result.([]*DebugBlockTraceEntry)
	if !ok {
		t.Fatalf("result not []*DebugBlockTraceEntry: %T", resp.Result)
	}
	if len(results) != 2 {
		t.Fatalf("want 2 trace results, got %d", len(results))
	}

	// First tx should succeed with 21000 gas.
	if results[0].Result.Gas != 21000 {
		t.Fatalf("want gas 21000 for tx1, got %d", results[0].Result.Gas)
	}
	if results[0].Result.Failed {
		t.Fatal("tx1 should not be marked as failed")
	}
	if results[0].TxHash != encodeHash(tx1.Hash()) {
		t.Fatalf("want tx1 hash %v, got %v", encodeHash(tx1.Hash()), results[0].TxHash)
	}

	// Second tx should fail with 15000 gas.
	if results[1].Result.Gas != 15000 {
		t.Fatalf("want gas 15000 for tx2, got %d", results[1].Result.Gas)
	}
	if !results[1].Result.Failed {
		t.Fatal("tx2 should be marked as failed")
	}
}

func TestDebugTraceBlockByNumber_Latest(t *testing.T) {
	mb := testutil.NewMockBackend()

	to := coretypes.HexToAddress("0xbbbb")
	tx := coretypes.NewTransaction(&coretypes.LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(1000000000),
		Gas:      21000,
		To:       &to,
		Value:    big.NewInt(100),
	})

	block := coretypes.NewBlock(mb.Headers[42], &coretypes.Body{
		Transactions: []*coretypes.Transaction{tx},
	})
	mb.Blocks[42] = block

	blockHash := block.Hash()
	mb.Receipts[blockHash] = []*coretypes.Receipt{
		{
			Status:  coretypes.ReceiptStatusSuccessful,
			GasUsed: 21000,
			TxHash:  tx.Hash(),
			Logs:    []*coretypes.Log{},
		},
	}

	api := NewDebugAPI(mb)
	resp := callDebugRPC(t, api, "debug_traceBlockByNumber", "latest")

	if resp.Error != nil {
		t.Fatalf("error: %v", resp.Error.Message)
	}
	results := resp.Result.([]*DebugBlockTraceEntry)
	if len(results) != 1 {
		t.Fatalf("want 1 trace result, got %d", len(results))
	}
}

func TestDebugTraceBlockByNumber_EmptyBlock(t *testing.T) {
	mb := testutil.NewMockBackend()

	block := coretypes.NewBlock(mb.Headers[42], nil)
	mb.Blocks[42] = block

	api := NewDebugAPI(mb)
	resp := callDebugRPC(t, api, "debug_traceBlockByNumber", "0x2a")

	if resp.Error != nil {
		t.Fatalf("error: %v", resp.Error.Message)
	}
	results := resp.Result.([]*DebugBlockTraceEntry)
	if len(results) != 0 {
		t.Fatalf("want 0 trace results for empty block, got %d", len(results))
	}
}

func TestDebugTraceBlockByNumber_NotFound(t *testing.T) {
	mb := testutil.NewMockBackend()
	api := NewDebugAPI(mb)

	resp := callDebugRPC(t, api, "debug_traceBlockByNumber", "0x999")
	if resp.Error == nil {
		t.Fatal("expected error for non-existent block")
	}
}

func TestDebugTraceBlockByNumber_MissingParam(t *testing.T) {
	api := NewDebugAPI(testutil.NewMockBackend())
	resp := callDebugRPC(t, api, "debug_traceBlockByNumber")
	if resp.Error == nil {
		t.Fatal("expected error for missing parameter")
	}
	if resp.Error.Code != rpctypes.ErrCodeInvalidParams {
		t.Fatalf("want error code %d, got %d", rpctypes.ErrCodeInvalidParams, resp.Error.Code)
	}
}

// ---------- debug_traceBlockByHash ----------

func TestDebugTraceBlockByHash(t *testing.T) {
	mb := testutil.NewMockBackend()

	to := coretypes.HexToAddress("0xbbbb")
	tx := coretypes.NewTransaction(&coretypes.LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(1000000000),
		Gas:      21000,
		To:       &to,
		Value:    big.NewInt(100),
	})
	sender := coretypes.HexToAddress("0xaaaa")
	tx.SetSender(sender)

	block := coretypes.NewBlock(mb.Headers[42], &coretypes.Body{
		Transactions: []*coretypes.Transaction{tx},
	})
	mb.Blocks[42] = block

	blockHash := block.Hash()
	mb.Receipts[blockHash] = []*coretypes.Receipt{
		{
			Status:            coretypes.ReceiptStatusSuccessful,
			GasUsed:           21000,
			CumulativeGasUsed: 21000,
			TxHash:            tx.Hash(),
			BlockHash:         blockHash,
			BlockNumber:       big.NewInt(42),
			TransactionIndex:  0,
			Logs:              []*coretypes.Log{},
		},
	}

	api := NewDebugAPI(mb)
	resp := callDebugRPC(t, api, "debug_traceBlockByHash", encodeHash(blockHash))

	if resp.Error != nil {
		t.Fatalf("error: %v", resp.Error.Message)
	}

	results, ok := resp.Result.([]*DebugBlockTraceEntry)
	if !ok {
		t.Fatalf("result not []*DebugBlockTraceEntry: %T", resp.Result)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 trace result, got %d", len(results))
	}
	if results[0].Result.Gas != 21000 {
		t.Fatalf("want gas 21000, got %d", results[0].Result.Gas)
	}
	if results[0].Result.Failed {
		t.Fatal("tx should not be marked as failed")
	}
	if results[0].TxHash != encodeHash(tx.Hash()) {
		t.Fatalf("want tx hash %v, got %v", encodeHash(tx.Hash()), results[0].TxHash)
	}
}

func TestDebugTraceBlockByHash_NotFound(t *testing.T) {
	mb := testutil.NewMockBackend()
	api := NewDebugAPI(mb)

	resp := callDebugRPC(t, api, "debug_traceBlockByHash",
		"0x0000000000000000000000000000000000000000000000000000000000000000")
	if resp.Error == nil {
		t.Fatal("expected error for non-existent block")
	}
}

func TestDebugTraceBlockByHash_MissingParam(t *testing.T) {
	api := NewDebugAPI(testutil.NewMockBackend())
	resp := callDebugRPC(t, api, "debug_traceBlockByHash")
	if resp.Error == nil {
		t.Fatal("expected error for missing parameter")
	}
	if resp.Error.Code != rpctypes.ErrCodeInvalidParams {
		t.Fatalf("want error code %d, got %d", rpctypes.ErrCodeInvalidParams, resp.Error.Code)
	}
}

// ---------- debug_traceBlockByNumber with no receipts ----------

func TestDebugTraceBlockByNumber_NoReceipts(t *testing.T) {
	mb := testutil.NewMockBackend()

	to := coretypes.HexToAddress("0xbbbb")
	tx := coretypes.NewTransaction(&coretypes.LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(1000000000),
		Gas:      21000,
		To:       &to,
		Value:    big.NewInt(100),
	})

	block := coretypes.NewBlock(mb.Headers[42], &coretypes.Body{
		Transactions: []*coretypes.Transaction{tx},
	})
	mb.Blocks[42] = block
	// No receipts stored -- should fallback to tx.Gas().

	api := NewDebugAPI(mb)
	resp := callDebugRPC(t, api, "debug_traceBlockByNumber", "0x2a")

	if resp.Error != nil {
		t.Fatalf("error: %v", resp.Error.Message)
	}
	results := resp.Result.([]*DebugBlockTraceEntry)
	if len(results) != 1 {
		t.Fatalf("want 1 trace result, got %d", len(results))
	}
	// Without receipts, falls back to tx gas limit.
	if results[0].Result.Gas != 21000 {
		t.Fatalf("want gas 21000, got %d", results[0].Result.Gas)
	}
}
