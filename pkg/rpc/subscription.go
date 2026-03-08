// subscription.go re-exports subscription types from rpc/subscription and
// rpc/filter for backward compatibility.
package rpc

import (
	rpcfilter "github.com/eth2030/eth2030/rpc/filter"
	rpcsub "github.com/eth2030/eth2030/rpc/subscription"
)

// Re-export filter types from rpc/filter.
type (
	FilterType  = rpcfilter.FilterType
	FilterQuery = rpcfilter.FilterQuery
)

// Re-export FilterType constants.
const (
	LogFilter       = rpcfilter.LogFilter
	BlockFilter     = rpcfilter.BlockFilter
	PendingTxFilter = rpcfilter.PendingTxFilter
)

// MatchFilter re-exports rpcfilter.MatchFilter.
var MatchFilter = rpcfilter.MatchFilter

// FilterLogs re-exports rpcfilter.FilterLogs.
var FilterLogs = rpcfilter.FilterLogs

// FilterLogsWithBloom re-exports rpcfilter.FilterLogsWithBloom.
var FilterLogsWithBloom = rpcfilter.FilterLogsWithBloom

// SubscriptionManager re-exports rpcsub.SubscriptionManager.
type SubscriptionManager = rpcsub.SubscriptionManager

// NewSubscriptionManager wraps rpcsub.NewSubscriptionManager.
var NewSubscriptionManager = rpcsub.NewSubscriptionManager
