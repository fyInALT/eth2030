# internal/testutil — Shared block-building helpers for tests

## Overview

`internal/testutil` provides lightweight helpers for constructing genesis blocks and child blocks against the ETH2030 execution layer. It is intended for use across test packages that need valid `types.Block` values without duplicating header-field wiring (blob gas, calldata gas, multidimensional gas vectors, BAL hash, EIP-4788 beacon root, EIP-7685 requests hash, etc.).

All exported symbols are test utilities only; the package lives under `internal/` and is not importable by external modules.

## Functionality

| Function | Description |
|---|---|
| `NewUint64(v uint64) *uint64` | Convenience pointer constructor. |
| `MakeGenesis(gasLimit uint64, baseFee *big.Int) *types.Block` | Builds block 0 with all post-Cancun header fields zeroed/empty (blob gas, calldata gas, withdrawals hash, beacon root, requests hash). |
| `MakeBlock(parent *types.Block, txs []*types.Transaction) *types.Block` | Builds a child block against a fresh empty `state.MemoryStateDB`. Use only for the first block after genesis. |
| `MakeBlockWithState(parent, txs, statedb) *types.Block` | Builds a child block by executing `txs` through `execution.StateProcessor.ProcessWithBAL`; updates `GasUsed`, `CalldataGasUsed`, `BlockAccessListHash`, `Bloom`, `ReceiptHash`, and `Root` from the execution result. The state is mutated in place, enabling chains of blocks. |

`MakeBlockWithState` derives:
- `ExcessBlobGas` via `gas.CalcExcessBlobGas`
- `CalldataExcessGas` via `gas.CalcCalldataExcessGas`
- `BaseFee` via `gas.CalcBaseFee`
- Block time is parent time + 12 seconds.

## Usage

```go
import "github.com/eth2030/eth2030/internal/testutil"

genesis := testutil.MakeGenesis(30_000_000, big.NewInt(1e9))

statedb := state.NewMemoryStateDB()
block1 := testutil.MakeBlockWithState(genesis, txs, statedb)
block2 := testutil.MakeBlockWithState(block1, moreTxs, statedb) // same statedb
```

---

Parent: [`internal`](../)
