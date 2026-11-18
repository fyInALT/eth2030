package node

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eth2030/eth2030/core/eips"
	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/engine"
	"github.com/eth2030/eth2030/engine/payload"
	"github.com/eth2030/eth2030/trie/bintrie"
)

// builderRequestGenesisJSON deploys the owned test contracts before root creation.
func builderRequestGenesisJSON(chainID uint64) []byte {
	var g map[string]any
	if err := json.Unmarshal(minimalGenesisJSON(chainID), &g); err != nil {
		panic(err)
	}
	alloc := g["alloc"].(map[string]any)
	alloc[types.BuilderDepositRequestAddress.Hex()] = map[string]any{"balance": "0", "nonce": "0x1", "code": "0x" + hex.EncodeToString(eips.BuilderDepositRuntime())}
	inhibitor := types.Hash{}
	for i := range inhibitor {
		inhibitor[i] = 255
	}
	alloc[types.BuilderExitRequestAddress.Hex()] = map[string]any{"balance": "0", "nonce": "0x1", "code": "0x" + hex.EncodeToString(eips.BuilderExitRuntime()), "storage": map[string]string{(types.Hash{}).Hex(): inhibitor.Hex()}}
	data, _ := json.Marshal(g)
	return data
}

// minimalGenesisJSON builds a small genesis.json payload for testing.
func minimalGenesisJSON(chainID uint64) []byte {
	g := map[string]any{
		"config": map[string]any{
			"chainId":                 chainID,
			"terminalTotalDifficulty": "0x0",
			"homesteadBlock":          0,
			"eip155Block":             0,
			"eip158Block":             0,
			"byzantiumBlock":          0,
			"constantinopleBlock":     0,
			"petersburgBlock":         0,
			"istanbulBlock":           0,
			"berlinBlock":             0,
			"londonBlock":             0,
			"shanghaiTime":            0,
			"cancunTime":              0,
			"pragueTime":              0,
			"osakaTime":               0,
		},
		"nonce":      "0x0",
		"timestamp":  "0x0",
		"extraData":  "0x",
		"gasLimit":   "0x1C9C380",
		"difficulty": "0x0",
		"alloc": map[string]any{
			"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266": map[string]any{
				"balance": "1000000000000000000000",
			},
		},
	}
	data, _ := json.Marshal(g)
	return data
}

func TestNodeRejectsIncoherentCanonicalScheduleBeforeGenesisAdoption(t *testing.T) {
	for _, field := range []string{"shanghaiTime", "cancunTime", "pragueTime", "osakaTime"} {
		t.Run(field, func(t *testing.T) {
			var input map[string]any
			if err := json.Unmarshal(minimalGenesisJSON(1337), &input); err != nil {
				t.Fatal(err)
			}
			forks := input["config"].(map[string]any)
			forks["amsterdamTime"] = 100
			delete(forks, field)
			data, err := json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}
			cfg := DefaultConfig()
			cfg.DataDir, cfg.RPCPort, cfg.EnginePort = t.TempDir(), 0, 0
			cfg.GenesisPath = filepath.Join(cfg.DataDir, "genesis.json")
			if err := os.WriteFile(cfg.GenesisPath, data, 0600); err != nil {
				t.Fatal(err)
			}
			if node, err := New(&cfg); err == nil || node != nil || !strings.Contains(err.Error(), "fork schedule") {
				if node != nil {
					_ = node.Stop()
				}
				t.Fatalf("incoherent production startup accepted: node=%v error=%v", node, err)
			}
		})
	}
}

func TestLoadGenesisFile_Basic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "genesis.json")
	if err := os.WriteFile(path, minimalGenesisJSON(1337), 0600); err != nil {
		t.Fatalf("write genesis: %v", err)
	}

	cfg := DefaultConfig()
	cfg.DataDir = dir
	cfg.GenesisPath = path

	genesis, err := loadGenesisFile(&cfg)
	if err != nil {
		t.Fatalf("loadGenesisFile() error: %v", err)
	}
	if genesis == nil {
		t.Fatal("genesis should not be nil")
	}
	if genesis.Config == nil {
		t.Fatal("genesis.Config should not be nil")
	}
	if genesis.Config.ChainID == nil || genesis.Config.ChainID.Uint64() != 1337 {
		t.Errorf("ChainID = %v, want 1337", genesis.Config.ChainID)
	}
}

