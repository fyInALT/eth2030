# rollup

Native rollup support per EIP-8079: EXECUTE precompile, anchor contract, cross-layer proofs, sequencer, and state bridge.

## Overview

The `rollup` package implements the EIP-8079 native rollup design, which exposes the Ethereum state transition function (STF) as an EVM precompile (`EXECUTE`). This allows rollup validity to be verified directly on L1 without a separate trusted bridge contract, enabling native proof-based settlement of L2 state transitions.

The package provides a `RollupRegistry` that tracks registered rollup chains by ID, processes batch submissions (which advance the rollup state root), verifies state transition proofs, and handles L1-to-L2 deposits and L2-to-L1 withdrawals with proof verification. An anchor system contract on each rollup stores the latest verified L1 state for the rollup chain.

The cross-layer proof system and fraud proof mechanism support both optimistic and validity-proof-based rollup models. The `Sequencer` manages ordered batch submission, while the `StateBridge` handles synchronization of L1 state events to L2.

## Table of Contents

- [Functionality](#functionality)
- [Subpackages](#subpackages)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### Core Types

```go
// Fixed addresses.
var ExecutePrecompileAddress = types.BytesToAddress([]byte{0x01, 0x01})
var AnchorAddress            = types.BytesToAddress([]byte{0x01, 0x02})

type RollupConfig struct {
    ChainID          *big.Int
    AnchorAddress    types.Address
    GenesisStateRoot types.Hash
    GasLimit         uint64
    BaseFee          *big.Int
    AllowBlobTx      bool      // disabled by default per EIP-8079
}

type ExecuteInput struct {
    ChainID      uint64
    PreStateRoot types.Hash
    BlockData    []byte    // RLP-encoded block
    Witness      []byte    // optional stateless witness
    AnchorData   []byte    // L1 state to inject via anchor system tx
}

type ExecuteOutput struct {
    PostStateRoot types.Hash
    ReceiptsRoot  types.Hash
    GasUsed       uint64
    BurnedFees    uint64
    Success       bool
}
```

### Rollup Registry

`RollupRegistry` is a thread-safe map of registered native rollups:

- `RegisterRollup(config NativeRollupConfig) (*NativeRollup, error)` — registers a rollup with a unique non-zero ID, human-readable name, bridge contract address, genesis state root, and gas limit.
- `GetRollupState(rollupID uint64) (*NativeRollup, error)` — returns a copy of the current rollup state (state root, block number, deposit/withdrawal lists).
- `SubmitBatch(rollupID uint64, batchData []byte, stateRoot Hash) (*BatchResult, error)` — processes a batch, derives a new post-state root as `Keccak256(preStateRoot || batchData || claimedRoot)`, and increments the block number. Maximum batch size is 2 MiB.
- `VerifyStateTransition(rollupID uint64, pre, post Hash, proof []byte) (bool, error)` — verifies a state transition proof (minimum 32 bytes). Uses `SHA256(pre || post || proof)` as a verification commitment.
- `ProcessDeposit(rollupID uint64, from Address, amount *big.Int) (*NativeDeposit, error)` — records an L1-to-L2 deposit with a deterministic deposit ID.
- `ProcessWithdrawal(rollupID uint64, to Address, amount *big.Int, proof []byte) (*NativeWithdrawal, error)` — processes an L2-to-L1 withdrawal after verifying the withdrawal proof.

### Anchor Contract

The `anchor` subpackage implements the predeploy anchor contract (`AnchorAddress = 0x0102`) which stores the latest L1 state on the rollup:

```go
type AnchorState struct {
    LatestBlockHash  types.Hash
    LatestStateRoot  types.Hash
    BlockNumber      uint64
    Timestamp        uint64
}
```

The anchor chain tracker follows L1 block progression and injects anchor data into rollup blocks as system transactions.

### EXECUTE Precompile

The `execute` subpackage implements the `EXECUTE` precompile at address `0x0101`. It accepts `ExecuteInput`, invokes the rollup STF, and returns `ExecuteOutput`. The execution context tracks the rollup's chain ID, gas limits, and EIP-7701 AA support.

### Cross-Layer Proofs

The `proof` subpackage provides cross-layer proof generation and verification. `ProofData` carries a validity proof with public inputs, pre- and post-state roots, and a rollup identifier:

```go
type ProofData struct {
    RollupID      uint64
    Proof         []byte
    PublicInputs  []byte
    PreStateRoot  types.Hash
    PostStateRoot types.Hash
}
```

### Fraud Proof System

The `rollup` package implements an optimistic fraud proof mechanism. When an invalid state transition is detected, a dispute trace is generated identifying the divergent step, which can be submitted to L1 for slashing.

### Sequencer

The `Sequencer` collects L2 transactions, orders them, and assembles batches for submission to L1 via the registry. It maintains an ordered queue and handles re-ordering on L1 reorgs.

### State Bridge

`StateBridge` synchronizes L1 events (deposits, contract calls, governance updates) to L2. It maintains a priority queue of pending state updates and a sync mechanism that tracks L1 confirmation depth before finalizing updates on L2.

## Subpackages

| Subpackage | Description |
|------------|-------------|
| [`anchor/`](./anchor/) | Anchor predeploy contract and L1 chain tracker |
| [`bridge/`](./bridge/) | L1-L2 bridge for cross-layer message passing |
| [`execute/`](./execute/) | EXECUTE precompile implementation and execution context |
| [`proof/`](./proof/) | Cross-layer validity proof generation and verification |
| [`registry/`](./registry/) | Registry compatibility shims |
| [`sequencer/`](./sequencer/) | Rollup transaction ordering and batch assembly |

## Usage

```go
// Create a registry and register a rollup.
reg := rollup.NewRollupRegistry()
cfg := rollup.NativeRollupConfig{
    ID:               1,
    Name:             "my-rollup",
    GenesisStateRoot: genesisRoot,
    GasLimit:         30_000_000,
}
r, err := reg.RegisterRollup(cfg)

// Submit a batch.
result, err := reg.SubmitBatch(1, batchRLP, claimedPostRoot)
// result.PostStateRoot is the new canonical state root.

// Verify a state transition proof.
valid, err := reg.VerifyStateTransition(1, result.PreStateRoot, result.PostStateRoot, zkProof)

// Process a deposit from L1.
deposit, err := reg.ProcessDeposit(1, l1Sender, depositAmount)

// Process a withdrawal to L1.
withdrawal, err := reg.ProcessWithdrawal(1, l1Recipient, amount, withdrawalProof)
```

## Documentation References

- [Roadmap](../../docs/ROADMAP.md)
- [GAP Analysis](../../docs/GAP_ANALYSIS.md)
- EIP-8079: Native rollups — EXECUTE precompile and anchor state
- I+ roadmap: native rollups, zkVM framework, STF executor
