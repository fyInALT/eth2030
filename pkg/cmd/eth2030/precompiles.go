package main

import (
	"log/slog"

	"github.com/eth2030/eth2030/node"
)

// logForkOverrides logs the configured fork override timestamps at startup.
// For the native eth2030 EVM, fork activations are driven entirely by the
// ChainConfig timestamps wired in node.New() via applyForkOverrides — no
// separate precompile injection is needed (unlike the geth adapter path).
func logForkOverrides(cfg *node.Config) {
	if cfg.GlamsterdamOverride != nil {
		slog.Info("fork override: Glamsterdam", "timestamp", *cfg.GlamsterdamOverride)
	}
	if cfg.HogotaOverride != nil {
		slog.Info("fork override: Hogota", "timestamp", *cfg.HogotaOverride)
	}
	if cfg.IPlusOverride != nil {
		slog.Info("fork override: I+", "timestamp", *cfg.IPlusOverride)
	}
}
