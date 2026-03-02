package node

import (
	"testing"
)

func TestNodeP2PBootnodes(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.DataDir = dir
	cfg.P2PPort = 0
	cfg.RPCPort = 0
	cfg.EnginePort = 0
	cfg.Bootnodes = "enode://abc@1.2.3.4:30303,enode://def@5.6.7.8:30303"

	n, err := New(&cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if n.Config().Bootnodes != cfg.Bootnodes {
		t.Errorf("Bootnodes not preserved in config")
	}
}

func TestNodeP2PDiscoveryPort(t *testing.T) {
	cfg := DefaultConfig()
	cfg.P2PPort = 30303
	cfg.DiscoveryPort = 0
	if cfg.EffectiveDiscoveryPort() != 30303 {
		t.Errorf("EffectiveDiscoveryPort with 0 = %d, want 30303", cfg.EffectiveDiscoveryPort())
	}

	cfg.DiscoveryPort = 30304
	if cfg.EffectiveDiscoveryPort() != 30304 {
		t.Errorf("EffectiveDiscoveryPort with 30304 = %d, want 30304", cfg.EffectiveDiscoveryPort())
	}
}

func TestNodeP2PNAT(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.DataDir = dir
	cfg.P2PPort = 0
	cfg.RPCPort = 0
	cfg.EnginePort = 0
	cfg.NAT = "extip:1.2.3.4"

	n, err := New(&cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if n.Config().NAT != "extip:1.2.3.4" {
		t.Errorf("NAT not preserved in config")
	}
}
