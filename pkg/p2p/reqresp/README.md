# Package reqresp

Request-response protocol abstraction: binary framing codec, request ID tracking, concurrent pending-request management, and a higher-level `RequestManager` for timeout-aware RPC over devp2p.

## Overview

The `reqresp` package provides the request-response layer that sits above raw devp2p message framing. `ReqRespCodec` serialises and deserialises requests and responses using a compact binary wire format: `method_len[2] || method || id[8] || payload_len[4] || payload` (responses append `err_len[2] || err`). Request IDs are monotonically incrementing `uint64` values; the codec tracks pending requests for correlation. `RequestManager` (in `request_manager.go`) adds timeout handling and callback dispatch. `RequestHandler` (in `request_handler.go`) is the server-side counterpart that routes incoming requests to registered method handlers.

## Functionality

### ReqRespCodec

- `NewReqRespCodec(config ReqRespConfig) *ReqRespCodec`
- `EncodeRequest(method string, payload []byte) (*Request, []byte, error)`
- `DecodeRequest(data []byte) (*Request, error)`
- `EncodeResponse(reqID RequestID, method string, payload []byte, errMsg string) ([]byte, error)`
- `DecodeResponse(data []byte) (*Response, error)`
- `PendingRequests() int`
- `DefaultReqRespConfig() ReqRespConfig` — max 1 MiB, 10 s timeout, 64 concurrent

### Supporting types

- `Request` — `{ID RequestID, Method string, Payload []byte, Timestamp time.Time}`
- `Response` — `{ID RequestID, Method string, Payload []byte, Error string, Timestamp time.Time}`
- Errors: `ErrRequestTooLarge`, `ErrInvalidEncoding`, `ErrMethodTooLong`

## Usage

```go
codec := reqresp.NewReqRespCodec(reqresp.DefaultReqRespConfig())

// client side
req, encoded, err := codec.EncodeRequest("GetBlockHeaders", payload)
transport.Write(encoded)

// server side
incoming, _ := codec.DecodeRequest(raw)
responseData := handleMethod(incoming.Method, incoming.Payload)
respEncoded, _ := codec.EncodeResponse(incoming.ID, incoming.Method, responseData, "")
transport.Write(respEncoded)
```

[← p2p](../README.md)
