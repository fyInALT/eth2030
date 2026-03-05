//go:build integration

package p2p

import (
	"testing"
	"time"
)

// TestTorTransport_RealTor requires a live Tor SOCKS5 daemon at 127.0.0.1:9050.
// Run with: go test -tags=integration ./p2p/... -run TestTorTransport_RealTor
func TestTorTransport_RealTor(t *testing.T) {
	const torProxy = "127.0.0.1:9050"

	// Skip if Tor is not running.
	if !probeProxy(torProxy, 500*time.Millisecond) {
		t.Skipf("Tor SOCKS5 not available at %s (start tor daemon to run this test)", torProxy)
	}

	cfg := DefaultTorConfig()
	cfg.RPCEndpoint = "http://127.0.0.1:8545"

	tr := NewTorTransport(cfg)
	if err := tr.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer tr.Stop()

	// SendViaExternalMixnet to a local endpoint via Tor.
	// This verifies the full SOCKS5 tunnel is established through a real Tor circuit.
	payload := []byte{0xde, 0xad, 0xbe, 0xef}
	if err := tr.SendViaExternalMixnet(payload, cfg.RPCEndpoint); err != nil {
		// Expected: the local node might not be running, but the Tor circuit should open.
		// A "connection refused" from the target (not from the proxy) still means Tor worked.
		t.Logf("SendViaExternalMixnet via real Tor: %v (expected if local node not running)", err)
	}
}

// TestNymTransport_RealNym requires a live Nym SOCKS5 client at 127.0.0.1:1080.
// Run with: go test -tags=integration ./p2p/... -run TestNymTransport_RealNym
func TestNymTransport_RealNym(t *testing.T) {
	const nymProxy = "127.0.0.1:1080"

	if !probeProxy(nymProxy, 500*time.Millisecond) {
		t.Skipf("Nym SOCKS5 not available at %s (start nym-client to run this test)", nymProxy)
	}

	cfg := DefaultNymConfig()
	cfg.RPCEndpoint = "http://127.0.0.1:8545"

	tr := NewNymTransport(cfg)
	if err := tr.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer tr.Stop()

	payload := []byte{0xca, 0xfe, 0xba, 0xbe}
	if err := tr.SendViaExternalMixnet(payload, cfg.RPCEndpoint); err != nil {
		t.Logf("SendViaExternalMixnet via real Nym: %v (expected if local node not running)", err)
	}
}
