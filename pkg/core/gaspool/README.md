# gaspool

Block gas limit tracking for transaction execution.

[← core](../README.md)

## Overview

Package `gaspool` provides `GasPool`, a simple counter that tracks the amount
of gas remaining in a block during transaction processing. It is decremented as
each transaction is applied and checked before applying the next one, ensuring
that the cumulative gas usage never exceeds the block gas limit.

## Functionality

### Types

- `GasPool` — a `uint64` value type representing available block gas.
- `ErrGasPoolExhausted` — returned by `SubGas` when the pool has insufficient
  gas for a requested amount.

### Methods

| Method | Description |
|---|---|
| `AddGas(amount uint64) *GasPool` | Adds gas to the pool; returns self for chaining. |
| `SubGas(amount uint64) error` | Subtracts gas; returns `ErrGasPoolExhausted` if insufficient. |
| `Gas() uint64` | Returns remaining gas. |

## Usage

```go
// Initialize pool from the block gas limit.
gp := new(gaspool.GasPool).AddGas(header.GasLimit)

for _, tx := range txs {
    if gp.Gas() < tx.Gas() {
        continue // not enough gas left for this tx
    }
    if err := gp.SubGas(tx.Gas()); err != nil {
        break
    }
    // apply transaction...
}
```
