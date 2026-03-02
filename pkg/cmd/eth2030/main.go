// Command eth2030 is the main entry point for the eth2030 Ethereum client.
//
// Usage:
//
//	eth2030 [flags]
//
// Flags match eth2030-geth for Kurtosis devnet compatibility.
package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/eth2030/eth2030/node"
)

// Build-time version info, overridable with ldflags.
var (
	version = "v0.1.0-dev"
	commit  = "unknown"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

// run is the actual entry point, returning an exit code.
func run(args []string) int {
	cfg, exit, code := parseFlags(args)
	if exit {
		return code
	}

	cfg.LogLevel = node.VerbosityToLogLevel(cfg.Verbosity)
	setupLogging(cfg.Verbosity)

	slog.Info("eth2030 starting", "version", version,
		"network", cfg.Network,
		"datadir", cfg.DataDir,
		"syncmode", cfg.SyncMode,
	)
	slog.Info("HTTP-RPC",
		"addr", cfg.HTTPListenAddr(),
		"vhosts", cfg.HTTPVhosts,
		"modules", cfg.HTTPModules,
	)
	slog.Info("Engine API",
		"addr", cfg.AuthListenAddr(),
		"vhosts", cfg.AuthVhosts,
		"jwt", cfg.JWTSecretPath(),
	)
	slog.Info("P2P",
		"port", cfg.P2PPort,
		"maxpeers", cfg.MaxPeers,
		"bootnodes", cfg.Bootnodes,
	)
	if cfg.Metrics {
		slog.Info("Metrics", "addr", cfg.MetricsListenAddr())
	}
	if cfg.WSEnabled {
		slog.Info("WebSocket RPC", "addr", cfg.WSListenAddr())
	}

	if err := cfg.Validate(); err != nil {
		slog.Error("invalid configuration", "err", err)
		return 1
	}

	if err := cfg.InitDataDir(); err != nil {
		slog.Error("failed to initialize datadir", "err", err)
		return 1
	}

	n, err := node.New(&cfg)
	if err != nil {
		slog.Error("failed to create node", "err", err)
		return 1
	}

	if err := n.Start(); err != nil {
		slog.Error("failed to start node", "err", err)
		return 1
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	slog.Info("received signal, shutting down", "signal", sig)

	if err := n.Stop(); err != nil {
		slog.Error("error during shutdown", "err", err)
		return 1
	}

	slog.Info("shutdown complete")
	return 0
}

// parseFlags parses CLI arguments into a node.Config.
// Returns (config, shouldExit, exitCode).
func parseFlags(args []string) (node.Config, bool, int) {
	cfg := node.DefaultConfig()
	fs := newFlagSet(&cfg)

	showVersion := fs.Bool("version", false, "print version and exit")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return cfg, true, 2
	}

	if *showVersion {
		fmt.Printf("eth2030 %s (commit %s)\n", version, commit)
		return cfg, true, 0
	}

	// Handle --override.*=0 special case: flag.Uint64 treats 0 as "not set"
	// because it is the zero value. We scan raw args so an explicit =0 activates
	// the fork at genesis.
	applyZeroForkOverrides(args, &cfg)

	return cfg, false, 0
}

// newFlagSet creates a flagSet binding all CLI flags to cfg.
func newFlagSet(cfg *node.Config) *flagSet {
	fs := newCustomFlagSet("eth2030")

	// --- Node ---
	fs.StringVar(&cfg.DataDir, "datadir", cfg.DataDir, "data directory path")
	fs.StringVar(&cfg.Network, "network", cfg.Network, "network to join (mainnet, sepolia, holesky)")
	fs.IntVar(&cfg.P2PPort, "port", cfg.P2PPort, "P2P listening port")
	fs.IntVar(&cfg.MaxPeers, "maxpeers", cfg.MaxPeers, "maximum number of P2P peers")
	fs.StringVar(&cfg.SyncMode, "syncmode", cfg.SyncMode, "sync mode (full, snap)")
	fs.StringVar(&cfg.GCMode, "gcmode", cfg.GCMode, "GC mode: archive (no pruning) or full")
	fs.Uint64Var(&cfg.NetworkID, "networkid", cfg.NetworkID, "network ID override (0 = use genesis chain ID)")
	fs.IntVar(&cfg.Verbosity, "verbosity", cfg.Verbosity, "log level 0-5 (0=silent, 5=trace)")

	// --- HTTP-RPC ---
	fs.Bool("http", false, "enable HTTP-RPC server (always on; kept for compatibility)")
	fs.StringVar(&cfg.HTTPAddr, "http.addr", cfg.HTTPAddr, "HTTP-RPC listen address")
	fs.IntVar(&cfg.RPCPort, "http.port", cfg.RPCPort, "HTTP-RPC port")
	fs.StringSliceVar(&cfg.HTTPVhosts, "http.vhosts", cfg.HTTPVhosts, "comma-separated virtual hostnames for HTTP-RPC")
	fs.StringSliceVar(&cfg.HTTPCORSDomain, "http.corsdomain", cfg.HTTPCORSDomain, "comma-separated CORS domains")
	fs.StringSliceVar(&cfg.HTTPModules, "http.api", cfg.HTTPModules, "comma-separated HTTP-RPC API modules")

	// --- Auth / Engine API ---
	fs.StringVar(&cfg.AuthAddr, "authrpc.addr", cfg.AuthAddr, "Engine API listen address")
	fs.IntVar(&cfg.EnginePort, "authrpc.port", cfg.EnginePort, "Engine API port")
	fs.StringSliceVar(&cfg.AuthVhosts, "authrpc.vhosts", cfg.AuthVhosts, "comma-separated virtual hostnames for Engine API")
	fs.StringVar(&cfg.JWTSecret, "authrpc.jwtsecret", cfg.JWTSecret, "path to JWT secret for Engine API auth")

	// --- WebSocket ---
	fs.BoolVar(&cfg.WSEnabled, "ws", cfg.WSEnabled, "enable WebSocket RPC server")
	fs.StringVar(&cfg.WSAddr, "ws.addr", cfg.WSAddr, "WebSocket RPC listen address")
	fs.IntVar(&cfg.WSPort, "ws.port", cfg.WSPort, "WebSocket RPC port")
	fs.StringSliceVar(&cfg.WSModules, "ws.api", cfg.WSModules, "comma-separated WebSocket API modules")
	fs.StringSliceVar(&cfg.WSOrigins, "ws.origins", cfg.WSOrigins, "comma-separated allowed WebSocket origins")

	// --- P2P ---
	fs.StringVar(&cfg.Bootnodes, "bootnodes", cfg.Bootnodes, "comma-separated enode URLs for bootstrap")
	fs.IntVar(&cfg.DiscoveryPort, "discovery.port", cfg.DiscoveryPort, "UDP port for node discovery")
	fs.StringVar(&cfg.NAT, "nat", cfg.NAT, "NAT traversal method (e.g. extip:1.2.3.4)")

	// --- Metrics ---
	fs.BoolVar(&cfg.Metrics, "metrics", cfg.Metrics, "enable metrics collection")
	fs.StringVar(&cfg.MetricsAddr, "metrics.addr", cfg.MetricsAddr, "metrics HTTP server address")
	fs.IntVar(&cfg.MetricsPort, "metrics.port", cfg.MetricsPort, "metrics HTTP server port")

	// --- Genesis / fork overrides ---
	fs.StringVar(&cfg.GenesisPath, "override.genesis", cfg.GenesisPath, "path to custom genesis.json (for Kurtosis devnets)")
	fs.Uint64PtrVar(&cfg.GlamsterdamOverride, "override.glamsterdam", "override Glamsterdam fork timestamp")
	fs.Uint64PtrVar(&cfg.HogotaOverride, "override.hogota", "override Hogota fork timestamp")
	fs.Uint64PtrVar(&cfg.IPlusOverride, "override.iplus", "override I+ fork timestamp (enables NTT/NII precompiles)")

	// --- Misc ---
	fs.BoolVar(&cfg.AllowUnprotectedTxs, "rpc.allow-unprotected-txs", cfg.AllowUnprotectedTxs, "allow non-EIP155 transactions over RPC")
	fs.Uint64Var(&cfg.MinerGasPrice, "miner.gasprice", cfg.MinerGasPrice, "minimum gas price for mining (wei)")
	fs.Uint64Var(&cfg.MinerGasLimit, "miner.gaslimit", cfg.MinerGasLimit, "target gas ceiling for mined blocks")

	return fs
}

// applyZeroForkOverrides scans raw args for --override.*=0 patterns so that
// an explicit zero value activates the fork at genesis (0 is otherwise
// indistinguishable from "flag not set" for Uint64Ptr flags).
func applyZeroForkOverrides(args []string, cfg *node.Config) {
	zero := uint64(0)
	for _, arg := range args {
		switch arg {
		case "--override.glamsterdam=0", "-override.glamsterdam=0":
			cfg.GlamsterdamOverride = &zero
		case "--override.hogota=0", "-override.hogota=0":
			cfg.HogotaOverride = &zero
		case "--override.iplus=0", "-override.iplus=0":
			cfg.IPlusOverride = &zero
		}
	}
}

// defaultDataDir returns the default data directory path.
func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".eth2030"
	}
	return filepath.Join(home, ".eth2030")
}

// setupLogging configures the slog default logger based on verbosity.
func setupLogging(verbosity int) {
	var level slog.Level
	switch {
	case verbosity <= 1:
		level = slog.LevelError
	case verbosity == 2:
		level = slog.LevelWarn
	case verbosity == 3:
		level = slog.LevelInfo
	default:
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
}
