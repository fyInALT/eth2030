package engine

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core"
	"github.com/eth2030/eth2030/core/state"
	"github.com/eth2030/eth2030/core/types"
)

// --- SPEC-5.5: engine_getPayloadBodiesByHashV2 / ByRangeV2 ---

func makePayloadBodiesAPI(t *testing.T) (*EngineAPI, *EngineBackend) {
	t.Helper()
	statedb := state.NewMemoryStateDB()
	genesis := makeGenesis()
	backend := NewEngineBackend(core.TestConfig, statedb, genesis)
	api := NewEngineAPI(backend)
	return api, backend
}

// TestGetPayloadBodiesByHashV2_UnknownHash verifies that unknown hashes return nil entries.
func TestGetPayloadBodiesByHashV2_UnknownHash(t *testing.T) {
	api, _ := makePayloadBodiesAPI(t)

	unknown := types.Hash{0xde, 0xad}
	results, err := api.GetPayloadBodiesByHashV2([]types.Hash{unknown})
	if err != nil {
		t.Fatalf("GetPayloadBodiesByHashV2: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}
	if results[0] != nil {
		t.Errorf("unknown hash: want nil body, got non-nil")
	}
}

// TestGetPayloadBodiesByHashV2_GenesisBlock verifies genesis block returns a body.
func TestGetPayloadBodiesByHashV2_GenesisBlock(t *testing.T) {
	api, _ := makePayloadBodiesAPI(t)
	genesis := makeGenesis()
	h := genesis.Hash()

	results, err := api.GetPayloadBodiesByHashV2([]types.Hash{h})
	if err != nil {
		t.Fatalf("GetPayloadBodiesByHashV2: %v", err)
	}
	if len(results) != 1 || results[0] == nil {
		t.Fatalf("want non-nil body for genesis, got nil")
	}
	body := results[0]
	if body.Transactions == nil {
		t.Error("Transactions must be non-nil slice")
	}
	if body.Withdrawals == nil {
		t.Error("Withdrawals must be non-nil slice")
	}
}

// TestGetPayloadBodiesByHashV2_NoBAL verifies that no BAL stored means nil BlockAccessList.
func TestGetPayloadBodiesByHashV2_NoBAL(t *testing.T) {
	api, _ := makePayloadBodiesAPI(t)
	genesis := makeGenesis()
	h := genesis.Hash()

	results, err := api.GetPayloadBodiesByHashV2([]types.Hash{h})
	if err != nil {
		t.Fatalf("GetPayloadBodiesByHashV2: %v", err)
	}
	if results[0] == nil {
		t.Fatal("expected non-nil body for genesis")
	}
	if results[0].BlockAccessList != nil {
		t.Errorf("expected nil BlockAccessList when no BAL stored")
	}
}

// TestGetPayloadBodiesByRangeV2_ZeroCount verifies count=0 returns an error.
func TestGetPayloadBodiesByRangeV2_ZeroCount(t *testing.T) {
	api, _ := makePayloadBodiesAPI(t)

	_, err := api.GetPayloadBodiesByRangeV2(0, 0)
	if err == nil {
		t.Fatal("expected error for count=0")
	}
}

// TestGetPayloadBodiesByRangeV2_ExceedsLimit verifies count>1024 returns an error.
func TestGetPayloadBodiesByRangeV2_ExceedsLimit(t *testing.T) {
	api, _ := makePayloadBodiesAPI(t)

	_, err := api.GetPayloadBodiesByRangeV2(0, 1025)
	if err == nil {
		t.Fatal("expected error for count=1025")
	}
}

// TestGetPayloadBodiesByRangeV2_MaxCount verifies count=1024 is accepted.
func TestGetPayloadBodiesByRangeV2_MaxCount(t *testing.T) {
	api, _ := makePayloadBodiesAPI(t)

	results, err := api.GetPayloadBodiesByRangeV2(0, 1024)
	if err != nil {
		t.Fatalf("GetPayloadBodiesByRangeV2(0, 1024): %v", err)
	}
	if uint64(len(results)) != 1024 {
		t.Errorf("want 1024 results, got %d", len(results))
	}
}

// TestGetPayloadBodiesByRangeV2_Genesis verifies genesis block is found by range.
func TestGetPayloadBodiesByRangeV2_Genesis(t *testing.T) {
	api, _ := makePayloadBodiesAPI(t)
	genesis := makeGenesis()
	genesisNum := genesis.NumberU64()

	results, err := api.GetPayloadBodiesByRangeV2(genesisNum, 1)
	if err != nil {
		t.Fatalf("GetPayloadBodiesByRangeV2: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}
	if results[0] == nil {
		t.Error("expected non-nil body for genesis block number")
	}
}

// TestHandleGetPayloadBodiesByHashV2_Handler verifies the JSON-RPC handler routing.
func TestHandleGetPayloadBodiesByHashV2_Handler(t *testing.T) {
	api, _ := makePayloadBodiesAPI(t)
	genesis := makeGenesis()
	h := genesis.Hash()

	hashesJSON, _ := json.Marshal([]types.Hash{h})
	params := []json.RawMessage{hashesJSON}

	result, rpcErr := api.handleGetPayloadBodiesByHashV2(params)
	if rpcErr != nil {
		t.Fatalf("handleGetPayloadBodiesByHashV2: code=%d msg=%s", rpcErr.Code, rpcErr.Message)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestHandleGetPayloadBodiesByRangeV2_Handler verifies the JSON-RPC handler routing.
func TestHandleGetPayloadBodiesByRangeV2_Handler(t *testing.T) {
	api, _ := makePayloadBodiesAPI(t)

	startJSON, _ := json.Marshal(uint64(0))
	countJSON, _ := json.Marshal(uint64(1))
	params := []json.RawMessage{startJSON, countJSON}

	result, rpcErr := api.handleGetPayloadBodiesByRangeV2(params)
	if rpcErr != nil {
		t.Fatalf("handleGetPayloadBodiesByRangeV2: code=%d msg=%s", rpcErr.Code, rpcErr.Message)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// TestExecutionPayloadBodyV2_Structure verifies the struct fields.
func TestExecutionPayloadBodyV2_Structure(t *testing.T) {
	body := &ExecutionPayloadBodyV2{
		Transactions:    [][]byte{{0x01, 0x02}},
		Withdrawals:     []*Withdrawal{{Index: 0, ValidatorIndex: 1, Amount: 100}},
		BlockAccessList: []byte(`{}`),
	}
	if len(body.Transactions) != 1 {
		t.Errorf("expected 1 tx, got %d", len(body.Transactions))
	}
	if len(body.Withdrawals) != 1 {
		t.Errorf("expected 1 withdrawal, got %d", len(body.Withdrawals))
	}
	if string(body.BlockAccessList) != `{}` {
		t.Errorf("BlockAccessList mismatch")
	}
}

// TestGetPayloadBodiesByHashV2_NonBackend verifies non-EngineBackend returns error.
func TestGetPayloadBodiesByHashV2_NonBackend(t *testing.T) {
	api := NewEngineAPI(&handlerMockBackend{})
	_, err := api.GetPayloadBodiesByHashV2(nil)
	if err == nil {
		t.Fatal("expected error when backend is not EngineBackend")
	}
}

// TestGetPayloadBodiesByRangeV2_NonBackend verifies non-EngineBackend returns error.
func TestGetPayloadBodiesByRangeV2_NonBackend(t *testing.T) {
	api := NewEngineAPI(&handlerMockBackend{})
	_, err := api.GetPayloadBodiesByRangeV2(0, 1)
	if err == nil {
		t.Fatal("expected error when backend is not EngineBackend")
	}
}

// TestBlockToPayloadBodyV2_EmptyBlock verifies conversion of an empty block.
func TestBlockToPayloadBodyV2_EmptyBlock(t *testing.T) {
	blobGas := uint64(0)
	excessBlobGas := uint64(0)
	header := &types.Header{
		Number:        big.NewInt(1),
		GasLimit:      30_000_000,
		BaseFee:       big.NewInt(1_000_000_000),
		Difficulty:    new(big.Int),
		UncleHash:     types.EmptyUncleHash,
		Root:          types.EmptyRootHash,
		TxHash:        types.EmptyRootHash,
		ReceiptHash:   types.EmptyRootHash,
		Time:          1700000001,
		BlobGasUsed:   &blobGas,
		ExcessBlobGas: &excessBlobGas,
	}
	block := types.NewBlock(header, nil)
	body := blockToPayloadBodyV2(block)
	if body == nil {
		t.Fatal("blockToPayloadBodyV2 returned nil")
	}
	if body.Transactions == nil {
		t.Error("Transactions must be non-nil")
	}
	if body.Withdrawals == nil {
		t.Error("Withdrawals must be non-nil")
	}
	if body.BlockAccessList != nil {
		t.Error("BlockAccessList should be nil for no BAL")
	}
}
