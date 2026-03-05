// Package node implements the ETH2030 full node lifecycle,
// wiring together blockchain, RPC, Engine API, P2P, and TxPool.
package node

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds all configuration for an ETH2030 node.
type Config struct {
	// DataDir is the root directory for all data storage.
	DataDir string

	// Name is a human-readable node identifier (used in logs).
	Name string

	// Network selects the Ethereum network (mainnet, sepolia, holesky).
	// Ignored when GenesisPath is set.
	Network string

	// NetworkID is the numeric network identifier.
	// 0 means derive from genesis chain ID.
	NetworkID uint64

	// SyncMode selects the sync strategy (full, snap).
	SyncMode string

	// GCMode controls state pruning: "full" (default) or "archive" (no pruning).
	GCMode string

	// P2PPort is the TCP port for devp2p connections.
	P2PPort int

	// DiscoveryPort is the UDP port for node discovery.
	// When 0, defaults to P2PPort.
	DiscoveryPort int

	// Bootnodes is a comma-separated list of enode URLs used for bootstrapping.
	Bootnodes string

	// NAT is the NAT traversal method string (e.g. "extip:1.2.3.4").
	NAT string

	// RPCAuthSecret requires this bearer token for JSON-RPC requests when set.
	RPCAuthSecret string

	// RPCRateLimitPerSec controls request rate limiting for JSON-RPC requests.
	RPCRateLimitPerSec int

	// RPCMaxRequestSize is the max body size for JSON-RPC requests, in bytes.
	RPCMaxRequestSize int64

	// RPCMaxBatchSize is the max number of requests in a single batch.
	RPCMaxBatchSize int

	// RPCCORSOrigins is a comma-separated list of allowed Origin headers.
	// Use "*" to allow all.
	RPCCORSOrigins string

	// EngineMaxRequestSize is the max body size for Engine API requests, in bytes.
	EngineMaxRequestSize int64

	// EngineAuthToken requires this bearer token for Engine API requests when set.
	EngineAuthToken string

	// EngineAuthTokenPath optionally loads the Engine API token from a file.
	EngineAuthTokenPath string

	// MaxPeers is the maximum number of P2P peers.
	MaxPeers int

	// HTTP-RPC server.
	RPCPort        int      // --http.port
	HTTPAddr       string   // --http.addr
	HTTPVhosts     []string // --http.vhosts
	HTTPCORSDomain []string // --http.corsdomain
	HTTPModules    []string // --http.api

	// Engine API (authenticated RPC).
	EnginePort int      // --authrpc.port
	AuthAddr   string   // --authrpc.addr
	AuthVhosts []string // --authrpc.vhosts
	JWTSecret  string   // --authrpc.jwtsecret  (path to file or hex string)

	// WebSocket RPC server.
	WSEnabled bool     // --ws
	WSAddr    string   // --ws.addr
	WSPort    int      // --ws.port
	WSModules []string // --ws.api
	WSOrigins []string // --ws.origins

	// Metrics server.
	Metrics     bool   // --metrics
	MetricsAddr string // --metrics.addr
	MetricsPort int    // --metrics.port

	// Genesis / fork overrides.
	GenesisPath         string  // --override.genesis (path to genesis.json)
	GlamsterdamOverride *uint64 // --override.glamsterdam
	HogotaOverride      *uint64 // --override.hogota
	IPlusOverride       *uint64 // --override.iplus

	// Misc RPC / miner settings.
	AllowUnprotectedTxs bool   // --rpc.allow-unprotected-txs
	MinerGasPrice       uint64 // --miner.gasprice
	MinerGasLimit       uint64 // --miner.gaslimit

	// FrameMempoolTier selects the frame tx ruleset: "conservative" (default) or "aggressive".
	// conservative: VERIFY frame gas capped at 50K, no external calls.
	// aggressive:   VERIFY frame gas capped at 200K when a staked paymaster is detected.
	FrameMempoolTier string // --frame-mempool

	// LeanAvailableChainMode enables PQ attestation for a per-slot validator subset (US-PQ-2).
	LeanAvailableChainMode bool // --lean-available-chain

	// LeanAvailableChainValidators is the per-slot PQ attestor count [256,1024] (default 512).
	LeanAvailableChainValidators int // --lean-available-validators

	// StarkValidationFrames enables STARK proof sealing for VERIFY frame transactions (US-PQ-5b).
	StarkValidationFrames bool // --stark-validation-frames

	// SlotDuration selects the slot timing: "4s" or "6s" (default "6s", LEAN-1.1).
	SlotDuration string // --slot-duration

	// AttesterSampleSize controls per-slot attester sampling (GAP-3.4).
	// 0 = full committee mode, 256/512/1024 = sampled mode.
	AttesterSampleSize int // --attester-sample-size

	// FinalityMode selects the finality engine: "ssf" (default) or "minimmit" (GAP-5.2).
	FinalityMode string // --finality-mode

	// BLSBackend selects the BLS signature backend: "blst" (default) or "pure-go" (GAP-7.2).
	BLSBackend string // --bls-backend

	// MixnetMode selects the anonymous transport: "simulated" (default), "tor", "nym".
	MixnetMode string // --mixnet

	// LogLevel controls log verbosity (debug, info, warn, error).
	LogLevel string

	// Verbosity controls numeric log level (0=silent, 1=error, 2=warn,
	// 3=info, 4=debug, 5=trace). When set, overrides LogLevel.
	Verbosity int
}

