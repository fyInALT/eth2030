package transport

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/eth2030/eth2030/core/types"
)

// jsonRPCRequest is the wire format for an Ethereum JSON-RPC call.
type jsonRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// sendViaSOCKS5 sends payload as an eth_sendRawTransaction JSON-RPC HTTP POST
// to endpoint, tunneled through proxyAddr via SOCKS5.
// payload must be the RLP-encoded signed transaction bytes.
func sendViaSOCKS5(proxyAddr string, dialTimeout time.Duration, payload []byte, endpoint string) error {
	body, err := json.Marshal(jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_sendRawTransaction",
		Params:  []interface{}{"0x" + hex.EncodeToString(payload)},
		ID:      1,
	})
	if err != nil {
		return fmt.Errorf("marshal rpc request: %w", err)
	}

	// HTTP client with a custom dialer that tunnels via SOCKS5.
	client := &http.Client{
		Timeout: dialTimeout * 20,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				h, p, splitErr := net.SplitHostPort(addr)
				if splitErr != nil {
					return nil, splitErr
				}
				portNum, convErr := strconv.ParseUint(p, 10, 16)
				if convErr != nil {
					return nil, convErr
				}
				return socks5Dial(proxyAddr, h, uint16(portNum), dialTimeout)
			},
		},
	}

	resp, err := client.Post(endpoint, "application/json", bytes.NewReader(body)) //nolint:noctx
	if err != nil {
		return fmt.Errorf("http post via socks5: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

// TorConfig configures the Tor SOCKS5 anonymous transport.
type TorConfig struct {
	// ProxyAddr is the Tor SOCKS5 proxy address (default: 127.0.0.1:9050).
	ProxyAddr string
	// RPCEndpoint is the node's JSON-RPC endpoint for tx submission via Tor.
	RPCEndpoint string
	// DialTimeout is the connection timeout for the SOCKS5 proxy dial.
	DialTimeout time.Duration
	// MaxPending is the channel buffer size for inbound txs.
	MaxPending int
}

// DefaultTorConfig returns sensible defaults for Tor SOCKS5 transport.
func DefaultTorConfig() *TorConfig {
	return &TorConfig{
		ProxyAddr:   "127.0.0.1:9050",
		RPCEndpoint: "http://127.0.0.1:8545",
		DialTimeout: 500 * time.Millisecond,
		MaxPending:  256,
	}
}

// TorTransport implements AnonymousTransport and ExternalMixnetTransport by
// routing transactions through a Tor SOCKS5 proxy, obscuring the sender IP.
type TorTransport struct {
	config  *TorConfig
	ch      chan *types.Transaction
	running bool
	mu      sync.Mutex
	stopCh  chan struct{}
}

// NewTorTransport creates a Tor SOCKS5 transport with the given config.
// If cfg is nil, DefaultTorConfig is used.
func NewTorTransport(cfg *TorConfig) *TorTransport {
	if cfg == nil {
		cfg = DefaultTorConfig()
	}
	return &TorTransport{
		config: cfg,
		ch:     make(chan *types.Transaction, cfg.MaxPending),
		stopCh: make(chan struct{}),
	}
}

// Name returns "tor".
func (t *TorTransport) Name() string { return "tor" }

// Submit sends a transaction via Tor SOCKS5 to the node's own RPC endpoint.
// The tx hash bytes are used as the payload (full RLP encoding in production).
func (t *TorTransport) Submit(tx *types.Transaction) error {
	if tx == nil {
		return ErrAnonTransportNilTx
	}
	t.mu.Lock()
	if !t.running {
		t.mu.Unlock()
		return ErrAnonTransportClosed
	}
	t.mu.Unlock()

	// Use tx hash bytes as the serialized payload.
	return t.SendViaExternalMixnet(tx.Hash().Bytes(), t.config.RPCEndpoint)
}

// SendViaExternalMixnet sends payload as an eth_sendRawTransaction JSON-RPC HTTP POST
// to endpoint, tunneled through the Tor SOCKS5 proxy. endpoint must be an HTTP URL.
func (t *TorTransport) SendViaExternalMixnet(payload []byte, endpoint string) error {
	if err := sendViaSOCKS5(t.config.ProxyAddr, t.config.DialTimeout, payload, endpoint); err != nil {
		return fmt.Errorf("tor: %w", err)
	}
	return nil
}

// Receive returns the channel of transactions received via this transport.
// For Tor (outbound-only), this channel is always empty.
func (t *TorTransport) Receive() <-chan *types.Transaction { return t.ch }

// Start activates the Tor transport.
func (t *TorTransport) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.running = true
	return nil
}

// Stop shuts down the Tor transport.
func (t *TorTransport) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.running {
		t.running = false
		close(t.stopCh)
	}
	return nil
}

// --- NymTransport ---

// NymConfig configures the Nym SOCKS5 anonymous transport.
// NymTransport uses the same SOCKS5 protocol as TorTransport but targets
// a Nym client proxy (default: 127.0.0.1:1080).
type NymConfig struct {
	ProxyAddr   string
	RPCEndpoint string
	DialTimeout time.Duration
	MaxPending  int
}

// DefaultNymConfig returns sensible defaults for Nym SOCKS5 transport.
func DefaultNymConfig() *NymConfig {
	return &NymConfig{
		ProxyAddr:   "127.0.0.1:1080",
		RPCEndpoint: "http://127.0.0.1:8545",
		DialTimeout: 500 * time.Millisecond,
		MaxPending:  256,
	}
}

