package node

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestMetricsServer_StartsAndResponds(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.DataDir = dir
	cfg.P2PPort = 0
	cfg.RPCPort = 0
	cfg.EnginePort = 0
	cfg.Metrics = true
	cfg.MetricsAddr = "127.0.0.1"
	cfg.MetricsPort = 0 // ephemeral — we'll probe directly

	// Use a fixed port in the high range to avoid conflicts.
	cfg.MetricsPort = freePort(t)

	n, err := New(&cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if err := n.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer func() { _ = n.Stop() }()

	// Give the server a moment to bind.
	time.Sleep(50 * time.Millisecond)

	url := fmt.Sprintf("http://%s/debug/vars", cfg.MetricsListenAddr())
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		t.Fatalf("GET %s error: %v", url, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("metrics status = %d, want 200", resp.StatusCode)
	}
}

func TestMetricsServer_DisabledByDefault(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.DataDir = dir
	cfg.P2PPort = 0
	cfg.RPCPort = 0
	cfg.EnginePort = 0
	// Metrics disabled by default.

	n, err := New(&cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if err := n.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer func() { _ = n.Stop() }()

	if n.metricsServer != nil {
		t.Error("metricsServer should be nil when metrics disabled")
	}
}

// freePort returns a deterministic test port derived from the test name.
func freePort(t *testing.T) int {
	t.Helper()
	h := 0
	for _, c := range t.Name() {
		h = h*31 + int(c)
	}
	if h < 0 {
		h = -h
	}
	return 40000 + (h % 10000)
}
