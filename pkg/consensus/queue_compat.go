package consensus

// queue_compat.go re-exports types from consensus/queue for backward compatibility.

import "github.com/eth2030/eth2030/consensus/queue"

// Queue type aliases.
type (
	DepositQueueConfig    = queue.DepositQueueConfig
	DepositEntry          = queue.DepositEntry
	DepositQueue          = queue.DepositQueue
	WithdrawalQueueConfig = queue.WithdrawalQueueConfig
	WithdrawalRequest     = queue.WithdrawalRequest
	QueueStats            = queue.QueueStats
	WithdrawalQueue       = queue.WithdrawalQueue
)

// Queue constants.
const (
	DepositPubkeyLen          = queue.DepositPubkeyLen
	DepositSigLen             = queue.DepositSigLen
	DepositWithdrawalCredsLen = queue.DepositWithdrawalCredsLen
)

// Queue error variables.
var (
	ErrDepositQueueBelowMinimum   = queue.ErrDepositQueueBelowMinimum
	ErrDepositQueueAboveMax       = queue.ErrDepositQueueAboveMax
	ErrDepositQueueInvalidPubkey  = queue.ErrDepositQueueInvalidPubkey
	ErrDepositQueueEmptyPubkey    = queue.ErrDepositQueueEmptyPubkey
	ErrDepositQueueInvalidSig     = queue.ErrDepositQueueInvalidSig
	ErrDepositQueueInvalidCreds   = queue.ErrDepositQueueInvalidCreds
	ErrDepositQueueDuplicateIndex = queue.ErrDepositQueueDuplicateIndex
	ErrDepositQueueZeroAmount     = queue.ErrDepositQueueZeroAmount
	ErrWithdrawalQueueFull        = queue.ErrWithdrawalQueueFull
	ErrWithdrawalAlreadyQueued    = queue.ErrWithdrawalAlreadyQueued
)

// Queue function wrappers.
func DefaultDepositQueueConfig() DepositQueueConfig {
	return queue.DefaultDepositQueueConfig()
}
func NewDepositQueue(config DepositQueueConfig) *DepositQueue {
	return queue.NewDepositQueue(config)
}
func DefaultWithdrawalQueueConfig() WithdrawalQueueConfig {
	return queue.DefaultWithdrawalQueueConfig()
}
func NewWithdrawalQueue(config WithdrawalQueueConfig) *WithdrawalQueue {
	return queue.NewWithdrawalQueue(config)
}
