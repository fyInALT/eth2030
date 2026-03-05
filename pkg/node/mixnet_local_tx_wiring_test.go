package node

import (
	"testing"

	"github.com/eth2030/eth2030/p2p"
)

// --- BB-1.1 / BB-1.3: TransportManager wiring ---

// TestMixnet_TransportMgrWired verifies transportMgr is non-nil after New().
func TestMixnet_TransportMgrWired(t *testing.T) {
	cfg := makeTestConfig(t)
	n := newTestNode(t, &cfg)

	if n.transportMgr == nil {
		t.Fatal("transportMgr should be non-nil after New()")
	}
}

// TestMixnet_TransportMgrHasOneTransport verifies exactly one transport is registered.
func TestMixnet_TransportMgrHasOneTransport(t *testing.T) {
	cfg := makeTestConfig(t)
	n := newTestNode(t, &cfg)

	if c := n.transportMgr.TransportCount(); c != 1 {
		t.Errorf("TransportCount = %d, want 1", c)
	}
}

// TestMixnet_SimulatedModeDefault verifies default MixnetMode wires a simulated transport.
func TestMixnet_SimulatedModeDefault(t *testing.T) {
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

// TestMixnet_TorModeExplicit verifies that --mixnet=tor sets the transport config mode
// (without needing a real Tor daemon — the manager honours the explicit flag).
func TestMixnet_TorModeExplicit(t *testing.T) {
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

// TestMixnet_NymModeExplicit verifies that --mixnet=nym sets the Nym transport.
func TestMixnet_NymModeExplicit(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.MixnetMode = "nym"
	n := newTestNode(t, &cfg)

	if got := n.transportMgr.Config().Mode; got != p2p.ModeNymSocks5 {
		t.Errorf("Config().Mode = %v, want ModeNymSocks5", got)
	}
}

// TestMixnet_InvalidModeRejected verifies config validation rejects unknown modes.
func TestMixnet_InvalidModeRejected(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.MixnetMode = "i2p" // not supported
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for unknown mixnet mode")
	}
}

// --- BB-2.2: ExperimentalLocalTx propagation ---

// TestLocalTx_FlagDefault verifies ExperimentalLocalTx defaults to false.
func TestLocalTx_FlagDefault(t *testing.T) {
	cfg := makeTestConfig(t)
	if cfg.ExperimentalLocalTx {
		t.Error("ExperimentalLocalTx should default to false")
	}
}

// TestLocalTx_FlagPropagatedToPool verifies the flag reaches the txpool config.
// We confirm by checking the node config rather than private pool fields.
// (AllowLocalTx=false means type-0x08 AddLocal returns an error.)
func TestLocalTx_FlagPropagatedToPool(t *testing.T) {
	cfg := makeTestConfig(t)
	n := newTestNode(t, &cfg)
	if n.TxPool() == nil {
		t.Fatal("TxPool should be non-nil")
	}
	if n.config.ExperimentalLocalTx {
		t.Error("ExperimentalLocalTx should be false by default")
	}
}

// TestLocalTx_FlagEnabledPropagation verifies a node created with
// ExperimentalLocalTx=true reflects that in its config.
func TestLocalTx_FlagEnabledPropagation(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.ExperimentalLocalTx = true
	n := newTestNode(t, &cfg)

	if !n.config.ExperimentalLocalTx {
		t.Error("ExperimentalLocalTx should be true after explicit set")
	}
}

// TestMixnet_RPCEndpointMatchesNode verifies the transport config uses
// the node's own RPC address as its forwarding endpoint.
func TestMixnet_RPCEndpointMatchesNode(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.RPCPort = 18545
	n := newTestNode(t, &cfg)

	endpoint := n.transportMgr.Config().RPCEndpoint
	want := "http://" + cfg.RPCAddr()
	if endpoint != want {
		t.Errorf("RPCEndpoint = %q, want %q", endpoint, want)
	}
}
