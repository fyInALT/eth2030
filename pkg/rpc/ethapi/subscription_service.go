package ethapi

import (
	"github.com/eth2030/eth2030/core/types"
	rpcfilter "github.com/eth2030/eth2030/rpc/filter"
)

// FilterType and FilterQuery are re-exported from rpc/filter.
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

// SubscriptionService is the interface that EthAPI uses to manage filters
// and subscriptions. SubscriptionManager in the top-level rpc package
// satisfies this interface.
type SubscriptionService interface {
	// Subscribe creates a WebSocket subscription and returns its ID.
	Subscribe(subType SubType, query FilterQuery) string
	// Unsubscribe removes a subscription by ID.
	Unsubscribe(id string) bool
	// NewLogFilter installs a log filter and returns its ID.
	NewLogFilter(query FilterQuery) string
	// NewBlockFilter installs a block filter and returns its ID.
	NewBlockFilter() string
	// NewPendingTxFilter installs a pending tx filter and returns its ID.
	NewPendingTxFilter() string
	// GetFilterChanges returns new results since the last poll.
	GetFilterChanges(id string) (interface{}, bool)
	// GetFilterLogs returns all logs matching a log filter.
	GetFilterLogs(id string) ([]*types.Log, bool)
	// Uninstall removes a filter by ID.
	Uninstall(id string) bool
}

// SubType distinguishes the kind of WebSocket subscription.
type SubType int

const (
	// SubNewHeads watches for new block headers.
	SubNewHeads SubType = iota
	// SubLogs watches for matching log events.
	SubLogs
	// SubPendingTx watches for new pending transaction hashes.
	SubPendingTx
)
