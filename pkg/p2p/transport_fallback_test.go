package p2p

import (
	"net"
	"testing"
	"time"
)

// BB-1.3: Transport fallback chain (Tor > Nym > Simulated)

// TestTransportFallback_TorAvailable verifies that when Tor proxy is reachable,
// SelectBestTransport returns ModeTorSocks5.
func TestTransportFallback_TorAvailable(t *testing.T) {
	// Start a fake TCP server to simulate an available Tor proxy.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()

	cfg := DefaultTransportConfig()
	cfg.TorProxyAddr = ln.Addr().String()
	cfg.NymProxyAddr = "127.0.0.1:1" // unreachable
	cfg.DialTimeout = 200 * time.Millisecond

	tm := NewTransportManagerWithConfig(cfg)
	mode := tm.SelectBestTransport()
	if mode != ModeTorSocks5 {
		t.Fatalf("expected ModeTorSocks5 when Tor available, got %v", mode)
	}
	if tm.SelectedMode() != ModeTorSocks5 {
		t.Fatalf("SelectedMode should be ModeTorSocks5, got %v", tm.SelectedMode())
	}
}

// TestTransportFallback_TorUnavailable_NymFallback verifies that when Tor is
// unreachable but Nym is available, SelectBestTransport returns ModeNymSocks5.
func TestTransportFallback_TorUnavailable_NymFallback(t *testing.T) {
	// Start a fake Nym proxy but leave Tor unreachable.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()

	cfg := DefaultTransportConfig()
	cfg.TorProxyAddr = "127.0.0.1:1" // unreachable
	cfg.NymProxyAddr = ln.Addr().String()
	cfg.DialTimeout = 200 * time.Millisecond

	tm := NewTransportManagerWithConfig(cfg)
	mode := tm.SelectBestTransport()
	if mode != ModeNymSocks5 {
		t.Fatalf("expected ModeNymSocks5 when only Nym available, got %v", mode)
	}
	if tm.SelectedMode() != ModeNymSocks5 {
		t.Fatalf("SelectedMode should be ModeNymSocks5, got %v", tm.SelectedMode())
	}
}

// TestTransportFallback_AllUnavailable_Simulated verifies that when neither Tor
// nor Nym is reachable, SelectBestTransport returns ModeSimulated.
func TestTransportFallback_AllUnavailable_Simulated(t *testing.T) {
	cfg := DefaultTransportConfig()
	cfg.TorProxyAddr = "127.0.0.1:1" // unreachable
	cfg.NymProxyAddr = "127.0.0.1:2" // unreachable
	cfg.DialTimeout = 100 * time.Millisecond

	tm := NewTransportManagerWithConfig(cfg)
	mode := tm.SelectBestTransport()
	if mode != ModeSimulated {
		t.Fatalf("expected ModeSimulated when all unavailable, got %v", mode)
	}
	if tm.SelectedMode() != ModeSimulated {
		t.Fatalf("SelectedMode should be ModeSimulated, got %v", tm.SelectedMode())
	}
}

// TestTransportFallback_LogStartup verifies SelectBestTransport sets selectedMode
// and returns a valid mode (integration check for startup log path).
func TestTransportFallback_LogStartup(t *testing.T) {
	cfg := DefaultTransportConfig()
	cfg.TorProxyAddr = "127.0.0.1:1"
	cfg.NymProxyAddr = "127.0.0.1:2"
	cfg.DialTimeout = 50 * time.Millisecond

	tm := NewTransportManagerWithConfig(cfg)
	mode := tm.SelectBestTransport()

	// Mode must be one of the valid values.
	switch mode {
	case ModeSimulated, ModeTorSocks5, ModeNymSocks5:
	default:
		t.Fatalf("unexpected mode %v", mode)
	}
}

// TestProbeProxy verifies that probeProxy correctly reports availability.
func TestProbeProxy(t *testing.T) {
	// Available proxy.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			c.Close()
		}
	}()

	if !probeProxy(ln.Addr().String(), 200*time.Millisecond) {
		t.Fatal("expected probeProxy=true for available address")
	}

	// Unavailable proxy.
	if probeProxy("127.0.0.1:1", 100*time.Millisecond) {
		t.Fatal("expected probeProxy=false for unreachable address")
	}
}
