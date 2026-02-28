// Package node implements the ETH2030 full node lifecycle,
// wiring together blockchain, RPC, Engine API, P2P, and TxPool.
package node

import (
	"errors"
	"fmt"
	"net"
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
	Network string

	// NetworkID is the numeric network identifier. Common values:
	// 1 = mainnet, 11155111 = sepolia, 17000 = holesky.
	NetworkID uint64

	// SyncMode selects the sync strategy (full, snap).
	SyncMode string

	// P2PPort is the TCP port for devp2p connections.
	P2PPort int

	// RPCPort is the HTTP port for the JSON-RPC server.
	RPCPort int

	// RPCHost is the interface that the HTTP-RPC server binds to.
	RPCHost string

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

	// EnginePort is the HTTP port for the Engine API server.
	EnginePort int

	// EngineHost is the interface that the Engine API server binds to.
	EngineHost string

	// EngineMaxRequestSize is the max body size for Engine API requests, in bytes.
	EngineMaxRequestSize int64

	// EngineAuthToken requires this bearer token for Engine API requests when set.
	EngineAuthToken string

	// EngineAuthTokenPath optionally loads the Engine API token from a file.
	EngineAuthTokenPath string

	// MaxPeers is the maximum number of P2P peers.
	MaxPeers int

	// LogLevel controls log verbosity (debug, info, warn, error).
	LogLevel string

	// Verbosity controls numeric log level (0=silent, 1=error, 2=warn,
	// 3=info, 4=debug, 5=trace). When set, overrides LogLevel.
	Verbosity int

	// Metrics enables the metrics collection subsystem.
	Metrics bool
}

// defaultDataDir returns the platform-specific default data directory.
// Falls back to ".ETH2030" in the current directory if the home directory
// cannot be determined.
func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".ETH2030"
	}
	return filepath.Join(home, ".ETH2030")
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		DataDir:              defaultDataDir(),
		Name:                 "ETH2030",
		Network:              "mainnet",
		NetworkID:            1,
		SyncMode:             "snap",
		P2PPort:              30303,
		RPCPort:              8545,
		RPCHost:              "127.0.0.1",
		RPCAuthSecret:        "",
		RPCRateLimitPerSec:   100,
		RPCMaxRequestSize:    5 * 1024 * 1024,
		RPCMaxBatchSize:      100,
		RPCCORSOrigins:       "*",
		EnginePort:           8551,
		EngineHost:           "127.0.0.1",
		EngineMaxRequestSize: 5 * 1024 * 1024,
		EngineAuthToken:      "",
		EngineAuthTokenPath:  "",
		MaxPeers:             50,
		LogLevel:             "info",
		Verbosity:            3,
		Metrics:              false,
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
	if c.RPCHost == "" {
		return errors.New("config: rpc host must not be empty")
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
	if c.EngineHost == "" {
		return errors.New("config: engine host must not be empty")
	}
	if c.EngineMaxRequestSize <= 0 {
		return fmt.Errorf("config: invalid engine max request size: %d", c.EngineMaxRequestSize)
	}
	if c.MaxPeers < 0 {
		return fmt.Errorf("config: invalid max peers: %d", c.MaxPeers)
	}
	if c.Verbosity < 0 || c.Verbosity > 5 {
		return fmt.Errorf("config: verbosity must be 0-5, got %d", c.Verbosity)
	}
	switch c.Network {
	case "mainnet", "sepolia", "holesky":
	default:
		return fmt.Errorf("config: unknown network %q", c.Network)
	}
	switch c.SyncMode {
	case "full", "snap":
	default:
		return fmt.Errorf("config: unknown sync mode %q", c.SyncMode)
	}
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("config: unknown log level %q", c.LogLevel)
	}
	return nil
}

// VerbosityToLogLevel converts a numeric verbosity level to a log level string.
func VerbosityToLogLevel(v int) string {
	switch {
	case v <= 0:
		return "error" // silent maps to error-only
	case v == 1:
		return "error"
	case v == 2:
		return "warn"
	case v == 3:
		return "info"
	default:
		return "debug" // 4 and 5 both map to debug
	}
}

// dataDirSubdirs lists subdirectories created inside the data directory.
var dataDirSubdirs = []string{
	"chaindata",
	"keystore",
	"nodes",
}

// InitDataDir creates the data directory and its standard subdirectories
// if they do not already exist. Returns an error if directory creation fails.
func (c *Config) InitDataDir() error {
	if c.DataDir == "" {
		return errors.New("config: datadir must not be empty")
	}

	// Create the root data directory.
	if err := os.MkdirAll(c.DataDir, 0700); err != nil {
		return fmt.Errorf("config: create datadir: %w", err)
	}

	// Create standard subdirectories.
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

// RPCAddr returns the RPC listen address string.
func (c *Config) RPCAddr() string {
	return net.JoinHostPort(c.RPCHost, fmt.Sprintf("%d", c.RPCPort))
}

// EngineAddr returns the Engine API listen address string.
func (c *Config) EngineAddr() string {
	return net.JoinHostPort(c.EngineHost, fmt.Sprintf("%d", c.EnginePort))
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
