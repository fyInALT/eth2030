# rollup/bridge — L1-L2 deposit and withdrawal bridge for native rollups

## Overview

This package manages the lifecycle of cross-layer value transfers for native
rollups. Deposits move ETH from L1 to L2 and require a configurable number of
L1 confirmation blocks before they are considered final. Withdrawals move ETH
from L2 to L1 and follow a three-phase flow: initiation on L2, proof submission
(optimistic or ZK), and finalization for L1 release.

All operations are keyed by a Keccak256-derived unique ID and guarded by a
`MaxPendingDeposits` limit to bound unbounded growth. The `Bridge` struct is
thread-safe via an internal mutex.

## Functionality

**Types**

- `Config` — `L1ContractAddr`, `L2ContractAddr`, `ConfirmationBlocks`,
  `MaxPendingDeposits`; `DefaultConfig()` returns sensible defaults
- `Deposit` — `ID`, `From`, `To`, `Amount`, `L1Block`, `Status`
- `Withdrawal` — `ID`, `From`, `To`, `Amount`, `ProofData`, `Status`
- Status constants: `StatusPending=0`, `StatusConfirmed=1`, `StatusFinalized=2`,
  `StatusProven=3`

**`Bridge` methods**

- `NewBridge(config) *Bridge`
- `Deposit(from, to, amount, l1Block) (*Deposit, error)` — initiates L1->L2 deposit
- `ConfirmDeposits(l1Block uint64) int` — confirms all deposits with sufficient
  L1 confirmations; returns count
- `InitiateWithdrawal(from, to, amount) (*Withdrawal, error)`
- `ProveWithdrawal(id, proofData) error` — attaches proof; moves to `StatusProven`
- `FinalizeWithdrawal(id) error` — finalizes a proven withdrawal
- `PendingDeposits() []*Deposit`
- `PendingWithdrawals() []*Withdrawal`

**Parent package:** [rollup](../)
