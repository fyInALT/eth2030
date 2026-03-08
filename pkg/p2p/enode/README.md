# Package enode

Ethereum node identification: `NodeID` (32-byte keccak256 of compressed secp256k1 pubkey), `Node` struct, and `enode://` URL parsing.

## Overview

The `enode` package defines the canonical identity for devp2p nodes. A `NodeID` is a 32-byte value derived from the keccak256 hash of the node's compressed public key. The `Node` struct bundles a `NodeID` with its TCP/UDP endpoints and an optional `enr.Record`. `ParseNode` decodes `enode://` URLs (supporting 32-byte raw IDs, 33-byte compressed keys, and 64/65-byte uncompressed keys). `Distance` and `DistCmp` implement the XOR metric used by Kademlia routing.

## Functionality

- `NodeID` — `[32]byte`
  - `ParseID(s string) (NodeID, error)` — hex with optional `0x` prefix
  - `HexID(s string) NodeID` — panics on invalid input
  - `(NodeID).String() string` / `IsZero() bool`

- `Node` — `{ID NodeID, IP net.IP, TCP uint16, UDP uint16, Record *enr.Record, Pubkey []byte}`
  - `NewNode(id NodeID, ip net.IP, tcp, udp uint16) *Node`
  - `(Node).String() string` — encodes as `enode://<id>@<ip>:<tcp>[?discport=<udp>]`
  - `(Node).Addr() net.UDPAddr` / `TCPAddr() net.TCPAddr`

- `ParseNode(rawurl string) (*Node, error)` — decodes an `enode://` URL

- `Distance(a, b NodeID) int` — XOR log distance (0 if equal)
- `DistCmp(target, a, b NodeID) int` — compares `a` vs `b` distance to target

## Usage

```go
node, err := enode.ParseNode("enode://abc123...@192.168.1.1:30303")
if err != nil {
    log.Fatal(err)
}
fmt.Println(node.ID, node.IP, node.TCP)

dist := enode.Distance(localID, remoteID)
```

[← p2p](../README.md)
