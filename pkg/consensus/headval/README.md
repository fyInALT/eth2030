# headval

Block header validation helpers for the consensus layer.

## Overview

Package `headval` provides stateless block header validation against Ethereum
consensus rules. `HeaderValidator` checks parent-hash linkage, block number
continuity, timestamp ordering, gas limit bounds, gas usage, and extra data
length. It is used by the Engine API block processor before importing a
payload.

Under Proof-of-Stake, `CalcDifficulty` always returns zero; the difficulty
field is kept for backward compatibility with the execution-layer type system.

## Functionality

### Types

| Name | Description |
|------|-------------|
| `HeaderValidator` | Stateless validator; all methods operate on `*types.Header` |

### Constants

| Name | Value | Description |
|------|-------|-------------|
| `MaxExtraDataBytes` | 32 | Maximum `Extra` field length in bytes |
| `GasLimitBoundDivisor` | 1024 | Max allowed gas limit change per block (`parent / 1024`) |

### Functions / methods

| Name | Description |
|------|-------------|
| `NewHeaderValidator() *HeaderValidator` | Create a validator |
| `(*HeaderValidator).ValidateHeader(header, parent *types.Header) error` | Full header validation |
| `ValidateGasLimit(parentLimit, headerLimit uint64) bool` | Gas limit bounds check |
| `ValidateTimestamp(parentTime, headerTime uint64) bool` | True if `headerTime > parentTime` |
| `CalcDifficulty(parentDiff *big.Int, parentTs, currentTs uint64) *big.Int` | Always returns zero (PoS) |

### Errors

`ErrInvalidParentHash`, `ErrInvalidNumber`, `ErrInvalidTimestamp`, `ErrInvalidGasLimit`, `ErrGasUsedExceedsLimit`, `ErrExtraDataTooLong`, `ErrNilHeader`, `ErrNilParent`

## Usage

```go
import "github.com/eth2030/eth2030/consensus/headval"

hv := headval.NewHeaderValidator()
if err := hv.ValidateHeader(newHeader, parentHeader); err != nil {
    return fmt.Errorf("invalid block header: %w", err)
}
```

[← consensus](../README.md)
