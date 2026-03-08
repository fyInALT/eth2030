# rollup/registry — Native rollup registration and batch submission

## Overview

This package implements the on-chain registry for native rollups (EIP-8079). It
tracks all registered rollups by a unique numeric ID, maintains their current
verified state root and block count, and processes batch submissions that advance
the state. Deposits and withdrawals are also processed here; withdrawal proofs
are verified before a withdrawal is recorded.

Each `SubmitBatch` call derives the new state root deterministically as
`Keccak256(preState || batchData || claimedRoot)`, making the state progression
auditable. State transition verification uses a SHA-256 commitment check.
The `Registry` is thread-safe via an internal read-write mutex.

## Functionality

**Types**

- `RollupConfig` — `ID`, `Name`, `BridgeContract`, `GenesisStateRoot`, `GasLimit`
- `NativeRollup` — full rollup state: `StateRoot`, `LastBlock`, `TotalBatches`,
  `TotalDeposits`, `TotalWithdrawals`, `Deposits`, `Withdrawals`
- `Deposit` — `ID`, `RollupID`, `From`, `Amount`, `BlockNumber`, `Finalized`
- `Withdrawal` — `ID`, `RollupID`, `To`, `Amount`, `Proof`, `Verified`
- `BatchResult` — `RollupID`, `BatchHash`, `PreStateRoot`, `PostStateRoot`,
  `BlockNumber`

**`Registry` methods**

- `NewRegistry() *Registry`
- `RegisterRollup(config) (*NativeRollup, error)`
- `GetRollupState(rollupID) (*NativeRollup, error)` — returns a defensive copy
- `SubmitBatch(rollupID, batchData, stateRoot) (*BatchResult, error)`
- `VerifyStateTransition(rollupID, pre, post, proof) (bool, error)`
- `ProcessDeposit(rollupID, from, amount) (*Deposit, error)`
- `ProcessWithdrawal(rollupID, to, amount, proof) (*Withdrawal, error)`
- `Count() int` / `IDs() []uint64`

**Parent package:** [rollup](../)
