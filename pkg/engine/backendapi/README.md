# engine/backendapi — Engine API backend interfaces

Defines the Go interfaces that the execution layer must implement for the Engine
API to function. Sub-packages import from here to avoid circular dependencies
with the top-level `engine` package.

## Overview

All Engine API handler packages (`engine/api`, `engine/blocks`, etc.) depend
only on the interfaces in this package, not on any concrete engine
implementation. This allows the EL backend to be swapped or mocked in tests.

## Functionality

**Interfaces**
- `Backend` — core Engine API backend: `ProcessBlock`, `ProcessBlockV4/V5`, `ForkchoiceUpdated/V4`, `GetPayloadByID`, `GetPayloadV4ByID/V6ByID`, fork predicates (`IsCancun`, `IsPrague`, `IsAmsterdam`), head/safe/finalized hash accessors
- `V4Backend` — minimal subset for `EngV4`: `GetPayloadV4ByID`, `IsPrague`
- `GlamsterdamBackend` — post-Glamsterdam: `NewPayloadV5`, `ForkchoiceUpdatedV4G`, `GetPayloadV5`, `GetBlobsV2`
- `EngineV7Backend` — K+ era: `NewPayloadV7`, `ForkchoiceUpdatedV7`, `GetPayloadV7`
- `InclusionListBackend` — EIP-7805 FOCIL: `ProcessInclusionList`, `GetInclusionList`
- `PayloadBodiesBackend` — `GetPayloadBodiesByHash`, `GetPayloadBodiesByRange`
- `UncoupledBackend` — EIP-7898 uncoupled payloads (extension point)

## Usage

Implement `Backend` on your execution layer and pass it to the engine server:

```go
// In your EL node:
var _ backendapi.Backend = (*MyELBackend)(nil)

// In engine server constructor:
func NewEngineServer(b backendapi.Backend) *EngineServer { ... }
```

[← engine](../README.md)
