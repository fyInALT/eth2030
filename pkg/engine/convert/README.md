# engine/convert — Execution payload version conversion utilities

Converts between execution payload versions (V1–V5) and `core/types.Header`,
extracts versioned hashes from transactions, and provides fork-aware payload
version selection.

## Overview

Each `PayloadToHeader*` function maps an Engine API payload struct to a
`types.Header` by copying relevant fields (parent hash, state root, gas fields,
blob gas, etc.). The inverse `HeaderToPayload*` functions extract payload fields
from a header for constructing Engine API responses.

`DeterminePayloadVersion` selects the highest applicable `PayloadVersion` for a
given timestamp given a `ForkTimestamps` configuration (Shanghai, Cancun, Prague,
Amsterdam). Upgrade helpers `ConvertV1ToV2`, `ConvertV2ToV3`, `ConvertV3ToV4`,
`ConvertV4ToV5` provide zero-value defaults for new fields.

## Functionality

**Types**
- `PayloadVersion` — `PayloadV1` … `PayloadV5`
- `ForkTimestamps` — `Shanghai`, `Cancun`, `Prague`, `Amsterdam`
- `WithdrawalsSummary` — `Count`, `TotalAmountGwei`, `UniqueValidators`, `UniqueAddresses`

**Functions**
- `PayloadToHeaderV1/V2/V3/V5` — payload → `*types.Header`
- `HeaderToPayloadV2/V3` — header → payload V2/V3
- `ExtractVersionedHashes(txBytes [][]byte) []types.Hash`
- `VersionedHashFromCommitment(commitment []byte) types.Hash`
- `BlobSidecarFromBundle(bundle, index, blockHash) (*blobsbundle.BlobSidecar, error)`
- `DeterminePayloadVersion(timestamp, forks) PayloadVersion`
- `ConvertV1ToV2`, `ConvertV2ToV3`, `ConvertV3ToV4`, `ConvertV4ToV5`
- `ValidatePayloadConsistency(p *ExecutionPayloadV3) bool`
- `ProcessWithdrawalsExt(withdrawals) (totalGwei, byValidator)`
- `CoreWithdrawalsFromPayload(p) []*types.Withdrawal`
- `SummarizeWithdrawals(withdrawals) WithdrawalsSummary`

## Usage

```go
header := convert.PayloadToHeaderV3(payload)

ver := convert.DeterminePayloadVersion(timestamp, &convert.ForkTimestamps{
    Cancun: cancunTime, Prague: pragueTime,
})

hashes := convert.ExtractVersionedHashes(payload.Transactions)
```

[← engine](../README.md)
