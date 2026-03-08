# engine/requests — EIP-7685 execution layer triggered request handling

## Overview

Package `requests` implements the EIP-7685 execution layer triggered request framework used by the Engine API. It provides parsing, validation, and hashing of the three canonical request types — deposits, withdrawals, and consolidations — that appear in `ExecutionPayloadV5` and later.

Each request is a typed byte blob. The package enforces the EIP-7685 rule that request types must appear in strictly ascending order within a payload, and it computes the canonical `RequestsHash` by Keccak-256-hashing the concatenated raw request bytes for each type in order.

## Functionality

**Types**
- `ExecutionRequest{Type byte, Data []byte}` — a single typed request blob

**Type constants**
- `ExecReqDepositType = 0x00`
- `ExecReqWithdrawalType = 0x01`
- `ExecReqConsolidationType = 0x02`

**Per-type data sizes** (bytes per item)
- Deposit: 192
- Withdrawal: 76
- Consolidation: 116

**Functions**
- `ParseExecutionRequests(raw [][]byte) ([]*ExecutionRequest, error)` — decodes wire-format request list into typed structs
- `ValidateExecutionRequestList(reqs []*ExecutionRequest) error` — enforces ascending type order required by EIP-7685
- `ComputeExecutionRequestsHash(reqs []*ExecutionRequest) (Hash, error)` — Keccak-256 of concatenated raw bytes per type; produces the `RequestsHash` field in the block header
- `SplitRequestsByType(reqs []*ExecutionRequest) map[byte][]*ExecutionRequest` — partitions a mixed list by type code
- `CountDepositRequests(reqs []*ExecutionRequest) int`
- `CountWithdrawalRequests(reqs []*ExecutionRequest) int`
- `CountConsolidationRequests(reqs []*ExecutionRequest) int`

Parent package: [`engine`](../README.md)
