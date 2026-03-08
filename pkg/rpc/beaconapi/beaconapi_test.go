package beaconapi

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/state"
	coretypes "github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/core/vm"
	rpctypes "github.com/eth2030/eth2030/rpc/types"
	"github.com/eth2030/eth2030/trie"
)

// mockBeaconBackend implements rpcbackend.Backend for BeaconAPI tests.
type mockBeaconBackend struct {
	headers map[int64]*coretypes.Header
	current *coretypes.Header
}

func newMockBeaconBackend() *mockBeaconBackend {
	h0 := &coretypes.Header{
		Number:     big.NewInt(0),
		ParentHash: coretypes.Hash{},
		Root:       coretypes.HexToHash("0xaaaa"),
		TxHash:     coretypes.HexToHash("0xbbbb"),
		Time:       1606824023,
		GasLimit:   30000000,
		GasUsed:    21000,
	}
	h1 := &coretypes.Header{
		Number:     big.NewInt(1),
		ParentHash: h0.Hash(),
		Root:       coretypes.HexToHash("0xcccc"),
		TxHash:     coretypes.HexToHash("0xdddd"),
		Time:       1606824035,
		GasLimit:   30000000,
		GasUsed:    42000,
	}
	return &mockBeaconBackend{
		headers: map[int64]*coretypes.Header{0: h0, 1: h1},
		current: h1,
	}
}

func (m *mockBeaconBackend) HeaderByNumber(number rpctypes.BlockNumber) *coretypes.Header {
	if number == rpctypes.LatestBlockNumber {
		return m.current
	}
	return m.headers[int64(number)]
}
func (m *mockBeaconBackend) HeaderByHash(hash coretypes.Hash) *coretypes.Header {
	for _, h := range m.headers {
		if h.Hash() == hash {
			return h
		}
	}
	return nil
}
func (m *mockBeaconBackend) CurrentHeader() *coretypes.Header { return m.current }
func (m *mockBeaconBackend) ChainID() *big.Int                { return big.NewInt(1) }

func (m *mockBeaconBackend) BlockByNumber(rpctypes.BlockNumber) *coretypes.Block  { return nil }
func (m *mockBeaconBackend) BlockByHash(coretypes.Hash) *coretypes.Block          { return nil }
func (m *mockBeaconBackend) StateAt(coretypes.Hash) (state.StateDB, error)        { return nil, nil }
func (m *mockBeaconBackend) SendTransaction(*coretypes.Transaction) error          { return nil }
func (m *mockBeaconBackend) GetTransaction(coretypes.Hash) (*coretypes.Transaction, uint64, uint64) {
	return nil, 0, 0
}
func (m *mockBeaconBackend) SuggestGasPrice() *big.Int                       { return big.NewInt(0) }
func (m *mockBeaconBackend) GetReceipts(coretypes.Hash) []*coretypes.Receipt  { return nil }
func (m *mockBeaconBackend) GetLogs(coretypes.Hash) []*coretypes.Log          { return nil }
func (m *mockBeaconBackend) GetBlockReceipts(uint64) []*coretypes.Receipt     { return nil }
func (m *mockBeaconBackend) GetProof(coretypes.Address, []coretypes.Hash, rpctypes.BlockNumber) (*trie.AccountProof, error) {
	return nil, nil
}
func (m *mockBeaconBackend) EVMCall(coretypes.Address, *coretypes.Address, []byte, uint64, *big.Int, rpctypes.BlockNumber) ([]byte, uint64, error) {
	return nil, 0, nil
}
func (m *mockBeaconBackend) TraceTransaction(coretypes.Hash) (*vm.StructLogTracer, error) {
	return nil, nil
}
func (m *mockBeaconBackend) HistoryOldestBlock() uint64 { return 0 }

