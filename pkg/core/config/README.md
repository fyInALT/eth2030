# config

Chain configuration, fork activation schedule, genesis, and message types.

[← core](../README.md)

## Overview

Package `config` defines `ChainConfig`, the central fork-activation record for
the ETH2030 chain. It covers the full upgrade path from Homestead through M+,
using block numbers for pre-merge forks and Unix timestamps for all post-merge
forks (Shanghai, Cancun, Prague, Amsterdam, Glamsterdam, Hogotá, I+, and the
BPO blob schedules).

It also provides the `Genesis` type for programmatic genesis creation,
`GenesisAlloc` for pre-funded accounts, and fork schedule utilities for
querying and diffing chain configurations.

## Functionality

### ChainConfig

```go
type ChainConfig struct {
    ChainID *big.Int
    // block-number forks: HomesteadBlock ... LondonBlock
    // timestamp forks:
    ShanghaiTime    *uint64
    CancunTime      *uint64
    PragueTime      *uint64
    AmsterdamTime   *uint64
    GlamsterdanTime *uint64
    HogotaTime      *uint64
    IPlusTime       *uint64
    BPO1Time        *uint64
    BPO2Time        *uint64
    EIP7864FinalHashTime *uint64
    BinaryTreeHashFunc   string // "sha256" | "blake3"
}
```

Fork predicate methods (all follow the pattern `Is<Fork>(time uint64) bool`):
`IsShanghai`, `IsCancun`, `IsPrague`, `IsAmsterdam`, `IsGlamsterdan`,
`IsHogota`, `IsIPlus`.

Block-number predicate methods: `IsHomestead`, `IsEIP150`, `IsEIP155`,
`IsEIP158`, `IsByzantium`, `IsConstantinople`, `IsPetersburg`, `IsIstanbul`,
`IsBerlin`, `IsLondon`.

### Fork Schedule Utilities (`chain_config_forks.go`)

- `ForkID` — `{Name string, Block *big.Int, Timestamp *uint64}` with
  `IsActive(num, time)` and `String()`.
- `ForkSchedule()` — returns the complete ordered `[]ForkID` for a config.
- `ActiveForks(num, time)` — returns only the currently active forks.
- `ConfigDiff(other)` — returns a human-readable list of differences between
  two chain configs (useful for detecting incompatible fork changes on startup).

### Chain Config Extensions (`chain_config_ext.go`)

Validation helpers and custom fork ordering utilities used by the Engine API
and devnet tooling.

### Genesis

```go
type Genesis struct {
    Config     *ChainConfig
    Timestamp  uint64
    GasLimit   uint64
    Difficulty *big.Int
    Alloc      GenesisAlloc
    // optional: BaseFee, ExcessBlobGas, BlobGasUsed, Number, ...
}

type GenesisAlloc map[types.Address]GenesisAccount

type GenesisAccount struct {
    Balance *big.Int
    Code    []byte
    Nonce   uint64
    Storage map[types.Hash]types.Hash
}
```

- `Genesis.ToBlock()` — produces the genesis `*types.Block` from the spec.
- `GenesisInit` (`genesis_init.go`) — wires a genesis into a state database
  and raw database in one call.
- `genesis_utils.go` — helpers for computing genesis hash, applying alloc to
  a `StateDB`, and deriving the genesis state root.

### Message (`message.go`)

`Message` is a processed transaction representation passed to the EVM,
containing sender, nonce, value, gas, data, and access list fields resolved
before execution.

## Usage

```go
t0 := uint64(0)
cfg := &config.ChainConfig{
    ChainID:         big.NewInt(1),
    HomesteadBlock:  big.NewInt(0),
    LondonBlock:     big.NewInt(0),
    ShanghaiTime:    &t0,
    CancunTime:      &t0,
    GlamsterdanTime: &t0,
}

if cfg.IsGlamsterdan(blockTimestamp) {
    // apply Glamsterdam-specific logic
}

schedule := cfg.ForkSchedule()
active := cfg.ActiveForks(blockNumber, blockTimestamp)
```
