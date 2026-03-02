package node

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWSServer_StartsAndAcceptsConnection(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.DataDir = dir
	cfg.P2PPort = 0
	cfg.RPCPort = 0
	cfg.EnginePort = 0
	cfg.WSEnabled = true
	cfg.WSAddr = "127.0.0.1"
	cfg.WSPort = freePortWS(t)
	cfg.WSOrigins = []string{"*"}

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

	wsURL := fmt.Sprintf("ws://%s", cfg.WSListenAddr())
	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(wsURL, http.Header{"Origin": []string{"*"}})
	if err != nil {
		t.Fatalf("Dial %s error: %v (resp: %v)", wsURL, err, resp)
	}
	defer conn.Close()

	// Send a JSON-RPC request.
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_blockNumber",
		"params":  []interface{}{},
		"id":      1,
	}
	if err := conn.WriteJSON(req); err != nil {
		t.Fatalf("WriteJSON error: %v", err)
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage error: %v", err)
	}
	var response map[string]interface{}
	if err := json.Unmarshal(msg, &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response["jsonrpc"] != "2.0" {
		t.Errorf("jsonrpc = %v, want 2.0", response["jsonrpc"])
	}
}

func TestWSServer_DisabledByDefault(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultConfig()
	cfg.DataDir = dir
	cfg.P2PPort = 0
	cfg.RPCPort = 0
	cfg.EnginePort = 0
	// WSEnabled defaults to false.

	n, err := New(&cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if err := n.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer func() { _ = n.Stop() }()

	if n.wsServer != nil {
		t.Error("wsServer should be nil when WebSocket is disabled")
	}
}

// freePortWS uses the same deterministic approach as freePort but with a
// different offset to avoid collisions with metrics ports.
func freePortWS(t *testing.T) int {
	t.Helper()
	h := 0
	for _, c := range t.Name() {
		h = h*31 + int(c)
	}
	if h < 0 {
		h = -h
	}
	return 50000 + (h % 10000)
}
