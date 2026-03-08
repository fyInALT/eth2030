# Package dispatch

Versioned protocol message dispatching with capability negotiation, per-peer rate limiting, request-response correlation, and a prioritised outbound queue.

## Overview

The `dispatch` package provides two complementary routing primitives. `ProtoDispatcher` handles multi-version protocol fanout: handlers are registered per `(version, message-code)` pair, version capabilities are exchanged during the devp2p handshake, and `NegotiateVersion` picks the highest mutually supported version. `MessageRouter` is a lower-level message bus that demultiplexes incoming messages by code, enforces per-peer token-bucket rate limits, correlates requests with responses via an embedded request ID, and manages a heap-based priority outbound queue (priorities: `PriorityHigh=0`, `PriorityNormal=1`, `PriorityLow=2`).

## Functionality

### ProtoDispatcher

- `NewProtoDispatcher(name string) *ProtoDispatcher`
- `RegisterVersion(spec ProtoVersionSpec) error`
- `RegisterHandler(version uint, code uint64, handler ProtoMsgHandler) error`
- `SetHandler(version uint, code uint64, handler ProtoMsgHandler) error`
- `Route(peerID string, version uint, code uint64, payload []byte) error`
- `RouteWithFallback(...)` — falls back to the highest version with a matching handler
- `NegotiateVersion(remoteCaps []Capability) (uint, error)`
- `SupportedVersions() []uint` / `HighestVersion() uint`
- `Capabilities() []Capability`

### MessageRouter

- `NewMessageRouter(cfg RouterConfig) *MessageRouter`
- `RegisterHandler(code uint64, handler RouterHandler) error`
- `SetHandler(code uint64, handler RouterHandler)`
- `Dispatch(peerID string, msg wire.Msg) error`
- `SendRequest(transport wire.Transport, requestCode, responseCode uint64, payload []byte, peerID string) (wire.Msg, error)`
- `Enqueue(msg wire.Msg, peerID string, priority int) error`
- `Dequeue() (*OutboundMsg, error)` / `DequeueNonBlocking() *OutboundMsg`
- `TrackPeer(peerID string)` / `UntrackPeer(peerID string)`
- `ExpireRequests(timeout time.Duration) int`
- `Stats() (dispatched, dropped, rateLimited, sent uint64)`

## Usage

```go
d := dispatch.NewProtoDispatcher("eth")
d.RegisterVersion(dispatch.ProtoVersionSpec{Name: "eth", Version: 68, MaxMsgCode: 0x14})
d.RegisterHandler(68, ethproto.NewBlockMsg, func(peerID string, code uint64, payload []byte) error {
    // handle new block
    return nil
})
negotiated, _ := d.NegotiateVersion(remoteCaps)
d.Route(peerID, negotiated, msgCode, payload)
```

[← p2p](../README.md)
