# queue

Deposit and withdrawal queue primitives for beacon chain processing.

## Overview

Package `queue` provides two independent thread-safe queues used by the beacon
chain state transition:

- **`DepositQueue`**: validates and batches EIP-6110 execution-layer deposit
  receipts for validator activation. Deposits are validated (pubkey length,
  amount bounds, signature and credential lengths), deduplicated by index, and
  a Keccak-256 Merkle root is computed over pending entries for on-chain
  commitment.

- **`WithdrawalQueue`**: manages EIP-4895 validator withdrawal requests with
  priority ordering (higher priority processed first), minimum withdrawal delay
  enforcement, and per-slot churn limits.

Both queues impose rate limits via configurable `MaxDepositsPerBlock` /
`MaxWithdrawalsPerSlot` caps.

## Functionality

### DepositQueue

| Name | Description |
|------|-------------|
| `DefaultDepositQueueConfig() DepositQueueConfig` | 16/block cap, 32 ETH min, 2048 ETH max (EIP-7251), mainnet contract address |
| `NewDepositQueue(config) *DepositQueue` | Create queue |
| `(*DepositQueue).AddDeposit(entry) error` | Validate and enqueue |
| `(*DepositQueue).ProcessDeposits(maxCount) []DepositEntry` | Dequeue up to `min(maxCount, MaxDepositsPerBlock)` deposits |
| `(*DepositQueue).GetDepositRoot() types.Hash` | Keccak-256 Merkle root of pending deposits |
| `(*DepositQueue).GetDepositCount() uint64` | Total deposits ever added |

### WithdrawalQueue

| Name | Description |
|------|-------------|
| `DefaultWithdrawalQueueConfig() WithdrawalQueueConfig` | 65536 cap, 16/slot, 256-slot delay, churn 8 |
| `NewWithdrawalQueue(config) *WithdrawalQueue` | Create queue |
| `(*WithdrawalQueue).Enqueue(request) error` | Add a withdrawal (sorted by priority, then slot, then index) |
| `(*WithdrawalQueue).ProcessSlot(slot) []WithdrawalRequest` | Process eligible withdrawals for a slot |
| `(*WithdrawalQueue).CancelWithdrawal(validatorIndex) bool` | Remove a pending request |
| `(*WithdrawalQueue).GetPosition(validatorIndex) int` | Queue position (0 = front) |
| `(*WithdrawalQueue).Stats() QueueStats` | Pending count, processed count, total Gwei |

## Usage

```go
import "github.com/eth2030/eth2030/consensus/queue"

dq := queue.NewDepositQueue(queue.DefaultDepositQueueConfig())
dq.AddDeposit(queue.DepositEntry{Index: 0, Pubkey: pubkey48, Amount: 32e9})
deposits := dq.ProcessDeposits(16)

wq := queue.NewWithdrawalQueue(queue.DefaultWithdrawalQueueConfig())
wq.Enqueue(queue.WithdrawalRequest{ValidatorIndex: 42, Amount: 32e9, ...})
processed := wq.ProcessSlot(currentSlot)
```

[← consensus](../README.md)
