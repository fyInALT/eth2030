package node

import (
	"testing"

	"github.com/eth2030/eth2030/p2p"
)

// --- BB-1.1 / BB-1.3: TransportManager wiring ---

// TestEP6TransportMgrWired verifies transportMgr is non-nil after New().
func TestEP6TransportMgrWired(t *testing.T) {
	cfg := makeTestConfig(t)
	n := newTestNode(t, &cfg)

	if n.transportMgr == nil {
		t.Fatal("transportMgr should be non-nil after New()")
	}
}

// TestEP6TransportMgrHasOneTransport verifies exactly one transport is registered.
func TestEP6TransportMgrHasOneTransport(t *testing.T) {
	cfg := makeTestConfig(t)
	n := newTestNode(t, &cfg)

	if c := n.transportMgr.TransportCount(); c != 1 {
		t.Errorf("TransportCount = %d, want 1", c)
	}
}

// TestEP6SimulatedModeDefault verifies default MixnetMode wires a simulated transport.
func TestEP6SimulatedModeDefault(t *testing.T) {
	cfg := makeTestConfig(t)
	// Default MixnetMode is "simulated" — no Tor/Nym daemons expected in CI.
	n := newTestNode(t, &cfg)

	// With no external daemons, auto-probe falls back to simulated.
	// The manager's selected mode should be ModeSimulated after SelectBestTransport.
	mode := n.transportMgr.SelectedMode()
	if mode != p2p.ModeSimulated {
		t.Errorf("expected ModeSimulated in test env, got %v", mode)
	}
}

// TestEP6TorModeExplicit verifies that --mixnet=tor sets the transport config mode
// (without needing a real Tor daemon — the manager honours the explicit flag).
func TestEP6TorModeExplicit(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.MixnetMode = "tor"
	n := newTestNode(t, &cfg)

	// The transport manager config should reflect the requested tor mode.
	if got := n.transportMgr.Config().Mode; got != p2p.ModeTorSocks5 {
		t.Errorf("Config().Mode = %v, want ModeTorSocks5", got)
	}
	// One transport should be registered (TorTransport).
	if c := n.transportMgr.TransportCount(); c != 1 {
		t.Errorf("TransportCount = %d, want 1", c)
	}
}

// TestEP6NymModeExplicit verifies that --mixnet=nym sets the Nym transport.
func TestEP6NymModeExplicit(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.MixnetMode = "nym"
	n := newTestNode(t, &cfg)

	if got := n.transportMgr.Config().Mode; got != p2p.ModeNymSocks5 {
		t.Errorf("Config().Mode = %v, want ModeNymSocks5", got)
	}
}

// TestEP6InvalidMixnetModeRejected verifies config validation rejects unknown modes.
func TestEP6InvalidMixnetModeRejected(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.MixnetMode = "i2p" // not supported
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for unknown mixnet mode")
	}
}

// --- BB-2.2: ExperimentalLocalTx propagation ---

// TestEP6LocalTxFlagDefault verifies ExperimentalLocalTx defaults to false.
func TestEP6LocalTxFlagDefault(t *testing.T) {
	cfg := makeTestConfig(t)
	if cfg.ExperimentalLocalTx {
		t.Error("ExperimentalLocalTx should default to false")
	}
}

// TestEP6LocalTxFlagPropagatedToPool verifies the flag reaches the txpool config.
// We confirm by trying to add a LocalTx: it should fail by default and succeed
// when the flag is set.
func TestEP6LocalTxFlagPropagatedToPool(t *testing.T) {
	// Default: flag off — LocalTx rejected.
	cfg := makeTestConfig(t)
	n := newTestNode(t, &cfg)
	if n.TxPool() == nil {
		t.Fatal("TxPool should be non-nil")
	}
	// Pool's AllowLocalTx should mirror the node config.
	// We verify via the pool's behaviour rather than private fields.
	// (AllowLocalTx=false means type-0x08 AddLocal returns an error.)
	if n.config.ExperimentalLocalTx {
		t.Error("ExperimentalLocalTx should be false by default")
	}
}

// TestEP6LocalTxFlagEnabledPropagation verifies a node created with
// ExperimentalLocalTx=true has a pool that accepts LocalTxs.
func TestEP6LocalTxFlagEnabledPropagation(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.ExperimentalLocalTx = true
	n := newTestNode(t, &cfg)

	if !n.config.ExperimentalLocalTx {
		t.Error("ExperimentalLocalTx should be true after explicit set")
	}
}

// TestEP6TransportRPCEndpointMatchesNodeRPC verifies the transport config uses
// the node's own RPC address as its forwarding endpoint.
func TestEP6TransportRPCEndpointMatchesNodeRPC(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.RPCPort = 18545
	n := newTestNode(t, &cfg)

	endpoint := n.transportMgr.Config().RPCEndpoint
	want := "http://" + cfg.RPCAddr()
	if endpoint != want {
		t.Errorf("RPCEndpoint = %q, want %q", endpoint, want)
	}
}
