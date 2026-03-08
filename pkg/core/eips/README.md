# eips

Self-contained EIP implementations for the ETH2030 execution layer.

[← core](../README.md)

## Overview

Package `eips` contains individual EIP implementations as focused, testable
modules. Each file corresponds to one EIP or a closely related group. They are
called by the block builder (`core/block`), the transaction executor
(`core/execution`), and system-contract initialization code. The package covers
the full range from Cancun-era system contracts through Glamsterdam/Hogotá and
I+ roadmap items.

## Functionality

### System Contracts

| File | EIP | Description |
|---|---|---|
| `beacon_root.go` | EIP-4788 | Stores `parentBeaconRoot` in the beacon root contract before each block |
| `eip2935.go` | EIP-2935 | Historical block hash storage contract (`ProcessParentBlockHash`) |
| `eip6110.go` | EIP-6110 | In-protocol deposit receipts from the deposit contract |
| `eip7002.go` | EIP-7002 | Execution-layer-triggered validator withdrawals |
| `eip7702.go` | EIP-7702 | SetCode — EOA delegation to contract code |
| `eip7997.go` | EIP-7997 | Deterministic CREATE2 factory deployment at Glamsterdam |

### Block Processing

| File | EIP | Description |
|---|---|---|
| `frame_execution.go` | EIP-8141 | Frame transaction execution (`ExecuteFrameTx`, `FrameExecutionContext`) |
| `inclusion_list.go` | EIP-7547/7805 | FOCIL inclusion list enforcement |
| `payload_chunking.go` | Hogotá | Payload chunking for large blocks |
| `payload_shrink.go` | Hogotá | Payload shrink logic for chunked payloads |
| `block_in_blobs.go` | Hogotá | Block-in-blobs encoding/decoding |

### Account Abstraction

| File | EIP | Description |
|---|---|---|
| `aa_entrypoint.go` | EIP-7701 | AA entry point execution |
| `paymaster_registry.go` | EIP-7701 | Paymaster registration and slash tracking |
| `elsa.go` | EIP-7701 | ELSA (execution layer smart accounts) helper |

### Gas

| File | Description |
|---|---|
| `access_gas.go` | Separate state-access gas accounting (I+ era multidim gas) |
| `tx_assertions.go` | Transaction assertion validation (pre/post-execution) |

### Key Constants

- `BeaconRootAddress`, `BeaconHistoryBufferLength` — EIP-4788 beacon root contract.
- `HistoryStorageAddress`, `HistoryServeWindow` — EIP-2935 block hash storage.
- `FactoryAddress` — EIP-7997 CREATE2 factory.
- `DepositContractAddr`, `MaxDepositsPerBlock` — EIP-6110 deposits.
- `DefaultReadGas` (2100), `DefaultWriteGas` (5000), `WarmReadGas` (100),
  `DefaultAccessGasLimit` (20M) — EIP-2929/access gas constants.

### Key Functions

- `ProcessBeaconBlockRoot(statedb, header)` — EIP-4788 system call.
- `ProcessParentBlockHash(statedb, num, hash)` — EIP-2935 system call.
- `ApplyEIP7997(statedb)` — deploys the deterministic CREATE2 factory.
- `ExecuteFrameTx(tx, stateNonce, callFn)` — EIP-8141 frame execution.
- `Uint64ToHash(n)` — packs a uint64 into a `types.Hash` for storage slot keys.

## Usage

```go
// EIP-4788: store beacon root before block transactions.
eips.ProcessBeaconBlockRoot(statedb, header)

// EIP-8141: execute a frame transaction.
ctx, err := eips.ExecuteFrameTx(frameTx, stateNonce, func(
    caller, target types.Address, gasLimit uint64, data []byte,
    mode uint8, frameIndex int,
) (uint64, uint64, []*types.Log, bool, uint8, error) {
    // invoke the EVM for each frame
    return evm.Call(caller, target, data, gasLimit, nil)
})
```
