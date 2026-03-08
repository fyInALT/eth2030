# engine/payload — Execution payload types, building, caching, and validation

## Overview

Package `payload` defines every Engine API payload type (V1 through V5) and the supporting infrastructure for constructing, caching, processing, and validating execution payloads. It covers the full lifecycle from `PayloadAttributes` received via `engine_forkchoiceUpdatedV*` through async block building to the final payload returned by `engine_getPayload*`.

Validation is split into two complementary layers: `PayloadProcessor` performs intrinsic header checks (base-fee formula, gas limits, timestamp ordering) and drives state-transition processing; `PayloadValidator` performs structural integrity checks on every field in isolation, including RLP-based block hash recomputation, blob gas accounting, and parent beacon block root verification. An optional STARK prover may replace validation frames in built payloads (US-PQ-5b).

## Functionality

**Core payload types**
- `PayloadID [8]byte`
- `Withdrawal{Index, ValidatorIndex, Address, Amount}`
- `ExecutionPayloadV1` through `ExecutionPayloadV5` — successive Engine API payload versions (V4 adds withdrawals+blob gas; V5 adds execution requests and block access list hash)
- `PayloadAttributesV1` through `PayloadAttributesV4` — V4 adds `SlotNumber` and `InclusionListTransactions` for FOCIL (EIP-7805)
- `GetPayloadV3Response`, `GetPayloadV4Response`, `GetPayloadV6Response` — response wrappers including blobs bundle and block value
- `BlobsBundleV1{Commitments, Proofs, Blobs [][]byte}`

**PayloadCache** (`cache.go`)
- `PayloadCacheConfig{MaxPayloads int, PayloadTTL time.Duration, MaxPayloadSize int64}` — default: 32 payloads, 120 s TTL, 10 MiB
- `NewPayloadCache(cfg PayloadCacheConfig) *PayloadCache`
- `Store(id PayloadID, payload *BuiltPayload) error`
- `Get(id PayloadID) (*BuiltPayload, bool)`
- `Delete(id PayloadID)`
- `Prune() int` — removes TTL-expired entries
- `Size() int` / `TotalBytes() int64`

**PayloadBuilder** (`builder.go`)
- `BuiltPayload{Block, Receipts, BlockValue *big.Int, BlobsBundle, ExecutionRequests, BAL}`
- `NewPayloadBuilder(stateDB, chainConfig, signer) *PayloadBuilder`
- `SetValidationFrameProver(prover)` — optional STARK prover for US-PQ-5b
- `StartBuild(ctx, attrs PayloadAttributes) (PayloadID, error)`
- `GetPayload(id PayloadID) (*BuiltPayload, error)`

**PayloadProcessor** (`processor.go`)
- Constants: `BaseFeeChangeDenominator=8`, `ElasticityMultiplier=2`, `MaxExtraDataSize=32`, `MinGasLimit=5000`, `GasLimitBoundDivisor=1024`
- `ProcessResult{StateRoot, ReceiptsRoot, LogsBloom, GasUsed}`
- `NewPayloadProcessor(cfg) *PayloadProcessor`
- `ValidatePayload(payload) error` — header-level checks
- `ValidateBlockHash`, `ValidateGasLimits`, `ValidateTimestamp`, `ValidateBaseFee`
- `ProcessPayload(payload) (*ProcessResult, error)` — drives state transition
- `CalcBaseFee(parentBaseFee, parentGasUsed, parentGasLimit *big.Int) *big.Int`

**PayloadValidator** (`validation.go`)
- `NewPayloadValidator(cfg) *PayloadValidator`
- `ValidatePayloadFull(payload) []error` — runs all checks, returns all failures
- `ValidateBlockHashComputed(payload) error` — RLP-encodes header, verifies Keccak-256
- `ValidateTimestamp`, `ValidateBaseFee`, `ValidateGasLimit`, `ValidateExtraData`
- `ValidateTransactions(txs) error` — max 1 M txs, 16 MiB per tx
- `ValidateBlobGasUsed`, `ValidateWithdrawals` — max 16 withdrawals
- `ValidateParentBeaconBlockRoot(payload, expected Hash) error`
- `CalcBaseFeeBig(parentBaseFee, parentGasUsed, parentGasLimit *big.Int) *big.Int`

Parent package: [`engine`](../README.md)
