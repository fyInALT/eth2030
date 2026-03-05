package p2p

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/eth2030/eth2030/core/types"
)

// --- BB-1.2: Tor SOCKS5 transport ---

func TestTorTransport_InterfaceCompliance(t *testing.T) {
	var _ AnonymousTransport = (*TorTransport)(nil)
	var _ ExternalMixnetTransport = (*TorTransport)(nil)
}

func TestTorTransport_Name(t *testing.T) {
	tr := NewTorTransport(nil)
	if tr.Name() != "tor" {
		t.Fatalf("expected 'tor', got %q", tr.Name())
	}
}

func TestTorTransport_DefaultConfig(t *testing.T) {
	cfg := DefaultTorConfig()
	if cfg.ProxyAddr != "127.0.0.1:9050" {
		t.Fatalf("expected tor proxy '127.0.0.1:9050', got %q", cfg.ProxyAddr)
	}
	if cfg.DialTimeout != 500*time.Millisecond {
		t.Fatalf("expected 500ms timeout, got %v", cfg.DialTimeout)
	}
	if cfg.MaxPending <= 0 {
		t.Fatalf("expected positive MaxPending, got %d", cfg.MaxPending)
	}
}

func TestTorTransport_NotStarted(t *testing.T) {
	tr := NewTorTransport(nil)
	tx := testTx()
	if err := tr.Submit(tx); err != ErrAnonTransportClosed {
		t.Fatalf("expected ErrAnonTransportClosed before start, got %v", err)
	}
}

func TestTorTransport_SubmitNilTx(t *testing.T) {
	tr := NewTorTransport(nil)
	_ = tr.Start()
	defer tr.Stop()
	if err := tr.Submit(nil); err != ErrAnonTransportNilTx {
		t.Fatalf("expected ErrAnonTransportNilTx, got %v", err)
	}
}

func TestTorTransport_Stop(t *testing.T) {
	tr := NewTorTransport(nil)
	_ = tr.Start()
	if err := tr.Stop(); err != nil {
		t.Fatalf("unexpected stop error: %v", err)
	}
	// Further submissions should fail.
	if err := tr.Submit(testTx()); err != ErrAnonTransportClosed {
		t.Fatalf("expected ErrAnonTransportClosed after stop, got %v", err)
	}
}

func TestTorTransport_Receive(t *testing.T) {
	tr := NewTorTransport(nil)
	ch := tr.Receive()
	if ch == nil {
		t.Fatal("expected non-nil receive channel")
	}
}

// TestTorTransport_MockSocks5 tests SOCKS5 protocol + HTTP POST JSON-RPC delivery.
func TestTorTransport_MockSocks5(t *testing.T) {
	// HTTP target: receives the forwarded eth_sendRawTransaction request.
	var httpHits int32
	httpSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&httpHits, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x0"}`)) //nolint:errcheck
	}))
	defer httpSrv.Close()

	mock := newMockSocks5Proxy(t)
	defer mock.close()

	cfg := &TorConfig{
		ProxyAddr:   mock.addr(),
		RPCEndpoint: httpSrv.URL,
		DialTimeout: 2 * time.Second,
		MaxPending:  16,
	}
	tr := NewTorTransport(cfg)
	_ = tr.Start()
	defer tr.Stop()

	payload := []byte{0x01, 0x02, 0x03}
	if err := tr.SendViaExternalMixnet(payload, httpSrv.URL); err != nil {
		t.Fatalf("SendViaExternalMixnet error: %v", err)
	}

	// Verify mock received a CONNECT to httpSrv's host.
	select {
	case target := <-mock.targets:
		if target == "" {
			t.Fatal("expected non-empty CONNECT target")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for SOCKS5 CONNECT")
	}

	// Verify HTTP server was actually called.
	if atomic.LoadInt32(&httpHits) == 0 {
		t.Fatal("HTTP server received no requests (JSON-RPC POST not delivered)")
	}
}

// TestTorTransport_JSONRPCFormat verifies the eth_sendRawTransaction JSON-RPC body.
func TestTorTransport_JSONRPCFormat(t *testing.T) {
	var capturedBody []byte
	httpSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0xabc"}`)) //nolint:errcheck
	}))
	defer httpSrv.Close()

	mock := newMockSocks5Proxy(t)
	defer mock.close()

	cfg := &TorConfig{
		ProxyAddr:   mock.addr(),
		RPCEndpoint: httpSrv.URL,
		DialTimeout: 2 * time.Second,
		MaxPending:  16,
	}
	tr := NewTorTransport(cfg)
	_ = tr.Start()
	defer tr.Stop()

	payload := []byte{0xde, 0xad, 0xbe, 0xef}
	if err := tr.SendViaExternalMixnet(payload, httpSrv.URL); err != nil {
		t.Fatalf("SendViaExternalMixnet error: %v", err)
	}

	// Wait for HTTP body capture (httpSrv is synchronous, no wait needed).
	var req map[string]interface{}
	if err := json.Unmarshal(capturedBody, &req); err != nil {
		t.Fatalf("body is not valid JSON: %v\nbody: %s", err, capturedBody)
	}
	if req["method"] != "eth_sendRawTransaction" {
		t.Fatalf("expected method 'eth_sendRawTransaction', got %v", req["method"])
	}
	if req["jsonrpc"] != "2.0" {
		t.Fatalf("expected jsonrpc '2.0', got %v", req["jsonrpc"])
	}
	params, ok := req["params"].([]interface{})
	if !ok || len(params) == 0 {
		t.Fatalf("expected params array, got %v", req["params"])
	}
	if params[0] != "0x"+fmt.Sprintf("%x", payload) {
		t.Fatalf("expected hex payload '0x%x', got %v", payload, params[0])
	}
}

