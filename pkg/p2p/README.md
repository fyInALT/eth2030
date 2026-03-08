# Package p2p

Peer-to-peer networking layer: devp2p server, peer management, discovery, gossip, snap protocol, Portal network, and anonymous transaction transport.

## Overview

The `p2p` package implements the full Ethereum peer-to-peer networking stack for ETH2030. At its core, `Server` manages TCP connections, performs the devp2p hello handshake, and runs protocol handlers via a `Multiplexer` that demultiplexes messages across concurrent sub-protocols. `PeerManager` tracks live peers and their transports; `NodeTable` holds the discovery node set.

The package is organized into focused subpackages. Discovery is handled by V5 Kademlia DHT (`discv5`), DNS-based discovery (`dnsdisc`), and ENR/enode utilities. The `gossip` subpackage implements the beacon chain pub/sub gossip protocol with topic-level scoring, peer banning, deduplication, and message validation. The `portal` subpackage implements the Portal network: a content-addressed DHT for history, state, and beacon chain data. The `snap` subpackage implements the Snap/1 state sync protocol. The `transport` subpackage provides pluggable anonymous transaction transports (Tor SOCKS5, Nym SOCKS5, simulated mixnet) for EIP-7702 SetCode broadcast and general privacy-preserving transaction propagation.

## Table of Contents

