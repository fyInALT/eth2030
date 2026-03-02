package node

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// minimalGenesisJSON builds a small genesis.json payload for testing.
func minimalGenesisJSON(chainID uint64) []byte {
	g := map[string]any{
		"config": map[string]any{
			"chainId":             chainID,
			"homesteadBlock":      0,
			"eip155Block":         0,
			"eip158Block":         0,
			"byzantiumBlock":      0,
			"constantinopleBlock": 0,
			"petersburgBlock":     0,
			"istanbulBlock":       0,
			"berlinBlock":         0,
			"londonBlock":         0,
			"shanghaiTime":        0,
			"cancunTime":          0,
			"pragueTime":          0,
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
