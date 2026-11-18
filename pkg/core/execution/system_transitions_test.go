package execution

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/eth2030/eth2030/core/config"
	"github.com/eth2030/eth2030/core/eips"
	"github.com/eth2030/eth2030/core/types"
	bridge "github.com/eth2030/eth2030/geth"
	gethcommon "github.com/ethereum/go-ethereum/common"
	gethcore "github.com/ethereum/go-ethereum/core"
	gethbal "github.com/ethereum/go-ethereum/core/types/bal"
	gethvm "github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

func TestCanonicalPreBlockSystemsAgainstIndependentSDK(t *testing.T) {
	history := types.HexToAddress("0x0000F90827F1C53a10cb7A02335B175320002935")
	beacon := types.HexToAddress("0x000F3df6D732807Ef1319fB7B8bB8522d0Beac02")
	for _, scenario := range []string{"deployed", "missing", "custom", "revert"} {
		t.Run(scenario, func(t *testing.T) {
			db := newBuilderRequestTestState()
			historyCode, beaconCode := []byte(nil), []byte(nil)
			switch scenario {
			case "deployed":
				historyCode, beaconCode = bytes.Clone(params.HistoryStorageCode), bytes.Clone(params.BeaconRootsCode)
			case "custom":
				historyCode, beaconCode = []byte{0x5f, 0x35, 0x60, 42, 0x55, 0}, []byte{0x5f, 0x35, 0x60, 43, 0x55, 0}
			case "revert":
				historyCode, beaconCode = []byte{0x60, 1, 0x60, 42, 0x55, 0x5f, 0x5f, 0xfd}, []byte{0x60, 1, 0x60, 43, 0x55, 0x5f, 0x5f, 0xfd}
			}
			for address, code := range map[types.Address][]byte{history: historyCode, beacon: beaconCode} {
				if code != nil {
					db.SetCode(address, code)
					db.SetNonce(address, 1)
				}
			}
			// Freeze complete owned allocations before either implementation runs.
			pre := make(map[string]bridge.PreAccount)
			for _, address := range []types.Address{history, beacon, types.BuilderDepositRequestAddress, types.BuilderExitRequestAddress} {
				if db.Exist(address) {
					pre[address.Hex()] = bridge.PreAccount{Balance: new(big.Int).Set(db.GetBalance(address)), Nonce: db.GetNonce(address), Code: bytes.Clone(db.GetCode(address)), Storage: map[types.Hash]types.Hash{{}: db.GetState(address, types.Hash{})}}
				}
			}
			reference, err := bridge.MakePreState(pre)
			if err != nil {
				t.Fatal(err)
			}
			defer reference.Close()
			root := types.Hash{31: 8}
			parent := types.Hash{31: 7}
			header := &types.Header{Number: big.NewInt(8192), Time: 8192, GasLimit: 30_000_000, BaseFee: big.NewInt(1), Difficulty: new(big.Int), ParentHash: parent, ParentBeaconRoot: &root}
			cfg := *config.TestConfig
			zero := uint64(0)
			cfg.AmsterdamTime, cfg.OsakaTime = &zero, &zero
			cfg.GlamsterdanTime = nil
			result, err := NewStateProcessor(&cfg).ProcessWithBAL(types.NewBlock(header, &types.Body{}), db)
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Receipts) != 0 {
				t.Fatal("system operations produced user receipts")
			}
			random := gethcommon.Hash{}
			context := gethvm.BlockContext{CanTransfer: gethcore.CanTransfer, Transfer: gethcore.Transfer, BlockNumber: header.Number, Time: header.Time, GasLimit: header.GasLimit, BaseFee: header.BaseFee, Difficulty: new(big.Int), Random: &random, CostPerStateByte: 1530}
			evm := gethvm.NewEVM(context, reference.StateDB, bridge.AmsterdamFixtureConfig(), gethvm.Config{})
			collector := gethbal.NewConstructionBlockAccessList()
			gethcore.ProcessBeaconBlockRoot(gethcommon.Hash(root), evm, collector)
			gethcore.ProcessParentBlockHash(gethcommon.Hash(parent), evm, collector)
			var requests [][]byte
			if err := gethcore.ProcessBuilderDepositQueue(&requests, evm.GetRules(), evm, 1, collector); err != nil {
				t.Fatal(err)
			}
			if err := gethcore.ProcessBuilderExitQueue(&requests, evm.GetRules(), evm, 1, collector); err != nil {
				t.Fatal(err)
			}
			nativeRoot, err := db.Commit()
			if err != nil {
				t.Fatal(err)
			}
			expected := reference.StateDB.IntermediateRoot(evm.GetRules())
			if nativeRoot != types.Hash(expected) {
				t.Fatalf("actual pre-block %s root %s differs from independent SDK %s", scenario, nativeRoot, expected)
			}
			if scenario == "deployed" && eips.GetHistoricalBlockHash(db, 8191) != parent {
				t.Fatal("native valid history readback used the wrong window")
			}
			if scenario == "deployed" && db.GetState(history, types.Hash{}) != parent {
				t.Fatal("8191 history wrap did not write canonical slot zero")
			}
			if scenario == "custom" && (db.GetState(history, types.Hash{31: 42}) != parent || db.GetState(beacon, types.Hash{31: 43}) != root) {
				t.Fatal("custom actual code was not executed")
			}
			if scenario == "missing" && (db.Exist(history) || db.Exist(beacon)) {
				t.Fatal("missing code was replaced by synthetic storage")
			}
			if db.Exist(types.SystemAddress) {
				t.Fatal("system caller state was fabricated")
			}
			for _, entry := range result.BlockAccessList.Entries {
				if scenario == "revert" && (entry.Address == history || entry.Address == beacon) && len(entry.StorageChanges) != 0 {
					t.Fatal("reverted system write survived in BAL")
				}
				if scenario == "deployed" && (entry.Address == history || entry.Address == beacon) && len(entry.StorageChanges) > 0 && entry.AccessIndex != 0 {
					t.Fatal("pre-block write used a transaction/post index")
				}
			}
			for _, entry := range result.BlockAccessList.Entries {
				if entry.Address == types.SystemAddress {
					t.Fatal("untouched system caller included in BAL")
				}
			}
			t.Logf("scenario=%s root=%s", scenario, nativeRoot)
		})
	}
}

