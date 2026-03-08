# debugapi — debug namespace JSON-RPC methods

[← rpc](../README.md)

## Overview

Package `debugapi` implements the `debug_` namespace JSON-RPC methods for
block introspection, transaction tracing, RLP encoding, database inspection,
and chain management. It is backed by the `rpcbackend.Backend` interface so it
stays decoupled from the chain internals.

An extended API (`debug_ext_api.go`) adds `dbg_` methods for lower-level
diagnostics, and `types.go` defines the trace result structures shared across
the package.

## Functionality

**Types**

- `DebugAPI` — constructed with `NewDebugAPI(backend rpcbackend.Backend)`
- `TraceResult` — opcode-level trace: `Gas uint64`, `Failed bool`, `ReturnValue string`, `StructLogs []StructLog`
- `StructLog` — single EVM step: `Pc`, `Op`, `Gas`, `GasCost`, `Depth`, `Stack`, `Memory`, `Storage`, `Reason`
- `DebugBlockTraceEntry` — per-transaction trace wrapper: `TxHash string`, `Result *TraceResult`

**Dispatched methods (via `DebugAPI.HandleDebugRequest`)**

| JSON-RPC method | Description |
|---|---|
| `debug_traceBlockByNumber` | Trace all txs in a block by number; returns `[]DebugBlockTraceEntry` |
| `debug_traceBlockByHash` | Same by block hash |
| `debug_getBlockRlp` | RLP-encoded block as hex string |
| `debug_printBlock` | Human-readable block summary |
| `debug_chaindbProperty` | DB property (`leveldb.stats`, `leveldb.iostats`, `version`) |
| `debug_chaindbCompact` | Trigger database compaction |
| `debug_setHead` | Rewind chain head to a block number |
| `debug_freeOSMemory` | Force GC and return memory to OS (`runtime.GC` + `debug.FreeOSMemory`) |
