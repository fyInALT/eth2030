package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eth2030/eth2030/node"
)

func TestParseFlags_Defaults(t *testing.T) {
	cfg, exit, code := parseFlags([]string{})
	if exit {
		t.Fatalf("unexpected exit with code %d", code)
	}

	defaults := node.DefaultConfig()
	if cfg.DataDir != defaults.DataDir {
		t.Errorf("DataDir = %q, want %q", cfg.DataDir, defaults.DataDir)
	}
	if cfg.P2PPort != 30303 {
		t.Errorf("P2PPort = %d, want 30303", cfg.P2PPort)
	}
	if cfg.RPCPort != 8545 {
		t.Errorf("RPCPort = %d, want 8545", cfg.RPCPort)
	}
	if cfg.EnginePort != 8551 {
		t.Errorf("EnginePort = %d, want 8551", cfg.EnginePort)
	}
	if cfg.HTTPAddr != "127.0.0.1" {
		t.Errorf("HTTPAddr = %q, want 127.0.0.1", cfg.HTTPAddr)
	}
	if cfg.AuthAddr != "127.0.0.1" {
		t.Errorf("AuthAddr = %q, want 127.0.0.1", cfg.AuthAddr)
	}
	if cfg.SyncMode != "snap" {
		t.Errorf("SyncMode = %q, want snap", cfg.SyncMode)
	}
	if cfg.NetworkID != 1 {
		t.Errorf("NetworkID = %d, want 1", cfg.NetworkID)
	}
	if cfg.MaxPeers != 50 {
		t.Errorf("MaxPeers = %d, want 50", cfg.MaxPeers)
	}
	if cfg.Verbosity != 3 {
		t.Errorf("Verbosity = %d, want 3", cfg.Verbosity)
	}
	if cfg.Metrics {
		t.Error("Metrics should be false by default")
	}
	if cfg.WSEnabled {
		t.Error("WSEnabled should be false by default")
	}
	if cfg.MetricsPort != 9001 {
		t.Errorf("MetricsPort = %d, want 9001", cfg.MetricsPort)
	}
}

func TestParseFlags_AllOriginalFlags(t *testing.T) {
	args := []string{
		"-datadir", "/tmp/testdata",
		"-port", "30304",
		"-http.port", "9545",
		"-authrpc.port", "9551",
		"-syncmode", "full",
		"-networkid", "11155111",
		"-maxpeers", "25",
		"-verbosity", "4",
		"-metrics",
	}

	cfg, exit, _ := parseFlags(args)
	if exit {
		t.Fatal("unexpected exit")
	}

	if cfg.DataDir != "/tmp/testdata" {
		t.Errorf("DataDir = %q, want /tmp/testdata", cfg.DataDir)
	}
	if cfg.P2PPort != 30304 {
		t.Errorf("P2PPort = %d, want 30304", cfg.P2PPort)
	}
	if cfg.RPCPort != 9545 {
		t.Errorf("RPCPort = %d, want 9545", cfg.RPCPort)
	}
	if cfg.EnginePort != 9551 {
		t.Errorf("EnginePort = %d, want 9551", cfg.EnginePort)
	}
	if cfg.SyncMode != "full" {
		t.Errorf("SyncMode = %q, want full", cfg.SyncMode)
	}
	if cfg.NetworkID != 11155111 {
		t.Errorf("NetworkID = %d, want 11155111", cfg.NetworkID)
	}
	if cfg.MaxPeers != 25 {
		t.Errorf("MaxPeers = %d, want 25", cfg.MaxPeers)
	}
	if cfg.Verbosity != 4 {
		t.Errorf("Verbosity = %d, want 4", cfg.Verbosity)
	}
	if !cfg.Metrics {
		t.Error("Metrics should be true")
	}
}

func TestParseFlags_DevnetP0(t *testing.T) {
	tmp := t.TempDir()
	jwtPath := filepath.Join(tmp, "jwt.hex")

	args := []string{
		"--http.addr", "0.0.0.0",
		"--authrpc.addr", "0.0.0.0",
		"--authrpc.jwtsecret", jwtPath,
		"--bootnodes", "enode://abc@1.2.3.4:30303,enode://def@5.6.7.8:30303",
		"--metrics",
		"--metrics.addr", "0.0.0.0",
		"--metrics.port", "9001",
	}

	cfg, exit, _ := parseFlags(args)
	if exit {
		t.Fatal("unexpected exit")
	}
	if cfg.HTTPAddr != "0.0.0.0" {
		t.Errorf("HTTPAddr = %q, want 0.0.0.0", cfg.HTTPAddr)
	}
	if cfg.AuthAddr != "0.0.0.0" {
		t.Errorf("AuthAddr = %q, want 0.0.0.0", cfg.AuthAddr)
	}
	if cfg.JWTSecret != jwtPath {
		t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, jwtPath)
	}
	if cfg.Bootnodes != "enode://abc@1.2.3.4:30303,enode://def@5.6.7.8:30303" {
		t.Errorf("Bootnodes = %q", cfg.Bootnodes)
	}
	if !cfg.Metrics {
		t.Error("Metrics should be true")
	}
	if cfg.MetricsAddr != "0.0.0.0" {
		t.Errorf("MetricsAddr = %q, want 0.0.0.0", cfg.MetricsAddr)
	}
	if cfg.MetricsPort != 9001 {
		t.Errorf("MetricsPort = %d, want 9001", cfg.MetricsPort)
	}
}

