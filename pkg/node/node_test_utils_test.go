package node

import "testing"

func makeTestConfig(t *testing.T) Config {
	t.Helper()
	cfg := DefaultConfig()
	cfg.DataDir = t.TempDir()
	cfg.P2PPort = 0
	cfg.RPCPort = 0
	cfg.EnginePort = 0
	return cfg
}

func newTestNode(t *testing.T, cfg *Config) *Node {
	t.Helper()
	n, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	t.Cleanup(func() {
		if err := n.Stop(); err != nil {
			t.Errorf("Stop() error: %v", err)
		}
	})
	return n
}
