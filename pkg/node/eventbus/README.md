# node/eventbus - Publish/subscribe event bus for node subsystems

## Overview

Package `eventbus` provides a lightweight publish/subscribe event bus used by
ETH2030 node subsystems to communicate without direct coupling. Subsystems
subscribe to named event types and receive typed `Event` values over Go channels.
The bus supports both blocking (`Publish`) and non-blocking (`PublishAsync`)
delivery, where the async variant silently drops events for subscribers whose
channels are full.

Ten standard event types cover the most common cross-subsystem signals: new
blocks, new transactions, chain head changes, peer connect/disconnect, sync
status, and tx-pool add/drop.

## Functionality

**Types**

- `EventType string` - named event identifier.
- `Event` - `Type EventType`, `Data interface{}`, `Timestamp time.Time`.
- `Subscription` - wraps a typed channel; obtain via `Chan() <-chan Event` and
  cancel via `Unsubscribe()` (idempotent).
- `EventBus` - main bus struct (RWMutex-protected subscription map).

**Predefined event types**

`EventNewBlock`, `EventNewTx`, `EventChainHead`, `EventChainSideHead`,
`EventNewPeer`, `EventDropPeer`, `EventSyncStarted`, `EventSyncCompleted`,
`EventTxPoolAdd`, `EventTxPoolDrop`

**Constructor**

- `NewEventBus(bufferSize int) *EventBus` - `bufferSize` is the channel buffer
  per subscription; 0 for unbuffered.

**Methods**

- `Subscribe(eventType EventType) *Subscription` - subscribe to a single type.
- `SubscribeMultiple(types ...EventType) *Subscription` - subscribe to several
  types with one channel.
- `Unsubscribe(sub *Subscription)` - remove and close; safe to call multiple times.
- `Publish(eventType EventType, data interface{})` - blocking fan-out to matching
  subscribers.
- `PublishAsync(eventType EventType, data interface{})` - non-blocking fan-out;
  drops for full channels.
- `SubscriberCount(eventType EventType) int`
- `Close()` - shuts down the bus and closes all subscriber channels.

Parent package: [`node`](../)
