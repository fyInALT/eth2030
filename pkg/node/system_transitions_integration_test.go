package node

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/eth2030/eth2030/core/eips"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/engine/payload"
	gethcore "github.com/ethereum/go-ethereum/core"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	gethrlp "github.com/ethereum/go-ethereum/rlp"
)

func TestCanonicalSystemContractsThroughNodeHTTPAndRestart(t *testing.T) {
	const token = "canonical-system-http"
	var genesis map[string]any
	if err := json.Unmarshal(builderRequestGenesisJSON(1337), &genesis); err != nil {
		t.Fatal(err)
	}
	forks := genesis["config"].(map[string]any)
	forks["amsterdamTime"], forks["osakaTime"], forks["terminalTotalDifficulty"] = 0, 0, 0
	alloc := genesis["alloc"].(map[string]any)
	for address, code := range map[types.Address][]byte{eips.HistoryStorageAddress: eips.HistoryStorageRuntime(), eips.BeaconRootAddress: eips.BeaconRootsRuntime()} {
		alloc[address.Hex()] = map[string]any{"balance": "0", "nonce": "0x1", "code": "0x" + hex.EncodeToString(code)}
	}
	input, err := json.Marshal(genesis)
	if err != nil {
		t.Fatal(err)
	}
	var sdkGenesis gethcore.Genesis
	if err := json.Unmarshal(input, &sdkGenesis); err != nil {
		t.Fatal(err)
	}
	wantGenesis := types.Hash(sdkGenesis.ToBlock().Hash())
	makeCfg := func() *Config {
		cfg := makeTestConfig(t)
		cfg.EngineAuthToken = token
		cfg.GenesisPath = filepath.Join(cfg.DataDir, "genesis.json")
		if err := os.WriteFile(cfg.GenesisPath, input, 0600); err != nil {
			t.Fatal(err)
		}
		return &cfg
	}
	sourceCfg, targetCfg := makeCfg(), makeCfg()
	source, target := newTestNode(t, sourceCfg), newTestNode(t, targetCfg)
	writer, reader := track4NodeRPC(t, source, token), track4NodeRPC(t, target, token)
	if source.blockchain.Genesis().Hash() != wantGenesis || target.blockchain.Genesis().Hash() != wantGenesis {
		t.Fatal("explicit allocations changed genesis after root computation")
	}
	if source.blockchain.State().GetState(eips.HistoryStorageAddress, types.Hash{}) != (types.Hash{}) {
		t.Fatal("history was written during genesis")
	}
	parent := wantGenesis
	roots := []types.Hash{{31: 0x55}, {31: 0x66}}
	times := []uint64{12, 98304}
	var latest payload.ExecutionPayloadV5
	for i, time := range times {
		attrs := map[string]any{"timestamp": fmt.Sprintf("0x%x", time), "slotNumber": fmt.Sprintf("0x%x", time/12), "prevRandao": types.Hash{}, "withdrawals": []any{}, "suggestedFeeRecipient": types.Address{19: 1}, "parentBeaconBlockRoot": roots[i], "targetGasLimit": "0x1c9c380"}
		var fc payload.ForkchoiceUpdatedResult
		writer.call(t, "engine_forkchoiceUpdatedV4", &fc, payload.ForkchoiceStateV1{HeadBlockHash: parent}, attrs)
		if fc.PayloadID == nil {
			t.Fatal("payload ID missing")
		}
		var built struct {
			ExecutionPayload  payload.ExecutionPayloadV5
			ExecutionRequests []string
		}
		writer.call(t, "engine_getPayloadV6", &built, fc.PayloadID)
		latest = built.ExecutionPayload
		if latest.GasUsed != 0 || len(latest.Transactions) != 0 {
			t.Fatal("pre-block systems debited user/block gas")
		}
		for _, rpc := range []track4RPC{writer, reader} {
			var status payload.PayloadStatusV1
			rpc.call(t, "engine_newPayloadV5", &status, latest, []types.Hash{}, roots[i], built.ExecutionRequests)
			if status.Status != "VALID" {
				t.Fatalf("actual system import: %+v", status)
			}
			rpc.call(t, "engine_forkchoiceUpdatedV4", &fc, payload.ForkchoiceStateV1{HeadBlockHash: latest.BlockHash}, nil)
			if fc.PayloadStatus.Status != "VALID" {
				t.Fatal("head selection failed")
			}
		}
		var history, beacon string
		reader.call(t, "eth_call", &history, map[string]any{"to": eips.HistoryStorageAddress, "data": eips.Uint64ToHash(uint64(i)).Hex()}, "latest")
		if history != parent.Hex() {
			t.Fatalf("canonical runtime history query %s want %s", history, parent)
		}
		reader.call(t, "eth_call", &beacon, map[string]any{"to": eips.BeaconRootAddress, "data": eips.Uint64ToHash(time).Hex()}, "latest")
		if beacon != roots[i].Hex() {
			t.Fatal("canonical beacon root getter disagrees")
		}
		if err := reader.invoke("eth_call", &history, map[string]any{"to": eips.HistoryStorageAddress, "data": eips.Uint64ToHash(uint64(i + 1)).Hex()}, "latest"); err == nil {
			t.Fatal("future/current block history query accepted")
		}
		if err := reader.invoke("eth_call", &history, map[string]any{"to": eips.HistoryStorageAddress, "data": "0x01"}, "latest"); err == nil {
			t.Fatal("short history query accepted")
		}
		if i == 1 {
			if err := reader.invoke("eth_call", &beacon, map[string]any{"to": eips.BeaconRootAddress, "data": eips.Uint64ToHash(times[0]).Hex()}, "latest"); err == nil {
				t.Fatal("overwritten beacon timestamp returned stale root")
			}
		}
		header := target.blockchain.CurrentBlock().Header()
		raw, err := header.EncodeRLP()
		if err != nil {
			t.Fatal(err)
		}
		var independent gethtypes.Header
		if err := gethrlp.DecodeBytes(raw, &independent); err != nil || types.Hash(independent.Hash()) != latest.BlockHash {
			t.Fatal("independent complete header hash differs")
		}
		parent = latest.BlockHash
	}
	root := target.blockchain.CurrentBlock().Root()
	if err := target.Stop(); err != nil {
		t.Fatal(err)
	}
	reopened := newTestNode(t, targetCfg)
	if reopened.blockchain.CurrentBlock().Hash() != latest.BlockHash || reopened.blockchain.CurrentBlock().Root() != root {
		t.Fatal("system transition changed after restart")
	}
	reread := track4NodeRPC(t, reopened, token)
	var value string
	reread.call(t, "eth_call", &value, map[string]any{"to": eips.BeaconRootAddress, "data": eips.Uint64ToHash(times[1]).Hex()}, "latest")
	if value != roots[1].Hex() {
		t.Fatal("canonical stored beacon root not readable after restart")
	}
	t.Logf("head=%s root=%s canonical history=8191 beacon wrap timestamp=%d", latest.BlockHash, root, times[1])
}