// defaultDataDir returns the platform-specific default data directory.
func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".ETH2030"
	}
	return filepath.Join(home, ".ETH2030")
}

// DefaultConfig returns a Config with sensible defaults matching eth2030-geth.
func DefaultConfig() Config {
	return Config{
		DataDir:   defaultDataDir(),
		Name:      "ETH2030",
		Network:   "mainnet",
		NetworkID: 1,
		SyncMode:  "snap",
		GCMode:    "full",
		P2PPort:   30303,
		MaxPeers:  50,
		LogLevel:  "info",
		Verbosity: 3,

		// HTTP-RPC
		RPCPort:            8545,
		HTTPAddr:           "0.0.0.0",
		RPCAuthSecret:      "",
		RPCRateLimitPerSec: 100,
		RPCMaxRequestSize:  5 * 1024 * 1024,
		RPCMaxBatchSize:    100,
		RPCCORSOrigins:     "*",
		HTTPVhosts:         []string{"*"},
		HTTPCORSDomain:     []string{"*"},
		HTTPModules:        []string{"eth", "net", "web3", "engine", "admin", "debug", "txpool"},

		// Engine API
		EnginePort:           8551,
		AuthAddr:             "0.0.0.0",
		EngineMaxRequestSize: 5 * 1024 * 1024,
		EngineAuthToken:      "",
		EngineAuthTokenPath:  "",
		AuthVhosts:           []string{"*"},
		JWTSecret:            "",

		// WebSocket
		WSEnabled: false,
		WSAddr:    "0.0.0.0",
		WSPort:    8546,
		WSModules: []string{},
		WSOrigins: []string{},

		// Metrics
		Metrics:     false,
		MetricsAddr: "0.0.0.0",
		MetricsPort: 9001,

		// Frame mempool (EIP-8141 AA-2.3)
		FrameMempoolTier: "conservative",

		// EP-3 Post-Quantum defaults.
		LeanAvailableChainValidators: 512,

		// LEAN-1.1: default to 6-second slots.
		SlotDuration: "6s",

		// GAP-3.4: default to full committee mode (no sampling).
		AttesterSampleSize: 0,

		// GAP-5.2: default to SSF finality.
		FinalityMode: "ssf",

		// GAP-7.2: default to blst BLS backend.
		BLSBackend: "blst",

		// BB-1.1: default to simulated mixnet (no external daemon required).
		MixnetMode: "simulated",
	}
}