func TestParseFlags_OverrideGenesis(t *testing.T) {
	args := []string{"--override.genesis", "/tmp/genesis.json"}
	cfg, exit, _ := parseFlags(args)
	if exit {
		t.Fatal("unexpected exit")
	}
	if cfg.GenesisPath != "/tmp/genesis.json" {
		t.Errorf("GenesisPath = %q, want /tmp/genesis.json", cfg.GenesisPath)
	}
}

func TestParseFlags_ForkOverrides(t *testing.T) {
	args := []string{
		"--override.glamsterdam", "1700000000",
		"--override.hogota", "1800000000",
		"--override.iplus", "1900000000",
	}
	cfg, exit, _ := parseFlags(args)
	if exit {
		t.Fatal("unexpected exit")
	}
	if cfg.GlamsterdamOverride == nil || *cfg.GlamsterdamOverride != 1700000000 {
		t.Errorf("GlamsterdamOverride = %v, want 1700000000", cfg.GlamsterdamOverride)
	}
	if cfg.HogotaOverride == nil || *cfg.HogotaOverride != 1800000000 {
		t.Errorf("HogotaOverride = %v, want 1800000000", cfg.HogotaOverride)
	}
	if cfg.IPlusOverride == nil || *cfg.IPlusOverride != 1900000000 {
		t.Errorf("IPlusOverride = %v, want 1900000000", cfg.IPlusOverride)
	}
}

func TestParseFlags_ForkOverrideZero(t *testing.T) {
	// --override.iplus=0 should activate at genesis (nil != 0).
	args := []string{"--override.iplus=0"}
	cfg, exit, _ := parseFlags(args)
	if exit {
		t.Fatal("unexpected exit")
	}
	if cfg.IPlusOverride == nil {
		t.Fatal("IPlusOverride should not be nil for --override.iplus=0")
	}
	if *cfg.IPlusOverride != 0 {
		t.Errorf("IPlusOverride = %d, want 0", *cfg.IPlusOverride)
	}
}

func TestParseFlags_HTTPModules(t *testing.T) {
	args := []string{"--http.api", "eth,net,web3"}
	cfg, exit, _ := parseFlags(args)
	if exit {
		t.Fatal("unexpected exit")
	}
	want := []string{"eth", "net", "web3"}
	if len(cfg.HTTPModules) != len(want) {
		t.Fatalf("HTTPModules = %v, want %v", cfg.HTTPModules, want)
	}
	for i, m := range want {
		if cfg.HTTPModules[i] != m {
			t.Errorf("HTTPModules[%d] = %q, want %q", i, cfg.HTTPModules[i], m)
		}
	}
}

func TestParseFlags_WebSocket(t *testing.T) {
	args := []string{
		"--ws",
		"--ws.addr", "0.0.0.0",
		"--ws.port", "8547",
		"--ws.api", "eth,net",
		"--ws.origins", "*",
	}
	cfg, exit, _ := parseFlags(args)
	if exit {
		t.Fatal("unexpected exit")
	}
	if !cfg.WSEnabled {
		t.Error("WSEnabled should be true")
	}
	if cfg.WSAddr != "0.0.0.0" {
		t.Errorf("WSAddr = %q, want 0.0.0.0", cfg.WSAddr)
	}
	if cfg.WSPort != 8547 {
		t.Errorf("WSPort = %d, want 8547", cfg.WSPort)
	}
}

func TestParseFlags_DoubleDash(t *testing.T) {
	args := []string{
		"--port", "30305",
		"--syncmode", "full",
		"--metrics",
	}
	cfg, exit, _ := parseFlags(args)
	if exit {
		t.Fatal("unexpected exit")
	}
	if cfg.P2PPort != 30305 {
		t.Errorf("P2PPort = %d, want 30305", cfg.P2PPort)
	}
	if cfg.SyncMode != "full" {
		t.Errorf("SyncMode = %q, want full", cfg.SyncMode)
	}
	if !cfg.Metrics {
		t.Error("Metrics should be true with --metrics")
	}
}

