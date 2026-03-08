// subscriptions.go re-exports WebSocket subscription types from rpc/subscription.
package rpc

import (
	"github.com/eth2030/eth2030/rpc/ethapi"
	rpcsub "github.com/eth2030/eth2030/rpc/subscription"
)

// Re-export WebSocket notification types from rpc/subscription.
type (
	WSNotification       = rpcsub.WSNotification
	WSSubscriptionResult = rpcsub.WSSubscriptionResult
	Subscription         = rpcsub.Subscription
)

// Re-export notification helper.
var FormatWSNotification = rpcsub.FormatWSNotification

// SubType is a type alias for ethapi.SubType.
type SubType = ethapi.SubType

// Re-export SubType constants from ethapi.
const (
	SubNewHeads  SubType = ethapi.SubNewHeads
	SubLogs      SubType = ethapi.SubLogs
	SubPendingTx SubType = ethapi.SubPendingTx
)
