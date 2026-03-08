# Package enr

Ethereum Node Records (EIP-778) — RLP encoding/decoding, "v4" identity-scheme signing, and signature verification for devp2p node identity.

## Overview

The `enr` package implements the Ethereum Node Record specification (EIP-778). A `Record` is an ordered set of key/value pairs (always sorted by key) plus a sequence number and a 64-byte secp256k1 signature. The wire format is an RLP list: `[sig, seq, k1, v1, k2, v2, ...]`. The maximum encoded size is 300 bytes.

`SignENR` sets the `"id"` and `"secp256k1"` entries, then signs the content list with a private key using keccak256. `VerifyENR` recovers the public key from the `"secp256k1"` entry and checks the signature. `NodeID` derives the 32-byte node identity by hashing the compressed public key.

## Functionality

- `Record` — `{Seq uint64, Pairs []Pair, Signature []byte}`
  - `Set(key string, value []byte)` — upserts and re-sorts; clears signature
  - `Get(key string) []byte`
  - `SetSeq(seq uint64)`
  - `NodeID() [32]byte` — keccak256 of the `"secp256k1"` entry

- `EncodeENR(r *Record) ([]byte, error)` — RLP-encode; enforces 300-byte size limit
- `DecodeENR(data []byte) (*Record, error)` — RLP-decode; validates sorted keys, no duplicates
- `SignENR(r *Record, key *ecdsa.PrivateKey) error` — "v4" identity scheme
- `VerifyENR(r *Record) error`

- Standard key constants: `KeyID`, `KeySecp256k1`, `KeyIP`, `KeyTCP`, `KeyUDP`, `KeyIP6`, `KeyTCP6`, `KeyUDP6`
- Errors: `ErrInvalidSig`, `ErrTooBig`, `ErrNotSigned`, `ErrNotSorted`, `ErrDuplicateKey`

## Usage

```go
r := &enr.Record{}
r.SetSeq(1)
r.Set(enr.KeyIP, ip.To4())
r.Set(enr.KeyTCP, []byte{0x75, 0xF0}) // port 30192

if err := enr.SignENR(r, privateKey); err != nil {
    log.Fatal(err)
}
encoded, _ := enr.EncodeENR(r)

decoded, _ := enr.DecodeENR(encoded)
if err := enr.VerifyENR(decoded); err != nil {
    log.Fatal("bad signature:", err)
}
```

[← p2p](../README.md)