// NymTransport implements AnonymousTransport and ExternalMixnetTransport via Nym SOCKS5.
type NymTransport struct {
	config  *NymConfig
	ch      chan *types.Transaction
	running bool
	mu      sync.Mutex
	stopCh  chan struct{}
}

// NewNymTransport creates a Nym SOCKS5 transport. If cfg is nil, DefaultNymConfig is used.
func NewNymTransport(cfg *NymConfig) *NymTransport {
	if cfg == nil {
		cfg = DefaultNymConfig()
	}
	return &NymTransport{
		config: cfg,
		ch:     make(chan *types.Transaction, cfg.MaxPending),
		stopCh: make(chan struct{}),
	}
}

// Name returns "nym".
func (n *NymTransport) Name() string { return "nym" }

// Submit sends a transaction via Nym SOCKS5 to the node's own RPC endpoint.
func (n *NymTransport) Submit(tx *types.Transaction) error {
	if tx == nil {
		return ErrAnonTransportNilTx
	}
	n.mu.Lock()
	if !n.running {
		n.mu.Unlock()
		return ErrAnonTransportClosed
	}
	n.mu.Unlock()
	return n.SendViaExternalMixnet(tx.Hash().Bytes(), n.config.RPCEndpoint)
}

// SendViaExternalMixnet sends payload as an eth_sendRawTransaction JSON-RPC HTTP POST
// to endpoint, tunneled through the Nym SOCKS5 proxy. endpoint must be an HTTP URL.
func (n *NymTransport) SendViaExternalMixnet(payload []byte, endpoint string) error {
	if err := sendViaSOCKS5(n.config.ProxyAddr, n.config.DialTimeout, payload, endpoint); err != nil {
		return fmt.Errorf("nym: %w", err)
	}
	return nil
}

// Receive returns the inbound tx channel (always empty for outbound-only Nym).
func (n *NymTransport) Receive() <-chan *types.Transaction { return n.ch }

// Start activates the Nym transport.
func (n *NymTransport) Start() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.running = true
	return nil
}

// Stop shuts down the Nym transport.
func (n *NymTransport) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.running {
		n.running = false
		close(n.stopCh)
	}
	return nil
}

// --- Shared SOCKS5 helpers ---

// socks5Dial establishes a SOCKS5 CONNECT tunnel to host:port through proxyAddr.
// It performs the full SOCKS5 handshake (no-auth, domain-name CONNECT).
func socks5Dial(proxyAddr, host string, port uint16, timeout time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", proxyAddr, timeout)
	if err != nil {
		return nil, fmt.Errorf("connect proxy %s: %w", proxyAddr, err)
	}

	// SOCKS5 greeting: version=5, nMethods=1, method=no-auth(0x00).
	if _, err := conn.Write([]byte{0x05, 0x01, 0x00}); err != nil {
		conn.Close()
		return nil, fmt.Errorf("greeting write: %w", err)
	}
	resp := make([]byte, 2)
	if _, err := io.ReadFull(conn, resp); err != nil {
		conn.Close()
		return nil, fmt.Errorf("greeting read: %w", err)
	}
	if resp[0] != 0x05 || resp[1] != 0x00 {
		conn.Close()
		return nil, fmt.Errorf("auth rejected: method 0x%02x", resp[1])
	}

	// SOCKS5 CONNECT request with domain-name ATYP (0x03).
	hostBytes := []byte(host)
	req := make([]byte, 7+len(hostBytes))
	req[0] = 0x05 // version
	req[1] = 0x01 // CMD: CONNECT
	req[2] = 0x00 // RSV
	req[3] = 0x03 // ATYP: domain
	req[4] = byte(len(hostBytes))
	copy(req[5:], hostBytes)
	req[5+len(hostBytes)] = byte(port >> 8)
	req[6+len(hostBytes)] = byte(port)

	if _, err := conn.Write(req); err != nil {
		conn.Close()
		return nil, fmt.Errorf("connect request: %w", err)
	}

	// Read SOCKS5 response (at least 10 bytes for IPv4 bound address).
	rsp := make([]byte, 10)
	if _, err := io.ReadFull(conn, rsp); err != nil {
		conn.Close()
		return nil, fmt.Errorf("connect response: %w", err)
	}
	if rsp[1] != 0x00 {
		conn.Close()
		return nil, fmt.Errorf("connect failed: code %d", rsp[1])
	}
	return conn, nil
}

// parseHTTPEndpoint parses an HTTP URL into host and port number.
// For URLs without an explicit port, defaults to 80 (http) or 443 (https).
func parseHTTPEndpoint(endpoint string) (host string, port uint16, err error) {
	u, parseErr := url.Parse(endpoint)
	if parseErr != nil {
		return "", 0, fmt.Errorf("parse endpoint %q: %w", endpoint, parseErr)
	}

	h, p, splitErr := net.SplitHostPort(u.Host)
	if splitErr != nil {
		// No explicit port — use scheme default.
		h = u.Host
		switch u.Scheme {
		case "https":
			p = "443"
		default:
			p = "80"
		}
	}
	if h == "" {
		return "", 0, fmt.Errorf("empty host in endpoint %q", endpoint)
	}
	portNum, convErr := strconv.ParseUint(p, 10, 16)
	if convErr != nil {
		return "", 0, fmt.Errorf("invalid port %q: %w", p, convErr)
	}
	return h, uint16(portNum), nil
}
