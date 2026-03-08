# engine/errors — Engine API error definitions and utilities

## Overview

Package `errors` defines the complete set of error sentinels, JSON-RPC error codes, and helper utilities for the Engine API. It codifies the error taxonomy specified in the execution-apis spec so that every other engine sub-package can import a single, consistent error vocabulary without pulling in heavier dependencies.

The package covers standard JSON-RPC 2.0 codes (-32700 to -32603), Engine API-specific codes (-38001 to -38005), and extended server codes (-32005/-32006). The structured `EngineError` type carries a numeric code and an optional cause, and marshals directly to a JSON-RPC error object.

## Functionality

**Error sentinels** — `ErrInvalidParams`, `ErrUnknownPayload`, `ErrInvalidForkchoiceState`, `ErrInvalidPayloadAttributes`, `ErrUnsupportedFork`, `ErrInvalidBlockHash`, `ErrInvalidBlobHashes`, `ErrMissingBeaconRoot`, `ErrMissingWithdrawals`, `ErrMissingExecutionRequests`, `ErrMissingBlockAccessList`, `ErrServerBusy`, `ErrRequestTimeout`, `ErrPayloadNotBuilding`, `ErrPayloadTimestamp`, `ErrInvalidTerminalBlock`

**Status strings** — `StatusValid`, `StatusInvalid`, `StatusSyncing`, `StatusAccepted`, `StatusInvalidBlockHash`

**Types** — `EngineError` with `Code`, `Message`, `Cause` fields; implements `error`, `Unwrap`, and `MarshalJSON`

**Functions**
- `NewEngineError(code int, message string) *EngineError`
- `WrapEngineError(code int, message string, cause error) *EngineError`
- `ErrorCodeFromError(err error) int` — maps known sentinels to their JSON-RPC code
- `ErrorName(code int) string` — human-readable code name
- `ErrorResponse(id json.RawMessage, code int, message string) []byte` — full JSON-RPC error response bytes
- `ValidatePayloadVersion(version int, hasWithdrawals, hasExecutionRequests, hasBlockAccessList bool) *EngineError`
- `IsClientError`, `IsServerError`, `IsEngineError` — code-range predicates

Parent package: [`engine`](../README.md)
