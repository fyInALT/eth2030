# Package ethproto

ETH wire protocol message types for ETH/68 through ETH/71, including EIP-7975 partial receipt lists and EIP-8159 Block Access List exchange.

## Overview

The `ethproto` package defines the message structures, message code constants, and validation utilities for the devp2p `eth` sub-protocol. It covers ETH/68 (the baseline with typed transaction announcements), ETH/70 (EIP-7975 partial block receipt lists), and ETH/71 (EIP-8159 Block Access List exchange). The `handler.go` file provides the peer handler that performs the `Status` handshake and dispatches messages to registered callbacks. `forkid.go` defines the EIP-2124 fork identifier used in the `Status` message to detect chain incompatibility.

## Functionality

### Protocol versions

`ETH68 = 68`, `ETH70 = 70`, `ETH71 = 71`

### Message codes

`StatusMsg`, `NewBlockHashesMsg`, `TransactionsMsg`, `GetBlockHeadersMsg`, `BlockHeadersMsg`, `GetBlockBodiesMsg`, `BlockBodiesMsg`, `NewBlockMsg`, `NewPooledTransactionHashesMsg`, `GetPooledTransactionsMsg`, `PooledTransactionsMsg`, `GetReceiptsMsg`, `ReceiptsMsg`, `GetPartialReceiptsMsg` (ETH/70), `PartialReceiptsMsg` (ETH/70), `GetBlockAccessListsMsg` (ETH/71), `BlockAccessListsMsg` (ETH/71).

### Key types

- `StatusData` — `ProtocolVersion`, `NetworkID`, `TD`, `Head`, `Genesis`, `ForkID`
- `GetBlockHeadersRequest` / `GetBlockHeadersPacket` / `BlockHeadersPacket`
- `GetBlockBodiesPacket` / `BlockBodiesPacket` / `BlockBody`
- `NewBlockData` — `Block`, `TD`
- `NewPooledTransactionHashesPacket68` — `Types`, `Sizes`, `Hashes`
- `GetPooledTransactionsPacket` / `PooledTransactionsPacket`
- `GetReceiptsPacket` / `ReceiptsPacket`
- `GetPartialReceiptsPacket` / `PartialReceiptsPacket` — with Merkle proofs
- `GetBlockAccessListsPacket` / `BlockAccessListsPacket` / `BlockAccessListData` / `AccessEntryData`
- `ForkID` — `Hash [4]byte`, `Next uint64`

### Utilities

- `ValidateMessageCode(code uint64) error`
- `MessageName(code uint64) string`

## Usage

```go
// Compose and send a GetBlockHeaders request
req := &ethproto.GetBlockHeadersPacket{
    RequestID: 1,
    Request: ethproto.GetBlockHeadersRequest{
        Origin:  ethproto.HashOrNumber{Number: 1000},
        Amount:  16,
        Reverse: false,
    },
}
// encode with RLP and write over the wire transport
```

[← p2p](../README.md)
