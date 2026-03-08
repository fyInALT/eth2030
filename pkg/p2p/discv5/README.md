# Package discv5

Ethereum Node Discovery Protocol V5 — Kademlia DHT with background table maintenance, iterative lookups, ping-based liveness checking, and stale-node eviction.

## Overview

The `discv5` package provides a self-contained implementation of the Discovery V5 Kademlia DHT. A `DiscV5` instance holds 256 k-buckets (one per XOR log-distance bit, capped at 16 entries each) and runs a background goroutine that periodically pings all known nodes, evicts nodes with three or more consecutive failures, and performs a random iterative lookup to discover new peers.

The `Transport` interface decouples network I/O from routing logic, allowing the implementation to be unit-tested without a live network. `New` wires together a configured instance; `Start` launches the background maintenance loop; `Stop` drains it and cleans up.

## Functionality

- `DiscV5` — main discovery instance
  - `New(selfID NodeID, cfg Config, transport Transport) *DiscV5`
  - `Start()` / `Stop()`
  - `AddNode(rec *NodeRecord) error`
  - `RemoveNode(id NodeID)`
  - `GetNode(id NodeID) *NodeRecord`
  - `Ping(node *NodeRecord) bool` — updates last-seen or increments failure counter
  - `FindNode(target *NodeRecord, distance int) []*NodeRecord` — queries via transport, inserts results
  - `RequestENR(node *NodeRecord) (*NodeRecord, error)`
  - `Lookup(target NodeID) []*NodeRecord` — iterative lookup returning up to BucketSize closest
  - `RandomLookup() []*NodeRecord`
  - `EvictStale(maxFailures int) int`
  - `ClosestNodes(target NodeID, count int) []*NodeRecord`
  - `NodesAtDistance(distance int) []*NodeRecord`
  - `Len() int` / `BucketLen(idx int) int`

- `Transport` interface — `Ping`, `FindNode`, `RequestENR`
- `NodeRecord` — `NodeID`, `IP`, `UDPPort`, `TCPPort`, `SeqNumber`, `ENRRecord`
- `Distance(a, b NodeID) int` — XOR log distance (0 if equal, 1–256 otherwise)
- `Config` — `BucketSize`, `MaxPeers`, `RefreshInterval`, `PingTimeout`

## Usage

```go
dv5 := discv5.New(selfID, discv5.Config{
    RefreshInterval: 30 * time.Second,
    PingTimeout:     5 * time.Second,
}, myTransport)
dv5.Start()
dv5.AddNode(&discv5.NodeRecord{NodeID: peerID, IP: ip, UDPPort: 30303})
closest := dv5.Lookup(targetID)
dv5.Stop()
```

[← p2p](../README.md)
