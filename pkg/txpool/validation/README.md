# txpool/validation — Multi-stage transaction validation pipeline

## Overview

Package `validation` provides two validation layers. `TxValidator` performs configurable single-transaction checks (gas price floor, gas limit ceiling, data size, chain ID, signature). `ValidationPipeline` chains five stages in order — per-peer rate limiting, syntax, signature (ECDSA + EIP-7702 SetCode authorization list), on-chain state (nonce and balance), and blob-specific checks (versioned hashes, blob fee cap) — each reporting a typed `ValidationErrorCode` on failure to enable precise pool-level error handling.

## Functionality

**Types**
- `TxValidator` — `ValidateTx`, `ValidateBasic`, `ValidateGas`, `ValidateSize`, `ValidateChainID`, `ValidateSignature`, `ValidateBatch`
- `TxValidationConfig{MinGasPrice, MaxGasLimit, MaxDataSize, MaxValueWei, ChainID}` — configurable bounds
- `ValidationPipeline` — `Validate(tx, sender, peerID)`, `ValidateBatch`, `RateLimiter()`
- `ValidationPipelineConfig{MaxGasLimit, MaxDataSize, MaxNonceGap, BaseFee, BlobBaseFee, MaxPerPeerRate, RateWindow}`
- `SyntaxCheck`, `SignatureVerify`, `StateCheck`, `BlobCheck`, `RateLimiter` — individual pipeline stages
- `ValidationResult{Valid, ErrorCode, Error, Stages}` — pipeline output
- `ValidationErrorCode` — `ValidationSyntaxErr`, `ValidationSignatureErr`, `ValidationStateErr`, `ValidationBlobErr`, `ValidationRateLimitErr`

**Functions**
- `NewValidationPipeline(config, state)` — wires all five stages
- `DefaultTxValidationConfig`, `DefaultValidationPipelineConfig`

## Usage

```go
vp := validation.NewValidationPipeline(validation.DefaultValidationPipelineConfig(), stateDB)
result := vp.Validate(tx, sender, peerID)
if !result.Valid {
    switch result.ErrorCode {
    case validation.ValidationRateLimitErr: // drop silently
    case validation.ValidationStateErr:     // penalize peer
    }
}
```

[← txpool](../README.md)
