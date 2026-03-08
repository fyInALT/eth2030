// filter_event_hub.go re-exports EventHub types from rpc/filter for
// backward compatibility.
package rpc

import (
	rpcfilter "github.com/eth2030/eth2030/rpc/filter"
)

// Re-export EventHub errors.
var (
	ErrHubClosed          = rpcfilter.ErrHubClosed
	ErrHubListenerFull    = rpcfilter.ErrHubListenerFull
	ErrHubInvalidListener = rpcfilter.ErrHubInvalidListener
	ErrHubDuplicateID     = rpcfilter.ErrHubDuplicateID
)

// Re-export event hub types.
type (
	ChainEventType = rpcfilter.ChainEventType
	ChainEvent     = rpcfilter.ChainEvent
	EventHubConfig = rpcfilter.EventHubConfig
	EventHub       = rpcfilter.EventHub
	EventHubStats  = rpcfilter.EventHubStats
	EventListener  = rpcfilter.EventListener
)

// Re-export chain event type constants.
const (
	EventNewBlock      = rpcfilter.EventNewBlock
	EventNewLogs       = rpcfilter.EventNewLogs
	EventPendingTx     = rpcfilter.EventPendingTx
	EventReorg         = rpcfilter.EventReorg
	EventFilterExpired = rpcfilter.EventFilterExpired
)

// Re-export constructors.
var (
	DefaultEventHubConfig = rpcfilter.DefaultEventHubConfig
	NewEventHub           = rpcfilter.NewEventHub
)
