# types — JSON-RPC 2.0 protocol types

[← rpc](../README.md)

## Overview

Package `rpctypes` defines the foundational JSON-RPC 2.0 wire types shared
across all RPC sub-packages. It provides `Request`, `Response`, `RPCError`,
and `BlockNumber` with their JSON marshaling, together with a set of encoding
helpers for hex addresses, hashes, and integers used throughout the API layer.

## Functionality

**Protocol types**

- `Request` — `JSONRPC string`, `Method string`, `Params []json.RawMessage`, `ID json.RawMessage`
- `Response` — `JSONRPC string`, `Result interface{}`, `Error *RPCError`, `ID json.RawMessage`
- `RPCError` — `Code int`, `Message string`; implements `error`

**Block number**

- `BlockNumber int64` — special sentinel values: `LatestBlockNumber=-1`, `PendingBlockNumber=-2`, `EarliestBlockNumber=0`, `SafeBlockNumber=-3`, `FinalizedBlockNumber=-4`
- `UnmarshalJSON` — accepts `"latest"`, `"pending"`, `"earliest"`, `"safe"`, `"finalized"`, or hex/decimal integers

**Standard error codes**

- `ErrCodeParseError = -32700`, `ErrCodeInvalidRequest = -32600`, `ErrCodeMethodNotFound = -32601`, `ErrCodeInvalidParams = -32602`, `ErrCodeInternal = -32603`

**Constructor helpers**

- `NewSuccessResponse(id json.RawMessage, result interface{}) *Response`
- `NewErrorResponse(id json.RawMessage, code int, message string) *Response`

**Encoding helpers**

- `EncodeHash(h types.Hash) string` — `0x`-prefixed hex
- `EncodeAddress(a types.Address) string` — `0x`-prefixed hex
- `EncodeUint64(n uint64) string` — `0x`-prefixed hex
- `EncodeBigInt(n *big.Int) string` — `0x`-prefixed hex
- `ParseHexUint64(s string) uint64` — tolerant hex parser