- [Functionality](#functionality)
- [Subpackages](#subpackages)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### Server

`Server` is the devp2p connection manager:

```go
NewServer(cfg Config) *Server
```

`Config` fields: `ListenAddr`, `MaxPeers`, `Protocols`, `StaticNodes`, `EnableRLPx`, `Name`, `NodeID`, `BootstrapNodes`, `DiscoveryPort`, `NAT`.

Key methods:
- `Start() error` — begins listening; loads static nodes; starts accept loop
- `Stop()` — shuts down listener and all connections
- `AddPeer(addr string)` — dials an address and adds it as a peer
- `PeerCount() int` / `PeersList() []*peermgr.Peer`
- `ListenAddr() net.Addr`
- `LocalID() string`
- `NodeTable() *NodeTable` / `Scores() *scoring.ScoreMap`

When `EnableRLPx` is set, connections are wrapped with RLPx encryption before the hello handshake. The `Multiplexer` runs each registered `Protocol` in its own goroutine, routing messages by code offset.

### Protocol

```go
type Protocol struct {
    Name    string
    Version uint
    Length  uint64 // message code count
    Run     func(peer *peermgr.Peer, t wire.Transport) error
}
```

`Protocol.Run` is invoked per-peer and should read/write typed messages over the transport until the peer disconnects.

### Peer Manager

`PeerManager` tracks live peers and their transports for message broadcasting:

- `AddPeer(peer, transport)` / `RemovePeer(id)`
- `GetPeer(id) *peermgr.Peer`
- `BroadcastNewBlock(block, td)` — sends `NewBlock` message to all peers
- `BroadcastTransactions(txs)` — fans out transactions to all peers

### Node Discovery

`NodeTable` implements the `NodeDiscovery` interface:
- `AllNodes()` / `StaticNodes()`
- `AddNode(n *Node)` / `Remove(id NodeID)`

`Node` and `NodeID` types are defined in `discovery.go`. `ParseEnode(url)` decodes an enode URL.

### Gossip Protocol

The `gossip` subpackage implements the beacon chain pub/sub gossip network:

- `TopicManager` — manages subscriptions to gossip topics with deduplication and per-topic scoring
- Topics: `BeaconBlock`, `BeaconAggregateAndProof`, `VoluntaryExit`, `ProposerSlashing`, `AttesterSlashing`, `BlobSidecar`, `SyncCommitteeContribution`, `STARKMempoolTick`, `PQAggRequest`, `PQAggResult`, `ProposerPreferences`
- `gossip_mesh_scoring.go` — mesh peer scoring with decay and cap parameters
- `gossip_v2.go` — v2 gossip with enhanced message validation
- `message_validator.go` — per-topic message validation pipeline

Package-level wrappers (`TopicManager`, `NewTopicManager`, `Subscribe`, `STARKMempoolTick`) are re-exported in the top-level `p2p` package for use by `node.go`.

### Anonymous Transport (Mixnet)

`TransportManager` selects and manages anonymous transaction transports:

```go
NewTransportManagerWithConfig(cfg TransportConfig) *TransportManager
```

Supported modes (`MixnetMode`):
- `ModeTorSocks5` — routes transactions via a Tor SOCKS5 proxy (`TorTransport`)
- `ModeNymSocks5` — routes transactions via the Nym mixnet SOCKS5 proxy (`NymTransport`)
- `ModeSimulated` — local simulated mixnet for testing (`MixnetTransport`)

`SelectBestTransport()` auto-probes Tor → Nym → simulated. `StartAll()` / `StopAll()` start/stop all registered transports.

`ParseMixnetMode(string)` converts CLI mode strings.

### ETH Wire Protocol

The `ethproto` subpackage implements ETH/66–ETH/72 wire protocol messages:
- Block headers, bodies, receipts, transactions
- EIP-8077 `AnnounceNonce` (ETH/72)
- EIP-7702 SetCode broadcast

### Portal Network

The `portal` subpackage implements the Portal network for light-weight data access:
- `ContentDB` — content-addressed key-value store
- `DHTRouter` — Kademlia routing for content lookup and storage
- `HistoryNetwork` — EIP-4444 historical block/receipt access
- `StateNetwork` — state trie data via Portal
- `EclipseResistance` — Kademlia eclipse attack mitigation

### Scoring

The `scoring` subpackage tracks per-peer reputation:
- `ScoreMap` — concurrent map of `PeerScore` by peer ID
- `PeerScore` — tracks `HandshakeOK`, `GoodResponse`, `BadResponse` with decay
- Scores are used by `Server.setupConn` to gate peer acceptance

### Snap Protocol

The `snap` subpackage implements the Snap/1 state sync protocol for downloading state trie data efficiently during snap sync.

### Dispatch, ReqResp, Wire

- `dispatch/` — request dispatching and routing layer
- `reqresp/` — request-response protocol abstraction
- `wire/` — low-level framing (`FrameTransport`), RLPx encrypted transport, `HelloPacket`, `TCPDialer`, `TCPListener`, `Msg` type

## Subpackages

| Subpackage | Description |
|------------|-------------|
| [`broadcast/`](./broadcast/) | Block and transaction broadcast to connected peers |
| [`discover/`](./discover/) | V4/V5 node discovery |
| [`discv5/`](./discv5/) | Discovery V5 Kademlia DHT implementation |
| [`dispatch/`](./dispatch/) | Protocol message dispatching |
| [`dnsdisc/`](./dnsdisc/) | DNS-based node discovery (EIP-1459) |
| [`enode/`](./enode/) | Enode URL parsing and encoding |
| [`enr/`](./enr/) | Ethereum Node Record (ENR, EIP-778) encoding |
| [`ethproto/`](./ethproto/) | ETH/66–ETH/72 wire protocol messages |
| [`gossip/`](./gossip/) | Beacon chain pub/sub gossip with topic scoring and deduplication |
| [`nat/`](./nat/) | NAT traversal (UPnP, extip, auto-detect) |
| [`netutil/`](./netutil/) | Network utility functions |
| [`nonce/`](./nonce/) | EIP-8077 announce nonce support |
| [`peermgr/`](./peermgr/) | Low-level peer set (`ManagedPeerSet`, `Peer`) |
| [`portal/`](./portal/) | Portal network: content DHT for history, state, and beacon data |
| [`reqresp/`](./reqresp/) | Request-response protocol abstraction |
| [`scoring/`](./scoring/) | Per-peer reputation scoring with decay |
| [`snap/`](./snap/) | Snap/1 state sync protocol |
| [`transport/`](./transport/) | Anonymous transaction transports: Tor, Nym, simulated mixnet |
| [`wire/`](./wire/) | devp2p framing, RLPx encryption, hello handshake, TCP dial/listen |

## Usage

```go
// Create and start a P2P server with the ETH protocol
srv := p2p.NewServer(p2p.Config{
    ListenAddr:     ":30303",
    MaxPeers:       50,
    BootstrapNodes: "enode://...",
    Protocols: []p2p.Protocol{{
        Name:    "eth",
        Version: 68,
        Length:  17,
        Run:     myETHProtocolHandler,
    }},
})
if err := srv.Start(); err != nil {
    log.Fatal(err)
}

// Subscribe to a gossip topic
topicMgr := p2p.NewTopicManager(p2p.DefaultTopicParams())
topicMgr.Subscribe(p2p.STARKMempoolTick, func(_ p2p.GossipTopic, _ p2p.MessageID, data []byte) {
    // handle incoming STARK mempool tick
})

// Create an anonymous transport manager (auto-probe Tor -> Nym -> simulated)
tmCfg := p2p.DefaultTransportConfig()
mgr := p2p.NewTransportManagerWithConfig(tmCfg)
mgr.SelectBestTransport()
```

## Documentation References

- [devp2p specification](https://github.com/ethereum/devp2p)
- [Discovery V5 specification](https://github.com/ethereum/devp2p/blob/master/discv5/discv5.md)
- [Portal Network specification](https://github.com/ethereum/portal-network-specs)
- [EIP-8077: Announce Nonce](https://eips.ethereum.org/EIPS/eip-8077)
- ETH2030 gossip integration: `pkg/node/node.go`
- ETH2030 sync: `pkg/sync/`
