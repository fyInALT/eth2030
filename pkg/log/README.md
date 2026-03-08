# Package log

Structured logging for the ETH2030 Ethereum execution client.

## Overview

The `log` package wraps Go's standard `log/slog` library with Ethereum-specific conveniences. It provides a `Logger` type that writes structured JSON to stderr by default, a process-wide default logger accessible via package-level functions, and a `Module` method for creating named child loggers that subsystems (EVM, txpool, p2p, etc.) use to tag their output.

The package is intentionally minimal and dependency-free. It does not introduce any log routing or aggregation beyond what `slog` provides. The `formatter` subpackage adds text-format output for human-readable logs in development environments.

## Table of Contents

- [Functionality](#functionality)
- [Subpackages](#subpackages)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### Logger

`Logger` wraps `*slog.Logger` and exposes the standard four levels:

```go
func New(level slog.Level) *Logger           // JSON output to stderr
func NewWithHandler(h slog.Handler) *Logger  // custom handler (testing, alternate sinks)
```

Methods:
- `Debug(msg, args...)` / `Info(msg, args...)` / `Warn(msg, args...)` / `Error(msg, args...)` — structured log at the respective level
- `Module(name string) *Logger` — returns a child logger with `"module": name` attached to all records
- `With(args ...any) *Logger` — returns a child logger with additional key-value context

### Default Logger

A process-wide default logger is initialized at `slog.LevelInfo` writing JSON to stderr:

- `Default() *Logger` — returns the current default logger
- `SetDefault(l *Logger)` — replaces the default logger (e.g., to change level at startup)
- Package-level `Debug`, `Info`, `Warn`, `Error` functions delegate to the default logger

## Subpackages

| Subpackage | Description |
|------------|-------------|
| [`formatter/`](./formatter/) | Text-format log handler for human-readable development output |

## Usage

```go
// Obtain a per-module logger for a subsystem
log := log.Default().Module("txpool")
log.Info("transaction added", "hash", tx.Hash(), "gas", tx.Gas())

// Create a logger at debug level with a custom handler
logger := log.New(slog.LevelDebug)
logger.Debug("detailed state info", "root", stateRoot)

// Replace the global default logger at startup
log.SetDefault(log.New(slog.LevelWarn))

// Add persistent context fields
rpcLog := log.Default().Module("rpc").With("method", "eth_getBalance")
rpcLog.Error("request failed", "err", err)
```

## Documentation References

- [Go log/slog documentation](https://pkg.go.dev/log/slog)
- ETH2030 node startup: `pkg/node/node.go`