func makeBeaconAPI(t *testing.T) *BeaconAPI {
	t.Helper()
	backend := newMockBeaconBackend()
	cs := NewConsensusState()
	cs.FinalizedEpoch = 2
	cs.FinalizedRoot = coretypes.HexToHash("0x1111")
	cs.JustifiedEpoch = 3
	cs.JustifiedRoot = coretypes.HexToHash("0x2222")
	cs.Peers = []*BeaconPeer{
		{PeerID: "peer1", State: "connected", Direction: "inbound", Address: "10.0.0.1:9000"},
	}
	cs.Validators = []*ValidatorEntry{
		{
			Index:   "0",
			Balance: "32000000000",
			Status:  "active_ongoing",
			Validator: &ValidatorData{
				Pubkey:                "0xaabb",
				WithdrawalCredentials: "0x00",
				EffectiveBalance:      "32000000000",
				Slashed:               false,
				ActivationEpoch:       "0",
				ExitEpoch:             "18446744073709551615",
			},
		},
	}
	return NewBeaconAPI(cs, backend)
}

func beaconReq(method string, params ...interface{}) *rpctypes.Request {
	rawParams := make([]json.RawMessage, len(params))
	for i, p := range params {
		b, _ := json.Marshal(p)
		rawParams[i] = b
	}
	return &rpctypes.Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  rawParams,
		ID:      json.RawMessage(`1`),
	}
}

func TestBeaconGetGenesis(t *testing.T) {
	api := makeBeaconAPI(t)
	resp := api.getGenesis(beaconReq("beacon_getGenesis"))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error.Message)
	}
	data, _ := json.Marshal(resp.Result)
	var genesis GenesisResponse
	json.Unmarshal(data, &genesis)
	if genesis.GenesisTime != "1606824023" {
		t.Errorf("genesis time = %q, want 1606824023", genesis.GenesisTime)
	}
	if genesis.GenesisForkVersion != "0x00000000" {
		t.Errorf("fork version = %q, want 0x00000000", genesis.GenesisForkVersion)
	}
}

func TestBeaconGetBlock(t *testing.T) {
	api := makeBeaconAPI(t)
	resp := api.getBlock(beaconReq("beacon_getBlock", "0x0"))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error.Message)
	}
	data, _ := json.Marshal(resp.Result)
	var block BlockResponse
	json.Unmarshal(data, &block)
	if block.Slot != "0" {
		t.Errorf("slot = %q, want 0", block.Slot)
	}
	resp = api.getBlock(beaconReq("beacon_getBlock", "0x999"))
	if resp.Error == nil {
		t.Fatal("expected error for non-existent slot")
	}
	if resp.Error.Code != BeaconErrNotFound {
		t.Errorf("error code = %d, want %d", resp.Error.Code, BeaconErrNotFound)
	}
}

func TestBeaconGetBlockHeader(t *testing.T) {
	api := makeBeaconAPI(t)
	resp := api.getBlockHeader(beaconReq("beacon_getBlockHeader", "0x1"))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error.Message)
	}
	data, _ := json.Marshal(resp.Result)
	var header HeaderResponse
	json.Unmarshal(data, &header)
	if !header.Canonical {
		t.Error("expected canonical = true")
	}
	if header.Header == nil || header.Header.Message == nil {
		t.Fatal("header message is nil")
	}
	if header.Header.Message.Slot != "1" {
		t.Errorf("slot = %q, want 1", header.Header.Message.Slot)
	}
}

func TestBeaconGetStateRoot(t *testing.T) {
	api := makeBeaconAPI(t)
	resp := api.getStateRoot(beaconReq("beacon_getStateRoot", "head"))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error.Message)
	}
	data, _ := json.Marshal(resp.Result)
	var sr StateRootResponse
	json.Unmarshal(data, &sr)
	if sr.Root == "" {
		t.Error("state root is empty")
	}
}

func TestBeaconGetStateFinalityCheckpoints(t *testing.T) {
	api := makeBeaconAPI(t)
	resp := api.getStateFinalityCheckpoints(beaconReq("beacon_getStateFinalityCheckpoints", "head"))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error.Message)
	}
	data, _ := json.Marshal(resp.Result)
	var cp FinalityCheckpointsResponse
	json.Unmarshal(data, &cp)
	if cp.Finalized == nil {
		t.Fatal("finalized checkpoint is nil")
	}
	if cp.Finalized.Epoch != "2" {
		t.Errorf("finalized epoch = %q, want 2", cp.Finalized.Epoch)
	}
	if cp.CurrentJustified == nil {
		t.Fatal("current justified checkpoint is nil")
	}
	if cp.CurrentJustified.Epoch != "3" {
		t.Errorf("justified epoch = %q, want 3", cp.CurrentJustified.Epoch)
	}
}