func TestLoadGenesisFile_NotFound(t *testing.T) {
	cfg := DefaultConfig()
	cfg.GenesisPath = "/nonexistent/genesis.json"
	_, err := loadGenesisFile(&cfg)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadGenesisFile_Invalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "genesis.json")
	if err := os.WriteFile(path, []byte("not json"), 0600); err != nil {
		t.Fatalf("write genesis: %v", err)
	}
	cfg := DefaultConfig()
	cfg.GenesisPath = path
	_, err := loadGenesisFile(&cfg)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadGenesisFile_ForkOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "genesis.json")
	if err := os.WriteFile(path, minimalGenesisJSON(9999), 0600); err != nil {
		t.Fatalf("write genesis: %v", err)
	}

	ts := uint64(1700000000)
	cfg := DefaultConfig()
	cfg.DataDir = dir
	cfg.GenesisPath = path
	cfg.HogotaOverride = &ts

	genesis, err := loadGenesisFile(&cfg)
	if err != nil {
		t.Fatalf("loadGenesisFile() error: %v", err)
	}
	if genesis.Config.HogotaTime == nil || *genesis.Config.HogotaTime != ts {
		t.Errorf("HogotaTime = %v, want %d", genesis.Config.HogotaTime, ts)
	}
}

func TestLoadGenesisFile_ToBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "genesis.json")
	if err := os.WriteFile(path, minimalGenesisJSON(42), 0600); err != nil {
		t.Fatalf("write genesis: %v", err)
	}

	cfg := DefaultConfig()
	cfg.DataDir = dir
	cfg.GenesisPath = path

	genesis, err := loadGenesisFile(&cfg)
	if err != nil {
		t.Fatalf("loadGenesisFile() error: %v", err)
	}
	block := genesis.ToBlock()
	if block == nil {
		t.Fatal("ToBlock() should not return nil")
	}
	if block.NumberU64() != 0 {
		t.Errorf("genesis block number = %d, want 0", block.NumberU64())
	}
}

func TestLoadGenesisFile_HexStringNonce(t *testing.T) {
	// Kurtosis devnet genesis files encode alloc nonce as a hex string, e.g. "0x1".
	// Ensure the loader doesn't error on this format.
	raw := []byte(`{
		"config": {"chainId": 3151908, "homesteadBlock": 0, "eip155Block": 0,
		           "eip158Block": 0, "byzantiumBlock": 0, "constantinopleBlock": 0,
		           "petersburgBlock": 0, "istanbulBlock": 0, "berlinBlock": 0,
		           "londonBlock": 0, "terminalTotalDifficulty": "0x0", "shanghaiTime": 0, "cancunTime": 0, "pragueTime": 0},
		"nonce": "0x0", "timestamp": "0x0", "extraData": "0x",
		"gasLimit": "0x1C9C380", "difficulty": "0x0",
		"alloc": {
			"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266": {
				"balance": "1000000000000000000000",
				"nonce": "0x1"
			}
		}
	}`)
	dir := t.TempDir()
	path := dir + "/genesis.json"
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatalf("write genesis: %v", err)
	}

	cfg := DefaultConfig()
	cfg.DataDir = dir
	cfg.GenesisPath = path

	genesis, err := loadGenesisFile(&cfg)
	if err != nil {
		t.Fatalf("loadGenesisFile() with hex nonce error: %v", err)
	}
	for _, acc := range genesis.Alloc {
		if acc.Nonce != 1 {
			t.Errorf("Nonce = %d, want 1", acc.Nonce)
		}
	}
}

