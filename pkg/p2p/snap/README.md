# Package snap

Snap/1 state synchronization protocol — message types and handler interface for downloading account ranges, storage ranges, bytecodes, and trie nodes during snap sync.

## Overview

The `snap` package defines the wire protocol for Ethereum's Snap synchronization protocol (snap/1). It provides the request and response packet structures for all four Snap message pairs, enforces soft (500 KB) and hard (2 MB) response size limits, and defines the `Handler` interface that a server must implement to serve state data. The `handler.go` file provides the client-side handler that drives snap sync by dispatching incoming responses.

## Functionality

### Protocol constants

`ProtocolName = "snap"`, `ProtocolVersion = 1`

`SoftResponseLimit = 500 KB`, `HardResponseLimit = 2 MB`

`MaxAccountRangeCount = 256`, `MaxStorageRangeCount = 512`, `MaxByteCodeCount = 64`, `MaxTrieNodeCount = 512`

### Message codes

`GetAccountRangeMsg = 0x00`, `AccountRangeMsg = 0x01`, `GetStorageRangesMsg = 0x02`, `StorageRangesMsg = 0x03`, `GetByteCodesMsg = 0x04`, `ByteCodesMsg = 0x05`, `GetTrieNodesMsg = 0x06`, `TrieNodesMsg = 0x07`

### Request/response types

- `GetAccountRangePacket` / `AccountRangePacket` — account hash range with Merkle boundary proof
- `GetStorageRangesPacket` / `StorageRangesPacket` — storage slots for a set of accounts
- `GetByteCodesPacket` / `ByteCodesPacket` — contract bytecode by code hash
- `GetTrieNodesPacket` / `TrieNodesPacket` — arbitrary trie nodes by path

### Handler interface

```go
type Handler interface {
    HandleGetAccountRange(req *GetAccountRangePacket) (*AccountRangePacket, error)
    HandleGetStorageRanges(req *GetStorageRangesPacket) (*StorageRangesPacket, error)
    HandleGetByteCodes(req *GetByteCodesPacket) (*ByteCodesPacket, error)
    HandleGetTrieNodes(req *GetTrieNodesPacket) (*TrieNodesPacket, error)
}
```

## Usage

```go
// Implement the Handler to serve snap sync requests
type myHandler struct{ db StateDatabase }

func (h *myHandler) HandleGetAccountRange(req *snap.GetAccountRangePacket) (*snap.AccountRangePacket, error) {
    accounts := h.db.AccountRange(req.Root, req.Origin, req.Limit, req.Bytes)
    return &snap.AccountRangePacket{ID: req.ID, Accounts: accounts}, nil
}
```

[← p2p](../README.md)
