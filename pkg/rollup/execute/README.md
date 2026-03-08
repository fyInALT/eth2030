# rollup/execute — EXECUTE precompile for native rollup state transitions (EIP-8079)

## Overview

This package implements the EIP-8079 `EXECUTE` precompile, which allows L1
contracts to verify a native rollup's state transition by re-executing a rollup
block. The precompile accepts a compact ABI-packed input containing a chain ID,
pre-state root, block data, optional witness, and anchor data; it returns the
post-state root, receipts root, gas used, burned fees, and a success flag.

Alongside the precompile, the package provides `ExecutionContext` for tracking
nested EXECUTE calls with gas metering, depth limits, and a full execution trace.
Gas cost is `ExecuteBaseGas (100,000) + len(blockData) * ExecutePerByteGas (16)`.
Blob transactions are rejected per the EIP-8079 specification.

## Functionality

**`ExecutePrecompile`**

- `RequiredGas(input) uint64` — returns gas cost from the 4-byte block-data-length field
- `Run(input) ([]byte, error)` — decodes input, executes STF, encodes 81-byte output
- `ValidateRollupExecution(input) error` — validates chain ID, pre-state root, block data size

**`ExecutionContext`** (context.go)

- `NewExecutionContext(rollupID, gasBudget, config) *ExecutionContext`
- `BeginCall(targetRollupID, caller, input, gas) error` — starts a nested call; checks depth and gas
- `EndCall(gasUsed, success, output) error` — closes the last call; refunds unused gas
- `Finish() (types.Hash, error)` — finalizes context; computes trace commitment hash
- `VerifyResult(claimedHash) (bool, error)`
- `GasRemaining() / GasUsed() / CurrentDepth() / TraceLength() uint64|int`
- `Trace() []ExecCallRecord`

**Input format** — `[chainID(8)] [preStateRoot(32)] [blockDataLen(4)] [witnessLen(4)] [anchorDataLen(4)] [blockData] [witness] [anchorData]`

**Output format** — `[postStateRoot(32)] [receiptsRoot(32)] [gasUsed(8)] [burnedFees(8)] [success(1)]`

**Parent package:** [rollup](../)