func TestLoadGenesisFile_Storage(t *testing.T) {
	// Genesis accounts with storage slots must have their storage parsed and
	// included in the alloc so the state root matches geth's computation.
	raw := []byte(`{
		"config": {"chainId": 1, "homesteadBlock": 0, "eip155Block": 0,
		           "eip158Block": 0, "byzantiumBlock": 0, "constantinopleBlock": 0,
		           "petersburgBlock": 0, "istanbulBlock": 0, "berlinBlock": 0,
		           "londonBlock": 0, "terminalTotalDifficulty": "0x0", "shanghaiTime": 0, "cancunTime": 0, "pragueTime": 0},
		"nonce": "0x0", "timestamp": "0x0", "extraData": "0x",
		"gasLimit": "0x1C9C380", "difficulty": "0x0",
		"alloc": {
			"0x000F3df6D732807Ef1319fB7B8bB8522d0Beac02": {
				"balance": "0x0",
				"code": "0x3d602d80600a3d3981f3363d3d373d3d3d363d30545af43d82803e903d91601e57fd5bf3",
				"storage": {
					"0x0000000000000000000000000000000000000000000000000000000000000000": "0x000000000000000000000000000000000000000000000000000000000000000f",
					"0x000000000000000000000000000000000000000000000000000000000000000b": "0x0000000000000000000000006d2b2e15339a2a3cc2fe8b7d28c92ea28c3e78ff"
				}
			}
		}
	}`)
	dir := t.TempDir()
	path := dir + "/genesis.json"
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatalf("write genesis: %v", err)
	}

	cfg := DefaultConfig()
	cfg.DataDir = dir
	cfg.GenesisPath = path

	genesis, err := loadGenesisFile(&cfg)
	if err != nil {
		t.Fatalf("loadGenesisFile() with storage: %v", err)
	}

	addr := types.HexToAddress("0x000F3df6D732807Ef1319fB7B8bB8522d0Beac02")
	acc, ok := genesis.Alloc[addr]
	if !ok {
		t.Fatal("expected alloc entry not found")
	}
	if len(acc.Storage) != 2 {
		t.Errorf("storage len = %d, want 2", len(acc.Storage))
	}

	slot0 := types.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000")
	if v := acc.Storage[slot0]; v != types.HexToHash("0x000000000000000000000000000000000000000000000000000000000000000f") {
		t.Errorf("slot 0 value = %v, want 0x...0f", v)
	}
}

func TestParseBalance(t *testing.T) {
	tests := []struct {
		input string
		want  string // decimal string
	}{
		{"0", "0"},
		{"1000000000000000000000", "1000000000000000000000"},
		{"0x0", "0"},
		{"0xDE0B6B3A7640000", "1000000000000000000"},
		{"0x3635C9ADC5DEA00000", "1000000000000000000000"},
		{"", "0"},
	}
	for _, tc := range tests {
		got := parseBalance(tc.input)
		if got.String() != tc.want {
			t.Errorf("parseBalance(%q) = %s, want %s", tc.input, got, tc.want)
		}
	}
}

// kurtosisGenesisPath points to the real Kurtosis devnet genesis bundled in the repo.
const kurtosisGenesisPath = "../devnet/kurtosis/el_cl_genesis_data/genesis.json"

func TestLoadGenesisFile_Kurtosis(t *testing.T) {
	if _, err := os.Stat(kurtosisGenesisPath); err != nil {
		t.Skipf("kurtosis genesis not found: %v", err)
	}

	cfg := DefaultConfig()
	cfg.DataDir = t.TempDir()
	cfg.GenesisPath = kurtosisGenesisPath

	genesis, err := loadGenesisFile(&cfg)
	if err != nil {
		t.Fatalf("loadGenesisFile() error: %v", err)
	}

	// Chain config.
	if genesis.Config == nil {
		t.Fatal("Config is nil")
	}
	if genesis.Config.ChainID == nil || genesis.Config.ChainID.Int64() != 3151908 {
		t.Errorf("ChainID = %v, want 3151908", genesis.Config.ChainID)
	}

	// Fork timestamps — all zero in this devnet genesis.
	for name, ts := range map[string]*uint64{
		"Shanghai": genesis.Config.ShanghaiTime,
		"Cancun":   genesis.Config.CancunTime,
		"Prague":   genesis.Config.PragueTime,
	} {
		if ts == nil || *ts != 0 {
			t.Errorf("%s time = %v, want 0", name, ts)
		}
	}

	// Alloc: 282 entries.
	if len(genesis.Alloc) != 282 {
		t.Errorf("alloc len = %d, want 282", len(genesis.Alloc))
	}

	// Timestamp and gas limit from the file.
	if genesis.Timestamp != 1772279152 {
		t.Errorf("Timestamp = %d, want 1772279152", genesis.Timestamp)
	}
	if genesis.GasLimit != 0x3938700 {
		t.Errorf("GasLimit = %x, want 3938700", genesis.GasLimit)
	}
}