func TestCanonicalSystemRuntimeAssetsAndExplicitAllocation(t *testing.T) {
	if eips.HistoryServeWindow != 8191 || eips.HistoryStorageAddress != types.Address(params.HistoryStorageAddress) || !bytes.Equal(eips.HistoryStorageRuntime(), params.HistoryStorageCode) || !bytes.Equal(eips.BeaconRootsRuntime(), params.BeaconRootsCode) {
		t.Fatal("native canonical system input differs from independent SDK")
	}
	db := newBuilderRequestTestState()
	custom := []byte{0}
	db.SetCode(eips.BeaconRootAddress, custom)
	db.SetNonce(eips.BeaconRootAddress, 7)
	eips.ApplyExecutionSystemGenesis(db)
	if !bytes.Equal(db.GetCode(eips.BeaconRootAddress), custom) || db.GetNonce(eips.BeaconRootAddress) != 7 || !bytes.Equal(db.GetCode(eips.HistoryStorageAddress), params.HistoryStorageCode) || db.GetNonce(eips.HistoryStorageAddress) != 1 {
		t.Fatal("explicit genesis allocation replaced custom state or omitted canonical history")
	}
	mutable := eips.HistoryStorageRuntime()
	mutable[0] ^= 1
	if bytes.Equal(mutable, eips.HistoryStorageRuntime()) {
		t.Fatal("runtime caller mutated canonical bytes")
	}
}