func TestParseFlags_Version(t *testing.T) {
	_, exit, code := parseFlags([]string{"-version"})
	if !exit {
		t.Fatal("expected exit for -version")
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
}

func TestParseFlags_InvalidFlag(t *testing.T) {
	_, exit, code := parseFlags([]string{"-unknown-flag"})
	if !exit {
		t.Fatal("expected exit for unknown flag")
	}
	if code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
}

func TestParseFlags_InvalidNetworkID(t *testing.T) {
	_, exit, code := parseFlags([]string{"-networkid", "notanumber"})
	if !exit {
		t.Fatal("expected exit for invalid networkid")
	}
	if code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
}

func TestParseFlags_PartialOverride(t *testing.T) {
	cfg, exit, _ := parseFlags([]string{"-maxpeers", "100"})
	if exit {
		t.Fatal("unexpected exit")
	}
	if cfg.MaxPeers != 100 {
		t.Errorf("MaxPeers = %d, want 100", cfg.MaxPeers)
	}
	if cfg.P2PPort != 30303 {
		t.Errorf("P2PPort = %d, want 30303", cfg.P2PPort)
	}
	if cfg.SyncMode != "snap" {
		t.Errorf("SyncMode = %q, want snap", cfg.SyncMode)
	}
}

func TestVerbosityMapping(t *testing.T) {
	tests := []struct {
		verbosity int
		wantLevel string
	}{
		{0, "error"},
		{1, "error"},
		{2, "warn"},
		{3, "info"},
		{4, "debug"},
		{5, "debug"},
	}
	for _, tt := range tests {
		got := node.VerbosityToLogLevel(tt.verbosity)
		if got != tt.wantLevel {
			t.Errorf("VerbosityToLogLevel(%d) = %q, want %q", tt.verbosity, got, tt.wantLevel)
		}
	}
}

func TestInitDataDir(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "eth2030-test")

	cfg := node.DefaultConfig()
	cfg.DataDir = dir

	if err := cfg.InitDataDir(); err != nil {
		t.Fatalf("InitDataDir() error: %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("datadir not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("datadir is not a directory")
	}

	for _, sub := range []string{"chaindata", "keystore", "nodes"} {
		subpath := filepath.Join(dir, sub)
		info, err := os.Stat(subpath)
		if err != nil {
			t.Errorf("subdir %q not created: %v", sub, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("subdir %q is not a directory", sub)
		}
	}
}

func TestInitDataDir_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "eth2030-test")

	cfg := node.DefaultConfig()
	cfg.DataDir = dir

	if err := cfg.InitDataDir(); err != nil {
		t.Fatalf("first InitDataDir() error: %v", err)
	}
	if err := cfg.InitDataDir(); err != nil {
		t.Fatalf("second InitDataDir() error: %v", err)
	}
}

func TestInitDataDir_EmptyPath(t *testing.T) {
	cfg := node.DefaultConfig()
	cfg.DataDir = ""
	if err := cfg.InitDataDir(); err == nil {
		t.Fatal("expected error for empty datadir")
	}
}

func TestListenAddrs(t *testing.T) {
	cfg := node.DefaultConfig()
	cfg.HTTPAddr = "0.0.0.0"
	cfg.RPCPort = 8545
	if got := cfg.HTTPListenAddr(); got != "0.0.0.0:8545" {
		t.Errorf("HTTPListenAddr = %q, want 0.0.0.0:8545", got)
	}

	cfg.AuthAddr = "0.0.0.0"
	cfg.EnginePort = 8551
	if got := cfg.AuthListenAddr(); got != "0.0.0.0:8551" {
		t.Errorf("AuthListenAddr = %q, want 0.0.0.0:8551", got)
	}

	cfg.MetricsAddr = "127.0.0.1"
	cfg.MetricsPort = 9001
	if got := cfg.MetricsListenAddr(); got != "127.0.0.1:9001" {
		t.Errorf("MetricsListenAddr = %q, want 127.0.0.1:9001", got)
	}
}

func TestJWTSecretPath_Default(t *testing.T) {
	cfg := node.DefaultConfig()
	cfg.DataDir = "/tmp/testdir"
	cfg.JWTSecret = ""
	got := cfg.JWTSecretPath()
	want := "/tmp/testdir/jwtsecret"
	if got != want {
		t.Errorf("JWTSecretPath = %q, want %q", got, want)
	}
}

func TestJWTSecretPath_Explicit(t *testing.T) {
	cfg := node.DefaultConfig()
	cfg.JWTSecret = "/custom/jwt.hex"
	got := cfg.JWTSecretPath()
	if got != "/custom/jwt.hex" {
		t.Errorf("JWTSecretPath = %q, want /custom/jwt.hex", got)
	}
}
