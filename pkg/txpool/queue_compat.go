package txpool

// queue_compat.go re-exports types from txpool/queue for backward compatibility.

import (
	"math/big"

	"github.com/eth2030/eth2030/core/types"
	"github.com/eth2030/eth2030/txpool/queue"
)

// Queue type aliases.
type (
	QueueManagerConfig = queue.QueueManagerConfig
	QueueManager       = queue.QueueManager
	PendingList        = queue.PendingList
)

// Queue constants.
const (
	DefaultQueueCapPerAccount = queue.DefaultQueueCapPerAccount
	DefaultGlobalQueueCap     = queue.DefaultGlobalQueueCap
)

// Queue function wrappers.
func NewQueueManager(config QueueManagerConfig, baseFee *big.Int) *QueueManager {
	return queue.NewQueueManager(config, baseFee)
}
func NewPendingList(baseFee *big.Int) *PendingList { return queue.NewPendingList(baseFee) }

// Ensure types.Address is used to avoid import cycle.
var _ = types.Address{}
