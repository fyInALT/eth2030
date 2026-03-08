# Package discover

Kademlia DHT routing table and iterative node lookup for Ethereum Discovery V5.

## Overview

The `discover` package implements the Kademlia-based routing table used by the Ethereum Discovery V5 protocol. A `Table` organises 256 k-buckets (one per XOR log-distance bit), each holding up to 16 entries with a replacement cache of up to 10 nodes. Node lookups are iterative: the closest known nodes are queried with an Alpha=3 concurrency factor, progressively converging toward the target.

The companion `v5.go` file provides a higher-level `DiscoveryV5` service that wires the table to a live UDP transport and runs bucket refresh via `Refresh`. A `KademliaTable` type (used by the Portal `DHTRouter`) exposes `FindClosest`, `AddNode`, `RecordFailure`, and `SelfID`.

## Functionality

- `Table` — Kademlia routing table
  - `NewTable(self enode.NodeID) *Table`
  - `AddNode(n *enode.Node)` — inserts or adds to replacement cache when bucket full
  - `RemoveNode(id enode.NodeID)` — evicts and promotes a replacement
  - `FindNode(target enode.NodeID, count int) []*enode.Node` — returns count closest nodes
  - `Lookup(target enode.NodeID, queryFn func(*enode.Node) []*enode.Node) []*enode.Node` — iterative lookup
  - `Refresh(queryFn ...)` — random-target lookup for table maintenance
  - `BucketIndex(id enode.NodeID) int` — XOR log-distance bucket index
  - `Nodes() []*enode.Node` / `Len() int` / `BucketEntries(idx int) []*enode.Node`

- `KademliaTable` — Portal-oriented variant exposing `FindClosest`, `AddNode`, `RecordFailure`, `SelfID`

- Constants: `BucketSize=16`, `NumBuckets=256`, `Alpha=3`, `MaxReplacements=10`

## Usage

```go
table := discover.NewTable(localID)
table.AddNode(enode.NewNode(peerID, ip, tcp, udp))

closest := table.Lookup(targetID, func(n *enode.Node) []*enode.Node {
    // send FINDNODE to n over UDP and return the results
    return queryPeer(n)
})
```

[← p2p](../README.md)
