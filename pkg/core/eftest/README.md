# eftest

Ethereum Foundation state test runner — 36,126/36,126 passing via go-ethereum backend.

[← core](../README.md)

## Overview

Package `eftest` implements the standard EF state test runner used to validate
ETH2030's execution layer against the official Ethereum test suite. Tests are
loaded from JSON fixtures in the EF format and executed through two backends:
the native ETH2030 `StateProcessor` (via `state_test_runner.go`) and a
go-ethereum v1.17.0 backend (via `geth_runner.go`).

The go-ethereum path achieves 36,126/36,126 (100%) pass rate across all
supported fork categories by delegating execution to `pkg/geth/`, which injects
ETH2030's custom precompiles and chain config into a vanilla go-ethereum
state transition.

## Functionality

### Fixture Loading (`fixture_loader.go`)

- `DiscoverFixtures(dir string) ([]string, error)` — walks a directory tree
  and returns sorted paths to all `.json` test files.
- `RunSingleFixture(path, forkFilter string) ([]*TestResult, error)` — parses
  and runs all subtests in one JSON file, optionally filtered by fork name.
- `BatchResult` — aggregate result: `Total`, `Passed`, `Failed`, `Skipped`,
  `Errors []*TestResult`.
- `TestResult` — per-subtest outcome: `File`, `Name`, `Fork`, `Index`,
  `Passed`, `Error`.

### Go-Ethereum Runner (`geth_runner.go`)

- `GethRunResult` — per-subtest outcome with `ExpectedRoot`, `GotRoot`,
  `ExpectedLogs`, `GotLogs`, and `Error`.
- `LoadGethTests(path string)` — parses the EF JSON format into
  `GethStateTest` records ready for execution.
- `GethStateTest.Run(t *testing.T)` — executes subtests through the
  go-ethereum state transition and verifies state root and log hash.

### Native Runner (`state_test_runner.go`)

- JSON structs: `stJSON`, `stEnv`, `stAccount`, `stTransaction`, `stPostState`.
- `RunStateTest(path, forkFilter string)` — runs the native ETH2030 path for
  smoke-testing individual EIP implementations.

### Test Organization

Tests in `geth_runner_test.go` group fixtures by fork category (Berlin,
London, Merge, Shanghai, Cancun, Prague, Byzantium, etc.) and run them in
parallel goroutines with a `sync.WaitGroup` for maximum throughput.

## Usage

```go
// Discover and run all fixtures in a directory.
files, err := eftest.DiscoverFixtures("/path/to/GeneralStateTests")
for _, f := range files {
    results, err := eftest.RunSingleFixture(f, "Cancun")
    for _, r := range results {
        if !r.Passed {
            log.Printf("FAIL %s/%s: %v", r.Fork, r.Name, r.Error)
        }
    }
}
```