func TestBeaconGetStateValidators(t *testing.T) {
	api := makeBeaconAPI(t)
	resp := api.getStateValidators(beaconReq("beacon_getStateValidators", "head"))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error.Message)
	}
	data, _ := json.Marshal(resp.Result)
	var vl ValidatorListResponse
	json.Unmarshal(data, &vl)
	if len(vl.Validators) != 1 {
		t.Fatalf("validators count = %d, want 1", len(vl.Validators))
	}
	if vl.Validators[0].Status != "active_ongoing" {
		t.Errorf("validator status = %q, want active_ongoing", vl.Validators[0].Status)
	}
}

func TestBeaconGetNodeVersion(t *testing.T) {
	api := makeBeaconAPI(t)
	resp := api.getNodeVersion(beaconReq("beacon_getNodeVersion"))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error.Message)
	}
	data, _ := json.Marshal(resp.Result)
	var ver VersionResponse
	json.Unmarshal(data, &ver)
	if ver.Version != "ETH2030/v0.1.0-beacon" {
		t.Errorf("version = %q, want ETH2030/v0.1.0-beacon", ver.Version)
	}
}

func TestBeaconGetNodeSyncing(t *testing.T) {
	api := makeBeaconAPI(t)
	resp := api.getNodeSyncing(beaconReq("beacon_getNodeSyncing"))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error.Message)
	}
	data, _ := json.Marshal(resp.Result)
	var sync SyncingResponse
	json.Unmarshal(data, &sync)
	if sync.HeadSlot != "1" {
		t.Errorf("head slot = %q, want 1", sync.HeadSlot)
	}
	if sync.IsSyncing {
		t.Error("expected is_syncing = false")
	}
}

func TestBeaconGetNodePeers(t *testing.T) {
	api := makeBeaconAPI(t)
	resp := api.getNodePeers(beaconReq("beacon_getNodePeers"))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error.Message)
	}
	data, _ := json.Marshal(resp.Result)
	var peers PeerListResponse
	json.Unmarshal(data, &peers)
	if len(peers.Peers) != 1 {
		t.Fatalf("peer count = %d, want 1", len(peers.Peers))
	}
	if peers.Peers[0].PeerID != "peer1" {
		t.Errorf("peer ID = %q, want peer1", peers.Peers[0].PeerID)
	}
}

func TestBeaconGetNodeHealth(t *testing.T) {
	api := makeBeaconAPI(t)
	resp := api.getNodeHealth(beaconReq("beacon_getNodeHealth"))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error.Message)
	}
	data, _ := json.Marshal(resp.Result)
	var health map[string]string
	json.Unmarshal(data, &health)
	if health["status"] != "healthy" {
		t.Errorf("status = %q, want healthy", health["status"])
	}
	api.state.IsSyncing = true
	resp = api.getNodeHealth(beaconReq("beacon_getNodeHealth"))
	data, _ = json.Marshal(resp.Result)
	json.Unmarshal(data, &health)
	if health["status"] != "syncing" {
		t.Errorf("status = %q, want syncing", health["status"])
	}
}

func TestRegisterBeaconRoutes(t *testing.T) {
	api := makeBeaconAPI(t)
	routes := RegisterBeaconRoutes(api)
	expected := []string{
		"beacon_getGenesis", "beacon_getBlock", "beacon_getBlockHeader",
		"beacon_getStateRoot", "beacon_getStateFinalityCheckpoints",
		"beacon_getStateValidators", "beacon_getNodeVersion",
		"beacon_getNodeSyncing", "beacon_getNodePeers", "beacon_getNodeHealth",
	}
	for _, method := range expected {
		if _, ok := routes[method]; !ok {
			t.Errorf("missing route: %s", method)
		}
	}
}

func TestBeaconGetBlockMissingParams(t *testing.T) {
	api := makeBeaconAPI(t)
	resp := api.getBlock(&rpctypes.Request{
		JSONRPC: "2.0",
		Method:  "beacon_getBlock",
		Params:  nil,
		ID:      json.RawMessage(`1`),
	})
	if resp.Error == nil {
		t.Fatal("expected error for missing params")
	}
	if resp.Error.Code != BeaconErrBadRequest {
		t.Errorf("error code = %d, want %d", resp.Error.Code, BeaconErrBadRequest)
	}
}
