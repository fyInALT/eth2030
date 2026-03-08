# engine/api — Engine API sub-handlers (V4 / Glamsterdam / V7)

Implements Engine API method logic for Prague (V4), Glamsterdam (V5), and the
V7 2028-era fork. Types and handlers here are imported by the main `engine`
package and re-exported via type aliases.

## Overview

`v4.go` provides `EngV4` with Prague/Electra-era logic: EIP-7685 execution
requests (deposits, withdrawals, consolidations) validation, encoding, decoding,
and `GetPayloadV4`. It enforces type-byte ascending order and per-type size
limits.

`glamsterdam.go` handles EIP-7805 (FOCIL) inclusion-list aware payloads and V5
payload attributes. `v7.go` exposes the V7 handler for the K+ era.

`uncoupled.go` provides an `UncoupledPayloadHandler` for EIP-7898 uncoupled
execution payloads, separating the CL and EL payload lifecycles.

`epbs.go` contains ePBS (EIP-7732) signing domain constants and helper
utilities used by the builder API.

## Functionality

**Types**
- `EngV4` — Prague payload handler backed by `V4Backend`
- `DepositRequest`, `WithdrawalRequest`, `ConsolidationRequest` — EIP-7685 structs
- `ExecutionRequestsV4` — categorised request container
- `GetPayloadV4Result` — payload + block value + blobs bundle + requests

**Functions**
- `NewEngV4(backend V4Backend) *EngV4`
- `(*EngV4).GetPayloadV4(payloadID) (*GetPayloadV4Result, error)`
- `ValidateExecutionRequests(requests [][]byte) error`
- `BuildExecutionRequestsList(deposits, withdrawals, consolidations) [][]byte`
- `ExecutionRequestsHash(requests) types.Hash`
- `ClassifyExecutionRequests(requests) (*ExecutionRequestsV4, error)`
- `DecodeDepositRequests`, `DecodeWithdrawalRequests`, `DecodeConsolidationRequests`
- `EncodeDepositRequest`, `EncodeWithdrawalRequest`, `EncodeConsolidationRequest`

## Usage

```go
v4 := api.NewEngV4(myBackend)
result, err := v4.GetPayloadV4(payloadID)

err = api.ValidateExecutionRequests(result.ExecutionRequests)
classified, _ := api.ClassifyExecutionRequests(result.ExecutionRequests)
fmt.Println("deposits:", len(classified.Deposits))
```

[← engine](../README.md)
