# Package geth

Adapter layer between ETH2030's type system and the go-ethereum execution engine.

## Overview

The `geth` package is the sole boundary between ETH2030 and go-ethereum (v1.17.0). All other ETH2030 packages use `eth2030/core/types`; this is the only package that imports `github.com/ethereum/go-ethereum` directly. Its role is to translate between the two type systems and to inject ETH2030-specific functionality (custom precompiles, fork detection) into go-ethereum's execution path.

The package provides bidirectional type conversion (address, hash, access list, logs, balance), a `GethBlockProcessor` that executes ETH2030 blocks through go-ethereum's state transition engine, pre-state creation for EF state test execution, and a `PrecompileAdapter` that wraps ETH2030 precompiles to satisfy go-ethereum's `PrecompiledContract` interface.

Custom opcode injection (CLZ, DUPN/SWAPN/EXCHANGE, APPROVE, TXPARAM\*, EOF, AA opcodes) is not possible from this package because go-ethereum's `operation` struct and `JumpTable` are unexported. Those opcodes are available only through ETH2030's native EVM interpreter in `pkg/core/vm/`.

## Table of Contents

- [Functionality](#functionality)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### Type Conversion

Zero-copy conversions between ETH2030 and go-ethereum types:

- `ToGethAddress` / `FromGethAddress` — `types.Address` ↔ `gethcommon.Address`
- `ToGethHash` / `FromGethHash` — `types.Hash` ↔ `gethcommon.Hash`
- `ToGethAccessList` — `types.AccessList` → `gethtypes.AccessList`
- `FromGethLogs` / `FromGethLog` — `[]*gethtypes.Log` → `[]*types.Log`
- `ToUint256` / `FromUint256` — `*big.Int` ↔ `*uint256.Int`

### Chain Config Mapping

- `ToGethChainConfig` — maps ETH2030 `ChainConfig` to go-ethereum `params.ChainConfig` (Frontier through Prague)
- `ToGethChainConfigWithEth2028Forks` — additionally maps ETH2030 forks (Glamsterdam, Hogota) onto `PragueTime`
- `EFTestChainConfig` / `EFTestForkSupported` — build cumulative chain configs for EF state test fork names

### Custom Precompile Injection

ETH2030 adds 18 custom precompiles beyond go-ethereum's standard set, grouped in four categories:

| Category | Fork | Addresses | Description |
|----------|------|-----------|-------------|
| Repricing | Glamsterdam | `0x06`, `0x08`, `0x09`, `0x0a` | Repriced ecAdd, ecPairing, BLAKE2f, KZG point eval |
| NTT | I+ | `0x0f`–`0x14` | Number Theoretic Transform: FW, INV, VecMulMod, VecAddMod, DotProduct, Butterfly |
| NII | I+ | `0x0201`–`0x0204` | ModExp, FieldMul, FieldInv, BatchVerify |
| Field | I+ | `0x0205`–`0x0208` | FieldMulExt, FieldInvExt, FieldExp, BatchFieldVerify |

Key functions:
- `InjectCustomPrecompiles(rules, forkLevel)` — returns a merged precompile map for the active fork
- `Eth2028ForkLevelFromConfig(config, time)` — detects the active ETH2030 fork (Prague / Glamsterdam / Hogota / I+)
- `CustomPrecompileAddresses(forkLevel)` — returns addresses for EIP-2929 access list warming
- `ListCustomPrecompiles()` — returns full metadata for all custom precompiles

### Block Processor

`GethBlockProcessor` executes ETH2030 blocks using go-ethereum's EVM and state transition:

- `NewGethBlockProcessor(config)` — standard processor
- `NewGethBlockProcessorWithEth2028(config, eth2030Cfg)` — processor with custom precompile injection
- `ProcessBlock(statedb, block, getHash)` — executes all transactions, processes EIP-4895 withdrawals, touches coinbase, commits state, returns receipts and post-state root

### State and Message Utilities

- `MakePreState(accounts)` — creates a go-ethereum `StateDB` from a map of pre-state accounts (used by EF test runner)
- `MakeBlockContext(header, getHash)` — builds a go-ethereum `BlockContext` from an ETH2030 header
- `MakeMessage(...)` — constructs a go-ethereum `Message` for EF test execution
- `ApplyMessage(statedb, config, blockCtx, msg, gasLimit)` — executes a message and returns the result

## Usage

```go
// Convert chain config for use with go-ethereum
gethCfg := geth.ToGethChainConfigWithEth2028Forks(eth2030ChainConfig)

// Detect active ETH2030 fork level and inject custom precompiles
forkLevel := geth.Eth2028ForkLevelFromConfig(eth2030ChainConfig, blockTime)
rules := gethCfg.Rules(blockNumber, true, blockTime)
precompiles := geth.InjectCustomPrecompiles(rules, forkLevel)
evm.SetPrecompiles(precompiles)

// Execute a block
processor := geth.NewGethBlockProcessorWithEth2028(gethCfg, eth2030ChainConfig)
receipts, stateRoot, err := processor.ProcessBlock(statedb, block, getHashFn)

// Type conversion
gethAddr := geth.ToGethAddress(eth2030Address)
eth2030Hash := geth.FromGethHash(gethHash)
```

## Documentation References

- [L1 Strawmap Roadmap](https://strawmap.org/)
- [go-ethereum v1.17.0](https://github.com/ethereum/go-ethereum)
- ETH2030 EF test runner: `pkg/core/eftest/`
- ETH2030 native EVM: `pkg/core/vm/`
