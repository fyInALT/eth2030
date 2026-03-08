# rlp

Recursive Length Prefix (RLP) encoding and decoding for Ethereum data structures.

## Overview

The `rlp` package implements the RLP serialization format used throughout the Ethereum execution layer for encoding transactions, blocks, receipts, and other wire-protocol data. RLP is a minimal, deterministic encoding that maps arbitrary nested byte arrays and lists onto a compact binary representation.

The package provides both a reflection-based general encoder/decoder suitable for arbitrary Go types, and a set of zero-allocation fast-path helpers for high-throughput scenarios such as transaction pool serialization and block batch encoding. A pooled encoder (`EncoderPool`) reduces GC pressure in hot paths by recycling internal buffers via `sync.Pool`.

The decoder exposes both a one-shot `DecodeBytes` function and a stateful `Stream` reader for incremental parsing of RLP-encoded data, supporting the list-scoped reading pattern used by Ethereum wire protocol handlers.

## Table of Contents

- [Functionality](#functionality)
- [Subpackages](#subpackages)
- [Usage](#usage)
- [Documentation References](#documentation-references)

## Functionality

### Encoding

`Encode(w io.Writer, val interface{}) error` writes an RLP-encoded value to a writer.

`EncodeToBytes(val interface{}) ([]byte, error)` returns the encoded bytes directly.

Supported types:
- `bool` — `0x01` (true) / `0x80` (false)
- `uint8`, `uint16`, `uint32`, `uint64`, integers — compact big-endian encoding
- `*big.Int` — no-leading-zero big-endian byte string
- `[]byte`, `[N]byte` — RLP string (byte array)
- `string` — RLP string
- slice / array of non-byte element type — RLP list
- struct (exported fields) — RLP list of fields in declaration order
- nil pointer / interface — `0x80` (empty string sentinel)

`WrapList(payload []byte) []byte` wraps an already-encoded RLP payload in a list header.

### Zero-Allocation Fast Paths

Standalone functions for common types, avoiding reflection overhead:

| Function | Description |
|----------|-------------|
| `EncodeUint64(v uint64) []byte` | Encode a uint64 |
| `EncodeBytes32(data [32]byte) []byte` | Encode a 32-byte hash/key (always 33 bytes output) |
| `EncodeBytes20(data [20]byte) []byte` | Encode a 20-byte address (always 21 bytes output) |
| `EncodeBool(v bool) []byte` | Encode a boolean |
| `AppendUint64(dst []byte, v uint64) []byte` | Append encoded uint64 to a slice |
| `AppendBytes(dst, data []byte) []byte` | Append encoded byte string |
| `AppendListHeader(dst []byte, payloadSize int) []byte` | Append a list header prefix |

Size estimation helpers: `EstimateListSize(payloadSize int) int` and `EstimateStringSize(dataLen int) int`.

### Encoder Pool

`EncoderPool` wraps `sync.Pool` to recycle encoding buffers:

```go
type EncoderPool struct { ... }

func NewEncoderPool() *EncoderPool
func (ep *EncoderPool) EncodeBytes(val interface{}) ([]byte, error)
func (ep *EncoderPool) EncodeBatch(items []interface{}) ([]byte, error)
func (ep *EncoderPool) Metrics() *EncoderMetrics
```

`EncodeBatch` encodes a slice of items and wraps them in a single RLP list — useful for transaction lists and log lists.

`EncoderMetrics` tracks pool hits/misses, total encode count, and total bytes produced via atomic counters. `Snapshot()` returns a frozen copy.

Buffer sizing: default 4 KiB, capped at 1 MiB before being discarded to avoid retaining oversized buffers.

### Decoding

`DecodeBytes(b []byte, val interface{}) error` — one-shot decode into a pointer.

`Decode(r io.Reader, val interface{}) error` — reads all bytes from the reader then decodes.

`Stream` — stateful decoder with list-scoped reading:

```go
type Stream struct { ... }

func NewStream(r io.Reader) *Stream
func NewStreamFromBytes(data []byte) *Stream
func (s *Stream) Kind() (Kind, uint64, error)   // peek type and payload size
func (s *Stream) Bytes() ([]byte, error)         // read a string/byte value
func (s *Stream) Uint64() (uint64, error)        // read an unsigned integer
func (s *Stream) BigInt() (*big.Int, error)      // read a big integer
func (s *Stream) List() (uint64, error)          // enter a list scope
func (s *Stream) ListEnd() error                 // exit and verify list scope
func (s *Stream) AtListEnd() bool                // check if at list boundary
func (s *Stream) RawItem() ([]byte, error)       // read raw RLP bytes including prefix
```

`Kind` constants: `Byte`, `String`, `List`.

### Error Handling

Canonical encoding violations are detected and returned as typed errors (see the `rlperrors` subpackage):

- `ErrCanonSize` — single-byte string encoded with long-string prefix
- `ErrCanonInt` — integer with leading zero byte
- `ErrNonCanonicalSize` — short-string encoded with long-string prefix
- `ErrExpectedString`, `ErrExpectedList` — type mismatch
- `ErrEOL` — list scope not fully consumed
- `ErrUint64Range` — integer exceeds 8 bytes

## Subpackages

| Subpackage | Description |
|------------|-------------|
| [`rlperrors/`](./rlperrors/) | Typed error constants for RLP decoding violations |

## Usage

```go
// Encode a struct.
type MyTx struct {
    Nonce    uint64
    GasPrice *big.Int
    To       [20]byte
}
tx := MyTx{Nonce: 1, GasPrice: big.NewInt(1e9)}
encoded, err := rlp.EncodeToBytes(tx)

// Decode back.
var decoded MyTx
if err := rlp.DecodeBytes(encoded, &decoded); err != nil {
    // handle error
}

// High-throughput encoding with a pool.
pool := rlp.NewEncoderPool()
batch, err := pool.EncodeBatch([]interface{}{tx1, tx2, tx3})

// Stream-based decoding.
s := rlp.NewStreamFromBytes(data)
if _, err := s.List(); err != nil { /* ... */ }
nonce, _ := s.Uint64()
s.ListEnd()

// Zero-allocation encoding for known types.
buf := rlp.AppendListHeader(nil, 33)
buf = rlp.AppendUint64(buf, 42)
buf = append(buf, rlp.EncodeBytes32(hashVal)...)
```

## Documentation References

- [Yellow Paper, Appendix B — RLP](https://ethereum.github.io/yellowpaper/paper.pdf)
- [go-ethereum RLP implementation](../../refs/go-ethereum/rlp/)
