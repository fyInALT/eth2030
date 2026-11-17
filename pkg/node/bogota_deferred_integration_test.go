package node

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/engine/payload"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
)

type bogotaIntegrationStatus struct {
	Status                 string
	LatestValidHash        *types.Hash
	InclusionListSatisfied *bool
}

type bogotaIntegrationPayload struct {
	ExecutionPayload  payload.ExecutionPayloadV5
	ExecutionRequests []string
}

func TestBogotaDeferredWireAcrossRestart(t *testing.T) {
	const token = "bogota-deferred-integration"
	key, err := gethcrypto.HexToECDSA("0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		t.Fatal(err)
	}
	var genesis map[string]any
	if err := json.Unmarshal(minimalGenesisJSON(1337), &genesis); err != nil {
		t.Fatal(err)
	}
	spec := genesis["config"].(map[string]any)
	spec["terminalTotalDifficulty"], spec["amsterdamTime"], spec["bogotaTime"] = 0, 0, 48
	genesis["alloc"].(map[string]any)[gethcrypto.PubkeyToAddress(key.PublicKey).Hex()] = map[string]string{"balance": "1000000000000000000"}
	wire, err := json.Marshal(genesis)
	if err != nil {
		t.Fatal(err)
	}
	config := func() *Config {
		t.Helper()
		cfg := makeTestConfig(t)
		cfg.GenesisPath, cfg.EngineAuthToken = filepath.Join(cfg.DataDir, "genesis.json"), token
		if err := os.WriteFile(cfg.GenesisPath, wire, 0600); err != nil {
			t.Fatal(err)
		}
		return &cfg
	}
	sourceNode := newTestNode(t, config())
	receiverConfig := config()
	receiverNode := newTestNode(t, receiverConfig)
	source := track4NodeRPC(t, sourceNode, token)
	receiver := track4NodeRPC(t, receiverNode, token)
	root := sourceNode.Blockchain().Genesis().Hash()
	beacon := types.Hash{31: 1}
	if !sourceNode.Blockchain().Config().IsHogota(60) {
		t.Fatal("fixture did not activate Bogotá")
	}
	build := func(parent types.Hash, ts uint64) bogotaIntegrationPayload {
		t.Helper()
		attrs := map[string]any{
			"timestamp": fmt.Sprintf("0x%x", ts), "slotNumber": fmt.Sprintf("0x%x", ts/12),
			"withdrawals": []any{}, "prevRandao": types.Hash{}, "suggestedFeeRecipient": types.Address{19: 1},
			"parentBeaconBlockRoot": beacon,
		}
		method := "engine_forkchoiceUpdatedV4"
		if ts >= 48 {
			method = "engine_forkchoiceUpdatedV5"
			attrs["targetGasLimit"], attrs["inclusionListTransactions"] = "0x1c9c380", []string{}
		}
		var fc payload.ForkchoiceUpdatedResult
		source.call(t, method, &fc, payload.ForkchoiceStateV1{HeadBlockHash: parent}, attrs)
		if fc.PayloadStatus.Status != "VALID" || fc.PayloadID == nil {
			t.Fatalf("build at %d: %+v", ts, fc)
		}
		var envelope bogotaIntegrationPayload
		source.call(t, "engine_getPayloadV6", &envelope, fc.PayloadID)
		return envelope
	}
	importAmsterdam := func(destination track4RPC, envelope bogotaIntegrationPayload, selectHead bool) {
		t.Helper()
		var status payload.PayloadStatusV1
		p := envelope.ExecutionPayload
		destination.call(t, "engine_newPayloadV5", &status, p, []types.Hash{}, beacon, envelope.ExecutionRequests)
		if status.Status != "VALID" {
			t.Fatalf("pre-Bogotá import %s: %+v", p.BlockHash, status)
		}
		if selectHead {
			var fc payload.ForkchoiceUpdatedResult
			destination.call(t, "engine_forkchoiceUpdatedV4", &fc, payload.ForkchoiceStateV1{HeadBlockHash: p.BlockHash}, nil)
			if fc.PayloadStatus.Status != "VALID" {
				t.Fatalf("select pre-Bogotá block: %+v", fc)
			}
		}
	}

	canonical := build(root, 12)
	side := build(root, 24)
	importAmsterdam(receiver, canonical, true)
	importAmsterdam(source, side, true)
	accepted := build(side.ExecutionPayload.BlockHash, 60)
	hash := accepted.ExecutionPayload.BlockHash
	newPayload := func(destination track4RPC, list []string) bogotaIntegrationStatus {
		t.Helper()
		var status bogotaIntegrationStatus
		destination.call(t, "engine_newPayloadV6", &status,
			accepted.ExecutionPayload, []types.Hash{}, beacon, accepted.ExecutionRequests, list)
		return status
	}
	assertStatus := func(status bogotaIntegrationStatus, want string, satisfied *bool) {
		t.Helper()
		if status.Status != want || (status.InclusionListSatisfied == nil) != (satisfied == nil) ||
			(satisfied != nil && *status.InclusionListSatisfied != *satisfied) {
			t.Fatalf("status %s, satisfaction %v, want %s with %v", status.Status, status.InclusionListSatisfied, want, satisfied)
		}
		if want == "VALID" && (status.LatestValidHash == nil || *status.LatestValidHash != hash) {
			t.Fatalf("VALID did not identify the executed block: %+v", status)
		}
	}
	assertStatus(newPayload(receiver, []string{}), "SYNCING", nil)
	importAmsterdam(receiver, side, false)
	badSlot := accepted.ExecutionPayload
	badSlot.SlotNumber = "0x6"
	var invalid bogotaIntegrationStatus
	receiver.call(t, "engine_newPayloadV6", &invalid, badSlot, []types.Hash{}, beacon, accepted.ExecutionRequests, []string{})
	assertStatus(invalid, "INVALID", nil)
	assertStatus(newPayload(receiver, []string{}), "ACCEPTED", nil)

	to := common.Address{19: 0x89}
	tx, err := gethtypes.SignTx(gethtypes.NewTx(&gethtypes.LegacyTx{
		To: &to, Gas: 21000, GasPrice: big.NewInt(2_000_000_000), Value: big.NewInt(1),
	}), gethtypes.NewEIP155Signer(big.NewInt(1337)), key)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if len(accepted.ExecutionPayload.Transactions) != 0 {
		t.Fatal("fixture must omit the required transfer")
	}
	list := []string{"0x" + hex.EncodeToString(raw)}
	assertStatus(newPayload(receiver, list), "ACCEPTED", nil)
	if receiverNode.Blockchain().HasValidatedBlock(hash) {
		t.Fatal("ACCEPTED executed the candidate before FCU")
	}
	restart := func() {
		t.Helper()
		if err := receiverNode.Stop(); err != nil {
			t.Fatal(err)
		}
		receiverNode = newTestNode(t, receiverConfig)
		receiver = track4NodeRPC(t, receiverNode, token)
	}
	restart()
	if receiverNode.Blockchain().CurrentBlock().Hash() != canonical.ExecutionPayload.BlockHash {
		t.Fatal("staged block changed canonical history before FCU")
	}
	selectAccepted := func(want bool) {
		t.Helper()
		var result struct{ PayloadStatus bogotaIntegrationStatus }
		receiver.call(t, "engine_forkchoiceUpdatedV5", &result, payload.ForkchoiceStateV1{HeadBlockHash: hash}, nil)
		assertStatus(result.PayloadStatus, "VALID", &want)
		if receiverNode.Blockchain().CurrentBlock().Hash() != hash {
			t.Fatal("FCU did not execute and select the retained block")
		}
	}
	selectAccepted(false)
	restart()
	selectAccepted(false)
	satisfied := true
	assertStatus(newPayload(receiver, []string{}), "VALID", &satisfied)
	selectAccepted(true)

	// A node that obtained the block through ordinary sync has execution state,
	// but no basis for claiming it satisfied this CL's list.
	syncedNode := newTestNode(t, config())
	synced := track4NodeRPC(t, syncedNode, token)
	importAmsterdam(synced, canonical, true)
	importAmsterdam(synced, side, false)
	if err := syncedNode.Blockchain().InsertSyncedBlock(receiverNode.Blockchain().GetBlock(hash)); err != nil {
		t.Fatal(err)
	}
	var unresolved struct{ PayloadStatus bogotaIntegrationStatus }
	synced.call(t, "engine_forkchoiceUpdatedV5", &unresolved, payload.ForkchoiceStateV1{HeadBlockHash: hash}, nil)
	assertStatus(unresolved.PayloadStatus, "SYNCING", nil)
	if syncedNode.Blockchain().CurrentBlock().Hash() != canonical.ExecutionPayload.BlockHash {
		t.Fatal("FCU selected a block with an unresolved list view")
	}
	assertStatus(newPayload(synced, list), "VALID", new(bool))
	synced.call(t, "engine_forkchoiceUpdatedV5", &unresolved, payload.ForkchoiceStateV1{HeadBlockHash: hash}, nil)
	assertStatus(unresolved.PayloadStatus, "VALID", new(bool))
	if syncedNode.Blockchain().CurrentBlock().Hash() != hash {
		t.Fatal("FCU did not select the block after receiving its list view")
	}
}