func TestSocks5Dial_InvalidProxy(t *testing.T) {
	// Port 1 is reserved and should refuse connections quickly.
	_, err := socks5Dial("127.0.0.1:1", "example.com", 80, 200*time.Millisecond)
	if err == nil {
		t.Fatal("expected error connecting to invalid proxy")
	}
}

func TestParseHTTPEndpoint(t *testing.T) {
	cases := []struct {
		endpoint string
		host     string
		port     uint16
		wantErr  bool
	}{
		{"http://127.0.0.1:8545", "127.0.0.1", 8545, false},
		{"http://localhost:9000", "localhost", 9000, false},
		{"https://example.com", "example.com", 443, false},
		{"http://example.com", "example.com", 80, false},
	}
	for _, tc := range cases {
		host, port, err := parseHTTPEndpoint(tc.endpoint)
		if tc.wantErr {
			if err == nil {
				t.Errorf("endpoint %q: expected error, got nil", tc.endpoint)
			}
			continue
		}
		if err != nil {
			t.Errorf("endpoint %q: unexpected error: %v", tc.endpoint, err)
			continue
		}
		if host != tc.host {
			t.Errorf("endpoint %q: host: got %q, want %q", tc.endpoint, host, tc.host)
		}
		if port != tc.port {
			t.Errorf("endpoint %q: port: got %d, want %d", tc.endpoint, port, tc.port)
		}
	}
}

// --- mockSocks5Proxy: simulates a SOCKS5 proxy for unit tests ---

type mockSocks5Proxy struct {
	ln      net.Listener
	targets chan string // receives "host:port" from each CONNECT request
	stopCh  chan struct{}
}

func newMockSocks5Proxy(t *testing.T) *mockSocks5Proxy {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("mock socks5: listen: %v", err)
	}
	m := &mockSocks5Proxy{
		ln:      ln,
		targets: make(chan string, 8),
		stopCh:  make(chan struct{}),
	}
	go m.serve()
	return m
}

func (m *mockSocks5Proxy) addr() string {
	return m.ln.Addr().String()
}

func (m *mockSocks5Proxy) close() {
	close(m.stopCh)
	m.ln.Close()
}

func (m *mockSocks5Proxy) serve() {
	for {
		conn, err := m.ln.Accept()
		if err != nil {
			select {
			case <-m.stopCh:
				return
			default:
			}
			continue
		}
		go m.handleConn(conn)
	}
}

// handleConn implements the SOCKS5 server-side protocol for testing.
func (m *mockSocks5Proxy) handleConn(conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(2 * time.Second)) //nolint:errcheck

	// Read greeting: [version, nMethods, methods...]
	hdr := make([]byte, 2)
	if _, err := io.ReadFull(conn, hdr); err != nil {
		return
	}
	methods := make([]byte, hdr[1])
	if _, err := io.ReadFull(conn, methods); err != nil {
		return
	}
	// Accept with no-auth.
	conn.Write([]byte{0x05, 0x00}) //nolint:errcheck

	// Read CONNECT request: [version, CMD, RSV, ATYP, ...]
	req := make([]byte, 4)
	if _, err := io.ReadFull(conn, req); err != nil {
		return
	}
	if req[1] != 0x01 { // only CONNECT supported
		return
	}

	var host string
	switch req[3] { // ATYP
	case 0x01: // IPv4
		addr := make([]byte, 4)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return
		}
		host = net.IP(addr).String()
	case 0x03: // domain
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			return
		}
		domain := make([]byte, lenBuf[0])
		if _, err := io.ReadFull(conn, domain); err != nil {
			return
		}
		host = string(domain)
	case 0x04: // IPv6
		addr := make([]byte, 16)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return
		}
		host = net.IP(addr).String()
	default:
		return
	}

	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBuf); err != nil {
		return
	}
	portNum := int(portBuf[0])<<8 | int(portBuf[1])

	target := fmt.Sprintf("%s:%d", host, portNum)
	select {
	case m.targets <- target:
	default:
	}

	// Try to connect to the actual target and forward traffic.
	targetConn, err := net.DialTimeout("tcp", target, 1*time.Second)
	if err != nil {
		// Target unreachable — send SOCKS5 failure (connection refused).
		conn.Write([]byte{0x05, 0x05, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //nolint:errcheck
		return
	}
	defer targetConn.Close()

	// Send SOCKS5 success before starting the bidirectional tunnel.
	conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //nolint:errcheck

	// Bidirectional tunnel: client ↔ target.
	done := make(chan struct{})
	go func() {
		io.Copy(targetConn, conn) //nolint:errcheck
		close(done)
	}()
	io.Copy(conn, targetConn) //nolint:errcheck
	<-done
}

// testLocalTxForTor builds a LocalTx for Tor transport tests.
func testLocalTxForTor() *types.Transaction {
	to := types.Address{}
	return types.NewLocalTx(
		big.NewInt(1), 0, &to, big.NewInt(0), 21000,
		big.NewInt(1), big.NewInt(1), nil, []byte{0x0a},
	)
}