func TestLoadGenesisFile_Kurtosis_AllocDetails(t *testing.T) {
	if _, err := os.Stat(kurtosisGenesisPath); err != nil {
		t.Skipf("kurtosis genesis not found: %v", err)
	}

	cfg := DefaultConfig()
	cfg.DataDir = t.TempDir()
	cfg.GenesisPath = kurtosisGenesisPath

	genesis, err := loadGenesisFile(&cfg)
	if err != nil {
		t.Fatalf("loadGenesisFile() error: %v", err)
	}

	// Accounts with code should have non-empty Code bytes.
	codeCount := 0
	for _, acc := range genesis.Alloc {
		if len(acc.Code) > 0 {
			codeCount++
		}
	}
	if codeCount != 5 {
		t.Errorf("accounts with code = %d, want 5", codeCount)
	}

	// Accounts with nonce > 0.
	nonceCount := 0
	for _, acc := range genesis.Alloc {
		if acc.Nonce > 0 {
			nonceCount++
		}
	}
	if nonceCount != 4 {
		t.Errorf("accounts with nonce > 0 = %d, want 4", nonceCount)
	}

	// All accounts must have a non-nil balance.
	for addr, acc := range genesis.Alloc {
		if acc.Balance == nil {
			t.Errorf("account %s has nil balance", addr)
		}
	}
}

func TestNodeWithGenesisFile_Kurtosis(t *testing.T) {
	if _, err := os.Stat(kurtosisGenesisPath); err != nil {
		t.Skipf("kurtosis genesis not found: %v", err)
	}

	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.DataDir = dir
	cfg.GenesisPath = kurtosisGenesisPath
	cfg.P2PPort = 0
	cfg.RPCPort = 0
	cfg.EnginePort = 0
	cfg.Network = ""

	n, err := New(&cfg)
	if err != nil {
		t.Fatalf("New() with kurtosis genesis error: %v", err)
	}
	if n.Blockchain() == nil {
		t.Error("blockchain should not be nil")
	}
	if n.Config().NetworkID != 3151908 {
		t.Errorf("NetworkID = %d, want 3151908", n.Config().NetworkID)
	}
}

func TestNodeWithGenesisFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "genesis.json")
	if err := os.WriteFile(path, minimalGenesisJSON(1337), 0600); err != nil {
		t.Fatalf("write genesis: %v", err)
	}

	cfg := DefaultConfig()
	cfg.DataDir = dir
	cfg.GenesisPath = path
	cfg.P2PPort = 0
	cfg.RPCPort = 0
	cfg.EnginePort = 0
	cfg.Network = "" // should not be needed with custom genesis

	n, err := New(&cfg)
	if err != nil {
		t.Fatalf("New() with genesis file error: %v", err)
	}
	if n.Blockchain() == nil {
		t.Error("blockchain should not be nil")
	}
	if n.Config().NetworkID != 1337 {
		t.Errorf("NetworkID = %d, want 1337 (from genesis chainId)", n.Config().NetworkID)
	}
}

