package node

// events_compat.go re-exports types from node/eventbus for backward compatibility.

import "github.com/eth2030/eth2030/node/eventbus"

// Type aliases.
type (
	EventType    = eventbus.EventType
	Event        = eventbus.Event
	Subscription = eventbus.Subscription
	EventBus     = eventbus.EventBus
)

// Event type constants.
const (
	EventNewBlock      = eventbus.EventNewBlock
	EventNewTx         = eventbus.EventNewTx
	EventChainHead     = eventbus.EventChainHead
	EventChainSideHead = eventbus.EventChainSideHead
	EventNewPeer       = eventbus.EventNewPeer
	EventDropPeer      = eventbus.EventDropPeer
	EventSyncStarted   = eventbus.EventSyncStarted
	EventSyncCompleted = eventbus.EventSyncCompleted
	EventTxPoolAdd     = eventbus.EventTxPoolAdd
	EventTxPoolDrop    = eventbus.EventTxPoolDrop
)

// NewEventBus creates a new EventBus.
func NewEventBus(bufferSize int) *EventBus { return eventbus.NewEventBus(bufferSize) }
