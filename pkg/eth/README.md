# Package eth

ETH wire protocol handler (eth/68–eth/72) connecting P2P networking to the blockchain and transaction pool.

## Overview

The `eth` package implements the Ethereum execution-layer wire protocol, currently at version eth/68, with extensions up to eth/72. It translates between raw P2P messages and the local blockchain/transaction pool interfaces, managing peer handshakes, block and transaction propagation, and the snap sync protocol.

The protocol evolution tracked in this package reflects the ETH2030 roadmap additions:

- **eth/70** (EIP-7975): Partial Block Receipt Lists — allows peers to request individual receipts by transaction index rather than fetching entire receipt trie
- **eth/71** (EIP-8159): Block Access List Exchange — propagates EIP-7928 BAL entries alongside blocks
- **eth/72** (EIP-8077): Announce Nonce — extends `NewPooledTransactionHashes` with sender address and nonce, enabling smarter fetch decisions

The `Handler` struct is the core message dispatcher. It is backed by `Blockchain` and `TxPool` interfaces, making it easy to wire into different EL implementations. An optional `SyncNotifier` callback triggers the sync subsystem when peers announce blocks above the local head.

## Table of Contents

- [Functionality](#functionality)
- [Subpackages](#subpackages)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### Protocol Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `ETH68` | 68 | Base eth protocol version |
| `ETH70` | 70 | EIP-7975: Partial Block Receipt Lists |
| `ETH71` | 71 | EIP-8159: Block Access List Exchange |
| `ETH72` | 72 | EIP-8077: Announce Nonce |
| `MaxHeaders` | 1024 | Maximum headers per response |
| `MaxBodies` | 512 | Maximum block bodies per response |
| `MaxPartialReceipts` | 256 | Maximum receipt indices per partial request |
| `MaxAccessLists` | 64 | Maximum BAL entries per response |

### Handler

`Handler` is the main entry point for eth protocol message dispatch:

```go
h := eth.NewHandler(chain, txPool, networkID)
h.SetSyncNotifier(syncMgr)
h.SetDownloader(downloader)
h.SetReceiptProvider(receiptDB)
h.SetAccessListProvider(balDB)
proto := h.Protocol() // register with p2p.Server
```

`Handler.Protocol()` returns a `p2p.Protocol` descriptor bound to `ETH68` that the P2P server calls for each new connection.

For each peer, the handler:
1. Performs the status handshake (protocol version, network ID, total difficulty, head hash, genesis hash)
2. Registers the peer in the `PeerSet`
3. Enters the message dispatch loop

### Message Dispatch

Messages handled by `handleMsg`:

| Message | Handler | Description |
|---------|---------|-------------|
| `GetBlockHeadersMsg` | `handleGetBlockHeaders` | Serve up to `MaxHeaders` headers by hash or number |
| `BlockHeadersMsg` | `handleBlockHeaders` | Process received headers response |
| `GetBlockBodiesMsg` | `handleGetBlockBodies` | Serve up to `MaxBodies` block bodies |
| `BlockBodiesMsg` | `handleBlockBodies` | Process received bodies response |
| `NewBlockHashesMsg` | `handleNewBlockHashes` | Log unknown block hashes for fetching |
| `NewBlockMsg` | `handleNewBlock` | Insert a full announced block, notify sync |
| `TransactionsMsg` | `handleTransactions` | Add received transactions to the pool |
| `GetPartialReceiptsMsg` | `handleGetPartialReceipts` | Serve partial receipts by tx index (eth/70) |
| `PartialReceiptsMsg` | `handlePartialReceipts` | Process received partial receipts |
| `GetBlockAccessListsMsg` | `handleGetBlockAccessLists` | Serve BAL entries for blocks (eth/71) |
| `BlockAccessListsMsg` | `handleBlockAccessLists` | Process received BAL entries |

### Exported Query Methods

These are exported to allow sync adapters and tests to call the handler directly:

- `HandleGetBlockHeaders(origin, amount, skip, reverse)` — collect headers from the local chain
- `HandleGetBlockBodies(hashes)` — retrieve bodies by hash
- `HandleNewBlock(peerID, block, td)` — process a block announcement and trigger sync if needed

### StatusInfo

`StatusInfo` carries the local chain's status for the handshake. It includes an `OldestBlock` field per EIP-4444 (history expiry) to advertise the lowest block number this node can serve, allowing peers to find alternative sources for older history.

### EIP-8077: Announce Nonce (ETH/72)

The `announce_nonce.go` file implements the extended transaction announcement message:

**`AnnounceNonceMsg`** extends the eth/68 `NewPooledTransactionHashes` format with two additional parallel arrays:
- `Sources []types.Address` — sender address for each announced transaction
- `Nonces []uint64` — sender nonce for each announced transaction

Wire format (RLP):
```
[txtypes: B, [sizes...], [hashes...], [sources...], [nonces...]]
```

Functions:
- `EncodeAnnounceNonce(msg *AnnounceNonceMsg) ([]byte, error)` — RLP encode
- `DecodeAnnounceNonce(data []byte) (*AnnounceNonceMsg, error)` — RLP decode
- `ProcessAnnounceMsg(tracker, msg) (int, error)` — record all announcements in a `NonceTracker`
- `ValidateAnnouncedNonce(tracker, sender, nonce, hash) error` — verify a specific announcement

**`NonceTracker`** is a thread-safe in-memory tracker of announced sender+nonce pairs (5-minute TTL):

- `Announce(sender, nonce, txHash) bool` — records an announcement; returns false on duplicate, handles RBF replacement
- `IsKnown(sender, nonce) bool` — predicate for deduplication
- `GetPending(sender) map[uint64]types.Hash` — all pending announcements for a sender
- `Remove(sender, nonce)` — remove after fetching/including
- `ExpireOld() int` — prune entries older than 5 minutes
- `Len() int` — total tracked announcement count

### Snap Protocol Handler

`snap_handler.go` integrates the snap sync protocol (Snap/1) for state trie data:

- Handles `GetAccountRange`, `AccountRange`, `GetStorageRanges`, `StorageRanges`, `GetByteCodes`, `ByteCodes`, `GetTrieNodes`, `TrieNodes`

### Block Fetcher

`block_fetcher.go` and `block_download.go` implement the block announcement fetcher that converts hash announcements into full block fetches:

- Tracks announced but not-yet-fetched hashes
- Deduplicates concurrent fetch requests from multiple peers
- Rate-limits fetches to avoid overwhelming a single peer

### EthPeer

`peer.go` and `peer_set.go` define:

- `EthPeer` — wraps a `p2p.Peer` and its `Transport`, adding typed send helpers (`SendBlockHeaders`, `SendBlockBodies`, `SendTransactions`, `SendPartialReceipts`, `SendBlockAccessLists`)
- `PeerSet` — thread-safe map of active peers keyed by peer ID

### PeerFetcher

`PeerFetcher` adapts the handler to the `sync.HeaderFetcher` and `sync.BodyFetcher` interfaces:

```go
pf := eth.NewPeerFetcher(chain)
headers, err := pf.FetchHeaders(fromBlock, count)
bodies, err := pf.FetchBodies(hashes)
```

## Subpackages

| Subpackage | Description |
|------------|-------------|
| [`ethversion/`](./ethversion/) | Protocol version negotiation and compatibility helpers |

## Usage

```go
import "github.com/eth2030/eth2030/eth"

// Create a handler with blockchain and tx pool backends.
handler := eth.NewHandler(blockchain, txPool, networkID)

// Wire in optional components.
handler.SetSyncNotifier(syncManager)
handler.SetDownloader(downloader)
handler.SetReceiptProvider(receiptDB)      // enables eth/70
handler.SetAccessListProvider(balStorage)  // enables eth/71

// Register the protocol with the P2P server.
proto := handler.Protocol()
p2pServer.Protocols = append(p2pServer.Protocols, proto)

// ETH/72: create a nonce tracker and process announcements.
tracker := eth.NewNonceTracker()
count, err := eth.ProcessAnnounceMsg(tracker, announceMsg)

// Validate a specific announced transaction.
if err := eth.ValidateAnnouncedNonce(tracker, sender, nonce, txHash); err != nil {
    log.Printf("stale announcement: %v", err)
}

// Clean up old entries periodically.
expired := tracker.ExpireOld()
```

## Documentation References

- [Design Doc](../../docs/DESIGN.md)
- [Roadmap Deep-Dive](../../docs/ROADMAP-DEEP-DIVE.md)
- [EIP-8077: Announce Nonce](https://eips.ethereum.org/EIPS/eip-8077)
- [EIP-7975: Partial Block Receipt Lists](https://eips.ethereum.org/EIPS/eip-7975)
- [EIP-4444: Bound Historical Data in Execution Clients](https://eips.ethereum.org/EIPS/eip-4444)