func TestLoadGenesisFile_ScheduledForksRoundTrip(t *testing.T) {
	const genesisTime = uint64(100)
	var input map[string]any
	if err := json.Unmarshal(builderRequestGenesisJSON(1337), &input); err != nil {
		t.Fatal(err)
	}
	input["timestamp"] = "0x64"
	cfgFields := input["config"].(map[string]any)
	cfgFields["amsterdamTime"] = genesisTime + 18
	cfgFields["glamsterdanTime"] = genesisTime + 31
	cfgFields["hogotaTime"] = genesisTime + 43
	cfgFields["bpo2Time"] = genesisTime + 61
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	cfg := DefaultConfig()
	cfg.GenesisPath = filepath.Join(t.TempDir(), "genesis.json")
	if err := os.WriteFile(cfg.GenesisPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	genesis, err := loadGenesisFile(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if genesis.Timestamp != genesisTime || genesis.ToBlock().Time() != genesisTime {
		t.Fatalf("genesis time %d", genesis.Timestamp)
	}
	for _, fork := range []struct {
		at     uint64
		active func(uint64) bool
	}{
		{118, genesis.Config.IsAmsterdam},
		{131, genesis.Config.IsGlamsterdan},
		{143, genesis.Config.IsHogota},
		{161, genesis.Config.IsBPO2},
	} {
		if fork.active(fork.at-1) || !fork.active(fork.at) {
			t.Fatalf("fork did not switch at %d", fork.at)
		}
	}
}

func TestGenesisLoadsExplicitPQDraftProfile(t *testing.T) {
	var genesis map[string]any
	if err := json.Unmarshal(minimalGenesisJSON(88), &genesis); err != nil {
		t.Fatal(err)
	}
	genesis["config"].(map[string]any)["pqDraft"] = map[string]any{
		"time":       100,
		"sigRecover": "0x0000000000000000000000000000000000007001",
		"mldsa":      "0x0000000000000000000000000000000000007002",
		"mldsaEth":   "0x0000000000000000000000000000000000007003",
	}
	data, err := json.Marshal(genesis)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "genesis.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	spec, err := loadGenesisFile(&Config{GenesisPath: path})
	if err != nil {
		t.Fatal(err)
	}
	pq := spec.Config.PQDraft
	if pq == nil || pq.Time == nil || *pq.Time != 100 ||
		pq.SigRecover != (types.Address{18: 0x70, 19: 1}) ||
		pq.MLDSA != (types.Address{18: 0x70, 19: 2}) ||
		pq.MLDSAETH != (types.Address{18: 0x70, 19: 3}) {
		t.Fatalf("unexpected PQ draft configuration: %+v", pq)
	}
}

func TestGenesisRejectsAmbiguousPQDraftAddress(t *testing.T) {
	for _, value := range []string{"0x7001", "0x100000000000000000000000000000000000007001", "0x000000000000000000000000000000000000700z"} {
		var draft pqDraftJSON
		input := []byte(`{"sigRecover":"` + value + `"}`)
		if err := json.Unmarshal(input, &draft); err == nil {
			t.Fatalf("accepted ambiguous address %q", value)
		}
	}
}

func TestBinaryOverlayGenesisThroughEngine(t *testing.T) {
	var genesis map[string]any
	if err := json.Unmarshal(minimalGenesisJSON(88), &genesis); err != nil {
		t.Fatal(err)
	}
	spec := genesis["config"].(map[string]any)
	spec["binaryOverlayTime"] = 1
	spec["binaryConversionTime"] = 2
	spec["binaryConversionStride"] = 1
	data, err := json.Marshal(genesis)
	if err != nil {
		t.Fatal(err)
	}
	cfg := makeTestConfig(t)
	cfg.GenesisPath = filepath.Join(cfg.DataDir, "genesis.json")
	if err := os.WriteFile(cfg.GenesisPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	n := newTestNode(t, &cfg)
	bc := n.Blockchain()
	funded := types.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	build := func(parent *types.Block, at uint64, beacon byte, withdrawal *payload.Withdrawal) *payload.ExecutionPayloadV4 {
		t.Helper()
		attrs := new(payload.PayloadAttributesV4)
		attrs.Timestamp = at
		attrs.Withdrawals = []*payload.Withdrawal{}
		if withdrawal != nil {
			attrs.Withdrawals = append(attrs.Withdrawals, withdrawal)
		}
		attrs.ParentBeaconBlockRoot = types.Hash{31: beacon}
		update, err := n.engBackend.ForkchoiceUpdatedV4(payload.ForkchoiceStateV1{HeadBlockHash: parent.Hash()}, attrs)
		if err != nil || update.PayloadStatus.Status != engine.StatusValid || update.PayloadID == nil {
			t.Fatalf("forkchoice: %+v, %v", update, err)
		}
		result, err := n.engBackend.GetPayloadByID(*update.PayloadID)
		if err != nil {
			t.Fatal(err)
		}
		return result.ExecutionPayload
	}

	first := build(bc.Genesis(), 1, 11, nil)
	parent := bc.GetBlock(first.BlockHash)
	if parent == nil || first.StateRoot != parent.Root() || first.StateRoot == bc.Genesis().Root() {
		t.Fatal("Engine did not publish and import the overlay state root")
	}
	before, err := bc.StateAtBlock(parent)
	if err != nil {
		t.Fatal(err)
	}
	balance := before.GetBalance(funded)
	if balance.Sign() == 0 {
		t.Fatal("first child cannot read the frozen genesis balance")
	}
	converted := build(parent, 2, 12, nil)
	if converted.StateRoot == first.StateRoot {
		t.Fatal("Engine did not apply the first conversion unit")
	}
	// Missing system code performs no writes. Reconstruct the overlay root
	// from the one frozen account selected by the conversion stride.
	want := bintrie.NewWithHashFunc(bintrie.HashFunctionBlake3)
	updateAccount := func(address types.Address) {
		t.Helper()
		account := &types.Account{Balance: before.GetBalance(address), CodeHash: types.EmptyCodeHash.Bytes()}
		if err := want.UpdateAccount(address, account, 0); err != nil {
			t.Fatal(err)
		}
	}

	if first.StateRoot != want.Hash() {
		t.Fatalf("first Engine root = %s, want %s", first.StateRoot, want.Hash())
	}
	updateAccount(funded)
	blk := bc.GetBlock(converted.BlockHash)
	if blk == nil || blk.Root() != want.Hash() || converted.StateRoot != want.Hash() {
		t.Fatalf("second Engine root = %s, want %s", converted.StateRoot, want.Hash())
	}
	after, err := bc.StateAtBlock(blk)
	if err != nil || after.GetBalance(funded).Cmp(balance) != 0 {
		t.Fatalf("conversion changed the logical balance: %v", err)
	}

	// Missing beacon code cannot distinguish the state of two payloads.
	// Use a real withdrawal difference to test independent overlay branches.
	other := build(parent, 2, 13, &payload.Withdrawal{Address: funded, Amount: 1})
	if other.BlockHash == converted.BlockHash || other.StateRoot == converted.StateRoot {
		t.Fatal("competing payloads did not produce independent overlays")
	}
	headOnly := payload.ForkchoiceStateV1{HeadBlockHash: converted.BlockHash}
	invalid := new(payload.PayloadAttributesV4)
	invalid.Timestamp = converted.Timestamp
	if _, err := n.engBackend.ForkchoiceUpdatedV4(headOnly, invalid); !errors.Is(err, engine.ErrInvalidPayloadAttributes) {
		t.Fatalf("invalid build timestamp: %v", err)
	}
	if bc.CurrentBlock().Hash() != converted.BlockHash {
		t.Fatal("invalid build attributes rolled back a valid forkchoice")
	}
	headOnly.HeadBlockHash = other.BlockHash
	if update, err := n.engBackend.ForkchoiceUpdatedV4(headOnly, nil); err != nil || update.PayloadStatus.Status != engine.StatusValid {
		t.Fatalf("forkchoice reorg: %+v, %v", update, err)
	}
	if bc.CurrentBlock().Hash() != other.BlockHash || bc.State().GetRoot() != other.StateRoot {
		t.Fatal("Engine forkchoice did not select the side block and its overlay")
	}
}