// Validate checks configuration values for correctness.
func (c *Config) Validate() error {
	if c.DataDir == "" {
		return errors.New("config: datadir must not be empty")
	}
	if c.P2PPort < 0 || c.P2PPort > 65535 {
		return fmt.Errorf("config: invalid p2p port: %d", c.P2PPort)
	}
	if c.RPCPort < 0 || c.RPCPort > 65535 {
		return fmt.Errorf("config: invalid rpc port: %d", c.RPCPort)
	}
	if c.EnginePort < 0 || c.EnginePort > 65535 {
		return fmt.Errorf("config: invalid engine port: %d", c.EnginePort)
	}
	if c.WSPort < 0 || c.WSPort > 65535 {
		return fmt.Errorf("config: invalid websocket port: %d", c.WSPort)
	}
	if c.MetricsPort < 0 || c.MetricsPort > 65535 {
		return fmt.Errorf("config: invalid metrics port: %d", c.MetricsPort)
	}
	if c.RPCMaxRequestSize <= 0 {
		return fmt.Errorf("config: invalid rpc max request size: %d", c.RPCMaxRequestSize)
	}
	if c.RPCRateLimitPerSec < 0 {
		return fmt.Errorf("config: invalid rpc rate limit: %d", c.RPCRateLimitPerSec)
	}
	if c.RPCMaxBatchSize <= 0 {
		return fmt.Errorf("config: invalid rpc max batch size: %d", c.RPCMaxBatchSize)
	}
	if c.EngineMaxRequestSize <= 0 {
		return fmt.Errorf("config: invalid engine max request size: %d", c.EngineMaxRequestSize)
	}
	if c.MaxPeers < 0 {
		return fmt.Errorf("config: invalid max peers: %d", c.MaxPeers)
	}
	if c.FrameMempoolTier != "" && c.FrameMempoolTier != "conservative" && c.FrameMempoolTier != "aggressive" {
		return fmt.Errorf("config: invalid frame-mempool tier %q (must be conservative or aggressive)", c.FrameMempoolTier)
	}
	if c.Verbosity < 0 || c.Verbosity > 5 {
		return fmt.Errorf("config: verbosity must be 0-5, got %d", c.Verbosity)
	}
	// Network is only validated when not using a custom genesis.
	if c.GenesisPath == "" {
		switch c.Network {
		case "mainnet", "sepolia", "holesky":
		default:
			return fmt.Errorf("config: unknown network %q", c.Network)
		}
	}
	switch c.SyncMode {
	case "full", "snap":
	default:
		return fmt.Errorf("config: unknown sync mode %q", c.SyncMode)
	}
	switch c.GCMode {
	case "", "full", "archive":
	default:
		return fmt.Errorf("config: unknown gc mode %q", c.GCMode)
	}
	// Validate slot duration (LEAN-1.1).
	switch c.SlotDuration {
	case "", "6s": // "" treated as default 6s
	case "4s":
	default:
		return fmt.Errorf("config: invalid slot-duration %q (must be 4s or 6s)", c.SlotDuration)
	}
	// Validate lean available chain validator count when mode is enabled.
	if c.LeanAvailableChainMode {
		if c.LeanAvailableChainValidators < 256 || c.LeanAvailableChainValidators > 1024 {
			return fmt.Errorf("config: lean-available-validators must be in [256,1024], got %d", c.LeanAvailableChainValidators)
		}
	}
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("config: unknown log level %q", c.LogLevel)
	}
	// GAP-3.4: validate attester sample size.
	switch c.AttesterSampleSize {
	case 0, 256, 512, 1024:
	default:
		return fmt.Errorf("config: invalid attester-sample-size %d (must be 0, 256, 512, or 1024)", c.AttesterSampleSize)
	}
	// GAP-5.2: validate finality mode.
	switch c.FinalityMode {
	case "", "ssf", "minimmit":
	default:
		return fmt.Errorf("config: invalid finality-mode %q (must be ssf or minimmit)", c.FinalityMode)
	}
	// GAP-7.2: validate BLS backend.
	switch c.BLSBackend {
	case "", "blst", "pure-go":
	default:
		return fmt.Errorf("config: invalid bls-backend %q (must be blst or pure-go)", c.BLSBackend)
	}
	// BB-1.1: validate mixnet mode.
	switch c.MixnetMode {
	case "", "simulated", "tor", "nym":
	default:
		return fmt.Errorf("config: invalid mixnet mode %q (must be simulated|tor|nym)", c.MixnetMode)
	}
	return nil
}

