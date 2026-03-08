# engine/builder — ePBS builder types (EIP-7732)

Defines the data types for Enshrined Proposer-Builder Separation (EIP-7732):
builder registration, sealed execution payload bids, signed envelopes, and BLS
key/signature types.

## Overview

In EIP-7732 (ePBS), builders commit to a block by submitting a signed
`ExecutionPayloadBid` that contains the block hash, gas limit, slot, and
payment value without revealing the full payload. After the proposer commits to
a winning bid the builder reveals the full `ExecutionPayloadEnvelope`.

`registry.go` provides a `BuilderRegistry` for managing registered builders and
their lifecycle status (active, exiting, withdrawn).

## Functionality

**Types**
- `BLSPubkey` — `[48]byte` BLS12-381 public key
- `BLSSignature` — `[96]byte` BLS12-381 signature
- `BuilderIndex` — uint64 registry index
- `BuilderStatus` — `Active`, `Exiting`, `Withdrawn`
- `Builder` — registered builder: `Pubkey`, `Index`, `FeeRecipient`, `GasLimit`, `Balance`, `Status`
- `ExecutionPayloadBid` — sealed bid: `ParentBlockHash`, `BlockHash`, `Slot`, `Value`, `BuilderIndex`, `BlobKZGCommitments`
- `SignedExecutionPayloadBid` — bid + BLS signature
- `ExecutionPayloadEnvelope` — revealed payload: `Payload`, `ExecutionRequests`, `BuilderIndex`, `Slot`
- `SignedExecutionPayloadEnvelope` — envelope + BLS signature
- `BuilderRegistrationV1` / `SignedBuilderRegistrationV1` — registration message

**Functions**
- `(*ExecutionPayloadBid).BidHash() types.Hash` — deterministic bid hash for signing

## Usage

```go
bid := &builder.ExecutionPayloadBid{
    Slot:         12345,
    BuilderIndex: myIndex,
    BlockHash:    blockHash,
    Value:        100, // Gwei to proposer
}
h := bid.BidHash() // sign h with builder's BLS key
signed := &builder.SignedExecutionPayloadBid{Message: *bid, Signature: sig}
```

[← engine](../README.md)