// VerbosityToLogLevel converts a numeric verbosity level to a log level string.
func VerbosityToLogLevel(v int) string {
	switch {
	case v <= 1:
		return "error"
	case v == 2:
		return "warn"
	case v == 3:
		return "info"
	default:
		return "debug"
	}
}

// dataDirSubdirs lists subdirectories created inside the data directory.
var dataDirSubdirs = []string{
	"chaindata",
	"keystore",
	"nodes",
}

// InitDataDir creates the data directory and its standard subdirectories.
func (c *Config) InitDataDir() error {
	if c.DataDir == "" {
		return errors.New("config: datadir must not be empty")
	}
	if err := os.MkdirAll(c.DataDir, 0700); err != nil {
		return fmt.Errorf("config: create datadir: %w", err)
	}
	for _, sub := range dataDirSubdirs {
		dir := filepath.Join(c.DataDir, sub)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("config: create %s: %w", sub, err)
		}
	}
	return nil
}

// ResolvePath resolves a path relative to the data directory.
func (c *Config) ResolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(c.DataDir, path)
}

// P2PAddr returns the P2P listen address string.
func (c *Config) P2PAddr() string {
	return fmt.Sprintf(":%d", c.P2PPort)
}

// HTTPListenAddr returns the HTTP-RPC listen address string.
func (c *Config) HTTPListenAddr() string {
	return fmt.Sprintf("%s:%d", c.HTTPAddr, c.RPCPort)
}

// AuthListenAddr returns the Engine API listen address string.
func (c *Config) AuthListenAddr() string {
	return fmt.Sprintf("%s:%d", c.AuthAddr, c.EnginePort)
}

// WSListenAddr returns the WebSocket RPC listen address string.
func (c *Config) WSListenAddr() string {
	return fmt.Sprintf("%s:%d", c.WSAddr, c.WSPort)
}

// MetricsListenAddr returns the metrics HTTP server listen address string.
func (c *Config) MetricsListenAddr() string {
	return fmt.Sprintf("%s:%d", c.MetricsAddr, c.MetricsPort)
}

// RPCAddr returns the HTTP-RPC listen address (alias for HTTPListenAddr).
func (c *Config) RPCAddr() string {
	return c.HTTPListenAddr()
}

// EngineAddr returns the Engine API listen address (alias for AuthListenAddr).
func (c *Config) EngineAddr() string {
	return c.AuthListenAddr()
}

// EffectiveDiscoveryPort returns the UDP discovery port, defaulting to P2PPort.
func (c *Config) EffectiveDiscoveryPort() int {
	if c.DiscoveryPort > 0 {
		return c.DiscoveryPort
	}
	return c.P2PPort
}

// JWTSecretPath returns the path to the JWT secret file, defaulting to
// <datadir>/jwtsecret when JWTSecret is empty.
func (c *Config) JWTSecretPath() string {
	if c.JWTSecret != "" {
		return c.JWTSecret
	}
	return filepath.Join(c.DataDir, "jwtsecret")
}

// SplitModules splits a comma-separated module string into a trimmed slice.
func SplitModules(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			result = append(result, v)
		}
	}
	return result
}

// RPCCORSAllowedOrigins returns the parsed CORS origin allowlist.
func (c *Config) RPCCORSAllowedOrigins() []string {
	out := make([]string, 0, 8)
	seen := make(map[string]struct{}, 8)
	for _, origin := range strings.Split(c.RPCCORSOrigins, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	if len(out) == 0 {
		out = append(out, "*")
	}
	return out
}
